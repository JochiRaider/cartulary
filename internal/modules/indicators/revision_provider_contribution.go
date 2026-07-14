package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/rollbackprovider"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	childProvider := rollbackprovider.NewChildProvider()
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerIndicators,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:     revisions.SourceOwnerIndicators,
			RecordType:            "indicator",
			DeleteRestoreProvider: deleterestore.NewProvider(),
			RowRollbackProvider:   rollbackprovider.NewProvider(),
		}},
		NonRowTargets: []revisions.NonRowProviderContribution{
			{SourceOwnerModule: revisions.SourceOwnerIndicators, TargetKind: "indicator_observation", RollbackProvider: childProvider},
			{SourceOwnerModule: revisions.SourceOwnerIndicators, TargetKind: "indicator_state_interval", RollbackProvider: childProvider},
		},
	}
}
