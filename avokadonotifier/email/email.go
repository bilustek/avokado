package email

import (
	"log/slog"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/consolemailer"
	"github.com/bilustek/avokado/avokadonotifier/email/resendmailer"
)

// Option is a functional option for configuring the notifier.
type Option func(*Notifier) error

// Notifier is the email notification hub that delegates to the active EmailSender implementation.
type Notifier struct {
	serverEnvironmentName string
	resendAPIKey          string
	logger                *slog.Logger
	emailer               avokadonotifier.EmailSender
}

// WithServerEnvironmentName sets the notifier's server environment.
func WithServerEnvironmentName(serverEnvironmentName string) Option {
	return func(n *Notifier) error {
		n.serverEnvironmentName = serverEnvironmentName

		return nil
	}
}

// WithResendAPIKey sets the Resend API key for production email delivery.
func WithResendAPIKey(apiKey string) Option {
	return func(n *Notifier) error {
		n.resendAPIKey = apiKey

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
		return nil, avokadoerror.New("[avokadonotifier.New] serverEnvironmentName required")
	}

	switch notifier.serverEnvironmentName {
	case "development":
		notifier.emailer = consolemailer.New()
	default:
		if notifier.resendAPIKey == "" {
			return nil, avokadoerror.New("[avokadonotifier.New] resendAPIKey required for non-development environments")
		}
		if notifier.logger == nil {
			return nil, avokadoerror.New(
				"[avokadonotifier.New] logger required for non-development environments, use WithLogger",
			)
		}

		mailer, mailerErr := resendmailer.New(
			resendmailer.WithAPIKey(notifier.resendAPIKey),
			resendmailer.WithLogger(notifier.logger),
		)
		if mailerErr != nil {
			return nil, mailerErr
		}

		notifier.emailer = mailer
	}

	return notifier, nil
}
