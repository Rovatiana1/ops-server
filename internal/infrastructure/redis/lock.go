package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Lock defines a distributed lock interface.
type Lock interface {
	Acquire(ctx context.Context, key string, ttl time.Duration) (string, error)
	Release(ctx context.Context, key, token string) error
}

type lock struct {
	client *redis.Client
}

// NewLock creates a Redis-backed distributed lock.
func NewLock(client *redis.Client) Lock {
	return &lock{client: client}
}

// Acquire tries to acquire the lock for the given key.
// Returns a unique token to be used for release, or an error if the lock is held.
func (l *lock) Acquire(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := uuid.New().String()
	lockKey := fmt.Sprintf("lock:%s", key)

	ok, err := l.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("lock acquire error: %w", err)
	}
	if !ok {
		return "", fmt.Errorf("lock already held for key: %s", key)
	}
	return token, nil
}

// Release releases the lock only if the token matches (prevents accidental release).
var releaseScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`)

func (l *lock) Release(ctx context.Context, key, token string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	result, err := releaseScript.Run(ctx, l.client, []string{lockKey}, token).Int()
	if err != nil {
		return fmt.Errorf("lock release error: %w", err)
	}
	if result == 0 {
		return fmt.Errorf("lock not released: token mismatch or already expired for key: %s", key)
	}
	return nil
}
