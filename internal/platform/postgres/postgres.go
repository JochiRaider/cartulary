package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

const (
	PostgresDSNEnv          = "CARTULARY_POSTGRES_DSN"
	FilesystemRootDSNFile   = "postgres.dsn"
	ManagedServiceDSNPrefix = "CARTULARY_POSTGRES_"
	ManagedServiceDSNSuffix = "_DSN"
	GooseLogFileEnv         = "CARTULARY_GOOSE_LOG_FILE"
)

var (
	errNilMigrateContext = errors.New("postgres migrate: nil context")
	gooseBaseFSSem       = make(chan struct{}, 1)
	gooseLoggerMu        sync.Mutex
	gooseRunContext      = goose.RunContext
	gooseSetBaseFS       = goose.SetBaseFS
)

type Settings struct {
	BindingKind string
	RootPath    string
	DSN         string
	ServiceRef  string
}

type MigrationSource struct {
	BaseFS                  fs.FS
	Path                    string
	Name                    string
	ExpectedLineageID       string
	ExpectedLineageBoundary string
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
	settings, err := ResolveSettings(cfg, nil)
	if err != nil {
		return ""
	}
	return settings.DSN
}

func ResolveSettings(cfg config.Config, env map[string]string) (Settings, error) {
	switch cfg.Roots.DatabaseStorage.BindingKind {
	case "filesystem_root":
		if cfg.Roots.DatabaseStorage.Path == "" {
			return Settings{}, fmt.Errorf("resolve postgres settings: roots.database_storage.path is required")
		}
		root, err := os.OpenRoot(cfg.Roots.DatabaseStorage.Path)
		if err != nil {
			return Settings{}, fmt.Errorf("resolve postgres settings: open database storage root %s: %w", cfg.Roots.DatabaseStorage.Path, err)
		}
		defer root.Close()

		payload, err := root.ReadFile(FilesystemRootDSNFile)
		if err != nil {
			return Settings{}, fmt.Errorf("resolve postgres settings: read root-bound DSN file %s: %w", filepath.Join(cfg.Roots.DatabaseStorage.Path, FilesystemRootDSNFile), err)
		}
		dsn := strings.TrimSpace(string(payload))
		if dsn == "" {
			return Settings{}, fmt.Errorf("resolve postgres settings: root-bound DSN file %s is empty", filepath.Join(cfg.Roots.DatabaseStorage.Path, FilesystemRootDSNFile))
		}
		return Settings{
			BindingKind: "filesystem_root",
			RootPath:    cfg.Roots.DatabaseStorage.Path,
			DSN:         dsn,
		}, nil
	case "managed_service":
		key, err := EnvKeyForServiceRef(cfg.Roots.DatabaseStorage.ServiceRef)
		if err != nil {
			return Settings{}, err
		}
		dsn, ok := lookupEnv(env, key)
		if !ok || dsn == "" {
			return Settings{}, fmt.Errorf("missing postgres DSN for managed service %q (%s)", cfg.Roots.DatabaseStorage.ServiceRef, key)
		}
		return Settings{
			BindingKind: "managed_service",
			DSN:         dsn,
			ServiceRef:  cfg.Roots.DatabaseStorage.ServiceRef,
		}, nil
	default:
		return Settings{}, fmt.Errorf("resolve postgres settings: roots.database_storage.binding_kind must be configured before postgres setup")
	}
}

func Setup(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	return SetupWithEnv(ctx, cfg, nil)
}

func SetupWithEnv(ctx context.Context, cfg config.Config, env map[string]string) (*pgxpool.Pool, error) {
	settings, err := ResolveSettings(cfg, env)
	if err != nil {
		return nil, err
	}
	poolConfig, err := pgxpool.ParseConfig(settings.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	return pgxpool.NewWithConfig(ctx, poolConfig)
}

func OpenSQL(cfg config.Config) (*sql.DB, error) {
	return OpenSQLWithEnv(cfg, nil)
}

func OpenSQLWithEnv(cfg config.Config, env map[string]string) (*sql.DB, error) {
	settings, err := ResolveSettings(cfg, env)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("pgx", settings.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres sql handle: %w", err)
	}

	return db, nil
}

func EnvKeyForServiceRef(serviceRef string) (string, error) {
	normalized := normalizeServiceRef(serviceRef)
	if normalized == "" {
		return "", fmt.Errorf("resolve managed postgres env key: service_ref must contain at least one letter or digit")
	}
	return ManagedServiceDSNPrefix + normalized + ManagedServiceDSNSuffix, nil
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

	if err := runMigrationPreflights(ctx, db, source, command, args...); err != nil {
		return status, err
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

	run := func() error {
		return runGooseWithConfiguredSource(ctx, command, db, source, args...)
	}
	return runGooseWithConfiguredLogger(os.Getenv(GooseLogFileEnv), run)
}

func runGooseWithConfiguredSource(ctx context.Context, command string, db *sql.DB, source MigrationSource, args ...string) error {
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

func runGooseWithConfiguredLogger(logPath string, run func() error) error {
	gooseLoggerMu.Lock()
	defer gooseLoggerMu.Unlock()

	if logPath == "" {
		goose.SetLogger(log.New(os.Stderr, "", log.LstdFlags))
		return run()
	}

	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return fmt.Errorf("create goose log directory: %w", err)
	}
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600) // #nosec G304 -- logPath is an artifact path provided by the repo-local runner.
	if err != nil {
		return fmt.Errorf("open goose log file: %w", err)
	}
	defer file.Close()

	goose.SetLogger(log.New(file, "", log.LstdFlags))
	defer goose.SetLogger(log.New(os.Stderr, "", log.LstdFlags))
	return run()
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

func normalizeServiceRef(value string) string {
	var builder strings.Builder
	previousUnderscore := false
	for i := 0; i < len(value); i++ {
		c := value[i]
		switch {
		case c >= 'a' && c <= 'z':
			builder.WriteByte(c - ('a' - 'A'))
			previousUnderscore = false
		case c >= 'A' && c <= 'Z' || c >= '0' && c <= '9':
			builder.WriteByte(c)
			previousUnderscore = false
		case !previousUnderscore:
			builder.WriteByte('_')
			previousUnderscore = true
		}
	}

	return strings.Trim(builder.String(), "_")
}
