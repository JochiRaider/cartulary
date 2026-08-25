package indicators

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

const indicatorImportContributionID = "indicators.import_create"

func NewImportContribution(application *Application) (ownerfacade.ImportOwnerCreateFacade, error) {
	if application == nil {
		return nil, fmt.Errorf("indicator import owner facade is required")
	}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: ViewSchemaID,
			FacadeID:           indicatorImportContributionID,
		},
		application.createImportRowTx,
	)
}

func (s *Application) createImportRowTx(ctx context.Context, tx pgx.Tx, command ownerfacade.ImportOwnerCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.ActorUserID == uuid.Nil {
		return ownerfacade.ImportOwnerCreateResponse{}, &IndicatorCreateValidationError{Field: "actor_user_id", ReasonCode: "missing_required_field"}
	}
	if request.TargetViewSchemaID != ViewSchemaID {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("indicator import surface %q not mapped", request.TargetViewSchemaID)
	}
	createCommand := indicatorImportCreateCommand(request.ClientTxnID, request.FieldValues)
	beforeSnapshot, err := s.captureIndicatorSnapshotBeforeUpsertTx(ctx, tx, request.IncidentID, createCommand)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	record, beforeRow, operationKind, _, err := s.upsertIndicatorTx(ctx, tx, request.ActorUserID, request.IncidentID, createCommand, command.Now.UTC())
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
	return ownerfacade.FinalizeLiveRecordTx(ctx, tx, s.revisions, s.publications, ownerfacade.FinalizeCommand{
		Request:         request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        record.RecordID,
		Operation:       operation,
		CreatedOrReused: createdOrReused,
		OwnerResultCode: resultCode,
		BeforeVersionID: beforeVersionID,
		BeforeValue:     beforeValue,
		BeforeSnapshot:  beforeSnapshot,
		Row:             row,
		CreatedAt:       command.Now,
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
