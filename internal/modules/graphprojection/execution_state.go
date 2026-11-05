package graphprojection

import "context"

// projectionState owns output capacity and indexes for every emission phase.
// Builders run only after capacity admission, so an over-limit object is never
// constructed, retained or visible to a subsequent rule.
type projectionState struct {
	budget                  *resultBudgetV2
	vertices                []Vertex
	edges                   []Edge
	verticesBySource        map[string][]Vertex
	verticesByRuleAndDigest map[string]map[string]Vertex
	kinds                   map[string]string
}

func (state *projectionState) emitVertex(ctx context.Context, issues *issueCollector, groupDigest string, build func() (Vertex, error)) error {
	if len(state.vertices) >= graphProjectionLimits.MaxProjectedVertices {
		return resourceProjectionErrorV2("maximum_projected_vertices", graphProjectionLimits.MaxProjectedVertices, graphProjectionLimits.MaxProjectedVertices+1)
	}
	if err := itemCheckpointV2(ctx); err != nil {
		return err
	}
	mark := state.budget.mark()
	before := len(issues.entries)
	vertex, err := build()
	if err != nil {
		return err
	}
	if len(issues.entries) != before {
		state.budget.restore(mark)
		return nil
	}
	if state.budget != nil {
		shell := vertex
		shell.Properties, shell.Metadata.MappedMetadata, shell.Metadata.AggregationSourceRefs = nil, nil, nil
		if err := state.budget.arrayEntry(vertexResource(shell), len(state.vertices)); err != nil {
			return err
		}
	}
	state.vertices = append(state.vertices, vertex)
	if ref := vertex.SourceEntityRef; ref != nil {
		if state.verticesBySource == nil {
			state.verticesBySource = map[string][]Vertex{}
		}
		state.verticesBySource[ref.SourceEntityID] = append(state.verticesBySource[ref.SourceEntityID], vertex)
	}
	if rule := vertex.Metadata.AggregationRuleID; rule != nil {
		if state.verticesByRuleAndDigest == nil {
			state.verticesByRuleAndDigest = map[string]map[string]Vertex{}
		}
		if state.verticesByRuleAndDigest[*rule] == nil {
			state.verticesByRuleAndDigest[*rule] = map[string]Vertex{}
		}
		state.verticesByRuleAndDigest[*rule][groupDigest] = vertex
	}
	return nil
}

func (state *projectionState) emitEdge(ctx context.Context, issues *issueCollector, build func() (Edge, error)) error {
	if len(state.edges) >= graphProjectionLimits.MaxProjectedEdges {
		return resourceProjectionErrorV2("maximum_projected_edges", graphProjectionLimits.MaxProjectedEdges, graphProjectionLimits.MaxProjectedEdges+1)
	}
	if err := itemCheckpointV2(ctx); err != nil {
		return err
	}
	mark := state.budget.mark()
	before := len(issues.entries)
	edge, err := build()
	if err != nil {
		return err
	}
	if len(issues.entries) != before {
		state.budget.restore(mark)
		return nil
	}
	if state.budget != nil {
		shell := edge
		shell.Properties, shell.Metadata.MappedMetadata, shell.Metadata.AggregationSourceRefs = nil, nil, nil
		if err := state.budget.arrayEntry(edgeResource(shell), len(state.edges)); err != nil {
			return err
		}
	}
	state.edges = append(state.edges, edge)
	return nil
}
