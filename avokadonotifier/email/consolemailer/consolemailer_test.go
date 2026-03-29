package consolemailer_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/consolemailer"
)

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

func TestConsoleImplementsEmailSender(t *testing.T) {
	t.Parallel()

	var _ avokadonotifier.EmailSender = consolemailer.New()
}
