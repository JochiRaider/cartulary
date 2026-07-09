package tasksdecisions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type ImportCreateCommand struct {
	Request     tabularingest.ImportOwnerCreateRequest
	ChangeSetID uuid.UUID
	SequenceNo  int
	Now         time.Time
}

func (s *Store) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand) (tabularingest.ImportOwnerCreateResponse, error) {
	request := command.Request
	values := taskDecisionValuesFromImport(ownerfacade.ValuesByField(request.FieldValues))
	now := command.Now.UTC()
	switch request.TargetViewSchemaID {
	case taskRequestsImportViewSchemaID:
		params := TaskCreateParams{Values: values}
		if err := ValidateTaskCreateParams(params); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		recordID, err := records.NewStore().InsertTx(ctx, tx, records.InsertParams{
			IncidentID:      request.IncidentID,
			RecordType:      "task_request",
			CreatedByUserID: request.ActorUserID,
			CreatedAt:       now,
			UpdatedByUserID: request.ActorUserID,
			UpdatedAt:       now,
			RowVersion:      1,
		})
		if err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		if err := s.InsertTaskRequestTx(ctx, tx, recordID, request.IncidentID, request.ActorUserID, params, now); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		if _, err := s.linkStore.SyncFieldReferenceCommandTx(ctx, tx, links.SyncFieldReferenceCommand{
			IncidentID:  request.IncidentID,
			SrcRecordID: recordID,
			TargetID:    values[TaskDecisionRecordFieldKey].UUID,
			FieldKey:    TaskDecisionRecordFieldKey,
			LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
			ActorUserID: request.ActorUserID,
			Now:         now,
		}); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		return s.finalizeImportRowTx(ctx, tx, command, recordID)
	case decisionsViewSchemaID:
		params := DecisionCreateParams{Values: values}
		if err := ValidateDecisionCreateParams(params); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		recordID, err := records.NewStore().InsertTx(ctx, tx, records.InsertParams{
			IncidentID:      request.IncidentID,
			RecordType:      "decision",
			CreatedByUserID: request.ActorUserID,
			CreatedAt:       now,
			UpdatedByUserID: request.ActorUserID,
			UpdatedAt:       now,
			RowVersion:      1,
		})
		if err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		if err := s.InsertDecisionTx(ctx, tx, recordID, request.IncidentID, request.ActorUserID, params, now); err != nil {
			return tabularingest.ImportOwnerCreateResponse{}, err
		}
		return s.finalizeImportRowTx(ctx, tx, command, recordID)
	default:
		return tabularingest.ImportOwnerCreateResponse{}, fmt.Errorf("tasks/decisions import surface %q not mapped", request.TargetViewSchemaID)
	}
}

func (s *Store) finalizeImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand, recordID uuid.UUID) (tabularingest.ImportOwnerCreateResponse, error) {
	row, err := s.RefreshImportRowTx(ctx, tx, command.Request.TargetViewSchemaID, recordID)
	if err != nil {
		return tabularingest.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeTx(ctx, tx, revisions.NewStore(), ownerfacade.FinalizeCommand{
		Request:         command.Request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        recordID,
		Operation:       "create",
		CreatedOrReused: "created",
		OwnerResultCode: "created",
		Row:             row,
	})
}

func taskDecisionValuesFromImport(values map[string]tabularingest.ImportScalarValue) map[string]FieldValue {
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
