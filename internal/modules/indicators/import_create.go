package indicators

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
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
	if beforeRow != nil {
		createdOrReused = "reused"
		resultCode = "reused"
		beforeVersion := record.RowVersion
		if jsonEqual(beforeRow, row) {
			operation = "reuse"
		} else {
			resultCode = "updated"
			if record.RowVersion > 1 {
				beforeVersion = record.RowVersion - 1
			}
		}
		value := entityVersionID("indicator", record.RecordID, beforeVersion)
		beforeVersionID = &value
	}
	afterSnapshot, err := s.revisions.CaptureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	afterVersionID := entityVersionID("indicator", record.RecordID, record.RowVersion)
	if err := s.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		TargetKind:      "indicator",
		RecordID:        record.RecordID,
		OperationKind:   operation,
		BeforeVersionID: beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if beforeRow == nil || !jsonEqual(beforeRow, row) {
		changedFieldKeys := indicatorChangedFieldKeys(beforeRow, row)
		if err := s.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
			ChangeSetID:    command.ChangeSetID,
			RecordID:       record.RecordID,
			RowVersion:     record.RowVersion,
			BeforeSnapshot: beforeSnapshot,
			AfterSnapshot:  &afterSnapshot,
			ConflictFacts:  indicatorRevisionFacts(beforeRow, row, changedFieldKeys),
		}); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := appendIndicatorPublicationTx(
			ctx, tx, s.publications, request.IncidentID, request.ActorUserID,
			request.ClientTxnID, command.ChangeSetID, record.RecordID,
			record.RowVersion, max(command.SequenceNo-1, 0), command.Now,
			row, changedFieldKeys,
		); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
	}
	return ownerfacade.ImportOwnerCreateResponse{
		RecordID:             record.RecordID,
		RowVersion:           record.RowVersion,
		ChangeSetMutationRef: fmt.Sprintf("change_set_mutation:%s:%d", command.ChangeSetID, command.SequenceNo),
		CreatedOrReused:      createdOrReused,
		OwnerResultCode:      resultCode,
		RowRefresh:           row,
	}, nil
}

func indicatorImportCreateCommand(clientTxnID string, fields []ownerfacade.ImportFieldValue) CreateCommand {
	command := CreateCommand{ClientTxnID: clientTxnID}
	for _, field := range fields {
		value, ok := field.NormalizedValue.Text()
		if !ok {
			continue
		}
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
