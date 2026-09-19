package networkflow_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	nf "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestTableLifecycleAdmissionAndReplay_Integration(t *testing.T) {
	t.Run("functional authorization controls", assertNetworkFlowAuthorizationConsumers)
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "table-lifecycle")
	login, actorText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	actor := uuid.MustParse(actorText)
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "table-lifecycle-incident", "incident_key": "IR-TABLE-LIFECYCLE", "title": "Table lifecycle"})
	id := uuid.MustParse(incident["incident_id"].(string))
	t.Run("major7 admission matrix", func(t *testing.T) {
		assertMajor7AdmissionMatrix(t, harness.Server.HTTP.URL, harness.Pool, id, actor, login.SessionCookie, login.CSRFCookie)
	})
	ctx := context.Background()
	store := newTestNetworkFlowStore(t, harness.Pool, harness.Revisions.Appender())
	table := createTestTable(t, harness.Pool, store, actor, id, "lifecycle.csv", nil, 1)
	path := harness.Server.HTTP.URL + "/api/v1/incidents/" + id.String() + "/network-flow/tables/" + table.TableID
	options := []func(*http.Request){httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value)}
	request := func(method string, body map[string]any) *http.Response {
		return httptestx.DoJSON(t, method, path, body, options...)
	}
	role := func(value string) {
		t.Helper()
		if _, err := harness.Pool.Exec(ctx, "UPDATE incident_memberships SET role=$3 WHERE incident_id=$1 AND user_id=$2", id, actor, value); err != nil {
			t.Fatal(err)
		}
	}
	body := map[string]any{"client_txn_id": "canonical-rename", "base_table_version": 1, "display_name": "\u00a0Cafe\u0301\u3000"}
	for _, denied := range []string{"viewer", "reviewer"} {
		role(denied)
		httptestx.RequireErrorEnvelope(t, request(http.MethodPatch, body), 403, "authorization_denied")
	}
	role("editor")
	receipt := httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, body), 200)["data"].(map[string]any)
	if receipt["schema_id"] != "cartulary.network_flow_table_mutation_result.v1" {
		t.Fatal(receipt)
	}
	body["display_name"] = "Café"
	if replay := httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, body), 200)["data"]; !reflect.DeepEqual(replay, receipt) {
		t.Fatal("normalized replay changed receipt")
	}
	body["display_name"] = "different"
	httptestx.RequireErrorEnvelope(t, request(http.MethodPatch, body), 409, "client_txn_conflict")
	body["display_name"] = "Café"
	noop := map[string]any{"client_txn_id": "noop", "base_table_version": 2, "display_name": "  Cafe\u0301  "}
	if result := httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, noop), 200)["data"]; !reflect.DeepEqual(result, receipt) {
		t.Fatal("normalized no-op changed metadata")
	}
	for index, name := range []string{"\t", strings.Repeat("😀", 65)} {
		invalid := map[string]any{"client_txn_id": fmt.Sprintf("invalid-%d", index), "base_table_version": 2, "display_name": name}
		httptestx.RequireErrorEnvelope(t, request(http.MethodPatch, invalid), 400, "network_flow_invalid_display_name")
	}
	valid := map[string]any{"client_txn_id": "scalar-boundary", "base_table_version": 2, "display_name": strings.Repeat("e\u0301", 64)}
	httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, valid), 200)
	deletion := map[string]any{"client_txn_id": "delete", "base_table_version": 3}
	for _, denied := range []string{"viewer", "editor"} {
		role(denied)
		httptestx.RequireErrorEnvelope(t, request(http.MethodDelete, deletion), 403, "authorization_denied")
	}
	role("reviewer")
	deleted := httptestx.RequireSuccessEnvelope(t, request(http.MethodDelete, deletion), 200)["data"]
	if replay := httptestx.RequireSuccessEnvelope(t, request(http.MethodDelete, deletion), 200)["data"]; !reflect.DeepEqual(replay, deleted) {
		t.Fatal("delete receipt changed")
	}
	role("editor")
	if replay := httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, body), 200)["data"]; !reflect.DeepEqual(replay, receipt) {
		t.Fatal("historical rename replay changed")
	}
	role("admin")
	// The request passes route admission then waits for the incident transaction.
	// Change membership while holding that lock; receipt replay must reauthorize.
	tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var pid int
	if err = tx.QueryRow(ctx, "SELECT pg_backend_pid() FROM incidents WHERE id=$1 FOR UPDATE", id).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	responses := make(chan *http.Response, 1)
	go func() { responses <- request(http.MethodPatch, body) }()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var blocked bool
		if err = harness.Pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE $1 = ANY(pg_blocking_pids(pid)))", pid).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("mutation never reached held incident lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if _, err = tx.Exec(ctx, "UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2", id, actor); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case response := <-responses:
		httptestx.RequireErrorEnvelope(t, response, 403, "authorization_denied")
	case <-time.After(10 * time.Second):
		t.Fatal("mutation did not settle")
	}
	role("admin")
	// Two requests admitted before either mutation can commit must allocate one
	// exact active name; normalized equivalence cannot evade transactional uniqueness.
	left := createTestTable(t, harness.Pool, store, actor, id, "race-left.csv", nil, 2)
	right := createTestTable(t, harness.Pool, store, actor, id, "race-right.csv", nil, 3)
	nameLock, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = nameLock.Rollback(ctx) }()
	if err = nameLock.QueryRow(ctx, "SELECT pg_backend_pid() FROM incidents WHERE id=$1 FOR UPDATE", id).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	nameResponses := make(chan error, 2)
	for index, target := range []string{left.TableID, right.TableID} {
		go func() {
			_, failure := store.RenameTable(ctx, nf.RenameTableParams{IncidentID: id, TableID: target, BaseTableVersion: 1, DisplayName: []string{" Shared name ", "Shared name"}[index], Now: time.Now()})
			nameResponses <- failure
		}()
	}
	deadline = time.Now().Add(10 * time.Second)
	for {
		var waiting int
		if err = harness.Pool.QueryRow(ctx, `WITH RECURSIVE blocked(pid) AS (SELECT pid FROM pg_stat_activity WHERE $1 = ANY(pg_blocking_pids(pid)) UNION SELECT a.pid FROM pg_stat_activity a JOIN blocked b ON b.pid = ANY(pg_blocking_pids(a.pid))) SELECT count(DISTINCT a.pid) FROM pg_stat_activity a JOIN blocked USING(pid)`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting >= 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("name requests did not reach incident lock: waiting=%d settled=%d", waiting, len(nameResponses))
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err = nameLock.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	accepted := 0
	for range 2 {
		select {
		case failure := <-nameResponses:
			if failure == nil {
				accepted++
			} else {
				var invalid *nf.InvalidDisplayNameError
				if !errors.As(failure, &invalid) || invalid.ReasonCode != "duplicate_display_name" {
					t.Fatalf("concurrent name rejection: %v", failure)
				}
			}
		case <-time.After(10 * time.Second):
			t.Fatal("name mutation did not settle")
		}
	}
	if accepted != 1 {
		t.Fatalf("accepted %d concurrent equivalent names", accepted)
	}
	if _, err = harness.Pool.Exec(ctx, "UPDATE incidents SET status='closed',closed_at=now() WHERE id=$1", id); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireErrorEnvelope(t, request(http.MethodPatch, body), 409, "incident_closed")
	httptestx.RequireErrorEnvelope(t, request(http.MethodDelete, deletion), 409, "incident_closed")
	httptestx.RequireErrorEnvelope(t, request(http.MethodGet, nil), 409, "incident_closed")
}
