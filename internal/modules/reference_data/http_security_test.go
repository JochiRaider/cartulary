package reference_data

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
	"github.com/google/uuid"
)

var (
	familyTestNow       = time.Date(2026, time.November, 19, 12, 0, 0, 0, time.UTC)
	familyTestSessionID = uuid.MustParse("20000000-0000-4000-8000-000000000001")
)

// admissionSessionStub resolves every session token to one session of a
// deployment administrator or an ordinary user, live or expired, and records
// lookups, revocations and slides.
type admissionSessionStub struct {
	admin, expired  bool
	lookups, slides int
	revoked         []string
}

func (s *admissionSessionStub) GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
	s.lookups++
	expiresAt := familyTestNow.Add(time.Hour)
	if s.expired {
		expiresAt = familyTestNow.Add(-time.Second)
	}
	return authn.SessionRecord{ID: familyTestSessionID, SessionExpiresAt: expiresAt, AbsoluteExpiresAt: expiresAt}, authn.UserRecord{IsActive: true, IsDeploymentAdmin: s.admin}, nil
}
func (*admissionSessionStub) GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
	return authn.BootstrapTokenRecord{}, authn.UserRecord{}, authn.ErrNotFound
}
func (s *admissionSessionStub) RevokeSession(_ context.Context, sessionID uuid.UUID, reasonCode string, _ time.Time) error {
	s.revoked = append(s.revoked, sessionID.String()+":"+reasonCode)
	return nil
}
func (s *admissionSessionStub) SlideSession(_ context.Context, _ uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
	s.slides++
	return timing, nil
}

// admissionHubStub records the session revocations published to sockets.
type admissionHubStub struct{ published []string }

func (hub *admissionHubStub) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	hub.published = append(hub.published, sessionID.String()+":"+reasonCode)
}

// admissionApplicationProbe records the reads that reach the application. No
// other application method is reachable in these tests.
type admissionApplicationProbe struct {
	AdministrativeApplication
	reads []string
}

func (probe *admissionApplicationProbe) ListVersions(context.Context) ([]AdministrativeVersion, error) {
	probe.reads = append(probe.reads, "list")
	return nil, nil
}
func (probe *admissionApplicationProbe) GetVersion(_ context.Context, packKey, packVersion string) (AdministrativeVersion, error) {
	probe.reads = append(probe.reads, packKey+"/"+packVersion)
	return AdministrativeVersion{}, ErrNotFound
}

type admissionNotifierProbe struct{}

func (admissionNotifierProbe) Notify(uuid.UUID) {}

type admissionBodyProbe struct{ reads int }

func (body *admissionBodyProbe) Read([]byte) (int, error) { body.reads++; return 0, io.EOF }

type familyAdmissionFixture struct {
	keys        authn.MasterKeys
	store       *admissionSessionStub
	hub         *admissionHubStub
	application *admissionApplicationProbe
	handler     http.Handler
}

// newFamilyAdmissionFixture binds the Reference Pack family through the route
// registry with the central authenticator over a session stub and a
// revocation hub, as the application does.
func newFamilyAdmissionFixture(t *testing.T, admin, expired bool) *familyAdmissionFixture {
	t.Helper()
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatal(err)
	}
	cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
	if err != nil {
		t.Fatal(err)
	}
	now := func() time.Time { return familyTestNow }
	fixture := &familyAdmissionFixture{keys: keys, store: &admissionSessionStub{admin: admin, expired: expired}, hub: &admissionHubStub{}, application: &admissionApplicationProbe{}}
	service := &httpAdapter{coordinator: fixture.application, sessions: fixture.store, jobRunner: admissionNotifierProbe{}, cursorCodec: pagination.NewCodec(cursorKey), now: now}
	deps := httpapiextensions.FromGeneratedRegistry(t, ProfileID).Dependencies(httpapi.DependencySet{
		Authenticator: httpauth.NewAuthenticator(httpauth.Options{Store: fixture.store, Keys: keys, Hub: fixture.hub, Now: now}),
	})
	fixture.handler, err = httpapi.NewHandler(httpapi.Options{Dependencies: deps, AdditionalRoutes: []httpapi.RouteRegistrar{
		func(_ *http.ServeMux, deps httpapi.DependencySet) error { return service.bindFamily(deps) },
	}})
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

// serve sends one request with an optional session cookie and CSRF proof and
// returns the status, error code and whether the body was read.
func (fixture *familyAdmissionFixture) serve(t *testing.T, method, target string, cookie, csrf bool) (int, string, int) {
	t.Helper()
	body := &admissionBodyProbe{}
	r := httptest.NewRequest(method, target, body)
	if cookie {
		r.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: "fixture-session"})
		if csrf {
			proof := authn.CSRFTokenForSessionToken(fixture.keys, "fixture-session")
			r.AddCookie(&http.Cookie{Name: authn.CSRFCookieName, Value: proof})
			r.Header.Set(authn.CSRFHeaderName, proof)
		}
	}
	w := httptest.NewRecorder()
	fixture.handler.ServeHTTP(w, r)
	var envelope httpapi.ErrorEnvelope
	if w.Code >= http.StatusBadRequest && method != http.MethodHead {
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("%s %s: unenveloped %d %q", method, target, w.Code, w.Body)
		}
	}
	return w.Code, envelope.Error.Code, body.reads
}

// testFamilyAdmissionEffects covers the effects of the Core 01 §17.4 family
// order as the registry binds it: a rejected request reads no body, reaches no
// application read and never slides the session; the gate's authentication
// publishes an expired session to its sockets; CSRF precedes query admission,
// which precedes path parameters; admitted values reach the handler decoded
// once; and a successful request slides the session.
func testFamilyAdmissionEffects(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		fixture := newFamilyAdmissionFixture(t, true, false)
		if status, code, reads := fixture.serve(t, http.MethodPost, "/api/v1/reference-packs/import?limit=1", false, false); status != 401 || code != "session_required" || reads != 0 {
			t.Fatalf("unauthenticated: %d %s reads=%d", status, code, reads)
		}
		if fixture.store.lookups != 0 || fixture.store.slides != 0 || len(fixture.application.reads) != 0 {
			t.Fatalf("unauthenticated effects: lookups=%d slides=%d reads=%v", fixture.store.lookups, fixture.store.slides, fixture.application.reads)
		}
	})
	t.Run("a non-admin is denied before every route check without CSRF", func(t *testing.T) {
		fixture := newFamilyAdmissionFixture(t, false, false)
		for _, tc := range []struct{ method, path string }{
			{http.MethodPost, "/api/v1/reference-packs/import?limit=1"},
			{http.MethodPatch, "/api/v1/reference-packs"},
			{http.MethodGet, "/api/v1/reference-packs/unknown/child/extra"},
			{http.MethodGet, "/api/v1/reference-packs//type_registry.host/1"},
			{http.MethodGet, "/api/v1/reference-packs/TYPE_REGISTRY/1?private=1"},
		} {
			if status, code, reads := fixture.serve(t, tc.method, tc.path, true, false); status != 403 || code != "authorization_denied" || reads != 0 {
				t.Fatalf("%s %s: %d %s reads=%d", tc.method, tc.path, status, code, reads)
			}
		}
		if fixture.store.slides != 0 || len(fixture.application.reads) != 0 {
			t.Fatalf("denial effects: slides=%d reads=%v", fixture.store.slides, fixture.application.reads)
		}
	})
	t.Run("an expired session is published to its sockets", func(t *testing.T) {
		fixture := newFamilyAdmissionFixture(t, true, true)
		if status, code, _ := fixture.serve(t, http.MethodGet, "/api/v1/reference-packs", true, false); status != 401 || code != "session_required" {
			t.Fatalf("expired: %d %s", status, code)
		}
		want := []string{familyTestSessionID.String() + ":session_expired"}
		if !slices.Equal(fixture.store.revoked, want) || !slices.Equal(fixture.hub.published, want) {
			t.Fatalf("revoked=%v published=%v, want %v", fixture.store.revoked, fixture.hub.published, want)
		}
	})
	t.Run("CSRF precedes the query, which precedes path parameters", func(t *testing.T) {
		fixture := newFamilyAdmissionFixture(t, true, false)
		for _, tc := range []struct {
			path   string
			csrf   bool
			status int
			code   string
		}{
			{"/api/v1/reference-packs/import?unknown=value", false, 403, "csrf_verification_failed"},
			{"/api/v1/reference-packs/import?unknown=value", true, 400, "invalid_reference_pack_request"},
			{"/api/v1/reference-packs/TYPE_REGISTRY/1/reverify?unknown=value", false, 403, "csrf_verification_failed"},
			{"/api/v1/reference-packs/TYPE_REGISTRY/1/reverify?unknown=value", true, 400, "invalid_reference_pack_request"},
			{"/api/v1/reference-packs/TYPE_REGISTRY/1/reverify", true, 404, "reference_pack_not_found"},
		} {
			if status, code, reads := fixture.serve(t, http.MethodPost, tc.path, true, tc.csrf); status != tc.status || code != tc.code || reads != 0 {
				t.Fatalf("POST %s csrf=%t: %d %s reads=%d", tc.path, tc.csrf, status, code, reads)
			}
		}
		if status, code, _ := fixture.serve(t, http.MethodGet, "/api/v1/reference-packs/validation-summaries/absent?unknown=value", true, false); status != 400 || code != "invalid_reference_pack_request" {
			t.Fatalf("summary query before path: %d %s", status, code)
		}
		if fixture.store.slides != 0 || len(fixture.application.reads) != 0 {
			t.Fatalf("rejection effects: slides=%d reads=%v", fixture.store.slides, fixture.application.reads)
		}
	})
	t.Run("admitted values reach the handler and success slides", func(t *testing.T) {
		fixture := newFamilyAdmissionFixture(t, true, false)
		for _, path := range []string{"/api/v1/reference-packs/type_registry.host/1", "/api/v1/reference-packs/type_registry.host/%61ctivate"} {
			if status, code, _ := fixture.serve(t, http.MethodGet, path, true, false); status != 404 || code != "reference_pack_not_found" {
				t.Fatalf("GET %s: %d %s", path, status, code)
			}
		}
		if want := []string{"type_registry.host/1", "type_registry.host/activate"}; !slices.Equal(fixture.application.reads, want) || fixture.store.slides != 0 {
			t.Fatalf("reads=%v slides=%d, want %v and no slide", fixture.application.reads, fixture.store.slides, want)
		}
		if status, _, _ := fixture.serve(t, http.MethodGet, "/api/v1/reference-packs", true, false); status != 200 || fixture.store.slides != 1 {
			t.Fatalf("list: %d slides=%d", status, fixture.store.slides)
		}
	})
}
