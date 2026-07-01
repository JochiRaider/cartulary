package parties

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
	if request.TargetViewSchemaID != ViewSchemaID {
		return tabularingest.ImportOwnerCreateResponse{}, fmt.Errorf("party import surface %q not mapped", request.TargetViewSchemaID)
	}
	params := CreateParams{Values: valuesFromImport(ownerfacade.ValuesByField(request.FieldValues))}
	if err := ValidateCreateParams(params); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	now := command.Now.UTC()
	if recordID, found, err := s.FindReusablePartyTx(ctx, tx, request.IncidentID, params); err != nil || found {
		if err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		row, err := s.RefreshImportRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
		if err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		return ownerfacade.FinalizeTx(ctx, tx, s.revisions(), ownerfacade.FinalizeCommand{
			Request:         request,
			ChangeSetID:     command.ChangeSetID,
			SequenceNo:      command.SequenceNo,
			RecordID:        recordID,
			Operation:       "reuse",
			CreatedOrReused: "reused",
			OwnerResultCode: "reused",
			Row:             row,
		})
	}
	recordID, err := s.records().InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      request.IncidentID,
		RecordType:      "party",
		CreatedByUserID: request.ActorUserID,
		CreatedAt:       now,
		UpdatedByUserID: request.ActorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	if err := s.InsertPartyTx(ctx, tx, recordID, request.IncidentID, params, now); err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	row, err := s.RefreshImportRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeTx(ctx, tx, s.revisions(), ownerfacade.FinalizeCommand{
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

func valuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]FieldValue {
	result := make(map[string]FieldValue, len(values))
	for field, value := range values {
		result[field] = FieldValue{
			Text:      value.Text,
			Timestamp: value.Timestamp,
			UUID:      value.UUID,
			Number:    value.Number,
			Bool:      value.Bool,
		}
	}
	return result
}
