package auth

import (
	"encoding/json"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func parseAdministrativeAuditScope(rawQuery string) (listquery.Result, *httpapi.APIError) {
	result, queryErr := administrativeaudit.ParseListScope(rawQuery, administrativeaudit.ScopeDeployment)
	if queryErr != nil {
		if queryErr.Kind == listquery.ErrorKindPagination {
			return listquery.Result{}, userPaginationError(queryErr.ReasonCode)
		}
		return listquery.Result{}, userListQueryError(queryErr.ReasonCode)
	}
	return result, nil
}

func administrativeAuditPageRequest(
	binding pagination.Binding,
	cursor *pagination.Cursor,
) (administrativeaudit.ListFilter, string) {
	return administrativeaudit.PageRequest(
		binding,
		cursor,
		administrativeaudit.ScopeDeployment,
		nil,
	)
}

func buildAdministrativeAuditPage(
	binding pagination.Binding,
	records []administrativeaudit.Record,
) ([]json.RawMessage, *pagination.Cursor, error) {
	return administrativeaudit.BuildPage(binding, records)
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
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
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
	pageRequest, reasonCode := administrativeAuditPageRequest(binding, cursor)
	if reasonCode != "" {
		writeAPIError(w, r, userPaginationError(reasonCode))
		return
	}
	records, err := s.deploymentAuditReader.ListAdministrativeAuditEvents(r.Context(), pageRequest)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	rows, nextCursor, err := buildAdministrativeAuditPage(binding, records)
	if err != nil {
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
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"audit_events": rows}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextCursorToken != nil,
		NextCursor: nextCursorToken,
	})
}
