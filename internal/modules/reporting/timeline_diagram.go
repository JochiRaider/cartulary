package reporting

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Diagram selection operates on retained events, never the raw snapshot or the
// flattened redaction decision index. Ordering is Core's typed tuple order.
func timelineDiagramFromEvents(decl compositionDiagramDecl, events []ReportingTimelineEvent, scope string) (reportingDiagramModel, error) {
	rule := decl.SelectionRule
	if hasDuplicateStrings(rule.TimelineRecordIDs) {
		return reportingDiagramModel{}, newRenderValidationError("graph_projection_unavailable", "diagram_selection_duplicate_ref")
	}
	raw, err := canonicalJSON(rule)
	var admitted diagramSelectionRule
	if err != nil || json.Unmarshal(raw, &admitted) != nil || !rule.CompatibleSource("timeline") || decl.DiagramSourceKind != "timeline" || decl.SourceGraphViewID != nil {
		return reportingDiagramModel{}, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
	}
	limit, err := diagramVertexLimit(decl.DiagramKind)
	if err != nil {
		return reportingDiagramModel{}, err
	}
	byID := make(map[string]ReportingTimelineEvent, len(events))
	for _, event := range events {
		if _, exists := byID[event.RecordID]; exists {
			return reportingDiagramModel{}, newRenderValidationError("export_model_invalid", "duplicate_stable_id")
		}
		byID[event.RecordID] = event
	}
	selected := make([]ReportingTimelineEvent, 0, len(rule.TimelineRecordIDs))
	for _, id := range rule.TimelineRecordIDs {
		event, present := byID[id]
		if !present {
			return reportingDiagramModel{}, newRenderValidationError("graph_projection_unavailable", "diagram_selection_missing_ref")
		}
		selected = append(selected, event)
	}
	slices.SortFunc(selected, func(a, b ReportingTimelineEvent) int { return compareTimelineEvents(a, b, false, false) })
	diagram := reportingDiagramModel{
		SourceRefs: []string{}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{},
		SchemaID: "cartulary.reporting_diagram.v1", DiagramID: decl.DeclID,
		DiagramKind: decl.DiagramKind, DiagramSourceKind: "timeline", LayoutMode: "auto",
		SelectionRule: rule, IncludedVertexRefs: []string{}, IncludedEdgeRefs: []string{},
		Nodes: []reportingDiagramNode{}, Edges: []reportingDiagramEdge{},
	}
	if len(selected) > limit {
		// A composition declaration cannot authorize the template-owned external
		// summarization exception. The current template has no such declaration.
		if rule.OverflowPolicy != "summarize" || scope == ReleaseScopeExternal {
			return reportingDiagramModel{}, newRenderValidationError("mermaid_invalid", "diagram_hard_limit_exceeded")
		}
		first := selected[limit].RecordID
		diagram.OverflowSummary = &diagramOverflowSummary{
			SchemaID: "diagram_overflow_summary.v1", SelectionRuleID: rule.RuleID,
			OmittedVertexCount: len(selected) - limit, OmittedEdgeCount: len(selected) - limit,
			FirstOmittedRef: &first,
		}
		selected = selected[:limit]
	}
	refs := emptyContainerRefs()
	for index, event := range selected {
		refs.include(event.SourceRefs, event.SupportRefs, event.DisclosurePartitionRefs)
		id := fmt.Sprintf("n%04d", index+1)
		label := event.RecordID
		if event.DisplayTimes.PrimaryDisplay != nil {
			label = *event.DisplayTimes.PrimaryDisplay
		}
		diagram.IncludedVertexRefs = append(diagram.IncludedVertexRefs, event.RecordID)
		diagram.Nodes = append(diagram.Nodes, reportingDiagramNode{NodeID: id, Label: label})
		if index == 0 {
			continue
		}
		previous := selected[index-1].RecordID
		ref, err := timelineAdjacencyRef(previous, event.RecordID)
		if err != nil {
			return reportingDiagramModel{}, err
		}
		diagram.IncludedEdgeRefs = append(diagram.IncludedEdgeRefs, ref)
		diagram.Edges = append(diagram.Edges, reportingDiagramEdge{
			EdgeID: fmt.Sprintf("e%04d", index), From: fmt.Sprintf("n%04d", index), To: id, Label: "next",
		})
	}
	refs = refs.normalized()
	diagram.SourceRefs, diagram.SupportRefs, diagram.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
	return diagram, nil
}

func diagramVertexLimit(kind string) (int, error) {
	switch kind {
	case "flowchart":
		return 80, nil
	case "sequence":
		return 24, nil
	default:
		return 0, newRenderValidationError("mermaid_invalid", "invalid_mermaid_construct")
	}
}

type diagramLabelOverride struct {
	Target struct {
		Kind string `json:"target_kind"`
		Ref  string `json:"ref"`
	} `json:"target"`
	Label string `json:"label"`
}

func applyDiagramLabelOverrides(diagram *reportingDiagramModel, overrides []diagramLabelOverride) error {
	for _, override := range overrides {
		matches := 0
		switch override.Target.Kind {
		case "vertex":
			for index, ref := range diagram.IncludedVertexRefs {
				if ref == override.Target.Ref {
					diagram.Nodes[index].Label = override.Label
					matches++
				}
			}
		case "edge":
			for index, ref := range diagram.IncludedEdgeRefs {
				if ref == override.Target.Ref {
					diagram.Edges[index].Label = override.Label
					matches++
				}
			}
		}
		if matches != 1 {
			return newRenderValidationError("composition_invalid", "diagram_selection_missing_ref")
		}
	}
	return nil
}

func timelineAdjacencyRef(from, to string) (string, error) {
	const schema = "cartulary.reporting_timeline_adjacency.v1"
	canonical, err := canonicalJSON(struct {
		SchemaID string `json:"schema_id"`
		From     string `json:"from_record_id"`
		To       string `json:"to_record_id"`
	}{schema, from, to})
	if err != nil {
		return "", err
	}
	return "timeline-edge-" + reportingObjectDigest(schema, canonical), nil
}
