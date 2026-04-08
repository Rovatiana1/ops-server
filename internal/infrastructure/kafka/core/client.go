package core

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"ops-server/configs"
	"ops-server/pkg/logger"

	"github.com/segmentio/kafka-go"
)

// DialBroker verifies connectivity to the first Kafka broker.
func DialBroker(cfg configs.KafkaConfig) error {
	if len(cfg.Brokers) == 0 {
		return fmt.Errorf("no kafka brokers configured")
	}

	host, portStr, err := net.SplitHostPort(cfg.Brokers[0])
	if err != nil {
		return fmt.Errorf("invalid kafka broker address: %w", err)
	}
	port, _ := strconv.Atoi(portStr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := kafka.DialLeader(ctx, "tcp", cfg.Brokers[0], "", 0)
	if err != nil {
		return fmt.Errorf("kafka dial failed (%s:%d): %w", host, port, err)
	}
	defer conn.Close()

	logger.S().Infow("kafka connected", "broker", cfg.Brokers[0])
	return nil
}

// NewReader creates a kafka-go reader for a given topic.
func NewReader(cfg configs.KafkaConfig, topic string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       topic,
		MinBytes:    cfg.Consumer.MinBytes,
		MaxBytes:    cfg.Consumer.MaxBytes,
		MaxWait:     time.Duration(cfg.Consumer.MaxWait) * time.Millisecond,
		StartOffset: kafka.FirstOffset,
		// Commit manually for at-least-once semantics
		CommitInterval: 0,
	})
}

// NewWriter creates a kafka-go writer for a given topic.
func NewWriter(cfg configs.KafkaConfig, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		BatchSize:    cfg.Producer.BatchSize,
		BatchTimeout: time.Duration(cfg.Producer.BatchTimeout) * time.Millisecond,
		RequiredAcks: kafka.RequiredAcks(cfg.Producer.RequiredAcks),
		Async:        false, // synchronous for reliability
	}
}
