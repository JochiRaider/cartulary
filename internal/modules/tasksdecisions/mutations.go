package tasksdecisions

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
)

const TaskDecisionRecordFieldKey = policy.TaskDecisionRecordField

type FieldValue = policy.FieldValue
type TaskCreateParams = policy.TaskCreateParams
type DecisionCreateParams = policy.DecisionCreateParams
type LifecycleValidationError = policy.LifecycleValidationError
type ValidationError = policy.ValidationError
type TaskLifecycleState = policy.TaskLifecycleState
type DecisionMachineState = policy.DecisionMachineState

func applyTaskDirectChangeTx(
	ctx context.Context,
	tx pgx.Tx,
	linkStore LinkCapability,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	actorID uuid.UUID,
	fieldKey string,
	value FieldValue,
	now time.Time,
) (bool, []links.RecordLinkMutation, error) {
	scalarChanged, err := tasksource.ApplyTaskDirectChangeTx(ctx, tx, recordID, fieldKey, value, now)
	if err != nil {
		return false, nil, err
	}
	if fieldKey != TaskDecisionRecordFieldKey {
		return scalarChanged, nil, nil
	}
	linkMutations, err := syncTaskDecisionReferenceTx(ctx, tx, linkStore, incidentID, recordID, actorID, value.UUID, now)
	if err != nil {
		return false, nil, err
	}
	return scalarChanged || len(linkMutations) > 0, linkMutations, nil
}

func syncTaskDecisionReferenceTx(
	ctx context.Context,
	tx pgx.Tx,
	linkStore interface {
		SyncFieldReferenceWithMutationValuesTx(context.Context, pgx.Tx, links.SyncFieldReferenceCommand) (links.CollectionMutationResult, error)
	},
	incidentID uuid.UUID,
	recordID uuid.UUID,
	actorID uuid.UUID,
	targetID *uuid.UUID,
	now time.Time,
) ([]links.RecordLinkMutation, error) {
	result, err := linkStore.SyncFieldReferenceWithMutationValuesTx(ctx, tx, links.SyncFieldReferenceCommand{
		IncidentID:  incidentID,
		SrcRecordID: recordID,
		TargetID:    targetID,
		FieldKey:    TaskDecisionRecordFieldKey,
		LinkType:    links.LinkType(links.LinkTypeReferencesRecord),
		ActorUserID: actorID,
		Now:         now,
	})
	if err != nil {
		return nil, err
	}
	return result.RecordLinks, nil
}

func validateDecisionMachineConsistentTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) error {
	state, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	return policy.ValidateDecisionMachineState(state)
}
