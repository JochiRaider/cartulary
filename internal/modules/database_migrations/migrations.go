package database_migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

const GooseLogFileEnv = "CARTULARY_GOOSE_LOG_FILE"

var errNilMigrateContext = errors.New("postgres migrate: nil context")

type MigrationSource struct {
	BaseFS                  fs.FS
	Path                    string
	Name                    string
	ExpectedLineageID       string
	ExpectedLineageBoundary string
}

type MigrationStatus struct {
	SourceName string
	Empty      bool
}

func NewMigrationSource(path string) MigrationSource {
	return MigrationSource{Path: path}
}

func NewEmbeddedMigrationSource(fsys fs.FS, path string, name string) MigrationSource {
	return MigrationSource{
		BaseFS: fsys,
		Path:   path,
		Name:   name,
	}
}

// Apply advances the database to the migration source head.
func Apply(ctx context.Context, db *sql.DB, source MigrationSource) (MigrationStatus, error) {
	return runMigrationOperation(ctx, db, source, migrationOperationApply, 0)
}

// ApplyThrough advances the database through the positive target version.
// It is intended for tests and repository test harnesses, not deployable CLI grammar.
func ApplyThrough(ctx context.Context, db *sql.DB, source MigrationSource, version int64) (MigrationStatus, error) {
	if version <= 0 {
		return migrationStatus(source), errors.New("migration apply-through version must be positive")
	}
	return runMigrationOperation(ctx, db, source, migrationOperationApplyThrough, version)
}

// RollbackThrough rolls the database back through the non-negative target version.
// It is intended for tests and repository test harnesses, not deployable CLI grammar.
func RollbackThrough(ctx context.Context, db *sql.DB, source MigrationSource, version int64) (MigrationStatus, error) {
	if version < 0 {
		return migrationStatus(source), errors.New("migration rollback-through version must be non-negative")
	}
	return runMigrationOperation(ctx, db, source, migrationOperationRollbackThrough, version)
}

type migrationOperation uint8

const (
	migrationOperationApply migrationOperation = iota
	migrationOperationApplyThrough
	migrationOperationRollbackThrough
)

func (operation migrationOperation) gooseName() string {
	switch operation {
	case migrationOperationApply:
		return "up"
	case migrationOperationApplyThrough:
		return "up-to"
	case migrationOperationRollbackThrough:
		return "down-to"
	default:
		return "unknown"
	}
}

func runMigrationOperation(ctx context.Context, db *sql.DB, source MigrationSource, operation migrationOperation, version int64) (MigrationStatus, error) {
	source = normalizeMigrationSource(source)
	status := migrationStatus(source)

	if ctx == nil {
		return status, errNilMigrateContext
	}
	if err := ctx.Err(); err != nil {
		return status, err
	}

	empty, err := migrationSourceEmpty(source)
	if err != nil {
		return status, fmt.Errorf("inspect migration directory: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return status, err
	}
	if empty {
		status.Empty = true
		return status, nil
	}

	if err := runMigrationPreflights(ctx, db, source, operation, version); err != nil {
		return status, err
	}
	if err := runGooseProvider(ctx, db, source, operation, version); err != nil {
		return status, fmt.Errorf("run goose %q: %w", operation.gooseName(), err)
	}

	return status, nil
}

func migrationStatus(source MigrationSource) MigrationStatus {
	source = normalizeMigrationSource(source)
	return MigrationStatus{SourceName: source.displayName()}
}

func normalizeMigrationSource(source MigrationSource) MigrationSource {
	if source.Path == "" {
		source.Path = "."
	}
	return source
}

func (source MigrationSource) displayName() string {
	if source.Name != "" {
		return source.Name
	}
	if source.Path == "" {
		return "."
	}
	return source.Path
}

func migrationSourceEmpty(source MigrationSource) (bool, error) {
	if source.BaseFS != nil {
		return migrationFSEmpty(source.BaseFS, source.Path)
	}
	return migrationDirectoryEmpty(source.Path)
}

func migrationDirectoryEmpty(directory string) (bool, error) {
	found := false
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == ".gitkeep" {
			return nil
		}

		found = true
		return fs.SkipAll
	})
	if err != nil && err != fs.SkipAll {
		return false, err
	}

	return !found, nil
}

func migrationFSEmpty(fsys fs.FS, directory string) (bool, error) {
	found := false
	err := fs.WalkDir(fsys, directory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Name() == ".gitkeep" {
			return nil
		}

		found = true
		return fs.SkipAll
	})
	if err != nil && err != fs.SkipAll {
		return false, err
	}

	return !found, nil
}

func runGooseProvider(ctx context.Context, db *sql.DB, source MigrationSource, operation migrationOperation, version int64) error {
	if ctx == nil {
		return errNilMigrateContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	providerFS, err := migrationProviderFS(source)
	if err != nil {
		return fmt.Errorf("resolve migration source: %w", err)
	}
	logger, closer, err := newGooseLogger(os.Getenv(GooseLogFileEnv))
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}

	provider, err := newGooseProvider(db, providerFS, logger)
	if err != nil {
		return fmt.Errorf("create goose provider: %w", err)
	}

	switch operation {
	case migrationOperationApply:
		_, err = provider.Up(ctx)
	case migrationOperationApplyThrough:
		_, err = provider.UpTo(ctx, version)
	case migrationOperationRollbackThrough:
		_, err = provider.DownTo(ctx, version)
	default:
		err = errors.New("unsupported migration operation")
	}
	return err
}

func newGooseProvider(db *sql.DB, sourceFS fs.FS, logger goose.Logger) (*goose.Provider, error) {
	return goose.NewProvider(
		goose.DialectPostgres,
		db,
		sourceFS,
		goose.WithDisableGlobalRegistry(true),
		goose.WithLogger(logger),
	)
}

func migrationProviderFS(source MigrationSource) (fs.FS, error) {
	if source.BaseFS == nil {
		return os.DirFS(source.Path), nil
	}
	if source.Path == "." {
		return source.BaseFS, nil
	}
	return fs.Sub(source.BaseFS, source.Path)
}

func newGooseLogger(logPath string) (goose.Logger, io.Closer, error) {
	if logPath == "" {
		return log.New(os.Stderr, "", log.LstdFlags), nil, nil
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return nil, nil, fmt.Errorf("create goose log directory: %w", err)
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- logPath is an artifact path provided by the repo-local runner.
	if err != nil {
		return nil, nil, fmt.Errorf("open goose log file: %w", err)
	}
	return log.New(file, "", log.LstdFlags), file, nil
}
