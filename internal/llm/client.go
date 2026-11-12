package llm

import (
	"context"
	"time"

	"github.com/repurpose-ai/re-purpose-be/config"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/sirupsen/logrus"
)

// Input is the normalized request passed to an LLM provider.
type Input struct {
	RawText  string
	Tone     string
	Audience string
	Reroll   bool
}

// Metrics captures inference telemetry that is persisted with each generation.
type Metrics struct {
	LatencyMS         int64
	Tokens            int
	Confidence        string
	SemanticDeviation string
}

// Client is the generation engine abstraction. Implementations must return
// output matching entity.PlatformOutputs (the frontend contract).
type Client interface {
	Generate(ctx context.Context, input Input) (*entity.PlatformOutputs, Metrics, error)
	Name() string
	Model() string
}

// NewFromConfig selects the configured provider. When the OpenAI provider is
// selected without an API key it degrades to the mock engine so local
// development never fails because of missing credentials.
func NewFromConfig() Client {
	switch config.LLMProvider() {
	case "openai":
		if config.LLMAPIKey() == "" {
			logrus.WithField("provider", "openai").Warn("llm.api_key is empty; falling back to mock engine")
			return NewMock()
		}
		return NewOpenAI(
			config.LLMBaseURL(),
			config.LLMAPIKey(),
			config.LLMModel(),
			config.LLMTimeout(),
		)
	default:
		return NewMock()
	}
}

// withTimeout derives a context with the configured LLM timeout.
func withTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}
