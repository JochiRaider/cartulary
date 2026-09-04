package networkflow_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

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
)

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
		"schema_id": "cartulary.network_flow.graph_view_create_request.v2", "client_txn_id": "txn-temporal-saved-create", "display_name": "Temporal graph",
		"semantic_query": map[string]any{
			"schema_id": "cartulary.network_flow.graph_semantic_query.v2", "selected_table_ids": []string{table.TableID}, "filters": []any{},
			"time_range":  map[string]any{"start_utc": "2026-07-10T09:00:00Z", "end_utc": "2026-07-10T09:04:00Z"},
			"aggregation": map[string]any{"mode": "time_bucket_v1", "bucket_width_seconds": 60, "include_example_row_refs": true},
		},
	}
	createResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	created := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusAccepted)["data"].(map[string]any)
	graphViewID := created["graph_view"].(map[string]any)["graph_view_id"].(string)
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, created["job_id"].(string), "succeeded")

	resourcePath := collectionPath + "/" + graphViewID
	resultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	resource := httptestx.RequireSuccessEnvelope(t, resultResp, http.StatusOK)["data"].(map[string]any)
	result := resource["result"].(map[string]any)
	projection := result["graph_projection_result"].(map[string]any)
	if resource["schema_id"] != "cartulary.network_flow.graph_view_result.v3" || result["schema_id"] != "cartulary.network_flow.graph_query_result.v2" || projection["projection_version"] != "network_flow_activity.time_bucket.v1" {
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
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, refreshed["job_id"].(string), "succeeded")
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
		"schema_id":     "cartulary.network_flow.graph_view_create_request.v2",
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
	jobID := created["job_id"].(string)
	if graphView["graph_view_version"] != float64(1) || graphView["materialization_generation"] != float64(1) || graphView["last_materialization_status"] != "queued" {
		t.Fatalf("unexpected created graph view: %#v", graphView)
	}
	replayResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	replayed := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusAccepted)["data"].(map[string]any)
	if replayed["job_id"] != jobID || replayed["graph_view"].(map[string]any)["graph_view_id"] != graphViewID {
		t.Fatalf("create replay drifted: %#v", replayed)
	}

	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, jobID, "succeeded")
	terminalReplayResp := httptestx.DoJSON(t, http.MethodPost, collectionPath, createBody, mutationOptions...)
	terminalReplay := httptestx.RequireSuccessEnvelope(t, terminalReplayResp, http.StatusAccepted)["data"].(map[string]any)
	if terminalReplay["job_id"] != jobID || terminalReplay["graph_view"].(map[string]any)["graph_view_id"] != graphViewID {
		t.Fatalf("terminal create replay drifted: %#v", terminalReplay)
	}
	resourcePath := collectionPath + "/" + graphViewID
	resultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	result := httptestx.RequireSuccessEnvelope(t, resultResp, http.StatusOK)["data"].(map[string]any)
	projection := result["result"].(map[string]any)["graph_projection_result"].(map[string]any)
	if projection["projection_result_id"] == "" || projection["graph_view_id"] != graphViewID || projection["source_owner_id"] != ProfileID {
		t.Fatalf("saved graph result binding drifted: %#v", projection)
	}
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
	restoreResult, err := restoreParticipant.Rebuild(context.Background(), graphrestore.RestoreRebuildRequest{
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
	})
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
		"schema_id": "cartulary.network_flow.graph_view_rename_request.v1", "client_txn_id": "txn-network-flow-graph-rename",
		"base_graph_view_version": 1, "display_name": "Renamed flow graph",
	}
	renameResp := httptestx.DoJSON(t, http.MethodPatch, resourcePath, renameBody, mutationOptions...)
	renamed := httptestx.RequireSuccessEnvelope(t, renameResp, http.StatusOK)["data"].(map[string]any)["graph_view"].(map[string]any)
	if renamed["graph_view_version"] != float64(2) || renamed["materialization_generation"] != float64(1) || renamed["selected_result"] == nil {
		t.Fatalf("rename changed materialization identity: %#v", renamed)
	}

	refreshBody := map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_refresh_request.v1", "client_txn_id": "txn-network-flow-graph-refresh",
		"base_graph_view_version": 2,
	}
	refreshResp := httptestx.DoJSON(t, http.MethodPost, resourcePath+"/refresh", refreshBody, mutationOptions...)
	refreshed := httptestx.RequireSuccessEnvelope(t, refreshResp, http.StatusAccepted)["data"].(map[string]any)
	refreshedGraph := refreshed["graph_view"].(map[string]any)
	if refreshedGraph["graph_view_version"] != float64(3) || refreshedGraph["materialization_generation"] != float64(2) || refreshedGraph["selected_result"] == nil {
		t.Fatalf("refresh did not preserve last-safe result: %#v", refreshedGraph)
	}
	waitForNetworkFlowJob(t, harness.Server.HTTP.URL, adminLogin, refreshed["job_id"].(string), "succeeded")
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
	if invalidated["graph_view_version"] != float64(4) || invalidated["materialization_generation"] != float64(3) || invalidated["selected_result"] != nil || invalidated["last_failure_code"] != "network_flow_source_table_deleted" {
		t.Fatalf("source retirement did not invalidate saved graph: %#v", invalidated)
	}
	invalidatedResultResp := httptestx.DoJSON(t, http.MethodGet, resourcePath+"/result", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, invalidatedResultResp, http.StatusConflict, "network_flow_graph_view_not_materialized")

	retireResp := httptestx.DoJSON(t, http.MethodDelete, resourcePath, map[string]any{
		"schema_id": "cartulary.network_flow.graph_view_retire_request.v1", "client_txn_id": "txn-network-flow-graph-retire",
		"base_graph_view_version": 4,
	}, mutationOptions...)
	retired := httptestx.RequireSuccessEnvelope(t, retireResp, http.StatusOK)["data"].(map[string]any)["graph_view"].(map[string]any)
	if retired["state"] != "retired" || retired["graph_view_version"] != float64(5) || retired["materialization_generation"] != float64(4) || retired["selected_result"] != nil {
		t.Fatalf("retired graph view drifted: %#v", retired)
	}
	getRetired := httptestx.DoJSON(t, http.MethodGet, resourcePath, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, getRetired, http.StatusConflict, "network_flow_graph_view_not_active")
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
`, jobID, now, uuid.MustParse("00000000-0000-0000-0000-000000009104"), now.Add(time.Hour)); err != nil {
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
`, jobID, now, uuid.MustParse("00000000-0000-0000-0000-000000009105"), now.Add(time.Hour)); err != nil {
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
