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
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

func (s *Service) handleTouch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}.Slide(s.now())
	persisted, err := s.sessionStore.SlideSession(r.Context(), principal.Session.ID, sliding)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	resource, err := s.buildSessionResource(r.Context(), principal.User, principal.Session)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (SessionPrincipal, *APIError) {
	context, apiErr := s.authenticateAuthRequest(r, stateChanging, false)
	if apiErr != nil {
		return SessionPrincipal{}, apiErr
	}
	if context.Principal == nil {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	return *context.Principal, nil
}

func (s *Service) authenticateAuthRequest(r *http.Request, stateChanging bool, allowBootstrap bool) (CredentialAuthContext, *APIError) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
		}
		if principal, apiErr := s.authenticateSessionToken(r, AuthSourceBearer, token, false); apiErr == nil {
			return CredentialAuthContext{Principal: &principal, User: principal.User}, nil
		} else if apiErr.Code != unauthorizedCode {
			return CredentialAuthContext{}, apiErr
		}

		bootstrap, matched, apiErr := s.bootstrapCredentialAuthenticator().Authenticate(r, token, allowBootstrap)
		if apiErr != nil {
			return CredentialAuthContext{}, apiErr
		}
		if matched {
			return CredentialAuthContext{
				BootstrapToken:     &bootstrap.Token,
				BootstrapTokenText: bootstrap.TokenText,
				User:               bootstrap.User,
			}, nil
		}
		return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if errors.Is(err, authn.ErrNotFound) {
		return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return CredentialAuthContext{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
		}
		return CredentialAuthContext{}, internalAPIError(err)
	}
	principal, apiErr := s.authenticateSessionToken(r, AuthSourceCookie, cookie.Value, stateChanging)
	if apiErr != nil {
		return CredentialAuthContext{}, apiErr
	}
	return CredentialAuthContext{Principal: &principal, User: principal.User}, nil
}

func (s *Service) bootstrapCredentialAuthenticator() bootstrapCredentialAuthenticator {
	return bootstrapCredentialAuthenticator{
		store: s.sessionStore,
		keys:  s.keys,
		now:   s.now,
	}
}

func (s *Service) authenticateSessionToken(r *http.Request, authSource AuthSource, sessionToken string, stateChanging bool) (SessionPrincipal, *APIError) {
	if stateChanging && authSource == AuthSourceCookie {
		csrfCookie, _ := r.Cookie(authn.CSRFCookieName)
		if csrfCookie == nil || csrfCookie.Value != authn.CSRFTokenForSessionToken(s.keys, sessionToken) {
			return SessionPrincipal{}, &APIError{Status: http.StatusForbidden, Code: "csrf_verification_failed", Details: map[string]any{}}
		}
		if apiErr := ValidateCSRF(r.Method, authSource, csrfCookie.Value, r.Header.Get(authn.CSRFHeaderName)); apiErr != nil {
			return SessionPrincipal{}, apiErr
		}
	}

	session, user, err := s.sessionStore.GetSessionByFingerprint(r.Context(), authn.FingerprintToken(s.keys, sessionToken))
	if errors.Is(err, authn.ErrNotFound) {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	if err != nil {
		return SessionPrincipal{}, internalAPIError(err)
	}
	if !user.IsActive {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	if session.RevokedAt != nil {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	now := s.now()
	if !session.SessionExpiresAt.After(now) {
		_ = s.sessionStore.RevokeSession(context.Background(), session.ID, sessionExpiredReasonCode, now)
		s.publishSessionRevocation(session.ID, sessionExpiredReasonCode)
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	return SessionPrincipal{
		AuthSource:   authSource,
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	}, nil
}

func (s *Service) issueBootstrapToken(ctx context.Context, userID uuid.UUID) (string, time.Time, error) {
	token, err := authn.GenerateOpaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	record, err := s.sessionStore.IssueBootstrapToken(ctx, userID, authn.FingerprintToken(s.keys, token), s.now())
	if err != nil {
		return "", time.Time{}, err
	}
	return token, record.ExpiresAt, nil
}

func (s *Service) setAuthCookies(w http.ResponseWriter, sessionToken string) {
	csrfToken := authn.CSRFTokenForSessionToken(s.keys, sessionToken)
	http.SetCookie(w, &http.Cookie{
		Name:     authn.SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- CSRF double-submit cookie must stay browser-readable; session cookie is HttpOnly and both cookies remain Secure/SameSite/Path.
		Name:     authn.CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	})
}

func (s *Service) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     authn.SessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{ // #nosec G124 -- CSRF double-submit cookie must stay browser-readable; session cookie is HttpOnly and both cookies remain Secure/SameSite/Path.
		Name:     authn.CSRFCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		MaxAge:   -1,
	})
}

func (s *Service) buildSessionResource(ctx context.Context, user authn.UserRecord, session authn.SessionRecord) (map[string]any, error) {
	mfaState := "not_required"
	if user.MFARequired {
		mfaState = "satisfied"
	}
	providerType := session.ProviderType
	if providerType == "" {
		providerType = "local"
	}

	memberships, err := s.sessionMembershipReader.ListIncidentMembershipSummaries(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	summaries := make([]map[string]any, 0, len(memberships))
	for _, membership := range memberships {
		summaries = append(summaries, map[string]any{
			"incident_id": membership.IncidentID,
			"role":        membership.Role,
		})
	}

	return map[string]any{
		"user_id":             user.ID,
		"display_name":        user.DisplayName,
		"provider_type":       providerType,
		"mfa_state":           mfaState,
		"is_deployment_admin": user.IsDeploymentAdmin,
		"authenticated_at":    session.AuthenticatedAt,
		"idle_expires_at":     session.IdleExpiresAt,
		"absolute_expires_at": session.AbsoluteExpiresAt,
		"session_expires_at":  session.SessionExpiresAt,
		"memberships":         summaries,
	}, nil
}

func (s *Service) buildSafeUserResource(ctx context.Context, user authn.UserRecord) (map[string]any, error) {
	bindings, err := s.userAdminStore.ListEnterpriseAuthBindingSummaries(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return BuildSafeUserResourceWithEnterpriseBindings(user, bindings), nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *SessionPrincipal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.sessionStore, principal, method, path, s.now)
}

func (s *Service) publishSessionRevocations(sessionIDs []uuid.UUID, reasonCode string) {
	for _, sessionID := range sessionIDs {
		s.publishSessionRevocation(sessionID, reasonCode)
	}
}

func (s *Service) publishSessionRevocation(sessionID uuid.UUID, reasonCode string) {
	if s.revocations == nil {
		return
	}
	s.revocations.RevokeSession(sessionID, reasonCode)
}

func requestSecretFingerprint(keys authn.MasterKeys, value string) []byte {
	return authn.FingerprintRequestValue(keys, value)
}

func requestSecondFactorHashPayload(keys authn.MasterKeys, secondFactor *SecondFactorAssertion) any {
	if secondFactor == nil {
		return nil
	}
	return map[string]any{
		"kind": secondFactor.Kind,
		"code": requestSecretFingerprint(keys, secondFactor.Code),
	}
}

func hashRequestPayload(payload any) []byte {
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func hashesEqual(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *APIError) {
	httpapi.WriteAPIError(w, r, apiErr)
}

func internalAPIError(err error) *APIError {
	return httpapi.InternalAPIError(err)
}
