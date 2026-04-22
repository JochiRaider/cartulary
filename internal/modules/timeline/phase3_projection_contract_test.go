package timeline

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPhase3_ProjectionContract_U_3_08(t *testing.T) {
	recordID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	incidentID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	replacementID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	actorID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	occurredAt := time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)
	recordedAt := occurredAt.Add(2 * time.Minute)
	summary := "Summary"
	details := "Details"
	sourceText := "Source"

	projected := projectRecord(sourceRecord{
		RecordID:        recordID,
		IncidentID:      incidentID,
		OccurredAt:      &occurredAt,
		Summary:         &summary,
		Details:         &details,
		SourceText:      &sourceText,
		CaptureState:    captureStateReviewed,
		RowVersion:      4,
		RecordedAt:      recordedAt,
		EditedAt:        recordedAt,
		CreatedByUserID: actorID,
		UpdatedByUserID: actorID,
	}, &replacementID)
	projected.Tags = []map[string]any{
		{
			"item_ref":     "record_tag:55555555-5555-5555-5555-555555555555",
			"item_kind":    "tag",
			"display_text": "critical-host",
			"raw_text":     "critical-host",
		},
	}
	row := BuildRow(projected)

	if row["record_id"] != recordID.String() || row["row_version"] != int64(4) {
		t.Fatalf("expected stable record identity binding, got %#v", row)
	}

	cells := row["cells"].(map[string]any)
	wantCellKeys := []string{
		"timeline.capture_state",
		"timeline.details",
		"timeline.edited_at",
		"timeline.evidence_count",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
		"timeline.host_refs",
		"timeline.identity_refs",
		"timeline.occurred_at",
		"timeline.occurred_day",
		"timeline.recorded_at",
		"timeline.recorded_day",
		"timeline.replacement_record_id",
		"timeline.sort_ts",
		"timeline.source_text",
		"timeline.summary",
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
	if cells["timeline.summary"].(map[string]any)["value"] != summary || cells["timeline.details"].(map[string]any)["value"] != details || cells["timeline.source_text"].(map[string]any)["value"] != sourceText {
		t.Fatalf("expected scalar field values in projection row, got %#v", cells)
	}
	if cells["timeline.capture_state"].(map[string]any)["value"] != captureStateReviewed {
		t.Fatalf("expected capture_state cell value, got %#v", cells["timeline.capture_state"])
	}
	tagValue := cells["timeline.tags"].(map[string]any)["value"].(map[string]any)
	tagItems := tagValue["items"].([]map[string]any)
	if len(tagItems) != 1 || tagItems[0]["display_text"] != "critical-host" {
		t.Fatalf("expected stable tag collection value, got %#v", cells["timeline.tags"])
	}
	if cells["timeline.occurred_day"].(map[string]any)["value"] != "2026-04-10" || cells["timeline.recorded_day"].(map[string]any)["value"] != "2026-04-10" {
		t.Fatalf("expected derived day cells, got %#v %#v", cells["timeline.occurred_day"], cells["timeline.recorded_day"])
	}
	if cells["timeline.replacement_record_id"].(map[string]any)["value"] != replacementID.String() {
		t.Fatalf("expected replacement_record_id cell, got %#v", cells["timeline.replacement_record_id"])
	}

	groupValues := row["group_values"].(map[string]any)
	wantGroupKeys := []string{
		"timeline.capture_state",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
		"timeline.occurred_day",
		"timeline.recorded_day",
	}
	gotGroupKeys := make([]string, 0, len(groupValues))
	for fieldKey := range groupValues {
		gotGroupKeys = append(gotGroupKeys, fieldKey)
	}
	slices.Sort(gotGroupKeys)
	if !slices.Equal(gotGroupKeys, wantGroupKeys) {
		t.Fatalf("unexpected group_values keys: got %v want %v", gotGroupKeys, wantGroupKeys)
	}
	if groupValues["timeline.capture_state"] != captureStateReviewed || groupValues["timeline.occurred_day"] != "2026-04-10" || groupValues["timeline.recorded_day"] != "2026-04-10" {
		t.Fatalf("unexpected derived group_values: %#v", groupValues)
	}
}
