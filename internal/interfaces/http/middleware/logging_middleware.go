package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"ops-server/pkg/logger"
)

// RequestLogger logs every HTTP request with method, path, status, latency and request ID.
func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()

		// Attach or generate request ID
		requestID := ctx.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx.Header("X-Request-ID", requestID)

		// Enrich context logger
		reqCtx := logger.WithRequestID(ctx.Request.Context(), requestID)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		ctx.Next()

		latency := time.Since(start)
		status := ctx.Writer.Status()

		log := logger.FromContext(ctx.Request.Context())

		logFn := log.Info
		if status >= 500 {
			logFn = log.Error
		} else if status >= 400 {
			logFn = log.Warn
		}

		logFn("http request",
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.FullPath()),
			zap.String("rawQuery", ctx.Request.URL.RawQuery),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", ctx.ClientIP()),
			zap.String("requestId", requestID),
		)
	}
}
