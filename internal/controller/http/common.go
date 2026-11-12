package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/repurpose-ai/re-purpose-be/internal/service"
	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

var successResponse = map[string]string{
	"ExtractURL":        "Success extract content from URL",
	"GenerateContent":   "Success generate multi-platform content",
	"RerollContent":     "Success reroll content variations",
	"GetGeneration":     "Success get generation detail",
	"ListGenerations":   "Success get generation history",
	"GetSystemStatus":   "Success get system status",
	"GetUserProfile":    "Success get user profile",
	"TrackSponsorClick": "Success track sponsor event",
	"GetListErrorCodes": "Success get list error codes",
}

func httpErrorHandler(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		type fieldError struct {
			Field string `json:"field"`
			Error string `json:"error"`
		}

		details := make([]fieldError, 0, len(validationErrors))
		for _, field := range validationErrors {
			details = append(details, fieldError{
				Field: utils.ToCamelCase(field.Field()),
				Error: msgForTag(field.Tag()),
			})
		}

		httpresponse.Error(c, httpresponse.NewHTTPError().
			WithCode(http.StatusBadRequest).
			WithErrorCode(httpresponse.ErrorValidationFailed.ErrorCode).
			WithMessage(httpresponse.ErrorValidationFailed.Message).
			WithData(details))
		return
	}

	switch {
	case errors.Is(err, service.ErrInvalidURL):
		httpresponse.Error(c, &httpresponse.ErrorInvalidURL)
	case errors.Is(err, service.ErrContentTooShort):
		httpresponse.Error(c, &httpresponse.ErrorContentTooShort)
	case errors.Is(err, service.ErrInvalidTone):
		httpresponse.Error(c, &httpresponse.ErrorInvalidTone)
	case errors.Is(err, service.ErrSponsorEventInvalid):
		httpresponse.Error(c, &httpresponse.ErrorInvalidSponsorEvent)
	case errors.Is(err, service.ErrExtractionFailed):
		httpresponse.Error(c, &httpresponse.ErrorURLExtractionFailed)
	case errors.Is(err, service.ErrLLMUnavailable):
		httpresponse.Error(c, &httpresponse.ErrorLLMUnavailable)
	case errors.Is(err, service.ErrGenerationNotFound):
		httpresponse.Error(c, &httpresponse.ErrorGenerationNotFound)
	case errors.Is(err, service.ErrUserNotFound):
		httpresponse.Error(c, &httpresponse.ErrorUserNotFound)
	case errors.Is(err, service.ErrBadRequest):
		httpresponse.Error(c, &httpresponse.ErrorBadRequest)
	case errors.Is(err, service.ErrNotFound):
		httpresponse.Error(c, &httpresponse.ErrorNotFound)
	default:
		httpresponse.Error(c, &httpresponse.ErrorInternalServer)
	}
}

func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "url":
		return "Invalid URL format"
	case "email":
		return "Invalid email address"
	case "min":
		return "The value is below the minimum allowed length"
	case "max":
		return "The value exceeds the maximum allowed length"
	case "oneof":
		return "The value must be one of the allowed options"
	default:
		return utils.WriteStringTemplate("Validation failed for the '%s' tag", tag)
	}
}

// paginationQuery parses and clamps page/size query parameters.
func paginationQuery(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size <= 0 {
		size = 10
	}
	if size > 100 {
		size = 100
	}

	return page, size
}
