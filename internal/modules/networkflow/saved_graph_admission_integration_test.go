package networkflow_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSavedGraphCutoverAdmission_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "network-flow-saved-graph-cutover")
	adminLogin, adminIDText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	adminID := uuid.MustParse(adminIDText)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-saved-graph-incident",
		"incident_key":  "IR-NF-SAVED-GRAPH",
		"title":         "Network Flow saved graph",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	store := newTestNetworkFlowStore(t, harness.Pool, harness.Revisions.Appender())
	sessionID, unitID := seedImportSessionUnit(t, harness.Pool, incidentID, adminID, "saved-graph.csv")
	table, err := store.CreateTable(context.Background(), CreateTableParams{
		IncidentID: incidentID, ActorUserID: adminID, ImportSessionID: sessionID, ImportUnitID: unitID,
		SourceContentSHA256: testSHA1, OriginalFilename: "saved-graph.csv",
		SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "route-test-key",
		MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{testFlowRow(1, "a")},
		Now: time.Date(2026, 7, 10, 12, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create saved-graph source table: %v", err)
	}

	collectionPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graph-views"
	createBody := map[string]any{
		"schema_id":     "cartulary.network_flow.graph_view_create_request.v3",
		"client_txn_id": "txn-network-flow-graph-create",
		"display_name":  "Shared flow graph",
		"semantic_query": map[string]any{
			"schema_id":          "cartulary.network_flow.graph_semantic_query.v2",
			"selected_table_ids": []string{table.TableID}, "filters": []any{},
			"time_range":  map[string]any{"start_utc": nil, "end_utc": nil},
			"aggregation": map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true},
		},
	}
	mutationOptions := []func(*http.Request){
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	}
	createResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	created := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusAccepted)["data"].(map[string]any)
	graphView := created["graph_view"].(map[string]any)
	graphViewID := graphView["graph_view_id"].(string)
	jobID := created["job"].(map[string]any)["job_id"].(string)
	if graphView["graph_view_version"] != float64(1) || graphView["materialization_generation"] != float64(1) || graphView["latest_job_id"] != jobID || created["job"].(map[string]any)["status_route"] != "/api/v1/jobs/"+jobID {
		t.Fatalf("unexpected created graph view: %#v", graphView)
	}

	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, jobID, "succeeded")
	definitions := appsupport.RecognizedExtensionJobDefinitions(t)
	var definition jobs.Definition
	for _, candidate := range definitions {
		if candidate.JobKind == GraphViewMaterializationJobKind {
			definition = candidate
		}
	}
	physical, err := extensionstore.New(harness.Pool, ExtensionStateFamilyCounters())
	if err != nil {
		t.Fatal(err)
	}
	check := func(wantError bool) {
		t.Helper()
		before := savedGraphCutoverBytes(t, harness.Pool)
		err := physical.WithAdmissionRead(context.Background(), func(reader extensionstore.Querier) error {
			return ValidateSavedGraphAdmission(context.Background(), reader, definition)
		})
		if (err != nil) != wantError {
			t.Fatalf("cutover error=%v want rejected=%v", err, wantError)
		}
		if wantError && !errors.Is(err, ErrSavedGraphCutoverIncompatible) {
			t.Fatalf("unsafe or unexpected failure: %v", err)
		}
		after := savedGraphCutoverBytes(t, harness.Pool)
		if !reflect.DeepEqual(before, after) {
			t.Fatal("read-only admission changed retained state")
		}
	}
	if _, err := harness.Pool.Exec(context.Background(), `INSERT INTO graph_projection_result_leases (lease_id,projection_result_id,lease_owner_id,lease_owner_resource_id,lease_purpose,leased_until,created_at,renewed_at) SELECT $1,selected_projection_result_id,'snapshot_reporting','cutover-proof','render',$2,$3,$3 FROM network_flow_graph_views WHERE graph_view_id=$4`, uuid.New(), time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC), graphViewID); err != nil {
		t.Fatal(err)
	}
	check(false)
	// Failed fixtures are installed explicitly by the test and restored after
	// checking. The admission algorithm itself never receives a write capability.
	for _, fixture := range []struct {
		name, change, restore string
		args                  []any
	}{
		{"old receipt", `UPDATE route_idempotency SET response_json=jsonb_set(response_json,'{schema_id}','"cartulary.network_flow.graph_view_accepted.v3"') WHERE client_txn_id=$1`, `UPDATE route_idempotency SET response_json=$2 WHERE client_txn_id=$1`, []any{createBody["client_txn_id"], mustJSON(t, created)}},
		{"invalid UTF8 byte name", `UPDATE network_flow_graph_views SET display_name=$2 WHERE graph_view_id=$1`, `UPDATE network_flow_graph_views SET display_name='Shared flow graph' WHERE graph_view_id=$1`, []any{graphViewID, strings.Repeat("é", 33)}},
		{"binding differs from result", `UPDATE network_flow_graph_views SET selected_canonical_output_sha256=repeat('a',64) WHERE graph_view_id=$1`, `UPDATE network_flow_graph_views SET selected_canonical_output_sha256=(SELECT canonical_output_sha256 FROM graph_projection_results WHERE projection_result_id=selected_projection_result_id) WHERE graph_view_id=$1`, []any{graphViewID}},
		{"invalid failure", `UPDATE network_flow_graph_views SET last_failure_code='network_flow_unsupported_failure',last_failed_at=updated_at WHERE graph_view_id=$1`, `UPDATE network_flow_graph_views SET last_failure_code=NULL,last_failed_at=NULL WHERE graph_view_id=$1`, []any{graphViewID}},
		{"contradictory worker", `UPDATE jobs SET handler_name='network_flow_activity.unknown_worker' WHERE job_id=$1`, `UPDATE jobs SET handler_name='network_flow_activity.graph_view_worker_v1' WHERE job_id=$1`, []any{uuid.MustParse(jobID)}},
		{"invalid proof", `UPDATE extension_job_commit_proofs SET terminal_result_sha256=repeat('a',64) WHERE job_id=$1`, `UPDATE extension_job_commit_proofs SET terminal_result_sha256=$2 WHERE job_id=$1`, nil},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			if fixture.name == "invalid proof" {
				var digest string
				if err := harness.Pool.QueryRow(context.Background(), `SELECT terminal_result_sha256 FROM extension_job_commit_proofs WHERE job_id=$1`, uuid.MustParse(jobID)).Scan(&digest); err != nil {
					t.Fatal(err)
				}
				fixture.args = []any{uuid.MustParse(jobID), digest}
			}
			execCutoverFixture(t, harness.Pool, fixture.change, fixture.args)
			defer execCutoverFixture(t, harness.Pool, fixture.restore, fixture.args)
			check(true)
		})
	}
	check(false)
	resourcePath := collectionPath + "/" + graphViewID
	for index, name := range []string{"Renamed", "  Renamed  "} {
		base := 1
		if index == 1 {
			base = 2
		}
		httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPatch, resourcePath, map[string]any{"schema_id": "cartulary.network_flow.graph_view_rename_request.v2", "client_txn_id": []string{"cutover-rename", "cutover-noop"}[index], "base_graph_view_version": base, "display_name": name}, mutationOptions...), http.StatusOK)
		check(false)
	}
	refreshed := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, resourcePath+"/refresh", map[string]any{"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "cutover-refresh", "base_graph_view_version": 2}, mutationOptions...), http.StatusAccepted)["data"].(map[string]any)
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, refreshed["job"].(map[string]any)["job_id"].(string), "succeeded")
	check(false)
	httptestx.RequireStatus(t, httptestx.DoJSON(t, http.MethodDelete, resourcePath, map[string]any{"schema_id": "cartulary.network_flow.graph_view_retire_request.v1", "client_txn_id": "cutover-retire", "base_graph_view_version": 3}, mutationOptions...), http.StatusNoContent)
	check(false)
	// This is the exact owner-defined compacted tombstone shape, including erased
	// handler payload and idempotency metadata. Proofs and route receipts remain.
	if _, err := harness.Pool.Exec(context.Background(), `UPDATE jobs SET expired_at=retained_until,handler_payload_json=NULL,handler_attempt_id=NULL,handler_lease_expires_at=NULL,handler_failure_count=0,handler_next_attempt_at=NULL,handler_last_attempted_at=NULL,handler_last_error=NULL,result_summary_json=NULL,error_summary_json=NULL,message=NULL,extension_idempotency_identity=NULL,extension_idempotency_route_key=NULL,extension_idempotency_scope_key=NULL,extension_normalized_request_sha256=NULL WHERE job_id=$1`, uuid.MustParse(jobID)); err != nil {
		t.Fatal(err)
	}
	check(false)
	if _, err := harness.Pool.Exec(context.Background(), `UPDATE route_idempotency SET response_json=jsonb_set(response_json,'{schema_id}','"cartulary.network_flow.graph_view_accepted.v3"') WHERE client_txn_id=$1`, createBody["client_txn_id"]); err != nil {
		t.Fatal(err)
	}
	check(true)
}

func mustJSON(t testing.TB, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func execCutoverFixture(t testing.TB, pool *pgxpool.Pool, query string, args []any) {
	t.Helper()
	count := 1
	if strings.Contains(query, "$2") {
		count = 2
	}
	if _, err := pool.Exec(context.Background(), query, args[:count]...); err != nil {
		t.Fatal(err)
	}
}
func savedGraphCutoverBytes(t testing.TB, pool *pgxpool.Pool) map[string]string {
	t.Helper()
	result := map[string]string{}
	for _, table := range []string{"network_flow_graph_views", "graph_projection_results", "graph_projection_result_vertices", "graph_projection_result_edges", "graph_projection_result_leases", "route_idempotency", "jobs", "extension_job_commit_proofs", "extension_state_metadata", "extension_migration_ledger"} {
		var raw string
		query := `SELECT COALESCE(jsonb_agg(row ORDER BY row::text),'[]'::jsonb)::text FROM (SELECT to_jsonb(t) AS row FROM ` + pgx.Identifier{table}.Sanitize() + ` t) rows`
		if err := pool.QueryRow(context.Background(), query).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		result[table] = raw
	}
	return result
}
