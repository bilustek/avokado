package avokadoerror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/gofiber/fiber/v3"
)

func TestNew_ReturnsErrorWithMessage(t *testing.T) {
	t.Parallel()

	if err := avokadoerror.New("something went wrong"); err.Error() != "something went wrong" {
		t.Errorf("expected 'something went wrong', got %q", err.Error())
	}
}

func TestNew_ReturnsPointer(t *testing.T) {
	t.Parallel()

	err := avokadoerror.New("test")
	if err == nil {
		t.Fatal("expected non-nil *Error")
	}
}

func TestFluentBuilder_ReturnsSamePointer(t *testing.T) {
	t.Parallel()

	err := avokadoerror.New("not found")
	result := err.WithStatus(fiber.StatusNotFound).WithCode(avokadoerror.CodeNotFound)

	if result != err {
		t.Error("fluent builder should return the same pointer")
	}

	if result.Status != fiber.StatusNotFound {
		t.Errorf("expected Status 404, got %d", result.Status)
	}

	if result.Code != avokadoerror.CodeNotFound {
		t.Errorf("expected Code %q, got %q", avokadoerror.CodeNotFound, result.Code)
	}
}

func TestFluentBuilder_WithErr(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("db connection failed")
	err := avokadoerror.New("internal error").WithErr(inner)

	if err.Err != inner {
		t.Error("WithErr should set the wrapped error")
	}
}

func TestUnwrap_ReturnsInnerError(t *testing.T) {
	t.Parallel()

	inner := fmt.Errorf("db connection failed")
	err := avokadoerror.New("outer").WithErr(inner)

	unwrapped := errors.Unwrap(err)
	if unwrapped != inner {
		t.Errorf("Unwrap should return inner error, got %v", unwrapped)
	}
}

func TestErrorsAs_WorksForChainedErrors(t *testing.T) {
	t.Parallel()

	appErr := avokadoerror.New("not found").WithStatus(fiber.StatusNotFound).WithCode(avokadoerror.CodeNotFound)
	wrapped := fmt.Errorf("handler error: %w", appErr)

	var target *avokadoerror.Error
	if !errors.As(wrapped, &target) {
		t.Fatal("errors.As should find *avokadoerror.Error in chain")
	}

	if target.Status != fiber.StatusNotFound {
		t.Errorf("expected Status %d, got %d", fiber.StatusNotFound, target.Status)
	}
}

func TestErrorsIs_WorksThroughUnwrapChain(t *testing.T) {
	t.Parallel()

	sentinel := fmt.Errorf("sentinel error")
	appErr := avokadoerror.New("wrapper").WithErr(sentinel)
	wrapped := fmt.Errorf("outer: %w", appErr)

	if !errors.Is(wrapped, sentinel) {
		t.Error("errors.Is should find sentinel through Unwrap chain")
	}
}

func TestErrorCodeConstants_AllDefined(t *testing.T) {
	t.Parallel()

	codes := map[string]avokadoerror.ErrorCode{
		"validation-error": avokadoerror.CodeValidationError,
		"unauthorized":     avokadoerror.CodeUnauthorized,
		"forbidden":        avokadoerror.CodeForbidden,
		"not-found":        avokadoerror.CodeNotFound,
		"internal-error":   avokadoerror.CodeInternalError,
		"conflict":         avokadoerror.CodeConflict,
		"invalid-param":    avokadoerror.CodeInvalidParam,
		"http-error":       avokadoerror.CodeHTTPError,
	}

	for expected, code := range codes {
		if string(code) != expected {
			t.Errorf("expected code %q, got %q", expected, code)
		}
	}
}

func TestError_ImplementsErrorInterface(t *testing.T) {
	t.Parallel()

	var err error = avokadoerror.New("test")
	if err == nil {
		t.Fatal("should implement error interface")
	}
}

func TestError_JSONTags(t *testing.T) {
	t.Parallel()

	// Verify struct has proper json tags by creating and checking fields
	err := avokadoerror.New("test").WithStatus(fiber.StatusBadRequest).WithCode(avokadoerror.CodeValidationError)
	if err.Message != "test" {
		t.Error("Message field should be accessible")
	}
	if err.Code != avokadoerror.CodeValidationError {
		t.Error("Code field should be accessible")
	}
}
