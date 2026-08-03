package parties

import (
	"github.com/JochiRaider/cartulary/internal/modules/parties/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/parties/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule:     revisions.SourceOwnerParties,
		ConflictFieldProvider: revisions.NewViewSchemaConflictFieldProvider(),
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:      revisions.SourceOwnerParties,
			RecordType:             "party",
			DeleteRestoreSource:    deleterestore.NewSource(),
			RowRollbackProvider:    rollbackprovider.NewPartyProvider(),
			LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "parties.parties",
				ViewSchemaIDs:  []string{ViewSchemaID},
			}},
		}},
	}
}
