package auth_test

import (
	"net/http"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// Each caller keeps these cases within its existing public-route selection.
func assertConcurrentLocalReceipts(t *testing.T, runtime *appsupport.Runtime, operation string) {
	t.Helper()
	for _, divergent := range []bool{false, true} {
		name := operation + " identical overlap"
		if divergent {
			name = operation + " divergent overlap"
		}
		t.Run(name, func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-local-overlap")
			defer db.Close()
			adminID := seedLocalUserFlags(t, db, "admin@example.test", "Admin", "LocalAdmin1!", false, true, true)
			login := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "admin@example.test", "LocalAdmin1!", nil)
			actorID, targetID := adminID, ""
			if operation != "create" {
				targetID = seedLocalUser(t, db, "target@example.test", "Target", "LocalTarget1!", false)
				targetLogin := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "target@example.test", "LocalTarget1!", nil)
				if operation == "password-change" {
					actorID, login = targetID, targetLogin
				}
			}
			path, route, scope := "/api/v1/users", "users.create", "actor"
			body := map[string]any{"client_txn_id": "local-overlap"}
			firstStatus, wantVersion := http.StatusOK, 2
			switch operation {
			case "create":
				body["auth_kind"], body["email"], body["display_name"] = "local", "created@example.test", "Created"
				body["initial_password"], body["mfa_required"] = "CreatedPassword1!", false
				firstStatus, wantVersion = http.StatusCreated, 1
			case "password-change":
				path, route = "/api/v1/auth/password/change", "auth.password.change"
				body["current_password"], body["new_password"] = "LocalTarget1!", "ChangedPassword1!"
			case "password-reset":
				path, route, scope = "/api/v1/users/"+targetID+"/password/reset", "users.password.reset", targetID
				body["base_user_version"], body["new_password"], body["reason"] = 1, "ChangedPassword1!", "original"
			case "totp-reset":
				path, route, scope = "/api/v1/users/"+targetID+"/mfa/totp/reset", "users.totp.reset", targetID
				body["base_user_version"], body["reason"] = 1, "original"
			case "revoke-all":
				path, route, scope = "/api/v1/users/"+targetID+"/sessions/revoke-all", "users.sessions.revoke_all", targetID
				body["reason"], wantVersion = "original", 1
			default:
				t.Fatal("unknown local receipt operation")
			}
			other := make(map[string]any, len(body))
			for key, value := range body {
				other[key] = value
			}
			if divergent {
				switch operation {
				case "create":
					other["display_name"] = "Other Created"
				case "password-change":
					other["new_password"] = "OtherChangedPassword1!"
				default:
					other["reason"] = "different"
				}
			}
			responses := overlapAuthMutationRequests(t, db, server, login, http.MethodPost, path, body, other)
			statuses := []int{responses[0].StatusCode, responses[1].StatusCode}
			wantStatuses := []int{firstStatus, http.StatusOK}
			if divergent {
				wantStatuses[1] = http.StatusConflict
			}
			slices.Sort(statuses)
			slices.Sort(wantStatuses)
			if !slices.Equal(statuses, wantStatuses) {
				t.Fatalf("overlap statuses = %v, want %v", statuses, wantStatuses)
			}
			receipt := lookupRouteIdempotency(t, db, route, actorID, scope, "local-overlap")
			if receipt.StatusCode != firstStatus {
				t.Fatalf("stored first status = %d, want %d", receipt.StatusCode, firstStatus)
			}
			if operation == "create" {
				if err := db.QueryRowContext(t.Context(), `SELECT id::text FROM users WHERE email='created@example.test'`).Scan(&targetID); err != nil {
					t.Fatal(err)
				}
			}
			for _, response := range responses {
				if response.StatusCode == http.StatusConflict {
					httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "client_txn_conflict")
					continue
				}
				data := httptestx.RequireSuccessEnvelope(t, response, response.StatusCode)["data"].(map[string]any)
				requireJSONEquivalent(t, data, receipt.Response)
				if data["user_id"] != targetID {
					t.Fatal("overlap returned another user")
				}
			}
			if queryUserVersion(t, db, targetID) != int64(wantVersion) {
				t.Fatal("overlap advanced the user version incorrectly")
			}
			for _, check := range []struct {
				name, query string
				args        []any
				want        int
			}{
				{"user", `SELECT count(*) FROM users WHERE id::text=$1`, []any{targetID}, 1},
				{"active sessions", `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, []any{targetID}, 0},
				{"receipt", `SELECT count(*) FROM route_idempotency WHERE route_key=$1 AND actor_user_id::text=$2 AND scope_key=$3 AND client_txn_id='local-overlap'`, []any{route, actorID, scope}, 1},
				{"audit", `SELECT count(*) FROM deployment_admin_audit_events WHERE event_source=$1 AND actor_user_id::text=$2 AND target_user_id::text=$3`, []any{route, actorID, targetID}, 1},
			} {
				if got := queryCount(t, db, check.query, check.args...); got != check.want {
					t.Errorf("%s = %d, want %d", check.name, got, check.want)
				}
			}
			password := "LocalTarget1!"
			switch operation {
			case "create":
				password = "CreatedPassword1!"
			case "password-reset":
				password = "ChangedPassword1!"
			case "password-change":
				password = "ChangedPassword1!"
				if divergent {
					// The winner is nondeterministic, but must be one of the
					// two independently authored passwords, never a third value.
					stored := queryUserPasswordHash(t, db, targetID)
					if ok, err := authn.VerifyPasswordHash(t.Context(), stored, password); err != nil {
						t.Fatal(err)
					} else if !ok {
						password = "OtherChangedPassword1!"
					}
				}
			}
			requireCurrentPasswordHash(t, queryUserPasswordHash(t, db, targetID), password, "UnrelatedPassword1!")
		})
	}
}
