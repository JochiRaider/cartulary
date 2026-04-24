package viewschema

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
)

func TestBaseRegistryPublicResources(t *testing.T) {
	resources := ListPublicResources()
	gotIDs := make([]string, 0, len(resources))
	for _, resource := range resources {
		gotIDs = append(gotIDs, resource.ViewSchemaID)
	}
	wantIDs := []string{
		"cartulary.view.assessments.v1",
		"cartulary.view.comm_log.v1",
		"cartulary.view.decisions.v1",
		"cartulary.view.evidence.v1",
		"cartulary.view.handoff.v1",
		"cartulary.view.hosts.v1",
		"cartulary.view.identities.v1",
		"cartulary.view.indicators.v1",
		"cartulary.view.lesson.v1",
		"cartulary.view.notes.v1",
		"cartulary.view.parties.v1",
		"cartulary.view.status_review.v1",
		"cartulary.view.task_requests.v1",
		"cartulary.view.timeline.v1",
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected Base view schema ids:\ngot  %v\nwant %v", gotIDs, wantIDs)
	}
	if !slices.IsSorted(gotIDs) {
		t.Fatalf("view schema ids must be sorted ascending: %v", gotIDs)
	}

	for _, resource := range resources {
		requirePublicResourceShape(t, resource)
		requireFieldOrderPreserved(t, resource)
		requireNoInternalMembers(t, resource)
	}
}

func requirePublicResourceShape(t testing.TB, resource ViewSchemaResource) {
	t.Helper()

	if resource.ViewSchemaID == "" || resource.SurfaceKind == "" || resource.Title == "" {
		t.Fatalf("resource has missing identity members: %#v", resource)
	}
	if !reflect.DeepEqual(resource.TechnicalFields, []string{"record_id", "row_version"}) {
		t.Fatalf("%s has unexpected technical fields: %v", resource.ViewSchemaID, resource.TechnicalFields)
	}
	if len(resource.DefaultSort) == 0 || resource.DefaultSort[len(resource.DefaultSort)-1] != (SortEntry{FieldKey: "record_id", Direction: "asc"}) {
		t.Fatalf("%s default_sort must end with record_id asc: %#v", resource.ViewSchemaID, resource.DefaultSort)
	}
	if len(resource.Fields) == 0 {
		t.Fatalf("%s must expose fields", resource.ViewSchemaID)
	}

	fieldKeys := make(map[string]struct{}, len(resource.Fields))
	for _, field := range resource.Fields {
		if field.FieldKey == "" || field.Label == "" || field.ReadKind == "" || field.WriteKind == "" || field.FilterOps == nil {
			t.Fatalf("%s exposes incomplete field entry: %#v", resource.ViewSchemaID, field)
		}
		if _, exists := fieldKeys[field.FieldKey]; exists {
			t.Fatalf("%s exposes duplicate field key %s", resource.ViewSchemaID, field.FieldKey)
		}
		fieldKeys[field.FieldKey] = struct{}{}
		if field.Sortable && field.HeaderSortFieldKey != nil && !slices.Contains(resource.SortFields, *field.HeaderSortFieldKey) {
			t.Fatalf("%s field %s has header sort key outside sort_fields: %s", resource.ViewSchemaID, field.FieldKey, *field.HeaderSortFieldKey)
		}
	}
	for _, fieldKey := range resource.FilterFields {
		if _, ok := fieldKeys[fieldKey]; !ok {
			t.Fatalf("%s filter field %s is not in fields[]", resource.ViewSchemaID, fieldKey)
		}
	}
	for _, predicate := range resource.SyntheticFilterPredicates {
		if _, ok := fieldKeys[predicate.FieldKey]; ok {
			t.Fatalf("%s synthetic predicate %s must not also be a field", resource.ViewSchemaID, predicate.FieldKey)
		}
	}
}

func requireFieldOrderPreserved(t testing.TB, resource ViewSchemaResource) {
	t.Helper()

	var artifactJSON string
	for _, artifact := range gencontracts.ViewSchemaArtifacts {
		if strings.HasSuffix(artifact.Path, "/index.json") {
			continue
		}
		var document schemaDocument
		if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
			t.Fatalf("unmarshal %s: %v", artifact.Path, err)
		}
		if document.ViewSchemaID == resource.ViewSchemaID {
			artifactJSON = artifact.JSON
			break
		}
	}
	if artifactJSON == "" {
		t.Fatalf("missing generated artifact for %s", resource.ViewSchemaID)
	}

	var document schemaDocument
	if err := json.Unmarshal([]byte(artifactJSON), &document); err != nil {
		t.Fatalf("unmarshal %s: %v", resource.ViewSchemaID, err)
	}
	want := make([]string, 0, len(document.Fields))
	for _, field := range document.Fields {
		want = append(want, field.FieldKey)
	}
	got := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		got = append(got, field.FieldKey)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s field order changed:\ngot  %v\nwant %v", resource.ViewSchemaID, got, want)
	}
}

func requireNoInternalMembers(t testing.TB, resource ViewSchemaResource) {
	t.Helper()

	payload, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("marshal resource: %v", err)
	}
	for _, forbidden := range []string{"write_target", "write_action", "base_projection", "read_model", "create_writable", "writable"} {
		if strings.Contains(string(payload), `"`+forbidden+`"`) {
			t.Fatalf("%s public resource leaked %s: %s", resource.ViewSchemaID, forbidden, payload)
		}
	}
}
