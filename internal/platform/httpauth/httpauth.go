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

// BootstrapCredential is a credential-setup bootstrap token admitted for an
// operation whose declared security includes credentialBootstrapBearer (Core 01
// REQ-01-522). It is not a session.
type BootstrapCredential struct {
	Token authn.BootstrapTokenRecord
	User  authn.UserRecord
}

// credential is exactly one of a session principal or a bootstrap credential.
type credential struct {
	session   *Principal
	bootstrap *BootstrapCredential
}

// AuthenticateRequest establishes the request principal from the bearer
// session or the session cookie, verifying CSRF for a cookie when
// StateChanging. A bootstrap token is never admitted here: it fails with
// credential_bootstrap_rejected. Each attempt, CSRF verification and principal
// is reported to the declared-security guard (Core 01 REQ-01-703) when one
// observes the request.
func AuthenticateRequest(r *http.Request, options Options) (Principal, *httpapi.APIError) {
	admitted, apiErr := authenticate(r, options, false)
	if apiErr != nil {
		return Principal{}, apiErr
	}
	if admitted.session == nil {
		return Principal{}, httpapi.InternalAPIError(errors.New("authentication admitted no session"))
	}
	return *admitted.session, nil
}

// authenticate applies the credential precedence of Core 01 REQ-01-522 and
// Table 3.3.2.2-D. A bearer header is tried first as a session. Only when it
// names no usable session is it tried as a bootstrap token: an invalid token is
// rejected with its reason, then a token presented where allowBootstrap is
// false with not_allowed_for_route, then a token presented together with a
// session cookie as a second auth mode. Without a bearer header the session
// cookie authenticates, with CSRF verified before the session lookup.
func authenticate(r *http.Request, options Options, allowBootstrap bool) (credential, *httpapi.APIError) {
	httpapi.ObserveAuthenticationAttempt(r.Context())
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return credential{}, httpapi.SessionRequiredError()
		}
		if principal, apiErr := authenticateSessionToken(r, options.Store, options.Keys, options.Hub, now, AuthSourceBearer, token, options.StateChanging); apiErr == nil {
			return credential{session: &principal}, nil
		} else if apiErr.Code != "session_required" {
			return credential{}, apiErr
		}

		bootstrapToken, user, err := options.Store.GetBootstrapTokenByFingerprint(r.Context(), authn.FingerprintToken(options.Keys, token))
		if errors.Is(err, authn.ErrNotFound) {
			return credential{}, httpapi.SessionRequiredError()
		}
		if err != nil {
			return credential{}, httpapi.InternalAPIError(err)
		}
		if reason := authn.BootstrapReasonCode(bootstrapToken, now()); reason != "" {
			return credential{}, BootstrapRejectedError(reason)
		}
		if !allowBootstrap {
			return credential{}, BootstrapRejectedError("not_allowed_for_route")
		}
		if cookie, _ := r.Cookie(authn.SessionCookieName); cookie != nil {
			return credential{}, BootstrapAuthModeConflictError()
		}
		httpapi.ObservePrincipal(r.Context(), httpapi.AuthenticatedByCredentialBootstrap)
		return credential{bootstrap: &BootstrapCredential{Token: bootstrapToken, User: user}}, nil
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return credential{}, httpapi.SessionRequiredError()
		}
		return credential{}, httpapi.InternalAPIError(err)
	}
	principal, apiErr := authenticateSessionToken(r, options.Store, options.Keys, options.Hub, now, AuthSourceCookie, cookie.Value, options.StateChanging)
	if apiErr != nil {
		return credential{}, apiErr
	}
	return credential{session: &principal}, nil
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

// validatePrincipalCSRF checks an operation's declared cookie-CSRF requirement
// against an already authenticated principal. It performs no session lookup and
// does not infer operation security from the HTTP method.
func validatePrincipalCSRF(r *http.Request, principal Principal, keys authn.MasterKeys) *httpapi.APIError {
	if principal.AuthSource == AuthSourceBearer {
		return nil
	}
	httpapi.ObserveCSRFVerification(r.Context())
	if principal.AuthSource != AuthSourceCookie || principal.SessionToken == "" {
		return httpapi.CSRFVerificationFailedError()
	}
	cookie, err := r.Cookie(authn.CSRFCookieName)
	if err != nil || cookie.Value != authn.CSRFTokenForSessionToken(keys, principal.SessionToken) || r.Header.Get(authn.CSRFHeaderName) != cookie.Value {
		return httpapi.CSRFVerificationFailedError()
	}
	return nil
}

// sessionInspectionOperationID is GET /api/v1/auth/session, which is not
// qualifying activity (Core 04 REQ-04-010).
const sessionInspectionOperationID = "getCurrentSession"

// QualifiesForIdleSlide reports whether a successful request is qualifying
// activity that slides idle expiry (Core 04 REQ-04-010). It is decided by the
// resolved operation: only session inspection is excluded. A request resolved
// to no public operation qualifies.
func QualifiesForIdleSlide(ctx context.Context) bool {
	metadata, ok := httpapi.OperationMetadataFromContext(ctx)
	return !ok || metadata.OperationID != sessionInspectionOperationID
}

func ShouldPersistIdleExpirySlide(timing authn.SessionTiming, now time.Time) bool {
	if now.After(timing.AbsoluteExpiresAt) || !timing.SessionExpiresAt.After(now) {
		return false
	}
	return !now.Before(timing.LastQualifyingActivityAt.Add(SessionSlideWriteInterval))
}

// SlideSessionIfNeeded slides the principal's idle expiry for the request's
// operation when it is qualifying activity (QualifiesForIdleSlide).
func SlideSessionIfNeeded(ctx context.Context, store SessionSlider, principal *Principal, now func() time.Time) error {
	if principal == nil || !QualifiesForIdleSlide(ctx) {
		return nil
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return slideSession(ctx, store, principal, principalTiming(principal).Slide(now()))
}

// SlideSessionIfPersistenceDue is SlideSessionIfNeeded that writes at most
// once per SessionSlideWriteInterval.
func SlideSessionIfPersistenceDue(ctx context.Context, store SessionSlider, principal *Principal, now func() time.Time) error {
	if principal == nil || !QualifiesForIdleSlide(ctx) {
		return nil
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	timing := principalTiming(principal)
	current := now()
	if !ShouldPersistIdleExpirySlide(timing, current) {
		return nil
	}
	return slideSession(ctx, store, principal, timing.Slide(current))
}

func principalTiming(principal *Principal) authn.SessionTiming {
	return authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}
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

// BootstrapAuthModeConflictError rejects a bootstrap token presented together
// with a session cookie: the TOTP setup routes accept exactly one auth mode
// (Core 01 Table 3.3.2.2-D).
func BootstrapAuthModeConflictError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_auth_request",
		Message: "exactly one auth mode is allowed",
		Details: map[string]any{"field": "authorization"},
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
		httpapi.ObserveCSRFVerification(r.Context())
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

	ObservePrincipal(r, authSource)
	return Principal{
		AuthSource:   authSource,
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	}, nil
}

// ObservePrincipal reports a principal established from a session cookie or
// bearer session to the declared-security guard, when one observes the
// request. Owners that authenticate without AuthenticateRequest call it.
func ObservePrincipal(r *http.Request, authSource AuthSource) {
	source := httpapi.AuthenticatedByBearerSession
	if authSource == AuthSourceCookie {
		source = httpapi.AuthenticatedBySessionCookie
	}
	httpapi.ObservePrincipal(r.Context(), source)
}

// Authenticator is the central admission authenticator (Core 01 REQ-01-707
// step 2). It authenticates exactly as AuthenticateRequest does, with cookie
// CSRF required when the operation's declared security pairs it, and admits a
// bootstrap token only for an operation whose declared security includes
// credentialBootstrapBearer. It attaches the session principal for
// AdmittedPrincipal or the bootstrap credential for
// AdmittedBootstrapCredential. With a revocation hub, a session found expired
// is published as session_expired to its sockets (Core 01 REQ-01-277). A
// family gate that authenticates before operation selection verifies the
// selected operation's cookie CSRF afterwards through VerifyCSRF (§17.4).
type Authenticator struct {
	options Options
}

var _ httpapi.Authenticator = (*Authenticator)(nil)

// NewAuthenticator returns an authenticator over the session store, keys,
// optional revocation hub and clock of options. Options.StateChanging is
// ignored; each operation's declared security decides it.
func NewAuthenticator(options Options) *Authenticator {
	options.StateChanging = false
	return &Authenticator{options: options}
}

// DependencyAuthenticator builds the authenticator from the application
// dependencies: the master keys in deps.Env, the session store over
// deps.PostgresHandle() and deps.Now. hub, when non-nil, receives the
// session_expired revocation of a session found expired.
func DependencyAuthenticator(deps httpapi.DependencySet, hub SessionRevoker) (*Authenticator, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	return NewAuthenticator(Options{Store: authn.NewStore(deps.PostgresHandle()), Keys: keys, Hub: hub, Now: deps.Now}), nil
}

type admittedPrincipalContextKey struct{}

type admittedBootstrapCredentialContextKey struct{}

// Authenticate implements httpapi.Authenticator.
func (authenticator *Authenticator) Authenticate(r *http.Request, requirement httpapi.SecurityRequirement) (*http.Request, *httpapi.APIError) {
	if authenticator == nil || requirement.Public {
		return r, httpapi.InternalAPIError(errors.New("central authentication does not apply to this operation"))
	}
	options := authenticator.options
	options.StateChanging = requirement.CookieCSRF
	admitted, apiErr := authenticate(r, options, requirement.Bootstrap)
	if apiErr != nil {
		return r, apiErr
	}
	if admitted.bootstrap != nil {
		return r.WithContext(context.WithValue(r.Context(), admittedBootstrapCredentialContextKey{}, *admitted.bootstrap)), nil
	}
	return r.WithContext(context.WithValue(r.Context(), admittedPrincipalContextKey{}, *admitted.session)), nil
}

// VerifyCSRF implements httpapi.Authenticator: it checks the double-submit
// proof of the session principal Authenticate admitted, without another session
// lookup. A bearer session needs none. A request without an admitted session
// principal is a binding error.
func (authenticator *Authenticator) VerifyCSRF(r *http.Request) *httpapi.APIError {
	principal, ok := AdmittedPrincipal(r)
	if authenticator == nil || !ok {
		return httpapi.InternalAPIError(errors.New("CSRF verification requires an admitted session principal"))
	}
	return validatePrincipalCSRF(r, principal, authenticator.options.Keys)
}

// AdmittedPrincipal returns the session principal central admission
// established. It reports false when the operation was not admitted centrally,
// which is a binding error, or when a bootstrap credential was admitted.
func AdmittedPrincipal(r *http.Request) (Principal, bool) {
	principal, ok := r.Context().Value(admittedPrincipalContextKey{}).(Principal)
	return principal, ok
}

// AdmittedBootstrapCredential returns the bootstrap credential central
// admission established for an operation whose declared security includes
// credentialBootstrapBearer. It reports false when a session principal was
// admitted instead.
func AdmittedBootstrapCredential(r *http.Request) (BootstrapCredential, bool) {
	bootstrap, ok := r.Context().Value(admittedBootstrapCredentialContextKey{}).(BootstrapCredential)
	return bootstrap, ok
}
