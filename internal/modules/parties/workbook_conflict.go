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
	ActorUserID uuid.UUID
	Admission   ConflictResolveAdmission
	RequestID   string
	Now         time.Time
}

func (f *MutationFacade) ResolveConflict(
	ctx context.Context,
	command ConflictCommand,
) (MutationResult, error) {
	admission := command.Admission
	if admission.claims.RecordID == uuid.Nil || admission.claims.ViewSchemaID != ViewSchemaID ||
		admission.claims.FieldKey == "" || admission.claims.CurrentRowVersion < 1 {
		return MutationResult{}, &ValidationError{Field: "conflict_token", ReasonCode: "invalid_value"}
	}
	if admission.resolutionKind != "keep_saved" {
		return f.Patch(ctx, PatchCommand{
			ActorUserID: command.ActorUserID,
			RecordID:    admission.claims.RecordID,
			Admission:   *admission.patch,
			RequestID:   command.RequestID,
			Now:         command.Now,
		})
	}
	if f.keepSaved == nil {
		return MutationResult{}, fmt.Errorf("parties: keep-saved idempotency is not configured")
	}
	mechanics := conflictresolution.Command{
		ActorUserID: command.ActorUserID,
		RecordID:    admission.claims.RecordID,
		Claims: conflictresolution.ConflictTokenClaims{
			RouteKey:          workbookConflictResolveOperation,
			RecordID:          admission.claims.RecordID.String(),
			ViewSchemaID:      admission.claims.ViewSchemaID,
			FieldKey:          admission.claims.FieldKey,
			CurrentRowVersion: admission.claims.CurrentRowVersion,
		},
		ClientTxnID: admission.clientTxnID,
		RequestHash: append([]byte(nil), admission.requestHash[:]...),
		RequestID:   command.RequestID,
		RouteKey:    workbookConflictResolveOperation,
	}
	result, err := f.keepSaved.KeepSaved(
		ctx,
		f.pool,
		mechanics,
		f.loadConflictTarget,
	)
	if err != nil {
		return MutationResult{}, err
	}
	rowVersion, parseErr := rowVersionFromGenericRow(result.Row)
	if parseErr != nil {
		return MutationResult{}, fmt.Errorf("parties: keep-saved result: %w", parseErr)
	}
	if result.RowVersion != 0 && result.RowVersion != rowVersion {
		return MutationResult{}, fmt.Errorf("parties: keep-saved result row version mismatch")
	}
	incidentID := result.IncidentID
	if incidentID == uuid.Nil {
		envelope, err := f.recordStore.LoadEnvelope(ctx, result.RecordID)
		if err != nil {
			return MutationResult{}, fmt.Errorf("parties: load immutable keep-saved record identity: %w", err)
		}
		if envelope.RecordType != "party" {
			return MutationResult{}, fmt.Errorf("parties: keep-saved record identity is not a Party")
		}
		incidentID = envelope.IncidentID
	}
	return MutationResult{
		Outcome:          conflictMutationOutcome(result.Replayed),
		Row:              result.Row,
		IncidentID:       incidentID,
		RecordID:         result.RecordID,
		ChangeSetID:      nil,
		RowVersion:       rowVersion,
		ChangedFieldKeys: []string{},
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
