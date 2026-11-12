package entity

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TokenTypeBearer is the token type returned by the auth endpoints.
	TokenTypeBearer = "Bearer"

	// Bcrypt cost bounds accepted by AuthConfig.WithDefaults.
	bcryptMinCost = 4
	bcryptMaxCost = 31
)

// AuthConfig holds the tunables required by the auth service.
type AuthConfig struct {
	JWTSecret       string
	JWTIssuer       string
	AccessTTL       time.Duration
	RefreshTTL      time.Duration
	RefreshTTLShort time.Duration
	BcryptCost      int
}

// WithDefaults replaces zero/invalid values with safe defaults.
func (c AuthConfig) WithDefaults() AuthConfig {
	if c.JWTIssuer == "" {
		c.JWTIssuer = "repurpose-ai"
	}
	if c.AccessTTL <= 0 {
		c.AccessTTL = 15 * time.Minute
	}
	if c.RefreshTTL <= 0 {
		c.RefreshTTL = 30 * 24 * time.Hour
	}
	if c.RefreshTTLShort <= 0 {
		c.RefreshTTLShort = 24 * time.Hour
	}
	if c.BcryptCost < bcryptMinCost || c.BcryptCost > bcryptMaxCost {
		c.BcryptCost = 12
	}
	return c
}

// RefreshTTLFor returns the refresh-token lifetime for the given
// "remember me" flag.
func (c AuthConfig) RefreshTTLFor(remember bool) time.Duration {
	if remember {
		return c.RefreshTTL
	}
	return c.RefreshTTLShort
}

// AccessTokenClaims is the JWT payload used by access tokens and parsed by the
// auth middleware.
type AccessTokenClaims struct {
	UserID    string `json:"uid"`
	Email     string `json:"email"`
	Tier      string `json:"tier"`
	SessionID string `json:"sid"`
	jwt.RegisteredClaims
}

// AuthSession represents one refresh-token session (one login/device).
// Refresh tokens are stored as SHA-256 hashes, never in plain text.
type AuthSession struct {
	ID               string     `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`
	UserID           string     `gorm:"column:user_id;type:varchar(64);index" json:"userId"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash;type:varchar(64);uniqueIndex" json:"-"`
	UserAgent        string     `gorm:"column:user_agent;type:varchar(512)" json:"userAgent,omitempty"`
	IPAddress        string     `gorm:"column:ip_address;type:varchar(64)" json:"ipAddress,omitempty"`
	ExpiresAt        time.Time  `gorm:"column:expires_at" json:"expiresAt"`
	RevokedAt        *time.Time `gorm:"column:revoked_at" json:"revokedAt,omitempty"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (AuthSession) TableName() string { return "auth_sessions" }

// IsActive reports whether the session can still be used for a refresh.
func (s *AuthSession) IsActive(now time.Time) bool {
	return s != nil && s.RevokedAt == nil && s.ExpiresAt.After(now)
}

// AuthUser is the user payload returned by the auth endpoints.
type AuthUser struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Email          string `json:"email"`
	AvatarURL      string `json:"avatarUrl"`
	Tier           string `json:"tier"`
	Plan           string `json:"plan"`
	PrimaryRole    string `json:"primaryRole,omitempty"`
	DailyCredits   int    `json:"dailyCredits"`
	MonthlyCredits int    `json:"monthlyCredits"`
}

// RegisterRequest is the payload for POST /auth/register.
type RegisterRequest struct {
	FullName    string `json:"fullName" binding:"required,min=2,max=255"`
	PrimaryRole string `json:"primaryRole" binding:"required,oneof=creator founder marketer agency"`
	Email       string `json:"email" binding:"required,email,max=255"`
	Password    string `json:"password" binding:"required,min=8,max=72"`
}

// LoginRequest is the payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	Remember bool   `json:"remember"`
}

// RefreshRequest is the payload for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// LogoutRequest is the payload for POST /auth/logout. The refresh token is
// optional so logout stays idempotent even when the client already dropped it.
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// GoogleAuthRequest is the payload for POST /auth/google. Credential is the
// Google ID token produced by Google Identity Services on the client.
type GoogleAuthRequest struct {
	Credential string `json:"credential" binding:"required"`
	Remember   bool   `json:"remember"`
}

// GoogleIdentity is the verified identity extracted from a Google ID token.
type GoogleIdentity struct {
	Subject       string
	Email         string
	Name          string
	Picture       string
	EmailVerified bool
}

// AuthResponse is returned by register, login, refresh, and Google auth.
type AuthResponse struct {
	AccessToken  string    `json:"accessToken"`
	TokenType    string    `json:"tokenType"`
	ExpiresIn    int64     `json:"expiresIn"`
	RefreshToken string    `json:"refreshToken"`
	User         *AuthUser `json:"user"`
}
