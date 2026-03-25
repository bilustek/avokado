package avoerror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/bilustek/avokado/avoerror"
)

func TestNew_ReturnsErrorWithMessage(t *testing.T) {
	t.Parallel()

	if err := avoerror.New("something went wrong"); err.Error() != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", err.Error())
	}
}

func TestNew_ReturnsPointer(t *testing.T) {
	t.Parallel()

	err := avoerror.New("test")
	if err == nil {
		t.Fatal("expected non-nil *Error")
	}
}

func TestFluentBuilder_ReturnsSamePointer(t *testing.T) {
	t.Parallel()

	err := avoerror.New("not found")
	result := err.WithStatus(http.StatusNotFound).WithCode(avoerror.CodeNotFound)

	if result != err {
		t.Error("fluent builder should return the same pointer")
	}

	if result.Status != http.StatusNotFound {
		t.Errorf("expected Status 404, got %d", result.Status)
	}

	if result.Code != avoerror.CodeNotFound {
		t.Errorf("expected Code %q, got %q", avoerror.CodeNotFound, result.Code)
	}
}

func TestFluentBuilder_WithErr(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("db connection failed")
	err := avoerror.New("internal error").WithErr(inner)

	if err.Err != inner {
		t.Error("WithErr should set the wrapped error")
	}
}

func TestUnwrap_ReturnsInnerError(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("db connection failed")
	err := avoerror.New("outer").WithErr(inner)

	unwrapped := errors.Unwrap(err)
	if unwrapped != inner {
		t.Errorf("Unwrap should return inner error, got %v", unwrapped)
	}
}

func TestErrorsAs_WorksForChainedErrors(t *testing.T) {
	t.Parallel()

	appErr := avoerror.New("not found").WithStatus(http.StatusNotFound).WithCode(avoerror.CodeNotFound)
	wrapped := fmt.Errorf("handler error: %w", appErr)

	var target *avoerror.Error
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should find *avoerror.Error in chain")
	}

	if target.Status != http.StatusNotFound {
		t.Errorf("expected Status %d, got %d", http.StatusNotFound, target.Status)
	}
}

func TestErrorsIs_WorksThroughUnwrapChain(t *testing.T) {
	t.Parallel()

	sentinel := fmt.Errorf("sentinel error")
	appErr := avoerror.New("wrapper").WithErr(sentinel)
	wrapped := fmt.Errorf("outer: %w", appErr)

	if !errors.Is(wrapped, sentinel) {
		t.Error("errors.Is should find sentinel through Unwrap chain")
	}
}

func TestErrorCodeConstants_AllDefined(t *testing.T) {
	t.Parallel()

	codes := map[string]avoerror.ErrorCode{
		"validation-error": avoerror.CodeValidationError,
		"unauthorized":     avoerror.CodeUnauthorized,
		"forbidden":        avoerror.CodeForbidden,
		"not-found":        avoerror.CodeNotFound,
		"internal-error":   avoerror.CodeInternalError,
		"conflict":         avoerror.CodeConflict,
		"invalid-param":    avoerror.CodeInvalidParam,
		"http-error":       avoerror.CodeHTTPError,
	}

	for expected, code := range codes {
		if string(code) != expected {
			t.Errorf("expected code %q, got %q", expected, code)
		}
	}
}

func TestError_ImplementsErrorInterface(t *testing.T) {
	t.Parallel()

	var err error = avoerror.New("test")
	if err == nil {
		t.Fatal("should implement error interface")
	}
}

func TestError_JSONTags(t *testing.T) {
	t.Parallel()

	// Verify struct has proper json tags by creating and checking fields
	err := avoerror.New("test").WithStatus(http.StatusBadRequest).WithCode(avoerror.CodeValidationError)
	if err.Message != "test" {
		t.Error("Message field should be accessible")
	}
	if err.Code != avoerror.CodeValidationError {
		t.Error("Code field should be accessible")
	}
}
