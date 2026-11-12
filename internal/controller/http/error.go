package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

func (r *Router) initErrorRoutes(app *gin.RouterGroup) {
	errorsGroup := app.Group("errors")
	{
		errorsGroup.GET("", r.GetListErrorCodes)
	}
}

// GetListErrorCodes returns the localized error catalog.
// GET /api/v1/errors?lang=id
func (r *Router) GetListErrorCodes(c *gin.Context) {
	lang := strings.ToUpper(strings.TrimSpace(c.Query("lang")))
	if lang == "" {
		lang = httpresponse.ErrorLangID
	}

	httpresponse.NewHttpResponse().
		WithMessage(successResponse["GetListErrorCodes"]).
		WithData(httpresponse.GetErrorList(lang)).
		ToWrapperResponseDTO(c, http.StatusOK)
}
