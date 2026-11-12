package job

import (
	"fmt"

	"github.com/hibiken/asynq"
)

// NewTaskQueue creates an asynq client for enqueueing background jobs.
func NewTaskQueue(redisURL string) (*asynq.Client, error) {
	options, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis worker url: %w", err)
	}
	return asynq.NewClient(options), nil
}
