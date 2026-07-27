package incidentbundles_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

func TestSourcePortCatalogCurrentOrderAndExactPathAccounting_Unit(t *testing.T) {
	catalog, err := incidentportabilityassembly.NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	var families []string
	for _, descriptor := range catalog.Descriptors() {
		families = append(families, descriptor.FamilyID)
		for _, path := range descriptor.Paths {
			for _, version := range path.Versions {
				consumer, ok := catalog.ConsumerFor(version, path.LogicalPath)
				if !ok || consumer != descriptor.FamilyID {
					t.Fatalf("path %q version %d consumer = %q, %v", path.LogicalPath, version, consumer, ok)
				}
			}
		}
	}
	want := []string{
		"incident", "records", "timeline", "parties", "entities", "indicators",
		"artifacts", "tasks_decisions", "evidence", "assessments", "links_tags",
		"revisions", "saved_views",
	}
	if !slices.Equal(families, want) {
		t.Fatalf("catalog order = %#v, want %#v", families, want)
	}
	for version, path := range map[int]string{1: "data/timeline_events.ndjson", 2: "data/timeline_source_provenance.ndjson"} {
		if consumer, ok := catalog.ConsumerFor(version, path); !ok || consumer != "timeline" {
			t.Fatalf("Timeline path %q version %d consumer = %q, %v", path, version, consumer, ok)
		}
	}
}

func TestSourcePortCatalogRejectsInvalidDescriptors_Unit(t *testing.T) {
	_, err := sourceport.NewCatalog(sourceport.CatalogOptions{
		Ports: []sourceport.Port{
			sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: sourceport.Descriptor{
				FamilyID: "duplicate", ContractMajor: sourceport.ContractMajor, OwnerID: "module.one",
				OwnerRelationIDs: []string{"owner"}, Paths: []sourceport.Path{{LogicalPath: "data/value.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}}},
				InvariantIDs: []string{"duplicate.valid"},
			}}),
			sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: sourceport.Descriptor{
				FamilyID: "duplicate", ContractMajor: sourceport.ContractMajor, OwnerID: "module.two",
				OwnerRelationIDs: []string{"owner"}, Paths: []sourceport.Path{{LogicalPath: "data/other.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}}},
				InvariantIDs: []string{"duplicate.valid"},
			}}),
		},
		RequiredPathsByVersion: map[int][]string{2: {"data/value.ndjson", "data/other.ndjson"}},
		AllowedRelationIDs:     map[string]struct{}{"owner": {}},
	})
	if err == nil {
		t.Fatal("duplicate source family must fail closed")
	}
}

func TestSourcePortDescriptorsAreImmutableAndPreparedValuesAreOperationBound_Unit(t *testing.T) {
	catalog, err := incidentportabilityassembly.NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	descriptors := catalog.Descriptors()
	descriptors[0].InvariantIDs[0] = "mutated"
	descriptors[0].Paths[0].StableIdentity[0] = "mutated"
	fresh := catalog.Descriptors()[0]
	if fresh.InvariantIDs[0] == "mutated" || fresh.Paths[0].StableIdentity[0] == "mutated" {
		t.Fatal("catalog descriptor mutation escaped the immutable catalog boundary")
	}

	descriptor := sourceport.Descriptor{
		FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
		Paths:        []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}}},
		InvariantIDs: []string{"fixture.identity"},
	}
	port := sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Prepare: func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) {
			return "prepared", nil
		},
		Apply: func(context.Context, pgx.Tx, any, sourceport.ImportContext) error {
			t.Fatal("operation-mismatched prepared value reached its owner apply function")
			return nil
		},
	})
	prepared, err := port.PrepareImport(context.Background(), sourceport.MapBundle{}, sourceport.ImportContext{OperationID: "operation-a"})
	if err != nil {
		t.Fatalf("prepare operation-bound source: %v", err)
	}
	err = port.ApplyImportTx(context.Background(), nil, prepared, sourceport.ImportContext{OperationID: "operation-b"})
	if !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("cross-operation prepared value error = %v; want ErrPreparedBinding", err)
	}
}
