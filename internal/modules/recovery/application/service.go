package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	restoreMinimumSchemaVersion       int64 = 22
	recoveryOperationAdvisoryLockKey  int64 = 401010
	restoreVerificationTargetSchemaID       = "cartulary.restore_verification_target.v1"

	RestoreVerificationTargetMarkerMaximumBytes int64 = 65536
)

var ErrTargetMarkerRequiresFilesystemStorage = errors.New("restore verification target marker requires filesystem backup storage")

type PostgresPool interface {
	postgres.DB
	Close()
}

type RootBinding struct {
	BindingKind string
	Path        string
	ServiceRef  string
}

// Deployment is the recovery-owned projection of an admitted deployment.
// Resource factories are application-owned closures so recovery never receives
// or interprets the complete deployment configuration.
type Deployment struct {
	DatabaseStorage  RootBinding
	ObjectStorage    RootBinding
	BackupStorage    RootBinding
	PostgresSettings postgres.Settings
	ObjectSettings   objectstore.Settings
	OpenPostgres     func(context.Context) (PostgresPool, error)
	OpenObjectStore  func(context.Context) (objectstore.Store, error)
	OpenBackup       func() (recovery.BackupStorage, error)
}

type DeploymentLoader func(string) (Deployment, error)
type TargetMarkerReader func(bindingKind string, rootPath string) ([]byte, error)
type ProjectionServicesFactory func(postgres.DB) (restorecontract.ProjectionRebuilder, recovery.WorkbookProjectionQuery)
type JournalKeyLoader func() (recovery.RecoveryEncryptionKey, error)
type FailureEvidenceProjector func(FailureKind) (code string, reasonCode string)

type Service struct {
	LoadDeployment         DeploymentLoader
	ReadTargetMarker       TargetMarkerReader
	NewProjectionServices  ProjectionServicesFactory
	LoadJournalKey         JournalKeyLoader
	ProjectFailureEvidence FailureEvidenceProjector
	ExtensionBackups       *recovery.ExtensionBackupCatalog
	Now                    func() time.Time
}

var _ Facade = Service{}

type operationRequest struct {
	OperationID        uuid.UUID
	Operation          Operation
	SourceConfigPath   string
	TargetConfigPath   string
	ConfirmedBackupSet uuid.UUID
	AttemptTimeout     time.Duration
}

func (service Service) BackupInspectLatest(ctx context.Context, request BackupInspectLatestRequest, progress ProgressSink) (Result, error) {
	result, err := service.backupInspectLatest(ctx, operationRequest{
		OperationID:      request.OperationID,
		Operation:        OperationBackupInspectLatest,
		SourceConfigPath: request.SourceConfigPath,
	}, progress)
	return result, EnsureFailure(FailureArtifactMissing, err)
}

func (service Service) BackupCreate(ctx context.Context, request BackupCreateRequest, progress ProgressSink) (Result, error) {
	result, err := service.backupCreate(ctx, operationRequest{
		OperationID:      request.OperationID,
		Operation:        OperationBackupCreate,
		SourceConfigPath: request.SourceConfigPath,
	}, progress)
	return result, EnsureFailure(FailureBackupPublication, err)
}

func (service Service) RestoreLatest(ctx context.Context, request RestoreLatestRequest, progress ProgressSink) (Result, error) {
	result, err := service.runRestoreLatest(ctx, operationRequest{
		OperationID:        request.OperationID,
		Operation:          OperationRestoreLatest,
		SourceConfigPath:   request.SourceConfigPath,
		TargetConfigPath:   request.TargetConfigPath,
		ConfirmedBackupSet: request.ConfirmedBackupSet,
	}, progress)
	return result, EnsureFailure(FailureRestoreInvariantCheck, err)
}

func (service Service) RestoreVerifyLatest(ctx context.Context, request RestoreVerifyLatestRequest, progress ProgressSink) (Result, error) {
	result, err := service.runRestoreVerifyLatest(ctx, operationRequest{
		OperationID:      request.OperationID,
		Operation:        OperationRestoreVerifyLatest,
		SourceConfigPath: request.SourceConfigPath,
		TargetConfigPath: request.TargetConfigPath,
	}, progress)
	return result, EnsureFailure(FailureVerificationInvariantCheck, err)
}

func (service Service) RestoreVerifyDue(ctx context.Context, request RestoreVerifyDueRequest, progress ProgressSink) (Result, error) {
	result, err := service.runRestoreVerifyDue(ctx, operationRequest{
		OperationID:      request.OperationID,
		Operation:        OperationRestoreVerifyDue,
		SourceConfigPath: request.SourceConfigPath,
		TargetConfigPath: request.TargetConfigPath,
		AttemptTimeout:   request.AttemptTimeout,
	}, progress)
	return result, EnsureFailure(FailureVerificationInvariantCheck, err)
}

func (service Service) backupInspectLatest(ctx context.Context, parsed operationRequest, progress ProgressSink) (Result, error) {
	ReportProgress(progress, "preflight", 0, nil)
	_, pool, backupStorage, closeFn, err := service.openSourceRuntime(ctx, parsed.SourceConfigPath)
	if err != nil {
		return Result{}, err
	}
	defer closeFn()
	ReportProgress(progress, "catalog_select", 0, nil)
	selection, err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage, service.ExtensionBackups).RestoreCandidateBackupSelection(ctx, service.now())
	if err != nil {
		return Result{}, classifyAdmissionFailure(err)
	}
	ReportProgress(progress, "artifact_validate", 0, nil)
	ReportProgress(progress, "finalize", 1, IntPtr(1))
	return ResultForBackupSet(selection.BackupSet, "backup_attestation", "cartulary.backup_attestation.v1"), nil
}

func (service Service) backupCreate(ctx context.Context, parsed operationRequest, progress ProgressSink) (outcome Result, err error) {
	ReportProgress(progress, "preflight", 0, nil)
	cfg, pool, backupStorage, closeFn, err := service.openSourceRuntime(ctx, parsed.SourceConfigPath)
	if err != nil {
		return Result{}, err
	}
	defer closeFn()
	unlock, err := service.acquireOperationLock(ctx, pool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, pool, parsed); err != nil {
		return Result{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, pool, parsed, outcome, &err) }()

	consistencyPointAt := service.now()
	backupSetID := uuid.New()
	ReportProgress(progress, "postgres_backup", 0, nil)
	postgresArtifact, err := recovery.CapturePostgresSnapshotArtifact(ctx, pool)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPostgres, fmt.Errorf("capture postgres artifact: %w", err))
	}
	blobIndex, err := loadBackupObjectBlobIndex(ctx, pool)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPostgres, err)
	}
	objectBucket, err := backupObjectStoreBucket(cfg)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupObject, err)
	}
	objectStore, err := service.setupObjectStore(ctx, cfg)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupObject, fmt.Errorf("open object store: %w", err))
	}
	defer func() { _ = objectStore.Close() }()
	ReportProgress(progress, "object_backup", 0, nil)
	objectArtifacts, err := recovery.CaptureSeaweedFSS3ObjectStoreBackupArtifacts(ctx, objectStore, recovery.ObjectStoreBackupCaptureParams{
		BackupSetID:               backupSetID,
		ConsistencyPointAt:        consistencyPointAt,
		Bucket:                    objectBucket,
		BlobObjectIDsByStorageRef: blobIndex,
	})
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupObject, fmt.Errorf("capture object-store artifacts: %w", err))
	}
	ReportProgress(progress, "attestation_write", 0, nil)
	backupSet, err := recovery.NewCaptureService(recovery.NewStore(pool), backupStorage, service.ExtensionBackups).CaptureBackupSet(ctx, recovery.CaptureBackupSetParams{
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
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, fmt.Errorf("capture backup set: %w", err))
	}
	if err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage, service.ExtensionBackups).VerifyBackupSetDurability(ctx, backupSet); err != nil {
		return ResultForBackupSet(backupSet, "backup_attestation", "cartulary.backup_attestation.v1"), NewFailure(FailureBackupArtifactReadback, fmt.Errorf("verify captured backup durability: %w", err))
	}
	ReportProgress(progress, "journal_write", 0, nil)
	ReportProgress(progress, "finalize", 1, IntPtr(1))
	return ResultForBackupSet(backupSet, "backup_attestation", "cartulary.backup_attestation.v1"), nil
}

func (service Service) runRestoreLatest(ctx context.Context, parsed operationRequest, progress ProgressSink) (outcome Result, err error) {
	ReportProgress(progress, "preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := service.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return Result{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	defer func() { _ = recovery.CloseBackupStorage(backupStorage) }()
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return Result{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()

	sourceStore := recovery.NewStore(sourcePool)
	backupSet, err := recovery.NewBackupCatalog(sourceStore, backupStorage, service.ExtensionBackups).RestoreCandidateBackup(ctx, service.now())
	if err != nil {
		return Result{}, classifyAdmissionFailure(err)
	}
	if backupSet.BackupSetID != parsed.ConfirmedBackupSet {
		return ResultForStoredBackupSet(backupSet), NewFailure(
			FailureConfirmationMismatch,
			errors.New("confirmed backup_set_id does not match latest retained backup"),
		)
	}
	if err := service.preflightRestoreTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	target, err := service.restoreTarget(targetPool, targetObjectStore)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureRestoreProjectionRebuild, err)
	}
	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	result, err := recovery.NewRestoreRunner(sourceStore, backupStorage, service.ExtensionBackups).RestoreBackupSet(ctx, target, backupSet)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), classifyRestoreFailure(err, false)
	}
	ReportProgress(progress, "journal_write", 0, nil)
	ReportProgress(progress, "finalize", 1, IntPtr(1))
	return ResultForBackupSet(result.BackupSet, "restore_operation", "cartulary.restore_operation.v1"), nil
}

func (service Service) runRestoreVerifyLatest(ctx context.Context, parsed operationRequest, progress ProgressSink) (outcome Result, err error) {
	ReportProgress(progress, "preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := service.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return Result{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	defer func() { _ = recovery.CloseBackupStorage(backupStorage) }()
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return Result{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()
	if err := service.preflightRestoreVerificationTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return Result{}, err
	}
	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, err)
	}
	target, err := service.restoreVerificationTarget(targetPool, targetObjectStore)
	if err != nil {
		return Result{}, NewFailure(FailureVerificationProjectionRebuild, err)
	}
	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	ReportProgress(progress, "workbook_probe", 0, nil)
	sourceStore := recovery.NewStore(sourcePool)
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage, service.ExtensionBackups))
	result, err := verify.VerifyLatestSuccessfulRetained(ctx, target, service.now(), basis)
	outcome = ResultForStoredBackupSet(result.BackupSet)
	if result.Run.RestoreVerificationRunID != uuid.Nil {
		outcome.ArtifactRefs = append(outcome.ArtifactRefs, ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+result.Run.RestoreVerificationRunID.String(), outcome.BackupSetID))
	}
	if err != nil {
		return outcome, classifyRestoreFailure(err, true)
	}
	ReportProgress(progress, "attestation_update", 0, nil)
	ReportProgress(progress, "journal_write", 0, nil)
	ReportProgress(progress, "finalize", 1, IntPtr(1))
	return outcome, nil
}

func (service Service) runRestoreVerifyDue(ctx context.Context, parsed operationRequest, progress ProgressSink) (outcome Result, err error) {
	ReportProgress(progress, "preflight", 0, nil)
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := service.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return Result{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	defer func() { _ = recovery.CloseBackupStorage(backupStorage) }()
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return Result{}, err
	}
	defer func() { service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()
	if err := service.preflightRestoreVerificationTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return Result{}, err
	}
	basis, err := restoreVerificationBasisForConfig(sourceCfg)
	if err != nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, err)
	}
	sourceStore := recovery.NewStore(sourcePool)
	due, err := recovery.NewBackupCatalog(sourceStore, backupStorage, service.ExtensionBackups).ListBackupsDueForRestoreVerification(ctx, service.now(), basis)
	if err != nil {
		return Result{}, classifyAdmissionFailure(err)
	}
	if len(due) == 0 {
		return Result{ArtifactRefs: []ArtifactRef{}, Status: ResultNoOp}, nil
	}
	verify := recovery.NewRestoreVerificationService(sourceStore, recovery.NewRestoreRunner(sourceStore, backupStorage, service.ExtensionBackups))
	for _, backupSet := range due {
		target, err := service.restoreVerificationTarget(targetPool, targetObjectStore)
		if err != nil {
			return outcome, NewFailure(FailureVerificationProjectionRebuild, err)
		}
		ReportProgress(progress, "postgres_restore", 0, nil)
		ReportProgress(progress, "object_restore", 0, nil)
		ReportProgress(progress, "projection_rebuild", 0, nil)
		ReportProgress(progress, "invariant_check", 0, nil)
		ReportProgress(progress, "workbook_probe", 0, nil)
		result, verifyErr := verify.VerifyBackupSet(ctx, target, backupSet, basis)
		if outcome.BackupSetID == nil {
			outcome = ResultForStoredBackupSet(result.BackupSet)
		}
		if result.Run.RestoreVerificationRunID != uuid.Nil {
			backupSetID := result.BackupSet.BackupSetID
			outcome.ArtifactRefs = append(outcome.ArtifactRefs, ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+result.Run.RestoreVerificationRunID.String(), &backupSetID))
		}
		if verifyErr != nil {
			return outcome, classifyRestoreFailure(verifyErr, true)
		}
		if err := recovery.ResetRestoreVerificationTarget(ctx, target.RestoreTarget, service.ExtensionBackups); err != nil {
			return outcome, NewFailure(FailureVerificationInvariantCheck, fmt.Errorf("reset disposable restore verification target: %w", err))
		}
		ReportProgress(progress, "attestation_update", 0, nil)
	}
	ReportProgress(progress, "journal_write", 0, nil)
	ReportProgress(progress, "finalize", len(due), IntPtr(len(due)))
	return outcome, nil
}

func (service Service) openSourceRuntime(ctx context.Context, sourceConfigPath string) (Deployment, PostgresPool, recovery.BackupStorage, func(), error) {
	deployment, err := service.loadDeployment(sourceConfigPath)
	if err != nil {
		return Deployment{}, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("load source deployment configuration: %w", err))
	}
	pool, err := service.setupPostgres(ctx, deployment)
	if err != nil {
		return Deployment{}, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("open source postgres: %w", err))
	}
	backupStorage, err := service.newBackupStorage(deployment)
	if err != nil {
		pool.Close()
		return Deployment{}, nil, nil, nil, classifyConfigOrSecretFailure(fmt.Errorf("open source backup storage: %w", err))
	}
	return deployment, pool, backupStorage, func() {
		_ = recovery.CloseBackupStorage(backupStorage)
		pool.Close()
	}, nil
}

func (service Service) openRestoreRuntime(ctx context.Context, parsed operationRequest) (Deployment, Deployment, PostgresPool, PostgresPool, objectstore.Store, objectstore.Store, recovery.BackupStorage, error) {
	sourceDeployment, err := service.loadDeployment(parsed.SourceConfigPath)
	if err != nil {
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("load source deployment configuration: %w", err))
	}
	targetDeployment, err := service.loadDeployment(parsed.TargetConfigPath)
	if err != nil {
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("load target deployment configuration: %w", err))
	}
	sourcePool, err := service.setupPostgres(ctx, sourceDeployment)
	if err != nil {
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("open source postgres: %w", err))
	}
	targetPool, err := service.setupPostgres(ctx, targetDeployment)
	if err != nil {
		sourcePool.Close()
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("open target postgres: %w", err))
	}
	sourceObjectStore, err := service.setupObjectStore(ctx, sourceDeployment)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("open source object store: %w", err))
	}
	targetObjectStore, err := service.setupObjectStore(ctx, targetDeployment)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		_ = sourceObjectStore.Close()
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("open target object store: %w", err))
	}
	backupStorage, err := service.newBackupStorage(sourceDeployment)
	if err != nil {
		sourcePool.Close()
		targetPool.Close()
		_ = sourceObjectStore.Close()
		_ = targetObjectStore.Close()
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, classifyConfigOrSecretFailure(fmt.Errorf("open source backup storage: %w", err))
	}
	return sourceDeployment, targetDeployment, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, nil
}

func (service Service) restoreTarget(targetPool postgres.DB, targetObjectStore objectstore.Store) (recovery.RestoreTarget, error) {
	rebuilder, _, err := service.newProjectionServices(targetPool)
	if err != nil {
		return recovery.RestoreTarget{}, err
	}
	return recovery.RestoreTarget{
		Stopped:     true,
		Postgres:    targetPool,
		ObjectStore: targetObjectStore,
		Projections: rebuilder,
	}, nil
}

func (service Service) restoreVerificationTarget(targetPool postgres.DB, targetObjectStore objectstore.Store) (recovery.RestoreVerificationTarget, error) {
	rebuilder, query, err := service.newProjectionServices(targetPool)
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	return recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			Stopped:     true,
			Postgres:    targetPool,
			ObjectStore: targetObjectStore,
			Projections: rebuilder,
		},
		Probe: recovery.RestoreVerificationWorkbookProbe{Postgres: targetPool, Query: query},
	}, nil
}

func (service Service) preflightRestoreTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceDeployment Deployment, targetDeployment Deployment, targetPool postgres.DB, targetObjectStore objectstore.Store) error {
	if sameConfigPath(sourceConfigPath, targetConfigPath) {
		return NewFailure(FailureSameDatabaseBinding, errors.New("restore target source and target configuration files must differ"))
	}
	sourcePostgres := sourceDeployment.PostgresSettings
	targetPostgres := targetDeployment.PostgresSettings
	if strings.TrimSpace(sourcePostgres.DSN) == strings.TrimSpace(targetPostgres.DSN) {
		return NewFailure(FailureSameDatabaseBinding, errors.New("restore target source and target database bindings must differ"))
	}
	sourceObject := sourceDeployment.ObjectSettings
	targetObject := targetDeployment.ObjectSettings
	if objectStoreBindingID(sourceObject) == objectStoreBindingID(targetObject) {
		return NewFailure(FailureSameObjectStoreBinding, errors.New("restore target source and target object-store bindings must differ"))
	}
	var schemaVersion int64
	if err := targetPool.QueryRow(ctx, `SELECT COALESCE(MAX(version_id), 0)::bigint FROM goose_db_version WHERE is_applied = true`).Scan(&schemaVersion); err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("inspect restore target schema version: %w", err))
	}
	if schemaVersion < restoreMinimumSchemaVersion {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("restore target schema version %d is below required %d", schemaVersion, restoreMinimumSchemaVersion))
	}
	var rowCount int64
	if err := targetPool.QueryRow(ctx, `
SELECT
    (SELECT COUNT(*) FROM incidents)
  + (SELECT COUNT(*) FROM records)
  + (SELECT COUNT(*) FROM object_blobs)
`).Scan(&rowCount); err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("inspect restore target data rows: %w", err))
	}
	if rowCount != 0 {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("restore target database is not empty (%d incident/record/blob rows)", rowCount))
	}
	objects, err := targetObjectStore.ListObjects(ctx, "")
	if err != nil {
		return NewFailure(FailureTargetObjectNamespaceNotFresh, fmt.Errorf("inspect restore target object store: %w", err))
	}
	if len(objects) != 0 {
		return NewFailure(FailureTargetObjectNamespaceNotFresh, fmt.Errorf("restore target object store is not empty (%d objects)", len(objects)))
	}
	return nil
}

func (service Service) preflightRestoreVerificationTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceDeployment Deployment, targetDeployment Deployment, targetPool postgres.DB, targetObjectStore objectstore.Store) error {
	if sameConfigPath(sourceConfigPath, targetConfigPath) {
		return NewFailure(FailureSameDatabaseBinding, errors.New("restore verification source and target configuration files must differ"))
	}
	sourcePostgres := sourceDeployment.PostgresSettings
	targetPostgres := targetDeployment.PostgresSettings
	if strings.TrimSpace(sourcePostgres.DSN) == strings.TrimSpace(targetPostgres.DSN) {
		return NewFailure(FailureSameDatabaseBinding, errors.New("restore verification source and target database bindings must differ"))
	}
	sourceObject := sourceDeployment.ObjectSettings
	targetObject := targetDeployment.ObjectSettings
	if objectStoreBindingID(sourceObject) == objectStoreBindingID(targetObject) {
		return NewFailure(FailureSameObjectStoreBinding, errors.New("restore verification source and target object-store bindings must differ"))
	}
	if service.ReadTargetMarker == nil {
		return NewFailure(FailureTargetMarkerMissing, errors.New("restore verification target marker reader is required"))
	}
	markerBody, err := service.ReadTargetMarker(
		targetDeployment.BackupStorage.BindingKind,
		targetDeployment.BackupStorage.Path,
	)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NewFailure(FailureTargetMarkerMissing, fmt.Errorf("read restore verification target marker: %w", err))
		}
		return NewFailure(FailureTargetMarkerInvalid, fmt.Errorf("read restore verification target marker: %w", err))
	}
	if err := requireRestoreVerificationTargetMarker(markerBody); err != nil {
		return err
	}
	var schemaVersion int64
	if err := targetPool.QueryRow(ctx, `SELECT COALESCE(MAX(version_id), 0)::bigint FROM goose_db_version WHERE is_applied = true`).Scan(&schemaVersion); err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("inspect restore verification target schema version: %w", err))
	}
	if schemaVersion < restoreMinimumSchemaVersion {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("restore verification target schema version %d is below required %d", schemaVersion, restoreMinimumSchemaVersion))
	}
	if _, err := targetObjectStore.ListObjects(ctx, ""); err != nil {
		return NewFailure(FailureTargetObjectNamespaceNotFresh, fmt.Errorf("inspect restore verification target object store: %w", err))
	}
	return nil
}

func requireRestoreVerificationTargetMarker(body []byte) error {
	var marker struct {
		SchemaID string `json:"schema_id"`
		Purpose  string `json:"purpose"`
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&marker); err != nil {
		return NewFailure(FailureTargetMarkerInvalid, fmt.Errorf("decode restore verification target marker: %w", err))
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return NewFailure(FailureTargetMarkerInvalid, errors.New("restore verification target marker has trailing JSON content"))
	}
	if marker.SchemaID != restoreVerificationTargetSchemaID || marker.Purpose != "restore_verification_target" {
		return NewFailure(FailureTargetMarkerInvalid, errors.New("restore verification target marker has the wrong schema or purpose"))
	}
	return nil
}

func (service Service) acquireOperationLock(ctx context.Context, pool PostgresPool) (func(), error) {
	locked, err := tryRecoveryOperationAdvisoryLock(ctx, pool)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, NewFailure(FailureOperationLockUnavailable, errors.New("recovery operation lock unavailable"))
	}
	return func() { _ = unlockRecoveryOperationAdvisoryLock(context.Background(), pool) }, nil
}

func tryRecoveryOperationAdvisoryLock(ctx context.Context, pool postgres.DB) (bool, error) {
	var locked bool
	if err := pool.QueryRow(ctx, `SELECT pg_try_advisory_lock($1)`, recoveryOperationAdvisoryLockKey).Scan(&locked); err != nil {
		return false, fmt.Errorf("acquire restore verification advisory lock: %w", err)
	}
	return locked, nil
}

func unlockRecoveryOperationAdvisoryLock(ctx context.Context, pool postgres.DB) error {
	var unlocked bool
	if err := pool.QueryRow(ctx, `SELECT pg_advisory_unlock($1)`, recoveryOperationAdvisoryLockKey).Scan(&unlocked); err != nil {
		return fmt.Errorf("release restore verification advisory lock: %w", err)
	}
	return nil
}

func (service Service) recordRecoveryStart(ctx context.Context, pool PostgresPool, parsed operationRequest) error {
	record := JournalRecord{
		OperationID: parsed.OperationID.String(),
		Operation:   string(parsed.Operation),
		Result:      "started",
		Summary: map[string]any{
			"source_config_supplied": parsed.SourceConfigPath != "",
			"target_config_supplied": parsed.TargetConfigPath != "",
		},
	}
	store := service.journalStore(pool)
	if err := store.Append(ctx, record); err != nil {
		return NewFailure(journalFailureKind(parsed.Operation), err)
	}
	if parsed.Operation == OperationRestoreLatest {
		if err := store.AppendAuditSummary(ctx, record); err != nil {
			return NewFailure(journalFailureKind(parsed.Operation), err)
		}
	}
	return nil
}

func (service Service) finishJournalAndAudit(ctx context.Context, pool PostgresPool, parsed operationRequest, outcome Result, operationErr *error) {
	result := string(outcome.Status)
	if result == "" {
		result = string(ResultSucceeded)
	}
	var errorCode string
	var reasonCode string
	if operationErr != nil && *operationErr != nil {
		result = "failed"
		*operationErr = EnsureFailure(defaultFailureKind(parsed.Operation), *operationErr)
		if kind, ok := FailureKindOf(*operationErr); ok {
			errorCode, reasonCode = service.failureEvidenceFields(kind)
		}
	}
	record := JournalRecord{
		OperationID: parsed.OperationID.String(),
		Operation:   string(parsed.Operation),
		Result:      result,
		BackupSetID: stringUUIDPtr(outcome.BackupSetID),
		ErrorCode:   errorCode,
		ReasonCode:  reasonCode,
		Summary: map[string]any{
			"artifact_ref_count": len(outcome.ArtifactRefs),
			"has_backup_set_id":  outcome.BackupSetID != nil,
		},
	}
	store := service.journalStore(pool)
	if err := store.Append(ctx, record); err != nil {
		mergeOperationError(operationErr, NewFailure(journalFailureKind(parsed.Operation), err))
		return
	}
	if err := store.AppendAuditSummary(ctx, record); err != nil {
		mergeOperationError(operationErr, NewFailure(journalFailureKind(parsed.Operation), err))
	}
}

func (service Service) journalStore(pool PostgresPool) JournalStore {
	return JournalStore{
		DB:      pool,
		LoadKey: service.LoadJournalKey,
		Now:     service.now,
	}
}

func loadBackupObjectBlobIndex(ctx context.Context, db postgres.DB) (map[string]uuid.UUID, error) {
	rows, err := db.Query(ctx, `
SELECT storage_key, object_blob_id
FROM object_blobs
WHERE storage_key IS NOT NULL AND storage_key <> ''
ORDER BY storage_key ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list object blobs for backup manifest: %w", err)
	}
	defer rows.Close()
	index := make(map[string]uuid.UUID)
	for rows.Next() {
		var storageKey string
		var objectBlobID uuid.UUID
		if err := rows.Scan(&storageKey, &objectBlobID); err != nil {
			return nil, fmt.Errorf("scan object blob for backup manifest: %w", err)
		}
		index[storageKey] = objectBlobID
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate object blobs for backup manifest: %w", err)
	}
	return index, nil
}

func backupObjectStoreBucket(deployment Deployment) (string, error) {
	settings := deployment.ObjectSettings
	switch settings.BindingKind {
	case "managed_service":
		if strings.TrimSpace(settings.Bucket) == "" {
			return "", errors.New("backup create requires configured object-store bucket")
		}
		return settings.Bucket, nil
	case "filesystem_root":
		if strings.TrimSpace(settings.RootPath) == "" {
			return "", errors.New("backup create requires configured filesystem object-store root")
		}
		return "filesystem-root:" + filepath.Clean(settings.RootPath), nil
	default:
		return "", fmt.Errorf("backup create unsupported object-store binding kind %q", settings.BindingKind)
	}
}

func restoreVerificationBasisForConfig(deployment Deployment) (string, error) {
	return recovery.RestoreVerificationBasisSHA256(map[string]string{
		"backup_mechanism":         "cartulary.backup.filesystem_snapshot.v1",
		"database_storage_binding": rootBindingBasis(deployment.DatabaseStorage),
		"object_storage_binding":   rootBindingBasis(deployment.ObjectStorage),
		"backup_storage_binding":   rootBindingBasis(deployment.BackupStorage),
	})
}

func rootBindingBasis(binding RootBinding) string {
	switch binding.BindingKind {
	case "filesystem_root":
		return "filesystem_root:" + filepath.Clean(binding.Path)
	case "managed_service":
		return "managed_service:" + strings.TrimSpace(binding.ServiceRef)
	default:
		return strings.TrimSpace(binding.BindingKind)
	}
}

func sameConfigPath(sourcePath string, targetPath string) bool {
	sourceAbs, sourceErr := filepath.Abs(strings.TrimSpace(sourcePath))
	targetAbs, targetErr := filepath.Abs(strings.TrimSpace(targetPath))
	if sourceErr == nil && targetErr == nil {
		return filepath.Clean(sourceAbs) == filepath.Clean(targetAbs)
	}
	return filepath.Clean(strings.TrimSpace(sourcePath)) == filepath.Clean(strings.TrimSpace(targetPath))
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

func mergeOperationError(operationErr *error, err error) {
	if operationErr == nil || err == nil {
		return
	}
	if *operationErr == nil {
		*operationErr = err
		return
	}
	*operationErr = fmt.Errorf("%w; additionally %v", *operationErr, err)
}

func classifyConfigOrSecretFailure(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, context.DeadlineExceeded):
		return NewFailure(FailureTimeoutElapsed, err)
	case errors.Is(err, recovery.ErrRecoveryMasterKeyRequired):
		return NewFailure(FailureSecretReferenceMissing, err)
	case errors.Is(err, recovery.ErrRecoveryMasterKeyInvalid):
		return NewFailure(FailureRecoveryKeyInvalid, err)
	default:
		return NewFailure(FailureLocalConfigInvalid, err)
	}
}

func classifyAdmissionFailure(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := FailureKindOf(err); ok {
		return err
	}
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return NewFailure(FailureTimeoutElapsed, err)
	case errors.Is(err, recovery.ErrRecoveryMasterKeyRequired):
		return NewFailure(FailureSecretReferenceMissing, err)
	case errors.Is(err, recovery.ErrRecoveryMasterKeyInvalid):
		return NewFailure(FailureRecoveryKeyInvalid, err)
	case errors.Is(err, recovery.ErrNoSuccessfulRetainedBackup),
		errors.Is(err, recovery.ErrLatestSuccessfulBackupStale),
		errors.Is(err, recovery.ErrAmbiguousBackupSelection),
		errors.Is(err, recovery.ErrBackupSetNotFound):
		return NewFailure(FailureNoSuccessfulRetainedBackup, err)
	case errors.Is(err, os.ErrNotExist):
		return NewFailure(FailureArtifactMissing, err)
	case errors.Is(err, recovery.ErrEncryptedBackupStorage):
		return NewFailure(FailureIntegrityProofMissing, err)
	case errors.Is(err, recovery.ErrInvalidBackupArtifact):
		return NewFailure(FailureChecksumMismatch, err)
	default:
		return NewFailure(FailureAttestationInvalid, err)
	}
}

func classifyRestoreFailure(err error, verification bool) error {
	if err == nil {
		return nil
	}
	if _, ok := FailureKindOf(err); ok {
		return err
	}
	if errors.Is(err, recovery.ErrWorkbookProbeFailed) {
		return NewFailure(FailureVerificationWorkbookProbe, err)
	}
	switch {
	case errors.Is(err, recovery.ErrRestoreTargetNotStopped):
		return NewFailure(FailureTargetServingTraffic, err)
	case errors.Is(err, recovery.ErrRestoreTargetNotEmpty):
		return NewFailure(FailureTargetDatabaseNotFresh, err)
	}
	var stageErr *recovery.RestoreStageError
	if errors.As(err, &stageErr) {
		if verification {
			switch stageErr.Stage {
			case recovery.RestoreStepPostgresRestore:
				return NewFailure(FailureVerificationPostgres, err)
			case recovery.RestoreStepObjectStoreRestore:
				return NewFailure(FailureVerificationObject, err)
			case recovery.RestoreStepProjectionRebuild:
				return NewFailure(FailureVerificationProjectionRebuild, err)
			default:
				return NewFailure(FailureVerificationInvariantCheck, err)
			}
		}
		switch stageErr.Stage {
		case recovery.RestoreStepPostgresRestore:
			return NewFailure(FailureRestorePostgres, err)
		case recovery.RestoreStepObjectStoreRestore:
			return NewFailure(FailureRestoreObject, err)
		case recovery.RestoreStepProjectionRebuild:
			return NewFailure(FailureRestoreProjectionRebuild, err)
		default:
			return NewFailure(FailureRestoreInvariantCheck, err)
		}
	}
	switch {
	case errors.Is(err, recovery.ErrNoSuccessfulRetainedBackup),
		errors.Is(err, recovery.ErrLatestSuccessfulBackupStale),
		errors.Is(err, recovery.ErrAmbiguousBackupSelection),
		errors.Is(err, recovery.ErrBackupSetNotFound),
		errors.Is(err, os.ErrNotExist),
		errors.Is(err, recovery.ErrEncryptedBackupStorage),
		errors.Is(err, recovery.ErrInvalidBackupArtifact):
		return classifyAdmissionFailure(err)
	}
	if verification {
		return NewFailure(FailureVerificationInvariantCheck, err)
	}
	return NewFailure(FailureRestoreInvariantCheck, err)
}

func defaultFailureKind(operation Operation) FailureKind {
	switch operation {
	case OperationBackupInspectLatest:
		return FailureArtifactMissing
	case OperationBackupCreate:
		return FailureBackupPublication
	case OperationRestoreLatest:
		return FailureRestoreInvariantCheck
	case OperationRestoreVerifyLatest, OperationRestoreVerifyDue:
		return FailureVerificationInvariantCheck
	default:
		panic(fmt.Sprintf("unsupported recovery operation %q", operation))
	}
}

func journalFailureKind(operation Operation) FailureKind {
	switch operation {
	case OperationBackupCreate:
		return FailureBackupJournalWrite
	case OperationRestoreLatest:
		return FailureRestoreJournalWrite
	case OperationRestoreVerifyLatest, OperationRestoreVerifyDue:
		return FailureVerificationJournalWrite
	default:
		return defaultFailureKind(operation)
	}
}

func (service Service) failureEvidenceFields(kind FailureKind) (string, string) {
	if service.ProjectFailureEvidence == nil {
		return "", string(kind)
	}
	return service.ProjectFailureEvidence(kind)
}

func stringUUIDPtr(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	rendered := value.String()
	return &rendered
}

func (service Service) loadDeployment(path string) (Deployment, error) {
	if service.LoadDeployment == nil {
		return Deployment{}, errors.New("operator recovery requires deployment loader")
	}
	return service.LoadDeployment(path)
}

func (service Service) setupPostgres(ctx context.Context, deployment Deployment) (PostgresPool, error) {
	if deployment.OpenPostgres == nil {
		return nil, errors.New("operator recovery requires postgres opener")
	}
	return deployment.OpenPostgres(ctx)
}

func (service Service) setupObjectStore(ctx context.Context, deployment Deployment) (objectstore.Store, error) {
	if deployment.OpenObjectStore == nil {
		return nil, errors.New("operator recovery requires object-store opener")
	}
	return deployment.OpenObjectStore(ctx)
}

func (service Service) newBackupStorage(deployment Deployment) (recovery.BackupStorage, error) {
	if deployment.OpenBackup == nil {
		return nil, errors.New("operator recovery requires backup-storage opener")
	}
	return deployment.OpenBackup()
}

func (service Service) newProjectionServices(db postgres.DB) (restorecontract.ProjectionRebuilder, recovery.WorkbookProjectionQuery, error) {
	if service.NewProjectionServices == nil {
		return nil, nil, errors.New("operator recovery requires projection services")
	}
	rebuilder, query := service.NewProjectionServices(db)
	if rebuilder == nil {
		return nil, nil, errors.New("operator recovery requires projection rebuilder")
	}
	if query == nil {
		return nil, nil, errors.New("operator recovery requires projection query")
	}
	return rebuilder, query, nil
}

func (service Service) now() time.Time {
	if service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}
