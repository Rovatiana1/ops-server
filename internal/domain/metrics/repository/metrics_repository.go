package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	metricModels "ops-server/internal/domain/metrics/models"
)

//go:generate mockgen -source=metrics_repository.go -destination=mocks/metrics_repository_mock.go
type MetricsRepository interface {
	CreateMetric(ctx context.Context, m *metricModels.Metric) error
	ListMetrics(ctx context.Context, filter metricModels.MetricFilterInput, from, to time.Time) ([]*metricModels.Metric, int64, error)
	CreateEvent(ctx context.Context, e *metricModels.Event) error
	ListEvents(ctx context.Context, filter metricModels.EventFilterInput, from, to time.Time) ([]*metricModels.Event, int64, error)
	FindEventByID(ctx context.Context, id uuid.UUID) (*metricModels.Event, error)
}

type metricsRepository struct{ db *gorm.DB }

func NewMetricsRepository(db *gorm.DB) MetricsRepository {
	return &metricsRepository{db: db}
}

func (r *metricsRepository) CreateMetric(ctx context.Context, m *metricModels.Metric) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *metricsRepository) ListMetrics(ctx context.Context, filter metricModels.MetricFilterInput, from, to time.Time) ([]*metricModels.Metric, int64, error) {
	var items []*metricModels.Metric
	var total int64

	q := r.db.WithContext(ctx).Model(&metricModels.Metric{})
	if filter.Name != "" {
		q = q.Where("name ILIKE ?", "%"+filter.Name+"%")
	}
	if filter.Type != "" {
		q = q.Where("type = ?", filter.Type)
	}
	if filter.Period != "" {
		q = q.Where("period = ?", filter.Period)
	}
	if !from.IsZero() {
		q = q.Where("recorded_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("recorded_at <= ?", to)
	}

	q.Count(&total)
	err := q.Order("recorded_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error
	return items, total, err
}

func (r *metricsRepository) CreateEvent(ctx context.Context, e *metricModels.Event) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *metricsRepository) ListEvents(ctx context.Context, filter metricModels.EventFilterInput, from, to time.Time) ([]*metricModels.Event, int64, error) {
	var items []*metricModels.Event
	var total int64

	q := r.db.WithContext(ctx).Model(&metricModels.Event{})
	if filter.Severity != "" {
		q = q.Where("severity = ?", filter.Severity)
	}
	if filter.Category != "" {
		q = q.Where("category = ?", filter.Category)
	}
	if filter.Source != "" {
		q = q.Where("source = ?", filter.Source)
	}
	if filter.UserID != "" {
		q = q.Where("user_id = ?", filter.UserID)
	}
	if !from.IsZero() {
		q = q.Where("occurred_at >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("occurred_at <= ?", to)
	}

	q.Count(&total)
	err := q.Order("occurred_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error
	return items, total, err
}

func (r *metricsRepository) FindEventByID(ctx context.Context, id uuid.UUID) (*metricModels.Event, error) {
	var e metricModels.Event
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error
	return &e, err
}
