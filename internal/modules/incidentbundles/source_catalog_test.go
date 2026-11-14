package incidentbundles_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestIncidentsSourcePortRejectsUnboundPreparedWithoutPanic_Unit(t *testing.T) {
	port, err := incidents.NewIncidentBundleSourcePort()
	if err != nil {
		t.Fatalf("construct Incidents source port: %v", err)
	}
	importContext := sourceport.ImportContext{OperationID: "unbound-prepared", IncidentID: uuid.New(), ActorUserID: uuid.New(), BundleVersion: 8}
	if err := port.ApplyImportTx(context.Background(), nil, sourceport.Prepared{}, importContext); !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("unbound prepared apply error = %v; want ErrPreparedBinding", err)
	}
	if err := port.ValidateImportTx(context.Background(), nil, sourceport.Prepared{}, importContext); !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("unbound prepared validation error = %v; want ErrPreparedBinding", err)
	}
}

func TestSourcePortCatalogCurrentOrderAndExactPathAccounting_Unit(t *testing.T) {
	catalog, revisionsPort := newSourcePortCatalog(t)
	var families []string
	for index, descriptor := range catalog.Descriptors() {
		families = append(families, descriptor.FamilyID)
		if descriptor.FamilyID == "revisions" && catalog.Ports()[index] != revisionsPort {
			t.Fatal("source catalog did not retain the supplied Revisions source port")
		}
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
		"incident", "records", "committed_sources", "timeline", "parties", "entities", "indicators",
		"artifacts", "tasks_decisions", "evidence", "assessments", "links_tags",
		"revisions", "saved_views",
	}
	if !slices.Equal(families, want) {
		t.Fatalf("catalog order = %#v, want %#v", families, want)
	}
	var assessmentDescriptor sourceport.Descriptor
	for _, descriptor := range catalog.Descriptors() {
		if descriptor.FamilyID == "assessments" {
			assessmentDescriptor = descriptor
			break
		}
	}
	if assessmentDescriptor.FamilyID != "assessments" ||
		assessmentDescriptor.OwnerID != "module.assessments" ||
		!slices.Equal(assessmentDescriptor.Dependencies, []string{"evidence"}) ||
		len(assessmentDescriptor.Paths) != 1 {
		t.Fatalf("assessment source descriptor drifted: %#v", assessmentDescriptor)
	}
	assessmentPath := assessmentDescriptor.Paths[0]
	if assessmentPath.LogicalPath != "data/compromise_assessments.ndjson" ||
		assessmentPath.ContentRole != "source_rows" ||
		!slices.Equal(assessmentPath.Versions, []int{8}) ||
		!slices.Equal(assessmentPath.StableIdentity, []string{"record_id"}) ||
		assessmentPath.StableIdentityInvariantID != "assessments.source_identity_admitted" {
		t.Fatalf("assessment source path drifted: %#v", assessmentPath)
	}
	assertAuthoredSourceCatalog(t)
	assertRuntimeSourceCatalogProjection(t, catalog.Descriptors())
	assertRevisionsCatalogProjection(t, catalog.Descriptors())
	assertSavedViewsCatalogProjection(t, catalog.Descriptors())
	if consumer, ok := catalog.ConsumerFor(8, "data/timeline_source_provenance.ndjson"); !ok || consumer != "timeline" {
		t.Fatalf("Timeline v7 provenance consumer = %q, %v", consumer, ok)
	}
	if consumer, ok := catalog.ConsumerFor(3, "data/timeline_source_provenance.ndjson"); ok || consumer != "" {
		t.Fatalf("retired bundle-version consumer = %q, %v", consumer, ok)
	}
}

func TestSourcePortCatalogRejectsInvalidDescriptors_Unit(t *testing.T) {
	completePort := func(family, owner, path, identityInvariant string) sourceport.Port {
		return sourceport.NewAdapter(sourceport.AdapterOptions{
			Descriptor: sourceport.Descriptor{
				FamilyID: family, ContractMajor: sourceport.ContractMajor, OwnerID: owner,
				OwnerRelationIDs: []string{"owner"},
				Paths:            []sourceport.Path{{LogicalPath: path, ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"id"}, StableIdentityInvariantID: identityInvariant}},
				InvariantIDs:     []string{family + ".source_identity_admitted"},
			},
			Export:   func(context.Context, sourceport.ExportContext) error { return nil },
			Prepare:  func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) { return nil, nil },
			Apply:    func(context.Context, pgx.Tx, any, sourceport.ImportContext) error { return nil },
			Validate: func(context.Context, pgx.Tx, any, sourceport.ImportContext) error { return nil },
		})
	}
	for _, secondFamily := range []string{"other", "fixture"} {
		t.Run("second family/"+secondFamily, func(t *testing.T) {
			_, err := sourceport.NewCatalog(sourceport.CatalogOptions{
				Ports: []sourceport.Port{
					completePort("fixture", "module.one", "data/value.ndjson", "fixture.source_identity_admitted"),
					completePort(secondFamily, "module.two", "data/other.ndjson", secondFamily+".source_identity_admitted"),
				},
				RequiredPathsByVersion: map[int][]string{8: {"data/value.ndjson", "data/other.ndjson"}},
				AllowedRelationIDs:     map[string]struct{}{"owner": {}},
			})
			if secondFamily == "other" {
				if err != nil {
					t.Fatalf("distinct complete source families must be admitted: %v", err)
				}
			} else if !errors.Is(err, sourceport.ErrInvalidCatalog) {
				t.Fatalf("duplicate source family must fail closed: %v", err)
			}
		})
	}

	for name, invariantID := range map[string]string{
		"declared stable identity invariant":     "fixture.source_identity_admitted",
		"missing stable identity invariant":      "",
		"undeclared stable identity invariant":   "fixture.undeclared",
		"cross-family stable identity invariant": "attacker.selected",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := sourceport.NewCatalog(sourceport.CatalogOptions{
				Ports:                  []sourceport.Port{completePort("fixture", "module.fixture", "data/fixture.ndjson", invariantID)},
				RequiredPathsByVersion: map[int][]string{8: {"data/fixture.ndjson"}},
				AllowedRelationIDs:     map[string]struct{}{"owner": {}},
			})
			if invariantID == "fixture.source_identity_admitted" {
				if err != nil {
					t.Fatalf("declared identity invariant must be admitted: %v", err)
				}
			} else if !errors.Is(err, sourceport.ErrInvalidCatalog) {
				t.Fatalf("invalid stable identity invariant error = %v; want ErrInvalidCatalog", err)
			}
		})
	}
	t.Run("version-disjoint path grammars", func(t *testing.T) {
		path := sourceport.Path{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", SchemaID: "cartulary.fixture.row.v1", Versions: []int{6}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "fixture.source_identity_admitted"}
		current := path
		current.SchemaID = "cartulary.fixture.row.v2"
		current.Versions = []int{8}
		for _, tc := range []struct {
			name  string
			paths []sourceport.Path
			valid bool
		}{
			{"disjoint", []sourceport.Path{path, current}, true},
			{"overlap", []sourceport.Path{path, path, current}, false},
			{"duplicate version", []sourceport.Path{func() sourceport.Path { p := path; p.Versions = []int{6, 6, 8}; return p }()}, false},
			{"changed identity", []sourceport.Path{path, func() sourceport.Path { p := current; p.StableIdentity = []string{"other"}; return p }()}, false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				_, err := sourceport.NewCatalog(sourceport.CatalogOptions{
					Ports: []sourceport.Port{sourceport.NewAdapter(sourceport.AdapterOptions{Descriptor: sourceport.Descriptor{FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture", OwnerRelationIDs: []string{"owner"}, Paths: tc.paths, InvariantIDs: []string{"fixture.source_identity_admitted"}},
						Export:   func(context.Context, sourceport.ExportContext) error { return nil },
						Prepare:  func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) { return nil, nil },
						Apply:    func(context.Context, pgx.Tx, any, sourceport.ImportContext) error { return nil },
						Validate: func(context.Context, pgx.Tx, any, sourceport.ImportContext) error { return nil },
					})},
					RequiredPathsByVersion: map[int][]string{6: {path.LogicalPath}, 8: {path.LogicalPath}}, AllowedRelationIDs: map[string]struct{}{"owner": {}},
				})
				if (err == nil) != tc.valid {
					t.Fatalf("valid=%v error=%v", tc.valid, err)
				}
			})
		}
	})

	t.Run("path identity selection and unknown path", assertSourcePathIdentityFailuresAreOrderIndependentAndUnknownPathsFailClosed)
}

func TestSourcePortDescriptorsAreImmutableAndPreparedValuesAreOperationBound_Unit(t *testing.T) {
	catalog, _ := newSourcePortCatalog(t)
	descriptors := catalog.Descriptors()
	if len(descriptors) == 0 {
		t.Fatal("current catalog must contain source descriptors")
	}
	before, err := json.Marshal(descriptors)
	if err != nil {
		t.Fatal(err)
	}
	for index := range descriptors {
		descriptor := &descriptors[index]
		descriptor.FamilyID, descriptor.OwnerID, descriptor.ContractMajor = "mutated", "mutated", -1
		for _, values := range [][]string{descriptor.Dependencies, descriptor.OwnerRelationIDs, descriptor.InvariantIDs} {
			for i := range values {
				values[i] = "mutated"
			}
		}
		for i := range descriptor.Paths {
			path := &descriptor.Paths[i]
			path.LogicalPath, path.ContentRole, path.SchemaID, path.StableIdentityInvariantID = "mutated", "mutated", "mutated", "mutated"
			for j := range path.Versions {
				path.Versions[j] = -1
			}
			for j := range path.StableIdentity {
				path.StableIdentity[j] = "mutated"
			}
		}
	}
	after, err := json.Marshal(catalog.Descriptors())
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("catalog changed through returned descriptors:\nbefore: %s\nafter: %s", before, after)
	}

	descriptor := sourceport.Descriptor{
		FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
		Paths:        []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "fixture.identity"}},
		InvariantIDs: []string{"fixture.identity"},
	}
	prepareCalls := 0
	port := sourceport.NewAdapter(sourceport.AdapterOptions{
		Descriptor: descriptor,
		Prepare: func(context.Context, sourceport.Bundle, sourceport.ImportContext) (any, error) {
			prepareCalls++
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
	importContext := sourceport.ImportContext{OperationID: "operation-a", IncidentID: uuid.New(), ActorUserID: uuid.New(), BundleVersion: 8}
	for _, version := range []int{0, 1, 2, 3, 4, 5, 6, 7, 9} {
		rejected := importContext
		rejected.BundleVersion = version
		_, err := port.PrepareImport(context.Background(), portabilitysupport.MapBundle{}, rejected)
		if !errors.Is(err, sourceport.ErrInvalidCatalog) || prepareCalls != 0 {
			t.Fatalf("unsupported version %d reached source preparation: calls=%d error=%v", version, prepareCalls, err)
		}
	}
	prepared, err := port.PrepareImport(context.Background(), portabilitysupport.MapBundle{}, importContext)
	if err != nil {
		t.Fatalf("prepare operation-bound source: %v", err)
	}
	importContext.OperationID = "operation-b"
	err = port.ApplyImportTx(context.Background(), nil, prepared, importContext)
	if !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("cross-operation prepared value error = %v; want ErrPreparedBinding", err)
	}
	err = port.ValidateImportTx(
		context.Background(),
		nil,
		prepared,
		importContext,
	)
	if !errors.Is(err, sourceport.ErrPreparedBinding) {
		t.Fatalf("cross-operation validation error = %v; want ErrPreparedBinding", err)
	}
}

func newSourcePortCatalog(t testing.TB) (*sourceport.Catalog, sourceport.Port) {
	t.Helper()
	revisionsPort := revisionsupport.MustRuntime(t).IncidentBundleSourcePort()
	catalog, err := incidentportabilityassembly.NewCatalog(revisionsPort, referencefixture.IndicatorRegistry{}, referencefixture.IndicatorRegistry{})
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	return catalog, revisionsPort
}

func assertSourcePathIdentityFailuresAreOrderIndependentAndUnknownPathsFailClosed(t *testing.T) {
	for _, invariants := range [][]string{
		{"fixture.first", "fixture.source_identity_admitted"},
		{"fixture.source_identity_admitted", "fixture.first"},
	} {
		descriptor := sourceport.Descriptor{
			FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
			Paths:        []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "fixture.source_identity_admitted"}},
			InvariantIDs: invariants,
		}
		for name, payload := range map[string][]byte{
			"compound":   []byte("{\"id\":{\"value\":1}}\n"),
			"fraction":   []byte("{\"id\":1.25}\n"),
			"decreasing": []byte("{\"id\":2}\n{\"id\":1}\n"),
			"missing":    []byte("{\"other\":\"value\"}\n"),
			"duplicate":  []byte("{\"id\":\"same\"}\n{\"id\":\"same\"}\n"),
		} {
			t.Run(name, func(t *testing.T) {
				_, err := sourceport.PrepareMembers(context.Background(), descriptor, portabilitysupport.MapBundle{"data/fixture.ndjson": payload}, 8)
				var failure *sourceport.Failure
				if !errors.As(err, &failure) || failure.FamilyID() != "fixture" || failure.InvariantID() != "fixture.source_identity_admitted" {
					t.Fatalf("identity failure = %#v, %v", failure, err)
				}
			})
		}
	}

	descriptor := sourceport.Descriptor{
		FamilyID: "fixture", ContractMajor: sourceport.ContractMajor, OwnerID: "module.fixture",
		Paths:        []sourceport.Path{{LogicalPath: "data/fixture.ndjson", ContentRole: "source_rows", Versions: []int{8}, StableIdentity: []string{"id"}, StableIdentityInvariantID: "fixture.source_identity_admitted"}},
		InvariantIDs: []string{"fixture.source_identity_admitted"},
	}
	if _, err := sourceport.PrepareMembers(context.Background(), descriptor, portabilitysupport.MapBundle{"data/fixture.ndjson": []byte("{\"id\":2}\n{\"id\":10}\n")}, 8); err != nil {
		t.Fatalf("numeric identity order: %v", err)
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
	importContext := sourceport.ImportContext{OperationID: "operation", IncidentID: uuid.New(), ActorUserID: uuid.New(), BundleVersion: 8}
	prepared, err := port.PrepareImport(context.Background(), portabilitysupport.MapBundle{}, importContext)
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
		FamilyID         string            `json:"family_id"`
		OwnerID          string            `json:"owner_id"`
		OwnerRelationIDs []string          `json:"owner_relation_ids"`
		Dependencies     []string          `json:"dependencies"`
		Paths            []sourceport.Path `json:"paths"`
		InvariantIDs     []string          `json:"invariant_ids"`
	} `json:"families"`
	SpecialConsumers []struct {
		FamilyID     string   `json:"family_id"`
		OwnerID      string   `json:"owner_id"`
		LogicalPath  string   `json:"logical_path"`
		Versions     []int    `json:"versions"`
		SchemaID     string   `json:"schema_id"`
		InvariantIDs []string `json:"invariant_ids"`
	} `json:"special_consumers"`
}

func assertRuntimeSourceCatalogProjection(t *testing.T, runtime []sourceport.Descriptor) {
	t.Helper()
	var authored sourceCatalogProjection
	readContractJSON(t, "source_catalog.json", &authored)
	if len(runtime) != len(authored.Families) {
		t.Fatalf("runtime source families = %d, authored = %d", len(runtime), len(authored.Families))
	}
	byFamily := make(map[string]sourceport.Descriptor, len(runtime))
	for _, descriptor := range runtime {
		byFamily[descriptor.FamilyID] = descriptor
	}
	for _, family := range authored.Families {
		descriptor, present := byFamily[family.FamilyID]
		if !present {
			t.Fatalf("runtime source catalog is missing %q", family.FamilyID)
		}
		authoredRelations := append([]string(nil), family.OwnerRelationIDs...)
		authoredDependencies := append([]string(nil), family.Dependencies...)
		authoredInvariants := append([]string(nil), family.InvariantIDs...)
		authoredPaths := append([]sourceport.Path(nil), family.Paths...)
		slices.Sort(authoredRelations)
		slices.Sort(authoredDependencies)
		slices.Sort(authoredInvariants)
		for index := range authoredPaths {
			slices.Sort(authoredPaths[index].Versions)
		}
		slices.SortFunc(authoredPaths, func(left, right sourceport.Path) int {
			return strings.Compare(left.LogicalPath, right.LogicalPath)
		})
		if descriptor.ContractMajor != authored.ContractMajor ||
			descriptor.OwnerID != family.OwnerID ||
			!slices.Equal(descriptor.OwnerRelationIDs, authoredRelations) ||
			!slices.Equal(descriptor.Dependencies, authoredDependencies) ||
			!slices.Equal(descriptor.InvariantIDs, authoredInvariants) ||
			!reflect.DeepEqual(descriptor.Paths, authoredPaths) {
			t.Fatalf("source catalog projection drift for %s:\nauthored=%#v\nruntime=%#v", family.FamilyID, family, descriptor)
		}
	}
}

func assertAuthoredSourceCatalog(t *testing.T) {
	t.Helper()
	var authored sourceCatalogProjection
	readContractJSON(t, "source_catalog.json", &authored)
	if authored.SchemaID != "cartulary.incident_bundle_source_catalog.v6" || authored.ContractMajor != sourceport.ContractMajor || len(authored.Families) != 14 {
		t.Fatalf("authored source catalog header = %q major %d families %d", authored.SchemaID, authored.ContractMajor, len(authored.Families))
	}
	for _, family := range authored.Families {
		invariantID := family.FamilyID + ".source_identity_admitted"
		if !slices.Contains(family.InvariantIDs, invariantID) {
			t.Fatalf("authored family %q does not declare %q", family.FamilyID, invariantID)
		}
		for _, path := range family.Paths {
			expectedVersions := []int{8}
			if family.FamilyID == "saved_views" && path.SchemaID != "cartulary.incident_bundle.saved_views.row.v2" {
				t.Fatal("unknown saved view row schema")
			}

			if !slices.Equal(path.Versions, expectedVersions) {
				t.Fatalf("authored path %q versions = %#v, want %v", path.LogicalPath, path.Versions, expectedVersions)
			}
			if path.StableIdentityInvariantID != invariantID {
				t.Fatalf("authored path %q invariant = %q, want %q", path.LogicalPath, path.StableIdentityInvariantID, invariantID)
			}
		}
	}
	wantSpecial := map[string]struct {
		owner, path, schema string
		invariants          []string
	}{
		"actors":              {"module.incidentbundles", "data/actors.ndjson", "", []string{"actors.inert", "actors.reference_complete"}},
		"reference_pack_refs": {"module.reference_data", "data/reference_pack_refs.json", "reference_pack_refs.v3", []string{"reference_pack_refs.degradation_bounded", "reference_pack_refs.exact_shape", "reference_pack_refs.identity_exact"}},
		"extension_payload":   {"module.extensions", "ext/**", "", []string{"extension_payload.contract_compatible", "extension_payload.participant_admitted", "extension_payload.resource_bounded", "extension_payload.schema_digest_valid"}},
		"committed_sources":   {"module.imports", "sources/sha256/**", "", []string{"committed_sources.blob_exact"}},
	}
	seen := map[string]bool{}
	for _, consumer := range authored.SpecialConsumers {
		want, exists := wantSpecial[consumer.FamilyID]
		if !exists || seen[consumer.FamilyID] {
			t.Errorf("unexpected or duplicate special consumer %q", consumer.FamilyID)
			continue
		}
		seen[consumer.FamilyID] = true
		invariants := slices.Clone(consumer.InvariantIDs)
		slices.Sort(invariants)
		if consumer.OwnerID != want.owner || consumer.LogicalPath != want.path || consumer.SchemaID != want.schema || !slices.Equal(consumer.Versions, []int{8}) || !slices.Equal(invariants, want.invariants) {
			t.Errorf("special consumer %q = %#v, want owner=%q path=%q schema=%q version8 invariants=%#v", consumer.FamilyID, consumer, want.owner, want.path, want.schema, want.invariants)
		}
	}
	for family := range wantSpecial {
		if !seen[family] {
			t.Errorf("missing special consumer %q", family)
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
			"data/change_sets.ndjson":          "change_sets.row.v3.schema.json",
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
		t.Fatal("authored and runtime saved_views catalogs must each expose one current-format path")
	}
	for index := range runtime.Paths {
		authoredPath := projection.Paths[index]
		runtimePath := runtime.Paths[index]
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

		var rowSchema struct {
			rowSchemaProjection
			Definitions map[string]struct {
				Pattern string `json:"pattern"`
			} `json:"$defs"`
		}
		readContractJSON(t, "saved_views.row.v2.schema.json", &rowSchema)
		required := []string{
			"saved_view_id", "incident_id", "view_schema_id", "scope", "display_name",
			"query_json", "layout_json", "owner_user_id", "created_at", "updated_at",
			"saved_view_version",
		}
		if rowSchema.ID != runtimePath.SchemaID || rowSchema.Type != "object" ||
			rowSchema.AdditionalProperties || !slices.Equal(rowSchema.Required, required) {
			t.Fatalf("saved_views row schema is not the exact closed runtime projection: %#v", rowSchema)
		}
		clock, err := regexp.Compile(rowSchema.Definitions["timestamp"].Pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, test := range []struct {
			value string
			valid bool
		}{
			{"2026-07-02T00:00:00Z", true},
			{"2026-07-02T00:00:00.1Z", true},
			{"2026-07-02T00:00:00.000001Z", true},
			{"2026-07-02T00:00:00.999999Z", true},
			{"2026-07-02T00:00:00.1234567Z", false},
			{"2026-07-02T00:00:00.000000001Z", false},
			{"2026-07-02T00:00:00x123456Z", false},
			{`2026-07-02T00:00:00\.123456Z`, false},
		} {
			if clock.MatchString(test.value) != test.valid {
				t.Errorf("Saved Views clock projection admits %q = %t, want %t", test.value, !test.valid, test.valid)
			}
		}
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
