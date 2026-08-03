package indicators

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	store *Store,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("indicator import surface %q not mapped", targetViewSchemaID)
	}
	if store == nil {
		return nil, fmt.Errorf("indicator import owner facade is required")
	}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		store.CreateImportRowTx,
	)
}

func (s *Store) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != ViewSchemaID {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("indicator import surface %q not mapped", request.TargetViewSchemaID)
	}
	createCommand := indicatorImportCreateCommand(request.ClientTxnID, request.FieldValues)
	record, beforeRow, operationKind, _, err := s.upsertIndicatorTx(ctx, tx, authn.UserRecord{ID: request.ActorUserID}, request.IncidentID, createCommand, command.Now.UTC())
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row, err := s.refreshAndLoadProjectionRowTx(ctx, tx, record.RecordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	createdOrReused := "created"
	resultCode := "created"
	operation := operationKind
	var beforeVersionID *string
	var beforeValue map[string]any
	if beforeRow != nil {
		createdOrReused = "reused"
		resultCode = "reused"
		if jsonEqual(beforeRow, row) {
			operation = "reuse"
		} else {
			resultCode = "updated"
			beforeValue = beforeRow
			if record.RowVersion > 1 {
				value := ownerfacade.VersionID(record.RecordID, record.RowVersion-1)
				beforeVersionID = &value
			}
		}
	}
	return ownerfacade.FinalizeTx(ctx, tx, s.revisionsStore, ownerfacade.FinalizeCommand{
		Request:         request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        record.RecordID,
		Operation:       operation,
		CreatedOrReused: createdOrReused,
		OwnerResultCode: resultCode,
		BeforeVersionID: beforeVersionID,
		BeforeValue:     beforeValue,
		Row:             row,
	})
}

func indicatorImportCreateCommand(clientTxnID string, fields []ownerfacade.ImportFieldValue) CreateCommand {
	command := CreateCommand{ClientTxnID: clientTxnID}
	for _, field := range fields {
		if field.NormalizedValue.Text == nil {
			continue
		}
		value := *field.NormalizedValue.Text
		switch field.FieldKey {
		case "indicator.indicator_type":
			command.IndicatorType = value
		case "indicator.value_kind":
			command.ValueKind = value
		case "indicator.display_value":
			command.DisplayValue = value
		case "indicator.normalized_value":
			command.NormalizedValue = &value
		case "indicator.defanged_value":
			command.DefangedValue = &value
		case "indicator.hash_algorithm":
			command.HashAlgorithm = &value
		case "indicator.hash_value":
			command.HashValue = &value
		case "indicator.stix_pattern":
			command.STIXPattern = &value
		}
	}
	return command
}
