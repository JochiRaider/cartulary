package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/operatorcli"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

const (
	OperatorRecoveryResultSchemaID   = operatorcli.ResultSchemaID
	OperatorRecoveryProgressSchemaID = operatorcli.ProgressSchemaID
)

type OperatorRecoveryResult = operatorcli.Result
type OperatorRecoveryArtifactRef = operatorcli.ArtifactRef
type OperatorRecoveryError = operatorcli.Error

func (runner operatorRunner) runRecoveryCLI(ctx context.Context, args []string) (bool, int) {
	return operatorcli.Runner{
		Stdout:     runner.stdout,
		Stderr:     runner.stderr,
		Now:        runner.now,
		Operations: operatorRecoveryOperations{runner: runner},
	}.Run(ctx, args)
}

type operatorRecoveryOperations struct {
	runner operatorRunner
}

func (operations operatorRecoveryOperations) BackupInspectLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (operatorcli.Outcome, error) {
	return operations.runner.runRecoveryBackupInspectLatest(ctx, parsed, progress)
}

func (operations operatorRecoveryOperations) BackupCreate(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (operatorcli.Outcome, error) {
	return operations.runner.runRecoveryBackupCreate(ctx, parsed, progress)
}

func (operations operatorRecoveryOperations) RestoreLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (operatorcli.Outcome, error) {
	return operations.runner.runRecoveryRestoreLatest(ctx, parsed, progress)
}

func (operations operatorRecoveryOperations) RestoreVerifyLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (operatorcli.Outcome, error) {
	return operations.runner.runRecoveryRestoreVerifyLatest(ctx, parsed, progress)
}

func (operations operatorRecoveryOperations) RestoreVerifyDue(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (operatorcli.Outcome, error) {
	return operations.runner.runRecoveryRestoreVerifyDue(ctx, parsed, progress)
}

func (runner operatorRunner) runRecoveryBackupInspectLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (operatorcli.Outcome, error) {
	progress.Emit("preflight", 0, nil)
	cfg, pool, backupStorage, closeFn, err := runner.openRecoverySourceRuntime(ctx, parsed.SourceConfigPath)
	_ = cfg
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer closeFn()
	progress.Emit("catalog_select", 0, nil)
	asOf := runner.now().UTC()
	selection, err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage).RestoreCandidateBackupSelection(ctx, asOf)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	progress.Emit("artifact_validate", 0, nil)
	backupSet := selection.BackupSet
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return operatorcli.OutcomeForBackupSet(backupSet, "backup_attestation", "cartulary.backup_attestation.v1"), nil
}

func (runner operatorRunner) runRecoveryBackupCreate(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	cfg, pool, backupStorage, closeFn, err := runner.openRecoverySourceRuntime(ctx, parsed.SourceConfigPath)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer closeFn()
	unlock, err := runner.acquireRecoveryOperationLock(ctx, pool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := runner.recordRecoveryStart(ctx, pool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { runner.finishRecoveryJournalAndAudit(ctx, pool, parsed, outcome, &err) }()

	consistencyPointAt := runner.now().UTC()
	backupSetID := uuid.New()
	progress.Emit("postgres_backup", 0, nil)
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, pool)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("capture postgres artifact: %w", err)
	}
	blobIndex, err := loadBackupCaptureObjectBlobIndex(ctx, pool)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), err
	}
	objectBucket, err := backupCaptureObjectStoreBucket(cfg)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), err
	}
	objectStore, err := runner.setupObjectStore(ctx, cfg)
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("open object store: %w", err)
	}
	defer func() { _ = objectStore.Close() }()
	progress.Emit("object_backup", 0, nil)
	objectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, objectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        consistencyPointAt,
		Bucket:                    objectBucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("capture object-store artifacts: %w", err)
	}
	progress.Emit("attestation_write", 0, nil)
	backupSet, err := recovery.NewCaptureService(recovery.NewStore(pool), backupStorage).CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
		BackupSetID:                       backupSetID,
		ConsistencyPointAt:                consistencyPointAt,
		CreatedAt:                         consistencyPointAt,
		RetainedUntil:                     consistencyPointAt.Add(recovery.MinimumRetentionDuration),
		PostgresArtifact:                  recovery.BackupArtifact{Body: postgresArtifact, ContentType: "application/json"},
		ObjectStoreArtifact:               recovery.BackupArtifact{Body: objectArtifacts.SnapshotBody, ContentType: "application/json"},
		ObjectStoreBackupManifestArtifact: recovery.BackupArtifact{Body: objectArtifacts.ManifestBody, ContentType: "application/json"},
		ObjectStoreBackupSummaryArtifact:  recovery.BackupArtifact{Body: objectArtifacts.SummaryBody, ContentType: "application/json"},
	})
	if err != nil {
		return operatorcli.OutcomeForCandidate(backupSetID, consistencyPointAt), fmt.Errorf("capture backup set: %w", err)
	}
	if err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage).VerifyBackupSetDurability(ctx, backupSet); err != nil {
		return operatorcli.OutcomeForBackupSet(backupSet, "backup_attestation", "cartulary.backup_attestation.v1"), fmt.Errorf("verify captured backup durability: %w", err)
	}
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return operatorcli.OutcomeForBackupSet(backupSet, "backup_attestation", "cartulary.backup_attestation.v1"), nil
}

func (runner operatorRunner) runRecoveryRestoreLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openRecoveryRestoreRuntime(ctx, parsed)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	unlock, err := runner.acquireRecoveryOperationLock(ctx, sourcePool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := runner.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { runner.finishRecoveryJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()

	asOf := runner.now().UTC()
	sourceStore := recovery.NewStore(sourcePool)
	backupSet, err := recovery.NewBackupCatalog(sourceStore, backupStorage).RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	confirmed, _ := uuid.Parse(parsed.ConfirmBackupSetID)
	if backupSet.BackupSetID != confirmed {
		return operatorcli.OutcomeForStoredBackupSet(backupSet), fmt.Errorf("confirmed backup_set_id does not match latest retained backup: %w", operatorcli.ErrConfirmationMismatch)
	}
	if err := runner.preflightRestoreTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return operatorcli.OutcomeForStoredBackupSet(backupSet), err
	}
	progress.Emit("postgres_restore", 0, nil)
	progress.Emit("object_restore", 0, nil)
	progress.Emit("projection_rebuild", 0, nil)
	progress.Emit("invariant_check", 0, nil)
	result, err := recovery.NewRestoreRunner(sourceStore, backupStorage).RestoreBackupSet(ctx, recovery.RestoreTarget{
		Postgres:    targetPool,
		ObjectStore: targetObjectStore,
		Projections: projectionadapters.NewRestoreRebuilder(targetPool),
	}, backupSet)
	if err != nil {
		return operatorcli.OutcomeForStoredBackupSet(backupSet), err
	}
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return operatorcli.OutcomeForBackupSet(result.BackupSet, "restore_operation", "cartulary.restore_operation.v1"), nil
}

func (runner operatorRunner) runRecoveryRestoreVerifyLatest(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openRecoveryRestoreRuntime(ctx, parsed)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	unlock, err := runner.acquireRecoveryOperationLock(ctx, sourcePool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := runner.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { runner.finishRecoveryJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()
	if err := runner.preflightRestoreVerificationTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return operatorcli.Outcome{}, err
	}
	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	progress.Emit("postgres_restore", 0, nil)
	progress.Emit("object_restore", 0, nil)
	progress.Emit("projection_rebuild", 0, nil)
	progress.Emit("invariant_check", 0, nil)
	progress.Emit("workbook_probe", 0, nil)
	sourceStore := recovery.NewStore(sourcePool)
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage))
	result, err := verify.VerifyLatestSuccessfulRetained(ctx, recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			Postgres:    targetPool,
			ObjectStore: targetObjectStore,
			Projections: projectionadapters.NewRestoreRebuilder(targetPool),
		},
		Probe: recovery.RestoreVerificationWorkbookProbe{Postgres: targetPool},
	}, runner.now().UTC(), basis)
	outcome = operatorcli.OutcomeForStoredBackupSet(result.BackupSet)
	if result.Run.RestoreVerificationRunID != uuid.Nil {
		outcome.ArtifactRefs = append(outcome.ArtifactRefs, operatorcli.ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+result.Run.RestoreVerificationRunID.String(), outcome.BackupSetID))
	}
	if err != nil {
		return outcome, err
	}
	progress.Emit("attestation_update", 0, nil)
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", 1, operatorcli.IntPtr(1))
	return outcome, nil
}

func (runner operatorRunner) runRecoveryRestoreVerifyDue(ctx context.Context, parsed operatorcli.Command, progress operatorcli.ProgressEmitter) (outcome operatorcli.Outcome, err error) {
	progress.Emit("preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openRecoveryRestoreRuntime(ctx, parsed)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	unlock, err := runner.acquireRecoveryOperationLock(ctx, sourcePool)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	defer unlock()
	if err := runner.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return operatorcli.Outcome{}, err
	}
	defer func() { runner.finishRecoveryJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()
	if err := runner.preflightRestoreVerificationTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return operatorcli.Outcome{}, err
	}
	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	sourceStore := recovery.NewStore(sourcePool)
	due, err := recovery.NewBackupCatalog(sourceStore, backupStorage).ListBackupsDueForRestoreVerification(ctx, runner.now().UTC(), basis)
	if err != nil {
		return operatorcli.Outcome{}, err
	}
	if len(due) == 0 {
		return operatorcli.Outcome{ArtifactRefs: []operatorcli.ArtifactRef{}, Result: "no_op"}, nil
	}
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage))
	for _, backupSet := range due {
		progress.Emit("postgres_restore", 0, nil)
		progress.Emit("object_restore", 0, nil)
		progress.Emit("projection_rebuild", 0, nil)
		progress.Emit("invariant_check", 0, nil)
		progress.Emit("workbook_probe", 0, nil)
		result, verifyErr := verify.VerifyBackupSet(ctx, recovery.RestoreVerificationTarget{
			RestoreTarget: recovery.RestoreTarget{
				Postgres:    targetPool,
				ObjectStore: targetObjectStore,
				Projections: projectionadapters.NewRestoreRebuilder(targetPool),
			},
			Probe: recovery.RestoreVerificationWorkbookProbe{Postgres: targetPool},
		}, backupSet, basis)
		if outcome.BackupSetID == nil {
			outcome = operatorcli.OutcomeForStoredBackupSet(result.BackupSet)
		}
		if result.Run.RestoreVerificationRunID != uuid.Nil {
			outcome.ArtifactRefs = append(outcome.ArtifactRefs, operatorcli.ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+result.Run.RestoreVerificationRunID.String(), outcome.BackupSetID))
		}
		if verifyErr != nil {
			return outcome, verifyErr
		}
		progress.Emit("attestation_update", 0, nil)
	}
	progress.Emit("journal_write", 0, nil)
	progress.Emit("finalize", len(due), operatorcli.IntPtr(len(due)))
	return outcome, nil
}

func (runner operatorRunner) openRecoverySourceRuntime(ctx context.Context, sourceConfigPath string) (config.Config, operatorPostgresPool, recovery.BackupStorage, func(), error) {
	cfg, err := runner.loadConfig(sourceConfigPath)
	if err != nil {
		return config.Config{}, nil, nil, nil, fmt.Errorf("load config: %w", err)
	}
	pool, err := runner.setupPostgres(ctx, cfg)
	if err != nil {
		return config.Config{}, nil, nil, nil, fmt.Errorf("open postgres: %w", err)
	}
	backupStorage, err := runner.newBackupStorage(cfg)
	if err != nil {
		pool.Close()
		return config.Config{}, nil, nil, nil, fmt.Errorf("open backup storage: %w", err)
	}
	return cfg, pool, backupStorage, func() { pool.Close() }, nil
}

func (runner operatorRunner) openRecoveryRestoreRuntime(ctx context.Context, parsed operatorcli.Command) (config.Config, config.Config, operatorPostgresPool, operatorPostgresPool, objectstore.Store, objectstore.Store, recovery.BackupStorage, error) {
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := runner.openRestoreRuntime(ctx, operatorCLIResult{
		sourceConfigPath: parsed.SourceConfigPath,
		targetConfigPath: parsed.TargetConfigPath,
	})
	return sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err
}

func (runner operatorRunner) acquireRecoveryOperationLock(ctx context.Context, pool operatorPostgresPool) (func(), error) {
	locked, err := tryRestoreVerificationAdvisoryLock(ctx, pool)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, operatorcli.ErrOperationLockUnavailable
	}
	return func() { _ = unlockRestoreVerificationAdvisoryLock(context.Background(), pool) }, nil
}

func (runner operatorRunner) recoveryJournalStore(pool operatorPostgresPool) operatorcli.JournalStore {
	return operatorcli.JournalStore{
		DB: pool,
		LoadKey: func() (recovery.RecoveryEncryptionKey, error) {
			return recovery.LoadRecoveryEncryptionKey(nil)
		},
		Now: runner.now,
	}
}

func (runner operatorRunner) recordRecoveryStart(ctx context.Context, pool operatorPostgresPool, parsed operatorcli.Command) error {
	return runner.recoveryJournalStore(pool).Append(ctx, operatorcli.JournalRecord{
		OperationID: parsed.OperationID,
		Operation:   parsed.Operation,
		Result:      "started",
		Summary: map[string]any{
			"source_config_supplied": parsed.SourceConfigPath != "",
			"target_config_supplied": parsed.TargetConfigPath != "",
		},
	})
}

func (runner operatorRunner) finishRecoveryJournalAndAudit(ctx context.Context, pool operatorPostgresPool, parsed operatorcli.Command, outcome operatorcli.Outcome, operationErr *error) {
	result := outcome.Result
	if result == "" {
		result = "succeeded"
	}
	var errorCode string
	var reasonCode string
	if operationErr != nil && *operationErr != nil {
		result = "failed"
		mapped, _ := operatorcli.MapError(parsed.Operation, *operationErr)
		errorCode = mapped.Code
		reasonCode = mapped.ReasonCode
	}
	record := operatorcli.JournalRecord{
		OperationID: parsed.OperationID,
		Operation:   parsed.Operation,
		Result:      result,
		BackupSetID: outcome.BackupSetID,
		ErrorCode:   errorCode,
		ReasonCode:  reasonCode,
		Summary: map[string]any{
			"artifact_ref_count": len(outcome.ArtifactRefs),
			"has_backup_set_id":  outcome.BackupSetID != nil,
		},
	}
	store := runner.recoveryJournalStore(pool)
	if err := store.Append(ctx, record); err != nil {
		mergeRecoveryOperationError(operationErr, err)
		return
	}
	if err := store.AppendAuditSummary(ctx, record); err != nil {
		mergeRecoveryOperationError(operationErr, err)
	}
}

func mergeRecoveryOperationError(operationErr *error, err error) {
	if operationErr == nil || err == nil {
		return
	}
	if *operationErr == nil {
		*operationErr = err
		return
	}
	*operationErr = fmt.Errorf("%w; additionally %v", *operationErr, err)
}
