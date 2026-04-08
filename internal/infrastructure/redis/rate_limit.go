package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter defines a token-bucket / sliding-window rate limiter.
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	Remaining(ctx context.Context, key string, limit int) (int, error)
}

type rateLimiter struct {
	client *redis.Client
}

// NewRateLimiter creates a Redis-backed sliding-window rate limiter.
func NewRateLimiter(client *redis.Client) RateLimiter {
	return &rateLimiter{client: client}
}

// slidingWindowScript increments a counter per window using a sorted set of timestamps.
var slidingWindowScript = redis.NewScript(`
	local key    = KEYS[1]
	local now    = tonumber(ARGV[1])
	local window = tonumber(ARGV[2])
	local limit  = tonumber(ARGV[3])
	local uid    = ARGV[4]

	-- Remove entries outside the window
	redis.call("ZREMRANGEBYSCORE", key, 0, now - window)

	local count = redis.call("ZCARD", key)
	if count < limit then
		redis.call("ZADD", key, now, uid)
		redis.call("EXPIRE", key, math.ceil(window / 1000))
		return 1
	end
	return 0
`)

// Allow checks if a request for the given key is within the rate limit.
// key: unique identifier (e.g., "rate:ip:192.168.1.1")
// limit: max requests per window
// window: sliding window duration
func (r *rateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	windowMs := window.Milliseconds()
	uid := fmt.Sprintf("%d", now) // unique member per request

	result, err := slidingWindowScript.Run(
		ctx, r.client,
		[]string{fmt.Sprintf("rate_limit:%s", key)},
		now, windowMs, limit, uid,
	).Int()
	if err != nil {
		return false, fmt.Errorf("rate limit error: %w", err)
	}
	return result == 1, nil
}

// Remaining returns the number of remaining allowed requests in the current window.
func (r *rateLimiter) Remaining(ctx context.Context, key string, limit int) (int, error) {
	count, err := r.client.ZCard(ctx, fmt.Sprintf("rate_limit:%s", key)).Result()
	if err != nil {
		return 0, err
	}
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}
