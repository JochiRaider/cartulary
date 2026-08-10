package artifacts

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

type ImportCreateCommand = ownerfacade.ImportOwnerCreateCommand

type artifactImportCreateAdapter struct {
	source           artifactSourceKernel
	revisionAppender ownerfacade.CapturedRevisionAppender
}

func NewImportCreateFacade(
	targetViewSchemaID string,
	facadeID string,
	appender *revisions.Appender,
	projectionRows artifactprojection.Rows,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if !IsArtifactBackedView(targetViewSchemaID) {
		return nil, fmt.Errorf("artifact import surface %q not mapped", targetViewSchemaID)
	}
	if projectionRows == nil {
		return nil, fmt.Errorf("artifact import projection rows are required")
	}
	adapter := &artifactImportCreateAdapter{
		source: artifactSourceKernel{
			records:     records.NewStore(),
			rows:        newSourceStore(appender),
			projections: projectionRows,
		},
		revisionAppender: appender,
	}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		adapter.createImportRowTx,
	)
}

func (a *artifactImportCreateAdapter) createImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	command ImportCreateCommand,
) (ownerfacade.ImportOwnerCreateResponse, error) {
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
	recordID, err := a.source.createRecordTx(
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
	row, err := a.source.refreshRowTx(ctx, tx, request.TargetViewSchemaID, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	return ownerfacade.FinalizeCapturedTx(ctx, tx, a.revisionAppender, ownerfacade.FinalizeCommand{
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
