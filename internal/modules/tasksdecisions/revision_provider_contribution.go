package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/rollbackprovider"
)

func NewRevisionContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerTasksDecisions,
		Records: []revisions.RecordProviderContribution{
			{
				SourceOwnerModule:      revisions.SourceOwnerTasksDecisions,
				RecordType:             "task_request",
				DeleteRestoreSource:    deleterestore.NewTaskRequestSource(),
				RowRollbackProvider:    rollbackprovider.NewTaskRequestProvider(),
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
				RowRollbackProvider:    rollbackprovider.NewDecisionProvider(),
				LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
				RecordViewRoutes: []revisions.RecordViewRouteContribution{{
					ContributionID: "tasksdecisions.decisions",
					ViewSchemaIDs:  []string{DecisionsViewSchemaID},
				}},
			},
		},
	}
}
