package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecoveryReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(NewMiddleware().Recovery())
	engine.GET("/panic", func(_ *gin.Context) {
		panic("boom")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}
