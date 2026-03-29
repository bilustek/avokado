package avokadodb

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/bilustek/avokado/avokadoerror"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationsFS exposes the embedded migration files for external use (e.g., CLI, testing).
var MigrationsFS = migrationsFS

const (
	defaultMigrationsTable = "avokado"
	migrationFilePerms     = 0o600
)

func newMigrate(databaseURL string, migrationsTable string, migrationsDir fs.FS) (*migrate.Migrate, error) {
	if migrationsTable == "" {
		migrationsTable = defaultMigrationsTable
	}

	if migrationsDir == nil {
		migrationsDir = migrationsFS
	}

	sourceDriver, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return nil, avokadoerror.New("[avokadodb.newMigrate iofs.New]: migration source").
			WithErr(err)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, avokadoerror.New("[avokadodb.newMigrate sql.Open]: database open").
			WithErr(err)
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		_ = db.Close()

		return nil, avokadoerror.New("[avokadodb.newMigrate postgres.WithInstance]: migration db driver").
			WithErr(err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return nil, avokadoerror.New("[avokadodb.newMigrate migrate.NewWithInstance]: migration init").
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
func RunMigrations(databaseURL string, migrationsTable string, migrationsDir fs.FS) error {
	m, err := newMigrate(databaseURL, migrationsTable, migrationsDir)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return avokadoerror.New("[avokadodb.RunMigrations m.Up]: migration up").WithErr(err)
	}

	return nil
}

// MigrationDown rolls back the last applied migration.
// It returns nil if the rollback was successful or if there are no migrations to roll back.
func MigrationDown(databaseURL string, migrationsTable string, migrationsDir fs.FS) error {
	m, err := newMigrate(databaseURL, migrationsTable, migrationsDir)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return avokadoerror.New("[avokadodb.MigrationDown m.Down]: migration down").WithErr(err)
	}

	return nil
}

// MigrationVersion returns the current migration version and dirty state.
// Returns version 0 with nil error if no migrations have been applied yet.
func MigrationVersion(
	databaseURL string,
	migrationsTable string,
	migrationsDir fs.FS,
) (version uint, dirty bool, err error) {
	m, err := newMigrate(databaseURL, migrationsTable, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	defer closeMigrate(m)

	ver, dirtyState, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}

		return 0, false, avokadoerror.New("[avokadodb.MigrationVersion m.Version]: migration version").WithErr(err)
	}

	return ver, dirtyState, nil
}

// HasMigrations reports whether the given FS contains any .up.sql migration files.
func HasMigrations(migrationsDir fs.FS) bool {
	entries, err := fs.ReadDir(migrationsDir, "migrations")
	if err != nil {
		return false
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			return true
		}
	}

	return false
}

// MigrationInfo represents a single migration file and its applied state.
type MigrationInfo struct {
	Version uint
	Name    string
	Applied bool
}

// MigrationStatus lists all available migrations and marks which ones have been applied.
// It reads .up.sql files from the given migrationsDir FS and compares against the current version.
// The migrationsDir must contain a "migrations" subdirectory with .up.sql/.down.sql files.
func MigrationStatus(databaseURL string, migrationsTable string, migrationsDir fs.FS) ([]MigrationInfo, error) {
	currentVersion, _, err := MigrationVersion(databaseURL, migrationsTable, migrationsDir)
	if err != nil {
		return nil, err
	}

	entries, err := fs.ReadDir(migrationsDir, "migrations")
	if err != nil {
		return nil, avokadoerror.New("[avokadodb.MigrationStatus fs.ReadDir]: reading migrations dir").
			WithErr(err)
	}

	var migrations []MigrationInfo

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}

		ver, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			continue
		}

		migName := strings.TrimSuffix(parts[1], ".up.sql")

		migrations = append(migrations, MigrationInfo{
			Version: uint(ver),
			Name:    fmt.Sprintf("%s_%s", parts[0], migName),
			Applied: ver <= uint64(currentVersion) && currentVersion > 0,
		})
	}

	slices.SortFunc(migrations, func(a, b MigrationInfo) int {
		switch {
		case a.Version < b.Version:
			return -1
		case a.Version > b.Version:
			return 1
		default:
			return 0
		}
	})

	return migrations, nil
}

// ShowMigrations prints a formatted migration status report to the given writer.
// It displays the app name, current version, dirty state, and a checklist of all migrations.
func ShowMigrations(w io.Writer, databaseURL, migrationsTable, appName string, migrationsDir fs.FS) error {
	version, dirty, err := MigrationVersion(databaseURL, migrationsTable, migrationsDir)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "%s, version: %d, dirty?: %t\n\n", appName, version, dirty)

	migrations, err := MigrationStatus(databaseURL, migrationsTable, migrationsDir)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		mark := " "
		if m.Applied {
			mark = "X"
		}

		fmt.Fprintf(w, "  [%s] - %s\n", mark, m.Name)
	}

	fmt.Fprintln(w)

	return nil
}

// AddMigration creates a new pair of migration files (up and down) in the given directory.
// It auto-detects the next sequence number from existing files.
// The migrationsPath must point to a "migrations" directory on disk.
func AddMigration(migrationsPath string, name string) error {
	if name == "" {
		return avokadoerror.New("[avokadodb.AddMigration] err: empty migration name")
	}

	var maxVersion uint64

	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return avokadoerror.New("[avokadodb.AddMigration os.ReadDir]: reading migrations dir").
			WithErr(err)
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}

		parts := strings.SplitN(e.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}

		ver, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			continue
		}

		if ver > maxVersion {
			maxVersion = ver
		}
	}

	nextVersion := maxVersion + 1
	prefix := fmt.Sprintf("%06d_%s", nextVersion, name)

	upPath := filepath.Join(migrationsPath, prefix+".up.sql")
	downPath := filepath.Join(migrationsPath, prefix+".down.sql")

	if err := os.WriteFile(upPath, []byte(""), migrationFilePerms); err != nil {
		return avokadoerror.New("[avokadodb.AddMigration] err: creating up migration").
			WithErr(err)
	}

	if err := os.WriteFile(downPath, []byte(""), migrationFilePerms); err != nil {
		return avokadoerror.New("[avokadodb.AddMigration] err: creating down migration").
			WithErr(err)
	}

	return nil
}

// MigrationForce forces the migration version, clearing the dirty state.
// This is used to fix a dirty database state after a failed migration.
func MigrationForce(databaseURL string, version int, migrationsTable string, migrationsDir fs.FS) error {
	m, err := newMigrate(databaseURL, migrationsTable, migrationsDir)
	if err != nil {
		return err
	}
	defer closeMigrate(m)

	if err := m.Force(version); err != nil {
		return avokadoerror.New("[avokadodb.MigrationForce m.Force]: migration force").WithErr(err)
	}

	return nil
}
