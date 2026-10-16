package reporting

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func testSlidevClickState(t *testing.T) {
	contract, _ := ResolveTemplateContract(DefaultTemplateID, DefaultTemplateVersion)
	limits, err := resolveRenderLimits(nil)
	if err != nil {
		t.Fatal(err)
	}
	paragraph := func(id string) ReportingBlock {
		return ReportingBlock{BlockID: id, BlockKind: "paragraph", Fields: []ReportingField{{FieldKey: "text", Value: id}}}
	}
	slide := reportingDeckSlide{SlideID: "sld-0001", Title: "Blocks", Layout: "default", Blocks: []ReportingBlock{paragraph("b1"), paragraph("b2"), paragraph("b3")}, clickProfile: "reveal_blocks"}
	if err := materializeSlideClicks(&slide, limits); err != nil {
		t.Fatal(err)
	}
	if len(slide.ClickSteps) != 2 || slide.ClickSteps[0].Targets[0] != "b2" || slide.ClickSteps[1].Targets[0] != "b3" {
		t.Fatal("wrong reveal targets")
	}
	withNotes := slide
	withNotes.Blocks = append([]ReportingBlock{{BlockID: "note", BlockKind: "speaker_note", Fields: []ReportingField{{Value: "Private notes"}}}}, slide.Blocks...)
	if err := materializeSlideClicks(&withNotes, limits); err != nil || !reflect.DeepEqual(withNotes.ClickSteps, slide.ClickSteps) {
		t.Fatal("notes consumed a body reveal", err)
	}
	// Literal independent canonical preimages ensure the state proof contains
	// every target, including later targets that remain hidden.
	for i, states := range []string{`[{"target_id":"b2","visible":true},{"target_id":"b3","visible":false}]`, `[{"target_id":"b2","visible":true},{"target_id":"b3","visible":true}]`} {
		preimage := `cartulary.click_state.v1` + "\n" + `{"schema_id":"cartulary.click_state.v1","slide_id":"sld-0001","states":` + states + `}`
		if slide.ClickSteps[i].ResultingStateHash != hashHex([]byte(preimage)) {
			t.Fatal("state hash preimage differs")
		}
	}
	deck := reportingDeckModel{Title: "Clicks", Slides: []reportingDeckSlide{slide}}
	source, err := serializeSlidevMarkdown(deck, contract)
	if err != nil || strings.Count(string(source), "<v-click at=") != 2 || !strings.Contains(string(source), "<v-click at=\"2\">\n\nb3\n\n</v-click>") {
		t.Fatal("block wrapper", err, string(source))
	}
	encoded, _ := json.Marshal(slide)
	for name, mutate := range map[string]func(*reportingDeckSlide){
		"state hash":     func(s *reportingDeckSlide) { s.ClickSteps[0].ResultingStateHash = strings.Repeat("0", 64) },
		"missing step":   func(s *reportingDeckSlide) { s.ClickSteps = s.ClickSteps[:1] },
		"reused target":  func(s *reportingDeckSlide) { s.ClickSteps[1].Targets[0] = "b2" },
		"foreign target": func(s *reportingDeckSlide) { s.ClickSteps[0].Targets[0] = "other" },
		"at":             func(s *reportingDeckSlide) { s.ClickSteps[0].At = 2 },
		"ordinal":        func(s *reportingDeckSlide) { s.ClickSteps[0].Ordinal = 0 },
		"component":      func(s *reportingDeckSlide) { s.ClickSteps[0].Component = "v-after" },
		"action":         func(s *reportingDeckSlide) { s.ClickSteps[0].Action = "hide" },
		"visibility":     func(s *reportingDeckSlide) { s.ClickSteps[0].InitialVisibility = "visible" },
		"changed slide":  func(s *reportingDeckSlide) { s.SlideID = "sld-0002" },
	} {
		t.Run(name, func(t *testing.T) {
			var bad reportingDeckSlide
			if json.Unmarshal(encoded, &bad) != nil {
				t.Fatal("fixture")
			}
			mutate(&bad)
			data, err := serializeSlidevMarkdown(reportingDeckModel{Title: "Clicks", Slides: []reportingDeckSlide{bad}}, contract)
			if err == nil || len(data) != 0 {
				t.Fatal("unbound click source emitted")
			}
		})
	}
	list := ReportingBlock{BlockID: "list", BlockKind: "bullet_list", Children: []ReportingBlock{paragraph("a"), paragraph("b"), paragraph("c")}}
	slide.Blocks = []ReportingBlock{list}
	slide.clickProfile = "reveal_list_items"
	if err := materializeSlideClicks(&slide, limits); err != nil || len(slide.ClickSteps) != 3 {
		t.Fatal("list counts", err)
	}
	source, err = serializeSlidevMarkdown(reportingDeckModel{Title: "List", Slides: []reportingDeckSlide{slide}}, contract)
	if err != nil || strings.Count(string(source), "<v-clicks at=") != 1 || !strings.Contains(string(source), "<v-clicks at=\"1\">\n\n- a\n- b\n- c\n\n</v-clicks>") {
		t.Fatal("list group", err, string(source))
	}
	limits["click_steps_per_slide"] = 2
	if materializeSlideClicks(&slide, limits) == nil {
		t.Fatal("click bound bypass")
	}
	contract.Layouts = append(contract.Layouts, TemplateLayout{LayoutID: "evidence", SlidevLayout: "section", AllowedBlockKinds: []string{"bullet_list", "paragraph"}})
	slide.Layout = "evidence"
	before, _ := json.Marshal(slide)
	source, err = serializeSlidevMarkdown(reportingDeckModel{Title: "Mapped", Slides: []reportingDeckSlide{slide}}, contract)
	if err != nil || !strings.Contains(string(source), `layout: "section"`) {
		t.Fatal("logical layout did not map", err)
	}
	after, _ := json.Marshal(slide)
	if !reflect.DeepEqual(before, after) || slide.Layout != "evidence" {
		t.Fatal("serialization changed logical layout")
	}
	contract.Layouts[len(contract.Layouts)-1].AllowedBlockKinds = []string{"bullet_list"}
	if _, err := resolveDeckLayout(contract, slide); err == nil {
		t.Fatal("nested forbidden block bypassed layout")
	}
}
