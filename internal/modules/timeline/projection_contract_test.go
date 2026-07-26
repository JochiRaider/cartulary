package timeline

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func TestProjectionContract_Unit(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	replacementID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	actorID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	occurredAt := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	recordedAt := occurredAt.Add(2 * time.Minute)
	dateEntered := "2026-04-10"
	analyst := "Analyst A"
	mitre := "TA0001"
	device := "HOST-1"
	ipAddress := "192.0.2.10"
	activityUTC := "2026-04-10T12:00:00Z"
	activityLocal := "2026-04-10T08:00:00-04:00"
	rawActivity := "Raw source text"
	synopsis := "Summary"
	dataSource := "Source"

	projected := projectRecord(sourcerepository.Snapshot{
		RecordID:              recordID,
		IncidentID:            incidentID,
		DateEnteredText:       &dateEntered,
		AnalystText:           &analyst,
		MitreStageText:        &mitre,
		DeviceObjectText:      &device,
		IPAddressText:         &ipAddress,
		ActivityUTCText:       &activityUTC,
		ActivityLocalText:     &activityLocal,
		RawActivityText:       &rawActivity,
		ActivitySynopsisText:  &synopsis,
		DataSourceText:        &dataSource,
		ActivityTimePairState: "paired_user_preserved",
		CaptureState:          captureStateReviewed,
		RowVersion:            4,
		RecordedAt:            recordedAt,
		EditedAt:              recordedAt,
		CreatedByUserID:       actorID,
		UpdatedByUserID:       actorID,
	}, &replacementID)
	projected.Tags = []workbookprojection.TagRef{{
		ItemRef:     "record_tag:" + recordID.String() + ":55555555-5555-5555-5555-555555555555",
		ItemKind:    "tag",
		DisplayText: "critical-host",
		TagID:       uuid.MustParse("55555555-5555-5555-5555-555555555555"),
	}}
	row := buildRow(projected)

	if row["record_id"] != recordID.String() || row["row_version"] != int64(4) {
		t.Fatalf("expected stable record identity binding, got %#v", row)
	}

	cells := row["cells"].(map[string]any)
	wantCellKeys := []string{
		"timeline.activity_local_text",
		"timeline.activity_sort_ts",
		"timeline.activity_synopsis_text",
		"timeline.activity_time_pair_state",
		"timeline.activity_utc_text",
		"timeline.analyst_text",
		"timeline.attached_evidence_ids",
		"timeline.capture_state",
		"timeline.data_source_text",
		"timeline.date_entered_sort_day",
		"timeline.date_entered_text",
		"timeline.device_object_text",
		"timeline.edited_at",
		"timeline.evidence_count",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
		"timeline.host_refs",
		"timeline.identity_refs",
		"timeline.ip_address_text",
		"timeline.mitre_stage_text",
		"timeline.raw_activity_text",
		"timeline.recorded_at",
		"timeline.replacement_record_id",
		"timeline.tags",
	}
	gotCellKeys := make([]string, 0, len(cells))
	for fieldKey := range cells {
		gotCellKeys = append(gotCellKeys, fieldKey)
	}
	slices.Sort(gotCellKeys)
	if !slices.Equal(gotCellKeys, wantCellKeys) {
		t.Fatalf("unexpected projection cell keys: got %v want %v", gotCellKeys, wantCellKeys)
	}
	if cells["timeline.activity_synopsis_text"].(map[string]any)["value"] != synopsis ||
		cells["timeline.raw_activity_text"].(map[string]any)["value"] != rawActivity ||
		cells["timeline.data_source_text"].(map[string]any)["value"] != dataSource {
		t.Fatalf("expected scalar field values in projection contract row shape, got %#v", cells)
	}
	if cells["timeline.capture_state"].(map[string]any)["value"] != captureStateReviewed {
		t.Fatalf("expected capture_state cell value, got %#v", cells["timeline.capture_state"])
	}
	tagValue := cells["timeline.tags"].(map[string]any)["value"].(map[string]any)
	tagItems := tagValue["items"].([]map[string]any)
	if len(tagItems) != 1 || tagItems[0]["display_text"] != "critical-host" {
		t.Fatalf("expected stable tag collection value, got %#v", cells["timeline.tags"])
	}
	if cells["timeline.date_entered_sort_day"].(map[string]any)["value"] != "2026-04-10" ||
		cells["timeline.activity_sort_ts"].(map[string]any)["value"] != "2026-04-10T12:00:00Z" {
		t.Fatalf("expected derived sort cells, got %#v %#v", cells["timeline.date_entered_sort_day"], cells["timeline.activity_sort_ts"])
	}
	if cells["timeline.replacement_record_id"].(map[string]any)["value"] != replacementID.String() {
		t.Fatalf("expected replacement_record_id cell, got %#v", cells["timeline.replacement_record_id"])
	}

	groupValues := row["group_values"].(map[string]any)
	wantGroupKeys := []string{
		"timeline.activity_time_pair_state",
		"timeline.capture_state",
		"timeline.date_entered_sort_day",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
	}
	gotGroupKeys := make([]string, 0, len(groupValues))
	for fieldKey := range groupValues {
		gotGroupKeys = append(gotGroupKeys, fieldKey)
	}
	slices.Sort(gotGroupKeys)
	if !slices.Equal(gotGroupKeys, wantGroupKeys) {
		t.Fatalf("unexpected group_values keys: got %v want %v", gotGroupKeys, wantGroupKeys)
	}
	if groupValues["timeline.capture_state"] != captureStateReviewed ||
		groupValues["timeline.date_entered_sort_day"] != "2026-04-10" ||
		groupValues["timeline.activity_time_pair_state"] != "paired_user_preserved" {
		t.Fatalf("unexpected derived group_values: %#v", groupValues)
	}
}
