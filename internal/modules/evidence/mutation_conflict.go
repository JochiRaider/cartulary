package evidence

// Conflict resolution remains part of the Evidence mutation contract.

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	conflictresolution "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ConflictCommand struct {
	Mechanics      conflictresolution.Command
	Actor          authn.UserRecord
	ResolutionKind string
	Patch          *PatchRequest
	Now            time.Time
}

func (f *mutationFacade) ResolveConflict(
	ctx context.Context,
	command ConflictCommand,
) (MutationResult, error) {
	if command.ResolutionKind != "keep_saved" {
		return f.Patch(ctx, PatchCommand{
			Actor:            command.Actor,
			RecordID:         command.Mechanics.RecordID,
			Request:          *command.Patch,
			RequestHash:      command.Mechanics.RequestHash,
			RequestID:        command.Mechanics.RequestID,
			RouteKey:         command.Mechanics.RouteKey,
			ConflictRouteKey: command.Mechanics.RouteKey,
			Now:              command.Now,
		})
	}
	if f.keepSaved == nil {
		return MutationResult{}, fmt.Errorf("evidence: keep-saved idempotency is not configured")
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
		Payload:      result.Payload,
		StatusCode:   http.StatusOK,
		Replayed:     result.Replayed,
		IncidentID:   result.IncidentID,
		RecordID:     result.RecordID,
		ClientTxnID:  result.ClientTxnID,
		RowVersion:   result.RowVersion,
		ViewSchemaID: result.ViewSchemaID,
	}, nil
}

func (f *mutationFacade) loadConflictTarget(
	ctx context.Context,
	tx pgx.Tx,
	command conflictresolution.Command,
) (conflictresolution.Target, error) {
	meta, err := loadEvidenceRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return conflictresolution.Target{}, err
	}
	if meta.RecordType != "evidence" || command.Claims.ViewSchemaID != ViewSchemaID {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	field, ok := viewschema.LookupField(command.Claims.ViewSchemaID, command.Claims.FieldKey)
	if !ok || !field.Writable {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	if err := f.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return conflictresolution.Target{}, err
	}
	row, err := f.projectionRows.LoadEvidenceTx(ctx, tx, command.RecordID)
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
