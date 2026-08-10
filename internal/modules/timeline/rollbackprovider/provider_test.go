package rollbackprovider

import (
	"reflect"
	"testing"
)

func TestTimelineProviderSourceForRollbackValueMapsTimelineCells(t *testing.T) {
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
	got, ok, err := sourceForRollbackValue(map[string]any{"source": want})
	if err != nil {
		t.Fatalf("SourceForRollbackValue returned error: %v", err)
	}
	if !ok {
		t.Fatal("SourceForRollbackValue returned ok=false")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("timeline rollback source got %#v want %#v", got, want)
	}
	if _, ok, err := sourceForRollbackValue(map[string]any{"cells": map[string]any{"timeline.raw_activity_text": map[string]any{"value": "legacy"}}}); err != nil || ok {
		t.Fatalf("schema-less projection row admission = %v, %v", ok, err)
	}
}
