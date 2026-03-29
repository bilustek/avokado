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
	webhookURL            string
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

// WithWebhookURL sets the Slack webhook URL for production delivery.
func WithWebhookURL(url string) Option {
	return func(c *config) error {
		c.webhookURL = url

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
		if cfg.webhookURL == "" {
			return nil, avokadoerror.New(
				"[avokadonotifier/slack.New] webhookURL required for non-development environments, use WithWebhookURL",
			)
		}

		return webhookslacker.New(
			webhookslacker.WithLogger(cfg.logger),
			webhookslacker.WithWebhookURL(cfg.webhookURL),
		)
	}
}
