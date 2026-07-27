package parties

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflictresolution"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type WorkbookConflictCommand struct {
	Mechanics      conflictresolution.Command
	ResolutionKind string
	Patch          *WorkbookPatchRequest
	Now            time.Time
}

func (f *WorkbookFacade) ResolveConflict(
	ctx context.Context,
	command WorkbookConflictCommand,
) (WorkbookMutationResult, error) {
	if command.ResolutionKind != "keep_saved" {
		return f.Patch(ctx, WorkbookPatchCommand{
			Actor:            command.Mechanics.Actor,
			RecordID:         command.Mechanics.RecordID,
			Request:          *command.Patch,
			RequestHash:      command.Mechanics.RequestHash,
			RequestID:        command.Mechanics.RequestID,
			RouteKey:         command.Mechanics.RouteKey,
			ConflictRouteKey: command.Mechanics.RouteKey,
			Now:              command.Now,
		})
	}
	result, err := conflictresolution.KeepSaved(
		ctx,
		f.pool,
		f.authStore,
		command.Mechanics,
		f.loadConflictTarget,
	)
	if err != nil {
		return WorkbookMutationResult{}, err
	}
	return WorkbookMutationResult{
		Payload:      result.Payload,
		StatusCode:   result.StatusCode,
		Replayed:     result.Replayed,
		IncidentID:   result.IncidentID,
		RecordID:     result.RecordID,
		ClientTxnID:  result.ClientTxnID,
		RowVersion:   result.RowVersion,
		ViewSchemaID: result.ViewSchemaID,
	}, nil
}

func (f *WorkbookFacade) loadConflictTarget(
	ctx context.Context,
	tx pgx.Tx,
	command conflictresolution.Command,
) (conflictresolution.Target, error) {
	meta, err := loadPartyRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return conflictresolution.Target{}, err
	}
	if meta.RecordType != "party" || command.Claims.ViewSchemaID != ViewSchemaID {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	field, ok := viewschema.LookupField(command.Claims.ViewSchemaID, command.Claims.FieldKey)
	if !ok || !field.Writable {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return conflictresolution.Target{}, err
	}
	row, err := f.projectionRows.LoadTx(ctx, tx, command.RecordID)
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
