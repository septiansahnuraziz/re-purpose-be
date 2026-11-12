package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/commons"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxPasswordBytes  = 72
	refreshTokenBytes = 32
	maxUserAgentRunes = 500
)

// AuthService implements register, login, refresh rotation, logout, and
// Google sign-in.
type AuthService struct {
	users    entity.IUserRepository
	sessions entity.ISessionRepository
	verifier entity.IGoogleVerifier
	cfg      entity.AuthConfig
}

func NewAuthService(users entity.IUserRepository, sessions entity.ISessionRepository, verifier entity.IGoogleVerifier, cfg entity.AuthConfig) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		verifier: verifier,
		cfg:      cfg.WithDefaults(),
	}
}

func (s *AuthService) Register(ctx context.Context, request entity.RegisterRequest, meta entity.RequestMeta) (*entity.AuthResponse, error) {
	fullName := strings.TrimSpace(request.FullName)
	email := normalizeEmail(request.Email)
	role := strings.ToLower(strings.TrimSpace(request.PrimaryRole))

	if fullName == "" || email == "" || !commons.ValidPrimaryRoles[role] {
		return nil, fmt.Errorf("%w: invalid registration payload", ErrBadRequest)
	}
	if len(request.Password) > maxPasswordBytes {
		return nil, ErrPasswordTooLong
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), s.cfg.BcryptCost)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	now := time.Now()
	user := &entity.User{
		ID:             newAuthID("usr_"),
		Email:          email,
		PasswordHash:   string(passwordHash),
		AuthProvider:   commons.AuthProviderPassword,
		PrimaryRole:    role,
		Name:           fullName,
		Tier:           commons.TierFree,
		DailyCredits:   commons.DefaultDailyCredits,
		MonthlyCredits: commons.DefaultMonthlyCredits,
		LastLoginAt:    &now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, entity.ErrDuplicateEmail) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	// A fresh registration is treated as a persistent session.
	return s.issueTokens(ctx, user, true, meta)
}

func (s *AuthService) Login(ctx context.Context, request entity.LoginRequest, meta entity.RequestMeta) (*entity.AuthResponse, error) {
	user, err := s.users.FindByEmail(ctx, normalizeEmail(request.Email))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	if user == nil || !user.HasPassword() {
		return nil, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.users.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	return s.issueTokens(ctx, user, request.Remember, meta)
}

// Refresh rotates the refresh token: the presented token is revoked and a new
// session is created with the same absolute expiry, so "remember me" duration
// is preserved without storing the flag.
func (s *AuthService) Refresh(ctx context.Context, request entity.RefreshRequest, meta entity.RequestMeta) (*entity.AuthResponse, error) {
	refreshToken := strings.TrimSpace(request.RefreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	session, err := s.sessions.FindByRefreshTokenHash(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	now := time.Now()
	if session == nil || !session.IsActive(now) {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.users.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	if user == nil {
		_ = s.sessions.Revoke(ctx, session.ID, now)
		return nil, ErrInvalidRefreshToken
	}

	if err := s.sessions.Revoke(ctx, session.ID, now); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	newSession := &entity.AuthSession{
		ID:               newAuthID("ses_"),
		UserID:           user.ID,
		RefreshTokenHash: hashRefreshToken(newRefreshToken),
		UserAgent:        utils.TruncateRunes(firstNonEmpty(meta.UserAgent, session.UserAgent), maxUserAgentRunes),
		IPAddress:        utils.TruncateRunes(firstNonEmpty(meta.IP, session.IPAddress), 64),
		ExpiresAt:        session.ExpiresAt,
	}
	if err := s.sessions.Create(ctx, newSession); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	accessToken, expiresIn, err := s.signAccessToken(user, newSession.ID, now)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	return &entity.AuthResponse{
		AccessToken:  accessToken,
		TokenType:    entity.TokenTypeBearer,
		ExpiresIn:    expiresIn,
		RefreshToken: newRefreshToken,
		User:         user.ToAuthUser(),
	}, nil
}

// Logout revokes the session tied to the refresh token. It is idempotent and
// never revokes the access token, which stays valid until it expires.
func (s *AuthService) Logout(ctx context.Context, request entity.LogoutRequest) error {
	refreshToken := strings.TrimSpace(request.RefreshToken)
	if refreshToken == "" {
		return nil
	}

	session, err := s.sessions.FindByRefreshTokenHash(ctx, hashRefreshToken(refreshToken))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	if session == nil || session.RevokedAt != nil {
		return nil
	}

	if err := s.sessions.Revoke(ctx, session.ID, time.Now()); err != nil {
		return fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	return nil
}

func (s *AuthService) GoogleAuth(ctx context.Context, request entity.GoogleAuthRequest, meta entity.RequestMeta) (*entity.AuthResponse, error) {
	if s.verifier == nil {
		return nil, ErrGoogleNotConfigured
	}

	identity, err := s.verifier.Verify(ctx, request.Credential)
	if err != nil {
		return nil, err
	}

	user, err := s.resolveGoogleUser(ctx, identity, time.Now())
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, user, request.Remember, meta)
}

// resolveGoogleUser finds the matching account, links Google to an existing
// password account only when Google vouches for the email, or creates a new
// account.
func (s *AuthService) resolveGoogleUser(ctx context.Context, identity *entity.GoogleIdentity, now time.Time) (*entity.User, error) {
	subject := strings.TrimSpace(identity.Subject)
	email := normalizeEmail(identity.Email)
	if subject == "" || email == "" {
		return nil, ErrInvalidGoogleCredential
	}

	user, err := s.users.FindByGoogleID(ctx, subject)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	if user == nil {
		user, err = s.users.FindByEmail(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
		}
	}

	if user == nil {
		return s.createGoogleUser(ctx, identity, email, subject, now)
	}

	if user.GoogleID != "" && user.GoogleID != subject {
		// The email already belongs to a different Google account.
		return nil, ErrInvalidGoogleCredential
	}

	if user.GoogleID == "" {
		if !identity.EmailVerified {
			return nil, ErrInvalidGoogleCredential
		}
		user.GoogleID = subject
		if user.HasPassword() {
			user.AuthProvider = commons.AuthProviderBoth
		} else {
			user.AuthProvider = commons.AuthProviderGoogle
		}
		if !user.IsEmailVerified() {
			user.EmailVerifiedAt = &now
		}
	}

	if user.Name == "" && identity.Name != "" {
		user.Name = identity.Name
	}
	if user.AvatarURL == "" && identity.Picture != "" {
		user.AvatarURL = identity.Picture
	}
	user.LastLoginAt = &now

	if err := s.users.Update(ctx, user); err != nil {
		if errors.Is(err, entity.ErrDuplicateGoogleID) {
			return nil, ErrInvalidGoogleCredential
		}
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	return user, nil
}

func (s *AuthService) createGoogleUser(ctx context.Context, identity *entity.GoogleIdentity, email, subject string, now time.Time) (*entity.User, error) {
	name := strings.TrimSpace(identity.Name)
	if name == "" {
		name = email
		if at := strings.IndexByte(name, '@'); at > 0 {
			name = name[:at]
		}
	}

	user := &entity.User{
		ID:             newAuthID("usr_"),
		Email:          email,
		GoogleID:       subject,
		AuthProvider:   commons.AuthProviderGoogle,
		Name:           name,
		AvatarURL:      strings.TrimSpace(identity.Picture),
		Tier:           commons.TierFree,
		DailyCredits:   commons.DefaultDailyCredits,
		MonthlyCredits: commons.DefaultMonthlyCredits,
		LastLoginAt:    &now,
	}
	if identity.EmailVerified {
		user.EmailVerifiedAt = &now
	}

	if err := s.users.Create(ctx, user); err != nil {
		if !errors.Is(err, entity.ErrDuplicateEmail) && !errors.Is(err, entity.ErrDuplicateGoogleID) {
			return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
		}

		// Another request may have created the account concurrently; reuse it.
		existing, findErr := s.users.FindByGoogleID(ctx, subject)
		if findErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalServerError, findErr)
		}
		if existing == nil {
			existing, findErr = s.users.FindByEmail(ctx, email)
			if findErr != nil {
				return nil, fmt.Errorf("%w: %v", ErrInternalServerError, findErr)
			}
		}
		if existing == nil {
			return nil, ErrEmailAlreadyExists
		}

		existing.LastLoginAt = &now
		if updateErr := s.users.Update(ctx, existing); updateErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrInternalServerError, updateErr)
		}
		return existing, nil
	}

	return user, nil
}

func (s *AuthService) issueTokens(ctx context.Context, user *entity.User, remember bool, meta entity.RequestMeta) (*entity.AuthResponse, error) {
	now := time.Now()
	return s.issueSessionTokens(ctx, user, now.Add(s.cfg.RefreshTTLFor(remember)), meta)
}

func (s *AuthService) issueSessionTokens(ctx context.Context, user *entity.User, expiresAt time.Time, meta entity.RequestMeta) (*entity.AuthResponse, error) {
	now := time.Now()
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	session := &entity.AuthSession{
		ID:               newAuthID("ses_"),
		UserID:           user.ID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		UserAgent:        utils.TruncateRunes(meta.UserAgent, maxUserAgentRunes),
		IPAddress:        utils.TruncateRunes(meta.IP, 64),
		ExpiresAt:        expiresAt,
	}
	if err := s.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	accessToken, expiresIn, err := s.signAccessToken(user, session.ID, now)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	return &entity.AuthResponse{
		AccessToken:  accessToken,
		TokenType:    entity.TokenTypeBearer,
		ExpiresIn:    expiresIn,
		RefreshToken: refreshToken,
		User:         user.ToAuthUser(),
	}, nil
}

func (s *AuthService) signAccessToken(user *entity.User, sessionID string, now time.Time) (string, int64, error) {
	expiresAt := now.Add(s.cfg.AccessTTL)
	claims := entity.AccessTokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Tier:      user.Tier,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        newAuthID(""),
			Subject:   user.ID,
			Issuer:    s.cfg.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", 0, err
	}
	return signed, int64(s.cfg.AccessTTL.Seconds()), nil
}

func generateRefreshToken() (string, error) {
	buffer := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func newAuthID(prefix string) string {
	return prefix + strings.ReplaceAll(uuid.NewString(), "-", "")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
