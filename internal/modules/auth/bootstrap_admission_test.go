package auth_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func assertBootstrapAdmission(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	now := time.Date(2026, time.April, 29, 19, 0, 0, 0, time.UTC)
	for _, initial := range []int{0, 1} {
		t.Run(fmt.Sprintf("bootstrap overlap with %d predecessor", initial), func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-bootstrap-overlap")
			defer db.Close()
			httptestx.SetClockFixed(t, server, now)
			userID := seedLocalUser(t, db, "setup@example.test", "Setup", "SetupPassword1!", true)
			if initial == 1 {
				flowtest.RequireBootstrapLogin(t, http.DefaultClient, server.HTTP.URL, "setup@example.test", "SetupPassword1!")
			}
			body := map[string]any{"username": "setup@example.test", "password": "SetupPassword1!"}
			responses := overlapAuthRequests(t, db, false, []*http.Request{
				httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", body),
				httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", body),
			}, nil)
			tokens := make(map[string]bool)
			for _, response := range responses {
				details := httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, "mfa_setup_required")["error"].(map[string]any)["details"].(map[string]any)
				token, ok := details["bootstrap_token"].(string)
				if !ok || token == "" || tokens[token] || len(response.Cookies()) != 0 {
					t.Fatal("bootstrap overlap returned duplicate authority or cookies")
				}
				expiresText, _ := details["bootstrap_expires_at"].(string)
				expires, err := time.Parse(time.RFC3339Nano, expiresText)
				if err != nil || !expires.Equal(now.Add(10*time.Minute)) {
					t.Fatalf("bootstrap expiry=%q, want exactly ten minutes after issuance", expiresText)
				}
				tokens[token] = true
			}
			for _, check := range []struct {
				query string
				want  int
			}{
				{`SELECT count(*) FROM bootstrap_tokens WHERE user_id::text=$1`, initial + 2},
				{`SELECT count(*) FROM bootstrap_tokens WHERE user_id::text=$1 AND superseded_at IS NULL AND consumed_at IS NULL`, 1},
				{`SELECT count(*) FROM bootstrap_tokens WHERE user_id::text=$1 AND superseded_at IS NOT NULL`, initial + 1},
				{`SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, 0},
			} {
				if got := queryCount(t, db, check.query, userID); got != check.want {
					t.Errorf("bootstrap overlap durable count=%d want=%d", got, check.want)
				}
			}
		})
	}
	for _, tc := range []struct {
		name, query string
		allowed     bool
	}{
		{"password changed", `UPDATE users SET password_hash=$2 WHERE id::text=$1`, false},
		{"email changed", `UPDATE users SET email='changed@example.test' WHERE id::text=$1`, false},
		{"disabled", `UPDATE users SET is_active=false WHERE id::text=$1`, false},
		{"MFA no longer required", `UPDATE users SET mfa_required=false WHERE id::text=$1`, false},
		{"factor enrolled", `UPDATE users SET totp_enrolled_at='2026-04-29T19:00:00Z', totp_secret_envelope=$2 WHERE id::text=$1`, false},
		{"profile changed", `UPDATE users SET display_name='Updated', user_version=user_version+1 WHERE id::text=$1`, true},
	} {
		t.Run("bootstrap admission after "+tc.name, func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-bootstrap-authority")
			defer db.Close()
			httptestx.SetClockFixed(t, server, now)
			userID := seedLocalUser(t, db, "setup@example.test", "Setup", "SetupPassword1!", true)
			flowtest.RequireBootstrapLogin(t, http.DefaultClient, server.HTTP.URL, "setup@example.test", "SetupPassword1!")
			args := []any{userID}
			switch tc.name {
			case "password changed":
				hash, err := authn.HashPassword(t.Context(), "ReplacementPassword1!")
				if err != nil {
					t.Fatal(err)
				}
				args = append(args, hash)
			case "factor enrolled":
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
			request := httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{"username": "setup@example.test", "password": "SetupPassword1!"})
			response := overlapAuthRequests(t, db, false, []*http.Request{request}, func(ctx context.Context, barrier *sql.Tx) error {
				_, err := barrier.ExecContext(ctx, tc.query, args...)
				return err
			})[0]
			wantTotal, wantSuperseded := 1, 0
			code := "invalid_credentials"
			if tc.allowed {
				wantTotal, wantSuperseded, code = 2, 1, "mfa_setup_required"
			}
			httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, code)
			if len(response.Cookies()) != 0 || queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, userID) != 0 {
				t.Fatal("bootstrap admission created session authority")
			}
			if queryCount(t, db, `SELECT count(*) FROM bootstrap_tokens WHERE user_id::text=$1`, userID) != wantTotal ||
				queryCount(t, db, `SELECT count(*) FROM bootstrap_tokens WHERE user_id::text=$1 AND superseded_at IS NOT NULL`, userID) != wantSuperseded {
				t.Fatal("bootstrap admission changed issuance or supersession incorrectly")
			}
			if queryCount(t, db, `SELECT count(*) FROM users WHERE id::text=$1 AND last_login_at IS NOT NULL`, userID) != 0 ||
				queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE target_user_id::text=$1`, userID) != 0 {
				t.Fatal("bootstrap admission changed login or audit state")
			}
		})
	}
}
