package avokadoerror

// ErrorCode represents a typed error code string for API error responses.
type ErrorCode string

// Predefined error codes for common API error scenarios.
const (
	CodeValidationError ErrorCode = "validation-error"
	CodeUnauthorized    ErrorCode = "unauthorized"
	CodeForbidden       ErrorCode = "forbidden"
	CodeNotFound        ErrorCode = "not-found"
	CodeInternalError   ErrorCode = "internal-error"
	CodeConflict        ErrorCode = "conflict"
	CodeInvalidParam    ErrorCode = "invalid-param"
	CodeHTTPError       ErrorCode = "http-error"
	CodeDatabaseError   ErrorCode = "database-error"
)

// Error is the core error type for avokado. It supports fluent builder chaining
// and implements the standard error interface with Unwrap support.
type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Status  int       `json:"-"`
	Err     error     `json:"-"`
}

// New creates a new Error with the given message.
func New(message string) *Error {
	return &Error{Message: message}
}

// WithStatus sets the HTTP status code and returns the same pointer for chaining.
func (e *Error) WithStatus(status int) *Error {
	e.Status = status

	return e
}

// WithCode sets the error code and returns the same pointer for chaining.
func (e *Error) WithCode(code ErrorCode) *Error {
	e.Code = code

	return e
}

// WithErr wraps an inner error and returns the same pointer for chaining.
func (e *Error) WithErr(err error) *Error {
	e.Err = err

	return e
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}

	return e.Message
}

// Unwrap returns the wrapped error for errors.Is/As/Unwrap compatibility.
func (e *Error) Unwrap() error {
	return e.Err
}
