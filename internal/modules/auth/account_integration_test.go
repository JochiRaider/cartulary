package auth_test

import (
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestCurrentAccountContracts_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	for _, profile := range []bool{true, false} {
		name, method, version, field, initial, changed := "preferences", http.MethodPut, "preferences_version", "density_mode", any(nil), any("compact")
		if profile {
			name, method, version, field, initial, changed = "profile", http.MethodPatch, "user_version", "display_name", "Account User", "Updated User"
		}
		path, route := "/api/v1/account/"+name, "account.preferences.put"
		if profile {
			route = "account.profile.patch"
		}
		t.Run(name+" lifecycle and admission", func(t *testing.T) {
			server, db := startServer(t, runtime, "auth-account-contract")
			defer db.Close()
			actorID := seedLocalUser(t, db, "account@example.test", "Account User", "AccountPassword1!", false)
			otherID := seedLocalUserFlags(t, db, "admin@example.test", "Other Admin", "OtherPassword1!", false, true, true)
			login := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "account@example.test", "AccountPassword1!", nil)
			other := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "admin@example.test", "OtherPassword1!", nil)
			// Fix the application clock within the existing sessions' lifetimes.
			var authenticated time.Time
			if err := db.QueryRowContext(t.Context(), `SELECT authenticated_at FROM user_sessions WHERE user_id::text=$1`, actorID).Scan(&authenticated); err != nil {
				t.Fatal(err)
			}
			now := authenticated.Add(time.Minute).UTC().Truncate(time.Microsecond)
			httptestx.SetClockFixed(t, server, now)
			read := func() map[string]any {
				response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+path, nil, httptestx.WithCookies(login.Session))
				return httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			}
			write := func(body map[string]any) *http.Response {
				return httptestx.DoJSON(t, http.DefaultClient, method, server.HTTP.URL+path, body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
			}
			request := func(key string, base int, value any) map[string]any {
				return map[string]any{"client_txn_id": key, "base_" + version: base, field: value}
			}
			before := read()
			want := map[string]any{"user_id": actorID, field: initial, version: float64(1)}
			var created, updated time.Time
			query := `SELECT created_at, updated_at FROM account_preferences WHERE user_id::text=$1`
			if profile {
				query = `SELECT created_at, updated_at FROM users WHERE id::text=$1`
				want["email"] = "account@example.test"
			}
			if err := db.QueryRowContext(t.Context(), query, actorID).Scan(&created, &updated); err != nil {
				t.Fatal(err)
			}
			want["created_at"], want["updated_at"] = created.UTC(), updated.UTC()
			requireAccountResource(t, before, want)
			requireJSONEquivalent(t, read(), before)
			if !profile && queryCount(t, db, `SELECT count(*) FROM account_preferences WHERE user_id::text=$1`, actorID) != 1 {
				t.Fatal("preferences were not durably initialized once")
			}
			for _, query := range []string{"limit=1", "cursor_token=x", "page=1", "offset=1", "page_size=1", "block_size=1"} {
				response := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+path+"?"+query, nil, httptestx.WithCookies(login.Session))
				envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_pagination_request")
				if envelope["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "pagination_not_supported" {
					t.Fatal("wrong singleton pagination reason")
				}
			}
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+path, nil), http.StatusUnauthorized, "session_required")
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, method, server.HTTP.URL+path, []any{}), http.StatusUnauthorized, "session_required")
			requireCookieCSRFRejected(t, server, method, path, []any{}, login.Session, login.CSRF)
			invalid := []struct {
				body  any
				field string
			}{
				{[]any{}, ""}, {map[string]any{}, "base_" + version},
				{request("bad", 0, initial), "base_" + version}, {request("bad", 1, 42), field},
			}
			for _, missing := range []string{"client_txn_id", field} {
				b := request("bad", 1, initial)
				delete(b, missing)
				invalid = append(invalid, struct {
					body  any
					field string
				}{b, missing})
			}
			for _, nullable := range []string{"base_" + version, "client_txn_id"} {
				b := request("bad", 1, initial)
				b[nullable] = nil
				invalid = append(invalid, struct {
					body  any
					field string
				}{b, nullable})
			}
			for _, forbidden := range []string{"user_id", "email", "locale", "time_zone", "theme", "password", "row_height", "home_sheet_ref"} {
				b := request("bad", 1, initial)
				b[forbidden] = otherID
				invalid = append(invalid, struct {
					body  any
					field string
				}{b, forbidden})
			}
			invalidValue := any("COMPACT")
			if profile {
				invalidValue = nil
			}
			invalid = append(invalid, struct {
				body  any
				field string
			}{request("bad", 1, invalidValue), field})
			for _, tc := range invalid {
				response := httptestx.DoJSON(t, http.DefaultClient, method, server.HTTP.URL+path, tc.body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
				envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
				if tc.field != "" && envelope["error"].(map[string]any)["details"].(map[string]any)["field"] != tc.field {
					t.Fatalf("wrong invalid field for %#v", tc.body)
				}
			}
			requireJSONEquivalent(t, read(), before)
			noOpValue := initial
			if profile {
				noOpValue = "  Account User  "
			}
			noop := httptestx.RequireSuccessEnvelope(t, write(request("noop", 1, noOpValue)), http.StatusOK)["data"]
			requireJSONEquivalent(t, noop, before)
			if queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_source=$1 AND target_user_id::text=$2`, route, actorID) != 0 {
				t.Fatal("invalid/no-op request emitted audit")
			}
			if lookupRouteIdempotency(t, db, route, actorID, "actor", "noop").StatusCode != http.StatusOK {
				t.Fatal("no-op receipt missing")
			}
			now = now.Add(time.Minute)
			httptestx.SetClockFixed(t, server, now)
			material := request("material", 1, changed)
			result := httptestx.RequireSuccessEnvelope(t, write(material), http.StatusOK)["data"].(map[string]any)
			want[field], want[version], want["updated_at"] = changed, float64(2), now
			requireAccountResource(t, result, want)
			requireAccountResource(t, read(), want)
			if queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_source=$1 AND target_user_id::text=$2`, route, actorID) != 1 {
				t.Fatal("material change must audit once")
			}
			state := observeUserAdministration(t, db, actorID, actorID)
			for _, prior := range []struct {
				body map[string]any
				want any
			}{{request("noop", 1, noOpValue), before}, {material, result}} {
				replay := httptestx.RequireSuccessEnvelope(t, write(prior.body), http.StatusOK)["data"]
				requireJSONEquivalent(t, replay, prior.want)
			}
			state.requireUnchanged(t, observeUserAdministration(t, db, actorID, actorID))
			requireAccountResource(t, read(), want)
			conflict := request("material", 1, initial)
			httptestx.RequireErrorEnvelope(t, write(conflict), http.StatusConflict, "client_txn_conflict")
			httptestx.RequireErrorEnvelope(t, write(request("stale", 1, changed)), http.StatusConflict, version+"_conflict")
			otherResponse := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+path, nil, httptestx.WithCookies(other.Session))
			if httptestx.RequireSuccessEnvelope(t, otherResponse, http.StatusOK)["data"].(map[string]any)["user_id"] != otherID {
				t.Fatal("admin read crossed actor scope")
			}
			session := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(login.Session)), http.StatusOK)["data"].(map[string]any)
			expectedName := "Account User"
			if profile {
				expectedName = "Updated User"
			}
			if session["display_name"] != expectedName {
				t.Fatal("session did not publish current display name")
			}
			if queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, actorID) != 1 {
				t.Fatal("account change revoked session")
			}
			if !profile {
				cleared := httptestx.RequireSuccessEnvelope(t, write(request("clear", 2, nil)), http.StatusOK)["data"].(map[string]any)
				if cleared["density_mode"] != nil || cleared[version] != float64(3) {
					t.Fatal("explicit null did not clear override")
				}
				for i, value := range []string{"default", "comfortable"} {
					data := httptestx.RequireSuccessEnvelope(t, write(request(value, 3+i, value)), http.StatusOK)["data"].(map[string]any)
					if data[field] != value || data[version] != float64(4+i) {
						t.Fatal("closed density token not persisted")
					}
				}
				if queryUserVersion(t, db, actorID) != 1 {
					t.Fatal("preferences changed user version")
				}
			}
		})
		for _, noop := range []bool{false, true} {
			for _, divergent := range []bool{false, true} {
				if noop && divergent {
					continue
				}
				title := name + " material identical overlap"
				if noop {
					title = name + " no-op identical overlap"
				}
				if divergent {
					title = name + " material divergent overlap"
				}
				t.Run(title, func(t *testing.T) {
					server, db := startServer(t, runtime, "auth-account-overlap")
					defer db.Close()
					actorID := seedLocalUser(t, db, "account@example.test", "Account User", "AccountPassword1!", false)
					login := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "account@example.test", "AccountPassword1!", nil)
					value := changed
					expectedVersion, audits := 2, 1
					if noop {
						value = initial
						expectedVersion, audits = 1, 0
					}
					body := map[string]any{"client_txn_id": "overlap", "base_" + version: 1, field: value}
					other := map[string]any{"client_txn_id": "overlap", "base_" + version: 1, field: value}
					if divergent {
						if profile {
							other[field] = "Other User"
						} else {
							other[field] = "comfortable"
						}
					}
					responses := overlapAuthMutationRequests(t, db, server, login, method, path, body, other)
					statuses := []int{responses[0].StatusCode, responses[1].StatusCode}
					slices.Sort(statuses)
					wantStatuses := []int{200, 200}
					if divergent {
						wantStatuses[1] = 409
					}
					if !slices.Equal(statuses, wantStatuses) {
						t.Fatalf("overlap statuses %v", statuses)
					}
					receipt := lookupRouteIdempotency(t, db, route, actorID, "actor", "overlap")
					for _, response := range responses {
						if response.StatusCode == 409 {
							httptestx.RequireErrorEnvelope(t, response, 409, "client_txn_conflict")
							continue
						}
						data := httptestx.RequireSuccessEnvelope(t, response, 200)["data"].(map[string]any)
						requireJSONEquivalent(t, data, receipt.Response)
						if data["user_id"] != actorID || data[version] != float64(expectedVersion) {
							t.Fatal("incorrect overlap resource")
						}
					}
					if queryCount(t, db, `SELECT count(*) FROM route_idempotency WHERE route_key=$1 AND actor_user_id::text=$2`, route, actorID) != 1 {
						t.Fatal("overlap did not store exactly one receipt")
					}
					if queryCount(t, db, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_source=$1 AND target_user_id::text=$2`, route, actorID) != audits {
						t.Fatal("overlap repeated audit")
					}
					if queryCount(t, db, `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, actorID) != 1 {
						t.Fatal("overlap revoked session")
					}
				})
			}
		}
	}
}

// Timestamp spelling may use an RFC3339 offset; resource equality compares
// the independently expected instant while keeping every other member exact.
func requireAccountResource(t testing.TB, got, want map[string]any) {
	t.Helper()
	actual := make(map[string]any, len(got))
	for key, value := range got {
		actual[key] = value
	}
	for _, field := range []string{"created_at", "updated_at"} {
		encoded, ok := got[field].(string)
		if !ok {
			t.Fatalf("%s is not a timestamp", field)
		}
		parsed, err := time.Parse(time.RFC3339Nano, encoded)
		expected := want[field].(time.Time)
		if err != nil || !parsed.Equal(expected) {
			t.Fatalf("%s did not equal the expected instant", field)
		}
		actual[field] = expected
	}
	requireJSONEquivalent(t, actual, want)
}
