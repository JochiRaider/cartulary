package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/crosscutting"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/wstest"
)

func TestPhase0_ReadyState_E_0_01(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-e-0-01")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer dropPhase0Database(t, postgresHarness, testDB.Name)

	db := openPhase0SQL(t, testDB.DSN)
	defer db.Close()

	bucket := phase0BucketName("phase0-e-0-01")
	defer cleanupPhase0Bucket(t, s3Harness, bucket)

	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))

	server := startPhase0Server(t, env)
	defer server.Stop(t)

	server.WaitForReady(t)
	server.RequireStatus(t, "/healthz", http.StatusOK)
	server.RequireStatus(t, "/readyz", http.StatusOK)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)

	payload := []byte("phase0 ready state proof")
	got, err := s3Harness.RoundTrip(context.Background(), bucket, "phase0-ready.txt", payload)
	if err != nil {
		t.Fatalf("round trip against ready deployment object store: %v", err)
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
		path       string
		reasonCode string
	}{
		{
			name:       "missing required runtime roots",
			configText: string(fixtures.MustRead("config", "invalid_missing_required.toml")),
			path:       "deployment_profile",
			reasonCode: "missing_required_key",
		},
		{
			name:       "invalid root path shape",
			configText: string(fixtures.MustRead("config", "valid.toml")),
			env: map[string]string{
				"CARTULARY__ROOTS__DATABASE_STORAGE__PATH": "relative/postgres",
			},
			path:       "roots.database_storage.path",
			reasonCode: "path_not_absolute",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-e-0-02")
			if err != nil {
				t.Fatalf("prepare postgres database: %v", err)
			}
			defer dropPhase0Database(t, postgresHarness, testDB.Name)

			bucket := phase0BucketName("phase0-e-0-02")
			defer cleanupPhase0Bucket(t, s3Harness, bucket)

			configPath := writePhase0Config(t, tc.configText)
			env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, "")
			for key, value := range tc.env {
				env[key] = value
			}

			server := startPhase0Server(t, env)
			err = server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected invalid config startup to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000/views/cartulary.view.timeline.v1/changes")
			server.RequireDiagnosticsCode(t, "invalid_deployment_config")
			server.RequireDiagnosticsField(t, tc.path, tc.reasonCode)
		})
	}
}

func TestPhase0_FirstAdminBootstrap_E_0_03(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-e-0-03")
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	defer dropPhase0Database(t, postgresHarness, testDB.Name)

	db := openPhase0SQL(t, testDB.DSN)
	defer db.Close()

	bucket := phase0BucketName("phase0-e-0-03")
	defer cleanupPhase0Bucket(t, s3Harness, bucket)

	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))

	server := startPhase0Server(t, env)
	defer server.Stop(t)
	server.WaitForReady(t)

	requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 1)
	requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)

	audit := lookupBootstrapAuditEvent(t, db)
	crosscutting.RequireSystemMutationAttribution(t, crosscutting.SystemMutationAttribution{
		ActorUserID: audit.ActorUserID,
		Source:      audit.EventSource,
		EventKind:   audit.EventKind,
		RequestID:   audit.RequestID,
		CreatedAt:   audit.CreatedAt,
	}, "bootstrap_manifest", "bootstrap_admin_created")
	crosscutting.RequireSecretSafePayload(t, audit.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32", "provider_subject", "provider_key"})
}

func TestPhase0_BootstrapFailures_E_0_04(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	cases := []struct {
		name           string
		configContent  func() string
		bootstrapPath  string
		wantReasonCode string
	}{
		{
			name: "missing bootstrap path",
			configContent: func() string {
				return stripConfigSection(t, string(fixtures.MustRead("config", "valid.toml")), "[bootstrap]")
			},
			wantReasonCode: "bootstrap_manifest_path_missing",
		},
		{
			name: "schema-invalid bootstrap manifest",
			configContent: func() string {
				return string(fixtures.MustRead("config", "valid.toml"))
			},
			bootstrapPath: func() string {
				path := filepath.Join(t.TempDir(), "bootstrap-admin.json")
				content := `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","mfa_required":false}`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatalf("write invalid bootstrap manifest: %v", err)
				}
				return path
			}(),
			wantReasonCode: "bootstrap_manifest_schema_invalid",
		},
		{
			name: "unknown-member bootstrap manifest",
			configContent: func() string {
				return string(fixtures.MustRead("config", "valid.toml"))
			},
			bootstrapPath: func() string {
				path := filepath.Join(t.TempDir(), "bootstrap-admin.json")
				content := `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","unexpected":"surprise"}`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatalf("write invalid bootstrap manifest: %v", err)
				}
				return path
			}(),
			wantReasonCode: "bootstrap_manifest_schema_invalid",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-e-0-04")
			if err != nil {
				t.Fatalf("prepare postgres database: %v", err)
			}
			defer dropPhase0Database(t, postgresHarness, testDB.Name)

			db := openPhase0SQL(t, testDB.DSN)
			defer db.Close()

			bucket := phase0BucketName("phase0-e-0-04")
			defer cleanupPhase0Bucket(t, s3Harness, bucket)

			configPath := writePhase0Config(t, tc.configContent())
			env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, tc.bootstrapPath)

			server := startPhase0Server(t, env)
			err = server.WaitForExit(t)
			if err == nil {
				t.Fatal("expected bootstrap failure to exit non-zero")
			}
			server.RequireConnectionRefused(t, "/healthz")
			server.RequireConnectionRefused(t, "/readyz")
			server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000/views/cartulary.view.timeline.v1/changes")
			server.RequireDiagnosticsCode(t, "invalid_deployment_config")
			server.RequireReasonCode(t, tc.wantReasonCode)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM users`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
			requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
		})
	}
}

func TestPhase0_BootstrapSkipAndRecovery_E_0_05(t *testing.T) {
	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	t.Run("existing active deployment admin skips stale and invalid bootstrap manifests", func(t *testing.T) {
		testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-e-0-05-skip")
		if err != nil {
			t.Fatalf("prepare postgres database: %v", err)
		}
		defer dropPhase0Database(t, postgresHarness, testDB.Name)

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
				name: "invalid manifest content",
				manifestPath: func() string {
					path := filepath.Join(t.TempDir(), "invalid-bootstrap.json")
					content := `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","mfa_required":false}`
					if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
						t.Fatalf("write invalid bootstrap manifest: %v", err)
					}
					return path
				}(),
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
				env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, tc.manifestPath)

				server := startPhase0Server(t, env)
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
		testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), "phase0-e-0-05-recovery")
		if err != nil {
			t.Fatalf("prepare postgres database: %v", err)
		}
		defer dropPhase0Database(t, postgresHarness, testDB.Name)

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

		server := startPhase0Server(t, env)
		err = server.WaitForExit(t)
		if err == nil {
			t.Fatal("expected lost-admin recovery startup to exit non-zero")
		}
		server.RequireConnectionRefused(t, "/readyz")
		server.RequireWebsocketConnectionRefused(t, "/ws/v1/incidents/00000000-0000-0000-0000-000000000000/views/cartulary.view.timeline.v1/changes")
		server.RequireReasonCode(t, "bootstrap_recovery_not_supported")
		requireCountSQL(t, db, `SELECT COUNT(*) FROM users WHERE is_active = true AND is_deployment_admin = true`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_bootstrap_state`, 1)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events`, 0)
		requireCountSQL(t, db, `SELECT COUNT(*) FROM incident_memberships`, 0)
	})
}

type phase0ServerProcess struct {
	address string
	cancel  context.CancelFunc
	done    chan error
	cmd     *exec.Cmd
	stderr  bytes.Buffer
	stdout  bytes.Buffer
}

type phase0AuditEvent struct {
	ActorUserID string
	EventSource string
	EventKind   string
	RequestID   string
	CreatedAt   time.Time
	After       map[string]any
}

func startPhase0Server(t testing.TB, env map[string]string) *phase0ServerProcess {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "go", "run", "./cmd/server")
	cmd.Dir = phase0RepoRoot()
	cmd.Env = append(os.Environ(), phase0EnvPairs(env)...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	process := &phase0ServerProcess{
		address: env["CARTULARY_HTTP_ADDR"],
		cancel:  cancel,
		done:    make(chan error, 1),
		cmd:     cmd,
	}
	cmd.Stdout = &process.stdout
	cmd.Stderr = &process.stderr

	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatalf("start cmd/server: %v", err)
	}

	go func() {
		process.done <- cmd.Wait()
	}()

	return process
}

func (p *phase0ServerProcess) WaitForReady(t testing.TB) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-p.done:
			t.Fatalf("cmd/server exited before readiness: %v\nstdout:\n%s\nstderr:\n%s", err, p.stdout.String(), p.stderr.String())
		default:
		}

		healthResp, healthErr := client.Get("http://" + p.address + "/healthz")
		if healthErr == nil {
			healthResp.Body.Close()
			readyResp, readyErr := client.Get("http://" + p.address + "/readyz")
			if readyErr == nil {
				readyResp.Body.Close()
				if healthResp.StatusCode == http.StatusOK && readyResp.StatusCode == http.StatusOK {
					return
				}
			}
		}

		time.Sleep(200 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for readiness\nstdout:\n%s\nstderr:\n%s", p.stdout.String(), p.stderr.String())
}

func (p *phase0ServerProcess) WaitForExit(t testing.TB) error {
	t.Helper()

	select {
	case err := <-p.done:
		return err
	case <-time.After(15 * time.Second):
		t.Fatalf("timed out waiting for cmd/server exit\nstdout:\n%s\nstderr:\n%s", p.stdout.String(), p.stderr.String())
		return nil
	}
}

func (p *phase0ServerProcess) Stop(t testing.TB) {
	t.Helper()

	if p == nil {
		return
	}
	p.cancel()
	if p.cmd != nil && p.cmd.Process != nil {
		_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGTERM)
	}
	select {
	case <-p.done:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out stopping cmd/server\nstdout:\n%s\nstderr:\n%s", p.stdout.String(), p.stderr.String())
	}
}

func (p *phase0ServerProcess) RequireStatus(t testing.TB, path string, want int) {
	t.Helper()

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + p.address + path)
	if err != nil {
		t.Fatalf("request %s: %v", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != want {
		t.Fatalf("unexpected status for %s: got %d want %d", path, resp.StatusCode, want)
	}
}

func (p *phase0ServerProcess) RequireConnectionRefused(t testing.TB, path string) {
	t.Helper()

	client := &http.Client{Timeout: 500 * time.Millisecond}
	if resp, err := client.Get("http://" + p.address + path); err == nil {
		resp.Body.Close()
		t.Fatalf("expected %s to be unreachable, got HTTP %d", path, resp.StatusCode)
	}
}

func (p *phase0ServerProcess) RequireWebsocketConnectionRefused(t testing.TB, path string) {
	t.Helper()

	_, _, err := wstest.TryConnect("http://"+p.address, path, nil)
	wstest.RequireConnectionRefused(t, err)
}

func (p *phase0ServerProcess) RequireDiagnosticsCode(t testing.TB, wantCode string) {
	t.Helper()

	payload := p.parseDiagnostics(t)
	errorPayload := payload["error"].(map[string]any)
	if errorPayload["code"] != wantCode {
		t.Fatalf("unexpected diagnostics code: got %v want %s", errorPayload["code"], wantCode)
	}
}

func (p *phase0ServerProcess) RequireReasonCode(t testing.TB, wantReasonCode string) {
	t.Helper()

	payload := p.parseDiagnostics(t)
	items := payload["error"].(map[string]any)["details"].(map[string]any)["items"].([]any)
	for _, item := range items {
		typed := item.(map[string]any)
		if typed["reason_code"] == wantReasonCode {
			return
		}
	}
	t.Fatalf("missing reason_code=%q in stderr payload %s", wantReasonCode, p.stderr.String())
}

func (p *phase0ServerProcess) RequireDiagnosticsField(t testing.TB, wantPath string, wantReasonCode string) {
	t.Helper()

	payload := p.parseDiagnostics(t)
	items := payload["error"].(map[string]any)["details"].(map[string]any)["items"].([]any)
	for _, item := range items {
		typed := item.(map[string]any)
		if typed["path"] == wantPath && typed["reason_code"] == wantReasonCode {
			return
		}
	}
	t.Fatalf("missing diagnostic path=%q reason_code=%q in stderr payload %s", wantPath, wantReasonCode, p.stderr.String())
}

func (p *phase0ServerProcess) parseDiagnostics(t testing.TB) map[string]any {
	t.Helper()

	stderr := strings.TrimSpace(p.stderr.String())
	if stderr == "" {
		t.Fatal("expected structured startup diagnostics on stderr")
	}
	if idx := strings.LastIndex(stderr, "\nexit status "); idx >= 0 {
		stderr = strings.TrimSpace(stderr[:idx])
	}
	if strings.HasPrefix(stderr, "exit status ") {
		t.Fatalf("missing structured startup diagnostics on stderr\nstderr:\n%s", p.stderr.String())
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("decode stderr diagnostics JSON: %v\nstderr:\n%s", err, stderr)
	}
	return payload
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
	env["CARTULARY_CONFIG_FILE"] = configPath
	env["CARTULARY_HTTP_ADDR"] = phase0FreeAddress(t)
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

func phase0FreeAddress(t testing.TB) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve free tcp port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().String()
}

func phase0EnvPairs(env map[string]string) []string {
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		pairs = append(pairs, key+"="+value)
	}
	return pairs
}

func phase0RepoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}

func phase0BucketName(prefix string) string {
	value := strings.ToLower(prefix)
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return fmt.Sprintf("%s-%d", value, time.Now().UnixNano())
}

func dropPhase0Database(t testing.TB, harness *pgtest.Harness, name string) {
	t.Helper()
	if err := harness.DropDatabase(context.Background(), name); err != nil {
		t.Fatalf("drop postgres database: %v", err)
	}
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
