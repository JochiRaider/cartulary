package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func assertTOTPActivationAdmission(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	for _, tc := range []struct {
		name, mode, query, code, reason string
		status                          int
		concurrent, mutationTime        bool
	}{
		{"bootstrap superseded", "bootstrap", `UPDATE bootstrap_tokens SET superseded_at=$2 WHERE user_id::text=$1`, "credential_bootstrap_rejected", "superseded", 409, false, true},
		{"bootstrap consumed", "bootstrap", `UPDATE bootstrap_tokens SET consumed_at=$2 WHERE user_id::text=$1`, "credential_bootstrap_rejected", "consumed", 409, false, true},
		{"bootstrap expired", "bootstrap", `UPDATE bootstrap_tokens SET expires_at=$2 WHERE user_id::text=$1`, "credential_bootstrap_rejected", "expired", 409, false, true},
		{"bootstrap account disabled", "bootstrap", `UPDATE users SET is_active=false WHERE id::text=$1`, "session_required", "", 401, false, false},
		{"session revoked", "session", `UPDATE user_sessions SET revoked_at=$2 WHERE user_id::text=$1`, "session_required", "", 401, false, true},
		{"session expired", "session", `UPDATE user_sessions SET session_expires_at=$2 WHERE user_id::text=$1`, "session_required", "", 401, false, true},
		{"session account disabled", "session", `UPDATE users SET is_active=false WHERE id::text=$1`, "session_required", "", 401, false, false},
		{"reset clears pending setup", "session", `DELETE FROM pending_totp_enrollments WHERE user_id::text=$1`, "totp_setup_not_pending", "not_found", 409, false, false},
		{"bootstrap concurrent completion", "bootstrap", "", "credential_bootstrap_rejected", "consumed", 409, true, false},
		{"session concurrent completion", "session", "", "totp_setup_not_pending", "consumed", 409, true, false},
	} {
		t.Run("activation admission "+tc.name, func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-totp-activation-admission")
			defer db.Close()
			now := time.Date(2026, time.April, 29, 20, 0, 0, 0, time.UTC)
			httptestx.SetClockFixed(t, server, now)
			userID := seedLocalUser(t, db, "activate@example.test", "Activate", "ActivationPassword1!", tc.mode == "bootstrap")
			var token string
			wantSessions := 0
			if tc.mode == "bootstrap" {
				token = flowtest.RequireBootstrapLogin(t, http.DefaultClient, server.HTTP.URL, "activate@example.test", "ActivationPassword1!")
			} else {
				token = flowtest.LoginLocalUser(t, http.DefaultClient, server.HTTP.URL, "activate@example.test", "ActivationPassword1!", nil).Session.Value
				wantSessions = 1
			}
			pending := flowtest.BeginTOTPEnrollment(t, http.DefaultClient, server.HTTP.URL, token, map[string]any{"client_txn_id": "activation-begin"})
			secret := pending["totp_setup"].(map[string]any)["secret_base32"].(string)
			code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA256})
			if err != nil {
				t.Fatal(err)
			}
			count := 1
			if tc.concurrent {
				count = 2
			}
			requests := make([]*http.Request, count)
			for index := range requests {
				requests[index] = httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
					"client_txn_id": "activation-complete", "enrollment_id": pending["enrollment_id"], "code": code,
				})
				requests[index].Header.Set("Authorization", "Bearer "+token)
			}
			var mutate func(context.Context, *sql.Tx) error
			if tc.query != "" {
				mutate = func(ctx context.Context, barrier *sql.Tx) error {
					args := []any{userID}
					if tc.mutationTime {
						args = append(args, now)
					}
					_, err := barrier.ExecContext(ctx, tc.query, args...)
					return err
				}
			}
			responses := overlapAuthRequests(t, db, false, requests, mutate)
			successes := 0
			for _, response := range responses {
				if response.StatusCode == http.StatusOK && tc.concurrent {
					data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
					if data["user_id"] != userID || data["sessions_revoked"] != false {
						t.Fatal("first enrollment returned incorrect user or revocation")
					}
					successes++
				} else {
					errorBody := httptestx.RequireErrorEnvelope(t, response, tc.status, tc.code)["error"].(map[string]any)
					if tc.reason != "" && errorBody["details"].(map[string]any)["reason_code"] != tc.reason {
						t.Fatalf("activation rejection reason=%v want=%s", errorBody["details"], tc.reason)
					}
				}
				if len(response.Cookies()) != 0 {
					t.Fatal("first enrollment or rejected activation changed cookies")
				}
			}
			wantActivations := 0
			if tc.concurrent {
				wantActivations = 1
			}
			if successes != wantActivations {
				t.Fatalf("successful activations=%d want=%d", successes, wantActivations)
			}
			for _, query := range []string{
				`SELECT count(*) FROM users WHERE id::text=$1 AND totp_enrolled_at IS NOT NULL`,
				`SELECT count(*) FROM pending_totp_enrollments WHERE user_id::text=$1 AND consumed_at IS NOT NULL`,
				`SELECT count(*) FROM deployment_admin_audit_events WHERE target_user_id::text=$1 AND event_source='auth.totp.complete'`,
			} {
				if got := queryCount(t, db, query, userID); got != wantActivations {
					t.Fatalf("activation effect count=%d want=%d", got, wantActivations)
				}
			}
			if queryUserVersion(t, db, userID) != int64(1+wantActivations) || queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, userID) != wantSessions {
				t.Fatal("activation changed user version or session count incorrectly")
			}
		})
	}
}
