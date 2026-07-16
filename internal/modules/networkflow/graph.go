package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	schemaGraphQueryRequest                 = "cartulary.network_flow.graph_query_request.v1"
	schemaGraphQueryResult                  = "cartulary.network_flow.graph_query_result.v1"
	schemaGraphSemanticQuery                = "cartulary.network_flow.graph_semantic_query.v1"
	schemaGraphContributorQueryRequest      = "cartulary.network_flow.graph_contributor_query_request.v1"
	schemaGraphContributorQueryContinuation = "cartulary.network_flow.graph_contributor_query_continuation.v1"
	schemaGraphContributorQueryResult       = "cartulary.network_flow.graph_contributor_query_result.v1"
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
}

type graphResultLimits struct {
	MaxVertices               int
	MaxEdges                  int
	MaxExampleRowRefsPerEdge  int
	MaxAggregateCounterDigits int
}

type graphQueryRequest struct {
	TableScope  TableScope
	Filters     []Filter
	TimeRange   graphTimeRange
	Aggregation graphAggregation
	Limits      graphResultLimits
}

type graphSelector struct {
	Kind     string `json:"kind"`
	VertexID string `json:"vertex_id,omitempty"`
	EdgeID   string `json:"edge_id,omitempty"`
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
	SelectedTableIDs []string
	Filters          []Filter
	TimeRange        graphTimeRange
	Aggregation      graphAggregation
	ResultLimits     graphResultLimits
	Raw              map[string]any
}

type graphComposition struct {
	Digest           string
	SemanticQuery    map[string]any
	ResultLimits     graphResultLimits
	SourceTables     []TableRecord
	SourceTableRefs  []any
	TableRanks       map[string]int
	Vertices         map[string]*graphVertex
	Edges            map[string]*graphEdge
	GraphProjection  map[string]any
	EdgeAnnotations  []any
	SelectedTableIDs []string
}

type graphVertex struct {
	EndpointID          string
	EndpointValue       string
	ContributingTableID map[string]struct{}
	MappingFingerprints map[string]struct{}
	Rows                []FlowRow
}

type graphEdge struct {
	EdgeID              string
	SrcEndpointID       string
	DstEndpointID       string
	SrcEndpointValue    string
	DstEndpointValue    string
	IPProtocol          int32
	DstPort             *int32
	Rows                []FlowRow
	BytesSum            big.Int
	PacketsSum          big.Int
	FirstFlowStartUTC   time.Time
	LastFlowEndUTC      time.Time
	ContributingTableID map[string]struct{}
	MappingFingerprints map[string]struct{}
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
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":               schemaGraphQueryResult,
		"graph_query_digest":      composition.Digest,
		"semantic_query":          composition.SemanticQuery,
		"graph_projection_result": composition.GraphProjection,
		"edge_annotations":        composition.EdgeAnnotations,
		"source_table_refs":       composition.SourceTableRefs,
		"result_limits":           graphResultLimitsResource(composition.ResultLimits),
	})
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

func decodeGraphQueryRequest(r *http.Request, limits Limits) (graphQueryRequest, *httpapi.APIError) {
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
	timeRange, apiErr := decodeGraphTimeRange(raw["time_range"])
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	aggregation, apiErr := decodeGraphAggregation(raw["aggregation"])
	if apiErr != nil {
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

func decodeGraphContributorQueryRequest(r *http.Request, limits Limits) (graphContributorQueryRequest, *httpapi.APIError) {
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
	limit := defaultQueryLimit(limits)
	if value, ok := raw["limit"]; ok {
		parsed, apiErr := decodePositiveInt(value, "limit")
		if apiErr != nil {
			return graphContributorQueryRequest{}, invalidLimit("limit", "not_integer")
		}
		if parsed < 1 {
			return graphContributorQueryRequest{}, invalidLimit("limit", "below_minimum")
		}
		if int64(parsed) > limits.MaxQueryLimit {
			return graphContributorQueryRequest{}, invalidLimit("limit", "above_maximum")
		}
		limit = parsed
	}
	return graphContributorQueryRequest{
		GraphQuery:       semantic,
		GraphQueryDigest: digest,
		Selector:         selector,
		Limit:            limit,
	}, nil
}

func decodeGraphTimeRange(raw json.RawMessage) (graphTimeRange, *httpapi.APIError) {
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
	start, apiErr := decodeOptionalTimestamp(object["start_utc"], "start_utc")
	if apiErr != nil {
		return graphTimeRange{}, apiErr
	}
	end, apiErr := decodeOptionalTimestamp(object["end_utc"], "end_utc")
	if apiErr != nil {
		return graphTimeRange{}, apiErr
	}
	if start == nil && end == nil {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "empty_range")
	}
	if start != nil && end != nil && !end.After(*start) {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "empty_range")
	}
	return graphTimeRange{StartUTC: start, EndUTC: end}, nil
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
	if err != nil {
		return nil, invalidNetworkFlowRequest(field, "invalid_timestamp")
	}
	value := parsed.UTC()
	return &value, nil
}

func decodeGraphAggregation(raw json.RawMessage) (graphAggregation, *httpapi.APIError) {
	out := graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true}
	if len(raw) == 0 {
		return out, nil
	}
	if bytes.Equal(raw, []byte("null")) {
		return graphAggregation{}, invalidNetworkFlowRequest("aggregation", "explicit_null")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphAggregation{}, invalidNetworkFlowRequest("aggregation", "type_mismatch")
	}
	if apiErr := ensureAllowedMembers(object, "mode", "include_example_row_refs"); apiErr != nil {
		return graphAggregation{}, invalidNetworkFlowRequest("aggregation", "unknown_member")
	}
	if value, ok := object["mode"]; ok {
		var mode string
		if err := json.Unmarshal(value, &mode); err != nil || mode != "default_flow_edge_v1" {
			return graphAggregation{}, invalidNetworkFlowRequest("aggregation.mode", "invalid_value")
		}
		out.Mode = mode
	}
	if value, ok := object["include_example_row_refs"]; ok {
		var include bool
		if err := json.Unmarshal(value, &include); err != nil {
			return graphAggregation{}, invalidNetworkFlowRequest("aggregation.include_example_row_refs", "type_mismatch")
		}
		out.IncludeExampleRowRefs = include
	}
	return out, nil
}

func decodeGraphResultLimits(raw json.RawMessage, limits Limits) (graphResultLimits, *httpapi.APIError) {
	out := graphResultLimits{
		MaxVertices:               int(limits.MaxGraphVertices),
		MaxEdges:                  int(limits.MaxGraphEdges),
		MaxExampleRowRefsPerEdge:  int(limits.MaxExampleRowRefsPerEdge),
		MaxAggregateCounterDigits: int(limits.MaxAggregateCounterDigits),
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
		"max_vertices":                  {},
		"max_edges":                     {},
		"max_example_row_refs_per_edge": {},
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
		parsed, apiErr := decodeLowerableGraphLimit(value, "max_edges", 1, out.MaxEdges)
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
		if apiErr := ensureAllowedMembers(object, "kind", "vertex_id"); apiErr != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector", "variant_member_conflict")
		}
		vertexID, apiErr := requiredJSONString(object, "vertex_id")
		if apiErr != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector.vertex_id", "missing_member")
		}
		return graphSelector{Kind: kind, VertexID: vertexID}, nil
	case "edge":
		if apiErr := ensureAllowedMembers(object, "kind", "edge_id"); apiErr != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector", "variant_member_conflict")
		}
		edgeID, apiErr := requiredJSONString(object, "edge_id")
		if apiErr != nil {
			return graphSelector{}, invalidNetworkFlowRequest("selector.edge_id", "missing_member")
		}
		return graphSelector{Kind: kind, EdgeID: edgeID}, nil
	default:
		return graphSelector{}, invalidNetworkFlowRequest("selector.kind", "unknown_selector_kind")
	}
}

func decodeGraphSemanticRequest(raw json.RawMessage, limits Limits) (graphSemanticRequest, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query", "missing_member")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query", "type_mismatch")
	}
	if apiErr := ensureAllowedMembers(object, "schema_id", "selected_table_ids", "filters", "time_range", "aggregation", "result_limits"); apiErr != nil {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query", "unknown_member")
	}
	schemaID, apiErr := requiredJSONString(object, "schema_id")
	if apiErr != nil || schemaID != schemaGraphSemanticQuery {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query.schema_id", "invalid_schema_id")
	}
	tableIDs, apiErr := decodeStringArray(object["selected_table_ids"], "selected_table_ids", int(limits.MaxSelectedTablesPerQuery))
	if apiErr != nil || len(tableIDs) == 0 {
		return graphSemanticRequest{}, invalidTableScope("selected_table_ids", "empty_resolved_scope")
	}
	filters, apiErr := decodeFilters(object["filters"], limits)
	if apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	timeRange, apiErr := decodeSemanticTimeRange(object["time_range"])
	if apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	aggregation, apiErr := decodeGraphAggregation(object["aggregation"])
	if apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	resultLimits, apiErr := decodeSemanticResultLimits(object["result_limits"])
	if apiErr != nil {
		return graphSemanticRequest{}, apiErr
	}
	var rawObject map[string]any
	if err := json.Unmarshal(raw, &rawObject); err != nil {
		return graphSemanticRequest{}, invalidNetworkFlowRequest("graph_query", "type_mismatch")
	}
	return graphSemanticRequest{
		SelectedTableIDs: tableIDs,
		Filters:          filters,
		TimeRange:        timeRange,
		Aggregation:      aggregation,
		ResultLimits:     resultLimits,
		Raw:              rawObject,
	}, nil
}

func decodeSemanticTimeRange(raw json.RawMessage) (graphTimeRange, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return graphTimeRange{Omitted: true}, nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "type_mismatch")
	}
	if apiErr := ensureAllowedMembers(object, "start_utc", "end_utc"); apiErr != nil {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "unknown_member")
	}
	start, apiErr := decodeOptionalTimestamp(object["start_utc"], "start_utc")
	if apiErr != nil {
		return graphTimeRange{}, apiErr
	}
	end, apiErr := decodeOptionalTimestamp(object["end_utc"], "end_utc")
	if apiErr != nil {
		return graphTimeRange{}, apiErr
	}
	if start != nil && end != nil && !end.After(*start) {
		return graphTimeRange{}, invalidNetworkFlowRequest("time_range", "empty_range")
	}
	return graphTimeRange{StartUTC: start, EndUTC: end, Omitted: start == nil && end == nil}, nil
}

func decodeSemanticResultLimits(raw json.RawMessage) (graphResultLimits, *httpapi.APIError) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return graphResultLimits{}, invalidNetworkFlowRequest("result_limits", "missing_member")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return graphResultLimits{}, invalidNetworkFlowRequest("result_limits", "type_mismatch")
	}
	if apiErr := ensureAllowedMembers(object, "max_vertices", "max_edges", "max_example_row_refs_per_edge", "max_aggregate_counter_digits"); apiErr != nil {
		return graphResultLimits{}, invalidNetworkFlowRequest("result_limits", "unknown_member")
	}
	read := func(field string, minimum int) (int, *httpapi.APIError) {
		parsed, apiErr := decodePositiveInt(object[field], field)
		if apiErr != nil || parsed < minimum {
			return 0, invalidLimit(field, "invalid_value")
		}
		return parsed, nil
	}
	maxVertices, apiErr := read("max_vertices", 1)
	if apiErr != nil {
		return graphResultLimits{}, apiErr
	}
	maxEdges, apiErr := read("max_edges", 1)
	if apiErr != nil {
		return graphResultLimits{}, apiErr
	}
	maxExamples, apiErr := read("max_example_row_refs_per_edge", 0)
	if apiErr != nil {
		return graphResultLimits{}, apiErr
	}
	maxDigits, apiErr := read("max_aggregate_counter_digits", 1)
	if apiErr != nil {
		return graphResultLimits{}, apiErr
	}
	return graphResultLimits{MaxVertices: maxVertices, MaxEdges: maxEdges, MaxExampleRowRefsPerEdge: maxExamples, MaxAggregateCounterDigits: maxDigits}, nil
}

func (s *Service) composeGraph(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, request graphQueryRequest) (graphComposition, *httpapi.APIError) {
	tables, tableIDs, tableRanks, apiErr := s.resolveGraphTables(ctx, incidentID, request.TableScope)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	rows, err := s.store.ListRowsForTables(ctx, incidentID, tableIDs)
	if err != nil {
		return graphComposition{}, httpapi.InternalAPIError(err)
	}
	filtered, apiErr := filterRows(rows, request.Filters)
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	filtered = filterRowsByGraphTimeRange(filtered, request.TimeRange)
	sortContributorRows(filtered, tableRanks)
	digest := graphQueryDigest(incidentID, tableIDs, request.Filters, request.TimeRange, request.Aggregation)
	composition := graphComposition{
		Digest:           digest,
		ResultLimits:     request.Limits,
		SourceTables:     tables,
		SourceTableRefs:  graphSourceTableRefs(tables),
		TableRanks:       tableRanks,
		Vertices:         map[string]*graphVertex{},
		Edges:            map[string]*graphEdge{},
		SelectedTableIDs: tableIDs,
	}
	tableByID := make(map[string]TableRecord, len(tables))
	for _, table := range tables {
		tableByID[table.TableID] = table
	}
	if apiErr := composeGraphObjects(incidentID, filtered, tableByID, &composition); apiErr != nil {
		return graphComposition{}, apiErr
	}
	if apiErr := validateGraphLimits(composition); apiErr != nil {
		return graphComposition{}, apiErr
	}
	composition.SemanticQuery = graphSemanticQueryResource(tableIDs, request.Filters, request.TimeRange, request.Aggregation, request.Limits)
	sourceSnapshotID := graphSourceSnapshotDigest(incidentID, tables, digest)
	projection, apiErr := s.projectNetworkFlowGraph(ctx, actorUserID, sourceSnapshotID, composition, s.now())
	if apiErr != nil {
		return graphComposition{}, apiErr
	}
	composition.GraphProjection = projection
	composition.EdgeAnnotations = graphEdgeAnnotations(composition)
	return composition, nil
}

func (s *Service) composeGraphFromSemantic(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, semantic graphSemanticRequest) (graphComposition, *httpapi.APIError) {
	request := graphQueryRequest{
		TableScope:  TableScope{Mode: "selected_tables", SelectedTableIDs: semantic.SelectedTableIDs},
		Filters:     semantic.Filters,
		TimeRange:   semantic.TimeRange,
		Aggregation: semantic.Aggregation,
		Limits:      semantic.ResultLimits,
	}
	return s.composeGraph(ctx, incidentID, actorUserID, request)
}

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

func filterRowsByGraphTimeRange(rows []FlowRow, timeRange graphTimeRange) []FlowRow {
	if timeRange.Omitted || (timeRange.StartUTC == nil && timeRange.EndUTC == nil) {
		return rows
	}
	out := make([]FlowRow, 0, len(rows))
	for _, row := range rows {
		if timeRange.StartUTC != nil && row.FlowEndUTC.Before(*timeRange.StartUTC) {
			continue
		}
		if timeRange.EndUTC != nil && !row.FlowStartUTC.Before(*timeRange.EndUTC) {
			continue
		}
		out = append(out, row)
	}
	return out
}

func composeGraphObjects(incidentID uuid.UUID, rows []FlowRow, tableByID map[string]TableRecord, composition *graphComposition) *httpapi.APIError {
	for _, row := range rows {
		srcID := EndpointID(incidentID, "ip", row.SrcIP)
		dstID := EndpointID(incidentID, "ip", row.DstIP)
		srcVertex := ensureGraphVertex(composition.Vertices, srcID, row.SrcIP)
		dstVertex := ensureGraphVertex(composition.Vertices, dstID, row.DstIP)
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
		}
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
		edge.Rows = append(edge.Rows, row)
		edge.ContributingTableID[row.NetworkFlowTableID] = struct{}{}
		if table, ok := tableByID[row.NetworkFlowTableID]; ok {
			edge.MappingFingerprints[table.MappingFingerprint] = struct{}{}
		}
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
	vertex.Rows = append(vertex.Rows, row)
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
		projector = newGraphProjectionAdapter(func() time.Time { return requestedAt })
	}
	graphViewID, err := projector.GraphViewID(graphViewKey)
	if err != nil {
		return nil, graphProjectionFailed("adapter_contract_rejected")
	}
	input := networkFlowProjectionInput(graphViewID, graphViewKey, actorUserID, sourceSnapshotID, composition, requestedAt)
	projectionResource, err := projector.ProjectEphemeral(ctx, canonicalJSON(input))
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

func networkFlowProjectionInput(graphViewID string, graphViewKey string, actorUserID uuid.UUID, sourceSnapshotID string, composition graphComposition, requestedAt time.Time) map[string]any {
	return map[string]any{
		"projection_schema_id": graphProjectionSchemaID,
		"graph_view_id":        graphViewID,
		"source_snapshot_id":   sourceSnapshotID,
		"projection_config":    networkFlowProjectionConfig(graphViewKey),
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
		"relationship_definitions": []any{},
		"property_definitions":     graphPropertyDefinitions(),
		"requested_at":             graphProjectionTimestamp(requestedAt),
		"requested_by":             actorUserID.String(),
	}
}

func graphProjectionTimestamp(value time.Time) string {
	return value.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)
}

func networkFlowProjectionConfig(graphViewKey string) map[string]any {
	return map[string]any{
		"graph_view_key":                     graphViewKey,
		"projection_version":                 "network_flow_activity.v1",
		"declared_source_entity_kinds":       []any{"network_flow.ip_endpoint.v1"},
		"declared_source_relationship_kinds": []any{"network_flow.flow_edge.v1"},
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
				"mapping_rule_id":          "nf.map.flow_edge.v1",
				"source_relationship_kind": "network_flow.flow_edge.v1",
				"projected_edge_kind":      "network_flow.flow_edge.v1",
				"inclusion_predicate":      "always",
				"direction_policy":         "preserve",
				"emit_reverse_edge":        false,
				"label_policy":             "mapping_only",
				"mapping_labels":           []any{},
				"required_property_keys":   []any{"bytes_sum", "contributing_table_ids", "dst_endpoint_id", "dst_port", "edge_id", "example_refs_total_count", "example_refs_truncated", "first_flow_start_utc", "flow_row_count", "ip_protocol", "last_flow_end_utc", "packets_sum", "src_endpoint_id"},
				"optional_property_keys":   []any{},
			},
		},
		"metadata_mappings":         graphMetadataMappings(),
		"aggregation_rules":         []any{},
		"default_vertex_labels":     []any{},
		"default_edge_labels":       []any{},
		"allow_empty_kind_registry": false,
		"retention_policy": map[string]any{
			"retain_replaced_results":           false,
			"retention_count":                   0,
			"retention_duration_seconds":        0,
			"retain_failed_results":             false,
			"failed_retention_count":            0,
			"failed_retention_duration_seconds": 0,
		},
		"custom_config": map[string]any{},
	}
}

func graphMetadataMappings() []any {
	type mapping struct {
		id, scope, kind, source, key, projectedType string
	}
	items := []mapping{
		{"nf.mm.edge.contributing_table_ids.v1", "edge", "network_flow.flow_edge.v1", "metadata.contributing_table_ids", "contributing_table_ids", "identifier_array"},
		{"nf.mm.edge.example_refs_total_count.v1", "edge", "network_flow.flow_edge.v1", "metadata.example_refs_total_count", "example_refs_total_count", "integer"},
		{"nf.mm.edge.mapping_fingerprints.v1", "edge", "network_flow.flow_edge.v1", "metadata.mapping_fingerprints", "mapping_fingerprints", "identifier_array"},
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

func graphPropertyDefinitions() []any {
	type definition struct {
		scope, kind, key, projectedType, sourceNull, nullOutput string
	}
	items := []definition{}
	vertexKeys := []string{"contributing_table_ids", "endpoint_kind", "endpoint_value", "flow_row_count", "indicator_candidate_value"}
	for _, key := range vertexKeys {
		items = append(items, definition{"vertex", "network_flow.ip_endpoint.v1", key, graphProjectedType(key), "error", "omit"})
	}
	edgeKeys := []string{"bytes_sum", "contributing_table_ids", "dst_endpoint_id", "dst_port", "edge_id", "example_refs_total_count", "example_refs_truncated", "first_flow_start_utc", "flow_row_count", "ip_protocol", "last_flow_end_utc", "packets_sum", "src_endpoint_id"}
	for _, key := range edgeKeys {
		sourceNull := "error"
		nullOutput := "omit"
		if key == "dst_port" {
			sourceNull = "emit_null"
			nullOutput = "emit_null"
		}
		items = append(items, definition{"edge", "network_flow.flow_edge.v1", key, graphProjectedType(key), sourceNull, nullOutput})
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
	case "first_flow_start_utc", "last_flow_end_utc":
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
			"flow_row_count":            len(vertex.Rows),
			"indicator_candidate_value": vertex.EndpointValue,
		}
		out = append(out, map[string]any{
			"source_entity_id":   vertex.EndpointID,
			"source_entity_kind": "network_flow.ip_endpoint.v1",
			"properties":         properties,
			"metadata": map[string]any{
				"contributing_table_ids": tableIDs,
				"mapping_fingerprints":   mappingFingerprints,
				"flow_row_count":         len(vertex.Rows),
			},
			"labels": []any{},
		})
	}
	return out
}

func graphProjectionRelationships(composition graphComposition) []any {
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
			"flow_row_count":           len(edge.Rows),
			"bytes_sum":                edge.BytesSum.String(),
			"packets_sum":              edge.PacketsSum.String(),
			"first_flow_start_utc":     edge.FirstFlowStartUTC.UTC().Format(time.RFC3339Nano),
			"last_flow_end_utc":        edge.LastFlowEndUTC.UTC().Format(time.RFC3339Nano),
			"contributing_table_ids":   tableIDs,
			"example_refs_truncated":   exampleRefsTruncated(edge, composition.ResultLimits.MaxExampleRowRefsPerEdge),
			"example_refs_total_count": len(edge.Rows),
		}
		out = append(out, map[string]any{
			"source_relationship_id":   edge.EdgeID,
			"source_relationship_kind": "network_flow.flow_edge.v1",
			"src_source_entity_id":     edge.SrcEndpointID,
			"dst_source_entity_id":     edge.DstEndpointID,
			"direction":                "forward",
			"properties":               properties,
			"metadata": map[string]any{
				"contributing_table_ids":   tableIDs,
				"mapping_fingerprints":     mappingFingerprints,
				"example_refs_total_count": len(edge.Rows),
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
		if composition.SemanticQuery["aggregation"].(map[string]any)["include_example_row_refs"].(bool) {
			limit := composition.ResultLimits.MaxExampleRowRefsPerEdge
			if limit > len(edge.Rows) {
				limit = len(edge.Rows)
			}
			for _, row := range edge.Rows[:limit] {
				refs = append(refs, rowRefResource(row))
			}
		}
		out = append(out, map[string]any{
			"edge_id":                  edge.EdgeID,
			"example_row_refs":         refs,
			"example_refs_truncated":   len(refs) < len(edge.Rows),
			"example_refs_total_count": len(edge.Rows),
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
	composition, apiErr := s.composeGraphFromSemantic(ctx, incidentID, uuid.MustParse(actorID), request.GraphQuery)
	if apiErr != nil {
		return nil, apiErr
	}
	if composition.Digest != request.GraphQueryDigest {
		return nil, graphQueryStale("digest_mismatch", request.GraphQueryDigest)
	}
	rows, apiErr := graphContributorRows(composition, request.Selector)
	if apiErr != nil {
		return nil, apiErr
	}
	page, hasMore := pageContributorRowsAfter(rows, composition.TableRanks, position, limit)
	contributors := make([]any, 0, len(page))
	for _, row := range page {
		contributors = append(contributors, map[string]any{
			"row_ref": rowRefResource(row),
			"row":     rowResource(row),
		})
	}
	queryEcho := graphContributorQueryEcho(request)
	queryEchoRaw, _ := json.Marshal(queryEcho)
	var nextToken *string
	if hasMore && len(page) > 0 {
		token, err := s.cursorProtector.Encode(CursorBinding{
			Route:       routeKeyGraphsContributorsQuery,
			ActorUserID: actorID,
			SessionID:   sessionID,
			IncidentID:  incidentID.String(),
			Scope:       map[string]string{"graph_query_digest": request.GraphQueryDigest},
			QueryHash:   queryHash(queryEcho),
			QueryEcho:   queryEchoRaw,
			Limit:       limit,
		}, "contributor_keyset_v1", newContributorCursorPosition(page[len(page)-1], composition.TableRanks))
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

func graphContributorRows(composition graphComposition, selector graphSelector) ([]FlowRow, *httpapi.APIError) {
	switch selector.Kind {
	case "vertex":
		vertex := composition.Vertices[selector.VertexID]
		if vertex == nil {
			return nil, graphQueryStale("vertex_not_found", composition.Digest)
		}
		rows := append([]FlowRow(nil), vertex.Rows...)
		sortContributorRows(rows, composition.TableRanks)
		return rows, nil
	case "edge":
		edge := composition.Edges[selector.EdgeID]
		if edge == nil {
			return nil, graphQueryStale("edge_not_found", composition.Digest)
		}
		rows := append([]FlowRow(nil), edge.Rows...)
		sortContributorRows(rows, composition.TableRanks)
		return rows, nil
	default:
		return nil, invalidNetworkFlowRequest("selector.kind", "unknown_selector_kind")
	}
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
		if len(edge.Rows) > limit {
			truncatedCount += len(edge.Rows) - limit
		}
	}
	err = s.store.transactionRun.WithinTx(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
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

func graphSemanticQueryResource(tableIDs []string, filters []Filter, timeRange graphTimeRange, aggregation graphAggregation, limits graphResultLimits) map[string]any {
	normalizedFilters := filters
	if normalizedFilters == nil {
		normalizedFilters = []Filter{}
	}
	return map[string]any{
		"schema_id":          schemaGraphSemanticQuery,
		"selected_table_ids": tableIDs,
		"filters":            normalizedFilters,
		"time_range":         graphTimeRangeResource(timeRange),
		"aggregation":        graphAggregationResource(aggregation),
		"result_limits":      graphResultLimitsResource(limits),
	}
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
	return map[string]any{
		"mode":                     aggregation.Mode,
		"include_example_row_refs": aggregation.IncludeExampleRowRefs,
	}
}

func graphResultLimitsResource(limits graphResultLimits) map[string]any {
	return map[string]any{
		"max_vertices":                  limits.MaxVertices,
		"max_edges":                     limits.MaxEdges,
		"max_example_row_refs_per_edge": limits.MaxExampleRowRefsPerEdge,
		"max_aggregate_counter_digits":  limits.MaxAggregateCounterDigits,
	}
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
	if selector.Kind == "vertex" {
		return map[string]any{"kind": selector.Kind, "vertex_id": selector.VertexID}
	}
	return map[string]any{"kind": selector.Kind, "edge_id": selector.EdgeID}
}

func graphQueryDigest(incidentID uuid.UUID, tableIDs []string, filters []Filter, timeRange graphTimeRange, aggregation graphAggregation) string {
	sortedTableIDs := append([]string(nil), tableIDs...)
	sort.Strings(sortedTableIDs)
	normalizedFilters := filters
	if normalizedFilters == nil {
		normalizedFilters = []Filter{}
	}
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow.graph_query_digest.v1")
	writeDigestPart(&b, incidentID.String())
	b.Write(canonicalJSON(map[string]any{"table_ids": sortedTableIDs}))
	b.WriteByte(0)
	b.Write(canonicalJSON(normalizedFilters))
	b.WriteByte(0)
	if timeRange.Omitted || (timeRange.StartUTC == nil && timeRange.EndUTC == nil) {
		b.Write(canonicalJSON(nil))
	} else {
		b.Write(canonicalJSON(graphTimeRangeResource(timeRange)))
	}
	b.WriteByte(0)
	b.Write(canonicalJSON(graphAggregationResource(aggregation)))
	b.WriteByte(0)
	sum := sha256.Sum256(b.Bytes())
	return hex.EncodeToString(sum[:])
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
	return len(edge.Rows) > limit
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
	return &httpapi.APIError{Status: http.StatusRequestEntityTooLarge, Code: "network_flow_graph_limit_exceeded", Message: "network_flow_graph_limit_exceeded", Details: map[string]any{
		"reason_code": reason,
		"limit_key":   limitKey,
		"limit":       limit,
		"actual":      actual,
		"phase":       "graph_composition",
	}}
}

func counterLimitExceeded(reason string, limitKey string, limit int, actual int) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusRequestEntityTooLarge, Code: "network_flow_counter_sum_limit_exceeded", Message: "network_flow_counter_sum_limit_exceeded", Details: map[string]any{
		"reason_code": reason,
		"limit_key":   limitKey,
		"limit":       limit,
		"actual":      actual,
		"phase":       "graph_composition",
	}}
}

func graphProjectionFailed(reason string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusBadGateway, Code: "network_flow_graph_projection_failed", Message: "network_flow_graph_projection_failed", Details: map[string]any{
		"reason_code":                 reason,
		"retry_action":                "do_not_retry",
		"projection_contract_version": "graph_projection.v1",
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
