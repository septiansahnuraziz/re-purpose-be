package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"gorm.io/gorm"
)

type sponsorClickRepository struct {
	db *gorm.DB
}

func NewSponsorClickRepository(db *gorm.DB) entity.ISponsorClickRepository {
	return &sponsorClickRepository{db: db}
}

func (r *sponsorClickRepository) Create(ctx context.Context, click *entity.SponsorClick) error {
	if click.ID == "" {
		click.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(click).Error
}
