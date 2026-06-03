package testruntime

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestPublicErrorFaultRouteDisabledByDefault(t *testing.T) {
	faults := NewPublicErrorFaultRegistry()
	server := startPublicErrorFaultHTTPServer(t, map[string]string{}, faults)

	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/public-error-faults", map[string]any{
		"method":       "PATCH",
		"path":         "/api/v1/records/record-1",
		"status":       http.StatusBadGateway,
		"code":         "unknown_public_error_probe",
		"message":      "unexpected public error",
		"consume_once": true,
	}))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
}

func TestPublicErrorFaultRouteRequiresHarnessAuthorization(t *testing.T) {
	service := &publicErrorFaultService{
		guard: httpapi.TestRouteGuard{
			Token:        testRuntimeResetToken,
			ExpectedHost: "127.0.0.1:8080",
			AllowedOrigins: map[string]struct{}{
				"http://127.0.0.1:8080": {},
				"http://127.0.0.1:4173": {},
			},
		},
		faults: NewPublicErrorFaultRegistry(),
	}

	missingOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/public-error-faults", publicErrorFaultBody()))
	recorder := httptest.NewRecorder()
	service.handleArm(recorder, missingOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/public-error-faults", publicErrorFaultBody()))
	wrongOrigin.Header.Set("Origin", "http://evil.example.test")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongHost := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/public-error-faults", publicErrorFaultBody()))
	wrongHost.Host = "evil.example.test"
	wrongHost.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongHost)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	allowed := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/public-error-faults", publicErrorFaultBody()))
	allowed.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, allowed)
	requireTestRuntimeResetStatus(t, recorder.Result(), http.StatusCreated)
}

func TestPublicErrorFaultRouteArmsOneShotOrdinaryPublicMutation(t *testing.T) {
	faults := NewPublicErrorFaultRegistry()
	server := startPublicErrorFaultHTTPServer(t, testRuntimeEnabledEnv(), faults)

	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/public-error-faults", map[string]any{
		"method":       "PATCH",
		"path":         "/api/v1/records/record-1",
		"status":       http.StatusBadGateway,
		"code":         "unknown_public_error_probe",
		"message":      "stack trace at handler (/home/cartulary/internal/private.go:42)",
		"retryable":    true,
		"details":      map[string]any{"reason_code": "unknown_public_error_probe", "private_path": "/home/cartulary/internal/private.go"},
		"consume_once": true,
	}))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	first := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPatch, server.URL+"/api/v1/records/record-1", map[string]any{}))
	body := requireTestRuntimeResetErrorEnvelope(t, first, http.StatusBadGateway, "unknown_public_error_probe")
	errorValue := body["error"].(map[string]any)
	if errorValue["retryable"] != true {
		t.Fatalf("fault must preserve configured retryable flag: %#v", errorValue)
	}
	if errorValue["request_id"] == "" {
		t.Fatalf("fault envelope must preserve public request id: %#v", errorValue)
	}

	second := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPatch, server.URL+"/api/v1/records/record-1", map[string]any{}))
	success := requireTestRuntimeResetSuccessEnvelope(t, second, http.StatusOK)
	if success["data"].(map[string]any)["ok"] != true {
		t.Fatalf("ordinary route did not run after one-shot fault was consumed: %#v", success)
	}
}

func TestPublicErrorFaultRouteRejectsInvalidTargets(t *testing.T) {
	service := &publicErrorFaultService{
		guard:  httpapi.TestRouteGuard{Token: testRuntimeResetToken},
		faults: NewPublicErrorFaultRegistry(),
	}
	for _, body := range []map[string]any{
		{"method": "PATCH", "path": "/api/v1/test/runtime/reset", "status": 500, "code": "bad", "consume_once": true},
		{"method": "PATCH", "path": "/api/v1/records/record-1?x=1", "status": 500, "code": "bad", "consume_once": true},
		{"method": "PATCH", "path": "/api/v1/records/record-1", "status": 200, "code": "bad", "consume_once": true},
		{"method": "PATCH", "path": "/api/v1/records/record-1", "status": 500, "code": "", "consume_once": true},
		{"method": "PATCH", "path": "/api/v1/records/record-1", "status": 500, "code": "bad", "consume_once": false},
	} {
		req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "/api/v1/test/runtime/public-error-faults", body))
		recorder := httptest.NewRecorder()
		service.handleArm(recorder, req)
		requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusBadRequest, "invalid_public_error_fault_request")
	}
}

func TestTestRuntimeResetClearsPublicErrorFaults(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, "test-public-error-fault-reset")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-public-error-fault-reset")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	faults := NewPublicErrorFaultRegistry()
	_, server := startTestRuntimeResetServerWithHTTPDeps(t, env, []httpapi.RouteRegistrar{
		RegisterTestRuntimeResetRoute(faults.Clear),
		RegisterPublicErrorFaultRoutes(faults),
	}, httpapi.DependencySet{PublicErrorFaults: faults})

	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/public-error-faults", publicErrorFaultBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	reset := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), reset), http.StatusOK)

	if _, ok := faults.ConsumePublicErrorFault("PATCH", "/api/v1/records/record-1"); ok {
		t.Fatal("runtime reset must clear armed public error faults")
	}
}

func startPublicErrorFaultHTTPServer(t testing.TB, env map[string]string, faults *PublicErrorFaultRegistry) *httptest.Server {
	t.Helper()
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: httpapi.DependencySet{
			Env:               env,
			PublicErrorFaults: faults,
		},
		AdditionalRoutes: []httpapi.RouteRegistrar{
			RegisterPublicErrorFaultRoutes(faults),
			func(mux *http.ServeMux, deps httpapi.DependencySet) error {
				mux.HandleFunc("PATCH /api/v1/records/record-1", func(w http.ResponseWriter, r *http.Request) {
					_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"ok": true})
				})
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("new public error fault handler: %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func publicErrorFaultBody() map[string]any {
	return map[string]any{
		"method":       "PATCH",
		"path":         "/api/v1/records/record-1",
		"status":       http.StatusInternalServerError,
		"code":         "unknown_public_error_probe",
		"message":      "unexpected public error",
		"consume_once": true,
	}
}

func testRuntimeEnabledEnv() map[string]string {
	return map[string]string{
		testRoutesEnabledEnv: "1",
		testRuntimeMarkerEnv: testRuntimeMarkerValue,
		testRouteTokenEnv:    testRuntimeResetToken,
	}
}
