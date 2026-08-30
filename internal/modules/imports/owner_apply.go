package imports

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const importApplyChangeSetSource = "imports.apply"

func (s *Service) applyGenericOwnerUnitTx(
	ctx context.Context,
	tx pgx.Tx,
	actor authn.UserRecord,
	start ApplyStartResult,
	unit ApplyUnitData,
	target importTarget,
) (appliedUnitCommit, error) {
	now := s.now().UTC()
	owner, ok := s.ownerCreateRegistry.Resolve(
		unit.ApprovedMapping.TargetViewSchemaID,
		target.CreateFacade,
	)
	if !ok {
		return appliedUnitCommit{}, importApplyBlockedError("owner_create_contract_unavailable")
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
		return appliedUnitCommit{}, err
	}

	mutationSequencer := ownerfacade.NewImportMutationSequencer()
	ownerResults := make([]map[string]any, 0, len(unit.SourceRows))
	for index, sourceRow := range unit.SourceRows {
		rowRef, _ := intFromAny(sourceRow["source_row_ref"])
		if rowRef <= 0 {
			return appliedUnitCommit{}, fmt.Errorf("import source row missing source_row_ref")
		}
		rowClientTxnID := fmt.Sprintf("%s:%d", clientTxnID, rowRef)
		request, err := importOwnerCreateRequest(
			ctx,
			start,
			unit,
			actor.ID,
			sourceRow,
			rowRef,
			rowClientTxnID,
			owner,
		)
		if err != nil {
			if errors.Is(err, tabularingest.ErrMappingKernelCanceled) {
				return appliedUnitCommit{}, errImportUnitCanceled
			}
			return appliedUnitCommit{}, &translatedImportUnitError{
				failure: importOwnerCreateFailure(err),
				cause:   err,
			}
		}
		response, err := owner.CreateImportRowTx(
			ctx,
			tx,
			ownerfacade.ImportOwnerCreateCommand{
				Request:           request,
				ChangeSetID:       changeSetID,
				SequenceNo:        index + 1,
				MutationSequencer: mutationSequencer,
				Now:               now,
			},
		)
		if err != nil {
			return appliedUnitCommit{}, &translatedImportUnitError{
				failure: importOwnerCreateFailure(err),
				cause:   err,
			}
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
			return appliedUnitCommit{}, err
		}
		ownerResults = append(ownerResults, map[string]any{
			"owner_response": ownerResponse,
			"row_refresh":    response.RowRefresh,
		})
	}
	return appliedUnitCommit{
		OwnerResult:  ownerResults,
		ResourceRefs: []jobs.ResourceRef{},
		ChangeSetID:  &changeSetID,
	}, nil
}

func importOwnerCreateRequest(
	ctx context.Context,
	start ApplyStartResult,
	unit ApplyUnitData,
	actorID uuid.UUID,
	sourceRow map[string]any,
	rowRef int,
	clientTxnID string,
	owner ownerfacade.ImportOwnerCreateFacade,
) (ownerfacade.ImportOwnerCreateRequest, error) {
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
	kernelRequest, err := importMappingKernelRequest(unit, sourceRow, rowRef)
	if err != nil {
		return ownerfacade.ImportOwnerCreateRequest{}, err
	}
	kernelPlan, err := tabularingest.BuildMappingKernelPlanV1(ctx, kernelRequest)
	if err != nil {
		return ownerfacade.ImportOwnerCreateRequest{}, err
	}
	if len(kernelPlan.Rows) != 1 {
		return ownerfacade.ImportOwnerCreateRequest{}, fmt.Errorf("import mapping kernel returned %d rows", len(kernelPlan.Rows))
	}
	for _, planned := range kernelPlan.Rows[0].Values {
		if planned.Disposition == tabularingest.MappingKernelOmitV1 {
			continue
		}
		value, include, err := owner.NormalizeImportField(
			planned.FieldKey,
			planned.TransformedValue,
			planned.EmptyValuePolicy,
		)
		if err != nil {
			return ownerfacade.ImportOwnerCreateRequest{}, err
		}
		if !include {
			continue
		}
		request.FieldValues = append(request.FieldValues, ownerfacade.ImportFieldValue{
			FieldKey:            planned.FieldKey,
			NormalizedValue:     value,
			SourceColumnOrdinal: planned.SourceColumnOrdinal,
			SourceHeaderText:    planned.SourceHeaderText,
			RawValue:            planned.RawValue,
			CellKind:            planned.CellKind,
			TransformID:         planned.TransformID,
			EmptyValuePolicy:    planned.EmptyValuePolicy,
			EntityBindingMode:   planned.EntityBindingMode,
		})
	}
	for _, unknown := range kernelPlan.Rows[0].UnknownValues {
		request.UnknownValues = append(request.UnknownValues, ownerfacade.ImportUnknownValue{
			SourceColumnOrdinal: unknown.SourceColumnOrdinal,
			SourceHeaderText:    unknown.SourceHeaderText,
			RawValue:            unknown.RawValue,
			CellKind:            unknown.CellKind,
		})
	}
	return request, nil
}

func importMappingKernelRequest(
	unit ApplyUnitData,
	sourceRow map[string]any,
	rowRef int,
) (tabularingest.MappingKernelRequestV1, error) {
	schema, ok := viewschema.Lookup(unit.ApprovedMapping.TargetViewSchemaID)
	if !ok {
		return tabularingest.MappingKernelRequestV1{}, fmt.Errorf("import mapping kernel target view is unavailable")
	}
	public, ok := viewschema.LookupPublicResource(unit.ApprovedMapping.TargetViewSchemaID)
	if !ok {
		return tabularingest.MappingKernelRequestV1{}, fmt.Errorf("import mapping kernel target view resource is unavailable")
	}
	fields := schema.Fields()
	targetFields := make([]tabularingest.MappingKernelTargetFieldV1, 0, len(public.Fields))
	for index, publicField := range public.Fields {
		field, exists := fields[publicField.FieldKey]
		if !exists {
			return tabularingest.MappingKernelRequestV1{}, fmt.Errorf("import mapping kernel field %q is unavailable", publicField.FieldKey)
		}
		targetFields = append(targetFields, tabularingest.MappingKernelTargetFieldV1{
			FieldKey:          field.FieldKey,
			Order:             index + 1,
			Writable:          field.Writable,
			CreateWritable:    field.CreateWritable,
			Clearable:         field.Clearable,
			EntityBindingMode: field.EntityBindingMode,
		})
	}
	columns := make([]tabularingest.MappingKernelSourceColumnV1, 0, len(unit.ApprovedMapping.SourceColumns))
	for _, column := range unit.ApprovedMapping.SourceColumns {
		columns = append(columns, tabularingest.MappingKernelSourceColumnV1{
			SourceColumnOrdinal: column.SourceColumnOrdinal,
			SourceHeaderText:    column.SourceHeaderText,
			FieldKey:            column.FieldKey,
			EntityBindingMode:   column.EntityBindingMode,
			TransformID:         column.TransformID,
			TransformOptions:    column.TransformOptions,
			EmptyValuePolicy:    column.EmptyValuePolicy,
		})
	}
	cellsByOrdinal := sourceRowCellsByOrdinal(sourceRow)
	cells := make([]tabularingest.MappingKernelScalarCellV1, 0, len(columns))
	for _, column := range columns {
		cell, exists := cellsByOrdinal[column.SourceColumnOrdinal]
		if !exists {
			return tabularingest.MappingKernelRequestV1{}, fmt.Errorf("import source row %d is missing column %d", rowRef, column.SourceColumnOrdinal)
		}
		rawValue, rawOK := cell["display_text"].(string)
		cellKind, kindOK := cell["cell_kind"].(string)
		if !rawOK || !kindOK || cellKind == "" {
			return tabularingest.MappingKernelRequestV1{}, fmt.Errorf("import source row %d column %d is not a scalar cell", rowRef, column.SourceColumnOrdinal)
		}
		classification := tabularingest.MappingKernelCellScalarV1
		if cellKind == "blank" || rawValue == "" {
			classification = tabularingest.MappingKernelCellEmptyV1
		}
		cells = append(cells, tabularingest.MappingKernelScalarCellV1{
			SourceColumnOrdinal: column.SourceColumnOrdinal,
			RawValue:            rawValue,
			CellKind:            cellKind,
			Classification:      classification,
			Present:             true,
		})
	}
	return tabularingest.MappingKernelRequestV1{
		TargetFields:        targetFields,
		SourceColumns:       columns,
		Rows:                []tabularingest.MappingKernelSourceRowV1{{SourceRowOrdinal: rowRef, Cells: cells}},
		UnknownColumnPolicy: unit.ApprovedMapping.UnknownColumnPolicy,
	}, nil
}
