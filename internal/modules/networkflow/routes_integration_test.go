package networkflow_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
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
	if sourceProfiles["schema_id"] != "cartulary.network_flow.source_profile_list.v1" {
		t.Fatalf("unexpected source profiles payload: %#v", sourceProfiles)
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

	invalidationMessages, unsubscribeInvalidations := harness.Collaboration.SubscribeIncident(incidentID, 4)
	defer unsubscribeInvalidations()

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
	requireNetworkFlowResourceChange(t, invalidationMessages, incidentID, table.TableID, platformws.ExtensionResourceChangeKindInvalidate, platformws.ExtensionResourceReasonRenamed)

	renameReplayResp := httptestx.DoJSON(t, http.MethodPatch, tablePath, renameBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	renameReplay := httptestx.RequireSuccessEnvelope(t, renameReplayResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if renameReplay["table_version"] != renamed["table_version"] {
		t.Fatalf("unexpected rename replay payload: %#v", renameReplay)
	}
	requireNoNetworkFlowResourceChange(t, invalidationMessages)

	divergentRenameResp := httptestx.DoJSON(t, http.MethodPatch, tablePath, map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename",
		"base_table_version": table.TableVersion,
		"display_name":       "Different routes flows",
	}, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	httptestx.RequireErrorEnvelope(t, divergentRenameResp, http.StatusConflict, "client_txn_conflict")

	noOpRenameResp := httptestx.DoJSON(t, http.MethodPatch, tablePath, map[string]any{
		"client_txn_id":      "txn-network-flow-route-rename-noop",
		"base_table_version": renamedVersion,
		"display_name":       "Routes flows",
	}, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	noOpRenamed := httptestx.RequireSuccessEnvelope(t, noOpRenameResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if int64(noOpRenamed["table_version"].(float64)) != renamedVersion {
		t.Fatalf("no-op rename changed table version: %#v", noOpRenamed)
	}
	requireNoNetworkFlowResourceChange(t, invalidationMessages)
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
	requireNetworkFlowResourceChange(t, invalidationMessages, incidentID, table.TableID, platformws.ExtensionResourceChangeKindRemove, platformws.ExtensionResourceReasonSoftDeleted)

	deleteReplayResp := httptestx.DoJSON(t, http.MethodDelete, tablePath, deleteBody, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	deleteReplay := httptestx.RequireSuccessEnvelope(t, deleteReplayResp, http.StatusOK)["data"].(map[string]any)["table"].(map[string]any)
	if deleteReplay["table_version"] != deleted["table_version"] {
		t.Fatalf("unexpected delete replay payload: %#v", deleteReplay)
	}
	requireNoNetworkFlowResourceChange(t, invalidationMessages)
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

func requireNetworkFlowResourceChange(t testing.TB, messages <-chan platformws.Message, incidentID uuid.UUID, tableID string, changeKind string, reasonCode string) {
	t.Helper()
	select {
	case message := <-messages:
		if message.Type != "extension_resource_changed" {
			t.Fatalf("message type = %q, want extension_resource_changed", message.Type)
		}
		if message.IncidentID != incidentID.String() {
			t.Fatalf("message incident_id = %q, want %q", message.IncidentID, incidentID)
		}
		if message.StreamSeq == nil || *message.StreamSeq <= 0 {
			t.Fatalf("message stream_seq missing or invalid: %#v", message)
		}
		var payload map[string]any
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			t.Fatalf("decode extension_resource_changed payload: %v", err)
		}
		if payload["extension_profile_id"] != ProfileID || payload["resource_kind"] != "network_flow_table" || payload["resource_id"] != tableID {
			t.Fatalf("unexpected extension resource identity: %#v", payload)
		}
		if payload["change_kind"] != changeKind || payload["reason_code"] != reasonCode {
			t.Fatalf("unexpected extension resource change semantics: %#v", payload)
		}
		if _, ok := payload["display_name"]; ok {
			t.Fatalf("invalidation payload must not disclose labels: %#v", payload)
		}
		if _, ok := payload["source_filename_display"]; ok {
			t.Fatalf("invalidation payload must not disclose import metadata: %#v", payload)
		}
		workspaceRefs := payload["workspace_refs"].([]any)
		if len(workspaceRefs) != 1 {
			t.Fatalf("unexpected workspace_refs: %#v", workspaceRefs)
		}
		ref := workspaceRefs[0].(map[string]any)
		if ref["kind"] != "extension_workspace" || ref["extension_profile_id"] != ProfileID || ref["workspace_key"] != WorkspaceKeyNetworkAnalysis {
			t.Fatalf("unexpected workspace ref: %#v", ref)
		}
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for Network Flow %s/%s invalidation", changeKind, reasonCode)
	}
}

func requireNoNetworkFlowResourceChange(t testing.TB, messages <-chan platformws.Message) {
	t.Helper()
	select {
	case message := <-messages:
		t.Fatalf("unexpected Network Flow invalidation message: %#v", message)
	default:
	}
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

	graphPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graphs/query"
	graphResp := httptestx.DoJSON(t, http.MethodPost, graphPath, map[string]any{
		"schema_id": "cartulary.network_flow.graph_query_request.v1",
		"table_scope": map[string]any{
			"mode":            "active_table",
			"active_table_id": table.TableID,
		},
	}, httptestx.WithCookies(adminLogin.SessionCookie))
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
	if projection["state"] != "ephemeral_available" || projection["projection_schema_id"] != "graph_projection.v1" {
		t.Fatalf("unexpected graph projection result: %#v", projection)
	}
	for _, field := range []string{"graph_view_id", "graph_view_key", "source_snapshot_id", "projection_version", "generated_at", "properties", "metadata", "schema_registry", "vertices", "edges", "consumer_capabilities"} {
		if _, ok := projection[field]; !ok {
			t.Fatalf("graph projection result omitted %s: %#v", field, projection)
		}
	}

	contributorPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/network-flow/graphs/contributors/query"
	contributorResp := httptestx.DoJSON(t, http.MethodPost, contributorPath, map[string]any{
		"schema_id":          "cartulary.network_flow.graph_contributor_query_request.v1",
		"graph_query":        semanticQuery,
		"graph_query_digest": graphDigest,
		"selector": map[string]any{
			"kind":    "edge",
			"edge_id": edgeID,
		},
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
	return runtime.StartServer(t, appsupport.ServerOptions{
		Prefix: prefix,
		Env: map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
		},
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
