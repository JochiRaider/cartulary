package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

type Operation string

const (
	OperationBackupInspectLatest Operation = "backup_inspect_latest"
	OperationBackupCreate        Operation = "backup_create"
	OperationRestoreLatest       Operation = "restore_latest"
	OperationRestoreVerifyLatest Operation = "restore_verify_latest"
	OperationRestoreVerifyDue    Operation = "restore_verify_due"
)

type BackupInspectLatestRequest struct {
	OperationID      uuid.UUID
	SourceConfigPath string
}

type BackupCreateRequest struct {
	OperationID      uuid.UUID
	SourceConfigPath string
}

type RestoreLatestRequest struct {
	OperationID        uuid.UUID
	SourceConfigPath   string
	TargetConfigPath   string
	ConfirmedBackupSet uuid.UUID
}

type RestoreVerifyLatestRequest struct {
	OperationID      uuid.UUID
	SourceConfigPath string
	TargetConfigPath string
}

type RestoreVerifyDueRequest struct {
	OperationID      uuid.UUID
	SourceConfigPath string
	TargetConfigPath string
	AttemptTimeout   time.Duration
}

type Facade interface {
	BackupInspectLatest(context.Context, BackupInspectLatestRequest, ProgressSink) (Result, error)
	BackupCreate(context.Context, BackupCreateRequest, ProgressSink) (Result, error)
	RestoreLatest(context.Context, RestoreLatestRequest, ProgressSink) (Result, error)
	RestoreVerifyLatest(context.Context, RestoreVerifyLatestRequest, ProgressSink) (Result, error)
	RestoreVerifyDue(context.Context, RestoreVerifyDueRequest, ProgressSink) (Result, error)
}

type ResultStatus string

const (
	ResultSucceeded ResultStatus = "succeeded"
	ResultNoOp      ResultStatus = "no_op"
	ResultFailed    ResultStatus = "failed"
)

type Result struct {
	BackupSetID               *uuid.UUID
	ConsistencyPointAt        *time.Time
	ArtifactRefs              []ArtifactRef
	Status                    ResultStatus
	graphProjectionCompletion *GraphProjectionCompletionEvidence
}

type ArtifactRef struct {
	Kind        string
	SchemaID    string
	RefID       string
	BackupSetID *uuid.UUID
}

type Progress struct {
	Phase     string
	Completed int
	Total     *int
}

type ProgressSink interface {
	ReportProgress(Progress)
}

type ProgressSinkFunc func(Progress)

func (sink ProgressSinkFunc) ReportProgress(progress Progress) {
	if sink != nil {
		sink(progress)
	}
}

func ReportProgress(sink ProgressSink, phase string, completed int, total *int) {
	if sink == nil {
		return
	}
	sink.ReportProgress(Progress{Phase: phase, Completed: completed, Total: total})
}

func IntPtr(value int) *int {
	return &value
}

type FailureKind string

const (
	FailureConfirmationMismatch          FailureKind = "confirmation_mismatch"
	FailureLocalConfigInvalid            FailureKind = "local_config_invalid"
	FailureSecretReferenceMissing        FailureKind = "secret_reference_missing"
	FailureSecretReferenceUnresolved     FailureKind = "secret_reference_unresolved"
	FailureRecoveryKeyInvalid            FailureKind = "recovery_key_invalid"
	FailureNoSuccessfulRetainedBackup    FailureKind = "no_successful_retained_backup"
	FailureSelectedBackupNotRetained     FailureKind = "selected_backup_not_retained"
	FailureArtifactMissing               FailureKind = "artifact_missing"
	FailureIntegrityProofMissing         FailureKind = "integrity_proof_missing"
	FailureChecksumMismatch              FailureKind = "checksum_mismatch"
	FailureAttestationInvalid            FailureKind = "attestation_invalid"
	FailureSameDatabaseBinding           FailureKind = "same_database_binding"
	FailureSameObjectStoreBinding        FailureKind = "same_object_store_binding"
	FailureTargetDatabaseNotFresh        FailureKind = "target_database_not_fresh"
	FailureTargetObjectNamespaceNotFresh FailureKind = "target_object_namespace_not_fresh"
	FailureTargetServingTraffic          FailureKind = "target_serving_traffic"
	FailureTargetMarkerMissing           FailureKind = "target_marker_missing"
	FailureTargetMarkerInvalid           FailureKind = "target_marker_invalid"
	FailureOperationLockUnavailable      FailureKind = "operation_lock_unavailable"
	FailureTimeoutElapsed                FailureKind = "timeout_elapsed"
	FailureBackupPostgres                FailureKind = "backup_postgres_failed"
	FailureBackupObject                  FailureKind = "backup_object_failed"
	FailureBackupIntegrityProof          FailureKind = "backup_integrity_proof_failed"
	FailureBackupArtifactReadback        FailureKind = "backup_artifact_readback_failed"
	FailureBackupAttestationWrite        FailureKind = "backup_attestation_write_failed"
	FailureBackupPublication             FailureKind = "backup_publication_failed"
	FailureBackupJournalWrite            FailureKind = "backup_journal_write_failed"
	FailureRestorePostgres               FailureKind = "restore_postgres_failed"
	FailureRestoreObject                 FailureKind = "restore_object_failed"
	FailureRestoreProjectionRebuild      FailureKind = "restore_projection_rebuild_failed"
	FailureRestoreInvariantCheck         FailureKind = "restore_invariant_check_failed"
	FailureRestoreJournalWrite           FailureKind = "restore_journal_write_failed"
	FailureVerificationPostgres          FailureKind = "verification_postgres_failed"
	FailureVerificationObject            FailureKind = "verification_object_failed"
	FailureVerificationProjectionRebuild FailureKind = "verification_projection_rebuild_failed"
	FailureVerificationInvariantCheck    FailureKind = "verification_invariant_check_failed"
	FailureVerificationWorkbookProbe     FailureKind = "verification_workbook_probe_failed"
	FailureVerificationAttestationUpdate FailureKind = "verification_attestation_update_failed"
	FailureVerificationJournalWrite      FailureKind = "verification_journal_write_failed"
)

var allFailureKinds = []FailureKind{
	FailureConfirmationMismatch,
	FailureLocalConfigInvalid,
	FailureSecretReferenceMissing,
	FailureSecretReferenceUnresolved,
	FailureRecoveryKeyInvalid,
	FailureNoSuccessfulRetainedBackup,
	FailureSelectedBackupNotRetained,
	FailureArtifactMissing,
	FailureIntegrityProofMissing,
	FailureChecksumMismatch,
	FailureAttestationInvalid,
	FailureSameDatabaseBinding,
	FailureSameObjectStoreBinding,
	FailureTargetDatabaseNotFresh,
	FailureTargetObjectNamespaceNotFresh,
	FailureTargetServingTraffic,
	FailureTargetMarkerMissing,
	FailureTargetMarkerInvalid,
	FailureOperationLockUnavailable,
	FailureTimeoutElapsed,
	FailureBackupPostgres,
	FailureBackupObject,
	FailureBackupIntegrityProof,
	FailureBackupArtifactReadback,
	FailureBackupAttestationWrite,
	FailureBackupPublication,
	FailureBackupJournalWrite,
	FailureRestorePostgres,
	FailureRestoreObject,
	FailureRestoreProjectionRebuild,
	FailureRestoreInvariantCheck,
	FailureRestoreJournalWrite,
	FailureVerificationPostgres,
	FailureVerificationObject,
	FailureVerificationProjectionRebuild,
	FailureVerificationInvariantCheck,
	FailureVerificationWorkbookProbe,
	FailureVerificationAttestationUpdate,
	FailureVerificationJournalWrite,
}

func AllFailureKinds() []FailureKind {
	return append([]FailureKind(nil), allFailureKinds...)
}

func (kind FailureKind) Valid() bool {
	for _, candidate := range allFailureKinds {
		if kind == candidate {
			return true
		}
	}
	return false
}

type Failure struct {
	Kind  FailureKind
	Cause error
}

func (failure *Failure) Error() string {
	if failure == nil {
		return "recovery operation failed"
	}
	if failure.Cause == nil {
		return fmt.Sprintf("recovery operation failed: %s", failure.Kind)
	}
	return failure.Cause.Error()
}

func (failure *Failure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.Cause
}

func NewFailure(kind FailureKind, cause error) error {
	if !kind.Valid() {
		panic(fmt.Sprintf("invalid recovery failure kind %q", kind))
	}
	if cause == nil {
		cause = errors.New(string(kind))
	}
	return &Failure{Kind: kind, Cause: cause}
}

func FailureKindOf(err error) (FailureKind, bool) {
	var failure *Failure
	if !errors.As(err, &failure) || failure == nil || !failure.Kind.Valid() {
		return "", false
	}
	return failure.Kind, true
}

func EnsureFailure(defaultKind FailureKind, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := FailureKindOf(err); ok {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return NewFailure(FailureTimeoutElapsed, err)
	}
	return NewFailure(defaultKind, err)
}

func ResultForBackupSet(backupSet recovery.BackupSet, kind string, schemaID string) Result {
	result := ResultForStoredBackupSet(backupSet)
	result.ArtifactRefs = append(result.ArtifactRefs, ArtifactRefFor(
		kind,
		schemaID,
		kind+":"+backupSet.BackupSetID.String(),
		result.BackupSetID,
	))
	return result
}

func ResultForStoredBackupSet(backupSet recovery.BackupSet) Result {
	if backupSet.BackupSetID == uuid.Nil {
		return Result{ArtifactRefs: []ArtifactRef{}}
	}
	backupSetID := backupSet.BackupSetID
	consistencyPointAt := backupSet.ConsistencyPointAt.UTC()
	return Result{
		BackupSetID:        &backupSetID,
		ConsistencyPointAt: &consistencyPointAt,
		ArtifactRefs:       []ArtifactRef{},
	}
}

func ResultForCandidate(backupSetID uuid.UUID, consistencyPointAt time.Time) Result {
	id := backupSetID
	at := consistencyPointAt.UTC()
	return Result{BackupSetID: &id, ConsistencyPointAt: &at, ArtifactRefs: []ArtifactRef{}}
}

func ArtifactRefFor(kind string, schemaID string, refID string, backupSetID *uuid.UUID) ArtifactRef {
	return ArtifactRef{Kind: kind, SchemaID: schemaID, RefID: refID, BackupSetID: backupSetID}
}
