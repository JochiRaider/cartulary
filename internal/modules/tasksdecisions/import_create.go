package tasksdecisions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
)

const (
	taskOwnerUserFieldKey     = "task.owner_user_id"
	decisionOwnerUserFieldKey = "decision.owner_user_id"
)

type ImportRecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
}

type ImportLinkCapability interface {
	SyncFieldReferenceWithMutationValuesTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (links.CollectionMutationResult, error)
}

type ImportRevisionCapability interface {
	ownerfacade.LiveRecordRevisionAppender
	AppendNonRowMutationTx(context.Context, pgx.Tx, revisions.AppendNonRowMutationParams) error
}

type ImportDependencies struct {
	RecordEnvelopes ImportRecordEnvelopeCapability
	Links           ImportLinkCapability
	Projections     taskdecisionprojection.MutationRows
	Revisions       ImportRevisionCapability
	Collaboration   collaboration.RecordChangedAppender
}

func (d ImportDependencies) validate() error {
	if d.RecordEnvelopes == nil {
		return fmt.Errorf("tasks/decisions import dependencies: Records insert is required")
	}
	if d.Links == nil {
		return fmt.Errorf("tasks/decisions import dependencies: Links synchronization is required")
	}
	if d.Projections == nil {
		return fmt.Errorf("tasks/decisions import dependencies: Projection refresh/load is required")
	}
	if d.Revisions == nil {
		return fmt.Errorf("tasks/decisions import dependencies: Revision finalization is required")
	}
	if d.Collaboration == nil {
		return fmt.Errorf("tasks/decisions import dependencies: Collaboration publication is required")
	}
	return nil
}

type importOwner struct {
	dependencies ImportDependencies
	catalog      *sourcecatalog.Catalog
}

func NewImportContribution(
	targetViewSchemaID string,
	facadeID string,
	dependencies ImportDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return nil, fmt.Errorf("compose Tasks/Decisions source catalog: %w", err)
	}
	if _, ok := catalog.SurfaceByViewID(targetViewSchemaID); !ok {
		return nil, fmt.Errorf("tasks/decisions import surface %q not mapped", targetViewSchemaID)
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	owner := &importOwner{dependencies: dependencies, catalog: catalog}
	return ownerfacade.NewImportOwnerCreateFacade(
		ownerfacade.ImportOwnerCreateBinding{
			TargetViewSchemaID: targetViewSchemaID,
			FacadeID:           facadeID,
		},
		owner.CreateImportRowTx,
	)
}

func (o *importOwner) CreateImportRowTx(ctx context.Context, tx pgx.Tx, command ownerfacade.ImportOwnerCreateCommand) (ownerfacade.ImportOwnerCreateResponse, error) {
	request := command.Request
	if err := validateImportedOwnerShape(o.catalog, request); err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	values := taskDecisionValuesFromImport(ownerfacade.ValuesByField(request.FieldValues))
	now := command.Now.UTC()
	surface, ok := o.catalog.SurfaceByViewID(request.TargetViewSchemaID)
	if !ok {
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("tasks/decisions import surface %q not mapped", request.TargetViewSchemaID)
	}
	switch request.TargetViewSchemaID {
	case TaskRequestsViewSchemaID:
		params := policy.TaskCreateParams{Values: values}
		if err := policy.ValidateTaskCreateParams(params); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := validateImportCreateReferencesTx(ctx, tx, o.catalog, request.IncidentID, request.ActorUserID, taskOwnerUserFieldKey, values); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		recordID, err := o.dependencies.RecordEnvelopes.InsertTx(ctx, tx, records.InsertParams{
			IncidentID:      request.IncidentID,
			RecordType:      surface.RecordType,
			CreatedByUserID: request.ActorUserID,
			CreatedAt:       now,
			UpdatedByUserID: request.ActorUserID,
			UpdatedAt:       now,
			RowVersion:      1,
		})
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := tasksource.InsertTaskRequestTx(ctx, tx, recordID, request.IncidentID, request.ActorUserID, params, now); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		linkMutations, err := syncTaskDecisionReferenceTx(
			ctx, tx, o.catalog, o.dependencies.Links, request.IncidentID, recordID,
			request.ActorUserID, values[taskDecisionRecordFieldKey].UUID, now,
		)
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		return o.finalizeImportRowTx(ctx, tx, command, recordID, linkMutations)
	case DecisionsViewSchemaID:
		params := policy.DecisionCreateParams{Values: values}
		if err := policy.ValidateDecisionCreateParams(params); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := validateImportCreateReferencesTx(ctx, tx, o.catalog, request.IncidentID, request.ActorUserID, decisionOwnerUserFieldKey, values); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		recordID, err := o.dependencies.RecordEnvelopes.InsertTx(ctx, tx, records.InsertParams{
			IncidentID:      request.IncidentID,
			RecordType:      surface.RecordType,
			CreatedByUserID: request.ActorUserID,
			CreatedAt:       now,
			UpdatedByUserID: request.ActorUserID,
			UpdatedAt:       now,
			RowVersion:      1,
		})
		if err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		if err := tasksource.InsertDecisionTx(ctx, tx, recordID, request.IncidentID, request.ActorUserID, params, now); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
		return o.finalizeImportRowTx(ctx, tx, command, recordID, nil)
	default:
		return ownerfacade.ImportOwnerCreateResponse{}, fmt.Errorf("tasks/decisions import surface %q not mapped", request.TargetViewSchemaID)
	}
}

func validateImportedOwnerShape(catalog *sourcecatalog.Catalog, request ownerfacade.ImportOwnerCreateRequest) error {
	for _, field := range request.FieldValues {
		policy, ok := catalog.Field(field.FieldKey)
		if !ok || policy.ViewSchemaID != request.TargetViewSchemaID || policy.Reference.Role != "incident_member_user" {
			continue
		}
		if field.NormalizedValue.Kind != "uuid" || field.NormalizedValue.UUID == nil {
			return &ValidationError{Field: field.FieldKey, ReasonCode: "invalid_value"}
		}
	}
	return nil
}

func (o *importOwner) finalizeImportRowTx(ctx context.Context, tx pgx.Tx, command ownerfacade.ImportOwnerCreateCommand, recordID uuid.UUID, linkMutations []links.RecordLinkMutation) (ownerfacade.ImportOwnerCreateResponse, error) {
	row, err := o.refreshImportRowTx(ctx, tx, command.Request.TargetViewSchemaID, recordID)
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	firstSequence, err := command.AllocateMutationSequence(1 + len(linkMutations))
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	response, err := ownerfacade.FinalizeLiveRecordTx(ctx, tx, o.dependencies.Revisions, o.dependencies.Collaboration, ownerfacade.FinalizeCommand{
		Request:         command.Request,
		ChangeSetID:     command.ChangeSetID,
		SequenceNo:      firstSequence,
		RecordID:        recordID,
		Operation:       "create",
		CreatedOrReused: "created",
		OwnerResultCode: "created",
		Row:             row,
		CreatedAt:       command.Now,
	})
	if err != nil {
		return ownerfacade.ImportOwnerCreateResponse{}, err
	}
	for index, mutation := range linkMutations {
		if err := o.dependencies.Revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
			ChangeSetID: command.ChangeSetID, SequenceNo: firstSequence + index + 1,
			TargetKind: "record_link", TargetID: mutation.RecordLinkID.String(), OperationKind: mutation.Operation,
			BeforeValue: mutation.BeforeValue, AfterValue: mutation.AfterValue,
		}); err != nil {
			return ownerfacade.ImportOwnerCreateResponse{}, err
		}
	}
	return response, nil
}

func (o *importOwner) refreshImportRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	switch viewSchemaID {
	case TaskRequestsViewSchemaID:
		if err := o.dependencies.Projections.RefreshTaskRequestTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
		return o.dependencies.Projections.LoadTaskRequestTx(ctx, tx, recordID)
	case DecisionsViewSchemaID:
		if err := o.dependencies.Projections.RefreshDecisionTx(ctx, tx, recordID); err != nil {
			return nil, err
		}
		return o.dependencies.Projections.LoadDecisionTx(ctx, tx, recordID)
	default:
		return nil, fmt.Errorf("tasks/decisions import projection surface %q not mapped", viewSchemaID)
	}
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
	catalog *sourcecatalog.Catalog,
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
		if isMemberUserReferenceField(catalog, fieldKey) {
			continue
		}
		if err := validateDirectReferenceTx(
			ctx,
			tx,
			catalog,
			incidentID,
			fieldKey,
			*value.UUID,
		); err != nil {
			return err
		}
	}
	return nil
}
