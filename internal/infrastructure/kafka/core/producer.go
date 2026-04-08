package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// Producer defines a generic Kafka producer.
type Producer interface {
	Publish(ctx context.Context, key string, event any) error
	Close() error
}

type producer struct {
	writer *kafka.Writer
}

// NewProducer wraps a kafka.Writer in the Producer interface.
func NewProducer(writer *kafka.Writer) Producer {
	return &producer{writer: writer}
}

// Publish serialises event to JSON and writes it to Kafka.
func (p *producer) Publish(ctx context.Context, key string, event any) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: payload,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write kafka message: %w", err)
	}
	return nil
}

func (p *producer) Close() error {
	return p.writer.Close()
}
