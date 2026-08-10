package evidence

import (
	"github.com/JochiRaider/cartulary/internal/modules/evidence/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerEvidence,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:      revisions.SourceOwnerEvidence,
			RecordType:             "evidence",
			SnapshotSchemaID:       "cartulary.revisions.snapshot.evidence.v1",
			HistoryTargetKinds:     []string{"evidence"},
			DeleteRestoreSource:    deleterestore.NewSource(),
			RowRollbackProvider:    rollbackprovider.NewProvider(),
			LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "evidence.evidence",
				ViewSchemaIDs:  []string{ViewSchemaID},
			}},
		}},
	}
}
