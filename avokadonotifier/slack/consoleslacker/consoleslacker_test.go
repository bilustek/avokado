package consoleslacker_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bilustek/avokado/avokadonotifier/slack/consoleslacker"
)

func TestConsoleNotify(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	c := consoleslacker.New(consoleslacker.WithWriter(&buf))

	err := c.Notify(context.Background(), "https://hooks.slack.com/services/xxx", "deploy completed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Webhook: https://hooks.slack.com/services/xxx") {
		t.Error("expected output to contain webhook URL")
	}
	if !strings.Contains(output, "Message: deploy completed") {
		t.Error("expected output to contain message")
	}

	separator := strings.Repeat("-", 72)
	if strings.Count(output, separator) != 2 {
		t.Errorf("expected 2 separators, got %d", strings.Count(output, separator))
	}
}
