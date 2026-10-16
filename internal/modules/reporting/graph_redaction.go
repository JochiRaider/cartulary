package reporting

import (
	"context"
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
)

// Selection retains topology and decision addresses only. Source values go
// through the same rule engine as snapshot fields; labels cannot read them here.
type selectedGraphDiagram struct {
	ref      sourceProjectionRef
	vertices []selectedGraphVertex
	edges    []selectedGraphEdge
	overflow *diagramOverflowSummary
}

type selectedGraphVertex struct{ id, endpoint string }
type selectedGraphEdge struct{ id, from, to, protocol, port, start, end string }
type graphRedactionSelection struct {
	diagrams map[string]selectedGraphDiagram
	fields   []ExportField
	sources  map[string]reportingFieldSource
}

func prepareGraphRedaction(ctx context.Context, input renderInput, composition *renderComposition) (graphRedactionSelection, error) {
	selection := graphRedactionSelection{diagrams: map[string]selectedGraphDiagram{}, sources: map[string]reportingFieldSource{}}
	decls, _, err := resolveDiagramDeclarations(input.TemplateContract, composition)
	if err != nil {
		return selection, err
	}
	refs, err := decodeSourceProjectionRefs(input.GraphProjectionRefs)
	if err != nil {
		return selection, newRenderValidationError("graph_projection_unavailable", "graph_projection_not_bound")
	}
	for _, decl := range decls {
		if err := ctx.Err(); err != nil {
			return selection, err
		}
		if decl.DiagramSourceKind != "graph" {
			continue
		}
		ref, err := resolveGraphProjectionRef(refs, decl.SourceGraphViewID)
		if err != nil {
			return selection, err
		}
		result, err := resolveLoadedGraphResult(input.GraphSources.Results, ref)
		if err != nil {
			return selection, err
		}
		if err := selection.include(decl, ref, result, input.ReleaseScope); err != nil {
			return selection, err
		}
	}
	return selection, nil
}

func (selection *graphRedactionSelection) include(decl compositionDiagramDecl, ref sourceProjectionRef, resolved resolvedGraphResult, scope string) error {
	invalid := newRenderValidationError("graph_projection_unavailable", "graph_projection_digest_mismatch")
	if resolved.Ref != ref || resolved.Result.Binding != ref.binding() || !validGraphLabelCandidates(graphsourcecontract.Result{Projection: resolved.Result, LabelCandidates: resolved.LabelCandidates}) {
		return invalid
	}
	if _, exists := selection.diagrams[decl.DeclID]; exists {
		return compositionFailure("duplicate_stable_id")
	}
	limit, err := diagramVertexLimit(decl.DiagramKind)
	if err != nil {
		return err
	}
	vertices, edges, overflow, err := selectGraphDiagramObjects(resolved.Result, decl.SelectionRule, limit)
	if err != nil {
		return err
	}
	if overflow != nil && scope == ReleaseScopeExternal {
		return newRenderValidationError("mermaid_invalid", "diagram_hard_limit_exceeded")
	}
	if scope == ReleaseScopeExternal && (len(vertices) > 0 || len(edges) > 0) {
		return newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
	}
	vertexLabels := map[string]graphsourcecontract.VertexLabelCandidate{}
	for _, candidate := range resolved.LabelCandidates.VertexLabelCandidates {
		vertexLabels[candidate.ProjectedVertexID] = candidate
	}
	edgeLabels := map[string]graphsourcecontract.EdgeLabelCandidate{}
	for _, candidate := range resolved.LabelCandidates.EdgeLabelCandidates {
		edgeLabels[candidate.ProjectedEdgeID] = candidate
	}
	diagram := selectedGraphDiagram{ref: ref, overflow: overflow}
	for _, vertex := range vertices {
		path, err := selection.component(decl.DeclID, "nodes", vertex.VertexID, vertexLabels[vertex.VertexID].Endpoint)
		if err != nil {
			return err
		}
		diagram.vertices = append(diagram.vertices, selectedGraphVertex{vertex.VertexID, path})
	}
	for _, edge := range edges {
		candidate := edgeLabels[edge.EdgeID]
		selected := selectedGraphEdge{id: edge.EdgeID, from: edge.SrcVertexID, to: edge.DstVertexID}
		selected.protocol, err = selection.component(decl.DeclID, "edges", edge.EdgeID, candidate.Protocol)
		if err != nil {
			return err
		}
		selected.port, err = selection.component(decl.DeclID, "edges", edge.EdgeID, candidate.DestinationPort)
		if err != nil {
			return err
		}
		if candidate.Kind == "time_bucket_v1" {
			selected.start, err = selection.component(decl.DeclID, "edges", edge.EdgeID, *candidate.BucketStartUTC)
			if err != nil {
				return err
			}
			selected.end, err = selection.component(decl.DeclID, "edges", edge.EdgeID, *candidate.BucketEndUTC)
			if err != nil {
				return err
			}
		}
		diagram.edges = append(diagram.edges, selected)
	}
	selection.diagrams[decl.DeclID] = diagram
	return nil
}

func (selection *graphRedactionSelection) component(diagramID, collection, id string, component graphsourcecontract.LabelComponent) (string, error) {
	path, err := keyedModelPath("$", "diagrams", diagramID)
	if err != nil {
		return "", err
	}
	path, err = keyedModelPath(path, collection, id)
	if err != nil {
		return "", err
	}
	path, err = keyedModelPath(path, "label_components", component.ComponentKind)
	if err != nil {
		return "", err
	}
	if _, exists := selection.sources[path]; exists {
		return "", newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
	}
	var value any
	if component.State == graphsourcecontract.ComponentStatePresent {
		switch component.ValueKind {
		case graphsourcecontract.ValueKindString:
			value = component.StringValue
		case graphsourcecontract.ValueKindInteger:
			value = component.IntegerValue
		}
	}
	// Typed absence never recovers a raw sentinel or unclassified Graph label.
	sourceState := "derived"
	selection.fields = append(selection.fields, ExportField{Path: path, SourceFamily: "network_flow_graph", SourceRecordID: component.SourceObjectRef, ContentClass: component.Classification, Value: value, DisclosurePartitionRefs: cloneStrings(component.DisclosurePartitions)})
	selection.sources[path] = reportingFieldSource{Field: ReportingField{
		SchemaID: "cartulary.reporting_field.v1", FieldKey: component.ComponentKind, FieldOrdinal: 1,
		SourceValueState: sourceState, RedactedValueState: "unchanged", Value: value,
		SourceRefs: []string{}, SupportRefs: []string{}, DisclosurePartitionRefs: cloneStrings(component.DisclosurePartitions),
	}, ContentClass: component.Classification}
	return path, nil
}

func graphDiagramFromSelection(decl compositionDiagramDecl, selected selectedGraphDiagram, model RedactedExportModel) (reportingDiagramModel, error) {
	fields := map[string]RedactedField{}
	for _, field := range model.Fields {
		fields[field.Path] = field
	}
	label := func(path string) (string, error) {
		if path == "" {
			return "", nil
		}
		if model.admittedFieldPaths[path] != path {
			return "", newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
		}
		field, retained := fields[path]
		if !retained || field.Value == nil {
			return "", nil
		}
		return formatFieldValue(field.Value), nil
	}
	ref := selected.ref
	diagram := reportingDiagramModel{SchemaID: "cartulary.reporting_diagram.v1", DiagramID: decl.DeclID, DiagramKind: decl.DiagramKind,
		DiagramSourceKind: "graph", LayoutMode: "auto", SourceProjectionRef: &ref, SelectionRule: decl.SelectionRule, OverflowSummary: selected.overflow,
		SourceRefs: []string{}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{"internal_only"},
		IncludedVertexRefs: []string{}, IncludedEdgeRefs: []string{}, Nodes: []reportingDiagramNode{}, Edges: []reportingDiagramEdge{}}
	for index, vertex := range selected.vertices {
		text, err := label(vertex.endpoint)
		if err != nil {
			return reportingDiagramModel{}, err
		}
		if text == "" {
			text = fmt.Sprintf("Vertex %04d", index+1)
		}
		diagram.IncludedVertexRefs = append(diagram.IncludedVertexRefs, vertex.id)
		diagram.Nodes = append(diagram.Nodes, reportingDiagramNode{NodeID: vertex.id, Label: text})
	}
	for index, edge := range selected.edges {
		parts := []string{}
		start, err := label(edge.start)
		if err != nil {
			return reportingDiagramModel{}, err
		}
		end, err := label(edge.end)
		if err != nil {
			return reportingDiagramModel{}, err
		}
		if start != "" && end != "" {
			parts = append(parts, "["+start+","+end+")")
		}
		for _, component := range []struct{ name, path string }{{"protocol", edge.protocol}, {"port", edge.port}} {
			text, err := label(component.path)
			if err != nil {
				return reportingDiagramModel{}, err
			}
			if text != "" {
				parts = append(parts, component.name+"="+text)
			}
		}
		text := strings.Join(parts, " ")
		if text == "" {
			text = fmt.Sprintf("Edge %04d", index+1)
		}
		diagram.IncludedEdgeRefs = append(diagram.IncludedEdgeRefs, edge.id)
		diagram.Edges = append(diagram.Edges, reportingDiagramEdge{EdgeID: edge.id, From: edge.from, To: edge.to, Label: text})
	}
	return diagram, nil
}
