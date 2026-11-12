package entity

import (
	"context"
	"time"
)

// IContentService handles the repurposing pipeline.
type IContentService interface {
	ExtractURL(ctx context.Context, request ExtractURLRequest) (*ExtractURLResponse, error)
	Generate(ctx context.Context, userID string, request GenerateContentRequest) (*GenerateContentResponse, error)
	Reroll(ctx context.Context, userID string, request GenerateContentRequest) (*GenerateContentResponse, error)
	GetGeneration(ctx context.Context, userID, generationID string) (*GenerationResponse, error)
	ListGenerations(ctx context.Context, userID string, page, size int) (*GenerationListResponse, error)
}

type ISystemService interface {
	Status(ctx context.Context) *SystemStatusResponse
}

type IUserService interface {
	GetProfile(ctx context.Context, userID string) (*UserProfile, error)
}

type IAuthService interface {
	Register(ctx context.Context, request RegisterRequest, meta RequestMeta) (*AuthResponse, error)
	Login(ctx context.Context, request LoginRequest, meta RequestMeta) (*AuthResponse, error)
	Refresh(ctx context.Context, request RefreshRequest, meta RequestMeta) (*AuthResponse, error)
	Logout(ctx context.Context, request LogoutRequest) error
	GoogleAuth(ctx context.Context, request GoogleAuthRequest, meta RequestMeta) (*AuthResponse, error)
}

// IGoogleVerifier validates a Google ID token and returns the identity it
// asserts.
type IGoogleVerifier interface {
	Verify(ctx context.Context, credential string) (*GoogleIdentity, error)
}

type ISponsorService interface {
	TrackClick(ctx context.Context, userID string, request SponsorClickRequest, meta RequestMeta) (*SponsorClickResponse, error)
}

// Repositories

type IGenerationRepository interface {
	Create(ctx context.Context, generation *Generation) error
	Update(ctx context.Context, generation *Generation) error
	FindByID(ctx context.Context, id string) (*Generation, error)
	FindByUser(ctx context.Context, userID string, page, size int) ([]*Generation, int64, error)
}

type IUserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByGoogleID(ctx context.Context, googleID string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}

type ISessionRepository interface {
	Create(ctx context.Context, session *AuthSession) error
	FindByID(ctx context.Context, id string) (*AuthSession, error)
	FindByRefreshTokenHash(ctx context.Context, hash string) (*AuthSession, error)
	Revoke(ctx context.Context, id string, revokedAt time.Time) error
	RevokeByUserID(ctx context.Context, userID string, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type ISponsorClickRepository interface {
	Create(ctx context.Context, click *SponsorClick) error
}
