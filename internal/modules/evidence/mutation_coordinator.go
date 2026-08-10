package evidence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type evidenceCreateTxResult struct {
	payload          map[string]any
	row              map[string]any
	recordID         uuid.UUID
	changeSetID      uuid.UUID
	changedFieldKeys []string
}

// evidenceMutationCoordinator orders cross-owner durable effects in a supplied
// transaction. Transaction begin, idempotency replay, commit, and HTTP response
// handling remain application/route concerns.
type evidenceMutationCoordinator struct {
	incidents     evidenceIncidentAdmissionPort
	source        evidenceSourceKernel
	revisions     revisionAppendPort
	collaboration collaboration.IntentAppender
}

func (coordinator evidenceMutationCoordinator) createTx(
	ctx context.Context,
	tx pgx.Tx,
	command WorkbookCreateCommand,
	createParams WorkbookCreateParams,
) (evidenceCreateTxResult, error) {
	request := command.Request
	if err := coordinator.incidents.EnsureOpenTx(ctx, tx, command.IncidentID); err != nil {
		return evidenceCreateTxResult{}, err
	}
	if err := validateEvidenceReferencesTx(ctx, tx, command.IncidentID, request.Values); err != nil {
		return evidenceCreateTxResult{}, err
	}
	now := command.Now.UTC()
	recordID, err := coordinator.source.createRecordTx(
		ctx,
		tx,
		command.IncidentID,
		command.Actor.ID,
		createParams,
		now,
	)
	if err != nil {
		return evidenceCreateTxResult{}, err
	}
	if createParams.InitialBlob != nil {
		if err := insertEvidenceCustodyEventTx(ctx, tx, evidenceCustodyEventParams{
			IncidentID:       command.IncidentID,
			EvidenceRecordID: recordID,
			CustodyEventType: "made_available",
			ActorUserID:      &command.Actor.ID,
			OccurredAt:       now,
			Metadata: map[string]any{
				"object_blob_id": createParams.InitialBlob.ObjectBlobID.String(),
			},
		}); err != nil {
			return evidenceCreateTxResult{}, err
		}
	}
	row, err := coordinator.source.refreshRowTx(ctx, tx, recordID)
	if err != nil {
		return evidenceCreateTxResult{}, err
	}
	afterSnapshot, err := coordinator.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return evidenceCreateTxResult{}, err
	}
	changeSetID, err := coordinator.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  command.IncidentID,
		ActorUserID: command.Actor.ID,
		Source:      command.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &command.RequestID,
		CreatedAt:   now,
	})
	if err != nil {
		return evidenceCreateTxResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := coordinator.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		RecordID:       recordID,
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return evidenceCreateTxResult{}, err
	}
	if err := coordinator.revisions.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams{
		ChangeSetID:   changeSetID,
		RecordID:      recordID,
		RowVersion:    1,
		AfterSnapshot: &afterSnapshot,
		LiveChange:    revisions.LiveRecordChange{AfterValue: row},
	}); err != nil {
		return evidenceCreateTxResult{}, err
	}
	changedFieldKeys := changedFieldKeys(nil, row)
	if err := appendEvidenceRecordChangeIntentsTx(
		ctx,
		tx,
		coordinator.collaboration,
		command.IncidentID,
		command.Actor.ID,
		request.ClientTxnID,
		changeSetID,
		AttachRecordChange{
			RecordID:         recordID,
			RowVersion:       1,
			ViewSchemaID:     request.ViewSchemaID,
			ChangedFieldKeys: changedFieldKeys,
		},
		row,
		nil,
		now,
	); err != nil {
		return evidenceCreateTxResult{}, err
	}
	return evidenceCreateTxResult{
		payload:          buildMutationPayload(request.ViewSchemaID, changeSetID, row),
		row:              row,
		recordID:         recordID,
		changeSetID:      changeSetID,
		changedFieldKeys: changedFieldKeys,
	}, nil
}
