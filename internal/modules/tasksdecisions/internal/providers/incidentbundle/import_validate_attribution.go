package incidentbundle

import (
	"context"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

func (p preparedTasksDecisionsImport) validateAttributions(ctx context.Context, input sourceport.ImportContext) error {
	if input.Attributions == nil {
		return tasksDecisionsInvariantFailure(policy.InvariantAttribution)
	}
	for _, table := range []string{"task_requests", "decisions"} {
		var expectedCount, actualCount uint64
		check := func(recordID uuid.UUID, owner *uuid.UUID) error {
			present := owner != nil
			if present {
				expectedCount++
			}
			if err := input.Attributions.VisitImportedAttributions(table, recordID.String(), func(actual incidentportability.ImportedAttribution) error {
				if !present || actual.SourceTable != table || actual.SourceRowID != recordID.String() ||
					actual.SourceColumn != "owner_user_id" || actual.SourceActorID != owner.String() || actual.LocalUserID != p.actorID {
					return tasksDecisionsInvariantFailure(policy.InvariantAttribution)
				}
				present = false
				return ctx.Err()
			}); err != nil {
				return err
			}
			if present {
				return tasksDecisionsInvariantFailure(policy.InvariantAttribution)
			}
			return nil
		}
		var err error
		if table == "task_requests" {
			err = p.walkTasks(ctx, input, func(row portableTaskRequest) error { return check(row.RecordID, row.PortableOwnerUserID) })
		} else {
			err = p.walkDecisions(ctx, input, func(row portableDecision) error { return check(row.RecordID, &row.PortableOwnerUserID) })
		}
		if err != nil {
			return err
		}
		if err := input.Attributions.VisitImportedAttributions(table, "", func(incidentportability.ImportedAttribution) error {
			actualCount++
			return ctx.Err()
		}); err != nil {
			return err
		}
		if actualCount != expectedCount {
			return tasksDecisionsInvariantFailure(policy.InvariantAttribution)
		}
	}
	return nil
}
