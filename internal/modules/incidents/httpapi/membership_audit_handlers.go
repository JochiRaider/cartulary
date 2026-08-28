package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type membershipAuditReader struct {
	db postgres.DB
}

func newMembershipAuditReader(db postgres.DB) (*membershipAuditReader, error) {
	if isNilRouteDependency(db) {
		return nil, errors.New("incidents: PostgreSQL dependency is required for membership audit reading")
	}
	return &membershipAuditReader{db: db}, nil
}

func (r *membershipAuditReader) list(
	ctx context.Context,
	filter administrativeaudit.ListFilter,
) ([]administrativeaudit.Record, error) {
	return administrativeaudit.List(ctx, r.db, filter)
}

func parseMembershipAuditScope(rawQuery string) (listquery.Result, *platformhttpapi.APIError) {
	result, queryErr := administrativeaudit.ParseListScope(rawQuery, administrativeaudit.ScopeIncident)
	if queryErr == nil {
		return result, nil
	}
	if queryErr.Kind == listquery.ErrorKindPagination {
		return listquery.Result{}, invalidPaginationRequest(queryErr.ReasonCode)
	}
	return listquery.Result{}, invalidListQuery(queryErr.ReasonCode)
}

func (s *service) handleMembershipAuditEvents(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{
		Store:         s.authStore,
		Keys:          s.keys,
		Now:           s.now,
		StateChanging: false,
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	listScope, apiErr := parseMembershipAuditScope(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	listScope.Scope["incident_id"] = incidentID.String()
	binding, cursor, reasonCode := s.cursorCodec.ResolveListRequest(
		listScope.Values,
		"incident.membership_audit_events.list",
		principal.User.ID.String(),
		listScope.Scope,
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	pageRequest, reasonCode := administrativeaudit.PageRequest(
		binding,
		cursor,
		administrativeaudit.ScopeIncident,
		&incidentID,
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	records, err := s.membershipAudit.list(r.Context(), pageRequest)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	rows, nextCursor, err := administrativeaudit.BuildPage(binding, records)
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
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"audit_events": rows}, platformhttpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextCursorToken != nil,
		NextCursor: nextCursorToken,
	})
}
