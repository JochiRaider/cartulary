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

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
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
		"schema_id":     "cartulary.network_flow.graph_view_create_request.v1",
		"client_txn_id": "txn-network-flow-graph-create",
		"display_name":  "Shared flow graph",
		"semantic_query": map[string]any{
			"schema_id":          "cartulary.network_flow.graph_semantic_query.v1",
			"selected_table_ids": []string{table.TableID}, "filters": []any{},
			"time_range":    map[string]any{"start_utc": nil, "end_utc": nil},
			"aggregation":   map[string]any{"mode": "default_flow_edge_v1", "include_example_row_refs": true},
			"result_limits": map[string]any{"max_vertices": 100, "max_edges": 100, "max_example_row_refs_per_edge": 10, "max_aggregate_counter_digits": 39},
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
	projection := result["projection_result"].(map[string]any)
	if projection["projection_result_id"] == "" || projection["graph_view_id"] != graphViewID || projection["source_owner_id"] != ProfileID {
		t.Fatalf("saved graph result binding drifted: %#v", projection)
	}
	vertices := projection["vertices"].([]any)
	if len(vertices) == 0 {
		t.Fatalf("saved graph result omitted vertices: %#v", projection)
	}
	beforeVertexIDs, beforeEdgeIDs := projectionObjectIDs(t, projection)
	reportingJobID := seedRestoredReportingGraphJob(t, harness, incidentID, adminID, projection)
	networkFlowJobID := seedRestoredNetworkFlowGraphJob(t, harness, incidentID, adminID, projection)
	restoreParticipant, err := recoveryassembly.NewGraphProjectionRestoreParticipant(harness.Pool)
	if err != nil {
		t.Fatalf("construct active saved-graph restore participant: %v", err)
	}
	recoveryCatalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatalf("construct active saved-graph recovery catalog: %v", err)
	}
	registryRef := restorecontract.CurrentGraphProjectionSourceRegistryRef()
	bindingRef := restorecontract.CurrentGraphProjectionImplementationBinding()
	restoreResult, err := restoreParticipant.Rebuild(context.Background(), restorecontract.GraphProjectionRebuildRequest{
		Context:             context.Background(),
		RestoreOperationID:  uuid.MustParse("00000000-0000-0000-0000-000000009101"),
		RestoredSourceState: restorecontract.RestoredGraphProjectionSourceState{},
		BackupSetID:         uuid.MustParse("00000000-0000-0000-0000-000000009102"),
		ConsistencyPointAt:  time.Date(2026, 7, 10, 12, 45, 0, 0, time.UTC),
		TargetGenerationID:  uuid.MustParse("00000000-0000-0000-0000-000000009103"),
		RecoveryStateCatalog: restorecontract.GraphProjectionRecoveryCatalogRef{
			DigestSHA256: recoveryCatalog.DigestSHA256(), AlgorithmID: restorecontract.GraphProjectionRestoreAlgorithmID,
			GraphTableIDs: restorecontract.GraphProjectionTableIDs(),
		},
		SourceRegistry: registryRef, ImplementationBinding: bindingRef,
	})
	if err != nil || !restoreResult.ReadinessSatisfied() || len(restoreResult.RebuiltViews) != 1 ||
		restoreResult.ReconciledNonterminalJobCount != 2 || restoreResult.ReconciledLeaseCount != 1 ||
		restoreResult.RebuiltViews[0].ProjectionResultID != projection["projection_result_id"] {
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
	restoredProjection := httptestx.RequireSuccessEnvelope(t, restoredResultResp, http.StatusOK)["data"].(map[string]any)["projection_result"].(map[string]any)
	afterVertexIDs, afterEdgeIDs := projectionObjectIDs(t, restoredProjection)
	if restoredProjection["projection_result_id"] != projection["projection_result_id"] ||
		strings.Join(afterVertexIDs, ",") != strings.Join(beforeVertexIDs, ",") || strings.Join(afterEdgeIDs, ",") != strings.Join(beforeEdgeIDs, ",") {
		t.Fatalf("restored saved graph object identity drifted: before=%v/%v after=%v/%v", beforeVertexIDs, beforeEdgeIDs, afterVertexIDs, afterEdgeIDs)
	}
	sourceEndpointID := vertices[0].(map[string]any)["source_entity_ref"].(map[string]any)["source_entity_id"].(string)
	contributorsResp := httptestx.DoJSON(t, http.MethodPost, resourcePath+"/contributors/query", map[string]any{
		"schema_id":            "cartulary.network_flow.graph_view_contributor_query_request.v1",
		"projection_result_id": projection["projection_result_id"],
		"selector":             map[string]any{"kind": "vertex", "vertex_id": sourceEndpointID}, "limit": 10,
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
	refreshedResult := httptestx.RequireSuccessEnvelope(t, refreshedResultResp, http.StatusOK)["data"].(map[string]any)["projection_result"].(map[string]any)
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
