package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/app"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPhase10_E_10_01_DeploymentLocalOperatorInspectLatestBackupMetadata(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-operator")
	env := operatorProcessEnv(t, testDB.Env())

	adminEmail := "phase10-e-10-01-admin@example.test"
	nonAdminEmail := "phase10-e-10-01-viewer@example.test"
	inactiveAdminEmail := "phase10-e-10-01-inactive-admin@example.test"
	incidentAdminEmail := "phase10-e-10-01-incident-admin@example.test"
	adminID := seedOperatorUser(t, testDB.DSN, adminEmail, true, true)
	seedOperatorUser(t, testDB.DSN, nonAdminEmail, false, true)
	seedOperatorUser(t, testDB.DSN, inactiveAdminEmail, true, false)
	incidentAdminID := seedOperatorUser(t, testDB.DSN, incidentAdminEmail, false, true)
	seedOperatorIncidentAdmin(t, testDB.DSN, adminID, incidentAdminID)

	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	createdAt := asOf.Add(-2 * time.Hour)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000102001")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatalf("open pgx pool for operator fixture: %v", err)
	}
	t.Cleanup(pool.Close)
	backupStorage, err := recovery.NewFilesystemBackupStorage(env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"])
	if err != nil {
		t.Fatalf("create operator backup storage: %v", err)
	}
	if _, err := recovery.NewCaptureService(recovery.NewStore(pool), backupStorage).CaptureBackupSet(context.Background(), recovery.CaptureBackupSetParams{
		BackupSetID:        backupSetID,
		ConsistencyPointAt: asOf.Add(-time.Hour),
		CreatedAt:          createdAt,
		RetainedUntil:      createdAt.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        []byte(`{"schema_id":"phase10.operator.postgres_artifact.v1"}`),
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        []byte(`{"schema_id":"phase10.operator.object_store_artifact.v1"}`),
			ContentType: "application/json",
		},
	}); err != nil {
		t.Fatalf("seed backup metadata for operator inspection: %v", err)
	}

	operatorBin := buildOperatorBinary(t)
	stdout, stderr, exitCode := runOperatorBinary(t, operatorBin, env,
		"backup-metadata", "latest",
		"-deployment-admin-email", adminEmail,
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if exitCode != 0 {
		t.Fatalf("operator inspection failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("decode operator JSON: %v\nstdout=%s", err, stdout)
	}
	if payload["schema_id"] != app.BackupMetadataInspectionSchemaID {
		t.Fatalf("unexpected operator schema_id: %#v", payload)
	}
	if payload["backup_set_id"] != backupSetID.String() {
		t.Fatalf("operator returned wrong backup_set_id: %#v", payload)
	}
	if !operatorPayloadStringHasPrefix(payload, "postgres_restore_anchor", "backup-storage://backup_sets/") ||
		!operatorPayloadStringHasPrefix(payload, "object_store_restore_anchor", "backup-storage://backup_sets/") {
		t.Fatalf("operator did not expose both restore anchors: %#v", payload)
	}
	requireOperatorArtifactProof(t, payload)
	if payload["verification_state"] != string(recovery.VerificationUnverified) {
		t.Fatalf("operator verification_state got %#v want %s", payload["verification_state"], recovery.VerificationUnverified)
	}
	if payload["last_verified_restore_at"] != nil {
		t.Fatalf("operator unverified metadata must expose null last_verified_restore_at: %#v", payload)
	}
	requireOperatorRetentionFloor(t, payload, "retained_until")
	requireOperatorRetentionFloor(t, payload, "postgres_restore_anchor_retained_until")
	requireOperatorRetentionFloor(t, payload, "object_store_restore_anchor_retained_until")

	_, nonAdminStderr, nonAdminExit := runOperatorBinary(t, operatorBin, env,
		"backup-metadata", "latest",
		"-deployment-admin-email", nonAdminEmail,
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if nonAdminExit == 0 {
		t.Fatalf("non-deployment-admin operator inspection unexpectedly succeeded")
	}
	if !strings.Contains(nonAdminStderr, "deployment admin authorization failed") {
		t.Fatalf("non-admin failure did not report deployment-admin authorization: %s", nonAdminStderr)
	}

	_, inactiveAdminStderr, inactiveAdminExit := runOperatorBinary(t, operatorBin, env,
		"backup-metadata", "latest",
		"-deployment-admin-email", inactiveAdminEmail,
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if inactiveAdminExit == 0 {
		t.Fatalf("inactive deployment-admin operator inspection unexpectedly succeeded")
	}
	if !strings.Contains(inactiveAdminStderr, "deployment admin authorization failed") {
		t.Fatalf("inactive admin failure did not report deployment-admin authorization: %s", inactiveAdminStderr)
	}

	_, incidentAdminStderr, incidentAdminExit := runOperatorBinary(t, operatorBin, env,
		"backup-metadata", "latest",
		"-deployment-admin-email", incidentAdminEmail,
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if incidentAdminExit == 0 {
		t.Fatalf("incident-admin-only operator inspection unexpectedly succeeded")
	}
	if !strings.Contains(incidentAdminStderr, "deployment admin authorization failed") {
		t.Fatalf("incident-admin-only failure did not report deployment-admin authorization: %s", incidentAdminStderr)
	}
}

func buildOperatorBinary(t testing.TB) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "cartulary-operator")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "make", "--no-print-directory", "build-operator", "OPERATOR_BIN="+bin)
	cmd.Dir = repoRoot()
	cmd.Env = append(os.Environ(),
		"CARTULARY_TEST_RESULTS_DIR="+filepath.Join(t.TempDir(), "results"),
		"CARTULARY_TEST_RUN_ID=operator-build",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build operator binary: %v\nstderr=%s", err, stderr.String())
	}
	return bin
}

func runOperatorBinary(t testing.TB, bin string, env map[string]string, args ...string) (string, string, int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = repoRoot()
	cmd.Env = append(os.Environ(), envPairs(env)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	}
	t.Fatalf("run operator binary: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	return "", "", 1
}

func operatorProcessEnv(t testing.TB, databaseEnv map[string]string) map[string]string {
	t.Helper()

	tempRoots := configtest.SetupTempRoots(t)
	env := make(map[string]string, len(databaseEnv)+len(tempRoots.Paths)+2)
	for key, value := range databaseEnv {
		env[key] = value
	}
	for key, value := range tempRoots.Paths {
		env[key] = value
	}
	configtest.BindPostgresEnvToDatabaseRoot(t, tempRoots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], env)
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, fixtures.MustRead("config", "valid.toml"), 0o644); err != nil {
		t.Fatalf("write operator config fixture: %v", err)
	}
	env["CARTULARY_CONFIG_FILE"] = configPath
	return env
}

func seedOperatorUser(t testing.TB, dsn string, email string, deploymentAdmin bool, active bool) uuid.UUID {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql DB for operator user seed: %v", err)
	}
	defer db.Close()
	var userID uuid.UUID
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'not-used-by-operator', false, $3, $4)
RETURNING id
`, email, email, active, deploymentAdmin).Scan(&userID); err != nil {
		t.Fatalf("seed operator user %s: %v", email, err)
	}
	return userID
}

func seedOperatorIncidentAdmin(t testing.TB, dsn string, creatorID uuid.UUID, incidentAdminID uuid.UUID) {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql DB for operator incident-admin seed: %v", err)
	}
	defer db.Close()
	var incidentID uuid.UUID
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO incidents (incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id)
VALUES ('phase10-e-10-01', 'phase10-e-10-01', 'Phase 10 operator auth boundary', 'active', $1, $1)
RETURNING id
`, creatorID).Scan(&incidentID); err != nil {
		t.Fatalf("seed operator incident: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO incident_memberships (incident_id, user_id, role, added_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'admin', $3, $3)
`, incidentID, incidentAdminID, creatorID); err != nil {
		t.Fatalf("seed incident-admin-only membership: %v", err)
	}
}

func requireOperatorRetentionFloor(t testing.TB, payload map[string]any, retainedField string) {
	t.Helper()

	createdRaw, ok := payload["created_at"].(string)
	if !ok {
		t.Fatalf("operator payload missing created_at string: %#v", payload)
	}
	retainedRaw, ok := payload[retainedField].(string)
	if !ok {
		t.Fatalf("operator payload missing %s string: %#v", retainedField, payload)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdRaw)
	if err != nil {
		t.Fatalf("parse created_at %q: %v", createdRaw, err)
	}
	retainedUntil, err := time.Parse(time.RFC3339Nano, retainedRaw)
	if err != nil {
		t.Fatalf("parse %s %q: %v", retainedField, retainedRaw, err)
	}
	if retainedUntil.Before(createdAt.Add(recovery.MinimumRetentionDuration)) {
		t.Fatalf("%s before 30-day retention floor: created_at=%s retained_until=%s", retainedField, createdAt, retainedUntil)
	}
}

func requireOperatorArtifactProof(t testing.TB, payload map[string]any) {
	t.Helper()
	for _, field := range []string{
		"postgres_artifact_key",
		"object_store_artifact_key",
		"integrity_manifest_key",
	} {
		if value, ok := payload[field].(string); !ok || value == "" {
			t.Fatalf("operator payload missing artifact key %s: %#v", field, payload)
		}
	}
	for _, field := range []string{
		"postgres_artifact_sha256",
		"object_store_artifact_sha256",
		"integrity_manifest_sha256",
	} {
		if value, ok := payload[field].(string); !ok || len(value) != 64 {
			t.Fatalf("operator payload missing sha256 proof %s: %#v", field, payload)
		}
	}
	for _, field := range []string{
		"postgres_artifact_size_bytes",
		"object_store_artifact_size_bytes",
		"integrity_manifest_size_bytes",
	} {
		if value, ok := payload[field].(float64); !ok || value <= 0 {
			t.Fatalf("operator payload missing positive size %s: %#v", field, payload)
		}
	}
}

func operatorPayloadStringHasPrefix(payload map[string]any, field string, prefix string) bool {
	value, ok := payload[field].(string)
	return ok && strings.HasPrefix(value, prefix)
}

func envPairs(env map[string]string) []string {
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		pairs = append(pairs, key+"="+value)
	}
	return pairs
}

func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..")
}
