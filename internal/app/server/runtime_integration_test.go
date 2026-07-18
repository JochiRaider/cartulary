package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/bootstraptest"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/securityassert"
)

func TestInvalidConfigNeverReachesReady_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-invalid-config")

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "bootstrap-invalid-config")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := IntegrationEnv(testDB.Env(), s3Harness.Env(bucket))
	cases := []struct {
		name       string
		mutate     func(config.Config) config.Config
		goldenFile string
	}{
		{
			name: "path-validation failure",
			mutate: func(cfg config.Config) config.Config {
				cfg.Roots.DatabaseStorage.Path = "relative/postgres"
				return cfg
			},
			goldenFile: "startup_path_not_absolute_database_storage_root.json",
		},
		{
			name: "missing required runtime root",
			mutate: func(cfg config.Config) config.Config {
				cfg.Roots.ExportOutputs = config.RootBinding{}
				return cfg
			},
			goldenFile: "startup_missing_export_outputs_root.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			counters := installStartupCounters(t)
			cfg := BindPostgres(t, tc.mutate(RuntimeConfig(t)), env)

			_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
			configtest.RequireDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", tc.goldenFile})
			counters.RequireNotStarted(t)
		})
	}

}

func TestFirstAdminBootstrap_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("commits one deployment admin, bootstrap marker, and startup audit before readiness", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-bootstrap-success")

		db := openSQL(t, testDB.DSN)
		defer db.Close()

		bucket := BucketName("bootstrap-bootstrap-success")
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := IntegrationEnv(testDB.Env(), s3Harness.Env(bucket))
		cfg := BindPostgres(t, RuntimeConfig(t), env)
		cfg.Bootstrap.FirstAdminManifestPath = fixtures.Path("bootstrap-admin", "canonical.json")

		runtime, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		if err != nil {
			t.Fatalf("start runtime with canonical bootstrap manifest: %v", err)
		}
		defer runtime.Close()

		requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)

		var userID string
		var email string
		var mfaRequired bool
		var passwordHash string
		if err := db.QueryRowContext(context.Background(), `SELECT id::text, email, mfa_required, password_hash FROM users WHERE is_active = true AND is_deployment_admin = true`).Scan(&userID, &email, &mfaRequired, &passwordHash); err != nil {
			t.Fatalf("query bootstrap-created user: %v", err)
		}
		if email != "bootstrap-admin@example.test" {
			t.Fatalf("unexpected bootstrap-created email: got %q", email)
		}
		if !mfaRequired {
			t.Fatal("expected bootstrap-created user to require MFA")
		}
		if passwordHash == "" || strings.Contains(passwordHash, "BootstrapPass1!") {
			t.Fatalf("expected persisted password hash without cleartext secret, got %q", passwordHash)
		}
		bootstraptest.RequireBootstrapUserLocalAuthOnly(t, testDB.DSN, userID, email)

		audit := lookupBootstrapAuditEvent(t, db)
		auditassert.RequireSystemMutationAttribution(t, auditassert.SystemMutationAttribution{
			ActorUserID: audit.ActorUserID,
			Source:      audit.EventSource,
			EventKind:   audit.EventKind,
			RequestID:   audit.RequestID,
			CreatedAt:   audit.CreatedAt,
		}, "bootstrap_manifest", "bootstrap_admin_created")
		if audit.TargetUserID != userID {
			t.Fatalf("unexpected startup audit target_user_id: got %q want %q", audit.TargetUserID, userID)
		}
		securityassert.RequireSecretSafePayload(t, audit.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32", "provider_subject", "provider_key"})
		if got := audit.After["email"]; got != email {
			t.Fatalf("unexpected startup audit email: got %v want %q", got, email)
		}
		if got := audit.After["mfa_required"]; got != true {
			t.Fatalf("unexpected startup audit MFA payload: %#v", audit.After)
		}

		server := httptest.NewServer(runtime.Handler)
		defer server.Close()
		resp, err := http.Get(server.URL + "/readyz")
		if err != nil {
			t.Fatalf("probe readyz after bootstrap commit: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected readyz status after bootstrap commit: got %d want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("rolls back the whole bootstrap transaction when the audit insert fails", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-bootstrap-rollback")

		db := openSQL(t, testDB.DSN)
		defer db.Close()
		if _, err := db.ExecContext(context.Background(), `
CREATE OR REPLACE FUNCTION bootstrap_fail_bootstrap_audit() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'bootstrap forced bootstrap audit failure';
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER bootstrap_fail_bootstrap_audit
BEFORE INSERT ON deployment_admin_audit_events
FOR EACH ROW
EXECUTE FUNCTION bootstrap_fail_bootstrap_audit();
`); err != nil {
			t.Fatalf("install bootstrap rollback trigger: %v", err)
		}

		bucket := BucketName("bootstrap-bootstrap-rollback")
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := IntegrationEnv(testDB.Env(), s3Harness.Env(bucket))
		cfg := BindPostgres(t, RuntimeConfig(t), env)
		cfg.Bootstrap.FirstAdminManifestPath = fixtures.Path("bootstrap-admin", "canonical.json")

		counters := installStartupCounters(t)
		_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		requireBootstrapReason(t, err, "bootstrap_persist_failed")
		counters.RequireNotStarted(t)

		requireNoBootstrapSideEffects(t, db)
	})
}

func TestBootstrapFailures_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	cases := []struct {
		name          string
		manifestPath  func(t *testing.T) string
		seed          func(t *testing.T, db *sql.DB)
		goldenFile    string
		wantUserCount int
	}{
		{
			name:       "missing configured bootstrap path",
			goldenFile: "bootstrap_manifest_path_missing.json",
		},
		{
			name: "unreadable regular bootstrap manifest",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteUnreadableBootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_not_readable_permission_denied.json",
		},
		{
			name: "non regular bootstrap path",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteNonRegularBootstrapManifestPath(t)
			},
			goldenFile: "bootstrap_manifest_not_regular_file.json",
		},
		{
			name: "malformed json manifest",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteMalformedBootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_parse_error.json",
		},
		{
			name: "wrong schema id manifest",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteWrongSchemaBootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_schema_invalid_wrong_schema.json",
		},
		{
			name: "explicit false mfa manifest",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteExplicitFalseMFABootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_schema_invalid_explicit_false_mfa.json",
		},
		{
			name: "unknown top level members",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteUnknownMemberBootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_schema_invalid_unknown_member.json",
		},
		{
			name: "forbidden incident membership fields",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteForbiddenIncidentMembershipBootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_schema_invalid_forbidden_incident_memberships.json",
		},
		{
			name: "forbidden provider binding fields",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteForbiddenProviderBootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_schema_invalid_forbidden_provider_subject.json",
		},
		{
			name: "forbidden client chosen admin fields",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.WriteForbiddenDeploymentAdminBootstrapManifest(t)
			},
			goldenFile: "bootstrap_manifest_schema_invalid_forbidden_deployment_admin.json",
		},
		{
			name: "email conflict",
			manifestPath: func(t *testing.T) string {
				return bootstraptest.CanonicalBootstrapManifestPath()
			},
			seed: func(t *testing.T, db *sql.DB) {
				bootstraptest.SeedBootstrapEmailConflict(t, db)
			},
			goldenFile:    "bootstrap_email_conflict.json",
			wantUserCount: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-bootstrap-failure")

			db := openSQL(t, testDB.DSN)
			defer db.Close()
			if tc.seed != nil {
				tc.seed(t, db)
			}

			bucket := BucketName(tc.name)
			defer func() {
				if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
					t.Logf("cleanup bucket: %v", err)
				}
			}()

			env := IntegrationEnv(testDB.Env(), s3Harness.Env(bucket))
			cfg := BindPostgres(t, RuntimeConfig(t), env)
			if tc.manifestPath != nil {
				cfg.Bootstrap.FirstAdminManifestPath = tc.manifestPath(t)
			}

			counters := installStartupCounters(t)
			_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
			configtest.RequireDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", tc.goldenFile})
			counters.RequireNotStarted(t)

			requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, tc.wantUserCount)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
		})
	}

	t.Run("startup failure leaves a borrowed postgres pool open", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-borrowed-postgres")
		pool, err := pgxpool.New(context.Background(), testDB.DSN)
		if err != nil {
			t.Fatalf("open borrowed postgres pool: %v", err)
		}
		defer pool.Close()

		bucket := BucketName("bootstrap-borrowed-postgres")
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := IntegrationEnv(testDB.Env(), s3Harness.Env(bucket))
		cfg := BindPostgres(t, RuntimeConfig(t), env)
		if _, err := NewRuntime(context.Background(), cfg, Options{Env: env, Postgres: pool}); err == nil {
			t.Fatal("expected missing bootstrap manifest to fail startup")
		}
		if err := pool.Ping(context.Background()); err != nil {
			t.Fatalf("borrowed postgres pool was closed after startup failure: %v", err)
		}
	})
}

func TestBootstrapSkipAndRecovery_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("existing active deployment admin skips stale and invalid manifests", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-bootstrap-skip")

		db := openSQL(t, testDB.DSN)
		defer db.Close()
		if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, true, true)`, "existing-admin@example.test", "Existing Admin", "existing-hash"); err != nil {
			t.Fatalf("seed active deployment admin: %v", err)
		}

		bucket := BucketName("bootstrap-bootstrap-skip")
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := IntegrationEnv(testDB.Env(), s3Harness.Env(bucket))
		cases := []struct {
			name         string
			manifestPath string
		}{
			{
				name:         "stale manifest path",
				manifestPath: filepath.Join(t.TempDir(), "missing-bootstrap.json"),
			},
			{
				name:         "invalid manifest content",
				manifestPath: bootstraptest.WriteExplicitFalseMFABootstrapManifest(t),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				cfg := BindPostgres(t, RuntimeConfig(t), env)
				cfg.Bootstrap.FirstAdminManifestPath = tc.manifestPath

				runtime, err := NewRuntime(context.Background(), cfg, Options{Env: env})
				if err != nil {
					t.Fatalf("start runtime with existing deployment admin: %v", err)
				}
				defer runtime.Close()

				requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, 1)
				requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
				requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
				requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
			})
		}
	})

	t.Run("bootstrap recovery remains fail-closed when completion state exists without an active admin", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-bootstrap-recovery")

		db := openSQL(t, testDB.DSN)
		defer db.Close()

		var userID string
		if err := db.QueryRowContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, false, true) RETURNING id`, "retired-admin@example.test", "Retired Admin", "existing-hash").Scan(&userID); err != nil {
			t.Fatalf("seed retired deployment admin: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), `INSERT INTO deployment_bootstrap_state (slot, bootstrap_schema_id, bootstrap_artifact_id, artifact_sha256, created_user_id) VALUES ('first_deployment_admin', $1, $2, $3, $4)`, bootstrap.ManifestSchemaID, "22222222-2222-2222-2222-222222222222", []byte{0x01, 0x02, 0x03}, userID); err != nil {
			t.Fatalf("seed bootstrap completion marker: %v", err)
		}

		bucket := BucketName("bootstrap-bootstrap-recovery")
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := IntegrationEnv(testDB.Env(), s3Harness.Env(bucket))
		cfg := BindPostgres(t, RuntimeConfig(t), env)
		cfg.Bootstrap.FirstAdminManifestPath = fixtures.Path("bootstrap-admin", "canonical.json")

		counters := installStartupCounters(t)
		_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		configtest.RequireDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", "bootstrap_recovery_not_supported.json"})
		counters.RequireNotStarted(t)

		requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 0)
	})
}

type StartupCounters struct {
	jobsManager int
	wsHub       int
	httpHandler int
}

type AuditEvent struct {
	ActorUserID  string
	TargetUserID string
	EventSource  string
	EventKind    string
	RequestID    string
	CreatedAt    time.Time
	After        map[string]any
}

func installStartupCounters(t testing.TB) *StartupCounters {
	t.Helper()

	counters := &StartupCounters{}
	originalJobsManager := newJobsManager
	originalWSHub := newWSHub
	originalHTTPHandler := newHTTPHandler

	newJobsManager = func() *jobs.Manager {
		counters.jobsManager++
		return originalJobsManager()
	}
	newWSHub = func() *platformws.Hub {
		counters.wsHub++
		return platformws.NewHub()
	}
	newHTTPHandler = func(options ...httpapi.Options) (http.Handler, error) {
		counters.httpHandler++
		return originalHTTPHandler(options...)
	}

	t.Cleanup(func() {
		newJobsManager = originalJobsManager
		newWSHub = originalWSHub
		newHTTPHandler = originalHTTPHandler
	})

	return counters
}

func (c *StartupCounters) RequireNotStarted(t testing.TB) {
	t.Helper()
	if c.jobsManager != 0 || c.wsHub != 0 || c.httpHandler != 0 {
		t.Fatalf("expected listeners and job shells to remain unstarted, got jobs=%d websocket=%d handler=%d", c.jobsManager, c.wsHub, c.httpHandler)
	}
}

func requireBootstrapReason(t testing.TB, err error, wantReasonCode string) {
	t.Helper()

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
	if !ok {
		t.Fatalf("expected diagnostics error, got %T", err)
	}
	for _, diagnostic := range diagnosticsErr.Diagnostics {
		if diagnostic.ReasonCode == wantReasonCode {
			return
		}
	}
	t.Fatalf("missing bootstrap reason_code=%q in %#v", wantReasonCode, diagnosticsErr.Diagnostics)
}

func lookupBootstrapAuditEvent(t testing.TB, db *sql.DB) AuditEvent {
	t.Helper()

	var actorUserID sql.NullString
	var targetUserID string
	var eventSource string
	var eventKind string
	var requestID sql.NullString
	var createdAt time.Time
	var afterJSON []byte
	if err := db.QueryRowContext(context.Background(), `
SELECT COALESCE(actor_user_id::text, ''),
       target_user_id::text,
       event_source,
       event_kind,
       COALESCE(request_id, ''),
       created_at,
       after_json
  FROM deployment_admin_audit_events
 ORDER BY created_at ASC
 LIMIT 1
`).Scan(&actorUserID, &targetUserID, &eventSource, &eventKind, &requestID, &createdAt, &afterJSON); err != nil {
		t.Fatalf("query bootstrap audit event: %v", err)
	}

	event := AuditEvent{
		ActorUserID:  actorUserID.String,
		TargetUserID: targetUserID,
		EventSource:  eventSource,
		EventKind:    eventKind,
		RequestID:    requestID.String,
		CreatedAt:    createdAt,
		After:        map[string]any{},
	}
	if len(afterJSON) > 0 {
		if err := json.Unmarshal(afterJSON, &event.After); err != nil {
			t.Fatalf("decode bootstrap audit after_json: %v", err)
		}
	}

	return event
}

func requireNoBootstrapSideEffects(t testing.TB, db *sql.DB) {
	t.Helper()
	requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, 0)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
}

func IntegrationEnv(databaseEnv map[string]string, objectStoreEnv map[string]string) map[string]string {
	env := make(map[string]string, len(databaseEnv)+len(objectStoreEnv))
	for key, value := range databaseEnv {
		env[key] = value
	}
	for key, value := range objectStoreEnv {
		env[key] = value
	}
	return env
}

func BindPostgres(t testing.TB, cfg config.Config, env map[string]string) config.Config {
	t.Helper()
	configtest.BindPostgresEnvToDatabaseRoot(t, cfg.Roots.DatabaseStorage.Path, env)
	return cfg
}

func openSQL(t testing.TB, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	return db
}

func requireCountSQL(t testing.TB, db *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}

func BucketName(prefix string) string {
	value := strings.ToLower(prefix)
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return fmt.Sprintf("%s-%d", value, time.Now().UnixNano())
}
