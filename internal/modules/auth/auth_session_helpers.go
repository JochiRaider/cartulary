package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

// handleTouch is the harness-only session touch route. It is not a public
// operation, so it authenticates itself and always slides idle expiry. Its
// registration admits only GET (TestRouteGuard.ProtectMethod).
func (s *Service) handleTouch(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.sessionStore, Keys: s.keys, Hub: s.revocations, Now: s.now})
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

// admittedPrincipal returns the session principal central admission
// established (Core 01 REQ-01-707 step 2). Its absence is a binding error.
func admittedPrincipal(w http.ResponseWriter, r *http.Request) (httpauth.Principal, bool) {
	principal, ok := httpauth.AdmittedPrincipal(r)
	if !ok {
		writeAPIError(w, r, internalAPIError(errors.New("auth operation has no admitted principal")))
		return httpauth.Principal{}, false
	}
	return principal, true
}

// admittedDeploymentAdmin returns the admitted principal after the
// deployment-administration gate.
func admittedDeploymentAdmin(w http.ResponseWriter, r *http.Request) (httpauth.Principal, bool) {
	principal, ok := admittedPrincipal(w, r)
	if !ok {
		return httpauth.Principal{}, false
	}
	if apiErr := httpauth.RequireDeploymentAdmin(principal.User); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, false
	}
	return principal, true
}

// admittedPathUUID returns a centrally admitted uuid path parameter. Its
// absence is a binding error.
func admittedPathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	value, ok := httpapi.PathUUID(r, name)
	if !ok {
		writeAPIError(w, r, internalAPIError(errors.New("auth operation has no admitted path parameter "+name)))
		return uuid.UUID{}, false
	}
	return value, true
}

// admittedUserTarget returns the deployment administrator and the addressed
// user_id of a /api/v1/users/{user_id} operation. The identifier's syntax was
// admitted before the gate (REQ-01-707 step 3); its existence is decided after
// it (REQ-01-118).
func admittedUserTarget(w http.ResponseWriter, r *http.Request) (httpauth.Principal, uuid.UUID, bool) {
	principal, ok := admittedDeploymentAdmin(w, r)
	if !ok {
		return httpauth.Principal{}, uuid.UUID{}, false
	}
	userID, ok := admittedPathUUID(w, r, "user_id")
	if !ok {
		return httpauth.Principal{}, uuid.UUID{}, false
	}
	return principal, userID, true
}

// totpCredential is the credential a TOTP setup operation was admitted with:
// exactly one of the current session or a bootstrap token (Core 01 Table
// 3.3.2.2-D).
type totpCredential struct {
	session   *httpauth.Principal
	bootstrap *httpauth.BootstrapCredential
	user      authn.UserRecord
}

func admittedTOTPCredential(w http.ResponseWriter, r *http.Request) (totpCredential, bool) {
	if principal, ok := httpauth.AdmittedPrincipal(r); ok {
		return totpCredential{session: &principal, user: principal.User}, true
	}
	if bootstrap, ok := httpauth.AdmittedBootstrapCredential(r); ok {
		return totpCredential{bootstrap: &bootstrap, user: bootstrap.User}, true
	}
	writeAPIError(w, r, internalAPIError(errors.New("TOTP operation has no admitted credential")))
	return totpCredential{}, false
}

func (s *Service) issueBootstrapToken(ctx context.Context, user authn.UserRecord) (string, time.Time, error) {
	token, err := authn.GenerateOpaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	record, err := s.sessionStore.IssueBootstrapToken(ctx, user, authn.FingerprintToken(s.keys, token), s.now())
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
	memberships, err := s.sessionMembershipReader.ListIncidentMembershipSummaries(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return assembleSessionResource(user, session, memberships), nil
}

func assembleSessionResource(user authn.UserRecord, session authn.SessionRecord, memberships []authn.IncidentMembershipSummary) map[string]any {
	mfaState := "not_required"
	if user.MFARequired {
		mfaState = "satisfied"
	}
	providerType := session.ProviderType
	if providerType == "" {
		providerType = "local"
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
	}
}

func (s *Service) buildSafeUserResource(ctx context.Context, user authn.UserRecord) (map[string]any, error) {
	bindings, err := s.userAdminStore.ListEnterpriseAuthBindingSummaries(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return BuildSafeUserResourceWithEnterpriseBindings(user, bindings), nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.sessionStore, principal, s.now)
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

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	httpapi.WriteAPIError(w, r, apiErr)
}

func internalAPIError(err error) *httpapi.APIError {
	return httpapi.InternalAPIError(err)
}
