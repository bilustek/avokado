package avokadodb

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/bilustek/avokado/avokadoerror"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	defaultMaxOpenConns           = 25
	defaultMaxIdleConns           = 25
	defaultConnMaxLifetimeMinutes = 5
	defaultConnMaxIdleTimeMinutes = 15
	defaultPingTimeoutSeconds     = 5

	minMaxOpenConns    = 1
	maxMaxOpenConns    = 100
	minMaxIdleConns    = 1
	maxMaxIdleConns    = 100
	minConnMaxLifetime = 1 * time.Minute
	maxConnMaxLifetime = 30 * time.Minute
	minConnMaxIdleTime = 1 * time.Minute
	maxConnMaxIdleTime = 30 * time.Minute
	minPingTimeout     = 1 * time.Second
	maxPingTimeout     = 30 * time.Second

	defaultGormLogLevel = gormlogger.Silent
)

type config struct {
	databaseURL     string
	maxOpenConns    int
	maxIdleConns    int
	gormLogLevel    gormlogger.LogLevel
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	pingTimeout     time.Duration
	logger          *slog.Logger
	gormConfig      *gorm.Config
}

// Option is a functional option for configuring the database connection.
type Option func(*config) error

// WithDatabaseURL sets the PostgreSQL connection URL.
func WithDatabaseURL(url string) Option {
	return func(c *config) error {
		if url == "" {
			return avokadoerror.New("[avokadodb.WithDatabaseURL] err: empty database url").
				WithCode(avokadoerror.CodeInvalidParam)
		}

		c.databaseURL = url

		return nil
	}
}

// WithMaxOpenConns sets the maximum number of open connections.
func WithMaxOpenConns(n int) Option {
	return func(c *config) error {
		if n < minMaxOpenConns {
			return fmt.Errorf(
				"%w, '%d' received, must > %d",
				avokadoerror.New("[avokadodb.WithMaxOpenConns] err:").WithCode(avokadoerror.CodeInvalidParam),
				n,
				minMaxOpenConns,
			)
		}
		if n > maxMaxOpenConns {
			return fmt.Errorf(
				"%w, '%d' received, must < %d",
				avokadoerror.New("[avokadodb.WithMaxOpenConns] err:").WithCode(avokadoerror.CodeInvalidParam),
				n,
				maxMaxOpenConns,
			)
		}

		c.maxOpenConns = n

		return nil
	}
}

// WithMaxIdleConns sets the maximum number of idle connections.
func WithMaxIdleConns(n int) Option {
	return func(c *config) error {
		if n < minMaxIdleConns {
			return fmt.Errorf(
				"%w, '%d' received, must > %d",
				avokadoerror.New("[avokadodb.WithMaxIdleConns] err:").WithCode(avokadoerror.CodeInvalidParam),
				n,
				minMaxIdleConns,
			)
		}
		if n > maxMaxIdleConns {
			return fmt.Errorf(
				"%w, '%d' received, must < %d",
				avokadoerror.New("[avokadodb.WithMaxIdleConns] err:").WithCode(avokadoerror.CodeInvalidParam),
				n,
				maxMaxIdleConns,
			)
		}

		c.maxIdleConns = n

		return nil
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection.
func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *config) error {
		if d < minConnMaxLifetime {
			return fmt.Errorf(
				"%w, '%s' received, must > %s",
				avokadoerror.New("[avokadodb.WithConnMaxLifetime] err:").WithCode(avokadoerror.CodeInvalidParam),
				d,
				minConnMaxLifetime,
			)
		}
		if d > maxConnMaxLifetime {
			return fmt.Errorf(
				"%w, '%s' received, must < %s",
				avokadoerror.New("[avokadodb.WithConnMaxLifetime] err:").WithCode(avokadoerror.CodeInvalidParam),
				d,
				maxConnMaxLifetime,
			)
		}

		c.connMaxLifetime = d

		return nil
	}
}

// WithConnMaxIdleTime sets the maximum idle time of a connection.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *config) error {
		if d < minConnMaxIdleTime {
			return fmt.Errorf(
				"%w, '%s' received, must > %s",
				avokadoerror.New("[avokadodb.WithConnMaxIdleTime] err:").WithCode(avokadoerror.CodeInvalidParam),
				d,
				minConnMaxIdleTime,
			)
		}
		if d > maxConnMaxIdleTime {
			return fmt.Errorf(
				"%w, '%s' received, must < %s",
				avokadoerror.New("[avokadodb.WithConnMaxIdleTime] err:").WithCode(avokadoerror.CodeInvalidParam),
				d,
				maxConnMaxIdleTime,
			)
		}

		c.connMaxIdleTime = d

		return nil
	}
}

// WithPingTimeout sets the ping timeout of a connection.
func WithPingTimeout(d time.Duration) Option {
	return func(c *config) error {
		if d < minPingTimeout {
			return fmt.Errorf(
				"%w, '%s' received, must > %s",
				avokadoerror.New("[avokadodb.WithPingTimeout] err:").WithCode(avokadoerror.CodeInvalidParam),
				d,
				minPingTimeout,
			)
		}
		if d > maxPingTimeout {
			return fmt.Errorf(
				"%w, '%s' received, must < %s",
				avokadoerror.New("[avokadodb.WithPingTimeout] err:").WithCode(avokadoerror.CodeInvalidParam),
				d,
				maxPingTimeout,
			)
		}

		c.pingTimeout = d

		return nil
	}
}

// WithLogger sets the slog.Logger for GORM SQL logging.
func WithLogger(l *slog.Logger) Option {
	return func(c *config) error {
		c.logger = l

		return nil
	}
}

// WithGormConfig sets a custom GORM configuration. When provided, WithLogger and WithGormLogLevel
// are ignored unless the given config has a nil Logger, in which case the configured logger is used.
func WithGormConfig(gcf *gorm.Config) Option {
	return func(c *config) error {
		c.gormConfig = gcf

		return nil
	}
}

// WithGormLogLevel sets the GORM logger log level (Silent=1, Error=2, Warn=3, Info=4).
func WithGormLogLevel(n int) Option {
	return func(c *config) error {
		if n < int(gormlogger.Silent) {
			return fmt.Errorf(
				"%w, '%d' received, must > %d",
				avokadoerror.New("[avokadodb.WithGormLogLevel] err:").WithCode(avokadoerror.CodeInvalidParam),
				n,
				gormlogger.Silent,
			)
		}

		if n > int(gormlogger.Info) {
			return fmt.Errorf(
				"%w, '%d' received, must < %d",
				avokadoerror.New("[avokadodb.WithGormLogLevel] err:").WithCode(avokadoerror.CodeInvalidParam),
				n,
				gormlogger.Info,
			)
		}

		c.gormLogLevel = gormlogger.LogLevel(n)

		return nil
	}
}

// New creates a new GORM database connection with the given options.
func New(opts ...Option) (*gorm.DB, error) {
	cfg := &config{
		maxOpenConns:    defaultMaxOpenConns,
		maxIdleConns:    defaultMaxIdleConns,
		connMaxLifetime: defaultConnMaxLifetimeMinutes * time.Minute,
		connMaxIdleTime: defaultConnMaxIdleTimeMinutes * time.Minute,
		pingTimeout:     defaultPingTimeoutSeconds * time.Second,
		gormLogLevel:    defaultGormLogLevel,
	}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.databaseURL == "" {
		return nil, avokadoerror.New("[avokadodb.New] err: databaseURL required").
			WithCode(avokadoerror.CodeInvalidParam)
	}

	if cfg.logger == nil {
		cfg.logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}))
	}

	grmLogger := gormlogger.NewSlogLogger(cfg.logger, gormlogger.Config{
		LogLevel: cfg.gormLogLevel,
	})

	if cfg.gormConfig == nil {
		cfg.gormConfig = &gorm.Config{
			Logger:               grmLogger,
			DisableAutomaticPing: true,
			NowFunc: func() time.Time {
				return time.Now().UTC()
			},
		}
	} else if cfg.gormConfig.Logger == nil {
		cfg.gormConfig.Logger = grmLogger
	}

	pgConfig := postgres.Config{
		DSN: cfg.databaseURL,
	}

	db, dbErr := gorm.Open(postgres.New(pgConfig), cfg.gormConfig)
	if dbErr != nil {
		return nil, fmt.Errorf("[avokadodb.New gorm.Open] err: %w", dbErr)
	}
	sqlDB, sqlDBErr := db.DB()
	if sqlDBErr != nil {
		return nil, fmt.Errorf("[avokadodb.New sqlDB] err: %w", sqlDBErr)
	}

	sqlDB.SetMaxOpenConns(cfg.maxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.maxIdleConns)
	sqlDB.SetConnMaxIdleTime(cfg.connMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.connMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.pingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("[avokadodb.New sqlDB.PingContext] err: %w", err)
	}

	return db, nil
}
