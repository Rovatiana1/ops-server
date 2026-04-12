package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ops-server/internal/domain/rbac/models"
	"ops-server/internal/domain/rbac/repository"
	appErrors "ops-server/pkg/errors"
	"ops-server/pkg/logger"
)

// RBACService définit toutes les opérations métier du système RBAC.
type RBACService interface {
	// Roles
	CreateRole(ctx context.Context, input *models.CreateRoleInput) (*models.RoleResponse, error)
	GetRole(ctx context.Context, id uuid.UUID) (*models.RoleResponse, error)
	GetRoleByName(ctx context.Context, name models.RoleName) (*models.RoleResponse, error)
	UpdateRole(ctx context.Context, id uuid.UUID, input *models.UpdateRoleInput) (*models.RoleResponse, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ListRoles(ctx context.Context, filter models.RoleFilterInput) ([]*models.RoleSummaryResponse, int64, error)

	// Permissions
	CreatePermission(ctx context.Context, input *models.CreatePermissionInput) (*models.PermissionResponse, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*models.PermissionResponse, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, input *models.UpdatePermissionInput) (*models.PermissionResponse, error)
	DeletePermission(ctx context.Context, id uuid.UUID) error
	ListPermissions(ctx context.Context, filter models.PermissionFilterInput) ([]*models.PermissionResponse, int64, error)

	// Role <-> Permission
	SetRolePermissions(ctx context.Context, roleID uuid.UUID, input *models.AssignPermissionsInput) (*models.RoleResponse, error)
	AddRolePermission(ctx context.Context, roleID uuid.UUID, input *models.AddPermissionInput) (*models.RoleResponse, error)
	RemoveRolePermission(ctx context.Context, roleID, permID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.PermissionResponse, error)

	// User <-> Role
	AssignRoleToUser(ctx context.Context, userID uuid.UUID, input *models.AssignRoleInput, assignedBy uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.UserRoleResponse, error)

	// Checks (utilisés par les middlewares)
	UserHasRole(ctx context.Context, userID uuid.UUID, roleName models.RoleName) (bool, error)
	UserHasPermission(ctx context.Context, userID uuid.UUID, resource string, action models.PermissionAction) (bool, error)
}

type rbacService struct {
	repo repository.RBACRepository
}

// NewRBACService crée un RBACService avec le repository injecté.
func NewRBACService(repo repository.RBACRepository) RBACService {
	return &rbacService{repo: repo}
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (s *rbacService) CreateRole(ctx context.Context, input *models.CreateRoleInput) (*models.RoleResponse, error) {
	log := logger.FromContext(ctx)

	exists, err := s.repo.ExistsRoleByName(ctx, input.Name)
	if err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to check role name", err)
	}
	if exists {
		return nil, appErrors.Conflict("role name '" + input.Name.String() + "' already exists")
	}

	role := &models.Role{
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
		IsSystem:    input.IsSystem,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to create role", err)
	}

	log.Info("role created", zap.String("roleId", role.ID.String()), zap.String("name", role.Name.String()))
	return role.ToResponse(), nil
}

func (s *rbacService) GetRole(ctx context.Context, id uuid.UUID) (*models.RoleResponse, error) {
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch role", err)
	}
	return role.ToResponse(), nil
}

func (s *rbacService) GetRoleByName(ctx context.Context, name models.RoleName) (*models.RoleResponse, error) {
	role, err := s.repo.FindRoleByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch role", err)
	}
	return role.ToResponse(), nil
}

func (s *rbacService) UpdateRole(ctx context.Context, id uuid.UUID, input *models.UpdateRoleInput) (*models.RoleResponse, error) {
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch role", err)
	}

	if input.DisplayName != nil {
		role.DisplayName = *input.DisplayName
	}
	if input.Description != nil {
		role.Description = *input.Description
	}

	if err := s.repo.UpdateRole(ctx, role); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to update role", err)
	}
	return role.ToResponse(), nil
}

func (s *rbacService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	role, err := s.repo.FindRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("role")
		}
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch role", err)
	}
	if role.IsSystem {
		return appErrors.Forbidden("system roles cannot be deleted")
	}
	if err := s.repo.DeleteRole(ctx, id); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to delete role", err)
	}
	logger.FromContext(ctx).Info("role deleted", zap.String("roleId", id.String()))
	return nil
}

func (s *rbacService) ListRoles(ctx context.Context, filter models.RoleFilterInput) ([]*models.RoleSummaryResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	roles, total, err := s.repo.ListRoles(ctx, filter)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list roles", err)
	}

	resp := make([]*models.RoleSummaryResponse, 0, len(roles))
	for _, r := range roles {
		resp = append(resp, r.ToSummaryResponse())
	}
	return resp, total, nil
}

// ── Permissions ───────────────────────────────────────────────────────────────

func (s *rbacService) CreatePermission(ctx context.Context, input *models.CreatePermissionInput) (*models.PermissionResponse, error) {
	// Vérifie l'unicité resource:action
	_, err := s.repo.FindPermissionBySlug(ctx, input.Resource, input.Action)
	if err == nil {
		return nil, appErrors.Conflict("permission '" + input.Resource + ":" + string(input.Action) + "' already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to check permission", err)
	}

	perm := &models.Permission{
		Resource:    input.Resource,
		Action:      input.Action,
		Description: input.Description,
	}

	if err := s.repo.CreatePermission(ctx, perm); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to create permission", err)
	}

	logger.FromContext(ctx).Info("permission created",
		zap.String("permissionId", perm.ID.String()),
		zap.String("slug", perm.Slug()),
	)
	return perm.ToResponse(), nil
}

func (s *rbacService) GetPermission(ctx context.Context, id uuid.UUID) (*models.PermissionResponse, error) {
	perm, err := s.repo.FindPermissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("permission")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch permission", err)
	}
	return perm.ToResponse(), nil
}

func (s *rbacService) UpdatePermission(ctx context.Context, id uuid.UUID, input *models.UpdatePermissionInput) (*models.PermissionResponse, error) {
	perm, err := s.repo.FindPermissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("permission")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch permission", err)
	}

	if input.Description != nil {
		perm.Description = *input.Description
	}

	if err := s.repo.UpdatePermission(ctx, perm); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to update permission", err)
	}
	return perm.ToResponse(), nil
}

func (s *rbacService) DeletePermission(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.FindPermissionByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("permission")
		}
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch permission", err)
	}
	if err := s.repo.DeletePermission(ctx, id); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to delete permission", err)
	}
	return nil
}

func (s *rbacService) ListPermissions(ctx context.Context, filter models.PermissionFilterInput) ([]*models.PermissionResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	}

	perms, total, err := s.repo.ListPermissions(ctx, filter)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list permissions", err)
	}

	resp := make([]*models.PermissionResponse, 0, len(perms))
	for _, p := range perms {
		resp = append(resp, p.ToResponse())
	}
	return resp, total, nil
}

// ── Role <-> Permission ───────────────────────────────────────────────────────

func (s *rbacService) SetRolePermissions(ctx context.Context, roleID uuid.UUID, input *models.AssignPermissionsInput) (*models.RoleResponse, error) {
	// Vérifie que le rôle existe
	if _, err := s.repo.FindRoleByID(ctx, roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch role", err)
	}

	// Vérifie que toutes les permissions existent
	if len(input.PermissionIDs) > 0 {
		perms, err := s.repo.FindPermissionsByIDs(ctx, input.PermissionIDs)
		if err != nil {
			return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to verify permissions", err)
		}
		if len(perms) != len(input.PermissionIDs) {
			return nil, appErrors.BadRequest("one or more permission IDs are invalid")
		}
	}

	if err := s.repo.SetRolePermissions(ctx, roleID, input.PermissionIDs); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to set role permissions", err)
	}

	// Recharger avec permissions mises à jour
	return s.GetRole(ctx, roleID)
}

func (s *rbacService) AddRolePermission(ctx context.Context, roleID uuid.UUID, input *models.AddPermissionInput) (*models.RoleResponse, error) {
	if _, err := s.repo.FindRoleByID(ctx, roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch role", err)
	}
	if _, err := s.repo.FindPermissionByID(ctx, input.PermissionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("permission")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch permission", err)
	}

	if err := s.repo.AddRolePermission(ctx, roleID, input.PermissionID); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to add permission to role", err)
	}
	return s.GetRole(ctx, roleID)
}

func (s *rbacService) RemoveRolePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	if err := s.repo.RemoveRolePermission(ctx, roleID, permID); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to remove permission from role", err)
	}
	return nil
}

func (s *rbacService) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.PermissionResponse, error) {
	perms, err := s.repo.GetRolePermissions(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.NotFound("role")
		}
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to get role permissions", err)
	}
	resp := make([]*models.PermissionResponse, 0, len(perms))
	for _, p := range perms {
		resp = append(resp, p.ToResponse())
	}
	return resp, nil
}

// ── User <-> Role ─────────────────────────────────────────────────────────────

func (s *rbacService) AssignRoleToUser(ctx context.Context, userID uuid.UUID, input *models.AssignRoleInput, assignedBy uuid.UUID) error {
	// Vérifie que le rôle existe
	if _, err := s.repo.FindRoleByID(ctx, input.RoleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.NotFound("role")
		}
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to fetch role", err)
	}

	if err := s.repo.AssignRoleToUser(ctx, userID, input.RoleID, assignedBy); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to assign role to user", err)
	}

	logger.FromContext(ctx).Info("role assigned to user",
		zap.String("userId", userID.String()),
		zap.String("roleId", input.RoleID.String()),
		zap.String("assignedBy", assignedBy.String()),
	)
	return nil
}

func (s *rbacService) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.repo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to remove role from user", err)
	}
	return nil
}

func (s *rbacService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.UserRoleResponse, error) {
	userRoles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to get user roles", err)
	}
	resp := make([]*models.UserRoleResponse, 0, len(userRoles))
	for _, ur := range userRoles {
		resp = append(resp, ur.ToResponse())
	}
	return resp, nil
}

// ── Checks ────────────────────────────────────────────────────────────────────

func (s *rbacService) UserHasRole(ctx context.Context, userID uuid.UUID, roleName models.RoleName) (bool, error) {
	return s.repo.UserHasRole(ctx, userID, roleName)
}

func (s *rbacService) UserHasPermission(ctx context.Context, userID uuid.UUID, resource string, action models.PermissionAction) (bool, error) {
	userRoles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to get user roles", err)
	}
	for _, ur := range userRoles {
		if ur.Role.HasPermission(resource, action) {
			return true, nil
		}
	}
	return false, nil
}
