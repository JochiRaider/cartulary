package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func testSnapshotContentClasses(t *testing.T) {
	build := func() SnapshotModel {
		t.Helper()
		fields := []ExportField{}
		for _, class := range []string{ContentClassSourceEvidence, ContentClassDerivedAnalytic, ContentClassCuratedNarrative, ContentClassWorkingMaterial} {
			fields = append(fields, ExportField{Path: "/incident/" + class, SourceFamily: "incident_metadata",
				ContentClass: class, Value: map[string]any{"a": "first", "b": "second"},
				DisclosurePartitionRefs: []string{"public"}, SupportRefs: []string{"/incident/source"}})
		}
		model, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID,
			time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC), "boundary", fixtureSnapshotSource(fields))
		if err != nil {
			t.Fatal(err)
		}
		return model
	}
	model := build()
	want := ReportingContentClassSummary{SourceEvidence: 1, DerivedAnalytic: 1, CuratedNarrative: 1, WorkingMaterial: 1}
	if model.Sections[0].ContentClassSummary != want {
		t.Fatal("summary must count actual blocks, not source fields or an assumed class")
	}
	before, _ := canonicalJSON(model)
	for _, scope := range []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal} {
		render, digest, err := bindRenderModel(model, fixtureRenderIdentity(), scope, nil)
		if err != nil || render.Sections[0].ContentClassSummary != want {
			t.Fatal("scope changed frozen classification", err)
		}
		profile := internalRedactionProfile()
		if scope == ReleaseScopeExternal {
			profile = externalRedactionProfile(nil)
		}
		fixtureAllowContainerRules(t, &profile, render.SnapshotContent)
		profileDigest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		redacted, err := RedactExportModel(render, fixtureSourceFactsForContent(render.SnapshotContent), profile, profileDigest, digest, scope, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateRedactionResult(redacted.Model, redacted.Manifest, scope); err != nil {
			t.Fatal(err)
		}
		wantFields := 8
		if scope == ReleaseScopeExternal {
			wantFields = 6
		}
		if len(redacted.Model.Fields) != wantFields {
			t.Fatal("unexpected class filtering", scope, len(redacted.Model.Fields))
		}
		for _, field := range redacted.Model.Fields {
			if !validContentClass(field.ContentClass) || (scope == ReleaseScopeExternal && field.ContentClass == ContentClassWorkingMaterial) {
				t.Fatal("source classification was relabeled or working text was released")
			}
		}
	}
	after, _ := canonicalJSON(model)
	if !bytes.Equal(before, after) {
		t.Fatal("classification or redaction changed the immutable source")
	}
	for name, change := range map[string]func(*SnapshotModel){
		"retired section": func(m *SnapshotModel) { m.Sections[0].SchemaID = "cartulary.reporting_section.v1" },
		"retired block":   func(m *SnapshotModel) { m.Sections[0].Blocks[0].SchemaID = "cartulary.reporting_block.v1" },
		"source reclassified as presentation": func(m *SnapshotModel) {
			m.Sections[0].Blocks[0].ContentClass = ContentClassPresentationText
			m.Sections[0].ContentClassSummary, _ = summarizeReportingBlocks(m.Sections[0].Blocks)
		},
		"retired case fact":        func(m *SnapshotModel) { m.Sections[0].Blocks[0].ContentClass = "case_fact" },
		"retired derived summary":  func(m *SnapshotModel) { m.Sections[0].Blocks[0].ContentClass = "derived_summary" },
		"negative count":           func(m *SnapshotModel) { m.Sections[0].ContentClassSummary.WorkingMaterial = -1 },
		"hidden working count":     func(m *SnapshotModel) { m.Sections[0].ContentClassSummary.WorkingMaterial = 0 },
		"invented generated count": func(m *SnapshotModel) { m.Sections[0].ContentClassSummary.PresentationText = 1 },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := build()
			change(&candidate)
			if _, _, err := bindRenderModel(candidate, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil); err == nil {
				t.Fatal("malformed source classification admitted for rendering")
			}
			raw, _ := canonicalJSON(candidate)
			decoded, _, err := decodePortableObject[SnapshotModel](raw, reportingObjectDigest(SnapshotModelSchemaID, raw))
			if err == nil && validateSnapshotModelIdentity(decoded) == nil {
				t.Fatal("matching hostile digest authorized malformed source classification")
			}
		})
	}
	// The retained summary has a closed complete member set even for zero
	// categories. Missing and null values must not become implied zero counts.
	for _, mutation := range []string{"omit", "null", "unknown"} {
		var raw map[string]any
		if err := json.Unmarshal(before, &raw); err != nil {
			t.Fatal(err)
		}
		section := raw["sections"].([]any)[0].(map[string]any)
		summary := section["content_class_summary"].(map[string]any)
		switch mutation {
		case "omit":
			delete(summary, "template_boilerplate")
		case "null":
			summary["template_boilerplate"] = nil
		case "unknown":
			summary["case_fact"] = 0
		}
		encoded, _ := canonicalJSON(raw)
		if _, _, err := decodePortableObject[SnapshotModel](encoded, reportingObjectDigest(SnapshotModelSchemaID, encoded)); err == nil {
			t.Fatal("invalid retained summary accepted", mutation)
		}
	}
	// Nested blocks count independently, without adding their fields as blocks.
	counts, err := summarizeReportingBlocks([]ReportingBlock{{ContentClass: ContentClassSourceEvidence, Children: []ReportingBlock{{ContentClass: ContentClassWorkingMaterial}}}})
	if err != nil || counts != (ReportingContentClassSummary{SourceEvidence: 1, WorkingMaterial: 1}) {
		t.Fatal("nested classification not counted exactly once", err)
	}
}

func testRedactionContentClassSelectors(t *testing.T) {
	for _, tc := range []struct {
		class    string
		accepted bool
	}{
		{"source_evidence", true}, {"derived_analytic", true}, {"curated_narrative", true}, {"working_material", true},
		{"presentation_text", true}, {"support_reference", true}, {"validation", true}, {"template_boilerplate", true},
		{"case_fact", false}, {"derived_summary", false}, {"unknown", false}, {" source_evidence ", false}, {"", false},
	} {
		class := tc.class
		profile := internalRedactionProfile()
		profile.Rules = []RedactionRule{{RuleID: "selected-class", ContentClass: &class, Action: RedactionActionSpec{Type: ActionDrop}}}
		_, err := ValidateRedactionProfile(profile)
		if tc.accepted {
			if err != nil {
				t.Fatal("valid class selector rejected", class, err)
			}
		} else if !errors.Is(err, ErrInvalidRedactionProfile) {
			t.Fatal("invalid class selector admitted", class, err)
		}
	}
}
