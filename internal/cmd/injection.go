package cmd

import (
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/llm"
	"github.com/repurpose-ai/re-purpose-be/internal/repository"
	"github.com/repurpose-ai/re-purpose-be/internal/service"
	"gorm.io/gorm"
)

// Container holds the wired application services.
type Container struct {
	ContentService    entity.IContentService
	SystemService     entity.ISystemService
	UserService       entity.IUserService
	AuthService       entity.IAuthService
	SponsorService    entity.ISponsorService
	SessionRepository entity.ISessionRepository
	EngineName        string
}

// InitContainer wires repositories -> services based on the available infrastructure.
func InitContainer(db *gorm.DB, redisClient *redis.Client, startedAt time.Time) *Container {
	var (
		generationRepository entity.IGenerationRepository
		userRepository       entity.IUserRepository
		sessionRepository    entity.ISessionRepository
		sponsorRepository    entity.ISponsorClickRepository
	)

	if db != nil {
		generationRepository = repository.NewGenerationRepository(db)
		userRepository = repository.NewUserRepository(db)
		sessionRepository = repository.NewSessionRepository(db)
		sponsorRepository = repository.NewSponsorClickRepository(db)
	} else {
		generationRepository, userRepository, sponsorRepository = repository.NewMemoryRepositories()
		sessionRepository = repository.NewMemorySessionRepository()
	}

	engine := llm.NewFromConfig()
	extractor := service.NewExtractor(config.ExtractorTimeout(), config.ExtractorMaxContentLength())

	authService := service.NewAuthService(
		userRepository,
		sessionRepository,
		service.NewGoogleVerifier(config.GoogleClientID()),
		entity.AuthConfig{
			JWTSecret:       config.JWTSecret(),
			JWTIssuer:       config.JWTIssuer(),
			AccessTTL:       config.JWTAccessTTL(),
			RefreshTTL:      config.JWTRefreshTTL(),
			RefreshTTLShort: config.JWTRefreshTTLShort(),
			BcryptCost:      config.AuthBcryptCost(),
		},
	)

	return &Container{
		ContentService:    service.NewContentService(generationRepository, engine, extractor),
		SystemService:     service.NewSystemService(db, redisClient, engine, startedAt),
		UserService:       service.NewUserService(userRepository),
		AuthService:       authService,
		SponsorService:    service.NewSponsorService(sponsorRepository),
		SessionRepository: sessionRepository,
		EngineName:        engine.Name(),
	}
}
