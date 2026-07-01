package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestSupportPhase1_SessionInspectionHelpers(t *testing.T) {
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

func TestSupportPhase1_AdminListQueriesUseCoreListQueryErrors(t *testing.T) {
	users, apiErr := parseUsersListScope("search=Admin+User&is_active=true&is_deployment_admin=false")
	if apiErr != nil {
		t.Fatalf("parse valid users query: %v", apiErr)
	}
	if users.Scope["search"] != "admin user" || users.Scope["is_active"] != "true" || users.Scope["is_deployment_admin"] != "false" {
		t.Fatalf("unexpected users scope: %#v", users.Scope)
	}

	_, apiErr = parseUsersListScope("is_active=yes")
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_list_query", "")
	if got := apiErr.Details["reason_code"]; got != "invalid_filter_value" {
		t.Fatalf("unexpected users invalid filter reason: %v", got)
	}
	_, apiErr = parseUsersListScope("search=a&search=b")
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_list_query", "")
	if got := apiErr.Details["reason_code"]; got != "duplicate_query_member" {
		t.Fatalf("unexpected users duplicate reason: %v", got)
	}

	audit, apiErr := parseAdministrativeAuditScope("target_kind=user&target_id=00000000-0000-0000-0000-000000000001&occurred_at_gte=2026-04-20T12:00:00Z&occurred_at_lt=2026-04-21T12:00:00Z")
	if apiErr != nil {
		t.Fatalf("parse valid administrative audit query: %v", apiErr)
	}
	if audit.Scope["target_kind"] != "user" || audit.Scope["target_id"] == "" {
		t.Fatalf("unexpected administrative audit scope: %#v", audit.Scope)
	}
	_, apiErr = parseAdministrativeAuditScope("search=user")
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_list_query", "")
	if got := apiErr.Details["reason_code"]; got != "unknown_query_member" {
		t.Fatalf("unexpected administrative audit unknown reason: %v", got)
	}
	_, apiErr = parseAdministrativeAuditScope("occurred_at_gte=2026-04-22T12:00:00Z&occurred_at_lt=2026-04-21T12:00:00Z")
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_list_query", "")
	if got := apiErr.Details["reason_code"]; got != "invalid_filter_range" {
		t.Fatalf("unexpected administrative audit range reason: %v", got)
	}
}

func TestSupportPhase1_CSRFHelpers(t *testing.T) {
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

func TestSupportPhase1_CredentialStateBuilders(t *testing.T) {
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
	requireNoSecretKeys(t, active, "password_hash", "bootstrap_token", "secret_base32", "otpauth_uri")

	apiErr := BootstrapRejectedError("not_allowed_for_route")
	requireAPIError(t, apiErr, http.StatusConflict, "credential_bootstrap_rejected", "")
	if got := apiErr.Details["reason_code"]; got != "not_allowed_for_route" {
		t.Fatalf("unexpected bootstrap rejection reason_code: got %v", got)
	}
}

func TestSupportPhase1_PasswordChangeDecode(t *testing.T) {
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

	conflict := ClientTxnConflictError("txn-password-2")
	requireAPIError(t, conflict, http.StatusConflict, "client_txn_conflict", "")
	if got := conflict.Details["client_txn_id"]; got != "txn-password-2" {
		t.Fatalf("unexpected client_txn_conflict details: got %v want txn-password-2", got)
	}
}

func TestSupportPhase1_TOTPBootstrapHelpers(t *testing.T) {
	if !AllowsBootstrapTokenRoute("/api/v1/auth/mfa/totp/begin") {
		t.Fatal("totp/begin must accept bootstrap_token")
	}
	if !AllowsBootstrapTokenRoute("/api/v1/auth/mfa/totp/complete") {
		t.Fatal("totp/complete must accept bootstrap_token")
	}
	if AllowsBootstrapTokenRoute("/api/v1/auth/credential-state") {
		t.Fatal("credential-state must reject bootstrap_token")
	}
	if AllowsBootstrapTokenRoute("/ws/v1/incidents/10000000-0000-0000-0000-000000000001") {
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

	_, apiErr := DecodeTOTPBeginRequest(strings.NewReader(`{
		"client_txn_id":"txn-totp-begin",
		"second_factor":{"kind":"totp","assertion":{"code":"123456"}}
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid totp begin request: %v", apiErr)
	}

	_, apiErr = DecodeTOTPCompleteRequest(strings.NewReader(`{
		"client_txn_id":"txn-totp-complete",
		"enrollment_id":"10000000-0000-0000-0000-000000000001",
		"code":"12 3456"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_auth_request", "code")
}

func TestSupportPhase1_UserCreateDefaultsAndSafeShape(t *testing.T) {
	defaults := ApplyUserCreateDefaults(nil, nil)
	if !defaults.MFARequired {
		t.Fatal("omitted mfa_required must default to true")
	}
	if defaults.IsDeploymentAdmin {
		t.Fatal("omitted is_deployment_admin must default to false")
	}

	userID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	createdAt := time.Date(2026, time.April, 17, 15, 0, 0, 0, time.UTC)
	resource := BuildSafeUserResource(authn.UserRecord{
		ID:                userID,
		Email:             "analyst@example.test",
		DisplayName:       "Analyst",
		IsActive:          true,
		MFARequired:       true,
		IsDeploymentAdmin: false,
		CreatedAt:         createdAt,
		UpdatedAt:         createdAt,
		UserVersion:       1,
	})

	if got := resource["email"]; got != "analyst@example.test" {
		t.Fatalf("unexpected email in safe user resource: got %v", got)
	}
	if got := resource["is_active"]; got != true {
		t.Fatalf("unexpected is_active in safe user resource: got %v", got)
	}
	authBindings, ok := resource["auth_bindings"].([]map[string]any)
	if !ok || len(authBindings) != 1 {
		t.Fatalf("expected one local auth binding summary, got %#v", resource["auth_bindings"])
	}
	if got := authBindings[0]["provider_type"]; got != "local" {
		t.Fatalf("unexpected local binding provider_type: got %v", got)
	}
	if _, ok := resource["password_hash"]; ok {
		t.Fatal("safe user resource must not expose password_hash")
	}
	requireNoSecretKeys(t, resource, "password_hash", "initial_password", "bootstrap_token", "secret_base32")
}

func TestSupportPhase1_AdminCredentialActionsRequireDeploymentAdmin(t *testing.T) {
	if apiErr := RequireDeploymentAdmin(authn.UserRecord{IsDeploymentAdmin: true}); apiErr != nil {
		t.Fatalf("deployment admin should pass admin-action guard: %v", apiErr)
	}
	apiErr := RequireDeploymentAdmin(authn.UserRecord{IsDeploymentAdmin: false})
	if apiErr == nil {
		t.Fatal("non-deployment-admin must fail admin-action guard")
	}
	if apiErr.Status != http.StatusForbidden || apiErr.Code != "authorization_denied" {
		t.Fatalf("unexpected admin-action guard error: %#v", apiErr)
	}
	if apiErr.Details["required_capability"] != "deployment_admin" {
		t.Fatalf("admin-action guard must identify required capability: %#v", apiErr.Details)
	}

	_, apiErr = DecodeAdminPasswordResetRequest(strings.NewReader(`{
		"base_user_version":1,
		"client_txn_id":"txn-reset",
		"new_password":"PasswordReset123!"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid admin password reset request: %v", apiErr)
	}
}
