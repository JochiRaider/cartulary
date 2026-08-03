package assessments

import (
	"github.com/JochiRaider/cartulary/internal/modules/assessments/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/assessments/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule:     revisions.SourceOwnerAssessments,
		ConflictFieldProvider: revisions.NewViewSchemaConflictFieldProvider(),
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:      revisions.SourceOwnerAssessments,
			RecordType:             "assessment",
			DeleteRestoreSource:    deleterestore.NewSource(),
			RowRollbackProvider:    rollbackprovider.NewProvider(),
			LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "assessments.assessments",
				ViewSchemaIDs:  []string{AssessmentsViewSchemaID},
			}},
		}},
	}
}
