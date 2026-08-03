package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/deleterestore"
	taskdecisionrollback "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/rollback"
)

func NewRevisionContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule:     revisions.SourceOwnerTasksDecisions,
		ConflictFieldProvider: revisions.NewViewSchemaConflictFieldProvider(),
		Records: []revisions.RecordProviderContribution{
			{
				SourceOwnerModule:      revisions.SourceOwnerTasksDecisions,
				RecordType:             "task_request",
				DeleteRestoreSource:    deleterestore.NewTaskRequestSource(),
				RowRollbackProvider:    taskdecisionrollback.NewTaskRequestProvider(),
				LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "tasksdecisions.task_requests",
					ViewSchemaIDs:  []string{TaskRequestsViewSchemaID},
				}},
			},
			{
				SourceOwnerModule:      revisions.SourceOwnerTasksDecisions,
				RecordType:             "decision",
				DeleteRestoreSource:    deleterestore.NewDecisionSource(),
				RowRollbackProvider:    taskdecisionrollback.NewDecisionProvider(),
				LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "tasksdecisions.decisions",
					ViewSchemaIDs:  []string{DecisionsViewSchemaID},
				}},
			},
		},
	}
}
