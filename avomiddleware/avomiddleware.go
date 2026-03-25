// Package avomiddleware provides HTTP middleware for the avokado framework.
// It includes security headers, Sentry panic recovery, CORS, and client-type
// detection middleware. All middleware follows the Fiber v3 Handler signature.
package avomiddleware

// Locals keys for storing middleware data in Fiber context.
const (
	// LocalsClientType is the Fiber Locals key for the client type string.
	LocalsClientType = "client_type"

	// LocalsSentryHub is the Fiber Locals key for the cloned Sentry hub.
	LocalsSentryHub = "sentry_hub"

	// LocalsUserClaims is the Fiber Locals key for the authenticated user's JWT claims.
	LocalsUserClaims = "user_claims"
)
