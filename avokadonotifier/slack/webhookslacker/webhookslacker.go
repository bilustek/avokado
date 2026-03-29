package webhookslacker

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/bilustek/avokado/avokadonotifier"
)

const (
	defaultMaxRetries = 5
	maxAllowedRetries = 10
)

// clientError represents a non-retryable HTTP client error (4xx).
type clientError struct {
	err error
}

func (e clientError) Error() string { return e.err.Error() }
func (e clientError) Unwrap() error { return e.err }

// compile-time proof of interface implementation.
var _ avokadonotifier.SlackNotifier = (*Webhook)(nil)

// Option is a functional option for configuring the webhookslacker.
type Option func(*Webhook) error

// Webhook is a SlackNotifier that delivers messages via Slack incoming webhooks.
type Webhook struct {
	client     *http.Client
	logger     *slog.Logger
	webhookURL string
	maxRetries int
}

// Notify sends a message to the configured Slack webhook URL with retry and backoff.
func (w *Webhook) Notify(ctx context.Context, message string) error {
	payload := `{"text":` + strconv.Quote(message) + `}`

	var lastErr error

	for attempt := range w.maxRetries + 1 {
		if err := ctx.Err(); err != nil {
			return avokadoerror.New("[Webhook.Notify] context cancelled").WithErr(err)
		}

		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			timer := time.NewTimer(backoff)

			select {
			case <-ctx.Done():
				timer.Stop()

				return avokadoerror.New("[Webhook.Notify] context cancelled during backoff").WithErr(ctx.Err())
			case <-timer.C:
			}
		}

		if lastErr = w.doRequest(ctx, payload); lastErr == nil {
			w.logger.InfoContext(ctx, "[Webhook.Notify] message sent")

			return nil
		}

		w.logger.WarnContext(ctx, "[Webhook.Notify] attempt failed",
			"attempt", attempt+1,
			"maxRetries", w.maxRetries,
			"error", lastErr,
		)

		var ce clientError
		if errors.As(lastErr, &ce) {
			break
		}
	}

	w.logger.ErrorContext(ctx, "[Webhook.Notify] all attempts failed", "error", lastErr)

	return lastErr
}

// NotifyAsync sends a message in a background goroutine, logging is handled by Notify.
func (w *Webhook) NotifyAsync(ctx context.Context, message string) {
	go func() {
		_ = w.Notify(context.WithoutCancel(ctx), message)
	}()
}

func (w *Webhook) doRequest(ctx context.Context, payload string) error {
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, w.webhookURL, bytes.NewBufferString(payload),
	)
	if err != nil {
		return avokadoerror.New("[Webhook.doRequest] create request err").WithErr(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return avokadoerror.New("[Webhook.doRequest] err").WithErr(err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= http.StatusInternalServerError {
		return avokadoerror.New("[Webhook.doRequest] server error: status " + strconv.Itoa(resp.StatusCode))
	}

	if resp.StatusCode >= http.StatusBadRequest {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return avokadoerror.New("[Webhook.doRequest] read response body err").WithErr(readErr)
		}

		return clientError{
			err: avokadoerror.New(
				"[Webhook.doRequest] client error: status " + strconv.Itoa(resp.StatusCode) + " body: " + string(body),
			),
		}
	}

	return nil
}

// WithWebhookURL sets the Slack webhook URL.
func WithWebhookURL(url string) Option {
	return func(w *Webhook) error {
		if url == "" {
			return avokadoerror.New("[webhookslacker.WithWebhookURL] webhookURL must not be empty")
		}

		w.webhookURL = url

		return nil
	}
}

// WithHTTPClient sets a custom HTTP client (useful for testing with mock RoundTripper).
func WithHTTPClient(client *http.Client) Option {
	return func(w *Webhook) error {
		if client == nil {
			return avokadoerror.New("[webhookslacker.WithHTTPClient] client must not be nil")
		}

		w.client = client

		return nil
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(w *Webhook) error {
		w.logger = logger

		return nil
	}
}

// WithMaxRetries sets the maximum number of retries on failure. Must be between 0 and maxAllowedRetries.
func WithMaxRetries(n int) Option {
	return func(w *Webhook) error {
		if n < 0 || n > maxAllowedRetries {
			return avokadoerror.New(
				"[webhookslacker.WithMaxRetries] maxRetries must be between 0 and " + strconv.Itoa(maxAllowedRetries),
			)
		}

		w.maxRetries = n

		return nil
	}
}

// New creates a Webhook slack notifier with the given options.
func New(opts ...Option) (*Webhook, error) {
	cfg := &Webhook{
		client:     &http.Client{Timeout: 10 * time.Second},
		maxRetries: defaultMaxRetries,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.logger == nil {
		return nil, avokadoerror.New("[webhookslacker.New] logger is required, use WithLogger")
	}
	if cfg.webhookURL == "" {
		return nil, avokadoerror.New("[webhookslacker.New] webhookURL is required, use WithWebhookURL")
	}

	return cfg, nil
}
