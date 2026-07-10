package networkflow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const routeRoot = "/api/v1/incidents/{incident_id}/network-flow"

const (
	routeKeyTablesPatch  = "nf.tables.patch"
	routeKeyTablesDelete = "nf.tables.delete"
)

type Service struct {
	store           *Store
	incidentAccess  incidents.Access
	authStore       *authn.Store
	keys            authn.MasterKeys
	cursorCodec     *CursorCodec
	hub             *platformws.Hub
	safeDigestKeyID string
	safeDigestKey   []byte
	now             func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.ExtensionProfileClaimedIn(deps.ExtensionProfiles, ProfileID) {
			return nil
		}
		service, err := newRouteService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("GET "+routeRoot+"/source-profiles", service.handleSourceProfiles)
		mux.HandleFunc("GET "+routeRoot+"/tables", service.handleTablesCollection)
		mux.HandleFunc("GET "+routeRoot+"/tables/{network_flow_table_id}", service.handleTableResource)
		mux.HandleFunc("PATCH "+routeRoot+"/tables/{network_flow_table_id}", service.handleTableResource)
		mux.HandleFunc("DELETE "+routeRoot+"/tables/{network_flow_table_id}", service.handleTableResource)
		mux.HandleFunc("POST "+routeRoot+"/tables/{network_flow_table_id}/query", service.handleTableRowsQuery)
		mux.HandleFunc("POST "+routeRoot+"/tables/{network_flow_table_id}/rejected-rows/query", service.handleRejectedRowsQuery)
		mux.HandleFunc("POST "+routeRoot+"/rows/query", service.handleRowsQuery)
		mux.HandleFunc("POST "+routeRoot+"/graphs/query", service.handleGraphQuery)
		mux.HandleFunc("POST "+routeRoot+"/graphs/contributors/query", service.handleGraphContributorsQuery)
		mux.HandleFunc("POST "+routeRoot+"/indicator-links", service.handleIndicatorLinks)
		return nil
	}
}

func newRouteService(deps httpapi.DependencySet) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cursorKey := authn.DerivePurposeKey(keys, "network-flow-cursor-v1")
	cursorCodec, err := NewCursorCodec("master-derived-v1", cursorKey[:], now)
	if err != nil {
		return nil, err
	}
	safeDigestKey := authn.DerivePurposeKey(keys, "network-flow-safe-digest-v1")
	return &Service{
		store:           NewStore(deps.PostgresHandle()),
		incidentAccess:  incidents.NewAccess(deps.PostgresHandle()),
		authStore:       authn.NewStore(deps.PostgresHandle()),
		keys:            keys,
		cursorCodec:     cursorCodec,
		hub:             deps.WSHub,
		safeDigestKeyID: "master-derived-v1",
		safeDigestKey:   safeDigestKey[:],
		now:             now,
	}, nil
}

func (s *Service) handleSourceProfiles(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := parseIncidentPathValue(w, r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"schema_id":        "cartulary.network_flow.source_profile_list.v1",
		"source_profiles":  []any{sourceProfileResource()},
		"effective_limits": effectiveLimitsResource(s.store.limits),
		"meta":             map[string]any{"count": 1},
	})
}

func (s *Service) handleTablesCollection(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := parseIncidentPathValue(w, r)
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
	tables, err := s.store.ListActiveTables(r.Context(), incidentID)
	if err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
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
		"schema_id": "cartulary.network_flow.table_list.v1",
		"tables":    resources,
		"meta":      map[string]any{"count": len(resources)},
	})
}

func (s *Service) handleTableResource(w http.ResponseWriter, r *http.Request) {
	incidentID, tableID, ok := parseIncidentTablePathValues(w, r)
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
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		table, err := s.store.GetActiveTable(r.Context(), incidentID, tableID)
		if err != nil {
			writeAPIError(w, r, tableReadError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, httpapi.InternalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
			"schema_id": "cartulary.network_flow.table_get.v1",
			"table":     tableResource(table),
		})
	case http.MethodPatch:
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
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
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "reviewer", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
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

func (s *Service) handleTableRowsQuery(w http.ResponseWriter, r *http.Request) {
	incidentID, tableID, ok := parseIncidentTablePathValues(w, r)
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
	request, apiErr := decodeAcceptedRowQueryRequest(r.Body, schemaTableQueryRequest, schemaTableQueryContinuation, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, apiErr := s.queryAcceptedRows(r.Context(), principal.User.ID.String(), "nf.tables.query", incidentID, []string{tableID}, "active_table", request)
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

func (s *Service) handleRowsQuery(w http.ResponseWriter, r *http.Request) {
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
	request, apiErr := decodeAcceptedRowQueryRequest(r.Body, schemaRowsQueryRequest, schemaRowsQueryContinuation, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	var tableIDs []string
	mode := ""
	if !request.Continuation {
		tableIDs, mode, apiErr = s.resolveInitialTableScope(r.Context(), incidentID, request.TableScope)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
	}
	result, apiErr := s.queryAcceptedRows(r.Context(), principal.User.ID.String(), "nf.rows.query", incidentID, tableIDs, mode, request)
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

func (s *Service) handleRejectedRowsQuery(w http.ResponseWriter, r *http.Request) {
	incidentID, tableID, ok := parseIncidentTablePathValues(w, r)
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
	request, apiErr := decodeRejectedRowsQueryRequest(r.Body, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, apiErr := s.queryRejectedRows(r.Context(), principal.User.ID.String(), incidentID, tableID, request)
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

func (s *Service) queryAcceptedRows(ctx context.Context, actorID string, route string, incidentID uuid.UUID, initialTableIDs []string, initialMode string, request RowQueryRequest) (map[string]any, *httpapi.APIError) {
	var offset int
	var tableIDs []string
	mode := initialMode
	filters := request.Filters
	sortSpecs := request.Sort
	limit := request.Limit
	var queryEcho map[string]any
	if request.Continuation {
		payload, reason := s.cursorCodec.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalid(reason)
		}
		if payload.Route != route || payload.ActorUserID != actorID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalid(payloadMismatchReason(payload, route, actorID, incidentID.String()))
		}
		tableIDs = splitTableIDs(payload.Scope["table_ids"])
		mode = payload.Scope["mode"]
		limit = payload.Limit
		offset = payload.Offset
		if len(tableIDs) == 0 {
			return nil, cursorInvalid("scope_stale")
		}
		if len(initialTableIDs) > 0 && !sameStringSet(tableIDs, initialTableIDs) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
		echo, echoMap, err := decodeAcceptedRowsQueryEcho(payload.QueryEcho)
		if err != nil {
			return nil, cursorInvalid("malformed")
		}
		filters = echo.Filters
		sortSpecs = echo.Sort
		queryEcho = echoMap
		if payload.QueryHash != queryHash(queryEcho) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
	} else {
		tableIDs = append([]string(nil), initialTableIDs...)
		sort.Strings(tableIDs)
		queryEcho = acceptedRowsQueryEcho(filters, sortSpecs, effectiveSort(sortSpecs), tableIDs)
	}
	if apiErr := s.ensureActiveTables(ctx, incidentID, tableIDs); apiErr != nil {
		return nil, apiErr
	}
	rows, err := s.store.ListRowsForTables(ctx, incidentID, tableIDs)
	if err != nil {
		return nil, httpapi.InternalAPIError(err)
	}
	filtered, apiErr := filterRows(rows, filters)
	if apiErr != nil {
		return nil, apiErr
	}
	sorted := sortRows(filtered, sortSpecs)
	page, nextOffset := pageFlowRows(sorted, offset, limit)
	rowResources := make([]any, 0, len(page))
	for _, row := range page {
		rowResources = append(rowResources, rowResource(row))
	}
	scope := map[string]string{"mode": mode, "table_ids": strings.Join(tableIDs, ",")}
	queryHashValue := queryHash(queryEcho)
	var queryEchoRaw json.RawMessage
	if encoded, err := json.Marshal(queryEcho); err == nil {
		queryEchoRaw = encoded
	}
	var nextToken *string
	if nextOffset != nil {
		token, err := s.cursorCodec.Encode(CursorBinding{
			Route:       route,
			ActorUserID: actorID,
			IncidentID:  incidentID.String(),
			Scope:       scope,
			QueryHash:   queryHashValue,
			QueryEcho:   queryEchoRaw,
			Limit:       limit,
		}, *nextOffset)
		if err != nil {
			return nil, httpapi.InternalAPIError(err)
		}
		nextToken = &token
	}
	return map[string]any{
		"table_scope": map[string]any{"mode": mode, "table_ids": tableIDs},
		"rows":        rowResources,
		"meta": map[string]any{
			"query": queryEcho,
			"paging": map[string]any{
				"limit":             limit,
				"returned_count":    len(rowResources),
				"next_cursor_token": nextToken,
			},
		},
	}, nil
}

func (s *Service) queryRejectedRows(ctx context.Context, actorID string, incidentID uuid.UUID, tableID string, request RejectedRowsQueryRequest) (map[string]any, *httpapi.APIError) {
	offset := 0
	limit := request.Limit
	var queryEcho map[string]any
	if request.Continuation {
		payload, reason := s.cursorCodec.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalid(reason)
		}
		if payload.Route != "nf.rejected_rows.query" || payload.ActorUserID != actorID || payload.IncidentID != incidentID.String() || payload.Scope["table_ids"] != tableID {
			return nil, cursorInvalid(payloadMismatchReason(payload, "nf.rejected_rows.query", actorID, incidentID.String()))
		}
		limit = payload.Limit
		offset = payload.Offset
		echoRequest, echoMap, err := decodeRejectedRowsQueryEcho(payload.QueryEcho)
		if err != nil {
			return nil, cursorInvalid("malformed")
		}
		request = echoRequest
		request.Continuation = true
		request.CursorToken = ""
		request.Limit = limit
		queryEcho = echoMap
		if payload.QueryHash != queryHash(queryEcho) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
	} else {
		queryEcho = rejectedRowsQueryEcho(request)
	}
	if apiErr := s.ensureActiveTables(ctx, incidentID, []string{tableID}); apiErr != nil {
		return nil, apiErr
	}
	diagnostics, err := s.store.ListRejectedRowDiagnostics(ctx, incidentID, tableID)
	if err != nil {
		return nil, tableReadError(err)
	}
	filtered := filterDiagnostics(diagnostics, request)
	page, nextOffset := pageDiagnostics(filtered, offset, limit)
	resources := make([]any, 0, len(page))
	for _, diagnostic := range page {
		resources = append(resources, diagnosticResource(diagnostic))
	}
	queryEchoRaw, _ := json.Marshal(queryEcho)
	var nextToken *string
	if nextOffset != nil {
		token, err := s.cursorCodec.Encode(CursorBinding{
			Route:       "nf.rejected_rows.query",
			ActorUserID: actorID,
			IncidentID:  incidentID.String(),
			Scope:       map[string]string{"mode": "active_table", "table_ids": tableID},
			QueryHash:   queryHash(queryEcho),
			QueryEcho:   queryEchoRaw,
			Limit:       limit,
		}, *nextOffset)
		if err != nil {
			return nil, httpapi.InternalAPIError(err)
		}
		nextToken = &token
	}
	return map[string]any{
		"schema_id":             "cartulary.network_flow.rejected_rows_query_result.v1",
		"network_flow_table_id": tableID,
		"diagnostics":           resources,
		"meta": map[string]any{
			"query": queryEcho,
			"paging": map[string]any{
				"limit":             limit,
				"returned_count":    len(resources),
				"next_cursor_token": nextToken,
			},
		},
	}, nil
}

func (s *Service) resolveInitialTableScope(ctx context.Context, incidentID uuid.UUID, scope TableScope) ([]string, string, *httpapi.APIError) {
	switch scope.Mode {
	case "active_table":
		if scope.ActiveTableID == "" {
			return nil, "", invalidTableScope("table_scope", "empty_resolved_scope")
		}
		return []string{scope.ActiveTableID}, scope.Mode, nil
	case "selected_tables":
		tableIDs := append([]string(nil), scope.SelectedTableIDs...)
		sort.Strings(tableIDs)
		if len(tableIDs) == 0 {
			return nil, "", invalidTableScope("table_scope", "empty_resolved_scope")
		}
		return tableIDs, scope.Mode, nil
	case "all_active_tables":
		tables, err := s.store.ListActiveTables(ctx, incidentID)
		if err != nil {
			return nil, "", httpapi.InternalAPIError(err)
		}
		tableIDs := make([]string, 0, len(tables))
		for _, table := range tables {
			tableIDs = append(tableIDs, table.TableID)
		}
		sort.Strings(tableIDs)
		if len(tableIDs) == 0 {
			return nil, "", invalidTableScope("table_scope", "empty_resolved_scope")
		}
		return tableIDs, scope.Mode, nil
	default:
		return nil, "", invalidTableScope("mode", "unknown_mode")
	}
}

func (s *Service) ensureActiveTables(ctx context.Context, incidentID uuid.UUID, tableIDs []string) *httpapi.APIError {
	for _, tableID := range tableIDs {
		if _, err := s.store.GetActiveTable(ctx, incidentID, tableID); err != nil {
			if errors.Is(err, ErrTableNotFound) {
				return networkFlowAPIError(http.StatusNotFound, "network_flow_table_not_found", "network_flow_table_id", "not_found")
			}
			if errors.Is(err, ErrTableNotActive) {
				return networkFlowAPIError(http.StatusConflict, "network_flow_table_not_active", "network_flow_table_id", "soft_deleted")
			}
			return httpapi.InternalAPIError(err)
		}
	}
	return nil
}

func filterDiagnostics(in []RejectedRowDiagnostic, request RejectedRowsQueryRequest) []RejectedRowDiagnostic {
	errorCodes := stringSet(request.ErrorCodes)
	fieldKeys := stringSet(request.FieldKeys)
	out := make([]RejectedRowDiagnostic, 0, len(in))
	for _, diagnostic := range in {
		if len(errorCodes) > 0 {
			if _, ok := errorCodes[diagnostic.ErrorCode]; !ok {
				continue
			}
		}
		if len(fieldKeys) > 0 {
			if diagnostic.FieldKey == nil {
				continue
			}
			if _, ok := fieldKeys[*diagnostic.FieldKey]; !ok {
				continue
			}
		}
		if request.SourceRowGTE != nil && diagnostic.SourceRowNumber < *request.SourceRowGTE {
			continue
		}
		if request.SourceRowLTE != nil && diagnostic.SourceRowNumber > *request.SourceRowLTE {
			continue
		}
		out = append(out, diagnostic)
	}
	return out
}

func acceptedRowsQueryEcho(filters []Filter, sortSpecs []SortSpec, effective []SortSpec, tableIDs []string) map[string]any {
	return map[string]any{
		"filters":        filters,
		"sort":           sortSpecs,
		"effective_sort": effective,
		"table_ids":      tableIDs,
	}
}

func rejectedRowsQueryEcho(request RejectedRowsQueryRequest) map[string]any {
	var sourceRange any
	if request.SourceRowGTE != nil || request.SourceRowLTE != nil {
		value := map[string]any{"gte": nil, "lte": nil}
		if request.SourceRowGTE != nil {
			value["gte"] = *request.SourceRowGTE
		}
		if request.SourceRowLTE != nil {
			value["lte"] = *request.SourceRowLTE
		}
		sourceRange = value
	}
	return map[string]any{
		"error_codes":      request.ErrorCodes,
		"field_keys":       request.FieldKeys,
		"source_row_range": sourceRange,
		"effective_sort": []map[string]string{
			{"field_key": "source_row_number", "direction": "asc"},
			{"field_key": "source_column_ordinal", "direction": "asc"},
			{"field_key": "field_key", "direction": "asc"},
			{"field_key": "error_code", "direction": "asc"},
			{"field_key": "reason_code", "direction": "asc"},
			{"field_key": "diagnostic_id", "direction": "asc"},
		},
	}
}

type acceptedRowsQueryEchoPayload struct {
	Filters       []Filter   `json:"filters"`
	Sort          []SortSpec `json:"sort"`
	EffectiveSort []SortSpec `json:"effective_sort"`
	TableIDs      []string   `json:"table_ids"`
}

func decodeAcceptedRowsQueryEcho(raw json.RawMessage) (acceptedRowsQueryEchoPayload, map[string]any, error) {
	var payload acceptedRowsQueryEchoPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return acceptedRowsQueryEchoPayload{}, nil, err
	}
	return payload, acceptedRowsQueryEcho(payload.Filters, payload.Sort, payload.EffectiveSort, payload.TableIDs), nil
}

func decodeRejectedRowsQueryEcho(raw json.RawMessage) (RejectedRowsQueryRequest, map[string]any, error) {
	var payload struct {
		ErrorCodes     []string        `json:"error_codes"`
		FieldKeys      []string        `json:"field_keys"`
		SourceRowRange json.RawMessage `json:"source_row_range"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return RejectedRowsQueryRequest{}, nil, err
	}
	request := RejectedRowsQueryRequest{
		ErrorCodes: payload.ErrorCodes,
		FieldKeys:  payload.FieldKeys,
	}
	if len(payload.SourceRowRange) > 0 && string(payload.SourceRowRange) != "null" {
		gte, lte, apiErr := decodeIntegerRange(payload.SourceRowRange)
		if apiErr != nil {
			return RejectedRowsQueryRequest{}, nil, fmt.Errorf("decode source row range")
		}
		request.SourceRowGTE = gte
		request.SourceRowLTE = lte
	}
	return request, rejectedRowsQueryEcho(request), nil
}

func pageFlowRows(rows []FlowRow, offset int, limit int) ([]FlowRow, *int) {
	if offset >= len(rows) {
		return []FlowRow{}, nil
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	page := rows[offset:end]
	if end >= len(rows) {
		return page, nil
	}
	return page, &end
}

func pageDiagnostics(rows []RejectedRowDiagnostic, offset int, limit int) ([]RejectedRowDiagnostic, *int) {
	if offset >= len(rows) {
		return []RejectedRowDiagnostic{}, nil
	}
	end := offset + limit
	if end > len(rows) {
		end = len(rows)
	}
	page := rows[offset:end]
	if end >= len(rows) {
		return page, nil
	}
	return page, &end
}

type tableRenameRequest struct {
	ClientTxnID      string
	BaseTableVersion int64
	DisplayName      string
}

type tableSoftDeleteRequest struct {
	ClientTxnID      string
	BaseTableVersion int64
}

func decodeRenameRequest(r *http.Request) (tableRenameRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
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
	return tableRenameRequest{ClientTxnID: clientTxnID, BaseTableVersion: int64(version), DisplayName: displayName}, nil
}

func decodeSoftDeleteRequest(r *http.Request) (tableSoftDeleteRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
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

func (s *Service) commitTableRenameRoute(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableRenameRequest, requestHash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	idempotencyKey := tableMutationIdempotencyKey(routeKeyTablesPatch, actorUserID, incidentID, tableID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	tx, err := s.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, 0, httpapi.InternalAPIError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	table, err := s.store.renameTableTx(ctx, tx, RenameTableParams{
		IncidentID:             incidentID,
		ActorUserID:            actorUserID,
		TableID:                tableID,
		BaseTableVersion:       request.BaseTableVersion,
		DisplayName:            request.DisplayName,
		ClientTxnID:            request.ClientTxnID,
		RequestID:              requestID,
		DisplayNameDigestKeyID: s.safeDigestKeyID,
		DisplayNameDigestKey:   s.safeDigestKey,
		Now:                    s.now(),
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
			return payload, status, apiErr
		}
		return nil, 0, tableMutationAPIError(err, request.ClientTxnID)
	}
	payload := tableMutationPayload(table)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return nil, 0, httpapi.ClientTxnConflictError(request.ClientTxnID)
		}
		return nil, 0, httpapi.InternalAPIError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, httpapi.InternalAPIError(fmt.Errorf("commit network flow table rename route: %w", err))
	}
	if table.TableVersion != request.BaseTableVersion {
		s.publishTableResourceChange(incidentID, table.TableID, platformws.ExtensionResourceChangeKindInvalidate, platformws.ExtensionResourceReasonRenamed)
	}
	return payload, http.StatusOK, nil
}

func (s *Service) commitTableSoftDeleteRoute(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableSoftDeleteRequest, requestHash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	idempotencyKey := tableMutationIdempotencyKey(routeKeyTablesDelete, actorUserID, incidentID, tableID, request.ClientTxnID)
	if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
		return payload, status, apiErr
	}
	tx, err := s.store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, 0, httpapi.InternalAPIError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	table, err := s.store.softDeleteTableTx(ctx, tx, SoftDeleteTableParams{
		IncidentID:       incidentID,
		ActorUserID:      actorUserID,
		TableID:          tableID,
		BaseTableVersion: request.BaseTableVersion,
		ClientTxnID:      request.ClientTxnID,
		RequestID:        requestID,
		Now:              s.now(),
	})
	if err != nil {
		if payload, status, replayed, apiErr := s.replayTableMutationIfPresent(ctx, idempotencyKey, requestHash); replayed || apiErr != nil {
			return payload, status, apiErr
		}
		return nil, 0, tableMutationAPIError(err, request.ClientTxnID)
	}
	payload := tableMutationPayload(table)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return nil, 0, httpapi.ClientTxnConflictError(request.ClientTxnID)
		}
		return nil, 0, httpapi.InternalAPIError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, 0, httpapi.InternalAPIError(fmt.Errorf("commit network flow table soft delete route: %w", err))
	}
	s.publishTableResourceChange(incidentID, table.TableID, platformws.ExtensionResourceChangeKindRemove, platformws.ExtensionResourceReasonSoftDeleted)
	return payload, http.StatusOK, nil
}

func (s *Service) publishTableResourceChange(incidentID uuid.UUID, tableID string, changeKind string, reasonCode string) {
	if s == nil || s.hub == nil || incidentID == uuid.Nil || strings.TrimSpace(tableID) == "" {
		return
	}
	_ = s.hub.PublishExtensionResourceChange(incidentID, platformws.ExtensionResourceChangePayload{
		ExtensionProfileID: ProfileID,
		ResourceKind:       "network_flow_table",
		ResourceID:         tableID,
		ChangeKind:         changeKind,
		ReasonCode:         reasonCode,
		WorkspaceRefs: []platformws.ExtensionWorkspaceRef{
			{
				Kind:               "extension_workspace",
				ExtensionProfileID: ProfileID,
				WorkspaceKey:       WorkspaceKeyNetworkAnalysis,
			},
		},
	})
}

func (s *Service) replayTableMutationIfPresent(ctx context.Context, key authn.RouteIdempotencyKey, requestHash []byte) (map[string]any, int, bool, *httpapi.APIError) {
	existing, err := s.authStore.GetRouteIdempotency(ctx, key)
	if err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return nil, 0, true, httpapi.ClientTxnConflictError(key.ClientTxnID)
		}
		payload, err := decodeStoredNetworkFlowResponse(existing.ResponseJSON)
		if err != nil {
			return nil, 0, true, httpapi.InternalAPIError(err)
		}
		return payload, existing.StatusCode, true, nil
	}
	if !errors.Is(err, authn.ErrNotFound) {
		return nil, 0, false, httpapi.InternalAPIError(err)
	}
	return nil, 0, false, nil
}

func tableMutationPayload(table TableRecord) map[string]any {
	return map[string]any{
		"schema_id": "cartulary.network_flow.table_mutation_result.v1",
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
	return networkFlowRequestHash(map[string]any{
		"route_key":          routeKeyTablesPatch,
		"network_flow_table": tableID,
		"client_txn_id":      request.ClientTxnID,
		"base_table_version": request.BaseTableVersion,
		"display_name":       request.DisplayName,
	})
}

func tableSoftDeleteRequestHash(tableID string, request tableSoftDeleteRequest) []byte {
	return networkFlowRequestHash(map[string]any{
		"route_key":          routeKeyTablesDelete,
		"network_flow_table": tableID,
		"client_txn_id":      request.ClientTxnID,
		"base_table_version": request.BaseTableVersion,
	})
}

func networkFlowRequestHash(value any) []byte {
	sum := sha256.Sum256(canonicalJSON(value))
	out := make([]byte, len(sum))
	copy(out, sum[:])
	return out
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

func (s *Service) authenticate(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	return httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

func parseIncidentPathValue(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	incidentID, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil {
		http.NotFound(w, r)
		return uuid.Nil, false
	}
	return incidentID, true
}

func parseIncidentTablePathValues(w http.ResponseWriter, r *http.Request) (uuid.UUID, string, bool) {
	incidentID, ok := parseIncidentPathValue(w, r)
	if !ok {
		return uuid.Nil, "", false
	}
	tableID := r.PathValue("network_flow_table_id")
	if !strings.HasPrefix(tableID, "nft_") {
		http.NotFound(w, r)
		return uuid.Nil, "", false
	}
	return incidentID, tableID, true
}

func tableReadError(err error) *httpapi.APIError {
	if errors.Is(err, ErrTableNotFound) {
		return networkFlowAPIError(http.StatusNotFound, "network_flow_table_not_found", "network_flow_table_id", "not_found")
	}
	if errors.Is(err, ErrTableNotActive) {
		return networkFlowAPIError(http.StatusConflict, "network_flow_table_not_active", "network_flow_table_id", "soft_deleted")
	}
	return httpapi.InternalAPIError(err)
}

func tableMutationError(err error) *httpapi.APIError {
	var versionConflict *TableVersionConflictError
	var displayName *InvalidDisplayNameError
	if errors.As(err, &versionConflict) {
		return networkFlowAPIError(http.StatusConflict, "network_flow_table_version_conflict", "base_table_version", "stale_version")
	}
	if errors.As(err, &displayName) {
		return networkFlowAPIError(http.StatusBadRequest, "network_flow_invalid_display_name", "display_name", displayName.ReasonCode)
	}
	if errors.Is(err, ErrTableNameExhausted) {
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

func stringSet(values []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func splitTableIDs(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	sort.Strings(out)
	return out
}

func sameStringSet(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	for index := range leftCopy {
		if leftCopy[index] != rightCopy[index] {
			return false
		}
	}
	return true
}

func payloadMismatchReason(payload CursorPayload, route string, actorID string, incidentID string) string {
	switch {
	case payload.ActorUserID != actorID:
		return "actor_mismatch"
	case payload.Route != route:
		return "route_mismatch"
	case payload.IncidentID != incidentID:
		return "semantic_query_mismatch"
	default:
		return "semantic_query_mismatch"
	}
}
