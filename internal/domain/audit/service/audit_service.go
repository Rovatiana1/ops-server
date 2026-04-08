package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"ops-server/internal/domain/audit/models"
	"ops-server/internal/domain/audit/repository"
	appErrors "ops-server/pkg/errors"
)

// AuditService définit les opérations de traçabilité.
type AuditService interface {
	Record(ctx context.Context, input *models.CreateAuditInput) error
	ListTrails(ctx context.Context, filter models.AuditFilterInput, userID *uuid.UUID, from, to time.Time) ([]*models.AuditTrailResponse, int64, error)
	GetTrail(ctx context.Context, id uuid.UUID) (*models.AuditTrailResponse, error)
	WriteLog(ctx context.Context, input *models.CreateLogInput) error
	ListLogs(ctx context.Context, filter models.LogFilterInput, from, to time.Time) ([]*models.LogResponse, int64, error)
}

type auditService struct{ repo repository.AuditRepository }

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Record(ctx context.Context, input *models.CreateAuditInput) error {
	trail, err := models.NewAuditTrailFromInput(input)
	if err != nil {
		return appErrors.Internal(err)
	}
	if err := s.repo.CreateTrail(ctx, trail); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to record audit trail", err)
	}
	return nil
}

func (s *auditService) ListTrails(ctx context.Context, filter models.AuditFilterInput, userID *uuid.UUID, from, to time.Time) ([]*models.AuditTrailResponse, int64, error) {
	items, total, err := s.repo.ListTrails(ctx, filter, userID, from, to)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list audit trails", err)
	}
	resp := make([]*models.AuditTrailResponse, 0, len(items))
	for _, a := range items {
		resp = append(resp, a.ToResponse())
	}
	return resp, total, nil
}

func (s *auditService) GetTrail(ctx context.Context, id uuid.UUID) (*models.AuditTrailResponse, error) {
	a, err := s.repo.FindTrailByID(ctx, id)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrCodeNotFound, "audit trail not found")
	}
	return a.ToResponse(), nil
}

func (s *auditService) WriteLog(ctx context.Context, input *models.CreateLogInput) error {
	l, err := models.NewLogFromInput(input)
	if err != nil {
		return appErrors.Internal(err)
	}
	if err := s.repo.CreateLog(ctx, l); err != nil {
		return appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to write log", err)
	}
	return nil
}

func (s *auditService) ListLogs(ctx context.Context, filter models.LogFilterInput, from, to time.Time) ([]*models.LogResponse, int64, error) {
	items, total, err := s.repo.ListLogs(ctx, filter, from, to)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list logs", err)
	}
	resp := make([]*models.LogResponse, 0, len(items))
	for _, l := range items {
		resp = append(resp, l.ToResponse())
	}
	return resp, total, nil
}
