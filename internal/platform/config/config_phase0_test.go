package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/todo/cartulary/internal/testutil/fixtures"
)

func TestPhase0_ConfigDiscovery_U_0_01(t *testing.T) {
	t.Run("uses the canonical default selector", func(t *testing.T) {
		path, err := ResolvePathWithOptions(LoadOptions{})
		if err != nil {
			t.Fatalf("resolve path: %v", err)
		}
		if path != DefaultConfigPath {
			t.Fatalf("unexpected default path: got %q want %q", path, DefaultConfigPath)
		}
	})

	t.Run("requires selector overrides to be absolute", func(t *testing.T) {
		_, err := ResolvePathWithOptions(LoadOptions{
			Env: map[string]string{
				ConfigFileEnv: "relative/config.toml",
			},
		})
		if err == nil {
			t.Fatal("expected relative selector override to fail")
		}
	})

	t.Run("applies nested overlays after file load", func(t *testing.T) {
		cfg, err := LoadWithOptions(LoadOptions{
			Path: fixtureConfigPath(),
			Env: map[string]string{
				"CARTULARY__ROOTS__BACKUP_STORAGE__PATH": "/srv/cartulary/backups",
			},
		})
		if err != nil {
			t.Fatalf("load valid config with overlay: %v", err)
		}

		if cfg.Roots.BackupStorage.Path != "/srv/cartulary/backups" {
			t.Fatalf("unexpected overlay result: got %q", cfg.Roots.BackupStorage.Path)
		}
	})

	t.Run("rejects unknown file keys", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml"))
		content = strings.Replace(content, `deployment_profile = "disconnected"`, strings.Join([]string{
			`deployment_profile = "disconnected"`,
			`unexpected_key = "surprise"`,
		}, "\n"), 1)

		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "unexpected_key", "unknown_key")
	})

	t.Run("rejects unknown overlay keys", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
			"CARTULARY__ROOTS__UNKNOWN_STORAGE__PATH": "/srv/cartulary/unknown",
		})
		requireDiagnostic(t, err, "roots.unknown_storage.path", "unknown_key")
	})

	t.Run("rejects unsupported config schema identifiers", func(t *testing.T) {
		err := loadInvalidConfig(t, strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `config_schema_id = "cartulary.deployment_config.v1"`, `config_schema_id = "cartulary.deployment_config.v2"`), nil)
		requireDiagnostic(t, err, "config_schema_id", "unsupported_config_schema_id")
	})

	t.Run("rejects invalid deployment profiles", func(t *testing.T) {
		err := loadInvalidConfig(t, strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `deployment_profile = "disconnected"`, `deployment_profile = "edge"`), nil)
		requireDiagnostic(t, err, "deployment_profile", "invalid_enum")
	})
}

func TestPhase0_RuntimeRoots_U_0_02(t *testing.T) {
	t.Run("requires the full runtime root registry", func(t *testing.T) {
		err := loadInvalidConfig(t, stripSection(t, string(fixtures.MustRead("config", "valid.toml")), "[roots.export_outputs]"), nil)
		requireDiagnostic(t, err, "roots.export_outputs", "missing_required_key")
	})

	t.Run("rejects unknown binding kinds", func(t *testing.T) {
		err := loadInvalidConfig(t, strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `binding_kind = "filesystem_root"`, `binding_kind = "object_store"`), nil)
		requireDiagnostic(t, err, "roots.backup_storage.binding_kind", "invalid_enum")
	})

	t.Run("rejects malformed filesystem root bindings", func(t *testing.T) {
		content := strings.Replace(string(fixtures.MustRead("config", "valid.toml")), strings.Join([]string{
			`[roots.database_storage]`,
			`binding_kind = "filesystem_root"`,
			`path = "/var/lib/cartulary/postgres"`,
		}, "\n"), strings.Join([]string{
			`[roots.database_storage]`,
			`binding_kind = "filesystem_root"`,
			`service_ref = "postgres-primary"`,
		}, "\n"), 1)

		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.database_storage.path", "missing_required_key")
		requireDiagnostic(t, err, "roots.database_storage.service_ref", "type_mismatch")
	})

	t.Run("rejects profile-incompatible managed services in disconnected mode", func(t *testing.T) {
		content := strings.Replace(string(fixtures.MustRead("config", "valid.toml")), strings.Join([]string{
			`[roots.database_storage]`,
			`binding_kind = "filesystem_root"`,
			`path = "/var/lib/cartulary/postgres"`,
		}, "\n"), strings.Join([]string{
			`[roots.database_storage]`,
			`binding_kind = "managed_service"`,
			`service_ref = "postgres-primary"`,
		}, "\n"), 1)

		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.database_storage.binding_kind", "profile_incompatible_binding")
	})

	t.Run("accepts managed services only on allowed profiles", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml"))
		content = strings.ReplaceAll(content, `deployment_profile = "disconnected"`, `deployment_profile = "on_prem"`)
		content = strings.Replace(content, strings.Join([]string{
			`[roots.database_storage]`,
			`binding_kind = "filesystem_root"`,
			`path = "/var/lib/cartulary/postgres"`,
		}, "\n"), strings.Join([]string{
			`[roots.database_storage]`,
			`binding_kind = "managed_service"`,
			`service_ref = "postgres-primary"`,
		}, "\n"), 1)
		content = strings.Replace(content, strings.Join([]string{
			`[roots.object_storage]`,
			`binding_kind = "filesystem_root"`,
			`path = "/var/lib/cartulary/object-store"`,
		}, "\n"), strings.Join([]string{
			`[roots.object_storage]`,
			`binding_kind = "managed_service"`,
			`service_ref = "minio-primary"`,
		}, "\n"), 1)
		content = strings.Replace(content, strings.Join([]string{
			`[roots.backup_storage]`,
			`binding_kind = "filesystem_root"`,
			`path = "/var/lib/cartulary/backups"`,
		}, "\n"), strings.Join([]string{
			`[roots.backup_storage]`,
			`binding_kind = "managed_service"`,
			`service_ref = "backup-vault"`,
		}, "\n"), 1)

		if _, err := LoadWithOptions(LoadOptions{Path: writeTempConfig(t, content)}); err != nil {
			t.Fatalf("load on-prem managed services config: %v", err)
		}
	})
}

func TestPhase0_FilesystemRootPaths_U_0_03(t *testing.T) {
	t.Run("rejects non-absolute filesystem roots", func(t *testing.T) {
		content := strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `path = "/var/lib/cartulary/postgres"`, `path = "relative/postgres"`)
		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.database_storage.path", "path_not_absolute")
	})

	t.Run("rejects forbidden lexical path segments", func(t *testing.T) {
		content := strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `path = "/var/lib/cartulary/postgres"`, `path = "/var/lib/cartulary/../postgres"`)
		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.database_storage.path", "path_forbidden_segment")
	})

	t.Run("rejects symlink overlap after startup canonicalization", func(t *testing.T) {
		base := t.TempDir()
		shared := filepath.Join(base, "shared")
		dbReal := filepath.Join(shared, "database")
		objectReal := filepath.Join(shared, "database", "objects")
		if err := os.MkdirAll(objectReal, 0o755); err != nil {
			t.Fatalf("create object root: %v", err)
		}

		dbAlias := filepath.Join(base, "database-alias")
		if err := os.Symlink(dbReal, dbAlias); err != nil {
			t.Fatalf("create symlink: %v", err)
		}

		cfg := mustLoadConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
			"CARTULARY__ROOTS__DATABASE_STORAGE__PATH":       dbAlias,
			"CARTULARY__ROOTS__OBJECT_STORAGE__PATH":         objectReal,
			"CARTULARY__ROOTS__BACKUP_STORAGE__PATH":         filepath.Join(base, "backups"),
			"CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH": filepath.Join(base, "reference-packs"),
			"CARTULARY__ROOTS__TEMPORARY_WORK__PATH":         filepath.Join(base, "tmp"),
			"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH":         filepath.Join(base, "exports"),
		})

		_, err := ValidateForStartup(cfg)
		requireDiagnostic(t, err, "roots.object_storage.path", "path_overlap")
	})

	t.Run("rejects non-writable filesystem roots at startup", func(t *testing.T) {
		base := t.TempDir()
		readonly := filepath.Join(base, "readonly")
		if err := os.MkdirAll(readonly, 0o755); err != nil {
			t.Fatalf("create readonly root: %v", err)
		}
		if err := os.Chmod(readonly, 0o555); err != nil {
			t.Fatalf("chmod readonly root: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chmod(readonly, 0o755)
		})

		cfg := mustLoadConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
			"CARTULARY__ROOTS__DATABASE_STORAGE__PATH":       readonly,
			"CARTULARY__ROOTS__OBJECT_STORAGE__PATH":         filepath.Join(base, "object-store"),
			"CARTULARY__ROOTS__BACKUP_STORAGE__PATH":         filepath.Join(base, "backups"),
			"CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH": filepath.Join(base, "reference-packs"),
			"CARTULARY__ROOTS__TEMPORARY_WORK__PATH":         filepath.Join(base, "tmp"),
			"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH":         filepath.Join(base, "exports"),
		})

		_, err := ValidateForStartup(cfg)
		requireDiagnostic(t, err, "roots.database_storage.path", "path_not_writable")
	})

	t.Run("rejects effective writes that escape a configured root", func(t *testing.T) {
		if !pathEscapesRoot("/srv/cartulary/tmp", "/srv/cartulary/tmp/../escape.txt") {
			t.Fatal("expected path escape detection to reject traversal outside the configured root")
		}
	})
}

func TestPhase0_DisconnectedDefaults_U_0_04(t *testing.T) {
	t.Run("accepts the canonical disconnected example without hidden rewrites", func(t *testing.T) {
		cfg, err := LoadWithOptions(LoadOptions{Path: fixtureConfigPath()})
		if err != nil {
			t.Fatalf("load canonical disconnected fixture: %v", err)
		}

		if cfg.Roots.DatabaseStorage.Path != "/var/lib/cartulary/postgres" {
			t.Fatalf("unexpected database storage path: got %q", cfg.Roots.DatabaseStorage.Path)
		}
		if cfg.Roots.ExportOutputs.Path != "/var/lib/cartulary/exports" {
			t.Fatalf("unexpected export root path: got %q", cfg.Roots.ExportOutputs.Path)
		}
	})

	t.Run("does not default missing required roots", func(t *testing.T) {
		err := loadInvalidConfig(t, stripSection(t, string(fixtures.MustRead("config", "valid.toml")), "[roots.backup_storage]"), nil)
		requireDiagnostic(t, err, "roots.backup_storage", "missing_required_key")
	})
}

func fixtureConfigPath() string {
	return fixtures.Path("config", "valid.toml")
}

func mustLoadConfig(t testing.TB, content string, env map[string]string) Config {
	t.Helper()

	cfg, err := LoadWithOptions(LoadOptions{
		Path: writeTempConfig(t, content),
		Env:  env,
	})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	return cfg
}

func loadInvalidConfig(t testing.TB, content string, env map[string]string) error {
	t.Helper()

	_, err := LoadWithOptions(LoadOptions{
		Path: writeTempConfig(t, content),
		Env:  env,
	})
	if err == nil {
		t.Fatal("expected invalid config to fail")
	}

	return err
}

func writeTempConfig(t testing.TB, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}

	return path
}

func stripSection(t testing.TB, content string, header string) string {
	t.Helper()

	lines := strings.Split(content, "\n")
	start := -1
	end := len(lines)
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			start = i
			continue
		}
		if start >= 0 && strings.HasPrefix(strings.TrimSpace(line), "[") {
			end = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("section %q not found", header)
	}

	return strings.Join(append(lines[:start], lines[end:]...), "\n")
}

func requireDiagnostic(t testing.TB, err error, wantPath string, wantReason string) {
	t.Helper()

	diagnosticsErr, ok := err.(*DiagnosticsError)
	if !ok {
		t.Fatalf("expected diagnostics error, got %T", err)
	}

	for _, diagnostic := range diagnosticsErr.Diagnostics {
		if diagnostic.Path == wantPath && diagnostic.ReasonCode == wantReason {
			return
		}
	}

	t.Fatalf("missing diagnostic path=%q reason=%q in %#v", wantPath, wantReason, diagnosticsErr.Diagnostics)
}
