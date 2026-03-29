package email

import (
	"log/slog"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/consolemailer"
	"github.com/bilustek/avokado/avokadonotifier/email/resendmailer"
)

// Option is a functional option for configuring the email notifier.
type Option func(*config) error

type config struct {
	serverEnvironmentName string
	resendAPIKey          string
	logger                *slog.Logger
}

// WithServerEnvironmentName sets the notifier's server environment.
func WithServerEnvironmentName(serverEnvironmentName string) Option {
	return func(c *config) error {
		c.serverEnvironmentName = serverEnvironmentName

		return nil
	}
}

// WithResendAPIKey sets the Resend API key for production email delivery.
func WithResendAPIKey(apiKey string) Option {
	return func(c *config) error {
		c.resendAPIKey = apiKey

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

// New creates an EmailSender configured via functional options.
func New(opts ...Option) (avokadonotifier.EmailSender, error) {
	cfg := &config{
		serverEnvironmentName: "development",
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.serverEnvironmentName == "" {
		return nil, avokadoerror.New("[avokadonotifier/email.New] serverEnvironmentName required")
	}

	switch cfg.serverEnvironmentName {
	case "development":
		return consolemailer.New(), nil
	default:
		if cfg.resendAPIKey == "" {
			return nil, avokadoerror.New(
				"[avokadonotifier/email.New] resendAPIKey required for non-development environments",
			)
		}
		if cfg.logger == nil {
			return nil, avokadoerror.New(
				"[avokadonotifier/email.New] logger required for non-development environments, use WithLogger",
			)
		}

		return resendmailer.New(
			resendmailer.WithAPIKey(cfg.resendAPIKey),
			resendmailer.WithLogger(cfg.logger),
		)
	}
}
