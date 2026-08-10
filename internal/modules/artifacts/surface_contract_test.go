package artifacts

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/surfacecatalog"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestArtifactSurfaceContractMatrix(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		CommLogViewSchemaID:              "comm_log",
		FindingsViewSchemaID:             "finding",
		ForensicKeywordsViewSchemaID:     "forensic_keyword",
		HandoffViewSchemaID:              "handoff",
		InvestigativeQueriesViewSchemaID: "investigative_query",
		LessonViewSchemaID:               "lesson",
		NotesViewSchemaID:                "note",
		StatusReviewViewSchemaID:         "status_review",
	}
	if len(want) != 8 {
		t.Fatalf("artifact surface fixture has %d entries, want 8", len(want))
	}
	catalog := surfacecatalog.All()
	if len(catalog) != len(want) {
		t.Fatalf("surface catalog has %d entries, want %d", len(catalog), len(want))
	}
	for _, surface := range catalog {
		if want[surface.ViewSchemaID] != surface.ArtifactType {
			t.Fatalf("surface catalog entry %#v does not match owner expectation", surface)
		}
		reverse, ok := surfacecatalog.LookupByArtifactType(surface.ArtifactType)
		if !ok || reverse != surface {
			t.Fatalf("reverse surface lookup for %q = %#v, %v", surface.ArtifactType, reverse, ok)
		}
	}
	if _, ok := surfacecatalog.LookupByArtifactType("future_unregistered_artifact"); ok {
		t.Fatal("unregistered artifact type was admitted")
	}

	for viewSchemaID, artifactType := range want {
		viewSchemaID, artifactType := viewSchemaID, artifactType
		t.Run(viewSchemaID, func(t *testing.T) {
			t.Parallel()
			if got := ArtifactTypeForView(viewSchemaID); got != artifactType {
				t.Fatalf("ArtifactTypeForView(%q) = %q, want %q", viewSchemaID, got, artifactType)
			}
			if !IsArtifactBackedView(viewSchemaID) {
				t.Fatalf("IsArtifactBackedView(%q) = false", viewSchemaID)
			}
			schema, ok := viewschema.Lookup(viewSchemaID)
			if !ok {
				t.Fatalf("viewschema.Lookup(%q) not found", viewSchemaID)
			}
			if schema.BaseProjection != "artifact_grid_projection" {
				t.Fatalf("base projection = %q, want artifact_grid_projection", schema.BaseProjection)
			}
			filter, ok := schema.CanonicalSourceFilter()
			if !ok || filter.Kind != "artifact_type" || filter.Field != "artifact_type" || filter.Value != artifactType {
				t.Fatalf("canonical filter = %#v, %v; want artifact_type=%q", filter, ok, artifactType)
			}
			for fieldKey, field := range schema.Fields() {
				policy, mapped := lookupArtifactSourceField(fieldKey)
				if !field.Writable && !field.CreateWritable {
					if mapped {
						t.Fatalf("read-only field %s was admitted to the artifact source policy", fieldKey)
					}
					continue
				}
				if !mapped || policy.viewSchemaID != viewSchemaID || !policy.writable {
					t.Fatalf("writable field %s has incomplete source policy %#v, %v", fieldKey, policy, mapped)
				}
				switch field.WriteKind {
				case "direct_value":
					if policy.kind != sourceFieldDirect || policy.storage.table == "" || policy.storage.column == "" {
						t.Fatalf("direct field %s has invalid source policy %#v", fieldKey, policy)
					}
				case "action_payload":
					if policy.kind != sourceFieldCollection || policy.collection.FieldKey != fieldKey {
						t.Fatalf("collection field %s has invalid source policy %#v", fieldKey, policy)
					}
				default:
					t.Fatalf("writable field %s has unsupported authored write kind %q", fieldKey, field.WriteKind)
				}
			}
		})
	}

	for _, viewSchemaID := range []string{"", "cartulary.view.timeline.v2", "cartulary.view.unknown.v1"} {
		if got := ArtifactTypeForView(viewSchemaID); got != "" {
			t.Fatalf("ArtifactTypeForView(%q) = %q, want empty", viewSchemaID, got)
		}
		if IsArtifactBackedView(viewSchemaID) {
			t.Fatalf("IsArtifactBackedView(%q) = true", viewSchemaID)
		}
	}

	t.Run("incident bundle descriptor is exact", func(t *testing.T) {
		descriptor := NewIncidentBundleSourcePort().Descriptor()
		if descriptor.FamilyID != "artifacts" || descriptor.ContractMajor != 1 || descriptor.OwnerID != "module.artifacts" {
			t.Fatalf("artifact incident-bundle descriptor identity = %#v", descriptor)
		}
		if !slices.Equal(descriptor.OwnerRelationIDs, []string{"artifacts-and-optional-surfaces"}) ||
			!slices.Equal(descriptor.Dependencies, []string{"indicators"}) {
			t.Fatalf("artifact incident-bundle descriptor ownership/dependencies = %#v", descriptor)
		}
		wantPaths := []struct {
			path     string
			identity string
		}{
			{"data/artifacts.ndjson", "record_id"},
			{"data/artifact_findings.ndjson", "record_id"},
			{"data/artifact_investigative_queries.ndjson", "record_id"},
			{"data/artifact_forensic_keywords.ndjson", "record_id"},
			{"data/handoff_risk_refs.ndjson", "risk_ref_id"},
		}
		if len(descriptor.Paths) != len(wantPaths) {
			t.Fatalf("artifact incident-bundle paths = %d, want %d", len(descriptor.Paths), len(wantPaths))
		}
		for index, wantPath := range wantPaths {
			got := descriptor.Paths[index]
			if got.LogicalPath != wantPath.path || got.ContentRole != "source_rows" ||
				!slices.Equal(got.Versions, []int{1, 2}) || !slices.Equal(got.StableIdentity, []string{wantPath.identity}) {
				t.Fatalf("artifact incident-bundle path %d = %#v, want %s/%s", index, got, wantPath.path, wantPath.identity)
			}
		}
		wantInvariants := []string{
			"artifacts.envelope_type_scope",
			"artifacts.subtype_exact",
			"artifacts.lifecycle_fields_legal",
			"artifacts.handoff_risk_target",
			"artifacts.references_same_incident",
		}
		if !slices.Equal(descriptor.InvariantIDs, wantInvariants) {
			t.Fatalf("artifact incident-bundle invariants = %#v, want %#v", descriptor.InvariantIDs, wantInvariants)
		}
	})

	t.Run("recovery contribution is exact", func(t *testing.T) {
		contribution := RecoveryStateContribution()
		if contribution.OwnerID != "module.artifacts" || len(contribution.ObjectFamilies) != 0 {
			t.Fatalf("artifact recovery contribution identity = %#v", contribution)
		}
		tables := make([]string, 0, len(contribution.Tables))
		for _, table := range contribution.Tables {
			tables = append(tables, table.TableName)
		}
		slices.Sort(tables)
		wantTables := []string{
			"artifact_findings",
			"artifact_forensic_keywords",
			"artifact_investigative_queries",
			"artifacts",
			"handoff_risk_refs",
		}
		if !slices.Equal(tables, wantTables) {
			t.Fatalf("artifact recovery tables = %#v, want %#v", tables, wantTables)
		}
	})

	t.Run("exact_write_admission", func(t *testing.T) {
		body := "Valid note signal"
		number := int64(1)
		for name, params := range map[string]CreateParams{
			"unknown_view": {
				ViewSchemaID: "cartulary.view.unknown.v1",
				Values:       map[string]FieldValue{"note.body": {Text: &body}},
			},
			"unknown_prefixed_field": {
				ViewSchemaID: NotesViewSchemaID,
				Values: map[string]FieldValue{
					"note.body":       {Text: &body},
					"note.unreviewed": {Text: &body},
				},
			},
			"read_only_field": {
				ViewSchemaID: NotesViewSchemaID,
				Values: map[string]FieldValue{
					"note.body":       {Text: &body},
					"note.updated_at": {Text: &body},
				},
			},
			"cross_surface_field": {
				ViewSchemaID: NotesViewSchemaID,
				Values: map[string]FieldValue{
					"note.body":        {Text: &body},
					"comm_log.summary": {Text: &body},
				},
			},
			"collection_as_scalar": {
				ViewSchemaID: NotesViewSchemaID,
				Values: map[string]FieldValue{
					"note.body": {Text: &body},
					"note.tags": {Text: &body},
				},
			},
			"wrong_value_kind": {
				ViewSchemaID: NotesViewSchemaID,
				Values: map[string]FieldValue{
					"note.body":  {Text: &body},
					"note.title": {Number: &number},
				},
			},
			"non_nullable_null": {
				ViewSchemaID: NotesViewSchemaID,
				Values: map[string]FieldValue{
					"note.body":  {Text: &body},
					"note.title": {},
				},
			},
		} {
			if err := ValidateCreateParams(params); err == nil {
				t.Fatalf("%s unexpectedly passed exact source admission", name)
			}
		}
		if table, column := tableColumnForField("finding.unreviewed"); table != "" || column != "" {
			t.Fatalf("unknown prefixed field mapped to %s.%s", table, column)
		}
	})

	contribution := RevisionProviderContribution()
	if len(contribution.Records) != 1 {
		t.Fatalf("revision record contributions = %d, want 1", len(contribution.Records))
	}
	routes := contribution.Records[0].RecordViewRoutes
	if len(routes) != len(want) {
		t.Fatalf("revision view routes = %d, want %d", len(routes), len(want))
	}
	seen := make([]string, 0, len(routes))
	for _, route := range routes {
		if len(route.ViewSchemaIDs) != 1 || route.Variant == nil {
			t.Fatalf("invalid artifact revision route: %#v", route)
		}
		viewSchemaID := route.ViewSchemaIDs[0]
		if slices.Contains(seen, viewSchemaID) {
			t.Fatalf("duplicate artifact revision route for %s", viewSchemaID)
		}
		seen = append(seen, viewSchemaID)
		if route.Variant.Kind != "artifact_type" || route.Variant.Value != want[viewSchemaID] {
			t.Fatalf("revision route %s variant = %#v, want artifact_type=%q", viewSchemaID, route.Variant, want[viewSchemaID])
		}
	}
}
