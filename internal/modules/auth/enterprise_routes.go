package auth

import (
	"context"
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

// The enterprise operations are bound only when the enterprise_authentication
// profile is claimed. Otherwise the unclaimed reserved-family gate answers their
// paths with extension_profile_not_claimed before route resolution.

func (s *Service) handleListEnterpriseAuthProviders(w http.ResponseWriter, r *http.Request) {
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
}

// admitEnterpriseProvider classifies the addressed provider before any body,
// RelayState, completion or browser-binding check, so a well-formed missing
// provider_key receives the malformed key's 404 auth_provider_not_found (Core
// 01 REQ-01-707 step 3). A provider of another type is not found for the
// route; a disabled one, or for begin a non-interactive one, is
// auth_provider_disabled.
func (s *Service) admitEnterpriseProvider(ctx context.Context, providerKey, providerType string, interactive bool) (authn.EnterpriseAuthProviderRecord, *httpapi.APIError) {
	provider, err := s.enterpriseStore.GetEnterpriseAuthProviderByKey(ctx, providerKey)
	switch {
	case errors.Is(err, authn.ErrAuthProviderNotFound) || err == nil && providerType != "" && provider.ProviderType != providerType:
		return authn.EnterpriseAuthProviderRecord{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_not_found", Details: map[string]any{}}
	case err != nil:
		return authn.EnterpriseAuthProviderRecord{}, internalAPIError(err)
	case !provider.IsEnabled || interactive && !provider.IsInteractive:
		return authn.EnterpriseAuthProviderRecord{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "auth_provider_disabled", Details: map[string]any{}}
	}
	return provider, nil
}

func (s *Service) handleBeginEnterpriseAuth(w http.ResponseWriter, r *http.Request) {
	providerKey, ok := admittedProviderKey(w, r)
	if !ok {
		return
	}
	provider, apiErr := s.admitEnterpriseProvider(r.Context(), providerKey, "", true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeEnterpriseAuthBeginRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	transactionID := uuid.New()
	var pkceEnvelope []byte
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
		pkceEnvelope, err = authn.SealSecret(s.keys, authn.SecretBinding{Purpose: authn.EnterprisePKCESecret, RecordID: transactionID, SubjectID: provider.ID}, []byte(verifier))
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

	transaction, err := s.enterpriseStore.CreateEnterpriseAuthTransaction(r.Context(), transactionID, provider, request.ReturnTo, state, nonce, pkceHash, pkceEnvelope, relayState, samlRequestID, browserHash, s.now())
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
	providerKey, ok := admittedProviderKey(w, r)
	if !ok {
		return
	}
	provider, apiErr := s.admitEnterpriseProvider(r.Context(), providerKey, "oidc", false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	clear, err := authn.OpenSecret(s.keys, authn.SecretBinding{Purpose: authn.EnterprisePKCESecret, RecordID: transaction.ID, SubjectID: transaction.ProviderID}, transaction.PKCEVerifierEnvelope)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	pkceVerifier := string(clear)
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
	sessionToken, err := authn.GenerateOpaqueToken()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.enterpriseStore.CompleteOIDCEnterpriseAuthTransaction(r.Context(), providerKey, verified.State, browserHash, &verified.Nonce, verified.ProviderSubject, authn.EnterpriseAuthSessionInput{TokenFingerprint: authn.FingerprintToken(s.keys, sessionToken), RequestID: httpapi.RequestIDFromContext(r.Context())}, s.now())
	if err != nil {
		writeAPIError(w, r, s.enterpriseCompletionError(err))
		return
	}
	s.finishEnterpriseLogin(w, r, result, sessionToken)
}

func (s *Service) handleEnterpriseSAMLACS(w http.ResponseWriter, r *http.Request) {
	providerKey, ok := admittedProviderKey(w, r)
	if !ok {
		return
	}
	provider, apiErr := s.admitEnterpriseProvider(r.Context(), providerKey, "saml", false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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

func (s *Service) handleEnterpriseSAMLComplete(w http.ResponseWriter, r *http.Request) {
	providerKey, ok := admittedProviderKey(w, r)
	if !ok {
		return
	}
	// The store rechecks the provider inside its transaction.
	if _, apiErr := s.admitEnterpriseProvider(r.Context(), providerKey, "saml", false); apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	sessionToken, err := authn.GenerateOpaqueToken()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.enterpriseStore.CompleteSAMLEnterpriseAuthTransaction(r.Context(), providerKey, authn.FingerprintToken(s.keys, completionToken), browserHash, authn.EnterpriseAuthSessionInput{TokenFingerprint: authn.FingerprintToken(s.keys, sessionToken), RequestID: httpapi.RequestIDFromContext(r.Context())}, s.now())
	if err != nil {
		writeAPIError(w, r, s.enterpriseCompletionError(err))
		return
	}
	s.finishEnterpriseLogin(w, r, result, sessionToken)
}

func (s *Service) handleCreateEnterpriseAuthBinding(w http.ResponseWriter, r *http.Request) {
	principal, userID, ok := admittedUserTarget(w, r)
	if !ok {
		return
	}
	if apiErr := s.requireTargetUser(r.Context(), userID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeEnterpriseBindingCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"base_user_version": request.BaseUserVersion,
		"provider_key":      request.ProviderKey,
		"provider_subject":  request.ProviderSubject,
		"reason":            request.Reason,
	})
	result, err := s.enterpriseStore.CreateEnterpriseAuthBinding(r.Context(), principal.User, userID, request.BaseUserVersion, request.ClientTxnID, request.ProviderKey, request.ProviderSubject, request.Reason, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	s.writeEnterpriseBindingResult(w, r, principal.Session.ID, request.ClientTxnID, result, err)
}

func (s *Service) handleRotateEnterpriseAuthBinding(w http.ResponseWriter, r *http.Request) {
	principal, userID, authBindingID, ok := admittedAuthBindingTarget(w, r)
	if !ok {
		return
	}
	request, apiErr := DecodeEnterpriseBindingRotateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"base_user_version":    request.BaseUserVersion,
		"new_provider_subject": request.NewProviderSubject,
		"reason":               request.Reason,
	})
	result, err := s.enterpriseStore.RotateEnterpriseAuthBinding(r.Context(), principal.User, userID, authBindingID, request.BaseUserVersion, request.ClientTxnID, request.NewProviderSubject, request.Reason, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	s.writeEnterpriseBindingResult(w, r, principal.Session.ID, request.ClientTxnID, result, bindingTargetError(err))
}

func (s *Service) handleRetireEnterpriseAuthBinding(w http.ResponseWriter, r *http.Request) {
	principal, userID, authBindingID, ok := admittedAuthBindingTarget(w, r)
	if !ok {
		return
	}
	request, apiErr := DecodeEnterpriseBindingRetireRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"base_user_version": request.BaseUserVersion,
		"reason":            request.Reason,
	})
	result, err := s.enterpriseStore.RetireEnterpriseAuthBinding(r.Context(), principal.User, userID, authBindingID, request.BaseUserVersion, request.ClientTxnID, request.Reason, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	s.writeEnterpriseBindingResult(w, r, principal.Session.ID, request.ClientTxnID, result, bindingTargetError(err))
}

// admittedAuthBindingTarget returns the deployment administrator and the
// addressed {user_id, auth_binding_id} pair of a binding rotate or retire.
func admittedAuthBindingTarget(w http.ResponseWriter, r *http.Request) (httpauth.Principal, uuid.UUID, uuid.UUID, bool) {
	principal, userID, ok := admittedUserTarget(w, r)
	if !ok {
		return httpauth.Principal{}, uuid.UUID{}, uuid.UUID{}, false
	}
	authBindingID, ok := admittedPathUUID(w, r, "auth_binding_id")
	if !ok {
		return httpauth.Principal{}, uuid.UUID{}, uuid.UUID{}, false
	}
	return principal, userID, authBindingID, true
}

// admittedProviderKey returns the centrally admitted provider_key of a public
// enterprise sign-in operation. Its absence is a binding error.
func admittedProviderKey(w http.ResponseWriter, r *http.Request) (string, bool) {
	providerKey, ok := httpapi.PathString(r, "provider_key")
	if !ok {
		writeAPIError(w, r, internalAPIError(errors.New("enterprise operation has no admitted provider_key")))
		return "", false
	}
	return providerKey, true
}

// bindingTargetError maps an absent target user on rotate or retire to the
// absent binding target: a {user_id, auth_binding_id} pair whose user does not
// exist identifies no current enterprise binding, which Core 01 REQ-01-539,
// REQ-01-540 and REQ-01-541 answer with 404 auth_binding_not_found.
func bindingTargetError(err error) error {
	if errors.Is(err, authn.ErrNotFound) {
		return authn.ErrAuthBindingNotFound
	}
	return err
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

func (s *Service) finishEnterpriseLogin(w http.ResponseWriter, r *http.Request, result authn.EnterpriseAuthCompletionResult, sessionToken string) {
	if result.RevokedSession != nil {
		s.publishSessionRevocation(result.RevokedSession.ID, authn.ConcurrencyLimitReasonCode)
	}
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
