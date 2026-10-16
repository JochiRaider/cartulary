package reporting

import (
	"encoding/json"
	"reflect"
	"slices"
	"sort"
	"testing"
)

func testStructuralRedactionRemovals(t *testing.T) {
	const section = "$.sections['section']"
	const parent = section + ".blocks['parent']"
	const child = parent + `.children['child\'quoted']`
	const field = child + ".fields['value']"
	model := SnapshotContent{Sections: []ReportingSection{{SectionID: "section", SectionKind: "appendix", Blocks: []ReportingBlock{{
		BlockID: "parent", Children: []ReportingBlock{{BlockID: "child'quoted", Fields: []ReportingField{{FieldKey: "value"}}}},
	}}}}}
	decisions := []RedactionManifestEntry{{Path: field, Outcome: "dropped"}}
	removals, err := structuralRedactionRemovals(model, decisions)
	want := []StructuralRedactionRemoval{
		{Path: section, Reason: "empty_section", CausalPaths: []string{parent}},
		{Path: parent, Reason: "empty_block", CausalPaths: []string{child}},
		{Path: child, Reason: "empty_block", CausalPaths: []string{field}},
	}
	if err != nil || !reflect.DeepEqual(removals, want) || validateStructuralRemovals(decisions, removals) != nil {
		t.Fatal("empty ancestor chain did not retain exact immediate causes", err, removals)
	}
	// A container may be allowed by its own rule, then disappear because all
	// children disappear. Both stage facts coexist without inventing a drop.
	if validateStructuralRemovals(append(slices.Clone(decisions), RedactionManifestEntry{Path: parent, Outcome: "unchanged"}), removals) != nil {
		t.Fatal("actual retaining rule conflicted with later empty-container removal")
	}
	for name, mutate := range map[string]func([]StructuralRedactionRemoval){
		"self cycle":       func(r []StructuralRedactionRemoval) { r[0].CausalPaths = []string{section} },
		"skipped ancestor": func(r []StructuralRedactionRemoval) { r[0].CausalPaths = []string{child} },
		"foreign cause": func(r []StructuralRedactionRemoval) {
			r[2].CausalPaths = []string{"$.sections['other'].blocks['b'].fields['value']"}
		},
		"unknown cause":       func(r []StructuralRedactionRemoval) { r[2].CausalPaths = []string{child + ".fields['missing']"} },
		"duplicate cause":     func(r []StructuralRedactionRemoval) { r[2].CausalPaths = []string{field, field} },
		"absent cause list":   func(r []StructuralRedactionRemoval) { r[2].CausalPaths = nil },
		"omitted known cause": func(r []StructuralRedactionRemoval) { r[2].CausalPaths = []string{} },
		"wrong kind":          func(r []StructuralRedactionRemoval) { r[0].Reason = "empty_block" },
		"unknown reason":      func(r []StructuralRedactionRemoval) { r[2].Reason = "profile_default" },
		"reordered":           func(r []StructuralRedactionRemoval) { r[0], r[1] = r[1], r[0] },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := append([]StructuralRedactionRemoval{}, removals...)
			mutate(candidate)
			if validateStructuralRemovals(decisions, candidate) == nil {
				t.Fatal("invalid causal evidence accepted")
			}
		})
	}
	decisions[0].Outcome = "unchanged"
	if validateStructuralRemovals(decisions, removals) == nil {
		t.Fatal("retained field could justify removing its ancestors")
	}
	if result, err := structuralRedactionRemovals(model, decisions); err != nil || len(result) != 0 {
		t.Fatal("retained child removed its ancestors", err)
	}
	if result, err := structuralRedactionRemovals(model, nil); err == nil || result != nil {
		t.Fatal("missing field decision produced partial structural evidence")
	}
	model.Sections[0].SectionKind = "validation"
	decisions[0].Outcome = "dropped"
	if result, err := structuralRedactionRemovals(model, decisions); err != nil || len(result) != 2 || result[0].Path == section {
		t.Fatal("empty validation section was removed", err)
	}
	// The production redaction path publishes the causal inventory under the
	// manifest digest. Removing all fields does not invent parent rule entries.
	input := fixtureStructuredRenderModel(t, []ExportField{{Path: "/incident/a", ContentClass: ContentClassSourceEvidence, Value: "private"}})
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, input.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	profile.DefaultAction = RedactionActionSpec{Type: ActionDrop}
	profileSHA, _ := ValidateRedactionProfile(profile)
	inputBytes, _ := canonicalJSON(input)
	result, err := RedactExportModel(input, fixtureSourceFactsForContent(input.SnapshotContent), profile, profileSHA, reportingObjectDigest(input.SchemaID, inputBytes), ReleaseScopeInternalReview, nil)
	if err != nil || len(result.Manifest.Entries) != 1 || len(result.Manifest.StructuralRemovals) != 2 || len(result.Model.Fields) != 0 {
		t.Fatal("production redaction did not bind structural evidence", err)
	}
	decode := func() map[string]any {
		var object map[string]any
		if err := json.Unmarshal(result.ManifestJSON, &object); err != nil {
			t.Fatal(err)
		}
		return object
	}
	keys := []string{"invented_rule"}
	for key := range decode()["structural_removals"].([]any)[0].(map[string]any) {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		candidate := decode()
		entry := candidate["structural_removals"].([]any)[0].(map[string]any)
		if key == "invented_rule" {
			entry[key] = "profile_default"
		} else {
			delete(entry, key)
		}
		raw, _ := canonicalJSON(candidate)
		if _, _, err := decodePortableObject[RedactionManifest](raw, reportingObjectDigest(RedactionManifestSchemaID, raw)); err == nil {
			t.Fatal("unclosed structural entry admitted", key)
		}
	}
}
