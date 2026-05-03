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

func TestRegisterTestClockRoutesSetFixedPayloadPinsClock(t *testing.T) {
	clock, server := startTestClockRouteServer(t)
	defer server.Close()

	fixed := time.Date(2026, time.May, 3, 13, 58, 59, 246615000, time.FixedZone("EDT", -4*60*60))
	body := postClockSet(t, server.URL, map[string]any{
		"fixed_now": fixed.Format(time.RFC3339Nano),
	}, http.StatusOK)
	data := body["data"].(map[string]any)
	want := fixed.UTC()

	if got := data["now"]; got != want.Format(time.RFC3339Nano) {
		t.Fatalf("unexpected fixed now response: got %v want %s", got, want.Format(time.RFC3339Nano))
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

func startTestClockRouteServer(t testing.TB) (*TestClock, *httptest.Server) {
	t.Helper()

	clock := NewTestClock()
	handler, err := NewHandler(Options{
		AdditionalRoutes: []RouteRegistrar{RegisterTestClockRoutes(clock)},
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

func postClockSetRaw(t testing.TB, baseURL string, payload string, wantStatus int) map[string]any {
	t.Helper()

	resp, err := http.Post(baseURL+"/api/v1/test/clock/set", "application/json", bytes.NewBufferString(payload))
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
