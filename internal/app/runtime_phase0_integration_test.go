package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestPhase0_InvalidConfigNeverReachesReady_I_0_03(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-invalid-config")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	}()

	s3Harness := s3test.Start(t)
	bucket, err := s3Harness.BootstrapBucket(context.Background(), "phase0-invalid-config")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	t.Run("rejects invalid filesystem roots even when services are healthy", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		cfg.Roots.DatabaseStorage.Path = "relative/postgres"

		_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		requireInvalidDeploymentConfig(t, err)
	})

	t.Run("rejects missing required runtime roots even when services are healthy", func(t *testing.T) {
		cfg := phase0RuntimeConfig(t)
		cfg.Roots.ExportOutputs = config.RootBinding{}

		_, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		requireInvalidDeploymentConfig(t, err)
	})
}

func TestPhase0_FirstAdminBootstrap_I_0_04(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-bootstrap-success")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	}()

	bucket := phase0BucketName("phase0-bootstrap-success")
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	}()

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}

	cfg := phase0RuntimeConfig(t)
	cfg.Bootstrap.FirstAdminManifestPath = fixtures.Path("bootstrap-admin", "canonical.json")

	runtime, err := NewRuntime(context.Background(), cfg, Options{Env: env})
	if err != nil {
		t.Fatalf("start runtime with canonical bootstrap manifest: %v", err)
	}
	defer runtime.Close()

	requireCountPool(t, runtime.Postgres, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireCountPool(t, runtime.Postgres, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireCountPool(t, runtime.Postgres, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)

	var email string
	var mfaRequired bool
	var passwordHash string
	if err := runtime.Postgres.QueryRow(context.Background(), `SELECT email, mfa_required, password_hash FROM users WHERE is_active = true AND is_deployment_admin = true`).Scan(&email, &mfaRequired, &passwordHash); err != nil {
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
}

func TestPhase0_BootstrapFailures_I_0_05(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	cases := []struct {
		name              string
		manifestPath      func(t *testing.T) string
		seed              func(t *testing.T, db *sql.DB)
		wantReasonCode    string
		wantUserCount     int
		wantBootstrapRows int
		wantAuditRows     int
	}{
		{
			name:           "missing configured bootstrap path",
			wantReasonCode: "bootstrap_manifest_path_missing",
		},
		{
			name:           "unreadable bootstrap file",
			manifestPath:   func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing-bootstrap.json") },
			wantReasonCode: "bootstrap_manifest_not_readable",
		},
		{
			name: "non regular bootstrap path",
			manifestPath: func(t *testing.T) string {
				return t.TempDir()
			},
			wantReasonCode: "bootstrap_manifest_not_regular_file",
		},
		{
			name: "malformed json manifest",
			manifestPath: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "bootstrap-admin.json")
				if err := os.WriteFile(path, []byte(`{"bootstrap_schema_id":`), 0o644); err != nil {
					t.Fatalf("write malformed manifest: %v", err)
				}
				return path
			},
			wantReasonCode: "bootstrap_manifest_parse_error",
		},
		{
			name: "schema invalid manifest",
			manifestPath: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "bootstrap-admin.json")
				content := `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","mfa_required":false}`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatalf("write invalid manifest: %v", err)
				}
				return path
			},
			wantReasonCode: "bootstrap_manifest_schema_invalid",
		},
		{
			name: "email conflict",
			manifestPath: func(t *testing.T) string {
				return fixtures.Path("bootstrap-admin", "canonical.json")
			},
			seed: func(t *testing.T, db *sql.DB) {
				if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash) VALUES ($1, $2, $3)`, "bootstrap-admin@example.test", "Existing User", "existing-hash"); err != nil {
					t.Fatalf("seed conflicting user: %v", err)
				}
			},
			wantReasonCode:    "bootstrap_email_conflict",
			wantUserCount:     1,
			wantBootstrapRows: 0,
			wantAuditRows:     0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-bootstrap-failure")
			if err != nil {
				t.Fatalf("prepare postgres database: %v", err)
			}
			defer func() {
				if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
					t.Fatalf("drop postgres database: %v", err)
				}
			}()

			db := openPhase0SQL(t, testDB.DSN)
			defer db.Close()
			if tc.seed != nil {
				tc.seed(t, db)
			}

			bucket := phase0BucketName(tc.name)
			defer func() {
				if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
					t.Logf("cleanup bucket: %v", err)
				}
			}()

			env := testDB.Env()
			for key, value := range s3Harness.Env(bucket) {
				env[key] = value
			}

			cfg := phase0RuntimeConfig(t)
			if tc.manifestPath != nil {
				cfg.Bootstrap.FirstAdminManifestPath = tc.manifestPath(t)
			}

			_, err = NewRuntime(context.Background(), cfg, Options{Env: env})
			requireBootstrapReason(t, err, tc.wantReasonCode)

			requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, tc.wantUserCount)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, tc.wantBootstrapRows)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, tc.wantAuditRows)
		})
	}
}

func TestPhase0_BootstrapSkipAndRecovery_I_0_06(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("existing active deployment admin skips manifest consumption", func(t *testing.T) {
		testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-bootstrap-skip")
		if err != nil {
			t.Fatalf("prepare postgres database: %v", err)
		}
		defer func() {
			if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
				t.Fatalf("drop postgres database: %v", err)
			}
		}()

		db := openPhase0SQL(t, testDB.DSN)
		defer db.Close()
		if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, true, true)`, "existing-admin@example.test", "Existing Admin", "existing-hash"); err != nil {
			t.Fatalf("seed active deployment admin: %v", err)
		}

		bucket := phase0BucketName("phase0-bootstrap-skip")
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := testDB.Env()
		for key, value := range s3Harness.Env(bucket) {
			env[key] = value
		}

		cfg := phase0RuntimeConfig(t)
		cfg.Bootstrap.FirstAdminManifestPath = filepath.Join(t.TempDir(), "missing-bootstrap.json")

		runtime, err := NewRuntime(context.Background(), cfg, Options{Env: env})
		if err != nil {
			t.Fatalf("start runtime with existing deployment admin: %v", err)
		}
		defer runtime.Close()

		requireCountPool(t, runtime.Postgres, `SELECT COUNT(*) FROM users`, 1)
		requireCountPool(t, runtime.Postgres, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
		requireCountPool(t, runtime.Postgres, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
	})

	t.Run("bootstrap recovery remains fail-closed when completion state exists without an active admin", func(t *testing.T) {
		testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-bootstrap-recovery")
		if err != nil {
			t.Fatalf("prepare postgres database: %v", err)
		}
		defer func() {
			if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
				t.Fatalf("drop postgres database: %v", err)
			}
		}()

		db := openPhase0SQL(t, testDB.DSN)
		defer db.Close()

		var userID string
		if err := db.QueryRowContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, false, true) RETURNING id`, "retired-admin@example.test", "Retired Admin", "existing-hash").Scan(&userID); err != nil {
			t.Fatalf("seed retired deployment admin: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), `INSERT INTO deployment_bootstrap_state (slot, bootstrap_schema_id, bootstrap_artifact_id, artifact_sha256, created_user_id) VALUES ('first_deployment_admin', $1, $2, $3, $4)`, bootstrapManifestSchemaID, "22222222-2222-2222-2222-222222222222", []byte{0x01, 0x02, 0x03}, userID); err != nil {
			t.Fatalf("seed bootstrap completion marker: %v", err)
		}

		bucket := phase0BucketName("phase0-bootstrap-recovery")
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := testDB.Env()
		for key, value := range s3Harness.Env(bucket) {
			env[key] = value
		}

		cfg := phase0RuntimeConfig(t)
		cfg.Bootstrap.FirstAdminManifestPath = fixtures.Path("bootstrap-admin", "canonical.json")

		_, err = NewRuntime(context.Background(), cfg, Options{Env: env})
		requireBootstrapReason(t, err, "bootstrap_recovery_not_supported")

		requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 0)
	})
}

func requireInvalidDeploymentConfig(t testing.TB, err error) {
	t.Helper()

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
	if !ok {
		t.Fatalf("expected diagnostics error, got %T", err)
	}
	if diagnosticsErr.Code != config.InvalidDeploymentConfigCode {
		t.Fatalf("unexpected diagnostics code: got %q want %q", diagnosticsErr.Code, config.InvalidDeploymentConfigCode)
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

func openPhase0SQL(t testing.TB, dsn string) *sql.DB {
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

func requireCountPool(t testing.TB, pool *pgxpool.Pool, query string, want int) {
	t.Helper()

	var got int
	if err := pool.QueryRow(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}

func phase0BucketName(prefix string) string {
	value := strings.ToLower(prefix)
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return fmt.Sprintf("%s-%d", value, time.Now().UnixNano())
}
