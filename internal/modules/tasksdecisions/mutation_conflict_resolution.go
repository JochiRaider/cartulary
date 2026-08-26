package tasksdecisions

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	conflictresolution "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

type ConflictCommand struct {
	Mechanics      conflictresolution.Command
	ResolutionKind string
	Patch          *PatchRequest
	Now            time.Time
}

func (f *MutationFacade) ResolveConflict(
	ctx context.Context,
	command ConflictCommand,
) (MutationResult, error) {
	if command.ResolutionKind != "keep_saved" {
		return f.Patch(ctx, PatchCommand{
			ActorUserID:      command.Mechanics.ActorUserID,
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
		f.keepSaved,
		command.Mechanics,
		f.loadConflictTarget,
	)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{
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
	meta, err := loadSupersedeRecordMetaForUpdateTx(ctx, tx, f.recordStore, command.RecordID)
	if err != nil {
		return conflictresolution.Target{}, err
	}
	if !recordTypeMatchesView(f.catalog, meta.RecordType, command.Claims.ViewSchemaID) {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	if _, err := f.conflictFields.ResolveWritableField(command.Claims.ViewSchemaID, command.Claims.FieldKey); err != nil {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return conflictresolution.Target{}, err
	}
	row, err := f.loadProjectionRowTx(
		ctx,
		tx,
		command.Claims.ViewSchemaID,
		command.RecordID,
	)
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
