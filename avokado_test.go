package avokado_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bilustek/avokado"
	"github.com/gofiber/fiber/v3"
)

func TestNew_Defaults(t *testing.T) {
	t.Parallel()

	server, err := avokado.New()
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}

	if server.App == nil {
		t.Fatal("expected non-nil *fiber.App")
	}
}

func TestNew_WithServerName(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(avokado.WithServerName("test-server"))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithListenAddr(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(avokado.WithListenAddr(":9090"))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithLogger(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	server, err := avokado.New(avokado.WithLogger(logger))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithServerEnvironmentName(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(avokado.WithServerEnvironmentName("production"))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithServerAPIPrefix(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(avokado.WithServerAPIPrefix("/api/v2"))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithServerVersion(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(avokado.WithServerVersion("1.2.3"))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithFiberConfig(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(avokado.WithFiberConfig(&fiber.Config{
		CaseSensitive: true,
	}))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithListenConfig(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(avokado.WithListenConfig(&fiber.ListenConfig{
		DisableStartupMessage: true,
	}))
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithValidTimeouts(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(
		avokado.WithIdleTimeout(30*time.Second),
		avokado.WithReadTimeout(10*time.Second),
		avokado.WithWriteTimeout(20*time.Second),
	)
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestNew_WithNegativeIdleTimeout(t *testing.T) {
	t.Parallel()

	_, err := avokado.New(avokado.WithIdleTimeout(-1 * time.Second))
	if err == nil {
		t.Fatal("expected error for negative idle timeout")
	}
}

func TestNew_WithNegativeReadTimeout(t *testing.T) {
	t.Parallel()

	_, err := avokado.New(avokado.WithReadTimeout(-1 * time.Second))
	if err == nil {
		t.Fatal("expected error for negative read timeout")
	}
}

func TestNew_WithNegativeWriteTimeout(t *testing.T) {
	t.Parallel()

	_, err := avokado.New(avokado.WithWriteTimeout(-1 * time.Second))
	if err == nil {
		t.Fatal("expected error for negative write timeout")
	}
}

func TestNew_WithAllOptions(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	server, err := avokado.New(
		avokado.WithServerName("full-test"),
		avokado.WithListenAddr(":8888"),
		avokado.WithServerEnvironmentName("staging"),
		avokado.WithServerAPIPrefix("/api/v3"),
		avokado.WithServerVersion("2.0.0"),
		avokado.WithHealthzURL("/health"),
		avokado.WithLogger(logger),
		avokado.WithIdleTimeout(20*time.Second),
		avokado.WithReadTimeout(10*time.Second),
		avokado.WithWriteTimeout(15*time.Second),
		avokado.WithFiberConfig(&fiber.Config{}),
		avokado.WithListenConfig(&fiber.ListenConfig{}),
	)
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	if server == nil {
		t.Fatal("expected non-nil *Server")
	}
}

func TestHealthzEndpoint_DefaultURL(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(
		avokado.WithServerName("healthz-test"),
		avokado.WithServerVersion("1.0.0"),
		avokado.WithLogger(slog.New(slog.NewJSONHandler(io.Discard, nil))),
	)
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	req := httptest.NewRequest("GET", "/healthz", nil)
	resp, err := server.App.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("json decode error: %v", err)
	}

	if body["name"] != "healthz-test" {
		t.Errorf("expected name 'healthz-test', got %q", body["name"])
	}

	if body["version"] != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", body["version"])
	}
}

func TestHealthzEndpoint_CustomURL(t *testing.T) {
	t.Parallel()

	server, err := avokado.New(
		avokado.WithServerName("custom-healthz"),
		avokado.WithHealthzURL("/health"),
		avokado.WithLogger(slog.New(slog.NewJSONHandler(io.Discard, nil))),
	)
	if err != nil {
		t.Fatalf("avokado.New error: %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := server.App.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("json decode error: %v", err)
	}

	if body["name"] != "custom-healthz" {
		t.Errorf("expected name 'custom-healthz', got %q", body["name"])
	}
}
