package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ops-server/internal/domain/user/models"
	"ops-server/internal/domain/user/service"
	"ops-server/internal/interfaces/http/response"
	appErrors "ops-server/pkg/errors"
	"ops-server/pkg/logger"
)

// UserController gère les requêtes HTTP du domaine utilisateur.
// La gestion des rôles est déléguée au RBACController (/api/v1/rbac).
type UserController struct {
	svc service.UserService
}

func NewUserController(svc service.UserService) *UserController {
	return &UserController{svc: svc}
}

// Register godoc
// @Summary      Créer un compte
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateUserInput  true  "Données d'inscription"
// @Success      201   {object}  response.APIResponse{data=models.UserResponse}
// @Failure      400   {object}  response.ErrorResponse
// @Failure      409   {object}  response.ErrorResponse
// @Router       /auth/register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var input models.CreateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	logger.FromContext(ctx.Request.Context()).Info("register", zap.String("identifier", input.Identifier))
	user, err := c.svc.Register(ctx.Request.Context(), &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Created(ctx, "user registered successfully", user)
}

// SignIn godoc
// @Summary      Authentification
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.SignInInput  true  "Identifiants"
// @Success      200   {object}  response.APIResponse{data=models.AuthResponse}
// @Failure      401   {object}  response.ErrorResponse
// @Router       /auth/signin [post]
func (c *UserController) SignIn(ctx *gin.Context) {
	var input models.SignInInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	auth, err := c.svc.SignIn(ctx.Request.Context(), &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "sign-in successful", auth)
}

// SignInLDAP godoc
// @Summary      Authentification LDAP
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.SignInInput  true  "Credentials LDAP"
// @Success      200   {object}  response.APIResponse{data=models.AuthResponse}
// @Failure      401   {object}  response.ErrorResponse
// @Router       /auth/signin-ldap [post]
func (c *UserController) SignInLDAP(ctx *gin.Context) {
	var input models.SignInInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}

	resp, err := c.svc.SignInLDAP(ctx.Request.Context(), &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "login success", resp)
}

// RefreshToken godoc
// @Summary      Renouveler le token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      models.RefreshTokenInput  true  "Refresh token"
// @Success      200   {object}  response.APIResponse{data=models.AuthResponse}
// @Failure      401   {object}  response.ErrorResponse
// @Router       /auth/refresh [post]
func (c *UserController) RefreshToken(ctx *gin.Context) {
	var input models.RefreshTokenInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	auth, err := c.svc.RefreshToken(ctx.Request.Context(), input.RefreshToken)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "token refreshed", auth)
}

// Logout godoc
// @Summary      Déconnexion
// @Tags         auth
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse
// @Router       /auth/logout [post]
func (c *UserController) Logout(ctx *gin.Context) {
	uid, err := currentUserID(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	if err := c.svc.Logout(ctx.Request.Context(), uid); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "logged out")
}

// GetMe godoc
// @Summary      Mon profil
// @Tags         users
// @Security     BearerAuth
// @Success      200  {object}  response.APIResponse{data=models.UserResponse}
// @Router       /users/me [get]
func (c *UserController) GetMe(ctx *gin.Context) {
	uid, err := currentUserID(ctx)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	user, err := c.svc.GetByID(ctx.Request.Context(), uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "user retrieved", user)
}

// GetUser godoc
// @Summary      Obtenir un utilisateur (admin)
// @Tags         users
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID utilisateur"
// @Success      200  {object}  response.APIResponse{data=models.UserResponse}
// @Failure      404  {object}  response.ErrorResponse
// @Router       /users/{id} [get]
func (c *UserController) GetUser(ctx *gin.Context) {
	uid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid user id"))
		return
	}
	user, err := c.svc.GetByID(ctx.Request.Context(), uid)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "user retrieved", user)
}

// UpdateUser godoc
// @Summary      Mettre à jour un utilisateur
// @Tags         users
// @Security     BearerAuth
// @Param        id    path  string                  true  "UUID utilisateur"
// @Param        body  body  models.UpdateUserInput  true  "Champs à modifier"
// @Success      200   {object}  response.APIResponse{data=models.UserResponse}
// @Failure      404   {object}  response.ErrorResponse
// @Router       /users/{id} [patch]
func (c *UserController) UpdateUser(ctx *gin.Context) {
	uid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid user id"))
		return
	}
	var input models.UpdateUserInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	user, err := c.svc.Update(ctx.Request.Context(), uid, &input)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, http.StatusOK, "user updated", user)
}

// DeleteUser godoc
// @Summary      Supprimer un utilisateur (soft-delete)
// @Tags         users
// @Security     BearerAuth
// @Param        id  path  string  true  "UUID utilisateur"
// @Success      200  {object}  response.APIResponse
// @Failure      404  {object}  response.ErrorResponse
// @Router       /users/{id} [delete]
func (c *UserController) DeleteUser(ctx *gin.Context) {
	uid, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, appErrors.BadRequest("invalid user id"))
		return
	}
	if err := c.svc.Delete(ctx.Request.Context(), uid); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Ok(ctx, "user deleted")
}

// ListUsers godoc
// @Summary      Lister les utilisateurs (admin)
// @Tags         users
// @Security     BearerAuth
// @Param        page   query  int  false  "Page (défaut: 1)"
// @Param        limit  query  int  false  "Limite (défaut: 20, max: 100)"
// @Success      200  {object}  response.APIResponse{data=response.PaginatedData[models.UserResponse]}
// @Router       /users [get]
func (c *UserController) ListUsers(ctx *gin.Context) {
	var q struct {
		Page  int `form:"page,default=1"`
		Limit int `form:"limit,default=20"`
	}
	if err := ctx.ShouldBindQuery(&q); err != nil {
		response.Error(ctx, appErrors.ValidationError(err.Error()))
		return
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit > 100 {
		q.Limit = 100
	}

	offset := response.PageToOffset(q.Page, q.Limit)
	users, total, err := c.svc.List(ctx.Request.Context(), offset, q.Limit)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Paginated(ctx, users, total, q.Page, q.Limit)
}

func currentUserID(ctx *gin.Context) (uuid.UUID, error) {
	val, exists := ctx.Get("userId")
	if !exists {
		return uuid.Nil, appErrors.Unauthorized("not authenticated")
	}
	uid, err := uuid.Parse(val.(string))
	if err != nil {
		return uuid.Nil, appErrors.BadRequest("invalid user id in token")
	}
	return uid, nil
}
