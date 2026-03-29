package consolenotifier

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
)

// Option is a functional option for configuring the consolenotifier.
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
	fmt.Fprintf(s.writer, "%s\n", strings.Repeat("-", 72))
	for k, vals := range msg.Header {
		for _, v := range vals {
			fmt.Fprintf(s.writer, "%s: %s\n", k, v)
		}
	}
	fmt.Fprintln(s.writer)

	body, err := io.ReadAll(msg.Body)
	if err != nil {
		return avokadoerror.New("[Console.Send] body err").WithErr(err)
	}
	fmt.Fprintf(s.writer, "%s\n", body)
	fmt.Fprintf(s.writer, "%s\n", strings.Repeat("-", 72))
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
