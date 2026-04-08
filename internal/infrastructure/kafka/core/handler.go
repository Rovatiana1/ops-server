package core

import "context"

// Handler processes a raw Kafka message payload.
// Implementations must be idempotent.
type Handler interface {
	Handle(ctx context.Context, payload []byte) error
	// EventType returns the event type this handler is responsible for.
	EventType() EventType
}
