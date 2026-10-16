package reporting

import (
	"fmt"
	"reflect"
	"sort"
)

type slideClickState struct {
	TargetID string `json:"target_id"`
	Visible  bool   `json:"visible"`
}

// Click identities and proofs are derived only after final chunk identities.
// The list profile reveals each direct list item, matching pinned VClicks.
func materializeSlideClicks(slide *reportingDeckSlide, limits renderLimits) error {
	steps := []reportingClickStep{}
	add := func(id, component string) {
		n := len(steps) + 1
		steps = append(steps, reportingClickStep{ClickStepID: fmt.Sprintf("clk-%04d", n), Ordinal: n, Action: "reveal", Targets: []string{id}, InitialVisibility: "hidden", Component: component, At: n})
	}
	switch slide.clickProfile {
	case "", "none":
	case "reveal_blocks":
		bodyBlocks := 0
		for _, block := range slide.Blocks {
			if block.BlockKind == "speaker_note" {
				continue
			}
			bodyBlocks++
			if bodyBlocks == 1 {
				continue
			}
			add(block.BlockID, "v-click")
		}
	case "reveal_list_items":
		for _, block := range slide.Blocks {
			if block.BlockKind == "bullet_list" {
				for _, child := range block.Children {
					add(child.BlockID, "v-clicks")
				}
			}
		}
	default:
		return invalidSlidevBlock()
	}
	if err := limits.check("click_steps_per_slide", int64(len(steps))); err != nil {
		return err
	}
	ids := map[string]bool{}
	for _, step := range steps {
		id := step.Targets[0]
		if !validReportingIdentifier(id) || ids[id] {
			return invalidSlidevBlock()
		}
		ids[id] = true
	}
	states := make([]slideClickState, 0, len(ids))
	for id := range ids {
		states = append(states, slideClickState{TargetID: id})
	}
	sort.Slice(states, func(i, j int) bool { return states[i].TargetID < states[j].TargetID })
	for i := range steps {
		for j := range states {
			if states[j].TargetID == steps[i].Targets[0] {
				states[j].Visible = true
			}
		}
		raw, err := canonicalJSON(struct {
			SchemaID string            `json:"schema_id"`
			SlideID  string            `json:"slide_id"`
			States   []slideClickState `json:"states"`
		}{"cartulary.click_state.v1", slide.SlideID, states})
		if err != nil {
			return err
		}
		steps[i].ResultingStateHash = reportingObjectDigest("cartulary.click_state.v1", raw)
	}
	slide.ClickSteps = steps
	return nil
}

// Validate serialized decks independently of private derivation metadata.
// Recompute the complete allowed profile, including all state hashes, before
// using any step to emit a component. Caller data never supplies HTML syntax.
func validatedClickWrappers(slide reportingDeckSlide, limits renderLimits) (map[string]reportingClickStep, error) {
	wrappers := map[string]reportingClickStep{}
	if len(slide.ClickSteps) == 0 {
		return wrappers, nil
	}
	expected := slide
	switch slide.ClickSteps[0].Component {
	case "v-click":
		expected.clickProfile = "reveal_blocks"
	case "v-clicks":
		expected.clickProfile = "reveal_list_items"
	default:
		return nil, invalidSlidevBlock()
	}
	if err := materializeSlideClicks(&expected, limits); err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(expected.ClickSteps, slide.ClickSteps) {
		return nil, invalidSlidevBlock()
	}
	if expected.clickProfile == "reveal_blocks" {
		for _, step := range expected.ClickSteps {
			wrappers[step.Targets[0]] = step
		}
		return wrappers, nil
	}
	stepIndex := 0
	for _, block := range slide.Blocks {
		if block.BlockKind != "bullet_list" || len(block.Children) == 0 {
			continue
		}
		wrappers[block.BlockID] = expected.ClickSteps[stepIndex]
		stepIndex += len(block.Children)
	}
	return wrappers, nil
}
