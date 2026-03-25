package avomiddleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bilustek/avokado/avomiddleware"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
)

func TestNewSentry_SetsHubOnContext(t *testing.T) {
	t.Parallel()

	_ = sentry.Init(sentry.ClientOptions{Dsn: ""})

	app := fiber.New()
	app.Use(avomiddleware.NewSentry())

	var hubFound bool

	app.Get("/test", func(c fiber.Ctx) error {
		hub := sentry.GetHubFromContext(c.Context())
		hubFound = hub != nil

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if !hubFound {
		t.Error("expected Sentry hub on context, got nil")
	}
}

func TestNewSentry_RecoverFromPanic(t *testing.T) {
	t.Parallel()

	_ = sentry.Init(sentry.ClientOptions{Dsn: ""})

	app := fiber.New()
	app.Use(avomiddleware.NewSentry())
	app.Get("/panic", func(_ fiber.Ctx) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		t.Error("expected non-200 status after panic, got 200")
	}
}

func TestNewSentry_PanicReturnsSentryHandledError(t *testing.T) {
	t.Parallel()

	_ = sentry.Init(sentry.ClientOptions{Dsn: ""})

	app := fiber.New(fiber.Config{
		ErrorHandler: func(_ fiber.Ctx, err error) error {
			if !avomiddleware.IsSentryHandled(err) {
				t.Error("expected error to be sentry-handled, got regular error")
			}

			// Verify we can unwrap to get the original error message.
			inner := errors.Unwrap(err)
			if inner == nil {
				t.Error("expected unwrappable error")
			}

			return nil
		},
	})
	app.Use(avomiddleware.NewSentry())
	app.Get("/panic", func(_ fiber.Ctx) error {
		panic("sentry test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
}

func TestNewSentry_HubInLocals(t *testing.T) {
	t.Parallel()

	_ = sentry.Init(sentry.ClientOptions{Dsn: ""})

	app := fiber.New()
	app.Use(avomiddleware.NewSentry())

	var hubFound bool

	app.Get("/test", func(c fiber.Ctx) error {
		val := c.Locals(avomiddleware.LocalsSentryHub)
		_, ok := val.(*sentry.Hub)
		hubFound = ok

		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if !hubFound {
		t.Error("expected Sentry hub in Locals, got nil")
	}
}
