package kafka_handler

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"ops-server/internal/infrastructure/kafka/core"
	redisInfra "ops-server/internal/infrastructure/redis"
	"ops-server/pkg/logger"
)

// SignupHandler processes UserSignupEvent messages.
// Idempotency is enforced via an Redis event-ID cache.
type SignupHandler struct {
	cache redisInfra.Cache
}

// NewSignupHandler creates a SignupHandler.
func NewSignupHandler(cache redisInfra.Cache) *SignupHandler {
	return &SignupHandler{cache: cache}
}

func (h *SignupHandler) EventType() core.EventType {
	return core.EventTypeUserSignup
}

// Handle deserialises and processes the signup event.
func (h *SignupHandler) Handle(ctx context.Context, payload []byte) error {
	log := logger.FromContext(ctx)

	var event core.UserSignupEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("invalid signup event payload: %w", err)
	}

	// Idempotency check — skip if already processed
	idempotencyKey := fmt.Sprintf("processed:signup:%s", event.EventID)
	already, err := h.cache.Exists(ctx, idempotencyKey)
	if err != nil {
		log.Warn("idempotency check failed, proceeding", zap.Error(err))
	}
	if already {
		log.Info("signup event already processed, skipping",
			zap.String("eventId", event.EventID),
		)
		return nil
	}

	log.Info("processing signup event",
		zap.String("eventId", event.EventID),
		zap.String("userId", event.UserID),
		zap.String("email", event.Email),
	)

	// Domain processing: e.g. send welcome email, provision resources…
	// This is intentionally a stub — real logic goes in a use-case called from here.

	// Mark as processed (TTL = 24h to avoid unbounded growth)
	_ = h.cache.Set(ctx, idempotencyKey, "1", 24*60*60*1_000_000_000)

	return nil
}
