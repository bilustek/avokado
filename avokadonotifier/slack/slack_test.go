package slack_test

import (
	"log/slog"
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

func TestNew_ProductionWithLogger(t *testing.T) {
	t.Parallel()

	n, err := slack.New(
		slack.WithServerEnvironmentName("production"),
		slack.WithLogger(slog.Default()),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == nil {
		t.Fatal("expected non-nil Notifier")
	}
}

func TestNew_ProductionWithoutLogger(t *testing.T) {
	t.Parallel()

	if _, err := slack.New(slack.WithServerEnvironmentName("production")); err == nil {
		t.Fatal("expected error when no logger provided for production")
	}
}

func TestNew_EmptyServerEnvironmentName(t *testing.T) {
	t.Parallel()

	if _, err := slack.New(slack.WithServerEnvironmentName("")); err == nil {
		t.Fatal("expected error when serverEnvironmentName is empty")
	}
}
