package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/internal/controller/middleware"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

func (r *Router) initUserRoutes(app *gin.RouterGroup) {
	user := app.Group("user", r.middleware.Auth())
	{
		user.GET("/me", r.GetUserProfile)
	}
}

// GetUserProfile handles TITIK INTEGRASI API #5.
// GET /api/v1/user/me
func (r *Router) GetUserProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	response, err := r.userService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		logger.FromContext(c.Request.Context()).
			WithField("userId", userID).
			WithError(err).
			Error("get user profile failed")
		httpErrorHandler(c, err)
		return
	}

	httpresponse.NewHttpResponse().
		WithMessage(successResponse["GetUserProfile"]).
		WithData(response).
		ToWrapperResponseDTO(c, http.StatusOK)
}
