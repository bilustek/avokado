package slack

import (
	"log/slog"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/slack/consoleslacker"
	"github.com/bilustek/avokado/avokadonotifier/slack/webhookslacker"
)

// Option is a functional option for configuring the notifier.
type Option func(*Notifier) error

// Notifier is the slack notification hub that delegates to the active SlackNotifier implementation.
type Notifier struct {
	serverEnvironmentName string
	logger                *slog.Logger
	slacker               avokadonotifier.SlackNotifier
}

// WithServerEnvironmentName sets the notifier's server environment.
func WithServerEnvironmentName(serverEnvironmentName string) Option {
	return func(n *Notifier) error {
		n.serverEnvironmentName = serverEnvironmentName

		return nil
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(n *Notifier) error {
		n.logger = logger

		return nil
	}
}

// New creates a Notifier configured via functional options.
func New(opts ...Option) (*Notifier, error) {
	notifier := &Notifier{
		serverEnvironmentName: "development",
	}

	for _, opt := range opts {
		if err := opt(notifier); err != nil {
			return nil, err
		}
	}

	if notifier.serverEnvironmentName == "" {
		return nil, avokadoerror.New("[avokadonotifier/slack.New] serverEnvironmentName required")
	}

	switch notifier.serverEnvironmentName {
	case "development":
		notifier.slacker = consoleslacker.New()
	default:
		if notifier.logger == nil {
			return nil, avokadoerror.New(
				"[avokadonotifier/slack.New] logger required for non-development environments, use WithLogger",
			)
		}

		slacker, slackerErr := webhookslacker.New(webhookslacker.WithLogger(notifier.logger))
		if slackerErr != nil {
			return nil, slackerErr
		}

		notifier.slacker = slacker
	}

	return notifier, nil
}
