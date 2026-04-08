package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Énumérations ─────────────────────────────────────────────────────────────

// EventSeverity représente la criticité d'un événement observable.
type EventSeverity string

const (
	EventSeverityInfo     EventSeverity = "info"
	EventSeverityWarning  EventSeverity = "warning"
	EventSeverityCritical EventSeverity = "critical"
)

func (s EventSeverity) IsValid() bool {
	switch s {
	case EventSeverityInfo, EventSeverityWarning, EventSeverityCritical:
		return true
	}
	return false
}

// EventCategory regroupe les événements par domaine fonctionnel.
type EventCategory string

const (
	EventCategoryAuth     EventCategory = "auth"
	EventCategoryUser     EventCategory = "user"
	EventCategorySystem   EventCategory = "system"
	EventCategoryIngestion EventCategory = "ingestion"
	EventCategoryAPI      EventCategory = "api"
)

// ─── Modèle GORM ──────────────────────────────────────────────────────────────

// Event est un événement métier observable persisté en base.
// Utilisé pour le monitoring, les dashboards et les alertes.
//
// Table: events
// Index: severity, category, occurred_at, user_id
type Event struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"                      json:"eventId"`
	Name        string         `gorm:"type:varchar(200);not null;index"          json:"name"`
	Severity    EventSeverity  `gorm:"type:varchar(20);not null;index"           json:"severity"`
	Category    EventCategory  `gorm:"type:varchar(50);index"                    json:"category,omitempty"`
	Source      string         `gorm:"type:varchar(100)"                         json:"source,omitempty"`
	UserID      *uuid.UUID     `gorm:"type:uuid;index"                           json:"userId,omitempty"`
	RequestID   string         `gorm:"type:varchar(100)"                         json:"requestId,omitempty"`
	Description string         `gorm:"type:text"                                 json:"description,omitempty"`
	Payload     []byte         `gorm:"type:jsonb"                                json:"payload,omitempty"`
	OccurredAt  time.Time      `gorm:"not null;index:idx_event_time"             json:"occurredAt"`
	CreatedAt   time.Time      `                                                 json:"createdAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                                     json:"-"`
}

func (Event) TableName() string { return "events" }

func (e *Event) BeforeCreate(_ *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.OccurredAt.IsZero() {
		e.OccurredAt = time.Now().UTC()
	}
	return nil
}

// ParsedPayload désérialise le payload JSONB en map générique.
func (e *Event) ParsedPayload() map[string]any {
	if e.Payload == nil {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal(e.Payload, &out)
	return out
}

// IsCritical retourne true si l'événement est critique.
func (e *Event) IsCritical() bool {
	return e.Severity == EventSeverityCritical
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// CreateEventInput est le payload de création d'un événement.
//
// @Description Données pour enregistrer un événement observable
type CreateEventInput struct {
	Name        string            `json:"name"        binding:"required,max=200"`
	Severity    EventSeverity     `json:"severity"    binding:"required"`
	Category    EventCategory     `json:"category"`
	Source      string            `json:"source"      binding:"max=100"`
	UserID      *uuid.UUID        `json:"userId"`
	RequestID   string            `json:"requestId"   binding:"max=100"`
	Description string            `json:"description"`
	Payload     map[string]any    `json:"payload"`
	OccurredAt  *time.Time        `json:"occurredAt"` // optionnel, défaut = now
}

// EventFilterInput regroupe les critères de filtre pour la liste.
type EventFilterInput struct {
	Severity EventSeverity `form:"severity"`
	Category EventCategory `form:"category"`
	Source   string        `form:"source"`
	UserID   string        `form:"userId"` // UUID string
	From     string        `form:"from"`   // RFC3339
	To       string        `form:"to"`     // RFC3339
	Offset   int           `form:"offset"`
	Limit    int           `form:"limit,default=20"`
}

// EventResponse est la vue publique (lowerCamelCase).
//
// @Description Représentation publique d'un événement observable
type EventResponse struct {
	ID          uuid.UUID         `json:"eventId"`
	Name        string            `json:"name"`
	Severity    EventSeverity     `json:"severity"`
	Category    EventCategory     `json:"category,omitempty"`
	Source      string            `json:"source,omitempty"`
	UserID      *uuid.UUID        `json:"userId,omitempty"`
	RequestID   string            `json:"requestId,omitempty"`
	Description string            `json:"description,omitempty"`
	Payload     map[string]any    `json:"payload,omitempty"`
	OccurredAt  time.Time         `json:"occurredAt"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// ToResponse convertit le modèle GORM en vue publique.
func (e *Event) ToResponse() *EventResponse {
	return &EventResponse{
		ID:          e.ID,
		Name:        e.Name,
		Severity:    e.Severity,
		Category:    e.Category,
		Source:      e.Source,
		UserID:      e.UserID,
		RequestID:   e.RequestID,
		Description: e.Description,
		Payload:     e.ParsedPayload(),
		OccurredAt:  e.OccurredAt,
		CreatedAt:   e.CreatedAt,
	}
}
