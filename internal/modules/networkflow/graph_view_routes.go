package networkflow

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

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

func (s *routeService) handleGraphViewsCollection(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := parseIncidentPathValue(w, r)
	if !ok {
		return
	}
	stateChanging := r.Method == http.MethodPost
	principal, apiErr := s.authenticate(r, stateChanging)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesMember, ""); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if r.URL.RawQuery != "" {
			writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
			return
		}
		declarations, err := s.store.ListActiveGraphViewDeclarations(r.Context(), incidentID)
		if err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		resources := make([]any, 0, len(declarations))
		for _, declaration := range declarations {
			resources = append(resources, graphViewResource(declaration))
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
			"schema_id": "cartulary.network_flow.graph_view_list.v4", "graph_views": resources,
		})
	case http.MethodPost:
		if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if r.URL.RawQuery != "" {
			writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
			return
		}
		request, apiErr := decodeGraphViewCreateRequest(r, s.store.limits)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		outcome, commandErr := s.savedGraphs.commitGraphViewCreate(
			r.Context(), incidentID, principal.User.ID, request,
			httpapi.RequestIDFromContext(r.Context()),
		)
		if commandErr != nil {
			writeAPIError(w, r, graphViewMutationAPIError(commandErr, request.ClientTxnID))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, savedGraphOutcomeStatus(outcome.kind), savedGraphOutcomePayload(outcome))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *routeService) handleGraphViewResource(w http.ResponseWriter, r *http.Request) {
	incidentID, graphViewID, ok := parseIncidentGraphViewPathValues(w, r)
	if !ok {
		return
	}
	stateChanging := r.Method == http.MethodPatch || r.Method == http.MethodDelete
	principal, apiErr := s.authenticate(r, stateChanging)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesMember, ""); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if r.URL.RawQuery != "" {
			writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
			return
		}
		declaration, apiErr := s.activeGraphView(r.Context(), incidentID, graphViewID)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
			"schema_id":  "cartulary.network_flow.graph_view_get.v4",
			"graph_view": graphViewResource(declaration),
		})
	case http.MethodPatch:
		if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if r.URL.RawQuery != "" {
			writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
			return
		}
		request, apiErr := decodeGraphViewRenameRequest(r)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		outcome, commandErr := s.savedGraphs.commitGraphViewRename(r.Context(), incidentID, graphViewID, principal.User.ID, request, httpapi.RequestIDFromContext(r.Context()))
		if commandErr != nil {
			writeAPIError(w, r, graphViewMutationAPIError(commandErr, request.ClientTxnID))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, savedGraphOutcomeStatus(outcome.kind), savedGraphOutcomePayload(outcome))
	case http.MethodDelete:
		if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesReviewerAdmin, "reviewer|admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if r.URL.RawQuery != "" {
			writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
			return
		}
		request, apiErr := decodeGraphViewVersionRequest(r, "cartulary.network_flow.graph_view_retire_request.v1")
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		outcome, commandErr := s.savedGraphs.commitGraphViewRetire(r.Context(), incidentID, graphViewID, principal.User.ID, request, httpapi.RequestIDFromContext(r.Context()))
		if commandErr != nil {
			writeAPIError(w, r, graphViewMutationAPIError(commandErr, request.ClientTxnID))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		w.WriteHeader(savedGraphOutcomeStatus(outcome.kind))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *routeService) handleGraphViewRefresh(w http.ResponseWriter, r *http.Request) {
	incidentID, graphViewID, ok := parseIncidentGraphViewPathValues(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if r.URL.RawQuery != "" {
		writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
		return
	}
	request, apiErr := decodeGraphViewVersionRequest(r, "cartulary.network_flow.graph_view_refresh_request.v1")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, commandErr := s.savedGraphs.commitGraphViewRefresh(r.Context(), incidentID, graphViewID, principal.User.ID, request, httpapi.RequestIDFromContext(r.Context()))
	if commandErr != nil {
		writeAPIError(w, r, graphViewMutationAPIError(commandErr, request.ClientTxnID))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, savedGraphOutcomeStatus(outcome.kind), savedGraphOutcomePayload(outcome))
}

func (s *routeService) handleGraphViewResult(w http.ResponseWriter, r *http.Request) {
	incidentID, graphViewID, ok := parseIncidentGraphViewPathValues(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesMember, ""); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if r.URL.RawQuery != "" {
		writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
		return
	}
	declaration, apiErr := s.activeGraphView(r.Context(), incidentID, graphViewID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if declaration.SelectedResult == nil {
		status, _ := s.graphViewMaterializationStatus(r.Context(), declaration)
		writeAPIError(w, r, graphViewNotMaterialized(status))
		return
	}
	reader, err := postgresresult.NewReader(s.store.pool)
	if err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	completed, err := reader.ReadExactResult(r.Context(), graphViewResultBinding(declaration))
	if err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	var result map[string]any
	if err := json.Unmarshal(completed.ResultJSON, &result); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	semantic, apiErr := decodeGraphSemanticRequestHTTP(declaration.SemanticQueryJSON, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, malformedStoredGraphResultHTTP())
		return
	}
	readerComposer := *s.graphComposer
	readerComposer.graphTelemetry = nil
	composition, compositionErr := readerComposer.composeGraphSourceFromSemantic(r.Context(), incidentID, semantic)
	apiErr = semanticHTTPError(compositionErr)
	if apiErr != nil || graphSourceSnapshotDigest(incidentID, composition.SourceTables, composition.Digest) != declaration.SelectedResult.SourceSnapshotID {
		writeAPIError(w, r, malformedStoredGraphResultHTTP())
		return
	}
	composition.GraphProjection = result
	if semantic.SchemaID == schemaGraphSemanticQueryV2 {
		if semantic.Aggregation.Mode == "time_bucket_v1" {
			composition.TimeBuckets, apiErr = deriveTimeBucketIndexFromExactResultHTTP(
				semantic.TimeRange, semantic.Aggregation.BucketWidthSeconds, composition.ResultLimits.MaxTimeBuckets, result,
			)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
		}
		if apiErr := bindGraphV2ResponseMetadataHTTP(&composition); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
	} else {
		composition.EdgeAnnotations = graphEdgeAnnotations(composition)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":  "cartulary.network_flow.graph_view_result.v4",
		"graph_view": graphViewResource(declaration), "result": graphQueryResultResource(composition),
	})
}

func (s *routeService) handleGraphViewContributorsQuery(w http.ResponseWriter, r *http.Request) {
	incidentID, graphViewID, ok := parseIncidentGraphViewPathValues(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireSavedGraphRole(r.Context(), incidentID, principal.User.ID, admission.RolesMember, ""); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if r.URL.RawQuery != "" {
		writeAPIError(w, r, invalidNetworkFlowRequestHTTP("query", "unknown_member"))
		return
	}
	declaration, apiErr := s.activeGraphView(r.Context(), incidentID, graphViewID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if declaration.SelectedResult == nil {
		status, _ := s.graphViewMaterializationStatus(r.Context(), declaration)
		writeAPIError(w, r, graphViewNotMaterialized(status))
		return
	}
	request, apiErr := decodeSavedGraphContributorRequest(r, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	semantic, apiErr := decodeGraphSemanticRequestHTTP(declaration.SemanticQueryJSON, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(errors.New("stored graph semantic query invalid")))
		return
	}
	digest := graphQueryDigestForSemantic(incidentID, semantic.SelectedTableIDs, semantic)
	result, apiErr := s.querySavedGraphContributors(
		r.Context(), principal.User.ID.String(), principal.Session.ID.String(), incidentID, graphViewID,
		declaration.SelectedResult.ProjectionResultID, semantic, digest, request,
	)
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

func (s *routeService) querySavedGraphContributors(
	ctx context.Context,
	actorID string,
	sessionID string,
	incidentID uuid.UUID,
	graphViewID string,
	selectedResultID string,
	semantic graphSemanticRequest,
	digest string,
	request graphViewContributorRequest,
) (map[string]any, *httpapi.APIError) {
	var position *contributorCursorPosition
	limit := request.Limit
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalidHTTP(reason)
		}
		if payload.Route != routeKeyGraphViewContributorsQuery || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalidHTTP(payloadMismatchReason(payload, routeKeyGraphViewContributorsQuery, actorID, incidentID.String()))
		}
		if payload.Scope["graph_view_id"] != graphViewID || payload.Scope["projection_result_id"] != selectedResultID || payload.Scope["graph_query_digest"] != digest {
			return nil, cursorInvalidHTTP("scope_stale")
		}
		if payload.PositionKind != "contributor_keyset_v1" {
			return nil, cursorInvalidHTTP("malformed")
		}
		decodedPosition, err := decodeContributorCursorPosition(payload.Position)
		if err != nil || !sameSortSpecs(decodedPosition.Row.EffectiveSort, effectiveSort(nil)) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
		position = &decodedPosition
		limit = payload.Limit
		var echo struct {
			GraphViewID        string          `json:"graph_view_id"`
			ProjectionResultID string          `json:"projection_result_id"`
			GraphQueryDigest   string          `json:"graph_query_digest"`
			Selector           json.RawMessage `json:"selector"`
		}
		if err := json.Unmarshal(payload.QueryEcho, &echo); err != nil || echo.GraphViewID != graphViewID || echo.ProjectionResultID != selectedResultID || echo.GraphQueryDigest != digest {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
		selector, apiErr := decodeGraphSelectorHTTP(echo.Selector)
		if apiErr != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		request.ProjectionResultID = echo.ProjectionResultID
		request.Selector = selector
		request.Limit = limit
		if payload.QueryHash != queryHash(savedGraphContributorQueryEcho(graphViewID, digest, request)) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
	} else if request.ProjectionResultID != selectedResultID {
		return nil, graphQueryStaleHTTP("digest_mismatch", request.ProjectionResultID)
	}

	rows, hasMore, tableRanks, apiErr := s.queryGraphContributorPageHTTP(ctx, incidentID, semantic, digest, request.Selector, position, limit)
	if apiErr != nil {
		return nil, apiErr
	}
	contributors := make([]any, 0, len(rows))
	for _, row := range rows {
		contributors = append(contributors, map[string]any{"row_ref": rowRefResource(row), "row": rowResource(row)})
	}
	echo := savedGraphContributorQueryEcho(graphViewID, digest, request)
	echoRaw, _ := json.Marshal(echo)
	var nextToken *string
	if hasMore && len(rows) > 0 {
		token, err := s.cursorProtector.Encode(cursorBinding{
			Route: routeKeyGraphViewContributorsQuery, ActorUserID: actorID, SessionID: sessionID, IncidentID: incidentID.String(),
			Scope: map[string]string{
				"graph_view_id": graphViewID, "projection_result_id": selectedResultID, "graph_query_digest": digest,
			},
			QueryHash: queryHash(echo), QueryEcho: echoRaw, Limit: limit,
		}, "contributor_keyset_v1", newContributorCursorPosition(rows[len(rows)-1], tableRanks))
		if err != nil {
			return nil, httpapi.InternalAPIError(err)
		}
		nextToken = &token
	}
	return map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_contributor_query_result.v2", "graph_view_id": graphViewID,
		"projection_result_id": selectedResultID, "selector": graphSelectorResource(request.Selector), "contributors": contributors,
		"meta": map[string]any{"paging": map[string]any{
			"limit": limit, "returned_count": len(contributors), "next_cursor_token": nextToken,
		}},
	}, nil
}

func savedGraphContributorQueryEcho(graphViewID string, digest string, request graphViewContributorRequest) map[string]any {
	return map[string]any{
		"graph_view_id": graphViewID, "projection_result_id": request.ProjectionResultID,
		"graph_query_digest": digest, "selector": graphSelectorResource(request.Selector),
	}
}

func decodeGraphViewCreateRequest(r *http.Request, limits EffectiveLimits) (graphViewCreateRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "client_txn_id", "display_name", "semantic_query"); apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	if schemaID, err := requiredJSONStringHTTP(raw, "schema_id"); err != nil || schemaID != "cartulary.network_flow.graph_view_create_request.v3" {
		return graphViewCreateRequest{}, invalidNetworkFlowRequestHTTP("schema_id", "invalid_schema_id")
	}
	clientTxnID, apiErr := requiredJSONStringHTTP(raw, "client_txn_id")
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	displayName, apiErr := requiredJSONStringHTTP(raw, "display_name")
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	semantic, apiErr := decodeGraphSemanticRequestHTTP(raw["semantic_query"], limits)
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	if semantic.SchemaID != schemaGraphSemanticQueryV2 {
		return graphViewCreateRequest{}, invalidNetworkFlowRequestHTTP("semantic_query.schema_id", "invalid_schema_id")
	}
	return graphViewCreateRequest{ClientTxnID: clientTxnID, DisplayName: displayName, Semantic: semantic}, nil
}

func decodeGraphViewRenameRequest(r *http.Request) (graphViewRenameRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "client_txn_id", "display_name", "base_graph_view_version"); apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	if schemaID, err := requiredJSONStringHTTP(raw, "schema_id"); err != nil || schemaID != "cartulary.network_flow.graph_view_rename_request.v2" {
		return graphViewRenameRequest{}, invalidNetworkFlowRequestHTTP("schema_id", "invalid_schema_id")
	}
	clientTxnID, apiErr := requiredJSONStringHTTP(raw, "client_txn_id")
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	displayName, apiErr := requiredJSONStringHTTP(raw, "display_name")
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	version, apiErr := decodePositiveIntHTTP(raw["base_graph_view_version"], "base_graph_view_version")
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	return graphViewRenameRequest{graphViewVersionRequest: graphViewVersionRequest{ClientTxnID: clientTxnID, BaseGraphViewVersion: int64(version)}, DisplayName: displayName}, nil
}

func decodeGraphViewVersionRequest(r *http.Request, schemaID string) (graphViewVersionRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "client_txn_id", "base_graph_view_version"); apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	if actual, err := requiredJSONStringHTTP(raw, "schema_id"); err != nil || actual != schemaID {
		return graphViewVersionRequest{}, invalidNetworkFlowRequestHTTP("schema_id", "invalid_schema_id")
	}
	clientTxnID, apiErr := requiredJSONStringHTTP(raw, "client_txn_id")
	if apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	version, apiErr := decodePositiveIntHTTP(raw["base_graph_view_version"], "base_graph_view_version")
	if apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	return graphViewVersionRequest{ClientTxnID: clientTxnID, BaseGraphViewVersion: int64(version)}, nil
}

func decodeSavedGraphContributorRequest(r *http.Request, limits EffectiveLimits) (graphViewContributorRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	schemaID, schemaErr := requiredJSONStringHTTP(raw, "schema_id")
	if schemaErr != nil {
		return graphViewContributorRequest{}, schemaErr
	}
	if schemaID == schemaGraphContributorQueryContinuation {
		if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "cursor_token"); apiErr != nil {
			return graphViewContributorRequest{}, apiErr
		}
		token, tokenErr := requiredJSONStringHTTP(raw, "cursor_token")
		if tokenErr != nil {
			return graphViewContributorRequest{}, tokenErr
		}
		return graphViewContributorRequest{Continuation: true, CursorToken: token}, nil
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "schema_id", "projection_result_id", "selector", "limit"); apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	if schemaID != "cartulary.network_flow.graph_view_contributor_query_request.v2" {
		return graphViewContributorRequest{}, invalidNetworkFlowRequestHTTP("schema_id", "invalid_schema_id")
	}
	resultID, apiErr := requiredJSONStringHTTP(raw, "projection_result_id")
	if apiErr != nil || !graphProjectionResultIDPattern.MatchString(resultID) {
		return graphViewContributorRequest{}, invalidNetworkFlowRequestHTTP("projection_result_id", "invalid_identifier")
	}
	selector, apiErr := decodeGraphSelectorHTTP(raw["selector"])
	if apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	if _, present := raw["limit"]; !present {
		return graphViewContributorRequest{}, invalidNetworkFlowRequestHTTP("limit", "missing_member")
	}
	limit, apiErr := decodePositiveIntHTTP(raw["limit"], "limit")
	if apiErr != nil {
		return graphViewContributorRequest{}, invalidLimitHTTP("limit", "not_integer")
	}
	if limit < 1 {
		return graphViewContributorRequest{}, invalidLimitHTTP("limit", "below_minimum")
	}
	if limit > 1000 || int64(limit) > limits.MaxQueryLimit {
		return graphViewContributorRequest{}, invalidLimitHTTP("limit", "above_maximum")
	}
	return graphViewContributorRequest{ProjectionResultID: resultID, Selector: selector, Limit: limit}, nil
}

func (s *routeService) activeGraphView(ctx context.Context, incidentID uuid.UUID, graphViewID string) (graphViewDeclaration, *httpapi.APIError) {
	declaration, err := s.store.GetGraphViewDeclaration(ctx, incidentID, graphViewID)
	if errors.Is(err, errGraphViewDeclarationNotFound) {
		return graphViewDeclaration{}, networkFlowAPIError(http.StatusNotFound, "network_flow_graph_view_not_found", "graph_view_id", "not_found")
	}
	if err != nil {
		return graphViewDeclaration{}, httpapi.InternalAPIError(err)
	}
	if declaration.DeclarationState != graphViewDeclarationStateActive {
		return graphViewDeclaration{}, networkFlowAPIError(http.StatusNotFound, "network_flow_graph_view_not_found", "graph_view_id", "not_found")
	}
	return declaration, nil
}

func (s *routeService) graphViewMaterializationStatus(ctx context.Context, declaration graphViewDeclaration) (string, error) {
	if declaration.LatestJobID == nil {
		if declaration.SelectedResult != nil {
			return "succeeded", nil
		}
		if declaration.LastFailureCode != nil {
			return "failed", nil
		}
		return "not_started", nil
	}
	if s.jobManager == nil {
		return "", errors.New("graph view job manager unavailable")
	}
	job, err := s.jobManager.Get(ctx, *declaration.LatestJobID)
	if errors.Is(err, jobs.ErrNotFound) {
		if declaration.SelectedResult != nil {
			return "succeeded", nil
		}
		if declaration.LastFailureCode != nil {
			return "failed", nil
		}
		return "not_started", nil
	}
	if err != nil {
		return "", err
	}
	return graphViewStatusFromJobStatus(job.Status), nil
}

func graphViewStatusFromJobStatus(status string) string {
	switch status {
	case jobs.StatusQueued:
		return "queued"
	case jobs.StatusRunning, jobs.StatusCancelRequested:
		return "running"
	case jobs.StatusSucceeded:
		return "succeeded"
	case jobs.StatusFailed:
		return "failed"
	case jobs.StatusCanceled:
		return "cancelled"
	default:
		return "not_started"
	}
}

func graphViewResource(declaration graphViewDeclaration) map[string]any {
	var selected any
	if declaration.SelectedResult != nil {
		selected = map[string]any{
			"projection_result_id":            declaration.SelectedResult.ProjectionResultID,
			"source_snapshot_id":              declaration.SelectedResult.SourceSnapshotID,
			"projection_schema_id":            declaration.SelectedResult.ProjectionSchemaID,
			"projection_version":              declaration.SelectedResult.ProjectionVersion,
			"normalized_configuration_sha256": declaration.SelectedResult.NormalizedConfigurationSHA256,
			"normalized_source_sha256":        declaration.SelectedResult.NormalizedSourceSHA256,
			"canonical_output_sha256":         declaration.SelectedResult.CanonicalOutputSHA256,
		}
	}
	var latestJobID any
	if declaration.LatestJobID != nil {
		latestJobID = declaration.LatestJobID.String()
	}
	var failure any
	if declaration.LastFailureCode != nil {
		failure = *declaration.LastFailureCode
	}
	return map[string]any{
		"schema_id": "cartulary.network_flow.graph_view.v4", "graph_view_id": declaration.GraphViewID,
		"incident_id": declaration.IncidentID.String(), "display_name": declaration.DisplayName,
		"semantic_query_sha256":      declaration.SemanticQuerySHA256,
		"desired_source_snapshot_id": declaration.DesiredSourceSnapshotID,
		"created_by":                 declaration.CreatedByUserID.String(),
		"graph_view_version":         declaration.GraphViewVersion, "materialization_generation": declaration.MaterializationGeneration,
		"state": declaration.DeclarationState, "semantic_query": rawJSONValue(declaration.SemanticQueryJSON),
		"selected_result_binding": selected, "latest_job_id": latestJobID,
		"last_failure_code": failure, "last_failed_at": nullableGraphViewTimestamp(declaration.LastFailedAt),
		"created_at": timestamp(declaration.CreatedAt), "updated_at": timestamp(declaration.UpdatedAt),
	}
}

func nullableGraphViewTimestamp(value *time.Time) any {
	if value == nil {
		return nil
	}
	return timestamp(*value)
}

func graphViewAcceptedPayload(declaration graphViewDeclaration, jobID uuid.UUID) map[string]any {
	return map[string]any{
		"schema_id":  "cartulary.network_flow.graph_view_accepted.v4",
		"graph_view": graphViewResource(declaration),
		"job":        map[string]any{"job_id": jobID.String(), "status_route": "/api/v1/jobs/" + jobID.String()},
	}
}

func graphViewMutationPayload(declaration graphViewDeclaration) map[string]any {
	return map[string]any{"schema_id": "cartulary.network_flow.graph_view_mutation_result.v4", "graph_view": graphViewResource(declaration)}
}

func graphViewResultBinding(declaration graphViewDeclaration) graphprojection.ResultBindingV2 {
	selected := declaration.SelectedResult
	return graphprojection.ResultBindingV2{
		ProjectionResultID: selected.ProjectionResultID, GraphViewID: declaration.GraphViewID,
		SourceOwnerID: graphSourceOwnerID, SourceSnapshotID: selected.SourceSnapshotID,
		ProjectionSchemaID: selected.ProjectionSchemaID, ProjectionVersion: selected.ProjectionVersion,
		NormalizedConfigurationSHA256: selected.NormalizedConfigurationSHA256,
		NormalizedSourceSHA256:        selected.NormalizedSourceSHA256, CanonicalOutputSHA256: selected.CanonicalOutputSHA256,
	}
}

func graphViewIdempotencyKey(routeKey string, actorUserID, incidentID uuid.UUID, resourceID, clientTxnID string) authn.RouteIdempotencyKey {
	if resourceID != "graph-views" {
		resourceID = "graph_view_id:" + resourceID
	}
	return authn.RouteIdempotencyKey{RouteKey: routeKey, ActorUserID: actorUserID, ScopeKey: incidentID.String() + ":" + resourceID, ClientTxnID: clientTxnID}
}

func sha256Bytes(value []byte) []byte {
	digest := sha256.Sum256(value)
	return digest[:]
}

func graphViewMutationError(err error) *httpapi.APIError {
	var displayName *invalidDisplayNameError
	var version *graphViewVersionConflictError
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible), admission.IsDenied(err, admission.DenialInsufficientRole), admission.IsDenied(err, admission.DenialIncidentClosed):
		return savedGraphAdmissionError(err, "")
	case errors.As(err, &displayName):
		return networkFlowAPIError(http.StatusBadRequest, "network_flow_invalid_display_name", "display_name", displayName.ReasonCode)
	case errors.As(err, &version):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_version_conflict", "base_graph_view_version", "stale_version")
	case errors.Is(err, errGraphViewDeclarationNotFound):
		return networkFlowAPIError(http.StatusNotFound, "network_flow_graph_view_not_found", "graph_view_id", "not_found")
	case errors.Is(err, errGraphViewDeclarationNotActive):
		return networkFlowAPIError(http.StatusNotFound, "network_flow_graph_view_not_found", "graph_view_id", "not_found")
	case errors.Is(err, errGraphViewActiveLimit):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_limit_exceeded", "graph_view_id", "active_graph_view_limit_exceeded")
	case errors.Is(err, errGraphViewRetainedLimit):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_limit_exceeded", "graph_view_id", "retained_graph_view_limit_exceeded")
	case errors.Is(err, errGraphViewJobLimit):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_materialization_limit_exceeded", "graph_view_id", "nonterminal_job_limit_exceeded")
	case errors.Is(err, errGraphViewPublicationStale):
		return graphQueryStaleHTTP("scope_stale", "")
	default:
		return httpapi.InternalAPIError(err)
	}
}

func graphViewMutationAPIError(err error, clientTxnID string) *httpapi.APIError {
	var receiptFailure *savedGraphReceiptFailure
	if errors.As(err, &receiptFailure) {
		return httpapi.InternalAPIError(err)
	}
	if errors.Is(err, authn.ErrClientTxnConflict) || authn.IsUniqueViolation(err) {
		return httpapi.ClientTxnConflictError(clientTxnID)
	}
	return graphViewMutationError(err)
}

func graphViewNotMaterialized(status string) *httpapi.APIError {
	reason := "initial_materialization_pending"
	if status == "failed" || status == "cancelled" {
		reason = "initial_materialization_failed"
	}
	return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_not_materialized", "graph_view_id", reason)
}

func parseIncidentGraphViewPathValues(w http.ResponseWriter, r *http.Request) (uuid.UUID, string, bool) {
	incidentID, ok := parseIncidentPathValue(w, r)
	if !ok {
		return uuid.Nil, "", false
	}
	graphViewID := r.PathValue("graph_view_id")
	if !graphViewIDPattern.MatchString(graphViewID) {
		http.NotFound(w, r)
		return uuid.Nil, "", false
	}
	return incidentID, graphViewID, true
}
