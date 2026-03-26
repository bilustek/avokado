package avokadomiddleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v3"
)

const sentryFlushTimeout = 2 * time.Second

type sentryHandledError struct {
	inner   error
	eventID *sentry.EventID
}

// Error returns the inner error message.
func (e *sentryHandledError) Error() string {
	return e.inner.Error()
}

// Unwrap returns the inner error for errors.Is/As/Unwrap support.
func (e *sentryHandledError) Unwrap() error {
	return e.inner
}

// IsSentryHandled reports whether the given error was already captured by
// the Sentry middleware panic recovery. This can be used by error handlers
// and loggers to prevent duplicate Sentry reports.
func IsSentryHandled(err error) bool {
	var handled *sentryHandledError

	return errors.As(err, &handled)
}

// NewSentry creates a Fiber middleware that clones the current Sentry hub,
// sets it on the request context and Locals, and recovers from panics.
//
// On panic, the middleware:
//  1. Captures the panic via hub.RecoverWithContext
//  2. Flushes the hub (2s timeout)
//  3. Returns a sentryHandledError wrapping the panic value
//
// The sentryHandledError prevents avokadologger's SentryHandler from
// reporting the same error again.
func NewSentry() fiber.Handler {
	return func(c fiber.Ctx) (returnErr error) {
		hub := sentry.CurrentHub().Clone()
		ctx := sentry.SetHubOnContext(c.Context(), hub)
		c.SetContext(ctx)
		c.Locals(LocalsSentryHub, hub)

		defer func() {
			if r := recover(); r != nil {
				eventID := hub.RecoverWithContext(c.Context(), r)
				hub.Flush(sentryFlushTimeout)

				panicErr, ok := r.(error)
				if !ok {
					panicErr = fmt.Errorf("%v", r)
				}

				returnErr = &sentryHandledError{
					inner:   panicErr,
					eventID: eventID,
				}
			}
		}()

		return c.Next()
	}
}
