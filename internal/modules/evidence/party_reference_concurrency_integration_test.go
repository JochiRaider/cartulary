package evidence_test

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
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestEvidencePartyReferenceConcurrency_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-party-reference-locks", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "reference-lock-incident", "incident_key": "REFERENCE-LOCK", "title": "Reference locks"})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	const view = "cartulary.view.evidence.v1"
	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + view + "/rows"
	for _, mode := range []string{"create", "patch"} {
		for _, fields := range [][]string{{"evidence.collector_party_id"}, {"evidence.source_party_id"}, {"evidence.collector_party_id", "evidence.source_party_id"}} {
			t.Run(mode+"/"+strings.Join(fields, "+"), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				mustExec := func(query string, args ...any) {
					t.Helper()
					if _, err := harness.DB.ExecContext(ctx, query, args...); err != nil {
						t.Fatal(err)
					}
				}
				refs := make(map[string]uuid.UUID)
				for _, field := range fields {
					id := uuid.New()
					refs[field] = id
					envelopetest.SeedRecordEnvelope(t, harness.DB, incidentID, actor, id, "party")
					mustExec(`INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Reference target','organization')`, id, incidentID)
				}
				party := refs[fields[0]]
				method, url := http.MethodPost, createURL
				body := map[string]any{"client_txn_id": uuid.NewString(), "evidence.title": "Reference write"}
				wantStatus, wantVersion := http.StatusCreated, float64(1)
				if mode == "patch" {
					created := requireHTTPWorkbookCreate(t, harness, login, incidentID, view, map[string]any{"client_txn_id": uuid.NewString(), "evidence.title": "Reference write"})
					method, url = http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+created["row"].(map[string]any)["record_id"].(string)
					changes := make([]map[string]any, 0, len(fields))
					for _, field := range fields {
						changes = append(changes, map[string]any{"field_key": field, "value": refs[field].String()})
					}
					body = map[string]any{"view_schema_id": view, "base_row_version": 1, "client_txn_id": uuid.NewString(), "changes": changes}
					wantStatus, wantVersion = http.StatusOK, 2
				} else {
					for _, field := range fields {
						body[field] = refs[field].String()
					}
				}
				before := snapshotEvidenceContractState(t, harness, incidentID)
				projectionState := func() string {
					t.Helper()
					var state string
					if err := harness.DB.QueryRowContext(ctx, `SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY record_id),'[]')::text FROM evidence_grid_projection p WHERE incident_id=$1`, incidentID).Scan(&state); err != nil {
						t.Fatal(err)
					}
					return state
				}
				beforeProjection := projectionState()
				barrier, err := harness.Pool.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer barrier.Rollback(context.Background())
				var holderPID int32
				if err := barrier.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&holderPID); err != nil {
					t.Fatal(err)
				}
				if _, err := barrier.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2,row_version=2 WHERE record_id=$1`, party, actor); err != nil {
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
				request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
				request.AddCookie(login.Session)
				request.AddCookie(login.CSRF)
				type result struct {
					response *http.Response
					err      error
				}
				finished := make(chan result, 1)
				go func() {
					response, err := http.DefaultClient.Do(request)
					finished <- result{response, err}
				}()
				consumed := false
				t.Cleanup(func() {
					cancel()
					if !consumed {
						select {
						case outcome := <-finished:
							if outcome.response != nil {
								_ = outcome.response.Body.Close()
							}
						case <-time.After(5 * time.Second):
							t.Error("reference HTTP worker did not drain")
						}
					}
				})
				for {
					var waiting bool
					if err := harness.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, holderPID).Scan(&waiting); err != nil {
						t.Fatal(err)
					}
					if waiting {
						break
					}
					select {
					case outcome := <-finished:
						consumed = true
						if outcome.response != nil {
							_ = outcome.response.Body.Close()
						}
						t.Fatalf("reference write did not wait on the held Party: %v", outcome.err)
					case <-ctx.Done():
						t.Fatal("reference write did not reach Party lock", ctx.Err())
					case <-time.After(10 * time.Millisecond):
					}
				}
				if err := barrier.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				var outcome result
				select {
				case outcome = <-finished:
					consumed = true
				case <-ctx.Done():
					t.Fatal("reference write did not finish after Party deletion", ctx.Err())
				}
				if outcome.err != nil {
					t.Fatal(outcome.err)
				}
				failure := httptestx.RequireErrorEnvelope(t, outcome.response, http.StatusBadRequest, "invalid_mutation_payload")["error"].(map[string]any)
				details := failure["details"].(map[string]any)
				if details["field"] != fields[0] || details["reason_code"] != "invalid_value" {
					t.Fatalf("current Party rejection=%#v", failure)
				}
				// Remove only the barrier's synthetic tombstone before comparing all
				// retained source/envelope/history/receipt/publication state.
				mustExec(`UPDATE records SET deleted_at=NULL,deleted_by_user_id=NULL,row_version=1 WHERE record_id=$1`, party)
				if snapshotEvidenceContractState(t, harness, incidentID) != before || projectionState() != beforeProjection {
					t.Fatal("rejected reference write left durable state")
				}
				success := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, method, url, body, authOptions(login)...), wantStatus)
				row := success["row"].(map[string]any)
				if row["row_version"] != wantVersion {
					t.Fatalf("reference retry version=%v want=%v", row["row_version"], wantVersion)
				}
				for _, field := range fields {
					if row["cells"].(map[string]any)[field].(map[string]any)["value"] != refs[field].String() {
						t.Fatalf("reference retry %s=%#v", field, row)
					}
				}
				before = snapshotEvidenceContractState(t, harness, incidentID)
				beforeProjection = projectionState()
				deleteURL := harness.Server.HTTP.URL + "/api/v1/records/" + party.String()
				blocked := httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, deleteURL, map[string]any{"base_row_version": 1, "client_txn_id": uuid.NewString()}, authOptions(login)...), http.StatusConflict, "record_delete_blocked")["error"].(map[string]any)
				if blocked["details"].(map[string]any)["reason_code"] != "active_incoming_party_reference" || snapshotEvidenceContractState(t, harness, incidentID) != before || projectionState() != beforeProjection {
					t.Fatalf("committed reference did not protect Party atomically: %#v", blocked)
				}
			})
		}
	}
}
