package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func testCanonicalRenderModelAssembly(t *testing.T) {
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/notes/n", SourceFamily: "note", SourceRecordID: "n", SourceRecordType: "artifact", ContentClass: ContentClassWorkingMaterial, DisclosurePartitionRefs: []string{"public"}, Value: map[string]any{"body": "Shared paragraph", "title": "UNSELECTED-TITLE"}},
		{Path: "/parties/a", SourceFamily: "party", SourceRecordID: "a", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Party A"}},
		{Path: "/parties/z", SourceFamily: "party", SourceRecordID: "z", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Party Z"}},
		{Path: "/timeline/e", SourceFamily: "timeline_event", SourceRecordID: "e", ContentClass: ContentClassDerivedAnalytic, Value: map[string]any{"capture_state": "reviewed", "value": "Selected event"}},
	})
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	contract.AllowedExportModelBindings = []string{"$.records[*]", "$.sections[*]"}
	contract.Sections[0].Bindings = []string{"$.records['n'].fields['body']"}
	contract.Sections[0].Expansion, contract.Sections[0].Title = "per_party", "For {expansion_label}"
	decl := fixtureTimelineDeclaration("flowchart", []string{"e"})
	contract.Sections[0].DiagramDecls = []TemplateDiagramDeclaration{{SchemaID: "template_diagram_decl.v1", DeclID: decl.DeclID, DiagramKind: decl.DiagramKind, DiagramSourceKind: decl.DiagramSourceKind, SelectionRule: decl.SelectionRule, Required: true}}
	second := contract.Sections[0]
	second.DeclID, second.Expansion, second.Title, second.DiagramDecls = "repeat", "none", "Again", []TemplateDiagramDeclaration{}
	contract.Sections = append(contract.Sections, second)
	encoded, _ := canonicalJSON(model)
	request := CreateReleaseRequest{ReleaseScope: ReleaseScopeInternalReview, OutputKind: OutputKindSlidev, OutputOptions: fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalReview), RedactionProfileID: InternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion}
	input := renderCandidateFixture(request, contract, model, reportingObjectDigest(ExportModelSchemaID, encoded))
	before, _ := canonicalJSON(input)
	result, err := redactRenderInput(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	var canonical ExportModel
	if err := json.Unmarshal(result.ModelJSON, &canonical); err != nil {
		t.Fatal(err)
	}
	if len(canonical.Sections) != 3 || len(canonical.Diagrams) != 1 || len(canonical.Diagrams[0].Nodes) != 1 || bytes.Contains(result.ModelJSON, []byte("UNSELECTED-TITLE")) {
		t.Fatal("canonical content is not the selected expanded model")
	}
	for index, want := range []struct {
		id, title string
		blocks    int
	}{{"sec-0001.0001", "For Party A", 2}, {"sec-0001.0002", "For Party Z", 2}, {"sec-0002", "Again", 1}} {
		section := canonical.Sections[index]
		if section.SectionID != want.id || section.Title != want.title || len(section.Blocks) != want.blocks || section.Blocks[0].Fields[0].Value != "Shared paragraph" {
			t.Fatal("expanded canonical section differs", section.SectionID)
		}
		if want.blocks == 2 && (section.Blocks[1].BlockKind != "diagram_ref" || section.Blocks[1].Fields[0].Value != "timeline") {
			t.Fatal("diagram reference absent from canonical section")
		}
	}
	if result.ModelSHA256 != reportingObjectDigest(ExportModelSchemaID, result.ModelJSON) {
		t.Fatal("digest precedes final assembly")
	}
	// Removing private selection indexes cannot change already assembled output.
	consumer := result.Model
	consumer.admittedFieldPaths, consumer.admittedPartyIDs, consumer.subjectDisplayTokens = nil, nil, nil
	deck, err := deriveDeckModel(contract, consumer, renderOutputOptions{}, nil)
	if err != nil || len(deck.Slides) != 3 {
		t.Fatal("deck attempted source reselection", err)
	}
	limits, _ := resolveRenderLimits(nil)
	prepared, err := prepareDiagrams(contract, consumer, OutputKindSlidev, limits)
	if err != nil || len(prepared.Diagrams) != 1 || !reflect.DeepEqual(prepared.Diagrams[0].Model.Nodes, canonical.Diagrams[0].Nodes) {
		t.Fatal("diagram did not consume canonical content", err)
	}
	deck.Slides[0].Blocks[0].Fields[0].Value = "mutated deck"
	prepared.Diagrams[0].Model.Nodes[0].Label = "mutated diagram"
	after, _ := canonicalJSON(result.Model)
	if !bytes.Equal(after, result.ModelJSON) {
		t.Fatal("consumer mutated canonical content")
	}
	copy, err := copyReportingModel(result.Model.ExportModel)
	if err != nil {
		t.Fatal(err)
	}
	consumer.ExportModel = copy
	consumer.Diagrams[0].Nodes[0].Label = "tampered canonical content"
	if _, err := prepareDiagrams(contract, consumer, OutputKindSlidev, limits); err == nil {
		t.Fatal("altered model was rendered")
	}
	unassembled := result.Model
	unassembled.assembly = nil
	if _, err := deriveDeckModel(contract, unassembled, renderOutputOptions{}, nil); err == nil {
		t.Fatal("unassembled model admitted")
	}
	changed := contract
	changed.DeckTitle = "changed template"
	if _, err := deriveDeckModel(changed, result.Model, renderOutputOptions{}, nil); err == nil {
		t.Fatal("different template admitted")
	}
	for _, bound := range []struct{ key, value, reason string }{{"sections.count", "2", "sections_count_exceeded"}, {"blocks.count", "4", "blocks_count_exceeded"}, {"fields.count", "4", "fields_count_exceeded"}, {"records.count", "1", "records_count_exceeded"}, {"subjects.count", "1", "subject_limit_exceeded"}, {"export_model.canonical_bytes", "1", "export_model_canonical_bytes_exceeded"}} {
		candidate := input
		candidate.TemplateContract.DeclaredLimits = map[string]json.RawMessage{bound.key: json.RawMessage(bound.value)}
		partial, err := redactRenderInput(context.Background(), candidate)
		assertTemplateDerivationReason(t, err, bound.reason)
		if len(partial.ModelJSON) != 0 {
			t.Fatal("resource failure published partial model")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if partial, err := redactRenderInput(ctx, input); !errors.Is(err, context.Canceled) || len(partial.ModelJSON) != 0 {
		t.Fatal("canceled assembly published bytes", err)
	}
	assertTemplateCandidateReplay(t, contract, model)
	if compareRenderMaterialization(result, result) != nil {
		t.Fatal("identical materialization differs")
	}
	for _, change := range []struct {
		reason string
		mutate func(*RedactionResult)
	}{
		{"export_model_mismatch", func(r *RedactionResult) { r.ModelJSON = []byte("changed") }},
		{"export_model_mismatch", func(r *RedactionResult) { r.ModelSHA256 = "changed" }},
		{"redaction_manifest_mismatch", func(r *RedactionResult) { r.ManifestJSON = []byte("changed") }},
		{"token_manifest_mismatch", func(r *RedactionResult) { r.TokenManifestJSON = []byte("changed") }},
	} {
		candidate := result
		change.mutate(&candidate)
		assertTemplateDerivationReason(t, compareRenderMaterialization(result, candidate), change.reason)
	}
	after, _ = canonicalJSON(input)
	if !bytes.Equal(before, after) {
		t.Fatal("assembly mutated admission input")
	}
}

// Low-level serializer fixtures intentionally enter after admission. They may
// contain invalid values whose rejection belongs to the serializer under test.
// Production never uses this fixture adapter or accepts unassembled content.
func fixtureAssembledModel(ctx context.Context, contract TemplateContract, model RedactedExportModel, kind string, composition json.RawMessage) (RedactedExportModel, error) {
	instances := []templateSectionInstance{}
	if kind == OutputKindSlidev {
		var err error
		instances, err = deriveTemplateSections(ctx, contract, model)
		if err != nil {
			return RedactedExportModel{}, err
		}
	}
	diagrams, err := deriveDiagramModels(contract, model, composition, model.ReleaseScope, kind, nil)
	if err != nil {
		return RedactedExportModel{}, err
	}
	model.Sections, model.Diagrams = []ReportingSection{}, diagrams
	// Timeline serializer fixtures supplied selected events directly. That
	// source metadata is no longer a renderer input once diagrams are assembled.
	model.TimelineEvents = []ReportingTimelineEvent{}
	model.assembly = &renderModelAssembly{sections: map[string]renderSectionOrigin{}}
	for _, instance := range instances {
		section := instance.Section
		for _, declared := range instance.Declaration.DiagramDecls {
			for _, diagram := range diagrams {
				if declared.DeclID == diagram.DiagramID {
					section.Blocks = append(section.Blocks, diagramReferenceBlock(diagram))
				}
			}
		}
		rebuildCompositionBlocks(section.Blocks, section.SectionID, false)
		model.Sections = append(model.Sections, section)
		for index, decl := range contract.Sections {
			if decl.DeclID == instance.Declaration.DeclID {
				model.assembly.sections[section.SectionID] = renderSectionOrigin{index, instance.ExpansionKey, instance.TitleRefs}
			}
		}
	}
	encoded, err := canonicalJSON(model)
	if err != nil {
		return RedactedExportModel{}, err
	}
	model.assembly.modelSHA256 = reportingObjectDigest(model.SchemaID, encoded)
	template, err := canonicalJSON(templateManifestView(contract))
	if err != nil {
		return RedactedExportModel{}, err
	}
	model.assembly.templateSHA256 = reportingObjectDigest("cartulary.reporting_template_pack_manifest.v3", template)
	return model, nil
}

func fixtureDerivedDeck(contract TemplateContract, model RedactedExportModel, options renderOutputOptions, composition json.RawMessage) (reportingDeckModel, error) {
	model, err := fixtureAssembledModel(context.Background(), contract, model, OutputKindSlidev, composition)
	if err != nil {
		return reportingDeckModel{}, err
	}
	return deriveDeckModel(contract, model, options, composition)
}

func (renderer *Renderer) fixtureRenderReportBundle(ctx context.Context, contract TemplateContract, kind string, model RedactedExportModel, profile RedactionProfileView, manifestSHA, scope string, options, composition json.RawMessage, artifacts RedactionBundleArtifacts) (RenderBundle, error) {
	if err := validateCompositionRenderBinding(composition, contract, model, scope, profile); err != nil {
		return RenderBundle{}, err
	}
	model.ReleaseScope = scope
	model, err := fixtureAssembledModel(ctx, contract, model, kind, composition)
	if err != nil {
		return RenderBundle{}, err
	}
	return renderer.renderReportBundle(ctx, contract, kind, model, profile, manifestSHA, scope, options, composition, artifacts)
}
