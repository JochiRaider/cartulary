package serverprocess

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestEffectiveConfigBackupRoot_Process(t *testing.T) {
	t.Run("process environment builder contract", requireProcessEnvBuilderContract)
	t.Run("partial harness startup retains cleanup error", requireProcessHarnessStartupErrorContract)
	t.Run("bounded I/O and cleanup failure contract", requireBoundedProcessIOContract)

	const websocketPath = "/ws/v1/incidents/00000000-0000-0000-0000-000000000000"

	validConfig := string(fixtures.MustRead("config", "valid.toml"))
	cases := []struct {
		name       string
		configText string
		mutateEnv  func(map[string]string)
		wantPath   string
		wantReason string
	}{
		{
			name:       "missing backup storage root",
			configText: stripConfigSection(t, validConfig, "[roots.backup_storage]"),
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__BINDING_KIND"] = ""
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = ""
				env["CARTULARY__ROOTS__BACKUP_STORAGE__SERVICE_REF"] = ""
			},
			wantPath:   "roots.backup_storage",
			wantReason: "missing_required_key",
		},
		{
			name:       "disconnected backup storage managed service",
			configText: validConfig,
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__BINDING_KIND"] = "managed_service"
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = ""
				env["CARTULARY__ROOTS__BACKUP_STORAGE__SERVICE_REF"] = "backup-vault"
			},
			wantPath:   "roots.backup_storage.binding_kind",
			wantReason: "profile_incompatible_binding",
		},
		{
			name:       "backup storage satisfied by export outputs",
			configText: validConfig,
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = env["CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH"]
			},
			wantPath:   "roots.backup_storage.path",
			wantReason: "path_overlap",
		},
		{
			name:       "backup storage satisfied by temporary work",
			configText: validConfig,
			mutateEnv: func(env map[string]string) {
				env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"] = env["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"]
			},
			wantPath:   "roots.backup_storage.path",
			wantReason: "path_overlap",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := writeConfig(t, tc.configText)
			env := newProcessEnv(t, processEnvOptions{ConfigPath: configPath})
			tc.mutateEnv(env)

			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			err := server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected invalid backup-root config startup to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, websocketPath)
			server.RequireDiagnosticsCode(t, config.InvalidDeploymentConfigCode)
			server.RequireDiagnosticsField(t, tc.wantPath, tc.wantReason)
		})
	}
}

type cleanupReporterProbe struct {
	failed bool
}

func (p *cleanupReporterProbe) Helper() {}

func (p *cleanupReporterProbe) Errorf(string, ...any) {
	p.failed = true
}

func requireBoundedProcessIOContract(t *testing.T) {
	ctx, cancel := newProcessCleanupContext()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("cleanup context has no deadline")
	}
	remaining := time.Until(deadline)
	if remaining <= 0 || remaining > processCleanupTimeout {
		t.Fatalf("cleanup context deadline remaining=%s", remaining)
	}
	cancel()
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("cleanup context cancellation got %v", ctx.Err())
	}
	if got := newProcessHTTPClient().Timeout; got != processHTTPTimeout {
		t.Fatalf("process HTTP timeout got %s want %s", got, processHTTPTimeout)
	}
	probe := new(cleanupReporterProbe)
	reportProcessCleanupFailure(probe, "injected cleanup", errors.New("cleanup failed"))
	if !probe.failed {
		t.Fatal("cleanup error did not produce a failing report")
	}
}

func requireProcessHarnessStartupErrorContract(t *testing.T) {
	startupErr := errors.New("object-store startup")
	cleanupErr := errors.New("postgres cleanup")
	cleanupCalled := false
	_, _, err := startProcessHarnesses(t.Context(), processHarnessStarters{
		startPostgres: func(context.Context) (*pgtest.Harness, error) {
			return new(pgtest.Harness), nil
		},
		startObjectStore: func(context.Context) (*s3test.Harness, error) {
			return nil, startupErr
		},
		stopPostgres: func(context.Context) error {
			cleanupCalled = true
			return cleanupErr
		},
	})
	if !cleanupCalled || !errors.Is(err, startupErr) || !errors.Is(err, cleanupErr) {
		t.Fatalf("partial startup got cleanup_called=%v err=%v", cleanupCalled, err)
	}
}

func requireProcessEnvBuilderContract(t *testing.T) {
	database := map[string]string{
		suiteservices.PostgresDSNEnv: "postgres://source",
		"COPY_DATABASE":              "database",
		"PRECEDENCE":                 "database",
	}
	objectStore := map[string]string{
		"COPY_OBJECT_STORE": "object-store",
		"PRECEDENCE":        "object-store",
	}
	overrideRoot := filepath.Join(t.TempDir(), "final-database-root")
	overrides := map[string]string{
		suiteservices.PostgresDSNEnv:               "postgres://override",
		"CARTULARY__ROOTS__DATABASE_STORAGE__PATH": overrideRoot,
		"CARTULARY_CONFIG_FILE":                    "override-config.toml",
		"PRECEDENCE":                               "override",
	}
	env := newProcessEnv(t, processEnvOptions{
		Database:      database,
		ObjectStore:   objectStore,
		ConfigPath:    "default-config.toml",
		BootstrapPath: "bootstrap.json",
		Overrides:     overrides,
	})

	database["COPY_DATABASE"] = "mutated"
	objectStore["COPY_OBJECT_STORE"] = "mutated"
	overrides["PRECEDENCE"] = "mutated"
	if env["COPY_DATABASE"] != "database" || env["COPY_OBJECT_STORE"] != "object-store" {
		t.Fatalf("environment builder retained a caller map: %#v", env)
	}
	if env["PRECEDENCE"] != "override" || env["CARTULARY_CONFIG_FILE"] != "override-config.toml" {
		t.Fatalf("environment override precedence failed: %#v", env)
	}
	if env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] != "bootstrap.json" {
		t.Fatalf("environment bootstrap default missing: %#v", env)
	}
	boundDSN, err := os.ReadFile(filepath.Join(overrideRoot, postgres.FilesystemRuntimeDSNFile))
	if err != nil {
		t.Fatalf("read final database-root binding: %v", err)
	}
	if string(boundDSN) != "postgres://override\n" {
		t.Fatalf("final database-root binding got %q", boundDSN)
	}

	isolated := newProcessEnv(t, processEnvOptions{})
	if env["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"] == isolated["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"] {
		t.Fatal("environment builders reused temporary roots")
	}
	for _, rootKey := range []string{
		"CARTULARY__ROOTS__DATABASE_STORAGE__PATH",
		"CARTULARY__ROOTS__OBJECT_STORAGE__PATH",
		"CARTULARY__ROOTS__BACKUP_STORAGE__PATH",
		"CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH",
		"CARTULARY__ROOTS__TEMPORARY_WORK__PATH",
		"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH",
	} {
		info, err := os.Stat(env[rootKey])
		if err != nil || !info.IsDir() {
			t.Fatalf("environment root %s is not a directory: path=%q err=%v", rootKey, env[rootKey], err)
		}
	}
}
