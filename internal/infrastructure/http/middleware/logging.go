package middleware

import (
	"log/slog"
	"time"

	"github.com/akaporn-katip/go-project-structure-template/internal/infrastructure/observability"
	"github.com/gin-gonic/gin"
)

type LoggingMiddleware struct {
	logger *slog.Logger
}

func NewLoggingMiddleware(logger *slog.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

func (l *LoggingMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Get trace context
		ctx := c.Request.Context()
		traceID := observability.GetTraceID(ctx)
		spanID := observability.GetSpanID(ctx)

		// Log attributes
		attrs := []any{
			"method", method,
			"path", path,
			"status", statusCode,
			"duration_ms", duration.Milliseconds(),
			"client_ip", clientIP,
			"user_agent", userAgent,
			"trace_id", traceID,
			"span_id", spanID,
		}

		// Check if there were any errors
		if len(c.Errors) > 0 {
			attrs = append(attrs, "errors", c.Errors.String())
		}

		// Log based on status code
		switch {
		case statusCode >= 500:
			l.logger.ErrorContext(ctx, "HTTP request completed with server error", attrs...)
		case statusCode >= 400:
			l.logger.WarnContext(ctx, "HTTP request completed with client error", attrs...)
		default:
			l.logger.InfoContext(ctx, "HTTP request completed", attrs...)
		}
	}
}
