package harnessruntime

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestNetworkFlowAuthTransitionRouteDisabledByDefault(t *testing.T) {
	transitions := NewNetworkFlowAuthTransitionRegistry()
	server := startNetworkFlowAuthTransitionHTTPServer(t, map[string]string{}, transitions)

	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
}

func TestNetworkFlowAuthTransitionRouteRequiresHarnessAuthorization(t *testing.T) {
	service := &networkFlowAuthTransitionService{
		guard: httpapi.TestRouteGuard{
			Token:        testRuntimeResetToken,
			ExpectedHost: "127.0.0.1:8080",
			AllowedOrigins: map[string]struct{}{
				"http://127.0.0.1:8080": {},
				"http://127.0.0.1:4173": {},
			},
		},
		transitions: NewNetworkFlowAuthTransitionRegistry(),
	}

	missingOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	recorder := httptest.NewRecorder()
	service.handleArm(recorder, missingOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	wrongOrigin.Header.Set("Origin", "http://evil.example.test")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongHost := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	wrongHost.Host = "evil.example.test"
	wrongHost.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongHost)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	allowed := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	allowed.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, allowed)
	requireTestRuntimeResetStatus(t, recorder.Result(), http.StatusCreated)
}

func TestNetworkFlowAuthTransitionRouteArmsExactHiddenResourceTransition(t *testing.T) {
	transitions := NewNetworkFlowAuthTransitionRegistry()
	server := startNetworkFlowAuthTransitionHTTPServer(t, testRuntimeEnabledEnv(), transitions)

	body := networkFlowAuthTransitionBody()
	body["correlation_key"] = "query:page-1"
	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-auth-transitions", body))
	response := requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)
	data := response["data"].(map[string]any)
	if data["schema_id"] != testNetworkFlowAuthTransitionSchemaID {
		t.Fatalf("unexpected schema_id: %#v", data)
	}
	if data["boundary"] != NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup || data["transition_kind"] != NetworkFlowAuthTransitionKindIncidentMembershipRevoked {
		t.Fatalf("unexpected transition response: %#v", data)
	}
	if data["resource_kind"] != NetworkFlowAuthResourceNetworkFlowTable || data["hidden_response_kind"] != NetworkFlowHiddenResponseNotFound {
		t.Fatalf("unexpected hidden-resource response: %#v", data)
	}
	if data["must_not_disclose_resource"] != true || data["consume_once"] != true || data["correlation_key"] != "query:page-1" {
		t.Fatalf("unexpected control flags: %#v", data)
	}

	if _, ok := transitions.ConsumeNetworkFlowAuthTransitionFor(NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-1", "query:page-1"); ok {
		t.Fatal("wrong boundary must not consume pending auth transition")
	}
	if _, ok := transitions.ConsumeNetworkFlowAuthTransitionFor(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-2", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-1", "query:page-1"); ok {
		t.Fatal("wrong actor must not consume pending auth transition")
	}
	if _, ok := transitions.ConsumeNetworkFlowAuthTransition(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowCursor, "network-flow-table:table-1"); ok {
		t.Fatal("wrong resource kind must not consume pending auth transition")
	}
	if _, ok := transitions.ConsumeNetworkFlowAuthTransition(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-1"); ok {
		t.Fatal("unscoped consume must not consume a correlation-scoped auth transition")
	}
	transition, ok := transitions.ConsumeNetworkFlowAuthTransitionFor(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-1", "query:page-1")
	if !ok {
		t.Fatal("expected exact boundary, refs, and correlation key to consume auth transition")
	}
	if transition.HiddenResponseKind != NetworkFlowHiddenResponseNotFound || !transition.MustNotDiscloseResource {
		t.Fatalf("unexpected consumed transition: %#v", transition)
	}
	if _, ok := transitions.ConsumeNetworkFlowAuthTransitionFor(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-1", "query:page-1"); ok {
		t.Fatal("auth transition must be consumed once")
	}
}

func TestNetworkFlowAuthTransitionRouteAllowsIndependentKeysAndRejectsDuplicate(t *testing.T) {
	transitions := NewNetworkFlowAuthTransitionRegistry()
	server := startNetworkFlowAuthTransitionHTTPServer(t, testRuntimeEnabledEnv(), transitions)

	first := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), first), http.StatusCreated)

	secondBody := networkFlowAuthTransitionBody()
	secondBody["resource_ref"] = "network-flow-table:table-2"
	secondBody["hidden_response_kind"] = NetworkFlowHiddenResponseEmptyCollection
	second := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-auth-transitions", secondBody))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), second), http.StatusCreated)

	duplicate := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), duplicate), http.StatusConflict, "test_network_flow_auth_transition_already_armed")

	if _, ok := transitions.ConsumeNetworkFlowAuthTransition(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-1"); !ok {
		t.Fatal("first transition must remain armed after duplicate rejection")
	}
	if transition, ok := transitions.ConsumeNetworkFlowAuthTransition(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-2"); !ok || transition.HiddenResponseKind != NetworkFlowHiddenResponseEmptyCollection {
		t.Fatalf("second independent transition missing or mutated: %#v ok=%v", transition, ok)
	}
}

func TestNetworkFlowAuthTransitionRouteRejectsInvalidRequests(t *testing.T) {
	service := &networkFlowAuthTransitionService{
		guard:       httpapi.TestRouteGuard{Token: testRuntimeResetToken},
		transitions: NewNetworkFlowAuthTransitionRegistry(),
	}
	for _, body := range []map[string]any{
		{"boundary": "network_flow.route.unknown", "transition_kind": NetworkFlowAuthTransitionKindIncidentMembershipRevoked, "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuthResourceNetworkFlowTable, "resource_ref": "network-flow-table:table-1", "hidden_response_kind": NetworkFlowHiddenResponseNotFound, "must_not_disclose_resource": true, "consume_once": true},
		{"boundary": NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "transition_kind": "unknown", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuthResourceNetworkFlowTable, "resource_ref": "network-flow-table:table-1", "hidden_response_kind": NetworkFlowHiddenResponseNotFound, "must_not_disclose_resource": true, "consume_once": true},
		{"boundary": NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "transition_kind": NetworkFlowAuthTransitionKindIncidentMembershipRevoked, "actor_ref": "bad actor", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuthResourceNetworkFlowTable, "resource_ref": "network-flow-table:table-1", "hidden_response_kind": NetworkFlowHiddenResponseNotFound, "must_not_disclose_resource": true, "consume_once": true},
		{"boundary": NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "transition_kind": NetworkFlowAuthTransitionKindIncidentMembershipRevoked, "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": "network_flow_secret", "resource_ref": "network-flow-table:table-1", "hidden_response_kind": NetworkFlowHiddenResponseNotFound, "must_not_disclose_resource": true, "consume_once": true},
		{"boundary": NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "transition_kind": NetworkFlowAuthTransitionKindIncidentMembershipRevoked, "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuthResourceNetworkFlowTable, "resource_ref": "network-flow-table:table-1", "hidden_response_kind": "raw_resource", "must_not_disclose_resource": true, "consume_once": true},
		{"boundary": NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "transition_kind": NetworkFlowAuthTransitionKindIncidentMembershipRevoked, "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuthResourceNetworkFlowTable, "resource_ref": "network-flow-table:table-1", "hidden_response_kind": NetworkFlowHiddenResponseNotFound, "must_not_disclose_resource": false, "consume_once": true},
		{"boundary": NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "transition_kind": NetworkFlowAuthTransitionKindIncidentMembershipRevoked, "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuthResourceNetworkFlowTable, "resource_ref": "network-flow-table:table-1", "hidden_response_kind": NetworkFlowHiddenResponseNotFound, "must_not_disclose_resource": true, "consume_once": false},
		{"boundary": NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "transition_kind": NetworkFlowAuthTransitionKindIncidentMembershipRevoked, "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuthResourceNetworkFlowTable, "resource_ref": "network-flow-table:table-1", "hidden_response_kind": NetworkFlowHiddenResponseNotFound, "must_not_disclose_resource": true, "consume_once": true, "unexpected": true},
	} {
		req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "/api/v1/test/runtime/network-flow-auth-transitions", body))
		recorder := httptest.NewRecorder()
		service.handleArm(recorder, req)
		requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusBadRequest, "invalid_network_flow_auth_transition_request")
	}
}

func TestTestRuntimeResetClearsNetworkFlowAuthTransitions(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, "test-network-flow-auth-transition-reset")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-network-flow-auth-transition-reset")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	transitions := NewNetworkFlowAuthTransitionRegistry()
	_, server := startTestRuntimeResetServerWithHTTPDeps(t, env, []httpapi.RouteRegistrar{
		RegisterTestRuntimeResetRoute(transitions.Clear),
		RegisterNetworkFlowAuthTransitionRoutes(transitions),
	}, httpapi.DependencySet{})

	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-auth-transitions", networkFlowAuthTransitionBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	reset := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), reset), http.StatusOK)

	if _, ok := transitions.ConsumeNetworkFlowAuthTransition(NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup, "actor:analyst-1", "incident:alpha", NetworkFlowAuthResourceNetworkFlowTable, "network-flow-table:table-1"); ok {
		t.Fatal("runtime reset must clear armed Network Flow auth transitions")
	}
}

func startNetworkFlowAuthTransitionHTTPServer(t testing.TB, env map[string]string, transitions *NetworkFlowAuthTransitionRegistry) *httptest.Server {
	t.Helper()
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: httpapi.DependencySet{
			Env: env,
		},
		AdditionalRoutes: []httpapi.RouteRegistrar{
			RegisterNetworkFlowAuthTransitionRoutes(transitions),
		},
	})
	if err != nil {
		t.Fatalf("new Network Flow auth-transition handler: %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func networkFlowAuthTransitionBody() map[string]any {
	return map[string]any{
		"boundary":                   NetworkFlowAuthTransitionBoundaryRouteAfterAuthorizationBeforeLookup,
		"transition_kind":            NetworkFlowAuthTransitionKindIncidentMembershipRevoked,
		"actor_ref":                  "actor:analyst-1",
		"incident_ref":               "incident:alpha",
		"resource_kind":              NetworkFlowAuthResourceNetworkFlowTable,
		"resource_ref":               "network-flow-table:table-1",
		"hidden_response_kind":       NetworkFlowHiddenResponseNotFound,
		"must_not_disclose_resource": true,
		"consume_once":               true,
	}
}
