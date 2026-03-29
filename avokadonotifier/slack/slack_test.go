package slack_test

import (
	"testing"

	"github.com/bilustek/avokado/avokadonotifier/slack"
)

func TestNew_DefaultIsDevelopment(t *testing.T) {
	t.Parallel()

	n, err := slack.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

func TestNew_DevelopmentExplicit(t *testing.T) {
	t.Parallel()

	n, err := slack.New(slack.WithServerEnvironmentName("development"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

func TestNew_ProductionNotYetSupported(t *testing.T) {
	t.Parallel()

	_, err := slack.New(slack.WithServerEnvironmentName("production"))
	if err == nil {
		t.Fatal("expected error for non-development environment")
	}
}

func TestNew_EmptyServerEnvironmentName(t *testing.T) {
	t.Parallel()

	_, err := slack.New(slack.WithServerEnvironmentName(""))
	if err == nil {
		t.Fatal("expected error when serverEnvironmentName is empty")
	}
}
