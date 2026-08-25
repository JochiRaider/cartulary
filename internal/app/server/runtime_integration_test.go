package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/bootstraptest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/processlease"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/securityassert"
)

func TestInvalidConfigNeverReachesReady_Integration(t *testing.T) {
	cases := []struct {
		name       string
		overlays   map[string]string
		goldenFile string
	}{
		{
			name: "path-validation failure",
			overlays: configtest.Overlay(
				"CARTULARY__ROOTS__DATABASE_STORAGE__PATH", "relative/postgres",
			),
			goldenFile: "startup_path_not_absolute_database_storage_root.json",
		},
		{
			name: "missing required runtime root",
			overlays: configtest.Overlay(
				"CARTULARY__ROOTS__EXPORT_OUTPUTS__BINDING_KIND", "",
				"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH", "",
			),
			goldenFile: "startup_missing_export_outputs_root.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			counters, _ := installStartupCounters()
			err := RuntimeConfigError(t, tc.overlays)
			configtest.RequireDiagnosticsMatchGolden(t, err, []string{"bootstrap", "diagnostics", tc.goldenFile})
			counters.RequireNotStarted(t)
		})
	}

}

func TestSingleActiveProcessAndRecoveryServingFencing_Integration(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "single-active-process")
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("open single-active postgres pool: %v", err)
	}
	defer pool.Close()

	s3Harness := s3test.Start(t)
	bucket := s3Harness.BootstrapBucketT(t, "single-active-process")
	s3Env := s3Harness.Env(bucket)
	store, err := objectstore.Setup(ctx, objectstore.Settings{
		BindingKind: "managed_service",
		Endpoint:    s3Env[objectstore.EndpointEnv],
		AccessKey:   s3Env[objectstore.AccessKeyEnv],
		SecretKey:   s3Env[objectstore.SecretKeyEnv],
		Secure:      s3Env[objectstore.SecureEnv] == "true",
		Bucket:      s3Env[objectstore.BucketEnv],
	}, objectstore.Instrumentation{})
	if err != nil {
		t.Fatalf("open single-active object store: %v", err)
	}
	defer store.Close()

	cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, configtest.Overlay(
		"CARTULARY__TIMEOUTS__EXTENSIONS__PROCESS_LEASE_ACQUIRE_SECONDS", "1",
		"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH", fixtures.Path("bootstrap-admin", "canonical.json"),
	)), testDB.Env())

	restoreAdmission, err := processlease.Acquire(
		ctx,
		processlease.PostgresBackend{
			Pool:        pool,
			AdvisoryKey: processlease.ServingAdvisoryKey,
			Purpose:     "restore target",
			Mode:        processlease.LockExclusive,
		},
		100*time.Millisecond,
		40*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("acquire restore admission fixture: %v", err)
	}
	defer restoreAdmission.Close()
	if blockedRuntime, err := NewRuntime(ctx, cfg, Options{Postgres: pool, ObjectStore: store}); err == nil {
		blockedRuntime.Close()
		t.Fatal("server runtime started while restore held the exclusive serving lease")
	} else if !errors.Is(err, processlease.ErrRecoveryServingLeaseActive) {
		t.Fatalf("server startup during restore failed with unexpected error: %v", err)
	}
	if err := restoreAdmission.Release(context.Background()); err != nil {
		t.Fatalf("release restore admission fixture: %v", err)
	}

	first, err := NewRuntime(ctx, cfg, Options{Postgres: pool, ObjectStore: store})
	if err != nil {
		t.Fatalf("start first single-active runtime: %v", err)
	}
	if first.processLease == nil || first.processLease.State() != processlease.StateHeld {
		first.Close()
		t.Fatal("first runtime did not hold the application-process lease")
	}
	if first.servingLease == nil || first.servingLease.State() != processlease.StateHeld {
		first.Close()
		t.Fatal("first runtime did not hold the application Recovery-serving lease")
	}
	if first.stagedJanitor == nil {
		first.Close()
		t.Fatal("single-active runtime did not compose the lifecycle-owned staged-object janitor")
	}

	counters, dependencies := installStartupCounters()
	if second, secondErr := newRuntimeWithTestDependencies(ctx, cfg, Options{Postgres: pool, ObjectStore: store}, dependencies); secondErr == nil {
		second.Close()
		first.Close()
		t.Fatal("overlapping application runtime acquired the deployment-global lease")
	} else if !errors.Is(secondErr, processlease.ErrApplicationProcessActive) {
		first.Close()
		t.Fatalf("overlapping application runtime failed with unexpected error: %v", secondErr)
	}
	counters.RequireNotStarted(t)

	if _, err := processlease.Acquire(
		ctx,
		processlease.PostgresBackend{
			Pool:        pool,
			AdvisoryKey: processlease.ServingAdvisoryKey,
			Purpose:     "restore target",
			Mode:        processlease.LockExclusive,
		},
		20*time.Millisecond,
		40*time.Millisecond,
	); !errors.Is(err, processlease.ErrApplicationProcessActive) {
		first.Close()
		t.Fatalf("active application serving lease did not block Recovery admission: %v", err)
	}
	if err := first.ActivatePublication(); err != nil {
		first.Close()
		t.Fatalf("activate first single-active runtime: %v", err)
	}
	first.Close()

	later, err := NewRuntime(ctx, cfg, Options{Postgres: pool, ObjectStore: store})
	if err != nil {
		t.Fatalf("orderly release did not permit later acquisition: %v", err)
	}
	later.Close()
}

func TestAllOptionalProfilesUnclaimedPublishesQuiescentJobs_Integration(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "all-optional-profiles-unclaimed")
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("open all-unclaimed postgres pool: %v", err)
	}
	defer pool.Close()
	store, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("open all-unclaimed object store: %v", err)
	}
	defer store.Close()

	cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, configtest.Overlay(
		"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH", fixtures.Path("bootstrap-admin", "canonical.json"),
		"CARTULARY__ENTERPRISE_AUTHENTICATION__CLAIMED", "false",
		"CARTULARY__IMPORT__CLAIMED", "false",
		"CARTULARY__INCIDENT_PORTABILITY__CLAIMED", "false",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED", "false",
		"CARTULARY__REFERENCE_PACK__CLAIMED", "false",
		"CARTULARY__SNAPSHOT_REPORTING__CLAIMED", "false",
	)), testDB.Env())

	var jobTransactions *jobs.TransactionService
	networkFlowCleanupComposed := false
	runtime, err := NewRuntime(ctx, cfg, Options{
		Postgres:    pool,
		ObjectStore: store,
		ObserveJobs: func(_ *jobs.Manager, transactions *jobs.TransactionService, _ *jobs.Runner, _ *pgxpool.Pool) {
			jobTransactions = transactions
		},
		ObserveNetworkFlowCleanup: func(*networkflow.GraphResultCleanupDispatcher) {
			networkFlowCleanupComposed = true
		},
	})
	if err != nil {
		t.Fatalf("start all-unclaimed runtime: %v", err)
	}
	defer runtime.Close()
	if err := runtime.ActivatePublication(); err != nil {
		t.Fatalf("activate all-unclaimed publication: %v", err)
	}

	claims := runtime.publication.claims()
	if len(claims) != 6 {
		t.Fatalf("all-unclaimed claim count = %d: %#v", len(claims), claims)
	}
	for _, claim := range claims {
		if claim.Claimed {
			t.Fatalf("all-unclaimed runtime admitted profile %q", claim.ProfileID)
		}
	}
	components := runtime.publication.expectedComponents()
	for _, componentID := range []string{"http", "job_dequeue", "websocket"} {
		if components[componentID] == "" {
			t.Fatalf("all-unclaimed publication omitted %q", componentID)
		}
	}
	for componentID := range components {
		if strings.HasPrefix(componentID, "worker:") {
			t.Fatalf("all-unclaimed publication admitted worker %q", componentID)
		}
	}
	if jobTransactions == nil {
		t.Fatal("all-unclaimed runtime did not expose Jobs composition")
	}
	if networkFlowCleanupComposed || runtime.networkFlowCleanupDispatcher != nil {
		t.Fatal("unclaimed Network Flow composed a graph-result cleanup dispatcher")
	}
	if _, err := jobTransactions.CreateQueuedTx(ctx, nil, jobs.EnqueueParams{
		JobKind: "import.discovery_v1",
	}, time.Now().UTC()); !errors.Is(err, jobs.ErrInvalidJobDefinition) {
		t.Fatalf("all-unclaimed Jobs admission error = %v", err)
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
		cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, configtest.Overlay(
			"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH", fixtures.Path("bootstrap-admin", "canonical.json"),
		)), env)

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
		if err := runtime.ActivatePublication(); err != nil {
			t.Fatalf("activate publication after bootstrap commit: %v", err)
		}

		server := httptest.NewServer(runtime.HTTPHandler())
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
		cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, configtest.Overlay(
			"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH", fixtures.Path("bootstrap-admin", "canonical.json"),
		)), env)

		counters, dependencies := installStartupCounters()
		_, err := newRuntimeWithTestDependencies(context.Background(), cfg, Options{Env: env}, dependencies)
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
			name: "symlinked bootstrap path",
			manifestPath: func(t *testing.T) string {
				target := bootstraptest.CanonicalBootstrapManifestPath()
				link := filepath.Join(t.TempDir(), "bootstrap-link.json")
				if err := os.Symlink(target, link); err != nil {
					t.Fatalf("create bootstrap manifest symlink: %v", err)
				}
				return link
			},
			goldenFile: "bootstrap_manifest_not_regular_file.json",
		},
		{
			name: "oversized bootstrap manifest",
			manifestPath: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "oversized-bootstrap.json")
				if err := os.WriteFile(path, make([]byte, bootstrap.ManifestMaximumBytes+1), 0o600); err != nil {
					t.Fatalf("write oversized bootstrap manifest: %v", err)
				}
				return path
			},
			goldenFile: "bootstrap_manifest_too_large.json",
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
			overlays := map[string]string{}
			if tc.manifestPath != nil {
				overlays["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = tc.manifestPath(t)
			}
			cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, overlays), env)

			counters, dependencies := installStartupCounters()
			_, err := newRuntimeWithTestDependencies(context.Background(), cfg, Options{Env: env}, dependencies)
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
				cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, configtest.Overlay(
					"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH", tc.manifestPath,
				)), env)

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
		cfg := BindPostgres(t, RuntimeConfigWithOverlays(t, configtest.Overlay(
			"CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH", fixtures.Path("bootstrap-admin", "canonical.json"),
		)), env)

		counters, dependencies := installStartupCounters()
		_, err := newRuntimeWithTestDependencies(context.Background(), cfg, Options{Env: env}, dependencies)
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

func installStartupCounters() (*StartupCounters, runtimeDependencies) {
	counters := &StartupCounters{}
	dependencies := productionRuntimeDependencies()
	originalJobsManager := dependencies.newJobsManager
	originalHTTPHandler := dependencies.newHTTPHandler

	dependencies.newJobsManager = func(options jobs.ManagerOptions) (*jobs.Manager, error) {
		counters.jobsManager++
		return originalJobsManager(options)
	}
	dependencies.newCollaborationRuntime = func(options collaboration.Options) (*collaboration.Runtime, error) {
		counters.wsHub++
		return collaboration.NewRuntime(options)
	}
	dependencies.newHTTPHandler = func(options ...httpapi.Options) (http.Handler, error) {
		counters.httpHandler++
		return originalHTTPHandler(options...)
	}

	return counters, dependencies
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
	configtest.EnsureRevisionsConflictTokenTestEnvironment(env)
	env["CARTULARY_SECRET_RUNTIME_TEST_REVISIONS_CONFLICT"] = "pVldGSpD5oEmYa9F85d3_iL2lzBgkyfiWcoJDhsSGpk"
	return env
}

func BindPostgres(t testing.TB, loaded configassembly.Loaded, env map[string]string) configassembly.Loaded {
	t.Helper()
	configtest.BindPostgresEnvToDatabaseRoot(t, loaded.Deployment().Roots.DatabaseStorage.Path, env, postgres.PurposeRuntime)
	return loaded
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
