package assessments

import (
	"github.com/JochiRaider/cartulary/internal/modules/assessments/internal/providers/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/assessments/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerAssessments,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:   revisions.SourceOwnerAssessments,
			RecordType:          "assessment",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.assessment.v1",
			HistoryTargetKinds:  []string{"assessment"},
			DeleteRestoreSource: deleterestore.NewSource(),
			RowRollbackProvider: rollback.NewProvider(),
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "assessments.assessments",
				ViewSchemaIDs:  []string{AssessmentsViewSchemaID},
			}},
		}},
	}
}
