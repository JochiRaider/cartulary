package httpapi

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

func (s *service) handleListIncidentMemberships(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := routeUUID(w, r, "incident_id")
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	query, apiErr := parseMembershipListQuery(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveListRequest(query, "incident.memberships.list", principal.User.ID.String(), map[string]string{"incident_id": incidentID.String()})
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}

	page, reasonCode := membershipListPageRequest(binding, cursor)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	records, listErr := s.application.ListMemberships(r.Context(), incidentID, page)
	if listErr != nil {
		writeAPIError(w, r, internalAPIError(listErr))
		return
	}
	rows, nextCursor, err := buildMembershipListPage(binding, records)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var nextToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextToken = &token
	}
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"memberships": rows}, platformhttpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *service) handleCreateIncidentMembership(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := routeUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := admitMembershipCreateJSON(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.application.CreateMembership(r.Context(), principal.User, incidentID, request, platformhttpapi.RequestIDFromContext(r.Context()))
	switch {
	case errors.Is(err, authn.ErrNotFound):
		writeAPIError(w, r, userNotFoundError())
		return
	case errors.Is(err, incidents.ErrMembershipTargetInactive):
		writeAPIError(w, r, userInactiveError())
		return
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, platformhttpapi.ClientTxnConflictError(request.ClientTxnID()))
		return
	case admission.IsDenied(err, admission.DenialNotVisible):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		writeAPIError(w, r, authorizationDeniedError("admin"))
		return
	case errors.Is(err, incidents.ErrMembershipExistsUsePatch):
		writeAPIError(w, r, membershipExistsUsePatchError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	_ = platformhttpapi.WriteSuccess(w, r, status, result.Payload)
}

func (s *service) handlePatchIncidentMembership(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := routeUUID(w, r, "incident_id")
	if !ok {
		return
	}
	userID, ok := routeUUID(w, r, "user_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := admitMembershipPatchJSON(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, _, err := s.application.UpdateMembership(r.Context(), principal.User, incidentID, userID, request, platformhttpapi.RequestIDFromContext(r.Context()))
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		writeAPIError(w, r, authorizationDeniedError("admin"))
		return
	case errors.Is(err, incidents.ErrMembershipNotFound):
		writeAPIError(w, r, membershipNotFoundError())
		return
	case errors.Is(err, incidents.ErrMembershipVersionConflict):
		writeAPIError(w, r, membershipVersionConflictError())
		return
	case errors.Is(err, incidents.ErrLastIncidentAdmin):
		writeAPIError(w, r, lastIncidentAdminError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, incidents.BuildMembershipResource(record))
}

func (s *service) handleDeleteIncidentMembership(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := routeUUID(w, r, "incident_id")
	if !ok {
		return
	}
	userID, ok := routeUUID(w, r, "user_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := admitMembershipDeleteJSON(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, err := s.terminalMutations.CoordinateMembershipDeletion(r.Context(), principal.User, incidentID, userID, request, platformhttpapi.RequestIDFromContext(r.Context())); err != nil {
		switch {
		case admission.IsDenied(err, admission.DenialNotVisible):
			writeAPIError(w, r, incidentNotFoundError())
			return
		case admission.IsDenied(err, admission.DenialInsufficientRole):
			writeAPIError(w, r, authorizationDeniedError("admin"))
			return
		case errors.Is(err, incidents.ErrMembershipNotFound):
			writeAPIError(w, r, membershipNotFoundError())
			return
		case errors.Is(err, incidents.ErrMembershipVersionConflict):
			writeAPIError(w, r, membershipVersionConflictError())
			return
		case errors.Is(err, incidents.ErrLastIncidentAdmin):
			writeAPIError(w, r, lastIncidentAdminError())
			return
		default:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
