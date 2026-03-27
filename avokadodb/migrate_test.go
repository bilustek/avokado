package avokadodb_test

import (
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/bilustek/avokado/avokadodb"
	"github.com/bilustek/avokado/avokadoerror"
)

func testDatabaseURL(t *testing.T) string {
	t.Helper()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	return url
}

func TestRunMigrations_RequiresValidURL(t *testing.T) {
	err := avokadodb.RunMigrations("")
	if err == nil {
		t.Fatal("expected error for empty database URL, got nil")
	}
}

func TestMigrationDown_RequiresValidURL(t *testing.T) {
	err := avokadodb.MigrationDown("")
	if err == nil {
		t.Fatal("expected error for empty database URL, got nil")
	}
}

func TestMigrationVersion_RequiresValidURL(t *testing.T) {
	_, _, err := avokadodb.MigrationVersion("")
	if err == nil {
		t.Fatal("expected error for empty database URL, got nil")
	}
}

func TestMigrationForce_RequiresValidURL(t *testing.T) {
	err := avokadodb.MigrationForce("", 1)
	if err == nil {
		t.Fatal("expected error for empty database URL, got nil")
	}
}

func TestMigrationsFS_ContainsMigrationFiles(t *testing.T) {
	entries, err := fs.ReadDir(avokadodb.MigrationsFS, "migrations")
	if err != nil {
		t.Fatalf("failed to read embedded migrations dir: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("expected embedded migration files, got none")
	}

	for _, entry := range entries {
		if entry.IsDir() {
			t.Errorf("unexpected directory in migrations: %s", entry.Name())
		}

		if !strings.HasSuffix(entry.Name(), ".sql") {
			t.Errorf("unexpected non-SQL file in migrations: %s", entry.Name())
		}
	}
}

func TestMigrationsFS_UpDownPairsMatch(t *testing.T) {
	entries, err := fs.ReadDir(avokadodb.MigrationsFS, "migrations")
	if err != nil {
		t.Fatalf("failed to read embedded migrations dir: %v", err)
	}

	ups := make(map[string]bool)
	downs := make(map[string]bool)

	for _, entry := range entries {
		name := entry.Name()

		switch {
		case strings.Contains(name, ".up.sql"):
			prefix := strings.SplitN(name, ".up.sql", 2)[0]
			ups[prefix] = true
		case strings.Contains(name, ".down.sql"):
			prefix := strings.SplitN(name, ".down.sql", 2)[0]
			downs[prefix] = true
		default:
			t.Errorf("migration file %q doesn't match .up.sql or .down.sql pattern", name)
		}
	}

	for prefix := range ups {
		if !downs[prefix] {
			t.Errorf("migration %q has .up.sql but missing .down.sql", prefix)
		}
	}

	for prefix := range downs {
		if !ups[prefix] {
			t.Errorf("migration %q has .down.sql but missing .up.sql", prefix)
		}
	}
}

func TestMigrationsFS_FilesAreNotEmpty(t *testing.T) {
	entries, err := fs.ReadDir(avokadodb.MigrationsFS, "migrations")
	if err != nil {
		t.Fatalf("failed to read embedded migrations dir: %v", err)
	}

	for _, entry := range entries {
		content, err := fs.ReadFile(avokadodb.MigrationsFS, "migrations/"+entry.Name())
		if err != nil {
			t.Errorf("failed to read migration file %s: %v", entry.Name(), err)

			continue
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			t.Errorf("migration file %s is empty", entry.Name())
		}
	}
}

func TestRunMigrations_MalformedURL(t *testing.T) {
	err := avokadodb.RunMigrations("not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for malformed database URL, got nil")
	}

	var avErr *avokadoerror.Error
	if !errors.As(err, &avErr) {
		t.Fatalf("expected *avokadoerror.Error, got %T", err)
	}

	if avErr.Code != avokadoerror.CodeDatabaseError {
		t.Errorf("expected error code %q, got %q", avokadoerror.CodeDatabaseError, avErr.Code)
	}
}

func TestMigrationDown_MalformedURL(t *testing.T) {
	err := avokadodb.MigrationDown("not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for malformed database URL, got nil")
	}

	var avErr *avokadoerror.Error
	if !errors.As(err, &avErr) {
		t.Fatalf("expected *avokadoerror.Error, got %T", err)
	}

	if avErr.Code != avokadoerror.CodeDatabaseError {
		t.Errorf("expected error code %q, got %q", avokadoerror.CodeDatabaseError, avErr.Code)
	}
}

func TestMigrationVersion_MalformedURL(t *testing.T) {
	_, _, err := avokadodb.MigrationVersion("not-a-valid-url")
	if err == nil {
		t.Fatal("expected error for malformed database URL, got nil")
	}

	var avErr *avokadoerror.Error
	if !errors.As(err, &avErr) {
		t.Fatalf("expected *avokadoerror.Error, got %T", err)
	}

	if avErr.Code != avokadoerror.CodeDatabaseError {
		t.Errorf("expected error code %q, got %q", avokadoerror.CodeDatabaseError, avErr.Code)
	}
}

func TestMigrationForce_MalformedURL(t *testing.T) {
	err := avokadodb.MigrationForce("not-a-valid-url", 1)
	if err == nil {
		t.Fatal("expected error for malformed database URL, got nil")
	}

	var avErr *avokadoerror.Error
	if !errors.As(err, &avErr) {
		t.Fatalf("expected *avokadoerror.Error, got %T", err)
	}

	if avErr.Code != avokadoerror.CodeDatabaseError {
		t.Errorf("expected error code %q, got %q", avokadoerror.CodeDatabaseError, avErr.Code)
	}
}

// Integration tests — require DATABASE_URL env var (skipped otherwise).

func TestRunMigrations_Integration(t *testing.T) {
	dbURL := testDatabaseURL(t)

	if err := avokadodb.RunMigrations(dbURL); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	// running again should be idempotent (no error, no change).
	if err := avokadodb.RunMigrations(dbURL); err != nil {
		t.Fatalf("RunMigrations (idempotent) failed: %v", err)
	}
}

func TestMigrationVersion_Integration(t *testing.T) {
	dbURL := testDatabaseURL(t)

	if err := avokadodb.RunMigrations(dbURL); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	version, dirty, err := avokadodb.MigrationVersion(dbURL)
	if err != nil {
		t.Fatalf("MigrationVersion failed: %v", err)
	}

	if version == 0 {
		t.Error("expected version > 0 after migrations")
	}

	if dirty {
		t.Error("expected dirty to be false after clean migration")
	}
}

func TestMigrationDownAndUp_Integration(t *testing.T) {
	dbURL := testDatabaseURL(t)

	// ensure migrations are applied first.
	if err := avokadodb.RunMigrations(dbURL); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	versionBefore, _, err := avokadodb.MigrationVersion(dbURL)
	if err != nil {
		t.Fatalf("MigrationVersion failed: %v", err)
	}

	// roll back one.
	if err = avokadodb.MigrationDown(dbURL); err != nil {
		t.Fatalf("MigrationDown failed: %v", err)
	}

	// after rolling back all migrations, Version() returns "no migration" error
	// which is expected — it means we're at version 0 (no migrations applied).
	_, _, err = avokadodb.MigrationVersion(dbURL)
	if err != nil {
		t.Logf("MigrationVersion after full rollback returned expected error: %v", err)
	}

	// re-apply.
	if err = avokadodb.RunMigrations(dbURL); err != nil {
		t.Fatalf("RunMigrations (re-apply) failed: %v", err)
	}

	versionRestored, _, err := avokadodb.MigrationVersion(dbURL)
	if err != nil {
		t.Fatalf("MigrationVersion after re-apply failed: %v", err)
	}

	if versionRestored != versionBefore {
		t.Errorf("expected version to restore to %d, got %d", versionBefore, versionRestored)
	}
}

func TestMigrationForce_Integration(t *testing.T) {
	dbURL := testDatabaseURL(t)

	if err := avokadodb.RunMigrations(dbURL); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	version, _, err := avokadodb.MigrationVersion(dbURL)
	if err != nil {
		t.Fatalf("MigrationVersion failed: %v", err)
	}

	// force to current version should not error.
	if err = avokadodb.MigrationForce(dbURL, int(version)); err != nil {
		t.Fatalf("MigrationForce failed: %v", err)
	}

	versionAfter, dirty, err := avokadodb.MigrationVersion(dbURL)
	if err != nil {
		t.Fatalf("MigrationVersion after force failed: %v", err)
	}

	if versionAfter != version {
		t.Errorf("expected version %d after force, got %d", version, versionAfter)
	}

	if dirty {
		t.Error("expected dirty to be false after force")
	}
}
