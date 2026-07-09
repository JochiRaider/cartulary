package auth

import (
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *Service) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	request, apiErr := DecodeLoginRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	user, err := s.loginStore.GetUserByNormalizedEmail(r.Context(), request.Username)
	if err != nil || !user.IsActive {
		writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_credentials", Details: map[string]any{}})
		return
	}

	ok, err := authn.VerifyPasswordHash(user.PasswordHash, request.Password)
	if err != nil || !ok {
		writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_credentials", Details: map[string]any{}})
		return
	}

	if user.MFARequired {
		hasActiveTOTP := user.TOTPEnrolledAt != nil && len(user.TOTPSecretCiphertext) > 0 && len(user.TOTPSecretNonce) > 0
		if !hasActiveTOTP {
			bootstrapToken, bootstrapExpiresAt, err := s.issueBootstrapToken(r.Context(), user.ID)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			writeAPIError(w, r, &APIError{
				Status: http.StatusUnauthorized,
				Code:   "mfa_setup_required",
				Details: map[string]any{
					"required_setup_kinds": []string{"totp"},
					"bootstrap_token":      bootstrapToken,
					"bootstrap_expires_at": bootstrapExpiresAt,
				},
			})
			return
		}

		if request.SecondFactor == nil {
			writeAPIError(w, r, &APIError{
				Status: http.StatusUnauthorized,
				Code:   "mfa_required",
				Details: map[string]any{
					"required_second_factor_kinds": []string{"totp"},
				},
			})
			return
		}

		secretBytes, err := authn.DecryptSecret(s.keys, user.TOTPSecretCiphertext, user.TOTPSecretNonce)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		secretBase32 := strings.ToUpper(authn.EncodeSecretBase32(secretBytes))
		if !authn.ValidateTOTPCode(secretBase32, request.SecondFactor.Code, s.now()) {
			writeAPIError(w, r, &APIError{Status: http.StatusUnauthorized, Code: "invalid_second_factor", Details: map[string]any{}})
			return
		}
	}

	sessionToken, err := authn.GenerateOpaqueToken()
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	session, revoked, err := s.sessionStore.CreateSessionWithConcurrency(
		r.Context(),
		user,
		authn.FingerprintToken(s.keys, sessionToken),
		authn.NewSessionTiming(s.now()),
		httpapi.RequestIDFromContext(r.Context()),
	)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if revoked != nil {
		s.publishSessionRevocation(revoked.ID, authn.ConcurrencyLimitReasonCode)
	}

	s.setAuthCookies(w, sessionToken)
	resource, err := s.buildSessionResource(r.Context(), user, session)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleSession(w http.ResponseWriter, r *http.Request) {
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

	resource, err := s.buildSessionResource(r.Context(), principal.User, principal.Session)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	if err := s.sessionStore.RevokeSession(r.Context(), principal.Session.ID, sessionRevokedReasonCode, s.now()); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.publishSessionRevocation(principal.Session.ID, sessionRevokedReasonCode)

	s.clearAuthCookies(w)
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"user_id":          principal.User.ID,
		"sessions_revoked": false,
		"logged_out":       true,
	})
}
