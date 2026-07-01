package evidence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type ImportCreateCommand struct {
	Request     tabularingest.ImportOwnerCreateRequest
	ChangeSetID uuid.UUID
	SequenceNo  int
	Now         time.Time
}

func (s *Store) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand) (tabularingest.ImportOwnerCreateResponse, error) {
	request := command.Request
	if request.TargetViewSchemaID != evidenceViewSchemaID {
		return tabularingest.ImportOwnerCreateResponse{}, fmt.Errorf("evidence import surface %q not mapped", request.TargetViewSchemaID)
	}
	params := WorkbookCreateParams{Values: evidenceValuesFromImport(ownerfacade.ValuesByField(request.FieldValues))}
	if err := ValidateWorkbookCreateParams(params); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	now := command.Now.UTC()
	recordID, err := records.NewStore().InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      request.IncidentID,
		RecordType:      "evidence",
		CreatedByUserID: request.ActorUserID,
		CreatedAt:       now,
		UpdatedByUserID: request.ActorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if err := s.InsertWorkbookRowTx(ctx, tx, recordID, request.IncidentID, params, now); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	row, err := s.RefreshImportRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeTx(ctx, tx, s.revisionStore, ownerfacade.FinalizeCommand{
		Request:         request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        recordID,
		Operation:       "create",
		CreatedOrReused: "created",
		OwnerResultCode: "created",
		Row:             row,
	})
}

func evidenceValuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]WorkbookFieldValue {
	result := make(map[string]WorkbookFieldValue, len(values))
	for field, value := range values {
		result[field] = WorkbookFieldValue{
			Text:      value.Text,
			Timestamp: value.Timestamp,
			UUID:      value.UUID,
			Number:    value.Number,
			Bool:      value.Bool,
		}
	}
	return result
}
