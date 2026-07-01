package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func parseAdministrativeAuditScope(rawQuery string) (listquery.Result, *APIError) {
	result, queryErr := listquery.Parse(rawQuery, listquery.Config{
		ExactFilters: map[string]listquery.ExactFilter{
			"actor_user_id": {},
			"action_code":   {},
			"target_kind":   {Allowed: []string{"user", "incident", "deployment"}},
			"target_id":     {},
		},
		RangeFilters: map[string]listquery.RangeFilter{
			"occurred_at_gte": {},
			"occurred_at_lt":  {},
		},
	})
	if queryErr != nil {
		if queryErr.Kind == listquery.ErrorKindPagination {
			return listquery.Result{}, userPaginationError(queryErr.ReasonCode)
		}
		return listquery.Result{}, userListQueryError(queryErr.ReasonCode)
	}
	if result.Scope["target_id"] != "" && result.Scope["target_kind"] == "" {
		return listquery.Result{}, userListQueryError(listquery.ReasonInvalidFilterValue)
	}
	for _, key := range []string{"occurred_at_gte", "occurred_at_lt"} {
		value := result.Scope[key]
		if value == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			return listquery.Result{}, userListQueryError(listquery.ReasonInvalidFilterRange)
		}
	}
	if result.Scope["occurred_at_gte"] != "" && result.Scope["occurred_at_lt"] != "" {
		gte, _ := time.Parse(time.RFC3339, result.Scope["occurred_at_gte"])
		lt, _ := time.Parse(time.RFC3339, result.Scope["occurred_at_lt"])
		if !gte.Before(lt) {
			return listquery.Result{}, userListQueryError(listquery.ReasonInvalidFilterRange)
		}
	}
	if actorUserID := result.Scope["actor_user_id"]; actorUserID != "" {
		if _, err := uuid.Parse(actorUserID); err != nil {
			return listquery.Result{}, userListQueryError(listquery.ReasonInvalidFilterValue)
		}
	}
	return result, nil
}

func administrativeAuditFilterFromScope(scope map[string]string) authn.AdministrativeAuditFilter {
	var filter authn.AdministrativeAuditFilter
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
		parsed, _ := time.Parse(time.RFC3339, value)
		parsed = parsed.UTC()
		filter.OccurredAtGTE = &parsed
	}
	if value := scope["occurred_at_lt"]; value != "" {
		parsed, _ := time.Parse(time.RFC3339, value)
		parsed = parsed.UTC()
		filter.OccurredAtLT = &parsed
	}
	return filter
}

func buildAdministrativeAuditResource(record authn.AdministrativeAuditRecord) map[string]any {
	targetKind := "deployment"
	var targetID any
	if record.IncidentID != nil {
		targetKind = "incident"
		targetID = record.IncidentID.String()
	} else if record.TargetUserID != nil {
		targetKind = "user"
		targetID = record.TargetUserID.String()
	}
	return map[string]any{
		"audit_event_id": record.ID,
		"scope_kind":     "deployment",
		"scope_id":       nil,
		"occurred_at":    record.OccurredAt,
		"actor_kind":     "user",
		"actor_user_id":  record.ActorUserID,
		"source":         record.EventSource,
		"action_code":    record.EventSource,
		"target_kind":    targetKind,
		"target_id":      targetID,
		"changes":        auditChanges(record.BeforeJSON, record.AfterJSON),
		"reason_code":    record.ReasonCode,
	}
}

func auditChanges(beforeRaw []byte, afterRaw []byte) []map[string]any {
	before := auditJSONObject(beforeRaw)
	after := auditJSONObject(afterRaw)
	keys := make(map[string]struct{}, len(before)+len(after))
	for key := range before {
		keys[key] = struct{}{}
	}
	for key := range after {
		keys[key] = struct{}{}
	}
	fieldPaths := make([]string, 0, len(keys))
	for key := range keys {
		fieldPaths = append(fieldPaths, key)
	}
	sort.Strings(fieldPaths)
	changes := make([]map[string]any, 0, len(fieldPaths))
	for _, fieldPath := range fieldPaths {
		changes = append(changes, map[string]any{
			"field_path":  fieldPath,
			"value_state": "visible",
			"before":      before[fieldPath],
			"after":       after[fieldPath],
		})
	}
	return changes
}

func auditJSONObject(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var object map[string]any
	if err := json.Unmarshal(raw, &object); err != nil {
		return map[string]any{}
	}
	return object
}

func (s *Service) handleAdministrativeAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	listScope, apiErr := parseAdministrativeAuditScope(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveListRequest(
		listScope.Values,
		"administrative_audit_events.list",
		principal.User.ID.String(),
		listScope.Scope,
	)
	if reasonCode != "" {
		writeAPIError(w, r, userPaginationError(reasonCode))
		return
	}
	records, err := s.store.ListAdministrativeAuditEvents(r.Context(), administrativeAuditFilterFromScope(binding.Scope))
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	resources := make([]map[string]any, 0, len(records))
	for _, record := range records {
		resources = append(resources, buildAdministrativeAuditResource(record))
	}
	rows, nextCursor, err := pagination.PageResources(binding, cursor, resources)
	switch {
	case errors.Is(err, pagination.ErrInvalidCursorToken):
		writeAPIError(w, r, userPaginationError(pagination.ReasonInvalidCursorToken))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	var nextCursorToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextCursorToken = &token
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"administrative_audit_events": rows}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextCursorToken != nil,
		NextCursor: nextCursorToken,
	})
}
