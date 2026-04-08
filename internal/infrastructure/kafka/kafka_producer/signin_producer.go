package kafka_producer

import (
	"context"

	"ops-server/internal/infrastructure/kafka/core"
	"ops-server/pkg/logger"

	"go.uber.org/zap"
)

// SigninProducer publishes UserSigninEvent to Kafka.
type SigninProducer struct {
	producer core.Producer
}

// NewSigninProducer creates a SigninProducer.
func NewSigninProducer(p core.Producer) *SigninProducer {
	return &SigninProducer{producer: p}
}

// Publish sends a signin event to the signin topic.
func (s *SigninProducer) Publish(ctx context.Context, event core.UserSigninEvent) error {
	log := logger.FromContext(ctx)
	log.Info("publishing signin event",
		zap.String("userId", event.UserID),
		zap.String("eventId", event.EventID),
	)

	if err := s.producer.Publish(ctx, event.UserID, event); err != nil {
		log.Error("failed to publish signin event", zap.Error(err))
		return err
	}
	return nil
}
