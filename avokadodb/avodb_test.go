package avokadodb_test

import (
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/bilustek/avokado/avoerror"
	"github.com/bilustek/avokado/avokadodb"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const testDSN = "postgres://localhost:5432/nonexistent_test_db"

func TestWithDatabaseURL(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		if _, err := avokadodb.New(avokadodb.WithDatabaseURL("")); err == nil {
			t.Fatal("expected error for empty URL")
		}
	})
}

func TestWithMaxOpenConns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"valid_min", 1, false},
		{"valid_mid", 50, false},
		{"valid_max", 100, false},
		{"below_min", 0, true},
		{"negative", -1, true},
		{"above_max", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := avokadodb.New(
				avokadodb.WithDatabaseURL(testDSN),
				avokadodb.WithMaxOpenConns(tt.value),
			)
			if tt.wantErr && err == nil {
				t.Errorf("WithMaxOpenConns(%d) expected error", tt.value)
			}
			if !tt.wantErr && err != nil && isOptionError(err) {
				t.Errorf("WithMaxOpenConns(%d) unexpected option error: %v", tt.value, err)
			}
		})
	}
}

func TestWithMaxIdleConns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"valid_min", 1, false},
		{"valid_mid", 50, false},
		{"valid_max", 100, false},
		{"below_min", 0, true},
		{"negative", -1, true},
		{"above_max", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := avokadodb.New(
				avokadodb.WithDatabaseURL(testDSN),
				avokadodb.WithMaxIdleConns(tt.value),
			)
			if tt.wantErr && err == nil {
				t.Errorf("WithMaxIdleConns(%d) expected error", tt.value)
			}
			if !tt.wantErr && err != nil && isOptionError(err) {
				t.Errorf("WithMaxIdleConns(%d) unexpected option error: %v", tt.value, err)
			}
		})
	}
}

func TestWithConnMaxLifetime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   time.Duration
		wantErr bool
	}{
		{"valid_min", 1 * time.Minute, false},
		{"valid_mid", 15 * time.Minute, false},
		{"valid_max", 30 * time.Minute, false},
		{"below_min", 30 * time.Second, true},
		{"above_max", 31 * time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := avokadodb.New(
				avokadodb.WithDatabaseURL(testDSN),
				avokadodb.WithConnMaxLifetime(tt.value),
			)
			if tt.wantErr && err == nil {
				t.Errorf("WithConnMaxLifetime(%s) expected error", tt.value)
			}
			if !tt.wantErr && err != nil && isOptionError(err) {
				t.Errorf("WithConnMaxLifetime(%s) unexpected option error: %v", tt.value, err)
			}
		})
	}
}

func TestWithConnMaxIdleTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   time.Duration
		wantErr bool
	}{
		{"valid_min", 1 * time.Minute, false},
		{"valid_mid", 15 * time.Minute, false},
		{"valid_max", 30 * time.Minute, false},
		{"below_min", 30 * time.Second, true},
		{"above_max", 31 * time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := avokadodb.New(
				avokadodb.WithDatabaseURL(testDSN),
				avokadodb.WithConnMaxIdleTime(tt.value),
			)
			if tt.wantErr && err == nil {
				t.Errorf("WithConnMaxIdleTime(%s) expected error", tt.value)
			}
			if !tt.wantErr && err != nil && isOptionError(err) {
				t.Errorf("WithConnMaxIdleTime(%s) unexpected option error: %v", tt.value, err)
			}
		})
	}
}

func TestWithPingTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   time.Duration
		wantErr bool
	}{
		{"valid_min", 1 * time.Second, false},
		{"valid_mid", 15 * time.Second, false},
		{"valid_max", 30 * time.Second, false},
		{"below_min", 500 * time.Millisecond, true},
		{"above_max", 31 * time.Second, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := avokadodb.New(
				avokadodb.WithDatabaseURL(testDSN),
				avokadodb.WithPingTimeout(tt.value),
			)
			if tt.wantErr && err == nil {
				t.Errorf("WithPingTimeout(%s) expected error", tt.value)
			}
			if !tt.wantErr && err != nil && isOptionError(err) {
				t.Errorf("WithPingTimeout(%s) unexpected option error: %v", tt.value, err)
			}
		})
	}
}

func TestWithLogger(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if _, err := avokadodb.New(
		avokadodb.WithDatabaseURL(testDSN),
		avokadodb.WithLogger(logger),
	); err != nil && isOptionError(err) {
		t.Fatalf("unexpected option error: %v", err)
	}
}

func TestWithGormConfig(t *testing.T) {
	t.Parallel()

	t.Run("with_config", func(t *testing.T) {
		t.Parallel()

		if _, err := avokadodb.New(
			avokadodb.WithDatabaseURL(testDSN),
			avokadodb.WithGormConfig(&gorm.Config{}),
		); err != nil && isOptionError(err) {
			t.Fatalf("unexpected option error: %v", err)
		}
	})

	t.Run("nil_config", func(t *testing.T) {
		t.Parallel()

		if _, err := avokadodb.New(
			avokadodb.WithDatabaseURL(testDSN),
			avokadodb.WithGormConfig(nil),
		); err != nil && isOptionError(err) {
			t.Fatalf("unexpected option error: %v", err)
		}
	})
}

func TestWithGormLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"silent", int(gormlogger.Silent), false},
		{"error", int(gormlogger.Error), false},
		{"warn", int(gormlogger.Warn), false},
		{"info", int(gormlogger.Info), false},
		{"below_silent", int(gormlogger.Silent) - 1, true},
		{"above_info", int(gormlogger.Info) + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := avokadodb.New(
				avokadodb.WithDatabaseURL(testDSN),
				avokadodb.WithGormLogLevel(tt.value),
			)
			if tt.wantErr && err == nil {
				t.Errorf("WithGormLogLevel(%d) expected error", tt.value)
			}
			if !tt.wantErr && err != nil && isOptionError(err) {
				t.Errorf("WithGormLogLevel(%d) unexpected option error: %v", tt.value, err)
			}
		})
	}
}

func TestNew_NoDatabaseURL(t *testing.T) {
	t.Parallel()

	if _, err := avokadodb.New(); err == nil {
		t.Fatal("expected error for missing database URL")
	}
}

func TestNew_OptionError(t *testing.T) {
	t.Parallel()

	if _, err := avokadodb.New(
		avokadodb.WithDatabaseURL(testDSN),
		avokadodb.WithMaxOpenConns(-1),
	); err == nil {
		t.Fatal("expected error from invalid option")
	}
}

// isOptionError checks if the error is from option validation (not from DB connection).
func isOptionError(err error) bool {
	var avoErr *avoerror.Error

	return errors.As(err, &avoErr)
}
