package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"example.com/todo/cartulary/internal/platform/authn"
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

func TestPhase1_SessionInspectionContracts_U_1_04(t *testing.T) {
	query := url.Values{
		"limit": []string{"10"},
	}
	apiErr := ValidateSingletonReadQuery(query)
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_pagination_request", "")
	if got := apiErr.Details["reason_code"]; got != "pagination_not_supported" {
		t.Fatalf("unexpected pagination reason_code: got %v", got)
	}

	if ShouldSlideIdleExpiry(http.MethodGet, "/api/v1/auth/session") {
		t.Fatal("session inspection must not extend idle expiry")
	}
	if !ShouldSlideIdleExpiry(http.MethodPost, "/api/v1/auth/logout") {
		t.Fatal("other authenticated API routes should qualify for idle sliding")
	}
}

func TestPhase1_CSRFFailClosed_U_1_09(t *testing.T) {
	if apiErr := ValidateCSRF(http.MethodPost, AuthSourceCookie, "csrf-cookie", ""); apiErr == nil {
		t.Fatal("expected missing csrf header to fail for cookie-authenticated state change")
	}
	if apiErr := ValidateCSRF(http.MethodPost, AuthSourceCookie, "csrf-cookie", "wrong-value"); apiErr == nil {
		t.Fatal("expected mismatched csrf header to fail for cookie-authenticated state change")
	}
	if apiErr := ValidateCSRF(http.MethodPost, AuthSourceCookie, "csrf-cookie", "csrf-cookie"); apiErr != nil {
		t.Fatalf("expected matching csrf proof to pass, got %v", apiErr)
	}
	if apiErr := ValidateCSRF(http.MethodPost, AuthSourceBearer, "", ""); apiErr != nil {
		t.Fatalf("bearer-authenticated state changes must not require csrf, got %v", apiErr)
	}
	if apiErr := ValidateCSRF(http.MethodGet, AuthSourceCookie, "", ""); apiErr != nil {
		t.Fatalf("safe methods must not require csrf, got %v", apiErr)
	}
}

func TestPhase1_CredentialStateInspection_U_1_10(t *testing.T) {
	userID := uuid.MustParse("20000000-0000-0000-0000-000000000001")
	changedAt := time.Date(2026, time.April, 17, 12, 30, 0, 0, time.UTC)
	enrolledAt := time.Date(2026, time.April, 17, 13, 0, 0, 0, time.UTC)
	pendingUntil := time.Date(2026, time.April, 17, 13, 15, 0, 0, time.UTC)

	notEnrolled := BuildCredentialStateResource(authn.UserRecord{
		ID:                userID,
		MFARequired:       true,
		PasswordChangedAt: changedAt,
	}, nil)
	if got := notEnrolled["totp"].(map[string]any)["state"]; got != "not_enrolled" {
		t.Fatalf("unexpected not-enrolled totp.state: got %v", got)
	}

	pending := BuildCredentialStateResource(authn.UserRecord{
		ID:                userID,
		MFARequired:       true,
		PasswordChangedAt: changedAt,
	}, &pendingUntil)
	if got := pending["totp"].(map[string]any)["state"]; got != "pending" {
		t.Fatalf("unexpected pending totp.state: got %v", got)
	}

	active := BuildCredentialStateResource(authn.UserRecord{
		ID:                userID,
		MFARequired:       true,
		PasswordChangedAt: changedAt,
		TOTPEnrolledAt:    &enrolledAt,
	}, nil)
	totp := active["totp"].(map[string]any)
	if got := totp["state"]; got != "active" {
		t.Fatalf("unexpected active totp.state: got %v", got)
	}
	if got := active["auth_kind"]; got != "local" {
		t.Fatalf("unexpected auth_kind: got %v", got)
	}
	if got := active["recovery_model"]; got != "admin_assisted" {
		t.Fatalf("unexpected recovery_model: got %v", got)
	}

	apiErr := BootstrapRejectedError("not_allowed_for_route")
	requireAPIError(t, apiErr, http.StatusConflict, "credential_bootstrap_rejected", "")
	if got := apiErr.Details["reason_code"]; got != "not_allowed_for_route" {
		t.Fatalf("unexpected bootstrap rejection reason_code: got %v", got)
	}
}

func TestPhase1_PasswordChangeRequest_U_1_11(t *testing.T) {
	request, apiErr := DecodePasswordChangeRequest(strings.NewReader(`{
		"client_txn_id":"txn-password-1",
		"current_password":"  Exact Current  ",
		"new_password":"Replacement passphrase 1",
		"second_factor":{"kind":"totp","assertion":{"code":"123456"}}
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid password-change request: %v", apiErr)
	}
	if request.CurrentPassword != "  Exact Current  " {
		t.Fatalf("current_password must remain exact after decode: got %q", request.CurrentPassword)
	}
	if request.NewPassword != "Replacement passphrase 1" {
		t.Fatalf("new_password must remain exact after decode: got %q", request.NewPassword)
	}

	_, apiErr = DecodePasswordChangeRequest(strings.NewReader(`{"current_password":"x","new_password":"y"}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_auth_request", "client_txn_id")

	_, apiErr = DecodePasswordChangeRequest(strings.NewReader(`{
		"client_txn_id":"txn-password-2",
		"current_password":"current password",
		"new_password":"Replacement passphrase 1",
		"second_factor":{"kind":"totp","assertion":{"code":"12 3456"}}
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_auth_request", "second_factor.assertion.code")
}

func TestPhase1_TOTPBootstrapAndSetupRules_U_1_12(t *testing.T) {
	if !AllowsBootstrapTokenRoute("/api/v1/auth/mfa/totp/begin") {
		t.Fatal("totp/begin must accept bootstrap_token")
	}
	if !AllowsBootstrapTokenRoute("/api/v1/auth/mfa/totp/complete") {
		t.Fatal("totp/complete must accept bootstrap_token")
	}
	if AllowsBootstrapTokenRoute("/api/v1/auth/credential-state") {
		t.Fatal("credential-state must reject bootstrap_token")
	}
	if AllowsBootstrapTokenRoute("/ws/v1/test/session-lifecycle") {
		t.Fatal("/ws/v1/* must reject bootstrap_token")
	}

	setup := BuildTOTPSetup("JBSWY3DPEHPK3PXP", "analyst@example.test")
	if got := setup["algorithm"]; got != "SHA1" {
		t.Fatalf("unexpected TOTP algorithm: got %v", got)
	}
	if got := setup["digits"]; got != 6 {
		t.Fatalf("unexpected TOTP digits: got %v", got)
	}
	if got := setup["period_seconds"]; got != 30 {
		t.Fatalf("unexpected TOTP period_seconds: got %v", got)
	}
	if got := setup["secret_base32"]; got != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("unexpected TOTP secret_base32: got %v", got)
	}
	if got := ShouldRevokeSessionsOnTOTPComplete(false); got {
		t.Fatal("first enrollment must not revoke sessions on complete")
	}
	if got := ShouldRevokeSessionsOnTOTPComplete(true); !got {
		t.Fatal("replacement enrollment must revoke sessions on complete")
	}
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
