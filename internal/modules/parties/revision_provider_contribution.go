package parties

import (
	"github.com/JochiRaider/cartulary/internal/modules/parties/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/parties/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerParties,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:   revisions.SourceOwnerParties,
			RecordType:          "party",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.party.v1",
			DeleteRestoreSource: deleterestore.NewSource(),
			RowRollbackProvider: rollbackprovider.NewPartyProvider(),
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "parties.parties",
				ViewSchemaIDs:  []string{ViewSchemaID},
			}},
		}},
	}
}
