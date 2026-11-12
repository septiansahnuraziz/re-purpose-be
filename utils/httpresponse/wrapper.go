package httpresponse

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/utils"
)

type HttpResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func NewHttpResponse() *HttpResponse {
	return new(HttpResponse)
}

func (h *HttpResponse) WithMessage(message string) *HttpResponse {
	h.Message = message
	return h
}

func (h *HttpResponse) WithData(data any) *HttpResponse {
	h.Data = data
	return h
}

// WrapperResponseDTO keeps the response envelope consistent with go-partner,
// plus a `success` flag that the RePurpose frontend expects.
type WrapperResponseDTO struct {
	ID      string `json:"id"`
	TraceID string `json:"traceId"`
	AppName string `json:"appName"`
	Version string `json:"version"`
	Build   string `json:"build"`
	Success bool   `json:"success"`
	HttpResponse
}

// ToWrapper builds the response envelope without writing it, so handlers can
// attach extra top-level fields (used to honour the flat frontend contracts).
func (h *HttpResponse) ToWrapper(c *gin.Context, httpStatus int) WrapperResponseDTO {
	traceID := utils.GetTraceID(c)
	return WrapperResponseDTO{
		ID:           traceID,
		TraceID:      traceID,
		AppName:      config.AppName(),
		Version:      config.AppVersion(),
		Build:        config.AppBuild(),
		Success:      httpStatus >= http.StatusOK && httpStatus < http.StatusMultipleChoices,
		HttpResponse: *h,
	}
}

func (h *HttpResponse) ToWrapperResponseDTO(c *gin.Context, httpStatus int) {
	c.JSON(httpStatus, h.ToWrapper(c, httpStatus))
}

func OK(c *gin.Context, response *HttpResponse) {
	response.ToWrapperResponseDTO(c, http.StatusOK)
}

func Created(c *gin.Context, response *HttpResponse) {
	response.ToWrapperResponseDTO(c, http.StatusCreated)
}

func NoContent(c *gin.Context) {
	NewHttpResponse().ToWrapperResponseDTO(c, http.StatusNoContent)
}
