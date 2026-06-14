package postgres_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestSupportPhase0_PostgresRootBindingResolution(t *testing.T) {
	t.Run("filesystem root uses root-bound DSN file and ignores generic DSN env", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "filesystem_root")
		if err := os.WriteFile(filepath.Join(cfg.Roots.DatabaseStorage.Path, postgres.FilesystemRootDSNFile), []byte("postgres://root-bound\n"), 0o600); err != nil {
			t.Fatalf("write root-bound postgres dsn: %v", err)
		}

		settings, err := postgres.ResolveSettings(cfg, map[string]string{
			postgres.PostgresDSNEnv: "postgres://generic-env",
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
		if err := os.Symlink(outsideDSN, filepath.Join(cfg.Roots.DatabaseStorage.Path, postgres.FilesystemRootDSNFile)); err != nil {
			t.Fatalf("create escaping postgres dsn symlink: %v", err)
		}

		_, err := postgres.ResolveSettings(cfg, nil)
		if err == nil {
			t.Fatal("expected filesystem-root postgres settings to reject escaping DSN symlink")
		}
		if !strings.Contains(err.Error(), "read root-bound DSN file") {
			t.Fatalf("unexpected filesystem-root postgres symlink error: %v", err)
		}
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
			postgres.PostgresDSNEnv: "postgres://generic-env",
		})
		if err == nil {
			t.Fatal("expected managed-service postgres settings to reject generic DSN env")
		}
		if !strings.Contains(err.Error(), "missing postgres DSN for managed service") {
			t.Fatalf("unexpected managed-service postgres error: %v", err)
		}
	})
}

func postgresSettingsConfig(t testing.TB, databaseBindingKind string) config.Config {
	t.Helper()

	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "postgres"), 0o700); err != nil {
		t.Fatalf("create postgres root: %v", err)
	}
	databaseRoot := config.RootBinding{BindingKind: "filesystem_root", Path: filepath.Join(base, "postgres")}
	profile := "disconnected"
	if databaseBindingKind == "managed_service" {
		databaseRoot = config.RootBinding{BindingKind: "managed_service", ServiceRef: "postgres-primary"}
		profile = "on_prem"
	}
	cfg, err := config.Validate(config.Config{
		ConfigSchemaID:    "cartulary.deployment_config.v1",
		DeploymentProfile: profile,
		Application: config.ApplicationConfig{
			PublicOrigin: "http://localhost:5173",
		},
		Roots: config.RootBindings{
			DatabaseStorage: databaseRoot,
			ObjectStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "object-store"),
			},
			BackupStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "backups"),
			},
			ReferencePackStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "reference-packs"),
			},
			TemporaryWork: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "tmp"),
			},
			ExportOutputs: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "exports"),
			},
		},
	})
	if err != nil {
		t.Fatalf("validate postgres settings config: %v", err)
	}
	return cfg
}
