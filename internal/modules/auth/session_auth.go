package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type SessionAuthStore interface {
	GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error)
	GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error)
	RevokeSession(context.Context, uuid.UUID, string, time.Time) error
}

type SessionRevoker interface {
	RevokeSession(uuid.UUID, string)
}

type SessionAuthOptions struct {
	Store         SessionAuthStore
	Keys          authn.MasterKeys
	Hub           SessionRevoker
	Now           func() time.Time
	StateChanging bool
}

func AuthenticateSessionRequest(r *http.Request, options SessionAuthOptions) (SessionPrincipal, *APIError) {
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
		}
		if principal, apiErr := authenticateSessionToken(r, options.Store, options.Keys, options.Hub, now, AuthSourceBearer, token, options.StateChanging); apiErr == nil {
			return principal, nil
		} else if apiErr.Code != unauthorizedCode {
			return SessionPrincipal{}, apiErr
		}

		bootstrapToken, _, err := options.Store.GetBootstrapTokenByFingerprint(r.Context(), authn.FingerprintToken(options.Keys, token))
		if err == nil {
			if reason := authn.BootstrapReasonCode(bootstrapToken, now()); reason != "" {
				return SessionPrincipal{}, BootstrapRejectedError(reason)
			}
			return SessionPrincipal{}, BootstrapRejectedError("not_allowed_for_route")
		}
		if err != nil && !errors.Is(err, authn.ErrNotFound) {
			return SessionPrincipal{}, internalAPIError(err)
		}
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
		}
		return SessionPrincipal{}, internalAPIError(err)
	}
	return authenticateSessionToken(r, options.Store, options.Keys, options.Hub, now, AuthSourceCookie, cookie.Value, options.StateChanging)
}

func authenticateSessionToken(
	r *http.Request,
	store SessionAuthStore,
	keys authn.MasterKeys,
	hub SessionRevoker,
	now func() time.Time,
	authSource AuthSource,
	sessionToken string,
	stateChanging bool,
) (SessionPrincipal, *APIError) {
	if stateChanging && authSource == AuthSourceCookie {
		csrfCookie, _ := r.Cookie(authn.CSRFCookieName)
		if csrfCookie == nil || csrfCookie.Value != authn.CSRFTokenForSessionToken(keys, sessionToken) {
			return SessionPrincipal{}, &APIError{Status: http.StatusForbidden, Code: "csrf_verification_failed", Details: map[string]any{}}
		}
		if apiErr := ValidateCSRF(r.Method, authSource, csrfCookie.Value, r.Header.Get(authn.CSRFHeaderName)); apiErr != nil {
			return SessionPrincipal{}, apiErr
		}
	}

	session, user, err := store.GetSessionByFingerprint(r.Context(), authn.FingerprintToken(keys, sessionToken))
	if errors.Is(err, authn.ErrNotFound) {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}
	if err != nil {
		return SessionPrincipal{}, internalAPIError(err)
	}
	if !user.IsActive || session.RevokedAt != nil {
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	current := now()
	if !session.SessionExpiresAt.After(current) {
		_ = store.RevokeSession(context.Background(), session.ID, "session_expired", current)
		if hub != nil {
			hub.RevokeSession(session.ID, "session_expired")
		}
		return SessionPrincipal{}, &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
	}

	return SessionPrincipal{
		AuthSource:   authSource,
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	}, nil
}
