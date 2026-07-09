package rollbackprovider

import (
	"reflect"
	"testing"
)

func TestTimelineProviderSourceForRollbackValueMapsTimelineCells(t *testing.T) {
	provider := NewTimelineProvider()
	value := map[string]any{
		"cells": map[string]any{
			"timeline.date_entered_text":        rollbackCell("2026-01-02"),
			"timeline.analyst_text":             rollbackCell("analyst"),
			"timeline.mitre_stage_text":         rollbackCell("TA0001"),
			"timeline.device_object_text":       rollbackCell("HOST-1"),
			"timeline.ip_address_text":          rollbackCell("192.0.2.10"),
			"timeline.activity_utc_text":        rollbackCell("2026-01-02T03:04:05Z"),
			"timeline.activity_local_text":      rollbackCell("2026-01-01T22:04:05-05:00"),
			"timeline.raw_activity_text":        rollbackCell("raw"),
			"timeline.activity_synopsis_text":   rollbackCell("summary"),
			"timeline.data_source_text":         rollbackCell("source"),
			"timeline.activity_time_pair_state": rollbackCell("paired_user_preserved"),
			"timeline.capture_state":            rollbackCell("reviewed"),
			"timeline.replacement_record_id":    rollbackCell("11111111-1111-4111-8111-111111111111"),
			"timeline.reviewed_at":              rollbackCell("2026-01-02T04:00:00Z"),
			"timeline.superseded_at":            rollbackCell("2026-01-02T05:00:00Z"),
		},
	}
	got, ok, err := provider.SourceForRollbackValue(value)
	if err != nil {
		t.Fatalf("SourceForRollbackValue returned error: %v", err)
	}
	if !ok {
		t.Fatal("SourceForRollbackValue returned ok=false")
	}
	want := map[string]any{
		"date_entered_text":        "2026-01-02",
		"analyst_text":             "analyst",
		"mitre_stage_text":         "TA0001",
		"device_object_text":       "HOST-1",
		"ip_address_text":          "192.0.2.10",
		"activity_utc_text":        "2026-01-02T03:04:05Z",
		"activity_local_text":      "2026-01-01T22:04:05-05:00",
		"raw_activity_text":        "raw",
		"activity_synopsis_text":   "summary",
		"data_source_text":         "source",
		"activity_time_pair_state": "paired_user_preserved",
		"capture_state":            "reviewed",
		"replacement_record_id":    "11111111-1111-4111-8111-111111111111",
		"reviewed_at":              "2026-01-02T04:00:00Z",
		"superseded_at":            "2026-01-02T05:00:00Z",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("timeline rollback source got %#v want %#v", got, want)
	}
}

func rollbackCell(value any) map[string]any {
	return map[string]any{"value": value}
}
