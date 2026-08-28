package sourcestate

import (
	"reflect"
	"testing"
)

func TestManifestProjectsExactSourceStateSurfaceAndDefensiveCopies(t *testing.T) {
	stateManifest, err := loadManifest()
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if got, want := stateManifest.tableNames(), []string{"record_links", "record_tags"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Recovery tables = %v, want %v", got, want)
	}
	descriptor := stateManifest.descriptor()
	gotPaths := make([]string, 0, len(descriptor.Paths))
	gotRoles := make([]string, 0, len(descriptor.Paths))
	for _, path := range descriptor.Paths {
		gotPaths = append(gotPaths, path.LogicalPath)
		gotRoles = append(gotRoles, path.ContentRole)
	}
	if want := []string{"data/record_links.ndjson", "data/tags.ndjson", "data/record_tags.ndjson"}; !reflect.DeepEqual(gotPaths, want) {
		t.Fatalf("descriptor paths = %v, want %v", gotPaths, want)
	}
	if want := []string{"source_rows", "validation_rows", "source_rows"}; !reflect.DeepEqual(gotRoles, want) {
		t.Fatalf("descriptor roles = %v, want %v", gotRoles, want)
	}

	input := authoredManifestInput()
	validated, err := validateManifest(input)
	if err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
	input.relations[0].table = "changed"
	input.relations[0].columns[0] = "changed"
	input.paths[0].logicalPath = "data/changed.ndjson"
	input.paths[0].versions[0] = 99
	input.invariants[0] = "changed"
	if got := validated.tableNames(); got[0] != "record_links" {
		t.Fatalf("manifest retained caller-owned relation slice: %v", got)
	}

	tables := validated.tableNames()
	paths := validated.pathSpecs()
	projected := validated.descriptor()
	tables[0] = "changed"
	paths[0].allowedColumns[0] = "changed"
	paths[0].stableIdentity[0] = "changed"
	projected.Paths[0].Versions[0] = 99
	projected.InvariantIDs[0] = "changed"
	if got := validated.tableNames(); got[0] != "record_links" {
		t.Fatalf("table accessor exposed manifest state: %v", got)
	}
	if got := validated.pathSpecs()[0]; got.allowedColumns[0] != "record_link_id" || got.stableIdentity[0] != "record_link_id" {
		t.Fatalf("path accessor exposed manifest state: %#v", got)
	}
	if got := validated.descriptor(); got.Paths[0].Versions[0] != 3 || got.InvariantIDs[0] != expectedInvariants[0] {
		t.Fatalf("descriptor exposed manifest state: %#v", got)
	}
}

func TestManifestRejectsMalformedAuthoringMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*manifestInput)
	}{
		{name: "no relations", mutate: func(input *manifestInput) { input.relations = nil }},
		{name: "wrong relation order", mutate: func(input *manifestInput) {
			input.relations[0], input.relations[1] = input.relations[1], input.relations[0]
		}},
		{name: "duplicate relation kind", mutate: func(input *manifestInput) { input.relations[1].kind = input.relations[0].kind }},
		{name: "unsafe relation", mutate: func(input *manifestInput) { input.relations[0].table = "record_links;drop" }},
		{name: "empty columns", mutate: func(input *manifestInput) { input.relations[0].columns = nil }},
		{name: "duplicate column", mutate: func(input *manifestInput) { input.relations[0].columns[1] = input.relations[0].columns[0] }},
		{name: "wrong path order", mutate: func(input *manifestInput) { input.paths[0], input.paths[1] = input.paths[1], input.paths[0] }},
		{name: "duplicate path", mutate: func(input *manifestInput) { input.paths[1].logicalPath = input.paths[0].logicalPath }},
		{name: "unsafe path", mutate: func(input *manifestInput) { input.paths[0].logicalPath = "data/../record_links.ndjson" }},
		{name: "wrong role", mutate: func(input *manifestInput) { input.paths[0].contentRole = "validation_rows" }},
		{name: "wrong version", mutate: func(input *manifestInput) { input.paths[0].versions = []int{2, 3} }},
		{name: "missing identity", mutate: func(input *manifestInput) { input.paths[0].stableIdentity = nil }},
		{name: "unknown relation", mutate: func(input *manifestInput) { input.paths[0].relation = relationInvalid }},
		{name: "required column absent", mutate: func(input *manifestInput) { input.paths[0].requiredColumns = []string{"absent"} }},
		{name: "source path embeds columns", mutate: func(input *manifestInput) { input.paths[0].allowedColumns = []string{"record_link_id"} }},
		{name: "catalog relation", mutate: func(input *manifestInput) { input.paths[1].relation = relationRecordTags }},
		{name: "catalog columns", mutate: func(input *manifestInput) { input.paths[1].allowedColumns = []string{"tag_name"} }},
		{name: "missing invariant", mutate: func(input *manifestInput) { input.invariants = input.invariants[:len(input.invariants)-1] }},
		{name: "reordered invariants", mutate: func(input *manifestInput) {
			input.invariants[0], input.invariants[1] = input.invariants[1], input.invariants[0]
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := authoredManifestInput()
			test.mutate(&input)
			if _, err := validateManifest(input); err == nil {
				t.Fatal("malformed manifest unexpectedly passed")
			}
		})
	}
}
