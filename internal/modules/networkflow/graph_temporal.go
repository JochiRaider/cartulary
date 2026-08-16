package networkflow

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
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

func graphTimeBuckets(timeRange graphTimeRange, width int64, limit int) ([]graphTimeBucket, *httpapi.APIError) {
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

func BucketEdgeID(
	incidentID uuid.UUID,
	bucketStart time.Time,
	bucketEnd time.Time,
	srcEndpointID string,
	dstEndpointID string,
	ipProtocol int32,
	dstPort *int32,
) string {
	var transcript bytes.Buffer
	fields := []string{
		"cartulary.network_flow.bucket_edge_id.v1",
		incidentID.String(),
		timestamp(bucketStart),
		timestamp(bucketEnd),
		srcEndpointID,
		dstEndpointID,
		strconv.FormatInt(int64(ipProtocol), 10),
	}
	if dstPort == nil {
		fields = append(fields, "n")
	} else {
		fields = append(fields, "p", strconv.FormatInt(int64(*dstPort), 10))
	}
	for _, field := range fields {
		writeLengthFramedPart(&transcript, field)
	}
	return "nfbe_" + sha256Hex(transcript.Bytes())
}

func composeTimeBucketGraphRow(
	incidentID uuid.UUID,
	row FlowRow,
	tableByID map[string]TableRecord,
	composition *graphComposition,
) *httpapi.APIError {
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
	srcID := EndpointID(incidentID, "ip", row.SrcIP)
	dstID := EndpointID(incidentID, "ip", row.DstIP)
	srcVertex := ensureGraphVertex(composition.Vertices, srcID, row.SrcIP)
	dstVertex := ensureGraphVertex(composition.Vertices, dstID, row.DstIP)
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

	edgeID := BucketEdgeID(incidentID, bucket.StartUTC, bucket.EndUTC, srcID, dstID, row.IPProtocol, row.DstPort)
	edge := composition.Edges[edgeID]
	if edge == nil {
		bucketStart := bucket.StartUTC
		bucketEnd := bucket.EndUTC
		edge = &graphEdge{
			EdgeID: edgeID, SrcEndpointID: srcID, DstEndpointID: dstID,
			SrcEndpointValue: row.SrcIP, DstEndpointValue: row.DstIP,
			IPProtocol: row.IPProtocol, DstPort: cloneInt32(row.DstPort),
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

func invalidGraphAggregation(field string, reason string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusBadRequest, Code: "network_flow_invalid_graph_aggregation", Message: "network_flow_invalid_graph_aggregation",
		Details: map[string]any{"field": field, "reason_code": reason, "retry_action": "correct_request"},
	}
}

func invalidTimeRange(field string, reason string, start *time.Time, end *time.Time) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusBadRequest, Code: "network_flow_invalid_time_range", Message: "network_flow_invalid_time_range",
		Details: map[string]any{
			"field": field, "reason_code": reason, "start_utc": nullableTimestamp(start),
			"end_utc": nullableTimestamp(end), "retry_action": "correct_request",
		},
	}
}

func graphLimitExceededAt(reason string, limitKey string, limit int, actual int, phase string) *httpapi.APIError {
	return &httpapi.APIError{
		Status: http.StatusRequestEntityTooLarge, Code: "network_flow_graph_limit_exceeded", Message: "network_flow_graph_limit_exceeded",
		Details: map[string]any{
			"reason_code": reason, "limit_key": limitKey, "limit": limit, "actual": actual,
			"phase": phase, "retry_action": "reduce_scope_or_limits",
		},
	}
}

func graphQueryDigestForSemantic(incidentID uuid.UUID, tableIDs []string, semantic graphSemanticRequest) string {
	if semantic.SchemaID == schemaGraphSemanticQueryV1 {
		return graphQueryDigest(incidentID, tableIDs, semantic.Filters, semantic.TimeRange, semantic.Aggregation)
	}
	return graphQueryDigestV2(incidentID, tableIDs, semantic.Filters, semantic.TimeRange, semantic.Aggregation)
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := value.UTC()
	return &cloned
}
