package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTestClockFixedTimeAndAdvance(t *testing.T) {
	clock := NewTestClock()
	fixed := time.Date(2026, time.April, 29, 13, 35, 4, 261991000, time.FixedZone("EDT", -4*60*60))
	wantFixed := fixed.UTC()

	if got := clock.SetFixed(fixed); !got.Equal(wantFixed) {
		t.Fatalf("SetFixed returned %s, want %s", got, wantFixed)
	}
	if got := clock.Now(); !got.Equal(wantFixed) {
		t.Fatalf("Now returned %s, want fixed %s", got, wantFixed)
	}

	wantAdvanced := wantFixed.Add(1500 * time.Millisecond)
	if got := clock.Advance(1500 * time.Millisecond); !got.Equal(wantAdvanced) {
		t.Fatalf("Advance returned %s, want %s", got, wantAdvanced)
	}
	if got := clock.Now(); !got.Equal(wantAdvanced) {
		t.Fatalf("Now returned %s, want advanced fixed %s", got, wantAdvanced)
	}
}

func TestTestClockSetOffsetClearsFixedTime(t *testing.T) {
	clock := NewTestClock()
	clock.SetFixed(time.Date(2035, time.January, 1, 0, 0, 0, 0, time.UTC))

	before := time.Now().UTC().Add(-100 * time.Millisecond)
	if got := clock.SetOffset(0); got.Before(before) {
		t.Fatalf("SetOffset returned stale time %s before %s", got, before)
	}
	after := time.Now().UTC().Add(100 * time.Millisecond)
	if got := clock.Now(); got.Before(before) || got.After(after) {
		t.Fatalf("SetOffset should clear fixed time; Now returned %s outside [%s, %s]", got, before, after)
	}
}

func TestTestClockResetClearsFixedAndOffset(t *testing.T) {
	clock := NewTestClock()
	clock.SetFixed(time.Date(2035, time.January, 1, 0, 0, 0, 0, time.UTC))
	clock.Advance(time.Hour)

	before := time.Now().UTC().Add(-100 * time.Millisecond)
	if got := clock.Reset(); got.Before(before) {
		t.Fatalf("Reset returned stale time %s before %s", got, before)
	}
	after := time.Now().UTC().Add(100 * time.Millisecond)
	snapshot := clock.Snapshot()
	if snapshot.Mode != testClockModeWall {
		t.Fatalf("Reset should restore wall mode, got %q", snapshot.Mode)
	}
	if snapshot.FixedNow != nil {
		t.Fatalf("Reset should clear fixed time, got %s", snapshot.FixedNow)
	}
	if snapshot.OffsetSeconds != 0 {
		t.Fatalf("Reset should clear offset, got %d", snapshot.OffsetSeconds)
	}
	if snapshot.Now.Before(before) || snapshot.Now.After(after) {
		t.Fatalf("Reset should restore wall time; snapshot now %s outside [%s, %s]", snapshot.Now, before, after)
	}
}

func TestRegisterTestClockRoutesRejectsUnauthenticatedMutation(t *testing.T) {
	_, server := startTestClockRouteServer(t)
	defer server.Close()

	body := postClockSetRawWithToken(t, server.URL, `{"offset_seconds":0}`, "", http.StatusForbidden)
	errorPayload := body["error"].(map[string]any)
	if got := errorPayload["code"]; got != "test_route_forbidden" {
		t.Fatalf("unexpected error code: got %v", got)
	}

	body = postClockSetRawWithToken(t, server.URL, `{"offset_seconds":0}`, "ABCDEFGabcdefghijklmnopqrstuvwxyz0123456789", http.StatusForbidden)
	errorPayload = body["error"].(map[string]any)
	if got := errorPayload["code"]; got != "test_route_forbidden" {
		t.Fatalf("unexpected wrong-token error code: got %v", got)
	}
}

func TestRegisterTestClockRoutesRequiresHarnessOwnedRuntime(t *testing.T) {
	clock := NewTestClock()
	_, err := NewHandler(Options{
		AdditionalRoutes: []RouteRegistrar{RegisterTestClockRoutes(clock)},
		Dependencies: testExtensionDependenciesWith(DependencySet{
			Env: map[string]string{
				TestRoutesEnabledEnv: "1",
				TestRouteTokenEnv:    testClockRouteToken,
			},
		}, nil),
	})
	if err == nil {
		t.Fatal("expected missing harness runtime marker to fail handler setup")
	}
}

func TestRegisterTestClockRoutesSetFixedPayloadPinsClock(t *testing.T) {
	clock, server := startTestClockRouteServer(t)
	defer server.Close()

	fixed := time.Date(2026, time.May, 3, 13, 58, 59, 246615000, time.FixedZone("EDT", -4*60*60))
	body := postClockSet(t, server.URL, map[string]any{
		"fixed_now": fixed.Format(time.RFC3339Nano),
	}, http.StatusOK)
	data := body["data"].(map[string]any)
	want := fixed.UTC()

	if got := data["schema_id"]; got != testClockControlSchemaID {
		t.Fatalf("unexpected schema_id: got %v", got)
	}
	if got := data["mode"]; got != testClockModeFixed {
		t.Fatalf("unexpected clock mode: got %v", got)
	}
	if got := data["now"]; got != want.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected fixed now response: got %v want %s", got, want.Format(time.RFC3339Nano))
	}
	if got := data["fixed_now"]; got != want.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected fixed_now response: got %v want %s", got, want.Format(time.RFC3339Nano))
	}
	if got := data["offset_seconds"]; got != float64(0) {
		t.Fatalf("fixed clock must report zero offset_seconds, got %v", got)
	}
	if got := clock.Now(); !got.Equal(want) {
		t.Fatalf("clock.Now() = %s, want %s", got, want)
	}
}

func TestRegisterTestClockRoutesSetOffsetPayloadClearsFixedClock(t *testing.T) {
	clock, server := startTestClockRouteServer(t)
	defer server.Close()

	clock.SetFixed(time.Date(2035, time.January, 1, 0, 0, 0, 0, time.UTC))
	before := time.Now().UTC().Add(-100 * time.Millisecond)
	body := postClockSet(t, server.URL, map[string]any{
		"offset_seconds": int64(0),
	}, http.StatusOK)
	after := time.Now().UTC().Add(100 * time.Millisecond)
	data := body["data"].(map[string]any)

	if got := data["schema_id"]; got != testClockControlSchemaID {
		t.Fatalf("unexpected schema_id: got %v", got)
	}
	if got := data["mode"]; got != testClockModeWall {
		t.Fatalf("unexpected clock mode: got %v", got)
	}
	if got := data["offset_seconds"]; got != float64(0) {
		t.Fatalf("unexpected offset_seconds response: got %v", got)
	}
	gotNow, err := time.Parse(time.RFC3339Nano, data["now"].(string))
	if err != nil {
		t.Fatalf("parse response now: %v", err)
	}
	if gotNow.Before(before) || gotNow.After(after) {
		t.Fatalf("SetOffset response now outside wall-clock window: got %s want between %s and %s", gotNow, before, after)
	}
	afterClockRead := time.Now().UTC().Add(100 * time.Millisecond)
	if got := clock.Now(); got.Before(before) || got.After(afterClockRead) {
		t.Fatalf("offset reset should clear fixed time; Now returned %s outside [%s, %s]", got, before, afterClockRead)
	}
}

func TestRegisterTestClockRoutesResetClearsClock(t *testing.T) {
	clock, server := startTestClockRouteServer(t)
	defer server.Close()

	clock.SetFixed(time.Date(2035, time.January, 1, 0, 0, 0, 0, time.UTC))
	before := time.Now().UTC().Add(-100 * time.Millisecond)
	body := postClockReset(t, server.URL, map[string]any{}, http.StatusOK)
	after := time.Now().UTC().Add(100 * time.Millisecond)
	data := body["data"].(map[string]any)

	if got := data["schema_id"]; got != testClockControlSchemaID {
		t.Fatalf("unexpected schema_id: got %v", got)
	}
	if got := data["mode"]; got != testClockModeWall {
		t.Fatalf("unexpected clock mode: got %v", got)
	}
	if _, ok := data["fixed_now"]; ok {
		t.Fatalf("reset response must omit fixed_now: %#v", data)
	}
	gotNow, err := time.Parse(time.RFC3339Nano, data["now"].(string))
	if err != nil {
		t.Fatalf("parse response now: %v", err)
	}
	if gotNow.Before(before) || gotNow.After(after) {
		t.Fatalf("Reset response now outside wall-clock window: got %s want between %s and %s", gotNow, before, after)
	}
	if got := clock.Snapshot(); got.Mode != testClockModeWall || got.OffsetSeconds != 0 || got.FixedNow != nil {
		t.Fatalf("reset route should clear fixed/offset state: %#v", got)
	}
}

func TestRegisterTestClockRoutesStateReportsWithoutMutation(t *testing.T) {
	clock, server := startTestClockRouteServer(t)
	defer server.Close()

	fixed := time.Date(2026, time.November, 1, 5, 30, 0, 123456789, time.UTC)
	clock.SetFixed(fixed)
	body := getClockState(t, server.URL, http.StatusOK)
	data := body["data"].(map[string]any)

	if got := data["schema_id"]; got != testClockControlSchemaID {
		t.Fatalf("unexpected schema_id: got %v", got)
	}
	if got := data["mode"]; got != testClockModeFixed {
		t.Fatalf("unexpected clock mode: got %v", got)
	}
	if got := data["now"]; got != fixed.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected now: got %v want %s", got, fixed.Format(time.RFC3339Nano))
	}
	if got := data["fixed_now"]; got != fixed.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected fixed_now: got %v want %s", got, fixed.Format(time.RFC3339Nano))
	}
	if got := clock.Now(); !got.Equal(fixed) {
		t.Fatalf("state route mutated clock: got %s want %s", got, fixed)
	}
}

func TestRegisterTestClockRoutesParsesRFC3339NanoFixedNow(t *testing.T) {
	clock, server := startTestClockRouteServer(t)
	defer server.Close()

	fixedText := "2026-05-03T17:58:59.246615123Z"
	postClockSet(t, server.URL, map[string]any{
		"fixed_now": fixedText,
	}, http.StatusOK)
	want, err := time.Parse(time.RFC3339Nano, fixedText)
	if err != nil {
		t.Fatalf("parse test fixed timestamp: %v", err)
	}
	if got := clock.Now(); !got.Equal(want) {
		t.Fatalf("clock.Now() = %s, want %s", got, want)
	}
}

func TestRegisterTestClockRoutesRejectsInvalidPayloads(t *testing.T) {
	_, server := startTestClockRouteServer(t)
	defer server.Close()

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "neither command", body: `{}`},
		{name: "multiple commands", body: `{"offset_seconds":0,"fixed_now":"2026-05-03T17:58:59Z"}`},
		{name: "unknown field", body: `{"offset_seconds":0,"unexpected":true}`},
		{name: "invalid fixed_now", body: `{"fixed_now":"not-a-time"}`},
		{name: "trailing json", body: `{"offset_seconds":0} {}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := postClockSetRaw(t, server.URL, tc.body, http.StatusBadRequest)
			errorPayload := body["error"].(map[string]any)
			if got := errorPayload["code"]; got != "invalid_mutation_payload" {
				t.Fatalf("unexpected error code: got %v", got)
			}
		})
	}
}

func TestRegisterTestClockRoutesRejectsInvalidResetPayloads(t *testing.T) {
	_, server := startTestClockRouteServer(t)
	defer server.Close()

	body := postClockResetRaw(t, server.URL, `{"unexpected":true}`, http.StatusBadRequest)
	errorPayload := body["error"].(map[string]any)
	if got := errorPayload["code"]; got != "invalid_mutation_payload" {
		t.Fatalf("unexpected error code: got %v", got)
	}
}

func startTestClockRouteServer(t testing.TB) (*TestClock, *httptest.Server) {
	t.Helper()

	clock := NewTestClock()
	handler, err := NewHandler(Options{
		AdditionalRoutes: []RouteRegistrar{RegisterTestClockRoutes(clock)},
		Dependencies: testExtensionDependenciesWith(DependencySet{
			Env: map[string]string{
				TestRoutesEnabledEnv: "1",
				TestRuntimeMarkerEnv: TestRuntimeMarkerValue,
				TestRouteTokenEnv:    testClockRouteToken,
			},
		}, nil),
	})
	if err != nil {
		t.Fatalf("new test clock handler: %v", err)
	}
	return clock, httptest.NewServer(handler)
}

func postClockSet(t testing.TB, baseURL string, payload any, wantStatus int) map[string]any {
	t.Helper()

	requestBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal clock set payload: %v", err)
	}
	return postClockSetRaw(t, baseURL, string(requestBody), wantStatus)
}

const testClockRouteToken = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"

func postClockSetRaw(t testing.TB, baseURL string, payload string, wantStatus int) map[string]any {
	t.Helper()
	return postClockSetRawWithToken(t, baseURL, payload, testClockRouteToken, wantStatus)
}

func postClockSetRawWithToken(t testing.TB, baseURL string, payload string, token string, wantStatus int) map[string]any {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/test/clock/set", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("build clock set request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set(TestRouteTokenHeader, token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post clock set: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, wantStatus)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode clock set response: %v", err)
	}
	return body
}

func postClockReset(t testing.TB, baseURL string, payload any, wantStatus int) map[string]any {
	t.Helper()

	requestBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal clock reset payload: %v", err)
	}
	return postClockResetRaw(t, baseURL, string(requestBody), wantStatus)
}

func postClockResetRaw(t testing.TB, baseURL string, payload string, wantStatus int) map[string]any {
	t.Helper()

	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/test/clock/reset", bytes.NewBufferString(payload))
	if err != nil {
		t.Fatalf("build clock reset request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(TestRouteTokenHeader, testClockRouteToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post clock reset: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, wantStatus)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode clock reset response: %v", err)
	}
	return body
}

func getClockState(t testing.TB, baseURL string, wantStatus int) map[string]any {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/test/clock/state", nil)
	if err != nil {
		t.Fatalf("build clock state request: %v", err)
	}
	req.Header.Set(TestRouteTokenHeader, testClockRouteToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("get clock state: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, wantStatus)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode clock state response: %v", err)
	}
	return body
}
