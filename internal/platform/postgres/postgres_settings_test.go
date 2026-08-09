package postgres_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestPostgresRootBindingResolution(t *testing.T) {
	t.Run("filesystem root uses root-bound DSN file and ignores generic DSN env", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "filesystem_root")
		if err := os.WriteFile(filepath.Join(cfg.RootPath, postgres.FilesystemRootDSNFile), []byte("postgres://root-bound\n"), 0o600); err != nil {
			t.Fatalf("write root-bound postgres dsn: %v", err)
		}

		settings, err := postgres.ResolveSettings(cfg, map[string]string{
			suiteservices.PostgresDSNEnv: "postgres://generic-env",
		})
		if err != nil {
			t.Fatalf("resolve filesystem-root postgres settings: %v", err)
		}
		if settings.BindingKind != "filesystem_root" || settings.DSN != "postgres://root-bound" {
			t.Fatalf("unexpected filesystem-root postgres settings: %#v", settings)
		}
	})

	t.Run("filesystem root rejects DSN symlink escapes", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "filesystem_root")
		outsideDir := t.TempDir()
		outsideDSN := filepath.Join(outsideDir, postgres.FilesystemRootDSNFile)
		if err := os.WriteFile(outsideDSN, []byte("postgres://escaped\n"), 0o600); err != nil {
			t.Fatalf("write escaped postgres dsn: %v", err)
		}
		if err := os.Symlink(outsideDSN, filepath.Join(cfg.RootPath, postgres.FilesystemRootDSNFile)); err != nil {
			t.Fatalf("create escaping postgres dsn symlink: %v", err)
		}

		_, err := postgres.ResolveSettings(cfg, nil)
		if err == nil {
			t.Fatal("expected filesystem-root postgres settings to reject escaping DSN symlink")
		}
		if !strings.Contains(err.Error(), "root-bound DSN file is unavailable or unsafe") {
			t.Fatalf("unexpected filesystem-root postgres symlink error: %v", err)
		}
		if strings.Contains(err.Error(), cfg.RootPath) || strings.Contains(err.Error(), outsideDir) {
			t.Fatalf("filesystem-root postgres error disclosed a host path: %v", err)
		}
	})

	t.Run("filesystem root rejects oversized hard-linked and non-regular DSN objects", func(t *testing.T) {
		t.Run("oversized", func(t *testing.T) {
			cfg := postgresSettingsConfig(t, "filesystem_root")
			if err := os.WriteFile(
				filepath.Join(cfg.RootPath, postgres.FilesystemRootDSNFile),
				make([]byte, 65537),
				0o600,
			); err != nil {
				t.Fatalf("write oversized DSN: %v", err)
			}
			if _, err := postgres.ResolveSettings(cfg, nil); err == nil {
				t.Fatal("oversized root-bound DSN was accepted")
			}
		})
		t.Run("hard link", func(t *testing.T) {
			cfg := postgresSettingsConfig(t, "filesystem_root")
			dsnPath := filepath.Join(cfg.RootPath, postgres.FilesystemRootDSNFile)
			if err := os.WriteFile(dsnPath, []byte("postgres://root-bound"), 0o600); err != nil {
				t.Fatalf("write DSN: %v", err)
			}
			if err := os.Link(dsnPath, filepath.Join(cfg.RootPath, "dsn-hard-link")); err != nil {
				t.Fatalf("create DSN hard link: %v", err)
			}
			if _, err := postgres.ResolveSettings(cfg, nil); err == nil {
				t.Fatal("hard-linked root-bound DSN was accepted")
			}
		})
		t.Run("directory", func(t *testing.T) {
			cfg := postgresSettingsConfig(t, "filesystem_root")
			if err := os.Mkdir(filepath.Join(cfg.RootPath, postgres.FilesystemRootDSNFile), 0o700); err != nil {
				t.Fatalf("create DSN directory: %v", err)
			}
			if _, err := postgres.ResolveSettings(cfg, nil); err == nil {
				t.Fatal("directory root-bound DSN was accepted")
			}
		})
	})

	t.Run("managed service requires service-ref DSN env", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "managed_service")
		key, err := postgres.EnvKeyForServiceRef("postgres-primary")
		if err != nil {
			t.Fatalf("resolve service-ref key: %v", err)
		}

		settings, err := postgres.ResolveSettings(cfg, map[string]string{
			key: "postgres://managed-service",
		})
		if err != nil {
			t.Fatalf("resolve managed-service postgres settings: %v", err)
		}
		if settings.BindingKind != "managed_service" || settings.DSN != "postgres://managed-service" {
			t.Fatalf("unexpected managed-service postgres settings: %#v", settings)
		}
	})

	t.Run("managed service env key normalization is ASCII only", func(t *testing.T) {
		key, err := postgres.EnvKeyForServiceRef("postgres.primary-1")
		if err != nil {
			t.Fatalf("resolve punctuation service-ref key: %v", err)
		}
		if key != "CARTULARY_POSTGRES_POSTGRES_PRIMARY_1_DSN" {
			t.Fatalf("unexpected normalized postgres key: got %q", key)
		}

		if _, err := postgres.EnvKeyForServiceRef("é"); err == nil {
			t.Fatal("expected non-ASCII-only service_ref to be rejected")
		}
	})

	t.Run("managed service ignores generic DSN env and fails closed", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "managed_service")
		_, err := postgres.ResolveSettings(cfg, map[string]string{
			suiteservices.PostgresDSNEnv: "postgres://generic-env",
		})
		if err == nil {
			t.Fatal("expected managed-service postgres settings to reject generic DSN env")
		}
		if !strings.Contains(err.Error(), "missing postgres DSN for managed service") {
			t.Fatalf("unexpected managed-service postgres error: %v", err)
		}
	})
}

func postgresSettingsConfig(t testing.TB, databaseBindingKind string) postgres.Binding {
	t.Helper()

	rootPath := filepath.Join(t.TempDir(), "postgres")
	if err := os.MkdirAll(rootPath, 0o700); err != nil {
		t.Fatalf("create postgres root: %v", err)
	}
	binding := postgres.Binding{BindingKind: "filesystem_root", RootPath: rootPath}
	if databaseBindingKind == "managed_service" {
		binding = postgres.Binding{BindingKind: "managed_service", ServiceRef: "postgres-primary"}
	}
	return binding
}
