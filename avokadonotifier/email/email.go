package email

import (
	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/consolemailer"
)

// Option is a functional option for configuring the notifier.
type Option func(*Notifier) error

// Notifier ...
type Notifier struct {
	serverEnvironmentName string
	emailer               avokadonotifier.EmailSender
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
		notifier.emailer = consolemailer.New()
	}

	return notifier, nil
}
