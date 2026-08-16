package networkflow

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTimeBucketArithmeticFixturesAndLimit_Unit(t *testing.T) {
	tests := []struct {
		name       string
		start      time.Time
		end        time.Time
		width      int64
		wantFirst  int64
		wantEnd    int64
		wantBucket int
	}{
		{name: "negative epoch crossing", start: time.Unix(-1, 0), end: time.Unix(1, 0), width: 60, wantFirst: -60, wantEnd: 60, wantBucket: 2},
		{name: "negative epoch single", start: time.Unix(-61, 0), end: time.Unix(-60, 0), width: 60, wantFirst: -120, wantEnd: -60, wantBucket: 1},
		{name: "exact boundary end", start: time.Unix(0, 0), end: time.Unix(60, 0), width: 60, wantFirst: 0, wantEnd: 60, wantBucket: 1},
		{name: "fractional bounds", start: time.Unix(0, 500_000_000), end: time.Unix(60, 1_000), width: 60, wantFirst: 0, wantEnd: 120, wantBucket: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buckets, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &test.start, EndUTC: &test.end}, test.width, 1024)
			if apiErr != nil {
				t.Fatalf("bucket arithmetic failed: %#v", apiErr)
			}
			if len(buckets) != test.wantBucket || buckets[0].StartUTC.Unix() != test.wantFirst || buckets[len(buckets)-1].EndUTC.Unix() != test.wantEnd {
				t.Fatalf("buckets = %#v want count=%d first=%d end=%d", buckets, test.wantBucket, test.wantFirst, test.wantEnd)
			}
		})
	}
	start := time.Unix(0, 0).UTC()
	end := time.Unix(121, 0).UTC()
	if _, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, 60, 2); apiErr == nil ||
		apiErr.Code != "network_flow_graph_limit_exceeded" || apiErr.Details["reason_code"] != "time_bucket_limit_exceeded" ||
		apiErr.Details["actual"] != 3 || apiErr.Details["phase"] != "graph_admission" {
		t.Fatalf("time bucket limit+1 = %#v", apiErr)
	}
	maximumEnd := start.Add(1024 * time.Minute)
	if buckets, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &maximumEnd}, 60, 1024); apiErr != nil || len(buckets) != 1024 {
		t.Fatalf("maximum admitted bucket count = %d err=%#v", len(buckets), apiErr)
	}
}

func TestTimeBucketAdmissionAndCanonicalTimestampErrors_Unit(t *testing.T) {
	limits := DefaultEffectiveLimits()
	tests := []struct {
		name       string
		raw        string
		wantCode   string
		wantReason string
	}{
		{name: "unknown mode", raw: `{"mode":"calendar_bucket"}`, wantCode: "network_flow_invalid_graph_aggregation", wantReason: "unknown_mode"},
		{name: "variant conflict", raw: `{"mode":"default_flow_edge_v1","bucket_width_seconds":60}`, wantCode: "network_flow_invalid_graph_aggregation", wantReason: "variant_member_conflict"},
		{name: "missing width", raw: `{"mode":"time_bucket_v1"}`, wantCode: "network_flow_invalid_graph_aggregation", wantReason: "missing_width"},
		{name: "unsupported width", raw: `{"mode":"time_bucket_v1","bucket_width_seconds":120}`, wantCode: "network_flow_invalid_graph_aggregation", wantReason: "unsupported_width"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, apiErr := decodeGraphAggregationV2(json.RawMessage(test.raw))
			if apiErr == nil || apiErr.Code != test.wantCode || apiErr.Details["reason_code"] != test.wantReason {
				t.Fatalf("aggregation error = %#v", apiErr)
			}
		})
	}
	semantic := json.RawMessage(`{
  "schema_id":"cartulary.network_flow.graph_semantic_query.v2",
  "selected_table_ids":["nft_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"],
  "filters":[],
  "time_range":{"start_utc":null,"end_utc":"2026-08-16T12:00:00Z"},
  "aggregation":{"mode":"time_bucket_v1","bucket_width_seconds":60}
}`)
	if _, apiErr := decodeGraphSemanticRequest(semantic, limits); apiErr == nil || apiErr.Code != "network_flow_invalid_time_range" || apiErr.Details["reason_code"] != "complete_range_required" {
		t.Fatalf("incomplete temporal range = %#v", apiErr)
	}
	for _, raw := range []json.RawMessage{
		json.RawMessage(`"2026-08-16T08:00:00-04:00"`),
		json.RawMessage(`"2026-08-16T12:00:00.000000Z"`),
		json.RawMessage(`"2026-08-16T12:00:00.0000001Z"`),
	} {
		if _, apiErr := decodeOptionalTimestamp(raw, "start_utc"); apiErr == nil {
			t.Fatalf("noncanonical timestamp admitted: %s", raw)
		}
	}
}

func TestStreamingTimeBucketAggregationConservesRowsAndCounters_Unit(t *testing.T) {
	incidentID := IncidentID()
	start := time.Unix(0, 0).UTC()
	end := time.Unix(180, 0).UTC()
	buckets, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, 60, 3)
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	port := int32(443)
	rows := []FlowRow{
		temporalTestRow(start, 1, "2", "1", &port),
		temporalTestRow(start.Add(59*time.Second+999*time.Millisecond), 2, "3", "2", &port),
		temporalTestRow(start.Add(60*time.Second), 3, "5", "3", &port),
	}
	composition := graphComposition{
		SemanticSchemaID: schemaGraphSemanticQueryV2,
		Aggregation:      graphAggregation{Mode: "time_bucket_v1", BucketWidthSeconds: 60, IncludeExampleRowRefs: true},
		Vertices:         map[string]*graphVertex{}, Edges: map[string]*graphEdge{}, TimeBuckets: buckets,
		ResultLimits: graphResultLimits{
			MaxVertices: 2, MaxEdges: 2, MaxExampleRowRefsPerEdge: 1,
			MaxAggregateCounterDigits: 39, MaxContributingRows: 3, MaxTimeBuckets: 3,
		},
		IncludeExamples: true,
	}
	for _, row := range rows {
		matched, matchErr := rowMatchesGraphQuery(row, nil, graphTimeRange{StartUTC: &start, EndUTC: &end}, composition.Aggregation)
		if matchErr != nil || !matched {
			t.Fatalf("included temporal row rejected: matched=%v err=%#v", matched, matchErr)
		}
		if apiErr := composeGraphRow(incidentID, row, nil, &composition); apiErr != nil {
			t.Fatalf("compose temporal row: %#v", apiErr)
		}
	}
	excluded := temporalTestRow(end, 4, "100", "100", &port)
	if matched, _ := rowMatchesGraphQuery(excluded, nil, graphTimeRange{StartUTC: &start, EndUTC: &end}, composition.Aggregation); matched {
		t.Fatal("half-open temporal end admitted a row")
	}
	if composition.ContributingRows != 3 || len(composition.Vertices) != 2 || len(composition.Edges) != 2 {
		t.Fatalf("temporal cardinality rows=%d vertices=%d edges=%d", composition.ContributingRows, len(composition.Vertices), len(composition.Edges))
	}
	if buckets := composition.TimeBuckets; buckets[0].ContributingRowCount != 2 || buckets[0].EdgeCount != 1 || len(buckets[0].UniqueVertexIDs) != 2 ||
		buckets[1].ContributingRowCount != 1 || buckets[1].EdgeCount != 1 || buckets[2].ContributingRowCount != 0 || buckets[2].EdgeCount != 0 {
		t.Fatalf("temporal bucket summaries = %#v", timeBucketSummariesResource(buckets))
	}
	bytesTotal := int64(0)
	packetsTotal := int64(0)
	for _, edge := range composition.Edges {
		bytesTotal += edge.BytesSum.Int64()
		packetsTotal += edge.PacketsSum.Int64()
		if !strings.HasPrefix(edge.EdgeID, "nfbe_") || edge.BucketStartUTC == nil || edge.BucketEndUTC == nil || len(edge.ExampleRows) != 1 {
			t.Fatalf("temporal edge shape = %#v", edge)
		}
	}
	if bytesTotal != 10 || packetsTotal != 6 {
		t.Fatalf("temporal counters duplicated or lost: bytes=%d packets=%d", bytesTotal, packetsTotal)
	}

	limitedBuckets, _ := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, 60, 3)
	limited := graphComposition{
		SemanticSchemaID: schemaGraphSemanticQueryV2,
		Aggregation:      graphAggregation{Mode: "time_bucket_v1", BucketWidthSeconds: 60},
		Vertices:         map[string]*graphVertex{}, Edges: map[string]*graphEdge{}, TimeBuckets: limitedBuckets,
		ResultLimits: graphResultLimits{MaxVertices: 2, MaxEdges: 1, MaxAggregateCounterDigits: 1, MaxContributingRows: 2, MaxTimeBuckets: 3},
	}
	for index, row := range rows {
		apiErr := composeGraphRow(incidentID, row, nil, &limited)
		if index < 2 && apiErr != nil {
			t.Fatalf("temporal limit setup row %d: %#v", index, apiErr)
		}
		if index == 2 && (apiErr == nil || apiErr.Code != "network_flow_graph_limit_exceeded" || apiErr.Details["reason_code"] != "contributing_row_limit_exceeded") {
			t.Fatalf("temporal contributing-row limit+1 = %#v", apiErr)
		}
	}
	edgeLimitedBuckets, _ := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, 60, 3)
	edgeLimited := graphComposition{
		SemanticSchemaID: schemaGraphSemanticQueryV2,
		Aggregation:      graphAggregation{Mode: "time_bucket_v1", BucketWidthSeconds: 60},
		Vertices:         map[string]*graphVertex{}, Edges: map[string]*graphEdge{}, TimeBuckets: edgeLimitedBuckets,
		ResultLimits: graphResultLimits{MaxVertices: 2, MaxEdges: 1, MaxAggregateCounterDigits: 39, MaxContributingRows: 3, MaxTimeBuckets: 3},
	}
	for index, row := range rows {
		apiErr := composeGraphRow(incidentID, row, nil, &edgeLimited)
		if index < 2 && apiErr != nil {
			t.Fatalf("temporal edge-limit setup row %d: %#v", index, apiErr)
		}
		if index == 2 && (apiErr == nil || apiErr.Code != "network_flow_graph_limit_exceeded" || apiErr.Details["reason_code"] != "edge_limit_exceeded") {
			t.Fatalf("temporal global edge limit+1 = %#v", apiErr)
		}
	}
	counterLimitedBuckets, _ := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, 60, 3)
	counterLimited := graphComposition{
		SemanticSchemaID: schemaGraphSemanticQueryV2,
		Aggregation:      graphAggregation{Mode: "time_bucket_v1", BucketWidthSeconds: 60},
		Vertices:         map[string]*graphVertex{}, Edges: map[string]*graphEdge{}, TimeBuckets: counterLimitedBuckets,
		ResultLimits: graphResultLimits{MaxVertices: 2, MaxEdges: 3, MaxAggregateCounterDigits: 1, MaxContributingRows: 3, MaxTimeBuckets: 3},
	}
	counterRow := temporalTestRow(start, 1, "9", "1", &port)
	if apiErr := composeGraphRow(incidentID, counterRow, nil, &counterLimited); apiErr != nil {
		t.Fatal(apiErr)
	}
	counterRow.SourceRowNumber = 2
	if apiErr := composeGraphRow(incidentID, counterRow, nil, &counterLimited); apiErr == nil || apiErr.Code != "network_flow_counter_sum_limit_exceeded" || apiErr.Details["reason_code"] != "bytes_sum_digit_limit_exceeded" {
		t.Fatalf("temporal counter digit limit = %#v", apiErr)
	}
}

func TestTimeBucketIdentityProjectionAndResponseMetadata_Unit(t *testing.T) {
	incidentID := IncidentID()
	start := time.Unix(-60, 0).UTC()
	end := time.Unix(0, 0).UTC()
	port := int32(443)
	srcID := EndpointID(incidentID, "ip", "192.0.2.10")
	dstID := EndpointID(incidentID, "ip", "198.51.100.20")
	edgeID := BucketEdgeID(incidentID, start, end, srcID, dstID, 6, &port)
	if edgeID != "nfbe_edcebe53934ace485ecfa2477086c6fb76e08045042a4161354b38c77e1bbeb5" {
		t.Fatalf("time-bucket edge identity fixture = %s", edgeID)
	}
	if edgeID == BucketEdgeID(incidentID, start, start.Add(5*time.Minute), srcID, dstID, 6, &port) {
		t.Fatal("bucket end did not separate widths")
	}
	zeroPort := int32(0)
	if BucketEdgeID(incidentID, start, end, srcID, dstID, 6, nil) == BucketEdgeID(incidentID, start, end, srcID, dstID, 6, &zeroPort) {
		t.Fatal("destination-port presence did not participate in bucket edge identity")
	}

	table := TableRecord{IncidentID: incidentID, TableID: "nft_" + strings.Repeat("a", 64), MappingFingerprint: strings.Repeat("1", 64)}
	buckets, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, 60, 1)
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	composition := graphComposition{
		SemanticSchemaID: schemaGraphSemanticQueryV2,
		Aggregation:      graphAggregation{Mode: "time_bucket_v1", BucketWidthSeconds: 60, IncludeExampleRowRefs: true},
		SourceTables:     []TableRecord{table}, SourceTableRefs: graphSourceTableRefs([]TableRecord{table}),
		SelectedTableIDs: []string{table.TableID}, Vertices: map[string]*graphVertex{}, Edges: map[string]*graphEdge{},
		TimeBuckets: buckets, IncludeExamples: true,
		ResultLimits: graphResultLimits{MaxVertices: 2, MaxEdges: 1, MaxExampleRowRefsPerEdge: 1, MaxAggregateCounterDigits: 39, MaxContributingRows: 1, MaxTimeBuckets: 1},
	}
	row := temporalTestRow(start, 1, "2", "1", &port)
	row.NetworkFlowTableID = table.TableID
	if apiErr := composeGraphRow(incidentID, row, map[string]TableRecord{table.TableID: table}, &composition); apiErr != nil {
		t.Fatal(apiErr)
	}
	composition.Digest = graphQueryDigestV2(incidentID, composition.SelectedTableIDs, nil, graphTimeRange{StartUTC: &start, EndUTC: &end}, composition.Aggregation)
	composition.SemanticQuery = graphSemanticQueryResource(schemaGraphSemanticQueryV2, composition.SelectedTableIDs, nil, graphTimeRange{StartUTC: &start, EndUTC: &end}, composition.Aggregation, composition.ResultLimits)
	snapshotID := graphSourceSnapshotDigest(incidentID, composition.SourceTables, composition.Digest)
	service := &Service{graphProjection: newGraphProjectionAdapter(), now: time.Now}
	projection, apiErr := service.projectNetworkFlowGraph(context.Background(), uuid.MustParse("22222222-2222-4222-8222-222222222222"), snapshotID, composition, time.Now())
	if apiErr != nil {
		t.Fatalf("project temporal graph: %#v", apiErr)
	}
	canceledContext, cancel := context.WithCancel(context.Background())
	cancel()
	if canceledProjection, canceledErr := service.projectNetworkFlowGraph(canceledContext, uuid.MustParse("22222222-2222-4222-8222-222222222222"), snapshotID, composition, time.Now()); canceledErr == nil || canceledProjection != nil || canceledErr.Code != "network_flow_graph_projection_failed" || canceledErr.Details["reason_code"] != "projection_cancelled" {
		t.Fatalf("temporal projection cancellation returned partial output=%#v error=%#v", canceledProjection, canceledErr)
	}
	composition.GraphProjection = projection
	derivedBuckets, apiErr := deriveTimeBucketIndexFromExactResult(
		graphTimeRange{StartUTC: &start, EndUTC: &end}, 60, 1, projection,
	)
	if apiErr != nil || len(derivedBuckets) != 1 || derivedBuckets[0].ContributingRowCount != 1 || derivedBuckets[0].EdgeCount != 1 || len(derivedBuckets[0].UniqueVertexIDs) != 2 {
		t.Fatalf("derive saved bucket index from exact result = %#v err=%#v", timeBucketSummariesResource(derivedBuckets), apiErr)
	}
	composition.TimeBuckets = derivedBuckets
	if apiErr := bindGraphV2ResponseMetadata(&composition); apiErr != nil {
		t.Fatalf("bind temporal response: %#v", apiErr)
	}
	result := graphQueryResultResource(composition)
	if result["schema_id"] != schemaGraphQueryResult || len(result["vertex_selectors"].([]any)) != 2 || len(result["edge_annotations"].([]any)) != 1 {
		t.Fatalf("temporal result metadata = %#v", result)
	}
	variant := result["result_variant"].(map[string]any)
	if variant["kind"] != "time_bucket_v1" || len(variant["time_buckets"].([]any)) != 1 {
		t.Fatalf("temporal result variant = %#v", variant)
	}
	projectedEdge := projection["edges"].([]any)[0].(map[string]any)
	if projectedEdge["edge_kind"] != "network_flow.bucketed_flow_edge.v1" ||
		projectedEdge["source_relationship_ref"].(map[string]any)["source_relationship_id"] != edgeID ||
		projectedEdge["properties"].(map[string]any)["bucket_start_utc"] != timestamp(start) ||
		projection["projection_version"] != "network_flow_activity.time_bucket.v1" {
		t.Fatalf("temporal Graph Projection mapping = %#v", projectedEdge)
	}
	annotation := result["edge_annotations"].([]any)[0].(map[string]any)
	selector := annotation["selector"].(map[string]any)
	if annotation["projected_edge_id"] != projectedEdge["edge_id"] || selector["source_edge_id"] != edgeID || selector["kind"] != "time_bucket_edge" {
		t.Fatalf("temporal selector binding = %#v", annotation)
	}
}

func TestTimeBucketV2DigestFixture_Unit(t *testing.T) {
	incidentID := IncidentID()
	start := time.Date(2026, 8, 16, 12, 0, 0, 123_000_000, time.UTC)
	end := time.Date(2026, 8, 16, 13, 0, 0, 456_000_000, time.UTC)
	tables := []string{"nft_" + strings.Repeat("b", 64), "nft_" + strings.Repeat("a", 64)}
	digest := graphQueryDigestV2(incidentID, tables, nil, graphTimeRange{StartUTC: &start, EndUTC: &end}, graphAggregation{
		Mode: "time_bucket_v1", BucketWidthSeconds: 300, IncludeExampleRowRefs: true,
	})
	if digest != "7b8fdab65896fe81ab9d11e1ca28274a83910deee90682a99b2962bfe4ad1770" {
		t.Fatalf("temporal semantic-query-v2 digest fixture = %s", digest)
	}
	otherWidth := graphQueryDigestV2(incidentID, tables, nil, graphTimeRange{StartUTC: &start, EndUTC: &end}, graphAggregation{
		Mode: "time_bucket_v1", BucketWidthSeconds: 900, IncludeExampleRowRefs: true,
	})
	if otherWidth == digest {
		t.Fatal("bucket width did not participate in graph digest identity")
	}
}

func FuzzTimeBucketArithmetic(f *testing.F) {
	f.Add(int64(-1), int64(1), int64(60))
	f.Add(int64(0), int64(60), int64(60))
	f.Add(int64(-61), int64(-60), int64(60))
	f.Fuzz(func(t *testing.T, startSeconds int64, durationSeconds int64, width int64) {
		if startSeconds < -62_135_596_800 || startSeconds > 253_402_300_799 || durationSeconds < 1 || durationSeconds > 86_400 || !allowedGraphBucketWidth(width) {
			t.Skip()
		}
		start := time.Unix(startSeconds, 0).UTC()
		end := start.Add(time.Duration(durationSeconds) * time.Second)
		buckets, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, width, 1024)
		if apiErr != nil {
			if apiErr.Details["reason_code"] == "time_bucket_limit_exceeded" {
				return
			}
			t.Fatalf("unexpected arithmetic failure: %#v", apiErr)
		}
		if len(buckets) == 0 || buckets[0].StartUTC.After(start) || !buckets[len(buckets)-1].EndUTC.After(start) || buckets[len(buckets)-1].EndUTC.Before(end) {
			t.Fatalf("bucket coverage invalid: start=%s end=%s buckets=%#v", start, end, buckets)
		}
		for index, bucket := range buckets {
			if bucket.EndUTC.Sub(bucket.StartUTC) != time.Duration(width)*time.Second || index > 0 && !bucket.StartUTC.Equal(buckets[index-1].EndUTC) {
				t.Fatalf("bucket continuity invalid: %#v", buckets)
			}
		}
	})
}

func TestTimeBucketArithmeticPropertySeeds_Unit(t *testing.T) {
	widths := []int64{60, 300, 900, 3600, 21600, 86400}
	for _, width := range widths {
		for startSeconds := int64(-2*width - 1); startSeconds <= 2*width+1; startSeconds += max(int64(1), width/7) {
			for _, durationSeconds := range []int64{1, width - 1, width, width + 1} {
				assertTimeBucketCoverage(t, startSeconds, durationSeconds, width)
			}
		}
	}
}

func assertTimeBucketCoverage(t testing.TB, startSeconds int64, durationSeconds int64, width int64) {
	t.Helper()
	start := time.Unix(startSeconds, 0).UTC()
	end := start.Add(time.Duration(durationSeconds) * time.Second)
	buckets, apiErr := graphTimeBuckets(graphTimeRange{StartUTC: &start, EndUTC: &end}, width, 1024)
	if apiErr != nil {
		t.Fatalf("unexpected arithmetic failure: %#v", apiErr)
	}
	if len(buckets) == 0 || buckets[0].StartUTC.After(start) || buckets[len(buckets)-1].EndUTC.Before(end) {
		t.Fatalf("bucket coverage invalid: start=%s end=%s buckets=%#v", start, end, buckets)
	}
	for index, bucket := range buckets {
		if bucket.EndUTC.Sub(bucket.StartUTC) != time.Duration(width)*time.Second || index > 0 && !bucket.StartUTC.Equal(buckets[index-1].EndUTC) {
			t.Fatalf("bucket continuity invalid: %#v", buckets)
		}
	}
}

func temporalTestRow(start time.Time, sourceRowNumber int64, bytesCount string, packetsCount string, port *int32) FlowRow {
	return FlowRow{
		NetworkFlowTableID: "nft_" + strings.Repeat("a", 64), RowID: "nfr_" + strings.Repeat(strconvDigit(sourceRowNumber), 64),
		SourceRowNumber: sourceRowNumber, FlowStartUTC: start.UTC(), FlowEndUTC: start.Add(time.Second).UTC(),
		SrcIP: "192.0.2.10", DstIP: "198.51.100.20", DstPort: port, IPProtocol: 6,
		BytesCount: bytesCount, PacketsCount: packetsCount,
	}
}

func strconvDigit(value int64) string {
	digit := byte('0' + value%10)
	return string([]byte{digit})
}
