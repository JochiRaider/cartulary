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

func (s *service) handleCloseIncident(w http.ResponseWriter, r *http.Request) {
	s.handleIncidentLifecycle(w, r, incidents.LifecycleActionClose)
}

func (s *service) handleReopenIncident(w http.ResponseWriter, r *http.Request) {
	s.handleIncidentLifecycle(w, r, incidents.LifecycleActionReopen)
}

func (s *service) handleIncidentLifecycle(
	w http.ResponseWriter,
	r *http.Request,
	action incidents.LifecycleAction,
) {
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
	request, apiErr := admitIncidentLifecycleJSON(action, r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.terminalMutations.CoordinateIncidentLifecycle(
		r.Context(),
		principal.User,
		incidentID,
		request,
		platformhttpapi.RequestIDFromContext(r.Context()),
	)
	var versionConflict *incidents.IncidentVersionConflictError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, platformhttpapi.ClientTxnConflictError(request.ClientTxnID()))
		return
	case admission.IsDenied(err, admission.DenialNotVisible):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		writeAPIError(w, r, authorizationDeniedError("admin"))
		return
	case errors.Is(err, incidents.ErrIncidentNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.As(err, &versionConflict):
		writeAPIError(w, r, incidentVersionConflictError(versionConflict))
		return
	case errors.Is(err, incidents.ErrIncidentIllegalTransition):
		writeAPIError(w, r, incidentIllegalTransitionError(action))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}
