package consoleslacker_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/bilustek/avokado/avokadonotifier/slack/consoleslacker"
)

type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestConsoleNotify(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := consoleslacker.New(consoleslacker.WithWriter(&buf))

	if err := c.Notify(context.Background(), "https://hooks.slack.com/services/xxx", "deploy completed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Webhook: [REDACTED]") {
		t.Error("expected output to contain redacted webhook")
	}
	if !strings.Contains(output, "Message: deploy completed") {
		t.Error("expected output to contain message")
	}

	separator := strings.Repeat("-", 72)
	if strings.Count(output, separator) != 2 {
		t.Errorf("expected 2 separators, got %d", strings.Count(output, separator))
	}
}

func TestConsoleNotify_WriterError(t *testing.T) {
	t.Parallel()

	c := consoleslacker.New(consoleslacker.WithWriter(failWriter{}))

	if err := c.Notify(context.Background(), "https://hooks.slack.com/test", "msg"); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestConsoleNotifyAsync(t *testing.T) {
	t.Parallel()

	r, w := io.Pipe()
	c := consoleslacker.New(consoleslacker.WithWriter(w))

	done := make(chan string)

	go func() {
		var buf bytes.Buffer

		_, _ = io.Copy(&buf, r)

		done <- buf.String()
	}()

	c.NotifyAsync(context.Background(), "https://hooks.slack.com/test", "async msg")
	time.Sleep(50 * time.Millisecond)
	_ = w.Close()

	output := <-done
	if !strings.Contains(output, "async msg") {
		t.Error("expected output to contain message")
	}
}
