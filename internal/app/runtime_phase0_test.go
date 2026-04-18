package app

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

func TestPhase0_FailClosedStartup_U_0_05(t *testing.T) {
	cfg := phase0RuntimeConfig(t)
	cfg.Roots.DatabaseStorage.Path = "relative/postgres"

	originalNewJobsManager := newJobsManager
	originalSetupPostgres := setupPostgres
	originalSetupObjectStore := setupObjectStore
	originalNewWSHub := newWSHub
	originalNewHTTPHandler := newHTTPHandler
	t.Cleanup(func() {
		newJobsManager = originalNewJobsManager
		setupPostgres = originalSetupPostgres
		setupObjectStore = originalSetupObjectStore
		newWSHub = originalNewWSHub
		newHTTPHandler = originalNewHTTPHandler
	})

	var jobsCalls int
	newJobsManager = func() *jobs.Manager {
		jobsCalls++
		return &jobs.Manager{}
	}

	var postgresCalls int
	setupPostgres = func(ctx context.Context, cfg config.Config, env map[string]string) (*pgxpool.Pool, error) {
		postgresCalls++
		return nil, nil
	}

	var objectStoreCalls int
	setupObjectStore = func(ctx context.Context, cfg config.Config, env map[string]string) (*minio.Client, error) {
		objectStoreCalls++
		return nil, nil
	}

	var wsHubCalls int
	newWSHub = func() *platformws.Hub {
		wsHubCalls++
		return platformws.NewHub()
	}

	var handlerCalls int
	newHTTPHandler = func(options ...httpapi.Options) (http.Handler, error) {
		handlerCalls++
		return http.NewServeMux(), nil
	}

	_, err := NewRuntime(context.Background(), cfg, Options{})
	if err == nil {
		t.Fatal("expected invalid config to fail closed")
	}

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
	if !ok {
		t.Fatalf("expected diagnostics error, got %T", err)
	}
	if diagnosticsErr.Code != "invalid_deployment_config" {
		t.Fatalf("unexpected diagnostics code: got %q", diagnosticsErr.Code)
	}

	if jobsCalls != 0 || postgresCalls != 0 || objectStoreCalls != 0 || wsHubCalls != 0 || handlerCalls != 0 {
		t.Fatalf("expected fail-closed startup before any dependency wiring, got jobs=%d postgres=%d object_store=%d websocket=%d handler=%d", jobsCalls, postgresCalls, objectStoreCalls, wsHubCalls, handlerCalls)
	}
}

func phase0RuntimeConfig(t testing.TB) config.Config {
	t.Helper()

	base := t.TempDir()
	return config.Config{
		ConfigSchemaID:    "cartulary.deployment_config.v1",
		DeploymentProfile: "disconnected",
		Roots: config.RootBindings{
			DatabaseStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "postgres"),
			},
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
	}
}
