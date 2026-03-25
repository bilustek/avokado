package avologger_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/bilustek/avokado/avologger"
)

func TestNewSentryHandler_WrapsInnerHandler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)

	handler := avologger.NewSentryHandler(inner, nil)
	if handler == nil {
		t.Fatal("expected non-nil SentryHandler")
	}

	// Should implement slog.Handler
	var _ slog.Handler = handler
}

func TestSentryHandler_PassesInfoToInner(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := avologger.NewSentryHandler(inner, nil)

	logger := slog.New(handler)
	logger.Info("info message")

	if buf.Len() == 0 {
		t.Error("expected Info message to be passed to inner handler")
	}
}

func TestSentryHandler_PassesWarnToInner(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := avologger.NewSentryHandler(inner, nil)

	logger := slog.New(handler)
	logger.Warn("warn message")

	if buf.Len() == 0 {
		t.Error("expected Warn message to be passed to inner handler")
	}
}

func TestSentryHandler_ErrorLevelTriggersSentry(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)

	captured := make([]string, 0)
	mockCapture := func(msg string, level slog.Level) {
		captured = append(captured, msg)
	}

	handler := avologger.NewSentryHandler(inner, mockCapture)

	logger := slog.New(handler)
	logger.Error("error message")

	if buf.Len() == 0 {
		t.Error("expected Error message to be passed to inner handler")
	}

	if len(captured) != 1 {
		t.Fatalf("expected 1 Sentry capture, got %d", len(captured))
	}

	if captured[0] != "error message" {
		t.Errorf("expected captured message 'error message', got %q", captured[0])
	}
}

func TestSentryHandler_InfoDoesNotTriggerSentry(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)

	captured := make([]string, 0)
	mockCapture := func(msg string, level slog.Level) {
		captured = append(captured, msg)
	}

	handler := avologger.NewSentryHandler(inner, mockCapture)

	logger := slog.New(handler)
	logger.Info("info message")

	if len(captured) != 0 {
		t.Errorf("expected 0 Sentry captures for Info, got %d", len(captured))
	}
}

func TestSentryHandler_DetectsSentryHandledError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)

	captured := make([]string, 0)
	mockCapture := func(msg string, level slog.Level) {
		captured = append(captured, msg)
	}

	handler := avologger.NewSentryHandler(inner, mockCapture)

	// Create a record at Error level with the sentry_handled sentinel attribute.
	ctx := context.Background()
	rec := slog.NewRecord(
		time.Now(),
		slog.LevelError,
		"error already handled by sentry",
		0,
	)
	rec.AddAttrs(slog.Bool(avologger.SentryHandledAttrKey, true))

	_ = handler.Handle(ctx, rec)

	if len(captured) != 0 {
		t.Errorf("expected 0 Sentry captures for sentryHandledError, got %d", len(captured))
	}
}

func TestSentryHandler_WithAttrs(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := avologger.NewSentryHandler(inner, nil)

	withAttrs := handler.WithAttrs([]slog.Attr{slog.String("service", "test")})
	if withAttrs == nil {
		t.Fatal("WithAttrs should return a new handler")
	}
}

func TestSentryHandler_WithGroup(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, nil)
	handler := avologger.NewSentryHandler(inner, nil)

	withGroup := handler.WithGroup("request")
	if withGroup == nil {
		t.Fatal("WithGroup should return a new handler")
	}
}

func TestSentryHandler_Enabled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	inner := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})
	handler := avologger.NewSentryHandler(inner, nil)

	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expected Info to be disabled when inner is at Warn level")
	}

	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("expected Warn to be enabled when inner is at Warn level")
	}
}
