package networkflow

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

var graphBucketWidths = map[int64]struct{}{
	60: {}, 300: {}, 900: {}, 3600: {}, 21600: {}, 86400: {},
}

type graphTimeBucket struct {
	StartUTC             time.Time
	EndUTC               time.Time
	UniqueVertexIDs      map[string]struct{}
	EdgeCount            int
	ContributingRowCount int
}

func allowedGraphBucketWidth(width int64) bool {
	_, ok := graphBucketWidths[width]
	return ok
}

func canonicalTimestampText(value string, parsed time.Time) bool {
	if value == "" || !strings.HasSuffix(value, "Z") || parsed.Year() < 1 || parsed.Year() > 9999 {
		return false
	}
	if dot := strings.IndexByte(value, '.'); dot >= 0 {
		fractionEnd := strings.IndexByte(value[dot:], 'Z')
		if fractionEnd < 2 || fractionEnd-1 > 6 {
			return false
		}
	}
	return timestamp(parsed) == value
}

func graphTimeBuckets(timeRange graphTimeRange, width int64, limit int) ([]graphTimeBucket, *semanticFailure) {
	if timeRange.StartUTC == nil || timeRange.EndUTC == nil || !timeRange.EndUTC.After(*timeRange.StartUTC) {
		return nil, invalidTimeRange("time_range", "complete_range_required", timeRange.StartUTC, timeRange.EndUTC)
	}
	if !allowedGraphBucketWidth(width) {
		return nil, invalidGraphAggregation("aggregation.bucket_width_seconds", "unsupported_width")
	}
	first := floorEpochBucket(timeRange.StartUTC.UTC(), width)
	endExclusive := ceilEpochBucket(timeRange.EndUTC.UTC(), width)
	if first.Year() < 1 || endExclusive.Year() > 9999 {
		return nil, invalidTimeRange("time_range", "invalid_bound", timeRange.StartUTC, timeRange.EndUTC)
	}
	count := (endExclusive.Unix() - first.Unix()) / width
	if count < 1 {
		return nil, invalidTimeRange("time_range", "empty_range", timeRange.StartUTC, timeRange.EndUTC)
	}
	if count > int64(limit) {
		return nil, graphLimitExceededAt(
			"time_bucket_limit_exceeded", "network_flow.max_time_buckets_per_graph", limit, limit+1, "graph_admission",
		)
	}
	buckets := make([]graphTimeBucket, 0, int(count))
	for index := int64(0); index < count; index++ {
		start := first.Add(time.Duration(index*width) * time.Second)
		buckets = append(buckets, graphTimeBucket{
			StartUTC: start, EndUTC: start.Add(time.Duration(width) * time.Second),
			UniqueVertexIDs: map[string]struct{}{},
		})
	}
	return buckets, nil
}

func floorEpochBucket(value time.Time, width int64) time.Time {
	seconds := value.UTC().Unix()
	quotient := seconds / width
	if seconds%width < 0 {
		quotient--
	}
	return time.Unix(quotient*width, 0).UTC()
}

func ceilEpochBucket(value time.Time, width int64) time.Time {
	floor := floorEpochBucket(value, width)
	if value.UTC().Equal(floor) {
		return floor
	}
	return floor.Add(time.Duration(width) * time.Second)
}

func composeTimeBucketGraphRow(
	incidentID uuid.UUID,
	row flowRow,
	tableByID map[string]tableRecord,
	composition *graphComposition,
) *semanticFailure {
	if composition == nil || len(composition.TimeBuckets) == 0 {
		return invalidGraphAggregation("aggregation", "variant_member_conflict")
	}
	bucketIndex := temporalBucketIndex(composition.TimeBuckets, row.FlowStartUTC)
	if bucketIndex < 0 {
		return invalidTimeRange("network_flow.flow_start_utc", "invalid_bound", nil, nil)
	}
	composition.ContributingRows++
	if composition.ContributingRows > composition.ResultLimits.MaxContributingRows {
		return graphLimitExceeded(
			"contributing_row_limit_exceeded", "network_flow.max_contributing_rows_per_graph",
			composition.ResultLimits.MaxContributingRows, composition.ResultLimits.MaxContributingRows+1,
		)
	}
	key, valid := edgeSemanticsForRow(row, tableByID[row.NetworkFlowTableID])
	if !valid {
		return newSemanticFailure(failureInvalidRequest, fieldDirection, "invalid_value")
	}
	srcID := endpointID(incidentID, "ip", key.First.IP)
	dstID := endpointID(incidentID, "ip", key.Second.IP)
	srcVertex := ensureGraphVertex(composition.Vertices, srcID, key.First.IP)
	dstVertex := ensureGraphVertex(composition.Vertices, dstID, key.Second.IP)
	if len(composition.Vertices) > composition.ResultLimits.MaxVertices {
		return graphLimitExceeded("vertex_limit_exceeded", "network_flow.max_graph_vertices", composition.ResultLimits.MaxVertices, composition.ResultLimits.MaxVertices+1)
	}
	addGraphVertexRow(srcVertex, row, tableByID[row.NetworkFlowTableID])
	if dstID != srcID {
		addGraphVertexRow(dstVertex, row, tableByID[row.NetworkFlowTableID])
	}
	bucket := &composition.TimeBuckets[bucketIndex]
	bucket.ContributingRowCount++
	bucket.UniqueVertexIDs[srcID] = struct{}{}
	bucket.UniqueVertexIDs[dstID] = struct{}{}

	edgeID := key.identity(incidentID, &bucket.StartUTC, &bucket.EndUTC)
	edge := composition.Edges[edgeID]
	if edge == nil {
		bucketStart := bucket.StartUTC
		bucketEnd := bucket.EndUTC
		edge = &graphEdge{Semantics: key,
			EdgeID: edgeID, SrcEndpointID: srcID, DstEndpointID: dstID,
			SrcEndpointValue: key.First.IP, DstEndpointValue: key.Second.IP,
			IPProtocol: row.IPProtocol, DstPort: cloneInt32(key.Second.Port),
			FirstFlowStartUTC: row.FlowStartUTC.UTC(), LastFlowEndUTC: row.FlowEndUTC.UTC(),
			ContributingTableID: map[string]struct{}{}, MappingFingerprints: map[string]struct{}{},
			BucketStartUTC: &bucketStart, BucketEndUTC: &bucketEnd,
		}
		composition.Edges[edgeID] = edge
		bucket.EdgeCount++
		if len(composition.Edges) > composition.ResultLimits.MaxEdges {
			return graphLimitExceeded("edge_limit_exceeded", "network_flow.max_graph_edges", composition.ResultLimits.MaxEdges, composition.ResultLimits.MaxEdges+1)
		}
	}
	return addGraphEdgeAggregate(row, tableByID, composition, edge)
}

func temporalBucketIndex(buckets []graphTimeBucket, value time.Time) int {
	if len(buckets) == 0 || value.Before(buckets[0].StartUTC) || !value.Before(buckets[len(buckets)-1].EndUTC) {
		return -1
	}
	width := int64(buckets[0].EndUTC.Sub(buckets[0].StartUTC) / time.Second)
	return int((floorEpochBucket(value.UTC(), width).Unix() - buckets[0].StartUTC.Unix()) / width)
}

func timeBucketSummariesResource(buckets []graphTimeBucket) []any {
	out := make([]any, 0, len(buckets))
	for _, bucket := range buckets {
		out = append(out, map[string]any{
			"start_utc": timestamp(bucket.StartUTC), "end_utc": timestamp(bucket.EndUTC),
			"unique_vertex_count": len(bucket.UniqueVertexIDs), "edge_count": bucket.EdgeCount,
			"contributing_row_count": bucket.ContributingRowCount,
		})
	}
	return out
}

func invalidGraphAggregation(field, reason string) *semanticFailure {
	return newSemanticFailure(failureInvalidGraphAggregation, field, reason)
}

func invalidTimeRange(field, reason string, start, end *time.Time) *semanticFailure {
	f := newSemanticFailure(failureInvalidTimeRange, field, reason)
	f.details.StartUTC = start
	f.details.EndUTC = end
	return f
}

func graphLimitExceededAt(reason, limitKey string, limit, actual int, phase string) *semanticFailure {
	f := newSemanticFailure(failureGraphLimitExceeded, "", reason)
	f.details.LimitKey = limitKey
	f.details.Limit = limit
	f.details.Actual = actual
	f.details.Phase = phase
	return f
}

func graphQueryDigestForSemantic(incidentID uuid.UUID, tableIDs []string, semantic graphSemanticRequest) (string, error) {
	return graphQueryDigestV2(incidentID, tableIDs, semantic.Filters, semantic.TimeRange, semantic.Aggregation)
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}
