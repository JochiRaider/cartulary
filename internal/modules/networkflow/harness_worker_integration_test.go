package networkflow_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func assertNetworkFlowWorkerFaultConsumers(t *testing.T) {
	controls := hc.NewControls()
	h := claimedNetworkFlowServerWithControlsForRouteTest(t, appsupport.StartRuntime(t), "network-flow-worker-faults", "", controls)
	httptestx.SetClockFixed(t, h.Server, time.Now())
	login, actorText := flowtest.ProvisionBootstrapAdmin(t, h.Server.HTTP.URL)
	actor := uuid.MustParse(actorText)
	incidentText := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "worker-fault-incident", "incident_key": "IR-NF-WORKER-FAULTS", "title": "Worker faults"})["incident_id"].(string)
	incident := uuid.MustParse(incidentText)
	session, unit := seedImportSessionUnit(t, h.Pool, incident, actor, "fault.csv")
	store := newTestNetworkFlowStore(t, h.Pool, h.Revisions.Appender())
	table, err := store.CreateTable(context.Background(), CreateTableParams{IncidentID: incident, ActorUserID: actor, ImportSessionID: session, ImportUnitID: unit, SourceContentSHA256: testSHA1, OriginalFilename: "fault.csv", SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "test-key", MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV, ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{testFlowRow(1, "a")}, Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	options := []func(*http.Request){httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value)}
	for i, boundary := range []string{hc.NetworkFlowFaultBoundaryWorkerBeforeHandlerStart, hc.NetworkFlowFaultBoundaryWorkerBeforeCancellationCheck, hc.NetworkFlowFaultBoundaryWorkerBeforeFinalCommit, hc.NetworkFlowFaultBoundaryWorkerAfterCompletedPublication} {
		kinds := []string{hc.NetworkFlowFaultKindReturnError}
		if i < 3 {
			kinds = append(kinds, hc.NetworkFlowFaultKindPanic, hc.NetworkFlowFaultKindCancelContext)
		}
		if i == 1 {
			kinds = append(kinds, hc.NetworkFlowFaultKindWorkerCancel)
		}
		for j, kind := range kinds {
			t.Run(fmt.Sprintf("%d/%s", i, kind), func(t *testing.T) {
				arm := map[string]any{"boundary": boundary, "fault_kind": kind, "consume_once": true}
				if kind == hc.NetworkFlowFaultKindReturnError {
					arm["error_code"] = "worker_fault_probe"
				}
				httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, h.Server.HTTP.URL+"/api/v1/test/runtime/network-flow-faults", arm, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)
				body := map[string]any{"schema_id": "cartulary.network_flow.graph_view_create_request.v3", "client_txn_id": fmt.Sprintf("worker-fault-%d-%d", i, j), "display_name": "Fault graph", "semantic_query": map[string]any{"schema_id": "cartulary.network_flow.graph_semantic_query.v2", "selected_table_ids": []string{table.TableID}, "filters": []any{}, "time_range": map[string]any{"start_utc": nil, "end_utc": nil}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"}}}
				path := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentText + "/network-flow/graph-views"
				accepted := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, path, body, options...), http.StatusAccepted)["data"].(map[string]any)
				job := accepted["job"].(map[string]any)["job_id"].(string)
				id := accepted["graph_view"].(map[string]any)["graph_view_id"].(string)
				want := "succeeded"
				if kind == hc.NetworkFlowFaultKindWorkerCancel {
					want = "canceled"
				}
				if i < 3 && kind != hc.NetworkFlowFaultKindWorkerCancel {
					failures := 0
					deadline := time.Now().Add(4 * time.Second)
					for time.Now().Before(deadline) {
						if err := h.Pool.QueryRow(context.Background(), `SELECT handler_failure_count FROM jobs WHERE job_id=$1`, job).Scan(&failures); err != nil {
							t.Fatal(err)
						}
						if failures > 0 {
							break
						}
						time.Sleep(10 * time.Millisecond)
					}
					if failures != 1 {
						t.Fatalf("fault did not fail exactly one handler attempt: %d", failures)
					}
					resource := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodGet, path+"/"+id, nil, options...), http.StatusOK)["data"].(map[string]any)
					if resource["graph_view"].(map[string]any)["selected_result_binding"] != nil {
						t.Fatal("failed attempt published selected result")
					}
					httptestx.SetClockFixed(t, h.Server, h.Server.Clock.Now().Add(time.Minute))
				}
				waitForNetworkFlowJob(t, h.Server.HTTP.URL, login, job, want)
				if _, pending := controls.Faults.ConsumeNetworkFlowFault(boundary); pending {
					t.Fatal("worker did not consume control")
				}
				resource := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodGet, path+"/"+id, nil, options...), http.StatusOK)["data"].(map[string]any)
				if want == "succeeded" && resource["graph_view"].(map[string]any)["selected_result_binding"] == nil {
					t.Fatal("recovery did not publish result")
				}
				if want == "canceled" && resource["graph_view"].(map[string]any)["selected_result_binding"] != nil {
					t.Fatal("cancellation published result")
				}

				replay := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, path, body, options...), http.StatusAccepted)["data"].(map[string]any)
				if replay["job"].(map[string]any)["job_id"] != job {
					t.Fatal("post-fault replay changed immutable job")
				}
			})
		}
	}
}
