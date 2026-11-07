package httpapi

import (
	"net/http"
	"net/http/httptest"
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
	newDeps := func() DependencySet {
		registry, err := newRouteRegistry(operations, []ExtensionClaim{{ProfileID: profile, Claimed: true}})
		if err != nil {
			t.Fatal(err)
		}
		return DependencySet{PublicRoutes: registry, ExtensionRoutes: familyTestRoutes{{ProfileID: profile, RouteFamily: root, Claimed: true}}}
	}
	calls := 0
	policy := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Fixture-Admitted") != "yes" {
				w.WriteHeader(403)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		calls++
		metadata, ok := OperationMetadataFromContext(r.Context())
		if !ok || metadata.OwnerID != owner || metadata.CookieCSRFRequired != (r.Method == "POST") {
			t.Fatal("lost operation security metadata")
		}
		_, _ = w.Write([]byte(r.PathValue("version")))
	}
	handlers := map[string]http.HandlerFunc{"readFixture": handler, "inspectFixture": handler}
	deps := newDeps()
	if err := BindOwnerFamilyRoutes(deps, owner, profile, root, policy, handlers); err != nil {
		t.Fatal(err)
	}
	if err := deps.PublicRoutes.ValidateActive(); err != nil {
		t.Fatal(err)
	}
	if err := BindOwnerFamilyRoutes(deps, owner, profile, root, policy, handlers); err == nil {
		t.Fatal("duplicate family admitted")
	}
	if _, err := deps.PublicRoutes.withOwnerFamilies(http.NotFoundHandler(), nil); err == nil {
		t.Fatal("lost claimed projection admitted")
	}
	dispatch, err := deps.PublicRoutes.withOwnerFamilies(http.NotFoundHandler(), extensionRoutesFromDependencies(deps))
	if err != nil {
		t.Fatal(err)
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
			r := httptest.NewRequest(method, path, nil)
			r.Header.Set("X-Fixture-Admitted", "yes")
			w := httptest.NewRecorder()
			dispatch.ServeHTTP(w, r)
			if w.Code != 404 || w.Header().Get("Location") != "" || calls != 0 {
				t.Fatalf("noncanonical path was redirected/dispatched: %s %s = %d", method, path, w.Code)
			}
		}
	}
	for _, method := range []string{"GET", "POST", "HEAD", "OPTIONS", "PATCH"} {
		r := httptest.NewRequest(method, root+"/%61ctivate", nil)
		r.Header.Set("X-Fixture-Admitted", "yes")
		w := httptest.NewRecorder()
		dispatch.ServeHTTP(w, r)
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
		r := httptest.NewRequest("GET", root+"/"+encoded, nil)
		r.Header.Set("X-Fixture-Admitted", "yes")
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
	for _, bad := range []struct {
		name, owner, profile, root string
		policy                     func(http.Handler) http.Handler
		handlers                   map[string]http.HandlerFunc
	}{
		{"nil policy", owner, profile, root, nil, handlers},
		{"nil result", owner, profile, root, func(http.Handler) http.Handler { return nil }, handlers},
		{"wrong owner", "module.other", profile, root, policy, handlers},
		{"wrong profile", owner, "other", root, policy, handlers},
		{"wrong root", owner, profile, "/api/v1/other", policy, handlers},
		{"missing operation", owner, profile, root, policy, map[string]http.HandlerFunc{"readFixture": handler}},
		{"unknown operation", owner, profile, root, policy, map[string]http.HandlerFunc{"readFixture": handler, "other": handler}},
		{"nil operation", owner, profile, root, policy, map[string]http.HandlerFunc{"readFixture": handler, "inspectFixture": nil}},
	} {
		t.Run(bad.name, func(t *testing.T) {
			deps := newDeps()
			if err := BindOwnerFamilyRoutes(deps, bad.owner, bad.profile, bad.root, bad.policy, bad.handlers); err == nil {
				t.Fatal("invalid family registration admitted")
			}
			if len(deps.PublicRoutes.Snapshot()) != 0 {
				t.Fatal("rejected registration partially published operations")
			}
		})
	}
}
