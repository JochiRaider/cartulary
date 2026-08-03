package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/deleterestore"
	indicatorrollback "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func NewRevisionContribution() revisions.ProviderContribution {
	childProvider := indicatorrollback.NewChildProvider()
	return revisions.ProviderContribution{
		SourceOwnerModule:     revisions.SourceOwnerIndicators,
		ConflictFieldProvider: revisions.NewViewSchemaConflictFieldProvider(),
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:      revisions.SourceOwnerIndicators,
			RecordType:             "indicator",
			DeleteRestoreSource:    deleterestore.NewSource(),
			RowRollbackProvider:    indicatorrollback.NewProvider(),
			LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "indicators.indicators",
				ViewSchemaIDs:  []string{ViewSchemaID},
			}},
		}},
		NonRowTargets: []revisions.NonRowProviderContribution{
			{SourceOwnerModule: revisions.SourceOwnerIndicators, TargetKind: "indicator_observation", RollbackProvider: childProvider},
			{SourceOwnerModule: revisions.SourceOwnerIndicators, TargetKind: "indicator_state_interval", RollbackProvider: childProvider},
		},
	}
}
