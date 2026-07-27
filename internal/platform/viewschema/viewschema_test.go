package viewschema

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestViewSchemaDiscovery_Unit(t *testing.T) {
	resources := ListPublicResources()
	if len(resources) == 0 {
		t.Fatal("expected public view-schema resources")
	}
	for _, resource := range resources {
		internal, ok := Lookup(resource.ViewSchemaID)
		if !ok {
			t.Fatalf("%s missing internal schema", resource.ViewSchemaID)
		}
		if !internal.CreateCapable {
			t.Fatalf("%s must be create-capable in the current profile", resource.ViewSchemaID)
		}
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
			if field.Sortable {
				sortFieldKey := field.FieldKey
				if field.HeaderSortFieldKey != nil {
					sortFieldKey = *field.HeaderSortFieldKey
				}
				if !slices.Contains(resource.SortFields, sortFieldKey) {
					t.Fatalf("%s sortable field %s is not backed by sort_fields key %s", resource.ViewSchemaID, field.FieldKey, sortFieldKey)
				}
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
		"timeline.evidence_count",
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

func TestOpenAPIViewSchemaPublicProjectionContract_Unit(t *testing.T) {
	document := contracttest.OpenAPIDocument(t)
	components := requireObject(t, document["components"], "components")
	schemas := requireObject(t, components["schemas"], "components.schemas")
	viewFieldEntry := requireObject(t, schemas["ViewFieldEntry"], "ViewFieldEntry")
	properties := requireObject(t, viewFieldEntry["properties"], "ViewFieldEntry.properties")
	gridEditable := requireObject(t, properties["grid_editable"], "ViewFieldEntry.properties.grid_editable")
	if gridEditable["type"] != "boolean" {
		t.Fatalf("grid_editable must have type boolean, got %#v", gridEditable["type"])
	}
	required := requireArray(t, viewFieldEntry["required"], "ViewFieldEntry.required")
	if !slices.Contains(required, "grid_editable") {
		t.Fatalf("ViewFieldEntry.required omits grid_editable: %#v", required)
	}

	for _, resource := range ListPublicResources() {
		content, err := json.Marshal(resource)
		if err != nil {
			t.Fatalf("marshal %s: %v", resource.ViewSchemaID, err)
		}
		var public map[string]any
		if err := json.Unmarshal(content, &public); err != nil {
			t.Fatalf("decode %s: %v", resource.ViewSchemaID, err)
		}
		if _, leaked := public["create_capable"]; leaked {
			t.Fatalf("%s leaked internal create_capable in public resource", resource.ViewSchemaID)
		}
		fields, ok := public["fields"].([]any)
		if !ok {
			t.Fatalf("%s fields must be an array, got %T", resource.ViewSchemaID, public["fields"])
		}
		for index, rawField := range fields {
			field := requireObject(t, rawField, resource.ViewSchemaID+" field")
			if _, ok := field["grid_editable"].(bool); !ok {
				t.Fatalf("%s fields[%d].grid_editable must be present and boolean, got %#v", resource.ViewSchemaID, index, field["grid_editable"])
			}
		}
	}
	verifyOpenAPIViewSchemaInlineCreateProjection(t, document)
}

func verifyOpenAPIViewSchemaInlineCreateProjection(t *testing.T, document map[string]any) {
	t.Helper()
	components := requireObject(t, document["components"], "components")
	schemas := requireObject(t, components["schemas"], "components.schemas")
	viewSchemaResource := requireObject(t, schemas["ViewSchemaResource"], "ViewSchemaResource")
	properties := requireObject(t, viewSchemaResource["properties"], "ViewSchemaResource.properties")
	inlineCreate := requireObject(t, properties["inline_create"], "ViewSchemaResource.properties.inline_create")
	if inlineCreate["type"] != "object" || inlineCreate["additionalProperties"] != false {
		t.Fatalf("inline_create must be a closed object, got %#v", inlineCreate)
	}
	required := requireArray(t, inlineCreate["required"], "inline_create.required")
	if !slices.Equal(required, []string{"minimum_create_field_sets", "permits_zero_field_create"}) {
		t.Fatalf("inline_create.required changed: %#v", required)
	}
	if !slices.Contains(requireArray(t, viewSchemaResource["required"], "ViewSchemaResource.required"), "inline_create") {
		t.Fatal("ViewSchemaResource.required omits inline_create")
	}
	inlineProperties := requireObject(t, inlineCreate["properties"], "inline_create.properties")
	if len(inlineProperties) != 2 {
		t.Fatalf("inline_create must have exactly two properties, got %#v", inlineProperties)
	}
	minimumSets := requireObject(t, inlineProperties["minimum_create_field_sets"], "inline_create.minimum_create_field_sets")
	setItems := requireObject(t, minimumSets["items"], "inline_create.minimum_create_field_sets.items")
	fieldItems := requireObject(t, setItems["items"], "inline_create.minimum_create_field_sets.items.items")
	if minimumSets["type"] != "array" || setItems["type"] != "array" || fieldItems["type"] != "string" {
		t.Fatalf("minimum_create_field_sets must be an array of string arrays, got %#v", minimumSets)
	}
	permitsZero := requireObject(t, inlineProperties["permits_zero_field_create"], "inline_create.permits_zero_field_create")
	if permitsZero["type"] != "boolean" {
		t.Fatalf("permits_zero_field_create must be boolean, got %#v", permitsZero)
	}

	for _, resource := range ListPublicResources() {
		content, err := json.Marshal(resource)
		if err != nil {
			t.Fatalf("marshal %s: %v", resource.ViewSchemaID, err)
		}
		var public map[string]any
		if err := json.Unmarshal(content, &public); err != nil {
			t.Fatalf("decode %s: %v", resource.ViewSchemaID, err)
		}
		policy := requireObject(t, public["inline_create"], resource.ViewSchemaID+" inline_create")
		if len(policy) != 2 {
			t.Fatalf("%s inline_create must have exactly two members, got %#v", resource.ViewSchemaID, policy)
		}
		if _, ok := policy["minimum_create_field_sets"].([]any); !ok {
			t.Fatalf("%s minimum_create_field_sets must be an array, got %#v", resource.ViewSchemaID, policy["minimum_create_field_sets"])
		}
		if _, ok := policy["permits_zero_field_create"].(bool); !ok {
			t.Fatalf("%s permits_zero_field_create must be boolean, got %#v", resource.ViewSchemaID, policy["permits_zero_field_create"])
		}
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

func requireObject(t *testing.T, value any, label string) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s must be an object, got %T", label, value)
	}
	return object
}

func requireArray(t *testing.T, value any, label string) []string {
	t.Helper()
	raw, ok := value.([]any)
	if !ok {
		t.Fatalf("%s must be an array, got %T", label, value)
	}
	result := make([]string, len(raw))
	for index, item := range raw {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("%s[%d] must be a string, got %T", label, index, item)
		}
		result[index] = text
	}
	return result
}
