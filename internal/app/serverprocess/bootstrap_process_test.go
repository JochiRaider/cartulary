package serverprocess

import (
	"bytes"
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
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/securityassert"
)

func TestReadyState_Process(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-e-0-01")

	db := openSQL(t, testDB.DSN)
	defer closeSQL(t, db)

	bucket := bucketName("bootstrap-e-0-01")
	defer cleanupBucket(t, s3Harness, bucket)

	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json")})

	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	defer server.Stop(t)

	server.WaitForReady(t)
	server.RequireStatus(t, "/healthz", http.StatusOK)
	server.RequireStatus(t, "/readyz", http.StatusOK)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)

	payload := []byte("bootstrap ready state proof")
	store, err := objectstore.NewFilesystemStore(env["CARTULARY__ROOTS__OBJECT_STORAGE__PATH"])
	if err != nil {
		t.Fatalf("open configured filesystem object store: %v", err)
	}
	if err := store.PutObject(t.Context(), "bootstrap-ready.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("write configured filesystem object store: %v", err)
	}
	object, _, err := store.ReadObject(t.Context(), "bootstrap-ready.txt", objectstore.ReadOptions{})
	if err != nil {
		t.Fatalf("read configured filesystem object store: %v", err)
	}
	defer func() {
		if err := object.Close(); err != nil {
			t.Errorf("close configured object-store payload: %v", err)
		}
	}()
	got, err := io.ReadAll(object)
	if err != nil {
		t.Fatalf("read configured object-store payload: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("unexpected object-store payload after ready state: got %q want %q", got, payload)
	}
}

func TestInvalidConfigDiagnostics_Process(t *testing.T) {
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
		{
			name:       "missing Revisions conflict token key ring",
			configText: string(fixtures.MustRead("config", "valid.toml")),
			env: map[string]string{
				"CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH": "/does/not/exist/revisions-conflict-token-key-ring.json",
			},
			goldenFile: "startup_revisions_conflict_token_manifest_missing.json",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-e-0-02")

			bucket := bucketName("bootstrap-e-0-02")
			defer cleanupBucket(t, s3Harness, bucket)

			configPath := writeConfig(t, tc.configText)
			env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, Overrides: tc.env})

			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			err := server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected invalid config startup to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000")
			server.RequireDiagnosticsMatchGolden(t, []string{"bootstrap", "diagnostics", tc.goldenFile})
		})
	}
}

func TestFirstAdminBootstrap_Process(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-e-0-03")

	db := openSQL(t, testDB.DSN)
	defer closeSQL(t, db)

	bucket := bucketName("bootstrap-e-0-03")
	defer cleanupBucket(t, s3Harness, bucket)

	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json")})

	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	defer server.Stop(t)
	server.WaitForReady(t)

	requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)

	var userID string
	var email string
	if err := db.QueryRowContext(t.Context(), `SELECT id::text, email FROM users WHERE is_active = true AND is_deployment_admin = true`).Scan(&userID, &email); err != nil {
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

func TestBootstrapFailures_Process(t *testing.T) {
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
			testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-e-0-04")

			db := openSQL(t, testDB.DSN)
			defer closeSQL(t, db)
			if tc.seed != nil {
				tc.seed(t, db)
			}

			bucket := bucketName("bootstrap-e-0-04")
			defer cleanupBucket(t, s3Harness, bucket)

			configPath := writeConfig(t, tc.configContent())
			env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: tc.bootstrapPath})

			server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
			err := server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected bootstrap failure to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000")
			server.RequireDiagnosticsMatchGolden(t, []string{"bootstrap", "diagnostics", tc.goldenFile})
			requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, tc.wantUserCount)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
		})
	}
}

func TestBootstrapSkipAndRecovery_Process(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	t.Run("existing active deployment admin skips stale and invalid bootstrap manifests", func(t *testing.T) {
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-e-0-05-skip")

		db := openSQL(t, testDB.DSN)
		defer closeSQL(t, db)
		if _, err := db.ExecContext(t.Context(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, true, true)`, "existing-admin@example.test", "Existing Admin", "existing-hash"); err != nil {
			t.Fatalf("seed active deployment admin: %v", err)
		}

		bucket := bucketName("bootstrap-e-0-05-skip")
		defer cleanupBucket(t, s3Harness, bucket)

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
				configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
				env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: tc.manifestPath})

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
		testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "bootstrap-e-0-05-recovery")

		db := openSQL(t, testDB.DSN)
		defer closeSQL(t, db)

		var userID string
		if err := db.QueryRowContext(t.Context(), `INSERT INTO users (email, display_name, password_hash, is_active, is_deployment_admin) VALUES ($1, $2, $3, false, true) RETURNING id`, "retired-admin@example.test", "Retired Admin", "existing-hash").Scan(&userID); err != nil {
			t.Fatalf("seed retired deployment admin: %v", err)
		}
		if _, err := db.ExecContext(t.Context(), `INSERT INTO deployment_bootstrap_state (slot, bootstrap_schema_id, bootstrap_artifact_id, artifact_sha256, created_user_id) VALUES ('first_deployment_admin', $1, $2, $3, $4)`, "cartulary.bootstrap_admin.v1", "33333333-3333-3333-3333-333333333333", []byte{0x04, 0x05, 0x06}, userID); err != nil {
			t.Fatalf("seed bootstrap completion state: %v", err)
		}

		bucket := bucketName("bootstrap-e-0-05-recovery")
		defer cleanupBucket(t, s3Harness, bucket)

		configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
		env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3Harness.Env(bucket), ConfigPath: configPath, BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json")})

		server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
		err := server.WaitForExit(t)
		if err == nil {
			t.Fatal("expected lost-admin recovery startup to exit non-zero")
		}
		server.RequireConnectionRefused(t, "/healthz")
		server.RequireConnectionRefused(t, "/readyz")
		server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000")
		server.RequireDiagnosticsMatchGolden(t, []string{"bootstrap", "diagnostics", "bootstrap_recovery_not_supported.json"})
		requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
	})
}

type auditEvent struct {
	ActorUserID string
	EventSource string
	EventKind   string
	RequestID   string
	CreatedAt   time.Time
	After       map[string]any
}

func writeConfig(t testing.TB, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write bootstrap config: %v", err)
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

func bucketName(prefix string) string {
	value := strings.ToLower(prefix)
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return fmt.Sprintf("%s-%d", value, time.Now().UnixNano())
}

func cleanupBucket(t testing.TB, harness *s3test.Harness, bucket string) {
	t.Helper()
	ctx, cancel := newProcessCleanupContext()
	defer cancel()
	reportProcessCleanupFailure(t, "cleanup bucket "+bucket, harness.CleanupBucket(ctx, bucket))
}

func openSQL(t testing.TB, dsn string) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	return db
}

func closeSQL(t testing.TB, db *sql.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("close postgres sql handle: %v", err)
	}
}

func requireCountSQL(t testing.TB, db *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(t.Context(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}

func lookupBootstrapAuditEvent(t testing.TB, db *sql.DB) auditEvent {
	t.Helper()

	var actorUserID string
	var eventSource string
	var eventKind string
	var requestID string
	var createdAt time.Time
	var afterJSON []byte
	if err := db.QueryRowContext(t.Context(), `
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

	event := auditEvent{
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
