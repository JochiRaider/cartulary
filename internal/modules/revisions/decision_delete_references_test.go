package revisions_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func requireDecisionDeleteTaskReferences(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	for _, variant := range []string{"active scalar", "materialized binding", "done", "canceled", "deleted Task", "null reference", "committed during wait"} {
		t.Run("Decision delete Task references/"+variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			decision, task := uuid.New(), uuid.New()
			envelopetest.SeedRecordEnvelope(t, h.DB, incident, actor, decision, "decision")
			envelopetest.SeedRecordEnvelope(t, h.DB, incident, actor, task, "task_request")
			mustExec(t, h.DB, `INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale) VALUES($1,$2,'Retained decision','approved',$3,'containment',now(),'Rationale')`, decision, incident, actor)
			var reference *uuid.UUID
			if variant != "null reference" && variant != "committed during wait" {
				reference = &decision
			}
			status := "open"
			if variant == "done" || variant == "canceled" {
				status = variant
			}
			mustExec(t, h.DB, `INSERT INTO task_requests(record_id,incident_id,title,status,owner_user_id,priority,task_kind,decision_record_id,completed_at) VALUES($1,$2,'Referencing Task',$3,$4,'normal','collection',$5,CASE WHEN $3='done' THEN now() ELSE NULL END)`, task, incident, status, actor, reference)
			if variant == "materialized binding" {
				mustExec(t, h.DB, `INSERT INTO record_links(incident_id,src_record_id,dst_record_id,link_type,field_key,provenance,owner_user_id,created_by_user_id,decided_at) VALUES($1,$2,$3,'references_record','task.decision_record_id','manual',$4,$4,now())`, incident, task, decision, actor)
			}
			if variant == "deleted Task" {
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, task, map[string]any{"base_row_version": 1, "client_txn_id": uuid.NewString()}), http.StatusOK)
			}
			body := map[string]any{"base_row_version": 1, "client_txn_id": uuid.NewString()}
			before := snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow)
			var response *http.Response
			if variant == "committed during wait" {
				response, before = deleteDecisionAfterTaskReferenceCommit(t, ctx, cancel, h, login, incident, task, decision, body)
			} else {
				response = deleteRecord(t, h, login, decision, body)
			}
			if variant != "deleted Task" && variant != "null reference" {
				failure := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "record_delete_blocked")["error"].(map[string]any)
				details := failure["details"].(map[string]any)
				if failure["retryable"] != false || len(details) != 2 || details["record_id"] != decision.String() || details["reason_code"] != "active_incoming_task_decision_reference" {
					t.Fatalf("unsafe Decision blocker=%#v", failure)
				}
				if after := snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow); after != before {
					t.Fatal("blocked Decision delete changed retained state or publication")
				}
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, task, map[string]any{"base_row_version": 1, "client_txn_id": uuid.NewString()}), http.StatusOK)
				response = deleteRecord(t, h, login, decision, body)
			}
			deleted := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			if deleted["deleted"] != true || deleted["row_version"] != float64(2) {
				t.Fatalf("Decision delete result=%#v", deleted)
			}
			if countRecordMutations(t, h.DB, decision, "soft_delete") != 1 || countRecordRevisions(t, h.DB, decision) != 1 {
				t.Fatal("Decision delete did not append exactly one transition")
			}
			requireSingleRecordChangeIntent(t, h.DB, deleted["change_set_id"].(string), decision, "remove")
			if countRows(t, h.DB, `SELECT count(*) FROM decisions WHERE record_id=$1 AND summary='Retained decision' AND status='approved' AND rationale='Rationale'`, decision) != 1 {
				t.Fatal("Decision deletion changed retained source")
			}
			beforeReplay := snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow)
			replay := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, decision, body), http.StatusOK)["data"].(map[string]any)
			if replay["change_set_id"] != deleted["change_set_id"] || replay["row_version"] != float64(2) || snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow) != beforeReplay {
				t.Fatal("exact Decision delete replay changed committed state")
			}
		})
	}
}

func snapshotDecisionReferenceState(t testing.TB, h *appsupport.ServerHarness, incident uuid.UUID, query func(context.Context, string, ...any) pgx.Row) string {
	t.Helper()
	values := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incident)}
	for _, table := range []string{"records", "task_requests", "decisions", "record_links", "task_request_grid_projection", "decision_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
		var value string
		if err := query(t.Context(), "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
			t.Fatal(err)
		}
		values[table] = value
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func deleteDecisionAfterTaskReferenceCommit(t *testing.T, ctx context.Context, cancel context.CancelFunc, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, task, decision uuid.UUID, body map[string]any) (*http.Response, string) {
	t.Helper()
	barrier, err := h.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer barrier.Rollback(context.Background())
	var holder int32
	if err := barrier.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&holder); err != nil {
		t.Fatal(err)
	}
	if _, err := barrier.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR SHARE`, decision); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, h.Server.HTTP.URL+"/api/v1/records/"+decision.String(), strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	type result struct {
		response *http.Response
		err      error
	}
	finished := make(chan result, 1)
	go func() { r, err := http.DefaultClient.Do(request); finished <- result{r, err} }()
	consumed := false
	t.Cleanup(func() {
		cancel()
		if consumed {
			return
		}
		select {
		case r := <-finished:
			if r.response != nil {
				_ = r.response.Body.Close()
			}
		case <-time.After(5 * time.Second):
			t.Error("Decision delete worker did not drain")
		}
	})
	for {
		var waiting bool
		if err := h.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, holder).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case r := <-finished:
			consumed = true
			if r.response != nil {
				_ = r.response.Body.Close()
			}
			t.Fatalf("Decision delete returned before waiting: %v", r.err)
		case <-ctx.Done():
			t.Fatal("Decision delete never waited", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if _, err := barrier.Exec(ctx, `UPDATE task_requests SET decision_record_id=$2 WHERE record_id=$1`, task, decision); err != nil {
		t.Fatal(err)
	}
	expected := snapshotDecisionReferenceState(t, h, incident, barrier.QueryRow)
	if err := barrier.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case r := <-finished:
		consumed = true
		if r.err != nil {
			t.Fatal(r.err)
		}
		return r.response, expected
	case <-ctx.Done():
		t.Fatal("Decision delete never finished", ctx.Err())
	}
	return nil, ""
}
