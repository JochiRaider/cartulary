package reporting

import (
	"bytes"
	"context"
	"testing"
)

func testTemplateReachability(t *testing.T) {
	t.Run("relationship and support closure", testTemplateRelationshipReachability)
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/notes/a", SourceFamily: "note", SourceRecordID: "a", SourceRecordType: "artifact", ContentClass: ContentClassWorkingMaterial, DisclosurePartitionRefs: []string{"public"}, Value: map[string]any{"body": "selected", "title": "PRIVATE-UNSELECTED-TITLE"}},
		{Path: "/hosts/b", SourceFamily: "host", SourceRecordID: "b", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "PRIVATE-UNSELECTED-HOST", "operating_system": "PRIVATE-OS"}},
		{Path: "/parties/p", SourceFamily: "party", SourceRecordID: "p", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Party label", "description": "PRIVATE-PARTY-DESCRIPTION"}},
		{Path: "/timeline/c", SourceFamily: "timeline_event", SourceRecordID: "c", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"capture_state": "reviewed", "value": "selected event"}},
		{Path: "/timeline/d", SourceFamily: "timeline_event", SourceRecordID: "d", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"capture_state": "reviewed", "value": "PRIVATE-UNSELECTED-EVENT"}},
	})
	before, _ := canonicalJSON(model)
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	contract.AllowedExportModelBindings = []string{"$.records[*]", "$.sections[*]"}
	contract.Sections[0].Bindings = []string{"$.records['a'].fields['body']"}
	result := templateDerivationResult(t, model, contract, func(profile *RedactionProfile) {
		// If an unselected Host enters action selection it allocates a token,
		// making this detect metadata leaks as well as literal-value leaks.
		profile.Rules = append(profile.Rules, RedactionRule{RuleID: "unused-host-token", ContentClass: fixtureNote(ContentClassSourceEvidence), Action: RedactionActionSpec{Type: ActionMask, TokenBacked: true}})
	})
	if len(result.Model.Records) != 1 || result.Model.Records[0].RecordID != "a" || len(result.Model.Records[0].Fields) != 1 || len(result.Model.Subjects) != 0 || len(result.Model.TimelineEvents) != 0 || len(result.Manifest.Entries) != 1 || result.TokenManifest != nil {
		t.Fatal("allowlist or unrelated source expanded the selected closure")
	}
	if bytes.Contains(result.ModelJSON, []byte("PRIVATE-")) || len(result.Manifest.StructuralRemovals) != 0 {
		t.Fatal("unselected values leaked or selection invented redaction removals")
	}
	assertTemplateCandidateReplay(t, contract, model)
	contract.Sections[0].Bindings = []string{"$.records['missing'].fields['body']"}
	empty := templateDerivationResult(t, model, contract, nil)
	if len(empty.Model.Records) != 0 || len(empty.Manifest.Entries) != 0 {
		t.Fatal("missing selection retained unrelated records")
	}
	_, err := deriveTemplateSections(context.Background(), contract, empty.Model)
	assertTemplateDerivationReason(t, err, "required_section_empty")

	// Expansion labels select just their display dependency, not the Party's
	// unrelated fields. The ordinary binding remains a different record.
	contract.Sections[0].Bindings = []string{"$.records['a'].fields['body']"}
	contract.Sections[0].Expansion, contract.Sections[0].Title = "per_party", "For {expansion_label}"
	labels := templateDerivationResult(t, model, contract, nil)
	if bytes.Contains(labels.ModelJSON, []byte("PRIVATE-")) || len(labels.Model.Records) != 2 {
		t.Fatal("label dependency copied unrelated fields")
	}
	instances, err := deriveTemplateSections(context.Background(), contract, labels.Model)
	if err != nil || len(instances) != 1 || instances[0].Section.Title != "For Party label" {
		t.Fatal("explicit label dependency was lost", err)
	}

	contract.Sections[0].Expansion, contract.Sections[0].Title = "none", "Selected"
	decl := fixtureTimelineDeclaration("flowchart", []string{"c"})
	contract.Sections[0].DiagramDecls = []TemplateDiagramDeclaration{{SchemaID: "template_diagram_decl.v1", DeclID: decl.DeclID, DiagramKind: decl.DiagramKind, DiagramSourceKind: decl.DiagramSourceKind, SelectionRule: decl.SelectionRule, Required: true}}
	diagram := templateDerivationResult(t, model, contract, nil)
	if len(diagram.Model.TimelineEvents) != 1 || diagram.Model.TimelineEvents[0].RecordID != "c" || bytes.Contains(diagram.ModelJSON, []byte("PRIVATE-")) {
		t.Fatal("diagram selection lost its event or selected unrelated events")
	}
	contract.Sections[0].DiagramDecls = []TemplateDiagramDeclaration{}

	// The independently bound composition may add a subject display use.
	limits, _ := resolveRenderLimits(nil)
	composition := &renderComposition{AuthoredTexts: []compositionAuthoredText{{AuthoredTextID: "text", TextRole: "authored_text", Body: "Host {{subject:" + canonicalSubjectRef("host", "b") + "}}", DisclosurePartitionRef: "internal_only"}}}
	profile, digest, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	authored, err := redactExportModel(context.Background(), model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil, renderSelectionInput{template: &contract, composition: composition, limits: limits})
	if err != nil {
		t.Fatal(err)
	}
	if len(authored.Model.PresentationTexts) != 1 || authored.Model.PresentationTexts[0].Body.Value != "Host PRIVATE-UNSELECTED-HOST" || bytes.Contains(authored.ModelJSON, []byte("PRIVATE-OS")) || bytes.Contains(authored.ModelJSON, []byte("PRIVATE-PARTY")) {
		t.Fatal("authored subject selection copied unrelated content")
	}
	after, _ := canonicalJSON(model)
	if !bytes.Equal(before, after) {
		t.Fatal("selection mutated frozen source")
	}
}

func testTemplateRelationshipReachability(t *testing.T) {
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/hosts/a", SourceFamily: "host", SourceRecordID: "a", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"display_name": "endpoint host"}},
		{Path: "/evidence/b", SourceFamily: "evidence", SourceRecordID: "b", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"description": "endpoint evidence"}},
		{Path: "/envelopes/a", SourceFamily: "record_envelope", SourceRecordType: "host", SourceRecordID: "a", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "host"}},
		{Path: "/envelopes/b", SourceFamily: "record_envelope", SourceRecordType: "evidence", SourceRecordID: "b", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"record_type": "evidence"}},
		{Path: "/links/link", SourceFamily: "record_link", SourceRecordID: "link", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"src_record_id": "a", "dst_record_id": "b", "link_type": "related"}},
		{Path: "/notes/c", SourceFamily: "note", SourceRecordType: "artifact", SourceRecordID: "c", ContentClass: ContentClassWorkingMaterial, Value: map[string]any{"body": "Cited statement"}, SupportRefs: []string{"/hosts/a"}},
	})
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	paths, err := redactionFieldModelPaths(model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	contract.Sections[0].Bindings = []string{paths[model.Relationships[0].SourceRefs[0]]}
	result := templateDerivationResult(t, model, contract, nil)
	if len(result.Model.Relationships) != 1 || len(result.Model.Records) != 2 || !bytes.Contains(result.ModelJSON, []byte("endpoint host")) || !bytes.Contains(result.ModelJSON, []byte("endpoint evidence")) || bytes.Contains(result.ModelJSON, []byte("Cited statement")) {
		t.Fatal("relationship dependencies did not close exactly")
	}
	for _, ref := range result.Model.SourceIndex {
		if ref.SourceFamily == "record_envelope" && ref.SourcePath != nil {
			t.Fatal("endpoint anchor selected envelope values")
		}
	}
	contract.Sections[0].Bindings = []string{"$.records['c'].fields['body']"}
	contract.AllowedExportModelBindings = []string{"$.records[*]"}
	citation := templateDerivationResult(t, model, contract, nil)
	if len(citation.Model.Records) != 1 || len(citation.Model.Relationships) != 0 || len(citation.Model.SupportIndex) != 1 || bytes.Contains(citation.ModelJSON, []byte("endpoint host")) {
		t.Fatal("support citation selected source values")
	}
	fields, err := model.RedactionFields()
	if err != nil {
		t.Fatal(err)
	}
	bindings, err := templateFieldBindings(model.SnapshotContent, paths)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	selected, err := selectReachableSourceFields(ctx, model, fixtureSourceFactsForContent(model.SnapshotContent), fields, bindings, RedactionProfile{}, ReleaseScopeInternalReview, nil, renderSelectionInput{template: &contract})
	if err != context.Canceled || selected != nil {
		t.Fatal("cancelled source selection produced a partial result", err)
	}
	contract.Sections[0].Bindings = []string{paths[model.Relationships[0].SourceRefs[0]]}
	for i := range fields {
		if fields[i].SourceFamily == "record_link" {
			fields[i].lifecycle = lifecycleSuperseded
		}
	}
	selected, err = selectReachableSourceFields(context.Background(), model, fixtureSourceFactsForContent(model.SnapshotContent), fields, bindings, RedactionProfile{}, ReleaseScopeExternal, nil, renderSelectionInput{template: &contract})
	if err != nil || len(selected) != 0 {
		t.Fatal("ineligible relationship selected active endpoint values", err)
	}
}
