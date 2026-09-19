package networkflow

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/google/uuid"
	"sort"
	"strings"
)

type readIdentity struct{ ActorID, SessionID, IncidentID uuid.UUID }

func admitRead(ctx context.Context, checker incidentAdmissionChecker, identity readIdentity) *semanticFailure {
	_, err := checker.Check(ctx, identity.IncidentID, identity.ActorID, admission.Requirement{AllowedRoles: admission.RolesMember, Lifecycle: admission.LifecycleOpen})
	if err != nil {
		return applicationAdmissionFailure(err, "")
	}
	return nil
}

type tableQueryApplication struct {
	store           *store
	incidentAccess  incidentAdmissionChecker
	cursorProtector cursorProtector
}
type acceptedRowsOutcome struct {
	Rows      []flowRow
	TableIDs  []string
	Mode      string
	Filters   []queryFilter
	Sort      []sortSpec
	Limit     int
	NextToken *string
}
type diagnosticsOutcome struct {
	Diagnostics []rejectedRowDiagnostic
	Request     rejectedRowsQueryRequest
	Limit       int
	NextToken   *string
}

func (s *tableQueryApplication) profiles(ctx context.Context, who readIdentity) (EffectiveLimits, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return EffectiveLimits{}, failure
	}
	return s.store.limits, nil
}
func (s *tableQueryApplication) list(ctx context.Context, who readIdentity) ([]tableRecord, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return nil, failure
	}
	records, err := s.store.ListActiveTables(ctx, who.IncidentID)
	if err != nil {
		return nil, internalSemanticFailure(err)
	}
	return records, nil
}
func (s *tableQueryApplication) get(ctx context.Context, who readIdentity, tableID string) (tableRecord, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return tableRecord{}, failure
	}
	record, err := s.store.GetActiveTable(ctx, who.IncidentID, tableID)
	if err != nil {
		return tableRecord{}, tableReadFailure(err)
	}
	return record, nil
}
func (s *tableQueryApplication) rows(ctx context.Context, who readIdentity, tableID string, request rowQueryRequest) (*acceptedRowsOutcome, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return nil, failure
	}
	route, mode, ids := "nf.tables.query", "active_table", []string{tableID}
	if tableID == "" {
		route = "nf.rows.query"
		ids = nil
		mode = ""
		if !request.Continuation {
			var failure *semanticFailure
			ids, mode, failure = s.resolveInitialTableScope(ctx, who.IncidentID, request.TableScope)
			if failure != nil {
				return nil, failure
			}
		}
	}
	return s.queryAcceptedRows(ctx, who.ActorID.String(), who.SessionID.String(), route, who.IncidentID, ids, mode, request)
}
func (s *tableQueryApplication) diagnostics(ctx context.Context, who readIdentity, tableID string, request rejectedRowsQueryRequest) (*diagnosticsOutcome, *semanticFailure) {
	if failure := admitRead(ctx, s.incidentAccess, who); failure != nil {
		return nil, failure
	}
	return s.queryRejectedRows(ctx, who.ActorID.String(), who.SessionID.String(), who.IncidentID, tableID, request)
}
func (s *tableQueryApplication) queryAcceptedRows(ctx context.Context, actorID string, sessionID string, route string, incidentID uuid.UUID, initialTableIDs []string, initialMode string, request rowQueryRequest) (*acceptedRowsOutcome, *semanticFailure) {
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
			return nil, cursorInvalid(reason)
		}
		if payload.Route != route || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() {
			return nil, cursorInvalid(payloadMismatchReason(payload, route, actorID, incidentID.String()))
		}
		tableIDs = splitTableIDs(payload.Scope["table_ids"])
		mode = payload.Scope["mode"]
		limit = payload.Limit
		if payload.PositionKind != "row_keyset_v1" {
			return nil, cursorInvalid("malformed")
		}
		decodedPosition, err := decodeRowCursorPosition(payload.Position)
		if err != nil {
			return nil, cursorInvalid("malformed")
		}
		position = &decodedPosition
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
		if !sameSortSpecs(position.EffectiveSort, effectiveSort(sortSpecs)) {
			return nil, cursorInvalid("semantic_query_mismatch")
		}
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
	page, hasMore, err := s.store.QueryRowsPage(ctx, incidentID, tableIDs, filters, sortSpecs, position, limit)
	if err != nil {
		return nil, internalSemanticFailure(err)
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
			return nil, internalSemanticFailure(err)
		}
		nextToken = &token
	}
	return &acceptedRowsOutcome{Rows: page, TableIDs: tableIDs, Mode: mode, Filters: filters, Sort: sortSpecs, Limit: limit, NextToken: nextToken}, nil
}

func (s *tableQueryApplication) queryRejectedRows(ctx context.Context, actorID string, sessionID string, incidentID uuid.UUID, tableID string, request rejectedRowsQueryRequest) (*diagnosticsOutcome, *semanticFailure) {
	var position *diagnosticCursorPosition
	limit := request.Limit
	var queryEcho map[string]any
	if request.Continuation {
		payload, reason := s.cursorProtector.Decode(request.CursorToken)
		if reason != "" {
			return nil, cursorInvalid(reason)
		}
		if payload.Route != "nf.rejected_rows.query" || payload.ActorUserID != actorID || payload.SessionID != sessionID || payload.IncidentID != incidentID.String() || payload.Scope["table_ids"] != tableID {
			return nil, cursorInvalid(payloadMismatchReason(payload, "nf.rejected_rows.query", actorID, incidentID.String()))
		}
		limit = payload.Limit
		if payload.PositionKind != "diagnostic_keyset_v1" {
			return nil, cursorInvalid("malformed")
		}
		var decodedPosition diagnosticCursorPosition
		if err := json.Unmarshal(payload.Position, &decodedPosition); err != nil || decodedPosition.SourceRowNumber < 1 || decodedPosition.ErrorCode == "" || decodedPosition.ReasonCode == "" || decodedPosition.DiagnosticID == "" {
			return nil, cursorInvalid("malformed")
		}
		position = &decodedPosition
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
	page, hasMore, err := s.store.QueryRejectedDiagnosticsPage(ctx, incidentID, tableID, request, position, limit)
	if err != nil {
		return nil, tableReadFailure(err)
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
			return nil, internalSemanticFailure(err)
		}
		nextToken = &token
	}
	return &diagnosticsOutcome{Diagnostics: page, Request: request, Limit: limit, NextToken: nextToken}, nil
}

func (s *tableQueryApplication) resolveInitialTableScope(ctx context.Context, incidentID uuid.UUID, scope tableScope) ([]string, string, *semanticFailure) {
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
			return nil, "", internalSemanticFailure(err)
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

func (s *tableQueryApplication) ensureActiveTables(ctx context.Context, incidentID uuid.UUID, tableIDs []string) *semanticFailure {
	return s.store.ensureActiveTables(ctx, incidentID, tableIDs)
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
		gte, lte, apiErr := decodeIntegerRange(payload.SourceRowRange)
		if apiErr != nil {
			return rejectedRowsQueryRequest{}, nil, fmt.Errorf("decode source row range")
		}
		request.SourceRowGTE = gte
		request.SourceRowLTE = lte
	}
	return request, rejectedRowsQueryEcho(request), nil
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
