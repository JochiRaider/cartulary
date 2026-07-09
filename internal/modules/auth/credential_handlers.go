package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *Service) handleCredentialState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if apiErr := ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	pending, err := s.credentialStore.GetPendingTOTPEnrollmentForUser(r.Context(), principal.User.ID, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var pendingExpiresAt *time.Time
	if pending != nil {
		pendingExpiresAt = &pending.ExpiresAt
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildCredentialStateResource(principal.User, pendingExpiresAt))
}

func (s *Service) handlePasswordChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	request, apiErr := DecodePasswordChangeRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	requestHash := hashRequestPayload(map[string]any{
		"current_password": requestSecretFingerprint(s.keys, request.CurrentPassword),
		"new_password":     requestSecretFingerprint(s.keys, request.NewPassword),
		"second_factor":    requestSecondFactorHashPayload(s.keys, request.SecondFactor),
	})
	key := authn.ActorOnlyRouteIdempotencyKey("auth.password.change", principal.User.ID, request.ClientTxnID)
	if existing, err := s.credentialStore.GetRouteIdempotency(r.Context(), key); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		s.clearAuthCookies(w)
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
		return
	} else if !errors.Is(err, authn.ErrNotFound) {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	ok, err := authn.VerifyPasswordHash(principal.User.PasswordHash, request.CurrentPassword)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !ok {
		writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_current_password", Details: map[string]any{}})
		return
	}
	if apiErr := s.validateActiveTOTP(principal.User, request.SecondFactor); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	newPasswordHash, err := authn.HashPassword(request.NewPassword)
	if err != nil {
		writeAPIError(w, r, invalidAuthRequest("new_password", err.Error()))
		return
	}

	result, err := s.credentialStore.ChangePassword(
		r.Context(),
		principal.User,
		request.ClientTxnID,
		requestHash,
		newPasswordHash,
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	s.publishSessionRevocations(result.RevokedSessionIDs, sessionRevokedReasonCode)

	payload, err := decodeStoredResponse(result.ResponseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.clearAuthCookies(w)
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
}

func (s *Service) handleTOTPBegin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authContext, apiErr := s.authenticateAuthRequest(r, true, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	request, apiErr := DecodeTOTPBeginRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	replacesActive := authContext.User.TOTPEnrolledAt != nil && len(authContext.User.TOTPSecretCiphertext) > 0 && len(authContext.User.TOTPSecretNonce) > 0
	if authContext.Principal != nil && replacesActive {
		if request.CurrentPassword == nil {
			writeAPIError(w, r, invalidAuthRequest("current_password", "current_password is required for totp replacement"))
			return
		}
		ok, err := authn.VerifyPasswordHash(authContext.User.PasswordHash, *request.CurrentPassword)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if !ok {
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_current_password", Details: map[string]any{}})
			return
		}
		if apiErr := s.validateActiveTOTP(authContext.User, request.SecondFactor); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
	}

	clearSecret, secretBase32, err := authn.GenerateTOTPSecret()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	ciphertext, nonce, err := authn.EncryptSecret(s.keys, clearSecret)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var sessionID *uuid.UUID
	var bootstrapTokenID *uuid.UUID
	authScopeKind := "bootstrap_token"
	if authContext.Principal != nil {
		sessionID = &authContext.Principal.Session.ID
		authScopeKind = "session"
	} else if authContext.BootstrapToken != nil {
		bootstrapTokenID = &authContext.BootstrapToken.ID
	}

	pending, replayed, err := s.credentialStore.BeginTOTPEnrollment(
		r.Context(),
		authContext.User.ID,
		authScopeKind,
		sessionID,
		bootstrapTokenID,
		request.ClientTxnID,
		ciphertext,
		nonce,
		replacesActive,
		s.now(),
	)
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if replayed {
		clearSecret, err = authn.DecryptSecret(s.keys, pending.SecretCiphertext, pending.SecretNonce)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		secretBase32 = authn.EncodeSecretBase32(clearSecret)
	}

	if authContext.Principal != nil {
		if err := s.slideSessionIfNeeded(r.Context(), authContext.Principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"enrollment_id": pending.ID,
		"expires_at":    pending.ExpiresAt,
		"totp_setup":    BuildTOTPSetup(secretBase32, authContext.User.Email),
	})
}

func (s *Service) handleTOTPComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authContext, apiErr := s.authenticateAuthRequest(r, true, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	request, apiErr := DecodeTOTPCompleteRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	enrollmentID, err := uuid.Parse(request.EnrollmentID)
	if err != nil {
		writeAPIError(w, r, invalidAuthRequest("enrollment_id", "enrollment_id must be a uuid"))
		return
	}

	pending, err := s.credentialStore.GetPendingTOTPEnrollmentByID(r.Context(), enrollmentID)
	if errors.Is(err, authn.ErrNotFound) {
		writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if pending.UserID != authContext.User.ID {
		if authContext.BootstrapToken != nil {
			writeAPIError(w, r, BootstrapRejectedError("subject_mismatch"))
		} else {
			writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		}
		return
	}
	if status := authn.PendingEnrollmentStatusAt(pending, s.now()); status == authn.PendingEnrollmentExpired {
		writeAPIError(w, r, TOTPSetupNotPendingError("expired"))
		return
	} else if status == authn.PendingEnrollmentConsumed {
		writeAPIError(w, r, TOTPSetupNotPendingError("consumed"))
		return
	}

	clearSecret, err := authn.DecryptSecret(s.keys, pending.SecretCiphertext, pending.SecretNonce)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !authn.ValidateTOTPCode(authn.EncodeSecretBase32(clearSecret), request.Code, s.now()) {
		writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}})
		return
	}

	var sessionID *uuid.UUID
	var bootstrapTokenID *uuid.UUID
	authScopeKind := "bootstrap_token"
	if authContext.Principal != nil {
		sessionID = &authContext.Principal.Session.ID
		authScopeKind = "session"
	} else if authContext.BootstrapToken != nil {
		bootstrapTokenID = &authContext.BootstrapToken.ID
	}

	result, err := s.credentialStore.ActivateTOTPEnrollment(
		r.Context(),
		authContext.User,
		enrollmentID,
		authScopeKind,
		sessionID,
		bootstrapTokenID,
		s.now(),
	)
	switch {
	case errors.Is(err, authn.ErrNotFound):
		writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		return
	case errors.Is(err, authn.ErrPendingExpired):
		writeAPIError(w, r, TOTPSetupNotPendingError("expired"))
		return
	case errors.Is(err, authn.ErrPendingConsumed):
		writeAPIError(w, r, TOTPSetupNotPendingError("consumed"))
		return
	case errors.Is(err, authn.ErrSubjectMismatch):
		if authContext.BootstrapToken != nil {
			writeAPIError(w, r, BootstrapRejectedError("subject_mismatch"))
		} else {
			writeAPIError(w, r, TOTPSetupNotPendingError("not_found"))
		}
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	s.publishSessionRevocations(result.RevokedSessionIDs, sessionRevokedReasonCode)
	if result.SessionsRevoked {
		s.clearAuthCookies(w)
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"user_id":          authContext.User.ID,
		"totp":             map[string]any{"enrolled_at": result.EnrolledAt},
		"sessions_revoked": result.SessionsRevoked,
	})
}

func (s *Service) validateActiveTOTP(user authn.UserRecord, secondFactor *SecondFactorAssertion) *APIError {
	if user.TOTPEnrolledAt == nil || len(user.TOTPSecretCiphertext) == 0 || len(user.TOTPSecretNonce) == 0 {
		return nil
	}
	if secondFactor == nil {
		return &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}}
	}
	secretBytes, err := authn.DecryptSecret(s.keys, user.TOTPSecretCiphertext, user.TOTPSecretNonce)
	if err != nil {
		return internalAPIError(err)
	}
	if !authn.ValidateTOTPCode(authn.EncodeSecretBase32(secretBytes), secondFactor.Code, s.now()) {
		return &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}}
	}
	return nil
}
