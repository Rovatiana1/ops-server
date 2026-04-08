package kafka_producer

import (
	"context"

	"ops-server/internal/infrastructure/kafka/core"
	"ops-server/pkg/logger"

	"go.uber.org/zap"
)

// SignupProducer publishes UserSignupEvent to Kafka.
type SignupProducer struct {
	producer core.Producer
}

// NewSignupProducer creates a SignupProducer.
func NewSignupProducer(p core.Producer) *SignupProducer {
	return &SignupProducer{producer: p}
}

// Publish sends a signup event to the signup topic.
func (s *SignupProducer) Publish(ctx context.Context, event core.UserSignupEvent) error {
	log := logger.FromContext(ctx)
	log.Info("publishing signup event",
		zap.String("userId", event.UserID),
		zap.String("eventId", event.EventID),
	)

	if err := s.producer.Publish(ctx, event.UserID, event); err != nil {
		log.Error("failed to publish signup event", zap.Error(err))
		return err
	}
	return nil
}
