package tasksdecisions

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/deleterestore"
	taskdecisionrollback "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func NewRevisionContribution() (revisions.ProviderContribution, error) {
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return revisions.ProviderContribution{}, err
	}
	records := make([]revisions.RecordProviderContribution, 0, 2)
	for _, recordType := range []string{"task_request", "decision"} {
		surface, ok := catalog.SurfaceByRecordType(recordType)
		if !ok {
			return revisions.ProviderContribution{}, &ValidationError{Field: "record_type", ReasonCode: "unknown_record_type"}
		}
		record := revisions.RecordProviderContribution{
			SourceOwnerModule: revisions.SourceOwnerTasksDecisions,
			RecordType:        surface.RecordType,
			SnapshotSchemaID:  surface.RevisionSnapshotSchemaID,
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "tasksdecisions." + surface.SourceTable,
				ViewSchemaIDs:  []string{surface.ViewSchemaID},
			}},
		}
		switch surface.RecordType {
		case "task_request":
			record.DeleteRestoreSource = deleterestore.NewTaskRequestSource()
			record.RowRollbackProvider = taskdecisionrollback.NewTaskRequestProvider()
		case "decision":
			record.DeleteRestoreSource = deleterestore.NewDecisionSource()
			record.RowRollbackProvider = taskdecisionrollback.NewDecisionProvider()
		default:
			return revisions.ProviderContribution{}, &ValidationError{Field: "record_type", ReasonCode: "unknown_record_type"}
		}
		records = append(records, record)
	}
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerTasksDecisions,
		Records:           records,
	}, nil
}
