package reference_data

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

type admissionSessionStub struct {
	admin           bool
	lookups, slides int
}

func (s *admissionSessionStub) GetSessionByFingerprint(context.Context, []byte) (authn.SessionRecord, authn.UserRecord, error) {
	s.lookups++
	return authn.SessionRecord{SessionExpiresAt: time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)}, authn.UserRecord{IsActive: true, IsDeploymentAdmin: s.admin}, nil
}
func (*admissionSessionStub) GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error) {
	return authn.BootstrapTokenRecord{}, authn.UserRecord{}, authn.ErrNotFound
}
func (*admissionSessionStub) RevokeSession(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}
func (s *admissionSessionStub) SlideSession(_ context.Context, _ uuid.UUID, timing authn.SessionTiming) (authn.SessionTiming, error) {
	s.slides++
	return timing, nil
}

type admissionBodyProbe struct{ reads int }

func (body *admissionBodyProbe) Read([]byte) (int, error) { body.reads++; return 0, io.EOF }

func testFamilyDenialEffects(t *testing.T) {
	for _, authenticated := range []bool{false, true} {
		store := &admissionSessionStub{}
		service := &httpAdapter{authStore: store}
		body := &admissionBodyProbe{}
		r := httptest.NewRequest("POST", "/api/v1/reference-packs/import?limit=1", body)
		if authenticated {
			r.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: "fixture-session"})
		}
		called := false
		w := httptest.NewRecorder()
		service.authorizeFamily(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(w, r)
		want := 401
		if authenticated {
			want = 403
		}
		if w.Code != want || called || body.reads != 0 || store.slides != 0 {
			t.Fatalf("denial effects: status=%d called=%t reads=%d slides=%d", w.Code, called, body.reads, store.slides)
		}
		if authenticated && store.lookups != 1 {
			t.Fatalf("session lookups=%d", store.lookups)
		}
	}
	store := &admissionSessionStub{admin: true}
	service := &httpAdapter{authStore: store}
	r := httptest.NewRequest("POST", "/api/v1/reference-packs/import", nil)
	r.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: "fixture-session"})
	w := httptest.NewRecorder()
	service.authorizeFamily(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, apiErr := service.authorizedPrincipal(r)
		if apiErr != nil {
			t.Fatal(apiErr)
		}
		if httpauth.ValidatePrincipalCSRF(r, principal, service.keys) == nil {
			t.Fatal("missing CSRF accepted")
		}
		proof := authn.CSRFTokenForSessionToken(service.keys, principal.SessionToken)
		r.AddCookie(&http.Cookie{Name: authn.CSRFCookieName, Value: proof})
		r.Header.Set(authn.CSRFHeaderName, proof)
		if apiErr := httpauth.ValidatePrincipalCSRF(r, principal, service.keys); apiErr != nil {
			t.Fatal(apiErr)
		}
		r.Header.Set(authn.CSRFHeaderName, "wrong")
		if httpauth.ValidatePrincipalCSRF(r, principal, service.keys) == nil {
			t.Fatal("incorrect CSRF accepted")
		}
		principal.AuthSource = httpauth.AuthSourceBearer
		if apiErr := httpauth.ValidatePrincipalCSRF(r, principal, service.keys); apiErr != nil {
			t.Fatal(apiErr)
		}
		w.WriteHeader(204)
	})).ServeHTTP(w, r)
	if w.Code != 204 || store.lookups != 1 {
		t.Fatal("admission must authenticate exactly once")
	}
	if _, apiErr := service.authorizedPrincipal(httptest.NewRequest("GET", "/", nil)); apiErr == nil {
		t.Fatal("missing admission accepted")
	}
	// Exercise the real operation security declaration. CSRF precedes query
	// admission; both stop the request before any body or application work.
	registry, err := httpapi.NewRouteRegistry([]httpapi.ExtensionClaim{{ProfileID: ProfileID, Claimed: true}})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	called := false
	if err := registry.Bind(mux, "module.reference_data", "importReferencePack", service.authorizeOperation(func(http.ResponseWriter, *http.Request) { called = true })); err != nil {
		t.Fatal(err)
	}
	for _, csrf := range []bool{false, true} {
		body := &admissionBodyProbe{}
		r := httptest.NewRequest("POST", "/api/v1/reference-packs/import?unknown=value", body)
		r.AddCookie(&http.Cookie{Name: authn.SessionCookieName, Value: "fixture-session"})
		if csrf {
			proof := authn.CSRFTokenForSessionToken(service.keys, "fixture-session")
			r.AddCookie(&http.Cookie{Name: authn.CSRFCookieName, Value: proof})
			r.Header.Set(authn.CSRFHeaderName, proof)
		}
		w := httptest.NewRecorder()
		service.authorizeFamily(mux).ServeHTTP(w, r)
		want := 403
		if csrf {
			want = 400
		}
		if w.Code != want || called || body.reads != 0 || store.slides != 0 {
			t.Fatalf("operation denial: status=%d called=%t reads=%d slides=%d", w.Code, called, body.reads, store.slides)
		}
	}
}
