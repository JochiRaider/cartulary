package serverprocess

import (
	"database/sql"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/processtest"
)

func assertNetworkFlowOwnedProcessCrashRecovery(t *testing.T) {
	pg, s3 := sharedProcessHarnesses(t)
	testDB := pg.PrepareIsolatedDatabaseT(t, "nf-owned-crash")
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { closeSQL(t, db) })
	bucket := bucketName("nf-owned-crash")
	t.Cleanup(func() { cleanupBucket(t, s3, bucket) })
	env := newProcessEnv(t, processEnvOptions{Database: testDB.Env(), ObjectStore: s3.Env(bucket), ConfigPath: writeConfig(t, string(fixtures.MustRead("config", "valid.toml"))), BootstrapPath: fixtures.Path("bootstrap-admin", "canonical.json"), Overrides: map[string]string{
		"CARTULARY_ENABLE_TEST_ROUTES": "1", "CARTULARY_TEST_RUNTIME_MARKER": "harness-owned", "CARTULARY_TEST_ROUTE_TOKEN": httptestx.TestRouteToken,
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED": "true", "CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE", "CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST": "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}})
	start := func() *processtest.Server {
		s := processtest.StartServer(t, processtest.ServerOptions{Env: env})
		t.Cleanup(func() { s.Stop(t) })
		s.WaitForReady(t)
		return s
	}
	s := start()
	login, actor := flowtest.ProvisionBootstrapAdmin(t, s.BaseURL)
	options := []func(*http.Request){httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value)}
	incident := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, s.BaseURL+"/api/v1/incidents", map[string]any{"client_txn_id": "crash-incident", "incident_key": "IR-NF-CRASH", "title": "Crash recovery"}, options...), http.StatusCreated)["data"].(map[string]any)["incident_id"].(string)
	table := networkflowsupport.SeedGraphSource(t, db, incident, actor)
	for i, boundary := range []string{hc.NetworkFlowFaultBoundaryWorkerBeforeHandlerStart, hc.NetworkFlowFaultBoundaryWorkerBeforeFinalCommit, hc.NetworkFlowFaultBoundaryWorkerAfterCompletedPublication} {
		var resultsBefore int
		if err := db.QueryRow(`SELECT count(*) FROM graph_projection_results`).Scan(&resultsBefore); err != nil {
			t.Fatal(err)
		}
		arm := map[string]any{"boundary": boundary, "fault_kind": hc.NetworkFlowFaultKindWorkerCrash, "consume_once": true}
		httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, s.BaseURL+"/api/v1/test/runtime/network-flow-faults", arm, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)
		body := map[string]any{"schema_id": "cartulary.network_flow.graph_view_create_request.v3", "client_txn_id": fmt.Sprintf("crash-create-%d", i), "display_name": "Crash graph", "semantic_query": map[string]any{"schema_id": "cartulary.network_flow.graph_semantic_query.v2", "selected_table_ids": []string{table}, "filters": []any{}, "time_range": map[string]any{"start_utc": fmt.Sprintf("2026-01-0%dT00:00:00Z", i+1), "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"}}}
		path := "/api/v1/incidents/" + incident + "/network-flow/graph-views"
		// The owned child may exit before its accepted response reaches the client.
		request := httptestx.NewJSONRequest(t, http.MethodPost, s.BaseURL+path, body)
		for _, option := range options {
			option(request)
		}
		response, requestErr := newProcessHTTPClient().Do(request)
		if requestErr == nil {
			response.Body.Close()
		}
		exitErr := s.WaitForExit(t)
		status, ok := exitErr.(*exec.ExitError)
		if !ok || status.ExitCode() != 86 {
			t.Fatalf("owned worker crash: %v", exitErr)
		}
		var jobID, graphID string
		if err := db.QueryRow(`SELECT latest_job_id::text,graph_view_id FROM network_flow_graph_views WHERE incident_id::text=$1 ORDER BY created_at DESC,graph_view_id DESC LIMIT 1`, incident).Scan(&jobID, &graphID); err != nil {
			t.Fatal(err)
		}
		var beforeStatus string
		if err := db.QueryRow(`SELECT status FROM jobs WHERE job_id::text=$1`, jobID).Scan(&beforeStatus); err != nil {
			t.Fatal(err)
		}
		if i == 2 && beforeStatus != "succeeded" {
			t.Fatalf("postpublication crash lost completed job: %s", beforeStatus)
		}
		var resultsAfter int
		if err := db.QueryRow(`SELECT count(*) FROM graph_projection_results`).Scan(&resultsAfter); err != nil {
			t.Fatal(err)
		}
		expected := resultsBefore
		if i == 2 {
			expected++
		}
		if resultsAfter != expected {
			t.Fatalf("crash split publication at %s: before=%d after=%d", boundary, resultsBefore, resultsAfter)
		}
		s = start()
		flowtest.SetClockOffset(t, s.BaseURL, int64((i+1)*120))
		deadline := time.Now().Add(20 * time.Second)
		terminal := false
		for time.Now().Before(deadline) {
			result := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodGet, s.BaseURL+"/api/v1/jobs/"+jobID, nil, options...), http.StatusOK)["data"].(map[string]any)
			if result["status"] == "succeeded" {
				terminal = true
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if !terminal {
			var retained string
			_ = db.QueryRow(`SELECT jsonb_build_object('status',status,'failures',handler_failure_count,'lease',handler_lease_expires_at,'retry',handler_next_attempt_at,'error',error_summary_json)::text FROM jobs WHERE job_id::text=$1`, jobID).Scan(&retained)
			t.Fatalf("replacement process did not recover job at %s: %s", boundary, retained)
		}
		replay := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, s.BaseURL+path, body, options...), http.StatusAccepted)["data"].(map[string]any)
		if replay["job"].(map[string]any)["job_id"] != jobID || replay["graph_view"].(map[string]any)["graph_view_id"] != graphID {
			t.Fatal("crash recovery changed receipt")
		}
		if err := db.QueryRow(`SELECT count(*) FROM graph_projection_results`).Scan(&resultsAfter); err != nil {
			t.Fatal(err)
		}
		if resultsAfter != resultsBefore+1 {
			t.Fatal("crash replay duplicated or lost publication")
		}
		// Replacement starts with a fresh control registry; recovery cannot crash again.
	}
}
