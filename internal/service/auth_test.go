package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/repurpose-ai/re-purpose-be/commons"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/repository"
	"github.com/repurpose-ai/re-purpose-be/internal/service"
	"golang.org/x/crypto/bcrypt"
)

type fakeGoogleVerifier struct {
	identity *entity.GoogleIdentity
	err      error
}

func (f *fakeGoogleVerifier) Verify(context.Context, string) (*entity.GoogleIdentity, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.identity, nil
}

func newTestAuthService(t *testing.T, verifier entity.IGoogleVerifier) (*service.AuthService, entity.IUserRepository) {
	t.Helper()

	_, users, _ := repository.NewMemoryRepositories()
	sessions := repository.NewMemorySessionRepository()

	auth := service.NewAuthService(users, sessions, verifier, entity.AuthConfig{
		JWTSecret:       "test-secret",
		JWTIssuer:       "test-issuer",
		AccessTTL:       time.Minute,
		RefreshTTL:      2 * time.Hour,
		RefreshTTLShort: time.Hour,
		BcryptCost:      bcrypt.MinCost,
	})
	return auth, users
}

func TestAuthRegisterLoginRefreshLogout(t *testing.T) {
	auth, users := newTestAuthService(t, nil)
	ctx := context.Background()
	meta := entity.RequestMeta{IP: "127.0.0.1", UserAgent: "go-test"}

	registerRequest := entity.RegisterRequest{
		FullName:    "Creator One",
		PrimaryRole: commons.PrimaryRoleCreator,
		Email:       "Creator@Example.com",
		Password:    "password123",
	}

	registered, err := auth.Register(ctx, registerRequest, meta)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if registered.AccessToken == "" || registered.RefreshToken == "" {
		t.Fatal("register should return both tokens")
	}
	if registered.TokenType != "Bearer" {
		t.Fatalf("token type = %q, want Bearer", registered.TokenType)
	}
	if registered.ExpiresIn != 60 {
		t.Fatalf("expiresIn = %d, want 60", registered.ExpiresIn)
	}
	if registered.User.Email != "creator@example.com" {
		t.Fatalf("email = %q, want normalized address", registered.User.Email)
	}
	if registered.User.Tier != commons.TierFree {
		t.Fatalf("tier = %q, want %q", registered.User.Tier, commons.TierFree)
	}
	assertAccessClaims(t, registered.AccessToken, "test-secret", registered.User.ID)

	stored, err := users.FindByEmail(ctx, "creator@example.com")
	if err != nil || stored == nil {
		t.Fatalf("find stored user: user=%+v err=%v", stored, err)
	}
	if !stored.HasPassword() || stored.PasswordHash == registerRequest.Password {
		t.Fatal("password should be stored as a bcrypt hash")
	}

	if _, err := auth.Register(ctx, registerRequest, meta); !errors.Is(err, service.ErrEmailAlreadyExists) {
		t.Fatalf("duplicate register error = %v, want ErrEmailAlreadyExists", err)
	}

	tooLong := registerRequest
	tooLong.Email = "long@example.com"
	tooLong.Password = strings.Repeat("a", 73)
	if _, err := auth.Register(ctx, tooLong, meta); !errors.Is(err, service.ErrPasswordTooLong) {
		t.Fatalf("long password error = %v, want ErrPasswordTooLong", err)
	}

	invalidRole := registerRequest
	invalidRole.Email = "role@example.com"
	invalidRole.PrimaryRole = "unknown"
	if _, err := auth.Register(ctx, invalidRole, meta); !errors.Is(err, service.ErrBadRequest) {
		t.Fatalf("invalid role error = %v, want ErrBadRequest", err)
	}

	if _, err := auth.Login(ctx, entity.LoginRequest{Email: "creator@example.com", Password: "wrong-password"}, meta); !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("wrong password error = %v, want ErrInvalidCredentials", err)
	}

	loggedIn, err := auth.Login(ctx, entity.LoginRequest{
		Email:    "CREATOR@example.com",
		Password: "password123",
		Remember: true,
	}, meta)
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	refreshed, err := auth.Refresh(ctx, entity.RefreshRequest{RefreshToken: loggedIn.RefreshToken}, meta)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.RefreshToken == loggedIn.RefreshToken {
		t.Fatal("refresh should rotate the refresh token")
	}
	assertAccessClaims(t, refreshed.AccessToken, "test-secret", loggedIn.User.ID)

	if _, err := auth.Refresh(ctx, entity.RefreshRequest{RefreshToken: loggedIn.RefreshToken}, meta); !errors.Is(err, service.ErrInvalidRefreshToken) {
		t.Fatalf("reused refresh error = %v, want ErrInvalidRefreshToken", err)
	}

	if err := auth.Logout(ctx, entity.LogoutRequest{RefreshToken: refreshed.RefreshToken}); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := auth.Refresh(ctx, entity.RefreshRequest{RefreshToken: refreshed.RefreshToken}, meta); !errors.Is(err, service.ErrInvalidRefreshToken) {
		t.Fatalf("refresh after logout error = %v, want ErrInvalidRefreshToken", err)
	}

	if err := auth.Logout(ctx, entity.LogoutRequest{}); err != nil {
		t.Fatalf("empty logout should be a no-op, got %v", err)
	}
}

func TestAuthGoogleCreatesAndLinksAccount(t *testing.T) {
	verifier := &fakeGoogleVerifier{
		identity: &entity.GoogleIdentity{
			Subject:       "google-1",
			Email:         "GUser@Example.com",
			Name:          "Google User",
			Picture:       "https://example.com/avatar.png",
			EmailVerified: true,
		},
	}
	auth, users := newTestAuthService(t, verifier)
	ctx := context.Background()
	meta := entity.RequestMeta{IP: "127.0.0.1", UserAgent: "go-test"}

	first, err := auth.GoogleAuth(ctx, entity.GoogleAuthRequest{Credential: "credential"}, meta)
	if err != nil {
		t.Fatalf("google auth: %v", err)
	}
	if first.User.Email != "guser@example.com" {
		t.Fatalf("email = %q, want normalized address", first.User.Email)
	}
	created, err := users.FindByGoogleID(ctx, "google-1")
	if err != nil || created == nil {
		t.Fatalf("find google user: user=%+v err=%v", created, err)
	}
	if created.AuthProvider != commons.AuthProviderGoogle {
		t.Fatalf("provider = %q, want %q", created.AuthProvider, commons.AuthProviderGoogle)
	}
	if !created.IsEmailVerified() {
		t.Fatal("google user with verified email should have email_verified_at")
	}

	second, err := auth.GoogleAuth(ctx, entity.GoogleAuthRequest{Credential: "credential"}, meta)
	if err != nil {
		t.Fatalf("second google auth: %v", err)
	}
	if second.User.ID != first.User.ID {
		t.Fatalf("second login user = %q, want %q", second.User.ID, first.User.ID)
	}

	if _, err := auth.Register(ctx, entity.RegisterRequest{
		FullName:    "Linked User",
		PrimaryRole: commons.PrimaryRoleFounder,
		Email:       "linked@example.com",
		Password:    "password123",
	}, meta); err != nil {
		t.Fatalf("register link target: %v", err)
	}

	verifier.identity = &entity.GoogleIdentity{
		Subject:       "google-2",
		Email:         "linked@example.com",
		Name:          "Linked User",
		EmailVerified: true,
	}
	linked, err := auth.GoogleAuth(ctx, entity.GoogleAuthRequest{Credential: "credential"}, meta)
	if err != nil {
		t.Fatalf("link google auth: %v", err)
	}
	linkedUser, err := users.FindByEmail(ctx, "linked@example.com")
	if err != nil || linkedUser == nil {
		t.Fatalf("find linked user: user=%+v err=%v", linkedUser, err)
	}
	if linkedUser.GoogleID != "google-2" {
		t.Fatalf("google id = %q, want google-2", linkedUser.GoogleID)
	}
	if linkedUser.AuthProvider != commons.AuthProviderBoth {
		t.Fatalf("provider = %q, want %q", linkedUser.AuthProvider, commons.AuthProviderBoth)
	}
	if linked.User.ID != linkedUser.ID {
		t.Fatalf("linked response user = %q, want %q", linked.User.ID, linkedUser.ID)
	}

	verifier.identity = &entity.GoogleIdentity{
		Subject:       "google-other",
		Email:         "linked@example.com",
		EmailVerified: true,
	}
	if _, err := auth.GoogleAuth(ctx, entity.GoogleAuthRequest{Credential: "credential"}, meta); !errors.Is(err, service.ErrInvalidGoogleCredential) {
		t.Fatalf("different google account error = %v, want ErrInvalidGoogleCredential", err)
	}

	if _, err := auth.Register(ctx, entity.RegisterRequest{
		FullName:    "Unverified User",
		PrimaryRole: commons.PrimaryRoleMarketer,
		Email:       "unverified@example.com",
		Password:    "password123",
	}, meta); err != nil {
		t.Fatalf("register unverified target: %v", err)
	}

	verifier.identity = &entity.GoogleIdentity{
		Subject:       "google-3",
		Email:         "unverified@example.com",
		EmailVerified: false,
	}
	if _, err := auth.GoogleAuth(ctx, entity.GoogleAuthRequest{Credential: "credential"}, meta); !errors.Is(err, service.ErrInvalidGoogleCredential) {
		t.Fatalf("unverified link error = %v, want ErrInvalidGoogleCredential", err)
	}
}

func TestGoogleVerifierConfiguration(t *testing.T) {
	verifier := service.NewGoogleVerifier("")
	if _, err := verifier.Verify(context.Background(), "credential"); !errors.Is(err, service.ErrGoogleNotConfigured) {
		t.Fatalf("unconfigured verifier error = %v, want ErrGoogleNotConfigured", err)
	}

	configured := service.NewGoogleVerifier("client-id")
	if _, err := configured.Verify(context.Background(), "  "); !errors.Is(err, service.ErrInvalidGoogleCredential) {
		t.Fatalf("empty credential error = %v, want ErrInvalidGoogleCredential", err)
	}
}

func assertAccessClaims(t *testing.T, token, secret, userID string) {
	t.Helper()

	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithIssuer("test-issuer"))
	if err != nil || !parsed.Valid {
		t.Fatalf("parse access token: valid=%v err=%v", parsed != nil && parsed.Valid, err)
	}

	if claims["uid"] != userID {
		t.Fatalf("uid claim = %v, want %q", claims["uid"], userID)
	}
	if claims["sid"] == "" || claims["sid"] == nil {
		t.Fatal("sid claim is missing")
	}
	if claims["jti"] == "" || claims["jti"] == nil {
		t.Fatal("jti claim is missing")
	}
}
