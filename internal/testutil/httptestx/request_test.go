package httptestx

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRequestEncodingAndOptions(t *testing.T) {
	type observed struct {
		method, body string
		header       http.Header
	}
	requests := make(chan observed, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read test request: %v", err)
		}
		requests <- observed{r.Method, string(body), r.Header.Clone()}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	}))
	t.Cleanup(server.Close)
	for _, tc := range []struct {
		name string
		body any
		want string
	}{
		{"absent", nil, ""},
		{"object", map[string]any{"value": "unchanged"}, `{"value":"unchanged"}`},
		{"array", []int{1, 2}, `[1,2]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp := DoJSON(t, server.Client(), http.MethodPatch, server.URL, tc.body,
				WithHeader("X-Order", "first"), WithHeader("X-Order", "last"),
				WithCookies(nil, &http.Cookie{Name: "session", Value: "test-value"}),
			)
			if string(ReadResponseBody(t, resp, http.StatusAccepted)) != "accepted" {
				t.Fatal("response bytes changed")
			}
			got := <-requests
			if got.method != http.MethodPatch || got.body != tc.want {
				t.Fatalf("method=%s body=%q", got.method, got.body)
			}
			if got.header.Get("Content-Type") != "application/json" || got.header.Get("X-Order") != "last" || got.header.Get("Cookie") != "session=test-value" {
				t.Fatalf("request options changed: %v", got.header)
			}
			if got.header.Get("Authorization") != "" {
				t.Fatal("unexpected implicit authentication")
			}
		})
	}
	for _, raw := range []string{`{"invalid":`, `{"duplicate":1,"duplicate":2}`, "null", ""} {
		resp := DoRawJSON(t, server.Client(), http.MethodPost, server.URL, raw, WithHeader("Content-Type", "application/problem+json"))
		ReadResponseBody(t, resp, http.StatusAccepted)
		got := <-requests
		if got.body != raw || got.header.Get("Content-Type") != "application/problem+json" {
			t.Fatal("raw payload/options interpreted")
		}
	}
	request := NewJSONRequest(t, http.MethodGet, server.URL, nil)
	if request.Body != nil {
		t.Fatal("nil input must leave body absent")
	}
	payload, err := io.ReadAll(JSONReader(t, nil))
	if err != nil || string(payload) != "null" {
		t.Fatalf("explicit JSON null = %q, %v", payload, err)
	}
}

func TestRequestsPreserveClientPolicy(t *testing.T) {
	var followed atomic.Bool
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/destination" {
			followed.Store(true)
			return
		}
		w.Header().Set("Location", "/destination")
		w.WriteHeader(http.StatusSeeOther)
	}))
	t.Cleanup(server.Close)
	client := server.Client()
	client.Timeout = time.Minute
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	transport := client.Transport
	for _, raw := range []bool{false, true} {
		var resp *http.Response
		if raw {
			resp = DoRawJSON(t, client, http.MethodPost, server.URL, `{"literal":`)
		} else {
			resp = DoJSON(t, client, http.MethodGet, server.URL, nil)
		}
		ReadResponseBody(t, resp, http.StatusSeeOther)
		if resp.Header.Get("Location") != "/destination" {
			t.Fatal("redirect response lost")
		}
	}
	request, err := http.NewRequest(http.MethodPost, server.URL, strings.NewReader("ordinary form"))
	if err != nil {
		t.Fatal(err)
	}
	ReadResponseBody(t, Do(t, client, request), http.StatusSeeOther)
	if followed.Load() || client.Transport != transport || client.Timeout != time.Minute || client.CheckRedirect == nil {
		t.Fatal("caller client policy changed")
	}
}
