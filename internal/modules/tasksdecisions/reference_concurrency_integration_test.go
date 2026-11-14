package tasksdecisions_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestTaskDirectReferenceConcurrency_Integration(t *testing.T) {
	h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "task-reference-admission", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "task-reference-incident", "incident_key": "IR-TASK-REFERENCES", "title": "Task references"})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	options := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
	const view = "cartulary.view.task_requests.v1"
	createURL := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + view + "/rows"
	for _, mode := range []string{"create", "patch"} {
		t.Run("overlapping Decision references/"+mode, func(t *testing.T) {
			exerciseTaskDecisionWriterOverlap(t, h, incidentID, actor, createURL, mode, options)
		})
	}
	for _, mode := range []string{"create", "patch"} {
		for _, referenceType := range []string{"party", "decision"} {
			for _, scenario := range []string{"deletion wait", "missing source"} {
				t.Run(mode+"/"+referenceType+"/"+scenario, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
					defer cancel()
					id := uuid.New()
					field := "task.requester_party_id"
					if referenceType == "decision" {
						field = "task.decision_record_id"
					}
					envelopetest.SeedRecordEnvelope(t, h.DB, incidentID, actor, id, referenceType)
					exec := func(query string, args ...any) {
						t.Helper()
						if _, err := h.DB.ExecContext(ctx, query, args...); err != nil {
							t.Fatal(err)
						}
					}
					seedSource := func() {
						t.Helper()
						if referenceType == "party" {
							exec(`INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Requester','organization')`, id, incidentID)
						} else {
							exec(`INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale) VALUES($1,$2,'Decision','approved',$3,'containment',now(),'Rationale')`, id, incidentID, actor)
						}
					}
					if scenario != "missing source" {
						seedSource()
					}
					method, url, wantStatus, wantVersion := http.MethodPost, createURL, http.StatusCreated, float64(1)
					body := map[string]any{"client_txn_id": uuid.NewString(), "task.title": "Reference task", "task.task_kind": "collection", "task.requester_party_text": "Preserved requester", field: id.String()}
					if mode == "patch" {
						created := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, createURL, map[string]any{"client_txn_id": uuid.NewString(), "task.title": "Reference task", "task.task_kind": "collection", "task.requester_party_text": "Preserved requester"}, options...), http.StatusCreated)
						method, url, wantStatus, wantVersion = http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+created["row"].(map[string]any)["record_id"].(string), http.StatusOK, 2
						body = map[string]any{"client_txn_id": uuid.NewString(), "view_schema_id": view, "base_row_version": 1, "changes": []map[string]any{{"field_key": field, "value": id.String()}}}
					}
					state := func() string {
						t.Helper()
						values := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incidentID)}
						for _, table := range []string{"records", "task_requests", "decisions", "parties", "record_links", "task_request_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
							var value string
							if err := h.DB.QueryRowContext(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
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
					before := state()
					var response *http.Response
					if scenario == "missing source" {
						response = httptestx.DoJSON(t, http.DefaultClient, method, url, body, options...)
					} else {
						barrier, err := h.Pool.Begin(ctx)
						if err != nil {
							t.Fatal(err)
						}
						defer barrier.Rollback(context.Background())
						var holder int32
						if err := barrier.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&holder); err != nil {
							t.Fatal(err)
						}
						if _, err := barrier.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2,row_version=2 WHERE record_id=$1`, id, actor); err != nil {
							t.Fatal(err)
						}
						payload, err := json.Marshal(body)
						if err != nil {
							t.Fatal(err)
						}
						request, err := http.NewRequestWithContext(ctx, method, url, strings.NewReader(string(payload)))
						if err != nil {
							t.Fatal(err)
						}
						request.Header.Set("Content-Type", "application/json")
						for _, option := range options {
							option(request)
						}
						type outcome struct {
							response *http.Response
							err      error
						}
						finished := make(chan outcome, 1)
						go func() { r, err := http.DefaultClient.Do(request); finished <- outcome{r, err} }()
						consumed := false
						t.Cleanup(func() {
							cancel()
							if !consumed {
								select {
								case r := <-finished:
									if r.response != nil {
										_ = r.response.Body.Close()
									}
								case <-time.After(5 * time.Second):
									t.Error("Task reference HTTP worker did not drain")
								}
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
								t.Fatalf("Task reference write did not wait on target deletion: %v", r.err)
							case <-ctx.Done():
								t.Fatal("Task reference lock not reached", ctx.Err())
							case <-time.After(10 * time.Millisecond):
							}
						}
						if err := barrier.Commit(ctx); err != nil {
							t.Fatal(err)
						}
						select {
						case r := <-finished:
							consumed = true
							if r.err != nil {
								t.Fatal(r.err)
							}
							response = r.response
						case <-ctx.Done():
							t.Fatal("Task reference write did not drain", ctx.Err())
						}
					}
					failure := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")["error"].(map[string]any)
					details := failure["details"].(map[string]any)
					if details["field"] != field || details["reason_code"] != "invalid_value" {
						t.Fatalf("reference admission details=%#v", failure)
					}
					if scenario == "deletion wait" {
						// Undo only the barrier fixture's synthetic deletion before comparing all effects.
						exec(`UPDATE records SET deleted_at=NULL,deleted_by_user_id=NULL,row_version=1 WHERE record_id=$1`, id)
					}
					if state() != before {
						t.Fatal("rejected reference write changed source/history/projection/receipt/publication")
					}
					if scenario == "missing source" {
						seedSource()
					}
					success := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, method, url, body, options...), wantStatus)
					row := success["row"].(map[string]any)
					cells := row["cells"].(map[string]any)
					if row["row_version"] != wantVersion || cells[field].(map[string]any)["value"] != id.String() || cells["task.requester_party_text"].(map[string]any)["value"] != "Preserved requester" {
						t.Fatalf("same-request reference retry=%#v", row)
					}
				})
			}
		}
	}
}
