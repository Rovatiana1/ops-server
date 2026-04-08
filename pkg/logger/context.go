package logger

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "logger"

// WithContext returns a context carrying the given logger.
func WithContext(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext extracts the logger from context, falling back to the global logger.
func FromContext(ctx context.Context) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return L()
}

// WithFields returns a context whose logger is enriched with the given fields.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	l := FromContext(ctx).With(fields...)
	return WithContext(ctx, l)
}

// WithRequestID adds a requestId field to the context logger.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return WithFields(ctx, zap.String("requestId", requestID))
}

// WithUserID adds a userId field to the context logger.
func WithUserID(ctx context.Context, userID string) context.Context {
	return WithFields(ctx, zap.String("userId", userID))
}
