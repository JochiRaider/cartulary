package hostidentity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflictresolution"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type WorkbookConflictCommand struct {
	Mechanics      conflictresolution.Command
	Actor          authn.UserRecord
	ResolutionKind string
	Patch          *PatchRequest
	Now            time.Time
}

func (s *Store) ResolveWorkbookConflict(
	ctx context.Context,
	command WorkbookConflictCommand,
) (PatchMutationResult, error) {
	if command.ResolutionKind != "keep_saved" {
		return s.PatchEntityRow(
			ctx,
			command.Actor,
			command.Mechanics.RecordID,
			*command.Patch,
			command.Mechanics.RequestHash,
			command.Mechanics.RequestID,
			command.Now,
			command.Mechanics.RouteKey,
		)
	}
	result, err := conflictresolution.KeepSaved(
		ctx,
		s.pool,
		conflictresolution.NewRouteIdempotencyAdapter(s.authStore),
		command.Mechanics,
		s.loadConflictTarget,
	)
	if err != nil {
		return PatchMutationResult{}, err
	}
	return PatchMutationResult{
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

func (s *Store) loadConflictTarget(
	ctx context.Context,
	tx pgx.Tx,
	command conflictresolution.Command,
) (conflictresolution.Target, error) {
	meta, err := loadEntityRecordMetaForUpdateTx(ctx, tx, command.RecordID)
	if err != nil {
		return conflictresolution.Target{}, err
	}
	if !entityRecordTypeMatchesView(meta.RecordType, command.Claims.ViewSchemaID) {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	field, ok := viewschema.LookupField(command.Claims.ViewSchemaID, command.Claims.FieldKey)
	if !ok || !field.Writable {
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return conflictresolution.Target{}, err
	}
	var row map[string]any
	switch meta.RecordType {
	case "host":
		record, err := LoadHostByRecordIDTx(ctx, tx, command.RecordID)
		if err != nil {
			return conflictresolution.Target{}, err
		}
		if err := hydrateHostRecordTx(ctx, tx, &record); err != nil {
			return conflictresolution.Target{}, err
		}
		row = BuildHostRow(record)
	case "identity":
		record, err := LoadIdentityByRecordIDTx(ctx, tx, command.RecordID)
		if err != nil {
			return conflictresolution.Target{}, err
		}
		if err := hydrateIdentityRecordTx(ctx, tx, &record); err != nil {
			return conflictresolution.Target{}, err
		}
		row = BuildIdentityRow(record)
	default:
		return conflictresolution.Target{}, pgx.ErrNoRows
	}
	return conflictresolution.Target{
		IncidentID: meta.IncidentID,
		RecordID:   command.RecordID,
		RowVersion: meta.RowVersion,
		Row:        row,
	}, nil
}
