package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func assertEnterpriseCompletionAdmission(t *testing.T, runtime *appsupport.Runtime, protocol string) {
	t.Helper()
	for _, scenario := range []string{"binding retired", "user disabled", "session write rejected"} {
		t.Run(protocol+" completion after "+scenario, func(t *testing.T) {
			server, db := startEnterpriseServer(t, runtime, "auth-enterprise-admission")
			defer db.Close()
			userID := seedLocalUser(t, db, "completion@example.test", "Completion", "CompletionPassword1!", false)
			wantSessions, oldestID := 0, ""
			if scenario == "session write rejected" {
				wantSessions = 5
				for index := 0; index < 5; index++ {
					flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "completion@example.test", "CompletionPassword1!", nil)
					if index == 0 {
						oldestID = querySessionRow(t, db, userID).sessionID
					}
				}
			}
			var priorLogin sql.NullTime
			if err := db.QueryRowContext(t.Context(), `SELECT last_login_at FROM users WHERE id::text=$1`, userID).Scan(&priorLogin); err != nil {
				t.Fatal(err)
			}
			providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{Key: "completion", Type: protocol, Name: "Completion", Enabled: true, Interactive: true, Issuer: "https://idp.example.test", Audience: "cartulary-sp"})
			bindingID := seedEnterpriseBinding(t, db, userID, providerID, "completion-subject", userID)
			begin := beginEnterpriseAuth(t, server, "completion")
			redirect := parseEnterpriseRedirect(t, begin.redirectURL)
			callback := oidcCallbackURL(server, "completion", redirect.Query().Get("state"), redirect.Query().Get("nonce"), "valid-code", "completion-subject")
			if protocol == "saml" {
				assertion := fakeSAMLResponse(t, map[string]any{"subject": "completion-subject", "issuer": "https://idp.example.test", "audience": "cartulary-sp", "signature_valid": true, "expires_at": time.Now().UTC().Add(time.Hour).Format(time.RFC3339)})
				staged := doFormNoRedirect(t, server.HTTP.URL+"/api/v1/auth/saml/completion/acs", url.Values{"RelayState": {redirect.Query().Get("RelayState")}, "SAMLResponse": {assertion}})
				if staged.StatusCode != http.StatusSeeOther {
					t.Fatalf("SAML staging status = %d", staged.StatusCode)
				}
				requireNoSessionCookie(t, staged)
				callback = server.HTTP.URL + staged.Header.Get("Location")
			}
			var response *http.Response
			if scenario == "session write rejected" {
				if _, err := db.ExecContext(t.Context(), `ALTER TABLE user_sessions ADD CONSTRAINT test_reject_enterprise_session CHECK (provider_type='local')`); err != nil {
					t.Fatal(err)
				}
				response = doNoRedirect(t, http.MethodGet, callback, nil, httptestx.WithCookies(begin.cookie))
				httptestx.RequireErrorEnvelope(t, response, http.StatusInternalServerError, "internal_error")
			} else {
				request := httptestx.NewJSONRequest(t, http.MethodGet, callback, nil)
				request.AddCookie(begin.cookie)
				response = overlapAuthRequests(t, db, false, []*http.Request{request}, func(ctx context.Context, barrier *sql.Tx) error {
					if scenario == "binding retired" {
						_, err := barrier.ExecContext(ctx, `UPDATE enterprise_auth_bindings SET retired_at=now() WHERE id::text=$1`, bindingID)
						return err
					}
					_, err := barrier.ExecContext(ctx, `UPDATE users SET is_active=false WHERE id::text=$1`, userID)
					return err
				})[0]
				reason := "no_linked_user"
				if scenario == "user disabled" {
					reason = "inactive_user"
				}
				requireErrorReason(t, response, http.StatusConflict, "provider_identity_rejected", reason)
			}
			requireNoSessionCookie(t, response)
			for _, check := range []struct {
				query string
				args  []any
			}{
				{`SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND provider_type=$2`, []any{userID, protocol}},
				{`SELECT count(*) FROM enterprise_auth_transactions WHERE provider_id::text=$1 AND consumed_at IS NOT NULL`, []any{providerID}},
				{`SELECT count(*) FROM enterprise_auth_bindings WHERE id::text=$1 AND last_auth_at IS NOT NULL`, []any{bindingID}},
				{`SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NOT NULL`, []any{userID}},
				{`SELECT count(*) FROM deployment_admin_audit_events WHERE target_user_id::text=$1`, []any{userID}},
			} {
				if queryCount(t, db, check.query, check.args...) != 0 {
					t.Fatal("rejected enterprise completion committed an auth effect")
				}
			}
			if got := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, userID); got != wantSessions {
				t.Fatalf("failed completion left %d sessions, want %d", got, wantSessions)
			}
			var afterLogin sql.NullTime
			if err := db.QueryRowContext(t.Context(), `SELECT last_login_at FROM users WHERE id::text=$1`, userID).Scan(&afterLogin); err != nil {
				t.Fatal(err)
			}
			if afterLogin.Valid != priorLogin.Valid || !afterLogin.Time.Equal(priorLogin.Time) {
				t.Fatal("failed completion changed last login")
			}
			if scenario == "session write rejected" {
				if _, err := db.ExecContext(t.Context(), `ALTER TABLE user_sessions DROP CONSTRAINT test_reject_enterprise_session`); err != nil {
					t.Fatal(err)
				}
				if protocol == "oidc" {
					// A real provider may already have redeemed the authorization
					// code. Recovery starts a fresh flow; it does not assume code reuse.
					begin = beginEnterpriseAuth(t, server, "completion")
					redirect = parseEnterpriseRedirect(t, begin.redirectURL)
					callback = oidcCallbackURL(server, "completion", redirect.Query().Get("state"), redirect.Query().Get("nonce"), "valid-code", "completion-subject")
				}
				retry := doNoRedirect(t, http.MethodGet, callback, nil, httptestx.WithCookies(begin.cookie))
				if retry.StatusCode != http.StatusSeeOther {
					t.Fatalf("unconsumed completion retry status = %d", retry.StatusCode)
				}
				if queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND auth_binding_id::text=$2`, userID, bindingID) != 1 {
					t.Fatal("retry did not create exactly one attributed session")
				}
				cookies := authcookietest.RequireAuthCookies(t, retry.Cookies())
				inspection := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(cookies.Session))
				data := httptestx.RequireSuccessEnvelope(t, inspection, http.StatusOK)["data"].(map[string]any)
				if data["user_id"] != userID || data["provider_type"] != protocol {
					t.Fatal("retry authenticated the wrong identity")
				}
				if queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, userID) != 5 || queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE actor_user_id::text=$1 AND target_user_id::text=$1 AND reason_code='concurrency_limit'`, userID) != 1 {
					t.Fatal("enterprise retry did not enforce the session cap with one attributed revocation")
				}
				victim := querySessionByID(t, db, oldestID)
				if !victim.revokedAt.Valid || victim.revokeReasonCode.String != "concurrency_limit" {
					t.Fatal("enterprise retry revoked the wrong predecessor")
				}
			}
		})
	}
}
