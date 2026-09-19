package networkflow_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
)

func assertNetworkFlowCommittedAuditConsumers(t *testing.T) {
	ctx := context.Background()
	controls := hc.NewControls()
	h := claimedNetworkFlowServerWithControlsForRouteTest(t, appsupport.StartRuntime(t), "network-flow-audit-controls", "", controls)
	login, actorText := flowtest.ProvisionBootstrapAdmin(t, h.Server.HTTP.URL)
	actor := uuid.MustParse(actorText)
	incidentText := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "audit-incident", "incident_key": "IR-NF-AUDIT", "title": "Committed audit"})["incident_id"].(string)
	incident := uuid.MustParse(incidentText)
	store := newTestNetworkFlowStore(t, h.Pool, h.Revisions.Appender())
	table := createTestTable(t, h.Pool, store, actor, incident, "audit.csv", nil, 1)
	root := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentText + "/network-flow"
	options := []func(*http.Request){httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value)}
	request := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, method, root+path, body, options...)
	}
	arm := func(t *testing.T, b networkflowsupport.AuditBinding, kind string, baseline, final int) *networkflowsupport.AuditFixture {
		t.Helper()
		f, err := networkflowsupport.NewAuditFixture(ctx, controls.AuditAssertions, h.Pool, b)
		if err != nil {
			t.Fatal(err)
		}
		s := b.Scope
		httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, h.Server.HTTP.URL+"/api/v1/test/runtime/network-flow-audit-assertions", map[string]any{"assertion_kind": kind, "event_code": s.EventCode, "operation_ref": s.OperationRef, "actor_ref": s.ActorRef, "incident_ref": s.IncidentRef, "resource_kind": s.ResourceKind, "resource_ref": s.ResourceRef, "baseline_count": baseline, "expected_final_count": final, "correlation_key": s.CorrelationKey, "consume_once": true}, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)
		return f
	}
	binding := func(event, operation, kind, resource string) networkflowsupport.AuditBinding {
		b := networkflowsupport.AuditBinding{Scope: hc.NetworkFlowAuditScope{EventCode: event, OperationRef: "operation:" + operation, ActorRef: "actor:" + actorText, IncidentRef: "incident:" + incidentText, ResourceKind: kind, ResourceRef: "resource:" + operation, CorrelationKey: operation}, ActorID: actor, IncidentID: incident, ResourceID: resource, ClientTxnID: operation}
		if event == hc.NetworkFlowAuditEventGraphQueryExecuted {
			b.ClientTxnID = ""
			b.RequestID = operation
		}
		return b
	}
	verify := func(t *testing.T, f *networkflowsupport.AuditFixture, resource string, replay func() error) {
		t.Helper()
		if err := f.Verify(ctx, resource, replay); err != nil {
			t.Fatal(err)
		}
	}
	path := "/tables/" + table.TableID
	t.Run("denied zero", func(t *testing.T) {
		f := arm(t, binding(hc.NetworkFlowAuditEventTableRenamed, "audit-denied", hc.NetworkFlowAuditResourceTable, table.TableID), hc.NetworkFlowAuditAssertionZeroOccurrences, 0, 0)
		if _, err := h.Pool.Exec(ctx, `UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2`, incident, actor); err != nil {
			t.Fatal(err)
		}
		httptestx.RequireErrorEnvelope(t, request(http.MethodPatch, path, map[string]any{"client_txn_id": "audit-denied", "base_table_version": 1, "display_name": "Denied"}), http.StatusForbidden, "authorization_denied")
		verify(t, f, table.TableID, nil)
		if _, err := h.Pool.Exec(ctx, `UPDATE incident_memberships SET role='admin' WHERE incident_id=$1 AND user_id=$2`, incident, actor); err != nil {
			t.Fatal(err)
		}
	})
	rename := map[string]any{"client_txn_id": "audit-rename", "base_table_version": 1, "display_name": "Audited"}
	t.Run("rename and immutable replay", func(t *testing.T) {
		f := arm(t, binding(hc.NetworkFlowAuditEventTableRenamed, "audit-rename", hc.NetworkFlowAuditResourceTable, table.TableID), hc.NetworkFlowAuditAssertionNoAuditReplay, 0, 1)
		receipt := httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, path, rename), http.StatusOK)["data"]
		verify(t, f, table.TableID, func() error {
			got := httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, path, rename), http.StatusOK)["data"]
			if !reflect.DeepEqual(got, receipt) {
				return fmt.Errorf("rename receipt drift")
			}
			return nil
		})
		// The baseline is now a real committed occurrence, not an assumed zero.
		f = arm(t, binding(hc.NetworkFlowAuditEventTableRenamed, "audit-rename", hc.NetworkFlowAuditResourceTable, table.TableID), hc.NetworkFlowAuditAssertionExactCount, 1, 1)
		httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, path, rename), http.StatusOK)
		verify(t, f, table.TableID, nil)
	})
	t.Run("no-op zero", func(t *testing.T) {
		f := arm(t, binding(hc.NetworkFlowAuditEventTableRenamed, "audit-noop", hc.NetworkFlowAuditResourceTable, table.TableID), hc.NetworkFlowAuditAssertionZeroOccurrences, 0, 0)
		httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, path, map[string]any{"client_txn_id": "audit-noop", "base_table_version": 2, "display_name": "Audited"}), http.StatusOK)
		verify(t, f, table.TableID, nil)
	})
	t.Run("rollback zero", func(t *testing.T) {
		migration, err := pgtest.OpenPurposeDatabase(h.Database.DSN, postgres.PurposeMigration)
		if err != nil {
			t.Fatal(err)
		}
		defer migration.Close()
		if _, err := migration.ExecContext(ctx, `CREATE FUNCTION nf_audit_fixture_failure() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.event_kind='network_flow_table_renamed' THEN RAISE EXCEPTION 'audit fixture failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER nf_audit_fixture_failure BEFORE INSERT ON deployment_admin_audit_events FOR EACH ROW EXECUTE FUNCTION nf_audit_fixture_failure()`); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if _, err := migration.ExecContext(ctx, `DROP TRIGGER nf_audit_fixture_failure ON deployment_admin_audit_events; DROP FUNCTION nf_audit_fixture_failure()`); err != nil {
				t.Error(err)
			}
		}()
		f := arm(t, binding(hc.NetworkFlowAuditEventTableRenamed, "audit-rollback", hc.NetworkFlowAuditResourceTable, table.TableID), hc.NetworkFlowAuditAssertionZeroOccurrences, 0, 0)
		httptestx.RequireErrorEnvelope(t, request(http.MethodPatch, path, map[string]any{"client_txn_id": "audit-rollback", "base_table_version": 2, "display_name": "Rolled back"}), http.StatusInternalServerError, "internal_error")
		verify(t, f, table.TableID, nil)
	})
	graphBody := map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": table.TableID}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"}}
	t.Run("graph query exact", func(t *testing.T) {
		f := arm(t, binding(hc.NetworkFlowAuditEventGraphQueryExecuted, "audit-graph", hc.NetworkFlowAuditResourceGraph, ""), hc.NetworkFlowAuditAssertionExactCount, 0, 1)
		graph := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, root+"/graphs/query", graphBody, httptestx.WithCookies(login.SessionCookie), httptestx.WithHeader("X-Request-Id", "audit-graph")), http.StatusOK)["data"].(map[string]any)
		verify(t, f, graph["graph_query_digest"].(string), nil)
	})
	t.Run("binding create reuse and replay", func(t *testing.T) {
		rows, err := store.ListRows(ctx, incident, table.TableID)
		if err != nil {
			t.Fatal(err)
		}
		body := map[string]any{"schema_id": "cartulary.network_flow.indicator_link_request.v1", "client_txn_id": "audit-binding-create", "selector": map[string]any{"kind": "row_field_value", "network_flow_table_id": table.TableID, "network_flow_row_id": rows[0].RowID, "field_key": "network_flow.src_ip"}, "target": map[string]any{"mode": "create_indicator", "indicator_type": "ipv4_addr"}, "observation_mode": "binding_only", "confirm_exact_value": "192.0.2.10"}
		f := arm(t, binding(hc.NetworkFlowAuditEventIndicatorBindingCreated, "audit-binding-create", hc.NetworkFlowAuditResourceIndicatorBinding, ""), hc.NetworkFlowAuditAssertionNoAuditReplay, 0, 1)
		created := httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/indicator-links", body), http.StatusCreated)["data"].(map[string]any)
		id := created["binding"].(map[string]any)["network_flow_indicator_binding_id"].(string)
		verify(t, f, id, func() error {
			httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/indicator-links", body), http.StatusCreated)
			return nil
		})
		body["client_txn_id"] = "audit-binding-reuse"
		f = arm(t, binding(hc.NetworkFlowAuditEventIndicatorBindingReused, "audit-binding-reuse", hc.NetworkFlowAuditResourceIndicatorBinding, id), hc.NetworkFlowAuditAssertionNoAuditReplay, 0, 1)
		reused := httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/indicator-links", body), http.StatusOK)["data"].(map[string]any)
		if reused["duplicate"] != true {
			t.Fatal("binding not reused")
		}
		verify(t, f, id, func() error {
			httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/indicator-links", body), http.StatusOK)
			return nil
		})
	})
	t.Run("concurrent actor and incident isolation", func(t *testing.T) {
		otherActorText := flowtest.SeedLocalUser(t, h.DB, "audit-other@example.test", "Other actor", "AuditOtherPass!", false)
		cookie, _ := flowtest.LoginLocalUser(t, h.Server.HTTP.URL, "audit-other@example.test", "AuditOtherPass!", nil)
		otherText := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "audit-other-incident", "incident_key": "IR-NF-AUDIT-OTHER", "title": "Other audit incident"})["incident_id"].(string)
		otherIncident := uuid.MustParse(otherText)
		otherActor := uuid.MustParse(otherActorText)
		if _, err := h.Pool.Exec(ctx, `INSERT INTO incident_memberships(incident_id,user_id,role,added_by_user_id,updated_by_user_id) VALUES($1,$2,'viewer',$3,$3)`, otherIncident, otherActor, actor); err != nil {
			t.Fatal(err)
		}
		otherTable := createTestTable(t, h.Pool, store, actor, otherIncident, "other.csv", nil, 2)
		b1 := binding(hc.NetworkFlowAuditEventGraphQueryExecuted, "audit-concurrent", hc.NetworkFlowAuditResourceGraph, "")
		b2 := b1
		b2.ActorID = otherActor
		b2.IncidentID = otherIncident
		b2.Scope.ActorRef = "actor:" + otherActorText
		b2.Scope.IncidentRef = "incident:" + otherText
		fixtures := []*networkflowsupport.AuditFixture{arm(t, b1, hc.NetworkFlowAuditAssertionExactCount, 0, 1), arm(t, b2, hc.NetworkFlowAuditAssertionExactCount, 0, 1)}
		responses := make([]*http.Response, 2)
		var wait sync.WaitGroup
		for i, input := range []struct {
			incident, table string
			cookie          *http.Cookie
		}{{incidentText, table.TableID, login.SessionCookie}, {otherText, otherTable.TableID, cookie}} {
			wait.Add(1)
			go func() {
				defer wait.Done()
				responses[i] = httptestx.DoJSON(t, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+input.incident+"/network-flow/graphs/query", map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": input.table}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"}}, httptestx.WithCookies(input.cookie), httptestx.WithHeader("X-Request-Id", "audit-concurrent"))
			}()
		}
		wait.Wait()
		for i, response := range responses {
			result := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			verify(t, fixtures[i], result["graph_query_digest"].(string), nil)
		}
	})
	t.Run("deletion and replay", func(t *testing.T) {
		f := arm(t, binding(hc.NetworkFlowAuditEventTableSoftDeleted, "audit-delete", hc.NetworkFlowAuditResourceTable, table.TableID), hc.NetworkFlowAuditAssertionNoAuditReplay, 0, 1)
		body := map[string]any{"client_txn_id": "audit-delete", "base_table_version": 2}
		httptestx.RequireSuccessEnvelope(t, request(http.MethodDelete, path, body), http.StatusOK)
		verify(t, f, table.TableID, func() error {
			httptestx.RequireSuccessEnvelope(t, request(http.MethodDelete, path, body), http.StatusOK)
			return nil
		})
	})
	t.Run("incorrect evidence fails the fixture", func(t *testing.T) {
		for _, check := range []struct {
			name            string
			baseline, final int
		}{{"baseline", 0, 1}, {"final count", 1, 2}} {
			t.Run(check.name, func(t *testing.T) {
				isolated := hc.NewNetworkFlowAuditAssertionRegistry()
				b := binding(hc.NetworkFlowAuditEventTableRenamed, "audit-rename", hc.NetworkFlowAuditResourceTable, table.TableID)
				f, err := networkflowsupport.NewAuditFixture(ctx, isolated, h.Pool, b)
				if err != nil {
					t.Fatal(err)
				}
				mux := http.NewServeMux()
				if err := hc.RegisterNetworkFlowAuditAssertionRoutes(isolated)(mux, httpapi.DependencySet{Env: map[string]string{"CARTULARY_ENABLE_TEST_ROUTES": "1", "CARTULARY_TEST_RUNTIME_MARKER": "harness-owned", "CARTULARY_TEST_ROUTE_TOKEN": httptestx.TestRouteToken}}); err != nil {
					t.Fatal(err)
				}
				endpoint := httptest.NewServer(mux)
				defer endpoint.Close()
				s := b.Scope
				httptestx.RequireStatus(t, httptestx.DoJSON(t, http.MethodPost, endpoint.URL+"/api/v1/test/runtime/network-flow-audit-assertions", map[string]any{"assertion_kind": hc.NetworkFlowAuditAssertionExactCount, "event_code": s.EventCode, "operation_ref": s.OperationRef, "actor_ref": s.ActorRef, "incident_ref": s.IncidentRef, "resource_kind": s.ResourceKind, "resource_ref": s.ResourceRef, "baseline_count": check.baseline, "expected_final_count": check.final, "correlation_key": s.CorrelationKey, "consume_once": true}, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)
				if err := isolated.RequireConsumed(); err == nil {
					t.Fatal("pending required assertion reported success")
				}
				httptestx.RequireSuccessEnvelope(t, request(http.MethodPatch, path, rename), http.StatusOK)
				if err := f.Verify(ctx, table.TableID, nil); err == nil || !strings.Contains(err.Error(), check.name) {
					t.Fatalf("false audit evidence admitted: %v", err)
				}
				if err := isolated.RequireConsumed(); err == nil {
					t.Fatal("consumed failed assertion reported success")
				}
			})
		}
	})
	if err := controls.AuditAssertions.RequireConsumed(); err != nil {
		t.Fatal(err)
	}
}
