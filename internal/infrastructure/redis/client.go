package redis

import (
	"context"
	"fmt"
	"time"

	"ops-server/configs"
	"ops-server/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// NewClient creates and validates a Redis client.
func NewClient(cfg configs.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	logger.S().Infow("redis connected",
		"host", cfg.Host,
		"port", cfg.Port,
	)

	return client, nil
}
