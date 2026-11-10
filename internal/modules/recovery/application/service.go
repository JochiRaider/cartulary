package application

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

const (
	recoveryOperationAdvisoryLockKey int64 = 401010
	terminalEvidenceTimeout                = 30 * time.Second
)

var ErrTargetMarkerRequiresFilesystemStorage = errors.New("restore target marker requires filesystem backup storage")

type PostgresPool interface {
	dbport.DB
	Close()
}

// Deployment is the recovery-owned projection of an admitted deployment.
// Resource factories are application-owned closures so recovery never receives
// or interprets the complete deployment configuration.
type Deployment struct {
	DatabaseStorage            recovery.RootBinding
	ObjectStorage              recovery.RootBinding
	ReferencePackStorage       recovery.RootBinding
	ExportOutputs              recovery.RootBinding
	BackupStorage              recovery.RootBinding
	PostgresSettings           postgres.Settings
	ObjectSettings             objectstore.Settings
	OpenPostgres               func(context.Context) (PostgresPool, error)
	OpenObjectStore            func(context.Context) (objectstore.Store, error)
	OpenReferencePacks         func() (recovery.ReferencePackStorage, error)
	OpenExportOutputs          func() (recovery.RootObjectStorage, error)
	OpenProofScratch           func() (*os.Root, error)
	OpenBackup                 func() (recovery.BackupStorage, error)
	ServingLeaseAcquireTimeout time.Duration
	ServingLeaseLossDetection  time.Duration
}

func (deployment Deployment) TargetBindings() recovery.TargetBindingDigests {
	return recovery.TargetBindingDigestsFor(recovery.TargetBindingInputs{
		DatabaseStorage: deployment.DatabaseStorage, ObjectStorage: deployment.ObjectStorage,
		ReferencePackStorage: deployment.ReferencePackStorage, ExportOutputs: deployment.ExportOutputs,
		PostgresSettings: deployment.PostgresSettings, ObjectSettings: deployment.ObjectSettings,
	})
}

type DeploymentLoader func(string) (Deployment, error)
type ProjectionServicesFactory func(dbport.DB) (restorecontract.ProjectionRebuilder, workbookprobe.Executor, error)
type EvidenceRecoveryProviderFactory func(dbport.DB) recovery.EvidenceRecoveryProvider
type FailureEvidenceProjector func(FailureKind) (code string, reasonCode string)
type RecoveryStateCoverageValidator func(context.Context, dbport.DB, *recoverystate.Catalog) error
type VNextCaptureFactory func(
	PostgresPool,
	objectstore.Store,
	recovery.VNextObjectSource,
	recovery.VNextObjectSource,
	recovery.BackupStorage,
	*recoverystate.Catalog,
	*os.Root,
) (*recovery.VNextCaptureService, error)

type Service struct {
	OpenTransferDirectory  func(string) (recovery.TransferDirectory, error)
	OpenPortableJournal    func(string) (PortableJournal, error)
	LoadRecoveryKey        func() (recovery.RecoveryEncryptionKey, error)
	NewTransferDirectory   func(string) (recovery.TransferDirectory, error)
	ReleaseIdentity        func() (string, error)
	LoadDeployment         DeploymentLoader
	ReadTargetMarker       TargetMarkerReader
	WriteTargetMarker      TargetMarkerWriter
	NewProjectionServices  ProjectionServicesFactory
	NewEvidenceProvider    EvidenceRecoveryProviderFactory
	NewEvidenceRepository  RecoveryEvidenceRepositoryFactory
	NewTargetAdmission     TargetServingAdmissionFactory
	InspectStoppedTarget   func(context.Context, PostgresPool) error
	ProjectFailureEvidence FailureEvidenceProjector
	ExtensionBackups       *recovery.ExtensionBackupCatalog
	RecoveryStateCatalog   *recoverystate.Catalog
	ValidateStateCoverage  RecoveryStateCoverageValidator
	NewVNextCapture        VNextCaptureFactory
	Now                    func() time.Time
}

var _ Facade = Service{}

type operationRequest struct {
	OperationID        uuid.UUID
	Operation          Operation
	AttemptID          *string
	StartedAt          time.Time
	SourceConfigPath   string
	TargetConfigPath   string
	ConfirmedBackupSet uuid.UUID
	AttemptTimeout     time.Duration
	BackupSetID        *uuid.UUID
	ConsistencyPointAt *time.Time
	ArtifactKinds      []string
}

func (service Service) BackupInspectLatest(ctx context.Context, request BackupInspectLatestRequest, progress ProgressSink) (Result, error) {
	result, err := service.backupInspectLatest(ctx, operationRequest{
		OperationID:      request.OperationID,
		Operation:        OperationBackupInspectLatest,
		StartedAt:        service.now(),
		SourceConfigPath: request.SourceConfigPath,
	}, progress)
	return result, EnsureFailure(FailureArtifactMissing, err)
}

func (service Service) BackupCreate(ctx context.Context, request BackupCreateRequest, progress ProgressSink) (Result, error) {
	result, err := service.backupCreate(ctx, operationRequest{
		OperationID:      request.OperationID,
		Operation:        OperationBackupCreate,
		StartedAt:        service.now(),
		SourceConfigPath: request.SourceConfigPath,
	}, progress)
	return result, EnsureFailure(FailureBackupPublication, err)
}

func (service Service) RestoreLatest(ctx context.Context, request RestoreLatestRequest, progress ProgressSink) (Result, error) {
	result, err := service.runRestoreLatest(ctx, operationRequest{
		OperationID:        request.OperationID,
		Operation:          OperationRestoreLatest,
		StartedAt:          service.now(),
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
		StartedAt:        service.now(),
		SourceConfigPath: request.SourceConfigPath,
		TargetConfigPath: request.TargetConfigPath,
	}, progress)
	return result, EnsureFailure(FailureVerificationInvariantCheck, err)
}

func (service Service) RestoreVerifyDue(ctx context.Context, request RestoreVerifyDueRequest, progress ProgressSink) (Result, error) {
	result, err := service.runRestoreVerifyDue(ctx, operationRequest{
		OperationID:      request.OperationID,
		Operation:        OperationRestoreVerifyDue,
		StartedAt:        service.now(),
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
	selection, err := recovery.NewBackupCatalog(
		recovery.NewStore(pool),
		backupStorage,
		service.ExtensionBackups,
		service.RecoveryStateCatalog,
	).RestoreCandidateBackupSelection(ctx, service.now())
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
	if err := service.validateRecoveryStateCatalog(ctx, pool); err != nil {
		return Result{}, NewFailure(FailureBackupPublication, err)
	}

	consistencyPointAt := service.now()
	backupSetID := uuid.New()
	objectStore, err := service.setupObjectStore(ctx, cfg)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupObject, fmt.Errorf("open object store: %w", err))
	}
	defer func() { _ = objectStore.Close() }()
	if service.NewVNextCapture == nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(
			FailureBackupPublication,
			errors.New("vNext capture assembly is unavailable"),
		)
	}
	if cfg.OpenReferencePacks == nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, errors.New("reference pack recovery storage is required"))
	}
	referencePacks, err := cfg.OpenReferencePacks()
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupObject, err)
	}
	defer referencePacks.Close()
	if cfg.OpenExportOutputs == nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, errors.New("export recovery storage is required"))
	}
	exports, err := cfg.OpenExportOutputs()
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupObject, err)
	}
	defer exports.Close()
	if cfg.OpenProofScratch == nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, errors.New("proof scratch capability is required"))
	}
	proofScratch, err := cfg.OpenProofScratch()
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, err)
	}
	defer proofScratch.Close()
	capture, err := service.NewVNextCapture(pool, objectStore, referencePacks, exports, backupStorage, service.RecoveryStateCatalog, proofScratch)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, err)
	}
	ReportProgress(progress, "postgres_backup", 0, nil)
	ReportProgress(progress, "object_backup", 0, nil)
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{
		SourceBindings:     cfg.TargetBindings(),
		BackupSetID:        backupSetID,
		ConsistencyPointAt: consistencyPointAt,
		CreatedAt:          consistencyPointAt,
		RetainedUntil:      consistencyPointAt.Add(recovery.MinimumRetentionDuration),
	})
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, fmt.Errorf("capture vNext backup: %w", err))
	}
	ReportProgress(progress, "attestation_write", 0, nil)
	backupSet, err := recovery.NewStore(pool).PublishVNextCapturedBackup(ctx, captured)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, fmt.Errorf("capture backup set: %w", err))
	}
	if err := recovery.NewBackupCatalog(recovery.NewStore(pool), backupStorage, service.ExtensionBackups, service.RecoveryStateCatalog).VerifyBackupSetDurability(ctx, backupSet); err != nil {
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
	var admission TargetServingAdmission
	defer func() { releaseTargetAdmission(admission, targetCfg.ServingLeaseLossDetection) }()
	if err := service.validateRecoveryStateCatalog(ctx, sourcePool); err != nil {
		return Result{}, NewFailure(FailureRestoreInvariantCheck, err)
	}
	sourceStore := recovery.NewStore(sourcePool)
	backupSet, err := recovery.NewBackupCatalog(
		sourceStore,
		backupStorage,
		service.ExtensionBackups,
		service.RecoveryStateCatalog,
	).RestoreCandidateBackup(ctx, service.now())
	if err != nil {
		return Result{}, classifyAdmissionFailure(err)
	}
	if backupSet.BackupSetID != parsed.ConfirmedBackupSet {
		return ResultForStoredBackupSet(backupSet), NewFailure(
			FailureConfirmationMismatch,
			errors.New("confirmed backup_set_id does not match latest retained backup"),
		)
	}
	generationIdentity, err := recovery.NewBackupCatalog(
		sourceStore,
		backupStorage,
		service.ExtensionBackups,
		service.RecoveryStateCatalog,
	).RecoveryGenerationIdentity(ctx, backupSet)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), classifyAdmissionFailure(err)
	}
	if err := requireDistinctRestoreTarget(parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	if err := service.requireStoppedTarget(ctx, targetPool); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	if replay, found, replayErr := service.replaySuccessfulRestore(ctx, sourcePool, parsed, backupSet, generationIdentity, targetCfg, targetPool); replayErr != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureRestoreJournalWrite, replayErr)
	} else if found {
		ReportProgress(progress, "journal_write", 0, nil)
		ReportProgress(progress, "finalize", 1, IntPtr(1))
		return replay, nil
	}
	if err := service.preflightRestoreTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.requireUnusedRestoreOperation(ctx, sourcePool, parsed); err != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureRestoreJournalWrite, err)
	}
	admission, err = service.acquireTargetAdmission(ctx, targetCfg, targetPool, RestoreTargetPurpose, targetObjectStore)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	targetGenerationID, ok := admittedTargetGenerationID(admission)
	if !ok {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureTargetMarkerInvalid, errors.New("restore target generation was not retained after admission"))
	}
	if err := service.preflightRestoreTarget(admission.Context(), parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	target, err := service.restoreTarget(targetCfg, admission.TargetDatabase(), targetObjectStore, parsed.OperationID, targetGenerationID)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureRestoreProjectionRebuild, err)
	}
	defer target.ReferencePacks.Close()
	defer target.ExportOutputs.Close()
	defer target.ProofScratch.Close()
	runner := recovery.NewVersionedRestoreRunner(sourceStore, backupStorage, service.ExtensionBackups, service.RecoveryStateCatalog)
	identity := restoreControlIdentity(parsed, backupSet, generationIdentity, targetCfg, targetGenerationID, RestoreTargetPurpose)
	var fence *recovery.FencedRestore
	started := false
	defer func() {
		if fence != nil {
			_ = fence.Close()
		}
	}()
	defer func() {
		if started {
			outcome = resultForAdmittedRestore(outcome, identity, backupSet)
			service.finishAdmittedRestore(admission.Context(), sourcePool, parsed, outcome, &err, fence, identity, generationIdentity)
		}
	}()
	fence, err = runner.AdmitRestore(admission.Context(), admission, target, backupSet, identity, func(work context.Context) error {
		err := service.recordRecoveryStart(work, sourcePool, parsed)
		started = err == nil
		return err
	})
	if err != nil {
		return ResultForStoredBackupSet(backupSet), classifyRestoreFailure(err, false)
	}
	target = fence.Target()
	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	result, err := runner.RestoreBackupSet(admission.Context(), target, backupSet)
	if admissionErr := admission.AssertHeld(); admissionErr != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureTargetServingTraffic, admissionErr)
	}
	if err != nil {
		return ResultForStoredBackupSet(backupSet), classifyRestoreFailure(err, false)
	}
	ReportProgress(progress, "journal_write", 0, nil)
	ReportProgress(progress, "finalize", 1, IntPtr(1))
	outcome = ResultForBackupSet(result.BackupSet, "restore_operation", "cartulary.restore_operation.v1")
	bindings := targetCfg.TargetBindings()
	outcome.targetBindings = &bindings
	return outcome, nil
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
	var admission TargetServingAdmission
	defer func() { releaseTargetAdmission(admission, targetCfg.ServingLeaseLossDetection) }()
	if err := service.validateRecoveryStateCatalog(ctx, sourcePool); err != nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, err)
	}
	if err := requireDistinctRestoreTarget(parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg); err != nil {
		return Result{}, err
	}
	sourceStore := recovery.NewStore(sourcePool)
	backupSet, err := recovery.NewBackupCatalog(sourceStore, backupStorage, service.ExtensionBackups, service.RecoveryStateCatalog).RestoreCandidateBackup(ctx, service.now())
	if err != nil {
		return Result{}, classifyAdmissionFailure(err)
	}
	generationIdentity, err := recovery.NewBackupCatalog(sourceStore, backupStorage, service.ExtensionBackups, service.RecoveryStateCatalog).RecoveryGenerationIdentity(ctx, backupSet)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), classifyAdmissionFailure(err)
	}
	if err := service.requireStoppedTarget(ctx, targetPool); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	if replay, found, replayErr := service.replaySuccessfulRestore(ctx, sourcePool, parsed, backupSet, generationIdentity, targetCfg, targetPool); replayErr != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureVerificationJournalWrite, replayErr)
	} else if found {
		return replay, nil
	}
	if err := service.preflightRestoreTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.requireUnusedRestoreOperation(ctx, sourcePool, parsed); err != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureVerificationJournalWrite, err)
	}
	admission, err = service.acquireTargetAdmission(ctx, targetCfg, targetPool, RestoreVerificationTargetPurpose, targetObjectStore)
	if err != nil {
		return Result{}, err
	}
	if err := service.preflightRestoreTarget(admission.Context(), parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return Result{}, err
	}
	basis, err := service.restoreVerificationBasisForConfigs(sourceCfg, targetCfg)
	if err != nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, err)
	}
	targetGenerationID, ok := admittedTargetGenerationID(admission)
	if !ok {
		return Result{}, NewFailure(FailureTargetMarkerInvalid, errors.New("restore target generation was not retained after admission"))
	}
	target, err := service.restoreVerificationTarget(targetCfg, admission.TargetDatabase(), targetObjectStore, parsed.OperationID, targetGenerationID)
	if err != nil {
		return Result{}, NewFailure(FailureVerificationProjectionRebuild, err)
	}
	defer target.ReferencePacks.Close()
	defer target.ExportOutputs.Close()
	defer target.ProofScratch.Close()
	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	ReportProgress(progress, "workbook_probe", 0, nil)
	runner := recovery.NewVersionedRestoreRunner(sourceStore, backupStorage, service.ExtensionBackups, service.RecoveryStateCatalog)
	verify := recovery.NewRestoreVerificationService(sourceStore, runner)
	identity := restoreControlIdentity(parsed, backupSet, generationIdentity, targetCfg, targetGenerationID, RestoreVerificationTargetPurpose)
	var fence *recovery.FencedRestore
	started := false
	defer func() {
		if fence != nil {
			_ = fence.Close()
		}
	}()
	defer func() {
		if started {
			outcome = resultForAdmittedRestore(outcome, identity, backupSet)
			service.finishAdmittedRestore(admission.Context(), sourcePool, parsed, outcome, &err, fence, identity, generationIdentity)
		}
	}()
	fence, err = runner.AdmitRestore(admission.Context(), admission, target.RestoreTarget, backupSet, identity, func(work context.Context) error {
		err := service.recordRecoveryStart(work, sourcePool, parsed)
		started = err == nil
		return err
	})
	if err != nil {
		return ResultForStoredBackupSet(backupSet), classifyRestoreFailure(err, true)
	}
	target.RestoreTarget = fence.Target()
	result, err := verify.VerifyBackupSetAttempt(admission.Context(), target, backupSet, basis, parsed.OperationID)
	outcome = ResultForStoredBackupSet(result.BackupSet)
	bindings := targetCfg.TargetBindings()
	outcome.targetBindings = &bindings
	if result.Run.RestoreVerificationRunID != uuid.Nil {
		outcome.ArtifactRefs = append(outcome.ArtifactRefs, ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+result.Run.RestoreVerificationRunID.String(), outcome.BackupSetID))
	}
	if admissionErr := admission.AssertHeld(); admissionErr != nil {
		return outcome, NewFailure(FailureTargetServingTraffic, admissionErr)
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
	if parsed.AttemptTimeout <= 0 {
		return Result{}, NewFailure(FailureLocalConfigInvalid, errors.New("restore verification attempt timeout must be positive"))
	}
	sourceCfg, targetCfg, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, err := service.openRestoreRuntime(ctx, parsed)
	if err != nil {
		return Result{}, err
	}
	defer sourcePool.Close()
	defer targetPool.Close()
	defer func() { _ = sourceObjectStore.Close() }()
	defer func() { _ = targetObjectStore.Close() }()
	defer func() { _ = recovery.CloseBackupStorage(backupStorage) }()
	if err := service.validateRecoveryStateCatalog(ctx, sourcePool); err != nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, err)
	}
	if err := requireDistinctRestoreTarget(parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg); err != nil {
		return Result{}, err
	}
	basis, err := service.restoreVerificationBasisForConfigs(sourceCfg, targetCfg)
	if err != nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, err)
	}
	sourceStore := recovery.NewStore(sourcePool)
	due, err := recovery.NewBackupCatalog(
		sourceStore,
		backupStorage,
		service.ExtensionBackups,
		service.RecoveryStateCatalog,
	).ListBackupsDueForRestoreVerification(ctx, service.now(), basis)
	if err != nil {
		return Result{}, classifyAdmissionFailure(err)
	}
	if len(due) == 0 {
		outcome = Result{ArtifactRefs: []ArtifactRef{}, Status: ResultNoOp}
		if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
			return Result{}, err
		}
		defer func() { service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err) }()
		ReportProgress(progress, "journal_write", 0, nil)
		ReportProgress(progress, "finalize", 0, IntPtr(0))
		return outcome, nil
	}
	verify := recovery.NewRestoreVerificationService(
		sourceStore,
		recovery.NewVersionedRestoreRunner(
			sourceStore,
			backupStorage,
			service.ExtensionBackups,
			service.RecoveryStateCatalog,
		),
	)
	outcome, err = executeDueVerificationBatch(
		ctx,
		due,
		parsed.AttemptTimeout,
		func() uuid.UUID { return uuid.New() },
		func(attemptCtx context.Context, backupSet recovery.BackupSet, attemptID uuid.UUID) (Result, error, bool) {
			if err := recovery.NewBackupCatalog(sourceStore, backupStorage, service.ExtensionBackups, service.RecoveryStateCatalog).VerifyBackupSetDurability(attemptCtx, backupSet); err != nil {
				return ResultForStoredBackupSet(backupSet), classifyAdmissionFailure(err), true
			}
			generationIdentity, err := recovery.NewBackupCatalog(sourceStore, backupStorage, service.ExtensionBackups, service.RecoveryStateCatalog).RecoveryGenerationIdentity(attemptCtx, backupSet)
			if err != nil {
				return ResultForStoredBackupSet(backupSet), classifyAdmissionFailure(err), true
			}
			return service.runRestoreVerifyDueAttempt(
				attemptCtx,
				parsed,
				progress,
				sourceCfg,
				targetCfg,
				sourcePool,
				targetPool,
				targetObjectStore,
				verify,
				backupSet,
				generationIdentity,
				basis,
				attemptID,
			)
		},
	)
	if err == nil {
		ReportProgress(progress, "finalize", len(due), IntPtr(len(due)))
	}
	return outcome, err
}

func (service Service) runRestoreVerifyDueAttempt(
	ctx context.Context,
	parsed operationRequest,
	progress ProgressSink,
	sourceCfg Deployment,
	targetCfg Deployment,
	sourcePool PostgresPool,
	targetPool PostgresPool,
	targetObjectStore objectstore.Store,
	verify *recovery.RestoreVerificationService,
	backupSet recovery.BackupSet,
	generationIdentity recovery.RecoveryGenerationIdentity,
	basis recovery.RestoreVerificationBasis,
	attemptID uuid.UUID,
) (outcome Result, attemptErr error, stop bool) {
	outcome = ResultForStoredBackupSet(backupSet)
	attemptIDText := attemptID.String()
	backupSetID := backupSet.BackupSetID
	consistencyPointAt := backupSet.ConsistencyPointAt.UTC()
	attemptRequest := parsed
	attemptRequest.AttemptID = &attemptIDText
	attemptRequest.StartedAt = service.now()
	attemptRequest.BackupSetID = &backupSetID
	attemptRequest.ConsistencyPointAt = &consistencyPointAt
	attemptRequest.ArtifactKinds = []string{"restore_verification"}

	if err := service.requireStoppedTarget(ctx, targetPool); err != nil {
		return outcome, dueAttemptContextFailure(ctx, err), true
	}
	if err := service.preflightRestoreTarget(ctx, parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return outcome, dueAttemptContextFailure(ctx, err), true
	}
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return outcome, dueAttemptContextFailure(ctx, err), true
	}
	defer unlock()
	if err := service.requireUnusedRestoreOperation(ctx, sourcePool, attemptRequest); err != nil {
		return outcome, NewFailure(FailureVerificationJournalWrite, err), true
	}

	var admission TargetServingAdmission
	defer func() { releaseTargetAdmission(admission, targetCfg.ServingLeaseLossDetection) }()

	admission, attemptErr = service.acquireTargetAdmission(ctx, targetCfg, targetPool, RestoreVerificationTargetPurpose, targetObjectStore)
	if attemptErr != nil {
		return outcome, dueAttemptContextFailure(ctx, attemptErr), true
	}
	if attemptErr = service.preflightRestoreTarget(
		admission.Context(),
		parsed.SourceConfigPath,
		parsed.TargetConfigPath,
		sourceCfg,
		targetCfg,
		targetPool,
		targetObjectStore,
	); attemptErr != nil {
		return outcome, dueAttemptContextFailure(admission.Context(), attemptErr), true
	}
	targetGenerationID, ok := admittedTargetGenerationID(admission)
	if !ok {
		return outcome, NewFailure(FailureTargetMarkerInvalid, errors.New("restore target generation was not retained after admission")), true
	}

	target, attemptErr := service.restoreVerificationTarget(targetCfg, admission.TargetDatabase(), targetObjectStore, parsed.OperationID, targetGenerationID)
	if attemptErr != nil {
		attemptErr = NewFailure(FailureVerificationProjectionRebuild, attemptErr)
		return outcome, attemptErr, true
	}
	defer target.ReferencePacks.Close()
	defer target.ExportOutputs.Close()
	defer target.ProofScratch.Close()
	identity := restoreControlIdentity(attemptRequest, backupSet, generationIdentity, targetCfg, targetGenerationID, RestoreVerificationTargetPurpose)
	var fence *recovery.FencedRestore
	started := false
	defer func() {
		if fence != nil {
			_ = fence.Close()
		}
	}()
	defer func() {
		if !started {
			return
		}
		outcome = resultForAdmittedRestore(outcome, identity, backupSet)
		ReportProgress(progress, "journal_write", 0, nil)
		service.finishJournalAndAudit(ctx, sourcePool, attemptRequest, outcome, &attemptErr)
		if kind, ok := FailureKindOf(attemptErr); ok && kind == FailureVerificationJournalWrite {
			stop = true
		}
	}()
	fence, attemptErr = verify.AdmitRestore(admission.Context(), admission, target, backupSet, identity, func(work context.Context) error {
		err := service.recordRecoveryStart(work, sourcePool, attemptRequest)
		started = err == nil
		return err
	})
	if attemptErr != nil {
		return outcome, dueAttemptContextFailure(admission.Context(), classifyRestoreFailure(attemptErr, true)), true
	}
	target.RestoreTarget = fence.Target()

	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	ReportProgress(progress, "workbook_probe", 0, nil)
	result, verifyErr := verify.VerifyBackupSetAttempt(admission.Context(), target, backupSet, basis, attemptID)
	bindings := targetCfg.TargetBindings()
	outcome.targetBindings = &bindings
	if result.ArtifactProof.Key != "" && result.Run.RestoreVerificationRunID == attemptID {
		outcome.ArtifactRefs = append(outcome.ArtifactRefs, ArtifactRefFor(
			"restore_verification",
			recovery.RestoreVerificationArtifactSchemaID,
			"restore_verification:"+attemptIDText,
			&backupSetID,
		))
	}
	if admissionErr := admission.AssertHeld(); admissionErr != nil {
		return outcome, NewFailure(FailureTargetServingTraffic, admissionErr), true
	}
	if verifyErr != nil {
		attemptErr = dueAttemptContextFailure(admission.Context(), classifyRestoreFailure(verifyErr, true))
		if dueAttemptMustStop(attemptErr) {
			return outcome, attemptErr, true
		}
	}

	if closeErr := fence.Close(); closeErr != nil {
		return outcome, NewFailure(FailureVerificationInvariantCheck, closeErr), true
	}
	if resetErr := recovery.ResetRestoreVerificationTarget(admission.Context(), admission, target.RestoreTarget, targetGenerationID, targetCfg.TargetBindings(), service.ExtensionBackups, service.RecoveryStateCatalog); resetErr != nil {
		attemptErr = dueAttemptContextFailure(
			admission.Context(),
			NewFailure(FailureVerificationInvariantCheck, fmt.Errorf("reset disposable restore verification target: %w", resetErr)),
		)
		return outcome, attemptErr, true
	}
	if admissionErr := admission.AssertHeld(); admissionErr != nil {
		return outcome, NewFailure(FailureTargetServingTraffic, admissionErr), true
	}
	if verifyErr == nil {
		ReportProgress(progress, "attestation_update", 0, nil)
	}
	return outcome, attemptErr, false
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
		targetPool.Close()
		sourcePool.Close()
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("open source object store: %w", err))
	}
	targetObjectStore, err := service.setupObjectStore(ctx, targetDeployment)
	if err != nil {
		_ = sourceObjectStore.Close()
		targetPool.Close()
		sourcePool.Close()
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, NewFailure(FailureLocalConfigInvalid, fmt.Errorf("open target object store: %w", err))
	}
	backupStorage, err := service.newBackupStorage(sourceDeployment)
	if err != nil {
		_ = targetObjectStore.Close()
		_ = sourceObjectStore.Close()
		targetPool.Close()
		sourcePool.Close()
		return Deployment{}, Deployment{}, nil, nil, nil, nil, nil, classifyConfigOrSecretFailure(fmt.Errorf("open source backup storage: %w", err))
	}
	return sourceDeployment, targetDeployment, sourcePool, targetPool, sourceObjectStore, targetObjectStore, backupStorage, nil
}

func (service Service) restoreTarget(targetCfg Deployment, targetPool dbport.DB, targetObjectStore objectstore.Store, restoreOperationID, targetGenerationID uuid.UUID) (recovery.RestoreTarget, error) {
	target, err := service.restoreVerificationTarget(targetCfg, targetPool, targetObjectStore, restoreOperationID, targetGenerationID)
	return target.RestoreTarget, err
}

func (service Service) restoreVerificationTarget(
	targetCfg Deployment,
	targetPool dbport.DB,
	targetObjectStore objectstore.Store,
	restoreOperationID uuid.UUID,
	targetGenerationID uuid.UUID,
) (recovery.RestoreVerificationTarget, error) {
	rebuilder, query, err := service.newProjectionServices(targetPool)
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	evidenceProvider, err := service.newEvidenceProvider(targetPool)
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	if targetCfg.OpenReferencePacks == nil {
		return recovery.RestoreVerificationTarget{}, errors.New("reference pack restore storage is required")
	}
	packs, err := targetCfg.OpenReferencePacks()
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	if targetCfg.OpenExportOutputs == nil {
		packs.Close()
		return recovery.RestoreVerificationTarget{}, errors.New("export restore storage is required")
	}
	exports, err := targetCfg.OpenExportOutputs()
	if err != nil {
		packs.Close()
		return recovery.RestoreVerificationTarget{}, err
	}
	if targetCfg.OpenProofScratch == nil {
		packs.Close()
		exports.Close()
		return recovery.RestoreVerificationTarget{}, errors.New("proof scratch capability is required")
	}
	proofScratch, err := targetCfg.OpenProofScratch()
	if err != nil {
		packs.Close()
		exports.Close()
		return recovery.RestoreVerificationTarget{}, err
	}
	return recovery.RestoreVerificationTarget{RestoreTarget: recovery.RestoreTarget{
		ProofScratch:       proofScratch,
		RestoreOperationID: restoreOperationID,
		TargetGenerationID: targetGenerationID,
		Postgres:           targetPool,
		ObjectStore:        targetObjectStore,
		ReferencePacks:     packs,
		ExportOutputs:      exports,
		EvidenceObjects:    evidenceProvider,
		Projections:        rebuilder,
	}, Probe: recovery.RestoreVerificationWorkbookProbe{Executor: query}}, nil
}

func (service Service) preflightRestoreTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceDeployment Deployment, targetDeployment Deployment, targetPool dbport.DB, targetObjectStore objectstore.Store) error {
	if err := requireDistinctRestoreTarget(sourceConfigPath, targetConfigPath, sourceDeployment, targetDeployment); err != nil {
		return err
	}
	return service.inspectPristineTarget(ctx, targetDeployment, targetPool, targetObjectStore)
}

func (service Service) requireStoppedTarget(ctx context.Context, target PostgresPool) error {
	if service.InspectStoppedTarget == nil {
		return NewFailure(FailureTargetServingTraffic, errors.New("stopped target inspection is required"))
	}
	if err := service.InspectStoppedTarget(ctx, target); err != nil {
		return NewFailure(FailureTargetServingTraffic, err)
	}
	return nil
}

func (service Service) inspectPristineTarget(ctx context.Context, deployment Deployment, pool dbport.DB, objects objectstore.Store) error {
	if deployment.OpenReferencePacks == nil || deployment.OpenExportOutputs == nil {
		return NewFailure(FailureTargetDatabaseNotFresh, errors.New("complete restore target roots are required"))
	}
	packs, err := deployment.OpenReferencePacks()
	if err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, err)
	}
	defer packs.Close()
	exports, err := deployment.OpenExportOutputs()
	if err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, err)
	}
	defer exports.Close()
	if err := recovery.InspectPristineRestoreTarget(ctx, recovery.RestoreTarget{Postgres: pool, ObjectStore: objects, ReferencePacks: packs, ExportOutputs: exports}, service.ExtensionBackups, service.RecoveryStateCatalog); err != nil {
		if errors.Is(err, recovery.ErrRestoreObjectNamespaceNotFresh) {
			return NewFailure(FailureTargetObjectNamespaceNotFresh, err)
		}
		return NewFailure(FailureTargetDatabaseNotFresh, err)
	}
	return nil
}

func requireDistinctRestoreTarget(sourceConfigPath string, targetConfigPath string, sourceDeployment Deployment, targetDeployment Deployment) error {
	if sameConfigPath(sourceConfigPath, targetConfigPath) {
		return NewFailure(FailureSameDatabaseBinding, errors.New("restore target source and target configuration files must differ"))
	}
	sourcePostgres := sourceDeployment.PostgresSettings
	targetPostgres := targetDeployment.PostgresSettings
	sourceDB, sourceParseErr := pgx.ParseConfig(sourcePostgres.DSN)
	targetDB, targetParseErr := pgx.ParseConfig(targetPostgres.DSN)
	samePhysical := sourceParseErr == nil && targetParseErr == nil && strings.EqualFold(sourceDB.Host, targetDB.Host) && sourceDB.Port == targetDB.Port && sourceDB.Database == targetDB.Database
	if samePhysical || strings.TrimSpace(sourcePostgres.DSN) == strings.TrimSpace(targetPostgres.DSN) {
		return NewFailure(FailureSameDatabaseBinding, errors.New("restore target source and target database bindings must differ"))
	}
	sourceObject := sourceDeployment.ObjectSettings
	targetObject := targetDeployment.ObjectSettings
	if objectStoreBindingID(sourceObject) == objectStoreBindingID(targetObject) {
		return NewFailure(FailureSameObjectStoreBinding, errors.New("restore target source and target object-store bindings must differ"))
	}
	if sourceDeployment.OpenReferencePacks != nil || targetDeployment.OpenReferencePacks != nil || sourceDeployment.OpenExportOutputs != nil || targetDeployment.OpenExportOutputs != nil {
		sourceRoot, targetRoot := sourceDeployment.ReferencePackStorage, targetDeployment.ReferencePackStorage
		if sourceRoot.BindingKind != "filesystem_root" || targetRoot.BindingKind != "filesystem_root" || !filepath.IsAbs(sourceRoot.Path) || !filepath.IsAbs(targetRoot.Path) {
			return NewFailure(FailureSameObjectStoreBinding, errors.New("restore source and target Reference Pack roots must be distinct admitted filesystem roots"))
		}
		if sourceDeployment.OpenExportOutputs != nil || targetDeployment.OpenExportOutputs != nil {
			for _, root := range []recovery.RootBinding{sourceDeployment.ExportOutputs, targetDeployment.ExportOutputs} {
				if root.BindingKind != "filesystem_root" || !filepath.IsAbs(root.Path) {
					return NewFailure(FailureSameObjectStoreBinding, errors.New("export restore roots must be admitted filesystem roots"))
				}
			}
		}
		// Both restore and disposable-target reset mutate these roots. Compare
		// across storage kinds as well, so target packs cannot alias source
		// objects or backups (nor target objects alias retained source packs).
		for _, destination := range []recovery.RootBinding{targetDeployment.ObjectStorage, targetRoot, targetDeployment.ExportOutputs} {
			for _, source := range []recovery.RootBinding{sourceDeployment.DatabaseStorage, sourceDeployment.ObjectStorage, sourceDeployment.BackupStorage, sourceRoot, sourceDeployment.ExportOutputs} {
				if restoreRootsOverlap(source, destination) {
					return NewFailure(FailureSameObjectStoreBinding, errors.New("restore destination overlaps retained source storage"))
				}
			}
		}
		for _, other := range []recovery.RootBinding{targetDeployment.DatabaseStorage, targetDeployment.ObjectStorage, targetDeployment.BackupStorage, targetRoot} {
			if restoreRootsOverlap(other, targetDeployment.ExportOutputs) {
				return NewFailure(FailureSameObjectStoreBinding, errors.New("export restore root overlaps another target storage binding"))
			}
		}
		if restoreRootsOverlap(targetDeployment.ObjectStorage, targetRoot) || restoreRootsOverlap(targetDeployment.BackupStorage, targetRoot) {
			return NewFailure(FailureSameObjectStoreBinding, errors.New("reference pack restore root overlaps another target storage binding"))
		}
	}
	return nil
}

func restoreRootsOverlap(a, b recovery.RootBinding) bool {
	if a.BindingKind != "filesystem_root" || b.BindingKind != "filesystem_root" {
		return false
	}
	aPath, bPath := filepath.Clean(a.Path), filepath.Clean(b.Path)
	return aPath == bPath || strings.HasPrefix(aPath, strings.TrimRight(bPath, string(filepath.Separator))+string(filepath.Separator)) || strings.HasPrefix(bPath, strings.TrimRight(aPath, string(filepath.Separator))+string(filepath.Separator))
}

func (service Service) acquireTargetAdmission(
	ctx context.Context,
	targetDeployment Deployment,
	targetPool PostgresPool,
	purpose string,
	objects objectstore.Store,
) (TargetServingAdmission, error) {
	if service.NewTargetAdmission == nil {
		return nil, NewFailure(FailureTargetServingTraffic, errors.New("restore target serving admission is unavailable"))
	}
	admission, err := service.NewTargetAdmission(
		ctx,
		targetPool,
		minPositiveDuration(targetDeployment.ServingLeaseAcquireTimeout, RestoreTargetServingLeaseAcquireMax),
		targetDeployment.ServingLeaseLossDetection,
	)
	if err != nil {
		return nil, NewFailure(FailureTargetServingTraffic, fmt.Errorf("acquire restore target serving lease: %w", err))
	}
	return service.prepareAdmittedTarget(targetDeployment, purpose, objects, admission)
}

func (service Service) prepareAdmittedTarget(targetDeployment Deployment, purpose string, objects objectstore.Store, admission TargetServingAdmission) (TargetServingAdmission, error) {
	releaseOnError := func() {
		releaseTargetAdmission(admission, targetDeployment.ServingLeaseLossDetection)
	}
	control, err := recovery.ReadTargetControl(admission.Context(), admission.TargetDatabase())
	if err != nil || (control.State != recovery.TargetInitialized && control.State != recovery.TargetVerificationReady) {
		releaseOnError()
		return nil, NewFailure(FailureTargetDatabaseNotFresh, recovery.ErrTargetIncomplete)
	}
	if control.State == recovery.TargetVerificationReady && (purpose != RestoreVerificationTargetPurpose || control.BindingDigests == nil || *control.BindingDigests != targetDeployment.TargetBindings()) {
		releaseOnError()
		return nil, NewFailure(FailureTargetMarkerInvalid, recovery.ErrTargetIncomplete)
	}
	if service.ReadTargetMarker == nil {
		releaseOnError()
		return nil, NewFailure(FailureTargetMarkerMissing, errors.New("restore target marker reader is required"))
	}
	material, err := service.ReadTargetMarker(
		targetDeployment.BackupStorage.BindingKind,
		targetDeployment.BackupStorage.Path,
	)

	targetGenerationID, err := prepareTargetMarker(material, err, purpose, targetDeployment.TargetBindings(), service.now(), func() error {
		if service.WriteTargetMarker == nil || objects == nil || targetDeployment.OpenReferencePacks == nil || targetDeployment.OpenExportOutputs == nil {
			return errors.New("target proof preparation unavailable")
		}
		if err := admission.AssertHeld(); err != nil {
			return err
		}
		return service.inspectPristineTarget(admission.Context(), targetDeployment, admission.TargetDatabase(), objects)
	}, func(previous, next TargetMarkerMaterial) error {
		if err := admission.AssertHeld(); err != nil {
			return err
		}
		return service.WriteTargetMarker(admission.Context(), targetDeployment.BackupStorage.Path, previous, next)
	})
	if err != nil {
		releaseOnError()
		return nil, NewFailure(FailureTargetMarkerInvalid, err)
	}

	if err := admission.AssertHeld(); err != nil {
		releaseOnError()
		return nil, NewFailure(FailureTargetServingTraffic, err)
	}
	if control.State == recovery.TargetVerificationReady && *control.TargetGenerationID != targetGenerationID.String() {
		if targetDeployment.OpenReferencePacks == nil || targetDeployment.OpenExportOutputs == nil {
			releaseOnError()
			return nil, NewFailure(FailureTargetMarkerInvalid, recovery.ErrTargetIncomplete)
		}
		packs, err := targetDeployment.OpenReferencePacks()
		if err != nil {
			releaseOnError()
			return nil, NewFailure(FailureTargetMarkerInvalid, err)
		}
		defer packs.Close()
		exports, err := targetDeployment.OpenExportOutputs()
		if err != nil {
			releaseOnError()
			return nil, NewFailure(FailureTargetMarkerInvalid, err)
		}
		defer exports.Close()
		target := recovery.RestoreTarget{ObjectStore: objects, ReferencePacks: packs, ExportOutputs: exports}
		if err := recovery.AdmitVerificationTargetGeneration(admission.Context(), admission, target, targetGenerationID, targetDeployment.TargetBindings(), service.ExtensionBackups, service.RecoveryStateCatalog); err != nil {
			releaseOnError()
			return nil, NewFailure(FailureTargetMarkerInvalid, err)
		}
	}
	return targetServingAdmissionWithGeneration{TargetServingAdmission: admission, targetGenerationID: targetGenerationID}, nil
}

type targetServingAdmissionWithGeneration struct {
	TargetServingAdmission
	targetGenerationID uuid.UUID
}

func (admission targetServingAdmissionWithGeneration) TargetGenerationID() uuid.UUID {
	return admission.targetGenerationID
}

func admittedTargetGenerationID(admission TargetServingAdmission) (uuid.UUID, bool) {
	typed, ok := admission.(interface{ TargetGenerationID() uuid.UUID })
	if !ok || typed.TargetGenerationID() == uuid.Nil {
		return uuid.Nil, false
	}
	return typed.TargetGenerationID(), true
}

func minPositiveDuration(configured time.Duration, maximum time.Duration) time.Duration {
	if configured <= 0 || configured > maximum {
		return maximum
	}
	return configured
}

func releaseTargetAdmission(admission TargetServingAdmission, lossDetection time.Duration) {
	if admission == nil {
		return
	}
	if lossDetection <= 0 {
		lossDetection = time.Second
	}
	releaseCtx, cancel := context.WithTimeout(context.Background(), lossDetection)
	defer cancel()
	_ = admission.Release(releaseCtx)
}

func (service Service) validateRecoveryStateCatalog(ctx context.Context, pool PostgresPool) error {
	if err := service.RecoveryStateCatalog.ValidateFrozen(); err != nil {
		return fmt.Errorf("validate frozen recovery state catalog: %w", err)
	}
	if service.ValidateStateCoverage == nil {
		return errors.New("validate frozen recovery state catalog: database coverage validator is unavailable")
	}
	if err := service.ValidateStateCoverage(ctx, pool, service.RecoveryStateCatalog); err != nil {
		return fmt.Errorf("validate frozen recovery state catalog database coverage: %w", err)
	}
	return nil
}

func (service Service) acquireOperationLock(ctx context.Context, pool PostgresPool) (func(), error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin recovery operation exclusion transaction: %w", err)
	}
	locked, err := tryRecoveryOperationAdvisoryLock(ctx, tx)
	if err != nil {
		_ = tx.Rollback(context.Background())
		return nil, err
	}
	if !locked {
		_ = tx.Rollback(context.Background())
		return nil, NewFailure(FailureOperationLockUnavailable, errors.New("recovery operation lock unavailable"))
	}
	return func() { _ = tx.Rollback(context.Background()) }, nil
}

type advisoryLockQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func tryRecoveryOperationAdvisoryLock(ctx context.Context, pool advisoryLockQueryer) (bool, error) {
	var locked bool
	if err := pool.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, recoveryOperationAdvisoryLockKey).Scan(&locked); err != nil {
		return false, fmt.Errorf("acquire recovery operation advisory lock: %w", err)
	}
	return locked, nil
}

func (service Service) recordRecoveryStart(ctx context.Context, pool PostgresPool, parsed operationRequest) error {
	repository, err := service.evidenceRepository(pool)
	if err != nil {
		return NewFailure(journalFailureKind(parsed.Operation), err)
	}
	if err := repository.AppendAdmission(ctx, RecoveryAdmissionRecord{
		OperationID:        parsed.OperationID,
		Operation:          parsed.Operation,
		AttemptID:          parsed.AttemptID,
		StartedAt:          parsed.StartedAt,
		BackupSetID:        parsed.BackupSetID,
		ConsistencyPointAt: parsed.ConsistencyPointAt,
		ArtifactKinds:      parsed.ArtifactKinds,
	}); err != nil {
		return NewFailure(journalFailureKind(parsed.Operation), err)
	}
	return nil
}

func (service Service) finishJournalAndAudit(ctx context.Context, pool PostgresPool, parsed operationRequest, outcome Result, operationErr *error) (RecoveryCompletionRecord, bool) {
	result := outcome.Status
	if result == "" {
		result = ResultSucceeded
	}
	var errorCode *string
	var reasonCode *string
	if operationErr != nil && *operationErr != nil {
		result = ResultFailed
		*operationErr = EnsureFailure(defaultFailureKind(parsed.Operation), *operationErr)
		if kind, ok := FailureKindOf(*operationErr); ok {
			code, reason := service.failureEvidenceFields(kind)
			errorCode = &code
			reasonCode = &reason
		}
	}
	repository, err := service.evidenceRepository(pool)
	if err != nil {
		replaceWithJournalFailure(operationErr, parsed.Operation, err)
		return RecoveryCompletionRecord{}, false
	}
	evidenceCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), terminalEvidenceTimeout)
	defer cancel()
	record, err := NormalizeCompletionRecord(RecoveryCompletionRecord{
		OperationID:        parsed.OperationID,
		Operation:          parsed.Operation,
		AttemptID:          parsed.AttemptID,
		StartedAt:          parsed.StartedAt,
		CompletedAt:        service.now(),
		Result:             result,
		BackupSetID:        outcome.BackupSetID,
		ConsistencyPointAt: outcome.ConsistencyPointAt,
		ArtifactCounts:     ArtifactCountsFor(outcome.ArtifactRefs),
		ErrorCode:          errorCode,
		ErrorReason:        reasonCode,
		RestoreGeneration:  outcome.restoreGeneration,
		TargetBindings:     outcome.targetBindings,
	})
	if err != nil {
		replaceWithJournalFailure(operationErr, parsed.Operation, err)
		return RecoveryCompletionRecord{}, false
	}
	if err := repository.AppendCompletion(evidenceCtx, record); err != nil {
		replaceWithJournalFailure(operationErr, parsed.Operation, err)
		return RecoveryCompletionRecord{}, false
	}
	return record, true
}

func (service Service) evidenceRepository(pool PostgresPool) (RecoveryEvidenceRepository, error) {
	if service.NewEvidenceRepository == nil {
		return nil, errors.New("operator recovery requires evidence repository factory")
	}
	return service.NewEvidenceRepository(pool)
}

func (service Service) replaySuccessfulRestore(
	ctx context.Context,
	pool PostgresPool,
	parsed operationRequest,
	backupSet recovery.BackupSet,
	generationIdentity recovery.RecoveryGenerationIdentity,
	targetDeployment Deployment,
	targetPool PostgresPool,
) (Result, bool, error) {
	repository, err := service.evidenceRepository(pool)
	if err != nil {
		return Result{}, false, err
	}
	record, err := repository.FindCompletion(ctx, parsed.OperationID, parsed.Operation, parsed.AttemptID)
	if err != nil || record == nil {
		return Result{}, false, err
	}
	if record.Result != ResultSucceeded || record.BackupSetID == nil || *record.BackupSetID != backupSet.BackupSetID {
		return Result{}, false, recovery.ErrTargetIncomplete
	}
	purpose := RestoreTargetPurpose
	if parsed.Operation == OperationRestoreVerifyLatest {
		purpose = RestoreVerificationTargetPurpose
	}
	targetBindings := targetDeployment.TargetBindings()
	if service.ReadTargetMarker == nil {
		return Result{}, false, errors.New("restore target marker reader is required")
	}
	material, err := service.ReadTargetMarker(targetDeployment.BackupStorage.BindingKind, targetDeployment.BackupStorage.Path)
	if err != nil {
		return Result{}, false, err
	}
	targetGenerationID, err := admitCompletedTargetMarker(material, purpose, targetBindings, record.StartedAt, record.CompletedAt)
	if err != nil {
		return Result{}, false, err
	}
	completion := record.RestoreGeneration
	if completion == nil || record.TargetBindings == nil || *record.TargetBindings != targetBindings || completion.TargetGenerationID != targetGenerationID ||
		record.ConsistencyPointAt == nil || !record.ConsistencyPointAt.Equal(backupSet.ConsistencyPointAt) || !generationIdentity.AdmitsRestoreGeneration(*completion) {
		return Result{}, false, recovery.ErrTargetIncomplete
	}
	unlock, err := service.acquireOperationLock(ctx, pool)
	if err != nil {
		return Result{}, false, err
	}
	defer unlock()
	if service.NewTargetAdmission == nil {
		return Result{}, false, ErrTargetServingLeaseLost
	}
	admission, err := service.NewTargetAdmission(ctx, targetPool, minPositiveDuration(targetDeployment.ServingLeaseAcquireTimeout, RestoreTargetServingLeaseAcquireMax), targetDeployment.ServingLeaseLossDetection)
	if err != nil {
		return Result{}, false, err
	}
	defer releaseTargetAdmission(admission, targetDeployment.ServingLeaseLossDetection)
	confirmed, err := repository.FindCompletion(admission.Context(), parsed.OperationID, parsed.Operation, parsed.AttemptID)
	if err != nil || confirmed == nil {
		if err == nil {
			err = recovery.ErrTargetIncomplete
		}
		return Result{}, false, err
	}
	original, err := EncodeRecoveryCompletion(*record)
	if err != nil {
		return Result{}, false, err
	}
	current, err := EncodeRecoveryCompletion(*confirmed)
	if err != nil || !bytes.Equal(original, current) {
		return Result{}, false, recovery.ErrTargetIncomplete
	}
	identity := restoreControlIdentity(parsed, backupSet, generationIdentity, targetDeployment, targetGenerationID, purpose)
	proof, err := targetCompletionProof(*confirmed, identity, generationIdentity)
	if err == nil {
		err = recovery.ReconcileTargetCompletion(admission.Context(), admission, proof)
	}
	if err != nil {
		return Result{}, false, err
	}
	result := ResultForBackupSet(backupSet, "restore_operation", "cartulary.restore_operation.v1")
	if parsed.Operation == OperationRestoreVerifyLatest {
		result = ResultForStoredBackupSet(backupSet)
		result.ArtifactRefs = append(result.ArtifactRefs, ArtifactRefFor("restore_verification", recovery.RestoreVerificationArtifactSchemaID, "restore_verification:"+parsed.OperationID.String(), result.BackupSetID))
	}
	result.restoreGeneration = completion
	result.targetBindings = &targetBindings
	return result, true, nil
}

func (service Service) restoreVerificationBasisForConfigs(
	source Deployment,
	target Deployment,
) (recovery.RestoreVerificationBasis, error) {
	if service.RecoveryStateCatalog == nil {
		return recovery.RestoreVerificationBasis{}, recoverystate.ErrInvalidCatalog
	}
	basis := recovery.RestoreVerificationBasis{
		ApplicationCryptoFormat:           recovery.ApplicationCryptoFormatID,
		MechanismID:                       recovery.VNextBackupMechanismID,
		CodecRegistrySHA256:               recovery.VNextCodecRegistrySHA256(),
		RecoveryStateCatalogSHA256:        service.RecoveryStateCatalog.DigestSHA256(),
		DatabaseBindingSHA256:             recovery.SHA256String(target.DatabaseStorage.Identity()),
		ObjectStoreBindingSHA256:          recovery.SHA256String(target.ObjectStorage.Identity()),
		ReferencePackStorageBindingSHA256: recovery.SHA256String(target.ReferencePackStorage.Identity()),
		ExportOutputsBindingSHA256:        recovery.SHA256String(target.ExportOutputs.Identity()),
		BackupStorageBindingSHA256:        recovery.SHA256String(source.BackupStorage.Identity()),
	}
	return basis, basis.Validate()
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

func replaceWithJournalFailure(operationErr *error, operation Operation, journalErr error) {
	if operationErr == nil || journalErr == nil {
		return
	}
	cause := journalErr
	if *operationErr != nil {
		cause = fmt.Errorf("operation failed before terminal evidence: %v; terminal evidence failed: %w", *operationErr, journalErr)
	}
	*operationErr = NewFailure(journalFailureKind(operation), cause)
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
	case errors.Is(err, recovery.ErrRestoreTargetNotEmpty), errors.Is(err, recovery.ErrTargetIncomplete), errors.Is(err, recovery.ErrTargetControlConflict):
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
	case OperationBackupExportLatest:
		return FailureTransferCopy
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
	case OperationBackupExportLatest:
		return FailureExportJournalWrite
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

func (service Service) newProjectionServices(db dbport.DB) (restorecontract.ProjectionRebuilder, workbookprobe.Executor, error) {
	if service.NewProjectionServices == nil {
		return nil, nil, errors.New("operator recovery requires projection services")
	}
	rebuilder, query, err := service.NewProjectionServices(db)
	if err != nil {
		return nil, nil, err
	}
	if rebuilder == nil {
		return nil, nil, errors.New("operator recovery requires projection rebuilder")
	}
	if query == nil {
		return nil, nil, errors.New("operator recovery requires projection query")
	}
	return rebuilder, query, nil
}

func (service Service) newEvidenceProvider(db dbport.DB) (recovery.EvidenceRecoveryProvider, error) {
	if service.NewEvidenceProvider == nil {
		return nil, errors.New("operator recovery requires Evidence recovery provider")
	}
	provider := service.NewEvidenceProvider(db)
	if provider == nil {
		return nil, errors.New("operator recovery Evidence recovery provider is unavailable")
	}
	return provider, nil
}

func (service Service) now() time.Time {
	if service.Now != nil {
		return service.Now().UTC()
	}
	return time.Now().UTC()
}
