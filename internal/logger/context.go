package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

// FromContext returns a log entry enriched with the request/trace and user ids
// stored in ctx. It is the preferred way to log inside handlers and services.
func FromContext(ctx context.Context) *logrus.Entry {
	return logrus.WithContext(ctx)
}

// WithFields returns a context-aware entry with additional structured fields.
func WithFields(ctx context.Context, fields logrus.Fields) *logrus.Entry {
	return FromContext(ctx).WithFields(fields)
}
