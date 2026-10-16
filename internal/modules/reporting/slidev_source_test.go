package reporting

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestSlidevCanonicalSourceV2(t *testing.T) {
	t.Run("manual and embedded diagrams", testManualDiagramLayout)
	t.Run("template manifest and selection", testTemplateManifestContract)
	t.Run("template section expansion", testTemplateSectionExpansion)
	t.Run("template narrative slots", testTemplateNarrativeSlots)
	t.Run("template field bindings", testTemplateFieldBindings)
	t.Run("template reachability", testTemplateReachability)
	t.Run("canonical render model assembly", testCanonicalRenderModelAssembly)
	t.Run("typed blocks", testTypedSlidevBlocks)
	t.Run("click state and layout", testSlidevClickState)
	// This authored contract vector is also consumed by the real Slidev parser.
	// It contains synthetic product input, never repository documentation.
	raw, err := os.ReadFile("../../../contracts/reporting/fixtures/slidev-source.v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Deck        reportingDeckModel `json:"deck"`
		Template    TemplateContract   `json:"template"`
		SourceLines []string           `json:"source_lines"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	got, err := serializeSlidevMarkdown(fixture.Deck, fixture.Template)
	if err != nil || string(got) != strings.Join(fixture.SourceLines, "\n") {
		t.Fatalf("canonical source differs: %q, error=%v", got, err)
	}
	for _, mutate := range []func(*reportingDeckModel){
		func(d *reportingDeckModel) { d.Slides = nil },
		func(d *reportingDeckModel) { d.Title = "bad\nheadmatter" },
		func(d *reportingDeckModel) { d.Slides[0].Layout = "default\nsrc: file" },
		func(d *reportingDeckModel) { d.Slides[0].Blocks[0].BlockKind = "future" },
		func(d *reportingDeckModel) { d.Slides[0].Blocks[0].BlockKind = "field_summary" },
	} {
		var candidate struct {
			Deck reportingDeckModel `json:"deck"`
		}
		if err := json.Unmarshal(raw, &candidate); err != nil {
			t.Fatal(err)
		}
		mutate(&candidate.Deck)
		got, err := serializeSlidevMarkdown(candidate.Deck, fixture.Template)
		assertSlidevSourceFailure(t, got, err, "invalid_slidev_construct")
	}
	for _, template := range []TemplateContract{
		{Author: "bad\u0085author", Theme: "default"},
		{Theme: "https://example.invalid/theme"},
	} {
		got, err := serializeSlidevMarkdown(fixture.Deck, template)
		assertSlidevSourceFailure(t, got, err, "invalid_slidev_construct")
	}
	if _, ok := ResolveTemplateContract(DefaultTemplateID, "1"); ok {
		t.Fatal("retired template resolved through a compatibility fallback")
	}
	contract, ok := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	if !ok || contract.Theme != "default" || contract.Author != "" {
		t.Fatal("current template lacks explicit renderer metadata")
	}
}
