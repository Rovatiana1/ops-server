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

// SigninHandler processes UserSigninEvent messages.
type SigninHandler struct {
	cache redisInfra.Cache
}

// NewSigninHandler creates a SigninHandler.
func NewSigninHandler(cache redisInfra.Cache) *SigninHandler {
	return &SigninHandler{cache: cache}
}

func (h *SigninHandler) EventType() core.EventType {
	return core.EventTypeUserSignin
}

// Handle deserialises and processes the signin event.
func (h *SigninHandler) Handle(ctx context.Context, payload []byte) error {
	log := logger.FromContext(ctx)

	var event core.UserSigninEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("invalid signin event payload: %w", err)
	}

	// Idempotency check
	idempotencyKey := fmt.Sprintf("processed:signin:%s", event.EventID)
	already, err := h.cache.Exists(ctx, idempotencyKey)
	if err != nil {
		log.Warn("idempotency check failed, proceeding", zap.Error(err))
	}
	if already {
		log.Info("signin event already processed, skipping",
			zap.String("eventId", event.EventID),
		)
		return nil
	}

	log.Info("processing signin event",
		zap.String("eventId", event.EventID),
		zap.String("userId", event.UserID),
	)

	// Domain processing: e.g. audit log, last-login update…

	_ = h.cache.Set(ctx, idempotencyKey, "1", 24*60*60*1_000_000_000)
	return nil
}
