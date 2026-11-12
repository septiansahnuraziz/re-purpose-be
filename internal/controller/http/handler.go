package http

import (
	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/internal/controller/middleware"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

type Router struct {
	middleware     *middleware.Middleware
	contentService entity.IContentService
	systemService  entity.ISystemService
	userService    entity.IUserService
	sponsorService entity.ISponsorService
}

// RouteService registers the API routes.
//
// Canonical routes live under /api/v1. The same handlers are also mounted under
// /api to match the endpoint paths documented in the re-purpose-fe integration
// comments, so the frontend can integrate without changing its paths.
func RouteService(
	app *gin.RouterGroup,
	middleware *middleware.Middleware,
	contentService entity.IContentService,
	systemService entity.ISystemService,
	userService entity.IUserService,
	sponsorService entity.ISponsorService,
) {
	router := &Router{
		middleware:     middleware,
		contentService: contentService,
		systemService:  systemService,
		userService:    userService,
		sponsorService: sponsorService,
	}

	api := app.Group("api")

	v1 := api.Group("v1")
	v1.GET("/ping", ping)
	router.registerRoutes(v1)

	// Frontend-compatible alias: /api/content/generate, /api/system/status, etc.
	api.GET("/ping", ping)
	router.registerRoutes(api)
}

func (r *Router) registerRoutes(group *gin.RouterGroup) {
	r.initContentRoutes(group)
	r.initSystemRoutes(group)
	r.initUserRoutes(group)
	r.initAnalyticsRoutes(group)
	r.initErrorRoutes(group)
}

func ping(c *gin.Context) {
	httpresponse.NoContent(c)
}
