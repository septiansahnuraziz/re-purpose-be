package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/repurpose-ai/re-purpose-be/commons"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/llm"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/sirupsen/logrus"
)

const minimumWordCount = 10

type ContentService struct {
	generationRepository entity.IGenerationRepository
	engine               llm.Client
	extractor            *Extractor
}

func NewContentService(repository entity.IGenerationRepository, engine llm.Client, extractor *Extractor) *ContentService {
	return &ContentService{
		generationRepository: repository,
		engine:               engine,
		extractor:            extractor,
	}
}

func (s *ContentService) ExtractURL(ctx context.Context, request entity.ExtractURLRequest) (*entity.ExtractURLResponse, error) {
	if !utils.IsHTTPURL(request.URL) {
		return nil, ErrInvalidURL
	}

	response, err := s.extractor.Extract(ctx, request.URL)
	if err != nil {
		return nil, err
	}

	logger.WithFields(ctx, logrus.Fields{
		"url":       request.URL,
		"wordCount": response.WordCount,
	}).Debug("content extracted")

	return response, nil
}

func (s *ContentService) Generate(ctx context.Context, userID string, request entity.GenerateContentRequest) (*entity.GenerateContentResponse, error) {
	return s.generate(ctx, userID, request, false)
}

func (s *ContentService) Reroll(ctx context.Context, userID string, request entity.GenerateContentRequest) (*entity.GenerateContentResponse, error) {
	return s.generate(ctx, userID, request, true)
}

func (s *ContentService) generate(ctx context.Context, userID string, request entity.GenerateContentRequest, reroll bool) (*entity.GenerateContentResponse, error) {
	request.RawText = strings.TrimSpace(request.RawText)
	request.Tone = strings.ToLower(strings.TrimSpace(request.Tone))
	request.Audience = utils.ExpectedString(request.Audience)

	if utils.WordCount(request.RawText) < minimumWordCount {
		return nil, ErrContentTooShort
	}
	if !commons.ValidTones[request.Tone] {
		return nil, ErrInvalidTone
	}
	if request.Audience == "" {
		return nil, ErrBadRequest
	}

	logger.WithFields(ctx, logrus.Fields{
		"userId":    userID,
		"tone":      request.Tone,
		"reroll":    reroll,
		"wordCount": utils.WordCount(request.RawText),
	}).Debug("generation started")

	outputs, metrics, err := s.engine.Generate(ctx, llm.Input{
		RawText:  request.RawText,
		Tone:     request.Tone,
		Audience: request.Audience,
		Reroll:   reroll,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLLMUnavailable, err)
	}

	generation := &entity.Generation{
		ID:                uuid.NewString(),
		UserID:            userID,
		SourceURL:         request.SourceURL,
		Tone:              request.Tone,
		Audience:          request.Audience,
		Status:            commons.GenerationStatusCompleted,
		RawText:           request.RawText,
		LatencyMS:         metrics.LatencyMS,
		Tokens:            metrics.Tokens,
		Confidence:        metrics.Confidence,
		SemanticDeviation: metrics.SemanticDeviation,
		Model:             s.engine.Model(),
	}
	if err := generation.SetResult(outputs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	if err := s.generationRepository.Create(ctx, generation); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	logger.WithFields(ctx, logrus.Fields{
		"userId":       userID,
		"generationId": generation.ID,
		"tone":         generation.Tone,
		"reroll":       reroll,
		"latencyMs":    metrics.LatencyMS,
		"tokens":       metrics.Tokens,
		"model":        generation.Model,
	}).Info("generation completed")

	return &entity.GenerateContentResponse{
		GenerationID: generation.ID,
		Data:         *outputs,
	}, nil
}

func (s *ContentService) GetGeneration(ctx context.Context, userID, generationID string) (*entity.GenerationResponse, error) {
	generation, err := s.generationRepository.FindByID(ctx, generationID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}
	if generation == nil || generation.UserID != userID {
		return nil, ErrGenerationNotFound
	}

	outputs, err := generation.ParseResult()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	response := generation.ToResponse(outputs)
	return &response, nil
}

func (s *ContentService) ListGenerations(ctx context.Context, userID string, page, size int) (*entity.GenerationListResponse, error) {
	page, size = normalizePagination(page, size)

	generations, total, err := s.generationRepository.FindByUser(ctx, userID, page, size)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInternalServerError, err)
	}

	items := make([]entity.GenerationResponse, 0, len(generations))
	for _, generation := range generations {
		outputs, parseErr := generation.ParseResult()
		if parseErr != nil {
			continue
		}
		items = append(items, generation.ToResponse(outputs))
	}

	return &entity.GenerationListResponse{
		Items: items,
		Meta:  entity.NewPaginationMeta(page, size, total),
	}, nil
}

func normalizePagination(page, size int) (int, int) {
	if page <= 0 {
		page = commons.DefaultPage
	}
	if size <= 0 {
		size = commons.DefaultPageSize
	}
	if size > commons.MaxPageSize {
		size = commons.MaxPageSize
	}
	return page, size
}
