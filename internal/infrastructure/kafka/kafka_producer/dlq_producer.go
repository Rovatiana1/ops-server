package kafka_producer

import (
	"context"

	"ops-server/internal/infrastructure/kafka/core"
	"ops-server/pkg/logger"

	"go.uber.org/zap"
)

// DLQProducer publishes unprocessable messages to the Dead Letter Queue.
type DLQProducer struct {
	producer core.Producer
}

// NewDLQProducer creates a DLQProducer.
func NewDLQProducer(p core.Producer) *DLQProducer {
	return &DLQProducer{producer: p}
}

// Publish sends an envelope to the DLQ topic.
func (d *DLQProducer) Publish(ctx context.Context, key string, payload any) error {
	log := logger.FromContext(ctx)
	log.Warn("sending message to DLQ", zap.String("key", key))

	if err := d.producer.Publish(ctx, key, payload); err != nil {
		log.Error("failed to publish to DLQ", zap.Error(err))
		return err
	}
	return nil
}
