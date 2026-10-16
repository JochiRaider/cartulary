package reporting

import (
	"errors"
	"slices"
)

func validateContainerDecisions(manifest RedactionManifest) error {
	invalid := errors.New("reporting: invalid container redaction evidence")
	if manifest.ContainerDecisions == nil {
		return invalid
	}
	owners := map[string]string{}
	containers := map[string]ContainerRedactionDecision{}
	for _, entry := range manifest.Entries {
		owners[entry.Path] = "field"
	}
	previous := ""
	for _, decision := range manifest.ContainerDecisions {
		steps, valid := parseExactModelPath(decision.Path)
		kind := modelContainerKind(steps)
		if !valid || kind == "" || kind != decision.Kind || decision.Path <= previous || owners[decision.Path] != "" ||
			decision.DisclosurePartitionRefs == nil || validateDisclosurePartitionSet(decision.DisclosurePartitionRefs) != nil ||
			!sameReportingValue(decision.DisclosurePartitionRefs, orderedDisclosurePartitions(decision.DisclosurePartitionRefs)) ||
			!orderedReportingRefs(decision.SuppressedPaths) || !validSelectedRuleTrace(decision.SelectedRuleTrace) {
			return invalid
		}
		previous = decision.Path
		if (kind == "section" && decision.ContentClass != nil) ||
			(kind == "block" && (decision.ContentClass == nil || !validReportingContentClass(*decision.ContentClass))) {
			return invalid
		}
		trace := decision.SelectedRuleTrace
		if trace.SelectorKind == "path" && (trace.SelectorValue == nil || *trace.SelectorValue != decision.Path) {
			return invalid
		}
		if trace.SelectorKind == "content_class" && (decision.ContentClass == nil || trace.SelectorValue == nil || *trace.SelectorValue != *decision.ContentClass) {
			return invalid
		}
		switch decision.Action {
		case ActionAllow:
			if decision.Outcome != "unchanged" || len(decision.SuppressedPaths) != 0 {
				return invalid
			}
		case ActionDrop:
			if decision.Outcome != "dropped" {
				return invalid
			}
		default:
			return invalid
		}
		owners[decision.Path], containers[decision.Path] = "container", decision
		for _, path := range decision.SuppressedPaths {
			steps, valid := parseExactModelPath(path)
			if !valid || !isModelDescendant(path, decision.Path) || !canonicalRedactionTreePath(steps) || owners[path] != "" {
				return invalid
			}
			owners[path] = decision.Path
		}
	}
	// Every published descendant has its complete ancestry, and a parent drop
	// owns all descendants without fictitious child actions or token uses.
	for path, owner := range owners {
		steps, _ := parseExactModelPath(path)
		if !canonicalRedactionTreePath(steps) {
			continue
		}
		for end := len(steps) - 2; end >= 2; end -= 2 {
			parent := formatExactModelPath(steps[:end])
			if owner == "field" || owner == "container" {
				decision, exists := containers[parent]
				if !exists || decision.Action != ActionAllow {
					return invalid
				}
			} else if parent == owner {
				break
			} else if owners[parent] != owner {
				return invalid
			}
		}
	}
	// Retaining envelopes carry exactly the surviving canonical fields' union,
	// not a stale pre-redaction set or a newly permissive classification.
	qualified := slices.Clone(manifest.ContainerDecisions)
	for i := range qualified {
		if qualified[i].Action == ActionAllow {
			qualified[i].DisclosurePartitionRefs = []string{}
		}
	}
	fields := []RedactedField{}
	for _, entry := range manifest.Entries {
		if entry.Outcome != "dropped" {
			fields = append(fields, RedactedField{Path: entry.Path, DisclosurePartitionRefs: entry.DisclosurePartitionRefs})
		}
	}
	qualifyRetainedContainers(qualified, fields)
	for i := range qualified {
		if !sameReportingValue(qualified[i].DisclosurePartitionRefs, manifest.ContainerDecisions[i].DisclosurePartitionRefs) {
			return invalid
		}
	}
	return nil
}

func canonicalRedactionTreePath(steps []modelPathStep) bool {
	if modelContainerKind(steps) != "" {
		return true
	}
	return len(steps) >= 6 && steps[len(steps)-2] == (modelPathStep{Value: "fields"}) && steps[len(steps)-1].Key && modelContainerKind(steps[:len(steps)-2]) == "block"
}
