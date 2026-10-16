package reporting

import (
	"errors"
	"slices"
	"sort"
)

// Structural removal records the consequence of field/child removal. It is
// deliberately separate from a selected Core rule: an empty container did not
// select a fictitious drop action or acquire its child's rule identity.
type StructuralRedactionRemoval struct {
	Path        string   `json:"export_model_path"`
	Reason      string   `json:"removal_reason"`
	CausalPaths []string `json:"causal_paths"`
}

func structuralRedactionRemovals(model SnapshotContent, decisions []RedactionManifestEntry) ([]StructuralRedactionRemoval, error) {
	invalid := newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
	dropped := map[string]bool{}
	for _, entry := range decisions {
		if _, exists := dropped[entry.Path]; exists {
			return nil, invalid
		}
		dropped[entry.Path] = entry.Outcome == "dropped"
	}
	removals := []StructuralRedactionRemoval{}
	seen := map[string]bool{}
	var block func(string, ReportingBlock) (bool, error)
	block = func(path string, value ReportingBlock) (bool, error) {
		if seen[path] {
			return false, invalid
		}
		seen[path] = true
		if dropped[path] {
			return true, nil
		}
		causes := []string{}
		retained := false
		for _, field := range value.Fields {
			fieldPath, err := keyedModelPath(path, "fields", field.FieldKey)
			if err != nil {
				return false, err
			}
			removed, exists := dropped[fieldPath]
			if !exists {
				return false, invalid
			}
			if removed {
				causes = append(causes, fieldPath)
			} else {
				retained = true
			}
		}
		for _, child := range value.Children {
			childPath, err := keyedModelPath(path, "children", child.BlockID)
			if err != nil {
				return false, err
			}
			removed, err := block(childPath, child)
			if err != nil {
				return false, err
			}
			if removed {
				causes = append(causes, childPath)
			} else {
				retained = true
			}
		}
		if !retained {
			sort.Strings(causes)
			removals = append(removals, StructuralRedactionRemoval{Path: path, Reason: "empty_block", CausalPaths: causes})
		}
		return !retained, nil
	}
	for _, section := range model.Sections {
		path, err := keyedModelPath("$", "sections", section.SectionID)
		if err != nil || seen[path] {
			return nil, invalid
		}
		seen[path] = true
		if dropped[path] {
			continue
		}
		causes := []string{}
		retained := false
		for _, item := range section.Blocks {
			blockPath, err := keyedModelPath(path, "blocks", item.BlockID)
			if err != nil {
				return nil, err
			}
			removed, err := block(blockPath, item)
			if err != nil {
				return nil, err
			}
			if removed {
				causes = append(causes, blockPath)
			} else {
				retained = true
			}
		}
		if !retained && section.SectionKind != "validation" {
			sort.Strings(causes)
			removals = append(removals, StructuralRedactionRemoval{Path: path, Reason: "empty_section", CausalPaths: causes})
		}
	}
	sort.Slice(removals, func(i, j int) bool { return removals[i].Path < removals[j].Path })
	return removals, nil
}

func modelContainerKind(steps []modelPathStep) string {
	if len(steps) < 2 || len(steps)%2 != 0 || steps[0] != (modelPathStep{Value: "sections"}) || !steps[1].Key {
		return ""
	}
	for i := 2; i < len(steps); i += 2 {
		member := "children"
		if i == 2 {
			member = "blocks"
		}
		if steps[i] != (modelPathStep{Value: member}) || !steps[i+1].Key {
			return ""
		}
	}
	if len(steps) == 2 {
		return "section"
	}
	return "block"
}

func validateStructuralRemovals(entries []RedactionManifestEntry, removals []StructuralRedactionRemoval) error {
	invalid := errors.New("reporting: invalid structural removal evidence")
	if removals == nil {
		return invalid
	}
	decisions := map[string]string{}
	removed := map[string]bool{}
	children := map[string]map[string]bool{}
	indexChild := func(path string) {
		steps, valid := parseExactModelPath(path)
		if !valid || len(steps) < 2 || steps[len(steps)-2].Key || !steps[len(steps)-1].Key {
			return
		}
		parent := formatExactModelPath(steps[:len(steps)-2])
		if children[parent] == nil {
			children[parent] = map[string]bool{}
		}
		children[parent][path] = true
	}
	for _, entry := range entries {
		decisions[entry.Path] = entry.Outcome
		indexChild(entry.Path)
	}
	previous := ""
	for _, removal := range removals {
		steps, valid := parseExactModelPath(removal.Path)
		kind := modelContainerKind(steps)
		if !valid || removal.Path <= previous || decisions[removal.Path] == "dropped" ||
			(kind != "section" || removal.Reason != "empty_section") && (kind != "block" || removal.Reason != "empty_block") ||
			!orderedReportingRefs(removal.CausalPaths) {
			return invalid
		}
		previous = removal.Path
		removed[removal.Path] = true
		indexChild(removal.Path)
	}
	for _, removal := range removals {
		// All published immediate children are known evidence. An omitted
		// known cause, or any surviving child, contradicts emptiness even
		// without the private original inventory available to native replay.
		if len(children[removal.Path]) != len(removal.CausalPaths) {
			return invalid
		}
		parent, _ := parseExactModelPath(removal.Path)
		for _, cause := range removal.CausalPaths {
			child, valid := parseExactModelPath(cause)
			if !valid || len(child) != len(parent)+2 || !slices.Equal(parent, child[:len(parent)]) || child[len(parent)].Key || !child[len(parent)+1].Key {
				return invalid
			}
			member := child[len(parent)].Value
			switch removal.Reason {
			case "empty_section":
				if member != "blocks" || (!removed[cause] && decisions[cause] != "dropped") {
					return invalid
				}
			case "empty_block":
				if (member != "fields" || decisions[cause] != "dropped") && (member != "children" || (!removed[cause] && decisions[cause] != "dropped")) {
					return invalid
				}
			}
		}
	}
	return nil
}
