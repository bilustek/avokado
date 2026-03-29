package slack

import (
	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/slack/consoleslacker"
)

// Option is a functional option for configuring the notifier.
type Option func(*Notifier) error

// Notifier is the slack notification hub that delegates to the active SlackNotifier implementation.
type Notifier struct {
	serverEnvironmentName string
	slacker               avokadonotifier.SlackNotifier
}

// WithServerEnvironmentName sets the notifier's server environment.
func WithServerEnvironmentName(serverEnvironmentName string) Option {
	return func(n *Notifier) error {
		n.serverEnvironmentName = serverEnvironmentName

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
		// TODO: webhookslacker
		return nil, avokadoerror.New("[avokadonotifier/slack.New] non-development environments not yet supported")
	}

	return notifier, nil
}
