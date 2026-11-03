package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func restoreControlIdentity(request operationRequest, backup recovery.BackupSet, generation recovery.RecoveryGenerationIdentity, deployment Deployment, targetGeneration uuid.UUID, purpose string) recovery.RestoreTargetIdentity {
	return recovery.RestoreTargetIdentity{
		TargetGenerationID: targetGeneration, Purpose: purpose, Operation: string(request.Operation), OperationID: request.OperationID, AttemptID: request.AttemptID, BackupSetID: backup.BackupSetID,
		RecoveryStateCatalogSHA256: generation.RecoveryStateCatalogSHA256, CodecRegistrySHA256: generation.CodecRegistrySHA256, BindingDigests: deployment.TargetBindings(),
	}
}

func targetCompletionProof(record RecoveryCompletionRecord, identity recovery.RestoreTargetIdentity, generation recovery.RecoveryGenerationIdentity) (recovery.TargetCompletionProof, error) {
	canonical, err := EncodeRecoveryCompletion(record)
	if err != nil {
		return recovery.TargetCompletionProof{}, err
	}
	completion := record.RestoreGeneration
	if record.Result != ResultSucceeded || record.OperationID != identity.OperationID || string(record.Operation) != identity.Operation ||
		!sameAttemptID(record.AttemptID, identity.AttemptID) || record.BackupSetID == nil || *record.BackupSetID != identity.BackupSetID ||
		record.TargetBindings == nil || *record.TargetBindings != identity.BindingDigests || completion == nil ||
		completion.TargetGenerationID != identity.TargetGenerationID ||
		generation.RecoveryStateCatalogSHA256 != identity.RecoveryStateCatalogSHA256 || generation.CodecRegistrySHA256 != identity.CodecRegistrySHA256 ||
		!generation.AdmitsRestoreGeneration(*completion) {
		return recovery.TargetCompletionProof{}, recovery.ErrTargetIncomplete
	}
	digest := sha256.Sum256(canonical)
	return recovery.TargetCompletionProof{Identity: identity, EvidenceSHA256: hex.EncodeToString(digest[:])}, nil
}

func sameAttemptID(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func (service Service) finishAdmittedRestore(ctx context.Context, pool PostgresPool, request operationRequest, outcome Result, operationErr *error, fence *recovery.FencedRestore, identity recovery.RestoreTargetIdentity, generation recovery.RecoveryGenerationIdentity) {
	record, written := service.finishJournalAndAudit(ctx, pool, request, outcome, operationErr)
	if !written || operationErr == nil || *operationErr != nil {
		return
	}
	proof, err := targetCompletionProof(record, identity, generation)
	if err == nil {
		err = fence.Complete(ctx, proof)
	}
	if err != nil {
		replaceWithJournalFailure(operationErr, request.Operation, err)
	}
}

// Completed or failed operation identities cannot start another payload pass,
// including after a disposable reset has cleared the old control tuple.
func (service Service) requireUnusedRestoreOperation(ctx context.Context, pool PostgresPool, request operationRequest) error {
	repository, err := service.evidenceRepository(pool)
	if err != nil {
		return err
	}
	record, err := repository.FindCompletion(ctx, request.OperationID, request.Operation, request.AttemptID)
	if err != nil {
		return err
	}
	if record != nil {
		return recovery.ErrTargetIncomplete
	}
	return nil
}

// Preserve the admitted identity even when a named return replaces the outcome
// after a payload failure. Call only once durable admission has been journaled.
func resultForAdmittedRestore(result Result, identity recovery.RestoreTargetIdentity, backup recovery.BackupSet) Result {
	result.restoreGeneration = &recovery.RestoreGenerationEvidence{TargetGenerationID: identity.TargetGenerationID, RecoveryStateCatalogSHA256: identity.RecoveryStateCatalogSHA256, CodecRegistrySHA256: identity.CodecRegistrySHA256}
	bindings := identity.BindingDigests
	result.targetBindings = &bindings
	backupID, point := backup.BackupSetID, backup.ConsistencyPointAt
	result.BackupSetID, result.ConsistencyPointAt = &backupID, &point
	return result
}
