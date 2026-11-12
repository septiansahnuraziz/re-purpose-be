package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/repurpose-ai/re-purpose-be/commons"
	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/llm"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/sirupsen/logrus"
)

// Processor handles asynchronous jobs (bulk generation, retries, webhooks, etc.).
// The synchronous HTTP endpoints call the service directly; this worker is the
// extension point for long-running work.
type Processor struct {
	engine llm.Client
}

func NewProcessor(engine llm.Client) *Processor {
	return &Processor{engine: engine}
}

func (p *Processor) Register(mux *asynq.ServeMux) {
	mux.HandleFunc(commons.TaskTypeContentGenerate, p.HandleContentGenerate)
}

// HandleContentGenerate is a scaffold for async generation.
// TODO: persist the result, notify the user (websocket/email), and deduct credits.
func (p *Processor) HandleContentGenerate(ctx context.Context, task *asynq.Task) error {
	log := logger.FromContext(ctx).WithField("task", task.Type())

	var request entity.GenerateContentRequest
	if err := json.Unmarshal(task.Payload(), &request); err != nil {
		log.WithError(err).Error("failed to decode task payload")
		return fmt.Errorf("decode %s payload: %w", task.Type(), err)
	}

	outputs, metrics, err := p.engine.Generate(ctx, llm.Input{
		RawText:  request.RawText,
		Tone:     request.Tone,
		Audience: request.Audience,
	})
	if err != nil {
		log.WithError(err).Error("async content generation failed")
		return fmt.Errorf("run async generation: %w", err)
	}

	log.WithFields(logrus.Fields{
		"tone":      request.Tone,
		"tweets":    len(outputs.Tweets),
		"latencyMs": metrics.LatencyMS,
		"tokens":    metrics.Tokens,
	}).Info("async content generation finished")

	return nil
}
