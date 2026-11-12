package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// LoadConfig reads config.yaml (or CONFIG_FILE) and enables env var overrides.
// Env keys use dots replaced by underscores, e.g. db.host -> DB_HOST.
func LoadConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if file := os.Getenv("CONFIG_FILE"); file != "" {
		viper.SetConfigFile(file)
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("app.name", DefaultAppName)
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("build", DefaultAppBuild)
	viper.SetDefault("mode", DefaultEnvironmentMode)
	viper.SetDefault("port", DefaultHTTPPort)
	viper.SetDefault("log_level", DefaultLogLevel)
	viper.SetDefault("log_format", DefaultLogFormat)
	viper.SetDefault("db.driver", DefaultDatabaseDriver)
	viper.SetDefault("auth.demo_user_id", DefaultDemoUserID)
	viper.SetDefault("auth.jwt.iss", DefaultJWTIssuer)
	viper.SetDefault("auth.jwt.secret", DefaultJWTSecret)
	viper.SetDefault("auth.jwt.access_ttl", DefaultJWTAccessTTL.String())
	viper.SetDefault("auth.refresh.ttl", DefaultJWTRefreshTTL.String())
	viper.SetDefault("auth.refresh.ttl_short", DefaultJWTRefreshTTLShort.String())
	viper.SetDefault("auth.bcrypt_cost", DefaultBcryptCost)
	viper.SetDefault("auth.rate_limit.requests", DefaultAuthRateLimit)
	viper.SetDefault("auth.rate_limit.window", DefaultAuthRateLimitWindow.String())
	viper.SetDefault("auth.google.redirect_url", DefaultGoogleRedirectURL)
	viper.SetDefault("llm.provider", DefaultLLMProvider)
	viper.SetDefault("llm.model", DefaultLLMModel)

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("failed to read config: %v", err)
	}
}

func AppName() string { return viper.GetString("app.name") }

func AppSlugName() string { return viper.GetString("app.slug_name") }

func AppVersion() string { return viper.GetString("app.version") }

func AppBuild() string { return viper.GetString("build") }

func EnvironmentMode() string { return viper.GetString("mode") }

func IsProduction() bool {
	return strings.EqualFold(EnvironmentMode(), "prod") || strings.EqualFold(EnvironmentMode(), "production")
}

func HttpPort() string { return fmt.Sprintf("%s", viper.GetString("port")) }

func LogLevel() string { return viper.GetString("log_level") }

func LogFormat() string {
	return utils.ValueOrDefault(viper.GetString("log_format"), DefaultLogFormat)
}

func CORSAllowedOrigins() []string {
	origins := viper.GetStringSlice("cors.allowed_origins")
	if len(origins) == 0 {
		return []string{"http://localhost:3000"}
	}
	return origins
}

// Database

func DatabaseDriver() string {
	return utils.ValueOrDefault(viper.GetString("db.driver"), DefaultDatabaseDriver)
}

func DatabaseTimeZone() string { return viper.GetString("db.timezone") }

func DatabaseUsername() string { return viper.GetString("db.user") }

func DatabasePassword() string { return viper.GetString("db.password") }

func DatabaseHost() string { return viper.GetString("db.host") }

func DatabasePort() string { return viper.GetString("db.port") }

func DatabaseName() string { return viper.GetString("db.name") }

func DatabaseSSLMode() string {
	return utils.ValueOrDefault(viper.GetString("db.ssl_mode"), DefaultDatabaseSSLMode)
}

func DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		DatabaseHost(),
		DatabasePort(),
		DatabaseUsername(),
		DatabasePassword(),
		DatabaseName(),
		DatabaseSSLMode(),
		urlQueryEscape(DatabaseTimeZone()),
	)
}

func DatabaseMaxIdleConns() int {
	return utils.ValueOrDefault(viper.GetInt("db.max_idle_conns"), DefaultDatabaseMaxIdle)
}

func DatabaseMaxOpenConns() int {
	return utils.ValueOrDefault(viper.GetInt("db.max_open_conns"), DefaultDatabaseMaxOpen)
}

func DatabaseConnMaxLifetime() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("db.conn_max_lifetime"), DefaultDatabaseLifetime)
}

func DatabaseConnMaxIdleTime() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("db.conn_max_idletime"), DefaultDatabaseIdleTime)
}

func DatabaseRetryAttempts() int {
	return utils.ValueOrDefault(viper.GetInt("db.retry_attempts"), 3)
}

func DatabasePingInterval() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("db.ping_interval"), DefaultDatabasePingEvery)
}

// Cache & Redis

func EnableCaching() bool { return viper.GetBool("enable_caching") }

func CacheTTL() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("cache_ttl"), DefaultCacheTTL)
}

func RedisCacheHost() string { return viper.GetString("redis.cache_host") }

func RedisWorkerHost() string { return viper.GetString("redis.worker_host") }

// Worker

func WorkerNamespace() string {
	return utils.ValueOrDefault(viper.GetString("worker.namespace"), DefaultWorkerNamespace)
}

func WorkerConcurrency() int {
	return utils.ValueOrDefault(viper.GetInt("worker.concurrency"), DefaultWorkerConcurrency)
}

func WorkerRetryAttempts() int {
	return utils.ValueOrDefault(viper.GetInt("worker.retry_attempts"), DefaultWorkerRetry)
}

func WorkerRetryDelay() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("worker.retry_delay"), DefaultWorkerRetryDelay)
}

func WorkerTimeout() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("worker.timeout"), DefaultWorkerTimeout)
}

// Auth

func AuthBypass() bool { return viper.GetBool("auth.bypass") }

func DemoUserID() string {
	return utils.ValueOrDefault(viper.GetString("auth.demo_user_id"), DefaultDemoUserID)
}

func JWTSecret() string {
	return utils.ValueOrDefault(viper.GetString("auth.jwt.secret"), DefaultJWTSecret)
}

func JWTIssuer() string {
	return utils.ValueOrDefault(viper.GetString("auth.jwt.iss"), DefaultJWTIssuer)
}

// JWTAccessTTL is the lifetime of access tokens. It falls back to the legacy
// auth.jwt.exp key so existing deployments keep working.
func JWTAccessTTL() time.Duration {
	if raw := strings.TrimSpace(viper.GetString("auth.jwt.access_ttl")); raw != "" {
		return utils.ParseDurationWithDefault(raw, DefaultJWTAccessTTL)
	}
	return utils.ParseDurationWithDefault(viper.GetString("auth.jwt.exp"), DefaultJWTAccessTTL)
}

// JWTExpiry is kept for backward compatibility. Prefer JWTAccessTTL.
func JWTExpiry() time.Duration { return JWTAccessTTL() }

func JWTRefreshTTL() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("auth.refresh.ttl"), DefaultJWTRefreshTTL)
}

func JWTRefreshTTLShort() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("auth.refresh.ttl_short"), DefaultJWTRefreshTTLShort)
}

// RefreshTTL returns the refresh-token lifetime for the given "remember me" flag.
func RefreshTTL(remember bool) time.Duration {
	if remember {
		return JWTRefreshTTL()
	}
	return JWTRefreshTTLShort()
}

// AuthBcryptCost returns the bcrypt cost, clamped to the valid bcrypt range.
func AuthBcryptCost() int {
	cost := viper.GetInt("auth.bcrypt_cost")
	if cost < minBcryptCost || cost > maxBcryptCost {
		return DefaultBcryptCost
	}
	return cost
}

// AuthRateLimitRequests is the number of credential requests allowed per IP
// inside AuthRateLimitWindow.
func AuthRateLimitRequests() int {
	requests := viper.GetInt("auth.rate_limit.requests")
	if requests <= 0 {
		return DefaultAuthRateLimit
	}
	return requests
}

// AuthRateLimitWindow is the fixed window used by the auth rate limiter.
func AuthRateLimitWindow() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("auth.rate_limit.window"), DefaultAuthRateLimitWindow)
}

func GoogleClientID() string {
	return strings.TrimSpace(viper.GetString("auth.google.client_id"))
}

func GoogleClientSecret() string {
	return viper.GetString("auth.google.client_secret")
}

func GoogleRedirectURL() string {
	return utils.ValueOrDefault(viper.GetString("auth.google.redirect_url"), DefaultGoogleRedirectURL)
}

func GoogleAllowedRedirects() []string {
	redirects := viper.GetStringSlice("auth.google.allowed_redirects")
	if len(redirects) == 0 {
		return CORSAllowedOrigins()
	}
	return redirects
}

func GoogleConfigured() bool { return GoogleClientID() != "" }

// ValidateAuthConfig fails fast when production auth is misconfigured.
// It is wired into server startup once the auth endpoints are enabled.
func ValidateAuthConfig() error {
	if !IsProduction() {
		return nil
	}
	if AuthBypass() {
		return errors.New("auth.bypass must be false in production")
	}
	secret := JWTSecret()
	if secret == DefaultJWTSecret || len(secret) < minJWTSecretLength {
		return errors.New("auth.jwt.secret must be set to a strong value in production")
	}
	return nil
}

// LLM

func LLMProvider() string {
	return utils.ValueOrDefault(viper.GetString("llm.provider"), DefaultLLMProvider)
}

func LLMBaseURL() string {
	return utils.ValueOrDefault(viper.GetString("llm.base_url"), DefaultLLMBaseURL)
}

func LLMAPIKey() string { return viper.GetString("llm.api_key") }

func LLMModel() string {
	return utils.ValueOrDefault(viper.GetString("llm.model"), DefaultLLMModel)
}

func LLMTimeout() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("llm.timeout"), DefaultLLMTimeout)
}

// Extractor

func ExtractorTimeout() time.Duration {
	return utils.ParseDurationWithDefault(viper.GetString("extractor.timeout"), DefaultExtractorTimeout)
}

func ExtractorMaxContentLength() int64 {
	return utils.ValueOrDefault(viper.GetInt64("extractor.max_content_length"), int64(DefaultExtractorMaxLength))
}

func urlQueryEscape(value string) string {
	return strings.ReplaceAll(value, " ", "_")
}
