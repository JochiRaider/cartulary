package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorops"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/configtest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestMVPObjectStoreInitOperatorCreatesConfiguredBucket(t *testing.T) {
	ctx := context.Background()
	s3Harness := s3test.Start(t)
	bucket := fmt.Sprintf("mvp-object-store-init-%d", time.Now().UnixNano())
	defer func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	}()

	configFixture := operatorManagedS3Config(t, "postgres://unused", "object_primary")
	cfg, err := config.LoadWithOptions(config.LoadOptions{Path: configFixture.path})
	if err != nil {
		t.Fatalf("load managed S3 config fixture: %v", err)
	}
	env := mergeOperatorEnv(operatorRecoveryEnv(), s3Harness.EnvForServiceRef("object_primary", bucket))
	if _, err := objectstore.SetupWithEnv(ctx, cfg, env); err == nil {
		t.Fatal("expected object-store setup to fail while configured bucket is missing")
	} else {
		adapterErr, ok := objectstore.AsAdapterError(err)
		if !ok || adapterErr.Reason != objectstore.ReasonBucketMissing {
			t.Fatalf("unexpected pre-init object-store error: %#v %v", adapterErr, err)
		}
	}

	operatorBin := buildOperatorBinary(t)
	stdout, stderr, exitCode := runOperatorBinary(t, operatorBin, env, "object-store", "init", "-config", configFixture.path)
	if exitCode != 0 {
		t.Fatalf("object-store init failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("decode object-store init JSON: %v\nstdout=%s", err, stdout)
	}
	if payload["schema_id"] != app.OperatorObjectStoreInitResultSchemaID || payload["result"] != "created" || payload["created"] != true {
		t.Fatalf("unexpected object-store init payload: %#v", payload)
	}
	if strings.Contains(stdout, bucket) || strings.Contains(stdout, s3Harness.Endpoint) || strings.Contains(stderr, bucket) || strings.Contains(stderr, s3Harness.Endpoint) {
		t.Fatalf("object-store init exposed storage details: stdout=%s stderr=%s", stdout, stderr)
	}

	store, err := objectstore.SetupWithEnv(ctx, cfg, env)
	if err != nil {
		t.Fatalf("object-store setup failed after init: %v", err)
	}
	_ = store.Close()
}

func TestPhase10_E_10_01_CanonicalOperatorBackupInspectLatest(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-inspect")
	sourceConfig := operatorExplicitConfig(t, sourceDB.DSN)
	sourcePool := mustOpenOperatorPool(t, sourceDB.DSN)
	backupStorage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000102501")
	seedOperatorRecoveryBackupSet(t, ctx, sourcePool, sourceConfig, backupStorage, backupSetID, time.Now().UTC().Add(-time.Minute), "phase10-canonical-inspect")

	operatorBin := buildOperatorBinary(t)
	stdout, stderr, exitCode := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"backup", "inspect", "latest",
		"--source-config-file", sourceConfig.path,
		"--progress", "jsonl",
	)
	if exitCode != 0 {
		t.Fatalf("canonical backup inspect failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	payload := decodeOperatorRecoveryResult(t, stdout)
	requireOperatorRecoverySuccess(t, payload, "backup_inspect_latest", backupSetID.String())
	requireOperatorRecoveryArtifactKind(t, payload, "backup_attestation", "cartulary.backup_attestation.v1", 1)
	requireOperatorRecoveryProgress(t, stderr, payload.OperationID, []string{"preflight", "catalog_select", "artifact_validate", "finalize"})
	requireOperatorRecoverySafeOutput(t, stdout, stderr, sourceDB.DSN, sourceConfig.path, sourceConfig.objectRoot, sourceConfig.backupRoot, operatorRecoveryMasterKey)

	noProgressStdout, noProgressStderr, noProgressExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"backup", "inspect", "latest",
		"--source-config-file", sourceConfig.path,
	)
	if noProgressExit != 0 {
		t.Fatalf("canonical backup inspect without progress failed: exit=%d stdout=%s stderr=%s", noProgressExit, noProgressStdout, noProgressStderr)
	}
	if noProgressStderr != "" {
		t.Fatalf("backup inspect without progress emitted stderr: %s", noProgressStderr)
	}
	noProgressPayload := decodeOperatorRecoveryResult(t, noProgressStdout)
	requireOperatorRecoverySuccess(t, noProgressPayload, "backup_inspect_latest", backupSetID.String())

	invalidStdout, invalidStderr, invalidExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"backup", "inspect", "latest",
		"--source-config-file", sourceConfig.path,
		"--deployment-admin-email", "phase10-admin@example.test",
	)
	requireOperatorRecoveryFailure(t, invalidStdout, invalidStderr, invalidExit, "backup_inspect_latest", 2, "invalid_operator_request", "invalid_flag_value")
}

func TestPhase10_E_10_01_CanonicalOperatorBackupCreate(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-create")
	s3Harness := s3test.Start(t)
	bucket := fmt.Sprintf("phase10-canonical-create-%d", time.Now().UnixNano())
	if err := s3Harness.CreateBucket(ctx, bucket); err != nil {
		t.Fatalf("create backup create bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	sourceConfig := operatorManagedS3Config(t, sourceDB.DSN, "backup-create")
	actorID := seedOperatorUser(t, sourceDB.DSN, "phase10-e-10-01-canonical-create@example.test", true, true)
	blob := seedOperatorMigrationBlob(t, sourceDB.DSN, actorID, []byte("phase10 canonical backup create object proof"))
	if _, err := s3Harness.RoundTrip(ctx, bucket, blob.storageKey, blob.body); err != nil {
		t.Fatalf("write source object for canonical backup create: %v", err)
	}

	env := mergeOperatorEnv(operatorRecoveryEnv(), s3Harness.EnvForServiceRef("backup-create", bucket))
	operatorBin := buildOperatorBinary(t)
	stdout, stderr, exitCode := runOperatorBinaryWithTimeout(t, 30*time.Second, operatorBin, env,
		"backup", "create",
		"--source-config-file", sourceConfig.path,
		"--progress", "jsonl",
	)
	if exitCode != 0 {
		t.Fatalf("canonical backup create failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	payload := decodeOperatorRecoveryResult(t, stdout)
	requireOperatorRecoverySuccess(t, payload, "backup_create", "")
	if payload.BackupSetID == nil {
		t.Fatalf("backup create result missing backup_set_id: %#v", payload)
	}
	requireOperatorRecoveryArtifactKind(t, payload, "backup_attestation", "cartulary.backup_attestation.v1", 1)
	requireOperatorRecoveryProgress(t, stderr, payload.OperationID, []string{"preflight", "postgres_backup", "object_backup", "attestation_write", "journal_write", "finalize"})
	requireOperatorRecoverySafeOutput(t, stdout, stderr, bucket, s3Harness.Endpoint, blob.storageKey, blob.storageRef, sourceDB.DSN, sourceConfig.path, operatorRecoveryMasterKey)
	requireOperatorRecoveryJournalAndAudit(t, sourceDB.DSN, payload, "backup_create", "succeeded", bucket, s3Harness.Endpoint, blob.storageKey, blob.storageRef, sourceDB.DSN, sourceConfig.path, operatorRecoveryMasterKey)

	rawArtifact := filepath.Join(sourceConfig.backupRoot, "backup_sets", *payload.BackupSetID, "object-store-artifact.json")
	rawBody, err := os.ReadFile(rawArtifact)
	if err != nil {
		t.Fatalf("read raw encrypted backup artifact: %v", err)
	}
	for _, forbidden := range []string{string(blob.body), blob.storageKey, blob.storageRef, bucket, s3Harness.Endpoint} {
		if strings.Contains(string(rawBody), forbidden) {
			t.Fatalf("raw encrypted backup artifact contains forbidden plaintext %q: %s", forbidden, rawBody)
		}
	}

	backupStorage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	reloaded, err := recovery.NewBackupCatalog(recovery.NewStore(mustOpenOperatorPool(t, sourceDB.DSN)), backupStorage).RestoreCandidateBackup(ctx, time.Now().UTC().Add(time.Minute))
	if err != nil {
		t.Fatalf("created backup did not remain selectable and durable: %v", err)
	}
	if reloaded.BackupSetID.String() != *payload.BackupSetID || reloaded.VerificationState != recovery.VerificationUnverified {
		t.Fatalf("reloaded backup changed identity or verification state: %#v payload=%#v", reloaded, payload)
	}

	missingKeyEnv := mergeOperatorEnv(env)
	delete(missingKeyEnv, recovery.RecoveryMasterKeyEnv)
	missingStdout, missingStderr, missingExit := runOperatorBinary(t, operatorBin, missingKeyEnv,
		"backup", "create",
		"--source-config-file", sourceConfig.path,
	)
	requireOperatorRecoveryFailure(t, missingStdout, missingStderr, missingExit, "backup_create", 3, "recovery_key_unavailable", "secret_reference_missing")
	missingPayload := decodeOperatorRecoveryResult(t, missingStdout)
	if missingPayload.BackupSetID != nil || missingPayload.ConsistencyPointAt != nil {
		t.Fatalf("backup create missing-key failure allocated a candidate: %#v", missingPayload)
	}
}

func TestPhase10_E_10_01_CanonicalOperatorRestoreLatest(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-restore-source")
	targetDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-restore-target")
	sourceConfig := operatorExplicitConfig(t, sourceDB.DSN)
	targetConfig := operatorExplicitConfig(t, targetDB.DSN)

	adminEmail := "phase10-e-10-01-canonical-restore@example.test"
	seedOperatorUser(t, sourceDB.DSN, adminEmail, true, true)
	sourcePool := mustOpenOperatorPool(t, sourceDB.DSN)
	backupStorage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000102601")
	seedOperatorRecoveryBackupSet(t, ctx, sourcePool, sourceConfig, backupStorage, backupSetID, time.Now().UTC().Add(-time.Minute), "phase10-canonical-restore")

	operatorBin := buildOperatorBinary(t)
	sameStdout, sameStderr, sameExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore", "latest",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", sourceConfig.path,
		"--confirm-backup-set-id", backupSetID.String(),
	)
	requireOperatorRecoveryFailure(t, sameStdout, sameStderr, sameExit, "restore_latest", 3, "unsafe_restore_target", "same_database_binding")

	mismatchStdout, mismatchStderr, mismatchExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore", "latest",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", targetConfig.path,
		"--confirm-backup-set-id", uuid.MustParse("00000000-0000-0000-0000-000000102699").String(),
	)
	requireOperatorRecoveryFailure(t, mismatchStdout, mismatchStderr, mismatchExit, "restore_latest", 2, "invalid_operator_request", "confirmation_mismatch")
	requireOperatorRestoreTargetUnmutated(t, targetDB.DSN)

	stdout, stderr, exitCode := runOperatorBinaryWithTimeout(t, 30*time.Second, operatorBin, operatorRecoveryEnv(),
		"restore", "latest",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", targetConfig.path,
		"--confirm-backup-set-id", backupSetID.String(),
		"--progress", "jsonl",
	)
	if exitCode != 0 {
		t.Fatalf("canonical restore latest failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	payload := decodeOperatorRecoveryResult(t, stdout)
	requireOperatorRecoverySuccess(t, payload, "restore_latest", backupSetID.String())
	requireOperatorRecoveryArtifactKind(t, payload, "restore_operation", "cartulary.restore_operation.v1", 1)
	requireOperatorRecoveryProgress(t, stderr, payload.OperationID, []string{"preflight", "postgres_restore", "object_restore", "projection_rebuild", "invariant_check", "journal_write", "finalize"})
	requireOperatorRecoverySafeOutput(t, stdout, stderr, sourceDB.DSN, targetDB.DSN, sourceConfig.path, targetConfig.path, sourceConfig.objectRoot, targetConfig.objectRoot, operatorRecoveryMasterKey)
	requireOperatorRecoveryJournalAndAudit(t, sourceDB.DSN, payload, "restore_latest", "succeeded", sourceDB.DSN, targetDB.DSN, sourceConfig.path, targetConfig.path, sourceConfig.objectRoot, targetConfig.objectRoot, operatorRecoveryMasterKey)

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
}

func TestPhase10_E_10_01_CanonicalOperatorRestoreVerifyLatest(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-verify-latest-source")
	targetDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-verify-latest-target")
	sourceConfig := operatorExplicitConfig(t, sourceDB.DSN)
	targetConfig := operatorExplicitConfig(t, targetDB.DSN)
	seedOperatorUser(t, sourceDB.DSN, "phase10-e-10-01-canonical-verify-latest@example.test", true, true)
	sourcePool := mustOpenOperatorPool(t, sourceDB.DSN)
	backupStorage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000102701")
	seedOperatorRecoveryBackupSet(t, ctx, sourcePool, sourceConfig, backupStorage, backupSetID, time.Now().UTC().Add(-time.Minute), "phase10-canonical-verify-latest")

	operatorBin := buildOperatorBinary(t)
	missingMarkerStdout, missingMarkerStderr, missingMarkerExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "latest",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", targetConfig.path,
	)
	requireOperatorRecoveryFailure(t, missingMarkerStdout, missingMarkerStderr, missingMarkerExit, "restore_verify_latest", 3, "unsafe_restore_target", "target_marker_missing")
	requireOperatorRestoreTargetUnmutated(t, targetDB.DSN)

	writeInvalidRestoreVerificationTargetMarker(t, loadOperatorConfig(t, targetConfig.path))
	invalidMarkerStdout, invalidMarkerStderr, invalidMarkerExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "latest",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", targetConfig.path,
	)
	requireOperatorRecoveryFailure(t, invalidMarkerStdout, invalidMarkerStderr, invalidMarkerExit, "restore_verify_latest", 3, "unsafe_restore_target", "target_marker_invalid")
	requireOperatorRestoreTargetUnmutated(t, targetDB.DSN)

	writeRestoreVerificationTargetMarker(t, loadOperatorConfig(t, targetConfig.path))
	stdout, stderr, exitCode := runOperatorBinaryWithTimeout(t, 30*time.Second, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "latest",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", targetConfig.path,
		"--progress", "jsonl",
	)
	if exitCode != 0 {
		t.Fatalf("canonical restore-verify latest failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	payload := decodeOperatorRecoveryResult(t, stdout)
	requireOperatorRecoverySuccess(t, payload, "restore_verify_latest", backupSetID.String())
	requireOperatorRecoveryArtifactKind(t, payload, "restore_verification", recovery.RestoreVerificationArtifactSchemaID, 1)
	requireOperatorRecoveryProgress(t, stderr, payload.OperationID, []string{"preflight", "postgres_restore", "object_restore", "projection_rebuild", "invariant_check", "workbook_probe", "attestation_update", "journal_write", "finalize"})
	requireOperatorRecoverySafeOutput(t, stdout, stderr, sourceDB.DSN, targetDB.DSN, sourceConfig.path, targetConfig.path, sourceConfig.objectRoot, targetConfig.objectRoot, operatorRecoveryMasterKey)
	requireOperatorRecoveryJournalAndAudit(t, sourceDB.DSN, payload, "restore_verify_latest", "succeeded", sourceDB.DSN, targetDB.DSN, sourceConfig.path, targetConfig.path, sourceConfig.objectRoot, targetConfig.objectRoot, operatorRecoveryMasterKey)
	requireOperatorBackupVerificationState(t, sourceDB.DSN, backupSetID, recovery.VerificationVerified)
}

func TestPhase10_E_10_01_CanonicalOperatorRestoreVerifyDue(t *testing.T) {
	ctx := context.Background()
	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-verify-due-source")
	targetDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-canonical-verify-due-target")
	sourceConfig := operatorExplicitConfig(t, sourceDB.DSN)
	targetConfig := operatorExplicitConfig(t, targetDB.DSN)
	seedOperatorUser(t, sourceDB.DSN, "phase10-e-10-01-canonical-verify-due@example.test", true, true)
	sourcePool := mustOpenOperatorPool(t, sourceDB.DSN)
	backupStorage := newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot)
	now := time.Now().UTC()
	olderBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000102801")
	newerBackupSetID := uuid.MustParse("00000000-0000-0000-0000-000000102802")
	seedOperatorRecoveryBackupSet(t, ctx, sourcePool, sourceConfig, backupStorage, newerBackupSetID, now.Add(-time.Minute), "phase10-canonical-verify-due-newer")
	seedOperatorRecoveryBackupSet(t, ctx, sourcePool, sourceConfig, backupStorage, olderBackupSetID, now.Add(-2*time.Minute), "phase10-canonical-verify-due-older")
	writeRestoreVerificationTargetMarker(t, loadOperatorConfig(t, targetConfig.path))

	operatorBin := buildOperatorBinary(t)
	sameObjectStoreConfig := operatorConfigVariant(t, targetConfig, map[string]string{
		targetConfig.objectRoot: sourceConfig.objectRoot,
	})
	sameObjectStdout, sameObjectStderr, sameObjectExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "due",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", sameObjectStoreConfig.path,
	)
	requireOperatorRecoveryFailure(t, sameObjectStdout, sameObjectStderr, sameObjectExit, "restore_verify_due", 3, "unsafe_restore_target", "same_object_store_binding")
	requireOperatorRestoreTargetUnmutated(t, targetDB.DSN)

	stdout, stderr, exitCode := runOperatorBinaryWithTimeout(t, 45*time.Second, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "due",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", targetConfig.path,
		"--progress", "jsonl",
	)
	if exitCode != 0 {
		t.Fatalf("canonical restore-verify due failed: exit=%d stdout=%s stderr=%s", exitCode, stdout, stderr)
	}
	payload := decodeOperatorRecoveryResult(t, stdout)
	requireOperatorRecoverySuccess(t, payload, "restore_verify_due", olderBackupSetID.String())
	requireOperatorRecoveryArtifactKind(t, payload, "restore_verification", recovery.RestoreVerificationArtifactSchemaID, 2)
	requireOperatorRecoveryProgress(t, stderr, payload.OperationID, []string{
		"preflight",
		"postgres_restore", "object_restore", "projection_rebuild", "invariant_check", "workbook_probe", "attestation_update",
		"postgres_restore", "object_restore", "projection_rebuild", "invariant_check", "workbook_probe", "attestation_update",
		"journal_write", "finalize",
	})
	requireOperatorRecoverySafeOutput(t, stdout, stderr, sourceDB.DSN, targetDB.DSN, sourceConfig.path, targetConfig.path, sourceConfig.objectRoot, targetConfig.objectRoot, operatorRecoveryMasterKey)
	requireOperatorRecoveryJournalAndAudit(t, sourceDB.DSN, payload, "restore_verify_due", "succeeded", sourceDB.DSN, targetDB.DSN, sourceConfig.path, targetConfig.path, sourceConfig.objectRoot, targetConfig.objectRoot, operatorRecoveryMasterKey)
	requireOperatorBackupVerificationState(t, sourceDB.DSN, olderBackupSetID, recovery.VerificationVerified)
	requireOperatorBackupVerificationState(t, sourceDB.DSN, newerBackupSetID, recovery.VerificationVerified)

	for _, ref := range payload.ArtifactRefs {
		if ref.BackupSetID == nil || (*ref.BackupSetID != olderBackupSetID.String() && *ref.BackupSetID != newerBackupSetID.String()) {
			t.Fatalf("restore-verify due artifact ref did not identify its backup set: %#v", ref)
		}
	}

	noOpStdout, noOpStderr, noOpExit := runOperatorBinary(t, operatorBin, operatorRecoveryEnv(),
		"restore-verify", "due",
		"--source-config-file", sourceConfig.path,
		"--target-config-file", targetConfig.path,
	)
	if noOpExit != 0 {
		t.Fatalf("second restore-verify due failed: exit=%d stdout=%s stderr=%s", noOpExit, noOpStdout, noOpStderr)
	}
	if noOpStderr != "" {
		t.Fatalf("restore-verify due no-op without progress emitted stderr: %s", noOpStderr)
	}
	noOpPayload := decodeOperatorRecoveryResult(t, noOpStdout)
	if noOpPayload.Operation != "restore_verify_due" || noOpPayload.Result != "no_op" || noOpPayload.BackupSetID != nil || noOpPayload.ConsistencyPointAt != nil || len(noOpPayload.ArtifactRefs) != 0 || noOpPayload.Error != nil {
		t.Fatalf("unexpected restore-verify due no-op result: %#v", noOpPayload)
	}
	requireOperatorRecoveryJournalAndAudit(t, sourceDB.DSN, noOpPayload, "restore_verify_due", "no_op", sourceDB.DSN, targetDB.DSN, sourceConfig.path, targetConfig.path, sourceConfig.objectRoot, targetConfig.objectRoot, operatorRecoveryMasterKey)
}

func TestSupportPhaseF_SeaweedFSMigrationPreservationPassEvidence(t *testing.T) {
	fixture := newOperatorMigrationFixture(t, "pass")
	passBucket := "phase-f-migration-pass"
	fixture.seedMigrationBucket(t, passBucket)

	passBackupID := uuid.MustParse("00000000-0000-0000-0000-000000130101")
	passEnv := operatorMigrationS3Env(fixture.sourceHarness, fixture.targetHarness, passBucket)
	captureOperatorMigrationBackup(t, fixture.ctx, fixture.sourceCfg, passEnv, fixture.sourcePool, fixture.backupStorage, passBackupID, fixture.asOf.Add(-2*time.Minute), passBucket, fixture.blobIndex())
	beforeRefs := loadOperatorMigrationEvidenceRefs(t, fixture.sourceDSN)

	passArtifacts := operatorPhaseFArtifactsDir(t, "pass")
	passResult := runSeaweedFSMigrationPreservationSupport(t, fixture, passEnv, passBackupID, uuid.MustParse("00000000-0000-0000-0000-000000130111"), fixture.asOf.Add(-90*time.Second), passArtifacts)
	if passResult.BlockingFailure || passResult.Run.CurrentState != recovery.ObjectStoreMigrationStateCutoverReady {
		t.Fatalf("support migration pass did not reach cutover readiness: blocking=%v state=%s", passResult.BlockingFailure, passResult.Run.CurrentState)
	}
	requireOperatorMigrationArtifactsPass(t, passResult, true)
	afterRefs := loadOperatorMigrationEvidenceRefs(t, fixture.sourceDSN)
	if !sameStringMap(beforeRefs, afterRefs) {
		t.Fatalf("migration mutated DB evidence storage_ref values: before=%#v after=%#v", beforeRefs, afterRefs)
	}
	targetStore, err := objectstore.SetupWithEnv(fixture.ctx, loadOperatorConfig(t, fixture.targetConfig.path), passEnv)
	if err != nil {
		t.Fatalf("open pass target store: %v", err)
	}
	t.Cleanup(func() {
		_ = targetStore.Close()
	})
	for _, blob := range fixture.blobs() {
		reader, _, err := targetStore.ReadObject(fixture.ctx, blob.storageKey, objectstore.ReadOptions{})
		if err != nil {
			t.Fatalf("read migrated target object %s: %v", blob.storageKey, err)
		}
		got, err := readAllAndCloseOperatorObject(reader)
		if err != nil {
			t.Fatalf("read migrated target body %s: %v", blob.storageKey, err)
		}
		if !bytes.Equal(got, blob.body) {
			t.Fatalf("target object bytes changed for %s: got %q want %q", blob.storageKey, got, blob.body)
		}
	}
}

func TestSupportPhaseF_SeaweedFSMigrationPreservationMismatchEvidence(t *testing.T) {
	fixture := newOperatorMigrationFixture(t, "mismatch")
	mismatchBucket := "phase-f-migration-mismatch"
	fixture.seedMigrationBucket(t, mismatchBucket)
	if _, err := fixture.targetHarness.RoundTrip(fixture.ctx, mismatchBucket, fixture.firstBlob.storageKey, []byte("target-side mismatch")); err != nil {
		t.Fatalf("seed mismatch target object: %v", err)
	}

	mismatchBackupID := uuid.MustParse("00000000-0000-0000-0000-000000130102")
	mismatchEnv := operatorMigrationS3Env(fixture.sourceHarness, fixture.targetHarness, mismatchBucket)
	captureOperatorMigrationBackup(t, fixture.ctx, fixture.sourceCfg, mismatchEnv, fixture.sourcePool, fixture.backupStorage, mismatchBackupID, fixture.asOf.Add(-time.Minute), mismatchBucket, fixture.blobIndex())
	mismatchArtifacts := operatorPhaseFArtifactsDir(t, "mismatch")
	mismatchResult := runSeaweedFSMigrationPreservationSupport(t, fixture, mismatchEnv, mismatchBackupID, uuid.MustParse("00000000-0000-0000-0000-000000130112"), fixture.asOf, mismatchArtifacts)
	if !mismatchResult.BlockingFailure || mismatchResult.Run.CurrentState != recovery.ObjectStoreMigrationStateFailed {
		t.Fatalf("support migration mismatch did not fail closed: blocking=%v state=%s", mismatchResult.BlockingFailure, mismatchResult.Run.CurrentState)
	}
	requireOperatorMigrationArtifactsPass(t, mismatchResult, false)
}

type operatorRecoveryProgressRecord struct {
	SchemaID    string    `json:"schema_id"`
	OperationID string    `json:"operation_id"`
	Phase       string    `json:"phase"`
	Completed   int       `json:"completed"`
	Total       *int      `json:"total"`
	EmittedAt   time.Time `json:"emitted_at"`
}

func seedOperatorRecoveryBackupSet(t testing.TB, ctx context.Context, pool *pgxpool.Pool, cfg operatorExplicitConfigFixture, backupStorage recovery.BackupStorage, backupSetID uuid.UUID, consistencyPointAt time.Time, label string) recovery.BackupSet {
	t.Helper()

	sourceObjectStore, err := objectstore.NewFilesystemStore(cfg.objectRoot)
	if err != nil {
		t.Fatalf("open source object store: %v", err)
	}
	defer func() {
		_ = sourceObjectStore.Close()
	}()
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, pool)
	if err != nil {
		t.Fatalf("capture source postgres artifact: %v", err)
	}
	objectArtifact, err := recovery.CaptureObjectStoreSnapshotArtifact(ctx, sourceObjectStore, "")
	if err != nil {
		t.Fatalf("capture source object artifact: %v", err)
	}
	objectManifestBody, objectSummaryBody := operatorPhaseEObjectArtifacts(t, backupSetID, consistencyPointAt, label, objectArtifact, nil)
	createdAt := consistencyPointAt.Add(-time.Minute)
	backupSet, err := recovery.NewCaptureService(recovery.NewStore(pool), backupStorage).CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
		BackupSetID:                       backupSetID,
		ConsistencyPointAt:                consistencyPointAt,
		CreatedAt:                         createdAt,
		RetainedUntil:                     createdAt.Add(31 * 24 * time.Hour),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: objectArtifact, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectSummaryBody, ContentType: "application/json"},
	})
	if err != nil {
		t.Fatalf("capture recovery backup set %s: %v", backupSetID, err)
	}
	return backupSet
}

func decodeOperatorRecoveryResult(t testing.TB, stdout string) app.OperatorRecoveryResult {
	t.Helper()
	if !strings.HasSuffix(stdout, "\n") || strings.Count(stdout, "\n") != 1 {
		t.Fatalf("operator recovery stdout must be exactly one JSON line, got %q", stdout)
	}
	decoder := json.NewDecoder(strings.NewReader(stdout))
	decoder.DisallowUnknownFields()
	var payload app.OperatorRecoveryResult
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode operator recovery result: %v\nstdout=%s", err, stdout)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("operator recovery stdout contained trailing JSON content: %v\nstdout=%s", err, stdout)
	}
	if payload.SchemaID != app.OperatorRecoveryResultSchemaID {
		t.Fatalf("operator recovery schema_id got %q want %q: %#v", payload.SchemaID, app.OperatorRecoveryResultSchemaID, payload)
	}
	if _, err := uuid.Parse(payload.OperationID); err != nil {
		t.Fatalf("operator recovery operation_id is not UUID: %#v", payload)
	}
	if payload.StartedAt.IsZero() || payload.CompletedAt.IsZero() || payload.CompletedAt.Before(payload.StartedAt) {
		t.Fatalf("operator recovery result has invalid timestamps: %#v", payload)
	}
	return payload
}

func requireOperatorRecoverySuccess(t testing.TB, payload app.OperatorRecoveryResult, operation string, backupSetID string) {
	t.Helper()
	if payload.Operation != operation || payload.Result != "succeeded" || payload.Error != nil {
		t.Fatalf("unexpected operator recovery success payload: %#v", payload)
	}
	if backupSetID != "" {
		if payload.BackupSetID == nil || *payload.BackupSetID != backupSetID {
			t.Fatalf("operator recovery backup_set_id got %#v want %s", payload.BackupSetID, backupSetID)
		}
	}
	if payload.ConsistencyPointAt == nil {
		t.Fatalf("successful operator recovery result missing consistency_point_at: %#v", payload)
	}
}

func requireOperatorRecoveryFailure(t testing.TB, stdout string, stderr string, exitCode int, operation string, wantExit int, code string, reasonCode string) {
	t.Helper()
	if exitCode != wantExit {
		t.Fatalf("operator recovery failure exit got %d want %d stdout=%s stderr=%s", exitCode, wantExit, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("operator recovery failure wrote stderr; stdout=%s stderr=%s", stdout, stderr)
	}
	payload := decodeOperatorRecoveryResult(t, stdout)
	if payload.Operation != operation || payload.Result != "failed" || payload.Error == nil {
		t.Fatalf("unexpected operator recovery failure payload: %#v", payload)
	}
	if payload.Error.Code != code || payload.Error.ReasonCode != reasonCode {
		t.Fatalf("operator recovery error got %#v want code=%s reason=%s", payload.Error, code, reasonCode)
	}
}

func requireOperatorRecoveryArtifactKind(t testing.TB, payload app.OperatorRecoveryResult, kind string, schemaID string, wantCount int) {
	t.Helper()
	count := 0
	for _, ref := range payload.ArtifactRefs {
		if ref.Kind == kind && ref.SchemaID == schemaID {
			count++
			if strings.Contains(ref.RefID, "/") || strings.Contains(ref.RefID, "\\") || strings.TrimSpace(ref.RefID) == "" {
				t.Fatalf("operator recovery artifact ref is not a safe logical ref: %#v", ref)
			}
		}
	}
	if count != wantCount {
		t.Fatalf("operator recovery artifact kind count got %d want %d for kind=%s schema=%s refs=%#v", count, wantCount, kind, schemaID, payload.ArtifactRefs)
	}
}

func requireOperatorRecoveryProgress(t testing.TB, stderr string, operationID string, wantPhases []string) {
	t.Helper()
	trimmed := strings.TrimSuffix(stderr, "\n")
	if trimmed == "" {
		t.Fatal("operator recovery progress stderr was empty")
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) != len(wantPhases) {
		t.Fatalf("operator recovery progress line count got %d want %d stderr=%s", len(lines), len(wantPhases), stderr)
	}
	for index, line := range lines {
		var record operatorRecoveryProgressRecord
		decoder := json.NewDecoder(strings.NewReader(line))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&record); err != nil {
			t.Fatalf("decode operator recovery progress line %d: %v\nline=%s", index, err, line)
		}
		if record.SchemaID != app.OperatorRecoveryProgressSchemaID || record.OperationID != operationID || record.Phase != wantPhases[index] {
			t.Fatalf("unexpected operator recovery progress line %d: %#v want phase=%s operation_id=%s", index, record, wantPhases[index], operationID)
		}
		if record.Completed < 0 || (record.Total != nil && (record.Completed > *record.Total || *record.Total < 0)) || record.EmittedAt.IsZero() {
			t.Fatalf("invalid operator recovery progress counters/timestamp: %#v", record)
		}
	}
}

func requireOperatorRecoverySafeOutput(t testing.TB, stdout string, stderr string, forbidden ...string) {
	t.Helper()
	combined := stdout + "\n" + stderr
	for _, value := range forbidden {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if strings.Contains(combined, value) {
			t.Fatalf("operator recovery output exposed forbidden value %q\nstdout=%s\nstderr=%s", value, stdout, stderr)
		}
	}
}

func requireOperatorRecoveryJournalAndAudit(t testing.TB, dsn string, payload app.OperatorRecoveryResult, operation string, terminalResult string, forbidden ...string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql DB for operator recovery journal check: %v", err)
	}
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), `
SELECT
    result,
    COALESCE(backup_set_id::text, ''),
    COALESCE(error_code, ''),
    COALESCE(reason_code, ''),
    envelope_schema_id,
    encryption_mode,
    octet_length(ciphertext)::int
FROM operator_recovery_journal
WHERE operation_id = $1::uuid AND operation = $2
ORDER BY created_at ASC, operator_recovery_journal_id ASC
`, payload.OperationID, operation)
	if err != nil {
		t.Fatalf("query operator recovery journal: %v", err)
	}
	defer rows.Close()
	type journalRow struct {
		result          string
		backupSetID     string
		errorCode       string
		reasonCode      string
		envelopeSchema  string
		encryptionMode  string
		ciphertextBytes int
	}
	var journalRows []journalRow
	for rows.Next() {
		var row journalRow
		if err := rows.Scan(&row.result, &row.backupSetID, &row.errorCode, &row.reasonCode, &row.envelopeSchema, &row.encryptionMode, &row.ciphertextBytes); err != nil {
			t.Fatalf("scan operator recovery journal: %v", err)
		}
		journalRows = append(journalRows, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate operator recovery journal: %v", err)
	}
	if len(journalRows) < 2 || journalRows[0].result != "started" || journalRows[len(journalRows)-1].result != terminalResult {
		t.Fatalf("operator recovery journal rows did not include started and terminal %s rows: %#v", terminalResult, journalRows)
	}
	for _, row := range journalRows {
		if row.envelopeSchema != recovery.OperatorRecoveryJournalSchemaID || row.encryptionMode != recovery.BackupStorageEncryptionModeAESGCM || row.ciphertextBytes <= 0 {
			t.Fatalf("operator recovery journal row is not encrypted envelope evidence: %#v", row)
		}
	}
	last := journalRows[len(journalRows)-1]
	if payload.BackupSetID != nil && last.backupSetID != *payload.BackupSetID {
		t.Fatalf("operator recovery journal terminal backup_set_id got %q want %s rows=%#v", last.backupSetID, *payload.BackupSetID, journalRows)
	}

	var eventSource string
	var eventKind string
	var afterJSON string
	var reasonCode string
	var requestID string
	if err := db.QueryRowContext(context.Background(), `
SELECT event_source, event_kind, COALESCE(after_json::text, ''), COALESCE(reason_code, ''), COALESCE(request_id, '')
FROM deployment_admin_audit_events
WHERE request_id = $1
`, payload.OperationID).Scan(&eventSource, &eventKind, &afterJSON, &reasonCode, &requestID); err != nil {
		t.Fatalf("query operator recovery audit summary: %v", err)
	}
	if eventSource != "operator.recovery."+operation || eventKind != "operator_recovery_"+terminalResult || requestID != payload.OperationID {
		t.Fatalf("unexpected operator recovery audit summary: source=%s kind=%s request=%s after=%s", eventSource, eventKind, requestID, afterJSON)
	}
	if payload.Error != nil && reasonCode != payload.Error.ReasonCode {
		t.Fatalf("operator recovery audit reason got %q want %q", reasonCode, payload.Error.ReasonCode)
	}
	for _, value := range forbidden {
		if strings.TrimSpace(value) == "" {
			continue
		}
		if strings.Contains(afterJSON, value) {
			t.Fatalf("operator recovery audit summary exposed forbidden value %q: %s", value, afterJSON)
		}
	}
}

func requireOperatorBackupVerificationState(t testing.TB, dsn string, backupSetID uuid.UUID, want recovery.VerificationState) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql DB for backup verification state check: %v", err)
	}
	defer db.Close()
	var state string
	var verifiedAt sql.NullTime
	if err := db.QueryRowContext(context.Background(), `
SELECT verification_state, last_verified_restore_at
FROM backup_sets
WHERE backup_set_id = $1
`, backupSetID).Scan(&state, &verifiedAt); err != nil {
		t.Fatalf("query backup verification state: %v", err)
	}
	if state != string(want) {
		t.Fatalf("backup verification state got %s want %s", state, want)
	}
	if want == recovery.VerificationVerified && !verifiedAt.Valid {
		t.Fatalf("verified backup is missing last_verified_restore_at")
	}
}

type operatorExplicitConfigFixture struct {
	path       string
	objectRoot string
	backupRoot string
}

type operatorMigrationBlobFixture struct {
	objectBlobID uuid.UUID
	incidentID   uuid.UUID
	recordID     uuid.UUID
	storageKey   string
	storageRef   string
	body         []byte
	sha256       string
}

type operatorMigrationFixture struct {
	ctx           context.Context
	sourceHarness *s3test.Harness
	targetHarness *s3test.Harness
	sourceDSN     string
	firstBlob     operatorMigrationBlobFixture
	zeroBlob      operatorMigrationBlobFixture
	sourceConfig  operatorExplicitConfigFixture
	targetConfig  operatorExplicitConfigFixture
	sourceCfg     config.Config
	sourcePool    *pgxpool.Pool
	backupStorage recovery.BackupStorage
	asOf          time.Time
}

func newOperatorMigrationFixture(t testing.TB, name string) operatorMigrationFixture {
	t.Helper()

	ctx := context.Background()
	sourceHarness, err := s3test.StartOwnedWithLabels(ctx, map[string]string{"cartulary.fixture": "phase-f-migration-source-" + name})
	if err != nil {
		t.Fatalf("start source S3 fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = sourceHarness.Close(context.Background())
	})
	targetHarness, err := s3test.StartOwnedWithLabels(ctx, map[string]string{"cartulary.fixture": "phase-f-migration-target-" + name})
	if err != nil {
		t.Fatalf("start target S3 fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = targetHarness.Close(context.Background())
	})

	postgresHarness := pgtest.Start(t)
	sourceDB := postgresHarness.PrepareDatabaseT(t, "phase10-e-10-01-object-store-migration-"+name)
	adminEmail := "phase-f-migration-admin@example.test"
	adminID := seedOperatorUser(t, sourceDB.DSN, adminEmail, true, true)
	firstBlob := seedOperatorMigrationBlob(t, sourceDB.DSN, adminID, []byte("phase f migration object"))
	zeroBlob := seedOperatorMigrationBlob(t, sourceDB.DSN, adminID, []byte{})

	sourceConfig := operatorManagedS3Config(t, sourceDB.DSN, "migration-source")
	targetConfig := operatorManagedS3Config(t, sourceDB.DSN, "migration-target")
	sourcePool, err := pgxpool.New(ctx, sourceDB.DSN)
	if err != nil {
		t.Fatalf("open source pgx pool: %v", err)
	}
	t.Cleanup(sourcePool.Close)

	return operatorMigrationFixture{
		ctx:           ctx,
		sourceHarness: sourceHarness,
		targetHarness: targetHarness,
		sourceDSN:     sourceDB.DSN,
		firstBlob:     firstBlob,
		zeroBlob:      zeroBlob,
		sourceConfig:  sourceConfig,
		targetConfig:  targetConfig,
		sourceCfg:     loadOperatorConfig(t, sourceConfig.path),
		sourcePool:    sourcePool,
		backupStorage: newOperatorEncryptedBackupStorage(t, sourceConfig.backupRoot),
		asOf:          time.Date(2026, 6, 4, 12, 0, 0, 0, time.UTC),
	}
}

func (fixture operatorMigrationFixture) blobs() []operatorMigrationBlobFixture {
	return []operatorMigrationBlobFixture{fixture.firstBlob, fixture.zeroBlob}
}

func (fixture operatorMigrationFixture) blobIndex() map[string]uuid.UUID {
	return map[string]uuid.UUID{
		fixture.firstBlob.storageKey: fixture.firstBlob.objectBlobID,
		fixture.zeroBlob.storageKey:  fixture.zeroBlob.objectBlobID,
	}
}

func (fixture operatorMigrationFixture) seedMigrationBucket(t testing.TB, bucket string) {
	t.Helper()
	if err := fixture.sourceHarness.CreateBucket(fixture.ctx, bucket); err != nil {
		t.Fatalf("create source bucket %s: %v", bucket, err)
	}
	if err := fixture.targetHarness.CreateBucket(fixture.ctx, bucket); err != nil {
		t.Fatalf("create target bucket %s: %v", bucket, err)
	}
	for _, blob := range fixture.blobs() {
		if _, err := fixture.sourceHarness.RoundTrip(fixture.ctx, bucket, blob.storageKey, blob.body); err != nil {
			t.Fatalf("seed source object %s/%s: %v", bucket, blob.storageKey, err)
		}
	}
}

func buildOperatorBinary(t testing.TB) string {
	t.Helper()

	if bin, ok := injectedOperatorBinary(t); ok {
		return bin
	}

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

func injectedOperatorBinary(t testing.TB) (string, bool) {
	t.Helper()

	bin := strings.TrimSpace(os.Getenv("CARTULARY_OPERATOR_BIN"))
	if bin == "" {
		return "", false
	}
	if !filepath.IsAbs(bin) {
		bin = filepath.Join(repoRoot(), bin)
	}
	info, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("CARTULARY_OPERATOR_BIN is not usable: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("CARTULARY_OPERATOR_BIN must be a file, got directory: %s", bin)
	}
	if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
		t.Fatalf("CARTULARY_OPERATOR_BIN is not executable: %s", bin)
	}
	return bin, true
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

	return runOperatorBinaryWithTimeout(t, 15*time.Second, bin, env, args...)
}

func runOperatorBinaryWithTimeout(t testing.TB, timeout time.Duration, bin string, env map[string]string, args ...string) (string, string, int) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

func operatorManagedS3Config(t testing.TB, dsn string, serviceRef string) operatorExplicitConfigFixture {
	t.Helper()

	roots := configtest.SetupTempRoots(t)
	configtest.BindPostgresDSNToDatabaseRoot(t, roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"], dsn)
	contents := string(fixtures.MustRead("config", "valid.toml"))
	contents = strings.Replace(contents, `deployment_profile = "disconnected"`, `deployment_profile = "on_prem"`, 1)
	replacements := map[string]string{
		"/var/lib/cartulary/postgres":         roots.Paths["CARTULARY__ROOTS__DATABASE_STORAGE__PATH"],
		"/var/lib/cartulary/backups":          roots.Paths["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"],
		"/var/lib/cartulary/reference-packs":  roots.Paths["CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH"],
		"/var/lib/cartulary/tmp":              roots.Paths["CARTULARY__ROOTS__TEMPORARY_WORK__PATH"],
		"/var/lib/cartulary/exports":          roots.Paths["CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH"],
		"/etc/cartulary/bootstrap-admin.json": fixtures.Path("bootstrap-admin", "canonical.json"),
	}
	for oldValue, newValue := range replacements {
		contents = strings.ReplaceAll(contents, oldValue, newValue)
	}
	objectBlock := strings.Join([]string{
		"[roots.object_storage]",
		"binding_kind = \"filesystem_root\"",
		"path = \"/var/lib/cartulary/object-store\"",
	}, "\n")
	managedBlock := strings.Join([]string{
		"[roots.object_storage]",
		"binding_kind = \"managed_service\"",
		"service_ref = \"" + serviceRef + "\"",
	}, "\n")
	contents = strings.Replace(contents, objectBlock, managedBlock, 1)
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write managed S3 operator config: %v", err)
	}
	return operatorExplicitConfigFixture{
		path:       configPath,
		objectRoot: roots.Paths["CARTULARY__ROOTS__OBJECT_STORAGE__PATH"],
		backupRoot: roots.Paths["CARTULARY__ROOTS__BACKUP_STORAGE__PATH"],
	}
}

func operatorConfigVariant(t testing.TB, base operatorExplicitConfigFixture, replacements map[string]string) operatorExplicitConfigFixture {
	t.Helper()

	body, err := os.ReadFile(base.path)
	if err != nil {
		t.Fatalf("read base operator config: %v", err)
	}
	contents := string(body)
	for oldValue, newValue := range replacements {
		contents = strings.ReplaceAll(contents, oldValue, newValue)
	}
	configPath := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(configPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write operator config variant: %v", err)
	}
	return operatorExplicitConfigFixture{
		path:       configPath,
		objectRoot: base.objectRoot,
		backupRoot: base.backupRoot,
	}
}

func operatorMigrationS3Env(source *s3test.Harness, target *s3test.Harness, bucket string) map[string]string {
	env := mergeOperatorEnv(source.EnvForServiceRef("migration-source", bucket), target.EnvForServiceRef("migration-target", bucket))
	return env
}

func mergeOperatorEnv(parts ...map[string]string) map[string]string {
	result := map[string]string{}
	for _, part := range parts {
		for key, value := range part {
			result[key] = value
		}
	}
	return result
}

func seedOperatorMigrationBlob(t testing.TB, dsn string, actorID uuid.UUID, body []byte) operatorMigrationBlobFixture {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql DB for migration blob seed: %v", err)
	}
	defer db.Close()
	now := time.Now().UTC()
	incidentID := uuid.New()
	recordID := uuid.New()
	blobID := uuid.New()
	storageKey := blobref.MustObjectBlobStorageKey(incidentID, blobID)
	sha := operatorSHA256Hex(body)
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO incidents (id, incident_key, incident_key_canonical, title, status, created_by_user_id, updated_by_user_id, created_at, updated_at)
VALUES ($1, $2, $2, $3, 'active', $4, $4, $5, $5)
`, incidentID, "phase-f-"+blobID.String(), "Phase F migration fixture", actorID, now); err != nil {
		t.Fatalf("insert migration incident: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, created_at, updated_by_user_id, updated_at, row_version)
VALUES ($1, $2, 'evidence', $3, $4, $3, $4, 1)
`, recordID, incidentID, actorID, now); err != nil {
		t.Fatalf("insert migration evidence record envelope: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, observed_size, observed_content_type, observed_sha256_hex,
    target_expires_at, pending_expires_at, finalized_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'available',
    $5, $5, 'application/octet-stream', $6,
    $7, $7, $8, $8, $8
)
`, blobID, incidentID, actorID, storageKey, int64(len(body)), sha, now.Add(time.Hour), now); err != nil {
		t.Fatalf("insert migration object blob: %v", err)
	}
	storageRef := "object://" + blobID.String()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, upload_state,
    storage_ref, blob_hash, object_blob_id, requested_at, received_at, created_at, updated_at
) VALUES (
    $1, $2, 'Phase F migration evidence', 'available', 'available',
    $3, $4, $5, $6, $6, $6, $6
)
`, recordID, incidentID, storageRef, sha, blobID, now); err != nil {
		t.Fatalf("insert migration evidence row: %v", err)
	}
	return operatorMigrationBlobFixture{
		objectBlobID: blobID,
		incidentID:   incidentID,
		recordID:     recordID,
		storageKey:   storageKey,
		storageRef:   storageRef,
		body:         append([]byte(nil), body...),
		sha256:       sha,
	}
}

func captureOperatorMigrationBackup(t testing.TB, ctx context.Context, cfg config.Config, env map[string]string, pool *pgxpool.Pool, backupStorage recovery.BackupStorage, backupSetID uuid.UUID, consistencyPointAt time.Time, bucket string, blobIndex map[string]uuid.UUID) {
	t.Helper()
	sourceObjectStore, err := objectstore.SetupWithEnv(ctx, cfg, env)
	if err != nil {
		t.Fatalf("open source object store for migration backup: %v", err)
	}
	defer func() {
		_ = sourceObjectStore.Close()
	}()
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, pool)
	if err != nil {
		t.Fatalf("capture migration postgres artifact: %v", err)
	}
	objectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, sourceObjectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        consistencyPointAt,
		Bucket:                    bucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		t.Fatalf("capture migration object-store backup artifacts: %v", err)
	}
	if _, err := recovery.NewCaptureService(recovery.NewStore(pool), backupStorage).CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
		BackupSetID:                       backupSetID,
		ConsistencyPointAt:                consistencyPointAt,
		CreatedAt:                         consistencyPointAt.Add(-time.Minute),
		RetainedUntil:                     consistencyPointAt.Add(31 * 24 * time.Hour),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: objectArtifacts.SnapshotBody, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectArtifacts.SummaryBody, ContentType: "application/json"},
	}); err != nil {
		t.Fatalf("capture migration backup set: %v", err)
	}
}

const seaweedFSMigrationSupportIdentity = "local_os_execution"

type seaweedFSMigrationSupportResult struct {
	Run                  recovery.ObjectStoreMigrationRun
	CopyLedger           recovery.ObjectStoreMigrationCopyLedger
	Validation           recovery.ObjectStoreMigrationValidation
	BlockingFailure      bool
	MigrationRunArtifact migrationSupportArtifact
	CopyLedgerArtifact   migrationSupportArtifact
	ValidationArtifact   migrationSupportArtifact
	ProbeArtifact        migrationSupportArtifact
	RollbackArtifact     migrationSupportArtifact
}

type migrationSupportArtifact struct {
	path string
	ref  recovery.ObjectStoreMigrationArtifactRef
}

func runSeaweedFSMigrationPreservationSupport(t testing.TB, fixture operatorMigrationFixture, env map[string]string, backupSetID uuid.UUID, runID uuid.UUID, asOf time.Time, artifactsDir string) seaweedFSMigrationSupportResult {
	t.Helper()

	sourceObjectStore, err := objectstore.SetupWithEnv(fixture.ctx, fixture.sourceCfg, env)
	if err != nil {
		t.Fatalf("open support migration source object store: %v", err)
	}
	defer func() {
		_ = sourceObjectStore.Close()
	}()
	targetCfg := loadOperatorConfig(t, fixture.targetConfig.path)
	targetObjectStore, err := objectstore.SetupWithEnv(fixture.ctx, targetCfg, env)
	if err != nil {
		t.Fatalf("open support migration target object store: %v", err)
	}
	defer func() {
		_ = targetObjectStore.Close()
	}()

	sourceSettings, targetSettings := preflightSeaweedFSMigrationSupportConfig(t, fixture.sourceCfg, targetCfg, env)
	proof := seaweedFSMigrationSupportQuiescenceProof()
	if err := recovery.ValidateObjectStoreMigrationWriteQuiescenceProof(proof); err != nil {
		t.Fatalf("support migration write-quiescence proof invalid: %v", err)
	}
	backupSet, err := recovery.NewBackupCatalog(recovery.NewStore(fixture.sourcePool), fixture.backupStorage).RestoreCandidateBackup(fixture.ctx, asOf)
	if err != nil {
		t.Fatalf("select support migration candidate backup: %v", err)
	}
	if backupSet.BackupSetID != backupSetID {
		t.Fatalf("support migration candidate backup got %s want %s", backupSet.BackupSetID, backupSetID)
	}
	backupRef, err := recovery.LoadObjectStoreMigrationBackupRefs(fixture.ctx, fixture.backupStorage, backupSet)
	if err != nil {
		t.Fatalf("load support migration backup refs: %v", err)
	}

	now := fixture.asOf
	run, err := recovery.NewObjectStoreMigrationRun(runID, now, seaweedFSMigrationSupportIdentity, sourceSettings.Endpoint, targetSettings.Endpoint, sourceSettings.Bucket, targetSettings.Bucket)
	if err != nil {
		t.Fatalf("create support migration run: %v", err)
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventPreflightPassed, now.Add(time.Millisecond), map[string]string{
		"source_backend": recovery.ObjectStoreBackendMinIOS3,
		"target_backend": recovery.ObjectStoreBackendSeaweedFSS3,
	}); err != nil {
		t.Fatalf("record support migration preflight: %v", err)
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventWriteQuiescenceVerified, proof.CheckedAt, map[string]string{
		"proof_kind": proof.ProofKind,
	}); err != nil {
		t.Fatalf("record support migration write quiescence: %v", err)
	}
	run.BackupRefs = append(run.BackupRefs, backupRef)
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventBackupCaptured, backupSet.ConsistencyPointAt, map[string]string{
		"backup_set_id": backupSet.BackupSetID.String(),
	}); err != nil {
		t.Fatalf("record support migration backup capture: %v", err)
	}

	artifactDir, err := filepath.Abs(artifactsDir)
	if err != nil {
		t.Fatalf("resolve support migration artifacts dir: %v", err)
	}
	probe, probeBody, err := recovery.ProbeObjectStoreMigrationTarget(fixture.ctx, runID, targetSettings.Bucket, targetObjectStore, now.Add(2*time.Millisecond))
	if err != nil {
		t.Fatalf("probe support migration target: %v", err)
	}
	probeArtifact := writeSeaweedFSMigrationSupportArtifact(t, artifactDir, "target-probe.json", probeBody)
	run.ProbeRef = &probeArtifact.ref
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventTargetPrepared, probe.CompletedAt, map[string]string{
		"probe_ref": probeArtifact.path,
	}); err != nil {
		t.Fatalf("record support migration target prepared: %v", err)
	}

	_, rollbackBody, err := recovery.BuildObjectStoreMigrationRollbackEvidence(runID, now.Add(3*time.Millisecond))
	if err != nil {
		t.Fatalf("build support migration rollback evidence: %v", err)
	}
	rollbackArtifact := writeSeaweedFSMigrationSupportArtifact(t, artifactDir, "rollback-evidence.json", rollbackBody)
	run.RollbackRef = &rollbackArtifact.ref

	objects, err := recovery.ListObjectStoreMigrationBlobs(fixture.ctx, fixture.sourcePool)
	if err != nil {
		t.Fatalf("list support migration blobs: %v", err)
	}
	if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventCopyStarted, now.Add(4*time.Millisecond), nil); err != nil {
		t.Fatalf("record support migration copy started: %v", err)
	}
	copyLedger, copyLedgerBody, err := recovery.CopyObjectStoreMigrationObjects(fixture.ctx, recovery.ObjectStoreMigrationCopyParams{
		RunID:         runID,
		SourceBackend: recovery.ObjectStoreBackendMinIOS3,
		TargetBackend: recovery.ObjectStoreBackendSeaweedFSS3,
		SourceBucket:  sourceSettings.Bucket,
		TargetBucket:  targetSettings.Bucket,
		SourceStore:   sourceObjectStore,
		TargetStore:   targetObjectStore,
		Objects:       objects,
	})
	if err != nil {
		t.Fatalf("copy support migration objects: %v", err)
	}
	copyArtifact := writeSeaweedFSMigrationSupportArtifact(t, artifactDir, "copy-ledger.json", copyLedgerBody)
	run.CopyLedgerRef = &copyArtifact.ref

	validationStartedAt := now.Add(5 * time.Millisecond)
	validation, validationBody, err := recovery.ValidateObjectStoreMigration(fixture.ctx, recovery.ObjectStoreMigrationValidationParams{
		RunID:         runID,
		StartedAt:     validationStartedAt,
		CompletedAt:   validationStartedAt.Add(time.Millisecond),
		SourceBackend: recovery.ObjectStoreBackendMinIOS3,
		TargetBackend: recovery.ObjectStoreBackendSeaweedFSS3,
		SourceBucket:  sourceSettings.Bucket,
		TargetBucket:  targetSettings.Bucket,
		SourceStore:   sourceObjectStore,
		TargetStore:   targetObjectStore,
		Objects:       objects,
	})
	if err != nil {
		t.Fatalf("validate support migration objects: %v", err)
	}
	validationArtifact := writeSeaweedFSMigrationSupportArtifact(t, artifactDir, "validation.json", validationBody)
	run.ValidationRef = &validationArtifact.ref

	blockingFailure := copyLedger.Result != "pass" || validation.Result != "pass"
	if blockingFailure {
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventBlockingFailure, now.Add(7*time.Millisecond), map[string]string{
			"copy_ledger_result": copyLedger.Result,
			"validation_result":  validation.Result,
		}); err != nil {
			t.Fatalf("record support migration blocking failure: %v", err)
		}
	} else {
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventCopyCompleted, now.Add(6*time.Millisecond), map[string]string{
			"copy_ledger_ref": copyArtifact.path,
		}); err != nil {
			t.Fatalf("record support migration copy complete: %v", err)
		}
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventValidationStarted, validationStartedAt, map[string]string{
			"validation_ref": validationArtifact.path,
		}); err != nil {
			t.Fatalf("record support migration validation started: %v", err)
		}
		if err := recovery.ApplyObjectStoreMigrationEvent(&run, recovery.ObjectStoreMigrationEventValidationPassed, *validation.CompletedAt, map[string]string{
			"validation_result": validation.Result,
		}); err != nil {
			t.Fatalf("record support migration validation passed: %v", err)
		}
	}

	runBody, err := recovery.EncodeObjectStoreMigrationRun(run)
	if err != nil {
		t.Fatalf("encode support migration run: %v", err)
	}
	runArtifact := writeSeaweedFSMigrationSupportArtifact(t, artifactDir, "migration-run.json", runBody)
	return seaweedFSMigrationSupportResult{
		Run:                  run,
		CopyLedger:           copyLedger,
		Validation:           validation,
		BlockingFailure:      blockingFailure,
		MigrationRunArtifact: runArtifact,
		CopyLedgerArtifact:   copyArtifact,
		ValidationArtifact:   validationArtifact,
		ProbeArtifact:        probeArtifact,
		RollbackArtifact:     rollbackArtifact,
	}
}

func preflightSeaweedFSMigrationSupportConfig(t testing.TB, sourceCfg config.Config, targetCfg config.Config, env map[string]string) (objectstore.Settings, objectstore.Settings) {
	t.Helper()
	sourceSettings, err := objectstore.ResolveSettings(sourceCfg, env)
	if err != nil {
		t.Fatalf("resolve support migration source object-store settings: %v", err)
	}
	targetSettings, err := objectstore.ResolveSettings(targetCfg, env)
	if err != nil {
		t.Fatalf("resolve support migration target object-store settings: %v", err)
	}
	if sourceSettings.BindingKind != "managed_service" || targetSettings.BindingKind != "managed_service" {
		t.Fatalf("support migration requires managed_service object stores, got source=%s target=%s", sourceSettings.BindingKind, targetSettings.BindingKind)
	}
	if objectStoreBindingID(sourceSettings) == objectStoreBindingID(targetSettings) {
		t.Fatalf("support migration source and target object stores must differ")
	}
	if strings.TrimSpace(sourceSettings.Bucket) != strings.TrimSpace(targetSettings.Bucket) {
		t.Fatalf("support migration buckets must match, got source=%q target=%q", sourceSettings.Bucket, targetSettings.Bucket)
	}
	return sourceSettings, targetSettings
}

func objectStoreBindingID(settings objectstore.Settings) string {
	switch settings.BindingKind {
	case "filesystem_root":
		return "filesystem_root:" + filepath.Clean(settings.RootPath)
	case "managed_service":
		return fmt.Sprintf("managed_service:%s:%t:%s", settings.Endpoint, settings.Secure, settings.Bucket)
	default:
		return settings.BindingKind
	}
}

func seaweedFSMigrationSupportQuiescenceProof() recovery.ObjectStoreMigrationWriteQuiescenceProof {
	return recovery.ObjectStoreMigrationWriteQuiescenceProof{
		SchemaID:                recovery.ObjectStoreMigrationProofSchemaID,
		ProofKind:               "process_stopped",
		CheckedAt:               time.Date(2026, 6, 4, 11, 59, 0, 0, time.UTC),
		ProcessState:            "absent",
		HTTPListenerClosed:      true,
		WebSocketListenerClosed: true,
	}
}

func writeSeaweedFSMigrationSupportArtifact(t testing.TB, dir string, name string, body []byte) migrationSupportArtifact {
	t.Helper()
	if len(body) == 0 {
		t.Fatalf("write support migration artifact %s: body is empty", name)
	}
	if strings.TrimSpace(dir) == "" {
		t.Fatalf("write support migration artifact %s: artifacts dir is required", name)
	}
	if strings.TrimSpace(name) == "" || filepath.Base(name) != name {
		t.Fatalf("write support migration artifact: invalid file name %q", name)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create support migration artifact dir: %v", err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write support migration artifact %s: %v", path, err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("resolve support migration artifact path: %v", err)
	}
	return migrationSupportArtifact{
		path: abs,
		ref:  recovery.ArtifactRefForBody(abs, body, "application/json"),
	}
}

func mustOpenOperatorPool(t testing.TB, dsn string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open operator pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func operatorPhaseFArtifactsDir(t testing.TB, name string) string {
	t.Helper()
	root := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RESULTS_DIR"))
	if root == "" {
		root = filepath.Join(t.TempDir(), "test-results")
	}
	if runID := strings.TrimSpace(os.Getenv("CARTULARY_TEST_RUN_ID")); runID != "" && filepath.Base(root) != runID {
		root = filepath.Join(root, runID)
	}
	dir := filepath.Join(root, "backend-process", "phase-f-object-store-migration", name)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create phase f artifacts dir: %v", err)
	}
	return dir
}

func requireOperatorMigrationArtifactsPass(t testing.TB, result seaweedFSMigrationSupportResult, wantPass bool) {
	t.Helper()
	ledger := readOperatorMigrationCopyLedger(t, result.CopyLedgerArtifact.path)
	validation := readOperatorMigrationValidation(t, result.ValidationArtifact.path)
	run := readOperatorMigrationRun(t, result.MigrationRunArtifact.path)
	if run.OperatorIdentity != seaweedFSMigrationSupportIdentity {
		t.Fatalf("object-store migration used unexpected operator identity: %q", run.OperatorIdentity)
	}
	if wantPass {
		if ledger.Result != "pass" || validation.Result != "pass" || run.CurrentState != recovery.ObjectStoreMigrationStateCutoverReady {
			t.Fatalf("pass artifacts did not prove cutover readiness: ledger=%s validation=%s run=%s", ledger.Result, validation.Result, run.CurrentState)
		}
		for _, item := range ledger.Items {
			if item.Status != recovery.ObjectStoreMigrationCopyCopied && item.Status != recovery.ObjectStoreMigrationCopyAlreadyCopied {
				t.Fatalf("pass ledger contains blocking status: %#v", item)
			}
		}
		return
	}
	if ledger.Result != "fail" || validation.Result != "fail" || run.CurrentState != recovery.ObjectStoreMigrationStateFailed {
		t.Fatalf("mismatch artifacts did not prove blocking failure: ledger=%s validation=%s run=%s", ledger.Result, validation.Result, run.CurrentState)
	}
	foundMismatch := false
	for _, item := range ledger.Items {
		if item.Status == recovery.ObjectStoreMigrationCopyTargetMismatch {
			foundMismatch = true
		}
	}
	if !foundMismatch {
		t.Fatalf("mismatch ledger did not contain target_mismatch: %#v", ledger.Items)
	}
}

func readOperatorMigrationRun(t testing.TB, path string) recovery.ObjectStoreMigrationRun {
	t.Helper()
	body := readOperatorMigrationArtifactBody(t, path)
	artifact, err := recovery.DecodeObjectStoreMigrationRun(body)
	if err != nil {
		t.Fatalf("decode migration run artifact: %v\nbody=%s", err, body)
	}
	return artifact
}

func readOperatorMigrationCopyLedger(t testing.TB, path string) recovery.ObjectStoreMigrationCopyLedger {
	t.Helper()
	body := readOperatorMigrationArtifactBody(t, path)
	artifact, err := recovery.DecodeObjectStoreMigrationCopyLedger(body)
	if err != nil {
		t.Fatalf("decode migration copy ledger artifact: %v\nbody=%s", err, body)
	}
	return artifact
}

func readOperatorMigrationValidation(t testing.TB, path string) recovery.ObjectStoreMigrationValidation {
	t.Helper()
	body := readOperatorMigrationArtifactBody(t, path)
	artifact, err := recovery.DecodeObjectStoreMigrationValidation(body)
	if err != nil {
		t.Fatalf("decode migration validation artifact: %v\nbody=%s", err, body)
	}
	return artifact
}

func readOperatorMigrationArtifactBody(t testing.TB, path string) []byte {
	t.Helper()
	if strings.TrimSpace(path) == "" {
		t.Fatalf("migration artifact path is empty")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration artifact %s: %v", path, err)
	}
	return body
}

func loadOperatorMigrationEvidenceRefs(t testing.TB, dsn string) map[string]string {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql DB for migration refs: %v", err)
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), `SELECT record_id::text, COALESCE(storage_ref, '') FROM evidence ORDER BY record_id::text`)
	if err != nil {
		t.Fatalf("query migration evidence refs: %v", err)
	}
	defer rows.Close()
	refs := map[string]string{}
	for rows.Next() {
		var recordID string
		var storageRef string
		if err := rows.Scan(&recordID, &storageRef); err != nil {
			t.Fatalf("scan migration evidence ref: %v", err)
		}
		refs[recordID] = storageRef
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate migration evidence refs: %v", err)
	}
	return refs
}

func sameStringMap(left map[string]string, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func operatorSHA256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func readAllAndCloseOperatorObject(reader io.ReadCloser) ([]byte, error) {
	body, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	return body, closeErr
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
	markerPath, err := operatorops.RestoreVerificationTargetMarkerPath(cfg)
	if err != nil {
		t.Fatalf("resolve restore verification target marker path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(markerPath), 0o700); err != nil {
		t.Fatalf("create restore verification target marker directory: %v", err)
	}
	body := []byte(`{"schema_id":"cartulary.restore_verification_target.v1","purpose":"restore_verification_target"}`)
	if err := os.WriteFile(markerPath, body, 0o600); err != nil {
		t.Fatalf("write restore verification target marker: %v", err)
	}
}

func writeInvalidRestoreVerificationTargetMarker(t testing.TB, cfg config.Config) {
	t.Helper()
	markerPath, err := operatorops.RestoreVerificationTargetMarkerPath(cfg)
	if err != nil {
		t.Fatalf("resolve invalid restore verification target marker path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(markerPath), 0o700); err != nil {
		t.Fatalf("create invalid restore verification target marker directory: %v", err)
	}
	body := []byte(`{"schema_id":"cartulary.restore_verification_target.v1","purpose":"production_target"}`)
	if err := os.WriteFile(markerPath, body, 0o600); err != nil {
		t.Fatalf("write invalid restore verification target marker: %v", err)
	}
}

func requireOperatorRestoreTargetUnmutated(t testing.TB, dsn string) {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open target sql for mutation check: %v", err)
	}
	defer db.Close()
	var rowCount int
	if err := db.QueryRowContext(context.Background(), `
SELECT
    (SELECT COUNT(*) FROM incidents)
  + (SELECT COUNT(*) FROM records)
  + (SELECT COUNT(*) FROM object_blobs)
`).Scan(&rowCount); err != nil {
		t.Fatalf("query target mutation count: %v", err)
	}
	if rowCount != 0 {
		t.Fatalf("unsafe restore-verification preflight mutated target rows, count=%d", rowCount)
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
