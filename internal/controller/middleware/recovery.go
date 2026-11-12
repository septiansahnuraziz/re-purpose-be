package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
	"github.com/sirupsen/logrus"
)

// Recovery recovers from panics, logs the stack trace, and returns the standard
// internal-server-error envelope instead of letting the process crash.
func (m *Middleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			logger.WithFields(c.Request.Context(), logrus.Fields{
				"panic": fmt.Sprintf("%v", recovered),
				"stack": string(debug.Stack()),
			}).Error("panic recovered")

			if !c.Writer.Written() {
				httpresponse.Error(c, &httpresponse.ErrorInternalServer)
			} else {
				c.Abort()
			}
		}()

		c.Next()
	}
}
