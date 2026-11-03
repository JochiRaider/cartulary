package recoveryassembly

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"strings"
	"testing"
	"time"
)

func TestRecoveryStructuralGenerationEvidenceAndConflictingTerminal_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "structural-generation-evidence")
	catalog, err := CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(recoveryEvidenceTestKey)
	if err != nil {
		t.Fatalf("parse Recovery evidence key: %v", err)
	}
	repository, err := NewRecoveryEvidenceRepository(db, func() (recovery.RecoveryEncryptionKey, error) { return key, nil })
	if err != nil {
		t.Fatalf("construct Recovery evidence repository: %v", err)
	}
	completion := &recovery.RestoreGenerationEvidence{TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000009003"), RecoveryStateCatalogSHA256: catalog.DigestSHA256(), CodecRegistrySHA256: strings.Repeat("5", 64)}
	operationID := uuid.MustParse("00000000-0000-0000-0000-000000009001")
	startedAt := time.Date(2026, 5, 30, 1, 0, 0, 0, time.UTC)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000009002")
	consistencyPoint := time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC)
	if err := repository.AppendCompletion(ctx, application.RecoveryCompletionRecord{
		OperationID: operationID, Operation: application.OperationRestoreLatest,
		StartedAt: startedAt, CompletedAt: startedAt.Add(time.Minute), Result: application.ResultSucceeded,
		BackupSetID: &backupSetID, ConsistencyPointAt: &consistencyPoint, ArtifactCounts: []application.ArtifactCount{},
		RestoreGeneration: completion,
		TargetBindings:    &recovery.TargetBindingDigests{DatabaseSHA256: strings.Repeat("1", 64), ObjectStoreSHA256: strings.Repeat("2", 64), ReferencePackStorageSHA256: strings.Repeat("3", 64), ExportOutputsSHA256: strings.Repeat("4", 64)},
	}); err != nil {
		t.Fatalf("persist structural terminal generation: %v", err)
	}
	replayed, err := repository.FindCompletion(ctx, operationID, application.OperationRestoreLatest, nil)
	if err != nil || replayed == nil || replayed.RestoreGeneration == nil ||
		*replayed.RestoreGeneration != *completion {
		t.Fatalf("replay protected structural terminal generation: record=%#v err=%v", replayed, err)
	}
	conflicting := *replayed
	conflicting.Result = application.ResultFailed
	code, reason := "restore_failed", "invariant_check_failed"
	conflicting.ErrorCode, conflicting.ErrorReason = &code, &reason
	if err := repository.AppendCompletion(ctx, conflicting); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.FindCompletion(ctx, operationID, application.OperationRestoreLatest, nil); err == nil {
		t.Fatal("conflicting terminal history authorized replay")
	}
}
