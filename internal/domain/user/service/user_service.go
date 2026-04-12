package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ops-server/configs"
	appErrors "ops-server/pkg/errors"
	"ops-server/pkg/logger"
	"ops-server/pkg/utils"

	"ops-server/internal/infrastructure/ldap"

	"ops-server/internal/domain/user/models"
	"ops-server/internal/domain/user/repository"
	redisInfra "ops-server/internal/infrastructure/redis"
)

// UserService définit les opérations métier sur les utilisateurs.
// La gestion des rôles et permissions est déléguée au domaine rbac.
type UserService interface {
	Register(ctx context.Context, input *models.CreateUserInput) (*models.UserResponse, error)
	SignIn(ctx context.Context, input *models.SignInInput) (*models.AuthResponse, error)
	SignInLDAP(ctx context.Context, input *models.SignInInput) (*models.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*models.AuthResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserResponse, error)
	Update(ctx context.Context, id uuid.UUID, input *models.UpdateUserInput) (*models.UserResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*models.UserResponse, int64, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

type userService struct {
	repo    repository.UserRepository
	cache   redisInfra.Cache
	jwtCfg  configs.JWTConfig
	ldapSvc ldap.LdapService
}

func NewUserService(
	repo repository.UserRepository,
	cache redisInfra.Cache,
	jwtCfg configs.JWTConfig,
	ldapSvc ldap.LdapService,
) UserService {
	return &userService{repo: repo, cache: cache, jwtCfg: jwtCfg, ldapSvc: ldapSvc}
}

func (s *userService) Register(ctx context.Context, input *models.CreateUserInput) (*models.UserResponse, error) {
	log := logger.FromContext(ctx)

	exists, err := s.repo.ExistsByIdentifier(ctx, input.Identifier)
	if err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to check identifier", err)
	}
	if exists {
		return nil, appErrors.New(appErrors.ErrCodeIdentifierTaken, "identifier already in use")
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, appErrors.Internal(err)
	}

	user := &models.User{
		Identifier: input.Identifier,
		Email:      input.Email,
		Password:   hash,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		IsActive:   true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to create user", err)
	}

	log.Info("user registered", zap.String("userId", user.ID.String()))
	return user.ToResponse(), nil
}

// SignIn authentifie un utilisateur et retourne des tokens JWT.
func (s *userService) SignIn(ctx context.Context, input *models.SignInInput) (*models.AuthResponse, error) {
	user, err := s.repo.FindByIdentifier(ctx, input.Identifier)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.ErrCodeInvalidCredentials, "invalid identifier or password")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to find user", err)
	}

	if !user.IsActive {
		return nil, appErrors.New(appErrors.ErrCodeUserDisabled, "account is disabled")
	}
	if !utils.CheckPasswordHash(input.Password, user.Password) {
		return nil, appErrors.New(appErrors.ErrCodeInvalidCredentials, "invalid identifier or password")
	}

	// Récupérer les rôles pour les inclure dans le token
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, appErrors.Internal(err)
	}
	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, appErrors.Internal(err)
	}

	ttl := time.Duration(s.jwtCfg.RefreshTTL) * time.Minute
	if err := s.cache.Set(ctx, refreshTokenKey(user.ID), refreshToken, ttl); err != nil {
		return nil, appErrors.Internal(fmt.Errorf("failed to store refresh token: %w", err))
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

func (s *userService) SignInLDAP(ctx context.Context, input *models.SignInInput) (*models.AuthResponse, error) {
	// 1. LDAP AUTH
	ldapUser, err := s.ldapSvc.Authenticate(ctx, input.Identifier, input.Password)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrCodeInvalidCredentials, "invalid ldap credentials")
	}

	// 2. CHECK USER EXISTS
	user, err := s.repo.FindByIdentifier(ctx, input.Identifier)

	if err != nil {
		// USER NOT FOUND → AUTO CREATE 🔥
		if errors.Is(err, gorm.ErrRecordNotFound) {

			user = &models.User{
				Identifier: ldapUser.Username,
				Email:      ldapUser.Email,
				FirstName:  ldapUser.FirstName,
				LastName:   ldapUser.LastName,
				IsActive:   true,
				Password:   "", // ⚠️ no local password
			}

			if err := s.repo.Create(ctx, user); err != nil {
				return nil, appErrors.Internal(err)
			}

		} else {
			return nil, appErrors.Internal(err)
		}
	}

	// 3. UPDATE USER (sync LDAP → DB)
	user.Email = ldapUser.Email
	user.FirstName = ldapUser.FirstName
	user.LastName = ldapUser.LastName

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, appErrors.Internal(err)
	}

	// 4. GENERATE TOKENS
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, appErrors.Internal(err)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, appErrors.Internal(err)
	}

	// 5. STORE REFRESH TOKEN (Redis)
	ttl := time.Duration(s.jwtCfg.RefreshTTL) * time.Minute
	if err := s.cache.Set(ctx, refreshTokenKey(user.ID), refreshToken, ttl); err != nil {
		return nil, appErrors.Internal(err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ToResponse(),
	}, nil
}

// RefreshToken renouvelle la paire de tokens.
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	claims, err := s.parseToken(refreshToken)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrCodeInvalidToken, "invalid or expired refresh token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrCodeInvalidToken, "malformed token subject")
	}

	stored, err := s.cache.Get(ctx, refreshTokenKey(userID))
	if err != nil || stored != refreshToken {
		return nil, appErrors.New(appErrors.ErrCodeInvalidToken, "refresh token revoked or not found")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrCodeUserNotFound, "user not found")
	}

	newAccess, err := s.generateAccessToken(user)
	if err != nil {
		return nil, appErrors.Internal(err)
	}
	newRefresh, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, appErrors.Internal(err)
	}

	ttl := time.Duration(s.jwtCfg.RefreshTTL) * time.Minute
	_ = s.cache.Set(ctx, refreshTokenKey(userID), newRefresh, ttl)

	return &models.AuthResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		User:         user.ToResponse(),
	}, nil
}

// GetByID retourne un utilisateur par UUID.
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*models.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.ErrCodeUserNotFound, "user not found")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch user", err)
	}
	return user.ToResponse(), nil
}

// Update applique une mise à jour partielle sur un utilisateur.
func (s *userService) Update(ctx context.Context, id uuid.UUID, input *models.UpdateUserInput) (*models.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.New(appErrors.ErrCodeUserNotFound, "user not found")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch user", err)
	}
	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to update user", err)
	}
	return user.ToResponse(), nil
}

// Delete effectue un soft-delete.
func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.New(appErrors.ErrCodeUserNotFound, "user not found")
		}
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch user", err)
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to delete user", err)
	}
	_ = s.cache.Delete(ctx, refreshTokenKey(id))
	return nil
}

// List retourne une liste paginée d'utilisateurs.
func (s *userService) List(ctx context.Context, offset, limit int) ([]*models.UserResponse, int64, error) {
	users, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list users", err)
	}
	resp := make([]*models.UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, u.ToResponse())
	}
	return resp, total, nil
}

// Logout révoque le refresh token Redis.
func (s *userService) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.cache.Delete(ctx, refreshTokenKey(userID))
}

// ── JWT ───────────────────────────────────────────────────────────────────────

type jwtClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (s *userService) generateAccessToken(user *models.User) (string, error) {
	claims := jwtClaims{
		Roles: user.RoleNames(),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtCfg.AccessTTL) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtCfg.Secret))
}

func (s *userService) generateRefreshToken(user *models.User) (string, error) {
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.jwtCfg.RefreshTTL) * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwtCfg.Secret))
}

func (s *userService) parseToken(tokenStr string) (*jwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}

func refreshTokenKey(userID uuid.UUID) string {
	return fmt.Sprintf("refresh_token:%s", userID.String())
}
