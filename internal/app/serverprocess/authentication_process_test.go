package serverprocess

import (
	"net/http"
	"testing"

	"github.com/coder/websocket"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// These are process-level smoke tests for the standalone server binary.
// They are intentionally not the authoritative Authentication browser E2E ledger.
func TestLoginSessionAndLogout_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-01")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.SessionCookie
	adminCSRF := adminLogin.CSRFCookie

	sessionResp := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if sessionBody["provider_type"] != "local" || sessionBody["is_deployment_admin"] != true || sessionBody["mfa_state"] != "satisfied" {
		t.Fatalf("unexpected session resource: %#v", sessionBody)
	}
	if memberships, ok := sessionBody["memberships"].([]any); !ok || len(memberships) != 0 {
		t.Fatalf("expected memberships to be present and empty, got %#v", sessionBody["memberships"])
	}

	logoutResp := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	logoutBody := httptestx.RequireSuccessEnvelope(t, logoutResp, http.StatusOK)["data"].(map[string]any)
	if logoutBody["logged_out"] != true || logoutBody["sessions_revoked"] != false {
		t.Fatalf("unexpected logout payload: %#v", logoutBody)
	}

	postLogout := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	httptestx.RequireErrorEnvelope(t, postLogout, http.StatusUnauthorized, "session_required")
}

func TestCSRFFailClosed_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-02")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.SessionCookie
	adminCSRF := adminLogin.CSRFCookie

	missingHeader := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
	)
	httptestx.RequireErrorEnvelope(t, missingHeader, http.StatusForbidden, "csrf_verification_failed")

	stillActive := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	httptestx.RequireSuccessEnvelope(t, stillActive, http.StatusOK)

	wrongHeader := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, "wrong-token"),
	)
	httptestx.RequireErrorEnvelope(t, wrongHeader, http.StatusForbidden, "csrf_verification_failed")

	validHeader := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, validHeader, http.StatusOK)
}

func TestConcurrencyCapRevokesSocket_Process(t *testing.T) {
	t.Parallel()

	server, db := startServerProcessWithDB(t, "authentication-e-1-03")
	t.Cleanup(func() {
		flowtest.ResetClockOffset(t, server.BaseURL)
	})

	initialLogin, adminSecret := provisionBootstrapAdmin(t, server)
	adminUserID := flowtest.QueryUserIDByEmail(t, db, authenticationBootstrapAdminEmail)
	firstSessionID := flowtest.QuerySessionRow(t, db, adminUserID).SessionID
	socketIncidentID := createSocketIncident(t, server, initialLogin, "e-1-03-socket")

	sessions := make([]flowtest.LoginResult, 0, 6)
	sessions = append(sessions, initialLogin)
	for i := 0; i < 4; i++ {
		flowtest.SetClockOffset(t, server.BaseURL, int64(i+1))
		sessions = append(sessions, loginLocalUserWithSecondFactor(t, server, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, flowtest.GenerateTOTPCode(t, adminSecret)))
	}

	socket := flowtest.ConnectSessionSocket(t, server.BaseURL, socketIncidentID, sessions[0].SessionCookie.Value)
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	flowtest.SetClockOffset(t, server.BaseURL, 5)
	sessions = append(sessions, loginLocalUserWithSecondFactor(t, server, authenticationBootstrapAdminEmail, authenticationBootstrapAdminPassword, flowtest.GenerateTOTPCode(t, adminSecret)))
	if err := flowtest.AwaitSessionRevoked(socket, authn.ConcurrencyLimitReasonCode); err != nil {
		firstSession := flowtest.QuerySessionByID(t, db, firstSessionID)
		activeCount := flowtest.QueryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text = $1 AND revoked_at IS NULL`, adminUserID)
		t.Fatalf(
			"await session_revoked for first session %s: %v (active_sessions=%d revoked_at_valid=%t revoke_reason_code=%q sessions=%s)",
			firstSessionID,
			err,
			activeCount,
			firstSession.RevokedAt.Valid,
			firstSession.RevokeReasonCode.String,
			flowtest.FormatUserSessions(t, db, adminUserID),
		)
	}

	revokedSession := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(sessions[0].SessionCookie))
	httptestx.RequireErrorEnvelope(t, revokedSession, http.StatusUnauthorized, "session_required")

	activeSession := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(sessions[5].SessionCookie))
	httptestx.RequireSuccessEnvelope(t, activeSession, http.StatusOK)
}

func TestFirstEnrollmentFlow_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-04")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.SessionCookie
	adminCSRF := adminLogin.CSRFCookie
	user := createUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-1-04-create",
		"auth_kind":        "local",
		"email":            "authentication-e-1-04@example.test",
		"display_name":     "Authentication E104",
		"initial_password": "AuthenticationE104Pass!",
	})

	bootstrapToken := requireBootstrapLogin(t, server, "authentication-e-1-04@example.test", "AuthenticationE104Pass!")
	begin := flowtest.BeginTOTPEnrollment(t, server.BaseURL, bootstrapToken, map[string]any{
		"client_txn_id": "txn-e-1-04-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	flowtest.CompleteInitialEnrollment(t, server.BaseURL, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-e-1-04-complete")

	mfaRequired := doJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "authentication-e-1-04@example.test",
		"password": "AuthenticationE104Pass!",
	})
	httptestx.RequireErrorEnvelope(t, mfaRequired, http.StatusUnauthorized, "mfa_required")

	userLogin := loginLocalUserWithSecondFactor(t, server, "authentication-e-1-04@example.test", "AuthenticationE104Pass!", flowtest.GenerateTOTPCode(t, secretBase32))
	stateResp := doJSON(t, server, http.MethodGet, "/api/v1/auth/credential-state", nil, withCookies(userLogin.SessionCookie))
	stateBody := httptestx.RequireSuccessEnvelope(t, stateResp, http.StatusOK)["data"].(map[string]any)
	if stateBody["user_id"] != user["user_id"] {
		t.Fatalf("unexpected credential-state user_id: got %v want %v", stateBody["user_id"], user["user_id"])
	}
	totpState := stateBody["totp"].(map[string]any)
	if totpState["state"] != "active" {
		t.Fatalf("unexpected totp state after enrollment: %#v", totpState)
	}
}

func TestPasswordChangeFlow_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-05")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.SessionCookie
	adminCSRF := adminLogin.CSRFCookie
	_, secretBase32 := provisionTOTPUser(t, server, adminSession, adminCSRF, "authentication-e-1-05@example.test", "Authentication E105", "AuthenticationE105Pass!")

	userLogin := loginLocalUserWithSecondFactor(t, server, "authentication-e-1-05@example.test", "AuthenticationE105Pass!", flowtest.GenerateTOTPCode(t, secretBase32))
	socket := connectSessionSocket(t, server, userLogin, "e-1-05-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	changeResp := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/password/change",
		map[string]any{
			"client_txn_id":    "txn-e-1-05-password-change",
			"current_password": "AuthenticationE105Pass!",
			"new_password":     "AuthenticationE105Changed!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": flowtest.GenerateTOTPCode(t, secretBase32),
				},
			},
		},
		withCookies(userLogin.SessionCookie, userLogin.CSRFCookie),
		withHeader(authn.CSRFHeaderName, userLogin.CSRFCookie.Value),
	)
	changeBody := httptestx.RequireSuccessEnvelope(t, changeResp, http.StatusOK)["data"].(map[string]any)
	if changeBody["sessions_revoked"] != true {
		t.Fatalf("unexpected password change response: %#v", changeBody)
	}

	flowtest.ExpectSessionRevoked(t, socket, "session_revoked")

	oldPassword := doJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "authentication-e-1-05@example.test",
		"password": "AuthenticationE105Pass!",
	})
	httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

	newPasswordNoFactor := doJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "authentication-e-1-05@example.test",
		"password": "AuthenticationE105Changed!",
	})
	httptestx.RequireErrorEnvelope(t, newPasswordNoFactor, http.StatusUnauthorized, "mfa_required")

	_ = loginLocalUserWithSecondFactor(t, server, "authentication-e-1-05@example.test", "AuthenticationE105Changed!", flowtest.GenerateTOTPCode(t, secretBase32))
}

func TestUserAdminAndRevokeAll_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-06")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.SessionCookie
	adminCSRF := adminLogin.CSRFCookie
	created := createUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-1-06-create",
		"auth_kind":        "local",
		"email":            "authentication-e-1-06@example.test",
		"display_name":     "Authentication E106",
		"initial_password": "AuthenticationE106Pass!",
		"mfa_required":     false,
	})
	createdUserID := created["user_id"].(string)

	listResp := doJSON(t, server, http.MethodGet, "/api/v1/users", nil, withCookies(adminSession))
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	users := listBody["users"].([]any)
	if len(users) < 2 {
		t.Fatalf("expected users list to include bootstrap admin and created user, got %#v", users)
	}

	getResp := doJSON(t, server, http.MethodGet, "/api/v1/users/"+createdUserID, nil, withCookies(adminSession))
	getBody := httptestx.RequireSuccessEnvelope(t, getResp, http.StatusOK)["data"].(map[string]any)
	if getBody["user_id"] != createdUserID {
		t.Fatalf("unexpected user lookup payload: %#v", getBody)
	}

	patchResp := doJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/users/"+createdUserID,
		map[string]any{
			"base_user_version": 1,
			"display_name":      "Authentication E106 Patched",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	if patchBody["display_name"] != "Authentication E106 Patched" || patchBody["user_version"] != float64(2) {
		t.Fatalf("unexpected patched user payload: %#v", patchBody)
	}

	userLogin := loginLocalUser(t, server, "authentication-e-1-06@example.test", "AuthenticationE106Pass!")
	nonAdminAction := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/users/"+createdUserID+"/sessions/revoke-all",
		map[string]any{
			"client_txn_id": "txn-e-1-06-non-admin",
			"reason":        "self revoke attempt",
		},
		withCookies(userLogin.SessionCookie, userLogin.CSRFCookie),
		withHeader(authn.CSRFHeaderName, userLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, nonAdminAction, http.StatusForbidden, "authorization_denied")

	socket := connectSessionSocket(t, server, userLogin, "e-1-06-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	revokeResp := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/users/"+createdUserID+"/sessions/revoke-all",
		map[string]any{
			"client_txn_id": "txn-e-1-06-revoke-all",
			"reason":        "explicit revoke all",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	revokeBody := httptestx.RequireSuccessEnvelope(t, revokeResp, http.StatusOK)["data"].(map[string]any)
	if revokeBody["sessions_revoked"] != true {
		t.Fatalf("unexpected revoke-all payload: %#v", revokeBody)
	}

	flowtest.ExpectSessionRevoked(t, socket, "session_revoked")
	_ = loginLocalUser(t, server, "authentication-e-1-06@example.test", "AuthenticationE106Pass!")
}

func TestAdminPasswordReset_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-07")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.SessionCookie
	adminCSRF := adminLogin.CSRFCookie
	user, secretBase32 := provisionTOTPUser(t, server, adminSession, adminCSRF, "authentication-e-1-07@example.test", "Authentication E107", "AuthenticationE107Pass!")
	targetUserID := user["user_id"].(string)

	targetLogin := loginLocalUserWithSecondFactor(t, server, "authentication-e-1-07@example.test", "AuthenticationE107Pass!", flowtest.GenerateTOTPCode(t, secretBase32))
	socket := connectSessionSocket(t, server, targetLogin, "e-1-07-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	resetResp := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/users/"+targetUserID+"/password/reset",
		map[string]any{
			"base_user_version": 2,
			"client_txn_id":     "txn-e-1-07-password-reset",
			"new_password":      "AuthenticationE107Reset!",
			"reason":            "e2e password reset",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	resetBody := httptestx.RequireSuccessEnvelope(t, resetResp, http.StatusOK)["data"].(map[string]any)
	if resetBody["user_version"] != float64(3) {
		t.Fatalf("unexpected password-reset payload: %#v", resetBody)
	}

	flowtest.ExpectSessionRevoked(t, socket, "session_revoked")

	oldPassword := doJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "authentication-e-1-07@example.test",
		"password": "AuthenticationE107Pass!",
	})
	httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

	newPasswordNoFactor := doJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "authentication-e-1-07@example.test",
		"password": "AuthenticationE107Reset!",
	})
	httptestx.RequireErrorEnvelope(t, newPasswordNoFactor, http.StatusUnauthorized, "mfa_required")

	_ = loginLocalUserWithSecondFactor(t, server, "authentication-e-1-07@example.test", "AuthenticationE107Reset!", flowtest.GenerateTOTPCode(t, secretBase32))
}

func TestAdminTOTPResetAndBootstrapBoundaries_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "authentication-e-1-08")

	adminLogin, _ := provisionBootstrapAdmin(t, server)
	adminSession := adminLogin.SessionCookie
	adminCSRF := adminLogin.CSRFCookie
	user, secretBase32 := provisionTOTPUser(t, server, adminSession, adminCSRF, "authentication-e-1-08@example.test", "Authentication E108", "AuthenticationE108Pass!")
	targetUserID := user["user_id"].(string)

	targetLogin := loginLocalUserWithSecondFactor(t, server, "authentication-e-1-08@example.test", "AuthenticationE108Pass!", flowtest.GenerateTOTPCode(t, secretBase32))
	socket := connectSessionSocket(t, server, targetLogin, "e-1-08-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	resetResp := doJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/users/"+targetUserID+"/mfa/totp/reset",
		map[string]any{
			"base_user_version": 2,
			"client_txn_id":     "txn-e-1-08-totp-reset",
			"reason":            "e2e totp reset",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	resetBody := httptestx.RequireSuccessEnvelope(t, resetResp, http.StatusOK)["data"].(map[string]any)
	if resetBody["user_version"] != float64(3) {
		t.Fatalf("unexpected totp-reset payload: %#v", resetBody)
	}

	flowtest.ExpectSessionRevoked(t, socket, "session_revoked")

	bootstrapLogin := doJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "authentication-e-1-08@example.test",
		"password": "AuthenticationE108Pass!",
	})
	bootstrapBody := httptestx.RequireErrorEnvelope(t, bootstrapLogin, http.StatusUnauthorized, "mfa_setup_required")
	bootstrapDetails := bootstrapBody["error"].(map[string]any)["details"].(map[string]any)
	bootstrapToken := bootstrapDetails["bootstrap_token"].(string)

	rejectedRead := doJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withHeader("Authorization", "Bearer "+bootstrapToken))
	rejectedBody := httptestx.RequireErrorEnvelope(t, rejectedRead, http.StatusConflict, "credential_bootstrap_rejected")
	rejectedDetails := rejectedBody["error"].(map[string]any)["details"].(map[string]any)
	if rejectedDetails["reason_code"] != "not_allowed_for_route" {
		t.Fatalf("unexpected bootstrap route rejection: %#v", rejectedDetails)
	}

	incidentID := createSocketIncident(t, server, adminLogin, "e-1-08-bootstrap")
	flowtest.RequireBootstrapWebsocketRejected(t, server.BaseURL, incidentID, bootstrapToken)

	begin := flowtest.BeginTOTPEnrollment(t, server.BaseURL, bootstrapToken, map[string]any{
		"client_txn_id": "txn-e-1-08-begin",
	})
	newSecretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	flowtest.CompleteInitialEnrollment(t, server.BaseURL, bootstrapToken, begin["enrollment_id"].(string), newSecretBase32, "txn-e-1-08-complete")

	_ = loginLocalUserWithSecondFactor(t, server, "authentication-e-1-08@example.test", "AuthenticationE108Pass!", flowtest.GenerateTOTPCode(t, newSecretBase32))
}
