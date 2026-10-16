package reporting

import (
	"context"
	"errors"
	"testing"
)

func testCompositionRecordAnchors(t *testing.T) {
	source, result := retainedRecordFixture(t)
	recordAnchors := newDeckRecordAnchors(result.Model.Records, result.Model.SourceIndex)
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	instances, err := deriveTemplateSections(context.Background(), contract, result.Model)
	if err != nil {
		t.Fatal(err)
	}
	base := reportingDeckModel{Slides: []reportingDeckSlide{}}
	for _, instance := range instances {
		base.Slides = append(base.Slides, reportingDeckSlide{Blocks: instance.Section.Blocks, sectionOrigin: &deckSectionOrigin{DeclarationID: instance.Declaration.DeclID}, derivedBlockIDs: derivedDeckBlockIDs(instance.Section.Blocks), recordAnchors: recordAnchors})
	}
	anchor := func(id string) *compositionBlockAnchor {
		return &compositionBlockAnchor{AnchorKind: "block_anchor", SectionAnchor: compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: contract.Sections[0].DeclID}, BlockKind: "paragraph", RecordAnchor: &compositionRecordAnchor{AnchorKind: "record_anchor", RecordID: id}}
	}
	check := func(slides []reportingDeckSlide, a *compositionBlockAnchor, reason string) {
		t.Helper()
		_, err := resolveCompositionBlock(slides, a)
		var failure *renderValidationError
		if !errors.As(err, &failure) || failure.ReasonCode != reason {
			t.Fatalf("want %s, got %v", reason, err)
		}
	}
	// A record qualifier disambiguates several paragraphs. Visible values,
	// arbitrary selectors and an auxiliary source with the same ID do not.
	if _, err := resolveCompositionBlock(base.Slides, anchor("one")); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"two", "raw first", "sec-0001", "/opaque/a"} {
		check(base.Slides, anchor(id), "composition_anchor_unresolved")
	}
	bare := anchor("one")
	bare.RecordAnchor = nil
	check(base.Slides, bare, "composition_anchor_ambiguous")
	bad := anchor("one")
	bad.RecordAnchor.AnchorKind = "block_anchor"
	check(base.Slides, bad, "invalid_scalar_value")
	// Exercise the digest-bound wire operation through the actual redaction
	// projection and deck derivation, before chunking changes structural IDs.
	composition := sealedReportingCompositionFixture(t, []byte(`{"schema_id":"cartulary.report_composition.v1","deck_ops":[{"op_id":"remove","op_kind":"exclude_block","on_unresolved":"fail","payload":{"block_anchor":{"anchor_kind":"block_anchor","section_anchor":{"anchor_kind":"section_anchor","template_section_decl_id":"`+contract.Sections[0].DeclID+`","expansion_key":null},"block_kind":"paragraph","record_anchor":{"anchor_kind":"record_anchor","record_id":"one"}}}}],"diagram_decls":[],"authored_texts":[]}`), source.IncidentID, contract)
	deck, err := fixtureDerivedDeck(contract, result.Model, renderOutputOptions{}, composition)
	if err != nil {
		t.Fatal(err)
	}
	if len(deck.Slides[0].Blocks) != len(base.Slides[0].Blocks)-1 {
		t.Fatal("record-qualified removal did not remove exactly one block")
	}
	for _, block := range deck.Slides[0].Blocks {
		for _, field := range block.Fields {
			if field.Value == DefaultMaskText {
				t.Fatal("targeted record survived the wire operation")
			}
		}
	}
	// An invalid retained summary set cannot turn multiple records into one
	// successful target. An ambiguous block is also never drop-eligible.
	duplicate := append([]ReportingRecordSummary{}, result.Model.Records...)
	duplicate = append(duplicate, duplicate[0])
	base.Slides[0].recordAnchors = newDeckRecordAnchors(duplicate, source.SourceIndex)
	check(base.Slides, anchor(duplicate[0].RecordID), "composition_anchor_ambiguous")
	base.Slides[0].recordAnchors = recordAnchors
	target, err := resolveCompositionBlock(base.Slides, anchor("one"))
	if err != nil {
		t.Fatal(err)
	}
	matching := (*target.siblings)[target.index]
	second := matching
	second.BlockID = "another-derived"
	base.Slides[0].Blocks = append(base.Slides[0].Blocks, second)
	base.Slides[0].derivedBlockIDs[second.BlockID] = true
	check(base.Slides, anchor("one"), "composition_anchor_ambiguous")
	base.Slides[0].Blocks = base.Slides[0].Blocks[:len(base.Slides[0].Blocks)-1]
	remove := compositionDeckOp{OpKind: "exclude_block", Payload: compositionOpPayload{BlockAnchor: anchor("one")}}
	if err := applyCompositionToDeck(&base.Slides, &renderComposition{DeckOps: []compositionDeckOp{remove, remove}}, contract, ReleaseScopeInternalReview); err == nil {
		t.Fatal("removed record block survived as a target")
	}
	testCompositionRecordSubtree(t, recordAnchors)
}

func testCompositionRecordSubtree(t *testing.T, records *deckRecordAnchors) {
	var ref string
	for candidate, id := range records.sources {
		if id == "one" {
			ref = candidate
			break
		}
	}
	child := ReportingBlock{BlockID: "child", BlockKind: "paragraph", Fields: []ReportingField{{SourceRefs: []string{ref}}}}
	parent := ReportingBlock{BlockID: "parent", BlockKind: "bullet_list", Children: []ReportingBlock{child}, SourceRefs: []string{ref}}
	derived := derivedDeckBlockIDs([]ReportingBlock{parent})
	if !records.blockContains(parent, "one", derived) {
		t.Fatal("nested source membership missing")
	}
	parent.Children = nil
	if records.blockContains(parent, "one", derived) {
		t.Fatal("removed child survived in cached aggregate refs")
	}
	child.BlockID = "authored"
	parent.Children = []ReportingBlock{child}
	if records.blockContains(parent, "one", derived) {
		t.Fatal("authored insertion manufactured derived source membership")
	}
}
