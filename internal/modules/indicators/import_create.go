package indicators

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ImportCreateCommand struct {
	Request     tabularingest.ImportOwnerCreateRequest
	ChangeSetID uuid.UUID
	SequenceNo  int
	Now         time.Time
}

func (s *Store) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand) (tabularingest.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != ViewSchemaID {
		return tabularingest.ImportOwnerCreateResponse{}, fmt.Errorf("indicator import surface %q not mapped", request.TargetViewSchemaID)
	}
	createRequest := CreateRequest{
		ClientTxnID: request.ClientTxnID,
		Values:      indicatorImportValuesByField(request.FieldValues),
	}
	record, beforeRow, operationKind, _, err := s.upsertIndicatorTx(ctx, tx, authn.UserRecord{ID: request.ActorUserID}, request.IncidentID, createRequest, command.Now.UTC())
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	projected, err := refreshIndicatorProjectionTx(ctx, tx, record.RecordID)
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
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

func indicatorImportValuesByField(fields []tabularingest.ImportFieldValue) map[string]string {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		if field.NormalizedValue.Text == nil {
			continue
		}
		values[field.FieldKey] = *field.NormalizedValue.Text
	}
	return values
}
