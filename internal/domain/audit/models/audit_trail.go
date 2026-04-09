package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Énumérations ─────────────────────────────────────────────────────────────

// AuditAction représente l'opération effectuée sur une ressource.
// @Enum CREATE, UPDATE, DELETE, READ, LOGIN, LOGOUT, EXPORT
type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
	AuditActionRead   AuditAction = "READ"
	AuditActionLogin  AuditAction = "LOGIN"
	AuditActionLogout AuditAction = "LOGOUT"
	AuditActionExport AuditAction = "EXPORT"
)

func (a AuditAction) IsValid() bool {
	switch a {
	case AuditActionCreate, AuditActionUpdate, AuditActionDelete,
		AuditActionRead, AuditActionLogin, AuditActionLogout, AuditActionExport:
		return true
	}
	return false
}

// AuditOutcome représente le résultat de l'opération.
// @Enum success,failure,denied
type AuditOutcome string

const (
	AuditOutcomeSuccess AuditOutcome = "success"
	AuditOutcomeFailure AuditOutcome = "failure"
	AuditOutcomeDenied  AuditOutcome = "denied"
)

// ─── Modèle GORM ──────────────────────────────────────────────────────────────

// AuditTrail enregistre chaque action sensible effectuée dans le système.
// Immuable par conception : pas de UpdatedAt, pas de soft-delete réel
// (DeletedAt présent uniquement pour GORM, jamais utilisé fonctionnellement).
//
// Table: audit_trails
// Index: action, resource, resource_id, user_id, created_at, outcome
type AuditTrail struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"                     json:"auditId"`
	UserID     *uuid.UUID     `gorm:"type:uuid;index"                          json:"userId,omitempty"`
	Action     AuditAction    `gorm:"type:varchar(20);not null;index"          json:"action"`
	Outcome    AuditOutcome   `gorm:"type:varchar(20);not null;default:'success'" json:"outcome"`
	Resource   string         `gorm:"type:varchar(100);not null;index"         json:"resource"`   // ex: "user", "notification"
	ResourceID string         `gorm:"type:varchar(100);index"                  json:"resourceId"` // UUID de la ressource
	OldValues  []byte         `gorm:"type:jsonb"                               json:"oldValues,omitempty"`
	NewValues  []byte         `gorm:"type:jsonb"                               json:"newValues,omitempty"`
	IPAddress  string         `gorm:"type:varchar(45)"                         json:"ipAddress,omitempty"`
	UserAgent  string         `gorm:"type:text"                                json:"userAgent,omitempty"`
	RequestID  string         `gorm:"type:varchar(100);index"                  json:"requestId,omitempty"`
	StatusCode int            `gorm:"not null;default:200"                     json:"statusCode"`
	Error      string         `gorm:"type:text"                                json:"error,omitempty"` // message d'erreur si failure
	CreatedAt  time.Time      `gorm:"not null;index:idx_audit_time"            json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                                    json:"-"`
}

func (AuditTrail) TableName() string { return "audit_trails" }

func (a *AuditTrail) BeforeCreate(_ *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.Outcome == "" {
		a.Outcome = AuditOutcomeSuccess
	}
	return nil
}

// ParsedOldValues désérialise OldValues en map.
func (a *AuditTrail) ParsedOldValues() map[string]any {
	if a.OldValues == nil {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal(a.OldValues, &out)
	return out
}

// ParsedNewValues désérialise NewValues en map.
func (a *AuditTrail) ParsedNewValues() map[string]any {
	if a.NewValues == nil {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal(a.NewValues, &out)
	return out
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// CreateAuditInput est utilisé par le service pour enregistrer une action.
// Jamais exposé en HTTP — appelé en interne par les services métier.
type CreateAuditInput struct {
	UserID     *uuid.UUID
	Action     AuditAction
	Outcome    AuditOutcome
	Resource   string
	ResourceID string
	OldValues  any // sera sérialisé en JSON
	NewValues  any // sera sérialisé en JSON
	IPAddress  string
	UserAgent  string
	RequestID  string
	StatusCode int
	Error      string
}

// AuditFilterInput regroupe les critères de filtre pour la liste.
type AuditFilterInput struct {
	Resource string       `form:"resource"`
	Action   AuditAction  `form:"action"`
	Outcome  AuditOutcome `form:"outcome"`
	UserID   string       `form:"userId"` // UUID string
	From     string       `form:"from"`   // RFC3339
	To       string       `form:"to"`     // RFC3339
	Offset   int          `form:"offset"`
	Limit    int          `form:"limit,default=20"`
}

// AuditTrailResponse est la vue publique (lowerCamelCase).
//
// @Description Représentation publique d'un audit trail
type AuditTrailResponse struct {
	ID         uuid.UUID      `json:"auditId"`
	UserID     *uuid.UUID     `json:"userId,omitempty"`
	Action     AuditAction    `json:"action"`
	Outcome    AuditOutcome   `json:"outcome"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resourceId"`
	OldValues  map[string]any `json:"oldValues,omitempty"`
	NewValues  map[string]any `json:"newValues,omitempty"`
	IPAddress  string         `json:"ipAddress,omitempty"`
	RequestID  string         `json:"requestId,omitempty"`
	StatusCode int            `json:"statusCode"`
	Error      string         `json:"error,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// ToResponse convertit le modèle GORM en vue publique.
func (a *AuditTrail) ToResponse() *AuditTrailResponse {
	return &AuditTrailResponse{
		ID:         a.ID,
		UserID:     a.UserID,
		Action:     a.Action,
		Outcome:    a.Outcome,
		Resource:   a.Resource,
		ResourceID: a.ResourceID,
		OldValues:  a.ParsedOldValues(),
		NewValues:  a.ParsedNewValues(),
		IPAddress:  a.IPAddress,
		RequestID:  a.RequestID,
		StatusCode: a.StatusCode,
		Error:      a.Error,
		CreatedAt:  a.CreatedAt,
	}
}

// NewAuditTrailFromInput construit un AuditTrail depuis un CreateAuditInput,
// en sérialisant OldValues et NewValues en JSONB.
func NewAuditTrailFromInput(input *CreateAuditInput) (*AuditTrail, error) {
	a := &AuditTrail{
		UserID:     input.UserID,
		Action:     input.Action,
		Outcome:    input.Outcome,
		Resource:   input.Resource,
		ResourceID: input.ResourceID,
		IPAddress:  input.IPAddress,
		UserAgent:  input.UserAgent,
		RequestID:  input.RequestID,
		StatusCode: input.StatusCode,
		Error:      input.Error,
	}
	if a.Outcome == "" {
		a.Outcome = AuditOutcomeSuccess
	}
	if input.OldValues != nil {
		b, err := json.Marshal(input.OldValues)
		if err != nil {
			return nil, err
		}
		a.OldValues = b
	}
	if input.NewValues != nil {
		b, err := json.Marshal(input.NewValues)
		if err != nil {
			return nil, err
		}
		a.NewValues = b
	}
	return a, nil
}
