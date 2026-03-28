// Package avokadodb provides GORM PostgreSQL connection builder with
// functional options.
package avokadodb

import (
	"context"
	"log/slog"
	"os"
	"strconv"
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
			return avokadoerror.New("[avokadodb.WithDatabaseURL] err: empty database url")
		}

		c.databaseURL = url

		return nil
	}
}

// WithMaxOpenConns sets the maximum number of open connections.
func WithMaxOpenConns(n int) Option {
	return func(c *config) error {
		if n < minMaxOpenConns {
			return avokadoerror.New(
				"[avokadodb.WithMaxOpenConns] err: '" + strconv.Itoa(
					n,
				) + "' received, must > " + strconv.Itoa(
					minMaxOpenConns,
				),
			)
		}
		if n > maxMaxOpenConns {
			return avokadoerror.New(
				"[avokadodb.WithMaxOpenConns] err: '" + strconv.Itoa(
					n,
				) + "' received, must < " + strconv.Itoa(
					maxMaxOpenConns,
				),
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
			return avokadoerror.New(
				"[avokadodb.WithMaxIdleConns] err: '" + strconv.Itoa(
					n,
				) + "' received, must > " + strconv.Itoa(
					minMaxIdleConns,
				),
			)
		}
		if n > maxMaxIdleConns {
			return avokadoerror.New(
				"[avokadodb.WithMaxIdleConns] err: '" + strconv.Itoa(
					n,
				) + "' received, must < " + strconv.Itoa(
					maxMaxIdleConns,
				),
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
			msg := "[avokadodb.WithConnMaxLifetime] err: '" +
				d.String() + "' received, must > " + minConnMaxLifetime.String()

			return avokadoerror.New(msg)
		}
		if d > maxConnMaxLifetime {
			msg := "[avokadodb.WithConnMaxLifetime] err: '" +
				d.String() + "' received, must < " + maxConnMaxLifetime.String()

			return avokadoerror.New(msg)
		}

		c.connMaxLifetime = d

		return nil
	}
}

// WithConnMaxIdleTime sets the maximum idle time of a connection.
func WithConnMaxIdleTime(d time.Duration) Option {
	return func(c *config) error {
		if d < minConnMaxIdleTime {
			msg := "[avokadodb.WithConnMaxIdleTime] err: '" +
				d.String() + "' received, must > " + minConnMaxIdleTime.String()

			return avokadoerror.New(msg)
		}
		if d > maxConnMaxIdleTime {
			msg := "[avokadodb.WithConnMaxIdleTime] err: '" +
				d.String() + "' received, must < " + maxConnMaxIdleTime.String()

			return avokadoerror.New(msg)
		}

		c.connMaxIdleTime = d

		return nil
	}
}

// WithPingTimeout sets the ping timeout of a connection.
func WithPingTimeout(d time.Duration) Option {
	return func(c *config) error {
		if d < minPingTimeout {
			return avokadoerror.New(
				"[avokadodb.WithPingTimeout] err: '" + d.String() + "' received, must > " + minPingTimeout.String(),
			)
		}
		if d > maxPingTimeout {
			return avokadoerror.New(
				"[avokadodb.WithPingTimeout] err: '" + d.String() + "' received, must < " + maxPingTimeout.String(),
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
			return avokadoerror.New(
				"[avokadodb.WithGormLogLevel] err: '" + strconv.Itoa(
					n,
				) + "' received, must > " + strconv.Itoa(
					int(gormlogger.Silent),
				),
			)
		}

		if n > int(gormlogger.Info) {
			return avokadoerror.New(
				"[avokadodb.WithGormLogLevel] err: '" + strconv.Itoa(
					n,
				) + "' received, must < " + strconv.Itoa(
					int(gormlogger.Info),
				),
			)
		}

		c.gormLogLevel = gormlogger.LogLevel(n)

		return nil
	}
}

// New creates a new GORM database connection with the given options.
func New(ctx context.Context, opts ...Option) (*gorm.DB, error) {
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
		return nil, avokadoerror.New("[avokadodb.New] err: databaseURL required")
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
		return nil, avokadoerror.New("[avokadodb.New gorm.Open] err").WithErr(dbErr)
	}
	sqlDB, sqlDBErr := db.DB()
	if sqlDBErr != nil {
		return nil, avokadoerror.New("[avokadodb.New sqlDB] err").WithErr(sqlDBErr)
	}

	sqlDB.SetMaxOpenConns(cfg.maxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.maxIdleConns)
	sqlDB.SetConnMaxIdleTime(cfg.connMaxIdleTime)
	sqlDB.SetConnMaxLifetime(cfg.connMaxLifetime)

	ctx, cancel := context.WithTimeout(ctx, cfg.pingTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, avokadoerror.New("[avokadodb.New sqlDB.PingContext] err").WithErr(err)
	}

	return db, nil
}
