package evidence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("evidence import surface %q not mapped", targetViewSchemaID)
	}
	store := NewStore(
		pool,
		WithRevisionAppender(appender),
		WithCollaborationIntents(intents),
	)
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
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("evidence import surface %q not mapped", request.TargetViewSchemaID)
	}
	params := WorkbookCreateParams{Values: evidenceValuesFromImport(ownerfacade.ValuesByField(request.FieldValues))}
	if err := ValidateWorkbookCreateParams(params); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if err := validateEvidenceReferencesTx(ctx, tx, request.IncidentID, params.Values); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
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
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if err := s.InsertWorkbookRowTx(ctx, tx, recordID, request.IncidentID, params, now); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row, err := s.RefreshImportRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
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

func evidenceValuesFromImport(values map[string]ownerfacade.ImportScalarValue) map[string]WorkbookFieldValue {
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
