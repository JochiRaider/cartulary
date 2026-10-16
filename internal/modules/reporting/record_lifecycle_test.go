package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func testSnapshotRecordLifecycle(t *testing.T) {
	build := func(family string, value map[string]any) (SnapshotModel, error) {
		source := fixtureSnapshotSource(fixtureOwnerFields([]ExportField{{Path: "/opaque/source", SourceFamily: family, SourceRecordID: "record", Value: value, ContentClass: ContentClassDerivedAnalytic}}))
		model, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", source)
		return model, err
	}
	for _, tc := range []struct {
		family, state string
		values        map[string]any
	}{
		{"timeline_event", "active", map[string]any{"capture_state": "rough", "value": "text", "nullable": nil}},
		{"timeline_event", "active", map[string]any{"capture_state": "enriched"}},
		{"timeline_event", "active", map[string]any{"capture_state": "reviewed"}},
		{"timeline_event", "superseded", map[string]any{"capture_state": "superseded", "value": "text", "nullable": nil}},
		{"decision", "superseded", map[string]any{"status": "executed", "is_superseded": true}},
		{"decision", "active", map[string]any{"status": "approved", "is_superseded": false}},
	} {
		model, err := build(tc.family, tc.values)
		if err != nil {
			t.Fatal(err)
		}
		if len(model.Records) != 1 || model.Records[0].DeletedState != tc.state {
			t.Fatal("owner lifecycle lost")
		}
		for _, field := range model.Records[0].Fields {
			want := "present"
			if tc.state == "superseded" {
				want = "superseded"
			}
			if field.Value == nil {
				want = "null"
			}
			if field.SourceValueState != want {
				t.Fatal("lifecycle corrupted nullable source state")
			}
		}
		before, _ := canonicalJSON(model)
		if err := validateSnapshotModelIdentity(model); err != nil {
			t.Fatal(err)
		}
		after, _ := canonicalJSON(model)
		if !bytes.Equal(before, after) {
			t.Fatal("admission repaired frozen lifecycle")
		}
		model.Records[0].DeletedState = "deleted"
		if validateSnapshotModelIdentity(model) == nil {
			t.Fatal("altered lifecycle admitted")
		}
	}
	for _, tc := range []struct {
		family string
		values map[string]any
	}{
		{"timeline_event", map[string]any{"value": "missing state"}},
		{"timeline_event", map[string]any{"capture_state": nil}},
		{"timeline_event", map[string]any{"capture_state": "future"}},
		{"decision", map[string]any{"status": "superseded"}},
		{"decision", map[string]any{"is_superseded": "true"}},
	} {
		if _, err := build(tc.family, tc.values); err == nil {
			t.Fatal("invalid owner state was inferred")
		}
	}
	model, err := build("timeline_event", map[string]any{"capture_state": "superseded", "value": "text"})
	if err != nil {
		t.Fatal(err)
	}
	model.Sections[0].Blocks[0].Fields[0].SourceValueState = "present"
	if validateSnapshotModelIdentity(model) == nil {
		t.Fatal("inconsistent superseded field admitted")
	}
}

func testReleaseLifecycleEligibility(t *testing.T) {
	fields := fixtureOwnerFields([]ExportField{
		{Path: "/incident/title", ContentClass: ContentClassDerivedAnalytic, Value: "kept"},
		{Path: "/timeline/current", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"capture_state": "reviewed", "value": "current"}},
		{Path: "/timeline/old", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"capture_state": "superseded", "value": "OLD-CONTENT", "nullable": nil}},
		{Path: "/record_envelopes/old", SourceFamily: "record_envelope", SourceRecordID: "old", SourceRecordType: "timeline_event", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "timeline_event", "row_version": 7}},
		{Path: "/tags/dependent", SourceFamily: "record_tag", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_id": "old", "tag_name": "OLD-TAG"}},
		{Path: "/tags/deleted", SourceFamily: "record_tag", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_id": "gone", "tag_name": "DELETED-TAG"}},
		{Path: "/tags/old", SourceFamily: "record_tag", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_id": "current", "tag_name": "KEPT-AUXILIARY"}},
		{Path: "/incident/cited", ContentClass: ContentClassDerivedAnalytic, Value: "OLD-CITATION", SupportRefs: []string{"/timeline/old"}},
		{Path: "/evidence/chain-a", ContentClass: ContentClassDerivedAnalytic, Value: "OLD-CHAIN-A", SupportRefs: []string{"/timeline/old", "/evidence/chain-b"}},
		{Path: "/evidence/chain-b", ContentClass: ContentClassDerivedAnalytic, Value: "OLD-CHAIN-B", SupportRefs: []string{"/evidence/chain-a"}},
		{Path: "/incident/chained", ContentClass: ContentClassDerivedAnalytic, Value: "OLD-CHAIN-USE", SupportRefs: []string{"/evidence/chain-b"}},
	})
	for i := range fields {
		fields[i].DisclosurePartitionRefs = []string{"public"}
	}
	source := fixtureSnapshotSource(fields)
	source.SourceFacts.DeletedRecords = []exportprovider.DeletedRecordFact{{RecordID: "gone", RecordType: "evidence"}}
	snapshot, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", source)
	if err != nil {
		t.Fatal(err)
	}
	model, _, err := bindRenderModel(snapshot, fixtureRenderIdentity(), ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := canonicalJSON(model)
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	profile := internalRedactionProfile()
	profile.AllowedReleaseScopes = append(profile.AllowedReleaseScopes, ReleaseScopeExternal)
	run := func(scope string, contract *TemplateContract, profile RedactionProfile) (RedactionResult, error) {
		digest, err := ValidateRedactionProfile(profile)
		if err != nil {
			t.Fatal(err)
		}
		return redactExportModel(context.Background(), model, snapshot.SourceFacts, profile, digest, "", scope, nil, renderSelectionInput{template: contract})
	}
	for _, scope := range []string{ReleaseScopeInternalDraft, ReleaseScopeInternalReview, ReleaseScopeExternal} {
		result, err := run(scope, &contract, profile)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(result.ModelJSON, []byte("DELETED-TAG")) {
			t.Fatal("deleted dependency released")
		}
		if !bytes.Contains(result.ModelJSON, []byte("KEPT-AUXILIARY")) {
			t.Fatal("auxiliary ID conflated with Core ID")
		}
		for _, marker := range []string{"OLD-CONTENT", "OLD-TAG", "OLD-CITATION", "OLD-CHAIN-A", "OLD-CHAIN-B", "OLD-CHAIN-USE"} {
			if bytes.Contains(result.ModelJSON, []byte(marker)) != (scope != ReleaseScopeExternal) {
				t.Fatal("lifecycle or reference restriction bypassed", scope, marker)
			}
		}
		if scope == ReleaseScopeExternal {
			for _, record := range result.Model.Records {
				if record.RecordID == "old" {
					t.Fatal("envelope resurrected superseded record")
				}
			}
			for _, entry := range result.Manifest.Entries {
				if entry.Outcome == "dropped" {
					t.Fatal("eligibility fabricated Core drop action")
				}
			}
			if len(result.Manifest.StructuralRemovals) != 0 {
				t.Fatal("eligibility fabricated redaction structural causes")
			}
		}
	}
	for _, permissions := range []struct{ profile, section bool }{{false, true}, {true, false}, {true, true}} {
		candidate := contract
		candidate.Sections = append([]TemplateSection{}, contract.Sections...)
		candidate.Sections[0].IncludeSuperseded = permissions.section
		profile.IncludeSuperseded = permissions.profile
		result, err := run(ReleaseScopeExternal, &candidate, profile)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(result.ModelJSON, []byte("OLD-CONTENT")) != (permissions.profile && permissions.section) {
			t.Fatal("missing either permission granted release")
		}
	}
	profile.IncludeSuperseded = true
	for _, binding := range []string{"$.records['old'].fields['value']", "$.timeline_events['old'].fields['value']"} {
		candidate := contract
		candidate.Sections = append([]TemplateSection{}, contract.Sections...)
		candidate.Sections[0].Bindings, candidate.Sections[0].IncludeSuperseded = []string{binding}, true
		result, err := run(ReleaseScopeExternal, &candidate, profile)
		if err != nil {
			t.Fatal(err)
		}
		instances, err := deriveTemplateSections(context.Background(), candidate, result.Model)
		if err != nil || len(instances) != 1 || len(instances[0].Section.Blocks) != 1 || instances[0].Section.Blocks[0].Fields[0].Value != "OLD-CONTENT" {
			t.Fatal("lifecycle did not recognize a canonical view binding", binding, err)
		}
		candidate.Sections[0].IncludeSuperseded = false
		result, err = run(ReleaseScopeExternal, &candidate, profile)
		if err != nil {
			t.Fatal(err)
		}
		_, err = deriveTemplateSections(context.Background(), candidate, result.Model)
		assertTemplateDerivationReason(t, err, "required_section_empty")
	}
	second := contract.Sections[0]
	second.DeclID, second.IncludeSuperseded = "opted-in", true
	contract.Sections = append(contract.Sections, second)
	result, err := run(ReleaseScopeExternal, &contract, profile)
	if err != nil {
		t.Fatal(err)
	}
	instances, err := deriveTemplateSections(context.Background(), contract, result.Model)
	if err != nil || len(instances) != 2 {
		t.Fatal("section selection failed", err)
	}
	for i, section := range instances {
		raw, _ := canonicalJSON(section.Section)
		for _, marker := range []string{"OLD-CONTENT", "OLD-TAG", "OLD-CITATION", "OLD-CHAIN-A", "OLD-CHAIN-B", "OLD-CHAIN-USE"} {
			if bytes.Contains(raw, []byte(marker)) != (i == 1) {
				t.Fatal("opt-in crossed section boundary", marker)
			}
		}
	}
	// Derived diagrams and authored insertion cannot use a different section's
	// permission. Check the final source-bound tree after composition as well.
	decl := fixtureTimelineDeclaration("flowchart", []string{"old"})
	decl.LabelOverrides = []diagramLabelOverride{}
	for i := range contract.Sections {
		candidate := contract
		candidate.Sections = append([]TemplateSection{}, contract.Sections...)
		candidate.Sections[i].DiagramDecls = []TemplateDiagramDeclaration{{DeclID: decl.DeclID, DiagramKind: decl.DiagramKind, DiagramSourceKind: decl.DiagramSourceKind, SelectionRule: decl.SelectionRule}}
		_, err := deriveDiagramModels(candidate, result.Model, nil, ReleaseScopeExternal, OutputKindSlidev, nil)
		if (err == nil) != (i == 1) {
			t.Fatal("diagram borrowed another section's opt-in", err)
		}
	}
	for i, section := range contract.Sections {
		composition := renderComposition{SchemaID: "cartulary.report_composition.v1", DiagramDecls: []compositionDiagramDecl{decl}, AuthoredTexts: []compositionAuthoredText{}, DeckOps: []compositionDeckOp{{OpID: "insert", OpKind: "insert_diagram_slide", OnUnresolved: "fail", Payload: compositionOpPayload{Position: "after", SectionAnchor: &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: section.DeclID}, DiagramAnchor: &compositionDiagramAnchor{AnchorKind: "diagram_anchor", DeclarationID: decl.DeclID}}}}}
		raw, _ := json.Marshal(composition)
		wire := sealedReportingCompositionFixture(t, raw, model.IncidentID, contract)
		_, err := deriveDiagramModels(contract, result.Model, wire, ReleaseScopeExternal, OutputKindSlidev, nil)
		if (err == nil) != (i == 1) {
			t.Fatal("composition diagram borrowed another section's opt-in", err)
		}
	}
	for i := range instances {
		slide := reportingDeckSlide{Blocks: instances[1].Section.Blocks, includeSuperseded: i == 1}
		if (validateSlideLifecycle(slide, result.Model) == nil) != (i == 1) {
			t.Fatal("post-operation content escaped section lifecycle")
		}
	}
	// Exact rules cannot grant eligibility to an excluded candidate.
	paths, _ := redactionFieldModelPaths(model.SnapshotContent)
	for _, record := range model.Records {
		if record.RecordID != "old" {
			continue
		}
		path := paths[record.Fields[0].SourceRefs[0]]
		denied := profile
		denied.IncludeSuperseded = false
		denied.Rules = []RedactionRule{{RuleID: "bypass", Path: &path, Action: RedactionActionSpec{Type: ActionAllow}}}
		if deniedResult, err := run(ReleaseScopeExternal, &contract, denied); err != nil || bytes.Contains(deniedResult.ModelJSON, []byte("OLD-CONTENT")) {
			t.Fatal("valid inactive rule blocked eligibility or restored excluded content", err)
		}
	}
	builtin, _, err := ResolveRedactionProfile(ExternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	if len(builtin.Rules) < 3 {
		t.Fatal("fixture lacks admitted nullable-field rule")
	}
	if deniedResult, err := run(ReleaseScopeExternal, &contract, builtin); err != nil || bytes.Contains(deniedResult.ModelJSON, []byte("OLD-CONTENT")) {
		t.Fatal("nullable profile rule blocked lifecycle exclusion", err)
	}
	for _, binding := range []string{"$.records['gone']", "$.timeline_events['gone'].fields"} {
		candidate := contract
		candidate.Sections = []TemplateSection{{Bindings: []string{binding}, Required: true}}
		_, err := run(ReleaseScopeExternal, &candidate, profile)
		var failure *renderValidationError
		if !errors.As(err, &failure) || failure.ReasonCode != "deleted_record_not_releasable" {
			t.Fatal("forced deleted inclusion misclassified", err)
		}
	}
	deletedDiagram := decl
	deletedDiagram.SelectionRule.TimelineRecordIDs = []string{"gone"}
	if err := validateDeletedRecordBindings(snapshot.SourceFacts, renderSelectionInput{composition: &renderComposition{DiagramDecls: []compositionDiagramDecl{deletedDiagram}}}); err == nil {
		t.Fatal("diagram forced a deleted record")
	}
	after, _ := canonicalJSON(model)
	if !bytes.Equal(before, after) {
		t.Fatal("eligibility mutated frozen admission")
	}
}
