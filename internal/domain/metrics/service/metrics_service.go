package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	metricModels "ops-server/internal/domain/metrics/models"
	"ops-server/internal/domain/metrics/repository"
	appErrors "ops-server/pkg/errors"
)

// MetricsService définit les opérations sur métriques et événements.
type MetricsService interface {
	RecordMetric(ctx context.Context, m *metricModels.Metric) (*metricModels.MetricResponse, error)
	ListMetrics(ctx context.Context, filter metricModels.MetricFilterInput, from, to time.Time) ([]*metricModels.MetricResponse, int64, error)
	RecordEvent(ctx context.Context, e *metricModels.Event) (*metricModels.EventResponse, error)
	ListEvents(ctx context.Context, filter metricModels.EventFilterInput, from, to time.Time) ([]*metricModels.EventResponse, int64, error)
	GetEvent(ctx context.Context, id uuid.UUID) (*metricModels.EventResponse, error)
}

type metricsService struct{ repo repository.MetricsRepository }

func NewMetricsService(repo repository.MetricsRepository) MetricsService {
	return &metricsService{repo: repo}
}

func (s *metricsService) RecordMetric(ctx context.Context, m *metricModels.Metric) (*metricModels.MetricResponse, error) {
	if err := s.repo.CreateMetric(ctx, m); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to record metric", err)
	}
	return m.ToResponse(), nil
}

func (s *metricsService) ListMetrics(ctx context.Context, filter metricModels.MetricFilterInput, from, to time.Time) ([]*metricModels.MetricResponse, int64, error) {
	items, total, err := s.repo.ListMetrics(ctx, filter, from, to)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list metrics", err)
	}
	resp := make([]*metricModels.MetricResponse, 0, len(items))
	for _, m := range items {
		resp = append(resp, m.ToResponse())
	}
	return resp, total, nil
}

func (s *metricsService) RecordEvent(ctx context.Context, e *metricModels.Event) (*metricModels.EventResponse, error) {
	if err := s.repo.CreateEvent(ctx, e); err != nil {
		return nil, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to record event", err)
	}
	return e.ToResponse(), nil
}

func (s *metricsService) ListEvents(ctx context.Context, filter metricModels.EventFilterInput, from, to time.Time) ([]*metricModels.EventResponse, int64, error) {
	items, total, err := s.repo.ListEvents(ctx, filter, from, to)
	if err != nil {
		return nil, 0, appErrors.Wrap(appErrors.ErrCodeDBQuery, "failed to list events", err)
	}
	resp := make([]*metricModels.EventResponse, 0, len(items))
	for _, e := range items {
		resp = append(resp, e.ToResponse())
	}
	return resp, total, nil
}

func (s *metricsService) GetEvent(ctx context.Context, id uuid.UUID) (*metricModels.EventResponse, error) {
	e, err := s.repo.FindEventByID(ctx, id)
	if err != nil {
		return nil, appErrors.New(appErrors.ErrCodeNotFound, "event not found")
	}
	return e.ToResponse(), nil
}
