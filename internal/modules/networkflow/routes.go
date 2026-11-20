package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

type routeService struct {
	savedGraphReads *savedGraphReadApplication
	graphQueries    *graphQueryApplication
	tableQueries    *tableQueryApplication
	limits          EffectiveLimits
	incidentAccess  incidentAdmissionChecker
	authStore       interface {
		httpauth.SessionStore
		httpauth.SessionSlider
	}
	now         func() time.Time
	tables      *tableApplication
	links       *indicatorLinkApplication
	savedGraphs *savedGraphApplication
}

func newRouteService(module *Module) (*routeService, error) {
	if err := module.prepareActiveApplications(); err != nil {
		return nil, err
	}
	return &routeService{
		savedGraphReads: module.active.savedGraphReads, graphQueries: module.active.graphQueries, tableQueries: module.active.tableQueries, tables: module.active.tables, links: module.active.links, savedGraphs: module.active.savedGraphs,
		limits: module.limits, incidentAccess: module.incidentAccess, authStore: module.authStore,
		now: module.now,
	}, nil
}

// operationHandlers is one handler per Network Flow route (NF NLSpec Table
// 17-A and §19.1), bound in the public route registry.
func (s *routeService) operationHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"nf.graphs.contributors.query":      s.handleGraphContributorsQuery,
		"nf.graphs.query":                   s.handleGraphQuery,
		"nf.graph_views.list":               s.handleGraphViewList,
		"nf.graph_views.create":             s.handleGraphViewCreate,
		"nf.graph_views.get":                s.handleGraphViewGet,
		"nf.graph_views.patch":              s.handleGraphViewPatch,
		"nf.graph_views.delete":             s.handleGraphViewDelete,
		"nf.graph_views.refresh":            s.handleGraphViewRefresh,
		"nf.graph_views.result":             s.handleGraphViewResult,
		"nf.graph_views.contributors.query": s.handleGraphViewContributorsQuery,
		"nf.indicator_links.create":         s.handleIndicatorLinks,
		"nf.rejected_rows.query":            s.handleRejectedRowsQuery,
		"nf.rows.query":                     s.handleRowsQuery,
		"nf.source_profiles.list":           s.handleSourceProfiles,
		"nf.tables.delete":                  s.handleTableDelete,
		"nf.tables.get":                     s.handleTableGet,
		"nf.tables.list":                    s.handleTablesCollection,
		"nf.tables.patch":                   s.handleTablePatch,
		"nf.tables.query":                   s.handleTableRowsQuery,
	}
}

func (s *routeService) handleSourceProfiles(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	limits, failure := s.tableQueries.profiles(r.Context(), admitted.identity())
	if failure != nil {
		writeAPIError(w, r, semanticHTTPError(failure))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	profiles := sourceProfileResources()
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":        "cartulary.network_flow.source_profile_list.v3",
		"source_profiles":  profiles,
		"effective_limits": effectiveLimitsResource(limits),
		"meta":             map[string]any{"count": len(profiles)},
	})
}

func (s *routeService) handleTablesCollection(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	tables, failure := s.tableQueries.list(r.Context(), admitted.identity())
	if failure != nil {
		writeAPIError(w, r, semanticHTTPError(failure))
		return
	}
	resources := make([]any, 0, len(tables))
	for _, table := range tables {
		resources = append(resources, tableResource(table))
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id": "cartulary.network_flow_table_list.v3",
		"tables":    resources,
		"meta":      map[string]any{"count": len(resources)},
	})
}

func (s *routeService) handleTableGet(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	table, failure := s.tableQueries.get(r.Context(), admitted.identity(), admitted.tableID)
	if failure != nil {
		writeAPIError(w, r, pathTableNotFound(semanticHTTPError(failure)))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id": "cartulary.network_flow_table_get.v3",
		"table":     tableResource(table),
	})
}

func (s *routeService) handleTablePatch(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, incidentID, tableID := admitted.principal, admitted.incidentID, admitted.tableID
	request, apiErr := decodeRenameRequest(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	hash, err := tableRenameRequestHash(tableID, request)
	if err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	payload, status, apiErr := s.commitTableRenameRoute(r.Context(), incidentID, tableID, principal.User.ID, request, hash, httpapi.RequestIDFromContext(r.Context()))
	if apiErr != nil {
		writeAPIError(w, r, pathTableNotFound(apiErr))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, status, payload)
}

func (s *routeService) handleTableDelete(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, incidentID, tableID := admitted.principal, admitted.incidentID, admitted.tableID
	request, apiErr := decodeSoftDeleteRequest(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	hash, err := tableSoftDeleteRequestHash(tableID, request)
	if err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	payload, status, apiErr := s.commitTableSoftDeleteRoute(r.Context(), incidentID, tableID, principal.User.ID, request, hash, httpapi.RequestIDFromContext(r.Context()))
	if apiErr != nil {
		writeAPIError(w, r, pathTableNotFound(apiErr))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, status, payload)
}

func (s *routeService) handleTableRowsQuery(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	tableID := admitted.tableID
	request, apiErr := decodeAcceptedRowQueryRequestHTTP(r.Body, schemaTableQueryRequest, schemaTableQueryContinuation, s.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, failure := s.tableQueries.rows(r.Context(), admitted.identity(), tableID, request)
	apiErr = pathTableNotFound(semanticHTTPError(failure))
	result := acceptedRowsResource(outcome)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	result["schema_id"] = "cartulary.network_flow.table_query_result.v3"
	result["network_flow_table_id"] = tableID
	delete(result, "table_scope")
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func (s *routeService) handleRowsQuery(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	request, apiErr := decodeAcceptedRowQueryRequestHTTP(r.Body, schemaRowsQueryRequest, schemaRowsQueryContinuation, s.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, failure := s.tableQueries.rows(r.Context(), admitted.identity(), "", request)
	apiErr = semanticHTTPError(failure)
	result := acceptedRowsResource(outcome)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	result["schema_id"] = "cartulary.network_flow.rows_query_result.v3"
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func (s *routeService) handleRejectedRowsQuery(w http.ResponseWriter, r *http.Request) {
	admitted, apiErr := admittedOperation(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	tableID := admitted.tableID
	request, apiErr := decodeRejectedRowsQueryRequestHTTP(r.Body, s.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, failure := s.tableQueries.diagnostics(r.Context(), admitted.identity(), tableID, request)
	apiErr = pathTableNotFound(semanticHTTPError(failure))
	result := diagnosticsResource(tableID, outcome)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func decodeRenameRequest(r *http.Request) (tableRenameRequest, *httpapi.APIError) {
	value, failure := decodeRenameRequestValue(r.Body)
	return value, semanticHTTPError(failure)
}

func decodeSoftDeleteRequest(r *http.Request) (tableSoftDeleteRequest, *httpapi.APIError) {
	value, failure := decodeSoftDeleteRequestValue(r.Body)
	return value, semanticHTTPError(failure)
}

func tableMutationPayload(table tableRecord) map[string]any {
	return map[string]any{
		"schema_id": "cartulary.network_flow_table_mutation_result.v3",
		"table":     tableResource(table),
	}
}

func tableMutationIdempotencyKey(routeKey string, actorUserID uuid.UUID, incidentID uuid.UUID, tableID string, clientTxnID string) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actorUserID,
		ScopeKey:    incidentID.String() + ":" + tableID,
		ClientTxnID: clientTxnID,
	}
}

func tableRenameRequestHash(tableID string, request tableRenameRequest) ([]byte, error) {
	name, err := normalizeTableDisplayNameInput(request.DisplayName)
	if err != nil {
		return nil, err
	}
	return mutationRequestHash(routeKeyTablesPatch, "network_flow_table_id:"+tableID, map[string]any{
		"base_table_version": request.BaseTableVersion, "display_name": name,
	})
}

func tableSoftDeleteRequestHash(tableID string, request tableSoftDeleteRequest) ([]byte, error) {
	return mutationRequestHash(routeKeyTablesDelete, "network_flow_table_id:"+tableID, map[string]any{"base_table_version": request.BaseTableVersion})
}

func decodeStoredNetworkFlowResponse(data []byte) (map[string]any, error) {
	if _, err := strictjson.DecodeUnicodeObject(bytes.NewReader(data)); err != nil {
		return nil, fmt.Errorf("decode stored network flow response: %w", err)
	}
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode stored network flow response: %w", err)
	}
	if payload == nil {
		return nil, errors.New("decode stored network flow response: empty payload")
	}
	return payload, nil
}

func (s *routeService) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleOpen})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, httpapi.NotFound("incident_not_found")
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Message: "authorization denied", Details: map[string]any{"required_role": requiredRole}}
	case err != nil:
		return admission.Grant{}, httpapi.InternalAPIError(err)
	default:
		return grant, nil
	}
}

func (s *routeService) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, s.now)
}

func tableMutationError(err error, tableID string) *httpapi.APIError {
	var versionConflict *tableVersionConflictError
	var displayName *invalidDisplayNameError
	if errors.As(err, &versionConflict) {
		apiErr := networkFlowAPIError(http.StatusConflict, "network_flow_table_version_conflict", "base_table_version", "stale_version")
		apiErr.Details["network_flow_table_id"] = versionConflict.TableID
		apiErr.Details["base_table_version"] = versionConflict.BaseTableVersion
		apiErr.Details["current_table_version"] = versionConflict.CurrentTableVersion
		apiErr.Details["retry_action"] = "refresh_resource"
		return apiErr
	}
	if errors.As(err, &displayName) {
		apiErr := networkFlowAPIError(http.StatusBadRequest, "network_flow_invalid_display_name", "display_name", displayName.ReasonCode)
		apiErr.Details["max_length"] = 64
		apiErr.Details["normalized_length"] = displayName.NormalizedLength
		apiErr.Details["retry_action"] = "correct_request"
		return apiErr
	}
	if errors.Is(err, errTableNameExhausted) {
		return networkFlowAPIError(http.StatusConflict, "network_flow_table_name_exhausted", "display_name", "suffix_space_exhausted")
	}
	return semanticHTTPError(tableReadFailure(err, tableID))
}

func tableMutationAPIError(err error, clientTxnID, tableID string) *httpapi.APIError {
	if errors.Is(err, authn.ErrClientTxnConflict) {
		return httpapi.ClientTxnConflictError(clientTxnID)
	}
	return tableMutationError(err, tableID)
}
