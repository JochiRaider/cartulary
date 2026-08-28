package sourcestate

import (
	"reflect"
	"slices"
	"testing"
)

func TestManifestProjectsDefensiveSourceAndRecoveryFacts(t *testing.T) {
	input := authoredManifestInput()
	validated, err := validateManifest(input)
	if err != nil {
		t.Fatalf("validate manifest: %v", err)
	}
	input.ownerRelationIDs[0] = "changed"
	input.path.columns[0] = "changed"
	input.path.versions[0] = 99
	input.invariants[0] = "changed"
	input.recoveryRelations[0] = "changed"

	descriptor := validated.descriptor()
	if descriptor.Path.SchemaID != "cartulary.incident_bundle.incident.row.v1" ||
		!slices.Equal(descriptor.InvariantIDs, expectedInvariants) {
		t.Fatalf("descriptor projection drifted: %#v", descriptor)
	}
	if !reflect.DeepEqual(validated.path.columns, expectedIncidentColumns) ||
		!reflect.DeepEqual(validated.recoveryRelations, expectedRecoveryRelations) {
		t.Fatalf("validated catalog retained caller slices: %#v", validated)
	}

	descriptor.OwnerRelationIDs[0] = "changed"
	descriptor.Path.Versions[0] = 99
	descriptor.Path.StableIdentity[0] = "changed"
	descriptor.InvariantIDs[0] = "changed"
	if projected := validated.descriptor(); projected.OwnerRelationIDs[0] != "incident-core" ||
		projected.Path.Versions[0] != 3 || projected.Path.StableIdentity[0] != "id" ||
		projected.InvariantIDs[0] != expectedInvariants[0] {
		t.Fatalf("descriptor accessor exposed catalog slices: %#v", projected)
	}
}

func TestManifestRejectsMalformedAuthoringMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*manifestInput)
	}{
		{name: "family", mutate: func(input *manifestInput) { input.familyID = "incidents" }},
		{name: "contract major", mutate: func(input *manifestInput) { input.contractMajor++ }},
		{name: "owner", mutate: func(input *manifestInput) { input.ownerID = "module.other" }},
		{name: "owner relation", mutate: func(input *manifestInput) { input.ownerRelationIDs[0] = "other" }},
		{name: "unsafe path", mutate: func(input *manifestInput) { input.path.logicalPath = "data/../incident.json" }},
		{name: "content role", mutate: func(input *manifestInput) { input.path.contentRole = "source_rows" }},
		{name: "schema", mutate: func(input *manifestInput) { input.path.schemaID = "incident-row" }},
		{name: "version", mutate: func(input *manifestInput) { input.path.versions[0] = 4 }},
		{name: "identity", mutate: func(input *manifestInput) { input.path.stableIdentity[0] = "incident_key" }},
		{name: "identity invariant", mutate: func(input *manifestInput) { input.path.stableIdentityInvariantID = expectedInvariants[1] }},
		{name: "missing column", mutate: func(input *manifestInput) { input.path.columns = input.path.columns[1:] }},
		{name: "reordered columns", mutate: func(input *manifestInput) {
			input.path.columns[0], input.path.columns[1] = input.path.columns[1], input.path.columns[0]
		}},
		{name: "duplicate column", mutate: func(input *manifestInput) { input.path.columns[1] = input.path.columns[0] }},
		{name: "unsafe column", mutate: func(input *manifestInput) { input.path.columns[0] = "id;drop" }},
		{name: "missing invariant", mutate: func(input *manifestInput) { input.invariants = input.invariants[1:] }},
		{name: "reordered invariants", mutate: func(input *manifestInput) {
			input.invariants[0], input.invariants[1] = input.invariants[1], input.invariants[0]
		}},
		{name: "duplicate invariant", mutate: func(input *manifestInput) { input.invariants[1] = input.invariants[0] }},
		{name: "missing Recovery relation", mutate: func(input *manifestInput) { input.recoveryRelations = input.recoveryRelations[1:] }},
		{name: "reordered Recovery relations", mutate: func(input *manifestInput) {
			input.recoveryRelations[0], input.recoveryRelations[1] = input.recoveryRelations[1], input.recoveryRelations[0]
		}},
		{name: "duplicate Recovery relation", mutate: func(input *manifestInput) { input.recoveryRelations[1] = input.recoveryRelations[0] }},
		{name: "unsafe Recovery relation", mutate: func(input *manifestInput) { input.recoveryRelations[0] = "incidents;drop" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := authoredManifestInput()
			test.mutate(&input)
			if _, err := validateManifest(input); err == nil {
				t.Fatal("malformed source-state catalog unexpectedly passed")
			}
		})
	}
}

func TestLoadedCatalogReturnsFreshProjectionCopies(t *testing.T) {
	firstDescriptor, err := Source()
	if err != nil {
		t.Fatalf("load first descriptor: %v", err)
	}
	firstColumns, err := IncidentColumns()
	if err != nil {
		t.Fatalf("load first columns: %v", err)
	}
	firstRecovery, err := Recovery()
	if err != nil {
		t.Fatalf("load first Recovery relations: %v", err)
	}
	firstDescriptor.InvariantIDs[0] = "changed"
	firstColumns[0] = "changed"
	firstRecovery.Relations[0] = "changed"

	secondDescriptor, _ := Source()
	secondColumns, _ := IncidentColumns()
	secondRecovery, _ := Recovery()
	if secondDescriptor.InvariantIDs[0] != expectedInvariants[0] ||
		secondColumns[0] != expectedIncidentColumns[0] ||
		secondRecovery.Relations[0] != expectedRecoveryRelations[0] {
		t.Fatal("loaded source-state catalog exposed mutable projection state")
	}
}
