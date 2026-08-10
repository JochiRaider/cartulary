package artifacts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	conflictresolution "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type WorkbookConflictCommand struct {
	Mechanics      conflictresolution.Command
	ActorUserID    uuid.UUID
	OperationID    OperationID
	ResolutionKind string
	Patch          *WorkbookPatchRequest
	Now            time.Time
}

func (f *MutationFacade) ResolveConflict(
	ctx context.Context,
	command WorkbookConflictCommand,
) (WorkbookMutationResult, error) {
	if command.OperationID != OperationConflictResolve {
		return WorkbookMutationResult{}, ErrStoredMutationKindMismatch
	}
	command.Mechanics.ActorUserID = command.ActorUserID
	command.Mechanics.RouteKey = string(command.OperationID)
	if command.ResolutionKind != "keep_saved" {
		return f.Patch(ctx, WorkbookPatchCommand{
			ActorUserID:         command.ActorUserID,
			RecordID:            command.Mechanics.RecordID,
			Request:             *command.Patch,
			RequestHash:         command.Mechanics.RequestHash,
			RequestID:           command.Mechanics.RequestID,
			OperationID:         command.OperationID,
			ConflictOperationID: command.OperationID,
			Now:                 command.Now,
		})
	}
	result, err := conflictresolution.KeepSaved(
		ctx,
		f.pool,
		f.keepSaved,
		command.Mechanics,
		f.loadConflictTarget,
	)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	return WorkbookMutationResult{
		Row:          conflictResultRow(result.Payload),
		Replayed:     result.Replayed,
		IncidentID:   result.IncidentID,
		RecordID:     result.RecordID,
		ClientTxnID:  result.ClientTxnID,
		RowVersion:   result.RowVersion,
		ViewSchemaID: result.ViewSchemaID,
	}, nil
}

func (f *MutationFacade) loadConflictTarget(
	ctx context.Context,
	tx pgx.Tx,
	command conflictresolution.Command,
) (conflictresolution.Target, error) {
	meta, err := f.loadArtifactRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return conflictresolution.Target{}, err
	}
	if meta.RecordType != "artifact" {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	if err := validateArtifactViewRecordTx(
		ctx,
		tx,
		command.RecordID,
		command.Claims.ViewSchemaID,
	); err != nil {
		return conflictresolution.Target{}, err
	}
	field, ok := viewschema.LookupField(command.Claims.ViewSchemaID, command.Claims.FieldKey)
	if !ok || !field.Writable {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return conflictresolution.Target{}, err
	}
	row, err := f.source.projections.LoadArtifactTx(ctx, tx, command.Claims.ViewSchemaID, command.RecordID)
	if err != nil {
		return conflictresolution.Target{}, err
	}
	return conflictresolution.Target{
		IncidentID: meta.IncidentID,
		RecordID:   command.RecordID,
		RowVersion: meta.RowVersion,
		Row:        row,
	}, nil
}

func conflictResultRow(payload map[string]any) map[string]any {
	row, _ := payload["row"].(map[string]any)
	return row
}
