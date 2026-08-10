package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/deleterestore"
	taskdecisionrollback "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/rollback"
)

func NewRevisionContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerTasksDecisions,
		Records: []revisions.RecordProviderContribution{
			{
				SourceOwnerModule:   revisions.SourceOwnerTasksDecisions,
				RecordType:          "task_request",
				SnapshotSchemaID:    "cartulary.revisions.snapshot.task_request.v1",
				DeleteRestoreSource: deleterestore.NewTaskRequestSource(),
				RowRollbackProvider: taskdecisionrollback.NewTaskRequestProvider(),
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "tasksdecisions.task_requests",
					ViewSchemaIDs:  []string{TaskRequestsViewSchemaID},
				}},
			},
			{
				SourceOwnerModule:   revisions.SourceOwnerTasksDecisions,
				RecordType:          "decision",
				SnapshotSchemaID:    "cartulary.revisions.snapshot.decision.v1",
				DeleteRestoreSource: deleterestore.NewDecisionSource(),
				RowRollbackProvider: taskdecisionrollback.NewDecisionProvider(),
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "tasksdecisions.decisions",
					ViewSchemaIDs:  []string{DecisionsViewSchemaID},
				}},
			},
		},
	}
}
