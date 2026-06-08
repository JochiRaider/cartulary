package collaboration

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
)

func TestWebSocketLifecycleTelemetryClassifiesPublicErrors(t *testing.T) {
	tests := []struct {
		name       string
		apiErr     *auth.APIError
		wantResult string
		wantCode   string
	}{
		{name: "none", apiErr: nil, wantResult: "success"},
		{name: "client", apiErr: &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required"}, wantResult: "rejected", wantCode: "session_required"},
		{name: "conflict", apiErr: &auth.APIError{Status: http.StatusConflict, Code: "client_txn_conflict"}, wantResult: "conflict", wantCode: "client_txn_conflict"},
		{name: "server", apiErr: &auth.APIError{Status: http.StatusInternalServerError, Code: "internal_error"}, wantResult: "failed", wantCode: "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotCode := webSocketLifecycleResultForAPIError(tt.apiErr)
			if gotResult != tt.wantResult || gotCode != tt.wantCode {
				t.Fatalf("classification = (%q, %q), want (%q, %q)", gotResult, gotCode, tt.wantResult, tt.wantCode)
			}
		})
	}
}

func TestWebSocketLifecycleTelemetryClosesVocabulary(t *testing.T) {
	if got := safeWebSocketLifecycleOperation("connect"); got != "connect" {
		t.Fatalf("connect operation = %q", got)
	}
	if got := safeWebSocketLifecycleOperation("connection:10000000-0000-4000-8000-000000000001"); got != "unknown" {
		t.Fatalf("unsafe operation = %q", got)
	}
	if got := safeWebSocketLifecycleResult("timeout"); got != "timeout" {
		t.Fatalf("timeout result = %q", got)
	}
	if got := safeWebSocketLifecycleResult("raw"); got != "failed" {
		t.Fatalf("unsafe result = %q", got)
	}
}
