package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDeploymentAdminAuthorizationRefresh_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	// Each operation needs a subject and a separate administrator to change its
	// capability. The returned closure performs only that real user mutation.
	start := func(t *testing.T, name string) (*httptestx.Server, *sql.DB, string, loginResult, func()) {
		t.Helper()
		server, db := startServer(t, runtime, "auth-current-admin-"+name)
		t.Cleanup(func() { _ = db.Close() })
		subjectID := seedLocalUserFlags(t, db, "subject@example.test", "Subject", "SubjectPassword1!", false, true, true)
		session, csrf := loginLocalUser(t, server, "subject@example.test", "SubjectPassword1!", nil)
		seedLocalUserFlags(t, db, "peer@example.test", "Peer", "PeerPassword1!", false, true, true)
		peerSession, peerCSRF := loginLocalUser(t, server, "peer@example.test", "PeerPassword1!", nil)
		return server, db, subjectID, loginResult{sessionCookie: session, csrfCookie: csrf}, func() {
			current := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+subjectID, nil, withCookies(peerSession))
			resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
			response := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+subjectID,
				map[string]any{"base_user_version": resource["user_version"], "is_deployment_admin": false},
				withCookies(peerSession, peerCSRF), withHeader(authn.CSRFHeaderName, peerCSRF.Value))
			demoted := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			if demoted["is_deployment_admin"] != false {
				t.Fatal("peer did not demote the subject")
			}
		}
	}

	t.Run("deployment audit read", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "audit")
		response := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/administrative-audit-events", nil, withCookies(subject.sessionCookie))
		httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
		demote()
		before := observeUserAdministration(t, db, subjectID, subjectID)
		denied := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/administrative-audit-events", nil, withCookies(subject.sessionCookie))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, subjectID, subjectID))
	})
	t.Run("users list", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "list")
		response := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users", nil, withCookies(subject.sessionCookie))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		findUserResource(t, data["users"].([]any), subjectID)
		demote()
		before := observeUserAdministration(t, db, subjectID, subjectID)
		denied := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users", nil, withCookies(subject.sessionCookie))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, subjectID, subjectID))
	})
	t.Run("user get", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "get")
		targetID := seedLocalUser(t, db, "get-target@example.test", "Get Target", "GetTargetPassword1!", false)
		response := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, withCookies(subject.sessionCookie))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if data["user_id"] != targetID {
			t.Fatal("get returned another user")
		}
		demote()
		before := observeUserAdministration(t, db, targetID, subjectID)
		denied := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, withCookies(subject.sessionCookie))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, targetID, subjectID))
	})
	t.Run("user create", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "create")
		body := map[string]any{"client_txn_id": "allowed-create", "auth_kind": "local", "email": "created@example.test", "display_name": "Created User", "initial_password": "CreatedPassword1!"}
		response := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusCreated)["data"].(map[string]any)
		requireCurrentPasswordHash(t, queryUserPasswordHash(t, db, data["user_id"].(string)), "CreatedPassword1!", "WrongPassword1!")
		demote()
		body["client_txn_id"], body["email"] = "denied-create", "denied-create@example.test"
		before := observeUserAdministration(t, db, "", subjectID)
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE email=$1`, body["email"]) != 0 {
			t.Fatal("denied-create identity already exists")
		}
		denied := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, "", subjectID))
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE email=$1`, body["email"]) != 0 {
			t.Fatal("denied create persisted a user")
		}
	})
	t.Run("user patch", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "patch")
		targetID := seedLocalUser(t, db, "patch-target@example.test", "Patch Target", "PatchTargetPassword1!", false)
		current := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, withCookies(subject.sessionCookie))
		resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body := map[string]any{"base_user_version": resource["user_version"], "display_name": "Allowed Change"}
		response := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+targetID, body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE id::text=$1 AND display_name=$2 AND user_version=$3`, targetID, "Allowed Change", data["user_version"]) != 1 {
			t.Fatal("allowed patch did not persist its returned version and value")
		}
		demote()
		body["base_user_version"], body["display_name"] = data["user_version"], "Denied Change"
		before := observeUserAdministration(t, db, targetID, subjectID)
		denied := doJSON(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+targetID, body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, targetID, subjectID))
	})
	t.Run("password reset", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "password")
		targetID, _ := seedAdministrativeCredentialTarget(t, server, db, "password-target@example.test")
		current := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, withCookies(subject.sessionCookie))
		resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body := map[string]any{"base_user_version": resource["user_version"], "client_txn_id": "allowed-password", "new_password": "AllowedPassword1!", "reason": "allowed reset"}
		response := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/password/reset", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		requireCurrentPasswordHash(t, queryUserPasswordHash(t, db, targetID), "AllowedPassword1!", "AdministrativeTarget1!")
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE id::text=$1 AND user_version=$2`, targetID, data["user_version"]) != 1 {
			t.Fatal("password reset did not persist its returned version")
		}
		deniedID, _ := seedAdministrativeCredentialTarget(t, server, db, "denied-password@example.test")
		current = doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+deniedID, nil, withCookies(subject.sessionCookie))
		resource = httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body["base_user_version"], body["client_txn_id"] = resource["user_version"], "denied-password"
		demote()
		before := observeUserAdministration(t, db, deniedID, subjectID)
		denied := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+deniedID+"/password/reset", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, deniedID, subjectID))
	})
	t.Run("totp reset", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "totp")
		targetID, _ := seedAdministrativeCredentialTarget(t, server, db, "totp-target@example.test")
		current := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, withCookies(subject.sessionCookie))
		resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body := map[string]any{"base_user_version": resource["user_version"], "client_txn_id": "allowed-totp", "reason": "allowed reset"}
		response := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/mfa/totp/reset", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE id::text=$1 AND totp_secret_envelope IS NULL AND totp_enrolled_at IS NULL AND user_version=$2`, targetID, data["user_version"]) != 1 {
			t.Fatal("allowed TOTP reset did not clear the factor at its returned version")
		}
		deniedID, _ := seedAdministrativeCredentialTarget(t, server, db, "denied-totp@example.test")
		current = doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+deniedID, nil, withCookies(subject.sessionCookie))
		resource = httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body["base_user_version"], body["client_txn_id"] = resource["user_version"], "denied-totp"
		demote()
		before := observeUserAdministration(t, db, deniedID, subjectID)
		denied := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+deniedID+"/mfa/totp/reset", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, deniedID, subjectID))
	})
	t.Run("revoke all", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "revoke")
		targetID, target := seedAdministrativeCredentialTarget(t, server, db, "revoke-target@example.test")
		beforeSuccess := observeUserAdministration(t, db, targetID, subjectID)
		body := map[string]any{"client_txn_id": "allowed-revoke", "reason": "allowed revoke"}
		response := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/sessions/revoke-all", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if data["sessions_revoked"] != true || queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, targetID) != 0 {
			t.Fatal("allowed revoke-all did not revoke target sessions")
		}
		afterSuccess := observeUserAdministration(t, db, targetID, subjectID)
		if beforeSuccess.user != afterSuccess.user || beforeSuccess.enrollments != afterSuccess.enrollments || beforeSuccess.bootstrap != afterSuccess.bootstrap {
			t.Fatal("revoke-all changed credential or enrollment state")
		}
		revoked := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(target.sessionCookie))
		httptestx.RequireErrorEnvelope(t, revoked, http.StatusUnauthorized, "session_required")
		deniedID, _ := seedAdministrativeCredentialTarget(t, server, db, "denied-revoke@example.test")
		body["client_txn_id"] = "denied-revoke"
		demote()
		before := observeUserAdministration(t, db, deniedID, subjectID)
		denied := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+deniedID+"/sessions/revoke-all", body, withCookies(subject.sessionCookie, subject.csrfCookie), withHeader(authn.CSRFHeaderName, subject.csrfCookie.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, deniedID, subjectID))
	})
}

// Credential actions need a meaningful active factor, session and pending
// replacement to expose unintended revocation, activation or reset effects.
func seedAdministrativeCredentialTarget(t testing.TB, server *httptestx.Server, db *sql.DB, email string) (string, loginResult) {
	t.Helper()
	const secret = "ZILKBF5BZGY567F262KQMFGNYPCBREN3OKX65SN4EMQ7ZDSSMBOA"
	userID := seedLocalUserWithActiveTOTP(t, db, email, "Administrative Target", "AdministrativeTarget1!", true, false, secret)
	login := loginLocalUserWithSecondFactor(t, server, email, "AdministrativeTarget1!", generateTOTPCode(t, secret))
	response := doJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{
		"client_txn_id": "pending-replacement", "current_password": "AdministrativeTarget1!",
		"second_factor": map[string]any{"kind": "totp", "assertion": map[string]any{"code": generateTOTPCode(t, secret)}},
	}, withCookies(login.sessionCookie, login.csrfCookie), withHeader(authn.CSRFHeaderName, login.csrfCookie.Value))
	httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	return userID, login
}

type userAdministrationObservation struct {
	user, sessions, enrollments, bootstrap, receipts, audit, projections string
}

// These concrete observations cover one target and the acting user's effects.
// Session activity timestamps are deliberately absent. Values can contain
// private credential state, so failure diagnostics name the changed area only.
func observeUserAdministration(t testing.TB, db *sql.DB, targetID, actorID string) userAdministrationObservation {
	t.Helper()
	var observation userAdministrationObservation
	if err := db.QueryRowContext(context.Background(), `SELECT
 (SELECT coalesce(jsonb_agg(to_jsonb(u) ORDER BY id),'[]')::text FROM users u WHERE id::text=$1),
 (SELECT coalesce(jsonb_agg(to_jsonb(s) ORDER BY id),'[]')::text FROM
   (SELECT id,user_id,revoked_at,revoke_reason_code FROM user_sessions WHERE user_id::text=$1) s),
 (SELECT coalesce(jsonb_agg(to_jsonb(e) ORDER BY id),'[]')::text FROM pending_totp_enrollments e WHERE user_id::text=$1),
 (SELECT coalesce(jsonb_agg(to_jsonb(b) ORDER BY id),'[]')::text FROM bootstrap_tokens b WHERE user_id::text=$1),
 (SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY id),'[]')::text FROM route_idempotency r WHERE actor_user_id::text=$2 OR target_user_id::text=$1),
 (SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY id),'[]')::text FROM deployment_admin_audit_events a WHERE actor_user_id::text=$2 OR target_user_id::text=$1),
 (SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY audit_event_id),'[]')::text FROM administrative_audit_projections p WHERE actor_user_id::text=$2 OR (target_kind='user' AND target_id=$1))`, targetID, actorID).Scan(
		&observation.user, &observation.sessions, &observation.enrollments, &observation.bootstrap,
		&observation.receipts, &observation.audit, &observation.projections,
	); err != nil {
		t.Fatalf("observe user administration: %v", err)
	}
	return observation
}

func (before userAdministrationObservation) requireUnchanged(t testing.TB, after userAdministrationObservation) {
	t.Helper()
	for _, area := range []struct{ name, before, after string }{
		{"user and credentials", before.user, after.user},
		{"session revocations", before.sessions, after.sessions},
		{"pending enrollment", before.enrollments, after.enrollments},
		{"bootstrap tokens", before.bootstrap, after.bootstrap},
		{"receipts", before.receipts, after.receipts},
		{"audit events", before.audit, after.audit},
		{"audit projections", before.projections, after.projections},
	} {
		if area.before != area.after {
			t.Errorf("denied administrative action changed %s", area.name)
		}
	}
}
