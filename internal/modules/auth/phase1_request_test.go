package auth

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
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

	t.Run("full login route rejects malformed auth requests before session side effects", func(t *testing.T) {
		now := time.Date(2026, time.April, 17, 11, 45, 0, 0, time.UTC)
		keys := loadUnitMasterKeys(t)

		storeCalls := 0
		store := &authStoreStub{
			getUserByNormalizedEmailFunc: func(context.Context, string) (authn.UserRecord, error) {
				storeCalls++
				t.Fatal("malformed login request must not look up users")
				return authn.UserRecord{}, nil
			},
			createSessionWithConcurrencyFunc: func(context.Context, authn.UserRecord, []byte, authn.SessionTiming, string) (authn.SessionRecord, *authn.SessionRecord, error) {
				storeCalls++
				t.Fatal("malformed login request must not create sessions")
				return authn.SessionRecord{}, nil, nil
			},
			issueBootstrapTokenFunc: func(context.Context, uuid.UUID, []byte, time.Time) (authn.BootstrapTokenRecord, error) {
				storeCalls++
				t.Fatal("malformed login request must not issue bootstrap tokens")
				return authn.BootstrapTokenRecord{}, nil
			},
		}
		service := newUnitService(t, store, &hubStub{}, keys, now)

		recorder := httptest.NewRecorder()
		request := newJSONRequest(t, http.MethodPost, "/api/v1/auth/login", `{
			"username":"analyst@example.test",
			"password":"exact password",
			"client_txn_id":"forbidden"
		}`)
		service.handleLogin(recorder, request)

		requireErrorEnvelope(t, recorder, http.StatusBadRequest, "invalid_auth_request", "client_txn_id")
		requireNoCookieByName(t, recorder.Result().Cookies(), authn.SessionCookieName)
		requireNoCookieByName(t, recorder.Result().Cookies(), authn.CSRFCookieName)
		if storeCalls != 0 {
			t.Fatalf("malformed login request must not call auth store, got %d calls", storeCalls)
		}
	})
}

func TestPhase1_StrictAuthRequestDecoding_U_1_01(t *testing.T) {
	for _, tc := range []struct {
		name       string
		body       string
		decode     func(io.Reader) *httpapi.APIError
		wantCode   string
		wantReason string
	}{
		{
			name: "login rejects duplicate top-level members",
			body: `{"username":"analyst@example.test","username":"other@example.test","password":"exact password"}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeLoginRequest(reader)
				return apiErr
			},
			wantCode: "invalid_auth_request",
		},
		{
			name: "login rejects trailing json",
			body: `{"username":"analyst@example.test","password":"exact password"}{}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeLoginRequest(reader)
				return apiErr
			},
			wantCode: "invalid_auth_request",
		},
		{
			name: "credential request rejects non-object bodies",
			body: `"not-object"`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodePasswordChangeRequest(reader)
				return apiErr
			},
			wantCode: "invalid_auth_request",
		},
		{
			name: "credential request rejects duplicate top-level members",
			body: `{"client_txn_id":"txn-password","client_txn_id":"txn-password-2","current_password":"current password","new_password":"Replacement passphrase 1"}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodePasswordChangeRequest(reader)
				return apiErr
			},
			wantCode: "invalid_auth_request",
		},
		{
			name: "credential request rejects trailing json",
			body: `{"client_txn_id":"txn-password","current_password":"current password","new_password":"Replacement passphrase 1"}{}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodePasswordChangeRequest(reader)
				return apiErr
			},
			wantCode: "invalid_auth_request",
		},
		{
			name: "account request rejects non-object bodies",
			body: `"not-object"`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeAccountProfilePatchRequest(reader)
				return apiErr
			},
			wantCode:   "invalid_mutation_payload",
			wantReason: "request_not_object",
		},
		{
			name: "account request rejects duplicate top-level members",
			body: `{"base_user_version":1,"client_txn_id":"txn-profile","display_name":"Analyst","display_name":"Other Analyst"}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeAccountProfilePatchRequest(reader)
				return apiErr
			},
			wantCode:   "invalid_mutation_payload",
			wantReason: "request_not_object",
		},
		{
			name: "account request rejects trailing json",
			body: `{"base_user_version":1,"client_txn_id":"txn-profile","display_name":"Analyst"}{}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeAccountProfilePatchRequest(reader)
				return apiErr
			},
			wantCode:   "invalid_mutation_payload",
			wantReason: "request_not_object",
		},
		{
			name: "deployment-user admin request rejects non-object bodies",
			body: `"not-object"`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeUserCreateRequest(reader)
				return apiErr
			},
			wantCode:   "invalid_mutation_payload",
			wantReason: "request_not_object",
		},
		{
			name: "deployment-user admin request rejects duplicate top-level members",
			body: `{"client_txn_id":"txn-user-create","auth_kind":"local","email":"new.user@example.test","email":"other.user@example.test","display_name":"New User","initial_password":"Initial passphrase 1"}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeUserCreateRequest(reader)
				return apiErr
			},
			wantCode:   "invalid_mutation_payload",
			wantReason: "request_not_object",
		},
		{
			name: "deployment-user admin request rejects trailing json",
			body: `{"client_txn_id":"txn-user-create","auth_kind":"local","email":"new.user@example.test","display_name":"New User","initial_password":"Initial passphrase 1"}{}`,
			decode: func(reader io.Reader) *httpapi.APIError {
				_, apiErr := DecodeUserCreateRequest(reader)
				return apiErr
			},
			wantCode:   "invalid_mutation_payload",
			wantReason: "request_not_object",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			apiErr := tc.decode(strings.NewReader(tc.body))
			requireAPIErrorReason(t, apiErr, http.StatusBadRequest, tc.wantCode, tc.wantReason)
		})
	}
}

func requireAPIError(t testing.TB, apiErr *httpapi.APIError, wantStatus int, wantCode string, wantField string) {
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

func requireAPIErrorReason(t testing.TB, apiErr *httpapi.APIError, wantStatus int, wantCode string, wantReason string) {
	t.Helper()
	requireAPIError(t, apiErr, wantStatus, wantCode, "")
	if wantReason == "" {
		return
	}
	if got := apiErr.Details["reason_code"]; got != wantReason {
		t.Fatalf("unexpected reason_code detail: got %v want %s", got, wantReason)
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

func requireNoCookieByName(t testing.TB, cookies []*http.Cookie, name string) {
	t.Helper()
	for _, cookie := range cookies {
		if cookie.Name == name {
			t.Fatalf("expected no %s cookie, got %#v", name, cookie)
		}
	}
}
