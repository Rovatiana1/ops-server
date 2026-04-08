package core

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"

	"ops-server/pkg/logger"
)

const (
	maxRetries    = 3
	retryBaseWait = 500 * time.Millisecond
)

// Consumer reads from a Kafka topic and dispatches messages to a Handler.
// It implements at-least-once processing with retry + DLQ.
type Consumer struct {
	reader      *kafka.Reader
	handler     Handler
	dlqProducer Producer
	topic       string
}

// NewConsumer creates a Consumer for the given topic.
func NewConsumer(reader *kafka.Reader, handler Handler, dlq Producer) *Consumer {
	return &Consumer{
		reader:      reader,
		handler:     handler,
		dlqProducer: dlq,
		topic:       reader.Config().Topic,
	}
}

// Run starts the consume loop. Blocks until ctx is cancelled.
func (c *Consumer) Run(ctx context.Context) {
	log := logger.L().With(zap.String("topic", c.topic))
	log.Info("consumer started")

	for {
		select {
		case <-ctx.Done():
			log.Info("consumer shutting down")
			_ = c.reader.Close()
			return
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error("fetch message error", zap.Error(err))
			continue
		}

		if err := c.processWithRetry(ctx, msg); err != nil {
			log.Error("message sent to DLQ",
				zap.String("key", string(msg.Key)),
				zap.Error(err),
			)
			c.sendToDLQ(ctx, msg, err)
		}

		// Manual commit — at-least-once semantics
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Error("commit error", zap.Error(err))
		}
	}
}

func (c *Consumer) processWithRetry(ctx context.Context, msg kafka.Message) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := c.handler.Handle(ctx, msg.Value); err != nil {
			lastErr = err
			wait := time.Duration(attempt) * retryBaseWait
			logger.L().Warn("handler error, retrying",
				zap.String("topic", c.topic),
				zap.Int("attempt", attempt),
				zap.Duration("wait", wait),
				zap.Error(err),
			)
			time.Sleep(wait)
			continue
		}
		return nil
	}
	return fmt.Errorf("all %d retries exhausted: %w", maxRetries, lastErr)
}

// DLQEnvelope wraps a failed message for the dead-letter queue.
type DLQEnvelope struct {
	OriginalTopic string    `json:"originalTopic"`
	Key           string    `json:"key"`
	Payload       string    `json:"payload"`
	Error         string    `json:"error"`
	FailedAt      time.Time `json:"failedAt"`
}

func (c *Consumer) sendToDLQ(ctx context.Context, msg kafka.Message, processingErr error) {
	envelope := DLQEnvelope{
		OriginalTopic: c.topic,
		Key:           string(msg.Key),
		Payload:       string(msg.Value),
		Error:         processingErr.Error(),
		FailedAt:      time.Now().UTC(),
	}
	payload, _ := json.Marshal(envelope)
	_ = c.dlqProducer.Publish(ctx, string(msg.Key), payload)
}
