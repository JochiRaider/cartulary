package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type familyTestRoutes []ExtensionRoute

func (routes familyTestRoutes) ExtensionRoutes() []ExtensionRoute { return routes }

func testOwnerFamilyRegistration(t *testing.T) {
	const root = "/api/v1/fixture-packs"
	const owner = "module.fixture"
	const profile = "fixture_profile"
	operations := []ContractOperation{
		{OwnerID: owner, Method: "GET", PathTemplate: root + "/{version}", Pattern: "GET " + root + "/{version}", OperationID: "readFixture", Availability: profile, SuccessStatuses: []int{200}, Security: [][]string{{"sessionCookie"}}},
		// CSRF is an independent security declaration, even on a read operation.
		{OwnerID: owner, Method: "POST", PathTemplate: root + "/{version}", Pattern: "POST " + root + "/{version}", OperationID: "inspectFixture", Availability: profile, SuccessStatuses: []int{200}, Security: [][]string{{"sessionCookie", "csrfCookie", "csrfHeader"}, {"bearerSession"}}},
	}
	newDeps := func(extra ...ContractOperation) DependencySet {
		registry, err := newRouteRegistry(append(append([]ContractOperation(nil), operations...), extra...), []ExtensionClaim{{ProfileID: profile, Claimed: true}})
		if err != nil {
			t.Fatal(err)
		}
		return DependencySet{PublicRoutes: registry, Authenticator: &fakeAuthenticator{}, ExtensionRoutes: familyTestRoutes{{ProfileID: profile, RouteFamily: root, Claimed: true}}}
	}
	calls := 0
	handler := func(w http.ResponseWriter, r *http.Request) {
		calls++
		metadata, ok := OperationMetadataFromContext(r.Context())
		if !ok || metadata.OwnerID != owner || metadata.CookieCSRFRequired != (r.Method == "POST") {
			t.Fatal("lost operation security metadata")
		}
		_, _ = w.Write([]byte(r.PathValue("version")))
	}
	handlers := map[string]http.HandlerFunc{"readFixture": handler, "inspectFixture": handler}
	family := func(handlers map[string]http.HandlerFunc) OwnerFamily {
		return OwnerFamily{OwnerID: owner, ProfileID: profile, Root: root, Authorize: fixtureFamilyPolicy, Handlers: handlers}
	}
	deps := newDeps()
	if err := BindAdmittedOwnerFamilyRoutes(deps, family(handlers)); err != nil {
		t.Fatal(err)
	}
	if err := deps.PublicRoutes.ValidateActive(); err != nil {
		t.Fatal(err)
	}
	if err := BindAdmittedOwnerFamilyRoutes(deps, family(handlers)); err == nil || err.Error() != fmt.Sprintf("owner family %q overlaps a bound family", root) {
		t.Fatalf("duplicate family error = %v", err)
	}
	if _, err := deps.PublicRoutes.withOwnerFamilies(http.NotFoundHandler(), nil); err == nil || err.Error() != "bound owner family is absent from the claimed route projection" {
		t.Fatalf("lost claimed projection error = %v", err)
	}
	dispatch, err := deps.PublicRoutes.withOwnerFamilies(http.NotFoundHandler(), extensionRoutesFromDependencies(deps))
	if err != nil {
		t.Fatal(err)
	}
	admitted := func(r *http.Request) *http.Request {
		r.Header.Set("X-Fixture-Auth", "ok")
		r.Header.Set("X-Fixture-Admin", "ok")
		r.Header.Set("X-Fixture-CSRF", "ok")
		return r
	}
	noncanonical := []string{
		"/api//v1/fixture-packs/activate",
		"/api/./v1/fixture-packs/activate",
		"/api/v1/other/../fixture-packs/activate",
		root + "//%61ctivate",
		root + "/../other",
	}
	for _, path := range append([]string{root, root + "/activate", root + "/unknown/descendant"}, noncanonical...) {
		for _, method := range []string{"GET", "HEAD", "OPTIONS", "PATCH", "POST"} {
			r := httptest.NewRequest(method, path, nil)
			r.Header.Set("X-Fixture-Auth", "ok")
			w := httptest.NewRecorder()
			dispatch.ServeHTTP(w, r)
			if w.Code != 403 {
				t.Fatalf("denial lost to routing: %s %s = %d", method, path, w.Code)
			}
		}
	}
	if calls != 0 {
		t.Fatal("denial invoked an operation")
	}
	for _, path := range noncanonical {
		for _, method := range []string{"GET", "POST", "HEAD", "OPTIONS", "PATCH"} {
			w := httptest.NewRecorder()
			dispatch.ServeHTTP(w, admitted(httptest.NewRequest(method, path, nil)))
			if w.Code != 404 || w.Header().Get("Location") != "" || calls != 0 {
				t.Fatalf("noncanonical path was redirected/dispatched: %s %s = %d", method, path, w.Code)
			}
		}
	}
	for _, method := range []string{"GET", "POST", "HEAD", "OPTIONS", "PATCH"} {
		w := httptest.NewRecorder()
		dispatch.ServeHTTP(w, admitted(httptest.NewRequest(method, root+"/%61ctivate", nil)))
		want := 405
		if method == "GET" || method == "POST" {
			want = 200
		}
		if w.Code != want || want == 405 && w.Header().Get("Allow") != "GET, POST" {
			t.Fatalf("method %s: %d %q", method, w.Code, w.Header().Get("Allow"))
		}
		if want == 200 && w.Body.String() != "activate" {
			t.Fatalf("wrong exact version %q", w.Body.String())
		}
	}
	for encoded, decoded := range map[string]string{"a%2Fb": "a/b", "%2e%2e": "..", "%2561ctivate": "%61ctivate"} {
		r := admitted(httptest.NewRequest("GET", root+"/"+encoded, nil))
		w := httptest.NewRecorder()
		dispatch.ServeHTTP(w, r)
		if w.Code != 200 || w.Body.String() != decoded || r.URL.EscapedPath() != root+"/"+encoded {
			t.Fatalf("encoded segment identity changed: %s = %d %q", encoded, w.Code, w.Body.String())
		}
	}
	for _, path := range []string{"/api%2Fv1/fixture-packs/activate", "/api/v1/other/%2e%2e/fixture-packs/activate", root + "-other"} {
		w := httptest.NewRecorder()
		dispatch.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("policy crossed a literal segment boundary: %s = %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	dispatch.ServeHTTP(w, httptest.NewRequest("GET", root+"-other", nil))
	if w.Code != 404 {
		t.Fatal("policy escaped its family boundary")
	}
	bootstrapRead := ContractOperation{OwnerID: owner, Method: "GET", PathTemplate: root + "/{version}/setup", Pattern: "GET " + root + "/{version}/setup", OperationID: "setupFixture", Availability: profile, SuccessStatuses: []int{200}, Security: [][]string{{"bearerSession"}, {"credentialBootstrapBearer"}}}
	anonymousRead := ContractOperation{OwnerID: owner, Method: "GET", PathTemplate: root + "/{version}/public", Pattern: "GET " + root + "/{version}/public", OperationID: "publicFixture", Availability: profile, SuccessStatuses: []int{200}, Security: [][]string{{"sessionCookie"}, {}}}
	badPattern := ContractOperation{OwnerID: owner, Method: "GET", PathTemplate: root + "/{version}/bad", Pattern: "GET " + root + "/{version}/bad", OperationID: "badFixture", Availability: profile, SuccessStatuses: []int{200}, Security: [][]string{{"sessionCookie"}},
		PathParameters: []PathParameter{{Name: "version", Pattern: "^(a$", NotFoundCode: "fixture_not_found", NotFoundStatus: http.StatusNotFound}}}
	withExtra := func(id string) map[string]http.HandlerFunc {
		return map[string]http.HandlerFunc{"readFixture": handler, "inspectFixture": handler, id: handler}
	}
	for _, bad := range []struct {
		name   string
		extra  []ContractOperation
		mutate func(*DependencySet, *OwnerFamily)
		want   string
	}{
		{"nil policy", nil, func(_ *DependencySet, family *OwnerFamily) { family.Authorize = nil }, "requires a route registry, root and policy"},
		{"no authenticator", nil, func(deps *DependencySet, _ *OwnerFamily) { deps.Authenticator = nil }, "requires an authenticator"},
		{"wrong owner", nil, func(_ *DependencySet, family *OwnerFamily) { family.OwnerID = "module.other" }, "overlaps another owner"},
		{"wrong profile", nil, func(_ *DependencySet, family *OwnerFamily) { family.ProfileID = "other" }, "not an admitted claimed route"},
		{"wrong root", nil, func(_ *DependencySet, family *OwnerFamily) { family.Root = "/api/v1/other" }, "not an admitted claimed route"},
		{"missing operation", nil, func(_ *DependencySet, family *OwnerFamily) {
			family.Handlers = map[string]http.HandlerFunc{"readFixture": handler}
		}, "complete operation catalog"},
		{"unknown operation", nil, func(_ *DependencySet, family *OwnerFamily) {
			family.Handlers = map[string]http.HandlerFunc{"readFixture": handler, "other": handler}
		}, "unknown or nil operation"},
		{"nil operation", nil, func(_ *DependencySet, family *OwnerFamily) {
			family.Handlers = map[string]http.HandlerFunc{"readFixture": handler, "inspectFixture": nil}
		}, "unknown or nil operation"},
		// The gate admits only a session, so no operation may admit a
		// bootstrap token or an anonymous caller.
		{"bootstrap operation", []ContractOperation{bootstrapRead}, func(_ *DependencySet, family *OwnerFamily) { family.Handlers = withExtra("setupFixture") }, "must require a session and no credential bootstrap"},
		{"anonymous operation", []ContractOperation{anonymousRead}, func(_ *DependencySet, family *OwnerFamily) { family.Handlers = withExtra("publicFixture") }, "must require a session and no credential bootstrap"},
		{"invalid pattern", []ContractOperation{badPattern}, func(_ *DependencySet, family *OwnerFamily) { family.Handlers = withExtra("badFixture") }, "path parameter \"version\" pattern"},
	} {
		t.Run(bad.name, func(t *testing.T) {
			deps := newDeps(bad.extra...)
			candidate := family(handlers)
			bad.mutate(&deps, &candidate)
			if err := BindAdmittedOwnerFamilyRoutes(deps, candidate); err == nil || !strings.Contains(err.Error(), bad.want) {
				t.Fatalf("bind error = %v, want %q", err, bad.want)
			}
			if len(deps.PublicRoutes.Snapshot()) != 0 || len(deps.PublicRoutes.AdmissionModes()) != 0 {
				t.Fatal("rejected registration partially published operations")
			}
		})
	}
}

// testOwnerFamilyAdmissionOrder covers an owner family that keeps its own
// admission order (Core 01 REQ-01-707, §17.4): the gate authenticates a
// session without CSRF or bootstrap, then the owner's policy runs, both
// before path canonicality, unknown descendants and method selection, which
// keep the enveloped step-1 rejections; the selected operation then verifies
// its declared cookie CSRF, runs the owner's query admission and admits its
// path parameters, in that order, before its handler; and the declared-
// security guard observes the gate's principal for the selected operation.
func testOwnerFamilyAdmissionOrder(t *testing.T) {
	const root = "/api/v1/fixture-packs"
	const profile = "fixture_profile"
	parameters := []PathParameter{
		{Name: "key", Pattern: "^[a-z]+$", NotFoundCode: "fixture_not_found", NotFoundStatus: http.StatusNotFound},
		{Name: "version", Pattern: "^v[0-9]+$", NotFoundCode: "fixture_not_found", NotFoundStatus: http.StatusNotFound},
	}
	operation := func(id, method, path string, security [][]string, parameters ...PathParameter) ContractOperation {
		return ContractOperation{
			OwnerID: admissionOwner, Method: method, PathTemplate: path, Pattern: method + " " + path, OperationID: id,
			Availability: profile, SuccessStatuses: []int{http.StatusOK}, Security: security, PathParameters: parameters,
			ErrorCodes: map[int][]string{
				http.StatusBadRequest: {"invalid_fixture_query"}, http.StatusUnauthorized: {"session_required"},
				http.StatusForbidden: {"authorization_denied", "csrf_verification_failed"}, http.StatusNotFound: {"fixture_not_found"},
			},
		}
	}
	operations := []ContractOperation{
		operation("listFixtures", http.MethodGet, root, admissionReadSecurity),
		operation("readFixture", http.MethodGet, root+"/{key}/{version}", admissionReadSecurity, parameters...),
		operation("actFixture", http.MethodPost, root+"/{key}/{version}/act", admissionWriteSecurity, parameters...),
	}
	type fixture struct {
		registry      *RouteRegistry
		authenticator *fakeAuthenticator
		events        []string
		values        []string
		handler       http.Handler
	}
	newFixture := func(t *testing.T, guard *SecurityGuard) *fixture {
		t.Helper()
		registry, err := newRouteRegistry(operations, []ExtensionClaim{{ProfileID: profile, Claimed: true}})
		if err != nil {
			t.Fatal(err)
		}
		f := &fixture{registry: registry}
		f.authenticator = &fakeAuthenticator{observe: true, events: &f.events}
		record := func(id string) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				f.events = append(f.events, "handle:"+id)
				key, _ := PathString(r, "key")
				version, _ := PathString(r, "version")
				f.values = append(f.values, key+"/"+version)
				_ = WriteSuccess(w, r, http.StatusOK, map[string]any{"operation": id})
			}
		}
		deps := DependencySet{PublicRoutes: registry, Authenticator: f.authenticator, ExtensionRoutes: familyTestRoutes{{ProfileID: profile, RouteFamily: root, Claimed: true}}}
		if err := BindAdmittedOwnerFamilyRoutes(deps, OwnerFamily{
			OwnerID: admissionOwner, ProfileID: profile, Root: root,
			Authorize: func(r *http.Request) *APIError {
				f.events = append(f.events, "authorize")
				return fixtureFamilyPolicy(r)
			},
			AdmitQuery: func(r *http.Request, operationID string) *APIError {
				f.events = append(f.events, "query:"+operationID)
				if operationID != "listFixtures" && r.URL.RawQuery != "" {
					return &APIError{Status: http.StatusBadRequest, Code: "invalid_fixture_query", Details: map[string]any{}}
				}
				return nil
			},
			Handlers: map[string]http.HandlerFunc{"listFixtures": record("listFixtures"), "readFixture": record("readFixture"), "actFixture": record("actFixture")},
		}); err != nil {
			t.Fatal(err)
		}
		families, err := registry.withOwnerFamilies(http.NotFoundHandler(), extensionRoutesFromDependencies(deps))
		if err != nil {
			t.Fatal(err)
		}
		f.handler = withRequestID(withSecurityGuard(families, guard), nil)
		return f
	}
	// serve sends a request as anonymous, "member" (authenticated, not
	// admitted by the policy) or "admin", with or without the CSRF proof.
	serve := func(f *fixture, method, target, caller string, csrf bool) *httptest.ResponseRecorder {
		f.events, f.values = nil, nil
		r := httptest.NewRequest(method, target, nil)
		if caller != "anonymous" {
			r.Header.Set("X-Fixture-Auth", "ok")
		}
		if caller == "admin" {
			r.Header.Set("X-Fixture-Admin", "ok")
		}
		if csrf {
			r.Header.Set("X-Fixture-CSRF", "ok")
		}
		recorder := httptest.NewRecorder()
		f.handler.ServeHTTP(recorder, r)
		return recorder
	}
	requireEvents := func(t *testing.T, f *fixture, want ...string) {
		t.Helper()
		if !reflect.DeepEqual(f.events, want) {
			t.Fatalf("events = %v, want %v", f.events, want)
		}
	}
	routingCases := []struct{ method, path string }{
		{http.MethodGet, root + "//a/v1"},
		{http.MethodPost, "/api/v1/other/../fixture-packs/a/v1/act"},
		{http.MethodGet, root + "/unknown/x/y/z"},
		{http.MethodPatch, root},
		{http.MethodPost, root + "/A/v1/act?bad=1"},
	}
	t.Run("authentication precedes the policy and routing", func(t *testing.T) {
		f := newFixture(t, nil)
		for _, tc := range routingCases {
			requireAdmissionError(t, serve(f, tc.method, tc.path, "anonymous", false), http.StatusUnauthorized, "session_required", map[string]any{})
			requireEvents(t, f, "authenticate")
		}
		// The gate authenticates a session without CSRF or a bootstrap
		// alternative; each operation verifies its own CSRF later.
		for _, requirement := range f.authenticator.requirements {
			if requirement != (SecurityRequirement{}) {
				t.Fatalf("gate requirement = %#v", requirement)
			}
		}
	})
	t.Run("the policy precedes routing and CSRF", func(t *testing.T) {
		f := newFixture(t, nil)
		for _, tc := range routingCases {
			requireAdmissionError(t, serve(f, tc.method, tc.path, "member", false), http.StatusForbidden, "authorization_denied", map[string]any{"required_capability": "fixture_admin"})
			requireEvents(t, f, "authenticate", "authorize")
		}
		if recorder := serve(f, http.MethodHead, root, "member", false); recorder.Code != http.StatusForbidden {
			t.Fatalf("member HEAD = %d", recorder.Code)
		}
	})
	t.Run("an admitted caller receives the enveloped step-1 rejections", func(t *testing.T) {
		f := newFixture(t, nil)
		for _, path := range []string{root + "//a/v1", root + "/a/../a/v1", "/api/v1/other/../fixture-packs", root + "/"} {
			recorder := serve(f, http.MethodGet, path, "admin", false)
			if recorder.Header().Get("Location") != "" {
				t.Fatalf("%s redirected", path)
			}
			requireAdmissionError(t, recorder, http.StatusNotFound, "route_not_found", map[string]any{})
			requireEvents(t, f, "authenticate", "authorize")
		}
		requireAdmissionError(t, serve(f, http.MethodGet, root+"/unknown/x/y/z", "admin", false), http.StatusNotFound, "route_not_found", map[string]any{})
		requireMethodNotAllowedEnvelope(t, serve(f, http.MethodPatch, root, "admin", false), "GET")
		requireMethodNotAllowedEnvelope(t, serve(f, http.MethodGet, root+"/a/v1/act", "admin", false), "POST")
		requireEvents(t, f, "authenticate", "authorize")
		if recorder := serve(f, http.MethodHead, root+"/a/v1", "admin", false); recorder.Code != http.StatusMethodNotAllowed || recorder.Header().Get("Allow") != "GET" || recorder.Body.Len() != 0 {
			t.Fatalf("HEAD = %d Allow=%q body=%q", recorder.Code, recorder.Header().Get("Allow"), recorder.Body)
		}
	})
	t.Run("CSRF precedes the query, which precedes path parameters", func(t *testing.T) {
		f := newFixture(t, nil)
		requireAdmissionError(t, serve(f, http.MethodPost, root+"/A/v1/act?bad=1", "admin", false), http.StatusForbidden, "csrf_verification_failed", map[string]any{})
		requireEvents(t, f, "authenticate", "authorize", "csrf")
		requireAdmissionError(t, serve(f, http.MethodPost, root+"/A/v1/act?bad=1", "admin", true), http.StatusBadRequest, "invalid_fixture_query", map[string]any{})
		requireEvents(t, f, "authenticate", "authorize", "csrf", "query:actFixture")
		for _, path := range []string{root + "/A/v1/act", root + "/a/1/act", root + "/a%2Fb/v1/act"} {
			requireAdmissionError(t, serve(f, http.MethodPost, path, "admin", true), http.StatusNotFound, "fixture_not_found", map[string]any{})
			requireEvents(t, f, "authenticate", "authorize", "csrf", "query:actFixture")
		}
		// A read declares no cookie CSRF, so none is verified.
		requireAdmissionError(t, serve(f, http.MethodGet, root+"/A/v1?bad=1", "admin", false), http.StatusBadRequest, "invalid_fixture_query", map[string]any{})
		requireEvents(t, f, "authenticate", "authorize", "query:readFixture")
	})
	t.Run("admitted values reach the handler decoded once", func(t *testing.T) {
		f := newFixture(t, nil)
		if recorder := serve(f, http.MethodPost, root+"/%61/v%31/act", "admin", true); recorder.Code != http.StatusOK {
			t.Fatalf("act = %d %s", recorder.Code, recorder.Body)
		}
		requireEvents(t, f, "authenticate", "authorize", "csrf", "query:actFixture", "handle:actFixture")
		if !reflect.DeepEqual(f.values, []string{"a/v1"}) {
			t.Fatalf("values = %v", f.values)
		}
		if recorder := serve(f, http.MethodGet, root+"?limit=1", "admin", false); recorder.Code != http.StatusOK {
			t.Fatalf("list = %d %s", recorder.Code, recorder.Body)
		}
		requireEvents(t, f, "authenticate", "authorize", "query:listFixtures", "handle:listFixtures")
		for id, mode := range f.registry.AdmissionModes() {
			if mode != AdmissionCentral {
				t.Fatalf("%s admission mode = %s", id, mode)
			}
		}
	})
	t.Run("the guard observes the gate's principal for the selected operation", func(t *testing.T) {
		var mismatches []SecurityMismatch
		f := newFixture(t, &SecurityGuard{Enforce: true, Report: func(mismatch SecurityMismatch) { mismatches = append(mismatches, mismatch) }})
		for _, tc := range []struct{ method, path string }{{http.MethodGet, root + "/a/v1"}, {http.MethodPost, root + "/a/v1/act"}, {http.MethodGet, root}} {
			if recorder := serve(f, tc.method, tc.path, "admin", true); recorder.Code != http.StatusOK {
				t.Fatalf("%s %s = %d %s", tc.method, tc.path, recorder.Code, recorder.Body)
			}
		}
		if len(mismatches) != 0 {
			t.Fatalf("mismatches = %v", mismatches)
		}
		// Control: a gate that establishes no principal is caught.
		f.authenticator.observe = false
		if recorder := serve(f, http.MethodGet, root+"/a/v1", "admin", false); recorder.Code != http.StatusInternalServerError || len(mismatches) != 1 || mismatches[0].OperationID != "readFixture" {
			t.Fatalf("unobserved gate principal: %d %v", recorder.Code, mismatches)
		}
	})
}
