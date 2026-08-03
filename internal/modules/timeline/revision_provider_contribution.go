package timeline

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/deleterestore"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/rollbackprovider"
)

func RevisionProviderContribution() revisions.ProviderContribution {
	return revisions.ProviderContribution{
		SourceOwnerModule:     revisions.SourceOwnerTimeline,
		ConflictFieldProvider: revisions.NewViewSchemaConflictFieldProvider(),
		Records: []revisions.RecordProviderContribution{{
			SourceOwnerModule:      revisions.SourceOwnerTimeline,
			RecordType:             "timeline_event",
			HistoryTargetKinds:     []string{"timeline_record"},
			DeleteRestoreSource:    deleterestore.NewSource(),
			RowRollbackProvider:    rollbackprovider.NewTimelineProvider(),
			LiveRecordChangePolicy: revisions.LiveRecordChangeRequired,
			RecordViewRoutes: []revisions.RecordViewRouteContribution{{
				ContributionID: "timeline.timeline",
				ViewSchemaIDs:  []string{TimelineViewSchemaID},
			}},
		}},
	}
}
