package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDeploymentAdminAuthorizationRefresh_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	// Each operation needs a subject and a separate administrator to change its
	// capability. The returned closure performs only that real user mutation.
	start := func(t *testing.T, name string) (*httptestx.Server, *sql.DB, string, authcookietest.AuthCookies, func()) {
		t.Helper()
		server, db := startServer(t, runtime, "auth-current-admin-"+name)
		t.Cleanup(func() { _ = db.Close() })
		subjectID := seedLocalUserFlags(t, db, "subject@example.test", "Subject", "SubjectPassword1!", false, true, true)
		login := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "subject@example.test", "SubjectPassword1!", nil)
		seedLocalUserFlags(t, db, "peer@example.test", "Peer", "PeerPassword1!", false, true, true)
		peerLogin := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "peer@example.test", "PeerPassword1!", nil)
		return server, db, subjectID, login, func() {
			current := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+subjectID, nil, httptestx.WithCookies(peerLogin.Session))
			resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
			response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+subjectID,
				map[string]any{"base_user_version": resource["user_version"], "is_deployment_admin": false},
				httptestx.WithCookies(peerLogin.Session, peerLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, peerLogin.CSRF.Value))
			demoted := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			if demoted["is_deployment_admin"] != false {
				t.Fatal("peer did not demote the subject")
			}
		}
	}

	t.Run("deployment audit read", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "audit")
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/administrative-audit-events", nil, httptestx.WithCookies(subject.Session))
		httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
		demote()
		before := observeUserAdministration(t, db, subjectID, subjectID)
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/administrative-audit-events", nil, httptestx.WithCookies(subject.Session))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, subjectID, subjectID))
	})
	t.Run("users list", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "list")
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users", nil, httptestx.WithCookies(subject.Session))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		findUserResource(t, data["users"].([]any), subjectID)
		demote()
		before := observeUserAdministration(t, db, subjectID, subjectID)
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users", nil, httptestx.WithCookies(subject.Session))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, subjectID, subjectID))
	})
	t.Run("user get", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "get")
		targetID := seedLocalUser(t, db, "get-target@example.test", "Get Target", "GetTargetPassword1!", false)
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, httptestx.WithCookies(subject.Session))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if data["user_id"] != targetID {
			t.Fatal("get returned another user")
		}
		demote()
		before := observeUserAdministration(t, db, targetID, subjectID)
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, httptestx.WithCookies(subject.Session))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, targetID, subjectID))
	})
	t.Run("user create", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "create")
		body := map[string]any{"client_txn_id": "allowed-create", "auth_kind": "local", "email": "created@example.test", "display_name": "Created User", "initial_password": "CreatedPassword1!"}
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusCreated)["data"].(map[string]any)
		requireCurrentPasswordHash(t, queryUserPasswordHash(t, db, data["user_id"].(string)), "CreatedPassword1!", "WrongPassword1!")
		demote()
		body["client_txn_id"], body["email"] = "denied-create", "denied-create@example.test"
		before := observeUserAdministration(t, db, "", subjectID)
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE email=$1`, body["email"]) != 0 {
			t.Fatal("denied-create identity already exists")
		}
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, "", subjectID))
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE email=$1`, body["email"]) != 0 {
			t.Fatal("denied create persisted a user")
		}
	})
	t.Run("user patch", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "patch")
		targetID := seedLocalUser(t, db, "patch-target@example.test", "Patch Target", "PatchTargetPassword1!", false)
		current := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, httptestx.WithCookies(subject.Session))
		resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body := map[string]any{"base_user_version": resource["user_version"], "display_name": "Allowed Change"}
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+targetID, body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE id::text=$1 AND display_name=$2 AND user_version=$3`, targetID, "Allowed Change", data["user_version"]) != 1 {
			t.Fatal("allowed patch did not persist its returned version and value")
		}
		demote()
		body["base_user_version"], body["display_name"] = data["user_version"], "Denied Change"
		before := observeUserAdministration(t, db, targetID, subjectID)
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+targetID, body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, targetID, subjectID))
	})
	t.Run("password reset", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "password")
		targetID, _ := seedAdministrativeCredentialTarget(t, server, db, "password-target@example.test")
		current := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, httptestx.WithCookies(subject.Session))
		resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body := map[string]any{"base_user_version": resource["user_version"], "client_txn_id": "allowed-password", "new_password": "AllowedPassword1!", "reason": "allowed reset"}
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/password/reset", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		requireCurrentPasswordHash(t, queryUserPasswordHash(t, db, targetID), "AllowedPassword1!", "AdministrativeTarget1!")
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE id::text=$1 AND user_version=$2`, targetID, data["user_version"]) != 1 {
			t.Fatal("password reset did not persist its returned version")
		}
		deniedID, _ := seedAdministrativeCredentialTarget(t, server, db, "denied-password@example.test")
		current = httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+deniedID, nil, httptestx.WithCookies(subject.Session))
		resource = httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body["base_user_version"], body["client_txn_id"] = resource["user_version"], "denied-password"
		demote()
		before := observeUserAdministration(t, db, deniedID, subjectID)
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+deniedID+"/password/reset", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, deniedID, subjectID))
	})
	t.Run("totp reset", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "totp")
		targetID, _ := seedAdministrativeCredentialTarget(t, server, db, "totp-target@example.test")
		current := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+targetID, nil, httptestx.WithCookies(subject.Session))
		resource := httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body := map[string]any{"base_user_version": resource["user_version"], "client_txn_id": "allowed-totp", "reason": "allowed reset"}
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/mfa/totp/reset", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if queryCount(t, db, `SELECT COUNT(*) FROM users WHERE id::text=$1 AND totp_secret_envelope IS NULL AND totp_enrolled_at IS NULL AND user_version=$2`, targetID, data["user_version"]) != 1 {
			t.Fatal("allowed TOTP reset did not clear the factor at its returned version")
		}
		deniedID, _ := seedAdministrativeCredentialTarget(t, server, db, "denied-totp@example.test")
		current = httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/users/"+deniedID, nil, httptestx.WithCookies(subject.Session))
		resource = httptestx.RequireSuccessEnvelope(t, current, http.StatusOK)["data"].(map[string]any)
		body["base_user_version"], body["client_txn_id"] = resource["user_version"], "denied-totp"
		demote()
		before := observeUserAdministration(t, db, deniedID, subjectID)
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+deniedID+"/mfa/totp/reset", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, deniedID, subjectID))
	})
	t.Run("revoke all", func(t *testing.T) {
		server, db, subjectID, subject, demote := start(t, "revoke")
		targetID, target := seedAdministrativeCredentialTarget(t, server, db, "revoke-target@example.test")
		beforeSuccess := observeUserAdministration(t, db, targetID, subjectID)
		body := map[string]any{"client_txn_id": "allowed-revoke", "reason": "allowed revoke"}
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+targetID+"/sessions/revoke-all", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		if data["sessions_revoked"] != true || queryCount(t, db, `SELECT COUNT(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, targetID) != 0 {
			t.Fatal("allowed revoke-all did not revoke target sessions")
		}
		afterSuccess := observeUserAdministration(t, db, targetID, subjectID)
		if beforeSuccess.user != afterSuccess.user || beforeSuccess.enrollments != afterSuccess.enrollments || beforeSuccess.bootstrap != afterSuccess.bootstrap {
			t.Fatal("revoke-all changed credential or enrollment state")
		}
		revoked := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(target.Session))
		httptestx.RequireErrorEnvelope(t, revoked, http.StatusUnauthorized, "session_required")
		deniedID, _ := seedAdministrativeCredentialTarget(t, server, db, "denied-revoke@example.test")
		body["client_txn_id"] = "denied-revoke"
		demote()
		before := observeUserAdministration(t, db, deniedID, subjectID)
		denied := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/users/"+deniedID+"/sessions/revoke-all", body, httptestx.WithCookies(subject.Session, subject.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, subject.CSRF.Value))
		httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
		before.requireUnchanged(t, observeUserAdministration(t, db, deniedID, subjectID))
	})
}

// Credential actions need a meaningful active factor, session and pending
// replacement to expose unintended revocation, activation or reset effects.
func seedAdministrativeCredentialTarget(t testing.TB, server *httptestx.Server, db *sql.DB, email string) (string, authcookietest.AuthCookies) {
	t.Helper()
	const secret = "ZILKBF5BZGY567F262KQMFGNYPCBREN3OKX65SN4EMQ7ZDSSMBOA"
	userID := seedLocalUserWithActiveTOTP(t, db, email, "Administrative Target", "AdministrativeTarget1!", true, false, secret)
	login := flowtest.LoginLocalUserWithSecondFactor(t, server.HTTP.Client(), server.HTTP.URL, email, "AdministrativeTarget1!", flowtest.GenerateTOTPCode(t, secret))
	response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{
		"client_txn_id": "pending-replacement", "current_password": "AdministrativeTarget1!",
		"second_factor": map[string]any{"kind": "totp", "assertion": map[string]any{"code": flowtest.GenerateTOTPCode(t, secret)}},
	}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
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
