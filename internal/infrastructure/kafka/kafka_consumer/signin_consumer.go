package kafka_consumer

import (
	"context"

	"ops-server/configs"
	"ops-server/internal/infrastructure/kafka/core"
	"ops-server/pkg/logger"

	"go.uber.org/zap"
)

// SigninConsumer wires a Reader + Handler + DLQ into a runnable consumer.
type SigninConsumer struct {
	consumer *core.Consumer
}

// NewSigninConsumer builds a SigninConsumer.
func NewSigninConsumer(cfg configs.KafkaConfig, handler core.Handler, dlq core.Producer) *SigninConsumer {
	reader := core.NewReader(cfg, cfg.Topics.Signin)
	c := core.NewConsumer(reader, handler, dlq)
	return &SigninConsumer{consumer: c}
}

// Run starts the consume loop. Blocks until ctx is done.
func (s *SigninConsumer) Run(ctx context.Context) {
	logger.L().Info("signin consumer starting", zap.String("topic", "user.signin"))
	s.consumer.Run(ctx)
}
