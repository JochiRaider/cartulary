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

func TestNetworkFlowFaultRouteDisabledByDefault(t *testing.T) {
	faults := NewNetworkFlowFaultRegistry()
	server := startNetworkFlowFaultHTTPServer(t, map[string]string{}, faults)

	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
}

func TestNetworkFlowFaultRouteRequiresHarnessAuthorization(t *testing.T) {
	service := &networkFlowFaultService{
		guard: httpapi.TestRouteGuard{
			Token:        testRuntimeResetToken,
			ExpectedHost: "127.0.0.1:8080",
			AllowedOrigins: map[string]struct{}{
				"http://127.0.0.1:8080": {},
				"http://127.0.0.1:4173": {},
			},
		},
		faults: NewNetworkFlowFaultRegistry(),
	}

	missingOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	recorder := httptest.NewRecorder()
	service.handleArm(recorder, missingOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	wrongOrigin.Header.Set("Origin", "http://evil.example.test")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongHost := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	wrongHost.Host = "evil.example.test"
	wrongHost.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongHost)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	allowed := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	allowed.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, allowed)
	requireTestRuntimeResetStatus(t, recorder.Result(), http.StatusCreated)
}

func TestNetworkFlowFaultRouteArmsOneShotBoundaryFault(t *testing.T) {
	faults := NewNetworkFlowFaultRegistry()
	server := startNetworkFlowFaultHTTPServer(t, testRuntimeEnabledEnv(), faults)

	body := networkFlowFaultBody()
	body["correlation_key"] = "apply:job-1"
	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-faults", body))
	response := requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)
	data := response["data"].(map[string]any)
	if data["schema_id"] != testNetworkFlowFaultSchemaID {
		t.Fatalf("unexpected schema_id: %#v", data)
	}
	if data["boundary"] != NetworkFlowFaultBoundaryImportBeforeTransactionCommit || data["fault_kind"] != NetworkFlowFaultKindReturnError {
		t.Fatalf("unexpected fault response: %#v", data)
	}
	if data["error_code"] != "network_flow_fault_probe" || data["correlation_key"] != "apply:job-1" {
		t.Fatalf("unexpected scoped fault response: %#v", data)
	}

	if _, ok := faults.ConsumeNetworkFlowFault(NetworkFlowFaultBoundaryImportAfterOwnerPrepare); ok {
		t.Fatal("wrong boundary must not consume pending Network Flow fault")
	}
	if _, ok := faults.ConsumeNetworkFlowFault(NetworkFlowFaultBoundaryImportBeforeTransactionCommit); ok {
		t.Fatal("unscoped consume must not consume a correlation-scoped Network Flow fault")
	}
	if _, ok := faults.ConsumeNetworkFlowFaultFor(NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "apply:job-2"); ok {
		t.Fatal("wrong correlation key must not consume pending Network Flow fault")
	}
	fault, ok := faults.ConsumeNetworkFlowFaultFor(NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "apply:job-1")
	if !ok {
		t.Fatal("expected exact boundary and correlation key to consume pending Network Flow fault")
	}
	if fault.FaultKind != NetworkFlowFaultKindReturnError || fault.ErrorCode != "network_flow_fault_probe" {
		t.Fatalf("unexpected consumed fault: %#v", fault)
	}
	if _, ok := faults.ConsumeNetworkFlowFaultFor(NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "apply:job-1"); ok {
		t.Fatal("Network Flow fault must be consumed once")
	}

	rearm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), rearm), http.StatusCreated)
}

func TestNetworkFlowFaultRouteRejectsSecondArmWhilePending(t *testing.T) {
	faults := NewNetworkFlowFaultRegistry()
	server := startNetworkFlowFaultHTTPServer(t, testRuntimeEnabledEnv(), faults)

	first := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), first), http.StatusCreated)

	secondBody := networkFlowFaultBody()
	secondBody["boundary"] = NetworkFlowFaultBoundaryWorkerBeforeFinalCommit
	secondBody["fault_kind"] = NetworkFlowFaultKindWorkerCrash
	delete(secondBody, "error_code")
	second := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-faults", secondBody))
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), second), http.StatusConflict, "test_network_flow_fault_already_armed")

	if _, ok := faults.ConsumeNetworkFlowFault(NetworkFlowFaultBoundaryImportBeforeTransactionCommit); !ok {
		t.Fatal("first fault must remain armed after rejected replacement")
	}
	third := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-faults", secondBody))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), third), http.StatusCreated)
}

func TestNetworkFlowFaultRouteRejectsInvalidRequests(t *testing.T) {
	service := &networkFlowFaultService{
		guard:  httpapi.TestRouteGuard{Token: testRuntimeResetToken},
		faults: NewNetworkFlowFaultRegistry(),
	}
	for _, body := range []map[string]any{
		{"boundary": "network_flow.import.unknown", "fault_kind": NetworkFlowFaultKindReturnError, "error_code": "network_flow_fault_probe", "consume_once": true},
		{"boundary": NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "fault_kind": "unknown", "consume_once": true},
		{"boundary": NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "fault_kind": NetworkFlowFaultKindReturnError, "consume_once": true},
		{"boundary": NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "fault_kind": NetworkFlowFaultKindReturnError, "error_code": "Bad-Code", "consume_once": true},
		{"boundary": NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "fault_kind": NetworkFlowFaultKindPanic, "error_code": "network_flow_fault_probe", "consume_once": true},
		{"boundary": NetworkFlowFaultBoundaryImportBeforeTransactionCommit, "fault_kind": NetworkFlowFaultKindWorkerCrash, "consume_once": true},
		{"boundary": NetworkFlowFaultBoundaryWorkerBeforeFinalCommit, "fault_kind": NetworkFlowFaultKindWorkerCrash, "correlation_key": "bad key", "consume_once": true},
		{"boundary": NetworkFlowFaultBoundaryWorkerBeforeFinalCommit, "fault_kind": NetworkFlowFaultKindWorkerCrash, "consume_once": false},
		{"boundary": NetworkFlowFaultBoundaryWorkerBeforeFinalCommit, "fault_kind": NetworkFlowFaultKindWorkerCrash, "consume_once": true, "unexpected": true},
	} {
		req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "/api/v1/test/runtime/network-flow-faults", body))
		recorder := httptest.NewRecorder()
		service.handleArm(recorder, req)
		requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusBadRequest, "invalid_network_flow_fault_request")
	}
}

func TestTestRuntimeResetClearsNetworkFlowFaults(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, "test-network-flow-fault-reset")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-network-flow-fault-reset")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	faults := NewNetworkFlowFaultRegistry()
	server := startTestRuntimeResetServerWithHTTPDeps(t, env, []httpapi.RouteRegistrar{
		harnessruntime.RegisterTestRuntimeResetRoute(faults.Clear),
		RegisterNetworkFlowFaultRoutes(faults),
	}, httpapi.DependencySet{})

	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-faults", networkFlowFaultBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	reset := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), reset), http.StatusOK)

	if _, ok := faults.ConsumeNetworkFlowFault(NetworkFlowFaultBoundaryImportBeforeTransactionCommit); ok {
		t.Fatal("runtime reset must clear armed Network Flow faults")
	}
}

func startNetworkFlowFaultHTTPServer(t testing.TB, env map[string]string, faults *NetworkFlowFaultRegistry) *httptest.Server {
	t.Helper()
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: httpapi.DependencySet{
			Env: env,
		},
		AdditionalRoutes: []httpapi.RouteRegistrar{
			RegisterNetworkFlowFaultRoutes(faults),
		},
	})
	if err != nil {
		t.Fatalf("new Network Flow fault handler: %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func networkFlowFaultBody() map[string]any {
	return map[string]any{
		"boundary":     NetworkFlowFaultBoundaryImportBeforeTransactionCommit,
		"fault_kind":   NetworkFlowFaultKindReturnError,
		"error_code":   "network_flow_fault_probe",
		"consume_once": true,
	}
}
