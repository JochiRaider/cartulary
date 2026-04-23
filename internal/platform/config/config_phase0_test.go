package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
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

	t.Run("selects an absolute CARTULARY_CONFIG_FILE override", func(t *testing.T) {
		alternate := writeTempConfig(t, string(fixtures.MustRead("config", "valid.toml")))
		path, err := ResolvePathWithOptions(LoadOptions{
			Env: map[string]string{
				ConfigFileEnv: alternate,
			},
		})
		if err != nil {
			t.Fatalf("resolve absolute selector override: %v", err)
		}
		if path != alternate {
			t.Fatalf("unexpected override path: got %q want %q", path, alternate)
		}

		cfg, err := LoadWithOptions(LoadOptions{
			Env: map[string]string{
				ConfigFileEnv: alternate,
			},
		})
		if err != nil {
			t.Fatalf("load config from absolute selector override: %v", err)
		}
		if cfg.Roots.DatabaseStorage.Path != "/var/lib/cartulary/postgres" {
			t.Fatalf("unexpected config loaded from override path: got %q", cfg.Roots.DatabaseStorage.Path)
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
	t.Run("accepts filesystem root bindings for every supported profile", func(t *testing.T) {
		for _, profile := range phase0SupportedDeploymentProfiles() {
			t.Run(profile, func(t *testing.T) {
				if _, err := Validate(phase0DeploymentProfileConfig(t, profile)); err != nil {
					t.Fatalf("validate %s filesystem-root config: %v", profile, err)
				}
			})
		}
	})

	t.Run("requires each runtime root for every supported profile", func(t *testing.T) {
		for _, profile := range phase0SupportedDeploymentProfiles() {
			for _, rootName := range phase0RuntimeRootNames() {
				t.Run(profile+"/"+rootName, func(t *testing.T) {
					cfg := setPhase0RootBinding(phase0DeploymentProfileConfig(t, profile), rootName, RootBinding{})

					_, err := Validate(cfg)
					requireDiagnostic(t, err, "roots."+rootName, "missing_required_key")
				})
			}
		}
	})

	t.Run("requires the full runtime root registry", func(t *testing.T) {
		err := loadInvalidConfig(t, stripSection(t, string(fixtures.MustRead("config", "valid.toml")), "[roots.export_outputs]"), nil)
		requireDiagnostic(t, err, "roots.export_outputs", "missing_required_key")
	})

	t.Run("rejects unknown runtime root keys in the config artifact", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml")) + "\n[roots.archive_storage]\nbinding_kind = \"filesystem_root\"\npath = \"/var/lib/cartulary/archive\"\n"
		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.archive_storage.binding_kind", "unknown_key")
	})

	t.Run("rejects unknown binding kinds", func(t *testing.T) {
		err := loadInvalidConfig(t, strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `binding_kind = "filesystem_root"`, `binding_kind = "object_store"`), nil)
		requireDiagnostic(t, err, "roots.backup_storage.binding_kind", "invalid_enum")
	})

	t.Run("rejects malformed filesystem root bindings", func(t *testing.T) {
		cfg := phase0BaseConfig(t)
		cfg.Roots.DatabaseStorage = RootBinding{
			BindingKind: "filesystem_root",
			ServiceRef:  "postgres-primary",
		}

		_, err := Validate(cfg)
		requireDiagnostic(t, err, "roots.database_storage.path", "missing_required_key")
		requireDiagnostic(t, err, "roots.database_storage.service_ref", "type_mismatch")
	})

	t.Run("accepts and rejects managed service bindings per root and supported profile", func(t *testing.T) {
		for _, profile := range phase0SupportedDeploymentProfiles() {
			for _, rootName := range phase0RuntimeRootNames() {
				t.Run(profile+"/"+rootName, func(t *testing.T) {
					cfg := setPhase0RootBinding(phase0DeploymentProfileConfig(t, profile), rootName, RootBinding{
						BindingKind: "managed_service",
						ServiceRef:  phase0ManagedServiceRef(rootName),
					})

					_, err := Validate(cfg)
					if phase0ManagedServiceAllowed(profile, rootName) {
						if err != nil {
							t.Fatalf("validate %s managed-service binding for %s: %v", profile, rootName, err)
						}
						return
					}

					requireDiagnostic(t, err, "roots."+rootName+".binding_kind", "profile_incompatible_binding")
				})
			}
		}
	})
}

func TestPhase0_FilesystemRootPaths_U_0_03(t *testing.T) {
	t.Run("rejects non-absolute filesystem roots", func(t *testing.T) {
		content := strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `path = "/var/lib/cartulary/postgres"`, `path = "relative/postgres"`)
		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.database_storage.path", "path_not_absolute")
	})

	t.Run("rejects empty and shell-expanded filesystem roots", func(t *testing.T) {
		cases := []struct {
			name   string
			path   string
			reason string
		}{
			{name: "empty", path: ``, reason: "missing_required_key"},
			{name: "home shorthand", path: `~/cartulary/postgres`, reason: "path_not_absolute"},
			{name: "shell variable", path: `$HOME/cartulary/postgres`, reason: "path_not_absolute"},
			{name: "shell variable braces", path: `${HOME}/cartulary/postgres`, reason: "path_not_absolute"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				content := strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `path = "/var/lib/cartulary/postgres"`, `path = "`+tc.path+`"`)
				err := loadInvalidConfig(t, content, nil)
				requireDiagnostic(t, err, "roots.database_storage.path", tc.reason)
			})
		}
	})

	t.Run("rejects forbidden lexical path segments", func(t *testing.T) {
		cases := []string{
			"/var/lib/cartulary/./postgres",
			"/var/lib/cartulary/../postgres",
		}

		for _, path := range cases {
			t.Run(path, func(t *testing.T) {
				content := strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `path = "/var/lib/cartulary/postgres"`, `path = "`+path+`"`)
				err := loadInvalidConfig(t, content, nil)
				requireDiagnostic(t, err, "roots.database_storage.path", "path_forbidden_segment")
			})
		}
	})

	t.Run("rejects NUL bytes when the runtime can construct them", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixtures.MustRead("config", "valid.toml")), nil)
		cfg.Roots.DatabaseStorage.Path = "/var/lib/cartulary/\x00postgres"

		_, err := Validate(cfg)
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

		env := phase0FilesystemRootEnv(base)
		env["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"] = readonly
		cfg := mustLoadConfig(t, string(fixtures.MustRead("config", "valid.toml")), env)

		_, err := ValidateForStartup(cfg)
		requireDiagnostic(t, err, "roots.database_storage.path", "path_not_writable")
	})

	t.Run("rejects effective writes that escape a configured root", func(t *testing.T) {
		base := t.TempDir()
		cfg := mustLoadConfig(t, string(fixtures.MustRead("config", "valid.toml")), phase0FilesystemRootEnv(base))

		validated, err := ValidateForStartup(cfg)
		if err != nil {
			t.Fatalf("validate startup config with temp roots: %v", err)
		}

		insideRelativePath := filepath.Join("nested", "proof.txt")
		if err := writeFileWithinFilesystemRoot(validated.Roots.TemporaryWork.Path, insideRelativePath, []byte("proof"), 0o644); err != nil {
			t.Fatalf("write inside configured root: %v", err)
		}

		insideTarget, err := resolvePathWithinFilesystemRoot(validated.Roots.TemporaryWork.Path, insideRelativePath)
		if err != nil {
			t.Fatalf("resolve in-root target: %v", err)
		}
		got, err := os.ReadFile(insideTarget)
		if err != nil {
			t.Fatalf("read in-root target: %v", err)
		}
		if string(got) != "proof" {
			t.Fatalf("unexpected in-root payload: got %q want %q", got, "proof")
		}

		escapeRelativePath := filepath.Join("..", "escape.txt")
		if err := writeFileWithinFilesystemRoot(validated.Roots.TemporaryWork.Path, escapeRelativePath, []byte("escape"), 0o644); err == nil {
			t.Fatal("expected attempted escape write to fail")
		}

		escapeTarget := filepath.Join(base, "escape.txt")
		if _, err := os.Stat(escapeTarget); !os.IsNotExist(err) {
			if err == nil {
				t.Fatalf("unexpected escaped write created %q", escapeTarget)
			}
			t.Fatalf("stat escaped target %q: %v", escapeTarget, err)
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

	t.Run("does not let disconnected defaults mask malformed configured roots", func(t *testing.T) {
		content := strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `path = "/var/lib/cartulary/object-store"`, `path = "relative/object-store"`)
		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.object_storage.path", "path_not_absolute")
	})

	t.Run("does not apply disconnected-only defaults to non-disconnected profiles", func(t *testing.T) {
		content := strings.ReplaceAll(string(fixtures.MustRead("config", "valid.toml")), `deployment_profile = "disconnected"`, `deployment_profile = "on_prem"`)
		content = stripSection(t, content, "[roots.reference_pack_storage]")
		err := loadInvalidConfig(t, content, nil)
		requireDiagnostic(t, err, "roots.reference_pack_storage", "missing_required_key")
	})
}

func TestPhase0_ResourceLimits_U_0_09(t *testing.T) {
	t.Run("resolves omitted resource-limit keys to the exact defaults", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml"))
		content = stripSection(t, content, "[limits.object_blobs]")
		content = stripSection(t, content, "[limits.imports]")
		content = stripSection(t, content, "[limits.archives]")
		content = stripSection(t, content, "[limits.reference_packs]")
		content = stripSection(t, content, "[limits.incident_bundles]")
		content = stripSection(t, content, "[limits.previews]")

		cfg := mustLoadConfig(t, content, nil)
		if cfg.Limits.ObjectBlobs.MaxDeclaredByteSize != DefaultObjectBlobMaxDeclaredByteSize {
			t.Fatalf("unexpected object blob default: got %d want %d", cfg.Limits.ObjectBlobs.MaxDeclaredByteSize, DefaultObjectBlobMaxDeclaredByteSize)
		}
		if cfg.Limits.Archives.MaxCompressionRatio != DefaultArchiveMaxCompressionRatio {
			t.Fatalf("unexpected archive compression ratio default: got %d want %d", cfg.Limits.Archives.MaxCompressionRatio, DefaultArchiveMaxCompressionRatio)
		}
		if cfg.Limits.Previews.MaxTextInlineBytes != DefaultPreviewMaxTextInlineBytes {
			t.Fatalf("unexpected preview inline default: got %d want %d", cfg.Limits.Previews.MaxTextInlineBytes, DefaultPreviewMaxTextInlineBytes)
		}
	})

	t.Run("applies explicit overrides to only the targeted declared key", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml"))
		content = stripSection(t, content, "[limits.object_blobs]")
		content = stripSection(t, content, "[limits.imports]")
		content = stripSection(t, content, "[limits.archives]")
		content = stripSection(t, content, "[limits.reference_packs]")
		content = stripSection(t, content, "[limits.incident_bundles]")
		content = stripSection(t, content, "[limits.previews]")
		content += "\n[limits.imports]\nmax_rows = 42\n"

		cfg := mustLoadConfig(t, content, nil)
		if cfg.Limits.Imports.MaxRows != 42 {
			t.Fatalf("unexpected overridden import row limit: got %d", cfg.Limits.Imports.MaxRows)
		}
		if cfg.Limits.Imports.MaxColumns != DefaultImportMaxColumns {
			t.Fatalf("unexpected import column default after override: got %d want %d", cfg.Limits.Imports.MaxColumns, DefaultImportMaxColumns)
		}
	})

	t.Run("rejects values below the closed numeric domain minimum", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
			"CARTULARY__LIMITS__ARCHIVES__MAX_COMPRESSION_RATIO": "0",
		})
		requireDiagnostic(t, err, "limits.archives.max_compression_ratio", "value_below_minimum")
	})

	t.Run("rejects values above the closed numeric domain maximum", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
			"CARTULARY__LIMITS__ARCHIVES__MAX_COMPRESSION_RATIO": "1001",
		})
		requireDiagnostic(t, err, "limits.archives.max_compression_ratio", "value_above_maximum")
	})

	t.Run("rejects non-integer resource-limit forms", func(t *testing.T) {
		err := loadInvalidConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
			"CARTULARY__LIMITS__PREVIEWS__MAX_PREVIEWABLE_PAYLOAD_BYTES": "32MB",
		})
		requireDiagnostic(t, err, "limits.previews.max_previewable_payload_bytes", "type_mismatch")
	})

	t.Run("rejects undeclared pseudo resource-limit keys", func(t *testing.T) {
		cases := []struct {
			envKey string
			path   string
		}{
			{"CARTULARY__LIMITS__VIEW_QUERY__MAX_SORT_ENTRIES", "limits.view_query.max_sort_entries"},
			{"CARTULARY__LIMITS__VIEW_QUERY__MAX_FILTER_ENTRIES", "limits.view_query.max_filter_entries"},
			{"CARTULARY__LIMITS__RECORDS__MAX_CHANGES", "limits.records.max_changes"},
			{"CARTULARY__LIMITS__RECORDS__MAX_COLLECTION_ACTIONS", "limits.records.max_collection_actions"},
		}

		for _, tc := range cases {
			t.Run(tc.path, func(t *testing.T) {
				err := loadInvalidConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
					tc.envKey: "9",
				})
				requireDiagnostic(t, err, tc.path, "unknown_key")
			})
		}
	})

	t.Run("keeps fixed public ceilings deployment-invariant", func(t *testing.T) {
		cfg := mustLoadConfig(t, string(fixtures.MustRead("config", "valid.toml")), map[string]string{
			"CARTULARY__LIMITS__IMPORTS__MAX_ROWS":     "777",
			"CARTULARY__LIMITS__ARCHIVES__MAX_MEMBERS": "123",
		})
		if cfg.Limits.Imports.MaxRows != 777 {
			t.Fatalf("unexpected overridden import row limit: got %d", cfg.Limits.Imports.MaxRows)
		}
		if PublicSortLimit != 8 || PublicFilterLimit != 16 || PublicChangeLimit != 32 || PublicCollectionActionLimit != 64 {
			t.Fatalf("unexpected fixed public ceilings: got sort=%d filters=%d changes=%d collection_actions=%d", PublicSortLimit, PublicFilterLimit, PublicChangeLimit, PublicCollectionActionLimit)
		}
	})
}

func fixtureConfigPath() string {
	return fixtures.Path("config", "valid.toml")
}

func phase0BaseConfig(t testing.TB) Config {
	t.Helper()
	return mustLoadConfig(t, string(fixtures.MustRead("config", "valid.toml")), nil)
}

func phase0DeploymentProfileConfig(t testing.TB, profile string) Config {
	t.Helper()

	cfg := phase0BaseConfig(t)
	cfg.DeploymentProfile = profile
	return cfg
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

func phase0SupportedDeploymentProfiles() []string {
	return []string{"disconnected", "on_prem", "cloud"}
}

func phase0RuntimeRootNames() []string {
	return []string{
		"database_storage",
		"object_storage",
		"backup_storage",
		"reference_pack_storage",
		"temporary_work",
		"export_outputs",
	}
}

func phase0ManagedServiceAllowed(profile string, rootName string) bool {
	if profile == "disconnected" {
		return false
	}

	switch rootName {
	case "database_storage", "object_storage", "backup_storage":
		return true
	default:
		return false
	}
}

func phase0ManagedServiceRef(rootName string) string {
	switch rootName {
	case "database_storage":
		return "postgres-primary"
	case "object_storage":
		return "minio-primary"
	case "backup_storage":
		return "backup-vault"
	default:
		return "shared-service"
	}
}

func setPhase0RootBinding(cfg Config, rootName string, binding RootBinding) Config {
	switch rootName {
	case "database_storage":
		cfg.Roots.DatabaseStorage = binding
	case "object_storage":
		cfg.Roots.ObjectStorage = binding
	case "backup_storage":
		cfg.Roots.BackupStorage = binding
	case "reference_pack_storage":
		cfg.Roots.ReferencePackStorage = binding
	case "temporary_work":
		cfg.Roots.TemporaryWork = binding
	case "export_outputs":
		cfg.Roots.ExportOutputs = binding
	default:
		panic("unknown phase0 root name: " + rootName)
	}

	return cfg
}

func phase0FilesystemRootEnv(base string) map[string]string {
	return map[string]string{
		"CARTULARY__ROOTS__DATABASE_STORAGE__PATH":       filepath.Join(base, "postgres"),
		"CARTULARY__ROOTS__OBJECT_STORAGE__PATH":         filepath.Join(base, "object-store"),
		"CARTULARY__ROOTS__BACKUP_STORAGE__PATH":         filepath.Join(base, "backups"),
		"CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH": filepath.Join(base, "reference-packs"),
		"CARTULARY__ROOTS__TEMPORARY_WORK__PATH":         filepath.Join(base, "tmp"),
		"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH":         filepath.Join(base, "exports"),
	}
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
