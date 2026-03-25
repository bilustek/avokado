package avologger

import (
	"io"
	"log/slog"
	"os"

	"github.com/getsentry/sentry-go"
)

// Option is a functional option for configuring the logger.
type Option func(*config) error

type config struct {
	level     slog.Level
	handler   slog.Handler
	writer    io.Writer
	sentryDSN string
}

// WithLevel sets the minimum log level.
func WithLevel(level slog.Level) Option {
	return func(c *config) error {
		c.level = level

		return nil
	}
}

// WithHandler sets a custom slog.Handler, replacing the default JSON handler.
// This allows developers to extend the logger with custom handlers.
func WithHandler(handler slog.Handler) Option {
	return func(c *config) error {
		c.handler = handler

		return nil
	}
}

// WithWriter sets a custom writer for the default JSON handler.
// Ignored if WithHandler is also provided.
func WithWriter(writer io.Writer) Option {
	return func(c *config) error {
		c.writer = writer

		return nil
	}
}

// WithSentryDSN enables Sentry integration. Error-level and above logs
// will be forwarded to Sentry as events. The sentryHandledError pattern
// prevents duplicate captures from panic recovery middleware.
func WithSentryDSN(dsn string) Option {
	return func(c *config) error {
		c.sentryDSN = dsn

		return nil
	}
}

// New creates a new *slog.Logger with the given options.
// By default, it uses a JSON handler writing to os.Stdout at Info level.
func New(opts ...Option) (*slog.Logger, error) {
	cfg := &config{
		level:  slog.LevelInfo,
		writer: os.Stdout,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	handler := cfg.handler
	if handler == nil {
		handler = slog.NewJSONHandler(cfg.writer, &slog.HandlerOptions{
			Level: cfg.level,
		})
	}

	// If Sentry DSN is provided, initialize Sentry and wrap the handler.
	if cfg.sentryDSN != "" {
		if err := initSentry(cfg.sentryDSN); err != nil {
			return nil, err
		}

		captureFunc := func(msg string, level slog.Level) {
			event := &sentry.Event{
				Message: msg,
				Level:   slogLevelToSentry(level),
			}
			sentry.CaptureEvent(event)
		}

		handler = NewSentryHandler(handler, captureFunc)
	}

	return slog.New(handler), nil
}

// initSentry initializes the Sentry SDK with the given DSN.
func initSentry(dsn string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
}

// IsSentryEnabled checks if the given logger has Sentry integration enabled
// by inspecting its handler chain for a SentryHandler.
func IsSentryEnabled(logger *slog.Logger) bool {
	_, ok := logger.Handler().(*SentryHandler)

	return ok
}

// slogLevelToSentry converts slog.Level to sentry.Level.
func slogLevelToSentry(level slog.Level) sentry.Level {
	switch {
	case level >= slog.LevelError:
		return sentry.LevelError
	case level >= slog.LevelWarn:
		return sentry.LevelWarning
	case level >= slog.LevelInfo:
		return sentry.LevelInfo
	default:
		return sentry.LevelDebug
	}
}
