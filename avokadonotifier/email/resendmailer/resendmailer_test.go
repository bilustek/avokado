package resendmailer_test

import (
	"log/slog"
	"testing"

	"github.com/bilustek/avokado/avokadonotifier/email/resendmailer"
)

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
