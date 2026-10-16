package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func testRetainedModelBytes(t *testing.T) {
	t.Run("retained references", testRetainedModelReferences)
	source, result := retainedRecordFixture(t)
	var object map[string]json.RawMessage
	if err := json.Unmarshal(result.ModelJSON, &object); err != nil {
		t.Fatal(err)
	}
	want := []string{"schema_id", "export_model_id", "snapshot_model_id", "release_id", "preview_attempt_id", "reference_packs", "incident_id", "snapshot_id", "snapshot_at", "render_admitted_at", "source_change_set_high_watermark", "snapshot_boundary_kind", "derivation_version", "export_model_created_at", "export_model_generator_id", "export_model_generator_version", "release_scope", "recipient_partition_refs", "sections", "records", "relationships", "timeline_events", "subjects", "diagrams", "assets", "presentation_texts", "source_index", "support_index", "validation_summary"}
	if len(object) != len(want) {
		t.Fatalf("canonical retained model has %d members, want %d", len(object), len(want))
	}
	for _, key := range want {
		if _, exists := object[key]; !exists {
			t.Fatal("canonical member absent", key)
		}
	}
	for _, collection := range []string{"sections", "records", "relationships", "timeline_events", "subjects", "diagrams", "assets", "presentation_texts", "source_index", "support_index", "recipient_partition_refs"} {
		if len(object[collection]) == 0 || object[collection][0] != '[' {
			t.Fatal("collection is not an explicit array", collection)
		}
	}
	for _, forbidden := range []string{"raw first", "raw second", "raw third", "source_facts", "admittedFieldPaths", "redacted_export_model", "export_model_sha256", "export_model_path"} {
		if bytes.Contains(result.ModelJSON, []byte(forbidden)) {
			t.Fatal("private or removed content in canonical bytes", forbidden)
		}
	}
	if result.Model.SchemaID != ExportModelSchemaID || reportingObjectDigest(ExportModelSchemaID, result.ModelJSON) != result.ModelSHA256 {
		t.Fatal("retained model digest is not bound to canonical bytes")
	}
	if result.Model.SnapshotModelID != source.SnapshotModelID || result.Model.ExportModelCreatedAt != source.RenderAdmittedAt || len(result.Model.Records) != 2 || len(result.Model.Subjects) != 2 {
		t.Fatal("retained identity or subject census differs")
	}
	before, _ := canonicalJSON(source)
	index := snapshotSourceIndex(result.Model.SnapshotContent)
	originalIndex := snapshotSourceIndex(source.SnapshotContent)
	for _, ref := range result.Model.SourceIndex {
		original := originalIndex[ref.SourceRefID]
		if !sameOptionalText(ref.SourcePath, original.SourcePath) {
			t.Fatal("retained ordinal rebuilt immutable source address")
		}
	}
	for _, record := range result.Model.Records {
		if record.SourceRecordRef.SourceRefID == nil || index[*record.SourceRecordRef.SourceRefID].SourcePath != nil {
			t.Fatal("record anchor not retained")
		}
	}
	// The value used by serializers is the same complete object as the hash
	// input. A private decision-index edit cannot change its canonical bytes.
	result.Model.Fields = []RedactedField{{Path: "not canonical", Value: "not serializable case content"}}
	encoded, err := canonicalJSON(result.Model)
	if err != nil || !bytes.Equal(encoded, result.ModelJSON) {
		t.Fatal("private decision index changed canonical retained content", err)
	}
	for i := range result.Model.SourceIndex {
		if result.Model.SourceIndex[i].SourcePath != nil {
			*result.Model.SourceIndex[i].SourcePath = "changed"
			break
		}
	}
	*result.Model.Records[0].SourceRecordRef.SourceRefID = "changed"
	after, _ := canonicalJSON(source)
	if !bytes.Equal(before, after) {
		t.Fatal("retained metadata aliases admitted source")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := projectRetainedMetadata(ctx, source, &result.Model, nil, nil, RedactionProfileView{}); err != context.Canceled {
		t.Fatal("canceled metadata projection continued", err)
	}
}

func testRetainedModelReferences(t *testing.T) {
	source := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/opaque/host", SourceFamily: "host", SourceRecordID: "a", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"display_name": "secret host"}},
		{Path: "/opaque/evidence", SourceFamily: "evidence", SourceRecordID: "b", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"description": "secret evidence"}},
		{Path: "/opaque/host-envelope", SourceFamily: "record_envelope", SourceRecordType: "host", SourceRecordID: "a", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "host"}},
		{Path: "/opaque/evidence-envelope", SourceFamily: "record_envelope", SourceRecordType: "evidence", SourceRecordID: "b", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "evidence"}},
		{Path: "/opaque/link", SourceFamily: "record_link", SourceRecordID: "link", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"src_record_id": "a", "dst_record_id": "b", "link_type": "related"}},
		{Path: "/incident/summary", ContentClass: ContentClassDerivedAnalytic, Value: "Supported statement", SupportRefs: []string{"/hosts/a"}},
	})
	paths, err := redactionFieldModelPaths(source.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	sourceFields := reportingFieldSources(source.SnapshotContent)
	render := func(drop string) RedactionResult {
		t.Helper()
		profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, source.SnapshotContent)
		if err != nil {
			t.Fatal(err)
		}
		for id, binding := range sourceFields {
			key := binding.Field.FieldKey
			if key != "display_name" && key != drop {
				continue
			}
			action := ActionMask
			if key == drop {
				action = ActionDrop
			}
			path := paths[id]
			profile.Rules = append(profile.Rules, RedactionRule{RuleID: "action-" + key, Path: &path, Action: RedactionActionSpec{Type: action}})
		}
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		result, err := RedactExportModel(source, fixtureSourceFactsForContent(source.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	masked := render("")
	if len(masked.Model.Relationships) != 1 || masked.Model.Relationships[0].SrcRecordRef.DisplayRef == nil || *masked.Model.Relationships[0].SrcRecordRef.DisplayRef != DefaultMaskText || len(masked.Model.SupportIndex) != 1 {
		t.Fatal("relationship label or support closure lost selected semantics")
	}
	if bytes.Contains(masked.ModelJSON, []byte("secret host")) {
		t.Fatal("raw source display copied into metadata")
	}
	for _, key := range []string{"description", "link_type", "display_name"} {
		result := render(key)
		if len(result.Model.Relationships) != 0 {
			t.Fatal("relationship survived a missing endpoint or structural input", key)
		}
		if len(result.Model.SupportIndex) != 1 || result.Model.SupportIndex[0].SourceRefID == nil {
			t.Fatal("retained citation lost its frozen metadata anchor")
		}
		if key == "display_name" {
			for _, record := range result.Model.Records {
				if record.RecordID == "a" {
					t.Fatal("citation resurrected removed source record")
				}
			}
			for _, subject := range result.Model.Subjects {
				if subject.SourceRecordID != nil && *subject.SourceRecordID == "a" {
					t.Fatal("removed subject retained through a citation")
				}
			}
		}
	}
	for name, mutate := range map[string]func(*ExportModel){
		"missing reference": func(m *ExportModel) { m.SourceIndex = m.SourceIndex[1:] },
		"foreign snapshot":  func(m *ExportModel) { m.SourceIndex[0].SourceSnapshotID = "foreign" },
		"changed source path": func(m *ExportModel) {
			for i := range m.SourceIndex {
				if m.SourceIndex[i].SourcePath != nil {
					*m.SourceIndex[i].SourcePath = "$.sections['invented']"
					break
				}
			}
		},
		"missing support":    func(m *ExportModel) { m.SupportIndex = nil },
		"altered field copy": func(m *ExportModel) { m.Records[0].Fields[0].Value = "different" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate, err := copyReportingModel(masked.Model.ExportModel)
			if err != nil {
				t.Fatal(err)
			}
			mutate(&candidate)
			if validateRetainedProjection(candidate) == nil {
				t.Fatal("invalid retained graph admitted")
			}
		})
	}
	// Runtime asset admission reads the same retained case-content tree as
	// serialization. The private decision index cannot bypass that boundary.
	masked.Model.Fields = []RedactedField{{Value: "harmless"}}
	masked.Model.Sections[0].Blocks[0].Fields[0].Value = "![image](https://example.test/image.png)"
	if !errors.Is(validateSourceValuesForRemoteAssets(masked.Model), ErrRemoteRuntimeAsset) {
		t.Fatal("private index bypassed canonical asset admission")
	}
}
