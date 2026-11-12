package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/repurpose-ai/re-purpose-be/commons"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/repurpose-ai/re-purpose-be/utils/httpresponse"
)

const (
	contextUserID    = "userId"
	contextEmail     = "email"
	contextTier      = "tier"
	contextSessionID = "sessionId"
	contextTokenID   = "tokenId"
)

// Auth validates a Bearer JWT and stores the identity in the request context.
// When auth.bypass is enabled (local development), the demo user is injected.
func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.bypass {
			m.applyIdentity(c, m.demoUserID, "demo@repurpose.ai", commons.TierPro, "", "")
			c.Next()
			return
		}

		claims, ok := m.parseBearerToken(c)
		if !ok {
			httpresponse.Error(c, &httpresponse.ErrorInvalidToken)
			return
		}

		tier := claims.Tier
		if tier == "" {
			tier = commons.TierFree
		}
		m.applyIdentity(c, claims.UserID, claims.Email, tier, claims.SessionID, claims.ID)
		c.Next()
	}
}

// OptionalAuth attaches identity when a valid token is present, but never rejects.
func (m *Middleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if m.bypass {
			m.applyIdentity(c, m.demoUserID, "demo@repurpose.ai", commons.TierPro, "", "")
			c.Next()
			return
		}

		if claims, ok := m.parseBearerToken(c); ok {
			tier := claims.Tier
			if tier == "" {
				tier = commons.TierFree
			}
			m.applyIdentity(c, claims.UserID, claims.Email, tier, claims.SessionID, claims.ID)
		}

		c.Next()
	}
}

func (m *Middleware) parseBearerToken(c *gin.Context) (*entity.AccessTokenClaims, bool) {
	header := c.GetHeader("Authorization")
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, false
	}

	claims := &entity.AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(m.jwtSecret), nil
	}, jwt.WithIssuer(m.jwtIssuer))
	if err != nil || !token.Valid || claims.UserID == "" {
		return nil, false
	}

	return claims, true
}

func (m *Middleware) applyIdentity(c *gin.Context, userID, email, tier, sessionID, tokenID string) {
	c.Set(contextUserID, userID)
	c.Set(contextEmail, email)
	c.Set(contextTier, tier)
	c.Set(contextSessionID, sessionID)
	c.Set(contextTokenID, tokenID)

	ctx := utils.WithUserID(c.Request.Context(), userID)
	c.Request = c.Request.WithContext(ctx)
}

// GetUserID returns the authenticated user id (empty for anonymous requests).
func GetUserID(c *gin.Context) string {
	return c.GetString(contextUserID)
}

// GetSessionID returns the `sid` claim of the access token.
func GetSessionID(c *gin.Context) string {
	return c.GetString(contextSessionID)
}

// GetTokenID returns the `jti` claim of the access token.
func GetTokenID(c *gin.Context) string {
	return c.GetString(contextTokenID)
}

func GetEmail(c *gin.Context) string {
	return c.GetString(contextEmail)
}

func GetTier(c *gin.Context) string {
	return c.GetString(contextTier)
}
