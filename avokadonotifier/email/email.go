package email

import (
	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/consolenotifier"
)

// Option is a functional option for configuring the notifier.
type Option func(*Notifier) error

// Notifier ...
type Notifier struct {
	serverEnvironmentName string
	emailer               avokadonotifier.EmailSender
	slacker               avokadonotifier.SlackNotifier
}

// WithServerEnvironmentName sets the notifier's server environment.
func WithServerEnvironmentName(serverEnvironmentName string) Option {
	return func(n *Notifier) error {
		n.serverEnvironmentName = serverEnvironmentName

		return nil
	}
}

// New ...
func New(opts ...Option) (*Notifier, error) {
	notifier := &Notifier{
		serverEnvironmentName: "development",
		slacker:               nil,
	}

	for _, opt := range opts {
		if err := opt(notifier); err != nil {
			return nil, err
		}
	}

	if notifier.serverEnvironmentName == "" {
		return nil, avokadoerror.New("[avokadonotifier.New] serverEnvironmentName required")
	}

	if notifier.serverEnvironmentName == "development" {
		notifier.emailer = consolenotifier.New()
	}

	return notifier, nil
}
