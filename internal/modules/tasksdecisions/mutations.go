package tasksdecisions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

const taskDecisionRecordFieldKey = policy.TaskDecisionRecordField

type FieldValue = policy.FieldValue
type LifecycleValidationError = policy.LifecycleValidationError
type ValidationError = policy.ValidationError

func applyTaskDirectChangeTx(
	ctx context.Context,
	tx pgx.Tx,
	catalog *sourcecatalog.Catalog,
	linkStore LinkCapability,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	actorID uuid.UUID,
	fieldKey string,
	value FieldValue,
	now time.Time,
) (bool, []links.Mutation, error) {
	scalarChanged, err := tasksource.ApplyTaskDirectChangeTx(ctx, tx, catalog, recordID, fieldKey, value, now)
	if err != nil {
		return false, nil, err
	}
	if fieldKey != taskDecisionRecordFieldKey {
		return scalarChanged, nil, nil
	}
	linkMutations, err := syncTaskDecisionReferenceTx(ctx, tx, catalog, linkStore, incidentID, recordID, actorID, value.UUID, now)
	if err != nil {
		return false, nil, err
	}
	return scalarChanged || len(linkMutations) > 0, linkMutations, nil
}

func syncTaskDecisionReferenceTx(
	ctx context.Context,
	tx pgx.Tx,
	catalog *sourcecatalog.Catalog,
	linkStore interface {
		SyncFieldReferenceWithMutationValuesTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (links.CollectionMutationResult, error)
	},
	incidentID uuid.UUID,
	recordID uuid.UUID,
	actorID uuid.UUID,
	targetID *uuid.UUID,
	now time.Time,
) ([]links.Mutation, error) {
	field, ok := catalog.Field(taskDecisionRecordFieldKey)
	if !ok || field.Reference.MirrorLinkType == "" {
		return nil, &ValidationError{Field: taskDecisionRecordFieldKey, ReasonCode: "unsupported_field_key"}
	}
	linkType, err := links.ParseLinkType(field.Reference.MirrorLinkType)
	if err != nil {
		return nil, err
	}
	result, err := linkStore.SyncFieldReferenceWithMutationValuesTx(ctx, tx, links.SyncFieldReferenceCommand{
		IncidentID:  incidentID,
		SrcRecordID: recordID,
		TargetID:    targetID,
		FieldKey:    taskDecisionRecordFieldKey,
		LinkType:    linkType,
		ActorUserID: actorID,
		Now:         now,
	})
	if err != nil {
		return nil, err
	}
	return result.Mutations(), nil
}

func validateDecisionMachineConsistentTx(ctx context.Context, tx pgx.Tx, linkFacts LinkFactsCapability, recordID uuid.UUID) error {
	state, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, linkFacts, recordID)
	if err != nil {
		return err
	}
	return policy.ValidateDecisionMachineState(state)
}
