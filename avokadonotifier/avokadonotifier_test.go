package avokadonotifier_test

import (
	"io"
	"strings"
	"testing"

	"github.com/bilustek/avokado/avokadonotifier"
)

func TestEmailSenderRequestToMailMessage_PlainText(t *testing.T) {
	t.Parallel()

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"recipient@example.com"},
		Subject: "Test Subject",
		Text:    "Hello, World!",
	}

	msg, err := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := msg.Header.Get("From"); got != "sender@example.com" {
		t.Errorf("expected From %q, got %q", "sender@example.com", got)
	}
	if got := msg.Header.Get("To"); got != "recipient@example.com" {
		t.Errorf("expected To %q, got %q", "recipient@example.com", got)
	}
	if got := msg.Header.Get("Subject"); got != "Test Subject" {
		t.Errorf("expected Subject %q, got %q", "Test Subject", got)
	}
	if got := msg.Header.Get("Content-Type"); got != "text/plain; charset=UTF-8" {
		t.Errorf("expected Content-Type %q, got %q", "text/plain; charset=UTF-8", got)
	}
	if got := msg.Header.Get("Date"); got == "" {
		t.Error("expected Date header to be set")
	}

	body, readErr := io.ReadAll(msg.Body)
	if readErr != nil {
		t.Fatalf("unexpected error reading body: %v", readErr)
	}
	if string(body) != "Hello, World!" {
		t.Errorf("expected body %q, got %q", "Hello, World!", string(body))
	}
}

func TestEmailSenderRequestToMailMessage_HTMLOverridesText(t *testing.T) {
	t.Parallel()

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "HTML Test",
		HTML:    "<h1>Hello</h1>",
		Text:    "fallback text",
	}

	msg, err := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := msg.Header.Get("Content-Type"); got != "text/html; charset=UTF-8" {
		t.Errorf("expected Content-Type %q, got %q", "text/html; charset=UTF-8", got)
	}

	body, readErr := io.ReadAll(msg.Body)
	if readErr != nil {
		t.Fatalf("unexpected error reading body: %v", readErr)
	}
	if string(body) != "<h1>Hello</h1>" {
		t.Errorf("expected body %q, got %q", "<h1>Hello</h1>", string(body))
	}
}

func TestEmailSenderRequestToMailMessage_MultipleRecipients(t *testing.T) {
	t.Parallel()

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com", "b@example.com"},
		Subject: "Multi",
		Text:    "body",
	}

	msg, err := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := msg.Header.Get("To"); got != "a@example.com, b@example.com" {
		t.Errorf("expected To %q, got %q", "a@example.com, b@example.com", got)
	}
}

func TestEmailSenderRequestToMailMessage_OptionalHeaders(t *testing.T) {
	t.Parallel()

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Headers Test",
		Text:    "body",
		Bcc:     []string{"bcc1@example.com", "bcc2@example.com"},
		Cc:      []string{"cc@example.com"},
		ReplyTo: "reply@example.com",
		Headers: map[string]string{
			"X-Custom": "custom-value",
		},
	}

	msg, err := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := msg.Header.Get("Bcc"); got != "bcc1@example.com, bcc2@example.com" {
		t.Errorf("expected Bcc %q, got %q", "bcc1@example.com, bcc2@example.com", got)
	}
	if got := msg.Header.Get("Cc"); got != "cc@example.com" {
		t.Errorf("expected Cc %q, got %q", "cc@example.com", got)
	}
	if got := msg.Header.Get("Reply-To"); got != "reply@example.com" {
		t.Errorf("expected Reply-To %q, got %q", "reply@example.com", got)
	}
	if got := msg.Header.Get("X-Custom"); got != "custom-value" {
		t.Errorf("expected X-Custom %q, got %q", "custom-value", got)
	}
}

func TestEmailSenderRequestToMailMessage_EmptyOptionalHeadersOmitted(t *testing.T) {
	t.Parallel()

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "No Optionals",
		Text:    "body",
	}

	msg, err := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := msg.Header.Get("Bcc"); got != "" {
		t.Errorf("expected empty Bcc, got %q", got)
	}
	if got := msg.Header.Get("Cc"); got != "" {
		t.Errorf("expected empty Cc, got %q", got)
	}
	if got := msg.Header.Get("Reply-To"); got != "" {
		t.Errorf("expected empty Reply-To, got %q", got)
	}
}

func TestEmailSenderRequestToMailMessage_WithAttachment(t *testing.T) {
	t.Parallel()

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "With Attachment",
		Text:    "see attached",
		Attachments: avokadonotifier.EmailAttachments{
			{
				Content:     []byte("file content here"),
				Filename:    "test.txt",
				ContentType: "text/plain",
			},
		},
	}

	msg, err := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ct := msg.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "multipart/mixed; boundary=") {
		t.Errorf("expected multipart/mixed Content-Type, got %q", ct)
	}

	body, readErr := io.ReadAll(msg.Body)
	if readErr != nil {
		t.Fatalf("unexpected error reading body: %v", readErr)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "see attached") {
		t.Error("expected body to contain text content")
	}
	if !strings.Contains(bodyStr, "Content-Type: text/plain") {
		t.Error("expected body to contain attachment content type")
	}
	if !strings.Contains(bodyStr, "Content-Transfer-Encoding: base64") {
		t.Error("expected body to contain base64 transfer encoding")
	}
	if !strings.Contains(bodyStr, "test.txt") {
		t.Error("expected body to contain attachment filename")
	}
}

func TestEmailSenderRequestToMailMessage_MultipleAttachments(t *testing.T) {
	t.Parallel()

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Multi Attach",
		HTML:    "<p>files</p>",
		Attachments: avokadonotifier.EmailAttachments{
			{
				Content:     []byte("pdf content"),
				Filename:    "doc.pdf",
				ContentType: "application/pdf",
			},
			{
				Content:     []byte("image content"),
				Filename:    "photo.png",
				ContentType: "image/png",
			},
		},
	}

	msg, err := avokadonotifier.EmailSenderRequestToMailMessage(request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, readErr := io.ReadAll(msg.Body)
	if readErr != nil {
		t.Fatalf("unexpected error reading body: %v", readErr)
	}

	bodyStr := string(body)
	if !strings.Contains(bodyStr, "text/html") {
		t.Error("expected multipart body to contain html content type")
	}
	if !strings.Contains(bodyStr, "doc.pdf") {
		t.Error("expected body to contain first attachment filename")
	}
	if !strings.Contains(bodyStr, "photo.png") {
		t.Error("expected body to contain second attachment filename")
	}
	if !strings.Contains(bodyStr, "application/pdf") {
		t.Error("expected body to contain first attachment content type")
	}
	if !strings.Contains(bodyStr, "image/png") {
		t.Error("expected body to contain second attachment content type")
	}
}
