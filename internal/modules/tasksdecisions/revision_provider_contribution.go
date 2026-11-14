package tasksdecisions

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/linkfacts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/deleterestore"
	taskdecisionrollback "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

type FieldReferenceFact = linkfacts.FieldReferenceFact
type FieldReferenceReader = linkfacts.FieldReferenceReader
type SupersessionFact = linkfacts.SupersessionFact
type SupersessionReader = linkfacts.SupersessionReader

func NewRevisionContribution(linkFacts LinkFactsCapability, fieldReferences FieldReferenceReader, supersession SupersessionReader, referenceTargets ReferenceTargetCapability) (revisions.ProviderContribution, error) {
	if linkFacts == nil {
		return revisions.ProviderContribution{}, errors.New("tasks/decisions: Links facts are required")
	}
	if fieldReferences == nil {
		return revisions.ProviderContribution{}, errors.New("tasks/decisions: field reference reader is required")
	}
	if supersession == nil {
		return revisions.ProviderContribution{}, errors.New("tasks/decisions: retained supersession reader is required")
	}
	if referenceTargets == nil {
		return revisions.ProviderContribution{}, errors.New("tasks/decisions: reference targets are required")
	}
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
			HistoryProjector:  taskDecisionHistoryProjector(catalog, surface.ViewSchemaID),
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "tasksdecisions." + surface.SourceTable,
				ViewSchemaIDs:  []string{surface.ViewSchemaID},
			}},
		}
		switch surface.RecordType {
		case "task_request":
			record.DeleteRestoreSource = deleterestore.NewTaskRequestSource(referenceTargets)
			record.ValidateSnapshotSource = deleterestore.ValidateTaskRequestSnapshotSource
			record.RowRollbackProvider = taskdecisionrollback.NewTaskRequestProvider(fieldReferences, referenceTargets)
		case "decision":
			record.DeleteRestoreSource = deleterestore.NewDecisionSource(linkFacts, supersession)
			record.ValidateSnapshotSource = deleterestore.ValidateDecisionSnapshotSource
			record.RowRollbackProvider = taskdecisionrollback.NewDecisionProvider(linkFacts)
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
