package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ops-server/internal/domain/config_general/models"
)

//go:generate mockgen -source=config_repository.go -destination=mocks/config_repository_mock.go -package=mocks
type ConfigGeneralRepository interface {
	Create(ctx context.Context, configGeneral *models.ConfigGeneral) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.ConfigGeneral, error)
	FindByEntityKey(ctx context.Context, entity, key string) (*models.ConfigGeneral, error)
	Update(ctx context.Context, configGeneral *models.ConfigGeneral) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*models.ConfigGeneral, int64, error)
}

type configRepository struct {
	db *gorm.DB
}

func NewConfigGeneralRepository(db *gorm.DB) ConfigGeneralRepository {
	return &configRepository{db: db}
}

func (r *configRepository) Create(ctx context.Context, configGeneral *models.ConfigGeneral) error {
	return r.db.WithContext(ctx).Create(configGeneral).Error
}

func (r *configRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.ConfigGeneral, error) {
	var configGeneral models.ConfigGeneral
	if err := r.db.WithContext(ctx).First(&configGeneral, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &configGeneral, nil
}

func (r *configRepository) FindByEntityKey(ctx context.Context, entity, key string) (*models.ConfigGeneral, error) {
	var configGeneral models.ConfigGeneral
	err := r.db.WithContext(ctx).
		Where("entity = ? AND key = ? AND is_active = true", entity, key).
		First(&configGeneral).Error
	if err != nil {
		return nil, err
	}
	return &configGeneral, nil
}

func (r *configRepository) Update(ctx context.Context, configGeneral *models.ConfigGeneral) error {
	return r.db.WithContext(ctx).Save(configGeneral).Error
}

func (r *configRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.ConfigGeneral{}, id).Error
}

func (r *configRepository) List(ctx context.Context, offset, limit int) ([]*models.ConfigGeneral, int64, error) {
	var configs []*models.ConfigGeneral
	var total int64

	query := r.db.WithContext(ctx).Model(&models.ConfigGeneral{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(limit).Find(&configs).Error; err != nil {
		return nil, 0, err
	}

	return configs, total, nil
}
