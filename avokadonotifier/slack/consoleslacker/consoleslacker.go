package consoleslacker

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/bilustek/avokado/avokadonotifier"
)

// compile-time proof of interface implementation.
var _ avokadonotifier.SlackNotifier = (*Console)(nil)

// Option is a functional option for configuring the consoleslacker.
type Option func(*Console)

// Console is a SlackNotifier that writes slack messages to an io.Writer for development use.
type Console struct {
	writer io.Writer
}

// Notify writes the slack message to the configured writer.
func (c *Console) Notify(_ context.Context, message string) error {
	separator := strings.Repeat("-", 72)

	if _, err := io.WriteString(c.writer, separator+"\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(c.writer, "Message: "+message+"\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(c.writer, separator+"\n"); err != nil {
		return err
	}

	return nil
}

// NotifyAsync writes the slack message to the configured writer in a background goroutine.
func (c *Console) NotifyAsync(ctx context.Context, message string) {
	go func() {
		_ = c.Notify(ctx, message)
	}()
}

// WithWriter overrides the default output destination (os.Stderr).
func WithWriter(w io.Writer) Option {
	return func(c *Console) {
		c.writer = w
	}
}

// New creates a console slack notifier with the given options.
func New(opts ...Option) *Console {
	cfg := &Console{
		writer: os.Stderr,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}
