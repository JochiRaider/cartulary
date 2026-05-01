package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

const PostgresDSNEnv = "CARTULARY_POSTGRES_DSN"

var (
	errNilMigrateContext = errors.New("postgres migrate: nil context")
	gooseBaseFSSem       = make(chan struct{}, 1)
	gooseRunContext      = goose.RunContext
	gooseSetBaseFS       = goose.SetBaseFS
)

type Settings struct {
	DSN string
}

type MigrationSource struct {
	BaseFS fs.FS
	Path   string
	Name   string
}

type MigrationStatus struct {
	Command          string
	Directory        string
	Empty            bool
	TemplateClone    bool
	TemplateDatabase string
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

func ConnectionString(cfg config.Config) string {
	return ResolveSettings(cfg, nil).DSN
}

func ResolveSettings(cfg config.Config, env map[string]string) Settings {
	_ = cfg
	if dsn, ok := lookupEnv(env, PostgresDSNEnv); ok && dsn != "" {
		return Settings{DSN: dsn}
	}

	return Settings{
		DSN: "postgres://cartulary:cartulary@localhost:5432/cartulary?sslmode=disable",
	}
}

func Setup(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	return SetupWithEnv(ctx, cfg, nil)
}

func SetupWithEnv(ctx context.Context, cfg config.Config, env map[string]string) (*pgxpool.Pool, error) {
	settings := ResolveSettings(cfg, env)
	poolConfig, err := pgxpool.ParseConfig(settings.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	// TODO: derive the runtime DSN from deployment configuration instead of fixed local defaults.
	return pgxpool.NewWithConfig(ctx, poolConfig)
}

func OpenSQL(cfg config.Config) (*sql.DB, error) {
	return OpenSQLWithEnv(cfg, nil)
}

func OpenSQLWithEnv(cfg config.Config, env map[string]string) (*sql.DB, error) {
	settings := ResolveSettings(cfg, env)
	db, err := sql.Open("pgx", settings.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres sql handle: %w", err)
	}

	return db, nil
}

func Migrate(ctx context.Context, db *sql.DB, source MigrationSource, command string, args ...string) (MigrationStatus, error) {
	source = normalizeMigrationSource(source)

	status := MigrationStatus{
		Command:   command,
		Directory: source.displayName(),
	}

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

	if err := runGoose(ctx, command, db, source, args...); err != nil {
		return status, fmt.Errorf("run goose %q: %w", command, err)
	}

	return status, nil
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

func runGoose(ctx context.Context, command string, db *sql.DB, source MigrationSource, args ...string) error {
	if ctx == nil {
		return errNilMigrateContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if source.BaseFS == nil {
		return gooseRunContext(ctx, command, db, source.Path, args...)
	}

	// goose stores the migration filesystem in package-global state, so embedded
	// runs are serialized to keep concurrent callers from trampling each other.
	release, err := acquireGooseBaseFS(ctx)
	if err != nil {
		return err
	}
	defer release()

	gooseSetBaseFS(source.BaseFS)
	defer gooseSetBaseFS(nil)

	if err := ctx.Err(); err != nil {
		return err
	}

	return gooseRunContext(ctx, command, db, source.Path, args...)
}

func acquireGooseBaseFS(ctx context.Context) (func(), error) {
	if ctx == nil {
		return nil, errNilMigrateContext
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case gooseBaseFSSem <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-gooseBaseFSSem
			return nil, err
		}
		return func() { <-gooseBaseFSSem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}

	return os.LookupEnv(key)
}
