package networkflow

import (
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
	keys        authn.MasterKeys
	now         func() time.Time
	tables      *tableApplication
	links       *indicatorLinkApplication
	savedGraphs *savedGraphApplication
}

func newRouteService(deps httpapi.DependencySet, module *Module) (*routeService, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	if err := module.prepareActiveApplications(); err != nil {
		return nil, err
	}
	return &routeService{
		savedGraphReads: module.active.savedGraphReads, graphQueries: module.active.graphQueries, tableQueries: module.active.tableQueries, tables: module.active.tables, links: module.active.links, savedGraphs: module.active.savedGraphs,
		limits: module.limits, incidentAccess: module.incidentAccess, authStore: module.authStore,
		keys: keys,
		now:  module.now,
	}, nil
}

func (s *routeService) handleSourceProfiles(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.source_profiles.list"
	admitted, apiErr := s.admitOperation(r, routeID)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":        "cartulary.network_flow.source_profile_list.v2",
		"source_profiles":  []any{sourceProfileResource()},
		"effective_limits": effectiveLimitsResource(limits),
		"meta":             map[string]any{"count": 1},
	})
}

func (s *routeService) handleTablesCollection(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.tables.list"
	admitted, apiErr := s.admitOperation(r, routeID)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id": "cartulary.network_flow_table_list.v1",
		"tables":    resources,
		"meta":      map[string]any{"count": len(resources)},
	})
}

func (s *routeService) handleTableResource(w http.ResponseWriter, r *http.Request) {
	routeID := map[string]string{"GET": "nf.tables.get", "PATCH": "nf.tables.patch", "DELETE": "nf.tables.delete"}[r.Method]
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, incidentID := admitted.principal, admitted.incidentID
	tableID := admitted.tableID
	switch r.Method {
	case http.MethodGet:
		table, failure := s.tableQueries.get(r.Context(), admitted.identity(), tableID)
		if failure != nil {
			writeAPIError(w, r, semanticHTTPError(failure))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
			"schema_id": "cartulary.network_flow_table_get.v1",
			"table":     tableResource(table),
		})
	case http.MethodPatch:
		request, apiErr := decodeRenameRequest(r)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		payload, status, apiErr := s.commitTableRenameRoute(r.Context(), incidentID, tableID, principal.User.ID, request, tableRenameRequestHash(tableID, request), httpapi.RequestIDFromContext(r.Context()))
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
		request, apiErr := decodeSoftDeleteRequest(r)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		payload, status, apiErr := s.commitTableSoftDeleteRoute(r.Context(), incidentID, tableID, principal.User.ID, request, tableSoftDeleteRequestHash(tableID, request), httpapi.RequestIDFromContext(r.Context()))
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

func (s *routeService) handleTableRowsQuery(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.tables.query"
	admitted, apiErr := s.admitOperation(r, routeID)
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
	apiErr = semanticHTTPError(failure)
	result := acceptedRowsResource(outcome)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	result["schema_id"] = "cartulary.network_flow.table_query_result.v1"
	result["network_flow_table_id"] = tableID
	delete(result, "table_scope")
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func (s *routeService) handleRowsQuery(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.rows.query"
	admitted, apiErr := s.admitOperation(r, routeID)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	result["schema_id"] = "cartulary.network_flow.rows_query_result.v1"
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func (s *routeService) handleRejectedRowsQuery(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.rejected_rows.query"
	admitted, apiErr := s.admitOperation(r, routeID)
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
	apiErr = semanticHTTPError(failure)
	result := diagnosticsResource(tableID, outcome)
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
		"schema_id": "cartulary.network_flow_table_mutation_result.v1",
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

func tableRenameRequestHash(tableID string, request tableRenameRequest) []byte {
	name, err := normalizeTableDisplayNameInput(request.DisplayName)
	if err != nil {
		return nil
	}
	return sha256Bytes(graphViewMutationBytes(routeKeyTablesPatch, "network_flow_table_id:"+tableID, map[string]any{
		"base_table_version": request.BaseTableVersion, "display_name": name,
	}))
}

func tableSoftDeleteRequestHash(tableID string, request tableSoftDeleteRequest) []byte {
	return sha256Bytes(graphViewMutationBytes(routeKeyTablesDelete, "network_flow_table_id:"+tableID, map[string]any{
		"base_table_version": request.BaseTableVersion,
	}))
}

func decodeStoredNetworkFlowResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode stored network flow response: %w", err)
	}
	if payload == nil {
		return nil, errors.New("decode stored network flow response: empty payload")
	}
	return payload, nil
}

func (s *routeService) authenticate(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	return httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
}

func (s *routeService) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleOpen})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
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

func (s *routeService) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

func tableReadError(err error) *httpapi.APIError {
	if errors.Is(err, errTableNotFound) {
		return networkFlowAPIError(http.StatusNotFound, "network_flow_table_not_found", "network_flow_table_id", "not_found")
	}
	if errors.Is(err, errTableNotActive) {
		return networkFlowAPIError(http.StatusConflict, "network_flow_table_not_active", "network_flow_table_id", "soft_deleted")
	}
	return httpapi.InternalAPIError(err)
}

func tableMutationError(err error) *httpapi.APIError {
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
	return tableReadError(err)
}

func tableMutationAPIError(err error, clientTxnID string) *httpapi.APIError {
	if errors.Is(err, authn.ErrClientTxnConflict) {
		return httpapi.ClientTxnConflictError(clientTxnID)
	}
	return tableMutationError(err)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	httpapi.WriteAPIError(w, r, apiErr)
}
