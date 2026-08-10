package rollback

import (
	"errors"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestSourceForRollbackValueExcludesDerivedChildEffects(t *testing.T) {
	t.Parallel()
	value := map[string]any{"source": map[string]any{
		"display_value":  "192.0.2.10",
		"hash_algorithm": nil,
	}}
	got, ok := sourceForRollbackValue(value)
	if !ok {
		t.Fatal("sourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"display_value": "192.0.2.10", "hash_algorithm": nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("indicator source = %#v, want %#v", got, want)
	}
	if _, ok := sourceForRollbackValue(map[string]any{"cells": map[string]any{"indicator.display_value": map[string]any{"value": "legacy"}}}); ok {
		t.Fatal("schema-less projection row was accepted")
	}
}

func TestParseChildValueRejectsNonRegistryObservationOrigin(t *testing.T) {
	t.Parallel()
	value := map[string]any{
		"incident_id":                  "00000000-0000-4000-8000-000000000001",
		"row_version":                  float64(1),
		"deleted_at":                   nil,
		"deleted_by_user_id":           nil,
		"indicator_observation_id":     "00000000-0000-4000-8000-000000000002",
		"source_record_id":             "00000000-0000-4000-8000-000000000003",
		"source_field_key":             "timeline.raw_activity_text",
		"resolution_status":            "unresolved",
		"origin_kind":                  "interactive_cell",
		"origin_locator":               "rollback-origin-test",
		"observed_text":                "192.0.2.10",
		"created_by_user_id":           "00000000-0000-4000-8000-000000000004",
		"created_at":                   "2026-08-03T16:00:00Z",
		"resolved_indicator_record_id": nil,
		"resolved_by_user_id":          nil,
		"resolved_at":                  nil,
		"resolution_method":            nil,
	}
	if _, err := parseChildValue("indicator_observation", value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("parse invalid historical origin = %v", err)
	}
}
