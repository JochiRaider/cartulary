package artifacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func (f *MutationFacade) Create(ctx context.Context, command CreateCommand) (MutationResult, error) {
	return f.create(ctx, command, nil)
}

func (f *MutationFacade) create(ctx context.Context, command CreateCommand, contextualSourceRecordID *uuid.UUID) (MutationResult, error) {
	request := command.Request
	wantOperation := OperationCreate
	wantKind := StoredMutationCreate
	if contextualSourceRecordID != nil {
		wantOperation = OperationLinkedNoteCreate
		wantKind = StoredMutationLinkedNote
	}
	if command.OperationID != wantOperation {
		return MutationResult{}, ErrStoredMutationKindMismatch
	}
	scopeKey := command.IncidentID.String() + ":" + request.ViewSchemaID
	if contextualSourceRecordID != nil {
		scopeKey = contextualSourceRecordID.String()
	}
	idempotencyKey := IdempotencyKey{
		OperationID: command.OperationID,
		ActorUserID: command.ActorUserID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	replayedStored, replayed, err := f.replayStoredMutation(ctx, idempotencyKey, command.RequestHash, "create", storedMutationExpectation{
		kind: wantKind, viewSchemaID: request.ViewSchemaID,
	})
	if err != nil {
		return MutationResult{}, err
	}
	if replayed {
		result := MutationResult{
			Row: replayedStored.Row, Replayed: true, IncidentID: command.IncidentID,
			RecordID: replayedStored.RecordID, ChangeSetID: replayedStored.ChangeSetID,
			ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID,
			RowVersion: rowVersionFromCanonicalRow(replayedStored.Row),
		}
		if replayedStored.SourceRecordID != nil {
			result.ContextualLink = &ContextualLink{SourceRecordID: *replayedStored.SourceRecordID, LinkType: replayedStored.LinkType}
		}
		return result, nil
	}
	if err := validateCreateParams(createParams{ViewSchemaID: request.ViewSchemaID, Values: request.Values}); err != nil {
		return MutationResult{}, err
	}

	tx, err := f.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin artifact create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	mutation, err := f.executeCreateTx(ctx, tx, command, contextualSourceRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	stored := StoredMutationPayload{
		ViewSchemaID: request.ViewSchemaID, RecordID: mutation.recordID,
		ChangeSetID: mutation.changeSetID, Row: mutation.row,
	}
	var storedResult StoredMutationResult
	if contextualSourceRecordID == nil {
		storedResult = NewStoredCreateResult(stored)
	} else {
		stored.SourceRecordID = contextualSourceRecordID
		stored.LinkType = "references_artifact"
		storedResult = NewStoredLinkedNoteResult(stored)
	}
	if err := f.idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, storedResult); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit artifact create transaction: %w", err)
	}
	return MutationResult{
		Row:              mutation.row,
		Created:          true,
		IncidentID:       mutation.incidentID,
		RecordID:         mutation.recordID,
		ChangeSetID:      mutation.changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, mutation.row),
		ContextualLink:   contextualLinkFacts(contextualSourceRecordID),
	}, nil
}

// artifactCreateTxResult is the durable output of the transaction-supplied
// source kernel. Route idempotency and commit remain coordinator concerns.
type artifactCreateTxResult struct {
	row         map[string]any
	incidentID  uuid.UUID
	recordID    uuid.UUID
	changeSetID uuid.UUID
}

// executeCreateTx performs all authoritative artifact-create effects in the
// supplied transaction. Both ordinary workbook creates and contextual-note
// creates use this sequence; this method never begins or commits a transaction.
func (f *MutationFacade) executeCreateTx(
	ctx context.Context,
	tx pgx.Tx,
	command CreateCommand,
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
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return artifactCreateTxResult{}, err
	}
	if err := validateArtifactReferencesTx(
		ctx,
		tx,
		f.memberReferences,
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
		command.ActorUserID,
		createParams{ViewSchemaID: request.ViewSchemaID, Values: request.Values},
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
		command.ActorUserID,
		request.ViewSchemaID,
		request.Collections,
		now,
	)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	var contextLinkMutation *links.RecordLinkMutation
	if contextualSourceRecordID != nil {
		contextLink, err := f.linkStore.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
			IncidentID:  incidentID,
			SrcRecordID: *contextualSourceRecordID,
			DstRecordID: recordID,
			LinkType:    links.LinkType(links.LinkTypeReferencesArtifact),
			Provenance:  links.LinkProvenance(links.LinkProvenanceManual),
			OwnerUserID: command.ActorUserID,
			Now:         now,
		})
		if err != nil {
			return artifactCreateTxResult{}, err
		}
		contextLinkMutation = contextLink.Mutation
	}
	row, err := f.source.refreshRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	changeSetID, err := f.revisions.AppendChangeSetTx(
		ctx,
		tx,
		revisions.AppendChangeSetParams{
			IncidentID:  incidentID,
			ActorUserID: command.ActorUserID,
			Source:      string(command.OperationID),
			ClientTxnID: &request.ClientTxnID,
			RequestID:   &command.RequestID,
			CreatedAt:   now,
		},
	)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	afterSnapshot, err := f.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return artifactCreateTxResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := f.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
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
	if contextLinkMutation != nil {
		if err := f.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    nextSequence,
			TargetKind:    "record_link",
			TargetID:      contextLinkMutation.RecordLinkID.String(),
			OperationKind: contextLinkMutation.Operation,
			BeforeValue:   contextLinkMutation.BeforeValue,
			AfterValue:    contextLinkMutation.AfterValue,
		}); err != nil {
			return artifactCreateTxResult{}, err
		}
	}
	if err := f.revisions.AppendRecordRevisionAndIntentTx(
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
	return artifactCreateTxResult{
		row:         row,
		incidentID:  incidentID,
		recordID:    recordID,
		changeSetID: changeSetID,
	}, nil
}
