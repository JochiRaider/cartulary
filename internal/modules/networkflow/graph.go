package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const (
	schemaGraphQueryRequest                 = "cartulary.network_flow.graph_query_request.v2"
	schemaGraphQueryResult                  = "cartulary.network_flow.graph_query_result.v2"
	schemaGraphSemanticQueryV2              = "cartulary.network_flow.graph_semantic_query.v2"
	schemaGraphContributorQueryRequest      = "cartulary.network_flow.graph_contributor_query_request.v2"
	schemaGraphContributorQueryContinuation = "cartulary.network_flow.graph_contributor_query_continuation.v1"
	schemaGraphContributorQueryResult       = "cartulary.network_flow.graph_contributor_query_result.v2"
	routeKeyGraphsContributorsQuery         = "nf.graphs.contributors.query"
)

type graphTimeRange struct {
	StartUTC *time.Time
	EndUTC   *time.Time
	Omitted  bool
}

type graphAggregation struct {
	Mode                  string
	IncludeExampleRowRefs bool
	BucketWidthSeconds    int64
}

type graphResultLimits struct {
	MaxVertices               int
	MaxEdges                  int
	MaxExampleRowRefsPerEdge  int
	MaxAggregateCounterDigits int
	MaxContributingRows       int
	MaxTimeBuckets            int
}

type graphQueryRequest struct {
	TableScope  TableScope
	Filters     []Filter
	TimeRange   graphTimeRange
	Aggregation graphAggregation
	Limits      graphResultLimits
}

type graphSelector struct {
	Kind                     string     `json:"kind"`
	SourceVertexID           string     `json:"source_vertex_id,omitempty"`
	SourceEdgeID             string     `json:"source_edge_id,omitempty"`
	EndpointValue            string     `json:"endpoint_value,omitempty"`
	SourceEndpointValue      string     `json:"source_endpoint_value,omitempty"`
	DestinationEndpointValue string     `json:"destination_endpoint_value,omitempty"`
	Protocol                 int32      `json:"protocol,omitempty"`
	DestinationPortPresent   bool       `json:"destination_port_present,omitempty"`
	DestinationPort          *int32     `json:"destination_port,omitempty"`
	BucketStartUTC           *time.Time `json:"bucket_start_utc,omitempty"`
	BucketEndUTC             *time.Time `json:"bucket_end_utc,omitempty"`
}

type graphContributorQueryRequest struct {
	Continuation     bool
	CursorToken      string
	GraphQuery       graphSemanticRequest
	GraphQueryDigest string
	Selector         graphSelector
	Limit            int
}

type graphSemanticRequest struct {
	SchemaID         string
	SelectedTableIDs []string
	Filters          []Filter
	TimeRange        graphTimeRange
	Aggregation      graphAggregation
	ResultLimits     graphResultLimits
	Raw              map[string]any
}

type graphComposition struct {
	SemanticSchemaID string
	Aggregation      graphAggregation
	Digest           string
	SemanticQuery    map[string]any
	ResultLimits     graphResultLimits
	SourceTables     []TableRecord
	SourceTableRefs  []any
	TableRanks       map[string]int
	Vertices         map[string]*graphVertex
	Edges            map[string]*graphEdge
	GraphProjection  map[string]any
	VertexSelectors  []any
	EdgeAnnotations  []any
	SelectedTableIDs []string
	ContributingRows int
	IncludeExamples  bool
	TimeBuckets      []graphTimeBucket
}

type graphVertex struct {
	EndpointID          string
	EndpointValue       string
	ContributingTableID map[string]struct{}
	MappingFingerprints map[string]struct{}
	FlowRowCount        int
}

type graphEdge struct {
	EdgeID              string
	SrcEndpointID       string
	DstEndpointID       string
	SrcEndpointValue    string
	DstEndpointValue    string
	IPProtocol          int32
	DstPort             *int32
	FlowRowCount        int
	ExampleRows         []FlowRow
	BytesSum            big.Int
	PacketsSum          big.Int
	FirstFlowStartUTC   time.Time
	LastFlowEndUTC      time.Time
	ContributingTableID map[string]struct{}
	MappingFingerprints map[string]struct{}
	BucketStartUTC      *time.Time
	BucketEndUTC        *time.Time
}

func (s *Service) handleGraphQuery(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := parseIncidentPathValue(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeGraphQueryRequest(r, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	composition, apiErr := s.composeGraph(r.Context(), incidentID, principal.User.ID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := s.recordGraphQueryAudit(r.Context(), incidentID, principal.User.ID, composition, httpapi.RequestIDFromContext(r.Context())); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, graphQueryResultResource(composition))
}

func (s *Service) handleGraphContributorsQuery(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := parseIncidentPathValue(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeGraphContributorQueryRequest(r, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, apiErr := s.queryGraphContributors(r.Context(), principal.User.ID.String(), principal.Session.ID.String(), incidentID, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func decodeGraphQueryRequest(r *http.Request, limits EffectiveLimits) (graphQueryRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONString(raw, "schema_id")
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	if schemaID != schemaGraphQueryRequest {
		return graphQueryRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "table_scope", "filters", "time_range", "aggregation", "limit_overrides"); apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	scope, apiErr := requiredTableScope(raw["table_scope"], limits)
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	filters, apiErr := decodeFilters(raw["filters"], limits)
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	timeRange, apiErr := decodeGraphTimeRangeV2(raw["time_range"])
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	aggregation, apiErr := decodeGraphAggregationV2(raw["aggregation"])
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	if apiErr := validateAggregationTimeRange(aggregation, timeRange); apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	resultLimits, apiErr := decodeGraphResultLimits(raw["limit_overrides"], limits)
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	return graphQueryRequest{
		TableScope:  scope,
		Filters:     filters,
		TimeRange:   timeRange,
		Aggregation: aggregation,
		Limits:      resultLimits,
	}, nil
}

func decodeGraphContributorQueryRequest(r *http.Request, limits EffectiveLimits) (graphContributorQueryRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONString(raw, "schema_id")
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	if schemaID == schemaGraphContributorQueryContinuation {
		if apiErr := ensureAllowedMembers(raw, "schema_id", "cursor_token"); apiErr != nil {
			return graphContributorQueryRequest{}, apiErr
		}
		token, apiErr := requiredJSONString(raw, "cursor_token")
		if apiErr != nil {
			return graphContributorQueryRequest{}, apiErr
		}
		return graphContributorQueryRequest{Continuation: true, CursorToken: token}, nil
	}
	if schemaID != schemaGraphContributorQueryRequest {
		return graphContributorQueryRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "graph_query", "graph_query_digest", "selector", "limit"); apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	semantic, apiErr := decodeGraphSemanticRequest(raw["graph_query"], limits)
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	digest, apiErr := requiredJSONString(raw, "graph_query_digest")
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	selector, apiErr := decodeGraphSelector(raw["selector"])
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	value, ok := raw["limit"]
	if !ok {
		return graphContributorQueryRequest{}, invalidNetworkFlowRequest("limit", "missing_member")
	}
	limit, apiErr := decodePositiveInt(value, "limit")
	if apiErr != nil {
		return graphContributorQueryRequest{}, invalidLimit("limit", "not_integer")
	}
	if limit < 1 {
		return graphContributorQueryRequest{}, invalidLimit("limit", "below_minimum")
	}
	if int64(limit) > limits.MaxQueryLimit {
		return graphContributorQueryRequest{}, invalidLimit("limit", "above_maximum")
	}
	return graphContributorQueryRequest{
		GraphQuery:       semantic,
		GraphQueryDigest: digest,
		Selector:         selector,
		Limit:            limit,
	}, nil
}

func decodeGraphTimeRangeV2(raw json.RawMessage) (graphTimeRange, *httpapi.APIError) {
	if len(raw) == 0 {
		return graphTimeRange{Omitted: true}, nil
	}
	if bytes.Equal(raw, []byte("null")) {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "explicit_null")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "type_mismatch")
	}
	if apiErr := ensureAllowedMembers(object, "start_utc", "end_utc"); apiErr != nil {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "unknown_member")
	}
	if _, ok := object["start_utc"]; !ok {
		return graphTimeRange{}, invalidTimeRange("start_utc", "invalid_bound", nil, nil)
	}
	if _, ok := object["end_utc"]; !ok {
		return graphTimeRange{}, invalidTimeRange("end_utc", "invalid_bound", nil, nil)
	}
	start, apiErr := decodeOptionalTimestamp(object["start_utc"], "start_utc")
	if apiErr != nil {
		return graphTimeRange{}, invalidTimeRange("start_utc", "invalid_bound", nil, nil)
	}
	end, apiErr := decodeOptionalTimestamp(object["end_utc"], "end_utc")
	if apiErr != nil {
		return graphTimeRange{}, invalidTimeRange("end_utc", "invalid_bound", start, nil)
	}
	if start == nil && end == nil {
		return graphTimeRange{}, invalidTimeRange("time_range", "both_bounds_null", start, end)
	}
	if start != nil && end != nil && !end.After(*start) {
		return graphTimeRange{}, invalidTimeRange("time_range", "empty_range", start, end)
	}
	return graphTimeRange{StartUTC: start, EndUTC: end}, nil
}

// decodeSemanticTimeRangeV2 accepts the normalized default-query representation
// with two null bounds. Public query requests still reject an explicitly empty
// range; persisted semantic queries must be able to represent an omitted range
// without losing their canonical shape.
func decodeSemanticTimeRangeV2(raw json.RawMessage) (graphTimeRange, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return graphTimeRange{}, invalidNetworkFlowRequest("graph_query.time_range", "missing_member")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphTimeRange{}, invalidNetworkFlowRequest("graph_query.time_range", "type_mismatch")
	}
	if apiErr := ensureAllowedMembers(object, "start_utc", "end_utc"); apiErr != nil {
		return graphTimeRange{}, invalidNetworkFlowRequest("graph_query.time_range", "unknown_member")
	}
	if _, ok := object["start_utc"]; !ok {
		return graphTimeRange{}, invalidTimeRange("start_utc", "invalid_bound", nil, nil)
	}
	if _, ok := object["end_utc"]; !ok {
		return graphTimeRange{}, invalidTimeRange("end_utc", "invalid_bound", nil, nil)
	}
	start, apiErr := decodeOptionalTimestamp(object["start_utc"], "graph_query.time_range.start_utc")
	if apiErr != nil {
		return graphTimeRange{}, invalidTimeRange("start_utc", "invalid_bound", nil, nil)
	}
	end, apiErr := decodeOptionalTimestamp(object["end_utc"], "graph_query.time_range.end_utc")
	if apiErr != nil {
		return graphTimeRange{}, invalidTimeRange("end_utc", "invalid_bound", start, nil)
	}
	if start != nil && end != nil && !end.After(*start) {
		return graphTimeRange{}, invalidTimeRange("time_range", "empty_range", start, end)
	}
	return graphTimeRange{StartUTC: start, EndUTC: end, Omitted: start == nil && end == nil}, nil
}

func decodeOptionalTimestamp(raw json.RawMessage, field string) (*time.Time, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil || text == "" {
		return nil, invalidNetworkFlowRequest(field, "type_mismatch")
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil || !canonicalTimestampText(text, parsed) {
		return nil, invalidNetworkFlowRequest(field, "invalid_timestamp")
	}
	value := parsed.UTC()
	return &value, nil
}

func decodeGraphAggregationV2(raw json.RawMessage) (graphAggregation, *httpapi.APIError) {
	out := graphAggregation{IncludeExampleRowRefs: true}
	if len(raw) == 0 {
		return graphAggregation{}, invalidGraphAggregation("aggregation", "unknown_mode")
	}
	if bytes.Equal(raw, []byte("null")) {
		return graphAggregation{}, invalidGraphAggregation("aggregation", "variant_member_conflict")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphAggregation{}, invalidGraphAggregation("aggregation", "variant_member_conflict")
	}
	for key := range object {
		if key != "mode" && key != "include_example_row_refs" && key != "bucket_width_seconds" {
			return graphAggregation{}, invalidGraphAggregation("aggregation."+key, "variant_member_conflict")
		}
	}
	mode, apiErr := requiredJSONString(object, "mode")
	if apiErr != nil {
		return graphAggregation{}, invalidGraphAggregation("aggregation.mode", "unknown_mode")
	}
	out.Mode = mode
	if value, ok := object["include_example_row_refs"]; ok {
		var include bool
		if err := json.Unmarshal(value, &include); err != nil {
			return graphAggregation{}, invalidGraphAggregation("aggregation.include_example_row_refs", "variant_member_conflict")
		}
		out.IncludeExampleRowRefs = include
	}
	switch mode {
	case "default_flow_edge_v1":
		if _, present := object["bucket_width_seconds"]; present {
			return graphAggregation{}, invalidGraphAggregation("aggregation.bucket_width_seconds", "variant_member_conflict")
		}
		return out, nil
	case "time_bucket_v1":
		rawWidth, present := object["bucket_width_seconds"]
		if !present {
			return graphAggregation{}, invalidGraphAggregation("aggregation.bucket_width_seconds", "missing_width")
		}
		width, decodeErr := decodePositiveInt(rawWidth, "aggregation.bucket_width_seconds")
		if decodeErr != nil || !allowedGraphBucketWidth(int64(width)) {
			return graphAggregation{}, invalidGraphAggregation("aggregation.bucket_width_seconds", "unsupported_width")
		}
		out.BucketWidthSeconds = int64(width)
		return out, nil
	default:
		return graphAggregation{}, invalidGraphAggregation("aggregation.mode", "unknown_mode")
	}
}

func validateAggregationTimeRange(aggregation graphAggregation, timeRange graphTimeRange) *httpapi.APIError {
	if aggregation.Mode != "time_bucket_v1" {
		return nil
	}
	if timeRange.Omitted || timeRange.StartUTC == nil || timeRange.EndUTC == nil {
		return invalidTimeRange("time_range", "complete_range_required", timeRange.StartUTC, timeRange.EndUTC)
	}
	return nil
}

func decodeGraphResultLimits(raw json.RawMessage, limits EffectiveLimits) (graphResultLimits, *httpapi.APIError) {
	out := graphResultLimits{
		MaxVertices:               int(limits.MaxGraphVertices),
		MaxEdges:                  int(limits.MaxGraphEdges),
		MaxExampleRowRefsPerEdge:  int(limits.MaxExampleRowRefsPerEdge),
		MaxAggregateCounterDigits: int(limits.MaxAggregateCounterDigits),
		MaxContributingRows:       int(limits.MaxContributingRowsPerGraph),
		MaxTimeBuckets:            int(limits.MaxTimeBucketsPerGraph),
	}
	if len(raw) == 0 {
		return out, nil
	}
	if bytes.Equal(raw, []byte("null")) {
		return graphResultLimits{}, invalidLimitOverride("limit_overrides", "explicit_null", "", 0, 0, 0)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphResultLimits{}, invalidLimitOverride("limit_overrides", "type_mismatch", "", 0, 0, 0)
	}
	allowed := map[string]struct{}{
		"max_vertices":                    {},
		"max_edges":                       {},
		"max_example_row_refs_per_edge":   {},
		"max_contributing_rows_per_graph": {},
		"max_time_buckets_per_graph":      {},
	}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return graphResultLimits{}, invalidLimitOverride(key, "unknown_limit_key", key, 0, 0, 0)
		}
	}
	if value, ok := object["max_vertices"]; ok {
		parsed, apiErr := decodeLowerableGraphLimit(value, "max_vertices", 1, out.MaxVertices)
		if apiErr != nil {
			return graphResultLimits{}, apiErr
		}
		out.MaxVertices = parsed
	}
	if value, ok := object["max_edges"]; ok {
		parsed, apiErr := decodeLowerableGraphLimit(value, "max_edges", 0, out.MaxEdges)
		if apiErr != nil {
			return graphResultLimits{}, apiErr
		}
		out.MaxEdges = parsed
	}
	if value, ok := object["max_example_row_refs_per_edge"]; ok {
		parsed, apiErr := decodeLowerableGraphLimit(value, "max_example_row_refs_per_edge", 0, out.MaxExampleRowRefsPerEdge)
		if apiErr != nil {
			return graphResultLimits{}, apiErr
		}
		out.MaxExampleRowRefsPerEdge = parsed
	}
	if value, ok := object["max_contributing_rows_per_graph"]; ok {
		parsed, apiErr := decodeLowerableGraphLimit(value, "max_contributing_rows_per_graph", 1, out.MaxContributingRows)
		if apiErr != nil {
			return graphResultLimits{}, apiErr
		}
		out.MaxContributingRows = parsed
	}
	if value, ok := object["max_time_buckets_per_graph"]; ok {
		parsed, apiErr := decodeLowerableGraphLimit(value, "max_time_buckets_per_graph", 1, out.MaxTimeBuckets)
		if apiErr != nil {
			return graphResultLimits{}, apiErr
		}
		out.MaxTimeBuckets = parsed
	}
	return out, nil
}

func decodeLowerableGraphLimit(raw json.RawMessage, key string, minimum int, maximum int) (int, *httpapi.APIError) {
	parsed, apiErr := decodePositiveInt(raw, key)
	if apiErr != nil {
		return 0, invalidLimitOverride(key, "not_integer", key, 0, minimum, maximum)
	}
	if parsed < minimum {
		return 0, invalidLimitOverride(key, "below_minimum", key, parsed, minimum, maximum)
	}
	if parsed > maximum {
		return 0, invalidLimitOverride(key, "above_maximum", key, parsed, minimum, maximum)
	}
	return parsed, nil
}

func decodeGraphSelector(raw json.RawMessage) (graphSelector, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return graphSelector{}, invalidNetworkFlowRequest("selector", "missing_member")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphSelector{}, invalidNetworkFlowRequest("selector", "type_mismatch")
	}
	kind, apiErr := requiredJSONString(object, "kind")
	if apiErr != nil {
		return graphSelector{}, invalidNetworkFlowRequest("selector.kind", "missing_member")
	}
	switch kind {
	case "vertex":
		if apiErr := ensureAllowedMembers(object, "kind", "source_vertex_id", "endpoint_value"); apiErr != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector", "variant_member_conflict")
		}
		sourceVertexID, apiErr := requiredJSONString(object, "source_vertex_id")
		if apiErr != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector.source_vertex_id", "missing_member")
		}
		endpointValue, apiErr := requiredCanonicalGraphIP(object, "endpoint_value")
		if apiErr != nil {
			return graphSelector{}, apiErr
		}
		return graphSelector{Kind: kind, SourceVertexID: sourceVertexID, EndpointValue: endpointValue}, nil
	case "default_edge", "time_bucket_edge":
		if apiErr := ensureAllowedMembers(object, "kind", "source_edge_id", "source_endpoint_value", "destination_endpoint_value", "protocol", "destination_port_present", "destination_port"); apiErr != nil {
			if kind != "time_bucket_edge" {
				return graphSelector{}, invalidNetworkFlowRequest("selector", "variant_member_conflict")
			}
			if apiErr := ensureAllowedMembers(object, "kind", "source_edge_id", "bucket_start_utc", "bucket_end_utc", "source_endpoint_value", "destination_endpoint_value", "protocol", "destination_port_present", "destination_port"); apiErr != nil {
				return graphSelector{}, invalidNetworkFlowRequest("selector", "variant_member_conflict")
			}
		}
		sourceEdgeID, apiErr := requiredJSONString(object, "source_edge_id")
		if apiErr != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector.source_edge_id", "missing_member")
		}
		sourceValue, apiErr := requiredCanonicalGraphIP(object, "source_endpoint_value")
		if apiErr != nil {
			return graphSelector{}, apiErr
		}
		destinationValue, apiErr := requiredCanonicalGraphIP(object, "destination_endpoint_value")
		if apiErr != nil {
			return graphSelector{}, apiErr
		}
		protocol, apiErr := decodePositiveInt(object["protocol"], "selector.protocol")
		if apiErr != nil || protocol < 0 || protocol > 255 {
			return graphSelector{}, invalidNetworkFlowRequest("selector.protocol", "invalid_value")
		}
		var portPresent bool
		if rawPresent, ok := object["destination_port_present"]; !ok || json.Unmarshal(rawPresent, &portPresent) != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector.destination_port_present", "missing_member")
		}
		var port *int32
		rawPort, hasPort := object["destination_port"]
		if portPresent {
			parsed, portErr := decodePositiveInt(rawPort, "selector.destination_port")
			if !hasPort || portErr != nil || parsed < 0 || parsed > 65535 {
				return graphSelector{}, invalidNetworkFlowRequest("selector.destination_port", "invalid_value")
			}
			value := int32(parsed)
			port = &value
		} else if hasPort {
			return graphSelector{}, invalidNetworkFlowRequest("selector.destination_port", "variant_member_conflict")
		}
		selector := graphSelector{
			Kind:                     kind,
			SourceEdgeID:             sourceEdgeID,
			SourceEndpointValue:      sourceValue,
			DestinationEndpointValue: destinationValue,
			Protocol:                 int32(protocol),
			DestinationPortPresent:   portPresent,
			DestinationPort:          port,
		}
		if kind == "time_bucket_edge" {
			bucketStart, timeErr := decodeRequiredGraphTimestamp(object, "bucket_start_utc")
			if timeErr != nil {
				return graphSelector{}, timeErr
			}
			bucketEnd, timeErr := decodeRequiredGraphTimestamp(object, "bucket_end_utc")
			if timeErr != nil || !bucketEnd.After(bucketStart) {
				return graphSelector{}, invalidNetworkFlowRequest("selector.bucket_end_utc", "invalid_value")
			}
			selector.BucketStartUTC = &bucketStart
			selector.BucketEndUTC = &bucketEnd
		}
		return selector, nil
	default:
		return graphSelector{}, invalidNetworkFlowRequest("selector.kind", "unknown_selector_kind")
	}
}

func decodeRequiredGraphTimestamp(object map[string]json.RawMessage, field string) (time.Time, *httpapi.APIError) {
	value, present := object[field]
	if !present || bytes.Equal(value, []byte("null")) {
		return time.Time{}, invalidNetworkFlowRequest("selector."+field, "missing_member")
	}
	parsed, apiErr := decodeOptionalTimestamp(value, "selector."+field)
	if apiErr != nil || parsed == nil {
		return time.Time{}, invalidNetworkFlowRequest("selector."+field, "invalid_value")
	}
	return parsed.UTC(), nil
}

func requiredCanonicalGraphIP(object map[string]json.RawMessage, field string) (string, *httpapi.APIError) {
	value, apiErr := requiredJSONString(object, field)
	if apiErr != nil {
		return "", invalidNetworkFlowRequest("selector."+field, "missing_member")
	}
	canonical, err := parseIPLiteral(value)
	if err != nil || canonical != value {
		return "", invalidNetworkFlowRequest("selector."+field, "invalid_value")
	}
	return canonical, nil
}

func decodeGraphSemanticRequest(raw json.RawMessage, limits EffectiveLimits) (graphSemanticRequest, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query", "missing_member")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query", "type_mismatch")
	}
	schemaID, apiErr := requiredJSONString(object, "schema_id")
	if apiErr != nil || schemaID != schemaGraphSemanticQueryV2 {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query.schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(object, "schema_id", "selected_table_ids", "filters", "time_range", "aggregation"); apiErr != nil {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query", "unknown_member")
	}
	tableIDs, apiErr := decodeStringArray(object["selected_table_ids"], "selected_table_ids", int(limits.MaxSelectedTablesPerQuery))
	if apiErr != nil || len(tableIDs) == 0 {
		return graphSemanticRequest{}, invalidTableScope("selected_table_ids", "empty_resolved_scope")
	}
	filters, apiErr := decodeFilters(object["filters"], limits)
	if apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	if _, present := object["time_range"]; !present {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query.time_range", "missing_member")
	}
	timeRange, apiErr := decodeSemanticTimeRangeV2(object["time_range"])
	var aggregation graphAggregation
	if apiErr == nil {
		aggregation, apiErr = decodeGraphAggregationV2(object["aggregation"])
	}
	if apiErr == nil {
		apiErr = validateAggregationTimeRange(aggregation, timeRange)
	}
	if apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	resultLimits := effectiveGraphResultLimits(limits)
	rawObject := graphSemanticQueryResource(schemaID, tableIDs, filters, timeRange, aggregation, resultLimits)
	return graphSemanticRequest{
		SchemaID:         schemaID,
		SelectedTableIDs: tableIDs,
		Filters:          filters,
		TimeRange:        timeRange,
		Aggregation:      aggregation,
		ResultLimits:     resultLimits,
		Raw:              rawObject,
	}, nil
}

func effectiveGraphResultLimits(limits EffectiveLimits) graphResultLimits {
	return graphResultLimits{
		MaxVertices:               int(limits.MaxGraphVertices),
		MaxEdges:                  int(limits.MaxGraphEdges),
		MaxExampleRowRefsPerEdge:  int(limits.MaxExampleRowRefsPerEdge),
		MaxAggregateCounterDigits: int(limits.MaxAggregateCounterDigits),
		MaxContributingRows:       int(limits.MaxContributingRowsPerGraph),
		MaxTimeBuckets:            int(limits.MaxTimeBucketsPerGraph),
	}
}

func (s *Service) composeGraph(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, request graphQueryRequest) (graphComposition, *httpapi.APIError) {
	composition, apiErr := s.composeGraphSource(ctx, incidentID, request)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	if err := ctx.Err(); err != nil {
		return graphComposition{}, graphProjectionFailedForContext(err)
	}
	sourceSnapshotID := graphSourceSnapshotDigest(incidentID, composition.SourceTables, composition.Digest)
	projectionStarted := time.Now()
	projection, apiErr := s.projectNetworkFlowGraph(ctx, actorUserID, sourceSnapshotID, composition, s.now())
	s.observeGraphPhase(ctx, graphTelemetryPhaseProjection, request.Aggregation.Mode, projectionStarted, apiErr)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	composition.GraphProjection = projection
	if apiErr := bindGraphV2ResponseMetadata(&composition); apiErr != nil {
		return graphComposition{}, apiErr
	}
	s.observeGraphComposition(ctx, composition)
	return composition, nil
}

func (s *Service) composeGraphSource(ctx context.Context, incidentID uuid.UUID, request graphQueryRequest) (graphComposition, *httpapi.APIError) {
	validationStarted := time.Now()
	tables, tableIDs, tableRanks, apiErr := s.resolveGraphTables(ctx, incidentID, request.TableScope)
	s.observeGraphPhase(ctx, graphTelemetryPhaseSourceValidation, request.Aggregation.Mode, validationStarted, apiErr)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	schemaID := schemaGraphSemanticQueryV2
	digest := graphQueryDigestV2(incidentID, tableIDs, request.Filters, request.TimeRange, request.Aggregation)
	composition := graphComposition{
		SemanticSchemaID: schemaID,
		Aggregation:      request.Aggregation,
		Digest:           digest,
		ResultLimits:     request.Limits,
		SourceTables:     tables,
		SourceTableRefs:  graphSourceTableRefs(tables),
		TableRanks:       tableRanks,
		Vertices:         map[string]*graphVertex{},
		Edges:            map[string]*graphEdge{},
		SelectedTableIDs: tableIDs,
		IncludeExamples:  request.Aggregation.IncludeExampleRowRefs,
	}
	if request.Aggregation.Mode == "time_bucket_v1" {
		buckets, bucketErr := graphTimeBuckets(request.TimeRange, request.Aggregation.BucketWidthSeconds, request.Limits.MaxTimeBuckets)
		if bucketErr != nil {
			s.observeGraphPhase(ctx, graphTelemetryPhaseSourceValidation, request.Aggregation.Mode, validationStarted, bucketErr)
			return graphComposition{}, bucketErr
		}
		composition.TimeBuckets = buckets
	}
	tableByID := make(map[string]TableRecord, len(tables))
	for _, table := range tables {
		tableByID[table.TableID] = table
	}
	if err := ctx.Err(); err != nil {
		return graphComposition{}, graphProjectionFailedForContext(err)
	}
	var compositionErr *httpapi.APIError
	scanStarted := time.Now()
	err := s.store.IterateRowsForTables(ctx, incidentID, tableIDs, func(row FlowRow) error {
		matched, rowErr := rowMatchesGraphQuery(row, request.Filters, request.TimeRange, request.Aggregation)
		if rowErr != nil {
			compositionErr = rowErr
			return errStopGraphIteration
		}
		if !matched {
			return nil
		}
		if rowErr = composeGraphRow(incidentID, row, tableByID, &composition); rowErr != nil {
			compositionErr = rowErr
			return errStopGraphIteration
		}
		return nil
	})
	if compositionErr != nil {
		s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, compositionErr)
		return graphComposition{}, compositionErr
	}
	if err != nil {
		var scanErr *httpapi.APIError
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			scanErr = graphProjectionFailedForContext(err)
		} else {
			scanErr = httpapi.InternalAPIError(err)
		}
		s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, scanErr)
		return graphComposition{}, scanErr
	}
	if apiErr := validateGraphLimits(composition); apiErr != nil {
		s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, apiErr)
		return graphComposition{}, apiErr
	}
	composition.SemanticQuery = graphSemanticQueryResource(schemaID, tableIDs, request.Filters, request.TimeRange, request.Aggregation, request.Limits)
	s.observeGraphPhase(ctx, graphTelemetryPhaseSourceScan, request.Aggregation.Mode, scanStarted, nil)
	return composition, nil
}

func (s *Service) composeGraphSourceFromSemantic(ctx context.Context, incidentID uuid.UUID, semantic graphSemanticRequest) (graphComposition, *httpapi.APIError) {
	if semantic.ResultLimits.MaxContributingRows == 0 {
		semantic.ResultLimits.MaxContributingRows = int(s.store.limits.MaxContributingRowsPerGraph)
	}
	if semantic.ResultLimits.MaxTimeBuckets == 0 {
		semantic.ResultLimits.MaxTimeBuckets = int(s.store.limits.MaxTimeBucketsPerGraph)
	}
	request := graphQueryRequest{
		TableScope:  TableScope{Mode: "selected_tables", SelectedTableIDs: semantic.SelectedTableIDs},
		Filters:     semantic.Filters,
		TimeRange:   semantic.TimeRange,
		Aggregation: semantic.Aggregation,
		Limits:      semantic.ResultLimits,
	}
	composition, apiErr := s.composeGraphSource(ctx, incidentID, request)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	composition.SemanticSchemaID = semantic.SchemaID
	composition.SemanticQuery = semantic.Raw
	composition.Digest = graphQueryDigestForSemantic(incidentID, composition.SelectedTableIDs, semantic)
	return composition, nil
}

var errStopGraphIteration = errors.New("stop Network Flow graph iteration")

func (s *Service) resolveGraphTables(ctx context.Context, incidentID uuid.UUID, scope TableScope) ([]TableRecord, []string, map[string]int, *httpapi.APIError) {
	activeTables, err := s.store.ListActiveTables(ctx, incidentID)
	if err != nil {
		return nil, nil, nil, httpapi.InternalAPIError(err)
	}
	byID := make(map[string]TableRecord, len(activeTables))
	for _, table := range activeTables {
		byID[table.TableID] = table
	}
	var selected []TableRecord
	switch scope.Mode {
	case "active_table":
		table, ok := byID[scope.ActiveTableID]
		if !ok {
			if _, err := s.store.GetActiveTable(ctx, incidentID, scope.ActiveTableID); err != nil {
				return nil, nil, nil, tableReadError(err)
			}
			return nil, nil, nil, networkFlowAPIError(http.StatusNotFound, "network_flow_table_not_found", "network_flow_table_id", "not_found")
		}
		selected = []TableRecord{table}
	case "selected_tables":
		if len(scope.SelectedTableIDs) == 0 {
			return nil, nil, nil, invalidTableScope("table_scope", "empty_resolved_scope")
		}
		selectedSet := stringSet(scope.SelectedTableIDs)
		for _, table := range activeTables {
			if _, ok := selectedSet[table.TableID]; ok {
				selected = append(selected, table)
				delete(selectedSet, table.TableID)
			}
		}
		if len(selectedSet) > 0 {
			missing := make([]string, 0, len(selectedSet))
			for tableID := range selectedSet {
				missing = append(missing, tableID)
			}
			sort.Strings(missing)
			if _, err := s.store.GetActiveTable(ctx, incidentID, missing[0]); err != nil {
				return nil, nil, nil, tableReadError(err)
			}
			return nil, nil, nil, networkFlowAPIError(http.StatusNotFound, "network_flow_table_not_found", "network_flow_table_id", "not_found")
		}
	case "all_active_tables":
		selected = activeTables
	default:
		return nil, nil, nil, invalidTableScope("mode", "unknown_mode")
	}
	if len(selected) == 0 {
		return nil, nil, nil, invalidTableScope("table_scope", "empty_resolved_scope")
	}
	tableIDs := make([]string, 0, len(selected))
	tableRanks := make(map[string]int, len(selected))
	for index, table := range selected {
		tableIDs = append(tableIDs, table.TableID)
		tableRanks[table.TableID] = index
	}
	return selected, tableIDs, tableRanks, nil
}

func rowMatchesGraphQuery(row FlowRow, filters []Filter, timeRange graphTimeRange, aggregation graphAggregation) (bool, *httpapi.APIError) {
	for _, filter := range filters {
		matched, apiErr := rowMatchesFilter(row, filter)
		if apiErr != nil || !matched {
			return matched, apiErr
		}
	}
	if timeRange.Omitted || (timeRange.StartUTC == nil && timeRange.EndUTC == nil) {
		return true, nil
	}
	if aggregation.Mode == "time_bucket_v1" {
		return timeRange.StartUTC != nil && timeRange.EndUTC != nil &&
			!row.FlowStartUTC.Before(*timeRange.StartUTC) && row.FlowStartUTC.Before(*timeRange.EndUTC), nil
	}
	if timeRange.StartUTC != nil && row.FlowEndUTC.Before(*timeRange.StartUTC) {
		return false, nil
	}
	if timeRange.EndUTC != nil && !row.FlowStartUTC.Before(*timeRange.EndUTC) {
		return false, nil
	}
	return true, nil
}

func composeGraphRow(incidentID uuid.UUID, row FlowRow, tableByID map[string]TableRecord, composition *graphComposition) *httpapi.APIError {
	if composition != nil && composition.Aggregation.Mode == "time_bucket_v1" {
		return composeTimeBucketGraphRow(incidentID, row, tableByID, composition)
	}
	return composeDefaultGraphRow(incidentID, row, tableByID, composition)
}

func composeDefaultGraphRow(incidentID uuid.UUID, row FlowRow, tableByID map[string]TableRecord, composition *graphComposition) *httpapi.APIError {
	composition.ContributingRows++
	if composition.ResultLimits.MaxContributingRows > 0 && composition.ContributingRows > composition.ResultLimits.MaxContributingRows {
		return graphLimitExceeded("contributing_row_limit_exceeded", "network_flow.max_contributing_rows_per_graph", composition.ResultLimits.MaxContributingRows, composition.ResultLimits.MaxContributingRows+1)
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
	edgeID := FlowEdgeID(incidentID, srcID, dstID, row.IPProtocol, row.DstPort)
	edge := composition.Edges[edgeID]
	if edge == nil {
		edge = &graphEdge{
			EdgeID:              edgeID,
			SrcEndpointID:       srcID,
			DstEndpointID:       dstID,
			SrcEndpointValue:    row.SrcIP,
			DstEndpointValue:    row.DstIP,
			IPProtocol:          row.IPProtocol,
			DstPort:             cloneInt32(row.DstPort),
			FirstFlowStartUTC:   row.FlowStartUTC.UTC(),
			LastFlowEndUTC:      row.FlowEndUTC.UTC(),
			ContributingTableID: map[string]struct{}{},
			MappingFingerprints: map[string]struct{}{},
		}
		composition.Edges[edgeID] = edge
		if len(composition.Edges) > composition.ResultLimits.MaxEdges {
			return graphLimitExceeded("edge_limit_exceeded", "network_flow.max_graph_edges", composition.ResultLimits.MaxEdges, composition.ResultLimits.MaxEdges+1)
		}
	}
	return addGraphEdgeAggregate(row, tableByID, composition, edge)
}

func addGraphEdgeAggregate(row FlowRow, tableByID map[string]TableRecord, composition *graphComposition, edge *graphEdge) *httpapi.APIError {
	if row.FlowStartUTC.Before(edge.FirstFlowStartUTC) {
		edge.FirstFlowStartUTC = row.FlowStartUTC.UTC()
	}
	if row.FlowEndUTC.After(edge.LastFlowEndUTC) {
		edge.LastFlowEndUTC = row.FlowEndUTC.UTC()
	}
	bytesValue, ok := new(big.Int).SetString(row.BytesCount, 10)
	if !ok {
		return networkFlowAPIError(http.StatusBadRequest, "network_flow_invalid_request", FieldBytesCount, "invalid_counter")
	}
	packetsValue, ok := new(big.Int).SetString(row.PacketsCount, 10)
	if !ok {
		return networkFlowAPIError(http.StatusBadRequest, "network_flow_invalid_request", FieldPacketsCount, "invalid_counter")
	}
	edge.BytesSum.Add(&edge.BytesSum, bytesValue)
	edge.PacketsSum.Add(&edge.PacketsSum, packetsValue)
	if digits := decimalDigitCount(edge.BytesSum.String()); digits > composition.ResultLimits.MaxAggregateCounterDigits {
		return counterLimitExceeded("bytes_sum_digit_limit_exceeded", "network_flow.max_aggregate_counter_digits", composition.ResultLimits.MaxAggregateCounterDigits, digits)
	}
	if digits := decimalDigitCount(edge.PacketsSum.String()); digits > composition.ResultLimits.MaxAggregateCounterDigits {
		return counterLimitExceeded("packets_sum_digit_limit_exceeded", "network_flow.max_aggregate_counter_digits", composition.ResultLimits.MaxAggregateCounterDigits, digits)
	}
	edge.FlowRowCount++
	if composition.IncludeExamples && len(edge.ExampleRows) < composition.ResultLimits.MaxExampleRowRefsPerEdge {
		edge.ExampleRows = append(edge.ExampleRows, row)
	}
	edge.ContributingTableID[row.NetworkFlowTableID] = struct{}{}
	if table, ok := tableByID[row.NetworkFlowTableID]; ok {
		edge.MappingFingerprints[table.MappingFingerprint] = struct{}{}
	}
	return nil
}

func ensureGraphVertex(vertices map[string]*graphVertex, endpointID string, endpointValue string) *graphVertex {
	vertex := vertices[endpointID]
	if vertex == nil {
		vertex = &graphVertex{
			EndpointID:          endpointID,
			EndpointValue:       endpointValue,
			ContributingTableID: map[string]struct{}{},
			MappingFingerprints: map[string]struct{}{},
		}
		vertices[endpointID] = vertex
	}
	return vertex
}

func addGraphVertexRow(vertex *graphVertex, row FlowRow, table TableRecord) {
	vertex.FlowRowCount++
	vertex.ContributingTableID[row.NetworkFlowTableID] = struct{}{}
	if table.MappingFingerprint != "" {
		vertex.MappingFingerprints[table.MappingFingerprint] = struct{}{}
	}
}

func validateGraphLimits(composition graphComposition) *httpapi.APIError {
	if len(composition.Vertices) > composition.ResultLimits.MaxVertices {
		return graphLimitExceeded("vertex_limit_exceeded", "network_flow.max_graph_vertices", composition.ResultLimits.MaxVertices, composition.ResultLimits.MaxVertices+1)
	}
	if len(composition.Edges) > composition.ResultLimits.MaxEdges {
		return graphLimitExceeded("edge_limit_exceeded", "network_flow.max_graph_edges", composition.ResultLimits.MaxEdges, composition.ResultLimits.MaxEdges+1)
	}
	edgeIDs := sortedGraphEdgeIDs(composition.Edges)
	for _, edgeID := range edgeIDs {
		digits := decimalDigitCount(composition.Edges[edgeID].BytesSum.String())
		if digits > composition.ResultLimits.MaxAggregateCounterDigits {
			return counterLimitExceeded("bytes_sum_digit_limit_exceeded", "network_flow.max_aggregate_counter_digits", composition.ResultLimits.MaxAggregateCounterDigits, digits)
		}
	}
	for _, edgeID := range edgeIDs {
		digits := decimalDigitCount(composition.Edges[edgeID].PacketsSum.String())
		if digits > composition.ResultLimits.MaxAggregateCounterDigits {
			return counterLimitExceeded("packets_sum_digit_limit_exceeded", "network_flow.max_aggregate_counter_digits", composition.ResultLimits.MaxAggregateCounterDigits, digits)
		}
	}
	return nil
}

func (s *Service) projectNetworkFlowGraph(ctx context.Context, actorUserID uuid.UUID, sourceSnapshotID string, composition graphComposition, requestedAt time.Time) (map[string]any, *httpapi.APIError) {
	if err := ctx.Err(); err != nil {
		return nil, graphProjectionFailedForContext(err)
	}
	graphViewKeySnapshot := sourceSnapshotID
	if len(graphViewKeySnapshot) > len("nfsnap_") && graphViewKeySnapshot[:len("nfsnap_")] == "nfsnap_" {
		graphViewKeySnapshot = graphViewKeySnapshot[len("nfsnap_"):]
	}
	graphViewKey := "network_flow_activity:" + composition.SourceTables[0].IncidentID.String() + ":" + graphViewKeySnapshot
	projector := s.graphProjection
	if projector == nil {
		projector = newGraphProjectionAdapter()
	}
	graphViewID, err := deriveNetworkFlowGraphViewID(graphViewKey)
	if err != nil {
		return nil, graphProjectionFailed("adapter_contract_rejected")
	}
	input := networkFlowProjectionInput(sourceSnapshotID, composition)
	projectionResource, err := projector.ProjectEphemeral(ctx, graphViewID, canonicalJSON(input))
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
			return nil, graphProjectionFailed("projection_cancelled")
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, graphProjectionFailed("projection_timeout")
		}
		var adapterErr *graphProjectionAdapterError
		if errors.As(err, &adapterErr) {
			return nil, graphProjectionFailed(adapterErr.reason)
		}
		return nil, graphProjectionFailed("projection_unavailable")
	}
	summary, ok := projectionResource["validation_summary"].(map[string]any)
	if !ok || summary["fatal_count"] != 0 || summary["error_count"] != 0 || summary["warning_count"] != 0 || summary["info_count"] != 0 {
		return nil, graphProjectionFailed("adapter_contract_rejected")
	}
	return projectionResource, nil
}

func networkFlowProjectionInput(sourceSnapshotID string, composition graphComposition) map[string]any {
	return map[string]any{
		"projection_schema_id": graphProjectionSchemaID,
		"source_snapshot_id":   sourceSnapshotID,
		"projection_config":    networkFlowProjectionConfig(composition.Aggregation.Mode),
		"source_entities":      graphProjectionEntities(composition),
		"source_relationships": graphProjectionRelationships(composition),
		"source_metadata": map[string]any{
			"incident_id":          composition.SourceTables[0].IncidentID.String(),
			"graph_query_digest":   composition.Digest,
			"source_snapshot_id":   sourceSnapshotID,
			"selected_table_ids":   composition.SelectedTableIDs,
			"mapping_fingerprints": orderedFingerprints(composition.SourceTables, nil),
		},
		"filters": map[string]any{
			"entity_filters":       []any{},
			"relationship_filters": []any{},
			"logic":                "and",
		},
		"property_definitions": graphPropertyDefinitions(composition.Aggregation.Mode),
	}
}

func networkFlowProjectionConfig(mode string) map[string]any {
	projectionVersion := "network_flow_activity.v1"
	relationshipKind := "network_flow.flow_edge.v1"
	mappingRuleID := "nf.map.flow_edge.v1"
	if mode == "time_bucket_v1" {
		projectionVersion = "network_flow_activity.time_bucket.v1"
		relationshipKind = "network_flow.bucketed_flow_edge.v1"
		mappingRuleID = "nf.map.bucketed_flow_edge.v1"
	}
	requiredEdgeProperties := []any{"bytes_sum", "contributing_table_ids", "dst_endpoint_id", "dst_port", "edge_id", "example_refs_total_count", "example_refs_truncated", "first_flow_start_utc", "flow_row_count", "ip_protocol", "last_flow_end_utc", "packets_sum", "src_endpoint_id"}
	if mode == "time_bucket_v1" {
		requiredEdgeProperties = append(requiredEdgeProperties, "bucket_start_utc", "bucket_end_utc")
	}
	return map[string]any{
		"projection_version":                 projectionVersion,
		"declared_source_entity_kinds":       []any{"network_flow.ip_endpoint.v1"},
		"declared_source_relationship_kinds": []any{relationshipKind},
		"entity_mappings": []any{
			map[string]any{
				"mapping_rule_id":        "nf.map.ip_endpoint.v1",
				"source_entity_kind":     "network_flow.ip_endpoint.v1",
				"projected_vertex_kind":  "network_flow.ip_endpoint.v1",
				"inclusion_predicate":    "always",
				"label_policy":           "mapping_only",
				"mapping_labels":         []any{},
				"required_property_keys": []any{"contributing_table_ids", "endpoint_kind", "endpoint_value", "flow_row_count", "indicator_candidate_value"},
				"optional_property_keys": []any{},
			},
		},
		"relationship_mappings": []any{
			map[string]any{
				"mapping_rule_id":          mappingRuleID,
				"source_relationship_kind": relationshipKind,
				"projected_edge_kind":      relationshipKind,
				"inclusion_predicate":      "always",
				"direction_policy":         "preserve",
				"emit_reverse_edge":        false,
				"label_policy":             "mapping_only",
				"mapping_labels":           []any{},
				"required_property_keys":   requiredEdgeProperties,
				"optional_property_keys":   []any{},
			},
		},
		"metadata_mappings":         graphMetadataMappings(mode),
		"aggregation_rules":         []any{},
		"default_vertex_labels":     []any{},
		"default_edge_labels":       []any{},
		"allow_empty_kind_registry": false,
	}
}

func graphMetadataMappings(mode string) []any {
	type mapping struct {
		id, scope, kind, source, key, projectedType string
	}
	relationshipKind := "network_flow.flow_edge.v1"
	if mode == "time_bucket_v1" {
		relationshipKind = "network_flow.bucketed_flow_edge.v1"
	}
	items := []mapping{
		{"nf.mm.edge.contributing_table_ids.v1", "edge", relationshipKind, "metadata.contributing_table_ids", "contributing_table_ids", "identifier_array"},
		{"nf.mm.edge.example_refs_total_count.v1", "edge", relationshipKind, "metadata.example_refs_total_count", "example_refs_total_count", "integer"},
		{"nf.mm.edge.mapping_fingerprints.v1", "edge", relationshipKind, "metadata.mapping_fingerprints", "mapping_fingerprints", "identifier_array"},
		{"nf.mm.vertex.contributing_table_ids.v1", "vertex", "network_flow.ip_endpoint.v1", "metadata.contributing_table_ids", "contributing_table_ids", "identifier_array"},
		{"nf.mm.vertex.flow_row_count.v1", "vertex", "network_flow.ip_endpoint.v1", "metadata.flow_row_count", "flow_row_count", "integer"},
		{"nf.mm.vertex.mapping_fingerprints.v1", "vertex", "network_flow.ip_endpoint.v1", "metadata.mapping_fingerprints", "mapping_fingerprints", "identifier_array"},
	}
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"metadata_mapping_id":    item.id,
			"target_scope":           item.scope,
			"target_kind":            item.kind,
			"source_field_path":      item.source,
			"projected_metadata_key": item.key,
			"projected_type":         item.projectedType,
			"required":               true,
			"missing_behavior":       "error",
			"source_null_behavior":   "error",
			"null_output_policy":     "omit",
			"merge_behavior":         "single_value",
		})
	}
	return out
}

func graphPropertyDefinitions(mode string) []any {
	type definition struct {
		scope, kind, key, projectedType, sourceNull, nullOutput string
	}
	items := []definition{}
	vertexKeys := []string{"contributing_table_ids", "endpoint_kind", "endpoint_value", "flow_row_count", "indicator_candidate_value"}
	for _, key := range vertexKeys {
		items = append(items, definition{"vertex", "network_flow.ip_endpoint.v1", key, graphProjectedType(key), "error", "omit"})
	}
	relationshipKind := "network_flow.flow_edge.v1"
	edgeKeys := []string{"bytes_sum", "contributing_table_ids", "dst_endpoint_id", "dst_port", "edge_id", "example_refs_total_count", "example_refs_truncated", "first_flow_start_utc", "flow_row_count", "ip_protocol", "last_flow_end_utc", "packets_sum", "src_endpoint_id"}
	if mode == "time_bucket_v1" {
		relationshipKind = "network_flow.bucketed_flow_edge.v1"
		edgeKeys = append(edgeKeys, "bucket_start_utc", "bucket_end_utc")
	}
	for _, key := range edgeKeys {
		sourceNull := "error"
		nullOutput := "omit"
		if key == "dst_port" {
			sourceNull = "emit_null"
			nullOutput = "emit_null"
		}
		items = append(items, definition{"edge", relationshipKind, key, graphProjectedType(key), sourceNull, nullOutput})
	}
	out := make([]any, 0, len(items))
	for _, item := range items {
		out = append(out, map[string]any{
			"property_definition_id": "nf.pd." + item.scope + "." + item.key + ".v1",
			"target_scope":           item.scope,
			"target_kind":            item.kind,
			"source_field_path":      "properties." + item.key,
			"projected_key":          item.key,
			"projected_type":         item.projectedType,
			"required":               true,
			"missing_behavior":       "error",
			"source_null_behavior":   item.sourceNull,
			"null_output_policy":     item.nullOutput,
			"merge_behavior":         "single_value",
		})
	}
	return out
}

func graphProjectedType(key string) string {
	switch key {
	case "endpoint_kind", "edge_id", "src_endpoint_id", "dst_endpoint_id":
		return "identifier"
	case "endpoint_value", "indicator_candidate_value", "bytes_sum", "packets_sum":
		return "string"
	case "ip_protocol", "dst_port", "flow_row_count", "example_refs_total_count":
		return "integer"
	case "first_flow_start_utc", "last_flow_end_utc", "bucket_start_utc", "bucket_end_utc":
		return "timestamp"
	case "contributing_table_ids":
		return "identifier_array"
	case "example_refs_truncated":
		return "boolean"
	default:
		return "string"
	}
}

func graphProjectionEntities(composition graphComposition) []any {
	ids := sortedGraphVertexIDs(composition.Vertices)
	out := make([]any, 0, len(ids))
	for _, id := range ids {
		vertex := composition.Vertices[id]
		tableIDs := orderedIDsFromSet(composition.SourceTables, vertex.ContributingTableID)
		mappingFingerprints := orderedFingerprints(composition.SourceTables, vertex.MappingFingerprints)
		properties := map[string]any{
			"endpoint_kind":             "ip",
			"endpoint_value":            vertex.EndpointValue,
			"contributing_table_ids":    tableIDs,
			"flow_row_count":            vertex.FlowRowCount,
			"indicator_candidate_value": vertex.EndpointValue,
		}
		out = append(out, map[string]any{
			"source_entity_id":   vertex.EndpointID,
			"source_entity_kind": "network_flow.ip_endpoint.v1",
			"properties":         properties,
			"metadata": map[string]any{
				"contributing_table_ids": tableIDs,
				"mapping_fingerprints":   mappingFingerprints,
				"flow_row_count":         vertex.FlowRowCount,
			},
			"labels": []any{},
		})
	}
	return out
}

func graphProjectionRelationships(composition graphComposition) []any {
	relationshipKind := "network_flow.flow_edge.v1"
	if composition.Aggregation.Mode == "time_bucket_v1" {
		relationshipKind = "network_flow.bucketed_flow_edge.v1"
	}
	ids := sortedGraphEdgeIDs(composition.Edges)
	out := make([]any, 0, len(ids))
	for _, id := range ids {
		edge := composition.Edges[id]
		tableIDs := orderedIDsFromSet(composition.SourceTables, edge.ContributingTableID)
		mappingFingerprints := orderedFingerprints(composition.SourceTables, edge.MappingFingerprints)
		properties := map[string]any{
			"edge_id":                  edge.EdgeID,
			"src_endpoint_id":          edge.SrcEndpointID,
			"dst_endpoint_id":          edge.DstEndpointID,
			"ip_protocol":              int64(edge.IPProtocol),
			"dst_port":                 nullableInt32Value(edge.DstPort),
			"flow_row_count":           edge.FlowRowCount,
			"bytes_sum":                edge.BytesSum.String(),
			"packets_sum":              edge.PacketsSum.String(),
			"first_flow_start_utc":     edge.FirstFlowStartUTC.UTC().Format(time.RFC3339Nano),
			"last_flow_end_utc":        edge.LastFlowEndUTC.UTC().Format(time.RFC3339Nano),
			"contributing_table_ids":   tableIDs,
			"example_refs_truncated":   exampleRefsTruncated(edge, composition.ResultLimits.MaxExampleRowRefsPerEdge),
			"example_refs_total_count": edge.FlowRowCount,
		}
		if edge.BucketStartUTC != nil && edge.BucketEndUTC != nil {
			properties["bucket_start_utc"] = timestamp(*edge.BucketStartUTC)
			properties["bucket_end_utc"] = timestamp(*edge.BucketEndUTC)
		}
		out = append(out, map[string]any{
			"source_relationship_id":   edge.EdgeID,
			"source_relationship_kind": relationshipKind,
			"src_source_entity_id":     edge.SrcEndpointID,
			"dst_source_entity_id":     edge.DstEndpointID,
			"direction":                "forward",
			"properties":               properties,
			"metadata": map[string]any{
				"contributing_table_ids":   tableIDs,
				"mapping_fingerprints":     mappingFingerprints,
				"example_refs_total_count": edge.FlowRowCount,
				"example_refs_truncated":   exampleRefsTruncated(edge, composition.ResultLimits.MaxExampleRowRefsPerEdge),
			},
			"labels": []any{},
		})
	}
	return out
}

func graphEdgeAnnotations(composition graphComposition) []any {
	edgeIDs := sortedGraphEdgeIDs(composition.Edges)
	out := make([]any, 0, len(edgeIDs))
	for _, edgeID := range edgeIDs {
		edge := composition.Edges[edgeID]
		refs := []any{}
		if composition.IncludeExamples {
			for _, row := range edge.ExampleRows {
				refs = append(refs, rowRefResource(row))
			}
		}
		out = append(out, map[string]any{
			"edge_id":                  edge.EdgeID,
			"example_row_refs":         refs,
			"example_refs_truncated":   len(refs) < edge.FlowRowCount,
			"example_refs_total_count": edge.FlowRowCount,
		})
	}
	return out
}

func (s *Service) queryGraphContributors(ctx context.Context, actorID string, sessionID string, incidentID uuid.UUID, request graphContributorQueryRequest) (map[string]any, *httpapi.APIError) {
	var position *contributorCursorPosition
	limit := request.Limit
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalid(reason)
		}
		if payload.Route != routeKeyGraphsContributorsQuery || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalid(payloadMismatchReason(payload, routeKeyGraphsContributorsQuery, actorID, incidentID.String()))
		}
		limit = payload.Limit
		if payload.PositionKind != "contributor_keyset_v1" {
			return nil, cursorInvalid("malformed")
		}
		decodedPosition, err := decodeContributorCursorPosition(payload.Position)
		if err != nil {
			return nil, cursorInvalid("malformed")
		}
		if !sameSortSpecs(decodedPosition.Row.EffectiveSort, effectiveSort(nil)) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
		position = &decodedPosition
		var echo struct {
			GraphQuery       graphSemanticRequest `json:"-"`
			GraphQueryDigest string               `json:"graph_query_digest"`
			Selector         graphSelector        `json:"selector"`
		}
		var rawEcho map[string]json.RawMessage
		if err := json.Unmarshal(payload.QueryEcho, &rawEcho); err != nil {
			return nil, cursorInvalid("malformed")
		}
		semantic, apiErr := decodeGraphSemanticRequest(rawEcho["graph_query"], s.store.limits)
		if apiErr != nil {
			return nil, cursorInvalid("malformed")
		}
		echo.GraphQuery = semantic
		if err := json.Unmarshal(rawEcho["graph_query_digest"], &echo.GraphQueryDigest); err != nil {
			return nil, cursorInvalid("malformed")
		}
		if err := json.Unmarshal(rawEcho["selector"], &echo.Selector); err != nil {
			return nil, cursorInvalid("malformed")
		}
		request.GraphQuery = echo.GraphQuery
		request.GraphQueryDigest = echo.GraphQueryDigest
		request.Selector = echo.Selector
		request.Limit = limit
		if payload.QueryHash != queryHash(graphContributorQueryEcho(request)) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
	}
	rows, hasMore, tableRanks, apiErr := s.queryGraphContributorPage(ctx, incidentID, request.GraphQuery, request.GraphQueryDigest, request.Selector, position, limit)
	if apiErr != nil {
		return nil, apiErr
	}
	contributors := make([]any, 0, len(rows))
	for _, row := range rows {
		contributors = append(contributors, map[string]any{
			"row_ref": rowRefResource(row),
			"row":     rowResource(row),
		})
	}
	queryEcho := graphContributorQueryEcho(request)
	queryEchoRaw, _ := json.Marshal(queryEcho)
	var nextToken *string
	if hasMore && len(rows) > 0 {
		token, err := s.cursorProtector.Encode(CursorBinding{
			Route:       routeKeyGraphsContributorsQuery,
			ActorUserID: actorID,
			SessionID:   sessionID,
			IncidentID:  incidentID.String(),
			Scope:       map[string]string{"graph_query_digest": request.GraphQueryDigest},
			QueryHash:   queryHash(queryEcho),
			QueryEcho:   queryEchoRaw,
			Limit:       limit,
		}, "contributor_keyset_v1", newContributorCursorPosition(rows[len(rows)-1], tableRanks))
		if err != nil {
			return nil, httpapi.InternalAPIError(err)
		}
		nextToken = &token
	}
	return map[string]any{
		"schema_id":          schemaGraphContributorQueryResult,
		"graph_query_digest": request.GraphQueryDigest,
		"selector":           graphSelectorResource(request.Selector),
		"contributors":       contributors,
		"meta": map[string]any{
			"paging": map[string]any{
				"limit":             limit,
				"returned_count":    len(contributors),
				"next_cursor_token": nextToken,
			},
		},
	}, nil
}

func (s *Service) queryGraphContributorPage(ctx context.Context, incidentID uuid.UUID, semantic graphSemanticRequest, expectedDigest string, selector graphSelector, position *contributorCursorPosition, limit int) ([]FlowRow, bool, map[string]int, *httpapi.APIError) {
	_, tableIDs, tableRanks, apiErr := s.resolveGraphTables(ctx, incidentID, TableScope{Mode: "selected_tables", SelectedTableIDs: semantic.SelectedTableIDs})
	if apiErr != nil {
		return nil, false, nil, apiErr
	}
	digest := graphQueryDigestForSemantic(incidentID, tableIDs, semantic)
	if digest != expectedDigest {
		return nil, false, nil, graphQueryStale("digest_mismatch", expectedDigest)
	}
	if apiErr := validateGraphSelectorForSemantic(semantic, selector); apiErr != nil {
		return nil, false, nil, apiErr
	}
	predicate, apiErr := canonicalGraphContributorPredicate(incidentID, selector)
	if apiErr != nil {
		return nil, false, nil, apiErr
	}

	rows := make([]FlowRow, 0, limit+1)
	matchedAny := false
	err := s.store.IterateGraphContributorRows(ctx, incidentID, tableIDs, predicate, func(row FlowRow) error {
		matched, matchErr := rowMatchesGraphQuery(row, semantic.Filters, semantic.TimeRange, semantic.Aggregation)
		if matchErr != nil {
			apiErr = matchErr
			return errStopGraphIteration
		}
		if !matched {
			return nil
		}
		matchedAny = true
		if position != nil {
			rank := tableRanks[row.NetworkFlowTableID]
			if rank < position.WorkspaceTableOrder || rank == position.WorkspaceTableOrder && compareRowToPosition(row, position.Row) <= 0 {
				return nil
			}
		}
		rows = append(rows, row)
		if len(rows) > limit {
			return errStopGraphIteration
		}
		return nil
	})
	if apiErr != nil {
		return nil, false, nil, apiErr
	}
	if err != nil && !errors.Is(err, errStopGraphIteration) {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, false, nil, graphProjectionFailedForContext(err)
		}
		return nil, false, nil, httpapi.InternalAPIError(err)
	}
	if !matchedAny {
		reason := "edge_not_found"
		if predicate.Kind == "vertex" {
			reason = "vertex_not_found"
		}
		return nil, false, nil, graphQueryStale(reason, digest)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	return rows, hasMore, tableRanks, nil
}

func canonicalGraphContributorPredicate(incidentID uuid.UUID, selector graphSelector) (graphContributorPredicate, *httpapi.APIError) {
	if selector.SourceVertexID != "" {
		expected := EndpointID(incidentID, "ip", selector.EndpointValue)
		if selector.SourceVertexID != expected {
			return graphContributorPredicate{}, invalidNetworkFlowRequest("selector.source_vertex_id", "id_key_mismatch")
		}
		return graphContributorPredicate{Kind: "vertex", EndpointValue: selector.EndpointValue}, nil
	}
	if selector.SourceEdgeID != "" {
		srcID := EndpointID(incidentID, "ip", selector.SourceEndpointValue)
		dstID := EndpointID(incidentID, "ip", selector.DestinationEndpointValue)
		expected := FlowEdgeID(incidentID, srcID, dstID, selector.Protocol, selector.DestinationPort)
		predicateKind := "default_edge"
		if selector.Kind == "time_bucket_edge" {
			if selector.BucketStartUTC == nil || selector.BucketEndUTC == nil {
				return graphContributorPredicate{}, invalidNetworkFlowRequest("selector", "variant_member_conflict")
			}
			expected = BucketEdgeID(incidentID, *selector.BucketStartUTC, *selector.BucketEndUTC, srcID, dstID, selector.Protocol, selector.DestinationPort)
			predicateKind = "time_bucket_edge"
		}
		if selector.SourceEdgeID != expected {
			return graphContributorPredicate{}, invalidNetworkFlowRequest("selector.source_edge_id", "id_key_mismatch")
		}
		return graphContributorPredicate{
			Kind:                     predicateKind,
			SourceEndpointValue:      selector.SourceEndpointValue,
			DestinationEndpointValue: selector.DestinationEndpointValue,
			Protocol:                 selector.Protocol,
			DestinationPort:          cloneInt32(selector.DestinationPort),
			BucketStartUTC:           cloneTime(selector.BucketStartUTC),
			BucketEndUTC:             cloneTime(selector.BucketEndUTC),
		}, nil
	}
	return graphContributorPredicate{}, invalidNetworkFlowRequest("selector.kind", "unknown_selector_kind")
}

func validateGraphSelectorForSemantic(semantic graphSemanticRequest, selector graphSelector) *httpapi.APIError {
	if selector.Kind == "vertex" {
		return nil
	}
	if semantic.Aggregation.Mode == "default_flow_edge_v1" {
		if selector.Kind != "default_edge" {
			return invalidNetworkFlowRequest("selector.kind", "variant_member_conflict")
		}
		return nil
	}
	if semantic.Aggregation.Mode != "time_bucket_v1" || selector.Kind != "time_bucket_edge" ||
		selector.BucketStartUTC == nil || selector.BucketEndUTC == nil ||
		semantic.TimeRange.StartUTC == nil || semantic.TimeRange.EndUTC == nil {
		return invalidNetworkFlowRequest("selector.kind", "variant_member_conflict")
	}
	width := time.Duration(semantic.Aggregation.BucketWidthSeconds) * time.Second
	if selector.BucketEndUTC.Sub(*selector.BucketStartUTC) != width ||
		!selector.BucketStartUTC.Equal(floorEpochBucket(*selector.BucketStartUTC, semantic.Aggregation.BucketWidthSeconds)) ||
		!selector.BucketEndUTC.After(*semantic.TimeRange.StartUTC) || !selector.BucketStartUTC.Before(*semantic.TimeRange.EndUTC) {
		return invalidNetworkFlowRequest("selector.bucket_start_utc", "id_key_mismatch")
	}
	return nil
}

func graphContributorQueryEcho(request graphContributorQueryRequest) map[string]any {
	return map[string]any{
		"graph_query":        request.GraphQuery.Raw,
		"graph_query_digest": request.GraphQueryDigest,
		"selector":           graphSelectorResource(request.Selector),
	}
}

func (s *Service) recordGraphQueryAudit(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, composition graphComposition, requestID string) *httpapi.APIError {
	graphDigestSafe, keyID, err := s.safeDigester.Digest("graph_query_digest", composition.Digest)
	if err != nil {
		return httpapi.InternalAPIError(err)
	}
	truncatedCount := 0
	for _, edge := range composition.Edges {
		limit := composition.ResultLimits.MaxExampleRowRefsPerEdge
		if !composition.SemanticQuery["aggregation"].(map[string]any)["include_example_row_refs"].(bool) {
			limit = 0
		}
		if edge.FlowRowCount > limit {
			truncatedCount += edge.FlowRowCount - limit
		}
	}
	err = withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		return s.store.appendAuditEventTx(ctx, tx, networkFlowAuditEvent{
			ActorUserID: &actorUserID,
			IncidentID:  &incidentID,
			EventKind:   "network_flow_graph_query_executed",
			RequestID:   optionalStringPtr(requestID),
			AfterJSON: map[string]any{
				"incident_id":                    incidentID.String(),
				"actor_user_id":                  actorUserID.String(),
				"graph_query_digest_safe":        graphDigestSafe,
				"graph_query_digest_safe_key_id": keyID,
				"selected_table_count":           len(composition.SelectedTableIDs),
				"result_vertex_count":            len(composition.Vertices),
				"result_edge_count":              len(composition.Edges),
				"truncated_example_ref_count":    truncatedCount,
				"network_flow.audit_event_code":  "network_flow_graph_query_executed",
				"network_flow.audit_resource_id": composition.Digest,
			},
		})
	})
	if err != nil {
		return httpapi.InternalAPIError(fmt.Errorf("record network flow graph query audit: %w", err))
	}
	return nil
}

func graphSemanticQueryResource(schemaID string, tableIDs []string, filters []Filter, timeRange graphTimeRange, aggregation graphAggregation, limits graphResultLimits) map[string]any {
	normalizedFilters := filters
	if normalizedFilters == nil {
		normalizedFilters = []Filter{}
	}
	resource := map[string]any{
		"schema_id":          schemaID,
		"selected_table_ids": tableIDs,
		"filters":            normalizedFilters,
		"time_range":         graphTimeRangeResource(timeRange),
		"aggregation":        graphAggregationResource(aggregation),
	}
	return resource
}

func graphTimeRangeResource(timeRange graphTimeRange) map[string]any {
	out := map[string]any{"start_utc": nil, "end_utc": nil}
	if timeRange.StartUTC != nil {
		out["start_utc"] = timeRange.StartUTC.UTC().Format(time.RFC3339Nano)
	}
	if timeRange.EndUTC != nil {
		out["end_utc"] = timeRange.EndUTC.UTC().Format(time.RFC3339Nano)
	}
	return out
}

func graphAggregationResource(aggregation graphAggregation) map[string]any {
	resource := map[string]any{
		"mode":                     aggregation.Mode,
		"include_example_row_refs": aggregation.IncludeExampleRowRefs,
	}
	if aggregation.Mode == "time_bucket_v1" {
		resource["bucket_width_seconds"] = aggregation.BucketWidthSeconds
	}
	return resource
}

func graphSourceTableRefs(tables []TableRecord) []any {
	out := make([]any, 0, len(tables))
	for _, table := range tables {
		out = append(out, map[string]any{
			"network_flow_table_id": table.TableID,
			"table_version":         table.TableVersion,
			"mapping_fingerprint":   table.MappingFingerprint,
			"row_count_accepted":    table.RowCountAccepted,
			"row_count_rejected":    table.RowCountRejected,
		})
	}
	return out
}

func graphSelectorResource(selector graphSelector) map[string]any {
	if selector.SourceVertexID != "" {
		return map[string]any{"kind": "vertex", "source_vertex_id": selector.SourceVertexID, "endpoint_value": selector.EndpointValue}
	}
	if selector.SourceEdgeID != "" {
		out := map[string]any{
			"kind":                       selector.Kind,
			"source_edge_id":             selector.SourceEdgeID,
			"source_endpoint_value":      selector.SourceEndpointValue,
			"destination_endpoint_value": selector.DestinationEndpointValue,
			"protocol":                   int64(selector.Protocol),
			"destination_port_present":   selector.DestinationPortPresent,
		}
		if selector.DestinationPortPresent {
			out["destination_port"] = int64(*selector.DestinationPort)
		}
		if selector.Kind == "time_bucket_edge" && selector.BucketStartUTC != nil && selector.BucketEndUTC != nil {
			out["bucket_start_utc"] = timestamp(*selector.BucketStartUTC)
			out["bucket_end_utc"] = timestamp(*selector.BucketEndUTC)
		}
		return out
	}
	return map[string]any{"kind": selector.Kind}
}

func graphQueryDigestV2(incidentID uuid.UUID, tableIDs []string, filters []Filter, timeRange graphTimeRange, aggregation graphAggregation) string {
	sortedTableIDs := append([]string(nil), tableIDs...)
	sort.Strings(sortedTableIDs)
	normalizedFilters := filters
	if normalizedFilters == nil {
		normalizedFilters = []Filter{}
	}
	var normalizedTimeRange any
	if !(timeRange.Omitted || timeRange.StartUTC == nil && timeRange.EndUTC == nil) {
		normalizedTimeRange = graphTimeRangeResource(timeRange)
	}
	var transcript bytes.Buffer
	for _, field := range []string{
		"cartulary.network_flow.graph_query_digest.v2",
		incidentID.String(),
		string(canonicalJSON(sortedTableIDs)),
		string(canonicalJSON(normalizedFilters)),
		string(canonicalJSON(normalizedTimeRange)),
		"cartulary.network_flow.graph_semantic_query.v2",
		string(canonicalJSON(graphAggregationResource(aggregation))),
	} {
		writeLengthFramedPart(&transcript, field)
	}
	return sha256Hex(transcript.Bytes())
}

func graphSourceSnapshotDigest(incidentID uuid.UUID, tables []TableRecord, graphDigest string) string {
	sortedTables := append([]TableRecord(nil), tables...)
	sort.SliceStable(sortedTables, func(i, j int) bool { return sortedTables[i].TableID < sortedTables[j].TableID })
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow.source_snapshot_digest.v1")
	writeDigestPart(&b, incidentID.String())
	writeDigestPart(&b, strconv.Itoa(len(sortedTables)))
	for _, table := range sortedTables {
		writeDigestPart(&b, table.TableID)
		writeDigestPart(&b, table.MappingFingerprint)
	}
	writeDigestPart(&b, graphDigest)
	return "nfsnap_" + sha256Hex(b.Bytes())
}

func sortedGraphVertexIDs(vertices map[string]*graphVertex) []string {
	out := make([]string, 0, len(vertices))
	for id := range vertices {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func sortedGraphEdgeIDs(edges map[string]*graphEdge) []string {
	out := make([]string, 0, len(edges))
	for id := range edges {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func sortContributorRows(rows []FlowRow, tableRanks map[string]int) {
	effective := effectiveSort(nil)
	sort.SliceStable(rows, func(i, j int) bool {
		leftRank := tableRanks[rows[i].NetworkFlowTableID]
		rightRank := tableRanks[rows[j].NetworkFlowTableID]
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		for _, spec := range effective {
			cmp := compareRowField(rows[i], rows[j], spec.FieldKey)
			if cmp == 0 {
				continue
			}
			if spec.Direction == "desc" {
				return cmp > 0
			}
			return cmp < 0
		}
		return rows[i].RowID < rows[j].RowID
	})
}

func orderedIDsFromSet(tables []TableRecord, set map[string]struct{}) []string {
	out := []string{}
	for _, table := range tables {
		if _, ok := set[table.TableID]; ok {
			out = append(out, table.TableID)
		}
	}
	return out
}

func orderedFingerprints(tables []TableRecord, set map[string]struct{}) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, table := range tables {
		if set != nil {
			if _, ok := set[table.MappingFingerprint]; !ok {
				continue
			}
		}
		if _, exists := seen[table.MappingFingerprint]; exists {
			continue
		}
		seen[table.MappingFingerprint] = struct{}{}
		out = append(out, table.MappingFingerprint)
	}
	return out
}

func exampleRefsTruncated(edge *graphEdge, limit int) bool {
	return edge.FlowRowCount > limit
}

func decimalDigitCount(value string) int {
	value = stringsTrimLeadingSign(value)
	value = stringsTrimLeadingZeros(value)
	if value == "" {
		return 1
	}
	return len(value)
}

func stringsTrimLeadingSign(value string) string {
	if len(value) > 0 && value[0] == '-' {
		return value[1:]
	}
	return value
}

func stringsTrimLeadingZeros(value string) string {
	for len(value) > 0 && value[0] == '0' {
		value = value[1:]
	}
	return value
}

func cloneInt32(value *int32) *int32 {
	if value == nil {
		return nil
	}
	out := *value
	return &out
}

func invalidLimitOverride(field string, reason string, limitKey string, limit int, minimum int, maximum int) *httpapi.APIError {
	details := map[string]any{"reason_code": reason}
	if field != "" {
		details["field"] = field
	}
	if limitKey != "" {
		details["limit_key"] = limitKey
		details["limit"] = limit
		details["minimum"] = minimum
		details["maximum"] = maximum
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: "network_flow_invalid_limit_override", Message: "network_flow_invalid_limit_override", Details: details}
}

func graphLimitExceeded(reason string, limitKey string, limit int, actual int) *httpapi.APIError {
	return graphLimitExceededAt(reason, limitKey, limit, actual, "graph_composition")
}

func counterLimitExceeded(reason string, limitKey string, limit int, actual int) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusRequestEntityTooLarge, Code: "network_flow_counter_sum_limit_exceeded", Message: "network_flow_counter_sum_limit_exceeded", Details: map[string]any{
		"reason_code":  reason,
		"limit_key":    limitKey,
		"limit":        limit,
		"actual":       actual,
		"phase":        "graph_composition",
		"retry_action": "reduce_scope_or_limits",
	}}
}

func graphProjectionFailed(reason string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusBadGateway, Code: "network_flow_graph_projection_failed", Message: "network_flow_graph_projection_failed", Details: map[string]any{
		"reason_code":                 reason,
		"retry_action":                "do_not_retry",
		"projection_contract_version": graphProjectionSchemaID,
	}}
}

func graphProjectionFailedForContext(err error) *httpapi.APIError {
	if errors.Is(err, context.Canceled) {
		return graphProjectionFailed("projection_cancelled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return graphProjectionFailed("projection_timeout")
	}
	return graphProjectionFailed("projection_unavailable")
}

func graphQueryStale(reason string, digest string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "network_flow_graph_query_stale", Message: "network_flow_graph_query_stale", Details: map[string]any{
		"reason_code":        reason,
		"graph_query_digest": digest,
		"retry_action":       "refresh_resource",
	}}
}
