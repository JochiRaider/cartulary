package serverprocess

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/bootstraptest"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/securityassert"
)

func TestPhase0_ReadyState_E_0_01(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "phase0-e-0-01")

	db := openPhase0SQL(t, testDB.DSN)
	defer db.Close()

	bucket := phase0BucketName("phase0-e-0-01")
	defer cleanupPhase0Bucket(t, s3Harness, bucket)

	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))

	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	defer server.Stop(t)

	server.WaitForReady(t)
	server.RequireStatus(t, "/healthz", http.StatusOK)
	server.RequireStatus(t, "/readyz", http.StatusOK)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)

	payload := []byte("phase0 ready state proof")
	store, err := objectstore.NewFilesystemStore(env["CARTULARY__ROOTS__OBJECT_STORAGE__PATH"])
	if err != nil {
		t.Fatalf("open configured filesystem object store: %v", err)
	}
	if err := store.PutObject(context.Background(), "phase0-ready.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("write configured filesystem object store: %v", err)
	}
	object, _, err := store.ReadObject(context.Background(), "phase0-ready.txt", objectstore.ReadOptions{})
	if err != nil {
		t.Fatalf("read configured filesystem object store: %v", err)
	}
	defer object.Close()
	got, err := io.ReadAll(object)
	if err != nil {
		t.Fatalf("read configured object-store payload: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("unexpected object-store payload after ready state: got %q want %q", got, payload)
	}
}

func TestPhase0_InvalidConfigDiagnostics_E_0_02(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	cases := []struct {
		name       string
		configText string
		env        map[string]string
		goldenFile string
	}{
		{
			name:       "missing required runtime root",
			configText: stripConfigSection(t, string(fixtures.MustRead("config", "valid.toml")), "[roots.export_outputs]"),
			env: map[string]string{
				"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH": "",
			},
			goldenFile: "startup_missing_export_outputs_root.json",
		},
		{
			name:       "invalid root path shape",
			configText: string(fixtures.MustRead("config", "valid.toml")),
			env: map[string]string{
				"CARTULARY__ROOTS__DATABASE_STORAGE__PATH": "relative/postgres",
			},
			goldenFile: "startup_path_not_absolute_database_storage_root.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "phase0-e-0-02")

			bucket := phase0BucketName("phase0-e-0-02")
			defer cleanupPhase0Bucket(t, s3Harness, bucket)

			configPath := writePhase0Config(t, tc.configText)
			env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, "")
			for key, value := range tc.env {
				env[key] = value
			}

			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			err := server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected invalid config startup to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000/views/cartulary.view.timeline.v2/changes")
			server.RequireDiagnosticsMatchGolden(t, []string{"phase0", "diagnostics", tc.goldenFile})
		})
	}
}

func TestPhase0_FirstAdminBootstrap_E_0_03(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "phase0-e-0-03")

	db := openPhase0SQL(t, testDB.DSN)
	defer db.Close()

	bucket := phase0BucketName("phase0-e-0-03")
	defer cleanupPhase0Bucket(t, s3Harness, bucket)

	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))

	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	defer server.Stop(t)
	server.WaitForReady(t)

	requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)

	var userID string
	var email string
	if err := db.QueryRowContext(context.Background(), `SELECT id::text, email FROM users WHERE is_active = true AND is_deployment_admin = true`).Scan(&userID, &email); err != nil {
		t.Fatalf("query bootstrap-created user: %v", err)
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
	securityassert.RequireSecretSafePayload(t, audit.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32", "provider_subject", "provider_key"})
}

func TestPhase0_BootstrapFailures_E_0_04(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	cases := []struct {
		name          string
		configContent func() string
		bootstrapPath string
		seed          func(t *testing.T, db *sql.DB)
		goldenFile    string
		wantUserCount int
	}{
		{
			name: "missing bootstrap path",
			configContent: func() string {
				return stripConfigSection(t, string(fixtures.MustRead("config", "valid.toml")), "[bootstrap]")
			},
			goldenFile: "bootstrap_manifest_path_missing.json",
		},
		{
			name: "non regular bootstrap path",
			configContent: func() string {
				return string(fixtures.MustRead("config", "valid.toml"))
			},
			bootstrapPath: bootstraptest.WriteNonRegularBootstrapManifestPath(t),
			goldenFile:    "bootstrap_manifest_not_regular_file.json",
		},
		{
			name: "malformed bootstrap manifest",
			configContent: func() string {
				return string(fixtures.MustRead("config", "valid.toml"))
			},
			bootstrapPath: bootstraptest.WriteMalformedBootstrapManifest(t),
			goldenFile:    "bootstrap_manifest_parse_error.json",
		},
		{
			name: "explicit false mfa bootstrap manifest",
			configContent: func() string {
				return string(fixtures.MustRead("config", "valid.toml"))
			},
			bootstrapPath: bootstraptest.WriteExplicitFalseMFABootstrapManifest(t),
			goldenFile:    "bootstrap_manifest_schema_invalid_explicit_false_mfa.json",
		},
		{
			name: "unknown-member bootstrap manifest",
			configContent: func() string {
				return string(fixtures.MustRead("config", "valid.toml"))
			},
			bootstrapPath: bootstraptest.WriteUnknownMemberBootstrapManifest(t),
			goldenFile:    "bootstrap_manifest_schema_invalid_unknown_member.json",
		},
		{
			name: "email conflict",
			configContent: func() string {
				return string(fixtures.MustRead("config", "valid.toml"))
			},
			bootstrapPath: bootstraptest.CanonicalBootstrapManifestPath(),
			seed: func(t *testing.T, db *sql.DB) {
				bootstraptest.SeedBootstrapEmailConflict(t, db)
			},
			goldenFile:    "bootstrap_email_conflict.json",
			wantUserCount: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "phase0-e-0-04")

			db := openPhase0SQL(t, testDB.DSN)
			defer db.Close()
			if tc.seed != nil {
				tc.seed(t, db)
			}

			bucket := phase0BucketName("phase0-e-0-04")
			defer cleanupPhase0Bucket(t, s3Harness, bucket)

			configPath := writePhase0Config(t, tc.configContent())
			env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, tc.bootstrapPath)

			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			err := server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected bootstrap failure to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000/views/cartulary.view.timeline.v2/changes")
			server.RequireDiagnosticsMatchGolden(t, []string{"phase0", "diagnostics", tc.goldenFile})
			requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, tc.wantUserCount)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
		})
	}
}

func TestPhase0_BootstrapSkipAndRecovery_E_0_05(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	t.Run("existing active deployment admin skips stale and invalid bootstrap manifests", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "phase0-e-0-05-skip")

		db := openPhase0SQL(t, testDB.DSN)
		defer db.Close()
		if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, true, true)`, "existing-admin@example.test", "Existing Admin", "existing-hash"); err != nil {
			t.Fatalf("seed active deployment admin: %v", err)
		}

		bucket := phase0BucketName("phase0-e-0-05-skip")
		defer cleanupPhase0Bucket(t, s3Harness, bucket)

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
				configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
				env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, tc.manifestPath)

				server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
				defer server.Stop(t)
				server.WaitForReady(t)
				server.RequireStatus(t, "/healthz", http.StatusOK)
				server.RequireStatus(t, "/readyz", http.StatusOK)
				requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, 1)
				requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
				requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
				requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
			})
		}
	})

	t.Run("bootstrap recovery remains fail-closed", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "phase0-e-0-05-recovery")

		db := openPhase0SQL(t, testDB.DSN)
		defer db.Close()

		var userID string
		if err := db.QueryRowContext(context.Background(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, false, true) RETURNING id`, "retired-admin@example.test", "Retired Admin", "existing-hash").Scan(&userID); err != nil {
			t.Fatalf("seed retired deployment admin: %v", err)
		}
		if _, err := db.ExecContext(context.Background(), `INSERT INTO deployment_bootstrap_state (slot, bootstrap_schema_id, bootstrap_artifact_id, artifact_sha256, created_user_id) VALUES ('first_deployment_admin', $1, $2, $3, $4)`, "cartulary.bootstrap_admin.v1", "33333333-3333-3333-3333-333333333333", []byte{0x04, 0x05, 0x06}, userID); err != nil {
			t.Fatalf("seed bootstrap completion state: %v", err)
		}

		bucket := phase0BucketName("phase0-e-0-05-recovery")
		defer cleanupPhase0Bucket(t, s3Harness, bucket)

		configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
		env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))

		server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
		err := server.WaitForExit(t)
		if err == nil {
			t.Fatal("expected lost-admin recovery startup to exit non-zero")
		}
		server.RequireConnectionRefused(t, "/healthz")
		server.RequireConnectionRefused(t, "/readyz")
		server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000/views/cartulary.view.timeline.v2/changes")
		server.RequireDiagnosticsMatchGolden(t, []string{"phase0", "diagnostics", "bootstrap_recovery_not_supported.json"})
		requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
	})
}

type phase0AuditEvent struct {
	ActorUserID string
	EventSource string
	EventKind   string
	RequestID   string
	CreatedAt   time.Time
	After       map[string]any
}

func phase0ServerEnv(t testing.TB, databaseEnv map[string]string, objectStoreEnv map[string]string, configPath string, bootstrapPath string) map[string]string {
	t.Helper()

	tempRoots := configtest.SetupTempRoots(t)
	env := make(map[string]string)
	for key, value := range databaseEnv {
		env[key] = value
	}
	for key, value := range objectStoreEnv {
		env[key] = value
	}
	for key, value := range tempRoots.Paths {
		env[key] = value
	}
	configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], env)
	env["CARTULARY_CONFIG_FILE"] = configPath
	if bootstrapPath != "" {
		env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = bootstrapPath
	}
	return env
}

func writePhase0Config(t testing.TB, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write phase 0 config: %v", err)
	}
	return path
}

func stripConfigSection(t testing.TB, content string, header string) string {
	t.Helper()

	lines := strings.Split(content, "\n")
	start := -1
	end := len(lines)
	for i, line := range lines {
		if strings.TrimSpace(line) == header {
			start = i
			continue
		}
		if start >= 0 && strings.HasPrefix(strings.TrimSpace(line), "[") {
			end = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("section %q not found", header)
	}
	return strings.Join(append(lines[:start], lines[end:]...), "\n")
}

func phase0BucketName(prefix string) string {
	value := strings.ToLower(prefix)
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return fmt.Sprintf("%s-%d", value, time.Now().UnixNano())
}

func cleanupPhase0Bucket(t testing.TB, harness *s3test.Harness, bucket string) {
	t.Helper()
	if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
		t.Logf("cleanup bucket: %v", err)
	}
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

func lookupBootstrapAuditEvent(t testing.TB, db *sql.DB) phase0AuditEvent {
	t.Helper()

	var actorUserID string
	var eventSource string
	var eventKind string
	var requestID string
	var createdAt time.Time
	var afterJSON []byte
	if err := db.QueryRowContext(context.Background(), `
SELECT COALESCE(actor_user_id::text, ''),
       event_source,
       event_kind,
       COALESCE(request_id, ''),
       created_at,
       after_json
  FROM deployment_admin_audit_events
 ORDER BY created_at ASC
 LIMIT 1
`).Scan(&actorUserID, &eventSource, &eventKind, &requestID, &createdAt, &afterJSON); err != nil {
		t.Fatalf("query bootstrap audit event: %v", err)
	}

	event := phase0AuditEvent{
		ActorUserID: actorUserID,
		EventSource: eventSource,
		EventKind:   eventKind,
		RequestID:   requestID,
		CreatedAt:   createdAt,
		After:       map[string]any{},
	}
	if len(afterJSON) > 0 {
		if err := json.Unmarshal(afterJSON, &event.After); err != nil {
			t.Fatalf("decode bootstrap audit after_json: %v", err)
		}
	}
	return event
}
