package workbook

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

// admittedIncidentRoute returns the principal and incident_id central
// admission established (Core 01 REQ-01-707 steps 2 and 3).
func admittedIncidentRoute(r *http.Request) (httpauth.Principal, uuid.UUID, *httpapi.APIError) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		return httpauth.Principal{}, uuid.UUID{}, apiErr
	}
	incidentID, apiErr := admittedPathUUID(r, "incident_id")
	if apiErr != nil {
		return httpauth.Principal{}, uuid.UUID{}, apiErr
	}
	return principal, incidentID, nil
}

func (s *service) handleGetDefaultWorkbookPreferences(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the singleton-query check is route handling, after
	// the incident's visibility (REQ-01-704).
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.startupStore.GetDefaultPreferences(r.Context(), incidentID)
	if errors.Is(err, workbookstartup.ErrPreferencesNotFound) {
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildDefaultPreferencesResource(record))
}

func (s *service) handlePutDefaultWorkbookPreferences(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildDefaultPreferencesResource(record))
}

func (s *service) handleGetUserWorkbookPreferences(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the singleton-query check is route handling, after
	// the incident's visibility (REQ-01-704).
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.startupStore.GetUserPreferences(r.Context(), incidentID, principal.User.ID)
	if errors.Is(err, workbookstartup.ErrPreferencesNotFound) {
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildUserPreferencesResource(record))
}

func (s *service) handlePutUserWorkbookPreferences(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildUserPreferencesResource(record))
}

func (s *service) handleWorkbookStartup(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildStartupResource(record))
}
