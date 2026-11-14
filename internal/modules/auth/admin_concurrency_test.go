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

func assertConcurrentLastAdminGuard(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	for _, field := range []string{"is_active", "is_deployment_admin"} {
		t.Run("concurrent last admin "+field, func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-last-admin-overlap")
			defer db.Close()
			// The server creates a bootstrap administrator. This fixture isolates
			// the two accounts whose concurrent removal is under observation.
			if _, err := db.ExecContext(t.Context(), `UPDATE users SET is_deployment_admin=false`); err != nil {
				t.Fatal(err)
			}
			ids := make([]string, 0, 2)
			requests := make([]*http.Request, 0, 2)
			for _, email := range []string{"first@example.test", "second@example.test"} {
				id := seedLocalUserFlags(t, db, email, "Admin", "ConcurrentAdmin1!", false, true, true)
				login := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, email, "ConcurrentAdmin1!", nil)
				request := httptestx.NewJSONRequest(t, http.MethodPatch, server.HTTP.URL+"/api/v1/users/"+id, map[string]any{"base_user_version": 1, field: false})
				request.AddCookie(login.Session)
				request.AddCookie(login.CSRF)
				request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
				ids = append(ids, id)
				requests = append(requests, request)
			}
			if queryCount(t, db, `SELECT count(*) FROM users WHERE is_active AND is_deployment_admin`) != 2 {
				t.Fatal("last-admin fixture must start with exactly two active administrators")
			}
			responses := overlapAuthRequests(t, db, true, requests, nil)
			statuses := []int{responses[0].StatusCode, responses[1].StatusCode}
			slices.Sort(statuses)
			if !slices.Equal(statuses, []int{http.StatusOK, http.StatusConflict}) {
				t.Fatalf("last-admin overlap statuses = %v", statuses)
			}
			var winner string
			for _, response := range responses {
				if response.StatusCode == http.StatusConflict {
					httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "last_deployment_admin")
					continue
				}
				data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
				winner, _ = data["user_id"].(string)
				if !slices.Contains(ids, winner) || data["user_version"] != float64(2) || data[field] != false {
					t.Fatal("incorrect successful admin mutation")
				}
			}
			for _, id := range ids {
				want := int64(1)
				if id == winner {
					want = 2
				}
				if queryUserVersion(t, db, id) != want {
					t.Fatal("denied admin advanced version")
				}
			}
			if queryCount(t, db, `SELECT count(*) FROM users WHERE is_active AND is_deployment_admin`) != 1 {
				t.Fatal("last-admin invariant lost")
			}
			if queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_source='users.patch'`) != 1 {
				t.Fatal("last-admin rejection emitted an audit effect")
			}
			expectedSessions := 2
			if field == "is_active" {
				expectedSessions = 1
			}
			if queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE revoked_at IS NULL`) != expectedSessions {
				t.Fatal("incorrect admin session revocation")
			}
		})
	}
}
