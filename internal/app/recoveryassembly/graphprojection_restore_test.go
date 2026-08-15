package recoveryassembly

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestGraphRestoreAcceptanceGPRA18RecoveryAssemblyUsesNarrowParticipant_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "recovery-graph-restore-assembly")
	participant, err := NewGraphProjectionRestoreParticipant(db)
	if err != nil {
		t.Fatalf("construct Graph restore participant: %v", err)
	}
	catalog, err := CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatalf("construct Recovery state catalog: %v", err)
	}
	registry := restorecontract.CurrentGraphProjectionSourceRegistryRef()
	binding := restorecontract.CurrentGraphProjectionImplementationBinding()
	ctx := context.Background()
	result, err := participant.Rebuild(ctx, restorecontract.GraphProjectionRebuildRequest{
		Context:             ctx,
		RestoreOperationID:  uuid.MustParse("00000000-0000-0000-0000-000000009001"),
		RestoredSourceState: restorecontract.RestoredGraphProjectionSourceState{},
		BackupSetID:         uuid.MustParse("00000000-0000-0000-0000-000000009002"),
		ConsistencyPointAt:  time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
		TargetGenerationID:  uuid.MustParse("00000000-0000-0000-0000-000000009003"),
		RecoveryStateCatalog: restorecontract.GraphProjectionRecoveryCatalogRef{
			DigestSHA256: catalog.DigestSHA256(), AlgorithmID: restorecontract.GraphProjectionRestoreAlgorithmID,
			GraphTableIDs: restorecontract.GraphProjectionTableIDs(),
		},
		SourceRegistry: registry, ImplementationBinding: binding,
	})
	if err != nil {
		t.Fatalf("execute assembled Graph restore participant: %v result=%#v catalog=%s binding=%s", err, result, catalog.DigestSHA256(), binding.Binding.RecoveryStateCatalogSHA256)
	}
	if !result.ReadinessSatisfied() || len(result.RebuiltViews) != 0 {
		t.Fatalf("assembled Graph restore result mismatch: %#v", result)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(recoveryEvidenceTestKey)
	if err != nil {
		t.Fatalf("parse Recovery evidence key: %v", err)
	}
	repository, err := NewRecoveryEvidenceRepository(db, func() (recovery.RecoveryEncryptionKey, error) { return key, nil })
	if err != nil {
		t.Fatalf("construct Recovery evidence repository: %v", err)
	}
	postcondition := *result.PostconditionSHA256
	completion := &restorecontract.GraphProjectionCompletionEvidence{
		TargetGenerationID: uuid.MustParse(result.TargetGenerationID), RestoreOperationID: uuid.MustParse(result.RestoreOperationID),
		BackupSetID: uuid.MustParse("00000000-0000-0000-0000-000000009002"), ConsistencyPointAt: time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
		RecoveryStateCatalogSHA256: catalog.DigestSHA256(), SourceRegistrySHA256: result.SourceRegistrySHA256,
		ImplementationBindingSHA256: result.ImplementationBindingSHA256, PostconditionSHA256: postcondition, ParticipantResult: result,
	}
	startedAt := time.Date(2026, 5, 30, 1, 0, 0, 0, time.UTC)
	backupSetID := completion.BackupSetID
	consistencyPoint := completion.ConsistencyPointAt
	if err := repository.AppendCompletion(ctx, application.RecoveryCompletionRecord{
		OperationID: completion.RestoreOperationID, Operation: application.OperationRestoreLatest,
		StartedAt: startedAt, CompletedAt: startedAt.Add(time.Minute), Result: application.ResultSucceeded,
		BackupSetID: &backupSetID, ConsistencyPointAt: &consistencyPoint, ArtifactCounts: []application.ArtifactCount{},
		GraphProjectionCompletion: completion,
	}); err != nil {
		t.Fatalf("persist Graph terminal completion: %v", err)
	}
	reader, ok := repository.(application.RecoveryEvidenceReplayReader)
	if !ok {
		t.Fatal("production Recovery evidence repository lacks terminal replay reader")
	}
	replayed, err := reader.FindSuccessfulCompletion(
		ctx, completion.RestoreOperationID, application.OperationRestoreLatest, nil, backupSetID,
	)
	if err != nil || replayed == nil || replayed.GraphProjectionCompletion == nil ||
		replayed.GraphProjectionCompletion.PostconditionSHA256 != postcondition {
		t.Fatalf("replay protected Graph terminal completion: record=%#v err=%v", replayed, err)
	}
}
