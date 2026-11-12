package entity

import "errors"

// Persistence sentinel errors shared by repository implementations and the
// services that consume them.
var (
	ErrNotFound          = errors.New("record not found")
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateGoogleID = errors.New("duplicate google id")
	ErrDuplicateRefresh  = errors.New("duplicate refresh token")
)
