package reporting

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func fixtureTextBlock(kind string, values ...any) ReportingBlock {
	fields := make([]ReportingField, len(values))
	for i, value := range values {
		fields[i] = ReportingField{FieldKey: "value", FieldOrdinal: i + 1, Value: value}
	}
	return ReportingBlock{BlockKind: kind, Fields: fields, Children: []ReportingBlock{}}
}

func testTypedSlidevBlocks(t *testing.T) {
	paragraph := fixtureTextBlock("paragraph", "Literal <script> | {{ text }}", []any{json.Number("12"), true, nil})
	metric := fixtureTextBlock("metric", json.Number("12"))
	metric.Fields[0].DisplayLabel = fixtureNote("Open | items")
	first := fixtureTextBlock("paragraph", "row | one", true)
	first.Fields[0].FieldKey, first.Fields[1].FieldKey = "name", "active"
	first.Fields[0].DisplayLabel = fixtureNote("Name | label")
	second := fixtureTextBlock("paragraph", "row two", nil)
	second.Fields[0].FieldKey, second.Fields[1].FieldKey = "name", "active"
	table := ReportingBlock{BlockKind: "table", Children: []ReportingBlock{first, second}}
	list := ReportingBlock{BlockKind: "bullet_list", Children: []ReportingBlock{fixtureTextBlock("paragraph", "first\ncontinued"), fixtureTextBlock("paragraph", "1. literal")}}
	nested := ReportingBlock{BlockKind: "bullet_list", Children: []ReportingBlock{fixtureTextBlock("paragraph", "nested")}}
	list.Children[0].Children = []ReportingBlock{nested}
	for _, test := range []struct {
		name  string
		block ReportingBlock
		want  string
	}{
		{"paragraph", paragraph, "Literal \\<script> \\| {{ text }}\n12, true, "},
		{"metric", metric, "**Open \\| items:** 12"},
		{"table", table, "| Name \\| label | active |\n| --- | --- |\n| row \\| one | true |\n| row two |  |"},
		{"list", list, "- first\n  continued\n  - nested\n- 1\\. literal"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := serializeSlidevBlock(test.block, 0)
			if err != nil || got != test.want {
				t.Fatalf("typed block serialization: %q, %v", got, err)
			}
		})
	}
	for _, block := range []ReportingBlock{
		{BlockKind: "table", Children: []ReportingBlock{first, fixtureTextBlock("paragraph", "short")}},
		{BlockKind: "table", Children: []ReportingBlock{first, fixtureTextBlock("paragraph", "wrong", false)}},
		{BlockKind: "table", Fields: []ReportingField{{Value: "unowned"}}, Children: []ReportingBlock{first}},
		{BlockKind: "table"},
		{BlockKind: "paragraph", Fields: paragraph.Fields, Children: []ReportingBlock{nested}},
		{BlockKind: "bullet_list", Children: []ReportingBlock{nested}},
		{BlockKind: "metric", Fields: metric.Fields, Children: []ReportingBlock{first}},
		{BlockKind: "asset_ref", Fields: []ReportingField{{Value: "https://example.invalid/raw"}}},
	} {
		data, err := serializeSlidevBlock(block, 0)
		assertSlidevSourceFailure(t, []byte(data), err, "invalid_slidev_construct")
	}
	for _, kind := range []string{"metric", "table"} {
		block := fixtureTextBlock(kind, "line\nbreak")
		if kind == "table" {
			block = ReportingBlock{BlockKind: kind, Children: []ReportingBlock{fixtureTextBlock("paragraph", "line\nbreak")}}
		}
		data, err := serializeSlidevBlock(block, 0)
		assertSlidevSourceFailure(t, []byte(data), err, "invalid_field_value_for_text")
	}
	deep := nested
	for depth := 2; depth <= 4; depth++ {
		item := fixtureTextBlock("paragraph", "item")
		item.Children = []ReportingBlock{deep}
		deep = ReportingBlock{BlockKind: "bullet_list", Children: []ReportingBlock{item}}
		data, err := serializeSlidevBlock(deep, 0)
		if depth == 4 {
			assertSlidevSourceFailure(t, []byte(data), err, "list_nesting_exceeded")
		} else if err != nil {
			t.Fatal("admitted nesting rejected", err)
		}
	}
	deck := reportingDeckModel{Title: "Typed", Slides: []reportingDeckSlide{{Title: "Slide", Layout: "default", SpeakerNotes: fixtureNote("whole slide"), Blocks: []ReportingBlock{
		paragraph, fixtureTextBlock("speaker_note", "first note"), metric, fixtureTextBlock("speaker_note", "second note"),
	}}}}
	template, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	data, err := serializeSlidevMarkdown(deck, template)
	if err != nil || strings.Count(string(data), "<!--") != 1 || !strings.HasSuffix(string(data), "<!--\nwhole slide\nfirst note\nsecond note\n-->\n") {
		t.Fatal("speaker-note blocks were rendered into the body or reordered", err)
	}
	deck.Slides[0].Blocks[3].Fields[0].Value = "-->"
	data, err = serializeSlidevMarkdown(deck, template)
	assertSlidevSourceFailure(t, data, err, "invalid_speaker_note_sequence")
	testTypedDeckProjection(t)
}

func testTypedDeckProjection(t *testing.T) {
	field := ReportingField{FieldKey: "value", FieldOrdinal: 1, Value: []any{"safe"}, SourceRefs: []string{"source-a"}, SupportRefs: []string{"support-a"}, DisclosurePartitionRefs: []string{"public"}}
	block := ReportingBlock{BlockID: "sec-0001.b0001", BlockOrdinal: 1, BlockKind: "paragraph", ContentClass: ContentClassSourceEvidence, Fields: []ReportingField{field}, SourceRefs: field.SourceRefs, SupportRefs: field.SupportRefs, DisclosurePartitionRefs: field.DisclosurePartitionRefs}
	model := fixtureStructuredRenderModel(t, []ExportField{{Path: "/incident/title", ContentClass: ContentClassSourceEvidence, Value: "safe"}})
	redacted := RedactedExportModel{ExportModel: ExportModel{RenderIdentity: model.RenderIdentity, SnapshotContent: SnapshotContent{SnapshotID: model.SnapshotID, Sections: []ReportingSection{{SectionID: "sec-0001", OrderingKey: "0001", SectionKind: "appendix", Title: "Retained", Blocks: []ReportingBlock{block}}}}}, admittedFieldPaths: map[string]string{"source-a": "$.sections['sec-0001'].blocks['sec-0001.b0001'].fields['value']"}}
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	options := renderOutputOptions{SchemaID: "cartulary.reporting_render_request_options.v1", SourceOnly: true}
	deck, err := fixtureDerivedDeck(contract, redacted, options, nil)
	if err != nil {
		t.Fatal(err)
	}
	if deck.OutputOptions != options || deck.SlideCount != 1 || deck.ClickStepCount != 0 || deck.ExpectedPageCount != 1 || deck.Slides[0].Layout != "cover" || deck.Slides[0].ClickSteps == nil || deck.Slides[0].SpeakerNotes != nil || !reflect.DeepEqual(deck.SourceRefs, field.SourceRefs) || !reflect.DeepEqual(deck.Slides[0].DisclosurePartitionRefs, field.DisclosurePartitionRefs) {
		t.Fatal("deck metadata differs from admitted input")
	}
	raw, err := canonicalJSON(deck)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	for _, retired := range []string{"derivation_algorithm", "composition_applied", "composition_summary"} {
		if _, ok := object[retired]; ok {
			t.Fatal("private implementation metadata remained in canonical deck", retired)
		}
	}
	var slides []map[string]json.RawMessage
	if err := json.Unmarshal(object["slides"], &slides); err != nil {
		t.Fatal(err)
	}
	if len(slides[0]) != 10 || string(slides[0]["slide_ordinal"]) != "1" || string(slides[0]["speaker_notes"]) != "null" || string(slides[0]["click_steps"]) != "[]" {
		t.Fatal("slide wire contract differs")
	}
	sourceOnly, err := serializeSlidevMarkdown(deck, contract)
	if err != nil {
		t.Fatal(err)
	}
	pdfOptions, err := parseRenderOutputOptions(fixtureOutputOptions(OutputKindSlidev, ReleaseScopeInternalReview), OutputKindSlidev, ReleaseScopeInternalReview)
	if err != nil {
		t.Fatal(err)
	}
	pdfDeck, err := fixtureDerivedDeck(contract, redacted, pdfOptions, nil)
	if err != nil {
		t.Fatal(err)
	}
	pdfSource, err := serializeSlidevMarkdown(pdfDeck, contract)
	if err != nil || string(pdfSource) != string(sourceOnly) || !strings.Contains(string(pdfSource), "format: \"pdf\"") {
		t.Fatal("supported output selection changed canonical Slidev source", err)
	}
	deck.Slides[0].Blocks[0].Fields[0].Value.([]any)[0] = "changed"
	deck.Slides[0].SourceRefs[0] = "changed"
	if field.Value.([]any)[0] != "safe" || block.SourceRefs[0] != "source-a" {
		t.Fatal("deck aliases immutable model")
	}
}
