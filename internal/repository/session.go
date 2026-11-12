package repository

import (
	"context"
	"errors"
	"time"

	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"gorm.io/gorm"
)

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) entity.ISessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *entity.AuthSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		if isDuplicateKey(err) {
			return entity.ErrDuplicateRefresh
		}
		return err
	}
	return nil
}

func (r *sessionRepository) FindByID(ctx context.Context, id string) (*entity.AuthSession, error) {
	var session entity.AuthSession
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) FindByRefreshTokenHash(ctx context.Context, hash string) (*entity.AuthSession, error) {
	var session entity.AuthSession
	err := r.db.WithContext(ctx).Where("refresh_token_hash = ?", hash).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Revoke is idempotent: revoking an unknown or already-revoked session succeeds.
func (r *sessionRepository) Revoke(ctx context.Context, id string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.AuthSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", revokedAt).
		Error
}

func (r *sessionRepository) RevokeByUserID(ctx context.Context, userID string, revokedAt time.Time) error {
	return r.db.WithContext(ctx).
		Model(&entity.AuthSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", revokedAt).
		Error
}

func (r *sessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", before).
		Delete(&entity.AuthSession{}).
		Error
}
