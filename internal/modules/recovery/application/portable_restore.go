package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

const PortableRestoreIntentSchemaID = "cartulary.portable_restore_intent.v1"
const portableJournalMaximumBytes = 16 << 20

// PortableJournal owns confined, exclusive files outside restored database state.
// Recovery owns their authenticated contents and replay decisions.
type PortableJournal interface {
	ReadArtifact(context.Context, string, int64) ([]byte, error)
	WriteArtifact(context.Context, string, []byte, string) (recovery.BackupArtifactProof, error)
	Close() error
}

type PortableRestoreIntent struct {
	BackupSetID        uuid.UUID            `json:"backup_set_id"`
	ConsistencyPointAt time.Time            `json:"consistency_point_at"`
	SchemaID           string               `json:"schema_id"`
	OperationID        uuid.UUID            `json:"operation_id"`
	TransferSHA256     string               `json:"transfer_sha256"`
	TargetGenerationID uuid.UUID            `json:"target_generation_id"`
	TargetBindings     TargetBindingDigests `json:"target_binding_digests"`
	StartedAt          time.Time            `json:"started_at"`
}

type portableReadiness func(context.Context, recovery.RestoreResult) error

func (f portableReadiness) MarkRestoreReady(ctx context.Context, result recovery.RestoreResult) error {
	return f(ctx, result)
}

func (service Service) RestoreBundle(ctx context.Context, request RestoreBundleRequest, progress ProgressSink) (outcome Result, operationErr error) {
	ReportProgress(progress, "preflight", 0, nil)
	if request.OperationID == uuid.Nil || service.OpenTransferDirectory == nil || service.NewTransferDirectory == nil || service.OpenPortableJournal == nil || service.LoadRecoveryKey == nil || service.ReleaseIdentity == nil || service.ReadTargetMarker == nil || service.ProjectFailureEvidence == nil {
		return Result{}, NewFailure(FailureLocalConfigInvalid, errors.New("portable restore composition unavailable"))
	}
	deployment, err := service.loadDeployment(request.TargetConfigPath)
	if err != nil {
		return Result{}, NewFailure(FailureLocalConfigInvalid, err)
	}
	if deployment.BackupStorage.BindingKind != "filesystem_root" {
		return Result{}, NewFailure(FailureLocalConfigInvalid, ErrTargetMarkerRequiresFilesystemStorage)
	}
	// These mutable owner roots must never contain each other or the transfer input.
	roots := []RootBinding{deployment.BackupStorage, deployment.ReferencePackStorage, deployment.ExportOutputs, deployment.ObjectStorage}
	inputRoot := RootBinding{BindingKind: "filesystem_root", Path: request.BundleDirectory}
	for i, root := range roots {
		if restoreRootsOverlap(root, inputRoot) {
			return Result{}, NewFailure(FailureSameObjectStoreBinding, recovery.ErrInvalidTransfer)
		}
		for _, other := range roots[i+1:] {
			if restoreRootsOverlap(root, other) {
				return Result{}, NewFailure(FailureSameObjectStoreBinding, recovery.ErrInvalidTransfer)
			}
		}
	}
	key, err := service.LoadRecoveryKey()
	if err != nil {
		return Result{}, classifyConfigOrSecretFailure(err)
	}
	release, err := service.ReleaseIdentity()
	if err != nil {
		return Result{}, NewFailure(FailureTransferReleaseMismatch, err)
	}
	input, err := service.OpenTransferDirectory(request.BundleDirectory)
	if err != nil {
		return Result{}, NewFailure(FailureTransferInvalid, err)
	}
	defer input.Close()
	captured, err := service.NewTransferDirectory(filepath.Join(deployment.BackupStorage.Path, "portable-input-"+uuid.NewString()))
	if err != nil {
		return Result{}, NewFailure(FailureTransferInvalid, err)
	}
	defer func() {
		if err := captured.Close(); err != nil && operationErr == nil {
			operationErr = NewFailure(FailureRestoreJournalWrite, err)
		}
	}()
	ReportProgress(progress, "artifact_validate", 0, nil)
	catalog := recovery.NewBackupCatalog(nil, nil, service.ExtensionBackups, service.RecoveryStateCatalog)
	manifest, digest, err := catalog.CaptureTransfer(ctx, input, captured, key, release)
	if err != nil {
		if errors.Is(err, recovery.ErrTransferReleaseMismatch) {
			return Result{}, NewFailure(FailureTransferReleaseMismatch, err)
		}
		return Result{}, NewFailure(FailureTransferInvalid, err)
	}
	backup := manifest.Selection.BackupSet()
	outcome = ResultForStoredBackupSet(backup)
	if backup.BackupSetID != request.ConfirmedBackupSet {
		return outcome, NewFailure(FailureConfirmationMismatch, errors.New("confirmed selection differs"))
	}
	if err := admitPortableAge(backup, request.AcknowledgedStaleBackup, service.now()); err != nil {
		return outcome, err
	}
	storage, err := recovery.NewEncryptedBackupStorage(captured, key)
	if err != nil {
		return outcome, NewFailure(FailureTransferInvalid, err)
	}
	// Only target services are opened, after the immutable encrypted input passes.
	pool, err := service.setupPostgres(ctx, deployment)
	if err != nil {
		return outcome, NewFailure(FailureLocalConfigInvalid, err)
	}
	defer pool.Close()
	unlock, err := service.acquireOperationLock(ctx, pool)
	if err != nil {
		return outcome, err
	}
	defer unlock()
	if service.NewTargetAdmission == nil {
		return outcome, NewFailure(FailureTargetServingTraffic, errors.New("target admission unavailable"))
	}
	admission, err := service.NewTargetAdmission(ctx, pool, minPositiveDuration(deployment.ServingLeaseAcquireTimeout, RestoreTargetServingLeaseAcquireMax), deployment.ServingLeaseLossDetection)
	if err != nil {
		return outcome, NewFailure(FailureTargetServingTraffic, err)
	}
	defer func() { releaseTargetAdmission(admission, deployment.ServingLeaseLossDetection) }()
	journal, err := service.OpenPortableJournal(deployment.BackupStorage.Path)
	if err != nil {
		return outcome, NewFailure(FailureRestoreJournalWrite, err)
	}
	defer journal.Close()
	bindings := TargetBindingDigestsFor(deployment)
	intentBody, err := readPortableRecord(admission.Context(), journal, key, request.OperationID, "intent")
	if err == nil {
		intent, err := decodePortableIntent(intentBody)
		if err != nil || intent.OperationID != request.OperationID || intent.TransferSHA256 != digest || intent.TargetBindings != bindings || intent.BackupSetID != backup.BackupSetID || !intent.ConsistencyPointAt.Equal(backup.ConsistencyPointAt) {
			return outcome, NewFailure(FailureTargetMarkerInvalid, recovery.ErrInvalidTransfer)
		}
		// A terminal retry checks the original admission instant; expiry cannot make
		// the same completed target look pristine or create a new generation.
		material, err := service.ReadTargetMarker(deployment.BackupStorage.BindingKind, deployment.BackupStorage.Path)
		if err != nil {
			return outcome, NewFailure(FailureTargetMarkerInvalid, err)
		}
		generation, err := AdmitRestoreTargetMarker(material, RestoreTargetPurpose, bindings, intent.StartedAt)
		if err != nil || generation != intent.TargetGenerationID {
			return outcome, NewFailure(FailureTargetMarkerInvalid, recovery.ErrInvalidTransfer)
		}
		body, err := readPortableRecord(admission.Context(), journal, key, request.OperationID, "completion")
		if err != nil {
			return outcome, NewFailure(FailureTargetDatabaseNotFresh, errors.New("indeterminate portable restore requires a fresh target"))
		}
		record, err := DecodeRecoveryCompletion(body)
		if err != nil || record.OperationID != intent.OperationID || record.Operation != OperationRestoreBundle || !record.StartedAt.Equal(intent.StartedAt) || record.BackupSetID == nil || *record.BackupSetID != backup.BackupSetID || record.TargetBindings == nil || *record.TargetBindings != bindings {
			return outcome, NewFailure(FailureRestoreJournalWrite, recovery.ErrInvalidTransfer)
		}
		if record.Result != ResultSucceeded {
			return outcome, NewFailure(FailureTargetDatabaseNotFresh, errors.New("failed portable restore requires a fresh target"))
		}
		completion := record.GraphProjectionCompletion
		generationIdentity, err := catalogForCaptured(service, storage).RecoveryGenerationIdentity(ctx, backup)
		if err != nil || completion == nil || completion.TargetGenerationID != generation || !generationIdentity.AdmitsGraphCompletion(completion.RecoveryStateCatalogSHA256, completion.SourceRegistrySHA256, completion.ImplementationBindingSHA256) || !completion.ParticipantResult.ReadinessSatisfied() {
			return outcome, NewFailure(FailureRestoreJournalWrite, recovery.ErrInvalidTransfer)
		}
		if err := admission.AssertHeld(); err != nil {
			return outcome, NewFailure(FailureTargetServingTraffic, err)
		}
		if err := service.publishPortableCompletion(admission.Context(), pool, record); err != nil {
			return outcome, NewFailure(FailureRestoreJournalWrite, err)
		}
		ReportProgress(progress, "journal_write", 1, IntPtr(1))
		ReportProgress(progress, "finalize", 1, IntPtr(1))
		return portableResult(backup, record), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return outcome, NewFailure(FailureRestoreJournalWrite, err)
	}
	objects, err := service.setupObjectStore(admission.Context(), deployment)
	if err != nil {
		return outcome, NewFailure(FailureLocalConfigInvalid, err)
	}
	defer objects.Close()
	admission, err = service.prepareAdmittedTarget(deployment, pool, RestoreTargetPurpose, objects, admission)
	if err != nil {
		return outcome, err
	}
	generation, ok := admittedTargetGenerationID(admission)
	if !ok {
		return outcome, NewFailure(FailureTargetMarkerInvalid, recovery.ErrInvalidTransfer)
	}
	target, err := service.restoreTarget(deployment, pool, objects, request.OperationID, generation)
	if err != nil {
		return outcome, NewFailure(FailureRestoreProjectionRebuild, err)
	}
	defer target.ReferencePacks.Close()
	defer target.ExportOutputs.Close()
	// A still-valid proof is not evidence that target data remained empty.
	if err := recovery.InspectPristineRestoreTarget(admission.Context(), target, service.ExtensionBackups, service.RecoveryStateCatalog); err != nil {
		return outcome, NewFailure(FailureTargetDatabaseNotFresh, err)
	}
	intent := PortableRestoreIntent{BackupSetID: backup.BackupSetID, ConsistencyPointAt: backup.ConsistencyPointAt, SchemaID: PortableRestoreIntentSchemaID, OperationID: request.OperationID, TransferSHA256: digest, TargetGenerationID: generation, TargetBindings: bindings, StartedAt: service.now()}
	body, err := canonicaljson.Marshal(intent)
	if err != nil {
		return outcome, NewFailure(FailureRestoreJournalWrite, err)
	}
	if err := writePortableRecord(admission.Context(), journal, key, request.OperationID, "intent", body); err != nil {
		return outcome, NewFailure(FailureRestoreJournalWrite, err)
	}
	var terminal *RecoveryCompletionRecord
	defer func() {
		if operationErr == nil || terminal != nil {
			return
		}
		kind, ok := FailureKindOf(operationErr)
		if !ok {
			kind = FailureRestoreInvariantCheck
		}
		code, reason := service.ProjectFailureEvidence(kind)
		record := RecoveryCompletionRecord{OperationID: request.OperationID, Operation: OperationRestoreBundle, StartedAt: intent.StartedAt, CompletedAt: service.now(), Result: ResultFailed, BackupSetID: outcome.BackupSetID, ConsistencyPointAt: outcome.ConsistencyPointAt, ArtifactCounts: ArtifactCountsFor(outcome.ArtifactRefs), ErrorCode: &code, ErrorReason: &reason, TargetBindings: &bindings}
		body, encodeErr := EncodeRecoveryCompletion(record)
		if encodeErr == nil {
			terminalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), terminalEvidenceTimeout)
			defer cancel()
			encodeErr = writePortableRecord(terminalCtx, journal, key, request.OperationID, "completion", body)
		}
		if encodeErr != nil {
			operationErr = NewFailure(FailureRestoreJournalWrite, encodeErr)
		}
	}()
	target.Readiness = portableReadiness(func(ctx context.Context, result recovery.RestoreResult) error {
		if err := admission.AssertHeld(); err != nil {
			return err
		}
		completed := ResultForBackupSet(result.BackupSet, "restore_operation", "cartulary.restore_operation.v1")
		record := RecoveryCompletionRecord{OperationID: request.OperationID, Operation: OperationRestoreBundle, StartedAt: intent.StartedAt, CompletedAt: service.now(), Result: ResultSucceeded, BackupSetID: completed.BackupSetID, ConsistencyPointAt: completed.ConsistencyPointAt, ArtifactCounts: ArtifactCountsFor(completed.ArtifactRefs), GraphProjectionCompletion: result.GraphProjectionCompletion, TargetBindings: &bindings}
		body, err := EncodeRecoveryCompletion(record)
		if err != nil {
			return err
		}
		if err := writePortableRecord(ctx, journal, key, request.OperationID, "completion", body); err != nil {
			return err
		}
		terminal = &record
		return nil
	})
	for _, phase := range []string{"postgres_restore", "object_restore", "projection_rebuild", "invariant_check"} {
		ReportProgress(progress, phase, 0, nil)
	}
	_, err = recovery.NewSelectedRestoreRunner(storage, service.ExtensionBackups, service.RecoveryStateCatalog).RestoreBackupSet(admission.Context(), target, backup)
	if err != nil {
		return outcome, classifyRestoreFailure(err, false)
	}
	if terminal == nil {
		return outcome, NewFailure(FailureRestoreJournalWrite, errors.New("portable completion missing"))
	}
	if err := admission.AssertHeld(); err != nil {
		return outcome, NewFailure(FailureTargetServingTraffic, err)
	}
	ReportProgress(progress, "journal_write", 0, nil)
	if err := service.publishPortableCompletion(admission.Context(), pool, *terminal); err != nil {
		return outcome, NewFailure(FailureRestoreJournalWrite, err)
	}
	ReportProgress(progress, "finalize", 1, IntPtr(1))
	return portableResult(backup, *terminal), nil
}

func catalogForCaptured(service Service, storage recovery.BackupStorage) *recovery.BackupCatalog {
	return recovery.NewBackupCatalog(nil, storage, service.ExtensionBackups, service.RecoveryStateCatalog)
}
func portableResult(backup recovery.BackupSet, record RecoveryCompletionRecord) Result {
	result := ResultForBackupSet(backup, "restore_operation", "cartulary.restore_operation.v1")
	result.graphProjectionCompletion, result.targetBindings = record.GraphProjectionCompletion, record.TargetBindings
	result.StartedAt, result.CompletedAt = &record.StartedAt, &record.CompletedAt
	return result
}
func (service Service) publishPortableCompletion(ctx context.Context, pool PostgresPool, record RecoveryCompletionRecord) error {
	repository, err := service.evidenceRepository(pool)
	if err != nil {
		return err
	}
	reader, ok := repository.(RecoveryEvidenceReplayReader)
	if !ok {
		return errors.New("target evidence requires durable replay lookup")
	}
	existing, err := reader.FindSuccessfulCompletion(ctx, record.OperationID, OperationRestoreBundle, nil, *record.BackupSetID)
	if err != nil {
		return err
	}
	if existing != nil {
		a, err := EncodeRecoveryCompletion(*existing)
		b, otherErr := EncodeRecoveryCompletion(record)
		if err != nil || otherErr != nil || !bytes.Equal(a, b) {
			return recovery.ErrInvalidTransfer
		}
		return nil
	}
	return repository.AppendCompletion(ctx, record)
}
func portableJournalAAD(operation uuid.UUID, kind string) string {
	return PortableRestoreIntentSchemaID + ":" + operation.String() + ":" + kind
}
func readPortableRecord(ctx context.Context, journal PortableJournal, key recovery.RecoveryEncryptionKey, operation uuid.UUID, kind string) ([]byte, error) {
	body, err := journal.ReadArtifact(ctx, operation.String()+"/"+kind+".sealed", portableJournalMaximumBytes)
	if err != nil {
		return nil, err
	}
	var envelope recovery.OperatorRecoveryJournalEnvelope
	if err := decodePortableStrict(body, &envelope); err != nil {
		return nil, err
	}
	return recovery.DecryptOperatorRecoveryJournalPayload(key, portableJournalAAD(operation, kind), envelope)
}
func writePortableRecord(ctx context.Context, journal PortableJournal, key recovery.RecoveryEncryptionKey, operation uuid.UUID, kind string, body []byte) error {
	envelope, err := recovery.EncryptOperatorRecoveryJournalPayload(key, uuid.NewString(), portableJournalAAD(operation, kind), body)
	if err != nil {
		return err
	}
	sealed, err := canonicaljson.Marshal(envelope)
	if err != nil || len(sealed) > portableJournalMaximumBytes {
		return recovery.ErrInvalidTransfer
	}
	if _, err := journal.WriteArtifact(ctx, operation.String()+"/"+kind+".sealed", sealed, "application/octet-stream"); err != nil {
		return err
	}
	readback, err := readPortableRecord(ctx, journal, key, operation, kind)
	if err != nil || !bytes.Equal(readback, body) {
		return recovery.ErrInvalidTransfer
	}
	return nil
}
func decodePortableStrict(body []byte, value any) error {
	admitted, err := canonicaljson.Canonicalize(body)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	projected, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(admitted, projected) {
		return recovery.ErrInvalidTransfer
	}
	return nil
}
func decodePortableIntent(body []byte) (PortableRestoreIntent, error) {
	var intent PortableRestoreIntent
	if err := decodePortableStrict(body, &intent); err != nil {
		return intent, err
	}
	if intent.SchemaID != PortableRestoreIntentSchemaID || intent.OperationID == uuid.Nil || intent.TargetGenerationID == uuid.Nil || intent.StartedAt.IsZero() || intent.BackupSetID == uuid.Nil || intent.ConsistencyPointAt.IsZero() || !isLowerSHA256(intent.TransferSHA256) {
		return intent, recovery.ErrInvalidTransfer
	}
	return intent, nil
}

func admitPortableAge(backup recovery.BackupSet, acknowledged uuid.UUID, now time.Time) error {
	age := now.Sub(backup.ConsistencyPointAt)
	if age < 0 {
		return NewFailure(FailureTransferInvalid, recovery.ErrInvalidTransfer)
	}
	if (age > 24*time.Hour && acknowledged != backup.BackupSetID) || (acknowledged != uuid.Nil && acknowledged != backup.BackupSetID) {
		return NewFailure(FailureStaleBackupUnacknowledged, errors.New("exact stale selection acknowledgement required"))
	}
	return nil
}
