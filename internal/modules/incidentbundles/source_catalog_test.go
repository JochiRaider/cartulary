package incidentbundles_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

func TestSourcePortCatalogCurrentOrderAndExactPathAccounting_Unit(t *testing.T) {
	catalog, err := incidentportabilityassembly.NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	var families []string
	for _, descriptor := range catalog.Descriptors() {
		families = append(families, descriptor.FamilyID)
		stableIdentityInvariantID := descriptor.FamilyID + ".source_identity_admitted"
		if !slices.Contains(descriptor.InvariantIDs, stableIdentityInvariantID) {
			t.Fatalf("family %q does not declare %q", descriptor.FamilyID, stableIdentityInvariantID)
		}
		for _, path := range descriptor.Paths {
			if path.StableIdentityInvariantID != stableIdentityInvariantID {
				t.Fatalf("path %q stable identity invariant = %q, want %q", path.LogicalPath, path.StableIdentityInvariantID, stableIdentityInvariantID)
			}
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
	assertAuthoredSourceCatalogV3(t)
	assertRevisionsCatalogProjection(t, catalog.Descriptors())
	assertSavedViewsCatalogProjection(t, catalog.Descriptors())
	if consumer, ok := catalog.ConsumerFor(2, "data/timeline_source_provenance.ndjson"); !ok || consumer != "timeline" {
		t.Fatalf("Timeline v2 provenance consumer = %q, %v", consumer, ok)
	}
	if consumer, ok := catalog.ConsumerFor(1, "data/timeline_source_provenance.ndjson"); ok || consumer != "" {
		t.Fatalf("retired bundle-version consumer = %q, %v", consumer, ok)
	}
}

func TestSourcePortCatalogRejectsInvalidDescriptors_Unit(t *testing.T) {
	_, err := sourceport.NewCatalog(sourceport.CatalogOptions{
		Ports: []sourceport.Port{
			sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: sourceport.Descriptor{
				FamilyID: "duplicate", ContractMajor: sourceport.ContractMajor, OwnerID: "module.one",
				OwnerRelationIDs: []string{"owner"}, Paths: []sourceport.Path{{LogicalPath: "data/value.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "duplicate.valid"}},
				InvariantIDs: []string{"duplicate.valid"},
			}}),
			sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: sourceport.Descriptor{
				FamilyID: "duplicate", ContractMajor: sourceport.ContractMajor, OwnerID: "module.two",
				OwnerRelationIDs: []string{"owner"}, Paths: []sourceport.Path{{LogicalPath: "data/other.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "duplicate.valid"}},
				InvariantIDs: []string{"duplicate.valid"},
			}}),
		},
		RequiredPathsByVersion: map[int][]string{2: {"data/value.ndjson", "data/other.ndjson"}},
		AllowedRelationIDs:     map[string]struct{}{"owner": {}},
	})
	if err == nil {
		t.Fatal("duplicate source family must fail closed")
	}

	for name, invariantID := range map[string]string{
		"missing stable identity invariant":      "",
		"undeclared stable identity invariant":   "fixture.undeclared",
		"cross-family stable identity invariant": "attacker.selected",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := sourceport.NewCatalog(sourceport.CatalogOptions{
				Ports: []sourceport.Port{sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: sourceport.Descriptor{
					FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
					OwnerRelationIDs: []string{"owner"},
					Paths:            []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}, StableIdentityInvariantID: invariantID}},
					InvariantIDs:     []string{"fixture.source_identity_admitted"},
				}})},
				RequiredPathsByVersion: map[int][]string{2: {"data/fixture.ndjson"}},
				AllowedRelationIDs:     map[string]struct{}{"owner": {}},
			})
			if !errors.Is(err, sourceport.ErrInvalidCatalog) {
				t.Fatalf("invalid stable identity invariant error = %v; want ErrInvalidCatalog", err)
			}
		})
	}
	t.Run("path identity selection and unknown path", assertSourcePathIdentityFailuresAreOrderIndependentAndUnknownPathsFailClosed)
}

func TestSourcePortDescriptorsAreImmutableAndPreparedValuesAreOperationBound_Unit(t *testing.T) {
	catalog, err := incidentportabilityassembly.NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	descriptors := catalog.Descriptors()
	descriptors[0].InvariantIDs[len(descriptors[0].InvariantIDs)-1] = "mutated"
	descriptors[0].Paths[0].StableIdentity[len(descriptors[0].Paths[0].StableIdentity)-1] = "mutated"
	fresh := catalog.Descriptors()[0]
	if slices.Contains(fresh.InvariantIDs, "mutated") || slices.Contains(fresh.Paths[0].StableIdentity, "mutated") {
		t.Fatal("catalog descriptor mutation escaped the immutable catalog boundary")
	}

	descriptor := sourceport.Descriptor{
		FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
		Paths:        []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "fixture.identity"}},
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
		Validate: func(context.Context, pgx.Tx, any, sourceport.ImportContext) error {
			t.Fatal("operation-mismatched prepared value reached its owner validate function")
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
	err = port.ValidateImportTx(
		context.Background(),
		nil,
		prepared,
		sourceport.ImportContext{OperationID: "operation-b"},
	)
	if !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("cross-operation validation error = %v; want ErrPreparedBinding", err)
	}
}

func assertSourcePathIdentityFailuresAreOrderIndependentAndUnknownPathsFailClosed(t *testing.T) {
	for _, invariants := range [][]string{
		{"fixture.first", "fixture.source_identity_admitted"},
		{"fixture.source_identity_admitted", "fixture.first"},
	} {
		descriptor := sourceport.Descriptor{
			FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
			Paths:        []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "fixture.source_identity_admitted"}},
			InvariantIDs: invariants,
		}
		for name, payload := range map[string][]byte{
			"missing":   []byte("{\"other\":\"value\"}\n"),
			"duplicate": []byte("{\"id\":\"same\"}\n{\"id\":\"same\"}\n"),
		} {
			t.Run(name, func(t *testing.T) {
				_, err := sourceport.PrepareFiles(descriptor, sourceport.MapBundle{"data/fixture.ndjson": payload}, 2)
				var failure *sourceport.Failure
				if !errors.As(err, &failure) || failure.FamilyID() != "fixture" || failure.InvariantID() != "fixture.source_identity_admitted" {
					t.Fatalf("identity failure = %#v, %v", failure, err)
				}
			})
		}
	}

	descriptor := sourceport.Descriptor{
		FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
		Paths:        []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{2}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "fixture.source_identity_admitted"}},
		InvariantIDs: []string{"fixture.source_identity_admitted"},
	}
	port := sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Prepare: func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) {
			return "prepared", nil
		},
		Apply: func(context.Context, pgx.Tx, any, sourceport.ImportContext) error {
			return incidentportability.FixedImportFailure("data/unknown.ndjson")
		},
	})
	importContext := sourceport.ImportContext{OperationID: "operation"}
	prepared, err := port.PrepareImport(context.Background(), sourceport.MapBundle{}, importContext)
	if err != nil {
		t.Fatalf("prepare unknown-path fixture: %v", err)
	}
	if err := port.ApplyImportTx(context.Background(), nil, prepared, importContext); !errors.Is(err, sourceport.ErrInvalidCatalog) {
		t.Fatalf("unknown fixed-import path error = %v; want ErrInvalidCatalog", err)
	}
}

type sourceCatalogProjection struct {
	SchemaID      string `json:"schema_id"`
	ContractMajor int    `json:"contract_major"`
	Families      []struct {
		FamilyID     string            `json:"family_id"`
		Paths        []sourceport.Path `json:"paths"`
		InvariantIDs []string          `json:"invariant_ids"`
	} `json:"families"`
	SpecialConsumers []struct {
		FamilyID string `json:"family_id"`
		Versions []int  `json:"versions"`
	} `json:"special_consumers"`
}

func assertAuthoredSourceCatalogV3(t *testing.T) {
	t.Helper()
	var authored sourceCatalogProjection
	readContractJSON(t, "source_catalog.json", &authored)
	if authored.SchemaID != "cartulary.incident_bundle_source_catalog.v3" || authored.ContractMajor != sourceport.ContractMajor || len(authored.Families) != 13 {
		t.Fatalf("authored source catalog header = %q major %d families %d", authored.SchemaID, authored.ContractMajor, len(authored.Families))
	}
	for _, family := range authored.Families {
		invariantID := family.FamilyID + ".source_identity_admitted"
		if !slices.Contains(family.InvariantIDs, invariantID) {
			t.Fatalf("authored family %q does not declare %q", family.FamilyID, invariantID)
		}
		for _, path := range family.Paths {
			if !slices.Equal(path.Versions, []int{2}) {
				t.Fatalf("authored path %q versions = %#v, want [2]", path.LogicalPath, path.Versions)
			}
			if path.StableIdentityInvariantID != invariantID {
				t.Fatalf("authored path %q invariant = %q, want %q", path.LogicalPath, path.StableIdentityInvariantID, invariantID)
			}
		}
	}
	if len(authored.SpecialConsumers) != 3 {
		t.Fatalf("authored special consumers = %d, want 3", len(authored.SpecialConsumers))
	}
	for _, consumer := range authored.SpecialConsumers {
		if !slices.Equal(consumer.Versions, []int{2}) {
			t.Fatalf("authored special consumer %q versions = %#v, want [2]", consumer.FamilyID, consumer.Versions)
		}
	}
}

type rowSchemaProjection struct {
	ID                   string   `json:"$id"`
	Type                 string   `json:"type"`
	AdditionalProperties bool     `json:"additionalProperties"`
	Required             []string `json:"required"`
}

func assertRevisionsCatalogProjection(t *testing.T, descriptors []sourceport.Descriptor) {
	t.Helper()
	var runtime sourceport.Descriptor
	for _, descriptor := range descriptors {
		if descriptor.FamilyID == "revisions" {
			runtime = descriptor
			break
		}
	}
	if runtime.FamilyID == "" || len(runtime.Paths) != 3 {
		t.Fatal("runtime source catalog must expose the three Revisions paths")
	}

	var authored sourceCatalogProjection
	readContractJSON(t, "source_catalog.json", &authored)
	var authoredPaths []sourceport.Path
	for _, family := range authored.Families {
		if family.FamilyID == "revisions" {
			authoredPaths = family.Paths
			break
		}
	}
	if len(authoredPaths) != len(runtime.Paths) {
		t.Fatalf("authored Revisions paths = %d, runtime = %d", len(authoredPaths), len(runtime.Paths))
	}
	slices.SortFunc(authoredPaths, func(left, right sourceport.Path) int {
		if left.LogicalPath < right.LogicalPath {
			return -1
		}
		if left.LogicalPath > right.LogicalPath {
			return 1
		}
		return 0
	})
	for index, authoredPath := range authoredPaths {
		runtimePath := runtime.Paths[index]
		if authoredPath.LogicalPath != runtimePath.LogicalPath ||
			authoredPath.ContentRole != runtimePath.ContentRole ||
			authoredPath.SchemaID != runtimePath.SchemaID ||
			authoredPath.StableIdentityInvariantID != runtimePath.StableIdentityInvariantID ||
			!slices.Equal(authoredPath.Versions, runtimePath.Versions) ||
			!slices.Equal(authoredPath.StableIdentity, runtimePath.StableIdentity) {
			t.Fatalf("Revisions path projection drift:\nauthored=%#v\nruntime=%#v", authoredPath, runtimePath)
		}
		var rowSchema rowSchemaProjection
		schemaFile := map[string]string{
			"data/change_sets.ndjson":          "change_sets.row.v1.schema.json",
			"data/change_set_mutations.ndjson": "change_set_mutations.row.v1.schema.json",
			"data/record_revisions.ndjson":     "record_revisions.row.v1.schema.json",
		}[runtimePath.LogicalPath]
		readContractJSON(t, schemaFile, &rowSchema)
		if rowSchema.ID != runtimePath.SchemaID || rowSchema.Type != "object" ||
			rowSchema.AdditionalProperties || len(rowSchema.Required) == 0 {
			t.Fatalf("Revisions row schema is not a closed runtime projection: %#v", rowSchema)
		}
	}
}

func assertSavedViewsCatalogProjection(t *testing.T, descriptors []sourceport.Descriptor) {
	t.Helper()
	var runtime sourceport.Descriptor
	for _, descriptor := range descriptors {
		if descriptor.FamilyID == "saved_views" {
			runtime = descriptor
			break
		}
	}
	if runtime.FamilyID == "" {
		t.Fatal("runtime source catalog is missing saved_views")
	}

	var authored sourceCatalogProjection
	readContractJSON(t, "source_catalog.json", &authored)
	var projection struct {
		FamilyID     string
		Paths        []sourceport.Path
		InvariantIDs []string
	}
	for _, family := range authored.Families {
		if family.FamilyID == "saved_views" {
			projection.FamilyID = family.FamilyID
			projection.Paths = family.Paths
			projection.InvariantIDs = family.InvariantIDs
			break
		}
	}
	if projection.FamilyID == "" || len(projection.Paths) != 1 || len(runtime.Paths) != 1 {
		t.Fatal("authored and runtime saved_views catalogs must each expose one path")
	}
	authoredPath := projection.Paths[0]
	runtimePath := runtime.Paths[0]
	if authoredPath.LogicalPath != runtimePath.LogicalPath ||
		authoredPath.ContentRole != runtimePath.ContentRole ||
		authoredPath.SchemaID != runtimePath.SchemaID ||
		authoredPath.StableIdentityInvariantID != runtimePath.StableIdentityInvariantID ||
		!slices.Equal(authoredPath.Versions, runtimePath.Versions) ||
		!slices.Equal(authoredPath.StableIdentity, runtimePath.StableIdentity) {
		t.Fatalf("saved_views path projection drift:\nauthored=%#v\nruntime=%#v", authoredPath, runtimePath)
	}
	authoredInvariants := append([]string(nil), projection.InvariantIDs...)
	slices.Sort(authoredInvariants)
	if !slices.Equal(authoredInvariants, runtime.InvariantIDs) {
		t.Fatalf("saved_views invariant projection drift:\nauthored=%#v\nruntime=%#v", authoredInvariants, runtime.InvariantIDs)
	}

	var rowSchema rowSchemaProjection
	readContractJSON(t, "saved_views.row.v1.schema.json", &rowSchema)
	required := []string{
		"saved_view_id", "incident_id", "view_schema_id", "scope", "display_name",
		"query_json", "layout_json", "owner_user_id", "created_at", "updated_at",
		"saved_view_version",
	}
	if rowSchema.ID != runtimePath.SchemaID || rowSchema.Type != "object" ||
		rowSchema.AdditionalProperties || !slices.Equal(rowSchema.Required, required) {
		t.Fatalf("saved_views row schema is not the exact closed runtime projection: %#v", rowSchema)
	}
}

func readContractJSON(t *testing.T, name string, target any) {
	t.Helper()
	path := filepath.Join("..", "..", "..", "contracts", "incident-bundles", name)
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(payload, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
