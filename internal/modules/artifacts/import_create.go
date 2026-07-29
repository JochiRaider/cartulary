package artifacts

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	appender *revisions.Appender,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if !IsArtifactBackedView(targetViewSchemaID) {
		return nil, fmt.Errorf("artifact import surface %q not mapped", targetViewSchemaID)
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
	if !IsArtifactBackedView(request.TargetViewSchemaID) {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("artifact import surface %q not mapped", request.TargetViewSchemaID)
	}
	values := artifactValuesFromImport(ownerfacade.ValuesByField(request.FieldValues))
	params := CreateParams{ViewSchemaID: request.TargetViewSchemaID, Values: values}
	if err := ValidateCreateParams(params); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	for fieldKey, value := range values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateImportActiveUserTx(ctx, tx, *value.UUID, fieldKey); err != nil {
				return ownerfacade.ImportOwnerCreateResponse{}, err
			}
		}
	}
	now := command.Now.UTC()
	recordID, err := records.NewStore().InsertTx(ctx, tx, records.InsertParams{
		IncidentID:      request.IncidentID,
		RecordType:      "artifact",
		CreatedByUserID: request.ActorUserID,
		CreatedAt:       now,
		UpdatedByUserID: request.ActorUserID,
		UpdatedAt:       now,
		RowVersion:      1,
	})
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	if err := s.InsertRowTx(ctx, tx, recordID, request.IncidentID, request.ActorUserID, params, now); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	row, err := s.RefreshImportRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeTx(ctx, tx, s.revisionAppender, ownerfacade.FinalizeCommand{
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

func artifactValuesFromImport(values map[string]ownerfacade.ImportScalarValue) map[string]FieldValue {
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
		return fmt.Errorf("validate artifact import user: %w", err)
	}
	if !exists {
		return &ValidationError{Field: field, ReasonCode: "invalid_value"}
	}
	return nil
}
