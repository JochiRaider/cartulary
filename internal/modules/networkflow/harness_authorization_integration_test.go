package networkflow_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/google/uuid"
)

func assertNetworkFlowAuthorizationConsumers(t *testing.T) {
	ctx := context.Background()
	controls := hc.NewControls()
	h := claimedNetworkFlowServerWithControlsForRouteTest(t, appsupport.StartRuntime(t), "network-flow-auth-controls", "", controls)
	login, actorText := flowtest.ProvisionBootstrapAdmin(t, h.Server.HTTP.URL)
	actor := uuid.MustParse(actorText)
	incidentText := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "auth-control-incident", "incident_key": "IR-NF-AUTH-CONTROL", "title": "Authorization controls"})["incident_id"].(string)
	incident := uuid.MustParse(incidentText)
	session, unit := seedImportSessionUnit(t, h.Pool, incident, actor, "auth.csv")
	store := newTestNetworkFlowStore(t, h.Pool, h.Revisions.Appender())
	table, err := store.CreateTable(ctx, CreateTableParams{IncidentID: incident, ActorUserID: actor, ImportSessionID: session, ImportUnitID: unit, SourceContentSHA256: testSHA1, OriginalFilename: "auth.csv", SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "test-key", MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV, ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{testFlowRow(1, "a"), testFlowRow(2, "b")}, Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	sessionID := uuid.MustParse(flowtest.QuerySessionRow(t, h.DB, actorText).SessionID)
	root := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentText + "/network-flow"
	request := func(method, path string, body any) *http.Response {
		return httptestx.DoJSON(t, method, root+path, body, httptestx.WithCookies(login.SessionCookie))
	}
	graphBody := map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": table.TableID}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"}}
	graph := httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/graphs/query", graphBody), http.StatusOK)["data"].(map[string]any)
	initial := map[string]any{"schema_id": schemaTableQueryRequestForTest, "limit": 1}
	first := httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/tables/"+table.TableID+"/query", initial), http.StatusOK)["data"].(map[string]any)
	token := first["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
	continuation := map[string]any{"schema_id": schemaTableQueryContinuationForTest, "cursor_token": token}
	newFixture := func(t *testing.T, b networkflowsupport.AuthorizationBinding) *networkflowsupport.AuthorizationFixture {
		t.Helper()
		f, err := networkflowsupport.NewAuthorizationFixture(ctx, controls.Transitions, h.Pool, b, h.Server.Clock.Now)
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	arm := func(t *testing.T, b networkflowsupport.AuthorizationBinding, boundary, kind string) {
		t.Helper()
		httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, h.Server.HTTP.URL+"/api/v1/test/runtime/network-flow-auth-transitions", map[string]any{"boundary": boundary, "transition_kind": kind, "actor_ref": b.ActorRef, "incident_ref": b.IncidentRef, "resource_kind": b.ResourceKind, "resource_ref": b.ResourceRef, "correlation_key": "owned-operation", "consume_once": true}, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)
	}
	before := func(t *testing.T, f *networkflowsupport.AuthorizationFixture, boundary string) {
		t.Helper()
		ok, err := f.Before(ctx, boundary, "owned-operation")
		if err != nil || !ok {
			t.Fatalf("apply owned transition: consumed=%v err=%v", ok, err)
		}
	}
	base := networkflowsupport.AuthorizationBinding{ActorRef: "actor:owned", IncidentRef: "incident:owned", ActorID: actor, IncidentID: incident, SessionID: sessionID, TableID: table.TableID}
	for _, resource := range []struct {
		kind, method, path, boundary string
		body                         any
	}{
		{hc.NetworkFlowAuthResourceIncident, http.MethodGet, "/source-profiles", hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization, nil},
		{hc.NetworkFlowAuthResourceNetworkFlowTable, http.MethodGet, "/tables/" + table.TableID, hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization, nil},
		{hc.NetworkFlowAuthResourceNetworkFlowWorkspace, http.MethodGet, "/tables", hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization, nil},
		{hc.NetworkFlowAuthResourceNetworkFlowGraph, http.MethodPost, "/graphs/query", hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization, graphBody},
		{hc.NetworkFlowAuthResourceNetworkFlowContributors, http.MethodPost, "/graphs/contributors/query", hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization, map[string]any{"schema_id": "cartulary.network_flow.graph_contributor_query_request.v2", "graph_query": graph["semantic_query"], "graph_query_digest": graph["graph_query_digest"], "selector": graph["vertex_selectors"].([]any)[0].(map[string]any)["selector"], "limit": 1}},
		{hc.NetworkFlowAuthResourceNetworkFlowCursor, http.MethodPost, "/tables/" + table.TableID + "/query", hc.NetworkFlowAuthTransitionBoundaryCursorBeforeAuthorizationRecheck, continuation},
	} {
		t.Run(resource.kind, func(t *testing.T) {
			b := base
			b.ResourceKind = resource.kind
			b.ResourceRef = "resource:" + resource.kind
			f := newFixture(t, b)
			arm(t, b, resource.boundary, hc.NetworkFlowAuthTransitionKindIncidentMembershipRevoked)
			if ok, err := f.Before(ctx, resource.boundary, "unrelated-operation"); err != nil || ok {
				t.Fatal("unrelated operation consumed transition")
			}
			httptestx.RequireSuccessEnvelope(t, request(resource.method, resource.path, resource.body), http.StatusOK)
			other := b
			other.ResourceRef = "resource:unrelated"
			unrelated := newFixture(t, other)
			if ok, err := unrelated.Before(ctx, resource.boundary, "owned-operation"); err != nil || ok {
				t.Fatal("unrelated resource consumed transition")
			}
			before(t, f, resource.boundary)
			denied := httptestx.RequireErrorEnvelope(t, request(resource.method, resource.path, resource.body), http.StatusNotFound, "incident_not_found")
			if len(denied["error"].(map[string]any)["details"].(map[string]any)) != 0 {
				t.Fatal("hidden resource details disclosed")
			}
			arm(t, b, resource.boundary, hc.NetworkFlowAuthTransitionKindIncidentMembershipRestored)
			before(t, f, resource.boundary)
			httptestx.RequireSuccessEnvelope(t, request(resource.method, resource.path, resource.body), http.StatusOK)
		})
	}
	t.Run("cursor observes table rename and deletion", func(t *testing.T) {
		b := base
		b.ResourceKind = hc.NetworkFlowAuthResourceNetworkFlowCursor
		b.ResourceRef = "cursor:table"
		f := newFixture(t, b)
		boundary := hc.NetworkFlowAuthTransitionBoundaryCursorBeforeAuthorizationRecheck
		arm(t, b, boundary, hc.NetworkFlowAuthTransitionKindNetworkFlowTableRenamed)
		before(t, f, boundary)
		httptestx.RequireSuccessEnvelope(t, request(http.MethodPost, "/tables/"+table.TableID+"/query", continuation), http.StatusOK)
		got := httptestx.RequireSuccessEnvelope(t, request(http.MethodGet, "/tables/"+table.TableID, nil), http.StatusOK)["data"].(map[string]any)
		if got["table"].(map[string]any)["display_name"] != "Fixture "+table.TableID {
			t.Fatal("rename did not change retained state")
		}
		arm(t, b, boundary, hc.NetworkFlowAuthTransitionKindNetworkFlowTableSoftDeleted)
		before(t, f, boundary)
		httptestx.RequireErrorEnvelope(t, request(http.MethodPost, "/tables/"+table.TableID+"/query", continuation), http.StatusConflict, "network_flow_table_not_active")
	})
	t.Run("incident deletion", func(t *testing.T) {
		text := uuid.NewString()
		// This disposable fixture has no retained audit/receipt references. We
		// preserve the FK protection for product-created, audited incidents.
		if _, err := h.Pool.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'IR-NF-AUTH-DELETE','ir-nf-auth-delete','Disposable owned incident','active',$2,$2)`, text, actor); err != nil {
			t.Fatal(err)
		}
		if _, err := h.Pool.Exec(ctx, `INSERT INTO incident_memberships(incident_id,user_id,role,added_by_user_id,updated_by_user_id) VALUES($1,$2,'admin',$2,$2)`, text, actor); err != nil {
			t.Fatal(err)
		}
		b := base
		b.IncidentID = uuid.MustParse(text)
		b.IncidentRef = "incident:deleted"
		b.TableID = ""
		b.ResourceKind = hc.NetworkFlowAuthResourceIncident
		b.ResourceRef = "incident:deleted"
		f := newFixture(t, b)
		boundary := hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization
		arm(t, b, boundary, hc.NetworkFlowAuthTransitionKindIncidentDeleted)
		before(t, f, boundary)
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.MethodGet, h.Server.HTTP.URL+"/api/v1/incidents/"+text+"/network-flow/source-profiles", nil, httptestx.WithCookies(login.SessionCookie)), http.StatusNotFound, "incident_not_found")
		httptestx.RequireSuccessEnvelope(t, request(http.MethodGet, "/tables", nil), http.StatusOK)
	})
	t.Run("session revocation and scope resolution", func(t *testing.T) {
		b := base
		b.ResourceKind = hc.NetworkFlowAuthResourceNetworkFlowWorkspace
		b.ResourceRef = "workspace:owned"
		f := newFixture(t, b)
		wrong := b
		wrong.IncidentID = uuid.New()
		if _, err := networkflowsupport.NewAuthorizationFixture(ctx, controls.Transitions, h.Pool, wrong, h.Server.Clock.Now); err == nil {
			t.Fatal("unowned fixture reference resolved")
		}
		boundary := hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization
		arm(t, b, boundary, hc.NetworkFlowAuthTransitionKindSessionRevoked)
		before(t, f, boundary)
		httptestx.RequireErrorEnvelope(t, request(http.MethodGet, "/tables", nil), http.StatusUnauthorized, "session_required")
	})
	if err := controls.Transitions.RequireConsumed(); err != nil {
		t.Fatal(err)
	}
}
