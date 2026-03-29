package consolemailer_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/consolemailer"
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

func TestConsoleSend_PlainText(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := consolemailer.New(consolemailer.WithWriter(&buf))

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Console Test",
		Text:    "Hello from console",
	}

	err := c.Send(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "From: sender@example.com") {
		t.Error("expected output to contain From header")
	}
	if !strings.Contains(output, "To: recipient@example.com") {
		t.Error("expected output to contain To header")
	}
	if !strings.Contains(output, "Subject: Console Test") {
		t.Error("expected output to contain Subject header")
	}
	if !strings.Contains(output, "Hello from console") {
		t.Error("expected output to contain body text")
	}

	separator := strings.Repeat("-", 72)
	separatorCount := strings.Count(output, separator)
	if separatorCount != 2 {
		t.Errorf("expected 2 separators, got %d", separatorCount)
	}
}

func TestConsoleSend_HTMLEmail(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := consolemailer.New(consolemailer.WithWriter(&buf))

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "HTML Console",
		HTML:    "<h1>Hello</h1>",
	}

	err := c.Send(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "<h1>Hello</h1>") {
		t.Error("expected output to contain HTML body")
	}
}

func TestConsoleSend_WithAttachment(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := consolemailer.New(consolemailer.WithWriter(&buf))

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Attachment Console",
		Text:    "see file",
		Attachments: avokadonotifier.EmailAttachments{
			{
				Content:     []byte("file data"),
				Filename:    "report.csv",
				ContentType: "text/csv",
			},
		},
	}

	err := c.Send(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "see file") {
		t.Error("expected output to contain body text")
	}
	if !strings.Contains(output, "report.csv") {
		t.Error("expected output to contain attachment filename")
	}
}

func TestConsoleSend_WriterError(t *testing.T) {
	t.Parallel()

	c := consolemailer.New(consolemailer.WithWriter(failWriter{}))

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Fail",
		Text:    "body",
	}

	if err := c.Send(context.Background(), request); err == nil {
		t.Fatal("expected error from failing writer")
	}
}

func TestConsoleSendAsync(t *testing.T) {
	t.Parallel()

	nw := newNotifyWriter("async body")
	c := consolemailer.New(consolemailer.WithWriter(nw))

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Async",
		Text:    "async body",
	}

	c.SendAsync(context.Background(), request)
	<-nw.done

	if !strings.Contains(nw.String(), "async body") {
		t.Error("expected output to contain body text")
	}
}
