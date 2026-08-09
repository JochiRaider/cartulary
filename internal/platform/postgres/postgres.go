package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

const (
	FilesystemRootDSNFile = "postgres.dsn"

	filesystemRootDSNMaximumBytes int64 = 65536
	managedServiceDSNPrefix             = "CARTULARY_POSTGRES_"
	managedServiceDSNSuffix             = "_DSN"
)

type Settings struct {
	BindingKind string
	RootPath    string
	DSN         string
	ServiceRef  string
}

type Binding struct {
	BindingKind string
	RootPath    string
	ServiceRef  string
}

func ResolveSettings(binding Binding, env map[string]string) (Settings, error) {
	switch binding.BindingKind {
	case "filesystem_root":
		if binding.RootPath == "" {
			return Settings{}, fmt.Errorf("resolve postgres settings: roots.database_storage.path is required")
		}
		root, err := rootedfs.Open(binding.RootPath)
		if err != nil {
			return Settings{}, fmt.Errorf("resolve postgres settings: database storage root is unavailable: %w", err)
		}
		defer root.Close()

		payload, _, err := root.ReadRegular(
			rootedfs.MustParseReference(FilesystemRootDSNFile),
			filesystemRootDSNMaximumBytes,
		)
		if err != nil {
			return Settings{}, fmt.Errorf("resolve postgres settings: root-bound DSN file is unavailable or unsafe: %w", err)
		}
		dsn := strings.TrimSpace(string(payload))
		if dsn == "" {
			return Settings{}, fmt.Errorf("resolve postgres settings: root-bound DSN file is empty")
		}
		return Settings{
			BindingKind: "filesystem_root",
			RootPath:    binding.RootPath,
			DSN:         dsn,
		}, nil
	case "managed_service":
		key, err := EnvKeyForServiceRef(binding.ServiceRef)
		if err != nil {
			return Settings{}, err
		}
		dsn, ok := lookupEnv(env, key)
		if !ok || dsn == "" {
			return Settings{}, fmt.Errorf("missing postgres DSN for managed service %q (%s)", binding.ServiceRef, key)
		}
		return Settings{
			BindingKind: "managed_service",
			DSN:         dsn,
			ServiceRef:  binding.ServiceRef,
		}, nil
	default:
		return Settings{}, fmt.Errorf("resolve postgres settings: roots.database_storage.binding_kind must be configured before postgres setup")
	}
}

func Setup(ctx context.Context, settings Settings) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(settings.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	return pgxpool.NewWithConfig(ctx, poolConfig)
}

func OpenSQL(settings Settings) (*sql.DB, error) {
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
	return managedServiceDSNPrefix + normalized + managedServiceDSNSuffix, nil
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
