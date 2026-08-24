package incidentbundle

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	tasksource "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/source"
)

func NewSourcePort() sourceport.Port {
	descriptor := tasksDecisionsSourceDescriptor()
	return sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor, Export: sourceport.QueryExport(exportIncidentBundleFiles),
		Prepare: func(_ context.Context, bundle sourceport.Bundle, importContext sourceport.ImportContext) (any, error) {
			return prepareTasksDecisionsImport(bundle, importContext)
		},
		Apply: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedTasksDecisionsImport)
			if !ok {
				return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
			}
			return applyPreparedTasksDecisionsImportTx(ctx, tx, prepared, importContext.ActorUserID, importContext.Attributions)
		},
		Validate: func(ctx context.Context, tx pgx.Tx, value any, importContext sourceport.ImportContext) error {
			prepared, ok := value.(preparedTasksDecisionsImport)
			if !ok {
				return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
			}
			return validatePreparedTasksDecisionsImportTx(ctx, tx, prepared, importContext)
		},
	})
}

func tasksDecisionsSourceDescriptor() sourceport.Descriptor {
	return sourceport.Descriptor{
		FamilyID: "tasks_decisions", ContractMajor: sourceport.ContractMajor,
		OwnerID: "module.tasksdecisions", OwnerRelationIDs: []string{"tasks-and-decisions"},
		Dependencies: []string{"artifacts"},
		Paths: []sourceport.Path{
			{LogicalPath: "data/task_requests.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "tasks_decisions.source_identity_admitted"},
			{LogicalPath: "data/decisions.ndjson", ContentRole: "source_rows", Versions: []int{3}, StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "tasks_decisions.source_identity_admitted"},
		},
		InvariantIDs: policy.PortabilityInvariantIDs(),
	}
}

type portableSourceIdentity struct {
	recordID   uuid.UUID
	recordType string
}

func validatePreparedTasksDecisionsImportTx(
	ctx context.Context,
	tx pgx.Tx,
	prepared preparedTasksDecisionsImport,
	importContext sourceport.ImportContext,
) error {
	identities := make([]portableSourceIdentity, 0, len(prepared.tasks)+len(prepared.decisions))
	for _, row := range prepared.tasks {
		identities = append(identities, portableSourceIdentity{recordID: row.RecordID, recordType: "task_request"})
	}
	for _, row := range prepared.decisions {
		identities = append(identities, portableSourceIdentity{recordID: row.RecordID, recordType: "decision"})
	}
	sort.Slice(identities, func(left, right int) bool {
		return identities[left].recordID.String() < identities[right].recordID.String()
	})
	for _, identity := range identities {
		valid, err := tasksource.EnvelopeValidTx(ctx, tx, identity.recordID, importContext.IncidentID, identity.recordType)
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.envelope_type_scope")
		}
	}

	decisionIDs := sortedPortableDecisionIDs(prepared.decisions)
	for _, recordID := range decisionIDs {
		state, err := tasksource.LoadDecisionMachineStateForUpdateTx(ctx, tx, recordID)
		if err != nil {
			return err
		}
		if err := policy.ValidateDecisionMachineState(state); err != nil {
			return tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
		}
		valid, err := tasksource.SupersessionRelationsValidTx(ctx, tx, recordID, importContext.IncidentID)
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.lifecycle_legal")
		}
	}

	taskIDs := sortedPortableTaskIDs(prepared.tasks)
	for _, recordID := range taskIDs {
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
	}

	for _, row := range sortedPortableTasks(prepared.tasks) {
		valid, err := portableTaskReferencesValidTx(ctx, tx, row, importContext.IncidentID)
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.references_same_incident")
		}
	}
	for _, row := range sortedPortableDecisions(prepared.decisions) {
		valid, err := tasksource.OwnedLinksValidTx(ctx, tx, row.RecordID, importContext.IncidentID, "decision")
		if err != nil {
			return err
		}
		if !valid {
			return tasksDecisionsInvariantFailure("tasks_decisions.references_same_incident")
		}
	}
	return nil
}

func portableTaskReferencesValidTx(
	ctx context.Context,
	tx pgx.Tx,
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
		valid, err = tasksource.TaskDecisionFieldLinkValidTx(ctx, tx, row.RecordID, *row.DecisionRecordID, incidentID)
		if err != nil || !valid {
			return valid, err
		}
	}
	return tasksource.OwnedLinksValidTx(ctx, tx, row.RecordID, incidentID, "task_request")
}

func sortedPortableTaskIDs(rows []portableTaskRequest) []uuid.UUID {
	result := make([]uuid.UUID, len(rows))
	for index, row := range rows {
		result[index] = row.RecordID
	}
	sort.Slice(result, func(left, right int) bool { return result[left].String() < result[right].String() })
	return result
}

func sortedPortableDecisionIDs(rows []portableDecision) []uuid.UUID {
	result := make([]uuid.UUID, len(rows))
	for index, row := range rows {
		result[index] = row.RecordID
	}
	sort.Slice(result, func(left, right int) bool { return result[left].String() < result[right].String() })
	return result
}

func sortedPortableTasks(rows []portableTaskRequest) []portableTaskRequest {
	result := append([]portableTaskRequest(nil), rows...)
	sort.Slice(result, func(left, right int) bool { return result[left].RecordID.String() < result[right].RecordID.String() })
	return result
}

func sortedPortableDecisions(rows []portableDecision) []portableDecision {
	result := append([]portableDecision(nil), rows...)
	sort.Slice(result, func(left, right int) bool { return result[left].RecordID.String() < result[right].RecordID.String() })
	return result
}
