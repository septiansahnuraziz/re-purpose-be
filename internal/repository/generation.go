package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"gorm.io/gorm"
)

type generationRepository struct {
	db *gorm.DB
}

func NewGenerationRepository(db *gorm.DB) entity.IGenerationRepository {
	return &generationRepository{db: db}
}

func (r *generationRepository) Create(ctx context.Context, generation *entity.Generation) error {
	if generation.ID == "" {
		generation.ID = uuid.NewString()
	}
	return r.db.WithContext(ctx).Create(generation).Error
}

func (r *generationRepository) Update(ctx context.Context, generation *entity.Generation) error {
	return r.db.WithContext(ctx).Save(generation).Error
}

func (r *generationRepository) FindByID(ctx context.Context, id string) (*entity.Generation, error) {
	var generation entity.Generation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&generation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &generation, nil
}

func (r *generationRepository) FindByUser(ctx context.Context, userID string, page, size int) ([]*entity.Generation, int64, error) {
	query := r.db.WithContext(ctx).Model(&entity.Generation{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var generations []*entity.Generation
	err := query.
		Order("created_at DESC").
		Offset((page - 1) * size).
		Limit(size).
		Find(&generations).Error
	if err != nil {
		return nil, 0, err
	}

	return generations, total, nil
}
