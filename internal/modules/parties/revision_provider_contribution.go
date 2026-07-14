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
			SourceOwnerModule:     revisions.SourceOwnerParties,
			RecordType:            "party",
			DeleteRestoreProvider: deleterestore.NewProvider(),
			RowRollbackProvider:   rollbackprovider.NewPartyProvider(),
		}},
	}
}
