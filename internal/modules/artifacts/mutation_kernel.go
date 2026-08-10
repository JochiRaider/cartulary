package artifacts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

// artifactCreateTxResult is the durable output of the transaction-supplied
// source kernel. Route idempotency and commit remain coordinator concerns.
type artifactCreateTxResult struct {
	payload     map[string]any
	row         map[string]any
	incidentID  uuid.UUID
	recordID    uuid.UUID
	changeSetID uuid.UUID
}

// executeCreateTx performs all authoritative artifact-create effects in the
// supplied transaction. Both ordinary workbook creates and contextual-note
// creates use this sequence; this method never begins or commits a transaction.
func (f *WorkbookFacade) executeCreateTx(
	ctx context.Context,
	tx pgx.Tx,
	command WorkbookCreateCommand,
	contextualSourceRecordID *uuid.UUID,
) (artifactCreateTxResult, error) {
	request := command.Request
	incidentID := command.IncidentID
	var err error
	if contextualSourceRecordID != nil {
		incidentID, err = f.contextIncidentTx(ctx, tx, *contextualSourceRecordID)
		if err != nil {
			return artifactCreateTxResult{}, err
		}
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, incidentID); err != nil {
		return artifactCreateTxResult{}, err
	}
	if err := validateArtifactReferencesTx(
		ctx,
		tx,
		f.linkStore,
		incidentID,
		request.ViewSchemaID,
		request.Values,
		request.Collections,
	); err != nil {
		return artifactCreateTxResult{}, err
	}
	now := command.Now.UTC()
	recordID, err := f.source.createRecordTx(
		ctx,
		tx,
		incidentID,
		command.Actor.ID,
		CreateParams{ViewSchemaID: request.ViewSchemaID, Values: request.Values},
		now,
	)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	collectionMutations, err := f.applyCollectionsTx(
		ctx,
		tx,
		incidentID,
		recordID,
		command.Actor.ID,
		request.ViewSchemaID,
		request.Collections,
		now,
	)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	var (
		contextLink         links.RecordLink
		contextLinkInserted bool
	)
	if contextualSourceRecordID != nil {
		contextLink, contextLinkInserted, err = f.linkStore.InsertLinkedNoteReferenceTx(
			ctx,
			tx,
			incidentID,
			*contextualSourceRecordID,
			recordID,
			command.Actor.ID,
			now,
		)
		if err != nil {
			return artifactCreateTxResult{}, err
		}
	}
	row, err := f.source.refreshRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	changeSetID, err := f.revisionAppender.AppendChangeSetTx(
		ctx,
		tx,
		revisions.AppendChangeSetParams{
			IncidentID:  incidentID,
			ActorUserID: command.Actor.ID,
			Source:      command.RouteKey,
			ClientTxnID: &request.ClientTxnID,
			RequestID:   &command.RequestID,
			CreatedAt:   now,
		},
	)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	afterSnapshot, err := f.revisionAppender.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := f.revisionAppender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		RecordID:       recordID,
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return artifactCreateTxResult{}, err
	}
	nextSequence, err := f.appendCollectionMutationsTx(ctx, tx, changeSetID, 2, collectionMutations)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	if contextLinkInserted {
		linkAfter, err := f.linkStore.LoadRecordLinkValueTx(ctx, tx, contextLink.RecordLinkID)
		if err != nil {
			return artifactCreateTxResult{}, err
		}
		if err := f.revisionAppender.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    nextSequence,
			TargetKind:    "record_link",
			TargetID:      contextLink.RecordLinkID.String(),
			OperationKind: "create",
			AfterValue:    linkAfter,
		}); err != nil {
			return artifactCreateTxResult{}, err
		}
	}
	if err := f.revisionAppender.AppendRecordRevisionAndIntentTx(
		ctx,
		tx,
		revisions.AppendRecordRevisionParams{
			ChangeSetID:   changeSetID,
			RecordID:      recordID,
			RowVersion:    1,
			AfterSnapshot: &afterSnapshot,
			LiveChange:    revisions.LiveRecordChange{AfterValue: row},
		},
	); err != nil {
		return artifactCreateTxResult{}, err
	}
	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, row)
	if contextualSourceRecordID != nil {
		payload["source_record_id"] = contextualSourceRecordID.String()
		payload["link_type"] = "references_artifact"
	}
	return artifactCreateTxResult{
		payload:     payload,
		row:         row,
		incidentID:  incidentID,
		recordID:    recordID,
		changeSetID: changeSetID,
	}, nil
}
