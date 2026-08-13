package evidence

import (
	"context"
	"time"

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

type evidenceCreateTxCommand struct {
	IncidentID    uuid.UUID
	ActorUserID   uuid.UUID
	ViewSchemaID  string
	ClientTxnID   string
	RequestID     string
	Source        string
	ChangeSetID   *uuid.UUID
	MutationOrder int
	Values        map[string]FieldValue
	Now           time.Time
}

// evidenceSourceMutationKernel orders owner and cross-owner durable effects in
// a supplied transaction. Transaction begin, idempotency replay, commit, and
// HTTP response handling remain application/route concerns.
type evidenceSourceMutationKernel struct {
	incidents     evidenceIncidentAdmissionPort
	source        evidenceSourceKernel
	revisions     revisionAppendPort
	collaboration collaboration.IntentAppender
}

func (coordinator evidenceSourceMutationKernel) createTx(
	ctx context.Context,
	tx pgx.Tx,
	command evidenceCreateTxCommand,
	createParams createParams,
) (evidenceCreateTxResult, error) {
	if err := coordinator.incidents.EnsureOpenTx(ctx, tx, command.IncidentID); err != nil {
		return evidenceCreateTxResult{}, err
	}
	if err := validateEvidenceReferencesTx(ctx, tx, command.IncidentID, command.Values); err != nil {
		return evidenceCreateTxResult{}, err
	}
	now := command.Now.UTC()
	recordID, err := coordinator.source.createRecordTx(
		ctx,
		tx,
		command.IncidentID,
		command.ActorUserID,
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
			ActorUserID:      &command.ActorUserID,
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
	var changeSetID uuid.UUID
	if command.ChangeSetID != nil {
		changeSetID = *command.ChangeSetID
	} else {
		requestID := command.RequestID
		changeSetID, err = coordinator.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
			IncidentID:  command.IncidentID,
			ActorUserID: command.ActorUserID,
			Source:      command.Source,
			ClientTxnID: &command.ClientTxnID,
			RequestID:   &requestID,
			CreatedAt:   now,
		})
		if err != nil {
			return evidenceCreateTxResult{}, err
		}
	}
	mutationOrder := command.MutationOrder
	if mutationOrder <= 0 {
		mutationOrder = 1
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := coordinator.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     mutationOrder,
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
		command.ActorUserID,
		command.ClientTxnID,
		changeSetID,
		attachRecordChange{
			RecordID:         recordID,
			RowVersion:       1,
			ViewSchemaID:     command.ViewSchemaID,
			ChangedFieldKeys: changedFieldKeys,
		},
		row,
		nil,
		now,
	); err != nil {
		return evidenceCreateTxResult{}, err
	}
	return evidenceCreateTxResult{
		payload:          buildMutationPayload(command.ViewSchemaID, changeSetID, row),
		row:              row,
		recordID:         recordID,
		changeSetID:      changeSetID,
		changedFieldKeys: changedFieldKeys,
	}, nil
}
