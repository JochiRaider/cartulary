package viewschemas

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestPhase8_ViewSchemaDiscovery_U_8_09(t *testing.T) {
	resources := viewschema.ListPublicResources()
	if len(resources) == 0 {
		t.Fatal("expected public view-schema resources")
	}
	for _, resource := range resources {
		if slices.Contains(resource.SortFields, "record_id") {
			t.Fatalf("%s exposed record_id as client-sortable: %#v", resource.ViewSchemaID, resource.SortFields)
		}
		if len(resource.DefaultSort) == 0 {
			t.Fatalf("%s must expose deterministic default_sort", resource.ViewSchemaID)
		}
		tail := resource.DefaultSort[len(resource.DefaultSort)-1]
		if tail.FieldKey != "record_id" || tail.Direction != "asc" {
			t.Fatalf("%s default_sort must end with record_id asc, got %#v", resource.ViewSchemaID, resource.DefaultSort)
		}
		for _, groupingKey := range resource.GroupingFields {
			if _, ok := fieldByKey(resource, groupingKey); !ok {
				t.Fatalf("%s grouping key %s is not a declared field", resource.ViewSchemaID, groupingKey)
			}
		}
	}

	timeline, ok := viewschema.LookupPublicResource("cartulary.view.timeline.v1")
	if !ok {
		t.Fatal("timeline view schema missing")
	}
	occurredAt, ok := fieldByKey(timeline, "timeline.occurred_at")
	if !ok {
		t.Fatal("timeline.occurred_at missing")
	}
	if occurredAt.HeaderSortFieldKey == nil || *occurredAt.HeaderSortFieldKey != "timeline.sort_ts" {
		t.Fatalf("timeline.occurred_at must sort through timeline.sort_ts, got %#v", occurredAt.HeaderSortFieldKey)
	}
	tags, ok := fieldByKey(timeline, "timeline.tags")
	if !ok {
		t.Fatal("timeline.tags missing")
	}
	if tags.Sortable || tags.HeaderSortFieldKey != nil {
		t.Fatalf("collection field timeline.tags must not synthesize client sort, got %#v", tags)
	}
}

func fieldByKey(resource viewschema.ViewSchemaResource, fieldKey string) (viewschema.ViewFieldEntry, bool) {
	for _, field := range resource.Fields {
		if field.FieldKey == fieldKey {
			return field, true
		}
	}
	return viewschema.ViewFieldEntry{}, false
}
