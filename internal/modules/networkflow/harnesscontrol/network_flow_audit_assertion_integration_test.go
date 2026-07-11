package harnesscontrol

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestNetworkFlowAuditAssertionRouteDisabledByDefault(t *testing.T) {
	assertions := NewNetworkFlowAuditAssertionRegistry()
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
		assertions: NewNetworkFlowAuditAssertionRegistry(),
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
	assertions := NewNetworkFlowAuditAssertionRegistry()
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
	if data["baseline_count"] != float64(0) || data["expected_final_count"] != float64(1) || data["expected_replay_increment"] != float64(0) {
		t.Fatalf("unexpected audit counts: %#v", data)
	}

	if _, ok := assertions.ConsumeNetworkFlowAuditAssertionFor(NetworkFlowAuditEventTableRenamed, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1"); ok {
		t.Fatal("wrong event code must not consume pending audit assertion")
	}
	if _, ok := assertions.ConsumeNetworkFlowAuditAssertionFor(NetworkFlowAuditEventTableCreated, "import:apply-2", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1"); ok {
		t.Fatal("wrong operation must not consume pending audit assertion")
	}
	if _, ok := assertions.ConsumeNetworkFlowAuditAssertion(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1"); ok {
		t.Fatal("unscoped consume must not consume a correlation-scoped audit assertion")
	}
	assertion, ok := assertions.ConsumeNetworkFlowAuditAssertionFor(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1")
	if !ok {
		t.Fatal("expected exact audit assertion consume")
	}
	if assertion.ExpectedFinalCount != 1 || assertion.ExpectedReplayIncrement != 0 {
		t.Fatalf("unexpected consumed assertion: %#v", assertion)
	}
	if _, ok := assertions.ConsumeNetworkFlowAuditAssertionFor(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1", "apply:job-1"); ok {
		t.Fatal("audit assertion must be consumed once")
	}
}

func TestNetworkFlowAuditAssertionRouteSupportsNoAuditReplayAndDuplicateProtection(t *testing.T) {
	assertions := NewNetworkFlowAuditAssertionRegistry()
	server := startNetworkFlowAuditAssertionHTTPServer(t, testRuntimeEnabledEnv(), assertions)

	replayBody := networkFlowAuditAssertionBody()
	replayBody["assertion_kind"] = NetworkFlowAuditAssertionNoAuditReplay
	replayBody["baseline_count"] = 1
	replayBody["expected_final_count"] = 1
	replayBody["expected_replay_increment"] = 0
	first := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", replayBody))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), first), http.StatusCreated)

	secondBody := networkFlowAuditAssertionBody()
	secondBody["resource_ref"] = "network-flow-table:table-2"
	secondBody["expected_final_count"] = 2
	second := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", secondBody))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), second), http.StatusCreated)

	duplicate := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", replayBody))
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), duplicate), http.StatusConflict, "test_network_flow_audit_assertion_already_armed")

	if assertion, ok := assertions.ConsumeNetworkFlowAuditAssertion(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1"); !ok || assertion.AssertionKind != NetworkFlowAuditAssertionNoAuditReplay {
		t.Fatalf("first no-audit replay assertion missing or mutated: %#v ok=%v", assertion, ok)
	}
	if assertion, ok := assertions.ConsumeNetworkFlowAuditAssertion(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-2"); !ok || assertion.ExpectedFinalCount != 2 {
		t.Fatalf("second independent assertion missing or mutated: %#v ok=%v", assertion, ok)
	}
}

func TestNetworkFlowAuditAssertionRouteRejectsInvalidRequests(t *testing.T) {
	service := &networkFlowAuditAssertionService{
		guard:      httpapi.TestRouteGuard{Token: testRuntimeResetToken},
		assertions: NewNetworkFlowAuditAssertionRegistry(),
	}
	for _, body := range []map[string]any{
		{"assertion_kind": "unknown", "event_code": NetworkFlowAuditEventTableCreated, "operation_ref": "import:apply-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceTable, "resource_ref": "network-flow-table:table-1", "baseline_count": 0, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": true},
		{"assertion_kind": NetworkFlowAuditAssertionExactCount, "event_code": "network_flow_secret_viewed", "operation_ref": "import:apply-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceTable, "resource_ref": "network-flow-table:table-1", "baseline_count": 0, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": true},
		{"assertion_kind": NetworkFlowAuditAssertionExactCount, "event_code": NetworkFlowAuditEventTableCreated, "operation_ref": "bad operation", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceTable, "resource_ref": "network-flow-table:table-1", "baseline_count": 0, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": true},
		{"assertion_kind": NetworkFlowAuditAssertionExactCount, "event_code": NetworkFlowAuditEventTableCreated, "operation_ref": "import:apply-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": "network_flow_secret", "resource_ref": "network-flow-table:table-1", "baseline_count": 0, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": true},
		{"assertion_kind": NetworkFlowAuditAssertionExactCount, "event_code": NetworkFlowAuditEventTableCreated, "operation_ref": "import:apply-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceTable, "resource_ref": "network-flow-table:table-1", "baseline_count": 2, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": true},
		{"assertion_kind": NetworkFlowAuditAssertionZeroOccurrences, "event_code": NetworkFlowAuditEventGraphQueryExecuted, "operation_ref": "graph:denied-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceGraph, "resource_ref": "network-flow-graph:graph-1", "baseline_count": 0, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": true},
		{"assertion_kind": NetworkFlowAuditAssertionNoAuditReplay, "event_code": NetworkFlowAuditEventTableCreated, "operation_ref": "import:apply-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceTable, "resource_ref": "network-flow-table:table-1", "baseline_count": 1, "expected_final_count": 1, "expected_replay_increment": 1, "consume_once": true},
		{"assertion_kind": NetworkFlowAuditAssertionExactCount, "event_code": NetworkFlowAuditEventTableCreated, "operation_ref": "import:apply-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceTable, "resource_ref": "network-flow-table:table-1", "baseline_count": 0, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": false},
		{"assertion_kind": NetworkFlowAuditAssertionExactCount, "event_code": NetworkFlowAuditEventTableCreated, "operation_ref": "import:apply-1", "actor_ref": "actor:analyst-1", "incident_ref": "incident:alpha", "resource_kind": NetworkFlowAuditResourceTable, "resource_ref": "network-flow-table:table-1", "baseline_count": 0, "expected_final_count": 1, "expected_replay_increment": 0, "consume_once": true, "unexpected": true},
	} {
		req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "/api/v1/test/runtime/network-flow-audit-assertions", body))
		recorder := httptest.NewRecorder()
		service.handleArm(recorder, req)
		requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusBadRequest, "invalid_network_flow_audit_assertion_request")
	}
}

func TestTestRuntimeResetClearsNetworkFlowAuditAssertions(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "test-network-flow-audit-assertion-reset")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-network-flow-audit-assertion-reset")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	assertions := NewNetworkFlowAuditAssertionRegistry()
	_, server := startTestRuntimeResetServerWithHTTPDeps(t, env, []httpapi.RouteRegistrar{
		harnessruntime.RegisterTestRuntimeResetRoute(assertions.Clear),
		RegisterNetworkFlowAuditAssertionRoutes(assertions),
	}, httpapi.DependencySet{})

	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-audit-assertions", networkFlowAuditAssertionBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	reset := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), reset), http.StatusOK)

	if _, ok := assertions.ConsumeNetworkFlowAuditAssertion(NetworkFlowAuditEventTableCreated, "import:apply-1", NetworkFlowAuditResourceTable, "network-flow-table:table-1"); ok {
		t.Fatal("runtime reset must clear armed Network Flow audit assertions")
	}
}

func startNetworkFlowAuditAssertionHTTPServer(t testing.TB, env map[string]string, assertions *NetworkFlowAuditAssertionRegistry) *httptest.Server {
	t.Helper()
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: httpapi.DependencySet{
			Env: env,
		},
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
		"assertion_kind":            NetworkFlowAuditAssertionExactCount,
		"event_code":                NetworkFlowAuditEventTableCreated,
		"operation_ref":             "import:apply-1",
		"actor_ref":                 "actor:analyst-1",
		"incident_ref":              "incident:alpha",
		"resource_kind":             NetworkFlowAuditResourceTable,
		"resource_ref":              "network-flow-table:table-1",
		"baseline_count":            0,
		"expected_final_count":      1,
		"expected_replay_increment": 0,
		"consume_once":              true,
	}
}
