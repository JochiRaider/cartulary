package imports

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const importApplyChangeSetSource = "imports.apply"

func (s *Service) applyGenericOwnerUnit(ctx context.Context, actor authn.UserRecord, start ApplyStartResult, unit ApplyUnitData, target importTarget) error {
	now := s.now().UTC()
	owner, ok := s.ownerCreateRegistry.Resolve(
		unit.ApprovedMapping.TargetViewSchemaID,
		target.CreateFacade,
	)
	if !ok {
		return importApplyBlockedError("owner_create_contract_unavailable")
	}
	tx, err := s.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin import apply unit transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, start.IncidentID); err != nil {
		return err
	}
	clientTxnID := fmt.Sprintf("import:%s:%s:%s", start.ImportSessionID, unit.UnitID, start.ClientTxnID)
	requestID := "req-" + clientTxnID
	changeSetID, err := newRevisionAppendAdapter(s.store.revisionAppender).AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  start.IncidentID,
		ActorUserID: actor.ID,
		Source:      importApplyChangeSetSource,
		ClientTxnID: &clientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now,
	})
	if err != nil {
		return err
	}

	for index, sourceRow := range unit.SourceRows {
		rowRef, _ := intFromAny(sourceRow["source_row_ref"])
		if rowRef <= 0 {
			return fmt.Errorf("import source row missing source_row_ref")
		}
		rowClientTxnID := fmt.Sprintf("%s:%d", clientTxnID, rowRef)
		request, err := importOwnerCreateRequest(start, unit, actor.ID, sourceRow, rowRef, rowClientTxnID)
		if err != nil {
			return err
		}
		response, err := owner.CreateImportRowTx(
			ctx,
			tx,
			ownerfacade.ImportOwnerCreateCommand{
				Request:     request,
				ChangeSetID: changeSetID,
				SequenceNo:  index + 1,
				Now:         now,
			},
		)
		if err != nil {
			return err
		}
		ownerResponse := map[string]any{
			"record_id":               response.RecordID.String(),
			"row_version":             response.RowVersion,
			"change_set_mutation_ref": response.ChangeSetMutationRef,
			"created_or_reused":       response.CreatedOrReused,
			"owner_result_code":       response.OwnerResultCode,
			"target_view_schema_id":   request.TargetViewSchemaID,
			"import_session_id":       request.ImportSessionID.String(),
			"import_unit_id":          request.ImportUnitID.String(),
			"mapping_fingerprint":     request.MappingFingerprint,
			"source_row_ref":          request.SourceRowRef,
		}
		if err := s.store.InsertApplyJournalTx(ctx, tx, ApplyJournalParams{
			ImportSessionID:      request.ImportSessionID,
			ImportUnitID:         request.ImportUnitID,
			MappingFingerprint:   request.MappingFingerprint,
			SourceRowRef:         request.SourceRowRef,
			TargetViewSchemaID:   request.TargetViewSchemaID,
			OwnerCreateFacade:    target.CreateFacade,
			RecordID:             response.RecordID,
			RowVersion:           response.RowVersion,
			ChangeSetID:          changeSetID,
			ChangeSetMutationRef: response.ChangeSetMutationRef,
			OwnerResultCode:      response.OwnerResultCode,
			CreatedOrReused:      response.CreatedOrReused,
			OwnerResponse:        ownerResponse,
			RowRefresh:           response.RowRefresh,
			CreatedAt:            now,
		}); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit import apply unit transaction: %w", err)
	}
	return nil
}

func importOwnerCreateRequest(start ApplyStartResult, unit ApplyUnitData, actorID uuid.UUID, sourceRow map[string]any, rowRef int, clientTxnID string) (ownerfacade.ImportOwnerCreateRequest, error) {
	cells := sourceRowCellsByOrdinal(sourceRow)
	request := ownerfacade.ImportOwnerCreateRequest{
		IncidentID:          start.IncidentID,
		ActorUserID:         actorID,
		TargetViewSchemaID:  unit.ApprovedMapping.TargetViewSchemaID,
		ImportSessionID:     start.ImportSessionID,
		ImportUnitID:        unit.UnitID,
		MappingFingerprint:  unit.MappingFingerprint,
		SourceFileKind:      unit.SourceFileKind,
		SourceContentSHA256: unit.SourceContentSHA256,
		ParserProfileID:     unit.ParserProfileID,
		ParserVersion:       unit.ParserVersion,
		LocatorKind:         unit.LocatorKind,
		Locator:             unit.Locator,
		SourceRectA1:        unit.SourceRectA1,
		SourceRowRef:        rowRef,
		ClientTxnID:         clientTxnID,
		SourceRowProvenance: ownerfacade.ImportSourceRowProvenance{SourceRowRef: rowRef},
	}
	for _, column := range unit.ApprovedMapping.SourceColumns {
		cell := cells[column.SourceColumnOrdinal]
		rawValue, _ := cell["display_text"].(string)
		cellKind, _ := cell["cell_kind"].(string)
		if column.FieldKey == nil {
			request.UnknownValues = append(request.UnknownValues, ownerfacade.ImportUnknownValue{
				SourceColumnOrdinal: column.SourceColumnOrdinal,
				SourceHeaderText:    column.SourceHeaderText,
				RawValue:            rawValue,
				CellKind:            cellKind,
			})
			continue
		}
		transformed, err := transformImportValue(rawValue, column)
		if err != nil {
			return ownerfacade.ImportOwnerCreateRequest{}, err
		}
		value, include, err := ownerfacade.NormalizeImportScalar(unit.ApprovedMapping.TargetViewSchemaID, *column.FieldKey, transformed, column.EmptyValuePolicy)
		if err != nil {
			return ownerfacade.ImportOwnerCreateRequest{}, err
		}
		if !include {
			continue
		}
		var transformID *string
		if column.TransformID != nil {
			transformID = column.TransformID
		}
		var entityBinding *string
		if column.EntityBindingMode != nil {
			entityBinding = column.EntityBindingMode
		}
		request.FieldValues = append(request.FieldValues, ownerfacade.ImportFieldValue{
			FieldKey:            *column.FieldKey,
			NormalizedValue:     value,
			SourceColumnOrdinal: column.SourceColumnOrdinal,
			SourceHeaderText:    column.SourceHeaderText,
			RawValue:            rawValue,
			CellKind:            cellKind,
			TransformID:         transformID,
			EmptyValuePolicy:    column.EmptyValuePolicy,
			EntityBindingMode:   entityBinding,
		})
	}
	return request, nil
}
