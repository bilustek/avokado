package consolemailer

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
)

// Option is a functional option for configuring the consolemailer.
type Option func(*Console)

// Console ...
type Console struct {
	writer io.Writer
}

// Send ...
func (s *Console) Send(_ context.Context, request *avokadonotifier.EmailSenderRequest) error {
	msg, msgErr := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if msgErr != nil {
		return avokadoerror.New("[Console.Send] err").WithErr(msgErr)
	}

	separator := strings.Repeat("-", 72)

	if _, err := io.WriteString(s.writer, separator+"\n"); err != nil {
		return avokadoerror.New("[Console.Send] write separator err").WithErr(err)
	}

	for k, vals := range msg.Header {
		for _, v := range vals {
			if _, err := io.WriteString(s.writer, k+": "+v+"\n"); err != nil {
				return avokadoerror.New("[Console.Send] write header err").WithErr(err)
			}
		}
	}

	if _, err := io.WriteString(s.writer, "\n"); err != nil {
		return avokadoerror.New("[Console.Send] write newline err").WithErr(err)
	}

	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return avokadoerror.New("[Console.Send] body err").WithErr(err)
	}

	if _, err := io.WriteString(s.writer, string(body)+"\n"); err != nil {
		return avokadoerror.New("[Console.Send] write body err").WithErr(err)
	}

	if _, err := io.WriteString(s.writer, separator+"\n"); err != nil {
		return avokadoerror.New("[Console.Send] write closing separator err").WithErr(err)
	}

	return nil
}

// WithWriter overrides the default output destination (os.Stderr).
func WithWriter(w io.Writer) Option {
	return func(c *Console) {
		c.writer = w
	}
}

// New creates a console email sender with the given options.
func New(opts ...Option) *Console {
	cfg := &Console{
		writer: os.Stderr,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
