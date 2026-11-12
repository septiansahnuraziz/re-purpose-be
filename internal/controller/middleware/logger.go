package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/sirupsen/logrus"
)

// Logger writes one structured access log line per request. It must be
// registered after RequestID so every line carries the request/trace id.
func (m *Middleware) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(startedAt)
		status := c.Writer.Status()

		fields := logrus.Fields{
			"method":    c.Request.Method,
			"path":      path,
			"status":    status,
			"latency":   latency.String(),
			"latencyMs": latency.Milliseconds(),
			"clientIp":  c.ClientIP(),
			"userAgent": c.Request.UserAgent(),
		}
		if query != "" {
			fields["query"] = query
		}
		if size := c.Writer.Size(); size > 0 {
			fields["responseSize"] = size
		}
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		entry := logger.WithFields(c.Request.Context(), fields)

		switch {
		case status >= 500:
			entry.Error("request completed")
		case status >= 400:
			entry.Warn("request completed")
		default:
			entry.Info("request completed")
		}
	}
}
