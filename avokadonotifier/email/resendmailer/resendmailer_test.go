package resendmailer_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/resendmailer"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestNew_WithoutAPIKey_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := resendmailer.New(); err == nil {
		t.Fatal("expected error when no API key provided")
	}
}

func TestNew_WithoutLogger_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := resendmailer.New(resendmailer.WithAPIKey("re_test_123")); err == nil {
		t.Fatal("expected error when no logger provided")
	}
}

func TestWithAPIKey_Empty_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := resendmailer.New(resendmailer.WithAPIKey("")); err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestWithHTTPClient_EmptyAPIKey_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := resendmailer.New(
		resendmailer.WithHTTPClient("", &http.Client{}),
	); err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestWithHTTPClient_NilClient_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := resendmailer.New(
		resendmailer.WithHTTPClient("re_test_123", nil),
	); err == nil {
		t.Fatal("expected error for nil HTTP client")
	}
}

func TestNew_WithAllOptions_Succeeds(t *testing.T) {
	t.Parallel()

	r, err := resendmailer.New(
		resendmailer.WithAPIKey("re_test_123"),
		resendmailer.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil Resend instance")
	}
}

func TestSend_Success(t *testing.T) {
	t.Parallel()

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		resp := map[string]string{"id": "test-id-123"}
		body, _ := json.Marshal(resp)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Header:     make(http.Header),
		}, nil
	})

	r, err := resendmailer.New(
		resendmailer.WithHTTPClient("re_test_123", client),
		resendmailer.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Test",
		Text:    "hello",
	}

	if err := r.Send(context.Background(), request); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSend_APIError(t *testing.T) {
	t.Parallel()

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnprocessableEntity,
			Body:       io.NopCloser(strings.NewReader(`{"message":"validation error"}`)),
			Header:     make(http.Header),
		}, nil
	})

	r, err := resendmailer.New(
		resendmailer.WithHTTPClient("re_test_123", client),
		resendmailer.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Test",
		Text:    "hello",
	}

	if err := r.Send(context.Background(), request); err == nil {
		t.Fatal("expected error from API")
	}
}

func TestSendAsync(t *testing.T) {
	t.Parallel()

	done := make(chan struct{})

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		defer close(done)

		resp := map[string]string{"id": "async-id"}
		body, _ := json.Marshal(resp)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(body))),
			Header:     make(http.Header),
		}, nil
	})

	r, err := resendmailer.New(
		resendmailer.WithHTTPClient("re_test_123", client),
		resendmailer.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	request := &avokadonotifier.EmailSenderRequest{
		From:    "sender@example.com",
		To:      []string{"a@example.com"},
		Subject: "Async Test",
		Text:    "async",
	}

	r.SendAsync(context.Background(), request)
	<-done
}
