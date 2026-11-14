package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func assertTOTPBeginAdmission(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	for _, mode := range []string{"bootstrap", "session"} {
		for _, divergent := range []bool{false, true} {
			name := mode + " matching begin overlap"
			if divergent {
				name = mode + " divergent begin overlap"
			}
			t.Run(name, func(t *testing.T) {
				server, db := startServer(t, runtime, "auth-totp-begin-overlap")
				defer db.Close()
				userID, token := prepareTOTPBeginActor(t, db, server, mode)
				keys := []string{"begin-key", "begin-key"}
				if divergent {
					keys[1] = "different-key"
				}
				requests := make([]*http.Request, 2)
				for index, key := range keys {
					requests[index] = httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{"client_txn_id": key})
					requests[index].Header.Set("Authorization", "Bearer "+token)
				}
				responses := overlapAuthRequests(t, db, false, requests, nil)
				var first map[string]any
				successes := 0
				for _, response := range responses {
					if response.StatusCode == http.StatusOK {
						data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
						if first != nil && !reflect.DeepEqual(first, data) {
							t.Fatal("matching begin overlap returned different enrollment or setup material")
						}
						first = data
						successes++
					} else {
						httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "client_txn_conflict")
					}
					if len(response.Cookies()) != 0 {
						t.Fatal("begin overlap changed cookies")
					}
				}
				wantSuccess := 2
				if divergent {
					wantSuccess = 1
				}
				if successes != wantSuccess || queryCount(t, db, `SELECT count(*) FROM pending_totp_enrollments WHERE user_id::text=$1`, userID) != 1 ||
					queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE target_user_id::text=$1 AND event_source='auth.totp.begin'`, userID) != 1 {
					t.Fatal("begin overlap did not produce exactly one pending enrollment and audit")
				}
			})
		}
		for _, replay := range []bool{false, true} {
			for _, disabled := range []bool{false, true} {
				name := mode + " fresh begin authority"
				if replay {
					name = mode + " replay authority"
				}
				if disabled {
					name += " disabled account"
				}
				t.Run(name, func(t *testing.T) {
					server, db := startServer(t, runtime, "auth-totp-begin-authority")
					defer db.Close()
					userID, token := prepareTOTPBeginActor(t, db, server, mode)
					wantPending := 0
					if replay {
						flowtest.BeginTOTPEnrollment(t, http.DefaultClient, server.HTTP.URL, token, map[string]any{"client_txn_id": "begin-key"})
						wantPending = 1
					}
					query := `UPDATE user_sessions SET revoked_at=created_at WHERE user_id::text=$1`
					status, code, reason := http.StatusUnauthorized, "session_required", ""
					if mode == "bootstrap" {
						query = `UPDATE bootstrap_tokens SET superseded_at=issued_at WHERE user_id::text=$1`
						status, code, reason = http.StatusConflict, "credential_bootstrap_rejected", "superseded"
					}
					if disabled {
						query = `UPDATE users SET is_active=false WHERE id::text=$1`
						status, code, reason = http.StatusUnauthorized, "session_required", ""
					}
					request := httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{"client_txn_id": "begin-key"})
					request.Header.Set("Authorization", "Bearer "+token)
					response := overlapAuthRequests(t, db, false, []*http.Request{request}, func(ctx context.Context, barrier *sql.Tx) error {
						_, err := barrier.ExecContext(ctx, query, userID)
						return err
					})[0]
					errorBody := httptestx.RequireErrorEnvelope(t, response, status, code)["error"].(map[string]any)
					if reason != "" && errorBody["details"].(map[string]any)["reason_code"] != reason {
						t.Fatal("begin rejection lost its exact bootstrap reason")
					}
					if len(response.Cookies()) != 0 || queryCount(t, db, `SELECT count(*) FROM pending_totp_enrollments WHERE user_id::text=$1`, userID) != wantPending ||
						queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE target_user_id::text=$1 AND event_source='auth.totp.begin'`, userID) != wantPending || queryUserVersion(t, db, userID) != 1 {
						t.Fatal("rejected begin changed enrollment, audit, user version or cookies")
					}
				})
			}
		}
	}
	assertTOTPReplacementBegin(t, runtime)
}

func prepareTOTPBeginActor(t *testing.T, db *sql.DB, server *httptestx.Server, mode string) (string, string) {
	t.Helper()
	userID := seedLocalUser(t, db, "begin@example.test", "Begin", "BeginPassword1!", mode == "bootstrap")
	if mode == "bootstrap" {
		return userID, flowtest.RequireBootstrapLogin(t, http.DefaultClient, server.HTTP.URL, "begin@example.test", "BeginPassword1!")
	}
	return userID, flowtest.LoginLocalUser(t, http.DefaultClient, server.HTTP.URL, "begin@example.test", "BeginPassword1!", nil).Session.Value
}

func assertTOTPReplacementBegin(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	for _, change := range []string{"none", "password", "factor"} {
		t.Run("replacement begin after "+change, func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-totp-replacement-admission")
			defer db.Close()
			now := time.Date(2026, time.April, 29, 21, 0, 0, 0, time.UTC)
			httptestx.SetClockFixed(t, server, now)
			secret := "ZILKBF5BZGY567F262KQMFGNYPCBREN3OKX65SN4EMQ7ZDSSMBOA"
			userID := seedLocalUserWithActiveTOTP(t, db, "replace@example.test", "Replace", "BeginPassword1!", false, false, secret)
			token := flowtest.LoginLocalUser(t, http.DefaultClient, server.HTTP.URL, "replace@example.test", "BeginPassword1!", nil).Session.Value
			code, err := totp.GenerateCodeCustom(secret, now, totp.ValidateOpts{Period: 30, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA256})
			if err != nil {
				t.Fatal(err)
			}
			body := map[string]any{"client_txn_id": "replace-key", "current_password": "BeginPassword1!", "second_factor": map[string]any{"kind": "totp", "assertion": map[string]any{"code": code}}}
			request := httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body)
			request.Header.Set("Authorization", "Bearer "+token)
			var mutate func(context.Context, *sql.Tx) error
			if change == "password" {
				hash, err := authn.HashPassword(t.Context(), "ReplacementPassword1!")
				if err != nil {
					t.Fatal(err)
				}
				mutate = func(ctx context.Context, tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, `UPDATE users SET password_hash=$2 WHERE id::text=$1`, userID, hash)
					return err
				}
			} else if change == "factor" {
				keys, err := authn.LoadMasterKeys(nil)
				if err != nil {
					t.Fatal(err)
				}
				id := uuid.MustParse(userID)
				envelope, err := authn.SealSecret(keys, authn.SecretBinding{Purpose: authn.ActiveTOTPSecret, RecordID: id, SubjectID: id}, []byte("12345678901234567890123456789012"))
				if err != nil {
					t.Fatal(err)
				}
				mutate = func(ctx context.Context, tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, `UPDATE users SET totp_secret_envelope=$2 WHERE id::text=$1`, userID, envelope)
					return err
				}
			}
			response := overlapAuthRequests(t, db, false, []*http.Request{request}, mutate)[0]
			wantAudits := 0
			if change == "none" {
				first := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
				replay := flowtest.BeginTOTPEnrollment(t, http.DefaultClient, server.HTTP.URL, token, map[string]any{"client_txn_id": "replace-key"})
				if !reflect.DeepEqual(first, replay) {
					t.Fatal("proof-free replay changed enrollment or seed")
				}
				otherToken := flowtest.LoginLocalUser(t, http.DefaultClient, server.HTTP.URL, "replace@example.test", "BeginPassword1!", nil).Session.Value
				otherScope := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{"client_txn_id": "replace-key"}, httptestx.WithHeader("Authorization", "Bearer "+otherToken))
				httptestx.RequireErrorEnvelope(t, otherScope, http.StatusConflict, "client_txn_conflict")
				// A replay whose pending state disappears under admission must not
				// silently turn into a fresh replacement without password/TOTP proof.
				retry := httptestx.NewJSONRequest(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", map[string]any{"client_txn_id": "replace-key"})
				retry.Header.Set("Authorization", "Bearer "+token)
				lost := overlapAuthRequests(t, db, false, []*http.Request{retry}, func(ctx context.Context, tx *sql.Tx) error {
					_, err := tx.ExecContext(ctx, `DELETE FROM pending_totp_enrollments WHERE user_id::text=$1`, userID)
					return err
				})[0]
				httptestx.RequireErrorEnvelope(t, lost, http.StatusConflict, "client_txn_conflict")
				wantAudits = 1 // One original begin audit; fixture removed its pending row.
			} else if change == "password" {
				httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "invalid_current_password")
			} else {
				httptestx.RequireErrorEnvelope(t, response, http.StatusUnauthorized, "invalid_second_factor")
			}
			if queryCount(t, db, `SELECT count(*) FROM pending_totp_enrollments WHERE user_id::text=$1`, userID) != 0 ||
				queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE target_user_id::text=$1 AND event_source='auth.totp.begin'`, userID) != wantAudits || queryUserVersion(t, db, userID) != 1 {
				t.Fatal("replacement admission changed durable effects incorrectly")
			}
		})
	}
}
