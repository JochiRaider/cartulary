package networkflow

import (
	"crypto/sha256"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *routeService) handleGraphViewsCollection(w http.ResponseWriter, r *http.Request) {
	routeID := map[string]string{"GET": "nf.graph_views.list", "POST": "nf.graph_views.create"}[r.Method]
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, incidentID := admitted.principal, admitted.incidentID
	switch r.Method {
	case http.MethodGet:
		declarations, failure := s.savedGraphReads.list(r.Context(), admitted.identity())
		if failure != nil {
			writeAPIError(w, r, semanticHTTPError(failure))
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
		request, apiErr := decodeGraphViewCreateRequest(r, s.limits)
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
	routeID := map[string]string{"GET": "nf.graph_views.get", "PATCH": "nf.graph_views.patch", "DELETE": "nf.graph_views.delete"}[r.Method]
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, incidentID := admitted.principal, admitted.incidentID
	graphViewID := admitted.graphViewID
	switch r.Method {
	case http.MethodGet:
		declaration, failure := s.savedGraphReads.get(r.Context(), admitted.identity(), graphViewID)
		apiErr = semanticHTTPError(failure)
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
	routeID := "nf.graph_views.refresh"
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, incidentID := admitted.principal, admitted.incidentID
	graphViewID := admitted.graphViewID
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
	routeID := "nf.graph_views.result"
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	graphViewID := admitted.graphViewID
	outcome, failure := s.savedGraphReads.result(r.Context(), admitted.identity(), graphViewID)
	if failure != nil {
		writeAPIError(w, r, semanticHTTPError(failure))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":  "cartulary.network_flow.graph_view_result.v4",
		"graph_view": graphViewResource(outcome.Declaration), "result": graphQueryResultResource(outcome.Composition),
	})
}

func (s *routeService) handleGraphViewContributorsQuery(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.graph_views.contributors.query"
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	graphViewID := admitted.graphViewID
	request, apiErr := decodeSavedGraphContributorRequest(r, s.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, failure := s.savedGraphReads.contributors(r.Context(), admitted.identity(), graphViewID, request)
	if failure != nil {
		writeAPIError(w, r, semanticHTTPError(failure))
		return
	}
	result := savedGraphContributorsResource(outcome)
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func decodeGraphViewCreateRequest(r *http.Request, limits EffectiveLimits) (graphViewCreateRequest, *httpapi.APIError) {
	value, failure := decodeGraphViewCreateRequestValue(r.Body, limits)
	return value, semanticHTTPError(failure)
}

func decodeGraphViewRenameRequest(r *http.Request) (graphViewRenameRequest, *httpapi.APIError) {
	value, failure := decodeGraphViewRenameRequestValue(r.Body)
	return value, semanticHTTPError(failure)
}

func decodeGraphViewVersionRequest(r *http.Request, schemaID string) (graphViewVersionRequest, *httpapi.APIError) {
	value, failure := decodeGraphViewVersionRequestValue(r.Body, schemaID)
	return value, semanticHTTPError(failure)
}

func decodeSavedGraphContributorRequest(r *http.Request, limits EffectiveLimits) (graphViewContributorRequest, *httpapi.APIError) {
	value, failure := decodeSavedGraphContributorRequestValue(r.Body, limits)
	return value, semanticHTTPError(failure)
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
