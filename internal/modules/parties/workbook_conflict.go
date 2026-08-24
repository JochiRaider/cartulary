package parties

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	partysource "github.com/JochiRaider/cartulary/internal/modules/parties/internal/source"
	conflictresolution "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ConflictCommand struct {
	Mechanics   conflictresolution.Command
	ActorUserID uuid.UUID
	Admission   ConflictResolveAdmission
	Now         time.Time
}

func (f *MutationFacade) ResolveConflict(
	ctx context.Context,
	command ConflictCommand,
) (MutationResult, error) {
	admission := command.Admission
	if command.Mechanics.RouteKey != workbookConflictResolveOperation ||
		command.Mechanics.ActorUserID != command.ActorUserID ||
		command.Mechanics.RecordID != admission.claims.RecordID ||
		command.Mechanics.Claims.ViewSchemaID != admission.claims.ViewSchemaID ||
		command.Mechanics.Claims.FieldKey != admission.claims.FieldKey ||
		command.Mechanics.Claims.CurrentRowVersion != admission.claims.CurrentRowVersion {
		return MutationResult{}, &ValidationError{Field: "conflict_token", ReasonCode: "invalid_value"}
	}
	command.Mechanics.RequestHash = admission.RequestHash()
	command.Mechanics.ClientTxnID = admission.clientTxnID
	if admission.resolutionKind != "keep_saved" {
		return f.Patch(ctx, PatchCommand{
			ActorUserID:      command.ActorUserID,
			RecordID:         command.Mechanics.RecordID,
			Admission:        *admission.patch,
			RequestID:        command.Mechanics.RequestID,
			RouteKey:         command.Mechanics.RouteKey,
			ConflictRouteKey: command.Mechanics.RouteKey,
			Now:              command.Now,
		})
	}
	if f.keepSaved == nil {
		return MutationResult{}, fmt.Errorf("parties: keep-saved idempotency is not configured")
	}
	result, err := f.keepSaved.KeepSaved(
		ctx,
		f.pool,
		command.Mechanics,
		f.loadConflictTarget,
	)
	if err != nil {
		return MutationResult{}, err
	}
	return MutationResult{
		Outcome:      conflictMutationOutcome(result.Replayed),
		Row:          result.Row,
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
	meta, err := partysource.LoadRecordMetaForUpdateTx(ctx, tx, command.RecordID)
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
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return conflictresolution.Target{}, err
	}
	row, err := f.projectionRows.LoadPartyTx(ctx, tx, command.RecordID)
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

func conflictMutationOutcome(replayed bool) MutationOutcome {
	if replayed {
		return MutationReplayed
	}
	return MutationKeptSaved
}
