package harnesscontrol

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestNetworkFlowAuditAssertionRouteDisabledByDefault(t *testing.T) {
	assertions := newBoundAuditRegistry(t)
	server := startNetworkFlowAuditAssertionHTTPServer(t, map[string]string{}, assertions)

	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", networkFlowAuditAssertionBody()))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
}

func TestNetworkFlowAuditAssertionRouteRequiresHarnessAuthorization(t *testing.T) {
	service := &networkFlowAuditAssertionService{
		guard: httpapi.TestRouteGuard{
			Token:        testRuntimeResetToken,
			ExpectedHost: "127.0.0.1:8080",
			AllowedOrigins: map[string]struct{}{
				"http://127.0.0.1:8080": {},
				"http://127.0.0.1:4173": {},
			},
		},
		assertions: newBoundAuditRegistry(t),
	}

	missingOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-audit-assertions", networkFlowAuditAssertionBody()))
	recorder := httptest.NewRecorder()
	service.handleArm(recorder, missingOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-audit-assertions", networkFlowAuditAssertionBody()))
	wrongOrigin.Header.Set("Origin", "http://evil.example.test")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongHost := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-audit-assertions", networkFlowAuditAssertionBody()))
	wrongHost.Host = "evil.example.test"
	wrongHost.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongHost)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	allowed := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-audit-assertions", networkFlowAuditAssertionBody()))
	allowed.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, allowed)
	requireTestRuntimeResetStatus(t, recorder.Result(), http.StatusCreated)
}

func TestNetworkFlowAuditAssertionRouteArmsExactCountAssertion(t *testing.T) {
	assertions := newBoundAuditRegistry(t)
	server := startNetworkFlowAuditAssertionHTTPServer(t, testRuntimeEnabledEnv(), assertions)

	body := networkFlowAuditAssertionBody()
	body["correlation_key"] = "apply:job-1"
	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", body))
	response := requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)
	data := response["data"].(map[string]any)
	if data["schema_id"] != testNetworkFlowAuditAssertionSchemaID {
		t.Fatalf("unexpected schema_id: %#v", data)
	}
	if data["assertion_kind"] != NetworkFlowAuditAssertionExactCount || data["event_code"] != NetworkFlowAuditEventTableCreated {
		t.Fatalf("unexpected audit assertion response: %#v", data)
	}
	if data["baseline_count"] != float64(0) || data["expected_final_count"] != float64(1) {
		t.Fatalf("unexpected audit counts: %#v", data)
	}

	if _, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableRenamed, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1")); ok {
		t.Fatal("wrong event code must not consume pending audit assertion")
	}
	if _, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-2", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1")); ok {
		t.Fatal("wrong operation must not consume pending audit assertion")
	}
	if _, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "")); ok {
		t.Fatal("unscoped consume must not consume a correlation-scoped audit assertion")
	}
	for _, field := range []string{"actor", "incident"} {
		wrong := testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1")
		if field == "actor" {
			wrong.ActorRef = "actor:other"
		} else {
			wrong.IncidentRef = "incident:other"
		}
		if _, ok := assertions.Consume(wrong); ok {
			t.Fatal("unrelated scope consumed audit assertion")
		}
	}
	assertion, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1"))
	if !ok {
		t.Fatal("expected exact audit assertion consume")
	}
	if assertion.ExpectedFinalCount != 1 {
		t.Fatalf("unexpected consumed assertion: %#v", assertion)
	}
	if _, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1")); ok {
		t.Fatal("audit assertion must be consumed once")
	}
}

func TestNetworkFlowAuditAssertionRouteSupportsNoAuditReplayAndDuplicateProtection(t *testing.T) {
	assertions := newBoundAuditRegistry(t)
	server := startNetworkFlowAuditAssertionHTTPServer(t, testRuntimeEnabledEnv(), assertions)

	replayBody := networkFlowAuditAssertionBody()
	replayBody["assertion_kind"] = NetworkFlowAuditAssertionNoAuditReplay
	replayBody["baseline_count"] = 1
	replayBody["expected_final_count"] = 1
	first := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", replayBody))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), first), http.StatusCreated)

	secondBody := networkFlowAuditAssertionBody()
	secondBody["resource_ref"] = "network-flow-table:table-2"
	secondBody["expected_final_count"] = 2
	second := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", secondBody))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), second), http.StatusCreated)

	duplicate := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", replayBody))
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), duplicate), http.StatusConflict, "test_network_flow_audit_assertion_already_armed")

	if assertion, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "")); !ok || assertion.AssertionKind != NetworkFlowAuditAssertionNoAuditReplay {
		t.Fatalf("first no-audit replay assertion missing or mutated: %#v ok=%v", assertion, ok)
	}
	if assertion, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-2", "")); !ok || assertion.ExpectedFinalCount != 2 {
		t.Fatalf("second independent assertion missing or mutated: %#v ok=%v", assertion, ok)
	}
}

func TestNetworkFlowAuditAssertionRouteRejectsInvalidRequests(t *testing.T) {
	if _, err := (networkFlowAuditAssertionRequest{AssertionKind: NetworkFlowAuditAssertionNoAuditReplay, EventCode: NetworkFlowAuditEventGraphQueryExecuted, OperationRef: "query", ActorRef: "actor", IncidentRef: "incident", ResourceKind: NetworkFlowAuditResourceGraph, ResourceRef: "graph", ConsumeOnce: true}).networkFlowAuditAssertion(); err == nil {
		t.Fatal("graph query replay assertion admitted")
	}
	service := &networkFlowAuditAssertionService{
		guard:      httpapi.TestRouteGuard{Token: testRuntimeResetToken},
		assertions: newBoundAuditRegistry(t),
	}
	for field, values := range map[string][]any{
		"assertion_kind": {"unknown"}, "event_code": {"unknown"}, "operation_ref": {"bad ref"}, "actor_ref": {"bad ref"}, "incident_ref": {"bad ref"}, "resource_ref": {"bad ref"},
		"resource_kind": {"network_flow_import", "network_flow_graph"}, "baseline_count": {-1, 1000001, 2}, "expected_final_count": {-1, 1000001}, "expected_replay_increment": {0, 1}, "consume_once": {false}, "unexpected": {true},
	} {
		for _, value := range values {
			body := networkFlowAuditAssertionBody()
			body[field] = value
			req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "/api/v1/test/runtime/network-flow-audit-assertions", body))
			recorder := httptest.NewRecorder()
			service.handleArm(recorder, req)
			requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusBadRequest, "invalid_network_flow_audit_assertion_request")
		}
	}
}

func TestNetworkFlowAuditAssertionRegistryClearRemovesArmedAssertions(t *testing.T) {
	assertions := newBoundAuditRegistry(t)
	server := startNetworkFlowAuditAssertionHTTPServer(t, testRuntimeEnabledEnv(), assertions)

	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", networkFlowAuditAssertionBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	assertions.Clear()
	if _, ok := assertions.Consume(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "")); ok {
		t.Fatal("clear must remove armed Network Flow audit assertions")
	}
}

func startNetworkFlowAuditAssertionHTTPServer(t testing.TB, env map[string]string, assertions *NetworkFlowAuditAssertionRegistry) *httptest.Server {
	t.Helper()
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: testHTTPDependencies(httpapi.DependencySet{
			Env: env,
		}),
		AdditionalRoutes: []httpapi.RouteRegistrar{
			RegisterNetworkFlowAuditAssertionRoutes(assertions),
		},
	})
	if err != nil {
		t.Fatalf("new Network Flow audit-assertion handler: %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func networkFlowAuditAssertionBody() map[string]any {
	return map[string]any{
		"assertion_kind":       NetworkFlowAuditAssertionExactCount,
		"event_code":           NetworkFlowAuditEventTableCreated,
		"operation_ref":        "import:apply-1",
		"actor_ref":            "actor:analyst-1",
		"incident_ref":         "incident:alpha",
		"resource_kind":        NetworkFlowAuditResourceTable,
		"resource_ref":         "network-flow-table:table-1",
		"baseline_count":       0,
		"expected_final_count": 1,
		"consume_once":         true,
	}
}

func testAuditScope(event, operation, kind, resource, correlation string) NetworkFlowAuditScope {
	return NetworkFlowAuditScope{EventCode: event, OperationRef: operation, ActorRef: "actor:analyst-1", IncidentRef: "incident:alpha", ResourceKind: kind, ResourceRef: resource, CorrelationKey: correlation}
}
func newBoundAuditRegistry(t testing.TB) *NetworkFlowAuditAssertionRegistry {
	t.Helper()
	r := NewNetworkFlowAuditAssertionRegistry()
	for _, ref := range []string{"network-flow-table:table-1", "network-flow-table:table-2"} {
		if err := r.BindFixture(testAuditScope(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, ref, ""), "owned-fixture:"+ref); err != nil {
			t.Fatal(err)
		}
	}
	return r
}
