package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/rollbackprovider"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerTasksDecisions,
		Records: []revisions.RecordProviderContribution{
			{SourceOwnerModule: revisions.SourceOwnerTasksDecisions, RecordType: "task_request", DeleteRestoreProvider: deleterestore.TaskRequestProvider(), RowRollbackProvider: rollbackprovider.NewTaskRequestProvider()},
			{SourceOwnerModule: revisions.SourceOwnerTasksDecisions, RecordType: "decision", DeleteRestoreProvider: deleterestore.DecisionProvider(), RowRollbackProvider: rollbackprovider.NewDecisionProvider()},
		},
	}
}
