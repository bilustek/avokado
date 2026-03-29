package resendmailer_test

import (
	"testing"

	"github.com/bilustek/avokado/avokadonotifier"
	"github.com/bilustek/avokado/avokadonotifier/email/resendmailer"
)

func TestNew_WithoutAPIKey_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := resendmailer.New()
	if err == nil {
		t.Fatal("expected error when no API key provided")
	}
}

func TestNew_WithAPIKey_Succeeds(t *testing.T) {
	t.Parallel()

	r, err := resendmailer.New(resendmailer.WithAPIKey("re_test_123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil Resend instance")
	}
}

func TestResendImplementsEmailSender(t *testing.T) {
	t.Parallel()

	r, err := resendmailer.New(resendmailer.WithAPIKey("re_test_123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var _ avokadonotifier.EmailSender = r
}
