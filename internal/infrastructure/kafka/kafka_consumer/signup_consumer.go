package kafka_consumer

import (
	"context"

	"ops-server/configs"
	"ops-server/internal/infrastructure/kafka/core"
	"ops-server/pkg/logger"

	"go.uber.org/zap"
)

// SignupConsumer wires a Reader + Handler + DLQ into a runnable consumer.
type SignupConsumer struct {
	consumer *core.Consumer
}

// NewSignupConsumer builds a SignupConsumer from config, handler and DLQ producer.
func NewSignupConsumer(cfg configs.KafkaConfig, handler core.Handler, dlq core.Producer) *SignupConsumer {
	reader := core.NewReader(cfg, cfg.Topics.Signup)
	c := core.NewConsumer(reader, handler, dlq)
	return &SignupConsumer{consumer: c}
}

// Run starts the consume loop. Blocks until ctx is done.
func (s *SignupConsumer) Run(ctx context.Context) {
	logger.L().Info("signup consumer starting", zap.String("topic", "user.signup"))
	s.consumer.Run(ctx)
}
