package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"google.golang.org/api/idtoken"
)

// googleVerifier validates Google ID tokens against the configured client ID.
// The contract lives in entity.IGoogleVerifier so the auth service can be
// tested without network access to Google's JWKS endpoint.
type googleVerifier struct {
	clientID string
}

// NewGoogleVerifier builds a verifier for tokens issued to the given Google
// Identity Services client ID. An empty client ID keeps the verifier
// unconfigured, and Verify returns ErrGoogleNotConfigured.
func NewGoogleVerifier(clientID string) entity.IGoogleVerifier {
	return &googleVerifier{clientID: strings.TrimSpace(clientID)}
}

func (v *googleVerifier) Verify(ctx context.Context, credential string) (*entity.GoogleIdentity, error) {
	if v.clientID == "" {
		return nil, ErrGoogleNotConfigured
	}

	credential = strings.TrimSpace(credential)
	if credential == "" {
		return nil, ErrInvalidGoogleCredential
	}

	payload, err := idtoken.Validate(ctx, credential, v.clientID)
	if err != nil {
		if isGoogleUnavailable(err) {
			return nil, fmt.Errorf("%w: %v", ErrGoogleUnavailable, err)
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidGoogleCredential, err)
	}

	identity := &entity.GoogleIdentity{
		Subject:       strings.TrimSpace(payload.Subject),
		Email:         claimString(payload.Claims, "email"),
		Name:          claimString(payload.Claims, "name"),
		Picture:       claimString(payload.Claims, "picture"),
		EmailVerified: claimBool(payload.Claims, "email_verified"),
	}
	if identity.Subject == "" || identity.Email == "" {
		return nil, ErrInvalidGoogleCredential
	}

	return identity, nil
}

// isGoogleUnavailable reports whether the validation failure was caused by a
// transport problem (JWKS fetch) rather than an invalid token.
func isGoogleUnavailable(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	return errors.As(err, &urlErr)
}

func claimString(claims map[string]any, key string) string {
	value, ok := claims[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func claimBool(claims map[string]any, key string) bool {
	switch value := claims[key].(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	default:
		return false
	}
}
