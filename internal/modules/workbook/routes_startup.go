package workbook

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

func (s *service) handleWorkbookPreferencesDefault(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
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
		record, err := s.startupStore.GetDefaultPreferences(r.Context(), incidentID)
		if errors.Is(err, workbookstartup.ErrPreferencesNotFound) {
			writeAPIError(w, r, incidentNotFoundError())
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildDefaultPreferencesResource(record))

	case http.MethodPut:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		membership, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin")
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := workbookstartup.DecodeDefaultPreferencesPutRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := s.startupStore.ValidatePreferenceSheetRef(request.DefaultSheetRef, membership.Role.String(), "default_sheet_ref"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.startupStore.PutDefaultPreferences(r.Context(), incidentID, principal.User.ID, request.DefaultSheetRef, s.now())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildDefaultPreferencesResource(record))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *service) handleWorkbookPreferencesMe(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
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
		record, err := s.startupStore.GetUserPreferences(r.Context(), incidentID, principal.User.ID)
		if errors.Is(err, workbookstartup.ErrPreferencesNotFound) {
			writeAPIError(w, r, incidentNotFoundError())
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildUserPreferencesResource(record))

	case http.MethodPut:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := workbookstartup.DecodeUserPreferencesPutRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := s.startupStore.ValidatePreferenceSheetRef(request.HomeSheetRef, membership.Role.String(), "home_sheet_ref"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.startupStore.PutUserPreferences(r.Context(), incidentID, principal.User.ID, request.HomeSheetRef, s.now())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildUserPreferencesResource(record))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *service) handleWorkbookStartup(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
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
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	explicitSheetRef, apiErr := workbookstartup.ParseExplicitSheetRef(r.URL.Query())
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := s.startupStore.ValidateExplicitSheetRef(explicitSheetRef, membership.Role.String()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.startupStore.Resolve(r.Context(), incidentID, principal.User.ID, membership.Role.String(), explicitSheetRef, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildStartupResource(record))
}
