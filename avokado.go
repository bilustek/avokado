package avokado

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bilustek/avokado/avoerror"
	"github.com/gofiber/fiber/v3"
)

// Server represents the fiber app and configuration values.
type Server struct {
	App    *fiber.App
	config *config
}

// ListenAndServe starts fiber server with grafeful shutdown.
func (s *Server) ListenAndServe(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		s.config.logger.Info(
			"start listening at",
			"addr", s.config.listenAddr,
			"name", s.config.serverName,
			"version", s.config.serverVersion,
		)
		if s.config.listenConfig != nil {
			errCh <- s.App.Listen(s.config.listenAddr, *s.config.listenConfig)
		} else {
			errCh <- s.App.Listen(s.config.listenAddr)
		}
	}()

	select {
	case <-ctx.Done():
		return s.App.ShutdownWithContext(context.Background())
	case err := <-errCh:
		return err
	}
}

const (
	defaultIdleTimeout  = 15 * time.Second
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

// Option is a functional option for configuring the avokado builder.
type Option func(*config) error

type config struct {
	logger                *slog.Logger
	serverName            string
	listenAddr            string
	serverEnvironmentName string
	serverAPIprefix       string
	serverVersion         string
	healthzURL            string
	idleTimeout           time.Duration
	readTimeout           time.Duration
	writeTimeout          time.Duration
	fiberConfig           *fiber.Config
	listenConfig          *fiber.ListenConfig
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *config) error {
		c.logger = logger

		return nil
	}
}

// WithServerName sets the api server name.
func WithServerName(serverName string) Option {
	return func(c *config) error {
		c.serverName = serverName

		return nil
	}
}

// WithListenAddr sets the api server's listen addr.
func WithListenAddr(addr string) Option {
	return func(c *config) error {
		c.listenAddr = addr

		return nil
	}
}

// WithServerEnvironmentName sets the api server environment name.
func WithServerEnvironmentName(serverEnvironmentName string) Option {
	return func(c *config) error {
		c.serverEnvironmentName = serverEnvironmentName

		return nil
	}
}

// WithServerAPIPrefix sets the api prefix.
func WithServerAPIPrefix(serverAPIprefix string) Option {
	return func(c *config) error {
		c.serverAPIprefix = serverAPIprefix

		return nil
	}
}

// WithServerVersion sets the api server version.
func WithServerVersion(serverVersion string) Option {
	return func(c *config) error {
		c.serverVersion = serverVersion

		return nil
	}
}

// WithHealthzURL sets the healthz endpoint url.
func WithHealthzURL(healthzURL string) Option {
	return func(c *config) error {
		c.healthzURL = healthzURL

		return nil
	}
}

// WithFiberConfig sets the custom fiber config.
func WithFiberConfig(fiberConfig *fiber.Config) Option {
	return func(c *config) error {
		c.fiberConfig = fiberConfig

		return nil
	}
}

// WithListenConfig sets the fiber server's listen configuration.
func WithListenConfig(listenConfig *fiber.ListenConfig) Option {
	return func(c *config) error {
		c.listenConfig = listenConfig

		return nil
	}
}

// WithIdleTimeout sets the fiber server's idle timeout.
func WithIdleTimeout(d time.Duration) Option {
	return func(c *config) error {
		if d < 0 {
			return fmt.Errorf(
				"%w '%s' received, must > 0",
				avoerror.New("[avokado.WithIdleTimeout] err:").WithCode(avoerror.CodeInvalidParam),
				d,
			)
		}
		c.idleTimeout = d

		return nil
	}
}

// WithReadTimeout sets the fiber server's read timeout.
func WithReadTimeout(d time.Duration) Option {
	return func(c *config) error {
		if d < 0 {
			return fmt.Errorf(
				"%w, '%s' received, must > 0",
				avoerror.New("[avokado.WithReadTimeout] err:").WithCode(avoerror.CodeInvalidParam),
				d,
			)
		}

		c.readTimeout = d

		return nil
	}
}

// WithWriteTimeout sets the fiber server's write timeout.
func WithWriteTimeout(d time.Duration) Option {
	return func(c *config) error {
		if d < 0 {
			return fmt.Errorf(
				"%w, '%s' received, must > 0",
				avoerror.New("[avokado.WithWriteTimeout] err:").WithCode(avoerror.CodeInvalidParam),
				d,
			)
		}

		c.writeTimeout = d

		return nil
	}
}

// New creates a new Fiber v3 application.
func New(opts ...Option) (*Server, error) {
	cfg := &config{}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.serverName == "" {
		cfg.serverName = "avokado"
	}
	if cfg.listenAddr == "" {
		cfg.listenAddr = ":7001"
	}
	if cfg.serverEnvironmentName == "" {
		cfg.serverEnvironmentName = "development"
	}
	if cfg.serverAPIprefix == "" {
		cfg.serverAPIprefix = "/api/v1"
	}
	if cfg.serverVersion == "" {
		cfg.serverVersion = "0.0.0"
	}
	if cfg.healthzURL == "" {
		cfg.healthzURL = "/healthz"
	}
	if cfg.idleTimeout == 0 {
		cfg.idleTimeout = defaultIdleTimeout
	}
	if cfg.readTimeout == 0 {
		cfg.readTimeout = defaultReadTimeout
	}
	if cfg.writeTimeout == 0 {
		cfg.writeTimeout = defaultWriteTimeout
	}
	if cfg.logger == nil {
		cfg.logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
	}
	if cfg.fiberConfig == nil {
		cfg.fiberConfig = &fiber.Config{}
	}

	fiberCfg := *cfg.fiberConfig
	fiberCfg.AppName = cfg.serverName

	app := fiber.New(fiberCfg)

	healthzHandlerArgs := &healthzHTTPHandlerArgs{}
	healthzHandlerArgs.baseArgs.logger = cfg.logger
	healthzHandlerArgs.baseArgs.serverEnvironmentName = cfg.serverEnvironmentName
	healthzHandlerArgs.baseArgs.serverVersion = cfg.serverVersion
	healthzHandlerArgs.serverName = cfg.serverName

	app.Get(cfg.healthzURL, healthzHandler(healthzHandlerArgs))

	return &Server{App: app, config: cfg}, nil
}

type baseHTTPHandlerArgs struct {
	logger                *slog.Logger
	serverEnvironmentName string
	serverVersion         string
}

type healthzHTTPHandlerArgs struct {
	baseArgs   baseHTTPHandlerArgs
	serverName string
}

func healthzHandler(args *healthzHTTPHandlerArgs) fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    args.serverName,
			"version": args.baseArgs.serverVersion,
		})
	}
}
