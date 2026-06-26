package viewschema

import (
	"reflect"
	"slices"
	"testing"
)

func TestPhase8_ViewSchemaDiscovery_U_8_09(t *testing.T) {
	resources := ListPublicResources()
	if len(resources) == 0 {
		t.Fatal("expected public view-schema resources")
	}
	for _, resource := range resources {
		if !reflect.DeepEqual(resource.TechnicalFields, []string{"record_id", "row_version"}) {
			t.Fatalf("%s has unexpected technical_fields: %#v", resource.ViewSchemaID, resource.TechnicalFields)
		}
		if slices.Contains(resource.SortFields, "record_id") {
			t.Fatalf("%s exposed record_id as client-sortable: %#v", resource.ViewSchemaID, resource.SortFields)
		}
		if slices.Contains(resource.SortFields, "row_version") {
			t.Fatalf("%s exposed row_version as client-sortable: %#v", resource.ViewSchemaID, resource.SortFields)
		}
		if resource.SortNullOrder != "last" {
			t.Fatalf("%s exposed unexpected sort_null_order: %q", resource.ViewSchemaID, resource.SortNullOrder)
		}
		if len(resource.DefaultSort) == 0 {
			t.Fatalf("%s must expose deterministic default_sort", resource.ViewSchemaID)
		}
		tail := resource.DefaultSort[len(resource.DefaultSort)-1]
		if tail.FieldKey != "record_id" || tail.Direction != "asc" {
			t.Fatalf("%s default_sort must end with record_id asc, got %#v", resource.ViewSchemaID, resource.DefaultSort)
		}
		for _, field := range resource.Fields {
			if field.HeaderSortFieldKey != nil && !slices.Contains(resource.SortFields, *field.HeaderSortFieldKey) {
				t.Fatalf("%s field %s has header_sort_field_key outside sort_fields: %s", resource.ViewSchemaID, field.FieldKey, *field.HeaderSortFieldKey)
			}
		}
		for _, groupingKey := range resource.GroupingFields {
			field, ok := viewFieldByKey(resource, groupingKey)
			if !ok {
				t.Fatalf("%s grouping key %s is not a declared field", resource.ViewSchemaID, groupingKey)
			}
			if !field.Groupable {
				t.Fatalf("%s grouping key %s must have groupable=true", resource.ViewSchemaID, groupingKey)
			}
		}
	}

	timeline, ok := LookupPublicResource("cartulary.view.timeline.v2")
	if !ok {
		t.Fatal("timeline view schema missing")
	}
	if !reflect.DeepEqual(timeline.SortFields, []string{
		"timeline.activity_sort_ts",
		"timeline.date_entered_sort_day",
		"timeline.activity_synopsis_text",
		"timeline.analyst_text",
		"timeline.mitre_stage_text",
		"timeline.device_object_text",
		"timeline.ip_address_text",
		"timeline.data_source_text",
		"timeline.edited_at",
		"timeline.capture_state",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
	}) {
		t.Fatalf("timeline sort_fields changed:\ngot  %#v", timeline.SortFields)
	}
	if !reflect.DeepEqual(timeline.GroupingFields, []string{
		"timeline.date_entered_sort_day",
		"timeline.activity_time_pair_state",
		"timeline.capture_state",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
	}) {
		t.Fatalf("timeline grouping_fields changed:\ngot  %#v", timeline.GroupingFields)
	}
	occurredAt, ok := viewFieldByKey(timeline, "timeline.activity_utc_text")
	if !ok {
		t.Fatal("timeline.activity_utc_text missing")
	}
	if occurredAt.HeaderSortFieldKey == nil || *occurredAt.HeaderSortFieldKey != "timeline.activity_sort_ts" {
		t.Fatalf("timeline.activity_utc_text must sort through timeline.activity_sort_ts, got %#v", occurredAt.HeaderSortFieldKey)
	}
	tags, ok := viewFieldByKey(timeline, "timeline.tags")
	if !ok {
		t.Fatal("timeline.tags missing")
	}
	if tags.Sortable || tags.HeaderSortFieldKey != nil || slices.Contains(timeline.SortFields, "timeline.tags") {
		t.Fatalf("collection field timeline.tags must not synthesize client sort, got %#v", tags)
	}
}

func viewFieldByKey(resource ViewSchemaResource, fieldKey string) (ViewFieldEntry, bool) {
	for _, field := range resource.Fields {
		if field.FieldKey == fieldKey {
			return field, true
		}
	}
	return ViewFieldEntry{}, false
}
