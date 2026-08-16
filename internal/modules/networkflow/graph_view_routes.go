package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const (
	routeKeyGraphViewsCreate  = "nf.graph_views.create"
	routeKeyGraphViewsPatch   = "nf.graph_views.patch"
	routeKeyGraphViewsDelete  = "nf.graph_views.delete"
	routeKeyGraphViewsRefresh = "nf.graph_views.refresh"
)

type graphViewCreateRequest struct {
	ClientTxnID string
	DisplayName string
	Semantic    graphSemanticRequest
}

type graphViewVersionRequest struct {
	ClientTxnID          string
	BaseGraphViewVersion int64
}

type graphViewRenameRequest struct {
	graphViewVersionRequest
	DisplayName string
}

type graphViewContributorRequest struct {
	ProjectionResultID string
	Selector           graphSelector
	Limit              int
}

func (s *Service) handleGraphViewsCollection(w http.ResponseWriter, r *http.Request) {
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
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		declarations, err := s.store.ListActiveGraphViewDeclarations(r.Context(), incidentID)
		if err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		resources := make([]any, 0, len(declarations))
		for _, declaration := range declarations {
			status, statusErr := s.graphViewMaterializationStatus(r.Context(), declaration)
			if statusErr != nil {
				writeAPIError(w, r, httpapi.InternalAPIError(statusErr))
				return
			}
			resources = append(resources, graphViewResource(declaration, status))
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
			"schema_id": "cartulary.network_flow.graph_view_list.v1", "graph_views": resources,
		})
	case http.MethodPost:
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := decodeGraphViewCreateRequest(r, s.store.limits)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		payload, status, jobID, apiErr := s.commitGraphViewCreate(
			r.Context(), incidentID, principal.User.ID, request,
			httpapi.RequestIDFromContext(r.Context()),
		)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if jobID != uuid.Nil && s.jobRunner != nil {
			s.jobRunner.Notify(jobID)
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, status, payload)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleGraphViewResource(w http.ResponseWriter, r *http.Request) {
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
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		declaration, apiErr := s.activeGraphView(r.Context(), incidentID, graphViewID)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		status, err := s.graphViewMaterializationStatus(r.Context(), declaration)
		if err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
			"schema_id":  "cartulary.network_flow.graph_view_get.v1",
			"graph_view": graphViewResource(declaration, status),
		})
	case http.MethodPatch:
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := decodeGraphViewRenameRequest(r)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		payload, status, apiErr := s.commitGraphViewRename(r.Context(), incidentID, graphViewID, principal.User.ID, request, httpapi.RequestIDFromContext(r.Context()))
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, status, payload)
	case http.MethodDelete:
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "reviewer", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := decodeGraphViewVersionRequest(r, "cartulary.network_flow.graph_view_retire_request.v1")
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		payload, status, apiErr := s.commitGraphViewRetire(r.Context(), incidentID, graphViewID, principal.User.ID, request, httpapi.RequestIDFromContext(r.Context()))
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, status, payload)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleGraphViewRefresh(w http.ResponseWriter, r *http.Request) {
	incidentID, graphViewID, ok := parseIncidentGraphViewPathValues(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeGraphViewVersionRequest(r, "cartulary.network_flow.graph_view_refresh_request.v1")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	payload, status, jobID, apiErr := s.commitGraphViewRefresh(r.Context(), incidentID, graphViewID, principal.User.ID, request, httpapi.RequestIDFromContext(r.Context()))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if jobID != uuid.Nil && s.jobRunner != nil {
		s.jobRunner.Notify(jobID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, status, payload)
}

func (s *Service) handleGraphViewResult(w http.ResponseWriter, r *http.Request) {
	incidentID, graphViewID, ok := parseIncidentGraphViewPathValues(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	status, err := s.graphViewMaterializationStatus(r.Context(), declaration)
	if err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":  "cartulary.network_flow.graph_view_result.v1",
		"graph_view": graphViewResource(declaration, status), "projection_result": result,
	})
}

func (s *Service) handleGraphViewContributorsQuery(w http.ResponseWriter, r *http.Request) {
	incidentID, graphViewID, ok := parseIncidentGraphViewPathValues(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	if request.ProjectionResultID != declaration.SelectedResult.ProjectionResultID {
		writeAPIError(w, r, graphQueryStale("projection_result_mismatch", request.ProjectionResultID))
		return
	}
	semantic, apiErr := decodeGraphSemanticRequest(declaration.SemanticQueryJSON, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(errors.New("stored graph semantic query invalid")))
		return
	}
	composition, apiErr := s.composeGraphSourceFromSemantic(r.Context(), incidentID, semantic)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	rows, apiErr := graphContributorRows(composition, request.Selector)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if len(rows) > request.Limit {
		rows = rows[:request.Limit]
	}
	contributors := make([]any, 0, len(rows))
	for _, row := range rows {
		contributors = append(contributors, map[string]any{"row_ref": rowRefResource(row), "row": rowResource(row)})
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":     "cartulary.network_flow.graph_view_contributor_query_result.v1",
		"graph_view_id": graphViewID, "projection_result_id": request.ProjectionResultID,
		"contributors": contributors,
	})
}

func decodeGraphViewCreateRequest(r *http.Request, limits Limits) (graphViewCreateRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "client_txn_id", "display_name", "semantic_query"); apiErr != nil {
		return graphViewCreateRequest{}, apiErr
	}
	if schemaID, err := requiredJSONString(raw, "schema_id"); err != nil || schemaID != "cartulary.network_flow.graph_view_create_request.v1" {
		return graphViewCreateRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
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
	return graphViewCreateRequest{ClientTxnID: clientTxnID, DisplayName: displayName, Semantic: semantic}, nil
}

func decodeGraphViewRenameRequest(r *http.Request) (graphViewRenameRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "client_txn_id", "display_name", "base_graph_view_version"); apiErr != nil {
		return graphViewRenameRequest{}, apiErr
	}
	if schemaID, err := requiredJSONString(raw, "schema_id"); err != nil || schemaID != "cartulary.network_flow.graph_view_rename_request.v1" {
		return graphViewRenameRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
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

func decodeGraphViewVersionRequest(r *http.Request, schemaID string) (graphViewVersionRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "client_txn_id", "base_graph_view_version"); apiErr != nil {
		return graphViewVersionRequest{}, apiErr
	}
	if actual, err := requiredJSONString(raw, "schema_id"); err != nil || actual != schemaID {
		return graphViewVersionRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
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

func decodeSavedGraphContributorRequest(r *http.Request, limits Limits) (graphViewContributorRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembers(raw, "schema_id", "projection_result_id", "selector", "limit"); apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	if schemaID, err := requiredJSONString(raw, "schema_id"); err != nil || schemaID != "cartulary.network_flow.graph_view_contributor_query_request.v1" {
		return graphViewContributorRequest{}, invalidNetworkFlowRequest("schema_id", "invalid_schema_id")
	}
	resultID, apiErr := requiredJSONString(raw, "projection_result_id")
	if apiErr != nil || !graphProjectionResultIDPattern.MatchString(resultID) {
		return graphViewContributorRequest{}, invalidNetworkFlowRequest("projection_result_id", "invalid_identifier")
	}
	selector, apiErr := decodeGraphSelector(raw["selector"])
	if apiErr != nil {
		return graphViewContributorRequest{}, apiErr
	}
	limit, apiErr := decodePositiveInt(raw["limit"], "limit")
	if apiErr != nil || limit < 1 || limit > 1000 || int64(limit) > limits.MaxQueryLimit {
		return graphViewContributorRequest{}, invalidLimit("limit", "above_maximum")
	}
	return graphViewContributorRequest{ProjectionResultID: resultID, Selector: selector, Limit: limit}, nil
}

func (s *Service) commitGraphViewCreate(ctx context.Context, incidentID, actorUserID uuid.UUID, request graphViewCreateRequest, requestID string) (map[string]any, int, uuid.UUID, *httpapi.APIError) {
	if s.graphViewJobs == nil || s.jobManager == nil {
		return nil, 0, uuid.Nil, httpapi.InternalAPIError(errors.New("graph view jobs unavailable"))
	}
	displayName, err := normalizeGraphViewDisplayName(request.DisplayName)
	if err != nil {
		return nil, 0, uuid.Nil, graphViewMutationError(err)
	}
	normalizedRequest := canonicalJSON(map[string]any{
		"route_key": routeKeyGraphViewsCreate, "client_txn_id": request.ClientTxnID,
		"display_name": displayName, "semantic_query": request.Semantic.Raw,
	})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsCreate, actorUserID, incidentID, "collection", request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
		return payload, status, uuid.Nil, apiErr
	}
	var payload map[string]any
	var jobID uuid.UUID
	err = withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		counts, err := s.store.CountGraphViewDeclarationsTx(ctx, tx, incidentID, s.store.limits.MaxRetainedGraphViewsPerIncident)
		if err != nil {
			return err
		}
		if counts.Active >= s.store.limits.MaxActiveGraphViewsPerIncident || counts.Retained >= s.store.limits.MaxRetainedGraphViewsPerIncident {
			return ErrGraphViewDeclarationLimit
		}
		if err := s.requireGraphViewJobCapacityTx(ctx, tx, incidentID); err != nil {
			return err
		}
		tables, err := graphViewSourceTablesTx(ctx, tx, incidentID, request.Semantic.SelectedTableIDs)
		if err != nil {
			return err
		}
		selectedTableIDs := make([]string, 0, len(tables))
		for _, table := range tables {
			selectedTableIDs = append(selectedTableIDs, table.TableID)
		}
		semantic := request.Semantic
		semantic.SelectedTableIDs = selectedTableIDs
		semantic.Raw = graphSemanticQueryResource(selectedTableIDs, semantic.Filters, semantic.TimeRange, semantic.Aggregation, semantic.ResultLimits)
		semanticJSON := canonicalJSON(semantic.Raw)
		snapshotID, err := graphViewSourceSnapshotTx(ctx, tx, incidentID, semantic)
		if err != nil {
			return err
		}
		graphViewID, err := NewGraphViewID()
		if err != nil {
			return err
		}
		now := s.now().UTC()
		declaration := GraphViewDeclaration{
			GraphViewID: graphViewID, IncidentID: incidentID, DisplayName: displayName,
			NormalizedDisplayName: strings.ToLower(displayName), DeclarationState: GraphViewDeclarationStateActive,
			SemanticQueryJSON: semanticJSON, SemanticQuerySHA256: GraphViewSemanticQuerySHA256(semanticJSON),
			DesiredSourceSnapshotID: snapshotID, GraphViewVersion: 1, MaterializationGeneration: 1,
			CreatedByUserID: actorUserID, CreatedAt: now, UpdatedAt: now,
		}
		if err := s.store.InsertGraphViewDeclarationTx(ctx, tx, declaration); err != nil {
			return err
		}
		job, err := s.enqueueGraphViewMaterializationTx(ctx, tx, key, normalizedRequest, declaration, actorUserID, now)
		if err != nil {
			return err
		}
		jobID = uuid.MustParse(job.JobID)
		declaration, err = s.store.SetGraphViewLatestJobTx(ctx, tx, incidentID, graphViewID, jobID)
		if err != nil {
			return err
		}
		payload = graphViewAcceptedPayload(declaration, jobs.StatusQueued, jobID)
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, payload); err != nil {
			return err
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_created", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, 1, 1)
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
			return payload, status, uuid.Nil, apiErr
		}
		return nil, 0, uuid.Nil, graphViewMutationAPIError(err, request.ClientTxnID)
	}
	return payload, http.StatusAccepted, jobID, nil
}

func (s *Service) commitGraphViewRefresh(ctx context.Context, incidentID uuid.UUID, graphViewID string, actorUserID uuid.UUID, request graphViewVersionRequest, requestID string) (map[string]any, int, uuid.UUID, *httpapi.APIError) {
	if s.graphViewJobs == nil || s.jobManager == nil {
		return nil, 0, uuid.Nil, httpapi.InternalAPIError(errors.New("graph view jobs unavailable"))
	}
	normalizedRequest := canonicalJSON(map[string]any{"route_key": routeKeyGraphViewsRefresh, "graph_view_id": graphViewID, "client_txn_id": request.ClientTxnID, "base_graph_view_version": request.BaseGraphViewVersion})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsRefresh, actorUserID, incidentID, graphViewID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
		return payload, status, uuid.Nil, apiErr
	}
	var payload map[string]any
	var jobID uuid.UUID
	err := withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		declaration, err := s.store.GetGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, true)
		if err != nil {
			return err
		}
		if declaration.DeclarationState != GraphViewDeclarationStateActive {
			return ErrGraphViewDeclarationNotActive
		}
		if declaration.GraphViewVersion != request.BaseGraphViewVersion {
			return &GraphViewVersionConflictError{Current: declaration.GraphViewVersion, Base: request.BaseGraphViewVersion}
		}
		if err := s.requireGraphViewJobCapacityTx(ctx, tx, incidentID); err != nil {
			return err
		}
		semantic, apiErr := decodeGraphSemanticRequest(declaration.SemanticQueryJSON, s.store.limits)
		if apiErr != nil {
			return ErrGraphViewDeclarationInvalid
		}
		snapshotID, err := graphViewSourceSnapshotTx(ctx, tx, incidentID, semantic)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		declaration, err = s.store.RefreshGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, request.BaseGraphViewVersion, snapshotID, now)
		if err != nil {
			return err
		}
		job, err := s.enqueueGraphViewMaterializationTx(ctx, tx, key, normalizedRequest, declaration, actorUserID, now)
		if err != nil {
			return err
		}
		jobID = uuid.MustParse(job.JobID)
		declaration, err = s.store.SetGraphViewLatestJobTx(ctx, tx, incidentID, graphViewID, jobID)
		if err != nil {
			return err
		}
		payload = graphViewAcceptedPayload(declaration, jobs.StatusQueued, jobID)
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusAccepted, payload); err != nil {
			return err
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_refresh_requested", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, declaration.GraphViewVersion, declaration.MaterializationGeneration)
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
			return payload, status, uuid.Nil, apiErr
		}
		return nil, 0, uuid.Nil, graphViewMutationAPIError(err, request.ClientTxnID)
	}
	return payload, http.StatusAccepted, jobID, nil
}

func (s *Service) commitGraphViewRename(ctx context.Context, incidentID uuid.UUID, graphViewID string, actorUserID uuid.UUID, request graphViewRenameRequest, requestID string) (map[string]any, int, *httpapi.APIError) {
	displayName, err := normalizeGraphViewDisplayName(request.DisplayName)
	if err != nil {
		return nil, 0, graphViewMutationError(err)
	}
	normalizedRequest := canonicalJSON(map[string]any{"route_key": routeKeyGraphViewsPatch, "graph_view_id": graphViewID, "client_txn_id": request.ClientTxnID, "base_graph_view_version": request.BaseGraphViewVersion, "display_name": displayName})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsPatch, actorUserID, incidentID, graphViewID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	var payload map[string]any
	err = withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		declaration, err := s.store.RenameGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, request.BaseGraphViewVersion, displayName, strings.ToLower(displayName), s.now())
		if err != nil {
			return err
		}
		status, err := s.graphViewMaterializationStatus(ctx, declaration)
		if err != nil {
			return err
		}
		payload = graphViewMutationPayload(declaration, status)
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
			return err
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_renamed", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, declaration.GraphViewVersion, declaration.MaterializationGeneration)
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
			return payload, status, apiErr
		}
		return nil, 0, graphViewMutationAPIError(err, request.ClientTxnID)
	}
	return payload, http.StatusOK, nil
}

func (s *Service) commitGraphViewRetire(ctx context.Context, incidentID uuid.UUID, graphViewID string, actorUserID uuid.UUID, request graphViewVersionRequest, requestID string) (map[string]any, int, *httpapi.APIError) {
	normalizedRequest := canonicalJSON(map[string]any{"route_key": routeKeyGraphViewsDelete, "graph_view_id": graphViewID, "client_txn_id": request.ClientTxnID, "base_graph_view_version": request.BaseGraphViewVersion})
	requestHash := sha256Bytes(normalizedRequest)
	key := graphViewIdempotencyKey(routeKeyGraphViewsDelete, actorUserID, incidentID, graphViewID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	var payload map[string]any
	err := withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		declaration, err := s.store.RetireGraphViewDeclarationTx(ctx, tx, incidentID, graphViewID, request.BaseGraphViewVersion, s.now())
		if err != nil {
			return err
		}
		payload = graphViewMutationPayload(declaration, "not_started")
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, http.StatusOK, payload); err != nil {
			return err
		}
		return s.appendGraphViewAuditTx(ctx, tx, "network_flow_graph_view_retired", actorUserID, incidentID, graphViewID, request.ClientTxnID, requestID, declaration.GraphViewVersion, declaration.MaterializationGeneration)
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayGraphViewMutationIfPresent(ctx, key, requestHash); replayed || apiErr != nil {
			return payload, status, apiErr
		}
		return nil, 0, graphViewMutationAPIError(err, request.ClientTxnID)
	}
	return payload, http.StatusOK, nil
}

func (s *Service) enqueueGraphViewMaterializationTx(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, normalizedRequest []byte, declaration GraphViewDeclaration, actorUserID uuid.UUID, now time.Time) (jobs.Resource, error) {
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &declaration.IncidentID}
	admission, err := jobs.NewExtensionJobAdmission(ProfileID, jobs.NewRouteIdempotencyKey(key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID), scope, normalizedRequest)
	if err != nil {
		return jobs.Resource{}, err
	}
	handlerPayload, err := json.Marshal(graphViewMaterializationPayload{
		SchemaID:   "cartulary.network_flow.graph_view_materialization_payload.v1",
		IncidentID: declaration.IncidentID, GraphViewID: declaration.GraphViewID,
		MaterializationGeneration: declaration.MaterializationGeneration,
		SourceSnapshotID:          declaration.DesiredSourceSnapshotID,
	})
	if err != nil {
		return jobs.Resource{}, err
	}
	total := 1
	return s.graphViewJobs.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{
		JobKind: GraphViewMaterializationJobKind, Scope: scope, SubmittedByUserID: actorUserID,
		AuthPolicy: jobs.AuthPolicyIncidentMembership, Cancelable: true,
		Progress: jobs.Progress{Completed: 0, Total: &total}, HandlerPayload: handlerPayload, Extension: admission,
	}, now.UTC())
}

func (s *Service) requireGraphViewJobCapacityTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	count, err := s.graphViewJobs.CountNonterminalIncidentJobsTx(ctx, tx, incidentID, GraphViewMaterializationJobKind, s.store.limits.MaxNonterminalGraphJobsPerIncident)
	if err != nil {
		return err
	}
	if count >= s.store.limits.MaxNonterminalGraphJobsPerIncident {
		return errGraphViewJobLimit
	}
	return nil
}

var errGraphViewJobLimit = errors.New("network flow graph materialization job limit exceeded")

func (s *Service) replayGraphViewMutationIfPresent(ctx context.Context, key authn.RouteIdempotencyKey, requestHash []byte) (map[string]any, int, bool, *httpapi.APIError) {
	existing, err := s.authStore.GetRouteIdempotency(ctx, key)
	if errors.Is(err, authn.ErrNotFound) {
		return nil, 0, false, nil
	}
	if err != nil {
		return nil, 0, false, httpapi.InternalAPIError(err)
	}
	if !bytes.Equal(existing.RequestHash, requestHash) {
		return nil, 0, true, httpapi.ClientTxnConflictError(key.ClientTxnID)
	}
	payload, err := decodeStoredNetworkFlowResponse(existing.ResponseJSON)
	if err != nil {
		return nil, 0, true, httpapi.InternalAPIError(err)
	}
	if _, present := payload["graph_view"]; present {
		return payload, existing.StatusCode, true, nil
	}
	jobIDText, _ := payload["job_id"].(string)
	jobID, parseErr := uuid.Parse(jobIDText)
	if parseErr != nil || s.jobManager == nil {
		return nil, 0, true, httpapi.InternalAPIError(errors.New("graph view replay payload unavailable"))
	}
	raw, err := s.jobManager.RetainedHandlerPayload(ctx, jobID)
	if err != nil {
		return nil, 0, true, httpapi.InternalAPIError(err)
	}
	var materialization graphViewMaterializationPayload
	if err := json.Unmarshal(raw, &materialization); err != nil || !materialization.valid() {
		return nil, 0, true, httpapi.InternalAPIError(errors.New("graph view replay materialization payload invalid"))
	}
	declaration, err := s.store.GetGraphViewDeclaration(ctx, materialization.IncidentID, materialization.GraphViewID)
	if err != nil {
		return nil, 0, true, httpapi.InternalAPIError(err)
	}
	status := graphViewStatusFromJobStatus(stringValue(payload["status"]))
	return graphViewAcceptedPayload(declaration, status, jobID), http.StatusAccepted, true, nil
}

func (s *Service) activeGraphView(ctx context.Context, incidentID uuid.UUID, graphViewID string) (GraphViewDeclaration, *httpapi.APIError) {
	declaration, err := s.store.GetGraphViewDeclaration(ctx, incidentID, graphViewID)
	if errors.Is(err, ErrGraphViewDeclarationNotFound) {
		return GraphViewDeclaration{}, networkFlowAPIError(http.StatusNotFound, "network_flow_graph_view_not_found", "graph_view_id", "not_found")
	}
	if err != nil {
		return GraphViewDeclaration{}, httpapi.InternalAPIError(err)
	}
	if declaration.DeclarationState != GraphViewDeclarationStateActive {
		return GraphViewDeclaration{}, networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_not_active", "graph_view_id", "retired")
	}
	return declaration, nil
}

func (s *Service) graphViewMaterializationStatus(ctx context.Context, declaration GraphViewDeclaration) (string, error) {
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

func graphViewResource(declaration GraphViewDeclaration, status string) map[string]any {
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
		"schema_id": "cartulary.network_flow.graph_view.v1", "graph_view_id": declaration.GraphViewID,
		"incident_id": declaration.IncidentID.String(), "display_name": declaration.DisplayName,
		"normalized_display_name": declaration.NormalizedDisplayName,
		"graph_view_version":      declaration.GraphViewVersion, "materialization_generation": declaration.MaterializationGeneration,
		"state": declaration.DeclarationState, "semantic_query": rawJSONValue(declaration.SemanticQueryJSON),
		"selected_result": selected, "last_materialization_job_id": latestJobID,
		"last_materialization_status": status, "last_failure_code": failure,
		"created_at": timestamp(declaration.CreatedAt), "updated_at": timestamp(declaration.UpdatedAt),
	}
}

func graphViewAcceptedPayload(declaration GraphViewDeclaration, status string, jobID uuid.UUID) map[string]any {
	materializationStatus := status
	switch status {
	case jobs.StatusQueued, jobs.StatusRunning, jobs.StatusCancelRequested, jobs.StatusSucceeded, jobs.StatusFailed, jobs.StatusCanceled:
		materializationStatus = graphViewStatusFromJobStatus(status)
	}
	return map[string]any{
		"schema_id":  "cartulary.network_flow.graph_view_accepted.v1",
		"graph_view": graphViewResource(declaration, materializationStatus),
		"job_id":     jobID.String(), "job_kind": GraphViewMaterializationJobKind,
	}
}

func graphViewMutationPayload(declaration GraphViewDeclaration, status string) map[string]any {
	return map[string]any{"schema_id": "cartulary.network_flow.graph_view_mutation_result.v1", "graph_view": graphViewResource(declaration, status)}
}

func graphViewResultBinding(declaration GraphViewDeclaration) graphprojection.ResultBindingV2 {
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
	return authn.RouteIdempotencyKey{RouteKey: routeKey, ActorUserID: actorUserID, ScopeKey: incidentID.String() + ":" + resourceID, ClientTxnID: clientTxnID}
}

func sha256Bytes(value []byte) []byte {
	digest := sha256.Sum256(value)
	return digest[:]
}

func normalizeGraphViewDisplayName(value string) (string, error) {
	normalized, err := normalizeExplicitDisplayName(value)
	if err != nil {
		return "", err
	}
	return normalized, nil
}

func graphViewMutationError(err error) *httpapi.APIError {
	var displayName *InvalidDisplayNameError
	var version *GraphViewVersionConflictError
	switch {
	case errors.As(err, &displayName):
		return networkFlowAPIError(http.StatusBadRequest, "network_flow_invalid_display_name", "display_name", displayName.ReasonCode)
	case errors.As(err, &version):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_version_conflict", "base_graph_view_version", "stale_version")
	case errors.Is(err, ErrGraphViewDeclarationNotFound):
		return networkFlowAPIError(http.StatusNotFound, "network_flow_graph_view_not_found", "graph_view_id", "not_found")
	case errors.Is(err, ErrGraphViewDeclarationNotActive):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_not_active", "graph_view_id", "retired")
	case errors.Is(err, ErrGraphViewDeclarationLimit):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_limit_exceeded", "graph_view_id", "declaration_limit_exceeded")
	case errors.Is(err, errGraphViewJobLimit):
		return networkFlowAPIError(http.StatusConflict, "network_flow_graph_materialization_limit_exceeded", "graph_view_id", "nonterminal_job_limit_exceeded")
	case errors.Is(err, ErrGraphViewPublicationStale):
		return graphQueryStale("source_state_changed", "")
	default:
		return httpapi.InternalAPIError(err)
	}
}

func graphViewMutationAPIError(err error, clientTxnID string) *httpapi.APIError {
	if errors.Is(err, authn.ErrClientTxnConflict) || authn.IsUniqueViolation(err) {
		return httpapi.ClientTxnConflictError(clientTxnID)
	}
	return graphViewMutationError(err)
}

func graphViewNotMaterialized(status string) *httpapi.APIError {
	reason := "never_materialized"
	switch status {
	case "queued", "running":
		reason = "materialization_pending"
	case "failed":
		reason = "failed_without_prior_result"
	case "cancelled":
		reason = "cancelled_without_prior_result"
	}
	return networkFlowAPIError(http.StatusConflict, "network_flow_graph_view_not_materialized", "graph_view_id", reason)
}

func (s *Service) appendGraphViewAuditTx(ctx context.Context, tx pgx.Tx, eventKind string, actorUserID, incidentID uuid.UUID, graphViewID, clientTxnID, requestID string, version, generation int64) error {
	return s.store.appendAuditEventTx(ctx, tx, networkFlowAuditEvent{
		ActorUserID: &actorUserID, IncidentID: &incidentID, EventKind: eventKind,
		ClientTxnID: optionalStringPtr(clientTxnID), RequestID: optionalStringPtr(requestID),
		AfterJSON: map[string]any{
			"incident_id": incidentID.String(), "actor_user_id": actorUserID.String(),
			"graph_view_id": graphViewID, "graph_view_version": version,
			"materialization_generation":    generation,
			"network_flow.audit_event_code": eventKind, "network_flow.audit_resource_id": graphViewID,
		},
	})
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

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
