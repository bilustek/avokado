package slack

import (
	"log/slog"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/slack/consoleslacker"
	"github.com/bilustek/avokado/avokadonotifier/slack/webhookslacker"
)

// Option is a functional option for configuring the slack notifier.
type Option func(*config) error

type config struct {
	serverEnvironmentName string
	logger                *slog.Logger
}

// WithServerEnvironmentName sets the notifier's server environment.
func WithServerEnvironmentName(serverEnvironmentName string) Option {
	return func(c *config) error {
		c.serverEnvironmentName = serverEnvironmentName

		return nil
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) error {
		c.logger = logger

		return nil
	}
}

// New creates a SlackNotifier configured via functional options.
func New(opts ...Option) (avokadonotifier.SlackNotifier, error) {
	cfg := &config{
		serverEnvironmentName: "development",
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.serverEnvironmentName == "" {
		return nil, avokadoerror.New("[avokadonotifier/slack.New] serverEnvironmentName required")
	}

	switch cfg.serverEnvironmentName {
	case "development":
		return consoleslacker.New(), nil
	default:
		if cfg.logger == nil {
			return nil, avokadoerror.New(
				"[avokadonotifier/slack.New] logger required for non-development environments, use WithLogger",
			)
		}

		return webhookslacker.New(webhookslacker.WithLogger(cfg.logger))
	}
}
