package email_test

import (
	"testing"

	"github.com/bilustek/avokado/avokadonotifier/email"
)

func TestNew_DefaultIsDevelopment(t *testing.T) {
	t.Parallel()

	n, err := email.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

func TestNew_DevelopmentExplicit(t *testing.T) {
	t.Parallel()

	n, err := email.New(email.WithServerEnvironmentName("development"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

func TestNew_ProductionWithAPIKey(t *testing.T) {
	t.Parallel()

	n, err := email.New(
		email.WithServerEnvironmentName("production"),
		email.WithResendAPIKey("re_test_123"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

func TestNew_ProductionWithoutAPIKey(t *testing.T) {
	t.Parallel()

	_, err := email.New(email.WithServerEnvironmentName("production"))
	if err == nil {
		t.Fatal("expected error when no API key provided for production")
	}
}

func TestNew_StagingWithAPIKey(t *testing.T) {
	t.Parallel()

	n, err := email.New(
		email.WithServerEnvironmentName("staging"),
		email.WithResendAPIKey("re_test_456"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

func TestNew_EmptyServerEnvironmentName(t *testing.T) {
	t.Parallel()

	_, err := email.New(email.WithServerEnvironmentName(""))
	if err == nil {
		t.Fatal("expected error when serverEnvironmentName is empty")
	}
}
