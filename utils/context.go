package utils

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

type ContextKey string

const (
	ContextKeyRequestID ContextKey = "requestId"
	ContextKeyUserID    ContextKey = "userId"
)

// WithTraceID stores the request/trace id in a context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, traceID)
}

// TraceIDFromContext returns the request/trace id stored in a context.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	traceID, _ := ctx.Value(ContextKeyRequestID).(string)
	return strings.TrimSpace(traceID)
}

// GetTraceID returns the request/trace id for a gin request. It checks the
// request context first, then the gin context, and finally the incoming header.
func GetTraceID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if c.Request != nil {
		if id := TraceIDFromContext(c.Request.Context()); id != "" {
			return id
		}
	}
	if id := strings.TrimSpace(c.GetString(string(ContextKeyRequestID))); id != "" {
		return id
	}
	return strings.TrimSpace(c.GetHeader("X-Request-ID"))
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

func UserIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	userID, _ := ctx.Value(ContextKeyUserID).(string)
	return userID
}

// DumpIncomingContext extracts a compact representation of the request context for logs.
func DumpIncomingContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	return Dump(map[string]any{
		"requestId": TraceIDFromContext(ctx),
		"userId":    UserIDFromContext(ctx),
	})
}
