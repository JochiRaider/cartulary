package server

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

func TestPhase0_FailClosedStartup_U_0_05(t *testing.T) {
	originalNewJobsManager := newJobsManager
	originalSetupPostgres := setupPostgres
	originalSetupObjectStore := setupObjectStore
	originalRunBootstrap := runBootstrap
	originalNewWSHub := newWSHub
	originalNewHTTPHandler := newHTTPHandler
	t.Cleanup(func() {
		newJobsManager = originalNewJobsManager
		setupPostgres = originalSetupPostgres
		setupObjectStore = originalSetupObjectStore
		runBootstrap = originalRunBootstrap
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
	var setupStore objectstore.Store
	setupObjectStore = func(ctx context.Context, cfg config.Config, env map[string]string) (objectstore.Store, error) {
		objectStoreCalls++
		return setupStore, nil
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

	t.Run("invalid deployment config stops before any dependency wiring", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		cfg.Roots.DatabaseStorage.Path = "relative/postgres"

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
	})

	t.Run("bootstrap preflight failures stop before jobs, websocket, and handler construction", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)

		var bootstrapCalls int
		runBootstrap = func(ctx context.Context, cfg config.Config, pool *pgxpool.Pool) error {
			bootstrapCalls++
			return config.NewDiagnosticsError(config.Diagnostic{
				Path:       "bootstrap.first_admin_manifest_path",
				ReasonCode: "bootstrap_recovery_not_supported",
				Message:    "bootstrap completion state exists but no active deployment admin remains",
			})
		}

		jobsCalls = 0
		postgresCalls = 0
		objectStoreCalls = 0
		wsHubCalls = 0
		handlerCalls = 0

		_, err := NewRuntime(context.Background(), cfg, Options{})
		if err == nil {
			t.Fatal("expected bootstrap preflight to fail closed")
		}

		diagnosticsErr, ok := err.(*config.DiagnosticsError)
		if !ok {
			t.Fatalf("expected diagnostics error, got %T", err)
		}
		if diagnosticsErr.Code != config.InvalidDeploymentConfigCode {
			t.Fatalf("unexpected diagnostics code: got %q", diagnosticsErr.Code)
		}
		if bootstrapCalls != 1 {
			t.Fatalf("expected exactly one bootstrap preflight call, got %d", bootstrapCalls)
		}
		if postgresCalls != 1 || objectStoreCalls != 1 {
			t.Fatalf("expected startup to reach storage setup before bootstrap preflight failure, got postgres=%d object_store=%d", postgresCalls, objectStoreCalls)
		}
		if jobsCalls != 0 || wsHubCalls != 0 || handlerCalls != 0 {
			t.Fatalf("expected bootstrap preflight failure to stop before listener construction, got jobs=%d websocket=%d handler=%d", jobsCalls, wsHubCalls, handlerCalls)
		}
	})

	t.Run("startup failure closes owned object store exactly once", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		ownedStore := &phase0CloseTrackingStore{}
		setupStore = ownedStore
		runBootstrap = func(context.Context, config.Config, *pgxpool.Pool) error {
			return config.NewDiagnosticsError(config.Diagnostic{Path: "bootstrap", ReasonCode: "forced_failure", Message: "forced failure"})
		}

		if _, err := NewRuntime(context.Background(), cfg, Options{}); err == nil {
			t.Fatal("expected forced startup failure")
		}
		if ownedStore.closeCalls != 1 {
			t.Fatalf("owned object-store close calls got %d want 1", ownedStore.closeCalls)
		}
	})

	t.Run("startup failure leaves borrowed object store open", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		borrowedStore := &phase0CloseTrackingStore{}
		setupStore = nil
		runBootstrap = func(context.Context, config.Config, *pgxpool.Pool) error {
			return config.NewDiagnosticsError(config.Diagnostic{Path: "bootstrap", ReasonCode: "forced_failure", Message: "forced failure"})
		}

		if _, err := NewRuntime(context.Background(), cfg, Options{ObjectStore: borrowedStore}); err == nil {
			t.Fatal("expected forced startup failure")
		}
		if borrowedStore.closeCalls != 0 {
			t.Fatalf("borrowed object-store close calls got %d want 0", borrowedStore.closeCalls)
		}
	})
}

type phase0CloseTrackingStore struct {
	objectstore.Store
	closeCalls int
}

func (s *phase0CloseTrackingStore) Close() error {
	s.closeCalls++
	return nil
}

func phase0RuntimeConfig(t testing.TB) config.Config {
	t.Helper()

	base := t.TempDir()
	return config.Config{
		ConfigSchemaID:    "cartulary.deployment_config.v1",
		DeploymentProfile: "disconnected",
		Application: config.ApplicationConfig{
			PublicOrigin: "http://localhost:5173",
		},
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
