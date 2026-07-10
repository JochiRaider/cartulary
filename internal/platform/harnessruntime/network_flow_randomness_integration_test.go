package harnessruntime

import (
	"bytes"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestNetworkFlowRandomnessRouteDisabledByDefault(t *testing.T) {
	random := NewNetworkFlowRandomnessRegistry()
	server := startNetworkFlowRandomnessHTTPServer(t, map[string]string{}, random)

	resp := doTestRuntimeResetRequest(t, server.Client(), newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-randomness", networkFlowRandomnessBody()))
	defer resp.Body.Close()
	requireTestRuntimeResetStatus(t, resp, http.StatusNotFound)
}

func TestNetworkFlowRandomnessRouteRequiresHarnessAuthorization(t *testing.T) {
	service := &networkFlowRandomnessService{
		guard: httpapi.TestRouteGuard{
			Token:        testRuntimeResetToken,
			ExpectedHost: "127.0.0.1:8080",
			AllowedOrigins: map[string]struct{}{
				"http://127.0.0.1:8080": {},
				"http://127.0.0.1:4173": {},
			},
		},
		random: NewNetworkFlowRandomnessRegistry(),
	}

	missingOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-randomness", networkFlowRandomnessBody()))
	recorder := httptest.NewRecorder()
	service.handleArm(recorder, missingOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongOrigin := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-randomness", networkFlowRandomnessBody()))
	wrongOrigin.Header.Set("Origin", "http://evil.example.test")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongOrigin)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	wrongHost := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-randomness", networkFlowRandomnessBody()))
	wrongHost.Host = "evil.example.test"
	wrongHost.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, wrongHost)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusForbidden, "test_route_forbidden")

	allowed := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "http://127.0.0.1:8080/api/v1/test/runtime/network-flow-randomness", networkFlowRandomnessBody()))
	allowed.Header.Set("Origin", "http://127.0.0.1:4173")
	recorder = httptest.NewRecorder()
	service.handleArm(recorder, allowed)
	requireTestRuntimeResetStatus(t, recorder.Result(), http.StatusCreated)
}

func TestNetworkFlowRandomnessRouteArmsDeterministicCollisionStream(t *testing.T) {
	random := NewNetworkFlowRandomnessRegistry()
	server := startNetworkFlowRandomnessHTTPServer(t, testRuntimeEnabledEnv(), random)

	body := networkFlowRandomnessBody()
	body["values"] = []string{
		"01234567-89ab-cdef-0123-456789abcdef",
		"01234567-89ab-cdef-0123-456789abcdef",
	}
	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-randomness", body))
	response := requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)
	data := response["data"].(map[string]any)
	if data["schema_id"] != testNetworkFlowRandomnessSchemaID {
		t.Fatalf("unexpected schema_id: %#v", data)
	}
	if data["stream"] != NetworkFlowRandomStreamTableID || data["value_kind"] != NetworkFlowRandomValueKindUUID {
		t.Fatalf("unexpected randomness response: %#v", data)
	}
	if _, leaked := data["values"]; leaked {
		t.Fatalf("randomness response must not echo deterministic values: %#v", data)
	}
	if data["value_count"] != float64(2) || data["remaining_count"] != float64(2) || data["consume_once"] != true || data["exhaustion"] != networkFlowRandomnessExhaustionFailClosed {
		t.Fatalf("unexpected count/control response: %#v", data)
	}

	if _, ok, err := random.ConsumeNetworkFlowRandomUUID(NetworkFlowRandomStreamRowID); err != nil || ok {
		t.Fatalf("wrong stream must be absent, got ok=%v err=%v", ok, err)
	}
	first, ok, err := random.ConsumeNetworkFlowRandomUUID(NetworkFlowRandomStreamTableID)
	if err != nil || !ok {
		t.Fatalf("expected first deterministic UUID, ok=%v err=%v", ok, err)
	}
	second, ok, err := random.ConsumeNetworkFlowRandomUUID(NetworkFlowRandomStreamTableID)
	if err != nil || !ok {
		t.Fatalf("expected second deterministic UUID, ok=%v err=%v", ok, err)
	}
	if first != second {
		t.Fatalf("duplicate fixture values must be preserved for collision injection: %s != %s", first, second)
	}
	if state, ok := random.NetworkFlowRandomnessState(NetworkFlowRandomStreamTableID); !ok || state.RemainingCount != 0 || state.ValueCount != 2 {
		t.Fatalf("unexpected exhausted stream state: %#v ok=%v", state, ok)
	}
	if _, ok, err := random.ConsumeNetworkFlowRandomUUID(NetworkFlowRandomStreamTableID); ok || !errors.Is(err, ErrNetworkFlowRandomnessExhausted) {
		t.Fatalf("exhausted deterministic stream must fail closed, ok=%v err=%v", ok, err)
	}
}

func TestNetworkFlowRandomnessRouteRejectsSecondArmWhileStreamRegistered(t *testing.T) {
	random := NewNetworkFlowRandomnessRegistry()
	server := startNetworkFlowRandomnessHTTPServer(t, testRuntimeEnabledEnv(), random)

	first := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-randomness", networkFlowRandomnessBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), first), http.StatusCreated)

	secondBody := networkFlowRandomnessBody()
	secondBody["values"] = []string{"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"}
	second := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-randomness", secondBody))
	requireTestRuntimeResetErrorEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), second), http.StatusConflict, "test_network_flow_random_stream_already_armed")

	value, ok, err := random.ConsumeNetworkFlowRandomUUID(NetworkFlowRandomStreamTableID)
	if err != nil || !ok || value.String() != "01234567-89ab-cdef-0123-456789abcdef" {
		t.Fatalf("first stream must remain armed after rejected replacement, value=%s ok=%v err=%v", value, ok, err)
	}
}

func TestNetworkFlowRandomnessRouteConsumesHexBytes(t *testing.T) {
	random := NewNetworkFlowRandomnessRegistry()
	server := startNetworkFlowRandomnessHTTPServer(t, testRuntimeEnabledEnv(), random)

	body := map[string]any{
		"stream":       NetworkFlowRandomStreamSafeDigestNonce,
		"value_kind":   NetworkFlowRandomValueKindHexBytes,
		"values":       []string{"000102ff"},
		"consume_once": true,
		"exhaustion":   networkFlowRandomnessExhaustionFailClosed,
	}
	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-randomness", body))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	value, ok, err := random.ConsumeNetworkFlowRandomHexBytes(NetworkFlowRandomStreamSafeDigestNonce)
	if err != nil || !ok {
		t.Fatalf("expected deterministic hex bytes, ok=%v err=%v", ok, err)
	}
	if got := hex.EncodeToString(value); got != "000102ff" {
		t.Fatalf("unexpected deterministic bytes: %s", got)
	}
	if _, ok, err := random.ConsumeNetworkFlowRandomString(NetworkFlowRandomStreamSafeDigestNonce); ok || !errors.Is(err, ErrNetworkFlowRandomnessKindMismatch) {
		t.Fatalf("wrong value-kind consumer must fail closed, ok=%v err=%v", ok, err)
	}
}

func TestNetworkFlowRandomnessRouteRejectsInvalidRequests(t *testing.T) {
	service := &networkFlowRandomnessService{
		guard:  httpapi.TestRouteGuard{Token: testRuntimeResetToken},
		random: NewNetworkFlowRandomnessRegistry(),
	}
	for _, body := range []map[string]any{
		{"stream": "network_flow.unknown", "value_kind": NetworkFlowRandomValueKindUUID, "values": []string{"01234567-89ab-cdef-0123-456789abcdef"}, "consume_once": true, "exhaustion": networkFlowRandomnessExhaustionFailClosed},
		{"stream": NetworkFlowRandomStreamTableID, "value_kind": "integer", "values": []string{"01234567-89ab-cdef-0123-456789abcdef"}, "consume_once": true, "exhaustion": networkFlowRandomnessExhaustionFailClosed},
		{"stream": NetworkFlowRandomStreamTableID, "value_kind": NetworkFlowRandomValueKindUUID, "values": []string{}, "consume_once": true, "exhaustion": networkFlowRandomnessExhaustionFailClosed},
		{"stream": NetworkFlowRandomStreamTableID, "value_kind": NetworkFlowRandomValueKindUUID, "values": []string{"01234567-89AB-cdef-0123-456789abcdef"}, "consume_once": true, "exhaustion": networkFlowRandomnessExhaustionFailClosed},
		{"stream": NetworkFlowRandomStreamCursorNonce, "value_kind": NetworkFlowRandomValueKindToken, "values": []string{"bad token"}, "consume_once": true, "exhaustion": networkFlowRandomnessExhaustionFailClosed},
		{"stream": NetworkFlowRandomStreamSafeDigestNonce, "value_kind": NetworkFlowRandomValueKindHexBytes, "values": []string{"FF"}, "consume_once": true, "exhaustion": networkFlowRandomnessExhaustionFailClosed},
		{"stream": NetworkFlowRandomStreamTableID, "value_kind": NetworkFlowRandomValueKindUUID, "values": []string{"01234567-89ab-cdef-0123-456789abcdef"}, "consume_once": false, "exhaustion": networkFlowRandomnessExhaustionFailClosed},
		{"stream": NetworkFlowRandomStreamTableID, "value_kind": NetworkFlowRandomValueKindUUID, "values": []string{"01234567-89ab-cdef-0123-456789abcdef"}, "consume_once": true, "exhaustion": "fallback"},
		{"stream": NetworkFlowRandomStreamTableID, "value_kind": NetworkFlowRandomValueKindUUID, "values": []string{"01234567-89ab-cdef-0123-456789abcdef"}, "consume_once": true, "exhaustion": networkFlowRandomnessExhaustionFailClosed, "unexpected": true},
	} {
		req := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, "/api/v1/test/runtime/network-flow-randomness", body))
		recorder := httptest.NewRecorder()
		service.handleArm(recorder, req)
		requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusBadRequest, "invalid_network_flow_randomness_request")
	}
}

func TestTestRuntimeResetClearsNetworkFlowRandomness(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PreparePackageDatabaseT(t, "test-network-flow-randomness-reset")
	s3Harness := s3test.Start(t)
	bucket := prepareTestRuntimeResetBucket(t, s3Harness, "test-network-flow-randomness-reset")

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	random := NewNetworkFlowRandomnessRegistry()
	_, server := startTestRuntimeResetServerWithHTTPDeps(t, env, []httpapi.RouteRegistrar{
		RegisterTestRuntimeResetRoute(random.Clear),
		RegisterNetworkFlowRandomnessRoutes(random),
	}, httpapi.DependencySet{})

	arm := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-randomness", networkFlowRandomnessBody()))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), arm), http.StatusCreated)

	reset := authorizeTestRuntimeResetRequest(newTestRuntimeResetJSONRequest(t, http.MethodPost, server.URL+"/api/v1/test/runtime/reset", nil))
	requireTestRuntimeResetSuccessEnvelope(t, doTestRuntimeResetRequest(t, server.Client(), reset), http.StatusOK)

	if _, ok, err := random.ConsumeNetworkFlowRandomUUID(NetworkFlowRandomStreamTableID); err != nil || ok {
		t.Fatalf("runtime reset must clear armed Network Flow randomness, ok=%v err=%v", ok, err)
	}
}

func startNetworkFlowRandomnessHTTPServer(t testing.TB, env map[string]string, random *NetworkFlowRandomnessRegistry) *httptest.Server {
	t.Helper()
	handler, err := httpapi.NewHandler(httpapi.Options{
		Dependencies: httpapi.DependencySet{
			Env: env,
		},
		AdditionalRoutes: []httpapi.RouteRegistrar{
			RegisterNetworkFlowRandomnessRoutes(random),
		},
	})
	if err != nil {
		t.Fatalf("new Network Flow randomness handler: %v", err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func networkFlowRandomnessBody() map[string]any {
	return map[string]any{
		"stream":       NetworkFlowRandomStreamTableID,
		"value_kind":   NetworkFlowRandomValueKindUUID,
		"values":       []string{"01234567-89ab-cdef-0123-456789abcdef"},
		"consume_once": true,
		"exhaustion":   networkFlowRandomnessExhaustionFailClosed,
	}
}

func TestNetworkFlowRandomnessValueValidationDoesNotAcceptBinaryBodies(t *testing.T) {
	service := &networkFlowRandomnessService{
		guard:  httpapi.TestRouteGuard{Token: testRuntimeResetToken},
		random: NewNetworkFlowRandomnessRegistry(),
	}
	req := authorizeTestRuntimeResetRequest(httptest.NewRequest(http.MethodPost, "/api/v1/test/runtime/network-flow-randomness", bytes.NewBuffer([]byte{0xff})))
	recorder := httptest.NewRecorder()
	service.handleArm(recorder, req)
	requireTestRuntimeResetErrorEnvelope(t, recorder.Result(), http.StatusBadRequest, "invalid_network_flow_randomness_request")
}
