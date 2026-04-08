package core

import (
	"time"

	"github.com/google/uuid"
)

// EventType identifies the kind of domain event.
type EventType string

const (
	EventTypeUserSignup EventType = "user.signup"
	EventTypeUserSignin EventType = "user.signin"
)

// BaseEvent is the envelope wrapping every Kafka message.
// All domain events embed or compose this struct.
type BaseEvent struct {
	EventID   string    `json:"eventId"`   // UUID v4 — used for idempotency
	EventType EventType `json:"eventType"`
	OccurredAt time.Time `json:"occurredAt"`
	Version   int       `json:"version"`   // schema version for forward-compat
	TraceID   string    `json:"traceId,omitempty"`
}

// NewBaseEvent creates a BaseEvent with a fresh UUID and current timestamp.
func NewBaseEvent(t EventType) BaseEvent {
	return BaseEvent{
		EventID:    uuid.New().String(),
		EventType:  t,
		OccurredAt: time.Now().UTC(),
		Version:    1,
	}
}

// --- Typed domain events ---

// UserSignupEvent is published when a user completes registration.
type UserSignupEvent struct {
	BaseEvent
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UserSigninEvent is published after a successful sign-in.
type UserSigninEvent struct {
	BaseEvent
	UserID string `json:"userId"`
	Email  string `json:"email"`
}
