package timeline

import conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"

func newTimelineConflictSnapshotProjector() conflicttokens.RevisionSnapshotProjector {
	projector, err := conflicttokens.NewRevisionSnapshotProjector(
		"cartulary.revisions.snapshot.timeline_event.v1",
		map[string]string{
			"timeline.date_entered_text":        "date_entered_text",
			"timeline.analyst_text":             "analyst_text",
			"timeline.mitre_stage_text":         "mitre_stage_text",
			"timeline.device_object_text":       "device_object_text",
			"timeline.ip_address_text":          "ip_address_text",
			"timeline.activity_utc_text":        "activity_utc_text",
			"timeline.activity_local_text":      "activity_local_text",
			"timeline.raw_activity_text":        "raw_activity_text",
			"timeline.activity_synopsis_text":   "activity_synopsis_text",
			"timeline.data_source_text":         "data_source_text",
			"timeline.activity_time_pair_state": "activity_time_pair_state",
			"timeline.capture_state":            "capture_state",
		},
	)
	if err != nil {
		panic(err)
	}
	return projector
}
