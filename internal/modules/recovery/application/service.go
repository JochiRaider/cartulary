package application

import (
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
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

const (
	restoreMinimumSchemaVersion      int64 = 22
	recoveryOperationAdvisoryLockKey int64 = 401010
	terminalEvidenceTimeout                = 30 * time.Second
)

var ErrTargetMarkerRequiresFilesystemStorage = errors.New("restore target marker requires filesystem backup storage")

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
	DatabaseStorage            RootBinding
	ObjectStorage              RootBinding
	BackupStorage              RootBinding
	PostgresSettings           postgres.Settings
	ObjectSettings             objectstore.Settings
	OpenPostgres               func(context.Context) (PostgresPool, error)
	OpenObjectStore            func(context.Context) (objectstore.Store, error)
	OpenBackup                 func() (recovery.BackupStorage, error)
	ServingLeaseAcquireTimeout time.Duration
	ServingLeaseLossDetection  time.Duration
}

type DeploymentLoader func(string) (Deployment, error)
type ProjectionServicesFactory func(postgres.DB) (restorecontract.ProjectionRebuilder, workbookprobe.Executor, error)
type GraphProjectionRestoreFactory func(postgres.DB) (restorecontract.GraphProjectionParticipant, error)
type EvidenceRecoveryProviderFactory func(postgres.DB) recovery.EvidenceRecoveryProvider
type FailureEvidenceProjector func(FailureKind) (code string, reasonCode string)
type RecoveryStateCoverageValidator func(context.Context, PostgresPool, *recoverystate.Catalog) error
type VNextCaptureFactory func(
	PostgresPool,
	objectstore.Store,
	recovery.BackupStorage,
	*recoverystate.Catalog,
) (*recovery.VNextCaptureService, error)

type Service struct {
	LoadDeployment            DeploymentLoader
	ReadTargetMarker          TargetMarkerReader
	NewProjectionServices     ProjectionServicesFactory
	NewGraphProjectionRestore GraphProjectionRestoreFactory
	NewEvidenceProvider       EvidenceRecoveryProviderFactory
	NewEvidenceRepository     RecoveryEvidenceRepositoryFactory
	NewTargetAdmission        TargetServingAdmissionFactory
	ProjectFailureEvidence    FailureEvidenceProjector
	ExtensionBackups          *recovery.ExtensionBackupCatalog
	RecoveryStateCatalog      *recoverystate.Catalog
	ValidateStateCoverage     RecoveryStateCoverageValidator
	NewVNextCapture           VNextCaptureFactory
	Now                       func() time.Time
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
	capture, err := service.NewVNextCapture(pool, objectStore, backupStorage, service.RecoveryStateCatalog)
	if err != nil {
		return ResultForCandidate(backupSetID, consistencyPointAt), NewFailure(FailureBackupPublication, err)
	}
	ReportProgress(progress, "postgres_backup", 0, nil)
	ReportProgress(progress, "object_backup", 0, nil)
	captured, err := capture.Capture(ctx, recovery.VNextCaptureParams{
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
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return Result{}, err
	}
	var admission TargetServingAdmission
	defer func() {
		service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err)
		releaseTargetAdmission(admission, targetCfg.ServingLeaseLossDetection)
	}()
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
	admission, err = service.acquireTargetAdmission(ctx, targetCfg, targetPool, RestoreTargetPurpose)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	targetGenerationID, ok := admittedTargetGenerationID(admission)
	if !ok {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureTargetMarkerInvalid, errors.New("restore target generation was not retained after admission"))
	}
	if replay, found, replayErr := service.replaySuccessfulRestore(ctx, sourcePool, parsed, backupSet, generationIdentity, targetGenerationID); replayErr != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureRestoreJournalWrite, replayErr)
	} else if found {
		if admissionErr := admission.AssertHeld(); admissionErr != nil {
			return ResultForStoredBackupSet(backupSet), NewFailure(FailureTargetServingTraffic, admissionErr)
		}
		ReportProgress(progress, "journal_write", 0, nil)
		ReportProgress(progress, "finalize", 1, IntPtr(1))
		return replay, nil
	}
	if err := service.preflightRestoreTarget(admission.Context(), parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg, targetPool, targetObjectStore); err != nil {
		return ResultForStoredBackupSet(backupSet), err
	}
	target, err := service.restoreTarget(targetPool, targetObjectStore, parsed.OperationID, targetGenerationID)
	if err != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureRestoreProjectionRebuild, err)
	}
	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	result, err := recovery.NewVersionedRestoreRunner(
		sourceStore,
		backupStorage,
		service.ExtensionBackups,
		service.RecoveryStateCatalog,
	).RestoreBackupSet(admission.Context(), target, backupSet)
	if admissionErr := admission.AssertHeld(); admissionErr != nil {
		return ResultForStoredBackupSet(backupSet), NewFailure(FailureTargetServingTraffic, admissionErr)
	}
	if err != nil {
		return ResultForStoredBackupSet(backupSet), classifyRestoreFailure(err, false)
	}
	ReportProgress(progress, "journal_write", 0, nil)
	ReportProgress(progress, "finalize", 1, IntPtr(1))
	outcome = ResultForBackupSet(result.BackupSet, "restore_operation", "cartulary.restore_operation.v1")
	outcome.graphProjectionCompletion = result.GraphProjectionCompletion
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
	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return Result{}, err
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, parsed); err != nil {
		return Result{}, err
	}
	var admission TargetServingAdmission
	defer func() {
		service.finishJournalAndAudit(ctx, sourcePool, parsed, outcome, &err)
		releaseTargetAdmission(admission, targetCfg.ServingLeaseLossDetection)
	}()
	if err := service.validateRecoveryStateCatalog(ctx, sourcePool); err != nil {
		return Result{}, NewFailure(FailureVerificationInvariantCheck, err)
	}
	if err := requireDistinctRestoreTarget(parsed.SourceConfigPath, parsed.TargetConfigPath, sourceCfg, targetCfg); err != nil {
		return Result{}, err
	}
	admission, err = service.acquireTargetAdmission(ctx, targetCfg, targetPool, RestoreVerificationTargetPurpose)
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
	target, err := service.restoreVerificationTarget(targetPool, targetObjectStore, parsed.OperationID, targetGenerationID)
	if err != nil {
		return Result{}, NewFailure(FailureVerificationProjectionRebuild, err)
	}
	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	ReportProgress(progress, "workbook_probe", 0, nil)
	sourceStore := recovery.NewStore(sourcePool)
	verify := recovery.NewRestoreVerificationService(
		sourceStore,
		recovery.NewVersionedRestoreRunner(
			sourceStore,
			backupStorage,
			service.ExtensionBackups,
			service.RecoveryStateCatalog,
		),
	)
	result, err := verify.VerifyLatestSuccessfulRetained(admission.Context(), target, service.now(), basis)
	outcome = ResultForStoredBackupSet(result.BackupSet)
	outcome.graphProjectionCompletion = result.RestoreResult.GraphProjectionCompletion
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

	unlock, err := service.acquireOperationLock(ctx, sourcePool)
	if err != nil {
		return outcome, dueAttemptContextFailure(ctx, err), true
	}
	defer unlock()
	if err := service.recordRecoveryStart(ctx, sourcePool, attemptRequest); err != nil {
		return outcome, dueAttemptContextFailure(ctx, err), true
	}

	var admission TargetServingAdmission
	defer func() {
		ReportProgress(progress, "journal_write", 0, nil)
		service.finishJournalAndAudit(ctx, sourcePool, attemptRequest, outcome, &attemptErr)
		if kind, ok := FailureKindOf(attemptErr); ok && kind == FailureVerificationJournalWrite {
			stop = true
		}
		releaseTargetAdmission(admission, targetCfg.ServingLeaseLossDetection)
	}()

	admission, attemptErr = service.acquireTargetAdmission(ctx, targetCfg, targetPool, RestoreVerificationTargetPurpose)
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
	target, attemptErr := service.restoreVerificationTarget(targetPool, targetObjectStore, parsed.OperationID, targetGenerationID)
	if attemptErr != nil {
		attemptErr = NewFailure(FailureVerificationProjectionRebuild, attemptErr)
		return outcome, attemptErr, true
	}

	ReportProgress(progress, "postgres_restore", 0, nil)
	ReportProgress(progress, "object_restore", 0, nil)
	ReportProgress(progress, "projection_rebuild", 0, nil)
	ReportProgress(progress, "invariant_check", 0, nil)
	ReportProgress(progress, "workbook_probe", 0, nil)
	result, verifyErr := verify.VerifyBackupSetAttempt(admission.Context(), target, backupSet, basis, attemptID)
	outcome.graphProjectionCompletion = result.RestoreResult.GraphProjectionCompletion
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

	if resetErr := recovery.ResetRestoreVerificationTarget(admission.Context(), target.RestoreTarget, service.ExtensionBackups); resetErr != nil {
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

func (service Service) restoreTarget(
	targetPool postgres.DB,
	targetObjectStore objectstore.Store,
	restoreOperationID uuid.UUID,
	targetGenerationID uuid.UUID,
) (recovery.RestoreTarget, error) {
	rebuilder, _, err := service.newProjectionServices(targetPool)
	if err != nil {
		return recovery.RestoreTarget{}, err
	}
	graphRestore, err := service.newGraphProjectionRestore(targetPool)
	if err != nil {
		return recovery.RestoreTarget{}, err
	}
	evidenceProvider, err := service.newEvidenceProvider(targetPool)
	if err != nil {
		return recovery.RestoreTarget{}, err
	}
	return recovery.RestoreTarget{
		RestoreOperationID: restoreOperationID,
		TargetGenerationID: targetGenerationID,
		Postgres:           targetPool,
		ObjectStore:        targetObjectStore,
		EvidenceObjects:    evidenceProvider,
		GraphProjection:    graphRestore,
		Projections:        rebuilder,
	}, nil
}

func (service Service) restoreVerificationTarget(
	targetPool postgres.DB,
	targetObjectStore objectstore.Store,
	restoreOperationID uuid.UUID,
	targetGenerationID uuid.UUID,
) (recovery.RestoreVerificationTarget, error) {
	rebuilder, query, err := service.newProjectionServices(targetPool)
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	graphRestore, err := service.newGraphProjectionRestore(targetPool)
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	evidenceProvider, err := service.newEvidenceProvider(targetPool)
	if err != nil {
		return recovery.RestoreVerificationTarget{}, err
	}
	return recovery.RestoreVerificationTarget{
		RestoreTarget: recovery.RestoreTarget{
			RestoreOperationID: restoreOperationID,
			TargetGenerationID: targetGenerationID,
			Postgres:           targetPool,
			ObjectStore:        targetObjectStore,
			EvidenceObjects:    evidenceProvider,
			GraphProjection:    graphRestore,
			Projections:        rebuilder,
		},
		Probe: recovery.RestoreVerificationWorkbookProbe{Executor: query},
	}, nil
}

func (service Service) preflightRestoreTarget(ctx context.Context, sourceConfigPath string, targetConfigPath string, sourceDeployment Deployment, targetDeployment Deployment, targetPool postgres.DB, targetObjectStore objectstore.Store) error {
	if err := requireDistinctRestoreTarget(sourceConfigPath, targetConfigPath, sourceDeployment, targetDeployment); err != nil {
		return err
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
`).Scan(&rowCount); err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("inspect restore target data rows: %w", err))
	}
	evidenceProvider, err := service.newEvidenceProvider(targetPool)
	if err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, err)
	}
	evidenceRows, err := evidenceProvider.CountRecoveryRows(ctx)
	if err != nil {
		return NewFailure(FailureTargetDatabaseNotFresh, fmt.Errorf("inspect restore target Evidence rows: %w", err))
	}
	rowCount += evidenceRows
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

func requireDistinctRestoreTarget(sourceConfigPath string, targetConfigPath string, sourceDeployment Deployment, targetDeployment Deployment) error {
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
	return nil
}

func (service Service) acquireTargetAdmission(
	ctx context.Context,
	targetDeployment Deployment,
	targetPool PostgresPool,
	purpose string,
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
	releaseOnError := func() {
		releaseTargetAdmission(admission, targetDeployment.ServingLeaseLossDetection)
	}
	if service.ReadTargetMarker == nil {
		releaseOnError()
		return nil, NewFailure(FailureTargetMarkerMissing, errors.New("restore target marker reader is required"))
	}
	material, err := service.ReadTargetMarker(
		targetDeployment.BackupStorage.BindingKind,
		targetDeployment.BackupStorage.Path,
	)
	if err != nil {
		releaseOnError()
		if errors.Is(err, os.ErrNotExist) {
			return nil, NewFailure(FailureTargetMarkerMissing, fmt.Errorf("read restore target marker: %w", err))
		}
		return nil, NewFailure(FailureTargetMarkerInvalid, fmt.Errorf("read restore target marker: %w", err))
	}
	targetGenerationID, err := AdmitRestoreTargetMarker(material, purpose, TargetBindingDigestsFor(targetDeployment), service.now())
	if err != nil {
		releaseOnError()
		return nil, NewFailure(FailureTargetMarkerInvalid, err)
	}
	if err := admission.AssertHeld(); err != nil {
		releaseOnError()
		return nil, NewFailure(FailureTargetServingTraffic, err)
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

func (service Service) finishJournalAndAudit(ctx context.Context, pool PostgresPool, parsed operationRequest, outcome Result, operationErr *error) {
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
		return
	}
	evidenceCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), terminalEvidenceTimeout)
	defer cancel()
	if err := repository.AppendCompletion(evidenceCtx, RecoveryCompletionRecord{
		OperationID:               parsed.OperationID,
		Operation:                 parsed.Operation,
		AttemptID:                 parsed.AttemptID,
		StartedAt:                 parsed.StartedAt,
		CompletedAt:               service.now(),
		Result:                    result,
		BackupSetID:               outcome.BackupSetID,
		ConsistencyPointAt:        outcome.ConsistencyPointAt,
		ArtifactCounts:            ArtifactCountsFor(outcome.ArtifactRefs),
		ErrorCode:                 errorCode,
		ErrorReason:               reasonCode,
		GraphProjectionCompletion: outcome.graphProjectionCompletion,
	}); err != nil {
		replaceWithJournalFailure(operationErr, parsed.Operation, err)
	}
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
	targetGenerationID uuid.UUID,
) (Result, bool, error) {
	repository, err := service.evidenceRepository(pool)
	if err != nil {
		return Result{}, false, err
	}
	reader, ok := repository.(RecoveryEvidenceReplayReader)
	if !ok {
		return Result{}, false, nil
	}
	record, err := reader.FindSuccessfulCompletion(ctx, parsed.OperationID, parsed.Operation, parsed.AttemptID, backupSet.BackupSetID)
	if err != nil || record == nil {
		return Result{}, false, err
	}
	completion := record.GraphProjectionCompletion
	if completion == nil || completion.TargetGenerationID != targetGenerationID ||
		completion.RestoreOperationID != parsed.OperationID || completion.BackupSetID != backupSet.BackupSetID ||
		!completion.ConsistencyPointAt.Equal(backupSet.ConsistencyPointAt) ||
		!generationIdentity.AdmitsGraphCompletion(
			completion.RecoveryStateCatalogSHA256,
			completion.SourceRegistrySHA256,
			completion.ImplementationBindingSHA256,
		) ||
		!completion.ParticipantResult.ReadinessSatisfied() {
		return Result{}, false, nil
	}
	result := ResultForBackupSet(backupSet, "restore_operation", "cartulary.restore_operation.v1")
	result.graphProjectionCompletion = completion
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
		MechanismID:                recovery.VNextBackupMechanismID,
		CodecRegistrySHA256:        recovery.VNextCodecRegistrySHA256(),
		RecoveryStateCatalogSHA256: service.RecoveryStateCatalog.DigestSHA256(),
		DatabaseBindingSHA256:      recovery.SHA256String(rootBindingBasis(target.DatabaseStorage)),
		ObjectStoreBindingSHA256:   recovery.SHA256String(rootBindingBasis(target.ObjectStorage)),
		BackupStorageBindingSHA256: recovery.SHA256String(rootBindingBasis(source.BackupStorage)),
	}
	return basis, basis.Validate()
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

func (service Service) newProjectionServices(db postgres.DB) (restorecontract.ProjectionRebuilder, workbookprobe.Executor, error) {
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

func (service Service) newGraphProjectionRestore(db postgres.DB) (restorecontract.GraphProjectionParticipant, error) {
	if service.NewGraphProjectionRestore == nil {
		return nil, errors.New("operator recovery requires Graph Projection restore participant")
	}
	participant, err := service.NewGraphProjectionRestore(db)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, errors.New("operator recovery Graph Projection restore participant is unavailable")
	}
	return participant, nil
}

func (service Service) newEvidenceProvider(db postgres.DB) (recovery.EvidenceRecoveryProvider, error) {
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
