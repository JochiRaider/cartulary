package tasksdecisions

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

const (
	taskOwnerUserFieldKey     = "task.owner_user_id"
	decisionOwnerUserFieldKey = "decision.owner_user_id"
)

func NewImportContribution(
	targetViewSchemaID string,
	facadeID string,
	appender *revisions.Appender,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if targetViewSchemaID != TaskRequestsViewSchemaID &&
		targetViewSchemaID != DecisionsViewSchemaID {
		return nil, fmt.Errorf("tasks/decisions import surface %q not mapped", targetViewSchemaID)
	}
	store := NewStore(appender)
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
	if err := validateImportedOwnerShape(request); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	values := taskDecisionValuesFromImport(ownerfacade.ValuesByField(request.FieldValues))
	now := command.Now.UTC()
	switch request.TargetViewSchemaID {
	case taskRequestsImportViewSchemaID:
		params := TaskCreateParams{Values: values}
		if err := ValidateTaskCreateParams(params); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := validateImportCreateReferencesTx(ctx, tx, request.IncidentID, request.ActorUserID, taskOwnerUserFieldKey, values); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
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
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.InsertTaskRequestTx(ctx, tx, recordID, request.IncidentID, request.ActorUserID, params, now); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
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
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		return s.finalizeImportRowTx(ctx, tx, command, recordID)
	case DecisionsViewSchemaID:
		params := DecisionCreateParams{Values: values}
		if err := ValidateDecisionCreateParams(params); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := validateImportCreateReferencesTx(ctx, tx, request.IncidentID, request.ActorUserID, decisionOwnerUserFieldKey, values); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
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
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := s.InsertDecisionTx(ctx, tx, recordID, request.IncidentID, request.ActorUserID, params, now); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		return s.finalizeImportRowTx(ctx, tx, command, recordID)
	default:
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("tasks/decisions import surface %q not mapped", request.TargetViewSchemaID)
	}
}

func validateImportedOwnerShape(request ownerfacade.ImportOwnerCreateRequest) error {
	ownerField := taskOwnerUserFieldKey
	if request.TargetViewSchemaID == DecisionsViewSchemaID {
		ownerField = decisionOwnerUserFieldKey
	}
	for _, field := range request.FieldValues {
		if field.FieldKey != ownerField {
			continue
		}
		if field.NormalizedValue.Kind != "uuid" || field.NormalizedValue.UUID == nil {
			return &ValidationError{Field: ownerField, ReasonCode: "invalid_value"}
		}
	}
	return nil
}

func (s *Store) finalizeImportRowTx(ctx context.Context, tx pgx.Tx, command ImportCreateCommand, recordID uuid.UUID) (ownerfacade.ImportOwnerCreateResponse, error) {
	row, err := s.RefreshImportRowTx(ctx, tx, command.Request.TargetViewSchemaID, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeTx(ctx, tx, s.revisionAppender, ownerfacade.FinalizeCommand{
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

func taskDecisionValuesFromImport(values map[string]ownerfacade.ImportScalarValue) map[string]FieldValue {
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

func validateImportCreateReferencesTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	ownerField string,
	values map[string]FieldValue,
) error {
	ownerUserID := actorUserID
	if value, ok := values[ownerField]; ok && value.UUID != nil {
		ownerUserID = *value.UUID
	}
	if err := validateIncidentMemberUserTx(ctx, tx, incidentID, ownerUserID, ownerField); err != nil {
		return err
	}
	for fieldKey, value := range values {
		if value.UUID == nil {
			continue
		}
		if strings.HasSuffix(fieldKey, "_user_id") {
			continue
		}
		if err := validateDirectReferenceTx(
			ctx,
			tx,
			incidentID,
			fieldKey,
			*value.UUID,
		); err != nil {
			return err
		}
	}
	return nil
}
