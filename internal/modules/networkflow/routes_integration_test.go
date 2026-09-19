package networkflow_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestNetworkFlowPaginationRecovery_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "network-flow-pagination-recovery")
	login, actorText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	start := harness.Server.Clock.Now().Truncate(time.Second)
	httptestx.SetClockFixed(t, harness.Server, start)
	actor := uuid.MustParse(actorText)
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "txn-pagination-incident", "incident_key": "IR-NF-PAGES", "title": "Pagination"})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	store := newTestNetworkFlowStore(t, harness.Pool, harness.Revisions.Appender())
	session, unit := seedImportSessionUnit(t, harness.Pool, incidentID, actor, "pages.csv")
	rows := []FlowRow{testFlowRow(1, "a"), testFlowRow(2, "b"), testFlowRow(3, "c")}
	for i := range rows {
		rows[i].SrcIP = "192.0.2.1"
		rows[i].DstIP = "192.0.2.2"
	}
	diagnostics := []RejectedRowDiagnostic{testDiagnostic(4, "network_flow_invalid_ip"), testDiagnostic(5, "network_flow_invalid_ip"), testDiagnostic(6, "network_flow_invalid_ip")}
	for i := range diagnostics {
		diagnostics[i].DiagnosticID = "nfd_" + strings.Repeat(string(rune('a'+i)), 64)
	}
	table, err := store.CreateTable(context.Background(), CreateTableParams{IncidentID: incidentID, ActorUserID: actor, ImportSessionID: session, ImportUnitID: unit, SourceContentSHA256: testSHA1, OriginalFilename: "pages.csv", SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "route-test-key", MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV, ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: rows, Diagnostics: diagnostics, Now: start})
	if err != nil {
		t.Fatal(err)
	}
	root := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow"
	query := func(path string, body map[string]any) map[string]any {
		response := httptestx.DoJSON(t, http.MethodPost, root+path, body, httptestx.WithCookies(login.SessionCookie))
		return httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
	}
	graph := query("/graphs/query", map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": table.TableID}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"}})
	selector := graph["vertex_selectors"].([]any)[0].(map[string]any)["selector"]
	for _, route := range []struct {
		name, path, continuation, items string
		initial                         map[string]any
	}{
		{"rows", "/tables/" + table.TableID + "/query", schemaTableQueryContinuationForTest, "rows", map[string]any{"schema_id": schemaTableQueryRequestForTest, "limit": 1}},
		{"diagnostics", "/tables/" + table.TableID + "/rejected-rows/query", "cartulary.network_flow.rejected_rows_query_continuation.v1", "diagnostics", map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_request.v1", "limit": 1}},
		{"contributors", "/graphs/contributors/query", "cartulary.network_flow.graph_contributor_query_continuation.v1", "contributors", map[string]any{"schema_id": "cartulary.network_flow.graph_contributor_query_request.v2", "graph_query": graph["semantic_query"], "graph_query_digest": graph["graph_query_digest"], "selector": selector, "limit": 1}},
	} {
		t.Run(route.name, func(t *testing.T) {
			issued := harness.Server.Clock.Now()
			first := query(route.path, route.initial)
			firstItems := first[route.items].([]any)
			token := first["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
			continuation := map[string]any{"schema_id": route.continuation, "cursor_token": token}
			rename := httptestx.DoJSON(t, http.MethodPatch, root+"/tables/"+table.TableID, map[string]any{
				"client_txn_id": "pagination-rename-" + route.name, "base_table_version": table.TableVersion, "display_name": "Pages " + route.name,
			}, httptestx.WithCookies(login.SessionCookie, login.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
			httptestx.RequireSuccessEnvelope(t, rename, http.StatusOK)
			table.TableVersion++
			second := query(route.path, continuation)
			if len(firstItems) != 1 || len(second[route.items].([]any)) != 1 || reflect.DeepEqual(firstItems, second[route.items]) {
				t.Fatalf("incorrect successive pages: %#v %#v", first, second)
			}
			// Previous uses the original producing request and returns the same authorized resources.
			if prior := query(route.path, route.initial); !reflect.DeepEqual(prior[route.items], firstItems) {
				t.Fatal("initial page replay changed rows")
			}
			httptestx.SetClockAfter(t, harness.Server, issued, 899*time.Second)
			query(route.path, continuation)
			httptestx.SetClockAfter(t, harness.Server, issued, 900*time.Second)
			response := httptestx.DoJSON(t, http.MethodPost, root+route.path, continuation, httptestx.WithCookies(login.SessionCookie))
			body := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "network_flow_cursor_invalid")
			details := body["error"].(map[string]any)["details"].(map[string]any)
			if details["reason_code"] != "expired" || details["retry_action"] != "restart_query" {
				t.Fatalf("expiry recovery details: %#v", details)
			}
			restarted := query(route.path, route.initial)
			if !reflect.DeepEqual(restarted[route.items], firstItems) {
				t.Fatal("restart changed first-page resources")
			}
			freshToken := restarted["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
			last := query(route.path, map[string]any{"schema_id": route.continuation, "cursor_token": freshToken})
			finalToken := last["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
			last = query(route.path, map[string]any{"schema_id": route.continuation, "cursor_token": finalToken})
			if last["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"] != nil {
				t.Fatal("terminal page has a continuation")
			}
		})
	}
}

func TestNetworkFlowRoutesRemainUnclaimedByDefault(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "network-flow-routes-unclaimed")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-routes-unclaimed-incident",
		"incident_key":  "IR-NF-UNCLAIMED",
		"title":         "Network Flow unclaimed",
	})
	incidentID := incident["incident_id"].(string)

	resp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/network-flow/source-profiles", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "extension_profile_not_claimed")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["profile_id"] != ProfileID {
		t.Fatalf("unexpected unclaimed extension details: %#v", details)
	}
}

func TestNetworkFlowEffectiveResourceLimitsReachDiscovery_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerWithLimitsForRouteTest(
		t,
		runtime,
		"network-flow-effective-limits",
		`{"max_graph_vertices":7000,"max_graph_edges":0,"max_query_limit":750,"max_time_buckets_per_graph":64}`,
	)
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-effective-limits-incident",
		"incident_key":  "IR-NF-LIMITS",
		"title":         "Network Flow configured limits",
	})
	incidentID := incident["incident_id"].(string)

	response := httptestx.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/network-flow/source-profiles",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	data := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
	limits := data["effective_limits"].(map[string]any)
	if len(limits) != 23 || limits["network_flow.max_graph_vertices"] != float64(7000) ||
		limits["network_flow.max_graph_edges"] != float64(0) ||
		limits["network_flow.max_query_limit"] != float64(750) ||
		limits["network_flow.max_time_buckets_per_graph"] != float64(64) ||
		limits["network_flow.max_header_scalar_length"] != float64(256) {
		t.Fatalf("configured effective-limit discovery = %#v", limits)
	}
}

func TestNetworkFlowStreamingGraphContributingRowLimit_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerWithLimitsForRouteTest(
		t,
		runtime,
		"network-flow-contributing-row-limit",
		`{"max_contributing_rows_per_graph":1}`,
	)
	adminLogin, adminIDText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	adminID := uuid.MustParse(adminIDText)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-contributing-limit-incident",
		"incident_key":  "IR-NF-CONTRIBUTING-LIMIT",
		"title":         "Network Flow contributing limit",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	store := newTestNetworkFlowStore(t, harness.Pool, harness.Revisions.Appender())
	sessionID, unitID := seedImportSessionUnit(t, harness.Pool, incidentID, adminID, "contributing-limit.csv")
	table, err := store.CreateTable(context.Background(), CreateTableParams{
		IncidentID: incidentID, ActorUserID: adminID, ImportSessionID: sessionID, ImportUnitID: unitID,
		SourceContentSHA256: testSHA1, OriginalFilename: "contributing-limit.csv",
		SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "route-test-key",
		MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{testFlowRow(2, "b"), testFlowRow(1, "a")},
		Now: time.Date(2026, 7, 10, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create contributing-limit table: %v", err)
	}
	response := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/graphs/query", map[string]any{
		"schema_id":   "cartulary.network_flow.graph_query_request.v2",
		"table_scope": map[string]any{"mode": "active_table", "active_table_id": table.TableID},
		"aggregation": map[string]any{"mode": "default_flow_edge_v1"},
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	body := httptestx.RequireErrorEnvelope(t, response, http.StatusRequestEntityTooLarge, "network_flow_graph_limit_exceeded")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != "contributing_row_limit_exceeded" || details["limit"] != float64(1) || details["actual"] != float64(2) {
		t.Fatalf("contributing-row limit details = %#v", details)
	}
}

func TestNetworkFlowRoutesQueryPageAndInvalidateAfterSoftDelete(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "network-flow-routes-query")
	adminLogin, adminIDText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	adminID := uuid.MustParse(adminIDText)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-routes-query-incident",
		"incident_key":  "IR-NF-QUERY",
		"title":         "Network Flow query",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	store := newTestNetworkFlowStore(
		t,
		harness.Pool,
		harness.Revisions.Appender(),
	)
	sessionID, unitID := seedImportSessionUnit(t, harness.Pool, incidentID, adminID, "flows.csv")
	first := testFlowRow(1, "1")
	first.SrcIP = "192.0.2.10"
	second := testFlowRow(2, "2")
	second.SrcIP = "198.51.100.7"
	second.BytesCount = "100"
	table, err := store.CreateTable(context.Background(), CreateTableParams{
		IncidentID:                incidentID,
		ActorUserID:               adminID,
		ImportSessionID:           sessionID,
		ImportUnitID:              unitID,
		SourceContentSHA256:       testSHA1,
		OriginalFilename:          "flows.csv",
		SourceFilenameDigest:      testSHA2,
		SourceFilenameDigestKeyID: "route-test-key",
		MappingFingerprint:        testSHA3,
		SourceProfileID:           SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID:           ParserProfileRFC4180HeaderedCSV,
		Rows:                      []FlowRow{second, first},
		Diagnostics:               []RejectedRowDiagnostic{testDiagnostic(3, "network_flow_invalid_ip")},
		Now:                       time.Date(2026, 7, 10, 12, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create network flow table for routes: %v", err)
	}

	sourceProfilesResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/source-profiles", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	sourceProfiles := httptestx.RequireSuccessEnvelope(t, sourceProfilesResp, http.StatusOK)["data"].(map[string]any)
	if sourceProfiles["schema_id"] != "cartulary.network_flow.source_profile_list.v2" {
		t.Fatalf("unexpected source profiles payload: %#v", sourceProfiles)
	}
	effectiveLimits := sourceProfiles["effective_limits"].(map[string]any)
	if len(effectiveLimits) != 23 || effectiveLimits["network_flow.max_header_scalar_length"] != float64(256) ||
		effectiveLimits["network_flow.max_contributing_rows_per_graph"] != float64(250000) ||
		effectiveLimits["network_flow.max_time_buckets_per_graph"] != float64(256) ||
		effectiveLimits["network_flow.graph_materialization_timeout_seconds"] != float64(300) {
		t.Fatalf("source-profile effective limits = %#v", effectiveLimits)
	}

	listResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/tables", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	list := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	tables := list["tables"].([]any)
	if len(tables) != 1 || tables[0].(map[string]any)["network_flow_table_id"] != table.TableID {
		t.Fatalf("unexpected table list: %#v", list)
	}

	queryPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/tables/" + table.TableID + "/query"
	firstPageResp := httptestx.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id": schemaTableQueryRequestForTest,
		"sort": []map[string]any{
			{"field_key": "source_row_number", "direction": "asc"},
		},
		"limit": 1,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	firstPage := httptestx.RequireSuccessEnvelope(t, firstPageResp, http.StatusOK)["data"].(map[string]any)
	firstRows := firstPage["rows"].([]any)
	if len(firstRows) != 1 || firstRows[0].(map[string]any)["source_row_number"] != float64(1) {
		t.Fatalf("unexpected first query page: %#v", firstPage)
	}
	nextToken := firstPage["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
	if !strings.HasPrefix(nextToken, "nfc2.route-cursor-v1.") {
		t.Fatalf("expected Network Flow cursor token with key id, got %q", nextToken)
	}

	secondPageResp := httptestx.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id":    schemaTableQueryContinuationForTest,
		"cursor_token": nextToken,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	secondPage := httptestx.RequireSuccessEnvelope(t, secondPageResp, http.StatusOK)["data"].(map[string]any)
	secondRows := secondPage["rows"].([]any)
	if len(secondRows) != 1 || secondRows[0].(map[string]any)["network_flow.src_ip"] != "198.51.100.7" {
		t.Fatalf("unexpected continuation page: %#v", secondPage)
	}

	filterResp := httptestx.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id": schemaTableQueryRequestForTest,
		"filters": []map[string]any{
			{"field_key": "network_flow.src_ip", "op": "eq", "value": "198.51.100.7"},
		},
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	filtered := httptestx.RequireSuccessEnvelope(t, filterResp, http.StatusOK)["data"].(map[string]any)
	filteredRows := filtered["rows"].([]any)
	if len(filteredRows) != 1 || filteredRows[0].(map[string]any)["source_row_number"] != float64(2) {
		t.Fatalf("unexpected filtered query: %#v", filtered)
	}

	rejectedResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/tables/"+table.TableID+"/rejected-rows/query", map[string]any{
		"schema_id":   "cartulary.network_flow.rejected_rows_query_request.v1",
		"error_codes": []string{"network_flow_invalid_ip"},
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	rejected := httptestx.RequireSuccessEnvelope(t, rejectedResp, http.StatusOK)["data"].(map[string]any)
	if diagnostics := rejected["diagnostics"].([]any); len(diagnostics) != 1 {
		t.Fatalf("unexpected rejected-row diagnostics: %#v", rejected)
	}

	socket := incidentwstest.ConnectExtensionWorkspaceSocket(
		t, harness.Server, incidentID.String(), ProfileID, WorkspaceKeyNetworkAnalysis, adminLogin.SessionCookie.Value,
	)
	defer socket.Close(1000, "test_complete")
	intentSelector := collaborationsupport.IntentSelector{
		IncidentID: incidentID.String(), EventFamily: "extension_resource_changed",
	}

	tablePath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/tables/" + table.TableID
	renameBody := map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename",
		"base_table_version": table.TableVersion,
		"display_name":       "Routes flows",
	}
	renameResp := httptestx.DoJSON(t, http.MethodPatch, tablePath, renameBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	renamed := httptestx.RequireSuccessEnvelope(t, renameResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	renamedVersion := int64(renamed["table_version"].(float64))
	if renamed["display_name"] != "Routes flows" || renamedVersion != table.TableVersion+1 {
		t.Fatalf("unexpected rename result: %#v", renamed)
	}
	requireNetworkFlowResourceChange(t, socket, incidentID, table.TableID, platformws.ExtensionResourceChangeKindInvalidate, platformws.ExtensionResourceReasonRenamed)
	wantIntentCount := collaborationsupport.CountIntents(t, harness.DB, intentSelector)

	renameReplayResp := httptestx.DoJSON(t, http.MethodPatch, tablePath, renameBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	renameReplay := httptestx.RequireSuccessEnvelope(t, renameReplayResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if renameReplay["table_version"] != renamed["table_version"] {
		t.Fatalf("unexpected rename replay payload: %#v", renameReplay)
	}
	collaborationsupport.RequireIntentCount(t, harness.DB, intentSelector, wantIntentCount)
	requireNoNetworkFlowResourceChange(t, socket, table.TableID)

	divergentRenameResp := httptestx.DoJSON(t, http.MethodPatch, tablePath, map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename",
		"base_table_version": table.TableVersion,
		"display_name":       "Different routes flows",
	}, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, divergentRenameResp, http.StatusConflict, "client_txn_conflict")
	collaborationsupport.RequireIntentCount(t, harness.DB, intentSelector, wantIntentCount)
	requireNoNetworkFlowResourceChange(t, socket, table.TableID)

	noOpRenameResp := httptestx.DoJSON(t, http.MethodPatch, tablePath, map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename-noop",
		"base_table_version": renamedVersion,
		"display_name":       "Routes flows",
	}, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	noOpRenamed := httptestx.RequireSuccessEnvelope(t, noOpRenameResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if int64(noOpRenamed["table_version"].(float64)) != renamedVersion {
		t.Fatalf("no-op rename changed table version: %#v", noOpRenamed)
	}
	collaborationsupport.RequireIntentCount(t, harness.DB, intentSelector, wantIntentCount)
	requireNoNetworkFlowResourceChange(t, socket, table.TableID)
	if got := networkFlowRouteCountRows(t, harness.DB, `
SELECT COUNT(*)
  FROM deployment_admin_audit_events
 WHERE incident_id = $1
   AND event_source = 'network_flow'
   AND event_kind = 'network_flow_table_renamed'
   AND after_json->>'network_flow_table_id' = $2
`, incidentID, table.TableID); got != 1 {
		t.Fatalf("expected one changed-rename audit event, got %d", got)
	}
	if got := networkFlowRouteCountRows(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'nf.tables.patch'
   AND actor_user_id = $1
   AND scope_key = $2
`, adminID, incidentID.String()+":"+table.TableID); got != 2 {
		t.Fatalf("expected changed and no-op rename idempotency rows only, got %d", got)
	}

	deleteBody := map[string]any{
		"client_txn_id":      "txn-network-flow-route-delete",
		"base_table_version": renamedVersion,
	}
	deleteResp := httptestx.DoJSON(t, http.MethodDelete, tablePath, deleteBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	deleted := httptestx.RequireSuccessEnvelope(t, deleteResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if deleted["table_status"] != TableStatusSoftDeleted || int64(deleted["table_version"].(float64)) != renamedVersion+1 {
		t.Fatalf("unexpected delete result: %#v", deleted)
	}
	requireNetworkFlowResourceChange(t, socket, incidentID, table.TableID, platformws.ExtensionResourceChangeKindRemove, platformws.ExtensionResourceReasonSoftDeleted)
	wantIntentCount++
	collaborationsupport.RequireIntentCount(t, harness.DB, intentSelector, wantIntentCount)

	deleteReplayResp := httptestx.DoJSON(t, http.MethodDelete, tablePath, deleteBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	deleteReplay := httptestx.RequireSuccessEnvelope(t, deleteReplayResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if deleteReplay["table_version"] != deleted["table_version"] {
		t.Fatalf("unexpected delete replay payload: %#v", deleteReplay)
	}
	collaborationsupport.RequireIntentCount(t, harness.DB, intentSelector, wantIntentCount)
	requireNoNetworkFlowResourceChange(t, socket, table.TableID)
	if got := networkFlowRouteCountRows(t, harness.DB, `
SELECT COUNT(*)
  FROM deployment_admin_audit_events
 WHERE incident_id = $1
   AND event_source = 'network_flow'
   AND event_kind = 'network_flow_table_soft_deleted'
   AND after_json->>'network_flow_table_id' = $2
`, incidentID, table.TableID); got != 1 {
		t.Fatalf("expected one soft-delete audit event, got %d", got)
	}
	if got := networkFlowRouteCountRows(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'nf.tables.delete'
   AND actor_user_id = $1
   AND scope_key = $2
`, adminID, incidentID.String()+":"+table.TableID); got != 1 {
		t.Fatalf("expected one delete idempotency row, got %d", got)
	}

	staleCursorResp := httptestx.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id":    schemaTableQueryContinuationForTest,
		"cursor_token": nextToken,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, staleCursorResp, http.StatusConflict, "network_flow_table_not_active")
}

func requireNetworkFlowResourceChange(t testing.TB, socket *incidentwstest.Client, incidentID uuid.UUID, tableID string, changeKind string, reasonCode string) {
	t.Helper()
	payload := incidentwstest.RequireExtensionResourceChanged(t, socket, incidentwstest.ExtensionResourceChangeExpectation{
		IncidentID: incidentID.String(), ExtensionProfileID: ProfileID,
		ResourceKind: "network_flow_table", ResourceID: tableID,
		ChangeKind: changeKind, ReasonCode: reasonCode,
		ForbiddenKeys: []string{"display_name", "source_filename_display"},
	})
	if len(payload.WorkspaceRefs) != 1 {
		t.Fatalf("unexpected workspace_refs: %#v", payload.WorkspaceRefs)
	}
	ref := payload.WorkspaceRefs[0]
	if ref.Kind != "extension_workspace" || ref.ExtensionProfileID != ProfileID || ref.WorkspaceKey != WorkspaceKeyNetworkAnalysis {
		t.Fatalf("unexpected workspace ref: %#v", ref)
	}
}

func requireNoNetworkFlowResourceChange(t testing.TB, socket *incidentwstest.Client, tableID string) {
	t.Helper()
	incidentwstest.ExpectNoExtensionResourceChanged(t, socket, tableID)
}

func TestNetworkFlowGraphContributorsAndIndicatorLinkRoutes(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "network-flow-routes-graph-link")
	adminLogin, adminIDText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	adminID := uuid.MustParse(adminIDText)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-routes-graph-incident",
		"incident_key":  "IR-NF-GRAPH",
		"title":         "Network Flow graph",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	store := newTestNetworkFlowStore(
		t,
		harness.Pool,
		harness.Revisions.Appender(),
	)
	sessionID, unitID := seedImportSessionUnit(t, harness.Pool, incidentID, adminID, "graph-flows.csv")
	first := testFlowRow(1, "a")
	second := testFlowRow(2, "b")
	second.BytesCount = "100"
	second.PacketsCount = "5"
	table, err := store.CreateTable(context.Background(), CreateTableParams{
		IncidentID:                incidentID,
		ActorUserID:               adminID,
		ImportSessionID:           sessionID,
		ImportUnitID:              unitID,
		SourceContentSHA256:       testSHA1,
		OriginalFilename:          "graph-flows.csv",
		SourceFilenameDigest:      testSHA2,
		SourceFilenameDigestKeyID: "route-test-key",
		MappingFingerprint:        testSHA3,
		SourceProfileID:           SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID:           ParserProfileRFC4180HeaderedCSV,
		Rows:                      []FlowRow{second, first},
		Now:                       time.Date(2026, 7, 10, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create network flow table for graph routes: %v", err)
	}
	secondSessionID, secondUnitID := seedImportSessionUnit(t, harness.Pool, incidentID, adminID, "graph-flows-second.csv")
	secondTableRow := testFlowRow(3, "c")
	secondTable, err := store.CreateTable(context.Background(), CreateTableParams{
		IncidentID:                incidentID,
		ActorUserID:               adminID,
		ImportSessionID:           secondSessionID,
		ImportUnitID:              secondUnitID,
		SourceContentSHA256:       testSHA4,
		OriginalFilename:          "graph-flows-second.csv",
		SourceFilenameDigest:      strings.Repeat("5", 64),
		SourceFilenameDigestKeyID: "route-test-key",
		MappingFingerprint:        strings.Repeat("6", 64),
		SourceProfileID:           SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID:           ParserProfileRFC4180HeaderedCSV,
		Rows:                      []FlowRow{secondTableRow},
		Now:                       time.Date(2026, 7, 10, 13, 1, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create second network flow table for ordered iteration: %v", err)
	}
	visited := make([]string, 0, 3)
	if err := store.IterateRowsForTables(context.Background(), incidentID, []string{secondTable.TableID, table.TableID}, func(row FlowRow) error {
		visited = append(visited, row.NetworkFlowTableID+"/"+row.RowID)
		return nil
	}); err != nil {
		t.Fatalf("iterate ordered graph rows: %v", err)
	}
	if len(visited) != 3 || !strings.HasPrefix(visited[0], secondTable.TableID+"/") || !strings.HasPrefix(visited[1], table.TableID+"/") || !strings.HasSuffix(visited[1], strings.Repeat("a", 64)) || !strings.HasSuffix(visited[2], strings.Repeat("b", 64)) {
		t.Fatalf("ordered graph iteration = %#v", visited)
	}
	cancelContext, cancelIteration := context.WithCancel(context.Background())
	visitedBeforeCancel := 0
	err = store.IterateRowsForTables(cancelContext, incidentID, []string{table.TableID}, func(FlowRow) error {
		visitedBeforeCancel++
		cancelIteration()
		return nil
	})
	if !errors.Is(err, context.Canceled) || visitedBeforeCancel != 1 {
		t.Fatalf("cancelled graph iteration visits=%d err=%v", visitedBeforeCancel, err)
	}

	graphPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graphs/query"
	graphResp := httptestx.DoJSON(t, http.MethodPost, graphPath, map[string]any{
		"schema_id": "cartulary.network_flow.graph_query_request.v2",
		"table_scope": map[string]any{
			"mode":            "active_table",
			"active_table_id": table.TableID,
		},
		"aggregation": map[string]any{"mode": "default_flow_edge_v1"},
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	graph := httptestx.RequireSuccessEnvelope(t, graphResp, http.StatusOK)["data"].(map[string]any)
	if graph["schema_id"] != "cartulary.network_flow.graph_query_result.v2" {
		t.Fatalf("unexpected graph schema: %#v", graph)
	}
	graphDigest := graph["graph_query_digest"].(string)
	semanticQuery := graph["semantic_query"].(map[string]any)
	edgeAnnotations := graph["edge_annotations"].([]any)
	if len(edgeAnnotations) != 1 {
		t.Fatalf("expected one aggregate edge annotation, got %#v", edgeAnnotations)
	}
	edge := edgeAnnotations[0].(map[string]any)
	canonicalSelector := edge["selector"].(map[string]any)
	edgeID := canonicalSelector["source_edge_id"].(string)
	examples := edge["example_row_refs"].([]any)
	if len(examples) != 2 || edge["example_refs_truncated"] != false || edge["example_refs_total_count"] != float64(2) {
		t.Fatalf("unexpected edge examples: %#v", edge)
	}
	projection := graph["graph_projection_result"].(map[string]any)
	if projection["projection_schema_id"] != graphprojection.ProjectionSchemaIDV2 || projection["source_owner_id"] != "network_flow_activity" {
		t.Fatalf("unexpected graph projection result: %#v", projection)
	}
	for _, field := range []string{"projection_result_id", "graph_view_id", "source_snapshot_id", "projection_version", "normalized_configuration_sha256", "normalized_source_sha256", "canonical_output_sha256", "properties", "mapped_metadata", "schema_registry", "vertices", "edges", "consumer_capabilities"} {
		if _, ok := projection[field]; !ok {
			t.Fatalf("graph projection result omitted %s: %#v", field, projection)
		}
	}
	for _, removed := range []string{"state", "ephemeral_projection_id", "graph_view_key", "generated_at", "metadata"} {
		if _, ok := projection[removed]; ok {
			t.Fatalf("graph projection result retained operational field %s: %#v", removed, projection)
		}
	}
	if selectors := graph["vertex_selectors"].([]any); len(selectors) != 2 {
		t.Fatalf("default graph omitted canonical vertex selectors: %#v", selectors)
	}
	if variant := graph["result_variant"].(map[string]any); variant["kind"] != "default_flow_edge_v1" {
		t.Fatalf("default graph result variant = %#v", variant)
	}

	contributorPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graphs/contributors/query"
	contributorResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector":           canonicalSelector,
		"limit":              10,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	contributorResult := httptestx.RequireSuccessEnvelope(t, contributorResp, http.StatusOK)["data"].(map[string]any)
	contributors := contributorResult["contributors"].([]any)
	if len(contributors) != 2 {
		t.Fatalf("expected two graph contributors, got %#v", contributorResult)
	}
	firstContributor := contributors[0].(map[string]any)["row"].(map[string]any)
	if firstContributor["source_row_number"] != float64(1) {
		t.Fatalf("contributors not ordered by source row: %#v", contributors)
	}

	canonicalResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector":           canonicalSelector,
		"limit":              1,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	canonicalResult := httptestx.RequireSuccessEnvelope(t, canonicalResp, http.StatusOK)["data"].(map[string]any)
	canonicalContributors := canonicalResult["contributors"].([]any)
	paging := canonicalResult["meta"].(map[string]any)["paging"].(map[string]any)
	if len(canonicalContributors) != 1 || canonicalResult["selector"].(map[string]any)["source_edge_id"] != edgeID || paging["next_cursor_token"] == nil {
		t.Fatalf("canonical contributor first page = %#v", canonicalResult)
	}
	continuationResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":    "cartulary.network_flow.graph_contributor_query_continuation.v1",
		"cursor_token": paging["next_cursor_token"],
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	continuationResult := httptestx.RequireSuccessEnvelope(t, continuationResp, http.StatusOK)["data"].(map[string]any)
	if continued := continuationResult["contributors"].([]any); len(continued) != 1 || continued[0].(map[string]any)["row"].(map[string]any)["source_row_number"] != float64(2) {
		t.Fatalf("canonical contributor continuation = %#v", continuationResult)
	}

	vertexSelector := map[string]any{
		"kind":             "vertex",
		"source_vertex_id": EndpointID(incidentID, "ip", first.SrcIP),
		"endpoint_value":   first.SrcIP,
	}
	vertexResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector":           vertexSelector,
		"limit":              10,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	if vertexResult := httptestx.RequireSuccessEnvelope(t, vertexResp, http.StatusOK)["data"].(map[string]any); len(vertexResult["contributors"].([]any)) != 2 {
		t.Fatalf("canonical vertex contributors = %#v", vertexResult)
	}

	tamperedSelector := map[string]any{}
	for key, value := range canonicalSelector {
		tamperedSelector[key] = value
	}
	tamperedSelector["source_edge_id"] = FlowEdgeID(incidentID, EndpointID(incidentID, "ip", first.DstIP), EndpointID(incidentID, "ip", first.SrcIP), first.IPProtocol, first.DstPort)
	tamperedResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector":           tamperedSelector,
		"limit":              1,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	tamperedError := httptestx.RequireErrorEnvelope(t, tamperedResp, http.StatusBadRequest, "network_flow_invalid_request")
	if tamperedError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "id_key_mismatch" {
		t.Fatalf("tampered selector error = %#v", tamperedError)
	}
	staleDigestResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        semanticQuery,
		"graph_query_digest": strings.Repeat("0", 64),
		"selector":           canonicalSelector,
		"limit":              1,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	staleDigestError := httptestx.RequireErrorEnvelope(t, staleDigestResp, http.StatusConflict, "network_flow_graph_query_stale")
	if staleDigestError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "digest_mismatch" {
		t.Fatalf("stale contributor digest error = %#v", staleDigestError)
	}

	temporalResp := httptestx.DoJSON(t, http.MethodPost, graphPath, map[string]any{
		"schema_id": "cartulary.network_flow.graph_query_request.v2",
		"table_scope": map[string]any{
			"mode":            "active_table",
			"active_table_id": table.TableID,
		},
		"time_range": map[string]any{
			"start_utc": "2026-07-10T09:00:00Z",
			"end_utc":   "2026-07-10T09:04:00Z",
		},
		"aggregation": map[string]any{"mode": "time_bucket_v1", "bucket_width_seconds": 60},
		"limit_overrides": map[string]any{
			"max_vertices": 2, "max_edges": 2, "max_contributing_rows_per_graph": 2, "max_time_buckets_per_graph": 4,
		},
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	temporal := httptestx.RequireSuccessEnvelope(t, temporalResp, http.StatusOK)["data"].(map[string]any)
	if temporal["schema_id"] != "cartulary.network_flow.graph_query_result.v2" || temporal["graph_query_digest"] == graphDigest {
		t.Fatalf("temporal graph did not use the v2 identity boundary: %#v", temporal)
	}
	temporalWithoutOverridesResp := httptestx.DoJSON(t, http.MethodPost, graphPath, map[string]any{
		"schema_id": "cartulary.network_flow.graph_query_request.v2",
		"table_scope": map[string]any{
			"mode":            "active_table",
			"active_table_id": table.TableID,
		},
		"time_range": map[string]any{
			"start_utc": "2026-07-10T09:00:00Z",
			"end_utc":   "2026-07-10T09:04:00Z",
		},
		"aggregation": map[string]any{"mode": "time_bucket_v1", "bucket_width_seconds": 60},
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	temporalWithoutOverrides := httptestx.RequireSuccessEnvelope(t, temporalWithoutOverridesResp, http.StatusOK)["data"].(map[string]any)
	if temporalWithoutOverrides["graph_query_digest"] != temporal["graph_query_digest"] ||
		temporalWithoutOverrides["graph_projection_result"].(map[string]any)["projection_result_id"] != temporal["graph_projection_result"].(map[string]any)["projection_result_id"] {
		t.Fatalf("lower limits entered temporal identity: lowered=%#v effective=%#v", temporal, temporalWithoutOverrides)
	}
	temporalVariant := temporal["result_variant"].(map[string]any)
	bucketSummaries := temporalVariant["time_buckets"].([]any)
	if temporalVariant["kind"] != "time_bucket_v1" || len(bucketSummaries) != 4 {
		t.Fatalf("temporal result variant = %#v", temporalVariant)
	}
	wantRows := []float64{0, 1, 1, 0}
	for index, raw := range bucketSummaries {
		bucket := raw.(map[string]any)
		if bucket["contributing_row_count"] != wantRows[index] {
			t.Fatalf("temporal bucket %d = %#v", index, bucket)
		}
	}
	temporalAnnotations := temporal["edge_annotations"].([]any)
	if len(temporalAnnotations) != 2 {
		t.Fatalf("temporal graph edge annotations = %#v", temporalAnnotations)
	}
	temporalSelector := temporalAnnotations[0].(map[string]any)["selector"].(map[string]any)
	if temporalSelector["kind"] != "time_bucket_edge" || !strings.HasPrefix(temporalSelector["source_edge_id"].(string), "nfbe_") {
		t.Fatalf("temporal graph selector = %#v", temporalSelector)
	}
	temporalContributorResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        temporal["semantic_query"],
		"graph_query_digest": temporal["graph_query_digest"],
		"selector":           temporalSelector,
		"limit":              10,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	temporalContributors := httptestx.RequireSuccessEnvelope(t, temporalContributorResp, http.StatusOK)["data"].(map[string]any)["contributors"].([]any)
	if len(temporalContributors) != 1 {
		t.Fatalf("temporal selector admitted rows outside its bucket: %#v", temporalContributors)
	}
	tamperedTemporalSelector := make(map[string]any, len(temporalSelector))
	for key, value := range temporalSelector {
		tamperedTemporalSelector[key] = value
	}
	tamperedTemporalSelector["bucket_end_utc"] = "2026-07-10T09:06:00Z"
	tamperedTemporalResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        temporal["semantic_query"],
		"graph_query_digest": temporal["graph_query_digest"],
		"selector":           tamperedTemporalSelector,
		"limit":              10,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	badTemporal := httptestx.RequireErrorEnvelope(t, tamperedTemporalResp, http.StatusBadRequest, "network_flow_invalid_request")
	if badTemporal["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "id_key_mismatch" {
		t.Fatalf("tampered temporal selector error = %#v", badTemporal)
	}

	linkPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/indicator-links"
	linkBody := map[string]any{
		"schema_id":     "cartulary.network_flow.indicator_link_request.v1",
		"client_txn_id": "txn-network-flow-link-create",
		"selector": map[string]any{
			"kind":               "graph_edge",
			"graph_query":        semanticQuery,
			"graph_query_digest": graphDigest,
			"edge_id":            edgeID,
			"field_key":          "network_flow.src_ip",
		},
		"target": map[string]any{
			"mode":           "create_indicator",
			"indicator_type": "ipv4_addr",
		},
		"observation_mode":    "binding_only",
		"confirm_exact_value": "192.0.2.10",
	}
	linkResp := httptestx.DoJSON(t, http.MethodPost, linkPath, linkBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	linkResult := httptestx.RequireSuccessEnvelope(t, linkResp, http.StatusCreated)["data"].(map[string]any)
	if linkResult["schema_id"] != "cartulary.network_flow_indicator_link_result.v1" {
		t.Fatal("incorrect indicator link result contract")
	}
	if linkResult["duplicate"] != false {
		t.Fatalf("new binding reported duplicate: %#v", linkResult)
	}
	binding := linkResult["binding"].(map[string]any)
	bindingID := binding["network_flow_indicator_binding_id"].(string)
	target := binding["target_indicator_ref"].(map[string]any)
	if binding["candidate_value"] != "192.0.2.10" || target["indicator_type"] != "ipv4_addr" || target["value_kind"] != "atomic" {
		t.Fatalf("unexpected binding result: %#v", linkResult)
	}

	linkReplayResp := httptestx.DoJSON(t, http.MethodPost, linkPath, linkBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	linkReplay := httptestx.RequireSuccessEnvelope(t, linkReplayResp, http.StatusCreated)["data"].(map[string]any)
	if linkReplay["binding"].(map[string]any)["network_flow_indicator_binding_id"] != bindingID {
		t.Fatalf("indicator-link replay changed binding: %#v", linkReplay)
	}

	duplicateBody := map[string]any{}
	for key, value := range linkBody {
		duplicateBody[key] = value
	}
	duplicateBody["client_txn_id"] = "txn-network-flow-link-duplicate"
	duplicateResp := httptestx.DoJSON(t, http.MethodPost, linkPath, duplicateBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	duplicateResult := httptestx.RequireSuccessEnvelope(t, duplicateResp, http.StatusOK)["data"].(map[string]any)
	if duplicateResult["duplicate"] != true || duplicateResult["binding"].(map[string]any)["network_flow_indicator_binding_id"] != bindingID {
		t.Fatalf("duplicate link did not reuse binding: %#v", duplicateResult)
	}

	if got := networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM network_flow_indicator_bindings WHERE incident_id = $1`, incidentID); got != 1 {
		t.Fatalf("expected one persisted binding, got %d", got)
	}
	if got := networkFlowRouteCountRows(t, harness.DB, `
SELECT COUNT(*)
  FROM deployment_admin_audit_events
 WHERE incident_id = $1
   AND event_source = 'network_flow'
   AND event_kind = 'network_flow_indicator_binding_created'
`, incidentID); got != 1 {
		t.Fatalf("expected one binding-created audit event, got %d", got)
	}
	if got := networkFlowRouteCountRows(t, harness.DB, `
SELECT COUNT(*)
  FROM deployment_admin_audit_events
 WHERE incident_id = $1
   AND event_source = 'network_flow'
   AND event_kind = 'network_flow_indicator_binding_reused'
`, incidentID); got != 1 {
		t.Fatalf("expected one binding-reused audit event for new txn duplicate, got %d", got)
	}
	if got := networkFlowRouteCountRows(t, harness.DB, `
SELECT COUNT(*) FROM deployment_admin_audit_events
 WHERE incident_id = $1 AND event_source = 'network_flow'
   AND event_kind IN ('network_flow_indicator_binding_created', 'network_flow_indicator_binding_reused')
   AND (after_json::text LIKE '%192.0.2.10%' OR after_json::text LIKE '%2001:db8::1%'
        OR NOT (after_json ? 'candidate_value_digest') OR NOT (after_json ? 'candidate_value_digest_key_id'))
`, incidentID); got != 0 {
		t.Fatalf("binding audit violated candidate privacy or digest key pairing: %d", got)
	}
	t.Run("link selector target and replay matrix", func(t *testing.T) {
		post := func(body map[string]any) *http.Response {
			return httptestx.DoJSON(t, http.MethodPost, linkPath, body, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
		}
		body := func(txn string, selector map[string]any, target map[string]any, value string) map[string]any {
			return map[string]any{"schema_id": "cartulary.network_flow.indicator_link_request.v1", "client_txn_id": txn, "selector": selector, "target": target, "confirm_exact_value": value, "observation_mode": "binding_only"}
		}
		create := map[string]any{"mode": "create_indicator", "indicator_type": "ipv4_addr"}
		existing := map[string]any{"mode": "existing_indicator", "indicator_id": target["indicator_id"]}
		row := map[string]any{"kind": "row_field_value", "network_flow_table_id": table.TableID, "network_flow_row_id": first.RowID, "field_key": FieldSrcIP}
		refs := map[string]any{"kind": "row_refs", "field_key": FieldSrcIP, "row_refs": examples}
		vertex := map[string]any{"kind": "graph_vertex", "graph_query": semanticQuery, "graph_query_digest": graphDigest, "vertex_id": EndpointID(incidentID, "ip", first.SrcIP)}
		temporalVertex := map[string]any{"kind": "graph_vertex", "graph_query": temporal["semantic_query"], "graph_query_digest": temporal["graph_query_digest"], "vertex_id": EndpointID(incidentID, "ip", first.SrcIP)}
		for index, selector := range []map[string]any{refs, vertex, temporalVertex} {
			result := httptestx.RequireSuccessEnvelope(t, post(body("matrix-reuse-"+string(rune('a'+index)), selector, existing, first.SrcIP)), http.StatusOK)["data"].(map[string]any)
			if result["duplicate"] != true || !reflect.DeepEqual(result["binding"], binding) {
				t.Fatal("selector-shape reuse changed original binding metadata")
			}
		}
		bucketEdge := map[string]any{"kind": "graph_edge", "graph_query": temporal["semantic_query"], "graph_query_digest": temporal["graph_query_digest"], "edge_id": temporalSelector["source_edge_id"], "field_key": FieldSrcIP}
		httptestx.RequireErrorEnvelope(t, post(body("matrix-bucket-edge", bucketEdge, existing, first.SrcIP)), http.StatusBadRequest, "network_flow_invalid_request")
		before := networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE incident_id=$1`, incidentID)
		one := httptestx.RequireSuccessEnvelope(t, post(body("matrix-single", row, create, first.SrcIP)), http.StatusCreated)["data"].(map[string]any)
		if one["binding"].(map[string]any)["target_indicator_ref"].(map[string]any)["indicator_id"] != target["indicator_id"] || networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE incident_id=$1`, incidentID) != before {
			t.Fatal("new binding must reuse the Core-owned canonical indicator")
		}
		oneRef := map[string]any{"kind": "row_refs", "field_key": FieldSrcIP, "row_refs": examples[:1]}
		oneReuse := httptestx.RequireSuccessEnvelope(t, post(body("matrix-one-ref", oneRef, existing, first.SrcIP)), http.StatusOK)["data"].(map[string]any)
		if !reflect.DeepEqual(oneReuse["binding"], one["binding"]) {
			t.Fatal("row-ref reuse must preserve original row-field metadata")
		}
		otherRow := map[string]any{"kind": "row_field_value", "network_flow_table_id": table.TableID, "network_flow_row_id": second.RowID, "field_key": FieldSrcIP}
		other := httptestx.RequireSuccessEnvelope(t, post(body("matrix-other-row", otherRow, existing, first.SrcIP)), http.StatusCreated)["data"].(map[string]any)
		if other["binding"].(map[string]any)["network_flow_indicator_binding_id"] == one["binding"].(map[string]any)["network_flow_indicator_binding_id"] {
			t.Fatal("different source sets must remain distinct bindings")
		}
		destination := map[string]any{"kind": "graph_edge", "graph_query": semanticQuery, "graph_query_digest": graphDigest, "edge_id": edgeID, "field_key": FieldDstIP}
		dst := httptestx.RequireSuccessEnvelope(t, post(body("matrix-destination", destination, map[string]any{"mode": "create_indicator", "indicator_type": "ipv6_addr"}, first.DstIP)), http.StatusCreated)["data"].(map[string]any)["binding"].(map[string]any)
		mismatch := map[string]any{"mode": "existing_indicator", "indicator_id": dst["target_indicator_ref"].(map[string]any)["indicator_id"]}
		for index, selector := range []map[string]any{
			{"kind": "row_refs", "field_key": FieldDstIP, "row_refs": examples},
			{"kind": "graph_vertex", "graph_query": semanticQuery, "graph_query_digest": graphDigest, "vertex_id": EndpointID(incidentID, "ip", first.DstIP)},
		} {
			result := httptestx.RequireSuccessEnvelope(t, post(body(fmt.Sprintf("matrix-destination-reuse-%d", index), selector, mismatch, first.DstIP)), http.StatusOK)["data"].(map[string]any)
			if !reflect.DeepEqual(result["binding"], dst) {
				t.Fatal("destination selector reuse changed binding identity")
			}
		}
		dstRow := map[string]any{"kind": "row_field_value", "network_flow_table_id": table.TableID, "network_flow_row_id": first.RowID, "field_key": FieldDstIP}
		httptestx.RequireSuccessEnvelope(t, post(body("matrix-destination-single", dstRow, mismatch, first.DstIP)), http.StatusCreated)
		coreResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.indicators.v1/rows", map[string]any{"client_txn_id": "matrix-other-canonical", "indicator.indicator_type": "ipv4_addr", "indicator.value_kind": "atomic", "indicator.display_value": "203.0.113.77"}, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
		incompatibleID := httptestx.RequireSuccessEnvelope(t, coreResp, http.StatusCreated)["data"].(map[string]any)["row"].(map[string]any)["record_id"]
		httptestx.RequireErrorEnvelope(t, post(body("matrix-existing-value", row, map[string]any{"mode": "existing_indicator", "indicator_id": incompatibleID}, first.SrcIP)), http.StatusBadRequest, "network_flow_invalid_indicator_target")
		for _, check := range []struct {
			txn, value, code, reason string
			selector, target         map[string]any
			status                   int
		}{
			{"space", first.SrcIP + " ", "network_flow_indicator_link_ambiguous", "candidate_mismatch", row, create, 400},
			{"existing-type", first.SrcIP, "network_flow_invalid_indicator_target", "target_type_mismatch", row, mismatch, 400},
			{"hidden", first.SrcIP, "network_flow_indicator_link_forbidden", "target_not_visible", row, map[string]any{"mode": "existing_indicator", "indicator_id": uuid.New().String()}, 403},
			{"type", first.SrcIP, "network_flow_invalid_indicator_target", "target_type_mismatch", row, map[string]any{"mode": "create_indicator", "indicator_type": "ipv6_addr"}, 400},
			{"field", first.SrcIP, "network_flow_invalid_indicator_selector", "field_not_linkable", map[string]any{"kind": "row_field_value", "network_flow_table_id": table.TableID, "network_flow_row_id": first.RowID, "field_key": FieldSrcPort}, create, 400},
		} {
			err := httptestx.RequireErrorEnvelope(t, post(body("matrix-invalid-"+check.txn, check.selector, check.target, check.value)), check.status, check.code)["error"].(map[string]any)["details"].(map[string]any)
			if err["reason_code"] != check.reason {
				t.Fatalf("unexpected link reason for %s", check.txn)
			}
			for _, key := range []string{"selector_kind", "field_key", "target_mode", "resolved_candidate_value", "retry_action"} {
				if _, ok := err[key]; !ok {
					t.Fatalf("missing link detail %s", key)
				}
			}
		}
		foreignIncident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{"client_txn_id": "matrix-foreign-incident", "incident_key": "IR-NF-LINK-FOREIGN", "title": "Foreign indicator"})
		foreignPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + foreignIncident["incident_id"].(string) + "/views/cartulary.view.indicators.v1/rows"
		foreignResp := httptestx.DoJSON(t, http.MethodPost, foreignPath, map[string]any{"client_txn_id": "matrix-foreign-indicator", "indicator.indicator_type": "ipv4_addr", "indicator.value_kind": "atomic", "indicator.display_value": first.SrcIP}, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
		foreignID := httptestx.RequireSuccessEnvelope(t, foreignResp, http.StatusCreated)["data"].(map[string]any)["row"].(map[string]any)["record_id"]
		httptestx.RequireErrorEnvelope(t, post(body("matrix-foreign", row, map[string]any{"mode": "existing_indicator", "indicator_id": foreignID}, first.SrcIP)), http.StatusForbidden, "network_flow_indicator_link_forbidden")
		v6Rows := make([]FlowRow, 0, 18)
		for index := 0; index < 18; index++ {
			r := testFlowRow(int64(index+1), "d")
			r.RowID = "nfr_" + fmt.Sprintf("%064x", index+100)
			r.SrcIP = "2001:db8::10"
			r.DstIP = "2001:db8::20"
			if index == 17 {
				r.SrcIP = "2001:db8::30"
			}
			v6Rows = append(v6Rows, r)
		}
		v6Session, v6Unit := seedImportSessionUnit(t, harness.Pool, incidentID, adminID, "ipv6-link.csv")
		v6Table, err := store.CreateTable(context.Background(), CreateTableParams{IncidentID: incidentID, ActorUserID: adminID, ImportSessionID: v6Session, ImportUnitID: v6Unit, SourceContentSHA256: strings.Repeat("7", 64), OriginalFilename: "ipv6-link.csv", SourceFilenameDigest: strings.Repeat("8", 64), SourceFilenameDigestKeyID: "route-test-key", MappingFingerprint: strings.Repeat("9", 64), SourceProfileID: SourceProfileCiscoSNANetFlowCSV, ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: v6Rows, Now: time.Now().UTC()})
		if err != nil {
			t.Fatal(err)
		}
		v6GraphResp := httptestx.DoJSON(t, http.MethodPost, graphPath, map[string]any{"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": v6Table.TableID}, "aggregation": map[string]any{"mode": "default_flow_edge_v1"}}, httptestx.WithCookies(adminLogin.SessionCookie))
		v6Graph := httptestx.RequireSuccessEnvelope(t, v6GraphResp, http.StatusOK)["data"].(map[string]any)
		v6Vertex := map[string]any{"kind": "graph_vertex", "graph_query": v6Graph["semantic_query"], "graph_query_digest": v6Graph["graph_query_digest"], "vertex_id": EndpointID(incidentID, "ip", "2001:db8::10")}
		v6Create := map[string]any{"mode": "create_indicator", "indicator_type": "ipv6_addr"}
		v6Receipt := httptestx.RequireSuccessEnvelope(t, post(body("matrix-v6-truncated", v6Vertex, v6Create, "2001:db8::10")), http.StatusCreated)["data"].(map[string]any)
		v6Binding := v6Receipt["binding"].(map[string]any)
		if v6Binding["source_row_refs_truncated"] != true || v6Binding["source_row_refs_total_count"] != float64(17) || len(v6Binding["source_row_refs"].([]any)) != 16 {
			t.Fatal("graph binding lost truncated full-source count")
		}
		v6Refs := map[string]any{"kind": "row_refs", "field_key": FieldSrcIP, "row_refs": v6Binding["source_row_refs"]}
		v6Reuse := httptestx.RequireSuccessEnvelope(t, post(body("matrix-v6-reuse-prefix", v6Refs, v6Create, "2001:db8::10")), http.StatusOK)["data"].(map[string]any)
		if !reflect.DeepEqual(v6Reuse["binding"], v6Binding) {
			t.Fatal("row-ref reuse changed original graph truncation metadata")
		}
		refFor := func(index int) map[string]any {
			return map[string]any{"network_flow_table_id": v6Table.TableID, "network_flow_row_id": v6Rows[index].RowID, "source_row_number": v6Rows[index].SourceRowNumber, "mapping_fingerprint": strings.Repeat("9", 64)}
		}
		mixed := map[string]any{"kind": "row_refs", "field_key": FieldSrcIP, "row_refs": []any{refFor(0), refFor(17)}}
		httptestx.RequireErrorEnvelope(t, post(body("matrix-mixed", mixed, v6Create, "2001:db8::10")), http.StatusBadRequest, "network_flow_indicator_link_ambiguous")
		tooMany := make([]any, 17)
		for index := range tooMany {
			tooMany[index] = refFor(index)
		}
		limitError := httptestx.RequireErrorEnvelope(t, post(body("matrix-limit", map[string]any{"kind": "row_refs", "field_key": FieldSrcIP, "row_refs": tooMany}, v6Create, "2001:db8::10")), http.StatusRequestEntityTooLarge, "network_flow_resource_limit_exceeded")
		if httptestx.RequireErrorDetails(t, limitError)["reason_code"] != "row_limit_exceeded" {
			t.Fatalf("missing link source-limit reason: %#v", httptestx.RequireErrorDetails(t, limitError))
		}
		staleVertex := map[string]any{"kind": "graph_vertex", "graph_query": semanticQuery, "graph_query_digest": strings.Repeat("0", 64), "vertex_id": EndpointID(incidentID, "ip", first.SrcIP)}
		httptestx.RequireErrorEnvelope(t, post(body("matrix-stale", staleVertex, create, first.SrcIP)), http.StatusConflict, "network_flow_graph_query_stale")
		for _, race := range []struct {
			name, change, restore, code string
			status                      int
		}{
			{"closure", `UPDATE incidents SET status='closed',closed_at=now() WHERE id=$1`, `UPDATE incidents SET status='active',closed_at=NULL WHERE id=$1`, "incident_closed", 409},
			{"role", `UPDATE incident_memberships SET role='reviewer' WHERE incident_id=$1 AND user_id=$2`, `UPDATE incident_memberships SET role='admin' WHERE incident_id=$1 AND user_id=$2`, "authorization_denied", 403},
			{"source", `UPDATE network_flow_tables SET table_status='soft_deleted',deleted_at=now() WHERE incident_id=$1 AND network_flow_table_id=$2`, `UPDATE network_flow_tables SET table_status='active',deleted_at=NULL WHERE incident_id=$1 AND network_flow_table_id=$2`, "network_flow_table_not_active", 409},
		} {
			t.Run("transaction rechecks "+race.name, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback(context.Background())
				var locked uuid.UUID
				if err = tx.QueryRow(ctx, `SELECT id FROM incidents WHERE id=$1 FOR UPDATE`, incidentID).Scan(&locked); err != nil {
					t.Fatal(err)
				}
				args := []any{incidentID}
				if race.name == "role" {
					args = append(args, adminID)
				}
				if race.name == "source" {
					args = append(args, v6Table.TableID)
				}
				if _, err = tx.Exec(ctx, race.change, args...); err != nil {
					t.Fatal(err)
				}
				defer func() {
					_ = tx.Rollback(context.Background())
					if _, err := harness.Pool.Exec(context.Background(), race.restore, args...); err != nil {
						t.Error(err)
					}
				}()
				racedBody := body("matrix-race-"+race.name, map[string]any{"kind": "row_field_value", "network_flow_table_id": v6Table.TableID, "network_flow_row_id": v6Rows[17].RowID, "field_key": FieldSrcIP}, v6Create, "2001:db8::30")
				encoded, _ := json.Marshal(racedBody)
				request, err := http.NewRequestWithContext(ctx, http.MethodPost, linkPath, bytes.NewReader(encoded))
				if err != nil {
					t.Fatal(err)
				}
				request.Header.Set("Content-Type", "application/json")
				request.Header.Set(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)
				request.AddCookie(adminLogin.SessionCookie)
				request.AddCookie(adminLogin.CSRFCookie)
				type response struct {
					value *http.Response
					err   error
				}
				settled := make(chan response, 1)
				go func() { value, err := http.DefaultClient.Do(request); settled <- response{value, err} }()
				ticker := time.NewTicker(10 * time.Millisecond)
				defer ticker.Stop()
				for {
					var waiting bool
					err = harness.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, int32(tx.Conn().PgConn().PID())).Scan(&waiting)
					if err != nil {
						t.Fatal(err)
					}
					if waiting {
						break
					}
					select {
					case <-ctx.Done():
						t.Fatal("link never reached the transaction lock")
					case result := <-settled:
						if result.value != nil {
							result.value.Body.Close()
						}
						t.Fatalf("link settled before transaction admission: %v", result.err)
					case <-ticker.C:
					}
				}
				if err = tx.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				result := <-settled
				if result.err != nil {
					t.Fatal(result.err)
				}
				httptestx.RequireErrorEnvelope(t, result.value, race.status, race.code)
				if networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency WHERE client_txn_id=$1`, racedBody["client_txn_id"]) != 0 {
					t.Fatal("rejected transaction retained a link receipt")
				}
				if networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE incident_id=$1 AND normalized_value='2001:db8::30'`, incidentID) != 0 {
					t.Fatal("rejected transaction committed a Core indicator")
				}
			})
		}
		t.Run("audit failure rolls back the complete link transaction", func(t *testing.T) {
			// A test-database constraint fails the real audit insert after Core
			// resolution and binding insertion, without a production fault hook.
			if _, err := harness.DB.Exec(`ALTER TABLE deployment_admin_audit_events ADD CONSTRAINT indicator_link_audit_failure CHECK (client_txn_id IS DISTINCT FROM 'matrix-audit-rollback') NOT VALID`); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if _, err := harness.DB.Exec(`ALTER TABLE deployment_admin_audit_events DROP CONSTRAINT indicator_link_audit_failure`); err != nil {
					t.Error(err)
				}
			}()
			bindings := networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM network_flow_indicator_bindings WHERE incident_id=$1`, incidentID)
			audits := networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE incident_id=$1`, incidentID)
			selector := map[string]any{"kind": "row_field_value", "network_flow_table_id": v6Table.TableID, "network_flow_row_id": v6Rows[17].RowID, "field_key": FieldSrcIP}
			httptestx.RequireErrorEnvelope(t, post(body("matrix-audit-rollback", selector, v6Create, "2001:db8::30")), http.StatusInternalServerError, "internal_error")
			if networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM network_flow_indicator_bindings WHERE incident_id=$1`, incidentID) != bindings || networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM deployment_admin_audit_events WHERE incident_id=$1`, incidentID) != audits || networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE incident_id=$1 AND normalized_value='2001:db8::30'`, incidentID) != 0 || networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency WHERE client_txn_id=$1`, "matrix-audit-rollback") != 0 {
				t.Fatal("failed audit retained an indicator, binding, receipt or audit occurrence")
			}
		})
		for _, role := range []string{"viewer", "reviewer", "editor", "admin"} {
			if _, err := harness.Pool.Exec(context.Background(), "UPDATE incident_memberships SET role=$3 WHERE incident_id=$1 AND user_id=$2", incidentID, adminID, role); err != nil {
				t.Fatal(err)
			}
			status := http.StatusCreated
			if role == "viewer" || role == "reviewer" {
				status = http.StatusForbidden
			}
			httptestx.RequireStatus(t, post(linkBody), status)
		}
		if _, err := harness.Pool.Exec(context.Background(), `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, target["indicator_id"], adminID); err != nil {
			t.Fatal(err)
		}
		httptestx.RequireErrorEnvelope(t, post(linkBody), http.StatusForbidden, "network_flow_indicator_link_forbidden")
		if _, err := harness.Pool.Exec(context.Background(), `UPDATE records SET deleted_at=NULL,deleted_by_user_id=NULL WHERE record_id=$1`, target["indicator_id"]); err != nil {
			t.Fatal(err)
		}
		if networkFlowRouteCountRows(t, harness.DB, `SELECT COUNT(*) FROM indicator_observations WHERE incident_id=$1`, incidentID) != 0 {
			t.Fatal("binding-only linking created Core observations")
		}
	})
	if err := ValidateRetainedExtensionState(context.Background(), harness.Pool); err != nil {
		t.Fatalf("current link receipts fail read-only admission: %v", err)
	}
	t.Run("incompatible receipt blocks admission without rewriting", func(t *testing.T) {
		tx, err := harness.Pool.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(context.Background())
		if _, err := tx.Exec(context.Background(), `UPDATE route_idempotency SET response_json = jsonb_set(response_json, '{schema_id}', '"cartulary.network_flow.indicator_link_result.v1"') WHERE client_txn_id = $1`, linkBody["client_txn_id"]); err != nil {
			t.Fatal(err)
		}
		read := func() string {
			var value string
			if err := tx.QueryRow(context.Background(), `SELECT response_json::text FROM route_idempotency WHERE client_txn_id = $1`, linkBody["client_txn_id"]).Scan(&value); err != nil {
				t.Fatal(err)
			}
			return value
		}
		before := read()
		if err := ValidateRetainedExtensionState(context.Background(), tx); err == nil {
			t.Fatal("old receipt admitted")
		}
		if after := read(); after != before {
			t.Fatal("admission rewrote retained receipt")
		}
	})
	if _, err := store.SoftDeleteTable(context.Background(), SoftDeleteTableParams{
		IncidentID: incidentID, ActorUserID: adminID, TableID: table.TableID,
		BaseTableVersion: table.TableVersion, ClientTxnID: "txn-network-flow-contributor-source-delete",
		RequestID: "req-network-flow-contributor-source-delete", Now: time.Date(2026, 7, 10, 14, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("soft-delete contributor source table: %v", err)
	}
	deletedSourceResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector":           canonicalSelector,
		"limit":              1,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	deletedSourceError := httptestx.RequireErrorEnvelope(t, deletedSourceResp, http.StatusConflict, "network_flow_table_not_active")
	if deletedSourceError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "soft_deleted" {
		t.Fatalf("deleted contributor source error = %#v", deletedSourceError)
	}
	replayAfterRemoval := httptestx.DoJSON(t, http.MethodPost, linkPath, linkBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	if receipt := httptestx.RequireSuccessEnvelope(t, replayAfterRemoval, http.StatusCreated)["data"]; !reflect.DeepEqual(receipt, linkResult) {
		t.Fatal("committed replay changed after source removal")
	}
	duplicateBody["client_txn_id"] = "txn-network-flow-link-after-removal"
	rejectedAfterRemoval := httptestx.DoJSON(t, http.MethodPost, linkPath, duplicateBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, rejectedAfterRemoval, http.StatusConflict, "network_flow_table_not_active")
	if _, err := harness.Pool.Exec(context.Background(), `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminID); err != nil {
		t.Fatalf("revoke contributor incident membership: %v", err)
	}
	deniedResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v2",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector":           canonicalSelector,
		"limit":              1,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, deniedResp, http.StatusNotFound, "incident_not_found")
}

func TestNetworkFlowTimeBucketSavedGraphLifecycle_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "network-flow-temporal-saved-graph")
	adminLogin, adminIDText := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	adminID := uuid.MustParse(adminIDText)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-temporal-saved-incident",
		"incident_key":  "IR-NF-TEMPORAL-SAVED",
		"title":         "Network Flow temporal saved graph",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	store := newTestNetworkFlowStore(t, harness.Pool, harness.Revisions.Appender())
	sessionID, unitID := seedImportSessionUnit(t, harness.Pool, incidentID, adminID, "temporal-saved.csv")
	first := testFlowRow(1, "a")
	second := testFlowRow(2, "b")
	third := testFlowRow(3, "c")
	third.FlowStartUTC = first.FlowStartUTC
	third.FlowEndUTC = first.FlowEndUTC
	table, err := store.CreateTable(context.Background(), CreateTableParams{
		IncidentID: incidentID, ActorUserID: adminID, ImportSessionID: sessionID, ImportUnitID: unitID,
		SourceContentSHA256: testSHA1, OriginalFilename: "temporal-saved.csv",
		SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "route-test-key",
		MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{third, second, first},
		Now: time.Date(2026, 7, 10, 12, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create temporal saved-graph source table: %v", err)
	}

	collectionPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graph-views"
	mutationOptions := []func(*http.Request){
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	}
	unsupportedCreate := map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_create_request.v1", "client_txn_id": "txn-temporal-unsupported-rejected", "display_name": "Unsupported graph",
		"semantic_query": map[string]any{"schema_id": unsupportedGraphSemanticQuerySchemaID()},
	}
	unsupportedResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, unsupportedCreate, mutationOptions...)
	httptestx.RequireErrorEnvelope(t, unsupportedResp, http.StatusBadRequest, "network_flow_invalid_request")

	createBody := map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_create_request.v3", "client_txn_id": "txn-temporal-saved-create", "display_name": "Temporal graph",
		"semantic_query": map[string]any{
			"schema_id": "cartulary.network_flow.graph_semantic_query.v2", "selected_table_ids": []string{table.TableID}, "filters": []any{},
			"time_range":  map[string]any{"start_utc": "2026-07-10T09:00:00Z", "end_utc": "2026-07-10T09:04:00Z"},
			"aggregation": map[string]any{"mode": "time_bucket_v1", "bucket_width_seconds": 60, "include_example_row_refs": true},
		},
	}
	createResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	created := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusAccepted)["data"].(map[string]any)
	graphViewID := created["graph_view"].(map[string]any)["graph_view_id"].(string)
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, created["job"].(map[string]any)["job_id"].(string), "succeeded")

	resourcePath := collectionPath + "/" + graphViewID
	resultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	resource := httptestx.RequireSuccessEnvelope(t, resultResp, http.StatusOK)["data"].(map[string]any)
	result := resource["result"].(map[string]any)
	projection := result["graph_projection_result"].(map[string]any)
	assertRestoreSourceComposition(t, harness.Pool, graphViewID, projection)
	query := createBody["semantic_query"].(map[string]any)
	ephemeralResponse := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/graphs/query", map[string]any{
		"schema_id": "cartulary.network_flow.graph_query_request.v2", "table_scope": map[string]any{"mode": "active_table", "active_table_id": table.TableID}, "aggregation": query["aggregation"], "time_range": query["time_range"],
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	ephemeral := httptestx.RequireSuccessEnvelope(t, ephemeralResponse, http.StatusOK)["data"].(map[string]any)
	if ephemeral["graph_query_digest"] != result["graph_query_digest"] || !reflect.DeepEqual(ephemeral["source_tables"], result["source_tables"]) {
		t.Fatalf("temporal HTTP and worker source composition differs: %#v / %#v", ephemeral, result)
	}

	if resource["schema_id"] != "cartulary.network_flow.graph_view_result.v4" || result["schema_id"] != "cartulary.network_flow.graph_query_result.v2" || projection["projection_version"] != "network_flow_activity.time_bucket.v1" {
		t.Fatalf("temporal saved result contract = %#v", resource)
	}
	buckets := result["result_variant"].(map[string]any)["time_buckets"].([]any)
	if len(buckets) != 4 || buckets[0].(map[string]any)["contributing_row_count"] != float64(0) || buckets[3].(map[string]any)["contributing_row_count"] != float64(0) {
		t.Fatalf("temporal saved bucket index = %#v", buckets)
	}
	annotations := result["edge_annotations"].([]any)
	if len(annotations) != 2 {
		t.Fatalf("temporal saved annotations = %#v", annotations)
	}
	var selector map[string]any
	for _, raw := range annotations {
		candidate := raw.(map[string]any)["selector"].(map[string]any)
		if candidate["bucket_start_utc"] == "2026-07-10T09:01:00Z" {
			selector = candidate
			break
		}
	}
	if selector == nil {
		t.Fatalf("temporal saved result omitted the populated 09:01 bucket selector: %#v", annotations)
	}
	contributorsResp := httptestx.DoJSON(t, http.MethodPost, resourcePath+"/contributors/query", map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_contributor_query_request.v2", "projection_result_id": projection["projection_result_id"],
		"selector": selector, "limit": 1,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	contributorResult := httptestx.RequireSuccessEnvelope(t, contributorsResp, http.StatusOK)["data"].(map[string]any)
	contributors := contributorResult["contributors"].([]any)
	page := contributorResult["meta"].(map[string]any)["paging"].(map[string]any)
	if len(contributors) != 1 || page["next_cursor_token"] == nil {
		t.Fatalf("temporal saved contributor first page = %#v", contributorResult)
	}
	continuationResp := httptestx.DoJSON(t, http.MethodPost, resourcePath+"/contributors/query", map[string]any{
		"schema_id": "cartulary.network_flow.graph_contributor_query_continuation.v1", "cursor_token": page["next_cursor_token"],
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	continuationResult := httptestx.RequireSuccessEnvelope(t, continuationResp, http.StatusOK)["data"].(map[string]any)
	if continued := continuationResult["contributors"].([]any); len(continued) != 1 || continuationResult["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"] != nil {
		t.Fatalf("temporal saved contributor continuation = %#v", continuationResult)
	}

	refreshResp := httptestx.DoJSON(t, http.MethodPost, resourcePath+"/refresh", map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "txn-temporal-saved-refresh", "base_graph_view_version": 1,
	}, mutationOptions...)
	refreshed := httptestx.RequireSuccessEnvelope(t, refreshResp, http.StatusAccepted)["data"].(map[string]any)
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, refreshed["job"].(map[string]any)["job_id"].(string), "succeeded")
	refreshedResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	refreshedProjection := httptestx.RequireSuccessEnvelope(t, refreshedResp, http.StatusOK)["data"].(map[string]any)["result"].(map[string]any)["graph_projection_result"].(map[string]any)
	if refreshedProjection["projection_result_id"] != projection["projection_result_id"] {
		t.Fatalf("temporal saved retry changed deterministic identity: first=%v refreshed=%v", projection["projection_result_id"], refreshedProjection["projection_result_id"])
	}
}

func TestNetworkFlowSavedGraphLifecycleRoutes_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := claimedNetworkFlowServerForRouteTest(t, runtime, "network-flow-saved-graph-routes")
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
	// Fail real transaction participants, rather than arming an unconsumed hook.
	migrationDB, err := pgtest.OpenPurposeDatabase(harness.Database.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = migrationDB.Close() })
	for _, failureTable := range []string{"network_flow_graph_views", "collaboration_intent", "jobs", "route_idempotency", "deployment_admin_audit_events"} {
		t.Run("command rollback at "+failureTable, func(t *testing.T) {
			ctx := context.Background()
			const snapshot = `SELECT jsonb_build_array(
			 (SELECT count(*) FROM network_flow_graph_views),
			 (SELECT count(*) FROM jobs WHERE job_kind = 'network_flow_activity.graph_view_materialize_v1'),
			 (SELECT count(*) FROM route_idempotency WHERE route_key LIKE 'nf.graph_views.%'),
			 (SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind = 'network_flow_graph_view_created'))::text`
			var before, after string
			if err := harness.Pool.QueryRow(ctx, snapshot).Scan(&before); err != nil {
				t.Fatal(err)
			}
			intentSelector := collaborationsupport.IntentSelector{IncidentID: incidentID.String(), EventFamily: "extension_resource_changed"}
			intentsBefore := collaborationsupport.CountIntents(t, harness.Pool, intentSelector)
			if failureTable == "collaboration_intent" {
				collaborationsupport.FailIntentInserts(t, migrationDB)
			} else {
				if _, err := migrationDB.ExecContext(ctx, `CREATE OR REPLACE FUNCTION nf_command_failure() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected saved graph participant failure'; END $$`); err != nil {
					t.Fatal(err)
				}
				quoted := pgx.Identifier{failureTable}.Sanitize()
				if _, err := migrationDB.ExecContext(ctx, `CREATE TRIGGER nf_command_failure BEFORE INSERT ON `+quoted+` FOR EACH ROW EXECUTE FUNCTION nf_command_failure()`); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					if _, err := migrationDB.ExecContext(ctx, `DROP TRIGGER nf_command_failure ON `+quoted); err != nil {
						t.Error(err)
					}
				})
			}

			response := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
			httptestx.RequireErrorEnvelope(t, response, http.StatusInternalServerError, "internal_error")
			if err := harness.Pool.QueryRow(ctx, snapshot).Scan(&after); err != nil {
				t.Fatal(err)
			}
			if before != after || collaborationsupport.CountIntents(t, harness.Pool, intentSelector) != intentsBefore {
				t.Fatalf("partial command commit at %s: %s -> %s", failureTable, before, after)
			}
		})
	}
	createResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	created := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusAccepted)["data"].(map[string]any)
	graphView := created["graph_view"].(map[string]any)
	graphViewID := graphView["graph_view_id"].(string)
	jobID := created["job"].(map[string]any)["job_id"].(string)
	if graphView["graph_view_version"] != float64(1) || graphView["materialization_generation"] != float64(1) || graphView["latest_job_id"] != jobID || created["job"].(map[string]any)["status_route"] != "/api/v1/jobs/"+jobID {
		t.Fatalf("unexpected created graph view: %#v", graphView)
	}
	replayResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	replayed := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusAccepted)["data"].(map[string]any)
	if !reflect.DeepEqual(replayed, created) {
		t.Fatalf("create replay drifted: %#v", replayed)
	}

	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, jobID, "succeeded")
	terminalReplayResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	terminalReplay := httptestx.RequireSuccessEnvelope(t, terminalReplayResp, http.StatusAccepted)["data"].(map[string]any)
	if !reflect.DeepEqual(terminalReplay, created) {
		t.Fatalf("terminal create replay drifted: %#v", terminalReplay)
	}
	resourcePath := collectionPath + "/" + graphViewID
	// Exercise authorization before request decoding without admitting additional work.
	for _, target := range []struct {
		method, path string
		body         any
	}{
		{http.MethodGet, collectionPath, nil},
		{http.MethodGet, resourcePath, nil},
		{http.MethodGet, resourcePath + "/result", nil},
		{http.MethodPost, collectionPath, createBody},
		{http.MethodPatch, resourcePath, map[string]any{"schema_id": "cartulary.network_flow.graph_view_rename_request.v2", "client_txn_id": "invalid-query-rename", "base_graph_view_version": 1, "display_name": "Invalid query"}},
		{http.MethodDelete, resourcePath, map[string]any{"schema_id": "cartulary.network_flow.graph_view_retire_request.v1", "client_txn_id": "invalid-query-retire", "base_graph_view_version": 1}},
		{http.MethodPost, resourcePath + "/refresh", map[string]any{"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "invalid-query-refresh", "base_graph_view_version": 1}},
		{http.MethodPost, resourcePath + "/contributors/query", map[string]any{}},
	} {
		resp := httptestx.DoJSON(t, target.method, target.path+"?undeclared=1", target.body, mutationOptions...)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "network_flow_invalid_request")
	}
	for _, role := range []string{"viewer", "editor", "reviewer", "admin"} {
		if _, err := harness.Pool.Exec(context.Background(), "UPDATE incident_memberships SET role = $3 WHERE incident_id = $1 AND user_id = $2", incidentID, adminID, role); err != nil {
			t.Fatal(err)
		}
		httptestx.RequireStatus(t, httptestx.DoJSON(t, http.MethodGet, collectionPath, nil, httptestx.WithCookies(adminLogin.SessionCookie)), http.StatusOK)
		for _, action := range []struct {
			method, path string
			allowed      bool
		}{
			{http.MethodPost, collectionPath, role == "editor" || role == "admin"},
			{http.MethodPatch, resourcePath, role == "editor" || role == "admin"},
			{http.MethodPost, resourcePath + "/refresh", role == "editor" || role == "admin"},
			{http.MethodDelete, resourcePath, role == "reviewer" || role == "admin"},
		} {
			status := http.StatusForbidden
			if action.allowed {
				status = http.StatusBadRequest
			}
			httptestx.RequireStatus(t, httptestx.DoJSON(t, action.method, action.path, map[string]any{}, mutationOptions...), status)
			httptestx.RequireStatus(t, httptestx.DoJSON(t, action.method, action.path+"?undeclared=1", map[string]any{}, mutationOptions...), status)
		}
	}
	if _, err := harness.Pool.Exec(context.Background(), "UPDATE incidents SET status = 'closed', closed_at = now() WHERE id = $1", incidentID); err != nil {
		t.Fatal(err)
	}
	for _, action := range []struct{ method, path string }{{http.MethodGet, collectionPath}, {http.MethodGet, resourcePath}, {http.MethodGet, resourcePath + "/result"}, {http.MethodPost, collectionPath}, {http.MethodPatch, resourcePath}, {http.MethodPost, resourcePath + "/refresh"}, {http.MethodDelete, resourcePath}} {
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, action.method, action.path, map[string]any{}, mutationOptions...), http.StatusConflict, "incident_closed")
	}
	if _, err := harness.Pool.Exec(context.Background(), "UPDATE incidents SET status = 'active', closed_at = NULL WHERE id = $1", incidentID); err != nil {
		t.Fatal(err)
	}
	resultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	result := httptestx.RequireSuccessEnvelope(t, resultResp, http.StatusOK)["data"].(map[string]any)
	projection := result["result"].(map[string]any)["graph_projection_result"].(map[string]any)
	if projection["projection_result_id"] == "" || projection["graph_view_id"] != graphViewID || projection["source_owner_id"] != ProfileID {
		t.Fatalf("saved graph result binding drifted: %#v", projection)
	}
	assertRestoreSourceComposition(t, harness.Pool, graphViewID, projection)
	vertices := projection["vertices"].([]any)
	if len(vertices) == 0 {
		t.Fatalf("saved graph result omitted vertices: %#v", projection)
	}
	beforeVertexIDs, beforeEdgeIDs := projectionObjectIDs(t, projection)
	unsupportedGraphViewID := "nfgv_00000000000000000000000000000090"
	unsupportedSemanticQuery := unsupportedDefaultGraphSemanticQuery(table.TableID)
	unsupportedDeclaration := graphViewDeclarationFixture(unsupportedGraphViewID, incidentID, adminID, time.Date(2026, 7, 10, 12, 31, 0, 0, time.UTC))
	unsupportedDeclaration.DisplayName = "Installed unsupported flow graph"
	unsupportedDeclaration.NormalizedDisplayName = "installed unsupported flow graph"
	unsupportedDeclaration.SemanticQueryJSON = unsupportedSemanticQuery
	unsupportedDeclaration.SemanticQuerySHA256 = GraphViewSemanticQuerySHA256(unsupportedSemanticQuery)
	unsupportedDeclaration.DesiredSourceSnapshotID = "pre-refresh-unsupported-placeholder"
	unsupportedTx, err := harness.Pool.Begin(context.Background())
	if err != nil {
		t.Fatalf("begin unsupported declaration transaction: %v", err)
	}
	if err := store.InsertGraphViewDeclarationTx(context.Background(), unsupportedTx, unsupportedDeclaration); err != nil {
		_ = unsupportedTx.Rollback(context.Background())
		t.Fatalf("insert unsupported declaration: %v", err)
	}
	if err := unsupportedTx.Commit(context.Background()); err != nil {
		t.Fatalf("commit unsupported declaration: %v", err)
	}
	unsupportedBytesBefore := persistedGraphSemanticBytes(t, harness.Pool, incidentID)
	unsupportedResourcePath := collectionPath + "/" + unsupportedGraphViewID
	unsupportedRefreshResp := httptestx.DoJSON(t, http.MethodPost, unsupportedResourcePath+"/refresh", map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "txn-installed-v1-refresh", "base_graph_view_version": 1,
	}, mutationOptions...)
	httptestx.RequireErrorEnvelope(t, unsupportedRefreshResp, http.StatusInternalServerError, "internal_error")
	unsupportedBytesAfter := persistedGraphSemanticBytes(t, harness.Pool, incidentID)
	if !slices.Equal(unsupportedBytesBefore, unsupportedBytesAfter) {
		t.Fatalf("unsupported refresh rejection changed saved declaration bytes: before=%q after=%q", unsupportedBytesBefore, unsupportedBytesAfter)
	}
	if _, err := harness.Pool.Exec(context.Background(), `DELETE FROM network_flow_graph_views WHERE graph_view_id = $1`, unsupportedGraphViewID); err != nil {
		t.Fatalf("remove intentionally unsupported declaration before current restore proof: %v", err)
	}
	reportingJobID := seedRestoredReportingGraphJob(t, harness, incidentID, adminID, projection)
	networkFlowJobID := seedRestoredNetworkFlowGraphJob(t, harness, incidentID, adminID, projection)
	t.Run("NF restore changes only attempt fields and rolls back", func(t *testing.T) {
		ctx := context.Background()
		var before, after []byte
		const snapshot = `SELECT (to_jsonb(jobs) - 'handler_attempt_id' - 'handler_lease_expires_at')::text FROM jobs WHERE job_id = $1`
		if err := harness.Pool.QueryRow(ctx, snapshot, networkFlowJobID).Scan(&before); err != nil {
			t.Fatal(err)
		}
		tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		for range 2 {
			count, err := ReconcileGraphRestoreJobsTx(ctx, tx)
			if err != nil || count != 1 {
				t.Fatalf("reconcile selected count: %d %v", count, err)
			}
		}
		if err := tx.QueryRow(ctx, snapshot, networkFlowJobID).Scan(&after); err != nil || string(before) != string(after) {
			t.Fatalf("non-attempt state changed: %s / %s / %v", before, after, err)
		}
		var attempt *uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT handler_attempt_id FROM jobs WHERE job_id = $1`, networkFlowJobID).Scan(&attempt); err != nil || attempt != nil {
			t.Fatalf("attempt not cleared: %v %v", attempt, err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		if err := harness.Pool.QueryRow(ctx, `SELECT handler_attempt_id FROM jobs WHERE job_id = $1`, networkFlowJobID).Scan(&attempt); err != nil || attempt == nil {
			t.Fatalf("rollback lost attempt: %v %v", attempt, err)
		}
	})

	otherIncident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-nf-restore-other-incident", "incident_key": "IR-NF-RESTORE-OTHER", "title": "Other restored jobs",
	})
	otherIncidentID := uuid.MustParse(otherIncident["incident_id"].(string))
	for _, count := range []int{0, 1, 256, 257, 512, 513} {
		t.Run(fmt.Sprintf("restore complete selected set %d", count), func(t *testing.T) {
			ctx := context.Background()
			tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if _, err := tx.Exec(ctx, `INSERT INTO jobs SELECT (jsonb_populate_record(NULL::jobs, to_jsonb(source) || jsonb_build_object('job_id', ('00000000-0000-0000-0001-' || lpad(n::text, 12, '0'))::uuid, 'status', (ARRAY['queued','running','cancel_requested'])[1 + n % 3], 'cancelable', false, 'incident_id', CASE WHEN n % 2 = 0 THEN $3::uuid ELSE source.incident_id END, 'handler_payload_json', source.handler_payload_json || jsonb_build_object('incident_id', CASE WHEN n % 2 = 0 THEN $3::uuid ELSE source.incident_id END)))).* FROM jobs source CROSS JOIN generate_series(1, $2::int) n WHERE source.job_id = $1 ORDER BY n DESC`, networkFlowJobID, count, otherIncidentID); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `UPDATE jobs SET extension_owner_profile_id = 'restore_neighbor' WHERE job_id = $1`, networkFlowJobID); err != nil {
				t.Fatal(err)
			}
			// Exclusion is independent in each scope dimension and terminal status.
			if _, err := tx.Exec(ctx, `INSERT INTO jobs SELECT (jsonb_populate_record(NULL::jobs, to_jsonb(source) || jsonb_build_object(
			 'job_id', ('00000000-0000-0000-0002-' || lpad(n::text, 12, '0'))::uuid,
			 'extension_owner_profile_id', $2::text, 'status', (ARRAY['succeeded','failed','canceled'])[n],
			 'cancelable', false, 'handler_attempt_id', NULL, 'handler_lease_expires_at', NULL,
			 'progress_completed', 1, 'finished_at', now(), 'retained_until', now()+interval '1 day',
			 'result_summary_json', CASE WHEN n = 2 THEN NULL ELSE '{}'::jsonb END,
			 'error_summary_json', CASE WHEN n = 2 THEN '{}'::jsonb ELSE NULL END))).*
			 FROM jobs source CROSS JOIN generate_series(1,3) n WHERE source.job_id = $1`, networkFlowJobID, ProfileID); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO jobs SELECT (jsonb_populate_record(NULL::jobs, to_jsonb(source) || jsonb_build_object('job_id', '00000000-0000-0000-0002-000000000004', 'extension_owner_profile_id', $2::text, 'job_kind', 'other.kind.v1'))).* FROM jobs source WHERE source.job_id = $1`, networkFlowJobID, ProfileID); err != nil {
				t.Fatal(err)
			}
			var cursor *uuid.UUID
			selected := 0
			for {
				page, err := jobs.ListRestoredNonterminalPageTx(ctx, tx, jobs.RestoredNonterminalScope{JobKind: GraphViewMaterializationJobKind, ExtensionOwnerProfileID: ProfileID}, cursor, 256)
				if err != nil {
					t.Fatal(err)
				}
				if len(page) == 0 {
					break
				}
				for _, job := range page {
					selected++
					if job.JobID.String() != fmt.Sprintf("00000000-0000-0000-0001-%012d", selected) {
						t.Fatalf("restore page scope/order: %s at %d", job.JobID, selected)
					}
				}
				last := page[len(page)-1].JobID
				cursor = &last
			}
			if selected != count {
				t.Fatalf("restored page count=%d want=%d", selected, count)
			}

			var before, after string
			const snapshot = `SELECT COALESCE(jsonb_agg(to_jsonb(j) - 'handler_attempt_id' - 'handler_lease_expires_at' ORDER BY job_id), '[]')::text FROM jobs j`
			if err := tx.QueryRow(ctx, snapshot).Scan(&before); err != nil {
				t.Fatal(err)
			}
			if count == 513 {
				for _, defect := range []string{
					"handler_payload_json = '{}'::jsonb",
					"handler_payload_json = handler_payload_json || '{\"unknown\":true}'::jsonb",
					"handler_payload_json = handler_payload_json || '{\"schema_id\":\"invalid\"}'::jsonb",
					"handler_payload_json = handler_payload_json || '{\"graph_view_id\":\"invalid\"}'::jsonb",
					"handler_payload_json = handler_payload_json || '{\"materialization_generation\":0}'::jsonb",
					"handler_payload_json = handler_payload_json || '{\"source_snapshot_id\":\"\"}'::jsonb",
					"handler_payload_json = handler_payload_json || '{\"incident_id\":\"00000000-0000-0000-0000-000000000001\"}'::jsonb",
					"handler_failure_count = 3",
				} {
					failedTx, err := tx.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := failedTx.Exec(ctx, `UPDATE jobs SET `+defect+` WHERE job_id = '00000000-0000-0000-0001-000000000513'`); err != nil {
						t.Fatal(err)
					}
					if got, err := ReconcileGraphRestoreJobsTx(ctx, failedTx); err == nil || got != 0 {
						t.Fatalf("later page failure returned success: %d %v", got, err)
					}
					var cleared bool
					if err := failedTx.QueryRow(ctx, `SELECT handler_attempt_id IS NULL FROM jobs WHERE job_id = '00000000-0000-0000-0001-000000000001'`).Scan(&cleared); err != nil || !cleared {
						t.Fatalf("fixture did not reach earlier-page writes: %v", err)
					}
					if err := failedTx.Rollback(ctx); err != nil {
						t.Fatal(err)
					}
					if err := tx.QueryRow(ctx, `SELECT handler_attempt_id IS NULL FROM jobs WHERE job_id = '00000000-0000-0000-0001-000000000001'`).Scan(&cleared); err != nil || cleared {
						t.Fatalf("later-page failure lost atomic rollback: %v", err)
					}
				}
			}

			for range 2 {
				got, err := ReconcileGraphRestoreJobsTx(ctx, tx)
				if err != nil || got != count {
					t.Fatalf("selected count=%d want=%d err=%v", got, count, err)
				}
			}
			if err := tx.QueryRow(ctx, snapshot).Scan(&after); err != nil || before != after {
				t.Fatalf("non-attempt state changed: %v", err)
			}
			var live int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM jobs WHERE extension_owner_profile_id = $1 AND job_kind = $2 AND handler_attempt_id IS NOT NULL`, ProfileID, GraphViewMaterializationJobKind).Scan(&live); err != nil || live != 0 {
				t.Fatalf("unreconciled jobs=%d err=%v", live, err)
			}
		})
	}

	recoveryDSN, err := harness.Database.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatalf("resolve Recovery-purpose DSN: %v", err)
	}
	recoveryHandle, err := postgres.Setup(context.Background(), postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          recoveryDSN,
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatalf("open admitted Recovery-purpose pool: %v", err)
	}
	t.Cleanup(recoveryHandle.Close)
	restoreParticipant, err := recoveryassembly.NewGraphProjectionRestoreParticipant(recoveryHandle.Pool())
	if err != nil {
		t.Fatalf("construct active saved-graph restore participant: %v", err)
	}
	recoveryCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatalf("construct active saved-graph recovery catalog: %v", err)
	}
	currentRegistry := graphrestore.CurrentRestoreSourceRegistry()
	registryRef := graphrestore.RestoreSourceRegistryRef{Registry: currentRegistry, SHA256: currentRegistry.DigestSHA256()}
	bindingRef := graphrestore.CurrentRestoreImplementationBinding()
	if recoveryCatalog.DigestSHA256() != bindingRef.Binding.RecoveryStateCatalogSHA256 ||
		bindingRef.Binding.AlgorithmID != graphrestore.RestoreAlgorithmID ||
		!slices.Equal(bindingRef.Binding.GraphTableIDs, graphrestore.RestoreGraphTableIDs()) {
		t.Fatalf("current Graph restore catalog/binding tuple drifted: catalog=%s binding=%#v", recoveryCatalog.DigestSHA256(), bindingRef.Binding)
	}
	restoreRequest := graphrestore.RestoreRebuildRequest{
		Context:             context.Background(),
		RestoreOperationID:  uuid.MustParse("00000000-0000-0000-0000-000000009101"),
		RestoredSourceState: restorecontract.RestoredGraphProjectionSourceState{},
		BackupSetID:         uuid.MustParse("00000000-0000-0000-0000-000000009102"),
		ConsistencyPointAt:  time.Date(2026, 7, 10, 12, 45, 0, 0, time.UTC),
		TargetGenerationID:  uuid.MustParse("00000000-0000-0000-0000-000000009103"),
		RecoveryStateCatalog: graphrestore.RestoreRecoveryCatalogRef{
			DigestSHA256: recoveryCatalog.DigestSHA256(), AlgorithmID: graphrestore.RestoreAlgorithmID,
			GraphTableIDs: graphrestore.RestoreGraphTableIDs(),
		},
		SourceRegistry: registryRef, ImplementationBinding: bindingRef,
	}
	// Reporting runs after NF reconciliation in the real atomic restore writer.
	var reportingPayload []byte
	if err := harness.Pool.QueryRow(context.Background(), `SELECT request_json FROM reporting_job_payloads WHERE job_id = $1`, reportingJobID).Scan(&reportingPayload); err != nil {
		t.Fatal(err)
	}
	if _, err := harness.Pool.Exec(context.Background(), `UPDATE reporting_job_payloads SET request_json = '{"graph_projection_refs":[{}]}'::jsonb WHERE job_id = $1`, reportingJobID); err != nil {
		t.Fatal(err)
	}
	const graphStateSnapshot = `SELECT jsonb_build_array((SELECT jsonb_agg(r ORDER BY projection_result_id) FROM graph_projection_results r), (SELECT jsonb_agg(l ORDER BY lease_id) FROM graph_projection_result_leases l))::text`
	var graphBefore, graphAfter string
	if err := harness.Pool.QueryRow(context.Background(), graphStateSnapshot).Scan(&graphBefore); err != nil {
		t.Fatal(err)
	}
	failedRestore, restoreErr := restoreParticipant.Rebuild(context.Background(), restoreRequest)
	if restoreErr == nil || failedRestore.ReadinessSatisfied() {
		t.Fatalf("failed Reporting reconciliation published readiness: %#v %v", failedRestore, restoreErr)
	}
	if err := harness.Pool.QueryRow(context.Background(), graphStateSnapshot).Scan(&graphAfter); err != nil || graphBefore != graphAfter {
		t.Fatalf("failed restore changed Graph state: %v", err)
	}
	var preservedAttempt *uuid.UUID
	if err := harness.Pool.QueryRow(context.Background(), `SELECT handler_attempt_id FROM jobs WHERE job_id = $1`, networkFlowJobID).Scan(&preservedAttempt); err != nil || preservedAttempt == nil {
		t.Fatalf("whole restore failed to roll back NF reconciliation: %v %v", preservedAttempt, err)
	}
	if _, err := harness.Pool.Exec(context.Background(), `UPDATE reporting_job_payloads SET request_json = $2 WHERE job_id = $1`, reportingJobID, reportingPayload); err != nil {
		t.Fatal(err)
	}
	restoreResult, err := restoreParticipant.Rebuild(context.Background(), restoreRequest)
	if err != nil || !restoreResult.ReadinessSatisfied() || len(restoreResult.RebuiltViews) != 1 ||
		restoreResult.ReconciledNonterminalJobCount != 2 || restoreResult.ReconciledLeaseCount != 1 ||
		!restoreContainsExactGraphBinding(restoreResult.RebuiltViews, graphViewID, projection["projection_result_id"].(string)) {
		t.Fatalf("active saved-graph restore did not reproduce exact identity: result=%#v err=%v", restoreResult, err)
	}
	var restoredLeaseCount int
	var restoredAttemptID *uuid.UUID
	if err := harness.Pool.QueryRow(context.Background(), `
SELECT
    (SELECT COUNT(*) FROM graph_projection_result_leases
      WHERE projection_result_id = $1
        AND lease_owner_id = 'snapshot_reporting'
        AND lease_owner_resource_id = $2
        AND lease_purpose = 'render'),
    (SELECT handler_attempt_id FROM jobs WHERE job_id = $3)
`, projection["projection_result_id"], reportingJobID.String(), reportingJobID).Scan(&restoredLeaseCount, &restoredAttemptID); err != nil || restoredLeaseCount != 1 || restoredAttemptID != nil {
		t.Fatalf("restore did not reconcile Reporting lease and Common Job attempt: leases=%d attempt=%v err=%v", restoredLeaseCount, restoredAttemptID, err)
	}
	if err := harness.Pool.QueryRow(context.Background(), `SELECT handler_attempt_id FROM jobs WHERE job_id = $1`, networkFlowJobID).Scan(&restoredAttemptID); err != nil || restoredAttemptID != nil {
		t.Fatalf("restore did not reconcile Network Flow Common Job attempt: attempt=%v err=%v", restoredAttemptID, err)
	}
	restoredResultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	restoredProjection := httptestx.RequireSuccessEnvelope(t, restoredResultResp, http.StatusOK)["data"].(map[string]any)["result"].(map[string]any)["graph_projection_result"].(map[string]any)
	afterVertexIDs, afterEdgeIDs := projectionObjectIDs(t, restoredProjection)
	if restoredProjection["projection_result_id"] != projection["projection_result_id"] ||
		strings.Join(afterVertexIDs, ",") != strings.Join(beforeVertexIDs, ",") || strings.Join(afterEdgeIDs, ",") != strings.Join(beforeEdgeIDs, ",") {
		t.Fatalf("restored saved graph object identity drifted: before=%v/%v after=%v/%v", beforeVertexIDs, beforeEdgeIDs, afterVertexIDs, afterEdgeIDs)
	}
	selectedVertex := vertices[0].(map[string]any)
	sourceEndpointID := selectedVertex["source_entity_ref"].(map[string]any)["source_entity_id"].(string)
	sourceEndpointValue := selectedVertex["properties"].(map[string]any)["endpoint_value"].(string)
	contributorsResp := httptestx.DoJSON(t, http.MethodPost, resourcePath+"/contributors/query", map[string]any{
		"schema_id":            "cartulary.network_flow.graph_view_contributor_query_request.v2",
		"projection_result_id": projection["projection_result_id"],
		"selector":             map[string]any{"kind": "vertex", "source_vertex_id": sourceEndpointID, "endpoint_value": sourceEndpointValue}, "limit": 10,
	}, httptestx.WithCookies(adminLogin.SessionCookie))
	contributors := httptestx.RequireSuccessEnvelope(t, contributorsResp, http.StatusOK)["data"].(map[string]any)["contributors"].([]any)
	if len(contributors) != 1 {
		t.Fatalf("saved graph contributors = %#v", contributors)
	}

	renameBody := map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_rename_request.v2", "client_txn_id": "txn-network-flow-graph-rename",
		"base_graph_view_version": 1, "display_name": "Renamed flow graph",
	}
	renameResp := httptestx.DoJSON(t, http.MethodPatch, resourcePath, renameBody, mutationOptions...)
	renamed := httptestx.RequireSuccessEnvelope(t, renameResp, http.StatusOK)["data"].(map[string]any)["graph_view"].(map[string]any)
	if renamed["graph_view_version"] != float64(2) || renamed["materialization_generation"] != float64(1) || renamed["selected_result_binding"] == nil {
		t.Fatalf("rename changed materialization identity: %#v", renamed)
	}

	noOpResp := httptestx.DoJSON(t, http.MethodPatch, resourcePath, map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_rename_request.v2", "client_txn_id": "txn-graph-noop",
		"base_graph_view_version": 2, "display_name": "  Renamed flow graph  ",
	}, mutationOptions...)
	noOp := httptestx.RequireSuccessEnvelope(t, noOpResp, http.StatusOK)["data"].(map[string]any)["graph_view"]
	if !reflect.DeepEqual(noOp, renamed) {
		t.Fatalf("same normalized name changed declaration: %#v", noOp)
	}
	renamedReplay := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...), http.StatusAccepted)["data"]
	if !reflect.DeepEqual(renamedReplay, created) {
		t.Fatal("rename altered original creation receipt")
	}

	refreshBody := map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "txn-network-flow-graph-refresh",
		"base_graph_view_version": 2,
	}
	refreshResp := httptestx.DoJSON(t, http.MethodPost, resourcePath+"/refresh", refreshBody, mutationOptions...)
	refreshed := httptestx.RequireSuccessEnvelope(t, refreshResp, http.StatusAccepted)["data"].(map[string]any)
	refreshedGraph := refreshed["graph_view"].(map[string]any)
	if refreshedGraph["graph_view_version"] != float64(3) || refreshedGraph["materialization_generation"] != float64(2) || refreshedGraph["selected_result_binding"] == nil {
		t.Fatalf("refresh did not preserve last-safe result: %#v", refreshedGraph)
	}
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, refreshed["job"].(map[string]any)["job_id"].(string), "succeeded")
	refreshedResultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	refreshedResult := httptestx.RequireSuccessEnvelope(t, refreshedResultResp, http.StatusOK)["data"].(map[string]any)["result"].(map[string]any)["graph_projection_result"].(map[string]any)
	if refreshedResult["projection_result_id"] != projection["projection_result_id"] {
		t.Fatalf("semantic retry produced a different immutable result: first=%v refreshed=%v", projection["projection_result_id"], refreshedResult["projection_result_id"])
	}

	tablePath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/tables/" + table.TableID
	deleteTableResp := httptestx.DoJSON(t, http.MethodDelete, tablePath, map[string]any{
		"client_txn_id": "txn-network-flow-saved-graph-source-delete", "base_table_version": 1,
	}, mutationOptions...)
	httptestx.RequireSuccessEnvelope(t, deleteTableResp, http.StatusOK)
	invalidatedResp := httptestx.DoJSON(t, http.MethodGet, resourcePath, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	invalidated := httptestx.RequireSuccessEnvelope(t, invalidatedResp, http.StatusOK)["data"].(map[string]any)["graph_view"].(map[string]any)
	if invalidated["graph_view_version"] != float64(4) || invalidated["materialization_generation"] != float64(3) || invalidated["selected_result_binding"] != nil || invalidated["last_failure_code"] != "network_flow_source_table_deleted" {
		t.Fatalf("source retirement did not invalidate saved graph: %#v", invalidated)
	}
	invalidatedResultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, invalidatedResultResp, http.StatusConflict, "network_flow_graph_view_not_materialized")

	retireBody := map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_retire_request.v1", "client_txn_id": "txn-network-flow-graph-retire",
		"base_graph_view_version": 4,
	}
	for range 2 {
		retireResp := httptestx.DoJSON(t, http.MethodDelete, resourcePath, retireBody, mutationOptions...)
		httptestx.RequireStatus(t, retireResp, http.StatusNoContent)
		body, err := io.ReadAll(retireResp.Body)
		_ = retireResp.Body.Close()
		if err != nil || len(body) != 0 {
			t.Fatalf("retirement must be bodyless: %q %v", body, err)
		}
	}
	retired, err := store.GetGraphViewDeclaration(context.Background(), incidentID, graphViewID)
	if err != nil || retired.DeclarationState != GraphViewDeclarationStateRetired || retired.GraphViewVersion != 5 || retired.MaterializationGeneration != 4 || retired.SelectedResult != nil {
		t.Fatalf("retired declaration drifted: %#v %v", retired, err)
	}
	getRetired := httptestx.DoJSON(t, http.MethodGet, resourcePath, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, getRetired, http.StatusNotFound, "network_flow_graph_view_not_found")
	retiredReplay := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...), http.StatusAccepted)["data"]
	if !reflect.DeepEqual(retiredReplay, created) {
		t.Fatal("retirement altered original creation receipt")
	}
	if _, err := harness.Pool.Exec(context.Background(), `UPDATE jobs SET retained_until = $2 WHERE job_id = $1`, uuid.MustParse(jobID), time.Now().UTC().Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	expired := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, expired, http.StatusNotFound, "job_not_found")
	expiredReplay := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...), http.StatusAccepted)["data"]
	if !reflect.DeepEqual(expiredReplay, created) {
		t.Fatal("job expiry altered original creation receipt")
	}
	intentSelector := collaborationsupport.IntentSelector{IncidentID: incidentID.String(), SourceIdentity: "network_flow_graph_view:" + graphViewID}
	intentCount := collaborationsupport.CountIntents(t, harness.Pool, intentSelector)
	intentSelector.PayloadReasonCode = "renamed"
	renamedCount := collaborationsupport.CountIntents(t, harness.Pool, intentSelector)
	intentSelector.PayloadReasonCode = ""
	intentSelector.PayloadChangeKind = "remove"
	removalCount := collaborationsupport.CountIntents(t, harness.Pool, intentSelector)
	if intentCount != 7 || renamedCount != 1 || removalCount != 1 {
		t.Fatalf("saved graph outbox duplicated or omitted lifecycle effects: total=%d renamed=%d removed=%d", intentCount, renamedCount, removalCount)
	}

}

func projectionObjectIDs(t testing.TB, projection map[string]any) ([]string, []string) {
	t.Helper()
	vertices, ok := projection["vertices"].([]any)
	if !ok {
		t.Fatalf("projection vertices have unexpected shape: %#v", projection["vertices"])
	}
	edges, ok := projection["edges"].([]any)
	if !ok {
		t.Fatalf("projection edges have unexpected shape: %#v", projection["edges"])
	}
	vertexIDs := make([]string, 0, len(vertices))
	for _, raw := range vertices {
		vertex, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("projection vertex has unexpected shape: %#v", raw)
		}
		vertexIDs = append(vertexIDs, vertex["vertex_id"].(string))
	}
	edgeIDs := make([]string, 0, len(edges))
	for _, raw := range edges {
		edge, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("projection edge has unexpected shape: %#v", raw)
		}
		edgeIDs = append(edgeIDs, edge["edge_id"].(string))
	}
	return vertexIDs, edgeIDs
}

func restoreContainsExactGraphBinding(views []graphrestore.RestoreRebuiltView, graphViewID string, projectionResultID string) bool {
	for _, view := range views {
		if view.GraphViewID == graphViewID && view.ProjectionResultID == projectionResultID {
			return true
		}
	}
	return false
}

func seedRestoredReportingGraphJob(
	t testing.TB,
	harness *appsupport.ServerHarness,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	projection map[string]any,
) uuid.UUID {
	t.Helper()
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}
	admission, err := jobs.NewExtensionJobAdmission(
		reporting.ProfileID,
		jobs.NewRouteIdempotencyKey("POST /api/v1/releases", actorID, incidentID.String(), "txn-network-flow-restore-report"),
		scope,
		[]byte(`{"graph_projection_refs":"restore-reconciliation"}`),
	)
	if err != nil {
		t.Fatalf("construct restored Reporting job admission: %v", err)
	}
	resource, err := harness.Jobs.Create(context.Background(), jobs.EnqueueParams{
		JobKind: reporting.ReleaseCreateJobKind, Scope: scope, SubmittedByUserID: actorID,
		Cancelable: true, Progress: jobs.Progress{Completed: 0}, Extension: admission,
	})
	if err != nil {
		t.Fatalf("create restored Reporting Graph job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)
	graphRefs := []map[string]any{{
		"source_owner_id": projection["source_owner_id"], "graph_view_id": projection["graph_view_id"],
		"projection_result_id": projection["projection_result_id"], "source_snapshot_id": projection["source_snapshot_id"],
		"projection_schema_id": projection["projection_schema_id"], "projection_version": projection["projection_version"],
		"normalized_configuration_sha256": projection["normalized_configuration_sha256"],
		"normalized_source_sha256":        projection["normalized_source_sha256"],
		"canonical_output_sha256":         projection["canonical_output_sha256"],
	}}
	payload, err := json.Marshal(map[string]any{"graph_projection_refs": graphRefs})
	if err != nil {
		t.Fatalf("encode restored Reporting Graph payload: %v", err)
	}
	now := time.Date(2026, 7, 10, 12, 40, 0, 0, time.UTC)
	if _, err := harness.Pool.Exec(context.Background(), `
INSERT INTO reporting_job_payloads (job_id, job_kind, incident_id, actor_user_id, request_json, created_at, updated_at)
VALUES ($1, 'release_create', $2, $3, $4::jsonb, $5, $5)
`, jobID, incidentID, actorID, payload, now); err != nil {
		t.Fatalf("persist restored Reporting Graph payload: %v", err)
	}
	if _, err := harness.Pool.Exec(context.Background(), `
UPDATE jobs
   SET status = 'running', started_at = $2, updated_at = $2,
       handler_attempt_id = $3, handler_lease_expires_at = $4
 WHERE job_id = $1
`, jobID, now, uuid.MustParse("00000000-0000-0000-0000-000000009104"), time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("seed restored Reporting execution lease: %v", err)
	}
	return jobID
}

func seedRestoredNetworkFlowGraphJob(
	t testing.TB,
	harness *appsupport.ServerHarness,
	incidentID uuid.UUID,
	actorID uuid.UUID,
	projection map[string]any,
) uuid.UUID {
	t.Helper()
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}
	admission, err := jobs.NewExtensionJobAdmission(
		ProfileID,
		jobs.NewRouteIdempotencyKey("POST /api/v1/network-flow/graph-views/refresh", actorID, projection["graph_view_id"].(string), "txn-network-flow-restore-job"),
		scope,
		[]byte(`{"graph_view":"restore-reconciliation"}`),
	)
	if err != nil {
		t.Fatalf("construct restored Network Flow job admission: %v", err)
	}
	payload, err := json.Marshal(map[string]any{
		"schema_id":   "cartulary.network_flow.graph_view_materialization_payload.v1",
		"incident_id": incidentID, "graph_view_id": projection["graph_view_id"],
		"materialization_generation": 1, "source_snapshot_id": projection["source_snapshot_id"],
	})
	if err != nil {
		t.Fatalf("encode restored Network Flow graph payload: %v", err)
	}
	total := 1
	resource, err := harness.Jobs.Create(context.Background(), jobs.EnqueueParams{
		JobKind: GraphViewMaterializationJobKind, Scope: scope, SubmittedByUserID: actorID,
		AuthPolicy: jobs.AuthPolicyIncidentMembership, Cancelable: true,
		Progress: jobs.Progress{Completed: 0, Total: &total}, HandlerPayload: payload, Extension: admission,
	})
	if err != nil {
		t.Fatalf("create restored Network Flow graph job: %v", err)
	}
	jobID := uuid.MustParse(resource.JobID)
	now := time.Date(2026, 7, 10, 12, 41, 0, 0, time.UTC)
	if _, err := harness.Pool.Exec(context.Background(), `
UPDATE jobs
   SET status = 'running', started_at = $2, updated_at = $2,
       handler_attempt_id = $3, handler_lease_expires_at = $4
 WHERE job_id = $1
`, jobID, now, uuid.MustParse("00000000-0000-0000-0000-000000009105"), time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("seed restored Network Flow execution lease: %v", err)
	}
	return jobID
}

func waitForNetworkFlowJob(t testing.TB, serverURL string, login flowtest.LoginResult, jobID string, wantStatus string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(t, http.MethodGet, serverURL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		last = data
		if data["status"] == wantStatus {
			return
		}
		if data["status"] == "failed" || data["status"] == "canceled" {
			t.Fatalf("job %s reached terminal status %#v", jobID, data)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach %s; last=%#v", jobID, wantStatus, last)
}

const (
	schemaTableQueryRequestForTest      = "cartulary.network_flow.table_query_request.v1"
	schemaTableQueryContinuationForTest = "cartulary.network_flow.table_query_continuation.v1"
)

func claimedNetworkFlowServerForRouteTest(
	t testing.TB,
	runtime *appsupport.Runtime,
	prefix string,
) *appsupport.ServerHarness {
	return claimedNetworkFlowServerWithLimitsForRouteTest(t, runtime, prefix, "")
}

func claimedNetworkFlowServerWithLimitsForRouteTest(
	t testing.TB,
	runtime *appsupport.Runtime,
	prefix string,
	resourceLimits string,
) *appsupport.ServerHarness {
	t.Helper()
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	rings, err := ParseKeyRings([]byte(`{
  "schema_id":"cartulary.network_flow_key_rings.v1",
  "cursor_key_ring":{"algorithm":"aes_256_gcm_v1","keys":[{"cursor_key_id":"route-cursor-v1","state":"active","secret_ref":{"kind":"env","name":"route-cursor"}}]},
  "safe_digest_key_ring":{"algorithm":"hmac_sha256_v1","keys":[{"safe_digest_key_id":"route-safe-v1","state":"active","secret_ref":{"kind":"env","name":"route-safe"}}]}
}`), map[string]string{
		"CARTULARY_SECRET_ROUTE_CURSOR": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_ROUTE_SAFE":   "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}, now)
	if err != nil {
		t.Fatalf("parse Network Flow route-test key rings: %v", err)
	}
	environment := map[string]string{
		"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
		"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
		"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
	}
	if resourceLimits != "" {
		environment["CARTULARY__NETWORK_FLOW_ACTIVITY__RESOURCE_LIMITS"] = resourceLimits
	}
	return runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: prefix,
		Env:    environment,
		Dependencies: httpapi.DependencySet{
			ModuleOverrides: map[string]any{KeyRingsOverrideKey: rings},
			Now:             func() time.Time { return now },
		},
		TestRouteMode: httptestx.TestRouteModeDisabled,
	})
}

func networkFlowRouteCountRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

// Exercise the real NF source contribution and Graph engine, without a fake
// enumerator. The full atomic Recovery writer is exercised by the lifecycle test.
func assertRestoreSourceComposition(t *testing.T, db postgres.DB, graphViewID string, expected map[string]any) {
	t.Helper()
	registration, err := NewGraphRestoreSourceRegistration(db)
	if err != nil {
		t.Fatal(err)
	}
	candidates, err := registration.Enumerate(context.Background(), nil, time.Now())
	if err != nil || len(candidates) != 1 {
		t.Fatalf("real restore candidates: %#v %v", candidates, err)
	}
	candidate := candidates[0]
	rebuilt, err := graphprojection.ProjectV2(context.Background(), graphprojection.InvocationContextV2{GraphViewID: graphViewID, SourceOwnerID: ProfileID}, candidate.SemanticInput)
	if err != nil || rebuilt.ResultBindingV2() != candidate.ExpectedBinding || rebuilt.ResultBindingV2().ProjectionResultID != expected["projection_result_id"] {
		t.Fatalf("HTTP/worker/restore composition identity differs: %#v %v", rebuilt, err)
	}

	// Enumerate through one borrowed transaction: out-of-order selected declarations
	// are deterministic; inactive and unselected declarations never enter restore.
	tx, err := db.BeginTx(context.Background(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	registration, err = NewGraphRestoreSourceRegistration(restoreSourceTestDB{Tx: tx})
	if err != nil {
		t.Fatal(err)
	}
	wantedIDs := []string{graphViewID}
	for _, n := range []int{4, 2, 3, 1} {
		id := "nfgv_" + strings.Repeat(fmt.Sprint(n), 32)
		projected, err := graphprojection.ProjectV2(context.Background(), graphprojection.InvocationContextV2{GraphViewID: id, SourceOwnerID: ProfileID}, candidate.SemanticInput)
		if err != nil {
			t.Fatal(err)
		}
		binding := projected.ResultBindingV2()
		overrides := map[string]any{
			"graph_view_id": id, "display_name": fmt.Sprintf("Restore clone %d", n), "normalized_display_name": fmt.Sprintf("restore clone %d", n),
			"selected_projection_result_id": binding.ProjectionResultID, "selected_source_snapshot_id": binding.SourceSnapshotID, "selected_projection_schema_id": binding.ProjectionSchemaID, "selected_projection_version": binding.ProjectionVersion,
			"selected_normalized_configuration_sha256": binding.NormalizedConfigurationSHA256, "selected_normalized_source_sha256": binding.NormalizedSourceSHA256, "selected_canonical_output_sha256": binding.CanonicalOutputSHA256,
		}
		if n == 4 {
			overrides["declaration_state"], overrides["retired_at"] = "retired", time.Now().UTC()
		}
		if n == 3 {
			for key := range overrides {
				if strings.HasPrefix(key, "selected_") {
					overrides[key] = nil
				}
			}
		}
		if n < 3 {
			wantedIDs = append(wantedIDs, id)
		}
		encoded, err := json.Marshal(overrides)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(context.Background(), `INSERT INTO network_flow_graph_views SELECT (jsonb_populate_record(NULL::network_flow_graph_views, to_jsonb(v) || $2::jsonb)).* FROM network_flow_graph_views v WHERE graph_view_id = $1`, graphViewID, encoded); err != nil {
			t.Fatal(err)
		}
	}
	slices.Sort(wantedIDs)
	ordered, err := registration.Enumerate(context.Background(), nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	actualIDs := make([]string, 0, len(ordered))
	for _, item := range ordered {
		actualIDs = append(actualIDs, item.GraphViewID)
	}
	if !slices.Equal(actualIDs, wantedIDs) {
		t.Fatalf("restore scope/order: %v want %v", actualIDs, wantedIDs)
	}
	for _, column := range []string{"desired_source_snapshot_id", "selected_source_snapshot_id"} {
		if _, err := tx.Exec(context.Background(), "UPDATE network_flow_graph_views SET "+column+" = 'mismatch' WHERE graph_view_id = $1", graphViewID); err != nil {
			t.Fatal(err)
		}
		if partial, err := registration.Enumerate(context.Background(), nil, time.Now()); err == nil || partial != nil {
			t.Fatalf("restore accepted snapshot mismatch: %#v %v", partial, err)
		}
		if _, err := tx.Exec(context.Background(), "UPDATE network_flow_graph_views SET "+column+" = $2 WHERE graph_view_id = $1", graphViewID, candidate.ExpectedBinding.SourceSnapshotID); err != nil {
			t.Fatal(err)
		}
	}

}

// Any unexpected nested transaction is a fixture failure, not silently tolerated.
type restoreSourceTestDB struct{ pgx.Tx }

func (restoreSourceTestDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	panic("restore source opened a transaction")
}
