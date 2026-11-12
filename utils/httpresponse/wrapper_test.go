package httpresponse

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/utils"
)

func testContext(traceID string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if traceID != "" {
		c.Request = c.Request.WithContext(utils.WithTraceID(c.Request.Context(), traceID))
	}
	return c, rec
}

func TestToWrapperIncludesTraceID(t *testing.T) {
	c, _ := testContext("trace-123")

	wrapper := NewHttpResponse().WithMessage("ok").ToWrapper(c, http.StatusOK)
	if wrapper.ID != "trace-123" || wrapper.TraceID != "trace-123" {
		t.Fatalf("unexpected trace ids: id=%q traceId=%q", wrapper.ID, wrapper.TraceID)
	}

	raw, err := json.Marshal(wrapper)
	if err != nil {
		t.Fatalf("marshal wrapper: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("unmarshal wrapper: %v", err)
	}
	if got := body["traceId"]; got != "trace-123" {
		t.Fatalf("traceId = %v, want trace-123", got)
	}
}

func TestErrorIncludesTraceID(t *testing.T) {
	c, rec := testContext("trace-err")

	Error(c, &ErrorInvalidToken)

	var body struct {
		ID        string `json:"id"`
		TraceID   string `json:"traceId"`
		ErrorCode string `json:"errorCode"`
		Success   bool   `json:"success"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal error envelope: %v", err)
	}
	if body.TraceID != "trace-err" || body.ID != "trace-err" {
		t.Fatalf("unexpected trace ids: id=%q traceId=%q", body.ID, body.TraceID)
	}
	if body.ErrorCode != ErrorInvalidToken.ErrorCode {
		t.Fatalf("errorCode = %q, want %q", body.ErrorCode, ErrorInvalidToken.ErrorCode)
	}
	if body.Success {
		t.Fatal("error envelope must not be successful")
	}
}

func TestGetTraceIDFallsBackToHeader(t *testing.T) {
	c, _ := testContext("")
	c.Request.Header.Set("X-Request-ID", "from-header")

	if got := utils.GetTraceID(c); got != "from-header" {
		t.Fatalf("trace id = %q, want from-header", got)
	}
}
