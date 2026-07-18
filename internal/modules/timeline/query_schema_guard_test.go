package timeline

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestUnit_TimelineQuerySchemaMappingGuard(t *testing.T) {
	schema, ok := viewschema.Lookup(TimelineViewSchemaID)
	if !ok {
		t.Fatalf("timeline schema %s not registered", TimelineViewSchemaID)
	}

	wantSortFields := uniqueStrings(schema.SortFields())
	for _, entry := range schema.DefaultSort() {
		wantSortFields = appendUnique(wantSortFields, entry.FieldKey)
	}
	schemaSortFields := uniqueStrings(schema.SortFields())
	for _, field := range schema.Fields() {
		if !field.Sortable {
			continue
		}
		sortFieldKey := field.FieldKey
		if field.HeaderSortFieldKey != nil {
			sortFieldKey = *field.HeaderSortFieldKey
		}
		if !slices.Contains(schemaSortFields, sortFieldKey) {
			t.Fatalf("timeline sortable field %s is not backed by sort_fields key %s", field.FieldKey, sortFieldKey)
		}
	}
	gotSortFields := mapKeys(timelineSortExpressions)
	if !sameStrings(gotSortFields, wantSortFields) {
		t.Fatalf("timeline sort SQL mapping drifted from schema/default sort: got %v want %v", gotSortFields, wantSortFields)
	}

	for _, fieldKey := range wantSortFields {
		_, _, err := buildTimelineQueryPageSQL(uuid.Nil, viewschema.QueryMeta{
			Sort: []viewschema.SortEntry{{FieldKey: fieldKey, Direction: "asc"}},
		}, querypage.Window{Limit: 100})
		if err != nil {
			t.Fatalf("timeline sort field %s from schema/default_sort is not mapped: %v", fieldKey, err)
		}
	}

	for _, fieldKey := range schema.FilterFields() {
		_, _, err := buildTimelineQueryPageSQL(uuid.Nil, viewschema.QueryMeta{
			Filters: []viewschema.Filter{sampleTimelineFilter(fieldKey)},
			Sort:    schema.DefaultSort(),
		}, querypage.Window{Limit: 100})
		if err != nil {
			t.Fatalf("timeline filter field %s from schema is not mapped: %v", fieldKey, err)
		}
	}

	row := buildRow(projectedRecord{})
	groupValues, ok := row["group_values"].(map[string]any)
	if !ok {
		t.Fatalf("timeline row group_values missing or wrong type: %#v", row["group_values"])
	}
	gotGroupFields := mapKeys(groupValues)
	if !sameStrings(gotGroupFields, schema.GroupingFields()) {
		t.Fatalf("timeline group_values drifted from schema grouping_fields: got %v want %v", gotGroupFields, schema.GroupingFields())
	}
}

func sampleTimelineFilter(fieldKey string) viewschema.Filter {
	switch fieldKey {
	case "timeline.date_entered_sort_day":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": "2026-01-02"}}
	case "timeline.activity_time_pair_state":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": "paired_user_preserved"}}
	case "timeline.capture_state":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": captureStateRough}}
	case "timeline.has_evidence", "timeline.has_unresolved_mentions":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": true}}
	case "timeline.tags":
		return viewschema.Filter{FieldKey: fieldKey, Op: "contains_any", Arg: map[string]any{"values": []any{"priority"}}}
	default:
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": "unsupported"}}
	}
}

func uniqueStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = appendUnique(out, value)
	}
	slices.Sort(out)
	return out
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	values = append(values, value)
	slices.Sort(values)
	return values
}

func mapKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func sameStrings(left []string, right []string) bool {
	left = uniqueStrings(left)
	right = uniqueStrings(right)
	return slices.Equal(left, right)
}
