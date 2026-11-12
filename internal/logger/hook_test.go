package logger

import (
	"context"
	"testing"

	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/sirupsen/logrus"
)

func TestContextHookAddsRequestAndUserFields(t *testing.T) {
	ctx := utils.WithTraceID(context.Background(), "trace-123")
	ctx = utils.WithUserID(ctx, "usr_demo")

	entry := &logrus.Entry{Context: ctx, Data: logrus.Fields{}}
	if err := (&contextHook{}).Fire(entry); err != nil {
		t.Fatalf("contextHook.Fire() error = %v", err)
	}

	if got := entry.Data[FieldRequestID]; got != "trace-123" {
		t.Fatalf("%s = %v, want trace-123", FieldRequestID, got)
	}
	if got := entry.Data[FieldTraceID]; got != "trace-123" {
		t.Fatalf("%s = %v, want trace-123", FieldTraceID, got)
	}
	if got := entry.Data[FieldUserID]; got != "usr_demo" {
		t.Fatalf("%s = %v, want usr_demo", FieldUserID, got)
	}
}

func TestContextHookKeepsExplicitFields(t *testing.T) {
	ctx := utils.WithTraceID(context.Background(), "trace-123")

	entry := &logrus.Entry{
		Context: ctx,
		Data:    logrus.Fields{FieldRequestID: "explicit-request"},
	}
	if err := (&contextHook{}).Fire(entry); err != nil {
		t.Fatalf("contextHook.Fire() error = %v", err)
	}

	if got := entry.Data[FieldRequestID]; got != "explicit-request" {
		t.Fatalf("%s = %v, want explicit-request", FieldRequestID, got)
	}
}

func TestDefaultFieldsHookSkipsEmptyValues(t *testing.T) {
	entry := &logrus.Entry{Data: logrus.Fields{}}
	hook := &defaultFieldsHook{fields: logrus.Fields{
		"service":     "re-purpose-be",
		"environment": "",
	}}

	if err := hook.Fire(entry); err != nil {
		t.Fatalf("defaultFieldsHook.Fire() error = %v", err)
	}

	if got := entry.Data["service"]; got != "re-purpose-be" {
		t.Fatalf("service = %v, want re-purpose-be", got)
	}
	if _, exists := entry.Data["environment"]; exists {
		t.Fatal("empty environment field should not be added")
	}
}
