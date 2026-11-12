package utils

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestTraceIDContext(t *testing.T) {
	ctx := WithTraceID(context.Background(), "trace-123")

	if got := TraceIDFromContext(ctx); got != "trace-123" {
		t.Fatalf("TraceIDFromContext() = %q, want trace-123", got)
	}
}

func TestGetTraceIDPrefersRequestContext(t *testing.T) {
	ctx := WithTraceID(context.Background(), "trace-from-context")

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", nil).WithContext(ctx)
	c.Set(string(ContextKeyRequestID), "trace-from-gin")

	if got := GetTraceID(c); got != "trace-from-context" {
		t.Fatalf("GetTraceID() = %q, want trace-from-context", got)
	}
}
