package auth_test

import (
	"context"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase1test"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
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
		if !stored.sessionExpiresAt.Equal(stored.idleExpiresAt) {
			t.Fatalf("expected initial session_expires_at to match idle_expires_at: session=%s idle=%s", stored.sessionExpiresAt, stored.idleExpiresAt)
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
		for _, key := range []string{"authenticated_at", "idle_expires_at", "absolute_expires_at", "session_expires_at"} {
			if got, ok := data[key].(string); !ok || got == "" {
				t.Fatalf("expected session resource to expose %s, got %#v", key, data[key])
			}
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

	t.Run("idle expiry fails closed and records session_expired", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-01-expiry")
		defer db.Close()

		userID := seedLocalUser(t, db, "idle-expiry@example.test", "Idle Expiry", "IdleExpiryPass1!", false)
		sessionCookie, _ := loginLocalUser(t, server, "idle-expiry@example.test", "IdleExpiryPass1!", nil)
		sessionBeforeExpiry := querySessionRow(t, db, userID)

		phase1test.WithClockOffset(
			t,
			server.HTTP.URL,
			int64((30*time.Minute/time.Second)+1),
		)

		expiredSession := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(sessionCookie))
		httptestx.RequireErrorEnvelope(t, expiredSession, http.StatusUnauthorized, "session_required")

		sessionAfterExpiry := querySessionByID(t, db, sessionBeforeExpiry.sessionID)
		if !sessionAfterExpiry.revokedAt.Valid {
			t.Fatal("expected expired session to be revoked on next authenticated request")
		}
		if sessionAfterExpiry.revokeReasonCode.String != "session_expired" {
			t.Fatalf("unexpected expiry revoke_reason_code: got %q", sessionAfterExpiry.revokeReasonCode.String)
		}
	})

	t.Run("sixth login revokes least recently used non-current session", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-01-concurrency")
		defer db.Close()
		t.Cleanup(func() {
			phase1test.ResetClockOffset(t, server.HTTP.URL)
		})

		userID := seedLocalUser(t, db, "analyst2@example.test", "Analyst Two", "ConcurrencyPass1!", false)

		var firstSessionID string
		for i := 0; i < 6; i++ {
			phase1test.SetClockOffset(t, server.HTTP.URL, int64(i))
			sessionCookie, _ := loginLocalUser(t, server, "analyst2@example.test", "ConcurrencyPass1!", nil)
			if i == 0 {
				firstSessionID = querySessionRow(t, db, userID).sessionID
			}
			_ = sessionCookie
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

	t.Run("logout revokes attached session socket", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-02-session-revoked")
		defer db.Close()

		seedLocalUser(t, db, "socket-owner@example.test", "Socket Owner", "SocketPass123!", false)
		sessionCookie, csrfCookie := loginLocalUser(t, server, "socket-owner@example.test", "SocketPass123!", nil)
		socket := connectSessionSocket(t, server, sessionCookie.Value)
		defer socket.Close(websocket.StatusNormalClosure, "integration_cleanup")

		logoutResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/auth/logout",
			nil,
			withCookies(sessionCookie, csrfCookie),
			withHeader(authn.CSRFHeaderName, csrfCookie.Value),
		)
		httptestx.RequireSuccessEnvelope(t, logoutResp, http.StatusOK)
		expectSessionRevoked(t, socket, "session_revoked")
	})

	t.Run("concurrency limit revokes attached least recently used session socket", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-02-concurrency-socket")
		defer db.Close()
		t.Cleanup(func() {
			phase1test.ResetClockOffset(t, server.HTTP.URL)
		})

		userID := seedLocalUser(t, db, "socket-concurrency@example.test", "Socket Concurrency", "SocketConcurrencyPass1!", false)
		sessionCookie, _ := loginLocalUser(t, server, "socket-concurrency@example.test", "SocketConcurrencyPass1!", nil)
		firstSessionID := querySessionRow(t, db, userID).sessionID
		if sessionCount := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1`, userID); sessionCount != 1 {
			t.Fatalf("expected one session row after initial login, got %d; sessions=%s", sessionCount, formatUserSessions(t, db, userID))
		}
		socket := connectSessionSocket(t, server, sessionCookie.Value)
		defer socket.Close(websocket.StatusNormalClosure, "integration_cleanup")

		for i := 0; i < 4; i++ {
			phase1test.SetClockOffset(t, server.HTTP.URL, int64(i+1))
			_, _ = loginLocalUser(t, server, "socket-concurrency@example.test", "SocketConcurrencyPass1!", nil)
		}

		phase1test.SetClockOffset(t, server.HTTP.URL, 5)
		activeSessionCookie, _ := loginLocalUser(t, server, "socket-concurrency@example.test", "SocketConcurrencyPass1!", nil)
		if activeCount := queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id = $1 AND revoked_at IS NULL`, userID); activeCount != 5 {
			t.Fatalf("expected five active sessions after sixth login, got %d; sessions=%s", activeCount, formatUserSessions(t, db, userID))
		}
		if err := phase1test.AwaitSessionRevoked(socket, authn.ConcurrencyLimitReasonCode); err != nil {
			firstSession := querySessionByID(t, db, firstSessionID)
			t.Fatalf(
				"await session_revoked for first session %s: %v (revoked_at_valid=%t revoke_reason_code=%q sessions=%s)",
				firstSessionID,
				err,
				firstSession.revokedAt.Valid,
				firstSession.revokeReasonCode.String,
				formatUserSessions(t, db, userID),
			)
		}

		firstSession := querySessionByID(t, db, firstSessionID)
		if !firstSession.revokedAt.Valid {
			t.Fatal("expected least-recently-used attached session to be revoked")
		}
		if firstSession.revokeReasonCode.String != authn.ConcurrencyLimitReasonCode {
			t.Fatalf("unexpected concurrency revoke_reason_code: got %q", firstSession.revokeReasonCode.String)
		}

		revokedSession := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(sessionCookie))
		httptestx.RequireErrorEnvelope(t, revokedSession, http.StatusUnauthorized, "session_required")

		activeSession := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(activeSessionCookie))
		httptestx.RequireSuccessEnvelope(t, activeSession, http.StatusOK)
	})
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

	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-boundary-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)

	for _, path := range []string{
		"/api/v1/auth/session",
		"/api/v1/auth/credential-state",
		"/api/v1/incidents",
		"/api/v1/test/auth/touch",
	} {
		resp := doJSON(t, http.MethodGet, server.HTTP.URL+path, nil, withHeader("Authorization", "Bearer "+bootstrapToken))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "credential_bootstrap_rejected")
		details := body["error"].(map[string]any)["details"].(map[string]any)
		if got := details["reason_code"]; got != "not_allowed_for_route" {
			t.Fatalf("unexpected bootstrap rejection reason for %s: got %v want not_allowed_for_route", path, got)
		}
	}

	phase1test.RequireBootstrapWebsocketRejected(t, server.HTTP.URL, bootstrapToken)

	completeInitialEnrollment(
		t,
		server,
		bootstrapToken,
		begin["enrollment_id"].(string),
		secretBase32,
		"txn-bootstrap-boundary-complete",
	)

	consumed := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{
		"client_txn_id": "txn-bootstrap-boundary-after-complete",
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	consumedBody := httptestx.RequireErrorEnvelope(t, consumed, http.StatusConflict, "credential_bootstrap_rejected")
	if got := consumedBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"]; got != "consumed" {
		t.Fatalf("unexpected consumed bootstrap rejection reason: got %v want consumed", got)
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
		defer targetSocket.Close(websocket.StatusNormalClosure, "integration_cleanup")

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
		defer targetSocket.Close(websocket.StatusNormalClosure, "integration_cleanup")

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
		defer targetSocket.Close(websocket.StatusNormalClosure, "integration_cleanup")

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

func TestPhase1_CredentialStateTransitions_I_1_04(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-04-state-transitions")
	defer db.Close()

	userID := seedLocalUser(t, db, "state-transitions@example.test", "State Transitions", "StateTransitions1!", false)
	sessionCookie, csrfCookie := loginLocalUser(t, server, "state-transitions@example.test", "StateTransitions1!", nil)

	initialStateResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/credential-state", nil, withCookies(sessionCookie))
	initialState := httptestx.RequireSuccessEnvelope(t, initialStateResp, http.StatusOK)["data"].(map[string]any)
	if got := initialState["user_id"]; got != userID {
		t.Fatalf("unexpected initial credential-state user_id: got %v want %s", got, userID)
	}
	if got := initialState["totp"].(map[string]any)["state"]; got != "not_enrolled" {
		t.Fatalf("unexpected initial totp.state: got %v want not_enrolled", got)
	}
	httptestx.RequireSecretSafePayload(t, initialState, []string{"password_hash", "bootstrap_token", "secret_base32", "otpauth_uri"})

	beginResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/mfa/totp/begin",
		map[string]any{
			"client_txn_id": "txn-stateful-begin",
		},
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	beginData := httptestx.RequireSuccessEnvelope(t, beginResp, http.StatusOK)["data"].(map[string]any)
	enrollmentID := beginData["enrollment_id"].(string)
	secretBase32 := beginData["totp_setup"].(map[string]any)["secret_base32"].(string)

	pendingStateResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/credential-state", nil, withCookies(sessionCookie))
	pendingState := httptestx.RequireSuccessEnvelope(t, pendingStateResp, http.StatusOK)["data"].(map[string]any)
	if got := pendingState["totp"].(map[string]any)["state"]; got != "pending" {
		t.Fatalf("unexpected pending totp.state: got %v want pending", got)
	}

	beginReplay := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/mfa/totp/begin",
		map[string]any{
			"client_txn_id": "txn-stateful-begin",
		},
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	replayed := httptestx.RequireSuccessEnvelope(t, beginReplay, http.StatusOK)["data"].(map[string]any)
	if replayed["enrollment_id"] != enrollmentID {
		t.Fatalf("expected begin replay enrollment_id %s, got %v", enrollmentID, replayed["enrollment_id"])
	}
	if got := replayed["totp_setup"].(map[string]any)["secret_base32"]; got != secretBase32 {
		t.Fatalf("expected begin replay secret_base32 %s, got %v", secretBase32, got)
	}

	divergentBegin := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/mfa/totp/begin",
		map[string]any{
			"client_txn_id": "txn-stateful-begin-divergent",
		},
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	divergentBody := httptestx.RequireErrorEnvelope(t, divergentBegin, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, divergentBegin.StatusCode, divergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

	completeResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/mfa/totp/complete",
		map[string]any{
			"client_txn_id": "txn-stateful-complete",
			"enrollment_id": enrollmentID,
			"code":          generateTOTPCode(t, secretBase32),
		},
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	completeData := httptestx.RequireSuccessEnvelope(t, completeResp, http.StatusOK)["data"].(map[string]any)
	if completeData["sessions_revoked"] != false {
		t.Fatalf("first session-scoped enrollment must not revoke sessions, got %v", completeData["sessions_revoked"])
	}

	activeStateResp := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/credential-state", nil, withCookies(sessionCookie))
	activeState := httptestx.RequireSuccessEnvelope(t, activeStateResp, http.StatusOK)["data"].(map[string]any)
	if got := activeState["totp"].(map[string]any)["state"]; got != "active" {
		t.Fatalf("unexpected active totp.state: got %v want active", got)
	}
	httptestx.RequireSecretSafePayload(t, activeState, []string{"password_hash", "bootstrap_token", "secret_base32", "otpauth_uri"})

	invalidCurrent := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/password/change",
		map[string]any{
			"client_txn_id":    "txn-stateful-password-invalid-current",
			"current_password": "WrongCurrentPass1!",
			"new_password":     "StateTransitionsChanged1!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": generateTOTPCode(t, secretBase32),
				},
			},
		},
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, invalidCurrent, http.StatusConflict, "invalid_current_password")

	missingFactor := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/password/change",
		map[string]any{
			"client_txn_id":    "txn-stateful-password-missing-factor",
			"current_password": "StateTransitions1!",
			"new_password":     "StateTransitionsChanged1!",
		},
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, missingFactor, http.StatusUnauthorized, "invalid_second_factor")

	changeResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/password/change",
		map[string]any{
			"client_txn_id":    "txn-stateful-password-change",
			"current_password": "StateTransitions1!",
			"new_password":     "StateTransitionsChanged1!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": generateTOTPCode(t, secretBase32),
				},
			},
		},
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	changeData := httptestx.RequireSuccessEnvelope(t, changeResp, http.StatusOK)["data"].(map[string]any)
	if changeData["sessions_revoked"] != true {
		t.Fatalf("password change must revoke all sessions, got %v", changeData["sessions_revoked"])
	}

	postChangeSession := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(sessionCookie))
	httptestx.RequireErrorEnvelope(t, postChangeSession, http.StatusUnauthorized, "session_required")

	oldPassword := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": "state-transitions@example.test",
		"password": "StateTransitions1!",
	})
	httptestx.RequireErrorEnvelope(t, oldPassword, http.StatusUnauthorized, "invalid_credentials")

	loginLocalUser(t, server, "state-transitions@example.test", "StateTransitionsChanged1!", nil)
}

func TestPhase1_BootstrapEnrollmentConsumption_I_1_04(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-04-bootstrap-consumption")
	defer db.Close()

	seedLocalUser(t, db, "bootstrap-consumption@example.test", "Bootstrap Consumption", "BootstrapConsumption1!", true)
	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-consumption@example.test", "BootstrapConsumption1!")

	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-consumption-begin",
	})
	enrollmentID := begin["enrollment_id"].(string)
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)

	beginReplay := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-bootstrap-consumption-begin",
	})
	if beginReplay["enrollment_id"] != enrollmentID {
		t.Fatalf("expected bootstrap begin replay enrollment_id %s, got %v", enrollmentID, beginReplay["enrollment_id"])
	}

	completeResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": "txn-bootstrap-consumption-complete",
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	completeData := httptestx.RequireSuccessEnvelope(t, completeResp, http.StatusOK)["data"].(map[string]any)
	if completeData["sessions_revoked"] != false {
		t.Fatalf("bootstrap completion must not revoke sessions, got %v", completeData["sessions_revoked"])
	}
	for _, cookie := range completeResp.Cookies() {
		if cookie.Name == authn.SessionCookieName {
			t.Fatal("bootstrap completion must not issue a session cookie")
		}
	}

	consumed := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{
		"client_txn_id": "txn-bootstrap-consumption-after-complete",
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	consumedBody := httptestx.RequireErrorEnvelope(t, consumed, http.StatusConflict, "credential_bootstrap_rejected")
	if got := consumedBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"]; got != "consumed" {
		t.Fatalf("unexpected consumed bootstrap reason_code: got %v want consumed", got)
	}

	loginWithoutFactor := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": "bootstrap-consumption@example.test",
		"password": "BootstrapConsumption1!",
	})
	httptestx.RequireErrorEnvelope(t, loginWithoutFactor, http.StatusUnauthorized, "mfa_required")

	_ = loginLocalUserWithSecondFactor(t, server, "bootstrap-consumption@example.test", "BootstrapConsumption1!", generateTOTPCode(t, secretBase32))
}

func TestPhase1_UserAdminAudit_I_1_03(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-03-audit")
	defer db.Close()

	adminID := seedLocalUserFlags(t, db, "audit-admin@example.test", "Audit Admin", "AuditAdminPass1!", false, true, true)
	adminSession, adminCSRF := loginLocalUser(t, server, "audit-admin@example.test", "AuditAdminPass1!", nil)

	createResp := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users", map[string]any{
		"client_txn_id":    "txn-user-audit-create",
		"auth_kind":        "local",
		"email":            "audit-target@example.test",
		"display_name":     "Audit Target",
		"initial_password": "AuditTargetPass1!",
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	if _, ok := createData["initial_password"]; ok {
		t.Fatal("user create response must not echo initial_password")
	}
	targetUserID := createData["user_id"].(string)

	patchResp := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+targetUserID, map[string]any{
		"base_user_version": 1,
		"display_name":      "Audit Target Patched",
		"mfa_required":      false,
	}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	events := lookupUserAuditEvents(t, db, targetUserID)
	if len(events) < 2 {
		t.Fatalf("expected at least create and patch audit events, got %d", len(events))
	}

	createEvent := requireAuditEventBySource(t, events, "users.create")
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: createEvent.ActorUserID,
		Source:      createEvent.EventSource,
		ClientTxnID: createEvent.ClientTxnID,
		RequestID:   createEvent.RequestID,
		CreatedAt:   createEvent.CreatedAt,
	}, adminID, "users.create", "txn-user-audit-create")
	httptestx.RequireSecretSafePayload(t, createEvent.Before, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32"})
	httptestx.RequireSecretSafePayload(t, createEvent.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32"})
	if got := createEvent.After["user_id"]; got != targetUserID {
		t.Fatalf("unexpected users.create audit after_json user_id: got %v want %s", got, targetUserID)
	}

	patchEvent := requireAuditEventBySource(t, events, "users.patch")
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: patchEvent.ActorUserID,
		Source:      patchEvent.EventSource,
		ClientTxnID: patchEvent.ClientTxnID,
		RequestID:   patchEvent.RequestID,
		CreatedAt:   patchEvent.CreatedAt,
	}, adminID, "users.patch", "")
	httptestx.RequireSecretSafePayload(t, patchEvent.Before, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32"})
	httptestx.RequireSecretSafePayload(t, patchEvent.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32"})
	if got := patchEvent.Before["user_version"]; got != float64(1) {
		t.Fatalf("unexpected users.patch before_json: %#v", patchEvent.Before)
	}
	if got := patchEvent.After["user_version"]; got != float64(2) {
		t.Fatalf("unexpected users.patch after_json: %#v", patchEvent.After)
	}
}

func TestPhase1_AdminCredentialAuditAndScope_I_1_05(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("password reset audit is deployment-local and incident admins are denied", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-audit-scope")
		defer db.Close()

		adminID := seedLocalUserFlags(t, db, "scope-admin@example.test", "Scope Admin", "ScopeAdminPass1!", false, true, true)
		adminSession, adminCSRF := loginLocalUser(t, server, "scope-admin@example.test", "ScopeAdminPass1!", nil)

		targetID := seedLocalUserWithActiveTOTP(t, db, "scope-target@example.test", "Scope Target", "ScopeTargetPass1!", true, false, "JBSWY3DPEHPK3QAC")
		incidentAdminID := seedLocalUserFlags(t, db, "incident-admin@example.test", "Incident Admin", "IncidentAdminPass1!", false, false, true)
		incidentAdminSession, incidentAdminCSRF := loginLocalUser(t, server, "incident-admin@example.test", "IncidentAdminPass1!", nil)

		incident := createIncidentResource(t, server, adminSession, adminCSRF, map[string]any{
			"client_txn_id": "txn-phase1-scope-incident",
			"incident_key":  "IR-PHASE1-SCOPE",
			"title":         "Phase 1 Scope",
		})
		incidentID := incident["incident_id"].(string)
		createIncidentMembership(t, server, incidentID, adminSession, adminCSRF, map[string]any{
			"client_txn_id": "txn-phase1-scope-membership",
			"email":         "incident-admin@example.test",
			"role":          "admin",
		})

		for _, path := range []string{
			"/api/v1/users/" + targetID + "/password/reset",
			"/api/v1/users/" + targetID + "/mfa/totp/reset",
			"/api/v1/users/" + targetID + "/sessions/revoke-all",
		} {
			var body map[string]any
			switch path {
			case "/api/v1/users/" + targetID + "/password/reset":
				body = map[string]any{
					"base_user_version": 1,
					"client_txn_id":     "txn-incident-admin-password-reset",
					"new_password":      "ScopeTargetChanged1!",
					"reason":            "incident admin denied",
				}
			case "/api/v1/users/" + targetID + "/mfa/totp/reset":
				body = map[string]any{
					"base_user_version": 1,
					"client_txn_id":     "txn-incident-admin-totp-reset",
					"reason":            "incident admin denied",
				}
			default:
				body = map[string]any{
					"client_txn_id": "txn-incident-admin-revoke-all",
					"reason":        "incident admin denied",
				}
			}
			resp := doJSON(
				t,
				http.MethodPost,
				server.HTTP.URL+path,
				body,
				withCookies(incidentAdminSession, incidentAdminCSRF),
				withHeader(authn.CSRFHeaderName, incidentAdminCSRF.Value),
			)
			httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "session_required")
		}

		passwordReset := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/password/reset", map[string]any{
			"base_user_version": 1,
			"client_txn_id":     "txn-phase1-admin-audit-password-reset",
			"new_password":      "ScopeTargetChanged1!",
			"reason":            "deployment admin reset",
		}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
		httptestx.RequireSuccessEnvelope(t, passwordReset, http.StatusOK)

		events := lookupUserAuditEvents(t, db, targetID)
		event := requireAuditEventBySource(t, events, "users.password.reset")
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: event.ActorUserID,
			Source:      event.EventSource,
			ClientTxnID: event.ClientTxnID,
			RequestID:   event.RequestID,
			CreatedAt:   event.CreatedAt,
		}, adminID, "users.password.reset", "txn-phase1-admin-audit-password-reset")
		httptestx.RequireSecretSafePayload(t, event.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32"})
		if got := event.After["user_version"]; got != float64(2) {
			t.Fatalf("unexpected users.password.reset after_json: %#v", event.After)
		}
		if got := queryCount(t, db, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2 AND role = 'admin'`, incidentID, incidentAdminID); got != 1 {
			t.Fatalf("expected incident admin membership to remain incident-scoped, got %d", got)
		}
	})

	t.Run("totp reset and revoke-all write safe deployment-admin audit records", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-audit-events")
		defer db.Close()

		adminID := seedLocalUserFlags(t, db, "audit-events-admin@example.test", "Audit Events Admin", "AuditEventsAdmin1!", false, true, true)
		adminSession, adminCSRF := loginLocalUser(t, server, "audit-events-admin@example.test", "AuditEventsAdmin1!", nil)

		totpTargetID := seedLocalUserWithActiveTOTP(t, db, "audit-events-totp@example.test", "Audit Events TOTP", "AuditEventsTotp1!", true, false, "JBSWY3DPEHPK3QAD")
		revokeTargetID := seedLocalUserWithActiveTOTP(t, db, "audit-events-revoke@example.test", "Audit Events Revoke", "AuditEventsRevoke1!", true, false, "JBSWY3DPEHPK3QAE")

		totpReset := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+totpTargetID+"/mfa/totp/reset", map[string]any{
			"base_user_version": 1,
			"client_txn_id":     "txn-phase1-admin-audit-totp-reset",
			"reason":            "deployment admin totp reset",
		}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
		httptestx.RequireSuccessEnvelope(t, totpReset, http.StatusOK)

		revokeAll := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+revokeTargetID+"/sessions/revoke-all", map[string]any{
			"client_txn_id": "txn-phase1-admin-audit-revoke-all",
			"reason":        "deployment admin revoke all",
		}, withCookies(adminSession, adminCSRF), withHeader(authn.CSRFHeaderName, adminCSRF.Value))
		httptestx.RequireSuccessEnvelope(t, revokeAll, http.StatusOK)

		totpEvents := lookupUserAuditEvents(t, db, totpTargetID)
		totpEvent := requireAuditEventBySource(t, totpEvents, "users.totp.reset")
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: totpEvent.ActorUserID,
			Source:      totpEvent.EventSource,
			ClientTxnID: totpEvent.ClientTxnID,
			RequestID:   totpEvent.RequestID,
			CreatedAt:   totpEvent.CreatedAt,
		}, adminID, "users.totp.reset", "")
		httptestx.RequireSecretSafePayload(t, totpEvent.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32"})
		if got := totpEvent.After["user_version"]; got != float64(2) {
			t.Fatalf("unexpected users.totp.reset after_json: %#v", totpEvent.After)
		}

		revokeEvents := lookupUserAuditEvents(t, db, revokeTargetID)
		revokeEvent := requireAuditEventBySource(t, revokeEvents, "users.sessions.revoke_all")
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: revokeEvent.ActorUserID,
			Source:      revokeEvent.EventSource,
			ClientTxnID: revokeEvent.ClientTxnID,
			RequestID:   revokeEvent.RequestID,
			CreatedAt:   revokeEvent.CreatedAt,
		}, adminID, "users.sessions.revoke_all", "")
		httptestx.RequireSecretSafePayload(t, revokeEvent.After, []string{"password_hash", "initial_password", "bootstrap_token", "secret_base32"})
		if _, ok := revokeEvent.After["revoked_at"].(string); !ok {
			t.Fatalf("unexpected users.sessions.revoke_all after_json: %#v", revokeEvent.After)
		}
	})
}

func TestPhase1_UserCreateReplayReturnsOriginalCommittedResource_I_1_03(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-03-create-replay")
	defer db.Close()

	adminID := seedLocalUserFlags(t, db, "create-replay-admin@example.test", "Create Replay Admin", "CreateReplayAdmin1!", false, true, true)
	adminSession, adminCSRF := loginLocalUser(t, server, "create-replay-admin@example.test", "CreateReplayAdmin1!", nil)

	createRequest := map[string]any{
		"client_txn_id":    "txn-user-create-replay",
		"auth_kind":        "local",
		"email":            "create-replay-target@example.test",
		"display_name":     "Create Replay Target",
		"initial_password": "CreateReplayTarget1!",
		"mfa_required":     false,
	}
	createResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users",
		createRequest,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	targetUserID := createData["user_id"].(string)

	requireArgon2PasswordHash(t, queryUserPasswordHash(t, db, targetUserID), "CreateReplayTarget1!", "WrongCreateReplay1!")

	idempotency := lookupRouteIdempotency(t, db, "users.create", adminID, "txn-user-create-replay")
	if idempotency.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected users.create idempotency status_code: got %d want %d", idempotency.StatusCode, http.StatusCreated)
	}
	requireJSONEquivalent(t, idempotency.Response, createData)
	httptestx.RequireSecretSafePayload(t, idempotency.Response, forbiddenSecretKeys())

	replayResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users",
		createRequest,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	requireJSONEquivalent(t, replayData, createData)

	divergentResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users",
		map[string]any{
			"client_txn_id":    "txn-user-create-replay",
			"auth_kind":        "local",
			"email":            "create-replay-target@example.test",
			"display_name":     "Create Replay Target Divergent",
			"initial_password": "CreateReplayTarget1!",
			"mfa_required":     false,
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	divergentBody := httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, divergentResp.StatusCode, divergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

	patchResp := doJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/users/"+targetUserID,
		map[string]any{
			"base_user_version": 1,
			"display_name":      "Create Replay Mutated",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	replayAfterPatchResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users",
		createRequest,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	replayAfterPatchData := httptestx.RequireSuccessEnvelope(t, replayAfterPatchResp, http.StatusOK)["data"].(map[string]any)
	requireJSONEquivalent(t, replayAfterPatchData, createData)

	if got := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND scope_key = $2 AND client_txn_id = $3`, "users.create", adminID, "txn-user-create-replay"); got != 1 {
		t.Fatalf("expected one users.create route_idempotency row, got %d", got)
	}
	if got := queryCount(t, db, `SELECT COUNT(*) FROM users WHERE email = $1`, "create-replay-target@example.test"); got != 1 {
		t.Fatalf("expected one created user row, got %d", got)
	}
}

func TestPhase1_PasswordChangeReplayAndStoredPayload_I_1_04(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-04-password-change-replay")
	defer db.Close()

	userID := seedLocalUserWithActiveTOTP(t, db, "password-replay@example.test", "Password Replay", "PasswordReplay1!", true, false, "JBSWY3DPEHPK3QBA")
	initialLogin := loginLocalUserWithSecondFactor(t, server, "password-replay@example.test", "PasswordReplay1!", generateTOTPCode(t, "JBSWY3DPEHPK3QBA"))
	sessionCookie := initialLogin.sessionCookie
	csrfCookie := initialLogin.csrfCookie

	changeRequest := map[string]any{
		"client_txn_id":    "txn-password-change-replay",
		"current_password": "PasswordReplay1!",
		"new_password":     "PasswordReplayChanged1!",
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": generateTOTPCode(t, "JBSWY3DPEHPK3QBA"),
			},
		},
	}
	changeResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/password/change",
		changeRequest,
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	changeData := httptestx.RequireSuccessEnvelope(t, changeResp, http.StatusOK)["data"].(map[string]any)

	requireArgon2PasswordHash(t, queryUserPasswordHash(t, db, userID), "PasswordReplayChanged1!", "PasswordReplay1!")

	idempotency := lookupRouteIdempotency(t, db, "auth.password.change", userID, "txn-password-change-replay")
	if idempotency.StatusCode != http.StatusOK {
		t.Fatalf("unexpected auth.password.change idempotency status_code: got %d want %d", idempotency.StatusCode, http.StatusOK)
	}
	requireJSONEquivalent(t, idempotency.Response, changeData)
	httptestx.RequireSecretSafePayload(t, idempotency.Response, forbiddenSecretKeys())

	replayLogin := loginLocalUserWithSecondFactor(t, server, "password-replay@example.test", "PasswordReplayChanged1!", generateTOTPCode(t, "JBSWY3DPEHPK3QBA"))
	replayResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/password/change",
		changeRequest,
		withCookies(replayLogin.sessionCookie, replayLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, replayLogin.csrfCookie.Value),
	)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	requireJSONEquivalent(t, replayData, changeData)

	divergentLogin := loginLocalUserWithSecondFactor(t, server, "password-replay@example.test", "PasswordReplayChanged1!", generateTOTPCode(t, "JBSWY3DPEHPK3QBA"))
	divergentResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/auth/password/change",
		map[string]any{
			"client_txn_id":    "txn-password-change-replay",
			"current_password": "PasswordReplay1!",
			"new_password":     "PasswordReplayChanged2!",
			"second_factor": map[string]any{
				"kind": "totp",
				"assertion": map[string]any{
					"code": generateTOTPCode(t, "JBSWY3DPEHPK3QBA"),
				},
			},
		},
		withCookies(divergentLogin.sessionCookie, divergentLogin.csrfCookie),
		withHeader(authn.CSRFHeaderName, divergentLogin.csrfCookie.Value),
	)
	divergentBody := httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, divergentResp.StatusCode, divergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

	if got := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND scope_key = $2 AND client_txn_id = $3`, "auth.password.change", userID, "txn-password-change-replay"); got != 1 {
		t.Fatalf("expected one auth.password.change route_idempotency row, got %d", got)
	}
}

func TestPhase1_AdminPasswordResetReplayReturnsOriginalCommittedResource_I_1_05(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-password-reset-replay")
	defer db.Close()

	adminID := seedLocalUserFlags(t, db, "admin-password-replay@example.test", "Admin Password Replay", "AdminPasswordReplay1!", false, true, true)
	adminSession, adminCSRF := loginLocalUser(t, server, "admin-password-replay@example.test", "AdminPasswordReplay1!", nil)
	targetUserID := seedLocalUser(t, db, "target-password-replay@example.test", "Target Password Replay", "TargetPasswordReplay1!", false)
	scopeKey := adminID + ":" + targetUserID

	resetRequest := map[string]any{
		"base_user_version": 1,
		"client_txn_id":     "txn-admin-password-replay",
		"new_password":      "TargetPasswordReplayChanged1!",
		"reason":            "admin reset replay",
	}
	resetResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users/"+targetUserID+"/password/reset",
		resetRequest,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	resetData := httptestx.RequireSuccessEnvelope(t, resetResp, http.StatusOK)["data"].(map[string]any)

	requireArgon2PasswordHash(t, queryUserPasswordHash(t, db, targetUserID), "TargetPasswordReplayChanged1!", "TargetPasswordReplay1!")

	idempotency := lookupRouteIdempotency(t, db, "users.password.reset", scopeKey, "txn-admin-password-replay")
	if idempotency.StatusCode != http.StatusOK {
		t.Fatalf("unexpected users.password.reset idempotency status_code: got %d want %d", idempotency.StatusCode, http.StatusOK)
	}
	requireJSONEquivalent(t, idempotency.Response, resetData)
	httptestx.RequireSecretSafePayload(t, idempotency.Response, forbiddenSecretKeys())

	replayResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users/"+targetUserID+"/password/reset",
		resetRequest,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	requireJSONEquivalent(t, replayData, resetData)

	divergentResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users/"+targetUserID+"/password/reset",
		map[string]any{
			"base_user_version": 1,
			"client_txn_id":     "txn-admin-password-replay",
			"new_password":      "TargetPasswordReplayChanged2!",
			"reason":            "admin reset replay divergent",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	divergentBody := httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, divergentResp.StatusCode, divergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

	patchResp := doJSON(
		t,
		http.MethodPatch,
		server.HTTP.URL+"/api/v1/users/"+targetUserID,
		map[string]any{
			"base_user_version": 2,
			"display_name":      "Target Password Replay Mutated",
		},
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	replayAfterPatchResp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/users/"+targetUserID+"/password/reset",
		resetRequest,
		withCookies(adminSession, adminCSRF),
		withHeader(authn.CSRFHeaderName, adminCSRF.Value),
	)
	replayAfterPatchData := httptestx.RequireSuccessEnvelope(t, replayAfterPatchResp, http.StatusOK)["data"].(map[string]any)
	requireJSONEquivalent(t, replayAfterPatchData, resetData)

	if got := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND scope_key = $2 AND client_txn_id = $3`, "users.password.reset", scopeKey, "txn-admin-password-replay"); got != 1 {
		t.Fatalf("expected one users.password.reset route_idempotency row, got %d", got)
	}
}

func TestPhase1_AdminTOTPResetAndRevokeAllReplay_I_1_05(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	s3Harness := s3test.Start(t)

	t.Run("totp reset replays the original response and rejects divergent reuse", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-totp-reset-replay")
		defer db.Close()

		adminID := seedLocalUserFlags(t, db, "admin-totp-replay@example.test", "Admin TOTP Replay", "AdminTotpReplay1!", false, true, true)
		adminSession, adminCSRF := loginLocalUser(t, server, "admin-totp-replay@example.test", "AdminTotpReplay1!", nil)
		targetUserID := seedLocalUserWithActiveTOTP(t, db, "target-totp-replay@example.test", "Target TOTP Replay", "TargetTotpReplay1!", true, false, "JBSWY3DPEHPK3QBB")
		scopeKey := adminID + ":" + targetUserID

		resetRequest := map[string]any{
			"base_user_version": 1,
			"client_txn_id":     "txn-admin-totp-replay",
			"reason":            "admin totp replay",
		}
		resetResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/users/"+targetUserID+"/mfa/totp/reset",
			resetRequest,
			withCookies(adminSession, adminCSRF),
			withHeader(authn.CSRFHeaderName, adminCSRF.Value),
		)
		resetData := httptestx.RequireSuccessEnvelope(t, resetResp, http.StatusOK)["data"].(map[string]any)

		idempotency := lookupRouteIdempotency(t, db, "users.totp.reset", scopeKey, "txn-admin-totp-replay")
		if idempotency.StatusCode != http.StatusOK {
			t.Fatalf("unexpected users.totp.reset idempotency status_code: got %d want %d", idempotency.StatusCode, http.StatusOK)
		}
		requireJSONEquivalent(t, idempotency.Response, resetData)
		httptestx.RequireSecretSafePayload(t, idempotency.Response, forbiddenSecretKeys())

		replayResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/users/"+targetUserID+"/mfa/totp/reset",
			resetRequest,
			withCookies(adminSession, adminCSRF),
			withHeader(authn.CSRFHeaderName, adminCSRF.Value),
		)
		replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
		requireJSONEquivalent(t, replayData, resetData)

		divergentResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/users/"+targetUserID+"/mfa/totp/reset",
			map[string]any{
				"base_user_version": 1,
				"client_txn_id":     "txn-admin-totp-replay",
				"reason":            "admin totp replay divergent",
			},
			withCookies(adminSession, adminCSRF),
			withHeader(authn.CSRFHeaderName, adminCSRF.Value),
		)
		divergentBody := httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
		httptestx.RequireDivergentReplayRejected(t, divergentResp.StatusCode, divergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

		if got := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND scope_key = $2 AND client_txn_id = $3`, "users.totp.reset", scopeKey, "txn-admin-totp-replay"); got != 1 {
			t.Fatalf("expected one users.totp.reset route_idempotency row, got %d", got)
		}
	})

	t.Run("revoke-all replays the original response and rejects divergent reuse", func(t *testing.T) {
		server, db := startPhase1Server(t, postgresHarness, s3Harness, "phase1-i-1-05-revoke-all-replay")
		defer db.Close()

		adminID := seedLocalUserFlags(t, db, "admin-revoke-replay@example.test", "Admin Revoke Replay", "AdminRevokeReplay1!", false, true, true)
		adminSession, adminCSRF := loginLocalUser(t, server, "admin-revoke-replay@example.test", "AdminRevokeReplay1!", nil)
		targetUserID := seedLocalUser(t, db, "target-revoke-replay@example.test", "Target Revoke Replay", "TargetRevokeReplay1!", false)
		targetSession, _ := loginLocalUser(t, server, "target-revoke-replay@example.test", "TargetRevokeReplay1!", nil)
		if targetSession == nil {
			t.Fatal("expected target session cookie before revoke-all")
		}
		scopeKey := adminID + ":" + targetUserID

		revokeRequest := map[string]any{
			"client_txn_id": "txn-admin-revoke-replay",
			"reason":        "admin revoke replay",
		}
		revokeResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/users/"+targetUserID+"/sessions/revoke-all",
			revokeRequest,
			withCookies(adminSession, adminCSRF),
			withHeader(authn.CSRFHeaderName, adminCSRF.Value),
		)
		revokeData := httptestx.RequireSuccessEnvelope(t, revokeResp, http.StatusOK)["data"].(map[string]any)

		idempotency := lookupRouteIdempotency(t, db, "users.sessions.revoke_all", scopeKey, "txn-admin-revoke-replay")
		if idempotency.StatusCode != http.StatusOK {
			t.Fatalf("unexpected users.sessions.revoke_all idempotency status_code: got %d want %d", idempotency.StatusCode, http.StatusOK)
		}
		requireJSONEquivalent(t, idempotency.Response, revokeData)
		httptestx.RequireSecretSafePayload(t, idempotency.Response, forbiddenSecretKeys())

		replayResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/users/"+targetUserID+"/sessions/revoke-all",
			revokeRequest,
			withCookies(adminSession, adminCSRF),
			withHeader(authn.CSRFHeaderName, adminCSRF.Value),
		)
		replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
		requireJSONEquivalent(t, replayData, revokeData)

		divergentResp := doJSON(
			t,
			http.MethodPost,
			server.HTTP.URL+"/api/v1/users/"+targetUserID+"/sessions/revoke-all",
			map[string]any{
				"client_txn_id": "txn-admin-revoke-replay",
				"reason":        "admin revoke replay divergent",
			},
			withCookies(adminSession, adminCSRF),
			withHeader(authn.CSRFHeaderName, adminCSRF.Value),
		)
		divergentBody := httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")
		httptestx.RequireDivergentReplayRejected(t, divergentResp.StatusCode, divergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

		if got := queryCount(t, db, `SELECT COUNT(*) FROM route_idempotency WHERE route_key = $1 AND scope_key = $2 AND client_txn_id = $3`, "users.sessions.revoke_all", scopeKey, "txn-admin-revoke-replay"); got != 1 {
			t.Fatalf("expected one users.sessions.revoke_all route_idempotency row, got %d", got)
		}
	})
}

type sessionRow struct {
	authenticatedAt          time.Time
	sessionID                string
	lastQualifyingActivityAt time.Time
	idleExpiresAt            time.Time
	absoluteExpiresAt        time.Time
	sessionExpiresAt         time.Time
	revokedAt                sql.NullTime
	revokeReasonCode         sql.NullString
}

func startPhase1Server(t testing.TB, postgresHarness *pgtest.Harness, s3Harness *s3test.Harness, prefix string) (*httptestx.Server, *sql.DB) {
	t.Helper()

	testDB := postgresHarness.PrepareDatabaseT(t, prefix)

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

	authCookies := httptestx.RequireAuthCookies(t, resp.Cookies())
	return authCookies.Session, authCookies.CSRF
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
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	authCookies := httptestx.RequireAuthCookies(t, resp.Cookies())
	return loginResult{sessionCookie: authCookies.Session, csrfCookie: authCookies.CSRF}
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

func connectSessionSocket(t testing.TB, server *httptestx.Server, sessionToken string) *phase1test.SessionSocketClient {
	t.Helper()
	return phase1test.ConnectSessionSocket(t, server.HTTP.URL, sessionToken)
}

func expectSessionRevoked(t testing.TB, conn *phase1test.SessionSocketClient, wantReasonCode string) {
	t.Helper()
	phase1test.ExpectSessionRevoked(t, conn, wantReasonCode)
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

	row := querySingleSession(t, db, `SELECT authenticated_at, id::text, last_qualifying_activity_at, idle_expires_at, absolute_expires_at, session_expires_at, revoked_at, revoke_reason_code FROM user_sessions WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`, userID)
	return row
}

func querySessionByID(t testing.TB, db *sql.DB, sessionID string) sessionRow {
	t.Helper()
	return querySingleSession(t, db, `SELECT authenticated_at, id::text, last_qualifying_activity_at, idle_expires_at, absolute_expires_at, session_expires_at, revoked_at, revoke_reason_code FROM user_sessions WHERE id::text = $1`, sessionID)
}

func querySingleSession(t testing.TB, db *sql.DB, query string, args ...any) sessionRow {
	t.Helper()

	var row sessionRow
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(
		&row.authenticatedAt,
		&row.sessionID,
		&row.lastQualifyingActivityAt,
		&row.idleExpiresAt,
		&row.absoluteExpiresAt,
		&row.sessionExpiresAt,
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

func formatUserSessions(t testing.TB, db *sql.DB, userID string) string {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT authenticated_at,
       id::text,
       last_qualifying_activity_at,
       revoked_at,
       revoke_reason_code
  FROM user_sessions
 WHERE user_id = $1
 ORDER BY authenticated_at ASC, id ASC
`, userID)
	if err != nil {
		t.Fatalf("query user sessions: %v", err)
	}
	defer rows.Close()

	summary := make([]string, 0, 6)
	for rows.Next() {
		var (
			authenticatedAt time.Time
			sessionID       string
			lastActivityAt  time.Time
			revokedAt       sql.NullTime
			reasonCode      sql.NullString
		)
		if err := rows.Scan(&authenticatedAt, &sessionID, &lastActivityAt, &revokedAt, &reasonCode); err != nil {
			t.Fatalf("scan user session: %v", err)
		}
		summary = append(summary, fmt.Sprintf(
			"{id=%s auth=%s last=%s revoked=%t reason=%q}",
			sessionID,
			authenticatedAt.UTC().Format(time.RFC3339Nano),
			lastActivityAt.UTC().Format(time.RFC3339Nano),
			revokedAt.Valid,
			reasonCode.String,
		))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate user sessions: %v", err)
	}
	return strings.Join(summary, ", ")
}

type auditEventRecord struct {
	EventKind    string
	ActorUserID  string
	TargetUserID string
	EventSource  string
	ReasonCode   string
	ClientTxnID  string
	RequestID    string
	CreatedAt    time.Time
	Before       map[string]any
	After        map[string]any
}

func lookupUserAuditEvents(t testing.TB, db *sql.DB, targetUserID string) []auditEventRecord {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT event_kind,
       actor_user_id::text,
       target_user_id::text,
       event_source,
       reason_code,
       COALESCE(client_txn_id, ''),
       COALESCE(request_id, ''),
       created_at,
       before_json,
       after_json
  FROM deployment_admin_audit_events
 WHERE target_user_id::text = $1
 ORDER BY created_at ASC
`, targetUserID)
	if err != nil {
		t.Fatalf("query user audit events: %v", err)
	}
	defer rows.Close()

	events := make([]auditEventRecord, 0, 4)
	for rows.Next() {
		var (
			record     auditEventRecord
			beforeJSON []byte
			afterJSON  []byte
		)
		if err := rows.Scan(
			&record.EventKind,
			&record.ActorUserID,
			&record.TargetUserID,
			&record.EventSource,
			&record.ReasonCode,
			&record.ClientTxnID,
			&record.RequestID,
			&record.CreatedAt,
			&beforeJSON,
			&afterJSON,
		); err != nil {
			t.Fatalf("scan user audit event: %v", err)
		}
		record.Before = decodeJSONMap(t, beforeJSON)
		record.After = decodeJSONMap(t, afterJSON)
		events = append(events, record)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate user audit events: %v", err)
	}
	return events
}

func decodeJSONMap(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	if len(payload) == 0 {
		return map[string]any{}
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode json map: %v", err)
	}
	return decoded
}

func requireAuditEventBySource(t testing.TB, events []auditEventRecord, source string) auditEventRecord {
	t.Helper()
	for _, event := range events {
		if event.EventSource == source {
			return event
		}
	}
	t.Fatalf("expected audit event source %q in %#v", source, events)
	return auditEventRecord{}
}

func createIncidentResource(t testing.TB, server *httptestx.Server, sessionCookie *http.Cookie, csrfCookie *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func createIncidentMembership(t testing.TB, server *httptestx.Server, incidentID string, sessionCookie *http.Cookie, csrfCookie *http.Cookie, body map[string]any) map[string]any {
	t.Helper()

	resp := doJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		body,
		withCookies(sessionCookie, csrfCookie),
		withHeader(authn.CSRFHeaderName, csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

type routeIdempotencyRecord struct {
	StatusCode int
	Response   map[string]any
}

func lookupRouteIdempotency(t testing.TB, db *sql.DB, routeKey string, scopeKey string, clientTxnID string) routeIdempotencyRecord {
	t.Helper()

	var (
		record       routeIdempotencyRecord
		responseJSON []byte
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT status_code, response_json
  FROM route_idempotency
 WHERE route_key = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, routeKey, scopeKey, clientTxnID).Scan(&record.StatusCode, &responseJSON); err != nil {
		t.Fatalf("lookup route idempotency: %v", err)
	}
	record.Response = decodeJSONMap(t, responseJSON)
	return record
}

func queryUserPasswordHash(t testing.TB, db *sql.DB, userID string) string {
	t.Helper()

	var passwordHash string
	if err := db.QueryRowContext(context.Background(), `
SELECT password_hash
  FROM users
 WHERE id::text = $1
`, userID).Scan(&passwordHash); err != nil {
		t.Fatalf("query user password hash: %v", err)
	}
	return passwordHash
}

func requireArgon2PasswordHash(t testing.TB, passwordHash string, acceptedPassword string, rejectedPassword string) {
	t.Helper()

	if !strings.HasPrefix(passwordHash, "argon2id$v=19$m=65536,t=1,p=4$") {
		t.Fatalf("expected argon2id password hash, got %q", passwordHash)
	}
	accepted, err := authn.VerifyPasswordHash(passwordHash, acceptedPassword)
	if err != nil {
		t.Fatalf("verify accepted password hash: %v", err)
	}
	if !accepted {
		t.Fatalf("expected password hash to accept %q", acceptedPassword)
	}
	rejected, err := authn.VerifyPasswordHash(passwordHash, rejectedPassword)
	if err != nil {
		t.Fatalf("verify rejected password hash: %v", err)
	}
	if rejected {
		t.Fatalf("expected password hash to reject %q", rejectedPassword)
	}
}

func requireJSONEquivalent(t testing.TB, got any, want any) {
	t.Helper()

	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got json: %v", err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want json: %v", err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("unexpected json payload:\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
}

func forbiddenSecretKeys() []string {
	return []string{
		"password_hash",
		"initial_password",
		"current_password",
		"new_password",
		"bootstrap_token",
		"secret_base32",
		"otpauth_uri",
	}
}
