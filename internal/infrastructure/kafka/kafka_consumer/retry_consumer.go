package kafka_consumer

import (
	"context"

	"ops-server/configs"
	"ops-server/internal/infrastructure/kafka/core"
	"ops-server/pkg/logger"

	"go.uber.org/zap"
)

// RetryConsumer processes messages from the retry topic.
type RetryConsumer struct {
	consumer *core.Consumer
}

// NewRetryConsumer builds a RetryConsumer.
// The handler passed in should route events back to the appropriate domain handler.
func NewRetryConsumer(cfg configs.KafkaConfig, handler core.Handler, dlq core.Producer) *RetryConsumer {
	reader := core.NewReader(cfg, cfg.Topics.Retry)
	c := core.NewConsumer(reader, handler, dlq)
	return &RetryConsumer{consumer: c}
}

// Run starts the retry consume loop.
func (r *RetryConsumer) Run(ctx context.Context) {
	logger.L().Info("retry consumer starting", zap.String("topic", "user.retry"))
	r.consumer.Run(ctx)
}
