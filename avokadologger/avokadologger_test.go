package avokadologger_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/bilustek/avokado/avokadologger"
)

func TestNew_ReturnsLogger(t *testing.T) {
	t.Parallel()

	logger, err := avokadologger.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if logger == nil {
		t.Fatal("expected non-nil *slog.Logger")
	}
}

func TestNew_DefaultIsJSONHandler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := avokadologger.New(avokadologger.WithWriter(&buf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger.Info("test message")

	var entry map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &entry); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}

	if _, ok := entry["time"]; !ok {
		t.Error("expected 'time' field in JSON output")
	}

	if _, ok := entry["level"]; !ok {
		t.Error("expected 'level' field in JSON output")
	}

	if _, ok := entry["msg"]; !ok {
		t.Error("expected 'msg' field in JSON output")
	}
}

func TestNew_WithLevel(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := avokadologger.New(
		avokadologger.WithLevel(slog.LevelWarn),
		avokadologger.WithWriter(&buf),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Debug and Info should be filtered
	logger.Debug("debug message")
	logger.Info("info message")

	if buf.Len() > 0 {
		t.Errorf("expected no output for Debug/Info at Warn level, got: %s", buf.String())
	}

	// Warn should pass
	logger.Warn("warn message")

	if buf.Len() == 0 {
		t.Error("expected output for Warn at Warn level")
	}
}

func TestNew_WithCustomHandler(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	customHandler := slog.NewTextHandler(&buf, nil)

	logger, err := avokadologger.New(avokadologger.WithHandler(customHandler))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger.Info("custom handler test")

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected output from custom text handler")
	}

	// Text handler outputs key=value format, not JSON
	var entry map[string]any
	if jsonErr := json.Unmarshal([]byte(output), &entry); jsonErr == nil {
		t.Error("expected non-JSON output from text handler")
	}
}

func TestNew_WithWriter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := avokadologger.New(avokadologger.WithWriter(&buf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger.Info("writer test")

	if buf.Len() == 0 {
		t.Error("expected output written to custom writer")
	}
}

func TestNew_StructuredJSONOutput(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger, err := avokadologger.New(avokadologger.WithWriter(&buf))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logger.Info("structured test", "key1", "value1", "key2", 42)

	var entry map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &entry); jsonErr != nil {
		t.Fatalf("expected JSON output, got: %s", buf.String())
	}

	if entry["msg"] != "structured test" {
		t.Errorf("expected msg 'structured test', got %v", entry["msg"])
	}

	if entry["key1"] != "value1" {
		t.Errorf("expected key1 'value1', got %v", entry["key1"])
	}

	if entry["key2"] != float64(42) {
		t.Errorf("expected key2 42, got %v", entry["key2"])
	}
}
