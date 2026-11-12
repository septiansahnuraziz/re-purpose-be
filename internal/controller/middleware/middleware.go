package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/repurpose-ai/re-purpose-be/config"
)

type Middleware struct {
	bypass          bool
	jwtSecret       string
	jwtIssuer       string
	demoUserID      string
	authRateLimiter *RateLimiter
}

func NewMiddleware() *Middleware {
	return &Middleware{
		bypass:          config.AuthBypass(),
		jwtSecret:       config.JWTSecret(),
		jwtIssuer:       config.JWTIssuer(),
		demoUserID:      config.DemoUserID(),
		authRateLimiter: NewRateLimiter(config.AuthRateLimitRequests(), config.AuthRateLimitWindow()),
	}
}

// AuthRateLimit throttles credential endpoints per client IP.
func (m *Middleware) AuthRateLimit() gin.HandlerFunc {
	if m.authRateLimiter == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return m.authRateLimiter.Middleware()
}
