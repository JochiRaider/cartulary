package networkflow

import (
	"math"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func graphQueryResultResource(composition graphComposition) map[string]any {
	if composition.SemanticSchemaID == schemaGraphSemanticQueryV1 {
		return map[string]any{
			"schema_id":          "cartulary.network_flow.graph_query_result.v1",
			"graph_query_digest": composition.Digest, "semantic_query": composition.SemanticQuery,
			"graph_projection_result": composition.GraphProjection,
			"edge_annotations":        composition.EdgeAnnotations, "source_table_refs": composition.SourceTableRefs,
			"result_limits": graphResultLimitsResourceV1(composition.ResultLimits),
		}
	}
	variant := map[string]any{"kind": composition.Aggregation.Mode}
	if composition.Aggregation.Mode == "time_bucket_v1" {
		variant["time_buckets"] = timeBucketSummariesResource(composition.TimeBuckets)
	}
	return map[string]any{
		"schema_id":          schemaGraphQueryResult,
		"graph_query_digest": composition.Digest, "semantic_query": composition.SemanticQuery,
		"graph_projection_result": composition.GraphProjection,
		"vertex_selectors":        composition.VertexSelectors, "edge_annotations": composition.EdgeAnnotations,
		"source_table_refs": composition.SourceTableRefs,
		"result_limits":     graphResultLimitsResourceV2(composition.ResultLimits),
		"result_variant":    variant,
	}
}

func bindGraphV2ResponseMetadata(composition *graphComposition) *httpapi.APIError {
	if composition == nil || composition.GraphProjection == nil {
		return graphProjectionFailed("adapter_contract_rejected")
	}
	rawVertices, ok := composition.GraphProjection["vertices"].([]any)
	if !ok || len(rawVertices) != len(composition.Vertices) {
		return graphProjectionFailed("adapter_contract_rejected")
	}
	vertexSelectors := make([]any, 0, len(rawVertices))
	seenVertices := make(map[string]struct{}, len(rawVertices))
	for _, raw := range rawVertices {
		projected, ok := raw.(map[string]any)
		if !ok {
			return graphProjectionFailed("adapter_contract_rejected")
		}
		projectedID, projectedOK := projected["vertex_id"].(string)
		sourceRef, refOK := projected["source_entity_ref"].(map[string]any)
		sourceID, sourceOK := sourceRef["source_entity_id"].(string)
		vertex := composition.Vertices[sourceID]
		if !projectedOK || projectedID == "" || !refOK || !sourceOK || vertex == nil {
			return graphProjectionFailed("adapter_contract_rejected")
		}
		if _, duplicate := seenVertices[sourceID]; duplicate {
			return graphProjectionFailed("adapter_contract_rejected")
		}
		seenVertices[sourceID] = struct{}{}
		vertexSelectors = append(vertexSelectors, map[string]any{
			"projected_vertex_id": projectedID,
			"selector": map[string]any{
				"kind": "vertex", "source_vertex_id": sourceID, "endpoint_value": vertex.EndpointValue,
			},
		})
	}

	rawEdges, ok := composition.GraphProjection["edges"].([]any)
	if !ok || len(rawEdges) != len(composition.Edges) {
		return graphProjectionFailed("adapter_contract_rejected")
	}
	edgeAnnotations := make([]any, 0, len(rawEdges))
	seenEdges := make(map[string]struct{}, len(rawEdges))
	for _, raw := range rawEdges {
		projected, ok := raw.(map[string]any)
		if !ok {
			return graphProjectionFailed("adapter_contract_rejected")
		}
		projectedID, projectedOK := projected["edge_id"].(string)
		sourceRef, refOK := projected["source_relationship_ref"].(map[string]any)
		sourceID, sourceOK := sourceRef["source_relationship_id"].(string)
		edge := composition.Edges[sourceID]
		if !projectedOK || projectedID == "" || !refOK || !sourceOK || edge == nil {
			return graphProjectionFailed("adapter_contract_rejected")
		}
		if _, duplicate := seenEdges[sourceID]; duplicate {
			return graphProjectionFailed("adapter_contract_rejected")
		}
		seenEdges[sourceID] = struct{}{}
		refs := []any{}
		if composition.IncludeExamples {
			for _, row := range edge.ExampleRows {
				refs = append(refs, rowRefResource(row))
			}
		}
		edgeAnnotations = append(edgeAnnotations, map[string]any{
			"projected_edge_id": projectedID, "selector": graphSelectorResource(graphSelectorForEdge(edge)),
			"example_row_refs": refs, "example_refs_truncated": len(refs) < edge.FlowRowCount,
			"example_refs_total_count": edge.FlowRowCount,
		})
	}
	composition.VertexSelectors = vertexSelectors
	composition.EdgeAnnotations = edgeAnnotations
	return nil
}

func deriveTimeBucketIndexFromExactResult(
	timeRange graphTimeRange,
	width int64,
	limit int,
	projection map[string]any,
) ([]graphTimeBucket, *httpapi.APIError) {
	buckets, apiErr := graphTimeBuckets(timeRange, width, limit)
	if apiErr != nil {
		return nil, malformedStoredGraphResult()
	}
	rawEdges, ok := projection["edges"].([]any)
	if !ok {
		return nil, malformedStoredGraphResult()
	}
	for _, raw := range rawEdges {
		edge, ok := raw.(map[string]any)
		if !ok {
			return nil, malformedStoredGraphResult()
		}
		properties, ok := edge["properties"].(map[string]any)
		if !ok {
			return nil, malformedStoredGraphResult()
		}
		startText, startOK := properties["bucket_start_utc"].(string)
		endText, endOK := properties["bucket_end_utc"].(string)
		srcID, srcOK := properties["src_endpoint_id"].(string)
		dstID, dstOK := properties["dst_endpoint_id"].(string)
		flowRows, rowsOK := exactGraphResultCount(properties["flow_row_count"])
		start, startErr := time.Parse(time.RFC3339Nano, startText)
		end, endErr := time.Parse(time.RFC3339Nano, endText)
		if !startOK || !endOK || !srcOK || !dstOK || !rowsOK || flowRows < 1 || startErr != nil || endErr != nil ||
			!canonicalTimestampText(startText, start) || !canonicalTimestampText(endText, end) {
			return nil, malformedStoredGraphResult()
		}
		index := temporalBucketIndex(buckets, start)
		if index < 0 || !buckets[index].StartUTC.Equal(start) || !buckets[index].EndUTC.Equal(end) {
			return nil, malformedStoredGraphResult()
		}
		bucket := &buckets[index]
		bucket.EdgeCount++
		bucket.ContributingRowCount += flowRows
		bucket.UniqueVertexIDs[srcID] = struct{}{}
		bucket.UniqueVertexIDs[dstID] = struct{}{}
	}
	return buckets, nil
}

func exactGraphResultCount(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, typed >= 0
	case int64:
		if typed < 0 || uint64(typed) > uint64(math.MaxInt) {
			return 0, false
		}
		return int(typed), true
	case float64:
		if typed < 0 || typed > float64(math.MaxInt) || math.Trunc(typed) != typed {
			return 0, false
		}
		return int(typed), true
	default:
		return 0, false
	}
}

func graphSelectorForEdge(edge *graphEdge) graphSelector {
	selector := graphSelector{
		Kind: "default_edge", SourceEdgeID: edge.EdgeID,
		SourceEndpointValue: edge.SrcEndpointValue, DestinationEndpointValue: edge.DstEndpointValue,
		Protocol: edge.IPProtocol, DestinationPortPresent: edge.DstPort != nil, DestinationPort: cloneInt32(edge.DstPort),
	}
	if edge.BucketStartUTC != nil && edge.BucketEndUTC != nil {
		selector.Kind = "time_bucket_edge"
		start := edge.BucketStartUTC.UTC()
		end := edge.BucketEndUTC.UTC()
		selector.BucketStartUTC = &start
		selector.BucketEndUTC = &end
	}
	return selector
}

func graphResultLimitsResourceV1(limits graphResultLimits) map[string]any {
	return map[string]any{
		"max_vertices": limits.MaxVertices, "max_edges": limits.MaxEdges,
		"max_example_row_refs_per_edge": limits.MaxExampleRowRefsPerEdge,
		"max_aggregate_counter_digits":  limits.MaxAggregateCounterDigits,
	}
}

func graphResultLimitsResourceV2(limits graphResultLimits) map[string]any {
	result := graphResultLimitsResourceV1(limits)
	result["max_contributing_rows_per_graph"] = limits.MaxContributingRows
	result["max_time_buckets_per_graph"] = limits.MaxTimeBuckets
	return result
}

func malformedStoredGraphResult() *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusInternalServerError, Code: "network_flow_graph_materialization_failed", Message: "network_flow_graph_materialization_failed",
		Details: map[string]any{"reason_code": "projection_rejected", "retry_action": "do_not_retry"},
	}
}
