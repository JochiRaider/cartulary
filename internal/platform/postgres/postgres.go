package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"example.com/todo/cartulary/internal/platform/config"
)

const PostgresDSNEnv = "CARTULARY_POSTGRES_DSN"

type Settings struct {
	DSN string
}

type MigrationStatus struct {
	Command   string
	Directory string
	Empty     bool
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

func Migrate(db *sql.DB, directory string, command string) (MigrationStatus, error) {
	status := MigrationStatus{
		Command:   command,
		Directory: directory,
	}

	empty, err := migrationDirectoryEmpty(directory)
	if err != nil {
		return status, fmt.Errorf("inspect migration directory: %w", err)
	}
	if empty {
		status.Empty = true
		return status, nil
	}

	if err := goose.Run(command, db, directory); err != nil {
		return status, fmt.Errorf("run goose %q: %w", command, err)
	}

	return status, nil
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

func lookupEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}

	return os.LookupEnv(key)
}
