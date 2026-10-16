package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func testRetainedMixedContent(t *testing.T) {
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/incident/mixed", ContentClass: ContentClassSourceEvidence,
		Value: map[string]any{"first": "eligible text", "second": "restricted text"}, DisclosurePartitionRefs: []string{"public"}}})
	block := &model.Sections[0].Blocks[0]
	block.Fields[1].DisclosurePartitionRefs = []string{"blocked"}
	block.DisclosurePartitionRefs = []string{"public", "blocked"}
	model.Sections[0].DisclosurePartitionRefs = []string{"public", "blocked"}
	for i := range model.SourceIndex {
		ref := &model.SourceIndex[i]
		if ref.SourcePath == nil {
			ref.DisclosurePartitionRefs = []string{"public", "blocked"}
		} else if ref.SourceRefID == block.Fields[1].SourceRefs[0] {
			ref.DisclosurePartitionRefs = []string{"blocked"}
		}
	}
	if err := validateRenderModelIdentity(model); err != nil {
		t.Fatal("invalid authored mixed fixture", err)
	}
	before, _ := canonicalJSON(model)
	paths, err := redactionFieldModelPaths(model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	restricted := paths[block.Fields[1].SourceRefs[0]]
	public := paths[block.Fields[0].SourceRefs[0]]
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	profile.Rules = []RedactionRule{{RuleID: "omit-restricted-field", Path: &restricted, Action: RedactionActionSpec{Type: ActionDrop}}}
	render := func(profile RedactionProfile) (RedactionResult, error) {
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		return RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, reportingObjectDigest(model.SchemaID, before), ReleaseScopeInternalReview, nil)
	}
	result, err := render(profile)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Model.Sections) != 1 || len(result.Model.Sections[0].Blocks) != 1 {
		t.Fatal("field removal invented fragments or discarded the eligible block")
	}
	retained := result.Model.Sections[0].Blocks[0]
	if retained.BlockKind != "paragraph" || retained.BlockID != "sec-0001.b0001" || len(retained.Fields) != 1 ||
		retained.Fields[0].Value != "eligible text" || retained.Fields[0].FieldOrdinal != 1 ||
		!sameReportingValue(retained.DisclosurePartitionRefs, []string{"public"}) ||
		!sameReportingValue(retained.SourceRefs, block.Fields[0].SourceRefs) {
		t.Fatal("retained structure, exact value, ordinal or reference union differs")
	}
	if len(result.Manifest.Entries) != 2 || result.Manifest.Entries[0].Path != public || result.Manifest.Entries[1].Path != restricted ||
		result.Manifest.Entries[1].Outcome != "dropped" || len(result.Manifest.StructuralRemovals) != 0 {
		t.Fatal("structural retention changed admitted rule addresses or invented removals")
	}
	if bytes.Contains(result.ModelJSON, []byte("restricted text")) || bytes.Contains(result.ModelJSON, []byte("split_from_block_id")) {
		t.Fatal("retired or restricted content survived")
	}
	for _, action := range []RedactionActionSpec{{Type: ActionAllow}, {Type: ActionMask}, {Type: ActionStub}} {
		profile.Rules[0].Action = action
		partial, err := render(profile)
		var failure *renderValidationError
		if !errors.As(err, &failure) || failure.ReasonCode != "disclosure_partition_unresolved" || len(partial.ModelJSON) != 0 {
			t.Fatal("selected action silently declassified or split restricted text", action.Type, err)
		}
	}
	after, _ := canonicalJSON(model)
	if !bytes.Equal(before, after) {
		t.Fatal("redaction mutated the immutable admitted structure")
	}

	// A recomputed digest must not rescue a retired member, including null.
	for _, retired := range []struct {
		key   string
		value any
	}{
		{"split_from_block_id", nil}, {"split_from_block_id", "sec-0001.b0001"},
		{"aggregate_only_non_identifying", false}, {"aggregate_only_non_identifying", true},
		{"aggregate_policy_id", nil}, {"aggregate_policy_id", "aggregate_public_v1"},
		{"contributor_count", nil}, {"contributor_count", 1000},
		{"excluded_field_keys", []string{}}, {"excluded_field_keys", []string{"secret"}},
	} {
		var object map[string]any
		if err := json.Unmarshal(before, &object); err != nil {
			t.Fatal(err)
		}
		item := object["sections"].([]any)[0].(map[string]any)["blocks"].([]any)[0].(map[string]any)
		item[retired.key] = retired.value
		raw, _ := canonicalJSON(object)
		if _, _, err := decodePortableObject[ExportModel](raw, reportingObjectDigest(ExportModelSchemaID, raw)); err == nil {
			t.Fatal("retired disclosure member admitted", retired.key, retired.value)
		}
	}
}
