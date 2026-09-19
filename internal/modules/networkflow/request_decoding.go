package networkflow

import "io"

func decodeRenameRequestValue(reader io.Reader) (tableRenameRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembers(raw, "client_txn_id", "base_table_version", "display_name"); apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONString(raw, "client_txn_id")
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	displayName, apiErr := requiredJSONString(raw, "display_name")
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	version, apiErr := decodePositiveInt(raw["base_table_version"], "base_table_version")
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	normalized, err := normalizeExplicitDisplayName(displayName)
	if err != nil {
		return tableRenameRequest{}, tableNameFailure(err)
	}
	return tableRenameRequest{ClientTxnID: clientTxnID, BaseTableVersion: int64(version), DisplayName: normalized}, nil
}

func decodeSoftDeleteRequestValue(reader io.Reader) (tableSoftDeleteRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembers(raw, "client_txn_id", "base_table_version"); apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONString(raw, "client_txn_id")
	if apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	version, apiErr := decodePositiveInt(raw["base_table_version"], "base_table_version")
	if apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	return tableSoftDeleteRequest{ClientTxnID: clientTxnID, BaseTableVersion: int64(version)}, nil
}

func decodeGraphQueryRequestValue(reader io.Reader, limits EffectiveLimits) (graphQueryRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
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

func decodeGraphContributorQueryRequestValue(reader io.Reader, limits EffectiveLimits) (graphContributorQueryRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
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

func decodeGraphViewCreateRequestValue(reader io.Reader, limits EffectiveLimits) (graphViewCreateRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	if schemaID, err := requiredJSONString(raw, "schema_id"); err != nil || schemaID != "cartulary.network_flow.graph_view_create_request.v3" {
		return graphViewCreateRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "client_txn_id", "display_name", "semantic_query"); apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONString(raw, "client_txn_id")
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	displayName, apiErr := requiredJSONString(raw, "display_name")
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	semantic, apiErr := decodeGraphSemanticRequest(raw["semantic_query"], limits)
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	if semantic.SchemaID != schemaGraphSemanticQueryV2 {
		return graphViewCreateRequest{}, invalidNetworkFlowRequest("semantic_query.schema_id", "invalid_schema_id")
	}
	return graphViewCreateRequest{ClientTxnID: clientTxnID, DisplayName: displayName, Semantic: semantic}, nil
}

func decodeGraphViewRenameRequestValue(reader io.Reader) (graphViewRenameRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	if schemaID, err := requiredJSONString(raw, "schema_id"); err != nil || schemaID != "cartulary.network_flow.graph_view_rename_request.v2" {
		return graphViewRenameRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "client_txn_id", "display_name", "base_graph_view_version"); apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONString(raw, "client_txn_id")
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	displayName, apiErr := requiredJSONString(raw, "display_name")
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	version, apiErr := decodePositiveInt(raw["base_graph_view_version"], "base_graph_view_version")
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	return graphViewRenameRequest{graphViewVersionRequest: graphViewVersionRequest{ClientTxnID: clientTxnID, BaseGraphViewVersion: int64(version)}, DisplayName: displayName}, nil
}

func decodeGraphViewVersionRequestValue(reader io.Reader, schemaID string) (graphViewVersionRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	if actual, err := requiredJSONString(raw, "schema_id"); err != nil || actual != schemaID {
		return graphViewVersionRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "client_txn_id", "base_graph_view_version"); apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONString(raw, "client_txn_id")
	if apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	version, apiErr := decodePositiveInt(raw["base_graph_view_version"], "base_graph_view_version")
	if apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	return graphViewVersionRequest{ClientTxnID: clientTxnID, BaseGraphViewVersion: int64(version)}, nil
}

func decodeSavedGraphContributorRequestValue(reader io.Reader, limits EffectiveLimits) (graphViewContributorRequest, *semanticFailure) {
	raw, apiErr := decodeNetworkFlowObject(reader)
	if apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	schemaID, schemaErr := requiredJSONString(raw, "schema_id")
	if schemaErr != nil {
		return graphViewContributorRequest{}, schemaErr
	}
	if schemaID == schemaGraphContributorQueryContinuation {
		if apiErr := ensureAllowedMembers(raw, "schema_id", "cursor_token"); apiErr != nil {
			return graphViewContributorRequest{}, apiErr
		}
		token, tokenErr := requiredJSONString(raw, "cursor_token")
		if tokenErr != nil {
			return graphViewContributorRequest{}, tokenErr
		}
		return graphViewContributorRequest{Continuation: true, CursorToken: token}, nil
	}
	if schemaID != "cartulary.network_flow.graph_view_contributor_query_request.v2" {
		return graphViewContributorRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "projection_result_id", "selector", "limit"); apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	resultID, apiErr := requiredJSONString(raw, "projection_result_id")
	if apiErr != nil || !graphProjectionResultIDPattern.MatchString(resultID) {
		return graphViewContributorRequest{}, invalidNetworkFlowRequest("projection_result_id", "invalid_identifier")
	}
	selector, apiErr := decodeGraphSelector(raw["selector"])
	if apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	if _, present := raw["limit"]; !present {
		return graphViewContributorRequest{}, invalidNetworkFlowRequest("limit", "missing_member")
	}
	limit, apiErr := decodePositiveInt(raw["limit"], "limit")
	if apiErr != nil {
		return graphViewContributorRequest{}, invalidLimit("limit", "not_integer")
	}
	if limit < 1 {
		return graphViewContributorRequest{}, invalidLimit("limit", "below_minimum")
	}
	if limit > 1000 || int64(limit) > limits.MaxQueryLimit {
		return graphViewContributorRequest{}, invalidLimit("limit", "above_maximum")
	}
	return graphViewContributorRequest{ProjectionResultID: resultID, Selector: selector, Limit: limit}, nil
}

const (
	routeKeyTablesPatch  = "nf.tables.patch"
	routeKeyTablesDelete = "nf.tables.delete"
)

type tableRenameRequest struct {
	ClientTxnID      string
	BaseTableVersion int64
	DisplayName      string
}

type tableSoftDeleteRequest struct {
	ClientTxnID      string
	BaseTableVersion int64
}

const (
	routeKeyGraphViewsCreate           = "nf.graph_views.create"
	routeKeyGraphViewsPatch            = "nf.graph_views.patch"
	routeKeyGraphViewsDelete           = "nf.graph_views.delete"
	routeKeyGraphViewsRefresh          = "nf.graph_views.refresh"
	routeKeyGraphViewContributorsQuery = "nf.graph_views.contributors.query"
)

type graphViewContributorRequest struct {
	ProjectionResultID string
	Selector           graphSelector
	Limit              int
	Continuation       bool
	CursorToken        string
}
