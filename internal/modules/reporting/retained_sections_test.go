package reporting

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func testRetainedSections(t *testing.T) {
	t.Run("mixed partition structure", testRetainedMixedContent)
	model := fixtureStructuredRenderModel(t, []ExportField{
		{Path: "/incident/a", ContentClass: ContentClassSourceEvidence, Value: "removed"},
		{Path: "/incident/b", ContentClass: ContentClassDerivedAnalytic, Value: []any{"retained", true}},
		{Path: "/incident/c", ContentClass: ContentClassSourceEvidence, Value: "masked original"},
	})
	before, _ := canonicalJSON(model)
	paths := fixtureModelPaths(t, model.SnapshotContent, "/incident/a", "/incident/b", "/incident/c")
	drop, mask := paths["/incident/a"], paths["/incident/c"]
	profile, _, err := ResolveRedactionProfile(InternalRedactionProfileID, DefaultRedactionProfileVersion, nil, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	profile.Rules = []RedactionRule{{RuleID: "drop", Path: &drop, Action: RedactionActionSpec{Type: ActionDrop}},
		{RuleID: "mask", Path: &mask, Action: RedactionActionSpec{Type: ActionMask}}}
	digest, err := ValidateRedactionProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	result, err := RedactExportModel(model, fixtureSourceFactsForContent(model.SnapshotContent), profile, digest, "", ReleaseScopeInternalReview, nil)
	if err != nil {
		t.Fatal(err)
	}
	sections := result.Model.Sections
	if len(sections) != 1 || len(sections[0].Blocks) != 2 || sections[0].ContentClassSummary.SourceEvidence != 1 || sections[0].ContentClassSummary.DerivedAnalytic != 1 {
		t.Fatal("retained section summary includes removed content")
	}
	for index, block := range sections[0].Blocks {
		if block.BlockID != fmt.Sprintf("sec-0001.b%04d", index+1) || block.BlockOrdinal != index+1 || block.ParentBlockID != nil || block.Fields[0].FieldOrdinal != 1 {
			t.Fatal("retained ordinals/identities were not rebuilt")
		}
	}
	if sections[0].Blocks[1].Fields[0].Value != DefaultMaskText || sections[0].Blocks[1].Fields[0].RedactedValueState != "masked" {
		t.Fatal("projection restored a raw source value")
	}
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	// The old decision index cannot inject content into deck derivation.
	result.Model.Fields = []RedactedField{{Value: "unadmitted index content"}}
	deck, err := fixtureDerivedDeck(contract, result.Model, renderOutputOptions{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(deck.Slides) != 1 || len(deck.Slides[0].Blocks) != 2 || !reflect.DeepEqual(deck.Slides[0].Blocks[0].Fields[0].Value, []any{"retained", true}) || deck.Slides[0].Blocks[1].Fields[0].Value != DefaultMaskText || deck.Slides[0].Title != sections[0].Title {
		t.Fatal("deck did not consume the retained section tree")
	}
	sections[0].Blocks[0].Fields[0].Value.([]any)[0] = "mutation"
	after, _ := canonicalJSON(model)
	if !bytes.Equal(before, after) {
		t.Fatal("retained projection aliases immutable source")
	}

	field := ReportingField{FieldKey: "value", SourceRefs: []string{"source"}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{"public"}, Value: "kept"}
	nested := []ReportingSection{{SectionID: "old", OrderingKey: "0002", SectionKind: "appendix", Blocks: []ReportingBlock{
		{BlockID: "drop", BlockKind: "paragraph", Fields: []ReportingField{{SourceRefs: []string{"removed"}}}},
		{BlockID: "parent", BlockKind: "paragraph", ContentClass: ContentClassDerivedAnalytic, Children: []ReportingBlock{
			{BlockID: "child-drop", BlockKind: "paragraph", Fields: []ReportingField{{SourceRefs: []string{"removed"}}}},
			{BlockID: "child-keep", BlockKind: "paragraph", ContentClass: ContentClassSourceEvidence, Fields: []ReportingField{field}},
		}},
	}}, {SectionID: "empty", OrderingKey: "0001", SectionKind: "appendix"}, {SectionID: "validation", OrderingKey: "9999.0001", SectionKind: "validation"}}
	projected, err := projectRetainedSections(context.Background(), nested, map[string]ReportingField{"source": field})
	if err != nil || len(projected) != 2 || projected[0].SectionID != "sec-0002" || len(projected[0].Blocks) != 1 || projected[1].SectionKind != "validation" {
		t.Fatal("empty section/container removal or ordering differs", err)
	}
	parent := projected[0].Blocks[0]
	child := parent.Children[0]
	if parent.BlockID != "sec-0002.b0001" || child.BlockID != "sec-0002.b0001.b0001" || child.ParentBlockID == nil || *child.ParentBlockID != parent.BlockID || child.BlockOrdinal != 1 || child.Fields[0].FieldOrdinal != 1 {
		t.Fatal("nested retained identities differ")
	}
	if projected[0].ContentClassSummary.SourceEvidence != 1 || projected[0].ContentClassSummary.DerivedAnalytic != 1 || len(parent.SourceRefs) != 1 || len(projected[0].SourceRefs) != 1 {
		t.Fatal("container summaries did not follow retained children")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err := projectRetainedSections(ctx, nested, nil); !errors.Is(err, context.Canceled) || result != nil {
		t.Fatal("canceled projection published partial sections")
	}
}

// These fixtures enter at the renderer boundary intentionally; malformed values
// exercise serializer admission without passing earlier source admission first.
func fixtureRetainedParagraphSections(fields []RedactedField) []ReportingSection {
	section := ReportingSection{SectionID: "sec-0001", OrderingKey: "0001", SectionKind: "appendix", Title: "Snapshot export", Blocks: []ReportingBlock{}}
	for i, field := range fields {
		section.Blocks = append(section.Blocks, ReportingBlock{BlockID: fmt.Sprintf("sec-0001.b%04d", i+1), BlockKind: "paragraph", BlockOrdinal: i + 1, ContentClass: field.ContentClass,
			Fields: []ReportingField{{FieldKey: "value", FieldOrdinal: 1, Value: field.Value, SourceRefs: []string{fmt.Sprintf("fixture-source-%04d", i+1)}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{"public"}}}, Children: []ReportingBlock{}})
	}
	return []ReportingSection{section}
}

func bindFixtureRetainedSections(model *RedactedExportModel) {
	model.Sections = fixtureRetainedParagraphSections(model.Fields)
	paths, err := redactionFieldModelPaths(SnapshotContent{Sections: model.Sections})
	if err != nil {
		panic(err)
	}
	model.admittedFieldPaths = paths
}
