package config

import "time"

const (
	DefaultAppName         = "RePurpose.ai API"
	DefaultAppBuild        = "1"
	DefaultEnvironmentMode = "dev"
	DefaultHTTPPort        = "8080"
	DefaultLogLevel        = "info"
	DefaultLogFormat       = "text"

	DefaultDatabaseDriver    = "memory"
	DefaultDatabaseSSLMode   = "disable"
	DefaultDatabaseMaxIdle   = 10
	DefaultDatabaseMaxOpen   = 50
	DefaultDatabaseLifetime  = 15 * time.Minute
	DefaultDatabaseIdleTime  = 10 * time.Minute
	DefaultDatabasePingEvery = 5 * time.Second

	DefaultRedisDialTimeout  = 5 * time.Second
	DefaultRedisWriteTimeout = 2 * time.Second
	DefaultRedisReadTimeout  = 2 * time.Second
	DefaultRedisMaxIdleConn  = 50
	DefaultRedisMaxActiveCon = 100

	DefaultWorkerConcurrency = 10
	DefaultWorkerRetry       = 3
	DefaultWorkerRetryDelay  = time.Minute
	DefaultWorkerTimeout     = 5 * time.Minute
	DefaultWorkerNamespace   = "repurpose-be"

	DefaultDemoUserID         = "usr_demo"
	DefaultJWTIssuer          = "repurpose-ai"
	DefaultJWTSecret          = "change-me-in-production"
	DefaultJWTAccessTTL       = 15 * time.Minute
	DefaultJWTRefreshTTL      = 30 * 24 * time.Hour
	DefaultJWTRefreshTTLShort = 24 * time.Hour
	DefaultBcryptCost         = 12
	DefaultGoogleRedirectURL  = "http://localhost:8080/api/v1/auth/google/callback"
	DefaultAuthRateLimit      = 30
	DefaultAuthRateLimitWindow = time.Minute

	minBcryptCost      = 4
	maxBcryptCost      = 31
	minJWTSecretLength = 32

	DefaultLLMProvider = "mock"
	DefaultLLMBaseURL  = "https://api.openai.com/v1"
	DefaultLLMModel    = "gpt-4o-mini"
	DefaultLLMTimeout  = 120 * time.Second

	DefaultExtractorTimeout   = 15 * time.Second
	DefaultExtractorMaxLength = 200_000

	DefaultCacheTTL = 15 * time.Minute
)
