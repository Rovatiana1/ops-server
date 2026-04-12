package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ops-server/internal/domain/config_general/models"
	"ops-server/internal/domain/config_general/repository"
	appErrors "ops-server/pkg/errors"
)

type ConfigGeneralService interface {
	Create(ctx context.Context, input *models.CreateConfigGeneralInput) (*models.ConfigGeneralResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.ConfigGeneralResponse, error)
	GetByEntityKey(ctx context.Context, entity, key string) (*models.ConfigGeneralResponse, error)
	Update(ctx context.Context, id uuid.UUID, input *models.UpdateConfigGeneralInput) (*models.ConfigGeneralResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*models.ConfigGeneralResponse, int64, error)
}

type configService struct {
	repo repository.ConfigGeneralRepository
}

func NewConfigGeneralService(repo repository.ConfigGeneralRepository) ConfigGeneralService {
	return &configService{repo: repo}
}

func (s *configService) Create(ctx context.Context, input *models.CreateConfigGeneralInput) (*models.ConfigGeneralResponse, error) {
	existing, err := s.repo.FindByEntityKey(ctx, input.Entity, input.Key)
	if err == nil && existing != nil {
		return nil, appErrors.New(appErrors.ErrCodeConflict, "configGeneral already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	configGeneral := &models.ConfigGeneral{
		Entity:   input.Entity,
		Key:      input.Key,
		Data:     input.Data,
		Version:  1,
		IsActive: true,
	}

	if err := s.repo.Create(ctx, configGeneral); err != nil {
		return nil, err
	}

	return configGeneral.ToResponse(), nil
}

func (s *configService) GetByID(ctx context.Context, id uuid.UUID) (*models.ConfigGeneralResponse, error) {
	configGeneral, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return configGeneral.ToResponse(), nil
}

func (s *configService) GetByEntityKey(ctx context.Context, entity, key string) (*models.ConfigGeneralResponse, error) {
	configGeneral, err := s.repo.FindByEntityKey(ctx, entity, key)
	if err != nil {
		return nil, err
	}
	return configGeneral.ToResponse(), nil
}

func (s *configService) Update(ctx context.Context, id uuid.UUID, input *models.UpdateConfigGeneralInput) (*models.ConfigGeneralResponse, error) {
	configGeneral, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Data != nil {
		configGeneral.Data = *input.Data
		configGeneral.Version++
	}

	if input.IsActive != nil {
		configGeneral.IsActive = *input.IsActive
	}

	if err := s.repo.Update(ctx, configGeneral); err != nil {
		return nil, err
	}

	return configGeneral.ToResponse(), nil
}

func (s *configService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *configService) List(ctx context.Context, offset, limit int) ([]*models.ConfigGeneralResponse, int64, error) {
	configs, total, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*models.ConfigGeneralResponse, 0, len(configs))
	for _, c := range configs {
		resp = append(resp, c.ToResponse())
	}

	return resp, total, nil
}
