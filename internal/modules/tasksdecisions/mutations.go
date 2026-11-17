package tasksdecisions

import (
	"context"
	"errors"
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

// validateDecisionMachineConsistentTx fails closed while the persisted Decision
// is inconsistent. requestedStatus, when known, is the status the request would
// produce; it becomes to_status (REQ-01-231).
func validateDecisionMachineConsistentTx(ctx context.Context, tx pgx.Tx, linkFacts LinkFactsCapability, recordID uuid.UUID, requestedStatus string) error {
	state, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, linkFacts, recordID)
	if err != nil {
		return err
	}
	err = policy.ValidateDecisionMachineState(state)
	var lifecycleErr *policy.LifecycleValidationError
	if requestedStatus != "" && errors.As(err, &lifecycleErr) {
		withRequested := *lifecycleErr
		withRequested.ToStatus = requestedStatus
		return &withRequested
	}
	return err
}

// requestedDecisionStatus returns the status a patch requests, or "" when the
// patch leaves decision.status untouched.
func requestedDecisionStatus(changes []PatchChange) string {
	for _, change := range changes {
		if change.FieldKey == "decision.status" && change.Value != nil && change.Value.Text != nil {
			return *change.Value.Text
		}
	}
	return ""
}
