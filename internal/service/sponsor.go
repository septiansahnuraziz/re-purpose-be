package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/commons"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/utils"
)

type SponsorService struct {
	repository entity.ISponsorClickRepository
}

func NewSponsorService(repository entity.ISponsorClickRepository) *SponsorService {
	return &SponsorService{repository: repository}
}

func (s *SponsorService) TrackClick(ctx context.Context, userID string, request entity.SponsorClickRequest, meta entity.RequestMeta) (*entity.SponsorClickResponse, error) {
	request.SponsorID = utils.ExpectedString(request.SponsorID)
	request.Event = utils.ExpectedString(request.Event)

	if request.SponsorID == "" || !commons.ValidSponsorEvents[request.Event] {
		return nil, ErrSponsorEventInvalid
	}

	occurredAt := time.Now()
	if request.Timestamp != nil && !request.Timestamp.IsZero() {
		occurredAt = *request.Timestamp
	}

	click := &entity.SponsorClick{
		ID:         uuid.NewString(),
		SponsorID:  request.SponsorID,
		Event:      request.Event,
		UserID:     userID,
		IP:         meta.IP,
		UserAgent:  utils.TruncateRunes(meta.UserAgent, 500),
		OccurredAt: occurredAt,
	}

	if err := s.repository.Create(ctx, click); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	return &entity.SponsorClickResponse{
		ID:        click.ID,
		SponsorID: click.SponsorID,
		Event:     click.Event,
		TrackedAt: occurredAt,
	}, nil
}
