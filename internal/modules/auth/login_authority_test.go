package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func assertLocalLoginAuthorityAtAdmission(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	for _, tc := range []struct {
		name, query  string
		mfa, allowed bool
	}{
		{"password changed", `UPDATE users SET password_hash=$2 WHERE id::text=$1`, false, false},
		{"login email changed", `UPDATE users SET email='changed@example.test' WHERE id::text=$1`, false, false},
		{"account disabled", `UPDATE users SET is_active=false WHERE id::text=$1`, false, false},
		{"MFA newly required", `UPDATE users SET mfa_required=true WHERE id::text=$1`, false, false},
		{"TOTP reset", `UPDATE users SET totp_enrolled_at=NULL, totp_secret_envelope=NULL WHERE id::text=$1`, true, false},
		{"TOTP replaced", `UPDATE users SET totp_secret_envelope=$2 WHERE id::text=$1`, true, false},
		{"profile changed", `UPDATE users SET display_name='Updated', user_version=user_version+1 WHERE id::text=$1`, false, true},
	} {
		t.Run("login admission after "+tc.name, func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-login-authority")
			defer db.Close()
			now := time.Date(2026, time.April, 29, 18, 30, 0, 0, time.UTC)
			httptestx.SetClockFixed(t, server, now)
			secret := "ZILKBF5BZGY567F262KQMFGNYPCBREN3OKX65SN4EMQ7ZDSSMBOA"
			userID := seedLocalUserWithActiveTOTP(t, db, "admission@example.test", "Admission", "AdmissionPassword1!", tc.mfa, false, secret)
			body := map[string]any{"username": "admission@example.test", "password": "AdmissionPassword1!"}
			if tc.mfa {
				code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA256})
				if err != nil {
					t.Fatal(err)
				}
				body["second_factor"] = map[string]any{"kind": "totp", "assertion": map[string]any{"code": code}}
			}
			args := []any{userID}
			switch tc.name {
			case "password changed":
				hash, err := authn.HashPassword(t.Context(), "ReplacementPassword1!")
				if err != nil {
					t.Fatal(err)
				}
				args = append(args, hash)
			case "TOTP replaced":
				keys, err := authn.LoadMasterKeys(nil)
				if err != nil {
					t.Fatal(err)
				}
				id := uuid.MustParse(userID)
				envelope, err := authn.SealSecret(keys, authn.SecretBinding{Purpose: authn.ActiveTOTPSecret, RecordID: id, SubjectID: id}, []byte("12345678901234567890123456789012"))
				if err != nil {
					t.Fatal(err)
				}
				args = append(args, envelope)
			}
			request := httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", body)
			responses := overlapAuthRequests(t, db, false, []*http.Request{request}, func(ctx context.Context, barrier *sql.Tx) error {
				_, err := barrier.ExecContext(ctx, tc.query, args...)
				return err
			})
			response := responses[0]
			wantSessions := 0
			if tc.allowed {
				if httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)["user_id"] != userID {
					t.Fatal("admitted another user")
				}
				authcookietest.RequireAuthCookies(t, response.Cookies())
				wantSessions = 1
			} else {
				httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, "invalid_credentials")
				if len(response.Cookies()) != 0 {
					t.Fatal("stale login emitted cookies")
				}
			}
			if got := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, userID); got != wantSessions {
				t.Fatalf("session count = %d, want %d", got, wantSessions)
			}
			if queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE target_user_id::text=$1`, userID) != 0 {
				t.Fatal("login admission emitted an unexpected audit effect")
			}
			if queryCount(t, db, `SELECT count(*) FROM bootstrap_tokens WHERE user_id::text=$1`, userID) != 0 {
				t.Fatal("stale admission created bootstrap authority")
			}
			if !tc.allowed && queryCount(t, db, `SELECT count(*) FROM users WHERE id::text=$1 AND last_login_at IS NOT NULL`, userID) != 0 {
				t.Fatal("rejected admission changed last login")
			}
		})
	}
}
