package auth

import (
	"net/http"
	"strings"
	"testing"
)

func TestPhase1_LoginRequestShape_U_1_01(t *testing.T) {
	t.Run("rejects unsupported second-factor kind", func(t *testing.T) {
		_, apiErr := DecodeLoginRequest(strings.NewReader(`{
			"username":"analyst@example.test",
			"password":"exact password",
			"second_factor":{"kind":"webauthn","assertion":{"code":"123456"}}
		}`))
		requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_auth_request", "second_factor.kind")
	})

	t.Run("rejects forbidden top-level member", func(t *testing.T) {
		_, apiErr := DecodeLoginRequest(strings.NewReader(`{
			"username":"analyst@example.test",
			"password":"exact password",
			"client_txn_id":"forbidden"
		}`))
		requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_auth_request", "client_txn_id")
	})

	t.Run("rejects malformed totp assertion", func(t *testing.T) {
		_, apiErr := DecodeLoginRequest(strings.NewReader(`{
			"username":"analyst@example.test",
			"password":"exact password",
			"second_factor":{"kind":"totp","assertion":{"code":"12 3456"}}
		}`))
		requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_auth_request", "second_factor.assertion.code")
	})
}

func requireAPIError(t testing.TB, apiErr *APIError, wantStatus int, wantCode string, wantField string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
	}
	if apiErr.Status != wantStatus {
		t.Fatalf("unexpected status: got %d want %d", apiErr.Status, wantStatus)
	}
	if apiErr.Code != wantCode {
		t.Fatalf("unexpected code: got %q want %q", apiErr.Code, wantCode)
	}
	if wantField == "" {
		return
	}
	if got := apiErr.Details["field"]; got != wantField {
		t.Fatalf("unexpected field detail: got %v want %s", got, wantField)
	}
}

func requireNoSecretKeys(t testing.TB, value any, forbidden ...string) {
	t.Helper()

	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			for _, forbiddenKey := range forbidden {
				if key == forbiddenKey {
					t.Fatalf("unexpected secret-bearing field %q in %#v", key, typed)
				}
			}
			requireNoSecretKeys(t, item, forbidden...)
		}
	case []any:
		for _, item := range typed {
			requireNoSecretKeys(t, item, forbidden...)
		}
	}
}
