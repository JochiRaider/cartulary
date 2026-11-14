package auth_test

import (
	"bytes"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func assertConcurrentSessionLimit(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	for _, initial := range []int{4, 5} {
		t.Run(fmt.Sprintf("overlapping logins with %d existing sessions", initial), func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-session-overlap")
			defer db.Close()
			userID := seedLocalUser(t, db, "overlap@example.test", "Overlap", "OverlapPassword1!", false)
			base := time.Date(2026, time.April, 29, 18, 30, 0, 0, time.UTC)
			oldIDs := make([]string, 0, initial)
			for index := 0; index < initial; index++ {
				httptestx.SetClockFixed(t, server, base.Add(time.Duration(index)*time.Minute))
				flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "overlap@example.test", "OverlapPassword1!", nil)
				var id string
				if err := db.QueryRowContext(t.Context(), `SELECT id::text FROM user_sessions WHERE user_id::text=$1 ORDER BY authenticated_at DESC LIMIT 1`, userID).Scan(&id); err != nil {
					t.Fatal(err)
				}
				oldIDs = append(oldIDs, id)
			}
			httptestx.SetClockFixed(t, server, base.Add(10*time.Minute))
			requests := make([]*http.Request, 2)
			for index := range requests {
				request, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", bytes.NewBufferString(`{"username":"overlap@example.test","password":"OverlapPassword1!"}`))
				if err != nil {
					t.Fatal(err)
				}
				request.Header.Set("Content-Type", "application/json")
				requests[index] = request
			}
			responses := overlapAuthRequests(t, db, false, requests, nil)
			var previousToken string
			for _, response := range responses {
				data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
				if data["user_id"] != userID || data["provider_type"] != "local" {
					t.Fatal("overlap returned an incorrect user or provider")
				}
				cookies := authcookietest.RequireAuthCookies(t, response.Cookies())
				if cookies.Session.Value == previousToken {
					t.Fatal("overlap reused a session credential")
				}
				previousToken = cookies.Session.Value
				inspection := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(cookies.Session))
				if httptestx.RequireSuccessEnvelope(t, inspection, http.StatusOK)["data"].(map[string]any)["user_id"] != userID {
					t.Fatal("issued session does not authenticate its user")
				}
			}
			rows, err := db.QueryContext(t.Context(), `SELECT id::text FROM user_sessions WHERE user_id::text=$1 AND authenticated_at=$2`, userID, base.Add(10*time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			var newIDs []string
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err != nil {
					t.Fatal(err)
				}
				if slices.Contains(oldIDs, id) || slices.Contains(newIDs, id) {
					t.Fatal("overlap reused a durable session identity")
				}
				newIDs = append(newIDs, id)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if len(newIDs) != 2 {
				t.Fatalf("overlap created %d new sessions, want 2", len(newIDs))
			}
			if got := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, userID); got != 5 {
				t.Fatalf("overlapping logins left %d active sessions, want 5", got)
			}
			if got := queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1`, userID); got != initial+2 {
				t.Fatalf("session count = %d, want %d", got, initial+2)
			}
			wantRevoked := initial + 2 - 5
			for index, id := range oldIDs {
				stored := querySessionByID(t, db, id)
				if index < wantRevoked {
					if !stored.revokedAt.Valid || stored.revokeReasonCode.String != "concurrency_limit" || !stored.revokedAt.Time.Equal(base.Add(10*time.Minute)) {
						t.Fatal("overlap did not revoke the exact oldest session with its reason and time")
					}
				} else if stored.revokedAt.Valid {
					t.Fatal("overlap revoked a retained predecessor")
				}
			}
			for _, id := range newIDs {
				if querySessionByID(t, db, id).revokedAt.Valid {
					t.Fatal("overlap revoked a newly issued session")
				}
			}
			if got := queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE actor_user_id::text=$1 AND target_user_id::text=$1 AND reason_code='concurrency_limit'`, userID); got != wantRevoked {
				t.Fatalf("attributed concurrency audit count = %d, want %d", got, wantRevoked)
			}
		})
	}
}
