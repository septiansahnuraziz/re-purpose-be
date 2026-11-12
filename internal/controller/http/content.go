package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/internal/controller/middleware"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
	"github.com/sirupsen/logrus"
)

func (r *Router) initContentRoutes(app *gin.RouterGroup) {
	content := app.Group("content")
	{
		content.POST("/extract-url", r.middleware.Auth(), r.ExtractURL)
		content.POST("/generate", r.middleware.Auth(), r.GenerateContent)
		content.POST("/reroll", r.middleware.Auth(), r.RerollContent)
		content.GET("/generations", r.middleware.Auth(), r.ListGenerations)
		content.GET("/generations/:id", r.middleware.Auth(), r.GetGeneration)
	}
}

// ExtractURL handles TITIK INTEGRASI API #1.
// POST /api/v1/content/extract-url
func (r *Router) ExtractURL(c *gin.Context) {
	ctx := c.Request.Context()
	log := logger.FromContext(ctx)

	var request entity.ExtractURLRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithError(err).Error("invalid extract-url payload")
		httpErrorHandler(c, err)
		return
	}

	response, err := r.contentService.ExtractURL(ctx, request)
	if err != nil {
		log.WithError(err).WithField("url", request.URL).Error("extract-url failed")
		httpErrorHandler(c, err)
		return
	}

	// Frontend contract (TITIK INTEGRASI API #1) reads title/content/wordCount
	// directly from the root object, so expose them flat in addition to `data`.
	envelope := struct {
		httpresponse.WrapperResponseDTO
		Title     string `json:"title"`
		Content   string `json:"content"`
		WordCount int    `json:"wordCount"`
	}{
		WrapperResponseDTO: httpresponse.NewHttpResponse().
			WithMessage(successResponse["ExtractURL"]).
			WithData(response).
			ToWrapper(c, http.StatusOK),
		Title:     response.Title,
		Content:   response.Content,
		WordCount: response.WordCount,
	}

	c.JSON(http.StatusOK, envelope)
}

// GenerateContent handles TITIK INTEGRASI API #2.
// POST /api/v1/content/generate
func (r *Router) GenerateContent(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)
	log := logger.FromContext(ctx).WithField("userId", userID)

	var request entity.GenerateContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithError(err).Error("invalid generate payload")
		httpErrorHandler(c, err)
		return
	}

	response, err := r.contentService.Generate(ctx, userID, request)
	if err != nil {
		log.WithError(err).WithField("tone", request.Tone).Error("generate failed")
		httpErrorHandler(c, err)
		return
	}

	writeGenerationResponse(c, successResponse["GenerateContent"], response)
}

// RerollContent handles TITIK INTEGRASI API #3.
// POST /api/v1/content/reroll
func (r *Router) RerollContent(c *gin.Context) {
	ctx := c.Request.Context()
	userID := middleware.GetUserID(c)
	log := logger.FromContext(ctx).WithField("userId", userID)

	var request entity.GenerateContentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithError(err).Error("invalid reroll payload")
		httpErrorHandler(c, err)
		return
	}

	response, err := r.contentService.Reroll(ctx, userID, request)
	if err != nil {
		log.WithError(err).WithField("tone", request.Tone).Error("reroll failed")
		httpErrorHandler(c, err)
		return
	}

	writeGenerationResponse(c, successResponse["RerollContent"], response)
}

// writeGenerationResponse keeps the frontend contract for TITIK INTEGRASI API
// #2/#3: `res.data` is the PlatformOutputs object, with generationId exposed as
// a sibling field on the envelope.
func writeGenerationResponse(c *gin.Context, message string, response *entity.GenerateContentResponse) {
	envelope := struct {
		httpresponse.WrapperResponseDTO
		GenerationID string `json:"generationId"`
	}{
		WrapperResponseDTO: httpresponse.NewHttpResponse().
			WithMessage(message).
			WithData(response.Data).
			ToWrapper(c, http.StatusOK),
		GenerationID: response.GenerationID,
	}

	c.JSON(http.StatusOK, envelope)
}

// ListGenerations returns the current user's generation history.
// GET /api/v1/content/generations?page=1&size=10
func (r *Router) ListGenerations(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, size := paginationQuery(c)

	response, err := r.contentService.ListGenerations(c.Request.Context(), userID, page, size)
	if err != nil {
		logger.FromContext(c.Request.Context()).WithField("userId", userID).WithError(err).Error("list generations failed")
		httpErrorHandler(c, err)
		return
	}

	httpresponse.NewHttpResponse().
		WithMessage(successResponse["ListGenerations"]).
		WithData(response).
		ToWrapperResponseDTO(c, http.StatusOK)
}

// GetGeneration returns one generation by id.
// GET /api/v1/content/generations/{id}
func (r *Router) GetGeneration(c *gin.Context) {
	userID := middleware.GetUserID(c)
	generationID := utils.ExpectedString(c.Param("id"))
	if generationID == "" {
		httpErrorHandler(c, &httpresponse.ErrorBadRequest)
		return
	}

	response, err := r.contentService.GetGeneration(c.Request.Context(), userID, generationID)
	if err != nil {
		logger.FromContext(c.Request.Context()).
			WithFields(logrus.Fields{"userId": userID, "generationId": generationID}).
			WithError(err).
			Error("get generation failed")
		httpErrorHandler(c, err)
		return
	}

	httpresponse.NewHttpResponse().
		WithMessage(successResponse["GetGeneration"]).
		WithData(response).
		ToWrapperResponseDTO(c, http.StatusOK)
}
