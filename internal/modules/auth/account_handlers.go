package auth

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *Service) handleGetAccountProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := admittedPrincipal(w, r)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildAccountProfileResource(principal.User))
}

func (s *Service) handlePatchAccountProfile(w http.ResponseWriter, r *http.Request) {
	principal, ok := admittedPrincipal(w, r)
	if !ok {
		return
	}
	request, apiErr := DecodeAccountProfilePatchRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"base_user_version": request.BaseUserVersion,
		"client_txn_id":     request.ClientTxnID,
		"display_name":      request.DisplayName,
	})
	result, err := s.accountStore.PatchAccountProfile(r.Context(), principal.User, request.BaseUserVersion, request.DisplayName, request.ClientTxnID, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, authn.ErrUserVersionConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	payload, err := decodeStoredResponse(result.ResponseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, payload)
}

func (s *Service) handleGetAccountPreferences(w http.ResponseWriter, r *http.Request) {
	principal, ok := admittedPrincipal(w, r)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.accountStore.GetOrCreateAccountPreferences(r.Context(), principal.User.ID, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildAccountPreferencesResource(record))
}

func (s *Service) handlePutAccountPreferences(w http.ResponseWriter, r *http.Request) {
	principal, ok := admittedPrincipal(w, r)
	if !ok {
		return
	}
	request, apiErr := DecodeAccountPreferencesPutRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"base_preferences_version": request.BasePreferencesVersion,
		"client_txn_id":            request.ClientTxnID,
		"density_mode":             request.DensityMode,
	})
	result, err := s.accountStore.PutAccountPreferences(r.Context(), principal.User, request.BasePreferencesVersion, request.ClientTxnID, request.DensityMode, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, authn.ErrPreferencesVersionConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "preferences_version_conflict", Details: map[string]any{}})
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	payload, err := decodeStoredResponse(result.ResponseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, payload)
}
