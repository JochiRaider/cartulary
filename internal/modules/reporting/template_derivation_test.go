package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func templateDerivationResult(t *testing.T, model ExportModel, contract TemplateContract, mutate func(*RedactionProfile)) RedactionResult {
	t.Helper()
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(&profile)
	}
	fixtureAllowContainerRules(t, &profile, model.SnapshotContent)
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := redactExportModel(context.Background(), model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil, renderSelectionInput{template: &contract})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertTemplateDerivationReason(t *testing.T, err error, reason string) {
	t.Helper()
	var validation *renderValidationError
	if !errors.As(err, &validation) || validation.ReasonCode != reason {
		t.Fatalf("wanted %s, got %v", reason, err)
	}
}

func assertTemplateCandidateReplay(t *testing.T, contract TemplateContract, model ExportModel) {
	t.Helper()
	encoded, err := canonicalJSON(model)
	if err != nil {
		t.Fatal(err)
	}
	request := CreateReleaseRequest{ReleaseScope: ReleaseScopeInternalReview, OutputKind: OutputKindSlidev, OutputOptions: fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalReview), RedactionProfileID: InternalRedactionProfileID, RedactionProfileVersion: DefaultRedactionProfileVersion}
	input := renderCandidateFixture(request, contract, model, reportingObjectDigest(ExportModelSchemaID, encoded))
	rendered, reason, err := fixtureRenderer().Render(context.Background(), input)
	if err != nil {
		t.Fatal("template candidate failed", reason, err)
	}
	if err := validateRenderCandidateBinding(input, rendered); err != nil {
		t.Fatal("template replay differs", err)
	}
}

func testTemplateSectionExpansion(t *testing.T) {
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/parties/z", SourceFamily: "party", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Secret Z"}},
		{Path: "/parties/a", SourceFamily: "party", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Secret A"}},
		{Path: "/hosts/h", SourceFamily: "host", ContentClass: ContentClassSourceEvidence, Value: map[string]any{"display_name": "Secret Host"}},
	})
	contract.Sections[0].Expansion, contract.Sections[0].Title = "per_party", "For {expansion_label}"
	result := templateDerivationResult(t, model, contract, nil)
	before, _ := canonicalJSON(result.Model)
	instances, err := deriveTemplateSections(context.Background(), contract, result.Model)
	if err != nil || len(instances) != 2 {
		t.Fatal("Party expansion", err)
	}
	for i, want := range []struct{ id, title, order string }{{"a", "For Secret A", "0001.0001"}, {"z", "For Secret Z", "0001.0002"}} {
		instance := instances[i]
		if instance.ExpansionKey == nil || *instance.ExpansionKey != want.id || instance.Section.OrderingKey != want.order || instance.Section.SectionID != "sec-"+want.order || instance.Section.Title != want.title || len(instance.TitleRefs.source) != 1 {
			t.Fatal("unstable dimension identity or label", instance)
		}
		for _, block := range instance.Section.Blocks {
			for _, field := range block.Fields {
				if field.Value == "Secret A" && want.id != "a" || field.Value == "Secret Z" && want.id != "z" {
					t.Fatal("another Party leaked into instance")
				}
			}
		}
	}
	deck, err := fixtureDerivedDeck(contract, result.Model, renderOutputOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	index, err := resolveCompositionSection(deck.Slides, &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: contract.Sections[0].DeclID, ExpansionKey: fixtureNote("z")})
	if err != nil || index != 1 {
		t.Fatal("semantic expansion anchor did not resolve", err)
	}
	_, err = resolveCompositionSection(deck.Slides, &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: contract.Sections[0].DeclID})
	assertTemplateDerivationReason(t, err, "composition_anchor_unresolved")
	for _, slide := range deck.Slides {
		if len(slide.titleSourceRefs) != 1 || !slices.Contains(slide.SourceRefs, slide.titleSourceRefs[0]) {
			t.Fatal("title evidence missing")
		}
	}
	assertTemplateCandidateReplay(t, contract, model)
	after, _ := canonicalJSON(result.Model)
	if !bytes.Equal(before, after) {
		t.Fatal("expansion mutated retained model")
	}
	slices.Reverse(result.Model.admittedPartyIDs)
	if reordered, err := deriveTemplateSections(context.Background(), contract, result.Model); err != nil || !reflect.DeepEqual(reordered, instances) {
		t.Fatal("Party input order affected expansion", err)
	}
	// The external dimension roster intersects frozen identities with recipients.
	external := result.Model
	external.ReleaseScope, external.RecipientPartitionRefs = ReleaseScopeExternal, []string{"party:z"}
	dimensions, err := templateSectionDimensions(contract.Sections[0], external)
	if err != nil || len(dimensions) != 1 || *dimensions[0].key != "z" || dimensions[0].includes(ReportingField{DisclosurePartitionRefs: []string{"party:z", "party:a"}}) || dimensions[0].includes(ReportingField{DisclosurePartitionRefs: []string{"internal_only"}}) {
		t.Fatal("recipient or all-partitions boundary failed", err)
	}
	// Removed Party labels never fall back to frozen raw values; Party identity
	// still participates in internal expansion independently of redaction.
	contract.Sections[0].SectionKind = "validation"
	dropped := templateDerivationResult(t, model, contract, func(p *RedactionProfile) { p.DefaultAction = RedactionActionSpec{Type: ActionDrop} })
	instances, err = deriveTemplateSections(context.Background(), contract, dropped.Model)
	if err != nil || len(instances) != 2 || instances[0].Section.Title != "For Party 0001" || len(instances[0].TitleRefs.source) != 0 {
		t.Fatal("removed Party did not use ordinal fallback", err)
	}
	contract.Sections[0].Expansion, contract.Sections[0].SectionKind = "per_subject", "appendix"
	tokenized := templateDerivationResult(t, model, contract, func(p *RedactionProfile) {
		p.Rules = nil
		p.DefaultAction = RedactionActionSpec{Type: ActionMask, TokenBacked: true}
	})
	instances, err = deriveTemplateSections(context.Background(), contract, tokenized.Model)
	if err != nil || len(instances) != 3 {
		t.Fatal("subject expansion", err)
	}
	wantTitles := []string{"For HOST-0001", "For PARTY-0001", "For PARTY-0002"}
	for i, instance := range instances {
		if instance.Section.Title != wantTitles[i] || len(instance.Section.Blocks) != 1 || len(instance.TitleRefs.source) != 1 {
			t.Fatal("subject token or membership differs", instance)
		}
	}
	slices.Reverse(tokenized.Model.Subjects)
	if reordered, err := deriveTemplateSections(context.Background(), contract, tokenized.Model); err != nil || !reflect.DeepEqual(reordered, instances) {
		t.Fatal("subject order affected output", err)
	}
	contract.DeclaredLimits["sections.count"] = json.RawMessage(`2`)
	_, err = deriveTemplateSections(context.Background(), contract, tokenized.Model)
	assertTemplateDerivationReason(t, err, "sections_count_exceeded")
	delete(contract.DeclaredLimits, "sections.count")
	contract.DeclaredLimits["fields.count"] = json.RawMessage(`2`)
	_, err = deriveTemplateSections(context.Background(), contract, tokenized.Model)
	assertTemplateDerivationReason(t, err, "fields_count_exceeded")
	delete(contract.DeclaredLimits, "fields.count")
	_, err = deriveTemplateSections(context.Background(), contract, dropped.Model)
	assertTemplateDerivationReason(t, err, "required_section_empty")
	contract.Sections[0].Required = false
	if got, err := deriveTemplateSections(context.Background(), contract, dropped.Model); err != nil || len(got) != 0 {
		t.Fatal("optional empty dimension emitted content", err)
	}
}

func testTemplateNarrativeSlots(t *testing.T) {
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	contract.AllowedExportModelBindings = []string{"$.records[*]", "$.sections[*]"}
	contract.Sections[0].Bindings = []string{}
	contract.Sections[0].SectionKind = "narrative"
	contract.NarrativeSlots = []TemplateNarrativeSlot{{SlotID: "analysis", SectionDeclID: contract.Sections[0].DeclID, Binding: "$.records[*].fields['body']", Required: true, MaxChars: 3}}
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/notes/z", SourceFamily: "note", SourceRecordID: "z", SourceRecordType: "artifact", ContentClass: ContentClassCuratedNarrative, Value: map[string]any{"title": "Not selected", "body": "Z"}},
		{Path: "/notes/a", SourceFamily: "note", SourceRecordID: "a", SourceRecordType: "artifact", ContentClass: ContentClassWorkingMaterial, Value: map[string]any{"body": "é🙂"}},
	})
	if err := validateTemplateContract(contract, OutputKindSlidev, ReleaseScopeInternalReview); err != nil {
		t.Fatal(err)
	}
	result := templateDerivationResult(t, model, contract, nil)
	before, _ := canonicalJSON(result.Model)
	deck, err := fixtureDerivedDeck(contract, result.Model, renderOutputOptions{}, nil)
	if err != nil || len(deck.Slides) != 1 || len(deck.Slides[0].Blocks) != 2 {
		t.Fatal("bound narrative was not emitted", err)
	}
	for i, value := range []string{"é🙂", "Z"} {
		block := deck.Slides[0].Blocks[i]
		if block.BlockKind != "paragraph" || len(block.Fields) != 1 || block.Fields[0].Value != value || len(block.SourceRefs) == 0 || block.Fields[0].SourceValueState != "present" {
			t.Fatal("narrative lost selected value or provenance", block)
		}
	}
	if deck.Slides[0].Blocks[0].ContentClass != ContentClassWorkingMaterial || deck.Slides[0].Blocks[1].ContentClass != ContentClassCuratedNarrative {
		t.Fatal("narrative reclassified as presentation")
	}
	assertTemplateCandidateReplay(t, contract, model)
	after, _ := canonicalJSON(result.Model)
	if !bytes.Equal(before, after) {
		t.Fatal("narrative mutated retained source")
	}
	contract.NarrativeSlots[0].MaxChars = 2
	_, err = fixtureDerivedDeck(contract, result.Model, renderOutputOptions{}, nil)
	assertTemplateDerivationReason(t, err, "invalid_scalar_value")
	contract.NarrativeSlots[0].MaxChars = 100
	paths := fixtureModelPaths(t, model.SnapshotContent, "/notes/a/body", "/notes/z/body", "/notes/z/title")
	masked := templateDerivationResult(t, model, contract, func(p *RedactionProfile) {
		a, z := paths["/notes/a/body"], paths["/notes/z/body"]
		p.Rules = []RedactionRule{{RuleID: "mask-a", Path: &a, Action: RedactionActionSpec{Type: ActionMask}}, {RuleID: "drop-z", Path: &z, Action: RedactionActionSpec{Type: ActionDrop}}}
	})
	deck, err = fixtureDerivedDeck(contract, masked.Model, renderOutputOptions{}, nil)
	if err != nil || len(deck.Slides[0].Blocks) != 1 || deck.Slides[0].Blocks[0].Fields[0].RedactedValueState != "masked" {
		t.Fatal("slot restored removed text", err)
	}
	encoded, _ := canonicalJSON(deck)
	if bytes.Contains(encoded, []byte("é🙂")) {
		t.Fatal("masked narrative restored raw content")
	}
	dropped := templateDerivationResult(t, model, contract, func(p *RedactionProfile) { p.DefaultAction = RedactionActionSpec{Type: ActionDrop} })
	_, err = fixtureDerivedDeck(contract, dropped.Model, renderOutputOptions{}, nil)
	assertTemplateDerivationReason(t, err, "narrative_slot_unbound")
	contract.NarrativeSlots[0].Required, contract.Sections[0].Required = false, false
	if instances, err := deriveTemplateSections(context.Background(), contract, dropped.Model); err != nil || len(instances) != 0 {
		t.Fatal("optional slot restored data", err)
	}
	contract.NarrativeSlots[0].Binding = "$.records[*]"
	if validateTemplateContract(contract, OutputKindSlidev, ReleaseScopeInternalReview) == nil {
		t.Fatal("ambiguous record-wide narrative selector admitted")
	}
	contract.NarrativeSlots[0].Binding = "$.records[*].fields['body']"
	nonArtifact := model.SnapshotContent
	nonArtifact.Records = slices.Clone(nonArtifact.Records)
	nonArtifact.Records[0].RecordType = "evidence"
	assertTemplateDerivationReason(t, validateNarrativeSourceBindings(&contract, nonArtifact), "invalid_scalar_value")
	contract.NarrativeSlots[0].Required = true
	for _, test := range []struct {
		name   string
		value  any
		reason string
	}{
		{"null", nil, "narrative_slot_unbound"}, {"empty", "", "narrative_slot_unbound"}, {"number", 42, "invalid_type"},
	} {
		t.Run(test.name, func(t *testing.T) {
			model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/notes/a", SourceFamily: "note", SourceRecordID: "a", SourceRecordType: "artifact", ContentClass: ContentClassWorkingMaterial, Value: map[string]any{"body": test.value}}})
			result := templateDerivationResult(t, model, contract, nil)
			_, err := deriveTemplateSections(context.Background(), contract, result.Model)
			assertTemplateDerivationReason(t, err, test.reason)
		})
	}
	if strings.Contains(string(result.ModelJSON), "admittedPartyIDs") || strings.Contains(string(result.ModelJSON), "subjectDisplayTokens") {
		t.Fatal("private expansion context serialized")
	}
}
