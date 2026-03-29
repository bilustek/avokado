package resendmailer

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/resend/resend-go/v3"
)

// compile-time proof of interface implementation.
var _ avokadonotifier.EmailSender = (*Resend)(nil)

// Option is a functional option for configuring the resendmailer.
type Option func(*Resend) error

// Resend is an EmailSender that delivers email via the Resend API.
type Resend struct {
	client *resend.Client
	logger *slog.Logger
}

// Send delivers an email through the Resend API.
func (r *Resend) Send(ctx context.Context, request *avokadonotifier.EmailSenderRequest) error {
	req := avokadonotifier.EmailSenderRequestToResendRequest(request)

	if _, err := r.client.Emails.SendWithContext(ctx, req); err != nil {
		r.logger.ErrorContext(ctx, "[Resend.Send] err", "error", err)

		return avokadoerror.New("[Resend.Send] err").WithErr(err)
	}

	r.logger.InfoContext(ctx, "[Resend.Send] email sent", "to", request.To, "subject", request.Subject)

	return nil
}

// SendAsync delivers an email in a background goroutine, logging is handled by Send.
func (r *Resend) SendAsync(ctx context.Context, request *avokadonotifier.EmailSenderRequest) {
	go func() {
		_ = r.Send(ctx, request)
	}()
}

// WithAPIKey sets the Resend API key.
func WithAPIKey(apiKey string) Option {
	return func(r *Resend) error {
		if apiKey == "" {
			return avokadoerror.New("[resendmailer.WithAPIKey] apiKey must not be empty")
		}

		r.client = resend.NewClient(apiKey)

		return nil
	}
}

// WithHTTPClient sets a custom HTTP client for the Resend API (useful for testing).
func WithHTTPClient(apiKey string, httpClient *http.Client) Option {
	return func(r *Resend) error {
		if apiKey == "" {
			return avokadoerror.New("[resendmailer.WithHTTPClient] apiKey must not be empty")
		}
		if httpClient == nil {
			return avokadoerror.New("[resendmailer.WithHTTPClient] httpClient must not be nil")
		}

		r.client = resend.NewCustomClient(httpClient, apiKey)

		return nil
	}
}

// WithLogger sets the logger for the resendmailer.
func WithLogger(logger *slog.Logger) Option {
	return func(r *Resend) error {
		r.logger = logger

		return nil
	}
}

// New creates a Resend email sender with the given options.
func New(opts ...Option) (*Resend, error) {
	cfg := &Resend{}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.client == nil {
		return nil, avokadoerror.New("[resendmailer.New] API key is required, use WithAPIKey")
	}
	if cfg.logger == nil {
		return nil, avokadoerror.New("[resendmailer.New] logger is required, use WithLogger")
	}

	return cfg, nil
}
