package workbook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func exerciseHostIdentityWaitingRebase(t *testing.T) {
	h, login, actor, incident := ConflictFixture(t, "entity-waiting-rebase", "IR-ENTITY-WAIT")
	for _, tc := range []struct{ kind, view, identifier string }{
		{"host", "cartulary.view.hosts.v1", "host.hostname"},
		{"identity", "cartulary.view.identities.v1", "identity.upn"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			field := tc.kind + ".display_name"
			row := requireWorkbookCreate(t, h, login, incident, tc.view, map[string]any{"client_txn_id": tc.kind + "-wait-create", field: "Original", tc.identifier: tc.kind + "-wait@example.test"})["row"].(map[string]any)
			id := appsupport.MustUUID(t, row["record_id"].(string))
			holder, err := h.Pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer holder.Rollback(context.Background())
			var holderPID int32
			if err := holder.QueryRow(ctx, `SELECT pg_backend_pid() FROM users WHERE id=$1 FOR UPDATE`, actor).Scan(&holderPID); err != nil {
				t.Fatal(err)
			}
			type result struct {
				response *http.Response
				err      error
			}
			done := []chan result{make(chan result, 1), make(chan result, 1)}
			started, consumed := [2]bool{}, [2]bool{}
			defer func() {
				cancel()
				_ = holder.Rollback(context.Background())
				for i := range done {
					if started[i] && !consumed[i] {
						select {
						case r := <-done[i]:
							if r.response != nil {
								_ = r.response.Body.Close()
							}
						case <-time.After(5 * time.Second):
							t.Error("entity writer did not drain")
						}
					}
				}
			}()
			requests := []map[string]any{
				{"view_schema_id": tc.view, "base_row_version": 1, "client_txn_id": tc.kind + "-wait-first", "changes": []map[string]any{{"field_key": field, "value": "First saved"}}},
				{"view_schema_id": tc.view, "base_row_version": 1, "client_txn_id": tc.kind + "-wait-second", "changes": []map[string]any{{"field_key": tc.identifier, "value": tc.kind + "-rebased@example.test"}}},
			}
			launch := func(i int) {
				t.Helper()
				encoded, err := json.Marshal(requests[i])
				if err != nil {
					t.Fatal(err)
				}
				request, err := http.NewRequestWithContext(ctx, http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+id.String(), strings.NewReader(string(encoded)))
				if err != nil {
					t.Fatal(err)
				}
				request.Header.Set("Content-Type", "application/json")
				request.AddCookie(login.Session)
				request.AddCookie(login.CSRF)
				request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
				started[i] = true
				go func() { r, e := http.DefaultClient.Do(request); done[i] <- result{r, e} }()
			}
			waitForBlock := func(blocker int32, i int) int32 {
				t.Helper()
				for {
					var blockedPID int32
					if err := h.Pool.QueryRow(ctx, `SELECT coalesce((SELECT pid FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)) ORDER BY pid LIMIT 1),0)`, blocker).Scan(&blockedPID); err != nil {
						t.Fatal(err)
					}
					if blockedPID != 0 {
						return blockedPID
					}
					select {
					case r := <-done[i]:
						consumed[i] = true
						if r.response != nil {
							defer r.response.Body.Close()
							t.Fatalf("writer %d bypassed wait: %d %v", i, r.response.StatusCode, r.err)
						}
						t.Fatal(r.err)
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
			}
			launch(0)
			firstPID := waitForBlock(holderPID, 0)
			launch(1)
			_ = waitForBlock(firstPID, 1)
			if err := holder.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			var responses [2]map[string]any
			for i := range done {
				var r result
				select {
				case r = <-done[i]:
					consumed[i] = true
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				}
				if r.err != nil {
					t.Fatal(r.err)
				}
				responses[i] = httptestx.RequireSuccessData(t, r.response, http.StatusOK)
				if responses[i]["row"].(map[string]any)["row_version"] != float64(i+2) {
					t.Fatalf("writer %d version=%#v", i, responses[i])
				}
			}
			final := responses[1]["row"].(map[string]any)
			requireCellValue(t, final, field, "First saved")
			requireCellValue(t, final, tc.identifier, tc.kind+"-rebased@example.test")
			before := snapshotWorkbookState(t, h, login, incident, tc.view)
			for i, request := range requests {
				replay := requireWorkbookPatch(t, h, login, id, request)
				if !reflect.DeepEqual(replay, responses[i]) {
					t.Fatalf("writer %d receipt changed", i)
				}
				requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, tc.view))
			}
		})
	}
}
