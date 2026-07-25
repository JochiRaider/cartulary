package bootstrap

import (
	"context"
	"io/fs"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/diagnosticstest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestBootstrapManifestValidation_Unit(t *testing.T) {
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

	t.Run("rejects missing bootstrap_schema_id", func(t *testing.T) {
		err := parseBootstrapManifestError(t, `{"bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!"}`)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest.bootstrap_schema_id", "bootstrap_manifest_schema_invalid")
	})

	t.Run("rejects client-chosen deployment-admin state", func(t *testing.T) {
		err := parseBootstrapManifestError(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","is_deployment_admin":true}`)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest.is_deployment_admin", "bootstrap_manifest_schema_invalid")
	})

	t.Run("rejects incident membership, provider binding, and credential lifecycle fields", func(t *testing.T) {
		cases := []struct {
			name    string
			content string
			path    string
		}{
			{
				name:    "incident membership",
				content: `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","incident_memberships":[{"incident_id":"11111111-1111-1111-1111-111111111111","role":"admin"}]}`,
				path:    "bootstrap.first_admin_manifest.incident_memberships",
			},
			{
				name:    "provider identity",
				content: `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","provider_subject":"oidc-subject"}`,
				path:    "bootstrap.first_admin_manifest.provider_subject",
			},
			{
				name:    "secret bearing extra",
				content: `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","totp_secret_ciphertext":"opaque-secret"}`,
				path:    "bootstrap.first_admin_manifest.totp_secret_ciphertext",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				err := parseBootstrapManifestError(t, tc.content)
				requireBootstrapDiagnostic(t, err, tc.path, "bootstrap_manifest_schema_invalid")
			})
		}
	})

	t.Run("rejects unsupported bootstrap schema identifiers", func(t *testing.T) {
		err := parseBootstrapManifestError(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v2","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!"}`)
		requireBootstrapDiagnostic(t, err, "bootstrap.first_admin_manifest.bootstrap_schema_id", "bootstrap_manifest_schema_invalid")
	})

	t.Run("rejects malformed manifest field values", func(t *testing.T) {
		cases := []struct {
			name    string
			content string
			path    string
		}{
			{
				name:    "email with whitespace",
				content: `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!"}`,
				path:    "bootstrap.first_admin_manifest.email",
			},
			{
				name:    "display name empty after trim",
				content: `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"   ","initial_password":"BootstrapPass1!"}`,
				path:    "bootstrap.first_admin_manifest.display_name",
			},
			{
				name:    "password too short",
				content: `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"short"}`,
				path:    "bootstrap.first_admin_manifest.initial_password",
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				err := parseBootstrapManifestError(t, tc.content)
				requireBootstrapDiagnostic(t, err, tc.path, "bootstrap_manifest_schema_invalid")
			})
		}
	})
}

func TestBootstrapPreflight_Unit(t *testing.T) {
	t.Run("skips manifest consumption when an active deployment admin already exists", func(t *testing.T) {
		store := &bootstrapStoreStub{
			state: bootstrapState{
				ActiveDeploymentAdmins: 1,
			},
		}
		manifestFS := &bootstrapManifestFSStub{}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/stale-bootstrap.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		if err != nil {
			t.Fatalf("bootstrap preflight with existing admin: %v", err)
		}
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if manifestFS.statCalls != 0 || manifestFS.readCalls != 0 {
			t.Fatalf("expected manifest filesystem access to be skipped, got stat=%d read=%d", manifestFS.statCalls, manifestFS.readCalls)
		}
		if store.createCalls != 0 {
			t.Fatalf("expected no bootstrap create call, got %d", store.createCalls)
		}
	})

	t.Run("skips stale and invalid manifests when an active deployment admin already exists", func(t *testing.T) {
		store := &bootstrapStoreStub{
			state: bootstrapState{
				ActiveDeploymentAdmins: 1,
			},
		}
		manifestFS := &bootstrapManifestFSStub{}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/invalid-bootstrap.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		if err != nil {
			t.Fatalf("bootstrap preflight with existing admin and invalid manifest: %v", err)
		}
		if manifestFS.statCalls != 0 || manifestFS.readCalls != 0 {
			t.Fatalf("expected manifest filesystem access to be skipped for invalid configured content, got stat=%d read=%d", manifestFS.statCalls, manifestFS.readCalls)
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
		manifestFS := &bootstrapManifestFSStub{}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/bootstrap.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		requireBootstrapDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", "bootstrap_recovery_not_supported.json"})
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if manifestFS.statCalls != 0 || manifestFS.readCalls != 0 {
			t.Fatalf("expected manifest filesystem access to be skipped during fail-closed recovery, got stat=%d read=%d", manifestFS.statCalls, manifestFS.readCalls)
		}
		if store.createCalls != 0 {
			t.Fatalf("expected no bootstrap create call, got %d", store.createCalls)
		}
	})

	t.Run("requires a configured manifest path when bootstrap is still needed", func(t *testing.T) {
		store := &bootstrapStoreStub{}
		manifestFS := &bootstrapManifestFSStub{}
		err := bootstrapPreflight(context.Background(), Settings{}, store, manifestFS, deriveBootstrapPasswordHash)
		requireBootstrapDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", "bootstrap_manifest_path_missing.json"})
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if manifestFS.statCalls != 0 || manifestFS.readCalls != 0 {
			t.Fatalf("expected manifest filesystem access to be skipped when path is missing, got stat=%d read=%d", manifestFS.statCalls, manifestFS.readCalls)
		}
		if store.createCalls != 0 {
			t.Fatalf("expected no bootstrap create call, got %d", store.createCalls)
		}
	})

	t.Run("returns a whole-payload golden for unreadable manifests via injected permission failure", func(t *testing.T) {
		store := &bootstrapStoreStub{}
		manifestFS := &bootstrapManifestFSStub{
			statInfo: stubFileInfo{mode: 0},
			readErr:  fs.ErrPermission,
		}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/bootstrap-admin.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		requireBootstrapDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", "bootstrap_manifest_not_readable_permission_denied.json"})
		if manifestFS.statCalls != 1 || manifestFS.readCalls != 1 {
			t.Fatalf("expected one manifest stat and one read, got stat=%d read=%d", manifestFS.statCalls, manifestFS.readCalls)
		}
	})

	t.Run("returns a whole-payload golden for non-regular manifests", func(t *testing.T) {
		store := &bootstrapStoreStub{}
		manifestFS := &bootstrapManifestFSStub{
			statInfo: stubFileInfo{mode: fs.ModeDir},
		}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/bootstrap-admin.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		requireBootstrapDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", "bootstrap_manifest_not_regular_file.json"})
		if manifestFS.statCalls != 1 || manifestFS.readCalls != 0 {
			t.Fatalf("expected one manifest stat and zero reads, got stat=%d read=%d", manifestFS.statCalls, manifestFS.readCalls)
		}
	})

	t.Run("returns a whole-payload golden for malformed bootstrap manifests", func(t *testing.T) {
		store := &bootstrapStoreStub{}
		manifestFS := &bootstrapManifestFSStub{
			statInfo: stubFileInfo{mode: 0},
			readData: []byte(`{"bootstrap_schema_id":`),
		}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/bootstrap-admin.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		requireBootstrapDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", "bootstrap_manifest_parse_error.json"})
	})

	t.Run("returns canonically ordered schema-invalid diagnostics", func(t *testing.T) {
		store := &bootstrapStoreStub{}
		manifestFS := &bootstrapManifestFSStub{
			statInfo: stubFileInfo{mode: 0},
			readData: []byte(`{"display_name":"   ","initial_password":"short","unexpected":"surprise"}`),
		}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/bootstrap-admin.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		requireBootstrapDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", "bootstrap_manifest_schema_invalid_multiple.json"})
	})

	t.Run("consumes the configured manifest when bootstrap is required", func(t *testing.T) {
		store := &bootstrapStoreStub{}
		manifestFS := &bootstrapManifestFSStub{
			statInfo: stubFileInfo{mode: 0},
			readData: fixtures.MustRead("bootstrap-admin", "canonical.json"),
		}
		err := bootstrapPreflight(context.Background(), Settings{
			ManifestPath: "/tmp/bootstrap-admin.json",
		}, store, manifestFS, deriveBootstrapPasswordHash)
		if err != nil {
			t.Fatalf("bootstrap preflight with valid manifest: %v", err)
		}
		if store.readCalls != 1 {
			t.Fatalf("expected exactly one bootstrap-state query, got %d", store.readCalls)
		}
		if store.createCalls != 1 {
			t.Fatalf("expected one bootstrap create call, got %d", store.createCalls)
		}
		if manifestFS.statCalls != 1 || manifestFS.readCalls != 1 {
			t.Fatalf("expected one manifest stat and one read, got stat=%d read=%d", manifestFS.statCalls, manifestFS.readCalls)
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

func requireBootstrapDiagnosticsMatchGolden(t testing.TB, err error, goldenParts []string) {
	t.Helper()

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
	if !ok {
		t.Fatalf("expected diagnostics error, got %T", err)
	}
	diagnosticstest.RequireJSONMatchesGolden(t, diagnosticsErr.JSON(), goldenParts)
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

type bootstrapManifestFSStub struct {
	statInfo  fs.FileInfo
	statErr   error
	readData  []byte
	readErr   error
	statCalls int
	readCalls int
}

func (s *bootstrapManifestFSStub) Stat(name string) (fs.FileInfo, error) {
	s.statCalls++
	if s.statErr != nil {
		return nil, s.statErr
	}
	return s.statInfo, nil
}

func (s *bootstrapManifestFSStub) ReadFile(name string) ([]byte, error) {
	s.readCalls++
	if s.readErr != nil {
		return nil, s.readErr
	}
	return append([]byte(nil), s.readData...), nil
}

type stubFileInfo struct {
	mode fs.FileMode
}

func (stubFileInfo) Name() string        { return "bootstrap-admin.json" }
func (stubFileInfo) Size() int64         { return 0 }
func (f stubFileInfo) Mode() fs.FileMode { return f.mode }
func (stubFileInfo) ModTime() time.Time  { return time.Time{} }
func (f stubFileInfo) IsDir() bool       { return f.mode.IsDir() }
func (stubFileInfo) Sys() any            { return nil }
