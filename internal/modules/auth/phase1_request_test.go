package auth

import (
	"net/http"
	"strings"
	"testing"
)

func TestPhase1_LoginRequestShape_U_1_01(t *testing.T) {
	for _, tc := range []struct {
		name      string
		body      string
		wantField string
	}{
		{
			name: "rejects non-object request bodies",
			body: `["analyst@example.test","exact password"]`,
		},
		{
			name: "rejects forbidden top-level members",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"client_txn_id":"forbidden"
			}`,
			wantField: "client_txn_id",
		},
		{
			name: "rejects wrong top-level username type",
			body: `{
				"username":42,
				"password":"exact password"
			}`,
			wantField: "username",
		},
		{
			name: "rejects wrong top-level password type",
			body: `{
				"username":"analyst@example.test",
				"password":null
			}`,
			wantField: "password",
		},
		{
			name: "rejects null second_factor container",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":null
			}`,
			wantField: "second_factor.kind",
		},
		{
			name: "rejects non-object second_factor container",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":"totp"
			}`,
			wantField: "second_factor",
		},
		{
			name: "rejects missing second_factor kind",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"assertion":{"code":"123456"}}
			}`,
			wantField: "second_factor.kind",
		},
		{
			name: "rejects non-string second_factor kind",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":7,"assertion":{"code":"123456"}}
			}`,
			wantField: "second_factor.kind",
		},
		{
			name: "rejects unsupported second-factor kind",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"webauthn","assertion":{"code":"123456"}}
			}`,
			wantField: "second_factor.kind",
		},
		{
			name: "rejects unknown second_factor members",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":{"code":"123456"},"proof":"extra"}
			}`,
			wantField: "second_factor.proof",
		},
		{
			name: "rejects missing second_factor assertion",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp"}
			}`,
			wantField: "second_factor.assertion",
		},
		{
			name: "rejects non-object second_factor assertion",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":null}
			}`,
			wantField: "second_factor.assertion.code",
		},
		{
			name: "rejects unknown second_factor assertion members",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":{"code":"123456","backup":"123456"}}
			}`,
			wantField: "second_factor.assertion.backup",
		},
		{
			name: "rejects missing second_factor assertion code",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":{}}
			}`,
			wantField: "second_factor.assertion.code",
		},
		{
			name: "rejects non-string second_factor assertion code",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":{"code":123456}}
			}`,
			wantField: "second_factor.assertion.code",
		},
		{
			name: "rejects malformed totp assertion shape",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":{"code":"12 3456"}}
			}`,
			wantField: "second_factor.assertion.code",
		},
		{
			name: "rejects non-digit totp assertion code",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":{"code":"ABC123"}}
			}`,
			wantField: "second_factor.assertion.code",
		},
		{
			name: "rejects wrong-length totp assertion code",
			body: `{
				"username":"analyst@example.test",
				"password":"exact password",
				"second_factor":{"kind":"totp","assertion":{"code":"12345"}}
			}`,
			wantField: "second_factor.assertion.code",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := DecodeLoginRequest(strings.NewReader(tc.body))
			requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_auth_request", tc.wantField)
		})
	}
}

func requireAPIError(t testing.TB, apiErr *APIError, wantStatus int, wantCode string, wantField string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
		return
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
