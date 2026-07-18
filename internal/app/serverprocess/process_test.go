package serverprocess

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/coder/websocket"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

const (
	phase1BootstrapAdminEmail    = "bootstrap-admin@example.test"
	phase1BootstrapAdminPassword = "BootstrapPass1!"
)

// These are process-level smoke tests for the standalone server binary.
// They are intentionally not the authoritative Authentication browser E2E ledger.
func TestLoginSessionAndLogout_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "phase1-e-1-01")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie

	sessionResp := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	sessionBody := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	if sessionBody["provider_type"] != "local" || sessionBody["is_deployment_admin"] != true || sessionBody["mfa_state"] != "satisfied" {
		t.Fatalf("unexpected session resource: %#v", sessionBody)
	}
	if memberships, ok := sessionBody["memberships"].([]any); !ok || len(memberships) != 0 {
		t.Fatalf("expected memberships to be present and empty, got %#v", sessionBody["memberships"])
	}

	logoutResp := DoJSON(
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

	postLogout := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	httptestx.RequireErrorEnvelope(t, postLogout, http.StatusUnauthorized, "session_required")
}

func TestCSRFFailClosed_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "phase1-e-1-02")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie

	missingHeader := DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
	)
	httptestx.RequireErrorEnvelope(t, missingHeader, http.StatusForbidden, "csrf_verification_failed")

	stillActive := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(adminSession))
	httptestx.RequireSuccessEnvelope(t, stillActive, http.StatusOK)

	wrongHeader := DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/auth/logout",
		map[string]any{},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, "wrong-token"),
	)
	httptestx.RequireErrorEnvelope(t, wrongHeader, http.StatusForbidden, "csrf_verification_failed")

	validHeader := DoJSON(
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

	server, db := startServerProcessWithDB(t, "phase1-e-1-03")
	t.Cleanup(func() {
		flowtest.ResetClockOffset(t, ServerURL(server))
	})

	initialLogin, adminSecret := ProvisionBootstrapAdmin(t, server)
	adminUserID := flowtest.QueryUserIDByEmail(t, db, phase1BootstrapAdminEmail)
	firstSessionID := flowtest.QuerySessionRow(t, db, adminUserID).SessionID
	socketIncidentID := CreateSocketIncident(t, server, initialLogin, "e-1-03-socket")

	sessions := make([]loginResult, 0, 6)
	sessions = append(sessions, initialLogin)
	for i := 0; i < 4; i++ {
		flowtest.SetClockOffset(t, ServerURL(server), int64(i+1))
		sessions = append(sessions, LoginLocalUserWithSecondFactor(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword, GenerateTOTPCode(t, adminSecret)))
	}

	socket := ConnectExistingIncidentSocket(t, server, socketIncidentID, sessions[0].sessionCookie.Value)
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	flowtest.SetClockOffset(t, ServerURL(server), 5)
	sessions = append(sessions, LoginLocalUserWithSecondFactor(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword, GenerateTOTPCode(t, adminSecret)))
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

	revokedSession := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(sessions[0].sessionCookie))
	httptestx.RequireErrorEnvelope(t, revokedSession, http.StatusUnauthorized, "session_required")

	activeSession := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withCookies(sessions[5].sessionCookie))
	httptestx.RequireSuccessEnvelope(t, activeSession, http.StatusOK)
}

func TestFirstEnrollmentFlow_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "phase1-e-1-04")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	user := CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-1-04-create",
		"auth_kind":        "local",
		"email":            "phase1-e-1-04@example.test",
		"display_name":     "Phase1 E104",
		"initial_password": "Phase1E104Pass!",
	})

	bootstrapToken := RequireBootstrapLogin(t, server, "phase1-e-1-04@example.test", "Phase1E104Pass!")
	begin := BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-e-1-04-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-e-1-04-complete")

	mfaRequired := DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-04@example.test",
		"password": "Phase1E104Pass!",
	})
	httptestx.RequireErrorEnvelope(t, mfaRequired, http.StatusUnauthorized, "mfa_required")

	userLogin := LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-04@example.test", "Phase1E104Pass!", GenerateTOTPCode(t, secretBase32))
	stateResp := DoJSON(t, server, http.MethodGet, "/api/v1/auth/credential-state", nil, withCookies(userLogin.sessionCookie))
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

	server := startServerProcess(t, "phase1-e-1-05")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	_, secretBase32 := ProvisionTOTPUser(t, server, adminSession, adminCSRF, "phase1-e-1-05@example.test", "Phase1 E105", "Phase1E105Pass!")

	userLogin := LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-05@example.test", "Phase1E105Pass!", GenerateTOTPCode(t, secretBase32))
	socket := ConnectSessionSocket(t, server, userLogin, "e-1-05-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	changeResp := DoJSON(
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
					"code": GenerateTOTPCode(t, secretBase32),
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

	ExpectSessionRevoked(t, socket, "session_revoked")

	oldPassword := DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-05@example.test",
		"password": "Phase1E105Pass!",
	})
	httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

	newPasswordNoFactor := DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-05@example.test",
		"password": "Phase1E105Changed!",
	})
	httptestx.RequireErrorEnvelope(t, newPasswordNoFactor, http.StatusUnauthorized, "mfa_required")

	_ = LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-05@example.test", "Phase1E105Changed!", GenerateTOTPCode(t, secretBase32))
}

func TestUserAdminAndRevokeAll_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "phase1-e-1-06")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	created := CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "txn-e-1-06-create",
		"auth_kind":        "local",
		"email":            "phase1-e-1-06@example.test",
		"display_name":     "Phase1 E106",
		"initial_password": "Phase1E106Pass!",
		"mfa_required":     false,
	})
	createdUserID := created["user_id"].(string)

	listResp := DoJSON(t, server, http.MethodGet, "/api/v1/users", nil, withCookies(adminSession))
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	users := listBody["users"].([]any)
	if len(users) < 2 {
		t.Fatalf("expected users list to include bootstrap admin and created user, got %#v", users)
	}

	getResp := DoJSON(t, server, http.MethodGet, "/api/v1/users/"+createdUserID, nil, withCookies(adminSession))
	getBody := httptestx.RequireSuccessEnvelope(t, getResp, http.StatusOK)["data"].(map[string]any)
	if getBody["user_id"] != createdUserID {
		t.Fatalf("unexpected user lookup payload: %#v", getBody)
	}

	patchResp := DoJSON(
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

	userLogin := LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-06@example.test", "Phase1E106Pass!", "")
	nonAdminAction := DoJSON(
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
	httptestx.RequireErrorEnvelope(t, nonAdminAction, http.StatusForbidden, "authorization_denied")

	socket := ConnectSessionSocket(t, server, userLogin, "e-1-06-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	revokeResp := DoJSON(
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

	ExpectSessionRevoked(t, socket, "session_revoked")
	_ = LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-06@example.test", "Phase1E106Pass!", "")
}

func TestAdminPasswordReset_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "phase1-e-1-07")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	user, secretBase32 := ProvisionTOTPUser(t, server, adminSession, adminCSRF, "phase1-e-1-07@example.test", "Phase1 E107", "Phase1E107Pass!")
	targetUserID := user["user_id"].(string)

	targetLogin := LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-07@example.test", "Phase1E107Pass!", GenerateTOTPCode(t, secretBase32))
	socket := ConnectSessionSocket(t, server, targetLogin, "e-1-07-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	resetResp := DoJSON(
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

	ExpectSessionRevoked(t, socket, "session_revoked")

	oldPassword := DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-07@example.test",
		"password": "Phase1E107Pass!",
	})
	httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

	newPasswordNoFactor := DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-07@example.test",
		"password": "Phase1E107Reset!",
	})
	httptestx.RequireErrorEnvelope(t, newPasswordNoFactor, http.StatusUnauthorized, "mfa_required")

	_ = LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-07@example.test", "Phase1E107Reset!", GenerateTOTPCode(t, secretBase32))
}

func TestAdminTOTPResetAndBootstrapBoundaries_Process(t *testing.T) {
	t.Parallel()

	server := startServerProcess(t, "phase1-e-1-08")

	adminLogin, _ := ProvisionBootstrapAdmin(t, server)
	adminSession := adminLogin.sessionCookie
	adminCSRF := adminLogin.csrfCookie
	user, secretBase32 := ProvisionTOTPUser(t, server, adminSession, adminCSRF, "phase1-e-1-08@example.test", "Phase1 E108", "Phase1E108Pass!")
	targetUserID := user["user_id"].(string)

	targetLogin := LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-08@example.test", "Phase1E108Pass!", GenerateTOTPCode(t, secretBase32))
	socket := ConnectSessionSocket(t, server, targetLogin, "e-1-08-socket")
	defer socket.Close(websocket.StatusNormalClosure, "process_smoke_cleanup")

	resetResp := DoJSON(
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

	ExpectSessionRevoked(t, socket, "session_revoked")

	bootstrapLogin := DoJSON(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "phase1-e-1-08@example.test",
		"password": "Phase1E108Pass!",
	})
	bootstrapBody := httptestx.RequireErrorEnvelope(t, bootstrapLogin, http.StatusUnauthorized, "mfa_setup_required")
	bootstrapDetails := bootstrapBody["error"].(map[string]any)["details"].(map[string]any)
	bootstrapToken := bootstrapDetails["bootstrap_token"].(string)

	rejectedRead := DoJSON(t, server, http.MethodGet, "/api/v1/auth/session", nil, withHeader("Authorization", "Bearer "+bootstrapToken))
	rejectedBody := httptestx.RequireErrorEnvelope(t, rejectedRead, http.StatusConflict, "credential_bootstrap_rejected")
	rejectedDetails := rejectedBody["error"].(map[string]any)["details"].(map[string]any)
	if rejectedDetails["reason_code"] != "not_allowed_for_route" {
		t.Fatalf("unexpected bootstrap route rejection: %#v", rejectedDetails)
	}

	incidentID := CreateSocketIncident(t, server, adminLogin, "e-1-08-bootstrap")
	RequireBootstrapWebsocketRejected(t, server, incidentID, bootstrapToken)

	begin := BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-e-1-08-begin",
	})
	newSecretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), newSecretBase32, "txn-e-1-08-complete")

	_ = LoginLocalUserWithSecondFactor(t, server, "phase1-e-1-08@example.test", "Phase1E108Pass!", GenerateTOTPCode(t, newSecretBase32))
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func startServerProcess(t testing.TB, prefix string) *processtest.Server {
	t.Helper()

	server, _ := startServerProcessWithDB(t, prefix)
	return server
}

func startServerProcessWithDB(t testing.TB, prefix string) (*processtest.Server, *sql.DB) {
	t.Helper()

	postgresHarness, s3Harness := sharedProcessHarnesses(t)

	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	bucket := BucketName(prefix)
	t.Cleanup(func() {
		cleanupBucket(t, s3Harness, bucket)
	})

	configPath := writeConfig(t, string(fixtures.MustRead("config", "valid.toml")))
	env := ServerEnv(t, testDB.Env(), s3Harness.Env(bucket), configPath, fixtures.Path("bootstrap-admin", "canonical.json"))
	env["CARTULARY_ENABLE_TEST_ROUTES"] = "1"
	env["CARTULARY_TEST_RUNTIME_MARKER"] = "harness-owned"
	env["CARTULARY_TEST_ROUTE_TOKEN"] = httptestx.TestRouteToken

	server := processtest.StartServer(t, processtest.ServerOptions{Env: env})
	t.Cleanup(func() {
		server.Stop(t)
	})
	server.WaitForReady(t)
	return server, db
}

func ServerURL(server *processtest.Server) string {
	return server.BaseURL
}

func DoJSON(t testing.TB, server *processtest.Server, method string, path string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()
	return flowtest.DoJSON(t, method, ServerURL(server)+path, body, options...)
}

func LoginLocalUserWithSecondFactor(t testing.TB, server *processtest.Server, username string, password string, code string) loginResult {
	t.Helper()
	if code == "" {
		sessionCookie, csrfCookie := flowtest.LoginLocalUser(t, ServerURL(server), username, password, nil)
		return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
	}
	login := flowtest.LoginLocalUserWithSecondFactor(t, ServerURL(server), username, password, code)
	return loginResult{sessionCookie: login.SessionCookie, csrfCookie: login.CSRFCookie}
}

func RequireBootstrapLogin(t testing.TB, server *processtest.Server, username string, password string) string {
	t.Helper()
	return flowtest.RequireBootstrapLogin(t, ServerURL(server), username, password)
}

func CreateUser(t testing.TB, server *processtest.Server, adminSession *http.Cookie, adminCSRF *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := DoJSON(
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

func ProvisionBootstrapAdmin(t testing.TB, server *processtest.Server) (loginResult, string) {
	t.Helper()

	bootstrapToken := RequireBootstrapLogin(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword)
	begin := BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-bootstrap-admin-complete")
	return LoginLocalUserWithSecondFactor(t, server, phase1BootstrapAdminEmail, phase1BootstrapAdminPassword, GenerateTOTPCode(t, secretBase32)), secretBase32
}

func ProvisionTOTPUser(t testing.TB, server *processtest.Server, adminSession *http.Cookie, adminCSRF *http.Cookie, email string, displayName string, password string) (map[string]any, string) {
	t.Helper()

	user := CreateUser(t, server, adminSession, adminCSRF, map[string]any{
		"client_txn_id":    "create-" + email,
		"auth_kind":        "local",
		"email":            email,
		"display_name":     displayName,
		"initial_password": password,
	})
	bootstrapToken := RequireBootstrapLogin(t, server, email, password)
	begin := BeginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "begin-" + email,
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	CompleteTOTPEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "complete-"+email)
	return user, secretBase32
}

func BeginTOTPEnrollment(t testing.TB, server *processtest.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()
	return flowtest.BeginTOTPEnrollment(t, ServerURL(server), bootstrapToken, body)
}

func CompleteTOTPEnrollment(t testing.TB, server *processtest.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()
	flowtest.CompleteInitialEnrollment(t, ServerURL(server), bootstrapToken, enrollmentID, secretBase32, clientTxnID)
}

func GenerateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()
	return flowtest.GenerateTOTPCode(t, secretBase32)
}

func ConnectSessionSocket(t testing.TB, server *processtest.Server, login loginResult, tag string) *flowtest.SessionSocketClient {
	t.Helper()
	incidentID := CreateSocketIncident(t, server, login, tag)
	return flowtest.ConnectSessionSocket(t, ServerURL(server), incidentID, login.sessionCookie.Value)
}

func ConnectExistingIncidentSocket(t testing.TB, server *processtest.Server, incidentID string, sessionToken string) *flowtest.SessionSocketClient {
	t.Helper()
	return flowtest.ConnectSessionSocket(t, ServerURL(server), incidentID, sessionToken)
}

func ExpectSessionRevoked(t testing.TB, conn *flowtest.SessionSocketClient, wantReasonCode string) {
	t.Helper()
	flowtest.ExpectSessionRevoked(t, conn, wantReasonCode)
}

func RequireBootstrapWebsocketRejected(t testing.TB, server *processtest.Server, incidentID string, bootstrapToken string) {
	t.Helper()
	flowtest.RequireBootstrapWebsocketRejected(t, ServerURL(server), incidentID, bootstrapToken)
}

func CreateSocketIncident(t testing.TB, server *processtest.Server, login loginResult, tag string) string {
	t.Helper()

	resp := DoJSON(
		t,
		server,
		http.MethodPost,
		"/api/v1/incidents",
		map[string]any{
			"client_txn_id": "txn-" + tag,
			"incident_key":  "IR-" + tag,
			"title":         "Process socket " + tag,
		},
		withCookies(login.sessionCookie, login.csrfCookie),
		withHeader(authn.CSRFHeaderName, login.csrfCookie.Value),
	)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
	return data["incident_id"].(string)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return flowtest.WithCookies(cookies...)
}

func withHeader(key string, value string) func(*http.Request) {
	return flowtest.WithHeader(key, value)
}
