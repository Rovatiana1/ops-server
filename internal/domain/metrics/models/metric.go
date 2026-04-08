package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Énumérations ─────────────────────────────────────────────────────────────

// MetricType catégorise le type de mesure Prometheus-compatible.
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"   // valeur toujours croissante
	MetricTypeGauge     MetricType = "gauge"     // valeur instantanée (peut baisser)
	MetricTypeHistogram MetricType = "histogram" // distribution de valeurs
)

func (t MetricType) IsValid() bool {
	switch t {
	case MetricTypeCounter, MetricTypeGauge, MetricTypeHistogram:
		return true
	}
	return false
}

// MetricPeriod représente la granularité temporelle d'agrégation.
type MetricPeriod string

const (
	MetricPeriodRealtime MetricPeriod = "realtime"
	MetricPeriodHourly   MetricPeriod = "hourly"
	MetricPeriodDaily    MetricPeriod = "daily"
	MetricPeriodWeekly   MetricPeriod = "weekly"
	MetricPeriodMonthly  MetricPeriod = "monthly"
)

// ─── Modèle GORM ──────────────────────────────────────────────────────────────

// Metric est une mesure agrégée persistée en base.
// Les labels JSONB permettent de stocker des dimensions arbitraires
// (ex: {"env":"prod","region":"eu-west","service":"api"}).
//
// Table: metrics
// Index: name, recorded_at, type
type Metric struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"                     json:"metricId"`
	Name       string         `gorm:"type:varchar(200);not null;index"         json:"name"`
	Type       MetricType     `gorm:"type:varchar(20);not null;index"          json:"type"`
	Value      float64        `gorm:"not null"                                 json:"value"`
	Labels     []byte         `gorm:"type:jsonb"                               json:"labels,omitempty"`
	Period     MetricPeriod   `gorm:"type:varchar(20)"                         json:"period,omitempty"`
	Source     string         `gorm:"type:varchar(100)"                        json:"source,omitempty"` // ex: "api", "worker"
	RecordedAt time.Time      `gorm:"not null;index:idx_metric_time"           json:"recordedAt"`
	CreatedAt  time.Time      `                                                json:"createdAt"`
	UpdatedAt  time.Time      `                                                json:"updatedAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                                    json:"-"`
}

func (Metric) TableName() string { return "metrics" }

func (m *Metric) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.RecordedAt.IsZero() {
		m.RecordedAt = time.Now().UTC()
	}
	return nil
}

// ParsedLabels désérialise les labels JSONB en map.
func (m *Metric) ParsedLabels() map[string]string {
	if m.Labels == nil {
		return nil
	}
	var out map[string]string
	_ = json.Unmarshal(m.Labels, &out)
	return out
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// CreateMetricInput est le payload de création d'une métrique.
//
// @Description Données pour enregistrer une métrique
type CreateMetricInput struct {
	Name       string         `json:"name"       binding:"required,max=200"`
	Type       MetricType     `json:"type"       binding:"required"`
	Value      float64        `json:"value"      binding:"required"`
	Labels     map[string]string `json:"labels"`
	Period     MetricPeriod   `json:"period"`
	Source     string         `json:"source"     binding:"max=100"`
	RecordedAt *time.Time     `json:"recordedAt"` // optionnel, défaut = now
}

// MetricFilterInput regroupe les critères de filtre pour la liste.
type MetricFilterInput struct {
	Name   string       `form:"name"`
	Type   MetricType   `form:"type"`
	Period MetricPeriod `form:"period"`
	From   string       `form:"from"` // RFC3339
	To     string       `form:"to"`   // RFC3339
	Offset int          `form:"offset"`
	Limit  int          `form:"limit,default=20"`
}

// MetricResponse est la vue publique (lowerCamelCase).
//
// @Description Représentation publique d'une métrique
type MetricResponse struct {
	ID         uuid.UUID         `json:"metricId"`
	Name       string            `json:"name"`
	Type       MetricType        `json:"type"`
	Value      float64           `json:"value"`
	Labels     map[string]string `json:"labels,omitempty"`
	Period     MetricPeriod      `json:"period,omitempty"`
	Source     string            `json:"source,omitempty"`
	RecordedAt time.Time         `json:"recordedAt"`
	CreatedAt  time.Time         `json:"createdAt"`
}

// ToResponse convertit le modèle GORM en vue publique.
func (m *Metric) ToResponse() *MetricResponse {
	return &MetricResponse{
		ID:         m.ID,
		Name:       m.Name,
		Type:       m.Type,
		Value:      m.Value,
		Labels:     m.ParsedLabels(),
		Period:     m.Period,
		Source:     m.Source,
		RecordedAt: m.RecordedAt,
		CreatedAt:  m.CreatedAt,
	}
}
