package consoleslacker_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/bilustek/avokado/avokadonotifier/slack/consoleslacker"
)

type failWriter struct{}

func (failWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

type notifyWriter struct {
	mu   sync.Mutex
	buf  bytes.Buffer
	done chan struct{}
	seen string
}

func newNotifyWriter(marker string) *notifyWriter {
	return &notifyWriter{
		done: make(chan struct{}),
		seen: marker,
	}
}

func (w *notifyWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err := w.buf.Write(p)
	if strings.Contains(w.buf.String(), w.seen) {
		select {
		case <-w.done:
		default:
			close(w.done)
		}
	}

	return n, err
}

func (w *notifyWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.buf.String()
}

func TestConsoleNotify(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := consoleslacker.New(consoleslacker.WithWriter(&buf))

	if err := c.Notify(context.Background(), "deploy completed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

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

	if err := c.Notify(context.Background(), "msg"); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestConsoleNotifyAsync(t *testing.T) {
	t.Parallel()

	nw := newNotifyWriter("async msg")
	c := consoleslacker.New(consoleslacker.WithWriter(nw))

	c.NotifyAsync(context.Background(), "async msg")
	<-nw.done

	if !strings.Contains(nw.String(), "async msg") {
		t.Error("expected output to contain message")
	}
}

func TestConsoleNotifyAsync_WriterError(t *testing.T) {
	t.Parallel()

	c := consoleslacker.New(consoleslacker.WithWriter(failWriter{}))

	// should not panic
	c.NotifyAsync(context.Background(), "msg")
}
