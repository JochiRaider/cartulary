package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"example.com/todo/cartulary/internal/platform/config"
	"example.com/todo/cartulary/internal/testutil/fixtures"
)

func TestPhase0_BootstrapManifestValidation_U_0_07(t *testing.T) {
	t.Run("accepts canonical manifest and defaults omitted mfa_required to true", func(t *testing.T) {
		manifest, err := parseBootstrapManifest(fixtures.MustRead("bootstrap-admin", "canonical.json"))
		if err != nil {
			t.Fatalf("parse canonical bootstrap manifest: %v", err)
		}

		if manifest.BootstrapSchemaID != bootstrapManifestSchemaID {
			t.Fatalf("unexpected bootstrap schema id: got %q", manifest.BootstrapSchemaID)
		}
		if manifest.BootstrapArtifactID == uuid.Nil {
			t.Fatal("expected bootstrap artifact id")
		}
		if manifest.Email != "bootstrap-admin@example.test" {
			t.Fatalf("unexpected normalized email: got %q", manifest.Email)
		}
		if manifest.DisplayName != "Bootstrap Admin" {
			t.Fatalf("unexpected normalized display name: got %q", manifest.DisplayName)
		}
		if !manifest.MFARequired {
			t.Fatal("expected omitted mfa_required to default to true")
		}
	})

	t.Run("rejects explicit false mfa_required", func(t *testing.T) {
		err := parseBootstrapManifestError(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","mfa_required":false}`)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest.mfa_required", "bootstrap_manifest_schema_invalid")
	})

	t.Run("rejects unknown top-level members", func(t *testing.T) {
		err := parseBootstrapManifestError(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","unexpected":"surprise"}`)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest.unexpected", "bootstrap_manifest_schema_invalid")
	})

	t.Run("rejects client-chosen deployment-admin state", func(t *testing.T) {
		err := parseBootstrapManifestError(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","is_deployment_admin":true}`)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest.is_deployment_admin", "bootstrap_manifest_schema_invalid")
	})

	t.Run("rejects unsupported bootstrap schema identifiers", func(t *testing.T) {
		err := parseBootstrapManifestError(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v2","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!"}`)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest.bootstrap_schema_id", "bootstrap_manifest_schema_invalid")
	})
}

func TestPhase0_BootstrapPreflight_U_0_08(t *testing.T) {
	t.Run("rejects lexically invalid configured bootstrap manifest paths before startup", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		cfg.Bootstrap.FirstAdminManifestPath = "relative/bootstrap-admin.json"

		_, err := config.Validate(cfg)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest_path", "path_not_absolute")
	})

	t.Run("skips manifest consumption when an active deployment admin already exists", func(t *testing.T) {
		store := &bootstrapStoreStub{
			state: bootstrapState{
				ActiveDeploymentAdmins: 1,
			},
		}

		var readCalls int
		err := bootstrapPreflight(context.Background(), config.Config{
			Bootstrap: config.BootstrapConfig{
				FirstAdminManifestPath: "/tmp/stale-bootstrap.json",
			},
		}, store, func(path string) ([]byte, error) {
			readCalls++
			return nil, errors.New("unexpected manifest read")
		}, deriveBootstrapPasswordHash)
		if err != nil {
			t.Fatalf("bootstrap preflight with existing admin: %v", err)
		}
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if readCalls != 0 {
			t.Fatalf("expected manifest reads to be skipped, got %d", readCalls)
		}
		if store.createCalls != 0 {
			t.Fatalf("expected no bootstrap create call, got %d", store.createCalls)
		}
	})

	t.Run("fails closed when bootstrap completion state exists but no active admin remains", func(t *testing.T) {
		store := &bootstrapStoreStub{
			state: bootstrapState{
				BootstrapCompleted: true,
			},
		}

		var readCalls int
		err := bootstrapPreflight(context.Background(), config.Config{
			Bootstrap: config.BootstrapConfig{
				FirstAdminManifestPath: "/tmp/bootstrap.json",
			},
		}, store, func(path string) ([]byte, error) {
			readCalls++
			return fixtures.MustRead("bootstrap-admin", "canonical.json"), nil
		}, deriveBootstrapPasswordHash)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest_path", "bootstrap_recovery_not_supported")
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if readCalls != 0 {
			t.Fatalf("expected manifest reads to be skipped during fail-closed recovery, got %d", readCalls)
		}
		if store.createCalls != 0 {
			t.Fatalf("expected no bootstrap create call, got %d", store.createCalls)
		}
	})

	t.Run("requires a configured manifest path when bootstrap is still needed", func(t *testing.T) {
		store := &bootstrapStoreStub{}

		err := bootstrapPreflight(context.Background(), config.Config{}, store, func(path string) ([]byte, error) {
			return fixtures.MustRead("bootstrap-admin", "canonical.json"), nil
		}, deriveBootstrapPasswordHash)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest_path", "bootstrap_manifest_path_missing")
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if store.createCalls != 0 {
			t.Fatalf("expected no bootstrap create call, got %d", store.createCalls)
		}
	})

	t.Run("consumes the configured manifest when bootstrap is required", func(t *testing.T) {
		store := &bootstrapStoreStub{}
		manifestPath := filepath.Join(t.TempDir(), "bootstrap-admin.json")
		if err := os.WriteFile(manifestPath, fixtures.MustRead("bootstrap-admin", "canonical.json"), 0o644); err != nil {
			t.Fatalf("write bootstrap manifest: %v", err)
		}

		err := bootstrapPreflight(context.Background(), config.Config{
			Bootstrap: config.BootstrapConfig{
				FirstAdminManifestPath: manifestPath,
			},
		}, store, os.ReadFile, deriveBootstrapPasswordHash)
		if err != nil {
			t.Fatalf("bootstrap preflight with valid manifest: %v", err)
		}
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if store.createCalls != 1 {
			t.Fatalf("expected one bootstrap create call, got %d", store.createCalls)
		}
		if store.created == nil {
			t.Fatal("expected bootstrap create request")
		}
		if !store.created.Manifest.MFARequired {
			t.Fatal("expected bootstrap create request to carry mfa_required=true")
		}
		if len(store.created.ArtifactSHA256) == 0 {
			t.Fatal("expected manifest artifact hash")
		}
		if store.created.PasswordHash == "" {
			t.Fatal("expected persisted password hash")
		}
	})
}

type bootstrapStoreStub struct {
	state       bootstrapState
	stateErr    error
	createErr   error
	readCalls   int
	createCalls int
	created     *bootstrapCreateRequest
}

func (s *bootstrapStoreStub) ReadBootstrapState(ctx context.Context) (bootstrapState, error) {
	s.readCalls++
	return s.state, s.stateErr
}

func (s *bootstrapStoreStub) CreateBootstrapAdmin(ctx context.Context, request bootstrapCreateRequest) error {
	s.createCalls++
	s.created = &request
	return s.createErr
}

func parseBootstrapManifestError(t testing.TB, content string) error {
	t.Helper()

	_, err := parseBootstrapManifest([]byte(content))
	if err == nil {
		t.Fatal("expected bootstrap manifest parse to fail")
	}

	return err
}

func requireBootstrapDiagnostic(t testing.TB, err error, wantPath string, wantReason string) {
	t.Helper()

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
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
