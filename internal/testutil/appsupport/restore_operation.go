package appsupport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type CurrentRestoreParams struct {
	Source   dbport.DB
	Target   recovery.RestoreTarget
	Runner   *recovery.RestoreRunner
	Catalog  *recovery.BackupCatalog
	Backup   recovery.BackupSet
	Key      recovery.RecoveryEncryptionKey
	Bindings recovery.TargetBindingDigests
}

// RestoreCurrentBackup composes the actual admission, mutation and durable
// completion ports for owned test installations. It does not reset a target,
// encode a snapshot, manufacture a control row or execute profile callbacks.
// The caller supplies the resource bindings used by any subsequent server.
func RestoreCurrentBackup(ctx context.Context, params CurrentRestoreParams) (result recovery.RestoreResult, resultErr error) {
	pool, ok := params.Target.Postgres.(postgres.AdmittedPool)
	if !ok || params.Runner == nil || params.Catalog == nil {
		return result, fmt.Errorf("current restore fixture requires an admitted Recovery pool, runner and catalog")
	}
	admission, err := recoveryassembly.AcquireTargetServingAdmission(ctx, pool, 5*time.Second, 5*time.Second)
	if err != nil {
		return result, err
	}
	defer func() {
		settle, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resultErr = errors.Join(resultErr, admission.Release(settle))
	}()
	generation, err := params.Catalog.RecoveryGenerationIdentity(ctx, params.Backup)
	if err != nil {
		return result, err
	}
	identity := recovery.RestoreTargetIdentity{TargetGenerationID: uuid.New(), Purpose: "restore_target", Operation: "restore_latest", OperationID: uuid.New(), BackupSetID: params.Backup.BackupSetID, RecoveryStateCatalogSHA256: generation.RecoveryStateCatalogSHA256, CodecRegistrySHA256: generation.CodecRegistrySHA256, BindingDigests: params.Bindings}
	sourceJournal, err := recoveryassembly.NewRecoveryEvidenceRepository(params.Source, func() (recovery.RecoveryEncryptionKey, error) { return params.Key, nil })
	if err != nil {
		return result, err
	}
	started := time.Now().UTC()
	fence, err := params.Runner.AdmitRestore(ctx, admission, params.Target, params.Backup, identity, func(work context.Context) error {
		return sourceJournal.AppendAdmission(work, application.RecoveryAdmissionRecord{OperationID: identity.OperationID, Operation: application.OperationRestoreLatest, StartedAt: started, BackupSetID: &identity.BackupSetID, ArtifactKinds: []string{}})
	})
	if err != nil {
		return result, err
	}
	defer func() { resultErr = errors.Join(resultErr, fence.Close()) }()
	result, err = params.Runner.RestoreBackupSet(ctx, fence.Target(), params.Backup)
	if err != nil {
		return result, err
	}
	targetJournal, err := recoveryassembly.NewRecoveryEvidenceRepository(pool, func() (recovery.RecoveryEncryptionKey, error) { return params.Key, nil })
	if err != nil {
		return result, err
	}
	point := params.Backup.ConsistencyPointAt
	record := application.RecoveryCompletionRecord{OperationID: identity.OperationID, Operation: application.OperationRestoreLatest, StartedAt: started, CompletedAt: time.Now().UTC(), Result: application.ResultSucceeded, BackupSetID: &identity.BackupSetID, ConsistencyPointAt: &point, ArtifactCounts: []application.ArtifactCount{}, RestoreGeneration: result.RestoreGeneration, TargetBindings: &params.Bindings}
	if err := targetJournal.AppendCompletion(ctx, record); err != nil {
		return result, err
	}
	body, err := application.EncodeRecoveryCompletion(record)
	if err != nil {
		return result, err
	}
	digest := sha256.Sum256(body)
	if err := fence.Complete(ctx, recovery.TargetCompletionProof{Identity: identity, EvidenceSHA256: hex.EncodeToString(digest[:])}); err != nil {
		return result, err
	}
	return result, nil
}
