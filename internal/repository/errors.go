package repository

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
)

const pgUniqueViolation = "23505"

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation
}

// mapUserWriteError converts PostgreSQL unique-violation errors into the
// sentinel errors used by the auth service.
func mapUserWriteError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgUniqueViolation {
		return err
	}

	constraint := strings.ToLower(pgErr.ConstraintName)
	switch {
	case strings.Contains(constraint, "uq_users_email"):
		return entity.ErrDuplicateEmail
	case strings.Contains(constraint, "uq_users_google_id"):
		return entity.ErrDuplicateGoogleID
	default:
		return err
	}
}
