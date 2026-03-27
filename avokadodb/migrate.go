package avokadodb

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationsFS exposes the embedded migration files for external use (e.g., CLI, testing).
var MigrationsFS = migrationsFS

const migrationsTable = "avokado"

func newMigrate(databaseURL string) (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, avokadoerror.New("[avokadodb.newMigrate iofs.New]: migration source").
			WithCode(avokadoerror.CodeDatabaseError).
			WithErr(err)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, avokadoerror.New("[avokadodb.newMigrate sql.Open]: database open").
			WithCode(avokadoerror.CodeDatabaseError).
			WithErr(err)
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		_ = db.Close()

		return nil, avokadoerror.New("[avokadodb.newMigrate postgres.WithInstance]: migration db driver").
			WithCode(avokadoerror.CodeDatabaseError).
			WithErr(err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return nil, avokadoerror.New("[avokadodb.newMigrate migrate.NewWithInstance]: migration init").
			WithCode(avokadoerror.CodeDatabaseError).
			WithErr(err)
	}

	return migrator, nil
}

func closeMigrate(m *migrate.Migrate) {
	sourceErr, databaseErr := m.Close()
	if sourceErr != nil {
		log.Printf("[avokadodb.closeMigrate]: migration source close error: %v", sourceErr)
	}

	if databaseErr != nil {
		log.Printf("[avokadodb.closeMigrate]: migration database close error: %v", databaseErr)
	}
}

// RunMigrations runs all pending migrations against the given database URL.
// It returns nil if all migrations ran successfully or if there are no new migrations.
func RunMigrations(databaseURL string) error {
	m, err := newMigrate(databaseURL)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("avokadodb: migration up: %w", err)
	}

	return nil
}

// MigrationDown rolls back the last applied migration.
// It returns nil if the rollback was successful or if there are no migrations to roll back.
func MigrationDown(databaseURL string) error {
	m, err := newMigrate(databaseURL)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("avokadodb: migration down: %w", err)
	}

	return nil
}

// MigrationVersion returns the current migration version and dirty state.
// Returns version 0 with nil error if no migrations have been applied yet.
func MigrationVersion(databaseURL string) (version uint, dirty bool, err error) {
	m, err := newMigrate(databaseURL)
	if err != nil {
		return 0, false, err
	}
	defer closeMigrate(m)

	ver, dirtyState, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("avokadodb: migration version: %w", err)
	}

	return ver, dirtyState, nil
}

// MigrationForce forces the migration version, clearing the dirty state.
// This is used to fix a dirty database state after a failed migration.
func MigrationForce(databaseURL string, version int) error {
	m, err := newMigrate(databaseURL)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Force(version); err != nil {
		return fmt.Errorf("avokadodb: migration force: %w", err)
	}

	return nil
}
