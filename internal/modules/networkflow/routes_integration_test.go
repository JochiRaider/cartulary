package networkflow_test

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestNetworkFlowRoutesRemainUnclaimedByDefault(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "network-flow-routes-unclaimed")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-routes-unclaimed-incident",
		"incident_key":  "IR-NF-UNCLAIMED",
		"title":         "Network Flow unclaimed",
	})
	incidentID := incident["incident_id"].(string)

	resp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/network-flow/source-profiles", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "extension_profile_not_claimed")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["profile_id"] != ProfileID {
		t.Fatalf("unexpected unclaimed extension details: %#v", details)
	}
}

func TestNetworkFlowRoutesQueryPageAndInvalidateAfterSoftDelete(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServerWithDependencies(t, "network-flow-routes-query", httpapi.DependencySet{
		ExtensionProfiles: claimedNetworkFlowProfilesForRouteTest(),
	})
	adminLogin, adminIDText := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	adminID := uuid.MustParse(adminIDText)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-routes-query-incident",
		"incident_key":  "IR-NF-QUERY",
		"title":         "Network Flow query",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	store := NewStore(harness.Server.Runtime.Postgres)
	sessionID, unitID := seedImportSessionUnit(t, harness.Server.Runtime.Postgres, incidentID, adminID, "flows.csv")
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

	sourceProfilesResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/source-profiles", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	sourceProfiles := httptestx.RequireSuccessEnvelope(t, sourceProfilesResp, http.StatusOK)["data"].(map[string]any)
	if sourceProfiles["schema_id"] != "cartulary.network_flow.source_profile_list.v1" {
		t.Fatalf("unexpected source profiles payload: %#v", sourceProfiles)
	}

	listResp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/tables", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	list := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	tables := list["tables"].([]any)
	if len(tables) != 1 || tables[0].(map[string]any)["network_flow_table_id"] != table.TableID {
		t.Fatalf("unexpected table list: %#v", list)
	}

	queryPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/tables/" + table.TableID + "/query"
	firstPageResp := phase2test.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id": schemaTableQueryRequestForTest,
		"sort": []map[string]any{
			{"field_key": "source_row_number", "direction": "asc"},
		},
		"limit": 1,
	}, phase2test.WithCookies(adminLogin.SessionCookie))
	firstPage := httptestx.RequireSuccessEnvelope(t, firstPageResp, http.StatusOK)["data"].(map[string]any)
	firstRows := firstPage["rows"].([]any)
	if len(firstRows) != 1 || firstRows[0].(map[string]any)["source_row_number"] != float64(1) {
		t.Fatalf("unexpected first query page: %#v", firstPage)
	}
	nextToken := firstPage["meta"].(map[string]any)["paging"].(map[string]any)["next_cursor_token"].(string)
	if !strings.HasPrefix(nextToken, "nfc1.master-derived-v1.") {
		t.Fatalf("expected Network Flow cursor token with key id, got %q", nextToken)
	}

	secondPageResp := phase2test.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id":    schemaTableQueryContinuationForTest,
		"cursor_token": nextToken,
	}, phase2test.WithCookies(adminLogin.SessionCookie))
	secondPage := httptestx.RequireSuccessEnvelope(t, secondPageResp, http.StatusOK)["data"].(map[string]any)
	secondRows := secondPage["rows"].([]any)
	if len(secondRows) != 1 || secondRows[0].(map[string]any)["network_flow.src_ip"] != "198.51.100.7" {
		t.Fatalf("unexpected continuation page: %#v", secondPage)
	}

	filterResp := phase2test.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id": schemaTableQueryRequestForTest,
		"filters": []map[string]any{
			{"field_key": "network_flow.src_ip", "op": "eq", "value": "198.51.100.7"},
		},
	}, phase2test.WithCookies(adminLogin.SessionCookie))
	filtered := httptestx.RequireSuccessEnvelope(t, filterResp, http.StatusOK)["data"].(map[string]any)
	filteredRows := filtered["rows"].([]any)
	if len(filteredRows) != 1 || filteredRows[0].(map[string]any)["source_row_number"] != float64(2) {
		t.Fatalf("unexpected filtered query: %#v", filtered)
	}

	rejectedResp := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/network-flow/tables/"+table.TableID+"/rejected-rows/query", map[string]any{
		"schema_id":   "cartulary.network_flow.rejected_rows_query_request.v1",
		"error_codes": []string{"network_flow_invalid_ip"},
	}, phase2test.WithCookies(adminLogin.SessionCookie))
	rejected := httptestx.RequireSuccessEnvelope(t, rejectedResp, http.StatusOK)["data"].(map[string]any)
	if diagnostics := rejected["diagnostics"].([]any); len(diagnostics) != 1 {
		t.Fatalf("unexpected rejected-row diagnostics: %#v", rejected)
	}

	tablePath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/tables/" + table.TableID
	renameBody := map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename",
		"base_table_version": table.TableVersion,
		"display_name":       "Routes flows",
	}
	renameResp := phase2test.DoJSON(t, http.MethodPatch, tablePath, renameBody, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	renamed := httptestx.RequireSuccessEnvelope(t, renameResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	renamedVersion := int64(renamed["table_version"].(float64))
	if renamed["display_name"] != "Routes flows" || renamedVersion != table.TableVersion+1 {
		t.Fatalf("unexpected rename result: %#v", renamed)
	}

	renameReplayResp := phase2test.DoJSON(t, http.MethodPatch, tablePath, renameBody, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	renameReplay := httptestx.RequireSuccessEnvelope(t, renameReplayResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if renameReplay["table_version"] != renamed["table_version"] {
		t.Fatalf("unexpected rename replay payload: %#v", renameReplay)
	}

	divergentRenameResp := phase2test.DoJSON(t, http.MethodPatch, tablePath, map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename",
		"base_table_version": table.TableVersion,
		"display_name":       "Different routes flows",
	}, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, divergentRenameResp, http.StatusConflict, "client_txn_conflict")

	noOpRenameResp := phase2test.DoJSON(t, http.MethodPatch, tablePath, map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename-noop",
		"base_table_version": renamedVersion,
		"display_name":       "Routes flows",
	}, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	noOpRenamed := httptestx.RequireSuccessEnvelope(t, noOpRenameResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if int64(noOpRenamed["table_version"].(float64)) != renamedVersion {
		t.Fatalf("no-op rename changed table version: %#v", noOpRenamed)
	}
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
	deleteResp := phase2test.DoJSON(t, http.MethodDelete, tablePath, deleteBody, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	deleted := httptestx.RequireSuccessEnvelope(t, deleteResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if deleted["table_status"] != TableStatusSoftDeleted || int64(deleted["table_version"].(float64)) != renamedVersion+1 {
		t.Fatalf("unexpected delete result: %#v", deleted)
	}

	deleteReplayResp := phase2test.DoJSON(t, http.MethodDelete, tablePath, deleteBody, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	deleteReplay := httptestx.RequireSuccessEnvelope(t, deleteReplayResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if deleteReplay["table_version"] != deleted["table_version"] {
		t.Fatalf("unexpected delete replay payload: %#v", deleteReplay)
	}
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

	staleCursorResp := phase2test.DoJSON(t, http.MethodPost, queryPath, map[string]any{
		"schema_id":    schemaTableQueryContinuationForTest,
		"cursor_token": nextToken,
	}, phase2test.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, staleCursorResp, http.StatusConflict, "network_flow_table_not_active")
}

func TestNetworkFlowGraphContributorsAndIndicatorLinkRoutes(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServerWithDependencies(t, "network-flow-routes-graph-link", httpapi.DependencySet{
		ExtensionProfiles: claimedNetworkFlowProfilesForRouteTest(),
	})
	adminLogin, adminIDText := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	adminID := uuid.MustParse(adminIDText)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-network-flow-routes-graph-incident",
		"incident_key":  "IR-NF-GRAPH",
		"title":         "Network Flow graph",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	store := NewStore(harness.Server.Runtime.Postgres)
	sessionID, unitID := seedImportSessionUnit(t, harness.Server.Runtime.Postgres, incidentID, adminID, "graph-flows.csv")
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

	graphPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graphs/query"
	graphResp := phase2test.DoJSON(t, http.MethodPost, graphPath, map[string]any{
		"schema_id": "cartulary.network_flow.graph_query_request.v1",
		"table_scope": map[string]any{
			"mode":            "active_table",
			"active_table_id": table.TableID,
		},
	}, phase2test.WithCookies(adminLogin.SessionCookie))
	graph := httptestx.RequireSuccessEnvelope(t, graphResp, http.StatusOK)["data"].(map[string]any)
	if graph["schema_id"] != "cartulary.network_flow.graph_query_result.v1" {
		t.Fatalf("unexpected graph schema: %#v", graph)
	}
	graphDigest := graph["graph_query_digest"].(string)
	semanticQuery := graph["semantic_query"].(map[string]any)
	edgeAnnotations := graph["edge_annotations"].([]any)
	if len(edgeAnnotations) != 1 {
		t.Fatalf("expected one aggregate edge annotation, got %#v", edgeAnnotations)
	}
	edge := edgeAnnotations[0].(map[string]any)
	edgeID := edge["edge_id"].(string)
	examples := edge["example_row_refs"].([]any)
	if len(examples) != 2 || edge["example_refs_truncated"] != false || edge["example_refs_total_count"] != float64(2) {
		t.Fatalf("unexpected edge examples: %#v", edge)
	}
	projection := graph["graph_projection_result"].(map[string]any)
	if projection["state"] != "ephemeral_available" || projection["schema_id"] != "graph_projection.ephemeral_projection_result.v1" {
		t.Fatalf("unexpected graph projection result: %#v", projection)
	}

	contributorPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graphs/contributors/query"
	contributorResp := phase2test.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v1",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector": map[string]any{
			"kind":    "edge",
			"edge_id": edgeID,
		},
	}, phase2test.WithCookies(adminLogin.SessionCookie))
	contributorResult := httptestx.RequireSuccessEnvelope(t, contributorResp, http.StatusOK)["data"].(map[string]any)
	contributors := contributorResult["contributors"].([]any)
	if len(contributors) != 2 {
		t.Fatalf("expected two graph contributors, got %#v", contributorResult)
	}
	firstContributor := contributors[0].(map[string]any)["row"].(map[string]any)
	if firstContributor["source_row_number"] != float64(1) {
		t.Fatalf("contributors not ordered by source row: %#v", contributors)
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
	linkResp := phase2test.DoJSON(t, http.MethodPost, linkPath, linkBody, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
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

	linkReplayResp := phase2test.DoJSON(t, http.MethodPost, linkPath, linkBody, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	linkReplay := httptestx.RequireSuccessEnvelope(t, linkReplayResp, http.StatusCreated)["data"].(map[string]any)
	if linkReplay["binding"].(map[string]any)["network_flow_indicator_binding_id"] != bindingID {
		t.Fatalf("indicator-link replay changed binding: %#v", linkReplay)
	}

	duplicateBody := map[string]any{}
	for key, value := range linkBody {
		duplicateBody[key] = value
	}
	duplicateBody["client_txn_id"] = "txn-network-flow-link-duplicate"
	duplicateResp := phase2test.DoJSON(t, http.MethodPost, linkPath, duplicateBody, phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
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
}

const (
	schemaTableQueryRequestForTest      = "cartulary.network_flow.table_query_request.v1"
	schemaTableQueryContinuationForTest = "cartulary.network_flow.table_query_continuation.v1"
)

func claimedNetworkFlowProfilesForRouteTest() []httpapi.ExtensionProfile {
	profiles := httpapi.CurrentExtensionProfiles()
	for index := range profiles {
		if profiles[index].ProfileID == ProfileID {
			profiles[index].Claimed = true
		}
	}
	return profiles
}

func networkFlowRouteCountRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}
