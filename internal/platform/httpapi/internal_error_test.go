package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestInternalErrorBoundaryRemovesPrivateDiagnostics(t *testing.T) {
	const private = "secret-marker SELECT private_relation /private/storage"
	for _, constructor := range []bool{false, true} {
		t.Run(map[bool]string{false: "writer", true: "constructor"}[constructor], func(t *testing.T) {
			for _, suppliedCorrelation := range []bool{false, true} {
				t.Run(map[bool]string{false: "generated-correlation", true: "supplied-correlation"}[suppliedCorrelation], func(t *testing.T) {
					response, body := readErrorBoundaryResponse(t, suppliedCorrelation, func(w http.ResponseWriter, r *http.Request) {
						if constructor {
							apiErr := InternalAPIError(errors.New(private))
							if apiErr.Message != "internal_error" {
								t.Error("constructor retains a private diagnostic")
							}
							WriteAPIError(w, r, apiErr)
						} else {
							_ = WriteErrorWithOptions(w, r, 503, "internal_error", private, map[string]any{"private": private}, ErrorOptions{
								Retryable: true, Conflict: map[string]any{"private": private},
							})
						}
					})
					for _, marker := range []string{"secret-marker", "SELECT private_relation", "/private/storage"} {
						if strings.Contains(string(body), marker) {
							t.Fatal("internal response discloses private marker")
						}
					}
					requireErrorBoundaryEnvelope(t, response, body, suppliedCorrelation, ErrorPayload{
						Code: "internal_error", Message: "internal_error", Status: 500, Details: map[string]any{},
					})
				})
			}
		})
	}
}

func TestInternalErrorBoundaryPreservesTypedPublicErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload ErrorPayload
	}{
		{"typed-conflict", ErrorPayload{Status: 409, Code: "job_cancel_rejected", Message: "Cancellation unavailable.", Details: map[string]any{"reason_code": "already_terminal"}}},
		{"evidence-access", ErrorPayload{Status: 409, Code: "evidence_access_unavailable", Message: "Evidence access unavailable.", Details: map[string]any{"reason_code": "blob_failed"}}},
		{"unknown-public-code", ErrorPayload{Status: 502, Code: "future_public_error", Message: "Public request unavailable.", Retryable: true, Details: map[string]any{"reason_code": "future_public_error"}}},
		{"empty-details", ErrorPayload{Status: 403, Code: "forbidden", Message: "Access denied.", Details: map[string]any{}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response, body := readErrorBoundaryResponse(t, true, func(w http.ResponseWriter, r *http.Request) {
				_ = WriteErrorWithOptions(w, r, tc.payload.Status, tc.payload.Code, tc.payload.Message, tc.payload.Details, ErrorOptions{Retryable: tc.payload.Retryable})
			})
			requireErrorBoundaryEnvelope(t, response, body, true, tc.payload)
		})
	}
}

// Exercise the production handler and request-ID middleware, not an injected envelope.
func readErrorBoundaryResponse(t *testing.T, suppliedCorrelation bool, write http.HandlerFunc) (*http.Response, []byte) {
	t.Helper()
	handler, err := NewHandler(Options{
		Dependencies:      testExtensionDependencies(nil),
		RequestIDSequence: func() string { return "server-safe-correlation" },
		AdditionalRoutes: []RouteRegistrar{func(mux *http.ServeMux, _ DependencySet) error {
			mux.HandleFunc("POST /api/v1/error-boundary", write)
			return nil
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/error-boundary", nil)
	if err != nil {
		t.Fatal(err)
	}
	if suppliedCorrelation {
		request.Header.Set(RequestIDHeader, "client-safe-correlation")
	}
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response, body
}

func requireErrorBoundaryEnvelope(t *testing.T, response *http.Response, body []byte, suppliedCorrelation bool, expected ErrorPayload) {
	t.Helper()
	correlation := "server-safe-correlation"
	if suppliedCorrelation {
		correlation = "client-safe-correlation"
	}
	expected.RequestID = correlation
	if response.StatusCode != expected.Status || response.Header.Get(RequestIDHeader) != correlation || response.Header.Get("Content-Type") != "application/json" {
		t.Fatal("public status, content type or request correlation changed")
	}
	var actual map[string]any
	if err := json.Unmarshal(body, &actual); err != nil {
		t.Fatal(err)
	}
	expectedJSON, err := json.Marshal(ErrorEnvelope{Error: expected})
	if err != nil {
		t.Fatal(err)
	}
	var expectedObject map[string]any
	if err := json.Unmarshal(expectedJSON, &expectedObject); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expectedObject) {
		t.Fatal("public error envelope differs from the complete expected contract")
	}
}
