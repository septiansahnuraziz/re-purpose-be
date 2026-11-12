package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/repurpose-ai/re-purpose-be/internal/controller/middleware"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

func (r *Router) initAnalyticsRoutes(app *gin.RouterGroup) {
	analytics := app.Group("analytics")
	{
		// Optional auth: analytics must also work for anonymous visitors.
		analytics.POST("/sponsor-click", r.middleware.OptionalAuth(), r.TrackSponsorClick)
	}
}

// TrackSponsorClick handles TITIK INTEGRASI API #6 (optional).
// POST /api/v1/analytics/sponsor-click
// Accepts JSON and navigator.sendBeacon(text/plain) bodies.
func (r *Router) TrackSponsorClick(c *gin.Context) {
	var request entity.SponsorClickRequest
	if err := c.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		logger.FromContext(c.Request.Context()).WithError(err).Error("invalid sponsor-click payload")
		httpErrorHandler(c, err)
		return
	}

	response, err := r.sponsorService.TrackClick(
		c.Request.Context(),
		middleware.GetUserID(c),
		request,
		entity.RequestMeta{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()},
	)
	if err != nil {
		logger.FromContext(c.Request.Context()).
			WithField("sponsorId", request.SponsorID).
			WithError(err).
			Error("track sponsor click failed")
		httpErrorHandler(c, err)
		return
	}

	httpresponse.NewHttpResponse().
		WithMessage(successResponse["TrackSponsorClick"]).
		WithData(response).
		ToWrapperResponseDTO(c, http.StatusOK)
}
