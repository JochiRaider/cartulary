package artifacts

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
)

type activeUserLookup interface {
	IsActiveUserTx(context.Context, pgx.Tx, uuid.UUID) (bool, error)
}

type ImportDependencies struct {
	RecordEnvelopes recordEnvelopeInserter
	ActiveUsers     activeUserLookup
	Projections     artifactProjectionRows
	Revisions       ownerfacade.LiveRecordRevisionAppender
	Collaboration   collaboration.RecordChangedAppender
}

func (d ImportDependencies) validate() error {
	required := []struct {
		name  string
		value any
	}{
		{name: "Records insertion", value: d.RecordEnvelopes},
		{name: "Active-user lookup", value: d.ActiveUsers},
		{name: "Projection refresh/load", value: d.Projections},
		{name: "Revision and intent appender", value: d.Revisions},
		{name: "Collaboration publication appender", value: d.Collaboration},
	}
	for _, dependency := range required {
		if dependency.value == nil {
			return fmt.Errorf("artifacts import dependencies: %s is required", dependency.name)
		}
	}
	return nil
}

type artifactImportCreateAdapter struct {
	source           artifactSourceKernel
	activeUsers      activeUserLookup
	revisionAppender ownerfacade.LiveRecordRevisionAppender
	publications     collaboration.RecordChangedAppender
}

func NewImportContribution(
	targetViewSchemaID string,
	facadeID string,
	dependencies ImportDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return nil, fmt.Errorf("compose Artifacts source catalog: %w", err)
	}
	if _, ok := catalog.SurfaceByViewID(targetViewSchemaID); !ok {
		return nil, fmt.Errorf("artifact import surface %q not mapped", targetViewSchemaID)
	}
	adapter := &artifactImportCreateAdapter{
		source: artifactSourceKernel{
			records:     dependencies.RecordEnvelopes,
			rows:        newSourceStore(catalog),
			projections: dependencies.Projections,
		},
		activeUsers:      dependencies.ActiveUsers,
		revisionAppender: dependencies.Revisions,
		publications:     dependencies.Collaboration,
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
	command ownerfacade.ImportOwnerCreateCommand,
) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if !isArtifactBackedView(request.TargetViewSchemaID) {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("artifact import surface %q not mapped", request.TargetViewSchemaID)
	}
	indexed, err := ownerfacade.IndexImportFieldValues(request.FieldValues)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	values := artifactValuesFromImport(indexed)
	params := createParams{ViewSchemaID: request.TargetViewSchemaID, Values: values}
	if err := validateCreateParams(params); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	for fieldKey, value := range values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			active, err := a.activeUsers.IsActiveUserTx(ctx, tx, *value.UUID)
			if err != nil {
				return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("validate artifact import user: %w", err)
			}
			if !active {
				return ownerfacade.ImportOwnerCreateResponse{}, &ValidationError{Field: fieldKey, ReasonCode: "invalid_value"}
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
	return ownerfacade.FinalizeLiveRecordTx(ctx, tx, a.revisionAppender, a.publications, ownerfacade.FinalizeCommand{
		Request:         request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      command.SequenceNo,
		RecordID:        recordID,
		Operation:       "create",
		CreatedOrReused: "created",
		OwnerResultCode: "created",
		Row:             row,
		CreatedAt:       command.Now,
	})
}

func artifactValuesFromImport(values map[string]ownerfacade.ImportScalarValue) map[string]FieldValue {
	result := make(map[string]FieldValue, len(values))
	for field, value := range values {
		converted := FieldValue{}
		if scalar, ok := value.Text(); ok {
			converted.Text = &scalar
		}
		if scalar, ok := value.Timestamp(); ok {
			converted.Timestamp = &scalar
		}
		if scalar, ok := value.UUID(); ok {
			converted.UUID = &scalar
		}
		if scalar, ok := value.Number(); ok {
			converted.Number = &scalar
		}
		if scalar, ok := value.Bool(); ok {
			converted.Bool = &scalar
		}
		result[field] = converted
	}
	return result
}
