package service

import "errors"

var (
	ErrNotFound            = errors.New("error not found")
	ErrBadRequest          = errors.New("error bad request")
	ErrInternalServerError = errors.New("error internal server")

	ErrContentTooShort     = errors.New("content is too short to repurpose")
	ErrInvalidURL          = errors.New("invalid or unsupported url")
	ErrInvalidTone         = errors.New("invalid tone")
	ErrExtractionFailed    = errors.New("failed to extract content from url")
	ErrLLMUnavailable      = errors.New("generation engine is unavailable")
	ErrGenerationNotFound  = errors.New("generation is not found")
	ErrUserNotFound        = errors.New("user is not found")
	ErrSponsorEventInvalid = errors.New("invalid sponsor event")

	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrInvalidRefreshToken     = errors.New("invalid or expired refresh token")
	ErrInvalidGoogleCredential = errors.New("invalid google credential")
	ErrEmailAlreadyExists      = errors.New("email is already registered")
	ErrPasswordTooLong         = errors.New("password exceeds 72 bytes")
	ErrGoogleNotConfigured     = errors.New("google sign-in is not configured")
	ErrGoogleUnavailable       = errors.New("google verification is unavailable")
)
