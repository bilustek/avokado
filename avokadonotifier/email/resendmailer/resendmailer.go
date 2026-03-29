package resendmailer

import (
	"context"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/resend/resend-go/v3"
)

// Option is a functional option for configuring the resendmailer.
type Option func(*Resend)

// Resend is an EmailSender that delivers email via the Resend API.
type Resend struct {
	client *resend.Client
}

// Send delivers an email through the Resend API.
func (r *Resend) Send(_ context.Context, request *avokadonotifier.EmailSenderRequest) error {
	req := avokadonotifier.EmailSenderRequestToResendRequest(request)

	_, err := r.client.Emails.Send(req)
	if err != nil {
		return avokadoerror.New("[Resend.Send] err").WithErr(err)
	}

	return nil
}

// WithAPIKey sets the Resend API key.
func WithAPIKey(apiKey string) Option {
	return func(r *Resend) {
		r.client = resend.NewClient(apiKey)
	}
}

// New creates a Resend email sender with the given options.
func New(opts ...Option) (*Resend, error) {
	cfg := &Resend{}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.client == nil {
		return nil, avokadoerror.New("[resendmailer.New] API key is required, use WithAPIKey")
	}

	return cfg, nil
}
