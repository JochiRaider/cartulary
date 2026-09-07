package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInternalErrorBoundaryRemovesPrivateDiagnostics(t *testing.T) {
	const private = "secret-marker SELECT private_relation /private/storage"
	for _, constructor := range []bool{false, true} {
		t.Run(map[bool]string{false: "writer", true: "constructor"}[constructor], func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/jobs/example", nil)
			request = request.WithContext(context.WithValue(request.Context(), requestIDContextKey{}, "request-safe-correlation"))
			recorder := httptest.NewRecorder()
			if constructor {
				apiErr := InternalAPIError(errors.New(private))
				if apiErr.Message != "internal_error" {
					t.Errorf("constructor retains a private diagnostic")
				}
				WriteAPIError(recorder, request, apiErr)
			} else {
				_ = WriteErrorWithOptions(recorder, request, 503, "internal_error", private, map[string]any{"private": private}, ErrorOptions{
					Retryable: true, Conflict: map[string]any{"private": private},
				})
			}
			if recorder.Code != 500 {
				t.Fatalf("unsafe internal response: status=%d", recorder.Code)
			}
			for _, marker := range []string{"secret-marker", "SELECT private_relation", "/private/storage"} {
				if strings.Contains(recorder.Body.String(), marker) {
					t.Fatal("internal response discloses private marker")
				}
			}
			var envelope struct {
				Error struct {
					Code      string
					Message   string
					Status    int
					RequestID string `json:"request_id"`
					Retryable bool
					Details   map[string]any
					Conflict  any
				}
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}
			value := envelope.Error
			if value.Code != "internal_error" || value.Message != "internal_error" || value.Status != 500 || value.RequestID != "request-safe-correlation" || value.Retryable || value.Details == nil || len(value.Details) != 0 || value.Conflict != nil {
				t.Fatal("internal error envelope is not canonical")
			}
		})
	}
}

func TestInternalErrorBoundaryPreservesTypedPublicErrors(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/jobs/example/cancel", nil)
	_ = WriteError(recorder, request, 409, "job_cancel_rejected", "Cancellation unavailable.", map[string]any{"reason_code": "already_terminal"})
	if recorder.Code != 409 || !strings.Contains(recorder.Body.String(), "already_terminal") || !strings.Contains(recorder.Body.String(), "Cancellation unavailable.") {
		t.Fatal("typed public error changed")
	}
}
