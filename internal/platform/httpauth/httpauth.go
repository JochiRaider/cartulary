package httpauth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type AuthSource string

const (
	AuthSourceCookie AuthSource = "cookie"
	AuthSourceBearer AuthSource = "bearer"
)

type Principal struct {
	AuthSource   AuthSource
	SessionToken string
	Session      authn.SessionRecord
	User         authn.UserRecord
}

type SessionStore interface {
	GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error)
	GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error)
	RevokeSession(context.Context, uuid.UUID, string, time.Time) error
}

type SessionSlider interface {
	SlideSession(context.Context, uuid.UUID, authn.SessionTiming) (authn.SessionTiming, error)
}

type SessionRevoker interface {
	RevokeSession(uuid.UUID, string)
}

type Options struct {
	Store         SessionStore
	Keys          authn.MasterKeys
	Hub           SessionRevoker
	Now           func() time.Time
	StateChanging bool
}

const SessionSlideWriteInterval = time.Minute

func AuthenticateRequest(r *http.Request, options Options) (Principal, *httpapi.APIError) {
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return Principal{}, httpapi.SessionRequiredError()
		}
		if principal, apiErr := authenticateSessionToken(r, options.Store, options.Keys, options.Hub, now, AuthSourceBearer, token, options.StateChanging); apiErr == nil {
			return principal, nil
		} else if apiErr.Code != "session_required" {
			return Principal{}, apiErr
		}

		bootstrapToken, _, err := options.Store.GetBootstrapTokenByFingerprint(r.Context(), authn.FingerprintToken(options.Keys, token))
		if err == nil {
			if reason := authn.BootstrapReasonCode(bootstrapToken, now()); reason != "" {
				return Principal{}, BootstrapRejectedError(reason)
			}
			return Principal{}, BootstrapRejectedError("not_allowed_for_route")
		}
		if err != nil && !errors.Is(err, authn.ErrNotFound) {
			return Principal{}, httpapi.InternalAPIError(err)
		}
		return Principal{}, httpapi.SessionRequiredError()
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return Principal{}, httpapi.SessionRequiredError()
		}
		return Principal{}, httpapi.InternalAPIError(err)
	}
	return authenticateSessionToken(r, options.Store, options.Keys, options.Hub, now, AuthSourceCookie, cookie.Value, options.StateChanging)
}

func ValidateCSRF(method string, authSource AuthSource, cookieValue string, headerValue string) *httpapi.APIError {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return nil
	}
	if authSource != AuthSourceCookie {
		return nil
	}
	if cookieValue == "" || headerValue == "" || cookieValue != headerValue {
		return httpapi.CSRFVerificationFailedError()
	}
	return nil
}

func ShouldSlideIdleExpiry(method string, path string) bool {
	return !(method == http.MethodGet && path == "/api/v1/auth/session")
}

func ShouldPersistIdleExpirySlide(timing authn.SessionTiming, now time.Time) bool {
	if now.After(timing.AbsoluteExpiresAt) || !timing.SessionExpiresAt.After(now) {
		return false
	}
	return !now.Before(timing.LastQualifyingActivityAt.Add(SessionSlideWriteInterval))
}

func SlideSessionIfNeeded(ctx context.Context, store SessionSlider, principal *Principal, method string, path string, now func() time.Time) error {
	if principal == nil || !ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	timing := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}
	current := now()
	return slideSession(ctx, store, principal, timing.Slide(current))
}

func SlideSessionIfPersistenceDue(ctx context.Context, store SessionSlider, principal *Principal, method string, path string, now func() time.Time) error {
	if principal == nil || !ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	timing := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}
	current := now()
	if !ShouldPersistIdleExpirySlide(timing, current) {
		return nil
	}
	return slideSession(ctx, store, principal, timing.Slide(current))
}

func slideSession(ctx context.Context, store SessionSlider, principal *Principal, timing authn.SessionTiming) error {
	persisted, err := store.SlideSession(ctx, principal.Session.ID, timing)
	if err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	return nil
}

func RequireDeploymentAdmin(user authn.UserRecord) *httpapi.APIError {
	if user.IsDeploymentAdmin {
		return nil
	}
	return httpapi.AuthorizationDeniedCapabilityError("deployment_admin")
}

func BootstrapRejectedError(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "credential_bootstrap_rejected",
		Message: "credential bootstrap token cannot be used",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func authenticateSessionToken(
	r *http.Request,
	store SessionStore,
	keys authn.MasterKeys,
	hub SessionRevoker,
	now func() time.Time,
	authSource AuthSource,
	sessionToken string,
	stateChanging bool,
) (Principal, *httpapi.APIError) {
	if stateChanging && authSource == AuthSourceCookie {
		csrfCookie, _ := r.Cookie(authn.CSRFCookieName)
		if csrfCookie == nil || csrfCookie.Value != authn.CSRFTokenForSessionToken(keys, sessionToken) {
			return Principal{}, httpapi.CSRFVerificationFailedError()
		}
		if apiErr := ValidateCSRF(r.Method, authSource, csrfCookie.Value, r.Header.Get(authn.CSRFHeaderName)); apiErr != nil {
			return Principal{}, apiErr
		}
	}

	session, user, err := store.GetSessionByFingerprint(r.Context(), authn.FingerprintToken(keys, sessionToken))
	if errors.Is(err, authn.ErrNotFound) {
		return Principal{}, httpapi.SessionRequiredError()
	}
	if err != nil {
		return Principal{}, httpapi.InternalAPIError(err)
	}
	if !user.IsActive || session.RevokedAt != nil {
		return Principal{}, httpapi.SessionRequiredError()
	}

	current := now()
	if !session.SessionExpiresAt.After(current) {
		_ = store.RevokeSession(context.Background(), session.ID, "session_expired", current)
		if hub != nil {
			hub.RevokeSession(session.ID, "session_expired")
		}
		return Principal{}, httpapi.SessionRequiredError()
	}

	return Principal{
		AuthSource:   authSource,
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	}, nil
}
