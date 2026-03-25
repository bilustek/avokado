package avologger

import (
	"context"
	"log/slog"
)

// SentryHandledAttrKey is the attribute key used to mark errors that have
// already been captured by Sentry (e.g., from panic recovery middleware).
// When this attribute is present and true, the SentryHandler skips duplicate
// Sentry capture.
const SentryHandledAttrKey = "sentry_handled"

// CaptureFunc is a function that captures a message to Sentry.
// This abstraction allows testing without a real Sentry connection.
type CaptureFunc func(msg string, level slog.Level)

// SentryHandler is an slog.Handler that wraps an inner handler and forwards
// Error-level and above log records to Sentry. It implements duplicate
// prevention via the sentryHandledError pattern.
type SentryHandler struct {
	inner   slog.Handler
	capture CaptureFunc
}

// NewSentryHandler creates a new SentryHandler wrapping the given inner handler.
// The capture function is called for Error+ level records. If capture is nil,
// Sentry capture is skipped (useful for testing the handler in isolation).
func NewSentryHandler(inner slog.Handler, capture CaptureFunc) *SentryHandler {
	return &SentryHandler{
		inner:   inner,
		capture: capture,
	}
}

// Enabled reports whether the inner handler is enabled for the given level.
func (h *SentryHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

// Handle processes the log record. For Error-level and above, it forwards
// the message to Sentry via the capture function, unless the record contains
// the sentryHandledError attribute (duplicate prevention).
func (h *SentryHandler) Handle(ctx context.Context, record slog.Record) error {
	// Always pass to inner handler first.
	if err := h.inner.Handle(ctx, record); err != nil {
		return err
	}

	// Only capture Error+ to Sentry.
	if record.Level < slog.LevelError {
		return nil
	}

	// Check for sentryHandledError attribute -- skip duplicate capture.
	if isSentryHandled(record) {
		return nil
	}

	// Capture to Sentry.
	if h.capture != nil {
		h.capture(record.Message, record.Level)
	}

	return nil
}

// WithAttrs returns a new handler with the given attributes added.
func (h *SentryHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SentryHandler{
		inner:   h.inner.WithAttrs(attrs),
		capture: h.capture,
	}
}

// WithGroup returns a new handler with the given group name.
func (h *SentryHandler) WithGroup(name string) slog.Handler {
	return &SentryHandler{
		inner:   h.inner.WithGroup(name),
		capture: h.capture,
	}
}

// isSentryHandled checks if the record contains the sentry_handled attribute
// set to true, indicating this error was already captured by Sentry
// (e.g., from panic recovery middleware).
func isSentryHandled(record slog.Record) bool {
	handled := false

	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == SentryHandledAttrKey {
			if attr.Value.Kind() == slog.KindBool && attr.Value.Bool() {
				handled = true

				return false // stop iteration
			}
		}

		return true // continue iteration
	})

	return handled
}
