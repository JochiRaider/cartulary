package reporting

import (
	"errors"
	"testing"
)

func testExactExportModelPaths(t *testing.T) {
	t.Run("bound selectors cannot silently disappear", testBoundRedactionPaths)
	for _, path := range []string{"$", "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']", "$.records['é'].fields['display_name']", "$.records['O\\'Brien']", "$.a_b2['with space']"} {
		if !validExactExportModelPath(path) {
			t.Fatalf("valid exact path rejected: %q", path)
		}
	}
	for _, path := range []string{"", "/incident/title", "$.records[0]", "$.records[*]", "$.records['*']..value", "$.**", "$.a[?(@.id)]", "$.a['unfinished", "$.a['id'", "$.a['id']junk", "$['a/b']", "$['a#b']", "$['a\\nb']", "$['a\\\\b']", "$['']", "$.a-b", "$..a", "$.a.", "$.a[' id']", "$.a['bad\nline']", "$.a['" + string([]byte{0xff}) + "']"} {
		if validExactExportModelPath(path) {
			t.Fatalf("invalid exact path admitted: %q", path)
		}
		profile := internalRedactionProfile()
		profile.Rules = []RedactionRule{{RuleID: "invalid-path", Path: &path, Action: RedactionActionSpec{Type: ActionAllow}}}
		if _, err := ValidateRedactionProfile(profile); err == nil {
			t.Fatal("invalid authored selector admitted")
		}
	}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/incident/a", SourceFamily: "incident_metadata", ContentClass: ContentClassDerivedAnalytic, Value: "first"},
		{Path: "/incident/z", SourceFamily: "incident_metadata", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"left": "second", "right": "third"}},
	})
	paths, err := redactionFieldModelPaths(model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	for source, expected := range map[string]string{
		model.Sections[0].Blocks[0].Fields[0].SourceRefs[0]: "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']",
		model.Sections[0].Blocks[1].Fields[0].SourceRefs[0]: "$.sections['sec-0001'].blocks['sec-0001.b0002'].fields['left']",
		model.Sections[0].Blocks[1].Fields[1].SourceRefs[0]: "$.sections['sec-0001'].blocks['sec-0001.b0002'].fields['right']",
	} {
		if paths[source] != expected {
			t.Fatalf("source/model address conflation: %s", source)
		}
	}
	// Nested objects keep explicit structural ancestry and escaped stable IDs.
	child := model.Sections[0].Blocks[1]
	child.BlockID = "child'quote"
	model.Sections[0].Blocks = model.Sections[0].Blocks[:1]
	model.Sections[0].Blocks[0].Children = []ReportingBlock{child}
	paths, err = redactionFieldModelPaths(model.SnapshotContent)
	if err != nil || paths[child.Fields[1].SourceRefs[0]] != "$.sections['sec-0001'].blocks['sec-0001.b0001'].children['child\\'quote'].fields['right']" {
		t.Fatal("nested field address lost", err)
	}
	// A source binding or model address may not silently overwrite a candidate.
	model.Sections = append(model.Sections, model.Sections[0])
	if _, err := redactionFieldModelPaths(model.SnapshotContent); err == nil {
		t.Fatal("duplicate candidate identity accepted")
	}
}

func testBoundRedactionPaths(t *testing.T) {
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/incident/a", SourceFamily: "incident_metadata", ContentClass: ContentClassDerivedAnalytic, Value: "retained"}})
	paths, err := redactionFieldModelPaths(model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	validPath := paths[model.Sections[0].Blocks[0].Fields[0].SourceRefs[0]]
	for _, path := range []string{validPath, "$", "$.sections['sec-0001']", "$.sections['sec-0001'].blocks['sec-0001.b0001']", "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['missing']", "$.records['absent'].fields['value']"} {
		t.Run(path, func(t *testing.T) {
			profile := internalRedactionProfile()
			profile.Rules = []RedactionRule{{RuleID: "required-drop", Path: &path, Action: RedactionActionSpec{Type: ActionDrop}}}
			digest, err := ValidateRedactionProfile(profile)
			if err != nil {
				t.Fatal("shape-valid profile rejected before binding", err)
			}
			payload := renderInput{ExportModel: model, RedactionProfile: profile, RedactionProfileSHA256: digest, RedactionProfileID: profile.ProfileID, RedactionProfileVersion: profile.Version, ReleaseScope: ReleaseScopeInternalReview}
			bindingErr := validateFrozenRenderProfile(payload)
			result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil)
			if path == validPath {
				if err != nil || bindingErr != nil || len(result.Model.Fields) != 0 || len(result.Manifest.Entries) != 1 || result.Manifest.Entries[0].Outcome != "dropped" {
					t.Fatal("bound exact drop lost", err, bindingErr)
				}
				return
			}
			if path == "$.sections['sec-0001']" || path == "$.sections['sec-0001'].blocks['sec-0001.b0001']" {
				if err != nil || bindingErr != nil || len(result.Model.Fields) != 0 || len(result.Manifest.Entries) != 0 {
					t.Fatal("bound container drop lost", err, bindingErr)
				}
				found := false
				for _, decision := range result.Manifest.ContainerDecisions {
					if decision.Path == path {
						found = decision.Action == ActionDrop && decision.Outcome == "dropped" && len(decision.SuppressedPaths) > 0
					}
				}
				if !found {
					t.Fatal("missing selected container drop")
				}
				return
			}
			var invalid *renderValidationError
			if !errors.As(err, &invalid) || invalid.ReasonCode != "redaction_action_unresolved" || bindingErr == nil || len(result.ManifestJSON) != 0 || result.ModelSHA256 != "" || len(result.Model.Fields) != 0 {
				t.Fatal("unsupported or unmatched drop silently ignored", err, bindingErr)
			}
		})
	}
}
