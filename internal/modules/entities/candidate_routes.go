package entities

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/entities/candidates"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

func (s *service) handleCandidates(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, err := parsePathUUID(r, "incident_id")
	if err != nil {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if _, apiErr = s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesMember, ""); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	parsed, failure := listquery.Parse(r.URL.RawQuery, listquery.Config{Search: true, ExactFilters: map[string]listquery.ExactFilter{"entity_type": {Allowed: []string{"host", "identity"}}}})
	if failure != nil {
		writeAPIError(w, r, candidateQueryError(failure.Kind, failure.ReasonCode))
		return
	}
	if parsed.Scope["entity_type"] == "" {
		writeAPIError(w, r, candidateQueryError(listquery.ErrorKindList, listquery.ReasonInvalidFilterValue))
		return
	}
	parsed.Scope["incident_id"] = incidentID.String()
	parsed.Scope["ordering"] = "display_name:asc,record_id:asc"
	binding, cursor, reason := s.cursorCodec.ResolveListRequest(parsed.Values, "entities.candidates", principal.User.ID.String(), parsed.Scope)
	if reason != "" {
		writeAPIError(w, r, candidateQueryError(listquery.ErrorKindPagination, reason))
		return
	}
	var after *candidates.Position
	if cursor != nil {
		name, ok := cursor.Position["display_name"]
		id, parseErr := uuid.Parse(cursor.Position["record_id"])
		if cursor.Mode != pagination.ModeKeyset || !ok || parseErr != nil || len(cursor.Position) != 2 {
			writeAPIError(w, r, candidateQueryError(listquery.ErrorKindPagination, pagination.ReasonInvalidCursorToken))
			return
		}
		after = &candidates.Position{DisplayName: name, RecordID: id}
	}
	page, err := s.candidateReader.Page(r.Context(), candidates.Request{IncidentID: incidentID, EntityType: parsed.Scope["entity_type"], Search: parsed.Scope["search"], Limit: binding.Limit, After: after})
	if err != nil {
		candidateReadFailure(w, r)
		return
	}
	// Recheck after a potentially long sparse scan, before releasing observations.
	if _, apiErr = s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesMember, ""); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	var next *string
	if page.Next != nil {
		token, err := s.cursorCodec.Encode(pagination.Cursor{Mode: pagination.ModeKeyset, Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: binding.Limit, Scope: binding.Scope, Position: map[string]string{"display_name": page.Next.DisplayName, "record_id": page.Next.RecordID.String()}})
		if err != nil {
			candidateReadFailure(w, r)
			return
		}
		next = &token
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		candidateReadFailure(w, r)
		return
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"incident_id": incidentID, "entity_type": parsed.Scope["entity_type"], "candidates": page.Candidates}, httpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next})
}
func candidateQueryError(kind, reason string) *httpapi.APIError {
	code := "invalid_list_query"
	if kind == listquery.ErrorKindPagination {
		code = "invalid_pagination_request"
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: code, Message: "Invalid candidate query", Details: map[string]any{"reason_code": reason}}
}
func candidateReadFailure(w http.ResponseWriter, r *http.Request) {
	writeAPIError(w, r, &httpapi.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "Candidate read failed", Details: map[string]any{}})
}
