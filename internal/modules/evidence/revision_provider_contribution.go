package evidence

import (
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/providers/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerEvidence,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:   revisions.SourceOwnerEvidence,
			RecordType:          "evidence",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.evidence.v1",
			HistoryTargetKinds:  []string{"evidence"},
			DeleteRestoreSource: deleterestore.NewSource(),
			RowRollbackProvider: rollback.NewProvider(),
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "evidence.evidence",
				ViewSchemaIDs:  []string{ViewSchemaID},
			}},
		}},
	}
}
