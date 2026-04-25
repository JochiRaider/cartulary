package app

import (
	"bytes"
	"database/sql"
	"errors"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestMigrateRunnerParsesPositionalCommandAndArgs(t *testing.T) {
	var gotCommand string
	var gotArgs []string

	runner := newTestMigrateRunner(t)
	runner.migrate = func(db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		gotCommand = command
		gotArgs = append([]string(nil), args...)
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	if exitCode := runner.runCLI([]string{"up-to", "5"}); exitCode != 0 {
		t.Fatalf("unexpected exit code: got %d want 0", exitCode)
	}
	if gotCommand != "up-to" {
		t.Fatalf("unexpected command: got %q want %q", gotCommand, "up-to")
	}
	if len(gotArgs) != 1 || gotArgs[0] != "5" {
		t.Fatalf("unexpected command args: got %#v want %#v", gotArgs, []string{"5"})
	}
}

func TestMigrateRunnerParsesCommandFlagAndArgs(t *testing.T) {
	var gotCommand string
	var gotArgs []string

	runner := newTestMigrateRunner(t)
	runner.migrate = func(db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		gotCommand = command
		gotArgs = append([]string(nil), args...)
		return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
	}

	if exitCode := runner.runCLI([]string{"-command", "up-to", "7"}); exitCode != 0 {
		t.Fatalf("unexpected exit code: got %d want 0", exitCode)
	}
	if gotCommand != "up-to" {
		t.Fatalf("unexpected command: got %q want %q", gotCommand, "up-to")
	}
	if len(gotArgs) != 1 || gotArgs[0] != "7" {
		t.Fatalf("unexpected command args: got %#v want %#v", gotArgs, []string{"7"})
	}
}

func TestMigrateRunnerConfigLoadFailure(t *testing.T) {
	stderr := &bytes.Buffer{}
	runner := newTestMigrateRunner(t)
	runner.stderr = stderr
	runner.loadConfig = func() (config.Config, error) {
		return config.Config{}, errors.New("config unavailable")
	}

	openCalled := false
	runner.openSQL = func(cfg config.Config) (*sql.DB, error) {
		openCalled = true
		return nil, nil
	}

	migrateCalled := false
	runner.migrate = func(db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		migrateCalled = true
		return postgres.MigrationStatus{}, nil
	}

	if exitCode := runner.runCLI(nil); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	if openCalled {
		t.Fatal("expected config failure to stop before opening postgres")
	}
	if migrateCalled {
		t.Fatal("expected config failure to stop before running migrations")
	}
	if output := stderr.String(); !strings.Contains(output, "load config: config unavailable") {
		t.Fatalf("expected config failure in stderr, got %q", output)
	}
}

func TestMigrateRunnerDBOpenFailure(t *testing.T) {
	stderr := &bytes.Buffer{}
	runner := newTestMigrateRunner(t)
	runner.stderr = stderr

	migrateCalled := false
	runner.openSQL = func(cfg config.Config) (*sql.DB, error) {
		return nil, errors.New("dsn rejected")
	}
	runner.migrate = func(db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
		migrateCalled = true
		return postgres.MigrationStatus{}, nil
	}

	if exitCode := runner.runCLI(nil); exitCode != 1 {
		t.Fatalf("unexpected exit code: got %d want 1", exitCode)
	}
	if migrateCalled {
		t.Fatal("expected db-open failure to stop before running migrations")
	}
	if output := stderr.String(); !strings.Contains(output, "open postgres: dsn rejected") {
		t.Fatalf("expected db-open failure in stderr, got %q", output)
	}
}

func newTestMigrateRunner(t testing.TB) migrateRunner {
	t.Helper()

	db, err := sql.Open("pgx", "postgres://cartulary:cartulary@127.0.0.1:1/cartulary?sslmode=disable")
	if err != nil {
		t.Fatalf("open test database handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	return migrateRunner{
		stderr: bytes.NewBuffer(nil),
		loadConfig: func() (config.Config, error) {
			return config.Config{}, nil
		},
		openSQL: func(cfg config.Config) (*sql.DB, error) {
			return db, nil
		},
		migrate: func(db *sql.DB, source postgres.MigrationSource, command string, args ...string) (postgres.MigrationStatus, error) {
			return postgres.MigrationStatus{Command: command, Directory: source.Name}, nil
		},
		source: postgres.MigrationSource{
			Path: "db/migrations",
			Name: "db/migrations",
		},
	}
}
