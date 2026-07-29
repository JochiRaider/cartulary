package indicators

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	pool postgres.DB,
	appender *revisions.Appender,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("indicator import surface %q not mapped", targetViewSchemaID)
	}
	store := NewStore(pool, appender)
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
	createRequest := CreateRequest{
		ClientTxnID: request.ClientTxnID,
		Values:      indicatorImportValuesByField(request.FieldValues),
	}
	record, beforeRow, operationKind, _, err := s.upsertIndicatorTx(ctx, tx, authn.UserRecord{ID: request.ActorUserID}, request.IncidentID, createRequest, command.Now.UTC())
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	projected, err := refreshIndicatorProjectionTx(ctx, tx, record.RecordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row := BuildIndicatorRow(projected)
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

func indicatorImportValuesByField(fields []ownerfacade.ImportFieldValue) map[string]string {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		if field.NormalizedValue.Text == nil {
			continue
		}
		values[field.FieldKey] = *field.NormalizedValue.Text
	}
	return values
}
