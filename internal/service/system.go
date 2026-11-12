package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/llm"
	"gorm.io/gorm"
)

type SystemService struct {
	db        *gorm.DB
	redis     *redis.Client
	engine    llm.Client
	startedAt time.Time
}

func NewSystemService(db *gorm.DB, redisClient *redis.Client, engine llm.Client, startedAt time.Time) *SystemService {
	return &SystemService{
		db:        db,
		redis:     redisClient,
		engine:    engine,
		startedAt: startedAt,
	}
}

func (s *SystemService) Status(ctx context.Context) *entity.SystemStatusResponse {
	dependencies := []entity.DependencyStatus{
		s.checkDatabase(ctx),
		s.checkRedis(ctx),
	}

	overall := entity.SystemStatusOperational
	for _, dependency := range dependencies {
		if dependency.Status == entity.SystemStatusDown {
			overall = entity.SystemStatusDegraded
			break
		}
	}

	return &entity.SystemStatusResponse{
		Status:        overall,
		Version:       config.AppVersion(),
		Environment:   config.EnvironmentMode(),
		UptimeSeconds: int64(time.Since(s.startedAt).Seconds()),
		Engine: entity.EngineStatus{
			Provider: s.engine.Name(),
			Model:    s.engine.Model(),
			Status:   entity.SystemStatusOperational,
		},
		Dependencies: dependencies,
		CheckedAt:    time.Now(),
	}
}

func (s *SystemService) checkDatabase(ctx context.Context) entity.DependencyStatus {
	status := entity.DependencyStatus{Name: "database", Status: entity.SystemStatusDisabled, Detail: "not configured"}

	if s.db == nil {
		return status
	}

	startedAt := time.Now()
	sqlDB, err := s.db.DB()
	if err != nil {
		status.Status = entity.SystemStatusDown
		status.Detail = err.Error()
		return status
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(pingCtx); err != nil {
		status.Status = entity.SystemStatusDown
		status.Detail = err.Error()
		status.LatencyMS = time.Since(startedAt).Milliseconds()
		return status
	}

	status.Status = entity.SystemStatusOperational
	status.LatencyMS = time.Since(startedAt).Milliseconds()
	return status
}

func (s *SystemService) checkRedis(ctx context.Context) entity.DependencyStatus {
	status := entity.DependencyStatus{Name: "redis", Status: entity.SystemStatusDisabled, Detail: "caching disabled"}

	if s.redis == nil {
		return status
	}

	startedAt := time.Now()
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.redis.Ping(pingCtx).Err(); err != nil {
		status.Status = entity.SystemStatusDown
		status.Detail = err.Error()
		status.LatencyMS = time.Since(startedAt).Milliseconds()
		return status
	}

	status.Status = entity.SystemStatusOperational
	status.LatencyMS = time.Since(startedAt).Milliseconds()
	return status
}
