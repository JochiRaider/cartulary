package auth

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func beginPasswordWorkflow(w http.ResponseWriter, r *http.Request) *authn.PasswordWorkflow {
	workflow, err := authn.BeginPasswordWorkflow(r.Context())
	if errors.Is(err, authn.ErrAuthenticationCapacity) {
		w.Header().Set("Retry-After", "1")
		_ = httpapi.WriteErrorWithOptions(w, r, http.StatusServiceUnavailable, "authentication_capacity_exhausted", "authentication capacity exhausted", map[string]any{}, httpapi.ErrorOptions{Retryable: true})
		return nil
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return nil
	}
	return workflow
}

// Completed mutation replay precedes capacity admission and password work.
// The store still performs its transactional replay check for concurrent races.
func (s *Service) replayCredentialMutation(w http.ResponseWriter, r *http.Request, key authn.RouteIdempotencyKey, requestHash []byte) bool {
	existing, err := s.credentialStore.GetRouteIdempotency(r.Context(), key)
	if errors.Is(err, authn.ErrNotFound) {
		return false
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return true
	}
	if !hashesEqual(existing.RequestHash, requestHash) {
		writeAPIError(w, r, httpapi.ClientTxnConflictError(key.ClientTxnID))
		return true
	}
	payload, err := decodeStoredResponse(existing.ResponseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return true
	}
	// Existing mutation replay contracts use 200, including first-create 201.
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	return true
}
