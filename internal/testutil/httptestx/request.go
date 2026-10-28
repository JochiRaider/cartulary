package httptestx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

// JSONReader serializes ordinary JSON data without interpreting its fields.
// Malformed-input tests should supply literal bytes instead.
func JSONReader(t testing.TB, body any) io.Reader {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	return bytes.NewReader(payload)
}

func NewJSONRequest(t testing.TB, method string, url string, body any) *http.Request {
	t.Helper()

	var reader io.Reader
	if body != nil {
		reader = JSONReader(t, body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// Do sends a request with the supplied client. The caller owns the response body.
func Do(t testing.TB, client *http.Client, req *http.Request) *http.Response {
	t.Helper()

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

// DoJSON encodes an optional JSON body and uses the supplied client unchanged.
func DoJSON(t testing.TB, client *http.Client, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return Do(t, client, req)
}

// DoRawJSON preserves literal request bytes, including intentionally malformed JSON.
func DoRawJSON(t testing.TB, client *http.Client, method string, url string, body string, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("create raw JSON request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(req)
	}
	return Do(t, client, req)
}

func WithCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func WithHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}
