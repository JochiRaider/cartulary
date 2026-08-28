package harnesscontrol

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
)

const testRuntimeResetToken = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFG"

func testRuntimeEnabledEnv() map[string]string {
	return map[string]string{
		httpapi.TestRoutesEnabledEnv: "1",
		httpapi.TestRuntimeMarkerEnv: httpapi.TestRuntimeMarkerValue,
		httpapi.TestRouteTokenEnv:    testRuntimeResetToken,
	}
}

func testHTTPDependencies(deps httpapi.DependencySet) httpapi.DependencySet {
	return httpapiextensions.New(nil).Dependencies(deps)
}

func newTestRuntimeResetJSONRequest(t testing.TB, method string, url string, body any) *http.Request {
	t.Helper()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func authorizeTestRuntimeResetRequest(req *http.Request) *http.Request {
	req.Header.Set(httpapi.TestRouteTokenHeader, testRuntimeResetToken)
	return req
}

func doTestRuntimeResetRequest(t testing.TB, client *http.Client, req *http.Request) *http.Response {
	t.Helper()
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func requireTestRuntimeResetStatus(t testing.TB, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Fatalf("unexpected status: got %d want %d", resp.StatusCode, want)
	}
}

func readTestRuntimeResetJSONBody(t testing.TB, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	return body
}

func requireTestRuntimeResetSuccessEnvelope(t testing.TB, resp *http.Response, want int) map[string]any {
	t.Helper()
	requireTestRuntimeResetStatus(t, resp, want)
	if resp.Header.Get(httpapi.RequestIDHeader) == "" {
		t.Fatal("missing request id header")
	}
	body := readTestRuntimeResetJSONBody(t, resp)
	if _, ok := body["data"].(map[string]any); !ok {
		t.Fatalf("expected success envelope data object, got %T", body["data"])
	}
	return body
}

func requireTestRuntimeResetErrorEnvelope(t testing.TB, resp *http.Response, want int, wantCode string) map[string]any {
	t.Helper()
	requireTestRuntimeResetStatus(t, resp, want)
	body := readTestRuntimeResetJSONBody(t, resp)
	errorValue, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error envelope, got %#v", body)
	}
	if errorValue["code"] != wantCode {
		t.Fatalf("unexpected error code: got %#v want %q", errorValue["code"], wantCode)
	}
	return body
}
