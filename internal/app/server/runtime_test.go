package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/secretpurpose"
	"github.com/JochiRaider/cartulary/internal/platform/securefile"
)

func TestFailClosedStartup_Unit(t *testing.T) {
	originalNewJobsManager := newJobsManager
	originalSetupPostgres := setupPostgres
	originalEnsureSchemaReady := ensureSchemaReady
	originalSetupObjectStore := setupObjectStore
	originalRunBootstrap := runBootstrap
	originalNewCollaborationHub := newCollaborationHub
	originalNewHTTPHandler := newHTTPHandler
	originalReadSecureFile := readSecureFile
	t.Cleanup(func() {
		newJobsManager = originalNewJobsManager
		setupPostgres = originalSetupPostgres
		ensureSchemaReady = originalEnsureSchemaReady
		setupObjectStore = originalSetupObjectStore
		runBootstrap = originalRunBootstrap
		newCollaborationHub = originalNewCollaborationHub
		newHTTPHandler = originalNewHTTPHandler
		readSecureFile = originalReadSecureFile
	})

	var jobsCalls int
	newJobsManager = func() *jobs.Manager {
		jobsCalls++
		return &jobs.Manager{}
	}

	var postgresCalls int
	setupPostgres = func(ctx context.Context, settings postgres.Settings) (*pgxpool.Pool, error) {
		postgresCalls++
		return nil, nil
	}
	ensureSchemaReady = func(context.Context, *pgxpool.Pool, postgres.MigrationSource) error {
		return nil
	}

	var objectStoreCalls int
	var setupStore objectstore.Store
	setupObjectStore = func(ctx context.Context, settings objectstore.Settings, instrumentation objectstore.Instrumentation) (objectstore.Store, error) {
		objectStoreCalls++
		return setupStore, nil
	}

	var wsHubCalls int
	newCollaborationHub = func() *collaboration.Hub {
		wsHubCalls++
		return collaboration.NewHub()
	}

	var handlerCalls int
	newHTTPHandler = func(options ...httpapi.Options) (http.Handler, error) {
		handlerCalls++
		return http.NewServeMux(), nil
	}

	t.Run("invalid deployment config stops before any dependency wiring", func(t *testing.T) {
		ensureSchemaReady = func(context.Context, *pgxpool.Pool, postgres.MigrationSource) error {
			t.Fatal("invalid config reached schema readiness")
			return nil
		}
		cfg := RuntimeConfig(t)
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

	t.Run("cross-purpose Revisions key reuse stops before listener publication", func(t *testing.T) {
		cfg := RuntimeConfig(t)
		t.Setenv(authn.AuthMasterKeyEnv, "pVldGSpD5oEmYa9F85d3/iL2lzBgkyfiWcoJDhsSGpk=")
		jobsCalls = 0
		postgresCalls = 0
		objectStoreCalls = 0
		wsHubCalls = 0
		handlerCalls = 0

		_, err := NewRuntime(context.Background(), cfg, Options{})
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 1 ||
			diagnostics[0].Path != "revisions.conflict_token_key_ring_manifest_path.keys[0].secret_ref" ||
			diagnostics[0].ReasonCode != "revisions_conflict_token_key_purpose_conflict" {
			t.Fatalf("cross-purpose secret diagnostics = %#v / %v", diagnostics, err)
		}
		if jobsCalls != 0 || postgresCalls != 1 || objectStoreCalls != 0 || wsHubCalls != 0 || handlerCalls != 0 {
			t.Fatalf("cross-purpose secret reuse crossed the startup preflight boundary: jobs=%d postgres=%d object_store=%d websocket=%d handler=%d", jobsCalls, postgresCalls, objectStoreCalls, wsHubCalls, handlerCalls)
		}
	})

	t.Run("inactive extension configuration stops before dependency wiring", func(t *testing.T) {
		t.Cleanup(func() { readSecureFile = originalReadSecureFile })
		cfg := RuntimeConfig(t)
		cfg.EnterpriseAuthentication.Claimed = false
		cfg.EnterpriseAuthentication.ProviderManifestPath = "/do-not-read/provider.json"

		var secureFileReads int
		readSecureFile = func(string, int64) (securefile.Document, error) {
			secureFileReads++
			t.Fatal("inactive Enterprise Authentication configuration reached secure file work")
			return securefile.Document{}, nil
		}
		jobsCalls = 0
		postgresCalls = 0
		objectStoreCalls = 0
		wsHubCalls = 0
		handlerCalls = 0
		_, err := NewRuntime(context.Background(), cfg, Options{})
		if err == nil {
			t.Fatal("expected inactive extension configuration to fail closed")
		}
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 1 ||
			diagnostics[0].Path != "enterprise_authentication.provider_manifest_path" ||
			diagnostics[0].ReasonCode != "extension_config_without_claim" {
			t.Fatalf("inactive extension diagnostics = %#v / %v", diagnostics, err)
		}
		if secureFileReads != 0 || jobsCalls != 0 || postgresCalls != 0 || objectStoreCalls != 0 || wsHubCalls != 0 || handlerCalls != 0 {
			t.Fatalf("inactive extension configuration reached effects: secure_file=%d jobs=%d postgres=%d object_store=%d websocket=%d handler=%d", secureFileReads, jobsCalls, postgresCalls, objectStoreCalls, wsHubCalls, handlerCalls)
		}
	})

	t.Run("unclaimed Network Flow configuration stops before active validation or file work", func(t *testing.T) {
		t.Cleanup(func() { readSecureFile = originalReadSecureFile })
		cfg := RuntimeConfig(t)
		cfg.NetworkFlowActivity.Claimed = false
		cfg.NetworkFlowActivity.KeyRingManifestPath = "relative/must-not-be-read.json"

		var secureFileReads int
		readSecureFile = func(string, int64) (securefile.Document, error) {
			secureFileReads++
			t.Fatal("unclaimed Network Flow configuration reached secure file work")
			return securefile.Document{}, nil
		}
		jobsCalls = 0
		postgresCalls = 0
		objectStoreCalls = 0
		wsHubCalls = 0
		handlerCalls = 0
		_, err := NewRuntime(context.Background(), cfg, Options{})
		if err == nil {
			t.Fatal("expected inactive Network Flow configuration to fail closed")
		}
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 1 {
			t.Fatalf("inactive Network Flow diagnostics = %#v / %v", diagnostics, err)
		}
		diagnostic := diagnostics[0]
		if diagnostic.Path != "network_flow_activity.key_ring_manifest_path" ||
			diagnostic.ReasonCode != "extension_config_without_claim" ||
			diagnostic.Message != "Extension configuration is present while the profile is inactive." ||
			diagnostic.Details["profile_id"] != "network_flow_activity" ||
			diagnostic.Details["config_path"] != "$.network_flow_activity.key_ring_manifest_path" {
			t.Fatalf("inactive Network Flow diagnostic = %#v", diagnostic)
		}
		if secureFileReads != 0 ||
			jobsCalls != 0 || postgresCalls != 0 || objectStoreCalls != 0 || wsHubCalls != 0 || handlerCalls != 0 {
			t.Fatalf(
				"inactive Network Flow reached effects: secure_file=%d jobs=%d postgres=%d object_store=%d websocket=%d handler=%d",
				secureFileReads,
				jobsCalls,
				postgresCalls,
				objectStoreCalls,
				wsHubCalls,
				handlerCalls,
			)
		}
	})

	t.Run("schema readiness failures stop before object store, bootstrap, jobs, websocket, and handler construction", func(t *testing.T) {
		cfg := RuntimeConfig(t)

		var schemaReadinessCalls int
		ensureSchemaReady = func(context.Context, *pgxpool.Pool, postgres.MigrationSource) error {
			schemaReadinessCalls++
			return config.NewDiagnosticsError(config.Diagnostic{
				Path:       "database.schema_version",
				ReasonCode: "schema_migration_required",
				Message:    "database schema version is behind",
			})
		}

		jobsCalls = 0
		postgresCalls = 0
		objectStoreCalls = 0
		wsHubCalls = 0
		handlerCalls = 0
		var bootstrapCalls int
		runBootstrap = func(context.Context, bootstrap.Settings, *pgxpool.Pool) error {
			bootstrapCalls++
			return nil
		}

		_, err := NewRuntime(context.Background(), cfg, Options{})
		if err == nil {
			t.Fatal("expected schema readiness failure")
		}

		diagnosticsErr, ok := err.(*config.DiagnosticsError)
		if !ok {
			t.Fatalf("expected diagnostics error, got %T", err)
		}
		if diagnosticsErr.Code != config.InvalidDeploymentConfigCode {
			t.Fatalf("unexpected diagnostics code: got %q", diagnosticsErr.Code)
		}
		if schemaReadinessCalls != 1 {
			t.Fatalf("expected one schema readiness call, got %d", schemaReadinessCalls)
		}
		if postgresCalls != 1 {
			t.Fatalf("expected startup to reach postgres setup before schema readiness failure, got postgres=%d", postgresCalls)
		}
		if objectStoreCalls != 0 || bootstrapCalls != 0 || jobsCalls != 0 || wsHubCalls != 0 || handlerCalls != 0 {
			t.Fatalf("expected schema readiness failure to stop before object store/listener construction, got object_store=%d bootstrap=%d jobs=%d websocket=%d handler=%d", objectStoreCalls, bootstrapCalls, jobsCalls, wsHubCalls, handlerCalls)
		}
	})

	t.Run("bootstrap preflight failures stop before jobs, websocket, and handler construction", func(t *testing.T) {
		ensureSchemaReady = func(context.Context, *pgxpool.Pool, postgres.MigrationSource) error {
			return nil
		}
		cfg := RuntimeConfig(t)

		var bootstrapCalls int
		runBootstrap = func(ctx context.Context, settings bootstrap.Settings, pool *pgxpool.Pool) error {
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
		ensureSchemaReady = func(context.Context, *pgxpool.Pool, postgres.MigrationSource) error {
			return nil
		}
		cfg := RuntimeConfig(t)
		ownedStore := &CloseTrackingStore{}
		setupStore = ownedStore
		runBootstrap = func(context.Context, bootstrap.Settings, *pgxpool.Pool) error {
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
		ensureSchemaReady = func(context.Context, *pgxpool.Pool, postgres.MigrationSource) error {
			return nil
		}
		cfg := RuntimeConfig(t)
		borrowedStore := &CloseTrackingStore{}
		setupStore = nil
		runBootstrap = func(context.Context, bootstrap.Settings, *pgxpool.Pool) error {
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

func TestEnterpriseAuthenticationManifestPreflight_Unit(t *testing.T) {
	originalReadSecureFile := readSecureFile
	t.Cleanup(func() { readSecureFile = originalReadSecureFile })

	t.Run("unclaimed configuration performs no file work", func(t *testing.T) {
		var reads int
		readSecureFile = func(string, int64) (securefile.Document, error) {
			reads++
			return securefile.Document{}, nil
		}
		definitions, err := loadEnterpriseProviderManifest(enterpriseauth.Configuration{
			ProviderManifestPath: "/must/not/be/read",
		}, nil)
		if err != nil || definitions != nil || reads != 0 {
			t.Fatalf("unclaimed preflight = definitions %#v, reads %d, error %v", definitions, reads, err)
		}
	})

	readSecureFile = securefile.Read
	root := t.TempDir()
	malformed := filepath.Join(root, "malformed.json")
	if err := os.WriteFile(malformed, []byte(`{"provider_manifest_schema_id":`), 0o600); err != nil {
		t.Fatal(err)
	}
	oversized := filepath.Join(root, "oversized.json")
	if err := os.WriteFile(oversized, make([]byte, enterpriseauth.ProviderManifestMaximumSize+1), 0o600); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(root, "manifest-link.json")
	if err := os.Symlink(malformed, symlink); err != nil {
		t.Fatal(err)
	}

	for name, testCase := range map[string]struct {
		path       string
		wantReason string
	}{
		"unreadable": {path: filepath.Join(root, "missing.json"), wantReason: "provider_manifest_not_readable"},
		"directory":  {path: root, wantReason: "provider_manifest_not_regular_file"},
		"oversized":  {path: oversized, wantReason: "provider_manifest_schema_invalid"},
		"symlink":    {path: symlink, wantReason: "provider_manifest_not_regular_file"},
		"malformed":  {path: malformed, wantReason: "provider_manifest_parse_error"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := loadEnterpriseProviderManifest(enterpriseauth.Configuration{
				Claimed:              true,
				ProviderManifestPath: testCase.path,
			}, nil)
			requireEnterpriseManifestDiagnostic(t, err, "enterprise_authentication.provider_manifest_path", testCase.wantReason)
			if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), testCase.path) {
				t.Fatalf("manifest error disclosed host path: %v", err)
			}
		})
	}

	for name, certificateSetup := range map[string]func(string) error{
		"certificate symlink": func(path string) error {
			target := filepath.Join(root, "certificate-target.pem")
			if err := os.WriteFile(target, []byte("not reached"), 0o600); err != nil {
				return err
			}
			return os.Symlink(target, path)
		},
		"oversized certificate": func(path string) error {
			return os.WriteFile(path, make([]byte, enterpriseauth.SigningCertificateMaximumSize+1), 0o600)
		},
	} {
		t.Run(name, func(t *testing.T) {
			certificatePath := filepath.Join(root, strings.ReplaceAll(name, " ", "-")+".pem")
			if err := certificateSetup(certificatePath); err != nil {
				t.Fatal(err)
			}
			manifestPath := filepath.Join(root, strings.ReplaceAll(name, " ", "-")+".json")
			if err := os.WriteFile(manifestPath, enterpriseSAMLManifest(certificatePath), 0o600); err != nil {
				t.Fatal(err)
			}
			_, err := loadEnterpriseProviderManifest(enterpriseauth.Configuration{
				Claimed:              true,
				ProviderManifestPath: manifestPath,
			}, nil)
			requireEnterpriseManifestDiagnostic(
				t,
				err,
				"enterprise_authentication.provider_manifest.providers[0].idp_signing_certificate_paths[0]",
				"provider_manifest_referenced_file_invalid",
			)
			if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), certificatePath) {
				t.Fatalf("certificate error disclosed host path: %v", err)
			}
		})
	}
}

func enterpriseSAMLManifest(certificatePath string) []byte {
	return []byte(fmt.Sprintf(`{
  "provider_manifest_schema_id": "cartulary.enterprise_auth_providers.v1",
  "providers": [{
    "provider_key": "corp-saml",
    "provider_type": "saml",
    "display_name": "Corporate SAML",
    "idp_entity_id": "https://idp.example.test/entity",
    "sso_url": "https://idp.example.test/sso",
    "idp_signing_certificate_paths": [%q],
    "sp_entity_id": "https://cartulary.example.test/saml/sp",
    "subject_source": { "kind": "name_id" }
  }]
}`, certificatePath))
}

func requireEnterpriseManifestDiagnostic(t testing.TB, err error, wantPath string, wantReason string) {
	t.Helper()
	diagnostics, ok := config.DiagnosticsFromError(err)
	if !ok || len(diagnostics) != 1 ||
		diagnostics[0].Path != wantPath ||
		diagnostics[0].ReasonCode != wantReason {
		t.Fatalf("enterprise manifest diagnostics = %#v / %v", diagnostics, err)
	}
}

func TestNetworkFlowManifestPreflight_Unit(t *testing.T) {
	originalReadSecureFile := readSecureFile
	t.Cleanup(func() { readSecureFile = originalReadSecureFile })

	t.Run("unclaimed configuration performs no file work", func(t *testing.T) {
		var reads int
		readSecureFile = func(string, int64) (securefile.Document, error) {
			reads++
			return securefile.Document{}, nil
		}
		rings, err := loadNetworkFlowKeyRings(
			networkflow.Configuration{KeyRingManifestPath: "/must/not/be/read"},
			nil,
			time.Time{},
			secretpurpose.NewRegistry(),
		)
		if err != nil || rings != nil || reads != 0 {
			t.Fatalf("unclaimed preflight = rings %v, reads %d, error %v", rings, reads, err)
		}
	})

	readSecureFile = securefile.Read
	root := t.TempDir()
	malformed := filepath.Join(root, "malformed.json")
	if err := os.WriteFile(malformed, []byte(`{"schema_id":`), 0o600); err != nil {
		t.Fatal(err)
	}
	oversized := filepath.Join(root, "oversized.json")
	if err := os.WriteFile(oversized, make([]byte, networkflow.KeyRingManifestMaximumSize+1), 0o600); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(root, "manifest-link.json")
	if err := os.Symlink(malformed, symlink); err != nil {
		t.Fatal(err)
	}

	for name, testCase := range map[string]struct {
		path       string
		wantReason string
	}{
		"unreadable": {path: filepath.Join(root, "missing.json"), wantReason: "network_flow_cursor_key_missing"},
		"directory":  {path: root, wantReason: "network_flow_cursor_key_invalid"},
		"oversized":  {path: oversized, wantReason: "network_flow_cursor_key_invalid"},
		"symlink":    {path: symlink, wantReason: "network_flow_cursor_key_invalid"},
		"malformed":  {path: malformed, wantReason: "network_flow_cursor_key_invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := loadNetworkFlowKeyRings(
				networkflow.Configuration{Claimed: true, KeyRingManifestPath: testCase.path},
				nil,
				time.Now().UTC(),
				secretpurpose.NewRegistry(),
			)
			diagnostics, ok := config.DiagnosticsFromError(err)
			if !ok || len(diagnostics) != 1 ||
				diagnostics[0].Path != "network_flow_activity.key_ring_manifest_path" ||
				diagnostics[0].ReasonCode != testCase.wantReason {
				t.Fatalf("manifest diagnostics = %#v / %v", diagnostics, err)
			}
			if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), testCase.path) {
				t.Fatalf("manifest error disclosed host path: %v", err)
			}
		})
	}
}

func TestRevisionsConflictTokenManifestPreflight_Unit(t *testing.T) {
	originalReadSecureFile := readSecureFile
	t.Cleanup(func() { readSecureFile = originalReadSecureFile })
	readSecureFile = securefile.Read

	root := t.TempDir()
	malformed := filepath.Join(root, "malformed.json")
	if err := os.WriteFile(malformed, []byte(`{"schema_id":`), 0o600); err != nil {
		t.Fatal(err)
	}
	oversized := filepath.Join(root, "oversized.json")
	if err := os.WriteFile(oversized, make([]byte, conflicts.ConflictTokenKeyRingManifestMaximumSize+1), 0o600); err != nil {
		t.Fatal(err)
	}
	symlink := filepath.Join(root, "manifest-link.json")
	if err := os.Symlink(malformed, symlink); err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		"CARTULARY_SECRET_RUNTIME_TEST_REVISIONS": "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY",
	}

	for name, testCase := range map[string]struct {
		path       string
		wantReason string
	}{
		"unreadable": {path: filepath.Join(root, "missing.json"), wantReason: "revisions_conflict_token_manifest_missing"},
		"directory":  {path: root, wantReason: "revisions_conflict_token_manifest_invalid"},
		"oversized":  {path: oversized, wantReason: "revisions_conflict_token_manifest_invalid"},
		"symlink":    {path: symlink, wantReason: "revisions_conflict_token_manifest_invalid"},
		"malformed":  {path: malformed, wantReason: "revisions_conflict_token_manifest_invalid"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := loadRevisionsConflictTokenKeyRing(
				conflicts.Configuration{ConflictTokenKeyRingManifestPath: testCase.path},
				env,
				time.Now().UTC(),
				secretpurpose.NewRegistry(),
			)
			diagnostics, ok := config.DiagnosticsFromError(err)
			if !ok || len(diagnostics) != 1 ||
				diagnostics[0].Path != "revisions.conflict_token_key_ring_manifest_path" ||
				diagnostics[0].ReasonCode != testCase.wantReason {
				t.Fatalf("manifest diagnostics = %#v / %v", diagnostics, err)
			}
			if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), testCase.path) {
				t.Fatalf("manifest error disclosed host path: %v", err)
			}
		})
	}
}

type CloseTrackingStore struct {
	objectstore.Store
	closeCalls int
}

func (s *CloseTrackingStore) Close() error {
	s.closeCalls++
	return nil
}

func RuntimeConfig(t testing.TB) configassembly.Deployment {
	t.Helper()

	base := t.TempDir()
	databaseRoot := filepath.Join(base, "postgres")
	if err := os.MkdirAll(databaseRoot, 0o700); err != nil {
		t.Fatalf("create database root fixture: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(databaseRoot, postgres.FilesystemRootDSNFile),
		[]byte("postgres://unit-test"),
		0o600,
	); err != nil {
		t.Fatalf("write database DSN fixture: %v", err)
	}
	conflictTokenManifestPath := filepath.Join(base, "revisions-conflict-token-key-ring.json")
	if err := os.WriteFile(conflictTokenManifestPath, []byte(`{"schema_id":"cartulary.revisions_conflict_token_key_ring.v1","algorithm":"aes_256_gcm_v1","keys":[{"conflict_token_key_id":"runtime-test","state":"active","secret_ref":{"kind":"env","name":"runtime-test-revisions-conflict"}}]}`), 0o600); err != nil {
		t.Fatalf("write Revisions conflict-token key-ring fixture: %v", err)
	}
	t.Setenv("CARTULARY_SECRET_RUNTIME_TEST_REVISIONS_CONFLICT", "pVldGSpD5oEmYa9F85d3_iL2lzBgkyfiWcoJDhsSGpk")
	return configassembly.Deployment{
		ConfigSchemaID:    "cartulary.deployment_config.v2",
		DeploymentProfile: "disconnected",
		Application: config.ApplicationConfig{
			PublicOrigin: "http://localhost:5173",
		},
		Roots: config.RootBindings{
			DatabaseStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        databaseRoot,
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
		Revisions: conflicts.Configuration{ConflictTokenKeyRingManifestPath: conflictTokenManifestPath},
	}
}
