package workers

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"ops-server/internal/infrastructure/kafka/kafka_consumer"
	"ops-server/pkg/logger"
)

// Runnable is anything that can be started with a context.
type Runnable interface {
	Run(ctx context.Context)
}

// WorkerPool manages a set of background workers (Kafka consumers, etc.).
type WorkerPool struct {
	workers []Runnable
}

// NewWorkerPool creates a WorkerPool from the provided consumers.
func NewWorkerPool(
	signupConsumer *kafka_consumer.SignupConsumer,
	signinConsumer *kafka_consumer.SigninConsumer,
	retryConsumer *kafka_consumer.RetryConsumer,
) *WorkerPool {
	return &WorkerPool{
		workers: []Runnable{
			signupConsumer,
			signinConsumer,
			retryConsumer,
		},
	}
}

// Run starts all workers concurrently and blocks until ctx is cancelled.
// Each worker runs in its own goroutine; the WaitGroup ensures clean shutdown.
func (wp *WorkerPool) Run(ctx context.Context) {
	log := logger.L()
	log.Info("worker pool starting", zap.Int("workers", len(wp.workers)))

	var wg sync.WaitGroup

	for i, w := range wp.workers {
		wg.Add(1)
		worker := w // capture loop variable
		idx := i

		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Error("worker panicked",
						zap.Int("worker", idx),
						zap.Any("panic", r),
					)
				}
			}()
			worker.Run(ctx)
		}()
	}

	<-ctx.Done()
	log.Info("worker pool shutting down — waiting for workers to finish")
	wg.Wait()
	log.Info("all workers stopped")
}
