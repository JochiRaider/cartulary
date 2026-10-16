package reporting

import (
	"bytes"
	"context"
	"testing"
)

func testTemplateFieldBindings(t *testing.T) {
	for _, path := range []string{"$.records", "$.records[*]", "$.records['one'].fields", "$.records[*].fields['body']", "$.timeline_events['one'].fields[*]", "$.sections[*]", "$.sections['a'].blocks['b'].children[*].fields['c']"} {
		if !supportedTemplateFieldBinding(path) {
			t.Fatal("supported address rejected", path)
		}
	}
	for _, path := range []string{"$", "$.source_facts", "$.records[0]", "$.records[*].fields[*]", "$.records['one'].record_type", "$.records['one'].fields['body'].value", "$.sections['a'].title", "$.sections['a'].fields['b']", "$.sections['a'].blocks['b'].blocks['c']"} {
		if supportedTemplateFieldBinding(path) {
			t.Fatal("non-field expression admitted", path)
		}
	}
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	contract.AllowedExportModelBindings = []string{"$.records[*]", "$.sections[*]", "$.timeline_events[*]"}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/notes/note", SourceFamily: "note", SourceRecordID: "note", SourceRecordType: "artifact", ContentClass: ContentClassWorkingMaterial, Value: map[string]any{"body": "selected narrative", "title": "unselected title"}},
		{Path: "/envelopes/note", SourceFamily: "record_envelope", SourceRecordID: "note", SourceRecordType: "artifact", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "artifact", "value": "unselected envelope"}},
		{Path: "/tags/note", SourceFamily: "record_tag", SourceRecordID: "note", SourceRecordType: "record_tag", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"tag_name": "unselected auxiliary"}},
		{Path: "/timeline/event", SourceFamily: "timeline_event", SourceRecordID: "event", SourceRecordType: "timeline_event", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"capture_state": "reviewed", "value": "selected event"}},
	})
	for _, test := range []struct{ binding, value string }{
		{"$.records['note'].fields['body']", "selected narrative"},
		{"$.records['event'].fields['value']", "selected event"},
		{"$.timeline_events['event'].fields['value']", "selected event"},
	} {
		contract.Sections[0].Bindings = []string{test.binding}
		if err := validateTemplateContract(contract, OutputKindSlidev, ReleaseScopeInternalReview); err != nil {
			t.Fatal(err)
		}
		result := templateDerivationResult(t, model, contract, nil)
		instances, err := deriveTemplateSections(context.Background(), contract, result.Model)
		if err != nil || len(instances) != 1 || len(instances[0].Section.Blocks) != 1 || len(instances[0].Section.Blocks[0].Fields) != 1 || instances[0].Section.Blocks[0].Fields[0].Value != test.value {
			t.Fatal("wrong field selected through stable view", test.binding, err)
		}
		selectedRef := instances[0].Section.Blocks[0].Fields[0].SourceRefs[0]
		// Selecting an alias and its canonical address must not duplicate data.
		contract.Sections[0].Bindings = append(contract.Sections[0].Bindings, result.Model.admittedFieldPaths[selectedRef])
		duplicates, err := deriveTemplateSections(context.Background(), contract, result.Model)
		if err != nil || len(duplicates) != 1 || !sameReportingValue(duplicates[0].Section, instances[0].Section) {
			t.Fatal("overlapping views duplicated a field", err)
		}
		encoded, _ := canonicalJSON(instances)
		if bytes.Contains(encoded, []byte("unselected")) {
			t.Fatal("view binding selected an envelope or auxiliary identity")
		}
		assertTemplateCandidateReplay(t, contract, model)
	}
	contract.Sections[0].Bindings = []string{"$.records['note'].record_type"}
	contract.Sections[0].Required = false
	if validateTemplateContract(contract, OutputKindSlidev, ReleaseScopeInternalReview) == nil {
		t.Fatal("optional metadata selector silently admitted")
	}
}
