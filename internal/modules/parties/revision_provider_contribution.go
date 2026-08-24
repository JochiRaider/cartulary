package parties

import (
	"github.com/JochiRaider/cartulary/internal/modules/parties/internal/providers/deleterestore"
	partyrollback "github.com/JochiRaider/cartulary/internal/modules/parties/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func NewRevisionContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerParties,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:   revisions.SourceOwnerParties,
			RecordType:          "party",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.party.v1",
			DeleteRestoreSource: deleterestore.NewSource(),
			RowRollbackProvider: partyrollback.NewProvider(),
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "parties.parties",
				ViewSchemaIDs:  []string{ViewSchemaID},
			}},
		}},
	}
}
