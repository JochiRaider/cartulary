package incidentbundle

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/linkfacts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
)

func NewSourcePort(linkFacts linkfacts.Capability) (sourceport.Port, error) {
	if linkFacts == nil {
		return nil, fmt.Errorf("tasks/decisions incident bundle source: Links facts are required")
	}
	descriptor := tasksDecisionsSourceDescriptor()
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: exportIncidentBundleFiles,
		Prepare: func(ctx context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareTasksDecisionsImport(ctx, bundle, importContext)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedTasksDecisionsImport)
			if !ok {
				return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
			}
			return applyPreparedTasksDecisionsImportTx(ctx, tx, prepared, importContext)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedTasksDecisionsImport)
			if !ok {
				return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
			}
			return validatePreparedTasksDecisionsImportTx(ctx, tx, linkFacts, prepared, importContext)
		},
	}), nil
}

func tasksDecisionsSourceDescriptor() sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "tasks_decisions", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.tasksdecisions", OwnerRelationIDs: []string{"tasks-and-decisions"},
		Dependencies: []string{"artifacts"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/task_requests.ndjson", ContentRole: "source_rows", Versions: []int{6}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "tasks_decisions.source_identity_admitted"},
			{LogicalPath: "data/decisions.ndjson", ContentRole: "source_rows", Versions: []int{6}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "tasks_decisions.source_identity_admitted"},
		},
		InvariantIDs: policy.PortabilityInvariantIDs(),
	}
}

func validatePreparedTasksDecisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	linkFacts linkfacts.Capability,
	prepared preparedTasksDecisionsImport,
	importContext sourceport.ImportContext,
) error {

	validateEnvelope := func(id uuid.UUID, kind string) error {
		valid, err := tasksource.EnvelopeValidTx(ctx, tx, id, importContext.IncidentID, kind)
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
		}
		return nil
	}
	if err := prepared.walkTasks(ctx, importContext, func(row portableTaskRequest) error { return validateEnvelope(row.RecordID, "task_request") }); err != nil {
		return err
	}
	if err := prepared.walkDecisions(ctx, importContext, func(row portableDecision) error { return validateEnvelope(row.RecordID, "decision") }); err != nil {
		return err
	}

	if err := prepared.walkDecisions(ctx, importContext, func(row portableDecision) error {
		recordID := row.RecordID
		state, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, linkFacts, recordID)
		if err != nil {
			return err
		}
		if err := policy.ValidateDecisionMachineState(state); err != nil {
			return tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
		}
		valid, err := tasksource.SupersessionRelationsValidTx(ctx, tx, linkFacts, recordID, importContext.IncidentID)
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
		}
		return nil
	}); err != nil {
		return err
	}

	if err := prepared.walkTasks(ctx, importContext, func(row portableTaskRequest) error {
		recordID := row.RecordID
		state, err := tasksource.LoadTaskLifecycleStateTx(ctx, tx, recordID)
		if err != nil {
			return err
		}
		if !policy.ValidTaskStatus(state.Status) {
			return tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
		}
		if err := policy.ValidateTaskState(state); err != nil {
			return tasksDecisionsInvariantFailure("tasks_decisions.dependent_fields_legal")
		}
		return nil
	}); err != nil {
		return err
	}

	if err := prepared.walkTasks(ctx, importContext, func(row portableTaskRequest) error {
		valid, err := portableTaskReferencesValidTx(ctx, tx, linkFacts, row, importContext.IncidentID)
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.references_same_incident")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := prepared.walkDecisions(ctx, importContext, func(row portableDecision) error {
		valid, err := tasksource.OwnedLinksValidTx(ctx, tx, linkFacts, row.RecordID, importContext.IncidentID, "decision")
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.references_same_incident")
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func portableTaskReferencesValidTx(
	ctx context.Context,
	tx pgx.Tx,
	linkFacts linkfacts.Capability,
	row portableTaskRequest,
	incidentID uuid.UUID,
) (bool, error) {
	if row.RequesterPartyID != nil {
		valid, err := tasksource.TargetValidTx(ctx, tx, *row.RequesterPartyID, incidentID, "party")
		if err != nil || !valid {
			return valid, err
		}
	}
	if row.DecisionRecordID != nil {
		valid, err := tasksource.TargetValidTx(ctx, tx, *row.DecisionRecordID, incidentID, "decision")
		if err != nil || !valid {
			return valid, err
		}
		valid, err = tasksource.TaskDecisionFieldLinkValidTx(ctx, tx, linkFacts, row.RecordID, *row.DecisionRecordID, incidentID)
		if err != nil || !valid {
			return valid, err
		}
	}
	return tasksource.OwnedLinksValidTx(ctx, tx, linkFacts, row.RecordID, incidentID, "task_request")
}
