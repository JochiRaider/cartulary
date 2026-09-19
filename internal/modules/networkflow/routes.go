package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

const (
	routeKeyTablesPatch  = "nf.tables.patch"
	routeKeyTablesDelete = "nf.tables.delete"
)

type routeService struct {
	store           *store
	incidentAccess  incidentAdmissionChecker
	authStore       *authn.Store
	keys            authn.MasterKeys
	cursorProtector cursorProtector
	safeDigester    safeDigester
	now             func() time.Time
	graphComposer   *graphSourceComposer
	transactions    *crossownertransaction.Coordinator
	savedGraphs     *savedGraphApplication
	jobManager      GraphViewJobManager
}

type incidentAdmissionChecker interface {
	Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
	CheckTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

func newRouteService(deps httpapi.DependencySet, module *Module) (*routeService, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if module.cursorProtector == nil || module.safeDigester == nil {
		return nil, fmt.Errorf("network flow configured key rings unavailable")
	}
	return &routeService{
		store:           module.store,
		incidentAccess:  admission.NewChecker(deps.PostgresHandle()),
		authStore:       authn.NewStore(deps.PostgresHandle()),
		keys:            keys,
		cursorProtector: module.cursorProtector,
		safeDigester:    module.safeDigester,
		now:             now,
		graphComposer:   module.graphComposer,
		transactions:    module.transactions,
		savedGraphs: &savedGraphApplication{
			store: module.store, incidentAccess: admission.NewChecker(deps.PostgresHandle()),
			receipts:      savedGraphReceiptAdapter{reader: authn.NewStore(deps.PostgresHandle())},
			graphViewJobs: module.graphViewJobs, jobRunner: module.jobRunner, now: now,
		},
		jobManager: module.jobManager,
	}, nil
}

func (s *routeService) handleSourceProfiles(w http.ResponseWriter, r *http.Request) {
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
		"schema_id":        "cartulary.network_flow.source_profile_list.v2",
		"source_profiles":  []any{sourceProfileResource()},
		"effective_limits": effectiveLimitsResource(s.store.limits),
		"meta":             map[string]any{"count": 1},
	})
}

func (s *routeService) handleTablesCollection(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, ok := parseIncidentPathValue(w, r)
	if !ok {
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
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
		"schema_id": "cartulary.network_flow_table_list.v1",
		"tables":    resources,
		"meta":      map[string]any{"count": len(resources)},
	})
}

func (s *routeService) handleTableResource(w http.ResponseWriter, r *http.Request) {
	stateChanging := r.Method == http.MethodPatch || r.Method == http.MethodDelete
	principal, apiErr := s.authenticate(r, stateChanging)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, tableID, ok := parseIncidentTablePathValues(w, r)
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
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
			"schema_id": "cartulary.network_flow_table_get.v1",
			"table":     tableResource(table),
		})
	case http.MethodPatch:
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
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
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesReviewerAdmin, "reviewer|admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
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

func (s *routeService) handleTableRowsQuery(w http.ResponseWriter, r *http.Request) {
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
	request, apiErr := decodeAcceptedRowQueryRequestHTTP(r.Body, schemaTableQueryRequest, schemaTableQueryContinuation, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, apiErr := s.queryAcceptedRows(r.Context(), principal.User.ID.String(), principal.Session.ID.String(), "nf.tables.query", incidentID, []string{tableID}, "active_table", request)
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
	request, apiErr := decodeAcceptedRowQueryRequestHTTP(r.Body, schemaRowsQueryRequest, schemaRowsQueryContinuation, s.store.limits)
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
	result, apiErr := s.queryAcceptedRows(r.Context(), principal.User.ID.String(), principal.Session.ID.String(), "nf.rows.query", incidentID, tableIDs, mode, request)
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
	request, apiErr := decodeRejectedRowsQueryRequestHTTP(r.Body, s.store.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, apiErr := s.queryRejectedRows(r.Context(), principal.User.ID.String(), principal.Session.ID.String(), incidentID, tableID, request)
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

func (s *routeService) queryAcceptedRows(ctx context.Context, actorID string, sessionID string, route string, incidentID uuid.UUID, initialTableIDs []string, initialMode string, request rowQueryRequest) (map[string]any, *httpapi.APIError) {
	var position *rowCursorPosition
	var tableIDs []string
	mode := initialMode
	filters := request.Filters
	sortSpecs := request.Sort
	limit := request.Limit
	var queryEcho map[string]any
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalidHTTP(reason)
		}
		if payload.Route != route || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalidHTTP(payloadMismatchReason(payload, route, actorID, incidentID.String()))
		}
		tableIDs = splitTableIDs(payload.Scope["table_ids"])
		mode = payload.Scope["mode"]
		limit = payload.Limit
		if payload.PositionKind != "row_keyset_v1" {
			return nil, cursorInvalidHTTP("malformed")
		}
		decodedPosition, err := decodeRowCursorPosition(payload.Position)
		if err != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		position = &decodedPosition
		if len(tableIDs) == 0 {
			return nil, cursorInvalidHTTP("scope_stale")
		}
		if len(initialTableIDs) > 0 && !sameStringSet(tableIDs, initialTableIDs) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
		echo, echoMap, err := decodeAcceptedRowsQueryEcho(payload.QueryEcho)
		if err != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		filters = echo.Filters
		sortSpecs = echo.Sort
		queryEcho = echoMap
		if !sameSortSpecs(position.EffectiveSort, effectiveSort(sortSpecs)) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
		if payload.QueryHash != queryHash(queryEcho) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
	} else {
		tableIDs = append([]string(nil), initialTableIDs...)
		sort.Strings(tableIDs)
		queryEcho = acceptedRowsQueryEcho(filters, sortSpecs, effectiveSort(sortSpecs), tableIDs)
	}
	if apiErr := s.ensureActiveTables(ctx, incidentID, tableIDs); apiErr != nil {
		return nil, apiErr
	}
	page, hasMore, err := s.store.QueryRowsPage(ctx, incidentID, tableIDs, filters, sortSpecs, position, limit)
	if err != nil {
		return nil, httpapi.InternalAPIError(err)
	}
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
	if hasMore && len(page) > 0 {
		token, err := s.cursorProtector.Encode(cursorBinding{
			Route:       route,
			ActorUserID: actorID,
			SessionID:   sessionID,
			IncidentID:  incidentID.String(),
			Scope:       scope,
			QueryHash:   queryHashValue,
			QueryEcho:   queryEchoRaw,
			Limit:       limit,
		}, "row_keyset_v1", newRowCursorPosition(page[len(page)-1], sortSpecs))
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

func (s *routeService) queryRejectedRows(ctx context.Context, actorID string, sessionID string, incidentID uuid.UUID, tableID string, request rejectedRowsQueryRequest) (map[string]any, *httpapi.APIError) {
	var position *diagnosticCursorPosition
	limit := request.Limit
	var queryEcho map[string]any
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalidHTTP(reason)
		}
		if payload.Route != "nf.rejected_rows.query" || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() || payload.Scope["table_ids"] != tableID {
			return nil, cursorInvalidHTTP(payloadMismatchReason(payload, "nf.rejected_rows.query", actorID, incidentID.String()))
		}
		limit = payload.Limit
		if payload.PositionKind != "diagnostic_keyset_v1" {
			return nil, cursorInvalidHTTP("malformed")
		}
		var decodedPosition diagnosticCursorPosition
		if err := json.Unmarshal(payload.Position, &decodedPosition); err != nil || decodedPosition.SourceRowNumber < 1 || decodedPosition.ErrorCode == "" || decodedPosition.ReasonCode == "" || decodedPosition.DiagnosticID == "" {
			return nil, cursorInvalidHTTP("malformed")
		}
		position = &decodedPosition
		echoRequest, echoMap, err := decodeRejectedRowsQueryEcho(payload.QueryEcho)
		if err != nil {
			return nil, cursorInvalidHTTP("malformed")
		}
		request = echoRequest
		request.Continuation = true
		request.CursorToken = ""
		request.Limit = limit
		queryEcho = echoMap
		if payload.QueryHash != queryHash(queryEcho) {
			return nil, cursorInvalidHTTP("semantic_query_mismatch")
		}
	} else {
		queryEcho = rejectedRowsQueryEcho(request)
	}
	if apiErr := s.ensureActiveTables(ctx, incidentID, []string{tableID}); apiErr != nil {
		return nil, apiErr
	}
	page, hasMore, err := s.store.QueryRejectedDiagnosticsPage(ctx, incidentID, tableID, request, position, limit)
	if err != nil {
		return nil, tableReadError(err)
	}
	resources := make([]any, 0, len(page))
	for _, diagnostic := range page {
		resources = append(resources, diagnosticResource(diagnostic))
	}
	queryEchoRaw, _ := json.Marshal(queryEcho)
	var nextToken *string
	if hasMore && len(page) > 0 {
		token, err := s.cursorProtector.Encode(cursorBinding{
			Route:       "nf.rejected_rows.query",
			ActorUserID: actorID,
			SessionID:   sessionID,
			IncidentID:  incidentID.String(),
			Scope:       map[string]string{"mode": "active_table", "table_ids": tableID},
			QueryHash:   queryHash(queryEcho),
			QueryEcho:   queryEchoRaw,
			Limit:       limit,
		}, "diagnostic_keyset_v1", newDiagnosticCursorPosition(page[len(page)-1]))
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

func (s *routeService) resolveInitialTableScope(ctx context.Context, incidentID uuid.UUID, scope tableScope) ([]string, string, *httpapi.APIError) {
	switch scope.Mode {
	case "active_table":
		if scope.ActiveTableID == "" {
			return nil, "", invalidTableScopeHTTP("table_scope", "empty_resolved_scope")
		}
		return []string{scope.ActiveTableID}, scope.Mode, nil
	case "selected_tables":
		tableIDs := append([]string(nil), scope.SelectedTableIDs...)
		sort.Strings(tableIDs)
		if len(tableIDs) == 0 {
			return nil, "", invalidTableScopeHTTP("table_scope", "empty_resolved_scope")
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
			return nil, "", invalidTableScopeHTTP("table_scope", "empty_resolved_scope")
		}
		return tableIDs, scope.Mode, nil
	default:
		return nil, "", invalidTableScopeHTTP("mode", "unknown_mode")
	}
}

func (s *routeService) ensureActiveTables(ctx context.Context, incidentID uuid.UUID, tableIDs []string) *httpapi.APIError {
	for _, tableID := range tableIDs {
		if _, err := s.store.GetActiveTable(ctx, incidentID, tableID); err != nil {
			if errors.Is(err, errTableNotFound) {
				return networkFlowAPIError(http.StatusNotFound, "network_flow_table_not_found", "network_flow_table_id", "not_found")
			}
			if errors.Is(err, errTableNotActive) {
				return networkFlowAPIError(http.StatusConflict, "network_flow_table_not_active", "network_flow_table_id", "soft_deleted")
			}
			return httpapi.InternalAPIError(err)
		}
	}
	return nil
}

func acceptedRowsQueryEcho(filters []queryFilter, sortSpecs []sortSpec, effective []sortSpec, tableIDs []string) map[string]any {
	if filters == nil {
		filters = []queryFilter{}
	}
	if sortSpecs == nil {
		sortSpecs = []sortSpec{}
	}
	if effective == nil {
		effective = []sortSpec{}
	}
	if tableIDs == nil {
		tableIDs = []string{}
	}
	return map[string]any{
		"filters":        filters,
		"sort":           sortSpecs,
		"effective_sort": effective,
		"table_ids":      tableIDs,
	}
}

func rejectedRowsQueryEcho(request rejectedRowsQueryRequest) map[string]any {
	errorCodes := request.ErrorCodes
	if errorCodes == nil {
		errorCodes = []string{}
	}
	fieldKeys := request.FieldKeys
	if fieldKeys == nil {
		fieldKeys = []string{}
	}
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
		"error_codes":      errorCodes,
		"field_keys":       fieldKeys,
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
	Filters       []queryFilter `json:"filters"`
	Sort          []sortSpec    `json:"sort"`
	EffectiveSort []sortSpec    `json:"effective_sort"`
	TableIDs      []string      `json:"table_ids"`
}

func decodeAcceptedRowsQueryEcho(raw json.RawMessage) (acceptedRowsQueryEchoPayload, map[string]any, error) {
	var payload acceptedRowsQueryEchoPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return acceptedRowsQueryEchoPayload{}, nil, err
	}
	return payload, acceptedRowsQueryEcho(payload.Filters, payload.Sort, payload.EffectiveSort, payload.TableIDs), nil
}

func decodeRejectedRowsQueryEcho(raw json.RawMessage) (rejectedRowsQueryRequest, map[string]any, error) {
	var payload struct {
		ErrorCodes     []string        `json:"error_codes"`
		FieldKeys      []string        `json:"field_keys"`
		SourceRowRange json.RawMessage `json:"source_row_range"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return rejectedRowsQueryRequest{}, nil, err
	}
	request := rejectedRowsQueryRequest{
		ErrorCodes: payload.ErrorCodes,
		FieldKeys:  payload.FieldKeys,
	}
	if len(payload.SourceRowRange) > 0 && string(payload.SourceRowRange) != "null" {
		gte, lte, apiErr := decodeIntegerRangeHTTP(payload.SourceRowRange)
		if apiErr != nil {
			return rejectedRowsQueryRequest{}, nil, fmt.Errorf("decode source row range")
		}
		request.SourceRowGTE = gte
		request.SourceRowLTE = lte
	}
	return request, rejectedRowsQueryEcho(request), nil
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
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "client_txn_id", "base_table_version", "display_name"); apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONStringHTTP(raw, "client_txn_id")
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	displayName, apiErr := requiredJSONStringHTTP(raw, "display_name")
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	version, apiErr := decodePositiveIntHTTP(raw["base_table_version"], "base_table_version")
	if apiErr != nil {
		return tableRenameRequest{}, apiErr
	}
	normalized, err := normalizeExplicitDisplayName(displayName)
	if err != nil {
		return tableRenameRequest{}, tableMutationError(err)
	}
	return tableRenameRequest{ClientTxnID: clientTxnID, BaseTableVersion: int64(version), DisplayName: normalized}, nil
}

func decodeSoftDeleteRequest(r *http.Request) (tableSoftDeleteRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObjectHTTP(r.Body)
	if apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	if apiErr := ensureAllowedMembersHTTP(raw, "client_txn_id", "base_table_version"); apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	clientTxnID, apiErr := requiredJSONStringHTTP(raw, "client_txn_id")
	if apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	version, apiErr := decodePositiveIntHTTP(raw["base_table_version"], "base_table_version")
	if apiErr != nil {
		return tableSoftDeleteRequest{}, apiErr
	}
	return tableSoftDeleteRequest{ClientTxnID: clientTxnID, BaseTableVersion: int64(version)}, nil
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

func (s *routeService) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *httpapi.APIError) {
	return s.requireIncidentRole(ctx, incidentID, userID, admission.RolesMember, "")
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

func payloadMismatchReason(payload cursorPayload, route string, actorID string, incidentID string) string {
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
