package sourcestate

import (
	"reflect"
	"testing"
)

func TestCatalogHasExactDeterministicSourceStateInventory(t *testing.T) {
	t.Parallel()
	catalog, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got, want := catalog.AuthoritativeRelations(), []AuthoritativeRelation{
		{TableName: "indicator_observations"},
		{TableName: "indicator_state_intervals"},
		{TableName: "indicators"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("authoritative relations = %#v, want %#v", got, want)
	}
	if got, want := catalog.RebuildableRelations(), []RebuildableRelation{
		{TableName: "indicator_active_identities", RebuildInvariantID: "indicators.restore_active_identities.v1"},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("rebuildable relations = %#v, want %#v", got, want)
	}
	if got, want := catalog.PortabilityDescriptors(), []PortabilityDescriptor{
		{
			Order:       0,
			LogicalPath: "data/indicators.ndjson", ContentRole: "source_rows",
			SchemaID: "cartulary.incident_bundle.indicators.row.v1", Versions: []int{3},
			StableIdentity: []string{"record_id"}, StableIdentityInvariantID: "indicators.source_identity_admitted",
		},
		{
			Order:       1,
			LogicalPath: "data/indicator_observations.ndjson", ContentRole: "source_rows",
			SchemaID: "cartulary.incident_bundle.indicator_observations.row.v1", Versions: []int{3},
			StableIdentity: []string{"indicator_observation_id"}, StableIdentityInvariantID: "indicators.source_identity_admitted",
		},
		{
			Order:       2,
			LogicalPath: "data/indicator_state_intervals.ndjson", ContentRole: "source_rows",
			SchemaID: "cartulary.incident_bundle.indicator_state_intervals.row.v1", Versions: []int{3},
			StableIdentity: []string{"indicator_state_interval_id"}, StableIdentityInvariantID: "indicators.source_identity_admitted",
		},
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("portability descriptors = %#v, want %#v", got, want)
	}
}

func TestCatalogRejectsInvalidSourceStateFacts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*definition)
	}{
		{name: "empty inventory", mutate: func(input *definition) { input.authoritative = nil }},
		{name: "empty relation", mutate: func(input *definition) { input.authoritative[0].TableName = "" }},
		{name: "duplicate authoritative relation", mutate: func(input *definition) { input.authoritative[1] = input.authoritative[0] }},
		{name: "relation in both classes", mutate: func(input *definition) { input.rebuildable[0].TableName = input.authoritative[0].TableName }},
		{name: "empty rebuild invariant", mutate: func(input *definition) { input.rebuildable[0].RebuildInvariantID = "" }},
		{name: "duplicate portable path", mutate: func(input *definition) {
			input.portability[1].LogicalPath = input.portability[0].LogicalPath
			input.portability[1].SchemaID = input.portability[0].SchemaID
		}},
		{name: "duplicate portable schema", mutate: func(input *definition) { input.portability[1].SchemaID = input.portability[0].SchemaID }},
		{name: "duplicate portable order", mutate: func(input *definition) { input.portability[1].Order = input.portability[0].Order }},
		{name: "missing portable order", mutate: func(input *definition) { input.portability[2].Order = 3 }},
		{name: "unsafe portable path", mutate: func(input *definition) { input.portability[0].LogicalPath = "data/../indicators.ndjson" }},
		{name: "bad path schema pairing", mutate: func(input *definition) { input.portability[0].SchemaID = "cartulary.incident_bundle.wrong.row.v1" }},
		{name: "bad version pairing", mutate: func(input *definition) { input.portability[0].Versions = []int{1, 2} }},
		{name: "empty stable identity", mutate: func(input *definition) { input.portability[0].StableIdentity = nil }},
		{name: "duplicate stable identity", mutate: func(input *definition) { input.portability[0].StableIdentity = []string{"record_id", "record_id"} }},
		{name: "unknown stable identity invariant", mutate: func(input *definition) { input.portability[0].StableIdentityInvariantID = "indicators.future" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := cloneDefinition(canonicalDefinition())
			tc.mutate(&input)
			if _, err := build(input); err == nil {
				t.Fatalf("build() accepted %s", tc.name)
			}
		})
	}
}

func TestCatalogAccessorsAreDefensive(t *testing.T) {
	t.Parallel()
	catalog, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	authoritative := catalog.AuthoritativeRelations()
	authoritative[0].TableName = "mutated"
	if catalog.AuthoritativeRelations()[0].TableName == "mutated" {
		t.Fatal("authoritative relation mutation escaped the catalog")
	}
	rebuildable := catalog.RebuildableRelations()
	rebuildable[0].RebuildInvariantID = "mutated"
	if catalog.RebuildableRelations()[0].RebuildInvariantID == "mutated" {
		t.Fatal("rebuildable relation mutation escaped the catalog")
	}
	portable := catalog.PortabilityDescriptors()
	portable[0].LogicalPath = "mutated"
	portable[0].Versions[0] = 999
	portable[0].StableIdentity[0] = "mutated"
	again := catalog.PortabilityDescriptors()[0]
	if again.LogicalPath == "mutated" || again.Versions[0] == 999 || again.StableIdentity[0] == "mutated" {
		t.Fatal("portability descriptor mutation escaped the catalog")
	}
}

func cloneDefinition(input definition) definition {
	return definition{
		authoritative: append([]AuthoritativeRelation(nil), input.authoritative...),
		rebuildable:   append([]RebuildableRelation(nil), input.rebuildable...),
		portability:   clonePortability(input.portability),
	}
}
