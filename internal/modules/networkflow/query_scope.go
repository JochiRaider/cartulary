package networkflow

import (
	"bytes"
	"context"
	"encoding/json"
	"sort"

	"github.com/google/uuid"
)

type tableScopeReader interface {
	ListActiveTables(context.Context, uuid.UUID) ([]tableRecord, error)
	GetActiveTable(context.Context, uuid.UUID, string) (tableRecord, error)
}

func invalidTableScope(scope tableScope, reason string) *semanticFailure {
	details := failureDetails{ScopeMode: scope.Mode, TableIDs: []string{}}
	switch scope.Mode {
	case "active_table":
		if scope.ActiveTableID != "" {
			details.TableIDs = []string{scope.ActiveTableID}
		}
	case "selected_tables":
		details.TableIDs = append(details.TableIDs, scope.SelectedTableIDs...)
	case "all_active_tables":
	default:
		details.ScopeMode = ""
	}
	if reason == "selected_table_limit_exceeded" {
		details.LimitKey = "network_flow.max_selected_tables_per_query"
	}
	return &semanticFailure{kind: failureInvalidTableScope, reason: reason, details: details}
}
func validateSelectedTableIDs(scope tableScope, limits EffectiveLimits) *semanticFailure {
	if len(scope.SelectedTableIDs) == 0 {
		return invalidTableScope(scope, "empty_resolved_scope")
	}
	if int64(len(scope.SelectedTableIDs)) > limits.MaxSelectedTablesPerQuery {
		return invalidTableScope(scope, "selected_table_limit_exceeded")
	}
	seen := map[string]struct{}{}
	for _, id := range scope.SelectedTableIDs {
		if id == "" {
			return invalidTableScope(scope, "empty_resolved_scope")
		}
		if _, ok := seen[id]; ok {
			return invalidTableScope(scope, "duplicate_table_id")
		}
		seen[id] = struct{}{}
	}
	return nil
}
func decodeSelectedTableIDs(raw json.RawMessage, limits EffectiveLimits) ([]string, *semanticFailure) {
	scope := tableScope{Mode: "selected_tables"}
	if err := json.Unmarshal(raw, &scope.SelectedTableIDs); err != nil {
		return nil, invalidTableScope(scope, "empty_resolved_scope")
	}
	if failure := validateSelectedTableIDs(scope, limits); failure != nil {
		return nil, failure
	}
	return scope.SelectedTableIDs, nil
}
func requiredTableScope(raw json.RawMessage, limits EffectiveLimits) (tableScope, *semanticFailure) {
	if len(raw) == 0 {
		return tableScope{}, invalidNetworkFlowRequest("table_scope", "missing_member")
	}
	if bytes.Equal(raw, []byte("null")) {
		return tableScope{}, invalidNetworkFlowRequest("table_scope", "explicit_null")
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || object == nil {
		return tableScope{}, invalidTableScope(tableScope{}, "unknown_mode")
	}
	mode, failure := requiredJSONString(object, "mode")
	if failure != nil {
		return tableScope{}, invalidTableScope(tableScope{}, "unknown_mode")
	}
	scope := tableScope{Mode: mode}
	switch mode {
	case "active_table":
		_ = json.Unmarshal(object["active_table_id"], &scope.ActiveTableID)
		if failure := ensureAllowedMembers(object, "mode", "active_table_id"); failure != nil {
			return tableScope{}, invalidTableScope(scope, "variant_member_conflict")
		}
		if scope.ActiveTableID == "" {
			return tableScope{}, invalidTableScope(scope, "empty_resolved_scope")
		}
	case "selected_tables":
		_ = json.Unmarshal(object["selected_table_ids"], &scope.SelectedTableIDs)
		if failure := ensureAllowedMembers(object, "mode", "selected_table_ids"); failure != nil {
			return tableScope{}, invalidTableScope(scope, "variant_member_conflict")
		}
		ids, failure := decodeSelectedTableIDs(object["selected_table_ids"], limits)
		if failure != nil {
			return tableScope{}, failure
		}
		scope.SelectedTableIDs = ids
	case "all_active_tables":
		if failure := ensureAllowedMembers(object, "mode"); failure != nil {
			return tableScope{}, invalidTableScope(scope, "variant_member_conflict")
		}
	default:
		return tableScope{}, invalidTableScope(scope, "unknown_mode")
	}
	return scope, nil
}

// Selection preserves the store's workspace order. Row, graph and contributor
// paths share the same bounds and never silently remove an explicit selection.
func selectActiveScopeTables(active []tableRecord, scope tableScope, limits EffectiveLimits) ([]tableRecord, []string, *semanticFailure) {
	selected := []tableRecord{}
	wanted := map[string]struct{}{}
	switch scope.Mode {
	case "active_table":
		if scope.ActiveTableID == "" {
			return nil, nil, invalidTableScope(scope, "empty_resolved_scope")
		}
		wanted[scope.ActiveTableID] = struct{}{}
	case "selected_tables":
		if failure := validateSelectedTableIDs(scope, limits); failure != nil {
			return nil, nil, failure
		}
		for _, id := range scope.SelectedTableIDs {
			wanted[id] = struct{}{}
		}
	case "all_active_tables":
	default:
		return nil, nil, invalidTableScope(scope, "unknown_mode")
	}
	for _, table := range active {
		_, explicit := wanted[table.TableID]
		if explicit || scope.Mode == "all_active_tables" {
			selected = append(selected, table)
			delete(wanted, table.TableID)
		}
	}
	missing := []string{}
	for id := range wanted {
		missing = append(missing, id)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return selected, missing, nil
	}
	if len(selected) == 0 {
		return nil, nil, invalidTableScope(scope, "empty_resolved_scope")
	}
	if int64(len(selected)) > limits.MaxSelectedTablesPerQuery {
		failure := invalidTableScope(scope, "selected_table_limit_exceeded")
		failure.details.TableIDs = []string{}
		for _, table := range selected {
			failure.details.TableIDs = append(failure.details.TableIDs, table.TableID)
		}
		return nil, nil, failure
	}
	return selected, missing, nil
}
func resolveActiveTableScope(ctx context.Context, store tableScopeReader, limits EffectiveLimits, incidentID uuid.UUID, scope tableScope) ([]tableRecord, []string, map[string]int, *semanticFailure) {
	active, err := store.ListActiveTables(ctx, incidentID)
	if err != nil {
		return nil, nil, nil, internalSemanticFailure(err)
	}
	selected, missing, failure := selectActiveScopeTables(active, scope, limits)
	if failure != nil {
		return nil, nil, nil, failure
	}
	if len(missing) > 0 {
		if _, err := store.GetActiveTable(ctx, incidentID, missing[0]); err != nil {
			return nil, nil, nil, tableReadFailure(err, missing[0])
		}
		return nil, nil, nil, tableReadFailure(errTableNotFound, missing[0])
	}
	ids := make([]string, 0, len(selected))
	ranks := map[string]int{}
	for index, table := range selected {
		ids = append(ids, table.TableID)
		ranks[table.TableID] = index
	}
	return selected, ids, ranks, nil
}
