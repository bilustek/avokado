package webhookslacker_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bilustek/avokado/avokadonotifier/slack/webhookslacker"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func mockClient(fn roundTripFunc) *http.Client {
	return &http.Client{Transport: fn}
}

func TestNotify_Success(t *testing.T) {
	t.Parallel()

	client := mockClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", req.Method)
		}
		if ct := req.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}

		body, _ := io.ReadAll(req.Body)
		if !strings.Contains(string(body), "hello slack") {
			t.Error("expected body to contain message")
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})

	w, err := webhookslacker.New(
		webhookslacker.WithHTTPClient(client),
		webhookslacker.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := w.Notify(context.Background(), "https://hooks.slack.com/test", "hello slack"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotify_ClientError_NoRetry(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		callCount.Add(1)

		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("invalid_payload")),
		}, nil
	})

	w, err := webhookslacker.New(
		webhookslacker.WithHTTPClient(client),
		webhookslacker.WithLogger(slog.Default()),
		webhookslacker.WithMaxRetries(2),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := w.Notify(context.Background(), "https://hooks.slack.com/test", "msg"); err == nil {
		t.Fatal("expected error for 400 response")
	}

	if got := callCount.Load(); got != 1 {
		t.Errorf("expected 1 call for client error (no retry), got %d", got)
	}

	if err := w.Notify(context.Background(), "https://hooks.slack.com/test", "msg"); err != nil {
		var unwrapped interface{ Unwrap() error }
		if errors.As(err, &unwrapped) {
			if unwrapped.Unwrap() == nil {
				t.Error("expected non-nil unwrapped error")
			}
		}
	}
}

func TestNotify_ServerError_Retries(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		callCount.Add(1)

		return &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(strings.NewReader("error")),
		}, nil
	})

	w, err := webhookslacker.New(
		webhookslacker.WithHTTPClient(client),
		webhookslacker.WithLogger(slog.Default()),
		webhookslacker.WithMaxRetries(2),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := w.Notify(context.Background(), "https://hooks.slack.com/test", "msg"); err == nil {
		t.Fatal("expected error after retries exhausted")
	}

	if got := callCount.Load(); got != 3 {
		t.Errorf("expected 3 calls (1 initial + 2 retries), got %d", got)
	}
}

func TestNotify_RetriesThenSuccess(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		n := callCount.Add(1)
		if n < 3 {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("error")),
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})

	w, err := webhookslacker.New(
		webhookslacker.WithHTTPClient(client),
		webhookslacker.WithLogger(slog.Default()),
		webhookslacker.WithMaxRetries(3),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := w.Notify(context.Background(), "https://hooks.slack.com/test", "msg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := callCount.Load(); got != 3 {
		t.Errorf("expected 3 calls (2 failures + 1 success), got %d", got)
	}
}

func TestNotify_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		t.Fatal("should not be called when context is cancelled")

		return nil, nil
	})

	w, err := webhookslacker.New(
		webhookslacker.WithHTTPClient(client),
		webhookslacker.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := w.Notify(ctx, "https://hooks.slack.com/test", "msg"); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestNew_WithoutLogger_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := webhookslacker.New(); err == nil {
		t.Fatal("expected error when no logger provided")
	}
}

func TestNew_DefaultHTTPClient(t *testing.T) {
	t.Parallel()

	w, err := webhookslacker.New(webhookslacker.WithLogger(slog.Default()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w == nil {
		t.Fatal("expected non-nil Webhook instance")
	}
}

func TestNotifyAsync(t *testing.T) {
	t.Parallel()

	var called atomic.Bool

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		called.Store(true)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("ok")),
		}, nil
	})

	w, err := webhookslacker.New(
		webhookslacker.WithHTTPClient(client),
		webhookslacker.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w.NotifyAsync(context.Background(), "https://hooks.slack.com/test", "async msg")
	time.Sleep(100 * time.Millisecond)

	if !called.Load() {
		t.Error("expected HTTP request to be made")
	}
}

func TestNotify_NetworkError_Retries(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int32

	client := mockClient(func(_ *http.Request) (*http.Response, error) {
		callCount.Add(1)

		return nil, io.ErrUnexpectedEOF
	})

	w, err := webhookslacker.New(
		webhookslacker.WithHTTPClient(client),
		webhookslacker.WithLogger(slog.Default()),
		webhookslacker.WithMaxRetries(1),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := w.Notify(context.Background(), "https://hooks.slack.com/test", "msg"); err == nil {
		t.Fatal("expected error after network failure")
	}

	if got := callCount.Load(); got != 2 {
		t.Errorf("expected 2 calls (1 initial + 1 retry), got %d", got)
	}
}

func TestWithMaxRetries_Negative_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := webhookslacker.New(
		webhookslacker.WithLogger(slog.Default()),
		webhookslacker.WithMaxRetries(-1),
	); err == nil {
		t.Fatal("expected error for negative maxRetries")
	}
}

func TestWithMaxRetries_ExceedsMax_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := webhookslacker.New(
		webhookslacker.WithLogger(slog.Default()),
		webhookslacker.WithMaxRetries(11),
	); err == nil {
		t.Fatal("expected error for maxRetries > 10")
	}
}

func TestWithHTTPClient_Nil_ReturnsError(t *testing.T) {
	t.Parallel()

	if _, err := webhookslacker.New(
		webhookslacker.WithLogger(slog.Default()),
		webhookslacker.WithHTTPClient(nil),
	); err == nil {
		t.Fatal("expected error for nil HTTP client")
	}
}
