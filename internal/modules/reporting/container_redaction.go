package reporting

import (
	"context"
	"sort"
	"strings"
)

// Containers carry structure, never a scalar replacement value. Their selected
// rules and excluded descendants are separate from field transformations and
// from the later removal of containers emptied by their children.
type ContainerRedactionDecision struct {
	Path                    string            `json:"export_model_path"`
	Kind                    string            `json:"element_kind"`
	ContentClass            *string           `json:"content_class"`
	Action                  string            `json:"action"`
	SelectedRuleTrace       SelectedRuleTrace `json:"selected_rule_trace"`
	DisclosurePartitionRefs []string          `json:"disclosure_partition_refs"`
	Outcome                 string            `json:"outcome"`
	SuppressedPaths         []string          `json:"suppressed_paths"`
}

type redactionTreeCandidate struct {
	path   string
	kind   string
	class  *string
	source string
	end    int // exclusive preorder subtree boundary
}

// This typed census is independent of rule selection. In particular, a drop
// never makes an otherwise invalid exact selector look like a valid fallback.
func redactionTreeCandidates(source SnapshotContent) ([]redactionTreeCandidate, error) {
	result := []redactionTreeCandidate{}
	seen := map[string]bool{}
	add := func(candidate redactionTreeCandidate) (int, error) {
		if seen[candidate.path] {
			return 0, newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
		}
		seen[candidate.path] = true
		index := len(result)
		result = append(result, candidate)
		return index, nil
	}
	var blocks func(string, string, []ReportingBlock) error
	blocks = func(parent, member string, values []ReportingBlock) error {
		for _, block := range values {
			path, err := keyedModelPath(parent, member, block.BlockID)
			if err != nil {
				return err
			}
			class := block.ContentClass
			index, err := add(redactionTreeCandidate{path: path, kind: "block", class: &class})
			if err != nil {
				return err
			}
			for _, field := range block.Fields {
				fieldPath, err := keyedModelPath(path, "fields", field.FieldKey)
				if err != nil {
					return err
				}
				if len(field.SourceRefs) == 0 {
					return newRenderValidationError("export_model_invalid", "dangling_source_ref")
				}
				fieldIndex, err := add(redactionTreeCandidate{path: fieldPath, kind: "field", source: field.SourceRefs[0]})
				if err != nil {
					return err
				}
				result[fieldIndex].end = fieldIndex + 1
			}
			if err := blocks(path, "children", block.Children); err != nil {
				return err
			}
			result[index].end = len(result)
		}
		return nil
	}
	for _, section := range source.Sections {
		path, err := keyedModelPath("$", "sections", section.SectionID)
		if err != nil {
			return nil, err
		}
		index, err := add(redactionTreeCandidate{path: path, kind: "section"})
		if err != nil {
			return nil, err
		}
		if err := blocks(path, "blocks", section.Blocks); err != nil {
			return nil, err
		}
		result[index].end = len(result)
	}
	return result, nil
}

func selectContainerRedactions(ctx context.Context, source SnapshotContent, fields []ExportField, profile RedactionProfileView) ([]ExportField, []ContainerRedactionDecision, error) {
	candidates, err := redactionTreeCandidates(source)
	if err != nil {
		return nil, nil, err
	}
	decisions := []ContainerRedactionDecision{}
	excluded := map[string]bool{}
	fieldPartitions := map[string][]string{}
	for _, field := range fields {
		fieldPartitions[field.Path] = field.DisclosurePartitionRefs
	}
	for i := 0; i < len(candidates); i++ {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		candidate := candidates[i]
		if candidate.kind == "field" {
			continue
		}
		class := ""
		if candidate.class != nil {
			class = *candidate.class
		}
		_, action, trace := selectRedactionRule(ExportField{ContentClass: class}, candidate.path, profile)
		decision := ContainerRedactionDecision{Path: candidate.path, Kind: candidate.kind, ContentClass: candidate.class,
			Action: action.Type, SelectedRuleTrace: trace, DisclosurePartitionRefs: []string{}, Outcome: "unchanged", SuppressedPaths: []string{}}
		switch action.Type {
		case ActionAllow:
			// Qualify the retained descendants after their own actions. A
			// removed child's labels cannot remain as stale container metadata.
		case ActionDrop:
			decision.Outcome = "dropped"
			for _, child := range candidates[i+1 : candidate.end] {
				decision.SuppressedPaths = append(decision.SuppressedPaths, child.path)
				if child.kind == "field" {
					excluded[child.source] = true
					decision.DisclosurePartitionRefs = append(decision.DisclosurePartitionRefs, fieldPartitions[child.source]...)
				}
			}
			decision.DisclosurePartitionRefs = orderedDisclosurePartitions(decision.DisclosurePartitionRefs)
			sort.Strings(decision.SuppressedPaths)
			i = candidate.end - 1
		default:
			// A container cannot become a scalar while retaining its declared
			// schema. Never stringify it or silently reinterpret the action.
			return nil, nil, newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
		}
		decisions = append(decisions, decision)
	}
	retained := make([]ExportField, 0, len(fields))
	for _, field := range fields {
		if !excluded[field.Path] {
			retained = append(retained, field)
		}
	}
	sort.Slice(decisions, func(i, j int) bool { return decisions[i].Path < decisions[j].Path })
	return retained, decisions, nil
}

func qualifyRetainedContainers(decisions []ContainerRedactionDecision, fields []RedactedField) {
	byPath := map[string]int{}
	for i, decision := range decisions {
		byPath[decision.Path] = i
	}
	for _, field := range fields {
		steps, _ := parseExactModelPath(field.Path)
		for end := len(steps) - 2; end >= 2; end -= 2 {
			if index, exists := byPath[formatExactModelPath(steps[:end])]; exists {
				decisions[index].DisclosurePartitionRefs = append(decisions[index].DisclosurePartitionRefs, field.DisclosurePartitionRefs...)
			}
		}
	}
	for i := range decisions {
		decisions[i].DisclosurePartitionRefs = orderedDisclosurePartitions(decisions[i].DisclosurePartitionRefs)
	}
}

func explicitlyRetainedTree(source SnapshotContent, decisions []ContainerRedactionDecision) SnapshotContent {
	dropped := map[string]bool{}
	for _, decision := range decisions {
		dropped[decision.Path] = decision.Outcome == "dropped"
	}
	var blocks func(string, string, []ReportingBlock) []ReportingBlock
	blocks = func(parent, member string, values []ReportingBlock) []ReportingBlock {
		retained := []ReportingBlock{}
		for _, value := range values {
			path, _ := keyedModelPath(parent, member, value.BlockID)
			if dropped[path] {
				continue
			}
			value.Children = blocks(path, "children", value.Children)
			retained = append(retained, value)
		}
		return retained
	}
	sections := []ReportingSection{}
	for _, section := range source.Sections {
		path, _ := keyedModelPath("$", "sections", section.SectionID)
		if dropped[path] {
			continue
		}
		section.Blocks = blocks(path, "blocks", section.Blocks)
		sections = append(sections, section)
	}
	source.Sections = sections
	return source
}

func structuralDecisionEntries(entries []RedactionManifestEntry, containers []ContainerRedactionDecision) []RedactionManifestEntry {
	result := append([]RedactionManifestEntry{}, entries...)
	for _, decision := range containers {
		result = append(result, RedactionManifestEntry{Path: decision.Path, Outcome: decision.Outcome})
	}
	return result
}

func isModelDescendant(path, parent string) bool { return strings.HasPrefix(path, parent+".") }
