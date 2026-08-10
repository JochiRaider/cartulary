package indicators

import (
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/deleterestore"
	indicatorrollback "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/rollback"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func NewRevisionContribution() revisions.ProviderContribution {
	childProvider := indicatorrollback.NewChildProvider()
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerIndicators,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:   revisions.SourceOwnerIndicators,
			RecordType:          "indicator",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.indicator.v1",
			HistoryTargetKinds:  []string{"indicator"},
			DeleteRestoreSource: deleterestore.NewSource(),
			RowRollbackProvider: indicatorrollback.NewProvider(),
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "indicators.indicators",
				ViewSchemaIDs:  []string{ViewSchemaID},
			}},
		}},
		NonRowTargets: []revisions.NonRowProviderContribution{
			{
				SourceOwnerModule: revisions.SourceOwnerIndicators,
				TargetKind:        "indicator_observation",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"source_record_id", "resolved_indicator_record_id"}, revisions.HistorySingleEntry),
				RollbackProvider:  childProvider,
			},
			{
				SourceOwnerModule: revisions.SourceOwnerIndicators,
				TargetKind:        "indicator_state_interval",
				HistoryFacet:      revisions.NewFieldAssociationHistoryFacet([]string{"indicator_record_id"}, revisions.HistorySingleEntry),
				RollbackProvider:  childProvider,
			},
		},
	}
}
