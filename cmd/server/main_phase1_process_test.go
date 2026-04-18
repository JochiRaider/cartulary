package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

const (
	phase1BootstrapAdminEmail    = "bootstrap-admin@example.test"
	phase1BootstrapAdminPassword = "BootstrapPass1!"
)

// These are process-level smoke tests for the standalone server binary.
// They are intentionally not the authoritative Phase 1 browser E2E ledger.
func TestPhase1_LoginSessionAndLogout_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-01")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie

	sessionResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if sessionBody["provider_type"] != "local" || sessionBody["is_deployment_admin"] != true || sessionBody["mfa_state"] != "satisfied" {
		t.Fatalf("unexpected session resource: %#v", sessionBody)
	}
	if memberships, ok := sessionBody["memberships"].([]any); !ok || len(memberships) != 0 {
		t.Fatalf("expected memberships to be present and empty, got %#v", sessionBody["memberships"])
	}

	logoutResp := phase1DoJSON(
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

	postLogout := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	httptestx.RequireErrorEnvelope(t, postLogout, http.StatusUnauthorized, "session_required")
}

func TestPhase1_CSRFFailClosed_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-02")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie

	missingHeader := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
	)
	httptestx.RequireErrorEnvelope(t, missingHeader, http.StatusForbidden, "csrf_verification_failed")

	stillActive := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	httptestx.RequireSuccessEnvelope(t, stillActive, http.StatusOK)

	wrongHeader := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, "wrong-token"),
	)
	httptestx.RequireErrorEnvelope(t, wrongHeader, http.StatusForbidden, "csrf_verification_failed")

	validHeader := phase1DoJSON(
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

func TestPhase1_ConcurrencyCapRevokesSocket_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-03")

	initialLogin, adminSecret := phase1ProvisionBootstrapAdmin(t, server)
	sessions := make([]loginResult, 0, 6)
	sessions = append(sessions, initialLogin)
	for i := 0; i < 4; i++ {
		sessions = append(sessions, phase1LoginLocalUserWithSecondFactor(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword, phase1GenerateTOTPCode(t, adminSecret)))
	}

	socket := phase1ConnectSessionSocket(t, server, sessions[0].sessionCookie.Value)
	defer socket.CloseNow()

	sessions = append(sessions, phase1LoginLocalUserWithSecondFactor(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword, phase1GenerateTOTPCode(t, adminSecret)))
	phase1ExpectSessionRevoked(t, socket, authn.ConcurrencyLimitReasonCode)

	revokedSession := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(sessions[0].sessionCookie))
	httptestx.RequireErrorEnvelope(t, revokedSession, http.StatusUnauthorized, "session_required")

	activeSession := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(sessions[5].sessionCookie))
	httptestx.RequireSuccessEnvelope(t, activeSession, http.StatusOK)
}

func TestPhase1_FirstEnrollmentFlow_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-04")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	user := phase1CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-1-04-create",
		"auth_kind":        "local",
		"email":            "phase1-e-1-04@example.test",
		"display_name":     "Phase1 E104",
		"initial_password": "Phase1E104Pass!",
	})

	bootstrapToken := phase1RequireBootstrapLogin(t, server, "phase1-e-1-04@example.test", "Phase1E104Pass!")
	begin := phase1BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-e-1-04-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	phase1CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-e-1-04-complete")

	mfaRequired := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-04@example.test",
		"password": "Phase1E104Pass!",
	})
	httptestx.RequireErrorEnvelope(t, mfaRequired, http.StatusUnauthorized, "mfa_required")

	userLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-04@example.test", "Phase1E104Pass!", phase1GenerateTOTPCode(t, secretBase32))
	stateResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/credential-state", nil, withCookies(userLogin.sessionCookie))
	stateBody := httptestx.RequireSuccessEnvelope(t, stateResp, http.StatusOK)["data"].(map[string]any)
	if stateBody["user_id"] != user["user_id"] {
		t.Fatalf("unexpected credential-state user_id: got %v want %v", stateBody["user_id"], user["user_id"])
	}
	totpState := stateBody["totp"].(map[string]any)
	if totpState["state"] != "active" {
		t.Fatalf("unexpected totp state after enrollment: %#v", totpState)
	}
}

func TestPhase1_PasswordChangeFlow_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-05")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	_, secretBase32 := phase1ProvisionTOTPUser(t, server, adminSession, adminCSRF, "phase1-e-1-05@example.test", "Phase1 E105", "Phase1E105Pass!")

	userLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-05@example.test", "Phase1E105Pass!", phase1GenerateTOTPCode(t, secretBase32))
	socket := phase1ConnectSessionSocket(t, server, userLogin.sessionCookie.Value)
	defer socket.CloseNow()

	changeResp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/password/change",
		map[string]any{
			"client_txn_id":    "txn-e-1-05-password-change",
			"current_password": "Phase1E105Pass!",
			"new_password":     "Phase1E105Changed!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": phase1GenerateTOTPCode(t, secretBase32),
				},
			},
		},
		withCookies(userLogin.sessionCookie, userLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, userLogin.csrfCookie.Value),
	)
	changeBody := httptestx.RequireSuccessEnvelope(t, changeResp, http.StatusOK)["data"].(map[string]any)
	if changeBody["sessions_revoked"] != true {
		t.Fatalf("unexpected password change response: %#v", changeBody)
	}

	phase1ExpectSessionRevoked(t, socket, "session_revoked")

	oldPassword := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-05@example.test",
		"password": "Phase1E105Pass!",
	})
	httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

	newPasswordNoFactor := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-05@example.test",
		"password": "Phase1E105Changed!",
	})
	httptestx.RequireErrorEnvelope(t, newPasswordNoFactor, http.StatusUnauthorized, "mfa_required")

	_ = phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-05@example.test", "Phase1E105Changed!", phase1GenerateTOTPCode(t, secretBase32))
}

func TestPhase1_UserAdminAndRevokeAll_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-06")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	created := phase1CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-1-06-create",
		"auth_kind":        "local",
		"email":            "phase1-e-1-06@example.test",
		"display_name":     "Phase1 E106",
		"initial_password": "Phase1E106Pass!",
		"mfa_required":     false,
	})
	createdUserID := created["user_id"].(string)

	listResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/users", nil, withCookies(adminSession))
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	users := listBody["users"].([]any)
	if len(users) < 2 {
		t.Fatalf("expected users list to include bootstrap admin and created user, got %#v", users)
	}

	getResp := phase1DoJSON(t, server, http.MethodGet, "/api/v1/users/"+createdUserID, nil, withCookies(adminSession))
	getBody := httptestx.RequireSuccessEnvelope(t, getResp, http.StatusOK)["data"].(map[string]any)
	if getBody["user_id"] != createdUserID {
		t.Fatalf("unexpected user lookup payload: %#v", getBody)
	}

	patchResp := phase1DoJSON(
		t,
		server,
		http.MethodPatch,
		"/api/v1/users/"+createdUserID,
		map[string]any{
			"base_user_version": 1,
			"display_name":      "Phase1 E106 Patched",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	if patchBody["display_name"] != "Phase1 E106 Patched" || patchBody["user_version"] != float64(2) {
		t.Fatalf("unexpected patched user payload: %#v", patchBody)
	}

	userLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-06@example.test", "Phase1E106Pass!", "")
	nonAdminAction := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/users/"+createdUserID+"/sessions/revoke-all",
		map[string]any{
			"client_txn_id": "txn-e-1-06-non-admin",
			"reason":        "self revoke attempt",
		},
		withCookies(userLogin.sessionCookie, userLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, userLogin.csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, nonAdminAction, http.StatusUnauthorized, "session_required")

	socket := phase1ConnectSessionSocket(t, server, userLogin.sessionCookie.Value)
	defer socket.CloseNow()

	revokeResp := phase1DoJSON(
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

	phase1ExpectSessionRevoked(t, socket, "session_revoked")
	_ = phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-06@example.test", "Phase1E106Pass!", "")
}

func TestPhase1_AdminPasswordReset_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-07")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	user, secretBase32 := phase1ProvisionTOTPUser(t, server, adminSession, adminCSRF, "phase1-e-1-07@example.test", "Phase1 E107", "Phase1E107Pass!")
	targetUserID := user["user_id"].(string)

	targetLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-07@example.test", "Phase1E107Pass!", phase1GenerateTOTPCode(t, secretBase32))
	socket := phase1ConnectSessionSocket(t, server, targetLogin.sessionCookie.Value)
	defer socket.CloseNow()

	resetResp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/users/"+targetUserID+"/password/reset",
		map[string]any{
			"base_user_version": 2,
			"client_txn_id":     "txn-e-1-07-password-reset",
			"new_password":      "Phase1E107Reset!",
			"reason":            "e2e password reset",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	resetBody := httptestx.RequireSuccessEnvelope(t, resetResp, http.StatusOK)["data"].(map[string]any)
	if resetBody["user_version"] != float64(3) {
		t.Fatalf("unexpected password-reset payload: %#v", resetBody)
	}

	phase1ExpectSessionRevoked(t, socket, "session_revoked")

	oldPassword := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-07@example.test",
		"password": "Phase1E107Pass!",
	})
	httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

	newPasswordNoFactor := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-07@example.test",
		"password": "Phase1E107Reset!",
	})
	httptestx.RequireErrorEnvelope(t, newPasswordNoFactor, http.StatusUnauthorized, "mfa_required")

	_ = phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-07@example.test", "Phase1E107Reset!", phase1GenerateTOTPCode(t, secretBase32))
}

func TestPhase1_AdminTOTPResetAndBootstrapBoundaries_ProcessSmoke(t *testing.T) {
	server := startPhase1ServerProcess(t, "phase1-e-1-08")

	adminLogin, _ := phase1ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	user, secretBase32 := phase1ProvisionTOTPUser(t, server, adminSession, adminCSRF, "phase1-e-1-08@example.test", "Phase1 E108", "Phase1E108Pass!")
	targetUserID := user["user_id"].(string)

	targetLogin := phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-08@example.test", "Phase1E108Pass!", phase1GenerateTOTPCode(t, secretBase32))
	socket := phase1ConnectSessionSocket(t, server, targetLogin.sessionCookie.Value)
	defer socket.CloseNow()

	resetResp := phase1DoJSON(
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

	phase1ExpectSessionRevoked(t, socket, "session_revoked")

	bootstrapLogin := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-08@example.test",
		"password": "Phase1E108Pass!",
	})
	bootstrapBody := httptestx.RequireErrorEnvelope(t, bootstrapLogin, http.StatusUnauthorized, "mfa_setup_required")
	bootstrapDetails := bootstrapBody["error"].(map[string]any)["details"].(map[string]any)
	bootstrapToken := bootstrapDetails["bootstrap_token"].(string)

	rejectedRead := phase1DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withHeader("Authorization", "Bearer "+bootstrapToken))
	rejectedBody := httptestx.RequireErrorEnvelope(t, rejectedRead, http.StatusConflict, "credential_bootstrap_rejected")
	rejectedDetails := rejectedBody["error"].(map[string]any)["details"].(map[string]any)
	if rejectedDetails["reason_code"] != "not_allowed_for_route" {
		t.Fatalf("unexpected bootstrap route rejection: %#v", rejectedDetails)
	}

	phase1RequireBootstrapWebsocketRejected(t, server, bootstrapToken)

	begin := phase1BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-e-1-08-begin",
	})
	newSecretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	phase1CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), newSecretBase32, "txn-e-1-08-complete")

	_ = phase1LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-08@example.test", "Phase1E108Pass!", phase1GenerateTOTPCode(t, newSecretBase32))
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func startPhase1ServerProcess(t testing.TB, prefix string) *phase0ServerProcess {
	t.Helper()

	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		dropPhase0Database(t, postgresHarness, testDB.Name)
	})

	bucket := phase0BucketName(prefix)
	t.Cleanup(func() {
		cleanupPhase0Bucket(t, s3Harness, bucket)
	})

	configPath := writePhase0Config(t, string(fixtures.MustRead("config", "valid.toml")))
	env := phase0ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	env[enableTestRoutesEnv] = "1"

	server := startPhase0Server(t, env)
	t.Cleanup(func() {
		server.Stop(t)
	})
	server.WaitForReady(t)
	return server
}

func phase1ServerURL(server *phase0ServerProcess) string {
	return "http://" + server.address
}

func phase1DoJSON(t testing.TB, server *phase0ServerProcess, method string, path string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, phase1ServerURL(server)+path, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func phase1LoginLocalUser(t testing.TB, server *phase0ServerProcess, username string, password string) (*http.Cookie, *http.Cookie) {
	t.Helper()

	login := phase1LoginLocalUserWithSecondFactor(t, server, username, password, "")
	return login.sessionCookie, login.csrfCookie
}

func phase1LoginLocalUserWithSecondFactor(t testing.TB, server *phase0ServerProcess, username string, password string, code string) loginResult {
	t.Helper()

	body := map[string]any{
		"username": username,
		"password": password,
	}
	if strings.TrimSpace(code) != "" {
		body["second_factor"] = map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		}
	}

	resp := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected session and csrf cookies, got %#v", resp.Cookies())
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
}

func phase1RequireBootstrapLogin(t testing.TB, server *phase0ServerProcess, username string, password string) string {
	t.Helper()

	resp := phase1DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	token, _ := details["bootstrap_token"].(string)
	if token == "" {
		t.Fatalf("expected bootstrap_token on setup-required response, got %#v", details)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == authn.SessionCookieName {
			t.Fatal("bootstrap login response must not set a session cookie")
		}
	}
	return token
}

func phase1CreateUser(t testing.TB, server *phase0ServerProcess, adminSession *http.Cookie, adminCSRF *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/users",
		body,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func phase1ProvisionBootstrapAdmin(t testing.TB, server *phase0ServerProcess) (loginResult, string) {
	t.Helper()

	bootstrapToken := phase1RequireBootstrapLogin(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword)
	begin := phase1BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	phase1CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	return phase1LoginLocalUserWithSecondFactor(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword, phase1GenerateTOTPCode(t, secretBase32)), secretBase32
}

func phase1ProvisionTOTPUser(t testing.TB, server *phase0ServerProcess, adminSession *http.Cookie, adminCSRF *http.Cookie, email string, displayName string, password string) (map[string]any, string) {
	t.Helper()

	user := phase1CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "create-" + email,
		"auth_kind":        "local",
		"email":            email,
		"display_name":     displayName,
		"initial_password": password,
	})
	bootstrapToken := phase1RequireBootstrapLogin(t, server, email, password)
	begin := phase1BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "begin-" + email,
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	phase1CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "complete-"+email)
	return user, secretBase32
}

func phase1BeginTOTPEnrollment(t testing.TB, server *phase0ServerProcess, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/mfa/totp/begin",
		body,
		withHeader("Authorization", "Bearer "+bootstrapToken),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func phase1CompleteTOTPEnrollment(t testing.TB, server *phase0ServerProcess, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := phase1DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/mfa/totp/complete",
		map[string]any{
			"client_txn_id": clientTxnID,
			"enrollment_id": enrollmentID,
			"code":          phase1GenerateTOTPCode(t, secretBase32),
		},
		withHeader("Authorization", "Bearer "+bootstrapToken),
	)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func phase1GenerateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secretBase32, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	return code
}

func phase1ConnectSessionSocket(t testing.TB, server *phase0ServerProcess, sessionToken string) *websocket.Conn {
	t.Helper()

	target, err := url.Parse(phase1ServerURL(server))
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	target.Scheme = strings.Replace(target.Scheme, "http", "ws", 1)
	target.Path = "/ws/v1/test/session-lifecycle"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+sessionToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, target.String(), &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}

	var connected platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &connected); err != nil {
		t.Fatalf("read connected websocket message: %v", err)
	}
	if connected.Type != "connected" {
		t.Fatalf("unexpected first websocket message: %#v", connected)
	}
	return conn
}

func phase1ExpectSessionRevoked(t testing.TB, conn *websocket.Conn, wantReasonCode string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var revoked platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &revoked); err != nil {
		t.Fatalf("read session_revoked: %v", err)
	}
	if revoked.Type != "session_revoked" {
		t.Fatalf("unexpected websocket event: got %q want session_revoked", revoked.Type)
	}

	var payload map[string]any
	if err := json.Unmarshal(revoked.Payload, &payload); err != nil {
		t.Fatalf("decode session_revoked payload: %v", err)
	}
	if payload["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected session_revoked payload: %#v", payload)
	}

	var trailing platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &trailing); websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatalf("expected websocket close after session_revoked, got err=%v status=%v", err, websocket.CloseStatus(err))
	}
}

func phase1RequireBootstrapWebsocketRejected(t testing.TB, server *phase0ServerProcess, bootstrapToken string) {
	t.Helper()

	target, err := url.Parse(phase1ServerURL(server))
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	target.Scheme = strings.Replace(target.Scheme, "http", "ws", 1)
	target.Path = "/ws/v1/test/session-lifecycle"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+bootstrapToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, target.String(), &websocket.DialOptions{HTTPHeader: header})
	if err == nil {
		t.Fatal("expected bootstrap-token websocket dial to fail")
	}
	if resp == nil {
		t.Fatalf("expected HTTP rejection response for bootstrap-token websocket dial, err=%v", err)
	}
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "credential_bootstrap_rejected")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "not_allowed_for_route" {
		t.Fatalf("unexpected websocket bootstrap rejection: %#v", details)
	}
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func withHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}
