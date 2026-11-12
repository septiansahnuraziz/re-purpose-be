package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) entity.IUserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	if strings.TrimSpace(googleID) == "" {
		return nil, nil
	}

	var user entity.User
	err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return mapUserWriteError(err)
	}
	return nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	result := r.db.WithContext(ctx).
		Model(&entity.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{
			"email":             user.Email,
			"password_hash":     user.PasswordHash,
			"google_id":         user.GoogleID,
			"auth_provider":     user.AuthProvider,
			"primary_role":      user.PrimaryRole,
			"name":              user.Name,
			"avatar_url":        user.AvatarURL,
			"tier":              user.Tier,
			"daily_credits":     user.DailyCredits,
			"monthly_credits":   user.MonthlyCredits,
			"email_verified_at": user.EmailVerifiedAt,
			"last_login_at":     user.LastLoginAt,
			"updated_at":        time.Now(),
		})
	if result.Error != nil {
		return mapUserWriteError(result.Error)
	}
	if result.RowsAffected == 0 {
		return entity.ErrNotFound
	}
	return nil
}
