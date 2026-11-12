package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/utils"
)

const headerRequestID = "X-Request-ID"

// RequestID ensures every request has a trace id, exposed via header and context.
// Downstream handlers can read it with utils.GetTraceID(c) or
// utils.TraceIDFromContext(ctx).
func (m *Middleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := strings.TrimSpace(c.GetHeader(headerRequestID))
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(string(utils.ContextKeyRequestID), requestID)
		c.Header(headerRequestID, requestID)

		c.Request = c.Request.WithContext(utils.WithTraceID(c.Request.Context(), requestID))
		c.Next()
	}
}
