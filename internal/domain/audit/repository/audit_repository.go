package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"ops-server/internal/domain/audit/models"
)

//go:generate mockgen -source=audit_repository.go -destination=mocks/audit_repository_mock.go
type AuditRepository interface {
	CreateTrail(ctx context.Context, a *models.AuditTrail) error
	ListTrails(ctx context.Context, filter models.AuditFilterInput, userID *uuid.UUID, from, to time.Time) ([]*models.AuditTrail, int64, error)
	FindTrailByID(ctx context.Context, id uuid.UUID) (*models.AuditTrail, error)
	CreateLog(ctx context.Context, l *models.Log) error
	ListLogs(ctx context.Context, filter models.LogFilterInput, from, to time.Time) ([]*models.Log, int64, error)
}

type auditRepository struct{ db *gorm.DB }

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) CreateTrail(ctx context.Context, a *models.AuditTrail) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *auditRepository) ListTrails(ctx context.Context, filter models.AuditFilterInput, userID *uuid.UUID, from, to time.Time) ([]*models.AuditTrail, int64, error) {
	var items []*models.AuditTrail
	var total int64

	q := r.db.WithContext(ctx).Model(&models.AuditTrail{})
	if filter.Resource != "" {
		q = q.Where("resource = ?", filter.Resource)
	}
	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.Outcome != "" {
		q = q.Where("outcome = ?", filter.Outcome)
	}
	if userID != nil {
		q = q.Where("user_id = ?", userID)
	}
	if !from.IsZero() {
		q = q.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("created_at <= ?", to)
	}

	q.Count(&total)
	err := q.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error
	return items, total, err
}

func (r *auditRepository) FindTrailByID(ctx context.Context, id uuid.UUID) (*models.AuditTrail, error) {
	var a models.AuditTrail
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error
	return &a, err
}

func (r *auditRepository) CreateLog(ctx context.Context, l *models.Log) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *auditRepository) ListLogs(ctx context.Context, filter models.LogFilterInput, from, to time.Time) ([]*models.Log, int64, error) {
	var items []*models.Log
	var total int64

	q := r.db.WithContext(ctx).Model(&models.Log{})
	if filter.Level != "" {
		q = q.Where("level = ?", filter.Level)
	}
	if filter.Service != "" {
		q = q.Where("service = ?", filter.Service)
	}
	if filter.TraceID != "" {
		q = q.Where("trace_id = ?", filter.TraceID)
	}
	if filter.RequestID != "" {
		q = q.Where("request_id = ?", filter.RequestID)
	}
	if !from.IsZero() {
		q = q.Where("created_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("created_at <= ?", to)
	}

	q.Count(&total)
	err := q.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error
	return items, total, err
}
