package httpauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

var (
	credentialTestNow       = time.Date(2026, time.November, 19, 12, 0, 0, 0, time.UTC)
	credentialTestUserID    = uuid.MustParse("10000000-0000-4000-8000-000000000001")
	credentialTestSessionID = uuid.MustParse("10000000-0000-4000-8000-000000000002")
	credentialTestTokenID   = uuid.MustParse("10000000-0000-4000-8000-000000000003")
)

// credentialStore resolves fingerprints of the named tokens and records every
// lookup and revocation.
type credentialStore struct {
	keys           authn.MasterKeys
	sessions       map[string]authn.SessionRecord
	bootstraps     map[string]authn.BootstrapTokenRecord
	bootstrapErr   error
	sessionLookups int
	revoked        []string
}

func (store *credentialStore) GetSessionByFingerprint(_ context.Context, fingerprint []byte) (authn.SessionRecord, authn.UserRecord, error) {
	store.sessionLookups++
	for token, session := range store.sessions {
		if string(authn.FingerprintToken(store.keys, token)) == string(fingerprint) {
			return session, authn.UserRecord{ID: credentialTestUserID, IsActive: true}, nil
		}
	}
	return authn.SessionRecord{}, authn.UserRecord{}, authn.ErrNotFound
}

func (store *credentialStore) GetBootstrapTokenByFingerprint(_ context.Context, fingerprint []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
	if store.bootstrapErr != nil {
		return authn.BootstrapTokenRecord{}, authn.UserRecord{}, store.bootstrapErr
	}
	for token, record := range store.bootstraps {
		if string(authn.FingerprintToken(store.keys, token)) == string(fingerprint) {
			return record, authn.UserRecord{ID: credentialTestUserID, IsActive: true}, nil
		}
	}
	return authn.BootstrapTokenRecord{}, authn.UserRecord{}, authn.ErrNotFound
}

func (store *credentialStore) RevokeSession(_ context.Context, sessionID uuid.UUID, reasonCode string, _ time.Time) error {
	store.revoked = append(store.revoked, sessionID.String()+":"+reasonCode)
	return nil
}

type revocationHub struct{ published []string }

func (hub *revocationHub) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	hub.published = append(hub.published, sessionID.String()+":"+reasonCode)
}

func credentialTestSession(expiresAt time.Time, revoked bool) authn.SessionRecord {
	timing := authn.NewSessionTiming(credentialTestNow.Add(-time.Minute))
	session := authn.SessionRecord{
		ID: credentialTestSessionID, UserID: credentialTestUserID,
		AuthenticatedAt: timing.AuthenticatedAt, LastQualifyingActivityAt: timing.LastQualifyingActivityAt,
		IdleExpiresAt: expiresAt, AbsoluteExpiresAt: timing.AbsoluteExpiresAt, SessionExpiresAt: expiresAt,
	}
	if revoked {
		revokedAt := credentialTestNow.Add(-time.Second)
		session.RevokedAt = &revokedAt
	}
	return session
}

func credentialTestBootstrap(status string) authn.BootstrapTokenRecord {
	record := authn.BootstrapTokenRecord{ID: credentialTestTokenID, UserID: credentialTestUserID, IssuedAt: credentialTestNow.Add(-time.Minute), ExpiresAt: credentialTestNow.Add(9 * time.Minute)}
	used := credentialTestNow.Add(-time.Second)
	switch status {
	case "expired":
		record.ExpiresAt = credentialTestNow.Add(-time.Second)
	case "consumed":
		record.ConsumedAt = &used
	case "superseded":
		record.SupersededAt = &used
	}
	return record
}

func newCredentialStore(t *testing.T) *credentialStore {
	t.Helper()
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &credentialStore{
		keys: keys,
		sessions: map[string]authn.SessionRecord{
			"live-session":    credentialTestSession(credentialTestNow.Add(20*time.Minute), false),
			"expired-session": credentialTestSession(credentialTestNow.Add(-time.Second), false),
			"revoked-session": credentialTestSession(credentialTestNow.Add(20*time.Minute), true),
		},
		bootstraps: map[string]authn.BootstrapTokenRecord{
			"valid-bootstrap":      credentialTestBootstrap("valid"),
			"expired-bootstrap":    credentialTestBootstrap("expired"),
			"consumed-bootstrap":   credentialTestBootstrap("consumed"),
			"superseded-bootstrap": credentialTestBootstrap("superseded"),
		},
	}
}

func credentialRequest(store *credentialStore, bearer string, cookie string, csrf bool) *http.Request {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/fixture", nil)
	switch bearer {
	case "":
	case " ":
		request.Header.Set("Authorization", "Bearer ")
	default:
		request.Header.Set("Authorization", "Bearer "+bearer)
	}
	if cookie != "" {
		request.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: cookie})
		if csrf {
			token := authn.CSRFTokenForSessionToken(store.keys, cookie)
			request.AddCookie(&http.Cookie{Name: authn.CSRFCookieName, Value: token})
			request.Header.Set(authn.CSRFHeaderName, token)
		}
	}
	return request
}

func requireCredentialError(t *testing.T, apiErr *httpapi.APIError, status int, code string, detail string, value string) {
	t.Helper()
	if apiErr == nil || apiErr.Status != status || apiErr.Code != code {
		t.Fatalf("error = %#v, want %d %s", apiErr, status, code)
	}
	if detail != "" && apiErr.Details[detail] != value {
		t.Fatalf("details = %#v, want %s=%s", apiErr.Details, detail, value)
	}
}

// TestCentralAuthenticatorCredentialPrecedence covers the credential order of
// Core 01 REQ-01-522, REQ-01-523 and Table 3.3.2.2-D as central admission
// applies it (REQ-01-703, REQ-01-707 step 2). A bearer header is tried first as
// a session and wins over a cookie. Only otherwise is it tried as a bootstrap
// token: an invalid token is rejected with its reason before the operation is
// considered, an operation that does not declare credentialBootstrapBearer
// rejects a valid one with not_allowed_for_route, and a valid one presented
// with a session cookie is a second auth mode. A cookie session verifies CSRF
// before the session lookup when the operation pairs it, and VerifyCSRF checks
// an admitted cookie principal's proof later without a second lookup.
func TestCentralAuthenticatorCredentialPrecedence(t *testing.T) {
	ordinary := httpapi.SecurityRequirement{CookieCSRF: true}
	setup := httpapi.SecurityRequirement{CookieCSRF: true, Bootstrap: true}
	for _, tc := range []struct {
		name         string
		requirement  httpapi.SecurityRequirement
		bearer       string
		cookie       string
		csrf         bool
		bootstrapErr error
		status       int
		code         string
		detail       string
		value        string
		source       AuthSource
		bootstrap    bool
		lookups      int
	}{
		{name: "no credential", requirement: setup, status: http.StatusUnauthorized, code: "session_required"},
		{name: "empty bearer", requirement: setup, bearer: " ", status: http.StatusUnauthorized, code: "session_required"},
		{name: "bearer session without csrf", requirement: setup, bearer: "live-session", source: AuthSourceBearer, lookups: 1},
		{name: "bearer session wins over a cookie", requirement: ordinary, bearer: "live-session", cookie: "revoked-session", source: AuthSourceBearer, lookups: 1},
		{name: "bootstrap where declared", requirement: setup, bearer: "valid-bootstrap", bootstrap: true, lookups: 1},
		{name: "bootstrap where not declared", requirement: ordinary, bearer: "valid-bootstrap", status: http.StatusConflict, code: "credential_bootstrap_rejected", detail: "reason_code", value: "not_allowed_for_route", lookups: 1},
		{name: "expired bootstrap before route", requirement: ordinary, bearer: "expired-bootstrap", status: http.StatusConflict, code: "credential_bootstrap_rejected", detail: "reason_code", value: "expired", lookups: 1},
		{name: "consumed bootstrap where declared", requirement: setup, bearer: "consumed-bootstrap", status: http.StatusConflict, code: "credential_bootstrap_rejected", detail: "reason_code", value: "consumed", lookups: 1},
		{name: "superseded bootstrap before a second mode", requirement: setup, bearer: "superseded-bootstrap", cookie: "live-session", csrf: true, status: http.StatusConflict, code: "credential_bootstrap_rejected", detail: "reason_code", value: "superseded", lookups: 1},
		{name: "bootstrap with a session cookie", requirement: setup, bearer: "valid-bootstrap", cookie: "live-session", csrf: true, status: http.StatusBadRequest, code: "invalid_auth_request", detail: "field", value: "authorization", lookups: 1},
		{name: "unknown bearer", requirement: setup, bearer: "unknown-token", status: http.StatusUnauthorized, code: "session_required", lookups: 1},
		{name: "bootstrap lookup failure", requirement: setup, bearer: "unknown-token", bootstrapErr: errors.New("bootstrap store unavailable"), status: http.StatusInternalServerError, code: "internal_error", lookups: 1},
		{name: "cookie without csrf proof", requirement: setup, cookie: "live-session", status: http.StatusForbidden, code: "csrf_verification_failed"},
		{name: "cookie with csrf proof", requirement: setup, cookie: "live-session", csrf: true, source: AuthSourceCookie, lookups: 1},
		{name: "cookie read without csrf", requirement: httpapi.SecurityRequirement{}, cookie: "live-session", source: AuthSourceCookie, lookups: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newCredentialStore(t)
			store.bootstrapErr = tc.bootstrapErr
			authenticator := NewAuthenticator(Options{Store: store, Keys: store.keys, Now: func() time.Time { return credentialTestNow }})
			admitted, apiErr := authenticator.Authenticate(credentialRequest(store, tc.bearer, tc.cookie, tc.csrf), tc.requirement)
			if store.sessionLookups != tc.lookups {
				t.Fatalf("session lookups = %d, want %d", store.sessionLookups, tc.lookups)
			}
			if tc.code != "" {
				requireCredentialError(t, apiErr, tc.status, tc.code, tc.detail, tc.value)
				return
			}
			if apiErr != nil {
				t.Fatalf("authenticate: %#v", apiErr)
			}
			principal, sessionAdmitted := AdmittedPrincipal(admitted)
			bootstrap, bootstrapAdmitted := AdmittedBootstrapCredential(admitted)
			if tc.bootstrap {
				if sessionAdmitted || !bootstrapAdmitted || bootstrap.Token.ID != credentialTestTokenID || bootstrap.User.ID != credentialTestUserID {
					t.Fatalf("bootstrap admission: session=%t bootstrap=%#v", sessionAdmitted, bootstrap)
				}
				return
			}
			if bootstrapAdmitted || !sessionAdmitted || principal.AuthSource != tc.source || principal.Session.ID != credentialTestSessionID {
				t.Fatalf("session admission: principal=%#v bootstrap=%t", principal, bootstrapAdmitted)
			}
		})
	}

	// A family gate authenticates without CSRF; the selected operation then
	// verifies its declared cookie CSRF on the admitted principal (§17.4).
	t.Run("operation-time CSRF verifies the admitted principal", func(t *testing.T) {
		for _, tc := range []struct {
			name, bearer, cookie string
			csrf, wrongHeader    bool
			status               int
			code                 string
		}{
			{name: "cookie without proof", cookie: "live-session", status: http.StatusForbidden, code: "csrf_verification_failed"},
			{name: "cookie with proof", cookie: "live-session", csrf: true},
			{name: "cookie with a mismatched header", cookie: "live-session", csrf: true, wrongHeader: true, status: http.StatusForbidden, code: "csrf_verification_failed"},
			{name: "bearer session needs no proof", bearer: "live-session"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				store := newCredentialStore(t)
				authenticator := NewAuthenticator(Options{Store: store, Keys: store.keys, Now: func() time.Time { return credentialTestNow }})
				request := credentialRequest(store, tc.bearer, tc.cookie, tc.csrf)
				if tc.wrongHeader {
					request.Header.Set(authn.CSRFHeaderName, "wrong")
				}
				admitted, apiErr := authenticator.Authenticate(request, httpapi.SecurityRequirement{})
				if apiErr != nil {
					t.Fatalf("authenticate: %#v", apiErr)
				}
				apiErr = authenticator.VerifyCSRF(admitted)
				if store.sessionLookups != 1 {
					t.Fatalf("session lookups = %d, want 1", store.sessionLookups)
				}
				if tc.code != "" {
					requireCredentialError(t, apiErr, tc.status, tc.code, "", "")
				} else if apiErr != nil {
					t.Fatalf("verify CSRF: %#v", apiErr)
				}
			})
		}
		t.Run("without an admitted principal", func(t *testing.T) {
			store := newCredentialStore(t)
			apiErr := NewAuthenticator(Options{Store: store, Keys: store.keys}).VerifyCSRF(credentialRequest(store, "", "live-session", true))
			requireCredentialError(t, apiErr, http.StatusInternalServerError, "internal_error", "", "")
		})
	})
	t.Run("a public operation is not authenticated centrally", func(t *testing.T) {
		store := newCredentialStore(t)
		_, apiErr := NewAuthenticator(Options{Store: store, Keys: store.keys}).Authenticate(credentialRequest(store, "live-session", "", false), httpapi.SecurityRequirement{Public: true})
		requireCredentialError(t, apiErr, http.StatusInternalServerError, "internal_error", "", "")
	})
	t.Run("AuthenticateRequest never admits a bootstrap token", func(t *testing.T) {
		store := newCredentialStore(t)
		_, apiErr := AuthenticateRequest(credentialRequest(store, "valid-bootstrap", "", false), Options{Store: store, Keys: store.keys, Now: func() time.Time { return credentialTestNow }})
		requireCredentialError(t, apiErr, http.StatusConflict, "credential_bootstrap_rejected", "reason_code", "not_allowed_for_route")
	})
}

// TestCentralAuthenticatorPublishesSessionExpiry covers Core 01 REQ-01-277 and
// Core 04 AC-157 for every centrally admitted owner: a session found expired is
// revoked with session_expired and published to its sockets through the hub,
// while a session already revoked is rejected without a second revocation or
// publication.
func TestCentralAuthenticatorPublishesSessionExpiry(t *testing.T) {
	for _, tc := range []struct {
		name, cookie string
		want         []string
	}{
		{"expired", "expired-session", []string{credentialTestSessionID.String() + ":session_expired"}},
		{"already revoked", "revoked-session", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := newCredentialStore(t)
			hub := &revocationHub{}
			authenticator := NewAuthenticator(Options{Store: store, Keys: store.keys, Hub: hub, Now: func() time.Time { return credentialTestNow }})
			_, apiErr := authenticator.Authenticate(credentialRequest(store, "", tc.cookie, false), httpapi.SecurityRequirement{})
			requireCredentialError(t, apiErr, http.StatusUnauthorized, "session_required", "", "")
			if !slices.Equal(store.revoked, tc.want) || !slices.Equal(hub.published, tc.want) {
				t.Fatalf("revoked=%v published=%v, want %v", store.revoked, hub.published, tc.want)
			}
		})
	}
	t.Run("without a hub the expiry is still revoked", func(t *testing.T) {
		store := newCredentialStore(t)
		_, apiErr := NewAuthenticator(Options{Store: store, Keys: store.keys, Now: func() time.Time { return credentialTestNow }}).Authenticate(credentialRequest(store, "", "expired-session", false), httpapi.SecurityRequirement{})
		requireCredentialError(t, apiErr, http.StatusUnauthorized, "session_required", "", "")
		if len(store.revoked) != 1 {
			t.Fatalf("revoked = %v", store.revoked)
		}
	})
}

type slideRecorder struct{ slides int }

func (slider *slideRecorder) SlideSession(_ context.Context, _ uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
	slider.slides++
	return timing, nil
}

// slideAt resolves a request to the generated operation, then applies slide in
// that request's context and reports how many slides it persisted.
func slideAt(t *testing.T, operationID string, slide func(context.Context, SessionSlider) error) int {
	t.Helper()
	registry, err := httpapi.NewRouteRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	var operation httpapi.ContractOperation
	for _, candidate := range httpapi.ContractOperations() {
		if candidate.OperationID == operationID {
			operation = candidate
		}
	}
	slider := &slideRecorder{}
	resolved := false
	mux := http.NewServeMux()
	if err := registry.Bind(mux, operation.OwnerID, operationID, func(_ http.ResponseWriter, r *http.Request) {
		resolved = true
		if err := slide(r.Context(), slider); err != nil {
			t.Error(err)
		}
	}); err != nil {
		t.Fatal(err)
	}
	registry.Handler(mux).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(operation.Method, operation.PathTemplate, nil))
	if !resolved {
		t.Fatalf("%s did not resolve", operationID)
	}
	return slider.slides
}

// TestSessionSlidingIsKeyedByOperation covers Core 04 REQ-04-010 and
// REQ-04-011: qualifying activity is decided by the resolved operation, so
// session inspection (getCurrentSession) never slides idle expiry whatever its
// path spelling, every other operation and a request resolved to no public
// operation slide, and the persistence-due cadence writes at most once per
// SessionSlideWriteInterval.
func TestSessionSlidingIsKeyedByOperation(t *testing.T) {
	now := func() time.Time { return credentialTestNow }
	recent := credentialTestSession(credentialTestNow.Add(20*time.Minute), false)
	recent.LastQualifyingActivityAt = credentialTestNow.Add(-time.Second)
	stale := credentialTestSession(credentialTestNow.Add(20*time.Minute), false)
	stale.LastQualifyingActivityAt = credentialTestNow.Add(-2 * SessionSlideWriteInterval)
	always := func(session authn.SessionRecord) func(context.Context, SessionSlider) error {
		return func(ctx context.Context, slider SessionSlider) error {
			return SlideSessionIfNeeded(ctx, slider, &Principal{Session: session}, now)
		}
	}
	due := func(session authn.SessionRecord) func(context.Context, SessionSlider) error {
		return func(ctx context.Context, slider SessionSlider) error {
			return SlideSessionIfPersistenceDue(ctx, slider, &Principal{Session: session}, now)
		}
	}
	for _, tc := range []struct {
		name, operationID string
		slide             func(context.Context, SessionSlider) error
		want              int
	}{
		{"session inspection never slides", "getCurrentSession", always(recent), 0},
		{"session inspection never slides when persistence is due", "getCurrentSession", due(stale), 0},
		{"credential state slides", "getCredentialState", always(recent), 1},
		{"another owner's read slides", "getJob", always(recent), 1},
		{"persistence not yet due", "getIncident", due(recent), 0},
		{"persistence due", "getIncident", due(stale), 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := slideAt(t, tc.operationID, tc.slide); got != tc.want {
				t.Fatalf("slides = %d, want %d", got, tc.want)
			}
		})
	}
	t.Run("a request resolved to no public operation qualifies", func(t *testing.T) {
		slider := &slideRecorder{}
		if !QualifiesForIdleSlide(context.Background()) {
			t.Fatal("an unresolved request must qualify")
		}
		if err := always(recent)(context.Background(), slider); err != nil || slider.slides != 1 {
			t.Fatalf("slides = %d err = %v", slider.slides, err)
		}
		if err := SlideSessionIfNeeded(context.Background(), slider, nil, now); err != nil || slider.slides != 1 {
			t.Fatal("a nil principal must not slide")
		}
	})
}
