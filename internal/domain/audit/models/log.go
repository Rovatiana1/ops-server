package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ─── Énumérations ─────────────────────────────────────────────────────────────

// LogLevel représente la sévérité d'un log applicatif persisté.
// @Enum debug,info,warning,error,fatal
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarning, LogLevelError, LogLevelFatal:
		return true
	}
	return false
}

// LogLevelSeverity retourne un entier pour comparer les niveaux.
func (l LogLevel) Severity() int {
	switch l {
	case LogLevelDebug:
		return 0
	case LogLevelInfo:
		return 1
	case LogLevelWarning:
		return 2
	case LogLevelError:
		return 3
	case LogLevelFatal:
		return 4
	}
	return -1
}

// ─── Modèle GORM ──────────────────────────────────────────────────────────────

// Log est un enregistrement de log applicatif persisté en base.
// Complémentaire aux logs JSON stdout (zap), il permet :
// - la recherche et le filtrage via l'UI
// - la corrélation par traceId / requestId
// - l'alerting sur les niveaux error/fatal
//
// Table: logs
// Index: level, service, trace_id, request_id, created_at
type Log struct {
	ID         uuid.UUID      `gorm:"type:uuid;primaryKey"             json:"logId"`
	Level      LogLevel       `gorm:"type:varchar(10);not null;index"  json:"level"`
	Message    string         `gorm:"type:text;not null"               json:"message"`
	Service    string         `gorm:"type:varchar(100);index"          json:"service,omitempty"`
	TraceID    string         `gorm:"type:varchar(100);index"          json:"traceId,omitempty"`
	RequestID  string         `gorm:"type:varchar(100);index"          json:"requestId,omitempty"`
	UserID     *uuid.UUID     `gorm:"type:uuid;index"                  json:"userId,omitempty"`
	Fields     []byte         `gorm:"type:jsonb"                       json:"fields,omitempty"`
	CallerFile string         `gorm:"type:varchar(200)"                json:"callerFile,omitempty"`
	CallerLine int            `gorm:"default:0"                        json:"callerLine,omitempty"`
	CreatedAt  time.Time      `gorm:"not null;index:idx_log_time"      json:"createdAt"`
	DeletedAt  gorm.DeletedAt `gorm:"index"                            json:"-"`
}

func (Log) TableName() string { return "logs" }

func (l *Log) BeforeCreate(_ *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// ParsedFields désérialise les fields JSONB en map générique.
func (l *Log) ParsedFields() map[string]any {
	if l.Fields == nil {
		return nil
	}
	var out map[string]any
	_ = json.Unmarshal(l.Fields, &out)
	return out
}

// IsError retourne true si le niveau est error ou fatal.
func (l *Log) IsError() bool {
	return l.Level == LogLevelError || l.Level == LogLevelFatal
}

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// CreateLogInput est le payload interne pour persister un log.
// Jamais exposé en HTTP — appelé par les services ou middlewares.
type CreateLogInput struct {
	Level      LogLevel       `json:"level"      binding:"required"`
	Message    string         `json:"message"    binding:"required"`
	Service    string         `json:"service"`
	TraceID    string         `json:"traceId"`
	RequestID  string         `json:"requestId"`
	UserID     *uuid.UUID     `json:"userId"`
	Fields     map[string]any `json:"fields"`
	CallerFile string         `json:"callerFile"`
	CallerLine int            `json:"callerLine"`
}

// LogFilterInput regroupe les critères de filtre pour la liste.
type LogFilterInput struct {
	Level     LogLevel `form:"level"`
	Service   string   `form:"service"`
	TraceID   string   `form:"traceId"`
	RequestID string   `form:"requestId"`
	From      string   `form:"from"` // RFC3339
	To        string   `form:"to"`   // RFC3339
	Offset    int      `form:"offset"`
	Limit     int      `form:"limit,default=50"`
}

// LogResponse est la vue publique (lowerCamelCase).
//
// @Description Représentation publique d'un log applicatif persisté
type LogResponse struct {
	ID         uuid.UUID      `json:"logId"`
	Level      LogLevel       `json:"level"`
	Message    string         `json:"message"`
	Service    string         `json:"service,omitempty"`
	TraceID    string         `json:"traceId,omitempty"`
	RequestID  string         `json:"requestId,omitempty"`
	UserID     *uuid.UUID     `json:"userId,omitempty"`
	Fields     map[string]any `json:"fields,omitempty"`
	CallerFile string         `json:"callerFile,omitempty"`
	CallerLine int            `json:"callerLine,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// ToResponse convertit le modèle GORM en vue publique.
func (l *Log) ToResponse() *LogResponse {
	return &LogResponse{
		ID:         l.ID,
		Level:      l.Level,
		Message:    l.Message,
		Service:    l.Service,
		TraceID:    l.TraceID,
		RequestID:  l.RequestID,
		UserID:     l.UserID,
		Fields:     l.ParsedFields(),
		CallerFile: l.CallerFile,
		CallerLine: l.CallerLine,
		CreatedAt:  l.CreatedAt,
	}
}

// NewLogFromInput construit un Log depuis un CreateLogInput,
// en sérialisant Fields en JSONB.
func NewLogFromInput(input *CreateLogInput) (*Log, error) {
	l := &Log{
		Level:      input.Level,
		Message:    input.Message,
		Service:    input.Service,
		TraceID:    input.TraceID,
		RequestID:  input.RequestID,
		UserID:     input.UserID,
		CallerFile: input.CallerFile,
		CallerLine: input.CallerLine,
	}
	if input.Fields != nil {
		b, err := json.Marshal(input.Fields)
		if err != nil {
			return nil, err
		}
		l.Fields = b
	}
	return l, nil
}
