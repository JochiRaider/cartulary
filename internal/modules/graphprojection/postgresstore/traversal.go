package postgresstore

import (
	"context"
	"sort"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func (s *Store) Traverse(ctx context.Context, request graphprojection.TraverseRequest) (graphprojection.TraverseResult, error) {
	graphView, err := s.GetGraphView(ctx, request.GraphViewID, request.ProjectionRunID)
	if err != nil {
		return graphprojection.TraverseResult{}, err
	}
	maxDepth := 1
	if request.MaxDepth != nil {
		maxDepth = *request.MaxDepth
	}
	if maxDepth < 0 || maxDepth > graphprojection.ResourceLimits().MaxTraversalDepth {
		return graphprojection.TraverseResult{}, &graphprojection.QueryError{Code: "invalid_argument", ReasonCode: "out_of_bounds", Field: "max_depth", Details: map[string]any{"field": "max_depth", "reason_code": "out_of_bounds"}}
	}
	direction := request.Direction
	if direction == "" {
		direction = "outbound"
	}
	if direction != "outbound" && direction != "inbound" && direction != "any" {
		return graphprojection.TraverseResult{}, &graphprojection.QueryError{Code: "invalid_argument", ReasonCode: "invalid_value", Field: "direction", Details: map[string]any{"field": "direction", "reason_code": "invalid_value"}}
	}
	if len(request.SeedVertexIDs) > graphprojection.ResourceLimits().MaxTraversalSeedVertices || len(request.VertexKinds) > graphprojection.ResourceLimits().MaxTraversalKindFilters || len(request.EdgeKinds) > graphprojection.ResourceLimits().MaxTraversalKindFilters {
		return graphprojection.TraverseResult{}, &graphprojection.QueryError{Code: "invalid_argument", ReasonCode: "out_of_bounds", Field: "traverse", Details: map[string]any{"field": "traverse", "reason_code": "out_of_bounds"}}
	}
	if hasDuplicateStrings(request.VertexKinds) || hasDuplicateStrings(request.EdgeKinds) {
		return graphprojection.TraverseResult{}, &graphprojection.QueryError{Code: "invalid_argument", ReasonCode: "duplicate_id", Field: "traverse", Details: map[string]any{"field": "traverse", "reason_code": "duplicate_id"}}
	}
	vertexKinds := stringSet(request.VertexKinds)
	edgeKinds := stringSet(request.EdgeKinds)
	filterVertexKinds := request.VertexKinds != nil
	filterEdgeKinds := request.EdgeKinds != nil
	vertexByID := map[string]graphprojection.Vertex{}
	vertexOrder := map[string]int{}
	for index, vertex := range graphView.Vertices {
		vertexByID[vertex.VertexID] = vertex
		vertexOrder[vertex.VertexID] = index
	}
	visited := map[string]bool{}
	frontier := []string{}
	omittedSeeds := []string{}
	seenSeeds := map[string]bool{}
	for _, seed := range request.SeedVertexIDs {
		if seenSeeds[seed] {
			return graphprojection.TraverseResult{}, &graphprojection.QueryError{Code: "invalid_argument", ReasonCode: "duplicate_id", Field: "seed_vertex_ids", Details: map[string]any{"field": "seed_vertex_ids", "reason_code": "duplicate_id"}}
		}
		seenSeeds[seed] = true
		if _, ok := vertexByID[seed]; ok {
			frontier = append(frontier, seed)
			visited[seed] = true
		} else {
			omittedSeeds = append(omittedSeeds, seed)
		}
	}
	sortVertexIDsByGraphOrder(frontier, vertexOrder)
	selectedEdges := map[string]graphprojection.Edge{}
	for depth := 0; depth < maxDepth && len(frontier) > 0; depth++ {
		next := []string{}
		for _, current := range frontier {
			for _, edge := range graphView.Edges {
				if filterEdgeKinds && !edgeKinds[edge.EdgeKind] {
					continue
				}
				candidate := traversalNeighbor(edge, current, direction)
				if candidate == "" {
					continue
				}
				vertex := vertexByID[candidate]
				if filterVertexKinds && !vertexKinds[vertex.VertexKind] {
					continue
				}
				selectedEdges[edge.EdgeID] = edge
				if !visited[candidate] {
					visited[candidate] = true
					next = append(next, candidate)
				}
			}
		}
		sortVertexIDsByGraphOrder(next, vertexOrder)
		frontier = next
	}
	vertices := []graphprojection.Vertex{}
	for vertexID := range visited {
		vertices = append(vertices, vertexByID[vertexID])
	}
	edges := []graphprojection.Edge{}
	for _, edge := range selectedEdges {
		edges = append(edges, edge)
	}
	graphprojection.SortVertices(vertices)
	graphprojection.SortEdges(edges)
	seeds := make([]string, 0, len(request.SeedVertexIDs)-len(omittedSeeds))
	requestedSeeds := stringSet(request.SeedVertexIDs)
	for _, vertex := range graphView.Vertices {
		if requestedSeeds[vertex.VertexID] {
			seeds = append(seeds, vertex.VertexID)
		}
	}
	return graphprojection.TraverseResult{GraphViewID: graphView.GraphViewID, ProjectionRunID: graphView.ProjectionRunID, SeedVertexIDs: seeds, OmittedSeedVertexIDs: omittedSeeds, Vertices: vertices, Edges: edges, Metadata: map[string]any{}}, nil
}

func sortVertexIDsByGraphOrder(vertexIDs []string, order map[string]int) {
	sort.SliceStable(vertexIDs, func(i, j int) bool {
		left, leftOK := order[vertexIDs[i]]
		right, rightOK := order[vertexIDs[j]]
		if leftOK && rightOK && left != right {
			return left < right
		}
		if leftOK != rightOK {
			return leftOK
		}
		return vertexIDs[i] < vertexIDs[j]
	})
}

func traversalNeighbor(edge graphprojection.Edge, current, requestedDirection string) string {
	if edge.Direction == "undirected" || edge.Direction == "bidirectional" || requestedDirection == "any" {
		if edge.SrcVertexID == current {
			return edge.DstVertexID
		}
		if edge.DstVertexID == current {
			return edge.SrcVertexID
		}
		return ""
	}
	if requestedDirection == "outbound" && edge.SrcVertexID == current {
		return edge.DstVertexID
	}
	if requestedDirection == "inbound" && edge.DstVertexID == current {
		return edge.SrcVertexID
	}
	return ""
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}
