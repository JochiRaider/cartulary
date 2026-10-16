package reporting

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func retainedRecordFixture(t *testing.T) (ExportModel, RedactionResult) {
	t.Helper()
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/opaque/a", SourceFamily: "host", SourceRecordID: "one", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"display_name": "raw first", "row_version": 1}},
		{Path: "/opaque/b", SourceFamily: "host", SourceRecordID: "two", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"display_name": "raw second"}},
		{Path: "/opaque/c", SourceFamily: "host", SourceRecordID: "three", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"display_name": "raw third", "row_version": 2}},
		{Path: "/opaque/d", SourceFamily: "record_tag", SourceRecordID: "one", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"tag_name": "tag shares Core ID"}},
		{Path: "/opaque/e", SourceFamily: "record_envelope", SourceRecordID: "two", SourceRecordType: "host", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "host"}},
	})
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	paths, err := redactionFieldModelPaths(model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range model.Records {
		for _, field := range record.Fields {
			if field.FieldKey != "display_name" {
				continue
			}
			path := paths[field.SourceRefs[0]]
			action := ActionDrop
			if record.RecordID == "one" {
				action = ActionMask
			}
			profile.Rules = append(profile.Rules, RedactionRule{RuleID: "selected-" + record.RecordID, Path: &path, Action: RedactionActionSpec{Type: action}})
		}
	}
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	return model, result
}

func testRetainedRecords(t *testing.T) {
	model, result := retainedRecordFixture(t)
	before, _ := canonicalJSON(model)
	if len(result.Model.Records) != 2 {
		t.Fatal("removed Core record was resurrected by its envelope")
	}
	for _, record := range result.Model.Records {
		switch record.RecordID {
		case "one":
			if record.DisplayName == nil || *record.DisplayName != DefaultMaskText || len(record.Fields) != 2 {
				t.Fatal("raw display survived masking")
			}
		case "three":
			if record.DisplayName != nil || len(record.Fields) != 1 || record.Fields[0].FieldOrdinal != 1 || len(record.SourceRefs) != 1 {
				t.Fatal("removed display survived projection or ordinals were not rebuilt")
			}
		default:
			t.Fatal("unexpected Core record", record.RecordID)
		}
	}
	encoded, _ := canonicalJSON(result.Model.Records)
	for _, raw := range []string{"raw first", "raw second", "raw third", "tag shares Core ID"} {
		if bytes.Contains(encoded, []byte(raw)) {
			t.Fatal("raw or auxiliary value entered retained Core summaries")
		}
	}
	result.Model.Records[0].Fields[0].Value = "modified"
	after, _ := canonicalJSON(model)
	if !bytes.Equal(before, after) {
		t.Fatal("retained summary mutates admitted snapshot")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if records, err := projectRetainedRecords(ctx, model.Records, nil); records != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("canceled projection published records", err)
	}
}
