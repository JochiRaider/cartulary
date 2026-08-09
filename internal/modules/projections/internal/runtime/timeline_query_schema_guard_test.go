package runtime

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestUnit_TimelineQuerySchemaMappingGuard(t *testing.T) {
	schema, ok := viewschema.Lookup(timelineViewSchemaID)
	if !ok {
		t.Fatalf("timeline schema %s not registered", timelineViewSchemaID)
	}
	surface, ok := querySurfacesForTest()[timelineViewSchemaID]
	if !ok {
		t.Fatal("timeline query surface is not registered")
	}

	wantSortFields := uniqueTimelineStrings(schema.SortFields())
	for _, entry := range schema.DefaultSort() {
		wantSortFields = appendUniqueTimelineString(wantSortFields, entry.FieldKey)
	}
	schemaSortFields := uniqueTimelineStrings(schema.SortFields())
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
	for _, fieldKey := range wantSortFields {
		if _, ok := surface.field(fieldKey); !ok {
			t.Fatalf("timeline sort field %s is not mapped", fieldKey)
		}
		_, _, err := buildGenericQueryPageSQL(uuid.Nil, surface, viewschema.QueryMeta{
			Sort: []viewschema.SortEntry{{FieldKey: fieldKey, Direction: "asc"}},
		}, querypage.Window{Limit: 100})
		if err != nil {
			t.Fatalf("timeline sort field %s from schema/default_sort is not mapped: %v", fieldKey, err)
		}
	}

	for _, fieldKey := range schema.FilterFields() {
		_, _, err := buildGenericQueryPageSQL(uuid.Nil, surface, viewschema.QueryMeta{
			Filters: []viewschema.Filter{sampleTimelineFilter(fieldKey)},
			Sort:    schema.DefaultSort(),
		}, querypage.Window{Limit: 100})
		if err != nil {
			t.Fatalf("timeline filter field %s from schema is not mapped: %v", fieldKey, err)
		}
	}
	if !sameTimelineStrings(surface.groupingFields, schema.GroupingFields()) {
		t.Fatalf("timeline grouping mapping drifted from schema: got %v want %v", surface.groupingFields, schema.GroupingFields())
	}
}

func sampleTimelineFilter(fieldKey string) viewschema.Filter {
	switch fieldKey {
	case "timeline.date_entered_sort_day":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": "2026-01-02"}}
	case "timeline.activity_time_pair_state":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": "paired_user_preserved"}}
	case "timeline.capture_state":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": "rough"}}
	case "timeline.has_evidence", "timeline.has_unresolved_mentions":
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": true}}
	case "timeline.tags":
		return viewschema.Filter{FieldKey: fieldKey, Op: "contains_any", Arg: map[string]any{"values": []any{"priority"}}}
	default:
		return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": "unsupported"}}
	}
}

func uniqueTimelineStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = appendUniqueTimelineString(out, value)
	}
	slices.Sort(out)
	return out
}

func appendUniqueTimelineString(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}
	values = append(values, value)
	slices.Sort(values)
	return values
}

func sameTimelineStrings(left []string, right []string) bool {
	left = uniqueTimelineStrings(left)
	right = uniqueTimelineStrings(right)
	return slices.Equal(left, right)
}
