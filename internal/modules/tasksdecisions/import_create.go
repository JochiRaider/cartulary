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

func NewImportCreateFacade(
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
	values := taskDecisionValuesFromImport(ownerfacade.ValuesByField(request.FieldValues))
	now := command.Now.UTC()
	switch request.TargetViewSchemaID {
	case taskRequestsImportViewSchemaID:
		params := TaskCreateParams{Values: values}
		if err := ValidateTaskCreateParams(params); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := validateImportCreateReferencesTx(ctx, tx, request.IncidentID, values); err != nil {
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
		if err := validateImportCreateReferencesTx(ctx, tx, request.IncidentID, values); err != nil {
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
	values map[string]FieldValue,
) error {
	for fieldKey, value := range values {
		if value.UUID == nil {
			continue
		}
		if strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateImportActiveUserTx(ctx, tx, *value.UUID, fieldKey); err != nil {
				return err
			}
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

func validateImportActiveUserTx(
	ctx context.Context,
	tx pgx.Tx,
	userID uuid.UUID,
	field string,
) error {
	var exists bool
	if err := tx.QueryRow(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`,
		userID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("validate task or decision import user: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}
