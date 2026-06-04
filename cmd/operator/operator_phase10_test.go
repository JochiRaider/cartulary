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
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
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
	objectSnapshotBody := []byte(`{"schema_id":"cartulary.object_store_snapshot_artifact.v2","objects":[]}`)
	objectManifestBody, objectSummaryBody := operatorPhaseEObjectArtifacts(t, backupSetID, asOf.Add(-time.Hour), "phase10-operator-inspection", objectSnapshotBody, nil)
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatalf("open pgx pool for operator fixture: %v", err)
	}
	t.Cleanup(pool.Close)
	backupStorage := newOperatorEncryptedBackupStorage(t, env["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"])
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
			Body:        objectSnapshotBody,
			ContentType: "application/json",
		},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectSummaryBody, ContentType: "application/json"},
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

func TestPhase10_E_10_01_DeploymentLocalOperatorRestoreLatestBackup(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-operator-restore-source")
	targetDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-operator-restore-target")
	sourceConfig := operatorExplicitConfig(t, sourceDB.DSN)
	targetConfig := operatorExplicitConfig(t, targetDB.DSN)

	adminEmail := "phase10-e-10-01-restore-admin@example.test"
	seedOperatorUser(t, sourceDB.DSN, adminEmail, true, true)

	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source pgx pool: %v", err)
	}
	t.Cleanup(sourcePool.Close)
	sourceObjectStore, err := objectstore.NewFilesystemStore(sourceConfig.objectRoot)
	if err != nil {
		t.Fatalf("open source object store: %v", err)
	}
	t.Cleanup(func() {
		_ = sourceObjectStore.Close()
	})
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, sourcePool)
	if err != nil {
		t.Fatalf("capture source postgres artifact: %v", err)
	}
	objectArtifact, err := recovery.CaptureObjectStoreSnapshotArtifact(ctx, sourceObjectStore, "")
	if err != nil {
		t.Fatalf("capture source object artifact: %v", err)
	}
	backupStorage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000102101")
	objectManifestBody, objectSummaryBody := operatorPhaseEObjectArtifacts(t, backupSetID, asOf.Add(-time.Minute), "phase10-operator-restore", objectArtifact, nil)
	if _, err := recovery.NewCaptureService(recovery.NewStore(sourcePool), backupStorage).CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
		BackupSetID:        backupSetID,
		ConsistencyPointAt: asOf.Add(-time.Minute),
		CreatedAt:          asOf.Add(-2 * time.Minute),
		RetainedUntil:      asOf.Add(31 * 24 * time.Hour),
		PostgresArtifact: recovery.BackupArtifact{
			Body:        postgresArtifact,
			ContentType: "application/json",
		},
		ObjectStoreArtifact: recovery.BackupArtifact{
			Body:        objectArtifact,
			ContentType: "application/json",
		},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectSummaryBody, ContentType: "application/json"},
	}); err != nil {
		t.Fatalf("capture restorable backup set: %v", err)
	}

	operatorBin := buildOperatorBinary(t)
	stdout, stderr, exitCode := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore", "latest",
		"-source-config", sourceConfig.path,
		"-target-config", targetConfig.path,
		"-deployment-admin-email", adminEmail,
		"-confirm-backup-set-id", backupSetID.String(),
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if exitCode != 0 {
		t.Fatalf("operator restore failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("decode restore JSON: %v\nstdout=%s", err, stdout)
	}
	if payload["schema_id"] != app.OperatorRestoreResultSchemaID || payload["backup_set_id"] != backupSetID.String() {
		t.Fatalf("unexpected restore payload: %#v", payload)
	}

	targetSQL, err := sql.Open("pgx", targetDB.DSN)
	if err != nil {
		t.Fatalf("open target sql: %v", err)
	}
	defer targetSQL.Close()
	var restoredAdminCount int
	if err := targetSQL.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE email = $1 AND is_deployment_admin = true`, adminEmail).Scan(&restoredAdminCount); err != nil {
		t.Fatalf("query restored deployment admin: %v", err)
	}
	if restoredAdminCount != 1 {
		t.Fatalf("operator restore did not copy authoritative source rows, restored admin count=%d", restoredAdminCount)
	}

	_, sameConfigStderr, sameConfigExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore", "latest",
		"-source-config", sourceConfig.path,
		"-target-config", sourceConfig.path,
		"-deployment-admin-email", adminEmail,
		"-confirm-backup-set-id", backupSetID.String(),
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if sameConfigExit == 0 {
		t.Fatal("operator restore with identical source and target configs unexpectedly succeeded")
	}
	if !strings.Contains(sameConfigStderr, "source-config and target-config must be different files") {
		t.Fatalf("same-config restore did not fail with target preflight error: %s", sameConfigStderr)
	}
}

func TestPhase10_E_10_01_DeploymentLocalOperatorRestoreVerifyDueRunner(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-operator-due-source")
	targetDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-operator-due-target")
	sourceConfig := operatorExplicitConfig(t, sourceDB.DSN)
	targetConfig := operatorExplicitConfig(t, targetDB.DSN)

	adminEmail := "phase10-e-10-01-due-admin@example.test"
	seedOperatorUser(t, sourceDB.DSN, adminEmail, true, true)

	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source pgx pool: %v", err)
	}
	t.Cleanup(sourcePool.Close)
	sourceObjectStore, err := objectstore.NewFilesystemStore(sourceConfig.objectRoot)
	if err != nil {
		t.Fatalf("open source object store: %v", err)
	}
	t.Cleanup(func() {
		_ = sourceObjectStore.Close()
	})
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, sourcePool)
	if err != nil {
		t.Fatalf("capture source postgres artifact: %v", err)
	}
	objectArtifact, err := recovery.CaptureObjectStoreSnapshotArtifact(ctx, sourceObjectStore, "")
	if err != nil {
		t.Fatalf("capture source object artifact: %v", err)
	}
	backupStorage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	sourceStore := recovery.NewStore(sourcePool)
	capture := recovery.NewCaptureService(sourceStore, backupStorage)
	asOf := time.Date(2026, 5, 22, 12, 0, 0, 0, time.UTC)
	sourceCfg := loadOperatorConfig(t, sourceConfig.path)
	targetCfg := loadOperatorConfig(t, targetConfig.path)
	basis := operatorRestoreVerificationBasis(t, sourceCfg)
	otherBasis := strings.Repeat("f", 64)
	dummyReport := recovery.RestoreConsistencyReport{
		AuthoritativeRowsSHA256: strings.Repeat("a", 64),
		AuthoritativeRowCount:   1,
		ChangeSetsSHA256:        strings.Repeat("b", 64),
		ChangeSetRowCount:       1,
		BlobHashesSHA256:        strings.Repeat("c", 64),
		BlobCount:               0,
	}
	captured := make([]recovery.BackupSet, 0, 4)
	for index, backupSetID := range []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000102201"),
		uuid.MustParse("00000000-0000-0000-0000-000000102202"),
		uuid.MustParse("00000000-0000-0000-0000-000000102203"),
		uuid.MustParse("00000000-0000-0000-0000-000000102204"),
	} {
		consistencyPointAt := asOf.Add(-time.Duration(index+1) * time.Minute)
		objectManifestBody, objectSummaryBody := operatorPhaseEObjectArtifacts(t, backupSetID, consistencyPointAt, "phase10-operator-due", objectArtifact, nil)
		backupSet, err := capture.CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
			BackupSetID:        backupSetID,
			ConsistencyPointAt: consistencyPointAt,
			CreatedAt:          asOf.Add(-time.Duration(index+2) * time.Minute),
			RetainedUntil:      asOf.Add(31 * 24 * time.Hour),
			PostgresArtifact:   recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
			ObjectStoreArtifact: recovery.BackupArtifact{
				Body:        objectArtifact,
				ContentType: "application/json",
			},
			ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectManifestBody, ContentType: "application/json"},
			ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectSummaryBody, ContentType: "application/json"},
		})
		if err != nil {
			t.Fatalf("capture due backup set %d: %v", index, err)
		}
		captured = append(captured, backupSet)
	}
	if _, _, err := sourceStore.RecordRestoreVerificationCompletion(ctx, recovery.CreateRestoreVerificationRunParams{
		BackupSetID:             captured[1].BackupSetID,
		StartedAt:               asOf.Add(-8*24*time.Hour - time.Minute),
		CompletedAt:             asOf.Add(-8 * 24 * time.Hour),
		VerificationState:       recovery.VerificationVerified,
		VerificationBasisSHA256: basis,
		ConsistencyReport:       dummyReport,
	}); err != nil {
		t.Fatalf("seed stale verification state: %v", err)
	}
	if _, _, err := sourceStore.RecordRestoreVerificationCompletion(ctx, recovery.CreateRestoreVerificationRunParams{
		BackupSetID:             captured[2].BackupSetID,
		StartedAt:               asOf.Add(-time.Hour - time.Minute),
		CompletedAt:             asOf.Add(-time.Hour),
		VerificationState:       recovery.VerificationVerified,
		VerificationBasisSHA256: otherBasis,
		ConsistencyReport:       dummyReport,
	}); err != nil {
		t.Fatalf("seed basis-changed verification state: %v", err)
	}
	if _, _, err := sourceStore.RecordRestoreVerificationCompletion(ctx, recovery.CreateRestoreVerificationRunParams{
		BackupSetID:             captured[3].BackupSetID,
		StartedAt:               asOf.Add(-30 * time.Minute),
		CompletedAt:             asOf.Add(-29 * time.Minute),
		VerificationState:       recovery.VerificationVerified,
		VerificationBasisSHA256: basis,
		ConsistencyReport:       dummyReport,
	}); err != nil {
		t.Fatalf("seed fresh verification state: %v", err)
	}
	operatorBin := buildOperatorBinary(t)
	_, unsafeStderr, unsafeExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "due",
		"-source-config", sourceConfig.path,
		"-target-config", targetConfig.path,
		"-deployment-admin-email", adminEmail,
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if unsafeExit == 0 {
		t.Fatalf("restore-verify due without target marker unexpectedly succeeded")
	}
	if !strings.Contains(unsafeStderr, "target marker") {
		t.Fatalf("unsafe restore-verification target failure did not mention marker: %s", unsafeStderr)
	}

	writeRestoreVerificationTargetMarker(t, targetCfg)
	stdout, stderr, exitCode := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "due",
		"-source-config", sourceConfig.path,
		"-target-config", targetConfig.path,
		"-deployment-admin-email", adminEmail,
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if exitCode != 0 {
		t.Fatalf("operator restore-verify due failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("decode due runner JSON: %v\nstdout=%s", err, stdout)
	}
	if payload["schema_id"] != app.OperatorRestoreVerificationDueSchemaID ||
		int(payload["due_count"].(float64)) != 3 ||
		int(payload["verified_count"].(float64)) != 3 ||
		int(payload["failed_count"].(float64)) != 0 {
		t.Fatalf("unexpected due runner payload: %#v", payload)
	}

	secondStdout, secondStderr, secondExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "due",
		"-source-config", sourceConfig.path,
		"-target-config", targetConfig.path,
		"-deployment-admin-email", adminEmail,
		"-as-of", asOf.Format(time.RFC3339Nano),
	)
	if secondExit != 0 {
		t.Fatalf("second operator restore-verify due failed: exit=%d stdout=%s stderr=%s", secondExit, secondStdout, secondStderr)
	}
	var secondPayload map[string]any
	if err := json.Unmarshal([]byte(secondStdout), &secondPayload); err != nil {
		t.Fatalf("decode second due runner JSON: %v\nstdout=%s", err, secondStdout)
	}
	if int(secondPayload["due_count"].(float64)) != 0 {
		t.Fatalf("fresh same-basis due runner should skip all backups, got %#v", secondPayload)
	}
}

type operatorExplicitConfigFixture struct {
	path       string
	objectRoot string
	backupRoot string
}

func buildOperatorBinary(t testing.TB) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "cartulary-operator")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "make", "--no-print-directory", "build-operator", "OPERATOR_BIN="+bin)
	cmd.Dir = repoRoot()
	cmd.Env = append(operatorBuildEnv(),
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

func operatorBuildEnv() []string {
	env := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "PHASE=") ||
			strings.HasPrefix(entry, "MAKEFLAGS=") ||
			strings.HasPrefix(entry, "MFLAGS=") {
			continue
		}
		env = append(env, entry)
	}
	return env
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
	env[recovery.RecoveryMasterKeyEnv] = operatorRecoveryMasterKey
	return env
}

func operatorExplicitConfig(t testing.TB, dsn string) operatorExplicitConfigFixture {
	t.Helper()

	roots := configtest.SetupTempRoots(t)
	configtest.BindPostgresDSNToDatabaseRoot(t, roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], dsn)
	contents := string(fixtures.MustRead("config", "valid.toml"))
	replacements := map[string]string{
		"/var/lib/cartulary/postgres":         roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"],
		"/var/lib/cartulary/object-store":     roots.Paths["CARTULARY__ROOTS__OBJECT_STORAGE__PATH"],
		"/var/lib/cartulary/backups":          roots.Paths["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"],
		"/var/lib/cartulary/reference-packs":  roots.Paths["CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH"],
		"/var/lib/cartulary/tmp":              roots.Paths["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"],
		"/var/lib/cartulary/exports":          roots.Paths["CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH"],
		"/etc/cartulary/bootstrap-admin.json": fixtures.Path("bootstrap-admin", "canonical.json"),
	}
	for oldValue, newValue := range replacements {
		contents = strings.ReplaceAll(contents, oldValue, newValue)
	}
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write explicit operator config: %v", err)
	}
	return operatorExplicitConfigFixture{
		path:       configPath,
		objectRoot: roots.Paths["CARTULARY__ROOTS__OBJECT_STORAGE__PATH"],
		backupRoot: roots.Paths["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"],
	}
}

func loadOperatorConfig(t testing.TB, path string) config.Config {
	t.Helper()
	cfg, err := config.LoadWithOptions(config.LoadOptions{Path: path})
	if err != nil {
		t.Fatalf("load operator config %s: %v", path, err)
	}
	return cfg
}

func operatorRestoreVerificationBasis(t testing.TB, cfg config.Config) string {
	t.Helper()
	basis, err := recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism":         "cartulary.phase10.filesystem_snapshot.v1",
		"database_storage_binding": operatorRootBindingBasis(cfg.Roots.DatabaseStorage),
		"object_storage_binding":   operatorRootBindingBasis(cfg.Roots.ObjectStorage),
		"backup_storage_binding":   operatorRootBindingBasis(cfg.Roots.BackupStorage),
	})
	if err != nil {
		t.Fatalf("compute restore verification basis: %v", err)
	}
	return basis
}

func operatorRootBindingBasis(binding config.RootBinding) string {
	switch binding.BindingKind {
	case "filesystem_root":
		return "filesystem_root:" + filepath.Clean(binding.Path)
	case "managed_service":
		return "managed_service:" + strings.TrimSpace(binding.ServiceRef)
	default:
		return strings.TrimSpace(binding.BindingKind)
	}
}

func writeRestoreVerificationTargetMarker(t testing.TB, cfg config.Config) {
	t.Helper()
	markerPath, err := app.RestoreVerificationTargetMarkerPath(cfg)
	if err != nil {
		t.Fatalf("resolve restore verification target marker path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(markerPath), 0o700); err != nil {
		t.Fatalf("create restore verification target marker directory: %v", err)
	}
	body := []byte(`{"schema_id":"` + app.RestoreVerificationTargetMarkerSchemaID + `","purpose":"restore_verification_target"}`)
	if err := os.WriteFile(markerPath, body, 0o600); err != nil {
		t.Fatalf("write restore verification target marker: %v", err)
	}
}

const operatorRecoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func operatorRecoveryEnv() map[string]string {
	return map[string]string{recovery.RecoveryMasterKeyEnv: operatorRecoveryMasterKey}
}

func newOperatorEncryptedBackupStorage(t testing.TB, root string) recovery.BackupStorage {
	t.Helper()
	rawStorage, err := recovery.NewFilesystemBackupStorage(root)
	if err != nil {
		t.Fatalf("create operator backup storage: %v", err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(operatorRecoveryMasterKey)
	if err != nil {
		t.Fatalf("parse operator recovery key: %v", err)
	}
	storage, err := recovery.NewEncryptedBackupStorage(rawStorage, key)
	if err != nil {
		t.Fatalf("create encrypted operator backup storage: %v", err)
	}
	return storage
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

func operatorPhaseEObjectArtifacts(t testing.TB, backupSetID uuid.UUID, consistencyPointAt time.Time, bucket string, objectSnapshotBody []byte, blobIndex map[string]uuid.UUID) ([]byte, []byte) {
	t.Helper()
	snapshot, err := recovery.DecodeObjectStoreSnapshotArtifact(objectSnapshotBody)
	if err != nil {
		t.Fatalf("decode operator object-store snapshot: %v", err)
	}
	manifest, manifestBody, err := recovery.BuildSeaweedFSS3ObjectStoreBackupManifest(snapshot, recovery.ObjectStoreBackupManifestParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        consistencyPointAt,
		Bucket:                    bucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("build operator object-store backup manifest: %v", err)
	}
	_, summaryBody, err := recovery.BuildObjectStoreBackupSummary(manifest)
	if err != nil {
		t.Fatalf("build operator object-store backup summary: %v", err)
	}
	return manifestBody, summaryBody
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
