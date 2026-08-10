package evidence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
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
	projectionRows evidenceprojection.Rows,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != ViewSchemaID {
		return nil, fmt.Errorf("evidence import surface %q not mapped", targetViewSchemaID)
	}
	if projectionRows == nil {
		return nil, fmt.Errorf("evidence import projection rows are required")
	}
	store := NewStore(
		pool,
		WithRevisionAppender(appender),
		WithCollaborationIntents(intents),
		WithWorkbookProjections(projectionRows),
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
	recordID, err := s.source.createRecordTx(
		ctx,
		tx,
		request.IncidentID,
		request.ActorUserID,
		params,
		now,
	)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row, err := s.source.refreshRowTx(ctx, tx, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeRecordRevisionTx(ctx, tx, s.revisionStore, ownerfacade.FinalizeCommand{
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
