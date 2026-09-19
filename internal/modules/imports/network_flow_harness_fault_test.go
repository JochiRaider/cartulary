package imports_test

import (
	"context"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func assertNetworkFlowImportFaultConsumers(t *testing.T) {
	controls := hc.NewControls()
	runtime := appsupport.StartRuntime(t)
	h := runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: "network-flow-functional-import-faults",
		Env: map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
		},
		TestRouteMode:    httptestx.TestRouteModeHarnessOwned,
		AdditionalRoutes: controls.Contribution().Routes,
		ConfigureRuntime: func(o *server.Options) { o.NetworkFlowComposition = controls },
	})
	httptestx.SetClockFixed(t, h.Server, time.Now())
	login, actorText := flowtest.ProvisionBootstrapAdmin(t, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "fault-incident", "incident_key": "IR-NF-FAULTS", "title": "Fault controls"})["incident_id"].(string)
	const csv = "Source IP Address,Destination IP Address,Source Port,Destination Port,Protocol,Bytes,Packets,Flow Start Time,Flow End Time,Input Interface,Output Interface\n192.0.2.10,192.0.2.20,443,51515,TCP,1234,12,2026-07-10T12:00:00Z,2026-07-10T12:00:05Z,,\n"
	for i, boundary := range []string{hc.NetworkFlowFaultBoundaryImportBeforeOwnerApply, hc.NetworkFlowFaultBoundaryImportAfterOwnerApply, hc.NetworkFlowFaultBoundaryImportBeforeTransactionCommit, hc.NetworkFlowFaultBoundaryImportAfterTransactionCommitBeforeReply} {
		kinds := []string{hc.NetworkFlowFaultKindReturnError}
		if i < 3 {
			kinds = append(kinds, hc.NetworkFlowFaultKindPanic, hc.NetworkFlowFaultKindCancelContext)
		}
		for j, kind := range kinds {
			t.Run(fmt.Sprintf("fault/%d/%s", i, kind), func(t *testing.T) {
				txn := fmt.Sprintf("fault-%d-%d", i, j)
				session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, incident, txn+"-upload", csv, "fault.csv")
				httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, "/api/v1/import-sessions/"+session+"/units/"+unit+"/mapping", networkFlowMappingPayload(txn+"-mapping")), http.StatusOK)
				httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, "/api/v1/import-sessions/"+session+"/units/"+unit+"/select", map[string]any{"client_txn_id": txn + "-select"}), http.StatusOK)
				body := map[string]any{"boundary": boundary, "fault_kind": kind, "consume_once": true, "correlation_key": unit}
				if kind == hc.NetworkFlowFaultKindReturnError {
					body["error_code"] = "harness_import_failure"
				}
				arm := httptestx.DoJSON(t, http.MethodPost, h.Server.HTTP.URL+"/api/v1/test/runtime/network-flow-faults", body, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken))
				httptestx.RequireSuccessEnvelope(t, arm, http.StatusCreated)
				auditBefore := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind='network_flow_table_created'`)
				applyBody := map[string]any{"client_txn_id": txn + "-apply"}
				auditBinding := networkflowsupport.AuditBinding{Scope: hc.NetworkFlowAuditScope{EventCode: hc.NetworkFlowAuditEventTableCreated, OperationRef: txn, ActorRef: "actor:import", IncidentRef: "incident:import", ResourceKind: hc.NetworkFlowAuditResourceTable, ResourceRef: "table:" + unit, CorrelationKey: unit}, ActorID: uuid.MustParse(actorText), IncidentID: uuid.MustParse(incident), ClientTxnID: txn + "-apply"}
				auditFixture, err := networkflowsupport.NewAuditFixture(context.Background(), controls.AuditAssertions, h.Pool, auditBinding)
				if err != nil {
					t.Fatal(err)
				}
				expectedAudit := 0
				if i == 3 || kind == hc.NetworkFlowFaultKindPanic || i == 2 {
					expectedAudit = 1
				}
				scope := auditBinding.Scope
				httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, h.Server.HTTP.URL+"/api/v1/test/runtime/network-flow-audit-assertions", map[string]any{"assertion_kind": hc.NetworkFlowAuditAssertionNoAuditReplay, "event_code": scope.EventCode, "operation_ref": scope.OperationRef, "actor_ref": scope.ActorRef, "incident_ref": scope.IncidentRef, "resource_kind": scope.ResourceKind, "resource_ref": scope.ResourceRef, "baseline_count": 0, "expected_final_count": expectedAudit, "correlation_key": scope.CorrelationKey, "consume_once": true}, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)

				path := "/api/v1/import-sessions/" + session + "/apply"
				accepted := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, path, applyBody), http.StatusAccepted)["data"].(map[string]any)
				jobID := accepted["job_id"].(string)
				failures := 0
				observedStatus := ""
				deadline := time.Now().Add(4 * time.Second)
				for time.Now().Before(deadline) {
					if err := h.Pool.QueryRow(context.Background(), `SELECT status,handler_failure_count FROM jobs WHERE job_id=$1`, jobID).Scan(&observedStatus, &failures); err != nil {
						t.Fatal(err)
					}
					if failures > 0 || observedStatus == "failed" || observedStatus == "succeeded" {
						break
					}
					time.Sleep(10 * time.Millisecond)
				}
				if failures > 0 {
					if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM network_flow_tables WHERE source_import_unit_id::text=$1`, unit); got != 0 {
						t.Fatal("failed attempt committed owner effects")
					}
					if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind='network_flow_table_created'`) - auditBefore; got != 0 {
						t.Fatal("failed attempt committed audit")
					}
					httptestx.SetClockFixed(t, h.Server, h.Server.Clock.Now().Add(time.Minute))
				}
				// Recovery scans use the real Jobs scheduler; wait across its five-second cadence.
				var job map[string]any
				deadline = time.Now().Add(12 * time.Second)
				for time.Now().Before(deadline) {
					job = httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodGet, h.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie)), http.StatusOK)["data"].(map[string]any)
					if job["status"] == "failed" || job["status"] == "succeeded" {
						break
					}
					time.Sleep(25 * time.Millisecond)
				}
				if _, pending := controls.Faults.ConsumeNetworkFlowFaultFor(boundary, unit); pending {
					t.Fatal("real import failed to consume its fault")
				}
				want := 0
				status := "failed"
				if i == 3 || failures > 0 {
					want = 1
					status = "succeeded"
				}
				if job["status"] != status {
					t.Fatalf("job outcome: %#v", job)
				}
				if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM network_flow_tables WHERE source_import_unit_id::text=$1`, unit); got != want {
					t.Fatalf("partial/absent table: %d", got)
				}
				if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM import_unit_apply_outcomes WHERE import_unit_id::text=$1 AND outcome_status='applied'`, unit); got != want {
					t.Fatalf("outcome count: %d", got)
				}
				if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind='network_flow_table_created'`) - auditBefore; got != want {
					t.Fatalf("committed audit delta: %d", got)
				}
				replay := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, path, applyBody), http.StatusAccepted)["data"].(map[string]any)
				if replay["job_id"] != accepted["job_id"] {
					t.Fatal("replay changed immutable job")
				}
				if got := dbassert.CountSQL(t, h.DB, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind='network_flow_table_created'`) - auditBefore; got != want {
					t.Fatal("replay appended audit")
				}
				resourceID := ""
				if want == 1 {
					if err := h.Pool.QueryRow(context.Background(), `SELECT network_flow_table_id FROM network_flow_tables WHERE source_import_unit_id::text=$1`, unit).Scan(&resourceID); err != nil {
						t.Fatal(err)
					}
				}
				if err := auditFixture.Verify(context.Background(), resourceID, func() error {
					httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, path, applyBody), http.StatusAccepted)
					return nil
				}); err != nil {
					t.Fatal(err)
				}

			})
		}
	}
	if err := controls.AuditAssertions.RequireConsumed(); err != nil {
		t.Fatal(err)
	}
}
