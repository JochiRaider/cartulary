package timeline

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/rollbackprovider"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule: revisions.SourceOwnerTimeline,
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:   revisions.SourceOwnerTimeline,
			RecordType:          "timeline_event",
			SnapshotSchemaID:    "cartulary.revisions.snapshot.timeline_event.v1",
			HistoryTargetKinds:  []string{"timeline_record"},
			DeleteRestoreSource: deleterestore.NewSource(),
			RowRollbackProvider: rollbackprovider.NewTimelineProvider(),
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "timeline.timeline",
				ViewSchemaIDs:  []string{TimelineViewSchemaID},
			}},
		}},
	}
}
