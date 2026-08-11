package postgres_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestPostgresPurposeContracts(t *testing.T) {
	tests := []struct {
		name    string
		purpose postgres.Purpose
		file    string
		segment string
		role    string
	}{
		{name: "runtime", purpose: postgres.PurposeRuntime, file: postgres.FilesystemRuntimeDSNFile, segment: "RUNTIME", role: "cartulary_runtime"},
		{name: "migration", purpose: postgres.PurposeMigration, file: postgres.FilesystemMigrationDSNFile, segment: "MIGRATION", role: "cartulary_schema_owner"},
		{name: "recovery", purpose: postgres.PurposeRecovery, file: postgres.FilesystemRecoveryDSNFile, segment: "RECOVERY", role: "cartulary_recovery"},
	}
	for _, test := range tests {
		t.Run(test.name+" filesystem", func(t *testing.T) {
			cfg := postgresSettingsConfig(t, "filesystem_root")
			if err := os.WriteFile(filepath.Join(cfg.RootPath, test.file), []byte("postgres://selected\r\n"), 0o600); err != nil {
				t.Fatalf("write selected DSN: %v", err)
			}
			settings, err := postgres.ResolveSettings(cfg, test.purpose, map[string]string{})
			if err != nil {
				t.Fatalf("resolve filesystem settings: %v", err)
			}
			if settings.Purpose != test.purpose || settings.ExpectedRole != test.role || settings.DSN != "postgres://selected" {
				t.Fatalf("unexpected settings: %#v", settings)
			}
		})
		t.Run(test.name+" managed", func(t *testing.T) {
			cfg := postgresSettingsConfig(t, "managed_service")
			key, err := postgres.EnvKeyForServiceRef(cfg.ServiceRef, test.purpose)
			if err != nil {
				t.Fatalf("resolve selected key: %v", err)
			}
			wantKey := "CARTULARY_POSTGRES_POSTGRES_PRIMARY_" + test.segment + "_DSN"
			if key != wantKey {
				t.Fatalf("unexpected key: got %q want %q", key, wantKey)
			}
			settings, err := postgres.ResolveSettings(cfg, test.purpose, map[string]string{key: "postgres://selected"})
			if err != nil {
				t.Fatalf("resolve managed settings: %v", err)
			}
			if settings.Purpose != test.purpose || settings.ExpectedRole != test.role || settings.DSN != "postgres://selected" {
				t.Fatalf("unexpected settings: %#v", settings)
			}
		})
	}
}

func TestPostgresResolutionPrecedence(t *testing.T) {
	t.Run("unknown purpose precedes binding inspection", func(t *testing.T) {
		assertConfigurationReason(t, func() error {
			_, err := postgres.ResolveSettings(postgres.Binding{}, postgres.Purpose(255), map[string]string{})
			return err
		}(), postgres.ReasonPurposeUnknown)
	})

	t.Run("invalid binding precedes credentials", func(t *testing.T) {
		assertConfigurationReason(t, func() error {
			_, err := postgres.ResolveSettings(postgres.Binding{BindingKind: "managed_service"}, postgres.PurposeRuntime, map[string]string{
				"CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN": "retired-secret",
			})
			return err
		}(), postgres.ReasonBindingInvalid)
	})

	t.Run("retired managed input precedes unselected and selected", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "managed_service")
		assertConfigurationReason(t, func() error {
			_, err := postgres.ResolveSettings(cfg, postgres.PurposeRuntime, map[string]string{
				"CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN":           "retired-secret",
				"CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN": "migration-secret",
				"CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN":   "runtime-secret",
			})
			return err
		}(), postgres.ReasonRetiredCredentialPresent)
	})

	t.Run("unselected managed input is rejected by presence only", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "managed_service")
		assertConfigurationReason(t, func() error {
			_, err := postgres.ResolveSettings(cfg, postgres.PurposeRuntime, map[string]string{
				"CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN": "",
				"CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN":   "runtime-secret",
			})
			return err
		}(), postgres.ReasonUnselectedPurposeCredentialPresent)
	})

	t.Run("missing and empty selected inputs differ", func(t *testing.T) {
		cfg := postgresSettingsConfig(t, "managed_service")
		assertConfigurationReason(t, func() error {
			_, err := postgres.ResolveSettings(cfg, postgres.PurposeRecovery, map[string]string{})
			return err
		}(), postgres.ReasonSelectedCredentialMissing)
		assertConfigurationReason(t, func() error {
			_, err := postgres.ResolveSettings(cfg, postgres.PurposeRecovery, map[string]string{
				"CARTULARY_POSTGRES_POSTGRES_PRIMARY_RECOVERY_DSN": "",
			})
			return err
		}(), postgres.ReasonSelectedCredentialInvalid)
	})
}

func TestPostgresFilesystemCredentialSafety(t *testing.T) {
	t.Run("retired and unselected objects are rejected without reading", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			file   string
			reason string
		}{
			{name: "retired", file: "postgres.dsn", reason: postgres.ReasonRetiredCredentialPresent},
			{name: "unselected", file: postgres.FilesystemRecoveryDSNFile, reason: postgres.ReasonUnselectedPurposeCredentialPresent},
		} {
			t.Run(test.name, func(t *testing.T) {
				cfg := postgresSettingsConfig(t, "filesystem_root")
				if err := os.Mkdir(filepath.Join(cfg.RootPath, test.file), 0o700); err != nil {
					t.Fatalf("create presence-only object: %v", err)
				}
				assertConfigurationReason(t, func() error {
					_, err := postgres.ResolveSettings(cfg, postgres.PurposeRuntime, nil)
					return err
				}(), test.reason)
			})
		}
	})

	t.Run("selected object must be no-follow bounded regular file", func(t *testing.T) {
		for _, test := range []struct {
			name   string
			create func(t *testing.T, path string)
		}{
			{name: "symlink", create: func(t *testing.T, path string) {
				outside := filepath.Join(t.TempDir(), "outside")
				if err := os.WriteFile(outside, []byte("postgres://secret"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "directory", create: func(t *testing.T, path string) {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "oversized", create: func(t *testing.T, path string) {
				if err := os.WriteFile(path, make([]byte, 65537), 0o600); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "hard-linked", create: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("postgres://secret"), 0o600); err != nil {
					t.Fatal(err)
				}
				if err := os.Link(path, path+"-second-name"); err != nil {
					t.Fatal(err)
				}
			}},
		} {
			t.Run(test.name, func(t *testing.T) {
				cfg := postgresSettingsConfig(t, "filesystem_root")
				test.create(t, filepath.Join(cfg.RootPath, postgres.FilesystemRuntimeDSNFile))
				assertConfigurationReason(t, func() error {
					_, err := postgres.ResolveSettings(cfg, postgres.PurposeRuntime, nil)
					return err
				}(), postgres.ReasonSelectedCredentialInvalid)
			})
		}
	})

	t.Run("selected payload grammar is exact", func(t *testing.T) {
		tests := []struct {
			name    string
			payload []byte
			valid   bool
		}{
			{name: "single line", payload: []byte("postgres://selected"), valid: true},
			{name: "terminal LF", payload: []byte("postgres://selected\n"), valid: true},
			{name: "terminal CRLF", payload: []byte("postgres://selected\r\n"), valid: true},
			{name: "empty", payload: nil},
			{name: "only LF", payload: []byte("\n")},
			{name: "multiple LF", payload: []byte("postgres://selected\n\n")},
			{name: "embedded LF", payload: []byte("postgres://selected\nsecond")},
			{name: "terminal CR", payload: []byte("postgres://selected\r")},
			{name: "NUL", payload: []byte("postgres://selected\x00")},
			{name: "invalid UTF-8", payload: []byte{0xff}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				cfg := postgresSettingsConfig(t, "filesystem_root")
				if err := os.WriteFile(filepath.Join(cfg.RootPath, postgres.FilesystemRuntimeDSNFile), test.payload, 0o600); err != nil {
					t.Fatalf("write selected DSN: %v", err)
				}
				_, err := postgres.ResolveSettings(cfg, postgres.PurposeRuntime, nil)
				if test.valid && err != nil {
					t.Fatalf("valid payload rejected: %v", err)
				}
				if !test.valid {
					assertConfigurationReason(t, err, postgres.ReasonSelectedCredentialInvalid)
				}
			})
		}
	})
}

func TestPostgresDiagnosticsAreOpaque(t *testing.T) {
	secret := "postgres://secret-user:secret-password@secret-host/secret-database"
	cfg := postgresSettingsConfig(t, "managed_service")
	_, err := postgres.ResolveSettings(cfg, postgres.PurposeRuntime, map[string]string{
		"CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN": secret,
	})
	if err == nil {
		t.Fatal("expected retired input rejection")
	}
	for _, forbidden := range []string{"secret-user", "secret-password", "secret-host", "secret-database", cfg.ServiceRef} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("diagnostic disclosed %q: %v", forbidden, err)
		}
	}
}

func assertConfigurationReason(t testing.TB, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s", want)
	}
	var configurationErr *postgres.ConfigurationError
	if !errors.As(err, &configurationErr) {
		t.Fatalf("expected configuration error, got %T: %v", err, err)
	}
	if configurationErr.Reason() != want || err.Error() != want {
		t.Fatalf("unexpected reason: got %q want %q", err.Error(), want)
	}
}

func postgresSettingsConfig(t testing.TB, databaseBindingKind string) postgres.Binding {
	t.Helper()
	rootPath := filepath.Join(t.TempDir(), "postgres")
	if err := os.MkdirAll(rootPath, 0o700); err != nil {
		t.Fatalf("create postgres root: %v", err)
	}
	if databaseBindingKind == "managed_service" {
		return postgres.Binding{BindingKind: "managed_service", ServiceRef: "postgres-primary"}
	}
	return postgres.Binding{BindingKind: "filesystem_root", RootPath: rootPath}
}
