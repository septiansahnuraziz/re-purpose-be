package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

func (r *Router) initSystemRoutes(app *gin.RouterGroup) {
	system := app.Group("system")
	{
		system.GET("/status", r.GetSystemStatus)
	}
}

// GetSystemStatus handles TITIK INTEGRASI API #4 (navbar + footer).
// GET /api/v1/system/status
func (r *Router) GetSystemStatus(c *gin.Context) {
	response := r.systemService.Status(c.Request.Context())

	httpresponse.NewHttpResponse().
		WithMessage(successResponse["GetSystemStatus"]).
		WithData(response).
		ToWrapperResponseDTO(c, http.StatusOK)
}
