package revisions

import (
	"slices"
	"testing"
)

func TestRollbackChangedFieldKeysIncludeEveryPublicCellDelta(t *testing.T) {
	before := map[string]any{
		"cells": map[string]any{
			"timeline.activity_synopsis_text": map[string]any{"value": "same"},
			"timeline.raw_activity_text":      map[string]any{"value": "before"},
			"timeline.capture_state":          map[string]any{"value": "rough"},
		},
	}
	after := map[string]any{
		"cells": map[string]any{
			"timeline.activity_synopsis_text":   map[string]any{"value": "same"},
			"timeline.raw_activity_text":        map[string]any{"value": "after"},
			"timeline.capture_state":            map[string]any{"value": "reviewed"},
			"timeline.has_unresolved_mentions":  map[string]any{"value": true},
			"timeline.activity_time_pair_state": map[string]any{"value": "paired_user_preserved"},
			"timeline.date_entered_sort_day":    map[string]any{"value": "2026-01-02"},
			"timeline.attached_evidence_ids":    map[string]any{"value": map[string]any{"items": []any{"evidence-1"}}},
		},
	}

	want := []string{
		"timeline.activity_time_pair_state",
		"timeline.attached_evidence_ids",
		"timeline.capture_state",
		"timeline.date_entered_sort_day",
		"timeline.has_unresolved_mentions",
		"timeline.raw_activity_text",
	}
	got := rollbackChangedFieldKeys(before, after)
	if !slices.Equal(got, want) {
		t.Fatalf("timeline rollback changed keys got %v want %v", got, want)
	}
}
