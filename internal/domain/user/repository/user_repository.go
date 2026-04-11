package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ops-server/internal/domain/user/models"
)

//go:generate mockgen -source=user_repository.go -destination=mocks/user_repository_mock.go -package=mocks
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByIdentifier(ctx context.Context, identifier string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*models.User, int64, error)
	ExistsByIdentifier(ctx context.Context, identifier string) (bool, error)
	AssignRole(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetRoles(ctx context.Context, userID uuid.UUID) ([]models.UserRole, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles.Role.Permissions").
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByIdentifier(ctx context.Context, identifier string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles.Role.Permissions").
		Where("identifier = ?", identifier).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.User{}).Error
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Preload("UserRoles.Role").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) ExistsByIdentifier(ctx context.Context, identifier string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("identifier = ?", identifier).
		Count(&count).Error
	return count > 0, err
}

func (r *userRepository) AssignRole(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	ur := models.UserRole{
		UserID:     userID,
		RoleID:     roleID,
		AssignedBy: assignedBy,
	}
	return r.db.WithContext(ctx).
		Where(models.UserRole{UserID: userID, RoleID: roleID}).
		FirstOrCreate(&ur).Error
}

func (r *userRepository) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&models.UserRole{}).Error
}

func (r *userRepository) GetRoles(ctx context.Context, userID uuid.UUID) ([]models.UserRole, error) {
	var roles []models.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}
