package evidence

// Conflict resolution remains part of the Evidence semantic mutation contract.

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	conflictresolution "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ConflictCommand struct {
	ActorUserID uuid.UUID
	Admission   ConflictResolveAdmission
	RequestID   string
	Now         time.Time
}

func (f *mutationFacade) ResolveConflict(ctx context.Context, command ConflictCommand) (MutationResult, error) {
	if !command.Admission.valid() || command.ActorUserID == uuid.Nil {
		return MutationResult{}, &ValidationError{Field: "payload", ReasonCode: "invalid_value"}
	}
	request := command.Admission.requestValue()
	conflictContext := command.Admission.contextValue()
	mechanics := conflictresolution.Command{
		ActorUserID: command.ActorUserID,
		RecordID:    conflictContext.RecordID,
		Claims: conflictresolution.ConflictTokenClaims{
			Version: conflictContext.Version, RecordID: conflictContext.RecordID.String(),
			ViewSchemaID: conflictContext.ViewSchemaID, RouteKey: conflictContext.RouteKey,
			FieldKey: conflictContext.FieldKey, ConflictResolutionClass: conflictContext.ConflictResolutionClass,
			BaseRowVersion: conflictContext.BaseRowVersion, CurrentRowVersion: conflictContext.CurrentRowVersion,
			RequestHash: conflictContext.OriginalRequestHash,
			IssuedAt:    conflictContext.IssuedAt, ExpiresAt: conflictContext.ExpiresAt,
		},
		ClientTxnID: request.ClientTxnID,
		RequestHash: command.Admission.requestHash(),
		RequestID:   command.RequestID,
		RouteKey:    string(OperationConflictResolve),
	}
	if request.ResolutionKind != "keep_saved" {
		return f.Patch(ctx, PatchCommand{
			ActorUserID: command.ActorUserID,
			RecordID:    conflictContext.RecordID,
			Admission:   command.Admission.patchAdmission(),
			RequestID:   command.RequestID,
			Now:         command.Now,
			operation:   OperationConflictResolve,
		})
	}
	if f.keepSaved == nil {
		return MutationResult{}, fmt.Errorf("evidence: keep-saved idempotency is not configured")
	}
	result, err := conflictresolution.KeepSaved(ctx, f.pool, f.keepSaved, mechanics, f.loadConflictTarget)
	if err != nil {
		return MutationResult{}, err
	}
	outcome := MutationOutcomeKeptSaved
	if result.Replayed {
		outcome = MutationOutcomeReplayed
	}
	row, ok := result.Payload["row"].(map[string]any)
	if !ok {
		return MutationResult{}, ErrStoredMutationKindMismatch
	}
	return MutationResult{
		Row: cloneStringAnyMap(row), Outcome: outcome,
		IncidentID: result.IncidentID, RecordID: result.RecordID,
		ClientTxnID: result.ClientTxnID, RowVersion: result.RowVersion, ViewSchemaID: result.ViewSchemaID,
	}, nil
}

func (f *mutationFacade) loadConflictTarget(
	ctx context.Context,
	tx pgx.Tx,
	command conflictresolution.Command,
) (conflictresolution.Target, error) {
	meta, err := f.loadEvidenceRecordMetaForUpdateTx(ctx, tx, command.RecordID)
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
	if err := f.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
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
