package administrativeaudit

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

const (
	listOrdering     = "occurred_at_desc,audit_event_id_desc"
	listContinuation = "exclusive_after"
)

func ParseListScope(rawQuery string, scopeKind string) (listquery.Result, *listquery.Error) {
	result, queryErr := listquery.Parse(rawQuery, listquery.Config{
		ExactFilters: map[string]listquery.ExactFilter{
			"actor_user_id": {},
			"action_code": {
				Allowed: ActionCodes(scopeKind),
			},
			"target_kind": {
				Allowed: TargetKinds(scopeKind),
			},
			"target_id": {},
		},
		RangeFilters: map[string]listquery.RangeFilter{
			"occurred_at_gte": {},
			"occurred_at_lt":  {},
		},
	})
	if queryErr != nil {
		return listquery.Result{}, queryErr
	}
	if result.Scope["target_id"] != "" && result.Scope["target_kind"] == "" {
		return listquery.Result{}, listQueryError(listquery.ReasonInvalidFilterValue)
	}
	for _, key := range []string{"occurred_at_gte", "occurred_at_lt"} {
		value := result.Scope[key]
		if value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			return listquery.Result{}, listQueryError(listquery.ReasonInvalidFilterRange)
		}
		result.Scope[key] = parsed.UTC().Format(time.RFC3339Nano)
	}
	if result.Scope["occurred_at_gte"] != "" && result.Scope["occurred_at_lt"] != "" {
		gte, _ := time.Parse(time.RFC3339Nano, result.Scope["occurred_at_gte"])
		lt, _ := time.Parse(time.RFC3339Nano, result.Scope["occurred_at_lt"])
		if !gte.Before(lt) {
			return listquery.Result{}, listQueryError(listquery.ReasonInvalidFilterRange)
		}
	}
	if actorUserID := result.Scope["actor_user_id"]; actorUserID != "" {
		parsed, err := uuid.Parse(actorUserID)
		if err != nil {
			return listquery.Result{}, listQueryError(listquery.ReasonInvalidFilterValue)
		}
		result.Scope["actor_user_id"] = parsed.String()
	}
	return result, nil
}

func PageRequest(
	binding pagination.Binding,
	cursor *pagination.Cursor,
	scopeKind string,
	scopeID *uuid.UUID,
) (ListFilter, string) {
	filter := filterFromScope(binding.Scope, scopeKind, scopeID)
	filter.Limit = binding.Limit + 1
	if cursor == nil {
		return filter, ""
	}
	if cursor.Mode != pagination.ModeKeyset {
		return ListFilter{}, pagination.ReasonInvalidCursorToken
	}
	if len(cursor.Position) != 4 ||
		cursor.Position["ordering"] != listOrdering ||
		cursor.Position["continuation"] != listContinuation {
		return ListFilter{}, pagination.ReasonInvalidCursorToken
	}
	occurredAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_occurred_at"])
	if err != nil {
		return ListFilter{}, pagination.ReasonInvalidCursorToken
	}
	auditEventID, err := uuid.Parse(cursor.Position["last_audit_event_id"])
	if err != nil {
		return ListFilter{}, pagination.ReasonInvalidCursorToken
	}
	filter.After = &ListPosition{
		OccurredAt:   occurredAt.UTC(),
		AuditEventID: auditEventID,
	}
	return filter, ""
}

func BuildPage(
	binding pagination.Binding,
	records []Record,
) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	pageRecords := records
	if hasMore {
		pageRecords = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(pageRecords))
	for _, record := range pageRecords {
		resources = append(resources, BuildResource(record))
	}
	rows, err := pagination.MarshalResources(resources)
	if err != nil {
		return nil, nil, err
	}
	if !hasMore || len(pageRecords) == 0 {
		return rows, nil, nil
	}
	last := pageRecords[len(pageRecords)-1]
	return rows, &pagination.Cursor{
		Version:     pagination.CursorVersion,
		Mode:        pagination.ModeKeyset,
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       binding.Scope,
		Position: map[string]string{
			"ordering":            listOrdering,
			"continuation":        listContinuation,
			"last_occurred_at":    last.OccurredAt.UTC().Format(time.RFC3339Nano),
			"last_audit_event_id": last.AuditEventID.String(),
		},
	}, nil
}

func BuildResource(record Record) map[string]any {
	return map[string]any{
		"audit_event_id": record.AuditEventID,
		"scope_kind":     record.ScopeKind,
		"scope_id":       record.ScopeID,
		"occurred_at":    record.OccurredAt.UTC(),
		"actor_kind":     record.ActorKind,
		"actor_user_id":  record.ActorUserID,
		"source":         record.Source,
		"action_code":    record.ActionCode,
		"target_kind":    record.TargetKind,
		"target_id":      record.TargetID,
		"changes":        record.Changes,
		"reason_code":    record.ReasonCode,
	}
}

func filterFromScope(scope map[string]string, scopeKind string, scopeID *uuid.UUID) ListFilter {
	filter := ListFilter{
		ScopeKind:          scopeKind,
		ScopeID:            scopeID,
		AllowedActionCodes: ActionCodes(scopeKind),
	}
	if value := scope["actor_user_id"]; value != "" {
		parsed := uuid.MustParse(value)
		filter.ActorUserID = &parsed
	}
	if value := scope["action_code"]; value != "" {
		filter.ActionCode = &value
	}
	if value := scope["target_kind"]; value != "" {
		filter.TargetKind = &value
	}
	if value := scope["target_id"]; value != "" {
		filter.TargetID = &value
	}
	if value := scope["occurred_at_gte"]; value != "" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		filter.OccurredAtGTE = &parsed
	}
	if value := scope["occurred_at_lt"]; value != "" {
		parsed, _ := time.Parse(time.RFC3339Nano, value)
		filter.OccurredAtLT = &parsed
	}
	return filter
}

func listQueryError(reasonCode string) *listquery.Error {
	return &listquery.Error{
		Kind:       listquery.ErrorKindList,
		ReasonCode: reasonCode,
	}
}
