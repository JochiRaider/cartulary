package viewquery

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const timelineViewSchemaID = "cartulary.view.timeline.v1"

func TestDecode_DefaultsToSchemaMeta(t *testing.T) {
	query, err := Decode(strings.NewReader(`{}`), timelineViewSchemaID)
	if err != nil {
		t.Fatalf("decode default query: %+v", err)
	}

	schema, ok := viewschema.Lookup(timelineViewSchemaID)
	if !ok {
		t.Fatalf("timeline schema not registered")
	}
	if len(query.Meta.Filters) != 0 {
		t.Fatalf("expected no filters, got %#v", query.Meta.Filters)
	}
	if query.Meta.GroupBy != nil {
		t.Fatalf("expected omitted group_by, got %#v", query.Meta.GroupBy)
	}
	if len(query.Meta.Sort) != len(schema.DefaultSort()) {
		t.Fatalf("expected default sort length %d, got %#v", len(schema.DefaultSort()), query.Meta.Sort)
	}
}

func TestDecode_NormalizesFiltersSortAndGrouping(t *testing.T) {
	query, err := Decode(strings.NewReader(`{
  "sort": [{"field_key": "timeline.summary", "direction": "asc"}],
  "filters": [
    {"field_key": "timeline.tags", "op": "contains_any", "arg": {"values": ["beta", "alpha", "alpha"]}},
    {"field_key": "timeline.capture_state", "op": "eq", "arg": {"value": "rough"}}
  ],
  "group_by": "timeline.capture_state"
}`), timelineViewSchemaID)
	if err != nil {
		t.Fatalf("decode normalized query: %+v", err)
	}

	if got := query.Meta.Sort[0]; got.FieldKey != "timeline.summary" || got.Direction != "asc" {
		t.Fatalf("unexpected primary sort: %#v", query.Meta.Sort)
	}
	if got := query.Meta.Sort[1]; got.FieldKey != "timeline.sort_ts" || got.Direction != "asc" {
		t.Fatalf("expected default tail sort_ts asc, got %#v", query.Meta.Sort)
	}
	if got := query.Meta.Sort[2]; got.FieldKey != "record_id" || got.Direction != "asc" {
		t.Fatalf("expected record_id asc tie-break, got %#v", query.Meta.Sort)
	}
	if query.Meta.GroupBy == nil || *query.Meta.GroupBy != "timeline.capture_state" {
		t.Fatalf("unexpected group_by: %#v", query.Meta.GroupBy)
	}
	if len(query.Meta.Filters) != 2 {
		t.Fatalf("expected two normalized filters, got %#v", query.Meta.Filters)
	}
	if query.Meta.Filters[0].FieldKey != "timeline.capture_state" {
		t.Fatalf("expected filters sorted by field_key, got %#v", query.Meta.Filters)
	}
	values, ok := query.Meta.Filters[1].Arg["values"].([]any)
	if !ok {
		t.Fatalf("expected canonical values array, got %#v", query.Meta.Filters[1].Arg)
	}
	if len(values) != 2 || values[0] != "alpha" || values[1] != "beta" {
		t.Fatalf("expected canonical deduped tag values, got %#v", values)
	}
}

func TestDecode_RejectsDuplicateFilterField(t *testing.T) {
	_, err := Decode(strings.NewReader(`{
  "filters": [
    {"field_key": "timeline.capture_state", "op": "eq", "arg": {"value": "rough"}},
    {"field_key": "timeline.capture_state", "op": "eq", "arg": {"value": "enriched"}}
  ]
}`), timelineViewSchemaID)
	if err == nil {
		t.Fatal("expected duplicate filter field rejection")
	}
	if err.ReasonCode != "duplicate_filter_field" {
		t.Fatalf("unexpected reason_code: %+v", err)
	}
	if err.FilterIndex == nil || *err.FilterIndex != 1 {
		t.Fatalf("expected duplicate rejection at second filter, got %+v", err)
	}
}

func TestDecode_RejectsOversizeSortAndInvalidGroupBy(t *testing.T) {
	_, sortErr := Decode(strings.NewReader(`{
  "sort": [
    {"field_key": "timeline.summary", "direction": "asc"},
    {"field_key": "timeline.sort_ts", "direction": "asc"},
    {"field_key": "timeline.evidence_count", "direction": "asc"},
    {"field_key": "timeline.edited_at", "direction": "asc"},
    {"field_key": "timeline.capture_state", "direction": "asc"},
    {"field_key": "timeline.occurred_day", "direction": "asc"},
    {"field_key": "timeline.recorded_day", "direction": "asc"},
    {"field_key": "timeline.has_evidence", "direction": "asc"},
    {"field_key": "timeline.has_unresolved_mentions", "direction": "asc"}
  ]
}`), timelineViewSchemaID)
	if sortErr == nil || sortErr.ReasonCode != "sort_count_exceeded" {
		t.Fatalf("expected sort_count_exceeded, got %+v", sortErr)
	}

	_, groupErr := Decode(strings.NewReader(`{"group_by": null}`), timelineViewSchemaID)
	if groupErr == nil || groupErr.ReasonCode != "invalid_group_by" {
		t.Fatalf("expected invalid_group_by, got %+v", groupErr)
	}
}
