package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const testOrigin = "http://localhost:5173"

func TestProxyPreflightExactPolicy(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(upstream.Close)
	handler := newProxyHandler(parseTestURL(t, upstream.URL), testOrigin)

	req := httptest.NewRequest(http.MethodOptions, "http://object-store.test/cartulary/object.bin", nil)
	req.Header.Set("Origin", testOrigin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPut)
	req.Header.Set("Access-Control-Request-Headers", "x-amz-checksum-sha256, content-type")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("preflight status got %d want %d", resp.Code, http.StatusNoContent)
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Fatalf("allow origin got %q want %q", got, testOrigin)
	}
	if got := resp.Header().Get("Access-Control-Allow-Methods"); got != "PUT, OPTIONS" {
		t.Fatalf("allow methods got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Headers"); got != "content-type, x-amz-checksum-sha256" {
		t.Fatalf("allow headers got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("credentials header must be absent, got %q", got)
	}
	if got := resp.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Fatalf("max age got %q want 600", got)
	}
}

func TestProxyPreflightRejectsDisallowedOriginMethodAndHeader(t *testing.T) {
	upstream := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(upstream.Close)
	handler := newProxyHandler(parseTestURL(t, upstream.URL), testOrigin)

	tests := []struct {
		name    string
		origin  string
		method  string
		headers string
	}{
		{name: "null origin", origin: "null", method: http.MethodPut, headers: "content-type"},
		{name: "extra method", origin: testOrigin, method: http.MethodGet, headers: "content-type"},
		{name: "extra header", origin: testOrigin, method: http.MethodPut, headers: "content-type, authorization"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "http://object-store.test/cartulary/object.bin", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Access-Control-Request-Method", tc.method)
			req.Header.Set("Access-Control-Request-Headers", tc.headers)
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			if resp.Code < 400 {
				t.Fatalf("preflight status got %d want rejection", resp.Code)
			}
			if got := resp.Header().Get("Access-Control-Allow-Origin"); got != "" {
				t.Fatalf("rejected preflight must not grant origin, got %q", got)
			}
		})
	}
}

func TestProxyPUTPreservesHostAndNormalizesCORS(t *testing.T) {
	var gotHost string
	var gotBody []byte
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHost = r.Host
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "etag, x-amz-request-id")
		w.Header().Set("ETag", `"abc"`)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(upstream.Close)
	handler := newProxyHandler(parseTestURL(t, upstream.URL), testOrigin)

	req := httptest.NewRequest(http.MethodPut, "http://signed-host.test/cartulary/object.bin", strings.NewReader("payload"))
	req.Header.Set("Origin", testOrigin)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("PUT status got %d want %d", resp.Code, http.StatusOK)
	}
	if gotHost != "signed-host.test" {
		t.Fatalf("upstream host got %q want signed host", gotHost)
	}
	if string(gotBody) != "payload" {
		t.Fatalf("upstream body got %q", string(gotBody))
	}
	if got := resp.Header().Get("Access-Control-Allow-Origin"); got != testOrigin {
		t.Fatalf("allow origin got %q want %q", got, testOrigin)
	}
	if got := resp.Header().Get("Access-Control-Expose-Headers"); got != "etag" {
		t.Fatalf("expose headers got %q want etag", got)
	}
}

func parseTestURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse test URL: %v", err)
	}
	return parsed
}
