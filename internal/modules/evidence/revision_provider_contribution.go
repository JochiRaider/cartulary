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
			SourceOwnerModule:     revisions.SourceOwnerEvidence,
			RecordType:            "evidence",
			DeleteRestoreProvider: deleterestore.NewProvider(),
			RowRollbackProvider:   rollbackprovider.NewProvider(),
		}},
	}
}
