package auth_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// These are public-route cases in the existing binding-lifecycle selection.
func assertConcurrentBindingReceipts(t *testing.T, runtime *appsupport.Runtime) {
	t.Helper()
	for _, operation := range []string{"create", "rotate", "noop", "retire"} {
		for _, divergent := range []bool{false, true} {
			name := operation + " identical overlap"
			if divergent {
				name = operation + " divergent overlap"
			}
			t.Run(name, func(t *testing.T) {
				server, db := startEnterpriseServer(t, runtime, "auth-binding-overlap")
				defer db.Close()
				adminID := seedLocalUserFlags(t, db, "admin@example.test", "Admin", "ConcurrentAdmin1!", false, true, true)
				targetID := seedLocalUser(t, db, "target@example.test", "Target", "ConcurrentTarget1!", false)
				login := flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "admin@example.test", "ConcurrentAdmin1!", nil)
				flowtest.LoginLocalUser(t, server.HTTP.Client(), server.HTTP.URL, "target@example.test", "ConcurrentTarget1!", nil)
				passwordBefore := queryUserPasswordHash(t, db, targetID)
				providerID := seedEnterpriseProvider(t, db, enterpriseProviderSeed{Key: "corp-overlap", Type: "oidc", Name: "Overlap"})
				path := "/api/v1/users/" + targetID + "/auth-bindings"
				method, route, scope := http.MethodPost, "users.auth_bindings.create", targetID
				body := map[string]any{"base_user_version": 1, "client_txn_id": "overlap", "reason": "original"}
				wantVersion, wantBindings, wantActive, wantAudit, wantSessions := 2, 1, 1, 1, 1
				firstStatus := http.StatusCreated
				if operation == "create" {
					body["provider_key"], body["provider_subject"] = "corp-overlap", "original-subject"
				} else {
					bindingID := seedEnterpriseBinding(t, db, targetID, providerID, "original-subject", adminID)
					path += "/" + bindingID
					scope, firstStatus = bindingID, http.StatusOK
					if operation == "retire" {
						method, route, wantActive, wantSessions = http.MethodDelete, "users.auth_bindings.retire", 0, 0
					} else {
						path += "/rotate"
						route, body["new_provider_subject"] = "users.auth_bindings.rotate", "replacement-subject"
						wantBindings, wantSessions = 2, 0
						if operation == "noop" {
							body["new_provider_subject"] = "original-subject"
							wantVersion, wantBindings, wantAudit, wantSessions = 1, 1, 0, 1
						}
					}
				}
				other := make(map[string]any, len(body))
				for key, value := range body {
					other[key] = value
				}
				if divergent {
					other["reason"] = "different"
				}
				responses := overlapAuthMutationRequests(t, db, server, login, method, path, body, other)
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
				receipt := lookupRouteIdempotency(t, db, route, adminID, scope, "overlap")
				if receipt.StatusCode != firstStatus {
					t.Fatalf("stored first status = %d, want %d", receipt.StatusCode, firstStatus)
				}
				for _, response := range responses {
					if response.StatusCode == http.StatusConflict {
						httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "client_txn_conflict")
						continue
					}
					data := httptestx.RequireSuccessEnvelope(t, response, response.StatusCode)["data"].(map[string]any)
					requireJSONEquivalent(t, data, receipt.Response)
					if data["user_id"] != targetID || data["user_version"] != float64(wantVersion) {
						t.Fatal("overlap returned an incorrect user or version")
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
					{"bindings", `SELECT count(*) FROM enterprise_auth_bindings WHERE user_id::text=$1`, []any{targetID}, wantBindings},
					{"active bindings", `SELECT count(*) FROM enterprise_auth_bindings WHERE user_id::text=$1 AND retired_at IS NULL`, []any{targetID}, wantActive},
					{"active sessions", `SELECT count(*) FROM user_sessions WHERE user_id::text=$1 AND revoked_at IS NULL`, []any{targetID}, wantSessions},
					{"receipt", `SELECT count(*) FROM route_idempotency WHERE route_key=$1 AND actor_user_id::text=$2 AND scope_key=$3 AND client_txn_id='overlap'`, []any{route, adminID, scope}, 1},
					{"audit", `SELECT count(*) FROM deployment_admin_audit_events WHERE event_source=$1 AND actor_user_id::text=$2 AND target_user_id::text=$3`, []any{route, adminID, targetID}, wantAudit},
				} {
					if got := queryCount(t, db, check.query, check.args...); got != check.want {
						t.Errorf("%s = %d, want %d", check.name, got, check.want)
					}
				}
				if queryUserPasswordHash(t, db, targetID) != passwordBefore {
					t.Fatal("binding overlap changed local credentials")
				}
			})
		}
	}
}

func overlapAuthMutationRequests(t *testing.T, db *sql.DB, server *httptestx.Server, login authcookietest.AuthCookies, method, path string, bodies ...map[string]any) []*http.Response {
	t.Helper()
	requests := make([]*http.Request, 0, len(bodies))
	for _, body := range bodies {
		encoded, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(method, server.HTTP.URL+path, bytes.NewReader(encoded))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
		request.AddCookie(login.Session)
		request.AddCookie(login.CSRF)
		requests = append(requests, request)
	}
	return overlapAuthRequests(t, db, false, requests, nil)
}

// afterMutation moves the barrier to audit insertion, allowing independent
// target writes to expose decisions made before either transaction commits.
func overlapAuthRequests(t *testing.T, db *sql.DB, afterMutation bool, requests []*http.Request, beforeRelease func(context.Context, *sql.Tx) error) []*http.Response {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	barrier, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer barrier.Rollback()
	// EXCLUSIVE permits authentication reads but blocks owner mutation or
	// target SELECT FOR UPDATE, including structural no-ops.
	if afterMutation {
		_, err = barrier.ExecContext(ctx, `LOCK TABLE deployment_admin_audit_events IN ACCESS EXCLUSIVE MODE`)
	} else {
		_, err = barrier.ExecContext(ctx, `LOCK TABLE users, user_sessions, account_preferences IN EXCLUSIVE MODE`)
	}
	if err != nil {
		t.Fatal(err)
	}
	for index, request := range requests {
		requests[index] = request.WithContext(ctx)
	}
	var pid int
	if err := barrier.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		response *http.Response
		err      error
	}
	results := make(chan outcome, len(requests))
	settled := 0
	defer func() {
		cancel()
		_ = barrier.Rollback()
		deadline := time.NewTimer(5 * time.Second)
		defer deadline.Stop()
		for settled < len(requests) {
			select {
			case result := <-results:
				settled++
				if result.response != nil {
					_ = result.response.Body.Close()
				}
			case <-deadline.C:
				t.Error("auth requests did not settle after cancellation")
				return
			}
		}
	}()
	for _, request := range requests {
		go func() { response, err := http.DefaultClient.Do(request); results <- outcome{response, err} }()
	}
	for {
		var waiting int
		if err := db.QueryRowContext(ctx, `WITH RECURSIVE blocked(pid) AS (
SELECT pid FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid))
UNION SELECT a.pid FROM pg_stat_activity a JOIN blocked b ON b.pid=ANY(pg_blocking_pids(a.pid))
) SELECT count(DISTINCT pid) FROM blocked`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting == len(requests) {
			break
		}
		select {
		case early := <-results:
			settled++
			if early.response != nil {
				_ = early.response.Body.Close()
			}
			t.Fatalf("auth request finished before observed overlap: %v", early.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if beforeRelease != nil {
		if err := beforeRelease(ctx, barrier); err != nil {
			t.Fatal(err)
		}
	}
	if err := barrier.Commit(); err != nil {
		t.Fatal(err)
	}
	responses := make([]*http.Response, 0, len(requests))
	for settled < len(requests) {
		select {
		case result := <-results:
			settled++
			if result.response != nil {
				t.Cleanup(func() { _ = result.response.Body.Close() })
			}
			if result.err != nil {
				t.Fatal(result.err)
			}
			// Consume the transport body while the request context is alive;
			// callers inspect the buffered response after this helper cancels it.
			payload, err := io.ReadAll(result.response.Body)
			_ = result.response.Body.Close()
			if err != nil {
				t.Fatal(err)
			}
			result.response.Body = io.NopCloser(bytes.NewReader(payload))
			responses = append(responses, result.response)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	return responses
}
