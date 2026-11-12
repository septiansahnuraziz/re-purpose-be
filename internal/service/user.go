package service

import (
	"context"
	"fmt"

	"github.com/repurpose-ai/re-purpose-be/internal/entity"
)

type UserService struct {
	repository entity.IUserRepository
}

func NewUserService(repository entity.IUserRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*entity.UserProfile, error) {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user.ToProfile(), nil
}
