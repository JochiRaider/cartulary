package artifacts

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerArtifacts,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:      revisions.SourceOwnerArtifacts,
			RecordType:             "artifact",
			DeleteRestoreProvider:  deleterestore.NewProvider(),
			RowRollbackProvider:    rollbackprovider.NewProvider(),
			LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
			RecordViewRoutes: []revisions.RecordViewRouteContribution{
				recordViewRoute("artifacts.comm_log", CommLogViewSchemaID),
				recordViewRoute("artifacts.findings", FindingsViewSchemaID),
				recordViewRoute("artifacts.forensic_keywords", ForensicKeywordsViewSchemaID),
				recordViewRoute("artifacts.handoff", HandoffViewSchemaID),
				recordViewRoute("artifacts.investigative_queries", InvestigativeQueriesViewSchemaID),
				recordViewRoute("artifacts.lesson", LessonViewSchemaID),
				recordViewRoute("artifacts.notes", NotesViewSchemaID),
				recordViewRoute("artifacts.status_review", StatusReviewViewSchemaID),
			},
		}},
	}
}

func recordViewRoute(contributionID string, viewSchemaID string) revisions.RecordViewRouteContribution {
	return revisions.RecordViewRouteContribution{
		ContributionID: contributionID,
		Variant: &revisions.RecordVariant{
			Kind:  "artifact_type",
			Value: ArtifactTypeForView(viewSchemaID),
		},
		ViewSchemaIDs: []string{viewSchemaID},
	}
}
