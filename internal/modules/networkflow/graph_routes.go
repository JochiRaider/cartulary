package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"net/http"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *routeService) handleGraphQuery(w http.ResponseWriter, r *http.Request) {
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
	composition, apiErr := s.composeGraphHTTP(r.Context(), incidentID, principal.User.ID, request)
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

func (s *routeService) handleGraphContributorsQuery(w http.ResponseWriter, r *http.Request) {
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
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONStringHTTP(raw, "schema_id")
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	if schemaID != schemaGraphQueryRequest {
		return graphQueryRequest{}, invalidNetworkFlowRequestHTTP("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "table_scope", "filters", "time_range", "aggregation", "limit_overrides"); apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	scope, apiErr := requiredTableScopeHTTP(raw["table_scope"], limits)
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	filters, apiErr := decodeFiltersHTTP(raw["filters"], limits)
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	timeRange, apiErr := decodeGraphTimeRangeV2HTTP(raw["time_range"])
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	aggregation, apiErr := decodeGraphAggregationV2HTTP(raw["aggregation"])
	if apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	if apiErr := validateAggregationTimeRangeHTTP(aggregation, timeRange); apiErr != nil {
		return graphQueryRequest{}, apiErr
	}
	resultLimits, apiErr := decodeGraphResultLimitsHTTP(raw["limit_overrides"], limits)
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
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	schemaID, apiErr := requiredJSONStringHTTP(raw, "schema_id")
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	if schemaID == schemaGraphContributorQueryContinuation {
		if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "cursor_token"); apiErr != nil {
			return graphContributorQueryRequest{}, apiErr
		}
		token, apiErr := requiredJSONStringHTTP(raw, "cursor_token")
		if apiErr != nil {
			return graphContributorQueryRequest{}, apiErr
		}
		return graphContributorQueryRequest{Continuation: true, CursorToken: token}, nil
	}
	if schemaID != schemaGraphContributorQueryRequest {
		return graphContributorQueryRequest{}, invalidNetworkFlowRequestHTTP("schema_id", "invalid_schema_id")
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "graph_query", "graph_query_digest", "selector", "limit"); apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	semantic, apiErr := decodeGraphSemanticRequestHTTP(raw["graph_query"], limits)
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	digest, apiErr := requiredJSONStringHTTP(raw, "graph_query_digest")
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	selector, apiErr := decodeGraphSelectorHTTP(raw["selector"])
	if apiErr != nil {
		return graphContributorQueryRequest{}, apiErr
	}
	value, ok := raw["limit"]
	if !ok {
		return graphContributorQueryRequest{}, invalidNetworkFlowRequestHTTP("limit", "missing_member")
	}
	limit, apiErr := decodePositiveIntHTTP(value, "limit")
	if apiErr != nil {
		return graphContributorQueryRequest{}, invalidLimitHTTP("limit", "not_integer")
	}
	if limit < 1 {
		return graphContributorQueryRequest{}, invalidLimitHTTP("limit", "below_minimum")
	}
	if int64(limit) > limits.MaxQueryLimit {
		return graphContributorQueryRequest{}, invalidLimitHTTP("limit", "above_maximum")
	}
	return graphContributorQueryRequest{
		GraphQuery:       semantic,
		GraphQueryDigest: digest,
		Selector:         selector,
		Limit:            limit,
	}, nil
}

func (s *routeService) queryGraphContributors(ctx context.Context, actorID string, sessionID string, incidentID uuid.UUID, request graphContributorQueryRequest) (map[string]any, *httpapi.APIError) {
	var position *contributorCursorPosition
	limit := request.Limit
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalidHTTP(reason)
		}
		if payload.Route != routeKeyGraphsContributorsQuery || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalidHTTP(payloadMismatchReason(payload, routeKeyGraphsContributorsQuery, actorID, incidentID.String()))
		}
		limit = payload.Limit
		if payload.PositionKind != "contributor_keyset_v1" {
			return nil, cursorInvalidHTTP("malformed")
		}
		decodedPosition, err := decodeContributorCursorPosition(payload.Position)
		if err != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		if !sameSortSpecs(decodedPosition.Row.EffectiveSort, effectiveSort(nil)) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
		position = &decodedPosition
		var echo struct {
			GraphQuery       graphSemanticRequest `json:"-"`
			GraphQueryDigest string               `json:"graph_query_digest"`
			Selector         graphSelector        `json:"selector"`
		}
		var rawEcho map[string]json.RawMessage
		if err := json.Unmarshal(payload.QueryEcho, &rawEcho); err != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		semantic, apiErr := decodeGraphSemanticRequestHTTP(rawEcho["graph_query"], s.store.limits)
		if apiErr != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		echo.GraphQuery = semantic
		if err := json.Unmarshal(rawEcho["graph_query_digest"], &echo.GraphQueryDigest); err != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		if err := json.Unmarshal(rawEcho["selector"], &echo.Selector); err != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		request.GraphQuery = echo.GraphQuery
		request.GraphQueryDigest = echo.GraphQueryDigest
		request.Selector = echo.Selector
		request.Limit = limit
		if payload.QueryHash != queryHash(graphContributorQueryEcho(request)) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
	}
	rows, hasMore, tableRanks, apiErr := s.queryGraphContributorPageHTTP(ctx, incidentID, request.GraphQuery, request.GraphQueryDigest, request.Selector, position, limit)
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
		token, err := s.cursorProtector.Encode(cursorBinding{
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

func (s *routeService) recordGraphQueryAudit(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, composition graphComposition, requestID string) *httpapi.APIError {
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
