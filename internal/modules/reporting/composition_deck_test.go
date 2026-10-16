package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func testCompositionDeckOperations(t *testing.T) {
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	anchor := func(id string) *compositionSectionAnchor {
		return &compositionSectionAnchor{AnchorKind: "section_anchor", DeclarationID: id}
	}
	blockAnchor := func(id, kind string) *compositionBlockAnchor {
		return &compositionBlockAnchor{AnchorKind: "block_anchor", SectionAnchor: *anchor(id), BlockKind: kind}
	}
	fixture := func() []reportingDeckSlide {
		result := []reportingDeckSlide{}
		for _, id := range []string{"first", "second", "third"} {
			blocks := []ReportingBlock{{BlockID: id + "-source", BlockKind: "paragraph", ContentClass: ContentClassSourceEvidence, Fields: []ReportingField{{FieldKey: "value", Value: id, SourceRefs: []string{id}, DisclosurePartitionRefs: []string{"public"}}}}}
			result = append(result, reportingDeckSlide{Title: id, Layout: "default", Blocks: blocks, sectionOrigin: &deckSectionOrigin{DeclarationID: id}, derivedBlockIDs: derivedDeckBlockIDs(blocks), clickProfile: "none"})
		}
		return result
	}
	texts := []compositionAuthoredText{
		{AuthoredTextID: "title", TextRole: "title_override", Body: "Changed second", DisclosurePartitionRef: "public"},
		{AuthoredTextID: "note", TextRole: "speaker_notes", Body: "Presenter note", DisclosurePartitionRef: "internal_only"},
		{AuthoredTextID: "one", TextRole: "authored_text", Body: "One", DisclosurePartitionRef: "public"},
		{AuthoredTextID: "two", TextRole: "authored_text", Body: "Two", DisclosurePartitionRef: "public"},
	}
	op := func(kind string, p compositionOpPayload) compositionDeckOp {
		return compositionDeckOp{OpKind: kind, Payload: p}
	}
	checkReason := func(slides []reportingDeckSlide, ops []compositionDeckOp, scope, want string) {
		t.Helper()
		err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: ops, AuthoredTexts: texts}, contract, scope)
		var failure *renderValidationError
		if !errors.As(err, &failure) || failure.ReasonCode != want {
			t.Fatalf("want %s, got %v", want, err)
		}
	}
	t.Run("ordered effects and actual target", func(t *testing.T) {
		slides := fixture()
		ops := []compositionDeckOp{
			op("override_title", compositionOpPayload{SectionAnchor: anchor("second"), AuthoredTextRef: "title"}),
			op("set_speaker_notes", compositionOpPayload{SectionAnchor: anchor("second"), AuthoredTextRef: "note"}),
			op("override_slide_layout", compositionOpPayload{SectionAnchor: anchor("second"), LayoutID: "section"}),
			op("reorder_sections", compositionOpPayload{SectionAnchors: []compositionSectionAnchor{*anchor("third"), *anchor("second")}}),
			op("insert_authored_block", compositionOpPayload{BlockAnchor: blockAnchor("second", "paragraph"), Position: "after", AuthoredTextRef: "one"}),
			op("insert_authored_block", compositionOpPayload{BlockAnchor: blockAnchor("second", "paragraph"), Position: "after", AuthoredTextRef: "two"}),
		}
		if err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: ops, AuthoredTexts: texts}, contract, ReleaseScopeInternalDraft); err != nil {
			t.Fatal(err)
		}
		if slides[0].Title != "third" || slides[1].Title != "Changed second" || slides[2].Title != "first" || slides[1].Layout != "section" || *slides[1].SpeakerNotes != "Presenter note" {
			t.Fatal("operation applied to wrong section")
		}
		values := []any{}
		for _, b := range slides[1].Blocks {
			values = append(values, b.Fields[0].Value)
		}
		if !reflect.DeepEqual(values, []any{"second", "One", "Two"}) {
			t.Fatal("insertion order", values)
		}
		if slides[1].Blocks[1].BlockID != "sec-0002.b0002" || slides[1].Blocks[1].ContentClass != ContentClassPresentationText {
			t.Fatal("authored block structure")
		}
	})
	t.Run("removals and drop policy", func(t *testing.T) {
		remove := op("exclude_section", compositionOpPayload{SectionAnchor: anchor("second")})
		checkReason(fixture(), []compositionDeckOp{remove, remove}, ReleaseScopeInternalDraft, "composition_anchor_unresolved")
		dropped := remove
		dropped.OnUnresolved = "drop"
		slides := fixture()
		if err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: []compositionDeckOp{remove, dropped}}, contract, ReleaseScopeInternalDraft); err != nil || len(slides) != 2 {
			t.Fatal(err)
		}
		checkReason(fixture(), []compositionDeckOp{dropped}, ReleaseScopeExternal, "composition_drop_invalid_for_external_release")
		slides = fixture()
		slides[1].sectionOrigin.Required = true
		checkReason(slides, []compositionDeckOp{remove}, ReleaseScopeInternalDraft, "required_section_empty")
		slides = fixture()
		slides = append(slides, slides[1])
		checkReason(slides, []compositionDeckOp{dropped}, ReleaseScopeInternalDraft, "composition_anchor_ambiguous")
		bad := op("override_slide_layout", compositionOpPayload{SectionAnchor: anchor("missing"), LayoutID: "unknown"})
		bad.OnUnresolved = "drop"
		checkReason(fixture(), []compositionDeckOp{bad}, ReleaseScopeInternalDraft, "invalid_scalar_value")
	})
	t.Run("block uniqueness and current membership", func(t *testing.T) {
		remove := op("exclude_block", compositionOpPayload{BlockAnchor: blockAnchor("second", "paragraph")})
		slides := fixture()
		slides[1].Blocks = append(slides[1].Blocks, ReportingBlock{BlockID: "other", BlockKind: "paragraph"})
		slides[1].derivedBlockIDs["other"] = true
		remove.OnUnresolved = "drop"
		checkReason(slides, []compositionDeckOp{remove}, ReleaseScopeInternalDraft, "composition_anchor_ambiguous")
		remove.OnUnresolved = "fail"
		checkReason(fixture(), []compositionDeckOp{remove, remove}, ReleaseScopeInternalDraft, "composition_anchor_unresolved")
		slides = fixture()
		slides[1].sectionOrigin.Required = true
		checkReason(slides, []compositionDeckOp{remove}, ReleaseScopeInternalDraft, "required_section_empty")
		remove.Payload.BlockAnchor.RecordAnchor = &compositionRecordAnchor{AnchorKind: "record_anchor", RecordID: "unretained"}
		checkReason(fixture(), []compositionDeckOp{remove}, ReleaseScopeInternalDraft, "composition_anchor_unresolved")
	})
	t.Run("atomic dropped reorder and repeated reorder", func(t *testing.T) {
		reorder := op("reorder_sections", compositionOpPayload{SectionAnchors: []compositionSectionAnchor{*anchor("third"), *anchor("missing")}})
		reorder.OnUnresolved = "drop"
		slides := fixture()
		if err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: []compositionDeckOp{reorder}}, contract, ReleaseScopeInternalDraft); err != nil || slides[0].Title != "first" {
			t.Fatal("partial reorder", err)
		}
		reorder.Payload.SectionAnchors = []compositionSectionAnchor{*anchor("third")}
		other := reorder
		other.Payload.SectionAnchors = []compositionSectionAnchor{*anchor("second")}
		if err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: []compositionDeckOp{reorder, other}}, contract, ReleaseScopeInternalDraft); err != nil || slides[0].Title != "second" || slides[1].Title != "third" || slides[2].Title != "first" {
			t.Fatal(err)
		}
	})
	t.Run("nested removal recomputes references", func(t *testing.T) {
		slides := fixture()
		leaf := slides[1].Blocks[0]
		slides[1].Blocks = []ReportingBlock{{BlockID: "list", BlockKind: "bullet_list", Children: []ReportingBlock{leaf}}, {BlockID: "metric", BlockKind: "metric", Fields: []ReportingField{{FieldKey: "count", Value: 1, SourceRefs: []string{"metric-source"}, DisclosurePartitionRefs: []string{"public"}}}}}
		slides[1].derivedBlockIDs = derivedDeckBlockIDs(slides[1].Blocks)
		remove := op("exclude_block", compositionOpPayload{BlockAnchor: blockAnchor("second", "paragraph")})
		if err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: []compositionDeckOp{remove}}, contract, ReleaseScopeInternalDraft); err != nil {
			t.Fatal(err)
		}
		if len(slides[1].Blocks) != 1 || slides[1].Blocks[0].BlockKind != "metric" || !reflect.DeepEqual(slides[1].Blocks[0].SourceRefs, []string{"metric-source"}) {
			t.Fatal("removed parent or source survived")
		}
	})
	t.Run("notes and title partitions follow their chunks", func(t *testing.T) {
		slides := fixture()
		ops := []compositionDeckOp{
			op("set_speaker_notes", compositionOpPayload{SectionAnchor: anchor("second"), AuthoredTextRef: "note"}),
			op("insert_authored_block", compositionOpPayload{BlockAnchor: blockAnchor("second", "paragraph"), Position: "before", AuthoredTextRef: "one"}),
			op("insert_authored_block", compositionOpPayload{BlockAnchor: blockAnchor("second", "paragraph"), Position: "before", AuthoredTextRef: "two"}),
		}
		if err := applyCompositionToDeck(&slides, &renderComposition{DeckOps: ops, AuthoredTexts: texts}, contract, ReleaseScopeInternalDraft); err != nil {
			t.Fatal(err)
		}
		if slides[1].Blocks[0].Fields[0].Value != "One" || slides[1].Blocks[1].Fields[0].Value != "Two" {
			t.Fatal("before order")
		}
		limits, _ := resolveRenderLimits(map[string]json.RawMessage{"blocks_per_slide": json.RawMessage(`1`)})
		chunks, err := chunkDeckSlides(slides, limits)
		if err != nil {
			t.Fatal(err)
		}
		if chunks[1].SpeakerNotes == nil || !reflect.DeepEqual(chunks[1].DisclosurePartitionRefs, []string{"public", "internal_only"}) {
			t.Fatal("note disclosure absent", chunks[1].DisclosurePartitionRefs)
		}
		if chunks[2].SpeakerNotes != nil || !reflect.DeepEqual(chunks[2].DisclosurePartitionRefs, []string{"public"}) {
			t.Fatal("note disclosure leaked into another chunk")
		}
	})
	t.Run("wire anchors survive projection and chunking", func(t *testing.T) {
		input := compositionTextFixture(t, []compositionAuthoredText{}, []compositionDeckOp{})
		input.ExportModel = fixtureStructuredRenderModel(t, []ExportField{{Path: "/incident/status", Value: "A", ContentClass: ContentClassDerivedAnalytic}, {Path: "/incident/title", Value: "B", ContentClass: ContentClassDerivedAnalytic}})
		encoded, _ := canonicalJSON(input.ExportModel)
		input.ExportModelSHA256 = reportingObjectDigest(ExportModelSchemaID, encoded)
		contract.Sections[0].DeclID = "named"
		contract.DeclaredLimits = map[string]json.RawMessage{"blocks_per_slide": json.RawMessage(`1`)}
		raw := sealedReportingCompositionFixture(t, []byte(`{"schema_id":"cartulary.report_composition.v1","authored_texts":[{"authored_text_id":"title","text_role":"title_override","body":"Semantic title","disclosure_partition_ref":"public"}],"diagram_decls":[],"deck_ops":[{"op_id":"title","op_kind":"override_title","on_unresolved":"fail","payload":{"section_anchor":{"anchor_kind":"section_anchor","template_section_decl_id":"named","expansion_key":null},"authored_text_ref":"title"}}]}`), input.ExportModel.IncidentID, contract)
		input.CompositionJSON, input.TemplateContract = raw, contract
		redacted, err := redactRenderInput(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		model := redacted.Model
		deck, err := deriveDeckModel(contract, model, renderOutputOptions{}, raw)
		if err != nil || len(deck.Slides) != 2 || deck.Slides[0].Title != "Semantic title (1/2)" || deck.Slides[1].Title != "Semantic title (2/2)" {
			t.Fatal("anchor projection or chunk order", err)
		}
		broken := strings.ReplaceAll(string(raw), `"named"`, `"sld-0001"`)
		brokenRaw := sealedReportingCompositionFixture(t, []byte(broken), model.IncidentID, contract)
		if _, err := deriveDeckModel(contract, model, renderOutputOptions{}, brokenRaw); err == nil {
			t.Fatal("structural identifier became semantic target")
		}
	})
}
