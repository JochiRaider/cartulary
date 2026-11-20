package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const (
	admissionThingID = "00000000-0000-4000-8000-0000000000c1"
	admissionOwner   = "module.fixture"
)

var (
	admissionReadSecurity  = [][]string{{"bearerSession"}, {"sessionCookie"}}
	admissionWriteSecurity = [][]string{{"bearerSession"}, {"csrfCookie", "csrfHeader", "sessionCookie"}}
	admissionThingParam    = PathParameter{Name: "thing_id", Format: "uuid", NotFoundCode: "thing_not_found", NotFoundStatus: http.StatusNotFound}
)

func admissionOperation(id, method, path string, security [][]string, parameters ...PathParameter) ContractOperation {
	return ContractOperation{
		OwnerID: admissionOwner, Method: method, PathTemplate: path, Pattern: method + " " + path, OperationID: id,
		Availability: BaseOperationAvailability, SuccessStatuses: []int{http.StatusOK}, Security: security,
		ErrorCodes: map[int][]string{
			http.StatusUnauthorized: {"session_required"}, http.StatusForbidden: {"csrf_verification_failed"},
			http.StatusNotFound: {"not_found", "thing_not_found"}, http.StatusConflict: {"credential_bootstrap_rejected", "thing_rejected"},
			http.StatusInternalServerError: {"internal_error"}, http.StatusServiceUnavailable: {"service_unavailable"},
		},
		PathParameters: parameters,
	}
}

func admissionOperations() []ContractOperation {
	return []ContractOperation{
		admissionOperation("getThing", http.MethodGet, "/api/v1/things/{thing_id}", admissionReadSecurity, admissionThingParam),
		admissionOperation("patchThing", http.MethodPatch, "/api/v1/things/{thing_id}", admissionWriteSecurity, admissionThingParam),
		admissionOperation("exportThings", http.MethodPost, "/api/v1/things/export", [][]string{}),
		admissionOperation("readThingVersion", http.MethodGet, "/api/v1/things/{thing_id}/versions/{version}", admissionReadSecurity, admissionThingParam,
			PathParameter{Name: "version", Pattern: "^v[1-9][0-9]*$", MaxLength: 8, NotFoundCode: "not_found", NotFoundStatus: http.StatusNotFound, NotFoundDetails: map[string]any{"validation_code": "version_not_found"}}),
		admissionOperation("verifyBlob", http.MethodPost, "/api/v1/blobs/{blob_id}/verify", admissionWriteSecurity,
			PathParameter{Name: "blob_id", Format: "uuid", NotFoundCode: "thing_rejected", NotFoundStatus: http.StatusConflict, NotFoundDetails: map[string]any{"reason_code": "blob_not_visible"}}),
		admissionOperation("bootstrapThing", http.MethodPost, "/api/v1/bootstrap", [][]string{{"bearerSession"}, {"credentialBootstrapBearer"}}),
		admissionOperation("getOwned", http.MethodGet, "/api/v1/owned/{owned_id}", admissionReadSecurity,
			PathParameter{Name: "owned_id", Format: "uuid", NotFoundCode: "thing_not_found", NotFoundStatus: http.StatusNotFound}),
	}
}

type admissionPrincipalKey struct{}

// fakeAuthenticator admits a request carrying X-Fixture-Auth: ok, observes
// the principal for the declared-security guard, and records each requirement.
// Its CSRF verification accepts X-Fixture-CSRF: ok. With events set, it logs
// "authenticate" and "csrf" in call order.
type fakeAuthenticator struct {
	requirements []SecurityRequirement
	observe      bool
	events       *[]string
}

func (authenticator *fakeAuthenticator) Authenticate(r *http.Request, requirement SecurityRequirement) (*http.Request, *APIError) {
	authenticator.requirements = append(authenticator.requirements, requirement)
	authenticator.log("authenticate")
	ObserveAuthenticationAttempt(r.Context())
	if r.Header.Get("X-Fixture-Auth") != "ok" {
		return r, SessionRequiredError()
	}
	if authenticator.observe {
		ObservePrincipal(r.Context(), AuthenticatedByBearerSession)
	}
	return r.WithContext(context.WithValue(r.Context(), admissionPrincipalKey{}, "principal")), nil
}

func (authenticator *fakeAuthenticator) VerifyCSRF(r *http.Request) *APIError {
	authenticator.log("csrf")
	ObserveCSRFVerification(r.Context())
	if r.Context().Value(admissionPrincipalKey{}) == nil {
		return InternalAPIError(nil)
	}
	if r.Header.Get("X-Fixture-CSRF") != "ok" {
		return &APIError{Status: http.StatusForbidden, Code: "csrf_verification_failed", Details: map[string]any{}}
	}
	return nil
}

func (authenticator *fakeAuthenticator) log(event string) {
	if authenticator.events != nil {
		*authenticator.events = append(*authenticator.events, event)
	}
}

type admissionFixture struct {
	registry      *RouteRegistry
	authenticator *fakeAuthenticator
	handler       http.Handler
	calls         map[string]int
	lastRequest   *http.Request
}

// newAdmissionFixture binds every fixture operation centrally, including the
// credential-bootstrap operation, and getOwned for owner admission, behind
// NewHandler.
func newAdmissionFixture(t *testing.T, options Options) *admissionFixture {
	t.Helper()
	registry, err := newRouteRegistry(admissionOperations(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fixture := &admissionFixture{registry: registry, authenticator: &fakeAuthenticator{observe: true}, calls: map[string]int{}}
	record := func(id string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fixture.calls[id]++
			fixture.lastRequest = r
			if r.Context().Value(admissionPrincipalKey{}) == nil && id != "exportThings" && id != "getOwned" {
				t.Errorf("%s ran without the admitted principal", id)
			}
			_ = WriteSuccess(w, r, http.StatusOK, map[string]any{"operation": id})
		}
	}
	options.Dependencies = testExtensionDependenciesWith(t, DependencySet{PublicRoutes: registry, Authenticator: fixture.authenticator}, nil)
	options.AdditionalRoutes = []RouteRegistrar{func(mux *http.ServeMux, deps DependencySet) error {
		if err := BindAdmittedOwnerRoutes(mux, deps, admissionOwner, map[string]http.HandlerFunc{
			"getThing": record("getThing"), "patchThing": record("patchThing"), "exportThings": record("exportThings"),
			"readThingVersion": record("readThingVersion"), "verifyBlob": record("verifyBlob"),
			"bootstrapThing": record("bootstrapThing"),
		}); err != nil {
			return err
		}
		if err := registry.Bind(mux, admissionOwner, "getOwned", record("getOwned")); err != nil {
			return err
		}
		return BindSocketRoute(mux, deps, "/ws/v1/things/{thing_id}", func(w http.ResponseWriter, r *http.Request) {
			fixture.calls["socket"]++
			_ = WriteSuccess(w, r, http.StatusOK, map[string]any{"operation": "socket"})
		})
	}}
	fixture.handler, err = NewHandler(options)
	if err != nil {
		t.Fatal(err)
	}
	return fixture
}

func (fixture *admissionFixture) serve(method, target string, authenticated bool) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, target, nil)
	if authenticated {
		request.Header.Set("X-Fixture-Auth", "ok")
	}
	recorder := httptest.NewRecorder()
	fixture.handler.ServeHTTP(recorder, request)
	return recorder
}

func requireAdmissionError(t *testing.T, recorder *httptest.ResponseRecorder, status int, code string, details map[string]any) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, status, recorder.Body)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Fatalf("content type = %q, want application/json", contentType)
	}
	var envelope ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope %q: %v", recorder.Body, err)
	}
	if envelope.Error.Code != code || envelope.Error.Status != status || envelope.Error.Message != code || envelope.Error.Retryable || envelope.Error.RequestID == "" {
		t.Fatalf("envelope = %#v, want %d %s", envelope.Error, status, code)
	}
	if !reflect.DeepEqual(envelope.Error.Details, details) {
		t.Fatalf("details = %#v, want %#v", envelope.Error.Details, details)
	}
}

func requireMethodNotAllowedEnvelope(t *testing.T, recorder *httptest.ResponseRecorder, allowed ...string) {
	t.Helper()
	if got := recorder.Header().Get("Allow"); got != strings.Join(allowed, ", ") {
		t.Fatalf("Allow = %q, want %q", got, strings.Join(allowed, ", "))
	}
	methods := make([]any, 0, len(allowed))
	for _, method := range allowed {
		methods = append(methods, method)
	}
	requireAdmissionError(t, recorder, http.StatusMethodNotAllowed, "method_not_allowed", map[string]any{"allowed_methods": methods})
}

// TestRouteResolutionRejectionsAreEnveloped covers Core 01 REQ-01-707 step 1
// and AC-617: under /api, a path no operation matches receives the enveloped
// 404 route_not_found and a path matched only under other methods, including
// HEAD on a GET path, the enveloped 405 method_not_allowed with an alphabetical
// Allow header and matching details. A non-canonical /api path is never
// redirected, an encoded slash stays inside its segment, a HEAD rejection
// carries headers only, paths outside /api keep their behaviour, and a claimed
// family envelopes its own rejections after its policy. Under /ws/v1, Core 01
// REQ-01-708 step 1 resolves a WebSocket route the same way before its handler:
// any method but GET, HEAD included, receives 405 with Allow: GET, and an
// unknown or non-canonical path 404 route_not_found.
func TestRouteResolutionRejectionsAreEnveloped(t *testing.T) {
	fixture := newAdmissionFixture(t, Options{})
	thing := "/api/v1/things/" + admissionThingID
	t.Run("unknown paths", func(t *testing.T) {
		for _, path := range []string{"/api", "/api/", "/api/v1", "/api/v1/nothing", thing + "/", thing + "/unknown", "/api/v1/things/" + admissionThingID + "/versions"} {
			requireAdmissionError(t, fixture.serve(http.MethodGet, path, true), http.StatusNotFound, "route_not_found", map[string]any{})
		}
	})
	t.Run("undeclared methods", func(t *testing.T) {
		requireMethodNotAllowedEnvelope(t, fixture.serve(http.MethodDelete, thing, true), "GET", "PATCH")
		requireMethodNotAllowedEnvelope(t, fixture.serve(http.MethodOptions, thing, true), "GET", "PATCH")
		// A literal segment and a parameter both match: every declared method is allowed.
		requireMethodNotAllowedEnvelope(t, fixture.serve(http.MethodPut, "/api/v1/things/export", true), "GET", "PATCH", "POST")
		requireMethodNotAllowedEnvelope(t, fixture.serve(http.MethodGet, "/api/v1/blobs/"+admissionThingID+"/verify", true), "POST")
		requireMethodNotAllowedEnvelope(t, fixture.serve(http.MethodPost, "/api/v1/owned/"+admissionThingID, true), "GET")
		// An encoded slash stays in the segment, so the path is the GET/PATCH template.
		requireMethodNotAllowedEnvelope(t, fixture.serve(http.MethodPost, thing+"%2Fexport", true), "GET", "PATCH")
		if fixture.calls["exportThings"] != 0 {
			t.Fatal("an encoded slash reached another operation")
		}
	})
	t.Run("HEAD carries headers only", func(t *testing.T) {
		for _, tc := range []struct {
			path   string
			status int
			allow  string
		}{
			{thing, http.StatusMethodNotAllowed, "GET, PATCH"},
			{"/api/v1/owned/" + admissionThingID, http.StatusMethodNotAllowed, "GET"},
			{"/api/v1/blobs/" + admissionThingID + "/verify", http.StatusMethodNotAllowed, "POST"},
			{"/api/v1/nothing", http.StatusNotFound, ""},
		} {
			recorder := fixture.serve(http.MethodHead, tc.path, true)
			if recorder.Code != tc.status || recorder.Header().Get("Allow") != tc.allow || recorder.Body.Len() != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q", tc.path, recorder.Code, recorder.Header().Get("Allow"), recorder.Body)
			}
		}
		if fixture.calls["getThing"] != 0 || fixture.calls["getOwned"] != 0 {
			t.Fatal("HEAD reached a GET operation")
		}
	})
	t.Run("non-canonical paths are never redirected", func(t *testing.T) {
		for _, path := range []string{"/api//v1/things/" + admissionThingID, "/api/v1/things/../things/" + admissionThingID, "/api/v1/./things/" + admissionThingID, "/api/v1/things/x/..", "/api/../healthz"} {
			for _, method := range []string{http.MethodGet, http.MethodPatch, http.MethodDelete} {
				recorder := fixture.serve(method, path, true)
				if recorder.Header().Get("Location") != "" {
					t.Fatalf("%s %s redirected to %s", method, path, recorder.Header().Get("Location"))
				}
				requireAdmissionError(t, recorder, http.StatusNotFound, "route_not_found", map[string]any{})
			}
		}
		if len(fixture.calls) != 0 {
			t.Fatalf("a non-canonical path reached an operation: %v", fixture.calls)
		}
	})
	t.Run("paths outside /api keep their behaviour", func(t *testing.T) {
		if recorder := fixture.serve(http.MethodGet, "/healthz", false); recorder.Code != http.StatusOK || !strings.HasPrefix(recorder.Header().Get("Content-Type"), "text/plain") {
			t.Fatalf("healthz = %d %q", recorder.Code, recorder.Header().Get("Content-Type"))
		}
		if recorder := fixture.serve(http.MethodGet, "/", false); recorder.Code != http.StatusOK || !strings.Contains(recorder.Header().Get("Content-Type"), "text/html") {
			t.Fatalf("root = %d %q", recorder.Code, recorder.Header().Get("Content-Type"))
		}
		if recorder := fixture.serve(http.MethodGet, "/apiary", false); recorder.Code != http.StatusNotFound || strings.Contains(recorder.Body.String(), "route_not_found") {
			t.Fatalf("a non-API path took the API envelope: %d %s", recorder.Code, recorder.Body)
		}
	})
	t.Run("a claimed family envelopes its rejections after its policy", testFamilyRejectionsAreEnveloped)
	t.Run("WebSocket routes under /ws/v1 resolve before their handler", func(t *testing.T) {
		socket := "/ws/v1/things/" + admissionThingID
		if recorder := fixture.serve(http.MethodGet, socket, false); recorder.Code != http.StatusOK || fixture.calls["socket"] != 1 {
			t.Fatalf("GET %s = %d, socket calls %d", socket, recorder.Code, fixture.calls["socket"])
		}
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions} {
			requireMethodNotAllowedEnvelope(t, fixture.serve(method, socket, false), "GET")
		}
		unknown := []string{
			"/ws/v1", "/ws/v1/", "/ws/v1/things", "/ws/v1/things/", socket + "/extra", "/ws/v1/other",
			"/ws/v1//things/" + admissionThingID, "/ws/v1/things/x/../" + admissionThingID, "/ws/v1/./things/" + admissionThingID,
		}
		for _, path := range unknown {
			for _, method := range []string{http.MethodGet, http.MethodPost} {
				recorder := fixture.serve(method, path, false)
				if recorder.Header().Get("Location") != "" {
					t.Fatalf("%s %s redirected to %s", method, path, recorder.Header().Get("Location"))
				}
				requireAdmissionError(t, recorder, http.StatusNotFound, "route_not_found", map[string]any{})
			}
		}
		for path, tc := range map[string]struct {
			status int
			allow  string
		}{socket: {http.StatusMethodNotAllowed, "GET"}, "/ws/v1/other": {http.StatusNotFound, ""}} {
			recorder := fixture.serve(http.MethodHead, path, false)
			if recorder.Code != tc.status || recorder.Header().Get("Allow") != tc.allow || recorder.Body.Len() != 0 {
				t.Fatalf("HEAD %s = %d Allow=%q body=%q", path, recorder.Code, recorder.Header().Get("Allow"), recorder.Body)
			}
		}
		if fixture.calls["socket"] != 1 {
			t.Fatalf("a rejected request reached the socket handler: %d calls", fixture.calls["socket"])
		}
		// Paths outside /ws/v1 keep the mux's behaviour.
		if recorder := fixture.serve(http.MethodGet, "/ws/v2/things/"+admissionThingID, false); recorder.Code != http.StatusNotFound || strings.Contains(recorder.Body.String(), "route_not_found") {
			t.Fatalf("a path outside /ws/v1 took the route envelope: %d %s", recorder.Code, recorder.Body)
		}
		handler := func(http.ResponseWriter, *http.Request) {}
		for _, path := range []string{"/api/v1/things", "ws/v1/things", "/ws/v1", "/ws/v1/things /x"} {
			if err := BindSocketRoute(http.NewServeMux(), DependencySet{}, path, handler); err == nil {
				t.Fatalf("BindSocketRoute accepted %q", path)
			}
		}
		// The route is recorded in the registry, once (REQ-01-708).
		registry, err := NewRouteRegistry(nil)
		if err != nil {
			t.Fatal(err)
		}
		deps := DependencySet{PublicRoutes: registry}
		if err := BindSocketRoute(http.NewServeMux(), deps, "/ws/v1/things/{thing_id}", handler); err != nil {
			t.Fatal(err)
		}
		if got := registry.SupplementaryRoutes(); !reflect.DeepEqual(got, []SupplementaryRoute{{Kind: SupplementarySocketRoute, Method: http.MethodGet, Path: "/ws/v1/things/{thing_id}"}}) {
			t.Fatalf("recorded socket routes = %#v", got)
		}
		if err := BindSocketRoute(http.NewServeMux(), deps, "/ws/v1/things/{thing_id}", handler); err == nil {
			t.Fatal("a second registration of the socket path was accepted")
		}
	})
}

func testFamilyRejectionsAreEnveloped(t *testing.T) {
	const root = "/api/v1/fixture-packs"
	const profile = "fixture_profile"
	operations := []ContractOperation{
		{OwnerID: admissionOwner, Method: "GET", PathTemplate: root + "/{version}", Pattern: "GET " + root + "/{version}", OperationID: "readFixture", Availability: profile, SuccessStatuses: []int{200}, Security: [][]string{{"sessionCookie"}}},
		{OwnerID: admissionOwner, Method: "POST", PathTemplate: root + "/{version}", Pattern: "POST " + root + "/{version}", OperationID: "inspectFixture", Availability: profile, SuccessStatuses: []int{200}, Security: [][]string{{"bearerSession"}}},
	}
	registry, err := newRouteRegistry(operations, []ExtensionClaim{{ProfileID: profile, Claimed: true}})
	if err != nil {
		t.Fatal(err)
	}
	deps := DependencySet{PublicRoutes: registry, Authenticator: &fakeAuthenticator{}, ExtensionRoutes: familyTestRoutes{{ProfileID: profile, RouteFamily: root, Claimed: true}}}
	calls := 0
	handler := func(w http.ResponseWriter, r *http.Request) { calls++ }
	if err := BindAdmittedOwnerFamilyRoutes(deps, OwnerFamily{
		OwnerID: admissionOwner, ProfileID: profile, Root: root, Authorize: fixtureFamilyPolicy,
		Handlers: map[string]http.HandlerFunc{"readFixture": handler, "inspectFixture": handler},
	}); err != nil {
		t.Fatal(err)
	}
	families, err := registry.withOwnerFamilies(http.NotFoundHandler(), extensionRoutesFromDependencies(deps))
	if err != nil {
		t.Fatal(err)
	}
	dispatch := withRequestID(families, nil)
	serve := func(method, path string, admitted bool) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, path, nil)
		request.Header.Set("X-Fixture-Auth", "ok")
		if admitted {
			request.Header.Set("X-Fixture-Admin", "ok")
		}
		recorder := httptest.NewRecorder()
		dispatch.ServeHTTP(recorder, request)
		return recorder
	}
	// The family gate still precedes every rejection (§17.4).
	for _, path := range []string{root, root + "/v1/unknown", root + "//v1"} {
		requireAdmissionError(t, serve(http.MethodGet, path, false), http.StatusForbidden, "authorization_denied", map[string]any{"required_capability": "fixture_admin"})
	}
	for _, path := range []string{root, root + "/v1/unknown", root + "/v1/a/b"} {
		requireAdmissionError(t, serve(http.MethodGet, path, true), http.StatusNotFound, "route_not_found", map[string]any{})
	}
	for _, path := range []string{root + "//v1", root + "/x/../v1", "/api/v1/other/../fixture-packs/v1"} {
		recorder := serve(http.MethodGet, path, true)
		if recorder.Header().Get("Location") != "" {
			t.Fatalf("%s redirected", path)
		}
		requireAdmissionError(t, recorder, http.StatusNotFound, "route_not_found", map[string]any{})
	}
	requireMethodNotAllowedEnvelope(t, serve(http.MethodPatch, root+"/v1", true), "GET", "POST")
	if recorder := serve(http.MethodHead, root+"/v1", true); recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != "GET, POST" || recorder.Body.Len() != 0 {
		t.Fatalf("family HEAD = %d Allow=%q body=%q", recorder.Code, recorder.Header().Get("Allow"), recorder.Body)
	}
	if calls != 0 {
		t.Fatal("a family rejection invoked an operation")
	}
	if modes := registry.AdmissionModes(); modes["readFixture"] != AdmissionCentral || modes["inspectFixture"] != AdmissionCentral {
		t.Fatalf("family admission modes = %v", modes)
	}
}

// fixtureFamilyPolicy admits a principal whose request carries
// X-Fixture-Admin: ok and otherwise denies the fixture_admin capability.
func fixtureFamilyPolicy(r *http.Request) *APIError {
	if r.Context().Value(admissionPrincipalKey{}) == nil {
		return InternalAPIError(nil)
	}
	if r.Header.Get("X-Fixture-Admin") != "ok" {
		return &APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_capability": "fixture_admin"}}
	}
	return nil
}

// TestCentralAdmissionOrderAndPathParameters covers Core 01 REQ-01-707 steps 2
// and 3 and AC-617 for a centrally admitted operation: authentication by the
// declared security precedes path parameters, so an unauthenticated malformed
// value receives session_required; an authenticated malformed value, including
// one with an encoded slash, receives the parameter's declared not-found
// response from the catalog projection (status, code and details) and never
// reaches the handler; typed values reach the handler; a public operation is
// never authenticated; an operation admitting a credential bootstrap bearer
// passes that alternative to the authenticator; binding refuses secured
// operations without an authenticator and invalid patterns; the
// declared-security and undeclared-code guards still observe a centrally
// admitted operation; and an owner family keeps its own order (§17.4).
func TestCentralAdmissionOrderAndPathParameters(t *testing.T) {
	thing := "/api/v1/things/" + admissionThingID
	t.Run("an owner family keeps its own admission order", testOwnerFamilyAdmissionOrder)
	t.Run("a no-store operation carries the header on every resolved response", func(t *testing.T) {
		// Core 01 REQ-01-144B and REQ-01-151.1: authentication, path-parameter
		// and handler responses carry no-store; route resolution's 405 belongs
		// to no operation, and another operation is unaffected.
		registry, err := newRouteRegistry(admissionOperations(), nil)
		if err != nil {
			t.Fatal(err)
		}
		authenticator := &fakeAuthenticator{}
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Fixture-Fail") == "ok" {
				WriteAPIError(w, r, NotFound("thing_not_found"))
				return
			}
			_ = WriteSuccess(w, r, http.StatusOK, map[string]any{})
		}
		options := Options{Dependencies: testExtensionDependenciesWith(t, DependencySet{PublicRoutes: registry, Authenticator: authenticator}, nil)}
		options.AdditionalRoutes = []RouteRegistrar{func(mux *http.ServeMux, deps DependencySet) error {
			return BindAdmittedOwnerRoutes(mux, deps, admissionOwner, map[string]http.HandlerFunc{"getThing": handler, "readThingVersion": handler}, NoStore("getThing"))
		}}
		served, err := NewHandler(options)
		if err != nil {
			t.Fatal(err)
		}
		serve := func(method, target string, headers ...string) *httptest.ResponseRecorder {
			request := httptest.NewRequest(method, target, nil)
			for _, header := range headers {
				request.Header.Set(header, "ok")
			}
			recorder := httptest.NewRecorder()
			served.ServeHTTP(recorder, request)
			return recorder
		}
		for _, tc := range []struct {
			name, method, target string
			headers              []string
			status               int
			cacheControl         string
		}{
			{"unauthenticated", http.MethodGet, thing, nil, http.StatusUnauthorized, "no-store"},
			{"malformed identifier", http.MethodGet, "/api/v1/things/malformed", []string{"X-Fixture-Auth"}, http.StatusNotFound, "no-store"},
			{"handler failure", http.MethodGet, thing, []string{"X-Fixture-Auth", "X-Fixture-Fail"}, http.StatusNotFound, "no-store"},
			{"success", http.MethodGet, thing, []string{"X-Fixture-Auth"}, http.StatusOK, "no-store"},
			{"route resolution 405", http.MethodHead, thing, []string{"X-Fixture-Auth"}, http.StatusMethodNotAllowed, ""},
			{"another operation", http.MethodGet, thing + "/versions/v1", []string{"X-Fixture-Auth"}, http.StatusOK, ""},
		} {
			recorder := serve(tc.method, tc.target, tc.headers...)
			if recorder.Code != tc.status || recorder.Header().Get("Cache-Control") != tc.cacheControl {
				t.Fatalf("%s = %d Cache-Control %q, want %d %q", tc.name, recorder.Code, recorder.Header().Get("Cache-Control"), tc.status, tc.cacheControl)
			}
		}
		err = BindAdmittedOwnerRoutes(http.NewServeMux(), DependencySet{PublicRoutes: registry, Authenticator: authenticator}, admissionOwner, map[string]http.HandlerFunc{"patchThing": handler}, NoStore("getThing"))
		if err == nil || !strings.Contains(err.Error(), "without binding it") {
			t.Fatalf("an unbound no-store operation was accepted: %v", err)
		}
	})
	t.Run("authentication precedes path parameters", func(t *testing.T) {
		fixture := newAdmissionFixture(t, Options{})
		for _, path := range []string{"/api/v1/things/malformed", "/api/v1/things/a%2Fb", thing} {
			requireAdmissionError(t, fixture.serve(http.MethodGet, path, false), http.StatusUnauthorized, "session_required", map[string]any{})
		}
		requireAdmissionError(t, fixture.serve(http.MethodPost, "/api/v1/blobs/malformed/verify", false), http.StatusUnauthorized, "session_required", map[string]any{})
		if len(fixture.calls) != 0 {
			t.Fatalf("an unauthenticated request reached an operation: %v", fixture.calls)
		}
	})
	t.Run("every security shape", func(t *testing.T) {
		fixture := newAdmissionFixture(t, Options{})
		for _, tc := range []struct {
			method, path string
			want         []SecurityRequirement
		}{
			{http.MethodGet, thing, []SecurityRequirement{{}}},
			{http.MethodPatch, thing, []SecurityRequirement{{CookieCSRF: true}}},
			{http.MethodPost, "/api/v1/blobs/" + admissionThingID + "/verify", []SecurityRequirement{{CookieCSRF: true}}},
			{http.MethodPost, "/api/v1/things/export", nil},
			// The bootstrap alternative reaches the authenticator, which decides it.
			{http.MethodPost, "/api/v1/bootstrap", []SecurityRequirement{{Bootstrap: true}}},
			// Owner admission authenticates in the handler, not centrally.
			{http.MethodGet, "/api/v1/owned/" + admissionThingID, nil},
		} {
			fixture.authenticator.requirements = nil
			if recorder := fixture.serve(tc.method, tc.path, true); recorder.Code != http.StatusOK {
				t.Fatalf("%s %s = %d %s", tc.method, tc.path, recorder.Code, recorder.Body)
			}
			if !reflect.DeepEqual(fixture.authenticator.requirements, tc.want) {
				t.Fatalf("%s %s authenticated with %#v, want %#v", tc.method, tc.path, fixture.authenticator.requirements, tc.want)
			}
		}
		if modes := fixture.registry.AdmissionModes(); modes["getThing"] != AdmissionCentral || modes["exportThings"] != AdmissionCentral || modes["getOwned"] != AdmissionOwner || modes["bootstrapThing"] != AdmissionCentral || len(modes) != 7 {
			t.Fatalf("admission modes = %v", modes)
		}
	})
	t.Run("a malformed value receives the declared not-found response", func(t *testing.T) {
		fixture := newAdmissionFixture(t, Options{})
		for _, value := range []string{"malformed", "a%2Fb", admissionThingID + "%2Fexport", strings.ReplaceAll(admissionThingID, "-", ""), "%7B" + admissionThingID + "%7D", "urn:uuid:" + admissionThingID, admissionThingID + "0", "%FF"} {
			requireAdmissionError(t, fixture.serve(http.MethodGet, "/api/v1/things/"+value, true), http.StatusNotFound, "thing_not_found", map[string]any{})
			requireAdmissionError(t, fixture.serve(http.MethodPatch, "/api/v1/things/"+value, true), http.StatusNotFound, "thing_not_found", map[string]any{})
		}
		// The first parameter in template order decides.
		requireAdmissionError(t, fixture.serve(http.MethodGet, "/api/v1/things/malformed/versions/v0", true), http.StatusNotFound, "thing_not_found", map[string]any{})
		for _, version := range []string{"v0", "1", "v123456789", "v1%2F2", "%C3%A91"} {
			requireAdmissionError(t, fixture.serve(http.MethodGet, thing+"/versions/"+version, true), http.StatusNotFound, "not_found", map[string]any{"validation_code": "version_not_found"})
		}
		// A non-404 not-found response comes from the projection, with its details.
		requireAdmissionError(t, fixture.serve(http.MethodPost, "/api/v1/blobs/malformed/verify", true), http.StatusConflict, "thing_rejected", map[string]any{"reason_code": "blob_not_visible"})
		if len(fixture.calls) != 0 {
			t.Fatalf("a malformed value reached an operation: %v", fixture.calls)
		}
	})
	t.Run("typed values reach the handler", func(t *testing.T) {
		fixture := newAdmissionFixture(t, Options{})
		if recorder := fixture.serve(http.MethodGet, "/api/v1/things/"+strings.ToUpper(admissionThingID)+"/versions/v12", true); recorder.Code != http.StatusOK {
			t.Fatalf("status = %d %s", recorder.Code, recorder.Body)
		}
		if value, ok := PathUUID(fixture.lastRequest, "thing_id"); !ok || value != uuid.MustParse(admissionThingID) {
			t.Fatalf("PathUUID = %v %t", value, ok)
		}
		if value, ok := PathString(fixture.lastRequest, "version"); !ok || value != "v12" {
			t.Fatalf("PathString = %q %t", value, ok)
		}
		if _, ok := PathString(fixture.lastRequest, "thing_id"); ok {
			t.Fatal("a uuid parameter was readable as a string")
		}
		if _, ok := PathUUID(fixture.lastRequest, "version"); ok {
			t.Fatal("a pattern parameter was readable as a uuid")
		}
		if recorder := fixture.serve(http.MethodGet, "/api/v1/owned/"+admissionThingID, true); recorder.Code != http.StatusOK {
			t.Fatalf("owner-admitted status = %d", recorder.Code)
		}
		if _, ok := PathUUID(fixture.lastRequest, "owned_id"); ok {
			t.Fatal("an owner-admitted operation received centrally admitted values")
		}
	})
	t.Run("binding refuses unsupported central admission", func(t *testing.T) {
		registry, err := newRouteRegistry(append(admissionOperations(),
			admissionOperation("badPattern", http.MethodGet, "/api/v1/bad/{bad_id}", admissionReadSecurity,
				PathParameter{Name: "bad_id", Pattern: "^(a$", NotFoundCode: "thing_not_found", NotFoundStatus: http.StatusNotFound})), nil)
		if err != nil {
			t.Fatal(err)
		}
		mux := http.NewServeMux()
		handler := func(http.ResponseWriter, *http.Request) {}
		for _, tc := range []struct {
			operationID   string
			authenticator Authenticator
			wantError     string
		}{
			{"getThing", nil, "requires an authenticator for central admission"},
			{"badPattern", &fakeAuthenticator{}, "path parameter \"bad_id\" pattern"},
		} {
			if err := registry.BindAdmitted(mux, admissionOwner, tc.operationID, handler, tc.authenticator); err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("%s: bind error = %v, want %q", tc.operationID, err, tc.wantError)
			}
		}
		if err := registry.BindAdmitted(mux, admissionOwner, "exportThings", handler, nil); err != nil {
			t.Fatalf("a public operation needs no authenticator: %v", err)
		}
		if modes := registry.AdmissionModes(); len(modes) != 1 || modes["exportThings"] != AdmissionCentral {
			t.Fatalf("refused bindings were recorded: %v", modes)
		}
	})
	t.Run("the guards still observe a centrally admitted operation", func(t *testing.T) {
		var mismatches []SecurityMismatch
		var undeclared []string
		fixture := newAdmissionFixture(t, Options{
			SecurityGuard: &SecurityGuard{Enforce: true, Report: func(mismatch SecurityMismatch) { mismatches = append(mismatches, mismatch) }},
			ErrorCodeGuard: &ErrorCodeGuard{Enforce: true, Report: func(operationID string, status int, code string) {
				undeclared = append(undeclared, operationID+" "+code)
			}},
		})
		for _, tc := range []struct {
			method, path  string
			authenticated bool
			status        int
		}{
			{http.MethodGet, thing, true, http.StatusOK},
			{http.MethodPost, "/api/v1/things/export", false, http.StatusOK},
			{http.MethodGet, thing, false, http.StatusUnauthorized},
			{http.MethodGet, "/api/v1/things/malformed", true, http.StatusNotFound},
			{http.MethodPost, "/api/v1/blobs/malformed/verify", true, http.StatusConflict},
			{http.MethodGet, "/api/v1/nothing", true, http.StatusNotFound},
			{http.MethodDelete, thing, true, http.StatusMethodNotAllowed},
		} {
			if recorder := fixture.serve(tc.method, tc.path, tc.authenticated); recorder.Code != tc.status {
				t.Fatalf("%s %s = %d, want %d: %s", tc.method, tc.path, recorder.Code, tc.status, recorder.Body)
			}
		}
		if len(mismatches) != 0 || len(undeclared) != 0 {
			t.Fatalf("guards reported mismatches=%v undeclared=%v", mismatches, undeclared)
		}
		// Control: an authenticator that establishes no principal is still caught
		// by the declared-security backstop.
		fixture.authenticator.observe = false
		if recorder := fixture.serve(http.MethodGet, thing, true); recorder.Code != http.StatusInternalServerError {
			t.Fatalf("unobserved principal status = %d, want the guard's 500", recorder.Code)
		}
		if len(mismatches) != 1 || mismatches[0].OperationID != "getThing" {
			t.Fatalf("mismatches = %v, want one for getThing", mismatches)
		}
	})
}
