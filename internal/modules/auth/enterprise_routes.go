package auth

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

const enterpriseAuthCookieName = "cartulary_enterprise_auth_txn"

type EnterpriseAuthBeginRequest struct {
	ReturnTo string
}

type EnterpriseBindingCreateRequest struct {
	BaseUserVersion int64
	ClientTxnID     string
	ProviderKey     string
	ProviderSubject string
	Reason          *string
}

type EnterpriseBindingRotateRequest struct {
	BaseUserVersion    int64
	ClientTxnID        string
	NewProviderSubject string
	Reason             *string
}

type EnterpriseBindingRetireRequest struct {
	BaseUserVersion int64
	ClientTxnID     string
	Reason          *string
}

func (s *Service) handleEnterpriseProviders(w http.ResponseWriter, r *http.Request) {
	if apiErr := s.requireEnterpriseProfileClaimed(r.URL.Path); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	if r.URL.Path == "/api/v1/auth/providers" {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		providers, err := s.enterpriseStore.ListEnterpriseAuthProviders(r.Context())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		rows := make([]map[string]any, 0, len(providers))
		for _, provider := range providers {
			rows = append(rows, map[string]any{
				"provider_key":  provider.ProviderKey,
				"provider_type": provider.ProviderType,
				"display_name":  provider.DisplayName,
			})
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"providers": rows})
		return
	}

	providerKey, ok := providerBeginKey(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	request, apiErr := DecodeEnterpriseAuthBeginRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, err := s.enterpriseStore.GetEnterpriseAuthProviderByKey(r.Context(), providerKey)
	switch {
	case errors.Is(err, authn.ErrAuthProviderNotFound):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_not_found", Details: map[string]any{}})
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !provider.IsEnabled || !provider.IsInteractive {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_disabled", Details: map[string]any{}})
		return
	}

	browserToken, err := authn.GenerateOpaqueToken()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	browserHash := authn.FingerprintToken(s.keys, browserToken)

	var state, nonce, relayState *string
	var pkceHash []byte
	var pkceCiphertext []byte
	var pkceNonce []byte
	var pkceVerifier string
	var samlRequestID *string
	if provider.ProviderType == "oidc" {
		stateValue, err := authn.GenerateOpaqueToken()
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nonceValue, err := authn.GenerateOpaqueToken()
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		verifier, err := authn.GenerateOpaqueToken()
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		sum := sha256.Sum256([]byte(verifier))
		pkceHash = sum[:]
		pkceCiphertext, pkceNonce, err = authn.EncryptSecret(s.keys, []byte(verifier))
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		pkceVerifier = verifier
		state = &stateValue
		nonce = &nonceValue
	} else {
		relayValue, err := authn.GenerateOpaqueToken()
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		relayState = &relayValue
		requestIDToken, err := authn.GenerateOpaqueToken()
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		requestID := "_" + requestIDToken
		samlRequestID = &requestID
	}

	transaction, err := s.enterpriseStore.CreateEnterpriseAuthTransaction(r.Context(), provider, request.ReturnTo, state, nonce, pkceHash, pkceCiphertext, pkceNonce, relayState, samlRequestID, browserHash, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	redirect, err := s.beginRedirect(provider, s.publicOrigin, transaction, pkceVerifier)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.setEnterpriseAuthCookie(w, browserToken, transaction.ExpiresAt)
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"provider_key":  provider.ProviderKey,
		"provider_type": provider.ProviderType,
		"redirect_url":  redirect.URL,
		"expires_at":    transaction.ExpiresAt,
	})
}

func (s *Service) handleEnterpriseOIDC(w http.ResponseWriter, r *http.Request) {
	if apiErr := s.requireEnterpriseProfileClaimed(r.URL.Path); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	providerKey, ok := providerCallbackKey(r.URL.Path, "/api/v1/auth/oidc/", "/callback")
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	provider, err := s.enterpriseStore.GetEnterpriseAuthProviderByKey(r.Context(), providerKey)
	if errors.Is(err, authn.ErrAuthProviderNotFound) || (err == nil && provider.ProviderType != "oidc") {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !provider.IsEnabled {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_disabled", Details: map[string]any{}})
		return
	}
	browserHash, apiErr := s.enterpriseBrowserBindingHash(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" {
		writeAPIError(w, r, providerResponseRejected("missing_required_field"))
		return
	}
	transaction, err := s.enterpriseStore.GetOIDCEnterpriseAuthTransactionForCallback(r.Context(), providerKey, state, browserHash, s.now())
	if err != nil {
		writeAPIError(w, r, s.enterpriseCompletionError(err))
		return
	}
	pkceVerifier := ""
	if len(transaction.PKCEVerifierCiphertext) > 0 || len(transaction.PKCEVerifierNonce) > 0 {
		clear, err := authn.DecryptSecret(s.keys, transaction.PKCEVerifierCiphertext, transaction.PKCEVerifierNonce)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		pkceVerifier = string(clear)
	}
	verified, apiErr := s.oidcVerifier.VerifyCallback(r.Context(), enterpriseauth.OIDCCallbackVerificationRequest{
		Provider:     provider,
		Transaction:  transaction,
		Values:       r.URL.Query(),
		PKCEVerifier: pkceVerifier,
		PublicOrigin: s.publicOrigin,
		Env:          s.env,
		Now:          s.now(),
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.enterpriseStore.CompleteOIDCEnterpriseAuthTransaction(r.Context(), providerKey, verified.State, browserHash, &verified.Nonce, verified.ProviderSubject, s.now())
	if err != nil {
		writeAPIError(w, r, s.enterpriseCompletionError(err))
		return
	}
	s.finishEnterpriseLogin(w, r, result)
}

func (s *Service) handleEnterpriseSAML(w http.ResponseWriter, r *http.Request) {
	if apiErr := s.requireEnterpriseProfileClaimed(r.URL.Path); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if providerKey, ok := providerCallbackKey(r.URL.Path, "/api/v1/auth/saml/", "/acs/complete"); ok {
		s.handleEnterpriseSAMLComplete(w, r, providerKey)
		return
	}
	providerKey, ok := providerCallbackKey(r.URL.Path, "/api/v1/auth/saml/", "/acs")
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		writeAPIError(w, r, providerResponseRejected("missing_required_field"))
		return
	}
	relayState := strings.TrimSpace(r.Form.Get("RelayState"))
	if relayState == "" {
		writeAPIError(w, r, enterpriseTransactionRejected("not_found"))
		return
	}
	provider, err := s.enterpriseStore.GetEnterpriseAuthProviderByKey(r.Context(), providerKey)
	if errors.Is(err, authn.ErrAuthProviderNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if provider.ProviderType != "saml" {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_not_found", Details: map[string]any{}})
		return
	}
	transaction, err := s.enterpriseStore.GetSAMLEnterpriseAuthTransactionForACS(r.Context(), providerKey, relayState, s.now())
	if err != nil {
		if errors.Is(err, authn.ErrEnterpriseTransactionNotFound) {
			writeAPIError(w, r, providerResponseRejected("relay_state_mismatch"))
			return
		}
		writeAPIError(w, r, s.enterpriseCompletionError(err))
		return
	}
	verified, apiErr := s.samlVerifier.VerifyACS(r.Context(), enterpriseauth.SAMLACSVerificationRequest{
		Provider:     provider,
		Transaction:  transaction,
		Values:       r.Form,
		PublicOrigin: s.publicOrigin,
		Now:          s.now(),
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	completionToken, err := authn.GenerateOpaqueToken()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	completionHash := authn.FingerprintToken(s.keys, completionToken)
	if _, err := s.enterpriseStore.StageSAMLEnterpriseAuthTransaction(r.Context(), providerKey, relayState, verified.ProviderSubject, completionHash, s.now()); err != nil {
		if errors.Is(err, authn.ErrEnterpriseTransactionNotFound) {
			writeAPIError(w, r, providerResponseRejected("relay_state_mismatch"))
			return
		}
		writeAPIError(w, r, s.enterpriseCompletionError(err))
		return
	}
	http.Redirect(w, r, enterpriseSAMLCompletionURL(providerKey, completionToken), http.StatusSeeOther)
}

func (s *Service) handleEnterpriseSAMLComplete(w http.ResponseWriter, r *http.Request, providerKey string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	completionToken := strings.TrimSpace(r.URL.Query().Get("completion"))
	if completionToken == "" {
		writeAPIError(w, r, enterpriseTransactionRejected("completion_mismatch"))
		return
	}
	browserHash, apiErr := s.enterpriseBrowserBindingHash(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.enterpriseStore.CompleteSAMLEnterpriseAuthTransaction(r.Context(), providerKey, authn.FingerprintToken(s.keys, completionToken), browserHash, s.now())
	if err != nil {
		writeAPIError(w, r, s.enterpriseCompletionError(err))
		return
	}
	s.finishEnterpriseLogin(w, r, result)
}

func (s *Service) handleEnterpriseAuthBindings(w http.ResponseWriter, r *http.Request, userID uuid.UUID, segments []string) bool {
	if len(segments) < 2 || segments[1] != "auth-bindings" {
		return false
	}
	if apiErr := s.requireEnterpriseProfileClaimed(r.URL.Path); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return true
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return true
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return true
	}

	switch {
	case len(segments) == 2 && r.Method == http.MethodPost:
		request, apiErr := DecodeEnterpriseBindingCreateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return true
		}
		requestHash := hashRequestPayload(map[string]any{
			"base_user_version": request.BaseUserVersion,
			"provider_key":      request.ProviderKey,
			"provider_subject":  request.ProviderSubject,
			"reason":            request.Reason,
		})
		result, err := s.enterpriseStore.CreateEnterpriseAuthBinding(r.Context(), principal.User, userID, request.BaseUserVersion, request.ClientTxnID, request.ProviderKey, request.ProviderSubject, request.Reason, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
		s.writeEnterpriseBindingResult(w, r, principal.Session.ID, request.ClientTxnID, result, err)
		return true
	case len(segments) == 4 && segments[3] == "rotate" && r.Method == http.MethodPost:
		authBindingID, err := uuid.Parse(segments[2])
		if err != nil {
			http.NotFound(w, r)
			return true
		}
		request, apiErr := DecodeEnterpriseBindingRotateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return true
		}
		requestHash := hashRequestPayload(map[string]any{
			"base_user_version":    request.BaseUserVersion,
			"new_provider_subject": request.NewProviderSubject,
			"reason":               request.Reason,
		})
		result, err := s.enterpriseStore.RotateEnterpriseAuthBinding(r.Context(), principal.User, userID, authBindingID, request.BaseUserVersion, request.ClientTxnID, request.NewProviderSubject, request.Reason, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
		s.writeEnterpriseBindingResult(w, r, principal.Session.ID, request.ClientTxnID, result, err)
		return true
	case len(segments) == 3 && r.Method == http.MethodDelete:
		authBindingID, err := uuid.Parse(segments[2])
		if err != nil {
			http.NotFound(w, r)
			return true
		}
		request, apiErr := DecodeEnterpriseBindingRetireRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return true
		}
		requestHash := hashRequestPayload(map[string]any{
			"base_user_version": request.BaseUserVersion,
			"reason":            request.Reason,
		})
		result, err := s.enterpriseStore.RetireEnterpriseAuthBinding(r.Context(), principal.User, userID, authBindingID, request.BaseUserVersion, request.ClientTxnID, request.Reason, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
		s.writeEnterpriseBindingResult(w, r, principal.Session.ID, request.ClientTxnID, result, err)
		return true
	default:
		if len(segments) == 2 || len(segments) == 3 || len(segments) == 4 {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return true
		}
		http.NotFound(w, r)
		return true
	}
}

func (s *Service) writeEnterpriseBindingResult(w http.ResponseWriter, r *http.Request, currentSessionID uuid.UUID, clientTxnID string, result authn.EnterpriseAuthBindingResult, err error) {
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(clientTxnID))
		return
	case errors.Is(err, authn.ErrUserVersionConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
		return
	case errors.Is(err, authn.ErrAuthProviderNotFound):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_not_found", Details: map[string]any{}})
		return
	case errors.Is(err, authn.ErrAuthBindingNotFound):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_binding_not_found", Details: map[string]any{}})
		return
	case errors.Is(err, authn.ErrAuthBindingNotActive):
		writeAPIError(w, r, authBindingConflict("binding_not_active"))
		return
	case errors.Is(err, authn.ErrAuthBindingProviderSubjectInUse):
		writeAPIError(w, r, authBindingConflict("provider_subject_in_use"))
		return
	case errors.Is(err, authn.ErrAuthBindingProviderAlreadyLinkedForUser):
		writeAPIError(w, r, authBindingConflict("provider_already_linked_for_user"))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	for _, sessionID := range result.RevokedSessionIDs {
		s.publishSessionRevocation(sessionID, sessionRevokedReasonCode)
		if sessionID == currentSessionID {
			s.clearAuthCookies(w)
		}
	}
	payload, err := decodeStoredResponse(result.ResponseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	statusCode := result.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	_ = httpapi.WriteSuccess(w, r, statusCode, payload)
}

func (s *Service) finishEnterpriseLogin(w http.ResponseWriter, r *http.Request, result authn.EnterpriseAuthCompletionResult) {
	sessionToken, err := authn.GenerateOpaqueToken()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	bindingID := result.Binding.ID
	session, revoked, err := s.sessionStore.CreateSessionWithProviderConcurrency(
		r.Context(),
		result.User,
		authn.FingerprintToken(s.keys, sessionToken),
		authn.NewSessionTiming(s.now()),
		httpapi.RequestIDFromContext(r.Context()),
		result.Binding.ProviderType,
		&bindingID,
	)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if revoked != nil {
		s.publishSessionRevocation(revoked.ID, authn.ConcurrencyLimitReasonCode)
	}
	_ = session
	s.setAuthCookies(w, sessionToken)
	s.clearEnterpriseAuthCookie(w)
	http.Redirect(w, r, result.ReturnTo, http.StatusSeeOther)
}

func (s *Service) enterpriseBrowserBindingHash(r *http.Request) ([]byte, *httpapi.APIError) {
	cookie, err := r.Cookie(enterpriseAuthCookieName)
	if err != nil || cookie.Value == "" {
		return nil, enterpriseTransactionRejected("browser_binding_mismatch")
	}
	return authn.FingerprintToken(s.keys, cookie.Value), nil
}

func (s *Service) enterpriseCompletionError(err error) *httpapi.APIError {
	switch {
	case errors.Is(err, authn.ErrAuthProviderNotFound):
		return &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_not_found", Details: map[string]any{}}
	case errors.Is(err, authn.ErrAuthProviderDisabled):
		return &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_disabled", Details: map[string]any{}}
	case errors.Is(err, authn.ErrEnterpriseTransactionNotFound):
		return enterpriseTransactionRejected("not_found")
	case errors.Is(err, authn.ErrEnterpriseTransactionExpired):
		return enterpriseTransactionRejected("expired")
	case errors.Is(err, authn.ErrEnterpriseTransactionUsed):
		return enterpriseTransactionRejected("already_used")
	case errors.Is(err, authn.ErrEnterpriseTransactionProviderMismatch):
		return enterpriseTransactionRejected("provider_mismatch")
	case errors.Is(err, authn.ErrEnterpriseTransactionStateMismatch):
		return providerResponseRejected("state_mismatch")
	case errors.Is(err, authn.ErrEnterpriseTransactionCompletionMismatch):
		return enterpriseTransactionRejected("completion_mismatch")
	case errors.Is(err, authn.ErrEnterpriseTransactionBrowserMismatch):
		return enterpriseTransactionRejected("browser_binding_mismatch")
	case errors.Is(err, authn.ErrSubjectMismatch):
		return providerResponseRejected("nonce_mismatch")
	case errors.Is(err, authn.ErrEnterpriseIdentityNoLinkedUser):
		return providerIdentityRejected("no_linked_user")
	case errors.Is(err, authn.ErrEnterpriseIdentityInactiveUser):
		return providerIdentityRejected("inactive_user")
	default:
		return internalAPIError(err)
	}
}

func (s *Service) requireEnterpriseProfileClaimed(path string) *httpapi.APIError {
	if s.enterpriseClaimed {
		return nil
	}
	return &httpapi.APIError{
		Status: http.StatusNotFound,
		Code:   "extension_profile_not_claimed",
		Details: map[string]any{
			"profile_id":   "enterprise_authentication",
			"route_family": enterpriseRouteFamily(path),
		},
	}
}

func enterpriseRouteFamily(path string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/auth/providers"):
		return "/api/v1/auth/providers"
	case strings.HasPrefix(path, "/api/v1/auth/oidc"):
		return "/api/v1/auth/oidc"
	case strings.HasPrefix(path, "/api/v1/auth/saml"):
		return "/api/v1/auth/saml"
	default:
		return "/api/v1/users/{user_id}/auth-bindings"
	}
}

func DecodeEnterpriseAuthBeginRequest(reader any) (EnterpriseAuthBeginRequest, *httpapi.APIError) {
	raw, apiErr := decodeEnterpriseObject(reader)
	if apiErr != nil {
		return EnterpriseAuthBeginRequest{}, apiErr
	}
	request := EnterpriseAuthBeginRequest{ReturnTo: "/"}
	for key, value := range raw {
		switch key {
		case "return_to":
			if string(value) == "null" {
				continue
			}
			var returnTo string
			if err := json.Unmarshal(value, &returnTo); err != nil {
				return EnterpriseAuthBeginRequest{}, invalidEnterpriseAuthRequest("return_to", "field_not_nullable")
			}
			normalized, ok := normalizeEnterpriseReturnTo(returnTo)
			if !ok {
				return EnterpriseAuthBeginRequest{}, invalidEnterpriseAuthRequest("return_to", "return_to_not_allowed")
			}
			request.ReturnTo = normalized
		default:
			return EnterpriseAuthBeginRequest{}, invalidEnterpriseAuthRequest(key, "unknown_field")
		}
	}
	return request, nil
}

func DecodeEnterpriseBindingCreateRequest(reader any) (EnterpriseBindingCreateRequest, *httpapi.APIError) {
	raw, apiErr := decodeMutationObject(reader)
	if apiErr != nil {
		return EnterpriseBindingCreateRequest{}, apiErr
	}
	request := EnterpriseBindingCreateRequest{}
	for key := range raw {
		switch key {
		case "base_user_version", "client_txn_id", "provider_key", "provider_subject", "reason":
		default:
			return EnterpriseBindingCreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	if apiErr := requireInt64(raw, "base_user_version", &request.BaseUserVersion); apiErr != nil {
		return EnterpriseBindingCreateRequest{}, apiErr
	}
	if apiErr := requireNonEmptyString(raw, "client_txn_id", &request.ClientTxnID); apiErr != nil {
		return EnterpriseBindingCreateRequest{}, apiErr
	}
	if apiErr := requireNonEmptyString(raw, "provider_key", &request.ProviderKey); apiErr != nil {
		return EnterpriseBindingCreateRequest{}, apiErr
	}
	if apiErr := requireNonEmptyString(raw, "provider_subject", &request.ProviderSubject); apiErr != nil {
		return EnterpriseBindingCreateRequest{}, apiErr
	}
	reason, apiErr := decodeOptionalBindingReason(raw["reason"])
	if apiErr != nil {
		return EnterpriseBindingCreateRequest{}, apiErr
	}
	request.Reason = reason
	return request, nil
}

func DecodeEnterpriseBindingRotateRequest(reader any) (EnterpriseBindingRotateRequest, *httpapi.APIError) {
	raw, apiErr := decodeMutationObject(reader)
	if apiErr != nil {
		return EnterpriseBindingRotateRequest{}, apiErr
	}
	request := EnterpriseBindingRotateRequest{}
	for key := range raw {
		switch key {
		case "base_user_version", "client_txn_id", "new_provider_subject", "reason":
		default:
			return EnterpriseBindingRotateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	if apiErr := requireInt64(raw, "base_user_version", &request.BaseUserVersion); apiErr != nil {
		return EnterpriseBindingRotateRequest{}, apiErr
	}
	if apiErr := requireNonEmptyString(raw, "client_txn_id", &request.ClientTxnID); apiErr != nil {
		return EnterpriseBindingRotateRequest{}, apiErr
	}
	if apiErr := requireNonEmptyString(raw, "new_provider_subject", &request.NewProviderSubject); apiErr != nil {
		return EnterpriseBindingRotateRequest{}, apiErr
	}
	reason, apiErr := decodeOptionalBindingReason(raw["reason"])
	if apiErr != nil {
		return EnterpriseBindingRotateRequest{}, apiErr
	}
	request.Reason = reason
	return request, nil
}

func DecodeEnterpriseBindingRetireRequest(reader any) (EnterpriseBindingRetireRequest, *httpapi.APIError) {
	raw, apiErr := decodeMutationObject(reader)
	if apiErr != nil {
		return EnterpriseBindingRetireRequest{}, apiErr
	}
	request := EnterpriseBindingRetireRequest{}
	for key := range raw {
		switch key {
		case "base_user_version", "client_txn_id", "reason":
		default:
			return EnterpriseBindingRetireRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	if apiErr := requireInt64(raw, "base_user_version", &request.BaseUserVersion); apiErr != nil {
		return EnterpriseBindingRetireRequest{}, apiErr
	}
	if apiErr := requireNonEmptyString(raw, "client_txn_id", &request.ClientTxnID); apiErr != nil {
		return EnterpriseBindingRetireRequest{}, apiErr
	}
	reason, apiErr := decodeOptionalBindingReason(raw["reason"])
	if apiErr != nil {
		return EnterpriseBindingRetireRequest{}, apiErr
	}
	request.Reason = reason
	return request, nil
}

func (s *Service) setEnterpriseAuthCookie(w http.ResponseWriter, value string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     enterpriseAuthCookieName,
		Value:    value,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Expires:  expiresAt,
	})
}

func (s *Service) clearEnterpriseAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     enterpriseAuthCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		MaxAge:   -1,
	})
}
