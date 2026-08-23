package indicators

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"testing"

	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

type indicatorSourcePathProjection struct {
	LogicalPath               string   `json:"logical_path"`
	ContentRole               string   `json:"content_role"`
	SchemaID                  string   `json:"schema_id,omitempty"`
	Versions                  []int    `json:"versions"`
	StableIdentity            []string `json:"stable_identity"`
	StableIdentityInvariantID string   `json:"stable_identity_invariant_id"`
}

type indicatorSourceDescriptorProjection struct {
	FamilyID         string
	ContractMajor    int
	OwnerID          string
	OwnerRelationIDs []string
	Dependencies     []string
	Paths            []indicatorSourcePathProjection
	InvariantIDs     []string
}

func TestIndicatorSourceStateProjectionsHaveExactParity(t *testing.T) {
	t.Parallel()

	contribution, err := NewIncidentBundleContribution()
	if err != nil {
		t.Fatalf("NewIncidentBundleContribution() error = %v", err)
	}
	runtimeDescriptor := contribution.SourcePort.Descriptor()
	gotDescriptor := indicatorSourceDescriptorProjection{
		FamilyID: runtimeDescriptor.FamilyID, ContractMajor: runtimeDescriptor.ContractMajor,
		OwnerID: runtimeDescriptor.OwnerID, OwnerRelationIDs: runtimeDescriptor.OwnerRelationIDs,
		Dependencies: runtimeDescriptor.Dependencies, InvariantIDs: runtimeDescriptor.InvariantIDs,
		Paths: make([]indicatorSourcePathProjection, 0, len(runtimeDescriptor.Paths)),
	}
	for _, path := range runtimeDescriptor.Paths {
		gotDescriptor.Paths = append(gotDescriptor.Paths, indicatorSourcePathProjection{
			LogicalPath: path.LogicalPath, ContentRole: path.ContentRole, SchemaID: path.SchemaID,
			Versions: path.Versions, StableIdentity: path.StableIdentity,
			StableIdentityInvariantID: path.StableIdentityInvariantID,
		})
	}
	wantDescriptor := authoredIndicatorSourceDescriptor(t)
	canonicalizeSourceDescriptor(&gotDescriptor)
	canonicalizeSourceDescriptor(&wantDescriptor)
	if !reflect.DeepEqual(gotDescriptor, wantDescriptor) {
		t.Fatalf("runtime Incident Bundle descriptor differs from authored projection:\ngot:  %#v\nwant: %#v", gotDescriptor, wantDescriptor)
	}

	recoveryContribution, err := RecoveryStateContribution()
	if err != nil {
		t.Fatalf("RecoveryStateContribution() error = %v", err)
	}
	wantTables := append(
		recoverystate.AuthoritativeTables("indicator_observations", "indicator_state_intervals", "indicators"),
		recoverystate.RebuildableTables("indicators.restore_active_identities.v1", "indicator_active_identities")...,
	)
	if recoveryContribution.OwnerID != "module.indicators" ||
		recoveryContribution.SchemaID != recoverystate.ContributionSchemaID ||
		!reflect.DeepEqual(recoveryContribution.Tables, wantTables) ||
		len(recoveryContribution.ObjectFamilies) != 0 {
		t.Fatalf("runtime Recovery contribution = %#v, want exact 3 authoritative / 1 rebuildable inventory", recoveryContribution)
	}
}

func TestIndicatorSourceStateProjectionOrderingAndCopiesAreStable(t *testing.T) {
	t.Parallel()
	first, err := NewIncidentBundleContribution()
	if err != nil {
		t.Fatalf("first contribution: %v", err)
	}
	second, err := NewIncidentBundleContribution()
	if err != nil {
		t.Fatalf("second contribution: %v", err)
	}
	firstBytes, err := json.Marshal(first.SourcePort.Descriptor().Paths)
	if err != nil {
		t.Fatalf("marshal first descriptor: %v", err)
	}
	secondBytes, err := json.Marshal(second.SourcePort.Descriptor().Paths)
	if err != nil {
		t.Fatalf("marshal second descriptor: %v", err)
	}
	if !slices.Equal(firstBytes, secondBytes) {
		t.Fatalf("source-state path bytes are nondeterministic:\nfirst:  %s\nsecond: %s", firstBytes, secondBytes)
	}
	paths := first.SourcePort.Descriptor().Paths
	wantPaths := []string{
		"data/indicators.ndjson",
		"data/indicator_observations.ndjson",
		"data/indicator_state_intervals.ndjson",
	}
	for index, wantPath := range wantPaths {
		if paths[index].LogicalPath != wantPath {
			t.Fatalf("source-state path %d = %q, want %q", index, paths[index].LogicalPath, wantPath)
		}
	}

	mutated := first.SourcePort.Descriptor()
	mutated.Paths[0].LogicalPath = "mutated"
	mutated.Paths[0].Versions[0] = 999
	mutated.Paths[0].StableIdentity[0] = "mutated"
	again := first.SourcePort.Descriptor().Paths[0]
	if again.LogicalPath == "mutated" || again.Versions[0] == 999 || again.StableIdentity[0] == "mutated" {
		t.Fatal("source-port descriptor mutation escaped the immutable contribution boundary")
	}
}

func authoredIndicatorSourceDescriptor(t testing.TB) indicatorSourceDescriptorProjection {
	t.Helper()
	type authoredFamily struct {
		FamilyID         string                          `json:"family_id"`
		OwnerID          string                          `json:"owner_id"`
		OwnerRelationIDs []string                        `json:"owner_relation_ids"`
		Dependencies     []string                        `json:"dependencies"`
		Paths            []indicatorSourcePathProjection `json:"paths"`
		InvariantIDs     []string                        `json:"invariant_ids"`
	}
	type authoredCatalog struct {
		ContractMajor int              `json:"contract_major"`
		Families      []authoredFamily `json:"families"`
	}
	payload, err := os.ReadFile(filepath.Join("..", "..", "..", "contracts", "incident-bundles", "source_catalog.json"))
	if err != nil {
		t.Fatalf("read authored Incident Bundle source catalog: %v", err)
	}
	var catalog authoredCatalog
	if err := json.Unmarshal(payload, &catalog); err != nil {
		t.Fatalf("decode authored Incident Bundle source catalog: %v", err)
	}
	for _, family := range catalog.Families {
		if family.FamilyID != "indicators" {
			continue
		}
		return indicatorSourceDescriptorProjection{
			FamilyID: family.FamilyID, ContractMajor: catalog.ContractMajor,
			OwnerID: family.OwnerID, OwnerRelationIDs: family.OwnerRelationIDs,
			Dependencies: family.Dependencies, Paths: family.Paths, InvariantIDs: family.InvariantIDs,
		}
	}
	t.Fatal("authored Incident Bundle source catalog has no Indicators family")
	return indicatorSourceDescriptorProjection{}
}

func canonicalizeSourceDescriptor(descriptor *indicatorSourceDescriptorProjection) {
	slices.Sort(descriptor.Dependencies)
	slices.Sort(descriptor.OwnerRelationIDs)
	slices.Sort(descriptor.InvariantIDs)
	sort.Slice(descriptor.Paths, func(left, right int) bool {
		return descriptor.Paths[left].LogicalPath < descriptor.Paths[right].LogicalPath
	})
	for index := range descriptor.Paths {
		slices.Sort(descriptor.Paths[index].Versions)
		slices.Sort(descriptor.Paths[index].StableIdentity)
	}
}
