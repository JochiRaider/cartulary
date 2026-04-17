package auth_test

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"example.com/todo/cartulary/internal/modules/auth"
	"example.com/todo/cartulary/internal/platform/authn"
	"example.com/todo/cartulary/internal/platform/httpapi"
	platformws "example.com/todo/cartulary/internal/platform/ws"
	"example.com/todo/cartulary/internal/testutil/fixtures"
	"example.com/todo/cartulary/internal/testutil/httptestx"
	"example.com/todo/cartulary/internal/testutil/pgtest"
	"example.com/todo/cartulary/internal/testutil/s3test"
)

func TestPhase1_LoginSessionLifecycle_I_1_01(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("persists login inspection idle sliding and logout", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-01-lifecycle")
		defer db.Close()

		userID := seedLocalUser(t, db, "analyst@example.test", "Analyst One", "  parol\u00e9 secret  ", false)
		sessionCookie, csrfCookie := loginLocalUser(t, server, "analyst@example.test", "  parol\u00e9 secret  ", nil)

		stored := querySessionRow(t, db, userID)
		if stored.revokedAt.Valid {
			t.Fatal("new login should create one active session")
		}

		sessionResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(sessionCookie))
		body := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)
		data := body["data"].(map[string]any)
		if got := data["user_id"]; got != userID {
			t.Fatalf("unexpected session user_id: got %v want %s", got, userID)
		}
		if got := data["provider_type"]; got != "local" {
			t.Fatalf("unexpected provider_type: got %v", got)
		}
		if memberships, ok := data["memberships"].([]any); !ok || memberships == nil {
			t.Fatalf("expected memberships[] to be present, got %T", data["memberships"])
		}

		afterInspect := querySessionByID(t, db, stored.sessionID)
		if !afterInspect.lastQualifyingActivityAt.Equal(stored.lastQualifyingActivityAt) {
			t.Fatalf("session inspection must not slide idle expiry: before=%s after=%s", stored.lastQualifyingActivityAt, afterInspect.lastQualifyingActivityAt)
		}

		time.Sleep(20 * time.Millisecond)
		touchResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/test/auth/touch", nil, withCookies(sessionCookie))
		httptestx.RequireSuccessEnvelope(t, touchResp, http.StatusOK)

		afterTouch := querySessionByID(t, db, stored.sessionID)
		if !afterTouch.lastQualifyingActivityAt.After(afterInspect.lastQualifyingActivityAt) {
			t.Fatalf("expected touch route to advance last_qualifying_activity_at: before=%s after=%s", afterInspect.lastQualifyingActivityAt, afterTouch.lastQualifyingActivityAt)
		}
		if !afterTouch.idleExpiresAt.After(afterInspect.idleExpiresAt) {
			t.Fatalf("expected touch route to slide idle expiry: before=%s after=%s", afterInspect.idleExpiresAt, afterTouch.idleExpiresAt)
		}

		logoutResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/logout", nil, withCookies(sessionCookie, csrfCookie), withHeader(authn.CSRFHeaderName, csrfCookie.Value))
		httptestx.RequireSuccessEnvelope(t, logoutResp, http.StatusOK)

		afterLogout := querySessionByID(t, db, stored.sessionID)
		if !afterLogout.revokedAt.Valid {
			t.Fatal("logout must revoke the current session")
		}
		if afterLogout.revokeReasonCode.String != "session_revoked" {
			t.Fatalf("unexpected logout revoke_reason_code: got %q", afterLogout.revokeReasonCode.String)
		}
	})

	t.Run("sixth login revokes least recently used non-current session", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-01-concurrency")
		defer db.Close()

		userID := seedLocalUser(t, db, "analyst2@example.test", "Analyst Two", "ConcurrencyPass1!", false)

		var firstSessionID string
		for i := 0; i < 6; i++ {
			sessionCookie, _ := loginLocalUser(t, server, "analyst2@example.test", "ConcurrencyPass1!", nil)
			if i == 0 {
				firstSessionID = querySessionRow(t, db, userID).sessionID
			}
			_ = sessionCookie
			time.Sleep(20 * time.Millisecond)
		}

		activeCount := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND revoked_at IS NULL`, userID)
		if activeCount != 5 {
			t.Fatalf("expected five active sessions after sixth login, got %d", activeCount)
		}

		firstSession := querySessionByID(t, db, firstSessionID)
		if !firstSession.revokedAt.Valid {
			t.Fatal("expected least-recently-used non-current session to be revoked")
		}
		if firstSession.revokeReasonCode.String != authn.ConcurrencyLimitReasonCode {
			t.Fatalf("unexpected concurrency revoke_reason_code: got %q", firstSession.revokeReasonCode.String)
		}

		auditCount := queryCount(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE target_user_id = $1 AND reason_code = $2`, userID, authn.ConcurrencyLimitReasonCode)
		if auditCount != 1 {
			t.Fatalf("expected one concurrency_limit audit event, got %d", auditCount)
		}
	})
}

func TestPhase1_SessionRevocationClosesAttachedSocket_I_1_02(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-02-session-revoked")
	defer db.Close()

	seedLocalUser(t, db, "socket-owner@example.test", "Socket Owner", "SocketPass123!", false)
	sessionCookie, csrfCookie := loginLocalUser(t, server, "socket-owner@example.test", "SocketPass123!", nil)

	socketURL, err := url.Parse(server.HTTP.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	socketURL.Scheme = strings.Replace(socketURL.Scheme, "http", "ws", 1)
	socketURL.Path = "/ws/v1/test/session-lifecycle"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+sessionCookie.Value)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, socketURL.String(), &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.CloseNow()

	var connected platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &connected); err != nil {
		t.Fatalf("read connected message: %v", err)
	}
	if connected.Type != "connected" {
		t.Fatalf("unexpected first websocket message type: got %q want %q", connected.Type, "connected")
	}

	logoutResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/logout",
		nil,
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, logoutResp, http.StatusOK)

	var revoked platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &revoked); err != nil {
		t.Fatalf("read session_revoked message: %v", err)
	}
	if revoked.Type != "session_revoked" {
		t.Fatalf("unexpected revocation message type: got %q want %q", revoked.Type, "session_revoked")
	}

	var payload map[string]any
	if err := json.Unmarshal(revoked.Payload, &payload); err != nil {
		t.Fatalf("decode session_revoked payload: %v", err)
	}
	if got := payload["reason_code"]; got != "session_revoked" {
		t.Fatalf("unexpected session_revoked reason_code: got %v want %q", got, "session_revoked")
	}

	var trailing platformws.Message
	err = platformws.ReadJSON(ctx, conn, &trailing)
	if status := websocket.CloseStatus(err); status != websocket.StatusPolicyViolation {
		t.Fatalf("expected websocket close after session_revoked, got err=%v status=%v", err, status)
	}
}

func TestPhase1_CredentialStateAndBootstrapFlows_I_1_04(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("first enrollment then password change revokes all sessions", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-04-first-enrollment")
		defer db.Close()

		userID := seedLocalUser(t, db, "mfa-user@example.test", "MFA User", "BootstrapPass123!", true)
		bootstrapToken := requireBootstrapLogin(t, server, "mfa-user@example.test", "BootstrapPass123!")

		begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
			"client_txn_id": "txn-totp-begin-1",
		})
		setup := begin["totp_setup"].(map[string]any)
		secretBase32 := setup["secret_base32"].(string)
		enrollmentID := begin["enrollment_id"].(string)

		beginReplay := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
			"client_txn_id": "txn-totp-begin-1",
		})
		if got := beginReplay["enrollment_id"]; got != enrollmentID {
			t.Fatalf("expected begin replay to return original enrollment_id, got %v want %s", got, enrollmentID)
		}
		replaySetup := beginReplay["totp_setup"].(map[string]any)
		if got := replaySetup["secret_base32"]; got != secretBase32 {
			t.Fatalf("expected begin replay to return original secret_base32, got %v want %s", got, secretBase32)
		}

		completeResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
			"client_txn_id": "txn-totp-complete-1",
			"enrollment_id": enrollmentID,
			"code":          generateTOTPCode(t, secretBase32),
		}, withHeader("Authorization", "Bearer "+bootstrapToken))
		completeBody := httptestx.RequireSuccessEnvelope(t, completeResp, http.StatusOK)
		if data := completeBody["data"].(map[string]any); data["sessions_revoked"] != false {
			t.Fatalf("first enrollment complete must not revoke sessions, got %v", data["sessions_revoked"])
		}

		loginWithoutFactor := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
			"username": "mfa-user@example.test",
			"password": "BootstrapPass123!",
		})
		httptestx.RequireErrorEnvelope(t, loginWithoutFactor, http.StatusUnauthorized, "mfa_required")

		loginWithFactor := loginLocalUserWithSecondFactor(t, server, "mfa-user@example.test", "BootstrapPass123!", generateTOTPCode(t, secretBase32))
		sessionCookie, csrfCookie := loginWithFactor.sessionCookie, loginWithFactor.csrfCookie

		credentialStateResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/credential-state", nil, withCookies(sessionCookie))
		credentialState := httptestx.RequireSuccessEnvelope(t, credentialStateResp, http.StatusOK)["data"].(map[string]any)
		totpState := credentialState["totp"].(map[string]any)
		if got := totpState["state"]; got != "active" {
			t.Fatalf("unexpected credential-state totp.state: got %v want active", got)
		}
		if credentialState["user_id"] != userID {
			t.Fatalf("unexpected credential-state user_id: got %v want %s", credentialState["user_id"], userID)
		}

		changeResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/password/change", map[string]any{
			"client_txn_id":    "txn-password-change-1",
			"current_password": "BootstrapPass123!",
			"new_password":     "ChangedPass12345!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": generateTOTPCode(t, secretBase32),
				},
			},
		}, withCookies(sessionCookie, csrfCookie), withHeader(authn.CSRFHeaderName, csrfCookie.Value))
		changeData := httptestx.RequireSuccessEnvelope(t, changeResp, http.StatusOK)["data"].(map[string]any)
		if changeData["sessions_revoked"] != true {
			t.Fatalf("password change must revoke all sessions, got %v", changeData["sessions_revoked"])
		}

		postChangeSession := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(sessionCookie))
		httptestx.RequireErrorEnvelope(t, postChangeSession, http.StatusUnauthorized, "session_required")

		activeSessions := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND revoked_at IS NULL`, userID)
		if activeSessions != 0 {
			t.Fatalf("expected password change to revoke all sessions, found %d active", activeSessions)
		}

		idempotencyCount := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND actor_user_id::text = $2`, "auth.password.change", userID)
		if idempotencyCount != 1 {
			t.Fatalf("expected one password-change idempotency record, got %d", idempotencyCount)
		}
	})

	t.Run("replacement enrollment revokes current session and swaps the active factor", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-04-replacement-enrollment")
		defer db.Close()

		seedLocalUser(t, db, "replace-user@example.test", "Replace User", "ReplacePass123!", true)
		bootstrapToken := requireBootstrapLogin(t, server, "replace-user@example.test", "ReplacePass123!")

		firstBegin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
			"client_txn_id": "txn-initial-begin",
		})
		firstSecret := firstBegin["totp_setup"].(map[string]any)["secret_base32"].(string)
		completeInitialEnrollment(t, server, bootstrapToken, firstBegin["enrollment_id"].(string), firstSecret, "txn-initial-complete")

		login := loginLocalUserWithSecondFactor(t, server, "replace-user@example.test", "ReplacePass123!", generateTOTPCode(t, firstSecret))

		replaceBegin := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{
			"client_txn_id":    "txn-replace-begin",
			"current_password": "ReplacePass123!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": generateTOTPCode(t, firstSecret),
				},
			},
		}, withCookies(login.sessionCookie), withHeader("Authorization", "Bearer "+login.sessionCookie.Value))
		replaceBeginData := httptestx.RequireSuccessEnvelope(t, replaceBegin, http.StatusOK)["data"].(map[string]any)
		replaceSecret := replaceBeginData["totp_setup"].(map[string]any)["secret_base32"].(string)
		replaceEnrollmentID := replaceBeginData["enrollment_id"].(string)

		replaceComplete := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
			"client_txn_id": "txn-replace-complete",
			"enrollment_id": replaceEnrollmentID,
			"code":          generateTOTPCode(t, replaceSecret),
		}, withHeader("Authorization", "Bearer "+login.sessionCookie.Value))
		replaceCompleteData := httptestx.RequireSuccessEnvelope(t, replaceComplete, http.StatusOK)["data"].(map[string]any)
		if replaceCompleteData["sessions_revoked"] != true {
			t.Fatalf("replacement complete must revoke all sessions, got %v", replaceCompleteData["sessions_revoked"])
		}

		revokedSession := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(login.sessionCookie))
		httptestx.RequireErrorEnvelope(t, revokedSession, http.StatusUnauthorized, "session_required")

		oldFactorLogin := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
			"username": "replace-user@example.test",
			"password": "ReplacePass123!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": generateTOTPCode(t, firstSecret),
				},
			},
		})
		httptestx.RequireErrorEnvelope(t, oldFactorLogin, http.StatusUnauthorized, "invalid_second_factor")

		_ = loginLocalUserWithSecondFactor(t, server, "replace-user@example.test", "ReplacePass123!", generateTOTPCode(t, replaceSecret))
	})
}

func TestPhase1_BootstrapTokenRouteBoundaries_I_1_06(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-06-bootstrap-boundaries")
	defer db.Close()

	seedLocalUser(t, db, "bootstrap-boundary@example.test", "Bootstrap Boundary", "BootstrapRoute123!", true)
	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-boundary@example.test", "BootstrapRoute123!")

	for _, path := range []string{
		"/api/v1/auth/session",
		"/api/v1/auth/credential-state",
		"/api/v1/test/auth/touch",
	} {
		resp := doJSON(t, http.MethodGet, server.HTTP.URL+path, nil, withHeader("Authorization", "Bearer "+bootstrapToken))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "credential_bootstrap_rejected")
		details := body["error"].(map[string]any)["details"].(map[string]any)
		if got := details["reason_code"]; got != "not_allowed_for_route" {
			t.Fatalf("unexpected bootstrap rejection reason for %s: got %v want not_allowed_for_route", path, got)
		}
	}

	socketURL, err := url.Parse(server.HTTP.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	socketURL.Scheme = strings.Replace(socketURL.Scheme, "http", "ws", 1)
	socketURL.Path = "/ws/v1/test/session-lifecycle"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+bootstrapToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, socketURL.String(), &websocket.DialOptions{HTTPHeader: header})
	if err == nil {
		t.Fatal("expected websocket dial with bootstrap_token to fail")
	}
	if resp == nil {
		t.Fatalf("expected non-upgrade HTTP response for bootstrap_token websocket dial, err=%v", err)
	}
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "credential_bootstrap_rejected")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if got := details["reason_code"]; got != "not_allowed_for_route" {
		t.Fatalf("unexpected websocket bootstrap rejection reason: got %v want not_allowed_for_route", got)
	}
}

func TestPhase1_UserAdminLifecycle_I_1_03(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-03-user-admin")
	defer db.Close()

	adminID := seedLocalUserFlags(t, db, "admin-users@example.test", "Users Admin", "AdminUsersPass123!", false, true, true)
	adminSession, adminCSRF := loginLocalUser(t, server, "admin-users@example.test", "AdminUsersPass123!", nil)

	createResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users", map[string]any{
		"client_txn_id":    "txn-user-create-1",
		"auth_kind":        "local",
		"email":            "created-user@example.test",
		"display_name":     "Created User",
		"initial_password": "CreatedPass123!",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("user create failed: status=%d body=%#v", createResp.StatusCode, httptestx.ReadJSONBody(t, createResp))
	}
	createBody := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	createdUserID := createBody["user_id"].(string)
	if createBody["mfa_required"] != true || createBody["is_deployment_admin"] != false || createBody["is_active"] != true {
		t.Fatalf("unexpected user-create defaults: %#v", createBody)
	}

	replayResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users", map[string]any{
		"client_txn_id":    "txn-user-create-1",
		"auth_kind":        "local",
		"email":            "created-user@example.test",
		"display_name":     "Created User",
		"initial_password": "CreatedPass123!",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	replayBody := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	if replayBody["user_id"] != createdUserID {
		t.Fatalf("expected create replay to return original user_id, got %v want %s", replayBody["user_id"], createdUserID)
	}

	listResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users", nil, withCookies(adminSession))
	listBody := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)
	users := listBody["data"].(map[string]any)["users"].([]any)
	if len(users) < 2 {
		t.Fatalf("expected users list to include admin and created user, got %#v", users)
	}
	meta := listBody["meta"].(map[string]any)
	paging := meta["paging"].(map[string]any)
	if paging["limit"] != float64(100) || paging["has_more"] != false || paging["next_cursor"] != nil {
		t.Fatalf("unexpected users paging contract: %#v", paging)
	}

	getResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+createdUserID, nil, withCookies(adminSession))
	getBody := httptestx.RequireSuccessEnvelope(t, getResp, http.StatusOK)["data"].(map[string]any)
	if getBody["user_id"] != createdUserID {
		t.Fatalf("unexpected get user_id: got %v want %s", getBody["user_id"], createdUserID)
	}

	patchResp := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+createdUserID, map[string]any{
		"base_user_version": 1,
		"display_name":      "Created User Patched",
		"mfa_required":      false,
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("user patch failed: status=%d body=%#v", patchResp.StatusCode, httptestx.ReadJSONBody(t, patchResp))
	}
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)
	if patchBody["display_name"] != "Created User Patched" || patchBody["user_version"] != float64(2) {
		t.Fatalf("unexpected patched user resource: %#v", patchBody)
	}

	stalePatch := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+createdUserID, map[string]any{
		"base_user_version": 1,
		"display_name":      "Stale Patch",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, stalePatch, http.StatusConflict, "user_version_conflict")

	var bootstrapAdminID string
	if err := db.QueryRowContext(context.Background(), `SELECT id::text FROM users WHERE email = $1`, "bootstrap-admin@example.test").Scan(&bootstrapAdminID); err != nil {
		t.Fatalf("lookup bootstrap admin user: %v", err)
	}
	bootstrapAdminDemotion := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+bootstrapAdminID, map[string]any{
		"base_user_version":   1,
		"is_deployment_admin": false,
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, bootstrapAdminDemotion, http.StatusOK)

	lastAdminGuard := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+adminID, map[string]any{
		"base_user_version":   1,
		"is_deployment_admin": false,
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, lastAdminGuard, http.StatusConflict, "last_deployment_admin")

	auditCount := queryCount(t, db, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE target_user_id::text = $1`, createdUserID)
	if auditCount < 2 {
		t.Fatalf("expected create and patch audit records, got %d", auditCount)
	}
}

func TestPhase1_AdminCredentialActions_I_1_05(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("password reset revokes attached sockets and preserves active totp", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-password-reset")
		defer db.Close()

		adminID := seedLocalUserFlags(t, db, "admin-reset@example.test", "Reset Admin", "ResetAdminPass123!", false, true, true)
		_ = adminID
		adminSession, adminCSRF := loginLocalUser(t, server, "admin-reset@example.test", "ResetAdminPass123!", nil)

		targetSecret := "JBSWY3DPEHPK3PXP"
		targetID := seedLocalUserWithActiveTOTP(t, db, "target-reset@example.test", "Target Reset", "TargetResetPass123!", true, false, targetSecret)
		targetLogin := loginLocalUserWithSecondFactor(t, server, "target-reset@example.test", "TargetResetPass123!", generateTOTPCode(t, targetSecret))
		targetSocket := connectSessionSocket(t, server, targetLogin.sessionCookie.Value)
		defer targetSocket.CloseNow()

		resetResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/password/reset", map[string]any{
			"base_user_version": 1,
			"client_txn_id":     "txn-admin-password-reset-1",
			"new_password":      "TargetResetChanged123!",
			"reason":            "admin reset",
		}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
		if resetResp.StatusCode != http.StatusOK {
			t.Fatalf("admin password reset failed: status=%d body=%#v", resetResp.StatusCode, httptestx.ReadJSONBody(t, resetResp))
		}
		resetBody := httptestx.RequireSuccessEnvelope(t, resetResp, http.StatusOK)["data"].(map[string]any)
		if resetBody["user_version"] != float64(2) {
			t.Fatalf("unexpected user_version after password reset: %#v", resetBody)
		}

		expectSessionRevoked(t, targetSocket, "session_revoked")

		oldPassword := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
			"username": "target-reset@example.test",
			"password": "TargetResetPass123!",
		})
		httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

		mfaRequired := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
			"username": "target-reset@example.test",
			"password": "TargetResetChanged123!",
		})
		httptestx.RequireErrorEnvelope(t, mfaRequired, http.StatusUnauthorized, "mfa_required")

		_ = loginLocalUserWithSecondFactor(t, server, "target-reset@example.test", "TargetResetChanged123!", generateTOTPCode(t, targetSecret))
	})

	t.Run("totp reset revokes attached sockets and reopens bootstrap flow", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-totp-reset")
		defer db.Close()

		seedLocalUserFlags(t, db, "admin-totp-reset@example.test", "TOTP Reset Admin", "TotpResetAdmin123!", false, true, true)
		adminSession, adminCSRF := loginLocalUser(t, server, "admin-totp-reset@example.test", "TotpResetAdmin123!", nil)

		targetSecret := "JBSWY3DPEHPK3QAA"
		targetID := seedLocalUserWithActiveTOTP(t, db, "target-totp-reset@example.test", "Target TOTP Reset", "TargetTotpPass123!", true, false, targetSecret)
		targetLogin := loginLocalUserWithSecondFactor(t, server, "target-totp-reset@example.test", "TargetTotpPass123!", generateTOTPCode(t, targetSecret))
		targetSocket := connectSessionSocket(t, server, targetLogin.sessionCookie.Value)
		defer targetSocket.CloseNow()

		resetResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/mfa/totp/reset", map[string]any{
			"base_user_version": 1,
			"client_txn_id":     "txn-admin-totp-reset-1",
			"reason":            "reset factor",
		}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
		if resetResp.StatusCode != http.StatusOK {
			t.Fatalf("admin totp reset failed: status=%d body=%#v", resetResp.StatusCode, httptestx.ReadJSONBody(t, resetResp))
		}
		resetBody := httptestx.RequireSuccessEnvelope(t, resetResp, http.StatusOK)["data"].(map[string]any)
		if resetBody["user_version"] != float64(2) {
			t.Fatalf("unexpected user_version after totp reset: %#v", resetBody)
		}

		expectSessionRevoked(t, targetSocket, "session_revoked")

		bootstrapBody := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
			"username": "target-totp-reset@example.test",
			"password": "TargetTotpPass123!",
		})
		httptestx.RequireErrorEnvelope(t, bootstrapBody, http.StatusUnauthorized, "mfa_setup_required")
	})

	t.Run("revoke-all revokes attached sockets without mutating credentials", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-revoke-all")
		defer db.Close()

		seedLocalUserFlags(t, db, "admin-revoke-all@example.test", "Revoke All Admin", "RevokeAllAdmin123!", false, true, true)
		adminSession, adminCSRF := loginLocalUser(t, server, "admin-revoke-all@example.test", "RevokeAllAdmin123!", nil)

		targetSecret := "JBSWY3DPEHPK3QAB"
		targetID := seedLocalUserWithActiveTOTP(t, db, "target-revoke-all@example.test", "Target Revoke All", "TargetRevokePass123!", true, false, targetSecret)
		targetLogin := loginLocalUserWithSecondFactor(t, server, "target-revoke-all@example.test", "TargetRevokePass123!", generateTOTPCode(t, targetSecret))
		targetSocket := connectSessionSocket(t, server, targetLogin.sessionCookie.Value)
		defer targetSocket.CloseNow()

		revokeResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/sessions/revoke-all", map[string]any{
			"client_txn_id": "txn-admin-revoke-all-1",
			"reason":        "revoke every session",
		}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
		if revokeResp.StatusCode != http.StatusOK {
			t.Fatalf("admin revoke-all failed: status=%d body=%#v", revokeResp.StatusCode, httptestx.ReadJSONBody(t, revokeResp))
		}
		revokeBody := httptestx.RequireSuccessEnvelope(t, revokeResp, http.StatusOK)["data"].(map[string]any)
		if revokeBody["sessions_revoked"] != true {
			t.Fatalf("unexpected revoke-all response: %#v", revokeBody)
		}

		expectSessionRevoked(t, targetSocket, "session_revoked")

		_ = loginLocalUserWithSecondFactor(t, server, "target-revoke-all@example.test", "TargetRevokePass123!", generateTOTPCode(t, targetSecret))
	})
}

type sessionRow struct {
	sessionID                string
	lastQualifyingActivityAt time.Time
	idleExpiresAt            time.Time
	revokedAt                sql.NullTime
	revokeReasonCode         sql.NullString
}

func startPhase1Server(t testing.TB, postgresHarness *pgtest.Harness, s3Harness *s3test.Harness, prefix string) (*httptestx.Server, *sql.DB) {
	t.Helper()

	testDB, _, err := postgresHarness.PrepareDatabase(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare postgres database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresHarness.DropDatabase(context.Background(), testDB.Name); err != nil {
			t.Fatalf("drop postgres database: %v", err)
		}
	})

	bucket, err := s3Harness.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	t.Cleanup(func() {
		if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Logf("cleanup bucket: %v", err)
		}
	})

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{
		Env:              env,
		AdditionalRoutes: []httpapi.RouteRegistrar{auth.RegisterTestRoutes()},
	})

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open postgres sql handle: %v", err)
	}
	return server, db
}

func seedLocalUser(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, true, false)
RETURNING id::text
`, email, displayName, hash, mfaRequired).Scan(&userID); err != nil {
		t.Fatalf("seed local user: %v", err)
	}
	return userID
}

func seedLocalUserFlags(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, isActive bool) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text
`, email, displayName, hash, mfaRequired, isActive, isDeploymentAdmin).Scan(&userID); err != nil {
		t.Fatalf("seed local user with flags: %v", err)
	}
	return userID
}

func seedLocalUserWithActiveTOTP(t testing.TB, db *sql.DB, email string, displayName string, password string, mfaRequired bool, isDeploymentAdmin bool, secretBase32 string) string {
	t.Helper()

	hash, err := authn.HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	keys, err := authn.LoadMasterKeys(nil)
	if err != nil {
		t.Fatalf("load auth master keys: %v", err)
	}
	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secretBase32)
	if err != nil {
		t.Fatalf("decode base32 totp secret: %v", err)
	}
	ciphertext, nonce, err := authn.EncryptSecret(keys, secretBytes)
	if err != nil {
		t.Fatalf("encrypt totp secret: %v", err)
	}

	var userID string
	if err := db.QueryRowContext(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin, totp_enrolled_at, totp_secret_ciphertext, totp_secret_nonce)
VALUES ($1, $2, $3, $4, true, $5, now(), $6, $7)
RETURNING id::text
`, email, displayName, hash, mfaRequired, isDeploymentAdmin, ciphertext, nonce).Scan(&userID); err != nil {
		t.Fatalf("seed local user with totp: %v", err)
	}
	return userID
}

func loginLocalUser(t testing.TB, server *httptestx.Server, username string, password string, headers func(*http.Request)) (*http.Cookie, *http.Cookie) {
	t.Helper()

	req := httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if headers != nil {
		headers(req)
	}
	resp := httptestx.Do(t, server.HTTP.Client(), req)
	if resp.StatusCode != http.StatusOK {
		body := httptestx.ReadJSONBody(t, resp)
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, body)
	}
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	_ = body

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
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}

	return sessionCookie, csrfCookie
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	req := httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		},
	})
	resp := httptestx.Do(t, server.HTTP.Client(), req)
	if resp.StatusCode != http.StatusOK {
		body := httptestx.ReadJSONBody(t, resp)
		t.Fatalf("login with second factor failed: status=%d body=%#v", resp.StatusCode, body)
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
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}

	return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	token, _ := details["bootstrap_token"].(string)
	if token == "" {
		t.Fatalf("expected bootstrap_token on mfa_setup_required response, got %#v", details)
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == authn.SessionCookieName {
			t.Fatal("mfa_setup_required must not set a session cookie")
		}
	}
	return token
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, withHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
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

func connectSessionSocket(t testing.TB, server *httptestx.Server, sessionToken string) *websocket.Conn {
	t.Helper()

	socketURL, err := url.Parse(server.HTTP.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	socketURL.Scheme = strings.Replace(socketURL.Scheme, "http", "ws", 1)
	socketURL.Path = "/ws/v1/test/session-lifecycle"

	header := http.Header{}
	header.Set("Authorization", "Bearer "+sessionToken)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, socketURL.String(), &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}

	var connected platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &connected); err != nil {
		t.Fatalf("read connected websocket message: %v", err)
	}
	if connected.Type != "connected" {
		t.Fatalf("unexpected first websocket message type: got %q want %q", connected.Type, "connected")
	}
	return conn
}

func expectSessionRevoked(t testing.TB, conn *websocket.Conn, wantReasonCode string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var revoked platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &revoked); err != nil {
		t.Fatalf("read session_revoked message: %v", err)
	}
	if revoked.Type != "session_revoked" {
		t.Fatalf("unexpected revocation message type: got %q want %q", revoked.Type, "session_revoked")
	}

	var payload map[string]any
	if err := json.Unmarshal(revoked.Payload, &payload); err != nil {
		t.Fatalf("decode session_revoked payload: %v", err)
	}
	if got := payload["reason_code"]; got != wantReasonCode {
		t.Fatalf("unexpected session_revoked reason_code: got %v want %q", got, wantReasonCode)
	}

	var trailing platformws.Message
	if err := platformws.ReadJSON(ctx, conn, &trailing); websocket.CloseStatus(err) != websocket.StatusPolicyViolation {
		t.Fatalf("expected websocket close after session_revoked, got err=%v status=%v", err, websocket.CloseStatus(err))
	}
}

func doJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
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

func querySessionRow(t testing.TB, db *sql.DB, userID string) sessionRow {
	t.Helper()

	row := querySingleSession(t, db, `SELECT id::text, last_qualifying_activity_at, idle_expires_at, revoked_at, revoke_reason_code FROM user_sessions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID)
	return row
}

func querySessionByID(t testing.TB, db *sql.DB, sessionID string) sessionRow {
	t.Helper()
	return querySingleSession(t, db, `SELECT id::text, last_qualifying_activity_at, idle_expires_at, revoked_at, revoke_reason_code FROM user_sessions WHERE id::text = $1`, sessionID)
}

func querySingleSession(t testing.TB, db *sql.DB, query string, args ...any) sessionRow {
	t.Helper()

	var row sessionRow
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(
		&row.sessionID,
		&row.lastQualifyingActivityAt,
		&row.idleExpiresAt,
		&row.revokedAt,
		&row.revokeReasonCode,
	); err != nil {
		t.Fatalf("query session row: %v", err)
	}
	return row
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}
