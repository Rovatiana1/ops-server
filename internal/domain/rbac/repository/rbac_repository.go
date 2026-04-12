package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ops-server/internal/domain/rbac/models"
	"ops-server/internal/interfaces/http/response"
)

// RBACRepository définit le contrat de persistance pour roles et permissions.
//
//go:generate mockgen -source=rbac_repository.go -destination=mocks/rbac_repository_mock.go -package=mocks
type RBACRepository interface {
	// ── Roles ─────────────────────────────────────────────────────────────────
	CreateRole(ctx context.Context, role *models.Role) error
	FindRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	FindRoleByName(ctx context.Context, name models.RoleName) (*models.Role, error)
	UpdateRole(ctx context.Context, role *models.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	ListRoles(ctx context.Context, filter models.RoleFilterInput) ([]*models.Role, int64, error)
	ExistsRoleByName(ctx context.Context, name models.RoleName) (bool, error)

	// ── Permissions ───────────────────────────────────────────────────────────
	CreatePermission(ctx context.Context, perm *models.Permission) error
	FindPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error)
	FindPermissionBySlug(ctx context.Context, resource string, action models.PermissionAction) (*models.Permission, error)
	UpdatePermission(ctx context.Context, perm *models.Permission) error
	DeletePermission(ctx context.Context, id uuid.UUID) error
	ListPermissions(ctx context.Context, filter models.PermissionFilterInput) ([]*models.Permission, int64, error)
	FindPermissionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Permission, error)

	// ── Role <-> Permission ───────────────────────────────────────────────────
	SetRolePermissions(ctx context.Context, roleID uuid.UUID, permIDs []uuid.UUID) error
	AddRolePermission(ctx context.Context, roleID, permID uuid.UUID) error
	RemoveRolePermission(ctx context.Context, roleID, permID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)

	// ── User <-> Role ─────────────────────────────────────────────────────────
	AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.UserRole, error)
	UserHasRole(ctx context.Context, userID uuid.UUID, roleName models.RoleName) (bool, error)
}

type rbacRepository struct {
	db *gorm.DB
}

// NewRBACRepository crée un RBACRepository GORM.
func NewRBACRepository(db *gorm.DB) RBACRepository {
	return &rbacRepository{db: db}
}

// ── Roles ─────────────────────────────────────────────────────────────────────

func (r *rbacRepository) CreateRole(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *rbacRepository) FindRoleByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("id = ?", id).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *rbacRepository) FindRoleByName(ctx context.Context, name models.RoleName) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("name = ?", name).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *rbacRepository) UpdateRole(ctx context.Context, role *models.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *rbacRepository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Role{}).Error
}

func (r *rbacRepository) ListRoles(ctx context.Context, filter models.RoleFilterInput) ([]*models.Role, int64, error) {
	var roles []*models.Role
	var total int64

	q := r.db.WithContext(ctx).Model(&models.Role{})

	if filter.Name != "" {
		q = q.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.IsSystem != nil {
		q = q.Where("is_system = ?", *filter.IsSystem)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := response.PageToOffset(filter.Page, filter.Limit)
	err := q.Preload("Permissions").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&roles).Error

	return roles, total, err
}

func (r *rbacRepository) ExistsRoleByName(ctx context.Context, name models.RoleName) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Role{}).
		Where("name = ?", name).
		Count(&count).Error
	return count > 0, err
}

// ── Permissions ───────────────────────────────────────────────────────────────

func (r *rbacRepository) CreatePermission(ctx context.Context, perm *models.Permission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *rbacRepository) FindPermissionByID(ctx context.Context, id uuid.UUID) (*models.Permission, error) {
	var perm models.Permission
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *rbacRepository) FindPermissionBySlug(ctx context.Context, resource string, action models.PermissionAction) (*models.Permission, error) {
	var perm models.Permission
	err := r.db.WithContext(ctx).
		Where("resource = ? AND action = ?", resource, action).
		First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func (r *rbacRepository) UpdatePermission(ctx context.Context, perm *models.Permission) error {
	return r.db.WithContext(ctx).Save(perm).Error
}

func (r *rbacRepository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.Permission{}).Error
}

func (r *rbacRepository) ListPermissions(ctx context.Context, filter models.PermissionFilterInput) ([]*models.Permission, int64, error) {
	var perms []*models.Permission
	var total int64

	q := r.db.WithContext(ctx).Model(&models.Permission{})

	if filter.Resource != "" {
		q = q.Where("resource ILIKE ?", "%"+filter.Resource+"%")
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := response.PageToOffset(filter.Page, filter.Limit)
	err := q.Order("resource ASC, action ASC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&perms).Error

	return perms, total, err
}

func (r *rbacRepository) FindPermissionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Permission, error) {
	var perms []*models.Permission
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&perms).Error
	return perms, err
}

// ── Role <-> Permission ───────────────────────────────────────────────────────

// SetRolePermissions remplace atomiquement TOUTES les permissions d'un rôle.
func (r *rbacRepository) SetRolePermissions(ctx context.Context, roleID uuid.UUID, permIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Charger le rôle
		var role models.Role
		if err := tx.First(&role, "id = ?", roleID).Error; err != nil {
			return err
		}

		// 2. Charger les nouvelles permissions
		var perms []*models.Permission
		if len(permIDs) > 0 {
			if err := tx.Where("id IN ?", permIDs).Find(&perms).Error; err != nil {
				return err
			}
		}

		// 3. Remplacement atomique via Association
		return tx.Model(&role).Association("Permissions").Replace(perms)
	})
}

func (r *rbacRepository) AddRolePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	rp := models.RolePermission{RoleID: roleID, PermissionID: permID}
	result := r.db.WithContext(ctx).
		Where(rp).
		FirstOrCreate(&rp)
	return result.Error
}

func (r *rbacRepository) RemoveRolePermission(ctx context.Context, roleID, permID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permID).
		Delete(&models.RolePermission{}).Error
}

func (r *rbacRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error) {
	var role models.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("id = ?", roleID).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	perms := make([]*models.Permission, len(role.Permissions))
	for i := range role.Permissions {
		perms[i] = &role.Permissions[i]
	}
	return perms, nil
}

// ── User <-> Role ─────────────────────────────────────────────────────────────

func (r *rbacRepository) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	ur := models.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
	}
	return r.db.WithContext(ctx).
		Where(models.UserRole{UserID: userID, RoleID: roleID}).
		FirstOrCreate(&ur).Error
}

func (r *rbacRepository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error
}

func (r *rbacRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*models.UserRole, error) {
	var roles []*models.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *rbacRepository) UserHasRole(ctx context.Context, userID uuid.UUID, roleName models.RoleName) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.name = ? AND roles.deleted_at IS NULL", userID, roleName).
		Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return count > 0, err
}
