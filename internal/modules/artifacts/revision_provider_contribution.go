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
			SourceOwnerModule:     revisions.SourceOwnerArtifacts,
			RecordType:            "artifact",
			DeleteRestoreProvider: deleterestore.NewProvider(),
			RowRollbackProvider:   rollbackprovider.NewProvider(),
		}},
	}
}
