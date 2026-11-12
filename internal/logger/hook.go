package logger

import (
	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/sirupsen/logrus"
)

// Field names shared by the logger hooks and helpers.
const (
	FieldRequestID = "requestId"
	FieldTraceID   = "traceId"
	FieldUserID    = "userId"
)

// defaultFieldsHook adds service metadata to every log entry.
type defaultFieldsHook struct {
	fields logrus.Fields
}

func (h *defaultFieldsHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *defaultFieldsHook) Fire(entry *logrus.Entry) error {
	for key, value := range h.fields {
		if value == "" {
			continue
		}
		if _, exists := entry.Data[key]; !exists {
			entry.Data[key] = value
		}
	}
	return nil
}

// contextHook enriches entries created with logrus.WithContext(ctx) with the
// request/trace id and user id stored in the context.
type contextHook struct{}

func (h *contextHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *contextHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}

	if _, exists := entry.Data[FieldRequestID]; !exists {
		if traceID := utils.TraceIDFromContext(entry.Context); traceID != "" {
			entry.Data[FieldRequestID] = traceID
			entry.Data[FieldTraceID] = traceID
		}
	}

	if _, exists := entry.Data[FieldUserID]; !exists {
		if userID := utils.UserIDFromContext(entry.Context); userID != "" {
			entry.Data[FieldUserID] = userID
		}
	}

	return nil
}
