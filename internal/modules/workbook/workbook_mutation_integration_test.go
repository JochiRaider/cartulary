package workbook_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestPartiesSurface_Integration(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase4-parties-surface")
	adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-parties-surface-incident",
		"incident_key":  "IR-PHASE4-PARTIES",
		"title":         "Record relationships parties surface",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

	listResp := workbookscenariotest.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/view-schemas", nil, workbookscenariotest.WithCookies(adminLogin.SessionCookie))
	listData := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	if !viewSchemaListContains(listData["view_schemas"].([]any), "cartulary.view.parties.v1") {
		t.Fatalf("Parties schema missing from base-profile discovery: %#v", listData["view_schemas"])
	}
	singleResp := workbookscenariotest.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/view-schemas/cartulary.view.parties.v1", nil, workbookscenariotest.WithCookies(adminLogin.SessionCookie))
	singleData := httptestx.RequireSuccessEnvelope(t, singleResp, http.StatusOK)["data"].(map[string]any)
	if singleData["view_schema_id"] != "cartulary.view.parties.v1" || singleData["surface_kind"] != "system_view" {
		t.Fatalf("unexpected Parties singleton schema: %#v", singleData)
	}

	beforeRecords := countIncidentRecords(t, harness, incidentID)
	beforeProjection := countViewRows(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1")
	invalidParty := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.parties.v1", uuid.Nil, map[string]any{
		"client_txn_id":    "txn-phase4-party-invalid",
		"party.party_kind": "organization",
	})
	httptestx.RequireErrorEnvelope(t, invalidParty, http.StatusBadRequest, "invalid_mutation_payload")
	if got := countIncidentRecords(t, harness, incidentID); got != beforeRecords {
		t.Fatalf("rejected party create wrote records: got %d want %d", got, beforeRecords)
	}
	if got := countViewRows(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1"); got != beforeProjection {
		t.Fatalf("rejected party create changed query rows: got %d want %d", got, beforeProjection)
	}

	routeCtx := workbookscenariotest.RouteInventoryContext{
		IncidentID:  incidentID.String(),
		ActorUserID: adminUserID.String(),
	}
	partyCreateRoute := workbookscenariotest.MustRoute(t, workbookscenariotest.RoutePartiesCreate, routeCtx)
	partyConformanceData := workbookscenariotest.RequireRouteReplayHistoryConformance(t, harness.DB, harness.Server.HTTP.URL, workbookscenariotest.RouteConformanceCase{
		Route:                  partyCreateRoute,
		Context:                routeCtx,
		ClientTxnID:            "txn-phase4-party-create-conformance",
		Login:                  adminLogin,
		ActorUserID:            adminUserID.String(),
		ExpectedMutationSource: "workbook.rows.create",
	})
	if got := partyConformanceData["row"].(map[string]any)["record_id"]; got == "" {
		t.Fatalf("party conformance create did not return a row record id: %#v", partyConformanceData)
	}

	partyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":       "txn-phase4-party-create",
		"party.display_name":  "  Acme Legal  ",
		"party.party_kind":    "organization",
		"party.primary_email": " legal@example.test ",
	})
	partyRow := partyData["row"].(map[string]any)
	partyID := workbookscenariotest.MustUUID(t, partyRow["record_id"].(string))
	requireCellValue(t, partyRow, "party.display_name", "Acme Legal")
	requireCellValue(t, partyRow, "party.party_kind", "organization")

	queryResp := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.parties.v1/query",
		map[string]any{"filters": []map[string]any{prefixFilter("party.display_name", "acme")}},
		workbookscenariotest.WithCookies(adminLogin.SessionCookie),
	)
	queryData := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)
	rows := queryData["rows"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["record_id"] != partyID.String() {
		t.Fatalf("expected incident-scoped Parties query to return created row, got %#v", rows)
	}
}

func TestWorkbook_EvidenceMutations(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "workbook-evidence-mutations")
	adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-evidence-incident",
		"incident_key":  "IR-WORKBOOK-EVIDENCE",
		"title":         "Workbook evidence mutations",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

	partyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-evidence-party-create",
		"party.display_name": "Forensics Vendor",
		"party.party_kind":   "organization",
	})
	partyID := workbookscenariotest.MustUUID(t, partyData["row"].(map[string]any)["record_id"].(string))

	beforeMinimumFailures := countIncidentRecords(t, harness, incidentID)
	blankEvidence := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.evidence.v1", uuid.Nil, map[string]any{
		"client_txn_id": "txn-workbook-evidence-blank",
	})
	httptestx.RequireErrorEnvelope(t, blankEvidence, http.StatusBadRequest, "invalid_mutation_payload")
	refOnlyEvidence := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.evidence.v1", uuid.Nil, map[string]any{
		"client_txn_id":               "txn-workbook-evidence-ref-only",
		"evidence.collector_party_id": partyID.String(),
	})
	httptestx.RequireErrorEnvelope(t, refOnlyEvidence, http.StatusBadRequest, "invalid_mutation_payload")
	if got := countIncidentRecords(t, harness, incidentID); got != beforeMinimumFailures {
		t.Fatalf("rejected evidence minimum creates wrote records: got %d want %d", got, beforeMinimumFailures)
	}

	otherIncident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-evidence-other-incident",
		"incident_key":  "IR-WORKBOOK-EVIDENCE-OTHER",
		"title":         "Workbook evidence other incident",
	})
	otherIncidentID := workbookscenariotest.MustUUID(t, otherIncident["incident_id"].(string))
	foreignPartyData := requireWorkbookCreate(t, harness, adminLogin, otherIncidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-evidence-foreign-party-create",
		"party.display_name": "Foreign Collector",
		"party.party_kind":   "person",
	})
	foreignPartyID := workbookscenariotest.MustUUID(t, foreignPartyData["row"].(map[string]any)["record_id"].(string))
	deletedPartyID := uuid.New()
	seedRecordEnvelope(t, harness, incidentID, adminUserID, deletedPartyID, "party")
	execSeed(t, harness, `
INSERT INTO parties (record_id, incident_id, display_name, party_kind)
VALUES ($1, $2, 'Deleted Party', 'person')
`, deletedPartyID, incidentID)
	execSeed(t, harness, `UPDATE records SET deleted_at = now(), deleted_by_user_id = $2 WHERE record_id = $1`, deletedPartyID, adminUserID)

	beforeReferenceFailures := countIncidentRecords(t, harness, incidentID)
	for _, invalid := range []struct {
		name string
		body map[string]any
	}{
		{
			name: "foreign-party",
			body: map[string]any{
				"client_txn_id":                 "txn-workbook-evidence-foreign-create",
				"evidence.collector_party_text": "External",
				"evidence.collector_party_id":   foreignPartyID.String(),
			},
		},
		{
			name: "deleted-party",
			body: map[string]any{
				"client_txn_id":                 "txn-workbook-evidence-deleted-create",
				"evidence.collector_party_text": "Deleted",
				"evidence.collector_party_id":   deletedPartyID.String(),
			},
		},
	} {
		resp := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.evidence.v1", uuid.Nil, invalid.body)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
	}
	if got := countIncidentRecords(t, harness, incidentID); got != beforeReferenceFailures {
		t.Fatalf("rejected evidence reference creates wrote records: got %d want %d", got, beforeReferenceFailures)
	}

	evidenceData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":                 "txn-workbook-evidence-create",
		"evidence.title":                "  Endpoint package  ",
		"evidence.storage_ref":          " s3://case/pkg ",
		"evidence.collector_party_text": "Forensics Vendor",
		"evidence.collector_party_id":   partyID.String(),
	})
	evidenceRow := evidenceData["row"].(map[string]any)
	evidenceID := workbookscenariotest.MustUUID(t, evidenceRow["record_id"].(string))
	requireCellValue(t, evidenceRow, "evidence.title", "Endpoint package")
	requireCellValue(t, evidenceRow, "evidence.storage_ref", "s3://case/pkg")
	requireCellValue(t, evidenceRow, "evidence.lifecycle_state", "requested")
	requireCellValue(t, evidenceRow, "evidence.collector_party_id", partyID.String())
	requireNonEmptyCellValue(t, evidenceRow, "evidence.requested_at")

	clearedCollector := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-evidence-clear-collector",
		"changes": []map[string]any{
			{"field_key": "evidence.collector_party_id", "value": nil},
		},
	})
	clearedCollectorRow := clearedCollector["row"].(map[string]any)
	requireCellValue(t, clearedCollectorRow, "evidence.collector_party_id", nil)
	requireCellValue(t, clearedCollectorRow, "evidence.collector_party_text", "Forensics Vendor")

	setSource := requireWorkbookPatch(t, harness, adminLogin, evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-workbook-evidence-set-source",
		"changes": []map[string]any{
			{"field_key": "evidence.source_party_id", "value": partyID.String()},
			{"field_key": "evidence.lifecycle_state", "value": "received"},
		},
	})
	setSourceRow := setSource["row"].(map[string]any)
	requireCellValue(t, setSourceRow, "evidence.source_party_id", partyID.String())
	requireCellValue(t, setSourceRow, "evidence.lifecycle_state", "received")

	invalidForeignPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", evidenceID, map[string]any{
		"view_schema_id":   "cartulary.view.evidence.v1",
		"base_row_version": 3,
		"client_txn_id":    "txn-workbook-evidence-foreign-patch",
		"changes": []map[string]any{
			{"field_key": "evidence.source_party_id", "value": foreignPartyID.String()},
		},
	})
	httptestx.RequireErrorEnvelope(t, invalidForeignPatch, http.StatusBadRequest, "invalid_mutation_payload")
}

func TestCoordinationDefaults_Integration(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase4-coordination-defaults")
	adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-coordination-defaults-incident",
		"incident_key":  "IR-PHASE4-COORD-DEFAULTS",
		"title":         "Record relationships coordination defaults",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

	routeCtx := workbookscenariotest.RouteInventoryContext{
		IncidentID:  incidentID.String(),
		ActorUserID: adminUserID.String(),
	}
	commCreateRoute := workbookscenariotest.MustRoute(t, workbookscenariotest.RouteCommLogCreate, routeCtx)
	commConformanceData := workbookscenariotest.RequireRouteReplayHistoryConformance(t, harness.DB, harness.Server.HTTP.URL, workbookscenariotest.RouteConformanceCase{
		Route:                  commCreateRoute,
		Context:                routeCtx,
		ClientTxnID:            "txn-phase4-comm-defaults-conformance",
		Login:                  adminLogin,
		ActorUserID:            adminUserID.String(),
		ExpectedMutationSource: "workbook.rows.create",
	})
	if got := commConformanceData["row"].(map[string]any)["record_id"]; got == "" {
		t.Fatalf("coordination conformance create did not return a row record id: %#v", commConformanceData)
	}

	beforeRecords := countIncidentRecords(t, harness, incidentID)
	for _, invalid := range []struct {
		viewSchemaID string
		body         map[string]any
	}{
		{"cartulary.view.comm_log.v1", map[string]any{"client_txn_id": "txn-phase4-comm-invalid", "comm_log.comm_type": "briefing"}},
		{"cartulary.view.handoff.v1", map[string]any{"client_txn_id": "txn-phase4-handoff-invalid", "handoff.incoming_owner_user_id": adminUserID.String()}},
		{"cartulary.view.status_review.v1", map[string]any{"client_txn_id": "txn-phase4-status-invalid"}},
		{"cartulary.view.lesson.v1", map[string]any{"client_txn_id": "txn-phase4-lesson-invalid"}},
	} {
		resp := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, invalid.viewSchemaID, uuid.Nil, invalid.body)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
	}
	if got := countIncidentRecords(t, harness, incidentID); got != beforeRecords {
		t.Fatalf("rejected coordination creates wrote records: got %d want %d", got, beforeRecords)
	}

	commData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.comm_log.v1", map[string]any{
		"client_txn_id":               "txn-phase4-comm-defaults-create",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "leadership",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Initial coordination update",
	})
	commRow := commData["row"].(map[string]any)
	commID := workbookscenariotest.MustUUID(t, commRow["record_id"].(string))
	requireNonEmptyCellValue(t, commRow, "comm_log.timestamp_utc")
	requireCollectionItemCount(t, commRow, "comm_log.decision_ids", 0)
	requireCollectionItemCount(t, commRow, "comm_log.action_task_ids", 0)
	requireCollectionItemCount(t, commRow, "comm_log.audience_party_ids", 0)
	requireCollectionItemCount(t, commRow, "comm_log.attendee_party_ids", 0)
	requireCellValue(t, commRow, "comm_log.next_report_at", nil)
	requireCellValue(t, commRow, "comm_log.privilege_tag", nil)

	immutableCommID := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-phase4-comm-immutable-id",
		"changes":          []map[string]any{{"field_key": "comm_log.comm_id", "value": "client-supplied"}},
	})
	httptestx.RequireErrorEnvelope(t, immutableCommID, http.StatusBadRequest, "invalid_mutation_payload")
	commSetNextReport := requireWorkbookPatch(t, harness, adminLogin, commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-phase4-comm-set-next-report",
		"changes":          []map[string]any{{"field_key": "comm_log.next_report_at", "value": "2026-04-24T18:00:00Z"}},
	})
	requireNonEmptyCellValue(t, commSetNextReport["row"].(map[string]any), "comm_log.next_report_at")
	commClear := requireWorkbookPatch(t, harness, adminLogin, commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-phase4-comm-clear-next-report",
		"changes":          []map[string]any{{"field_key": "comm_log.next_report_at", "value": nil}},
	})
	requireCellValue(t, commClear["row"].(map[string]any), "comm_log.next_report_at", nil)
	commNullTimestamp := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 3,
		"client_txn_id":    "txn-phase4-comm-null-timestamp",
		"changes":          []map[string]any{{"field_key": "comm_log.timestamp_utc", "value": nil}},
	})
	httptestx.RequireErrorEnvelope(t, commNullTimestamp, http.StatusBadRequest, "invalid_mutation_payload")

	handoffData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.handoff.v1", map[string]any{
		"client_txn_id":                  "txn-phase4-handoff-defaults-create",
		"handoff.incoming_owner_user_id": adminUserID.String(),
		"handoff.current_state_summary":  "Night shift owns containment",
	})
	handoffRow := handoffData["row"].(map[string]any)
	requireNonEmptyCellValue(t, handoffRow, "handoff.timestamp_utc")
	requireCellValue(t, handoffRow, "handoff.outgoing_owner_user_id", adminUserID.String())
	requireCollectionItemCount(t, handoffRow, "handoff.open_task_ids", 0)
	requireCollectionItemCount(t, handoffRow, "handoff.open_decision_ids", 0)
	requireCollectionItemCount(t, handoffRow, "handoff.open_risk_refs", 0)
	requireCellValue(t, handoffRow, "handoff.next_checks", nil)
	requireCellValue(t, handoffRow, "handoff.acknowledged_at", nil)

	statusData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.status_review.v1", map[string]any{
		"client_txn_id":                       "txn-phase4-status-defaults-create",
		"status_review.current_state_summary": "Containment is stable",
	})
	statusRow := statusData["row"].(map[string]any)
	requireNonEmptyCellValue(t, statusRow, "status_review.timestamp_utc")
	requireCellValue(t, statusRow, "status_review.review_owner_user_id", adminUserID.String())
	requireCellValue(t, statusRow, "status_review.current_state_summary", "Containment is stable")
	requireCollectionItemCount(t, statusRow, "status_review.blocked_task_ids", 0)
	requireCollectionItemCount(t, statusRow, "status_review.pending_evidence_ids", 0)
	requireCollectionItemCount(t, statusRow, "status_review.open_decision_ids", 0)
	requireCellValue(t, statusRow, "status_review.active_risks_summary", nil)
	requireCellValue(t, statusRow, "status_review.next_report_at", nil)

	lessonData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.lesson.v1", map[string]any{
		"client_txn_id":  "txn-phase4-lesson-defaults-create",
		"lesson.summary": "Preserve VPN logs earlier",
	})
	lessonRow := lessonData["row"].(map[string]any)
	requireNonEmptyCellValue(t, lessonRow, "lesson.timestamp_utc")
	requireCellValue(t, lessonRow, "lesson.summary", "Preserve VPN logs earlier")
	requireCellValue(t, lessonRow, "lesson.owner_user_id", adminUserID.String())
	requireCellValue(t, lessonRow, "lesson.closure_state", "open")
	requireCollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 0)
	requireCollectionItemCount(t, lessonRow, "lesson.evidence_refs", 0)
}

func TestGenericCollectionPatch_Integration(t *testing.T) {
	testCoordinationCollections(t)
}

func TestCoordinationCollections_Integration(t *testing.T) {
	testCoordinationCollections(t)
}

func testCoordinationCollections(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "workbook-coordination-mutations")
	adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-coordination-incident",
		"incident_key":  "IR-WORKBOOK-MUTATE",
		"title":         "Workbook coordination mutations",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

	beforePartyFailureRecords := countIncidentRecords(t, harness, incidentID)
	invalidParty := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.parties.v1", uuid.Nil, map[string]any{
		"client_txn_id":    "txn-workbook-party-invalid",
		"party.party_kind": "organization",
	})
	httptestx.RequireErrorEnvelope(t, invalidParty, http.StatusBadRequest, "invalid_mutation_payload")
	if got := countIncidentRecords(t, harness, incidentID); got != beforePartyFailureRecords {
		t.Fatalf("rejected party create wrote records: got %d want %d", got, beforePartyFailureRecords)
	}

	partyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":       "txn-workbook-party-create",
		"party.display_name":  "  Acme Legal  ",
		"party.party_kind":    "organization",
		"party.primary_email": " legal@example.test ",
	})
	partyRow := partyData["row"].(map[string]any)
	partyID := workbookscenariotest.MustUUID(t, partyRow["record_id"].(string))
	requireCellValue(t, partyRow, "party.display_name", "Acme Legal")
	requireCellValue(t, partyRow, "party.party_kind", "organization")

	secondPartyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-party-second-create",
		"party.display_name": "Security Lead",
		"party.party_kind":   "person",
	})
	secondPartyID := workbookscenariotest.MustUUID(t, secondPartyData["row"].(map[string]any)["record_id"].(string))
	thirdPartyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-party-third-create",
		"party.display_name": "Legal Observer",
		"party.party_kind":   "person",
	})
	thirdPartyID := workbookscenariotest.MustUUID(t, thirdPartyData["row"].(map[string]any)["record_id"].(string))

	decisionID := seedDecisionRecord(t, harness, incidentID, adminUserID, "Approve containment")
	autoRebaseDecisionID := seedDecisionRecord(t, harness, incidentID, adminUserID, "Approve status page")
	secondDecisionID := seedDecisionRecord(t, harness, incidentID, adminUserID, "Approve credential rotation")
	thirdDecisionID := seedDecisionRecord(t, harness, incidentID, adminUserID, "Approve network block")
	taskID := seedTaskRecord(t, harness, incidentID, adminUserID, "Collect endpoint logs")
	secondTaskID := seedTaskRecord(t, harness, incidentID, adminUserID, "Collect VPN logs")
	thirdTaskID := seedTaskRecord(t, harness, incidentID, adminUserID, "Collect firewall logs")
	evidenceID := seedEvidenceRecord(t, harness, incidentID, adminUserID, "Packet capture")
	secondEvidenceID := seedEvidenceRecord(t, harness, incidentID, adminUserID, "VPN log bundle")
	thirdEvidenceID := seedEvidenceRecord(t, harness, incidentID, adminUserID, "Firewall log bundle")

	commData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.comm_log.v1", map[string]any{
		"client_txn_id":               "txn-workbook-comm-create",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "leadership",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Initial coordination update",
		"comm_log.decision_ids":       collectionActions(addRecordRef(decisionID), addRecordRef(decisionID)),
		"comm_log.action_task_ids":    collectionActions(addRecordRef(taskID)),
		"comm_log.audience_party_ids": collectionActions(addPartyRef(partyID)),
		"comm_log.attendee_party_ids": collectionActions(addPartyRef(partyID)),
	})
	commRow := commData["row"].(map[string]any)
	commID := workbookscenariotest.MustUUID(t, commRow["record_id"].(string))
	requireCellValue(t, commRow, "comm_log.comm_type", "briefing")
	requireCollectionItemCount(t, commRow, "comm_log.decision_ids", 1)
	requireCollectionItemCount(t, commRow, "comm_log.action_task_ids", 1)
	requireCollectionItemCount(t, commRow, "comm_log.audience_party_ids", 1)
	requireCollectionItemCount(t, commRow, "comm_log.attendee_party_ids", 1)

	commPatchData := requireWorkbookPatch(t, harness, adminLogin, commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-comm-patch",
		"changes": []map[string]any{
			{"field_key": "comm_log.summary", "value": "Updated coordination summary"},
			{"field_key": "comm_log.privilege_tag", "value": " attorney-client "},
		},
	})
	commPatchedRow := commPatchData["row"].(map[string]any)
	requireCellValue(t, commPatchedRow, "comm_log.summary", "Updated coordination summary")
	requireCellValue(t, commPatchedRow, "comm_log.privilege_tag", "attorney-client")

	staleNonOverlappingCollection := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-comm-stale-non-overlap",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "action_payload": collectionActions(addRecordRef(autoRebaseDecisionID))},
		},
	})
	staleNonOverlappingData := httptestx.RequireSuccessEnvelope(t, staleNonOverlappingCollection, http.StatusOK)["data"].(map[string]any)
	staleNonOverlappingRow := staleNonOverlappingData["row"].(map[string]any)
	commVersion := int64(staleNonOverlappingRow["row_version"].(float64))
	if commVersion != 3 {
		t.Fatalf("stale non-overlapping collection patch should auto-rebase to row_version 3, got %#v", staleNonOverlappingData)
	}
	requireCollectionValueHasRecordRef(t, cellMapValue(t, staleNonOverlappingRow, "comm_log.decision_ids"), autoRebaseDecisionID)

	beforeDuplicateAdd := snapshotWorkbookConflictSideEffects(t, harness, incidentID, commID)
	duplicateAdd := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": commVersion,
		"client_txn_id":    "txn-workbook-comm-duplicate-add-no-op",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "action_payload": collectionActions(addRecordRef(autoRebaseDecisionID))},
		},
	})
	duplicateAddBody := httptestx.RequireErrorEnvelope(t, duplicateAdd, http.StatusBadRequest, "invalid_mutation_payload")
	duplicateAddDetails := duplicateAddBody["error"].(map[string]any)["details"].(map[string]any)
	if duplicateAddDetails["reason_code"] != "no_effective_change" {
		t.Fatalf("duplicate collection add should fail as no_effective_change, got %#v", duplicateAddBody)
	}
	if afterDuplicateAdd := snapshotWorkbookConflictSideEffects(t, harness, incidentID, commID); beforeDuplicateAdd != afterDuplicateAdd {
		t.Fatalf("duplicate collection add wrote durable side effects: before=%+v after=%+v", beforeDuplicateAdd, afterDuplicateAdd)
	}
	rawArrayPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": commVersion,
		"client_txn_id":    "txn-workbook-comm-raw-array",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "value": []any{}},
		},
	})
	httptestx.RequireErrorEnvelope(t, rawArrayPatch, http.StatusBadRequest, "invalid_mutation_payload")
	rawNullPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": commVersion,
		"client_txn_id":    "txn-workbook-comm-raw-null",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "value": nil},
		},
	})
	httptestx.RequireErrorEnvelope(t, rawNullPatch, http.StatusBadRequest, "invalid_mutation_payload")
	wrongTarget := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": commVersion,
		"client_txn_id":    "txn-workbook-comm-wrong-target",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "action_payload": collectionActions(addRecordRef(partyID))},
		},
	})
	httptestx.RequireErrorEnvelope(t, wrongTarget, http.StatusBadRequest, "invalid_mutation_payload")

	otherIncident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-coordination-other-incident",
		"incident_key":  "IR-WORKBOOK-MUTATE-OTHER",
		"title":         "Workbook coordination other incident",
	})
	otherIncidentID := workbookscenariotest.MustUUID(t, otherIncident["incident_id"].(string))
	foreignPartyData := requireWorkbookCreate(t, harness, adminLogin, otherIncidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-foreign-party-create",
		"party.display_name": "Foreign Party",
		"party.party_kind":   "person",
	})
	foreignPartyID := workbookscenariotest.MustUUID(t, foreignPartyData["row"].(map[string]any)["record_id"].(string))
	foreignPartyRef := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": commVersion,
		"client_txn_id":    "txn-workbook-comm-foreign-party",
		"changes": []map[string]any{
			{"field_key": "comm_log.audience_party_ids", "action_payload": collectionActions(addPartyRef(foreignPartyID))},
		},
	})
	httptestx.RequireErrorEnvelope(t, foreignPartyRef, http.StatusBadRequest, "invalid_mutation_payload")

	handoffData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.handoff.v1", map[string]any{
		"client_txn_id":                  "txn-workbook-handoff-create",
		"handoff.incoming_owner_user_id": adminUserID.String(),
		"handoff.current_state_summary":  "Night shift owns containment",
		"handoff.open_task_ids":          collectionActions(addRecordRef(taskID)),
		"handoff.open_decision_ids":      collectionActions(addRecordRef(decisionID)),
		"handoff.open_risk_refs":         collectionActions(addRiskRef("VPN logs may expire"), addRiskRef(" VPN logs may expire ")),
	})
	handoffRow := handoffData["row"].(map[string]any)
	requireCellValue(t, handoffRow, "handoff.outgoing_owner_user_id", adminUserID.String())
	requireCollectionItemCount(t, handoffRow, "handoff.open_task_ids", 1)
	riskItems := requireCollectionItemCount(t, handoffRow, "handoff.open_risk_refs", 1)
	firstRiskRef := riskItems[0].(map[string]any)["risk_ref_id"].(string)

	secondHandoffData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.handoff.v1", map[string]any{
		"client_txn_id":                  "txn-workbook-handoff-create-second",
		"handoff.incoming_owner_user_id": adminUserID.String(),
		"handoff.current_state_summary":  "Day shift owns containment",
		"handoff.open_risk_refs":         collectionActions(addRiskRef("VPN logs may expire")),
	})
	secondRiskItems := requireCollectionItemCount(t, secondHandoffData["row"].(map[string]any), "handoff.open_risk_refs", 1)
	if got := secondRiskItems[0].(map[string]any)["risk_ref_id"].(string); got == firstRiskRef {
		t.Fatalf("risk refs must be scoped per handoff record, reused %s", got)
	}

	statusData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.status_review.v1", map[string]any{
		"client_txn_id":                       "txn-workbook-status-review-create",
		"status_review.current_state_summary": "Containment is stable",
		"status_review.blocked_task_ids":      collectionActions(addRecordRef(taskID)),
		"status_review.pending_evidence_ids":  collectionActions(addRecordRef(evidenceID)),
		"status_review.open_decision_ids":     collectionActions(addRecordRef(decisionID)),
	})
	statusRow := statusData["row"].(map[string]any)
	requireCellValue(t, statusRow, "status_review.review_owner_user_id", adminUserID.String())
	requireCellValue(t, statusRow, "status_review.current_state_summary", "Containment is stable")
	requireCollectionItemCount(t, statusRow, "status_review.pending_evidence_ids", 1)

	lessonData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.lesson.v1", map[string]any{
		"client_txn_id":             "txn-workbook-lesson-create",
		"lesson.summary":            "Preserve VPN logs earlier",
		"lesson.follow_up_task_ids": collectionActions(addRecordRef(taskID)),
		"lesson.evidence_refs":      collectionActions(addRecordRef(evidenceID)),
	})
	lessonRow := lessonData["row"].(map[string]any)
	lessonID := workbookscenariotest.MustUUID(t, lessonRow["record_id"].(string))
	requireCellValue(t, lessonRow, "lesson.summary", "Preserve VPN logs earlier")
	requireCellValue(t, lessonRow, "lesson.owner_user_id", adminUserID.String())
	requireCellValue(t, lessonRow, "lesson.closure_state", "open")
	requireCollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 1)
	requireCollectionItemCount(t, lessonRow, "lesson.evidence_refs", 1)

	commVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, commID, "cartulary.view.comm_log.v1", "comm_log.decision_ids", commVersion, collectionActions(addRecordRef(secondDecisionID)), collectionActions(addRecordRef(thirdDecisionID)), adminUserID, "record_ref", "comm-decision")
	commVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, commID, "cartulary.view.comm_log.v1", "comm_log.action_task_ids", commVersion, collectionActions(addRecordRef(secondTaskID)), collectionActions(addRecordRef(thirdTaskID)), adminUserID, "record_ref", "comm-task")
	commVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, commID, "cartulary.view.comm_log.v1", "comm_log.audience_party_ids", commVersion, collectionActions(addPartyRef(secondPartyID)), collectionActions(addPartyRef(thirdPartyID)), adminUserID, "party_ref", "comm-audience")
	_ = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, commID, "cartulary.view.comm_log.v1", "comm_log.attendee_party_ids", commVersion, collectionActions(addPartyRef(secondPartyID)), collectionActions(addPartyRef(thirdPartyID)), adminUserID, "party_ref", "comm-attendee")

	handoffID := workbookscenariotest.MustUUID(t, handoffRow["record_id"].(string))
	handoffVersion := int64(1)
	handoffVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, handoffID, "cartulary.view.handoff.v1", "handoff.open_task_ids", handoffVersion, collectionActions(addRecordRef(secondTaskID)), collectionActions(addRecordRef(thirdTaskID)), adminUserID, "record_ref", "handoff-task")
	handoffVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, handoffID, "cartulary.view.handoff.v1", "handoff.open_decision_ids", handoffVersion, collectionActions(addRecordRef(secondDecisionID)), collectionActions(addRecordRef(thirdDecisionID)), adminUserID, "record_ref", "handoff-decision")
	_ = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, handoffID, "cartulary.view.handoff.v1", "handoff.open_risk_refs", handoffVersion, collectionActions(addRiskRef("VPN logs may expire soon")), collectionActions(addRiskRef("Rotate VPN logs sooner")), adminUserID, "risk_ref", "handoff-risk")

	statusID := workbookscenariotest.MustUUID(t, statusRow["record_id"].(string))
	statusVersion := int64(1)
	statusVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, statusID, "cartulary.view.status_review.v1", "status_review.blocked_task_ids", statusVersion, collectionActions(addRecordRef(secondTaskID)), collectionActions(addRecordRef(thirdTaskID)), adminUserID, "record_ref", "status-task")
	statusVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, statusID, "cartulary.view.status_review.v1", "status_review.pending_evidence_ids", statusVersion, collectionActions(addRecordRef(secondEvidenceID)), collectionActions(addRecordRef(thirdEvidenceID)), adminUserID, "record_ref", "status-evidence")
	_ = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, statusID, "cartulary.view.status_review.v1", "status_review.open_decision_ids", statusVersion, collectionActions(addRecordRef(secondDecisionID)), collectionActions(addRecordRef(thirdDecisionID)), adminUserID, "record_ref", "status-decision")

	lessonVersion := int64(1)
	lessonVersion = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, lessonID, "cartulary.view.lesson.v1", "lesson.follow_up_task_ids", lessonVersion, collectionActions(addRecordRef(secondTaskID)), collectionActions(addRecordRef(thirdTaskID)), adminUserID, "record_ref", "lesson-task")
	_ = requireCollectionSameFieldConflict(t, harness, adminLogin, incidentID, lessonID, "cartulary.view.lesson.v1", "lesson.evidence_refs", lessonVersion, collectionActions(addRecordRef(secondEvidenceID)), collectionActions(addRecordRef(thirdEvidenceID)), adminUserID, "record_ref", "lesson-evidence")

	immutableLessonID := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", lessonID, map[string]any{
		"view_schema_id":   "cartulary.view.lesson.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-lesson-immutable-id",
		"changes": []map[string]any{
			{"field_key": "lesson.lesson_id", "value": "client-supplied"},
		},
	})
	httptestx.RequireErrorEnvelope(t, immutableLessonID, http.StatusBadRequest, "invalid_mutation_payload")
}

func TestWorkbook_NotesTasksAndDecisionsMutations(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "workbook-notes-tasks-decisions-mutations")
	adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-required-surfaces-incident",
		"incident_key":  "IR-WORKBOOK-REQUIRED-MUTATE",
		"title":         "Workbook required surface mutations",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

	beforeNoteFailures := countIncidentRecords(t, harness, incidentID)
	blankNote := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.notes.v1", uuid.Nil, map[string]any{
		"client_txn_id": "txn-workbook-note-blank",
	})
	httptestx.RequireErrorEnvelope(t, blankNote, http.StatusBadRequest, "invalid_mutation_payload")
	tagOnlyNote := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.notes.v1", uuid.Nil, map[string]any{
		"client_txn_id": "txn-workbook-note-tag-only",
		"note.tags":     collectionActions(addToken("triage")),
	})
	httptestx.RequireErrorEnvelope(t, tagOnlyNote, http.StatusBadRequest, "invalid_mutation_payload")
	if got := countIncidentRecords(t, harness, incidentID); got != beforeNoteFailures {
		t.Fatalf("rejected note creates wrote records: got %d want %d", got, beforeNoteFailures)
	}

	noteData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1", map[string]any{
		"client_txn_id": "txn-workbook-note-create",
		"note.title":    "  Analyst note  ",
		"note.body":     "  First line\n\nSecond line  ",
		"note.tags":     collectionActions(addToken(" Investigation "), addToken("Investigation")),
	})
	noteRow := noteData["row"].(map[string]any)
	noteID := workbookscenariotest.MustUUID(t, noteRow["record_id"].(string))
	requireCellValue(t, noteRow, "note.title", "Analyst note")
	requireCollectionItemCount(t, noteRow, "note.tags", 1)
	noteReplay := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.notes.v1", uuid.Nil, map[string]any{
		"client_txn_id": "txn-workbook-note-create",
		"note.title":    "  Analyst note  ",
		"note.body":     "  First line\n\nSecond line  ",
		"note.tags":     collectionActions(addToken(" Investigation "), addToken("Investigation")),
	})
	httptestx.RequireSuccessEnvelope(t, noteReplay, http.StatusOK)
	notePatch := requireWorkbookPatch(t, harness, adminLogin, noteID, map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-note-patch",
		"changes": []map[string]any{
			{"field_key": "note.body", "value": "Updated note body"},
			{"field_key": "note.tags", "action_payload": collectionActions(addToken("follow-up"))},
		},
	})
	notePatchedRow := notePatch["row"].(map[string]any)
	requireCellValue(t, notePatchedRow, "note.body", "Updated note body")
	requireCollectionItemCount(t, notePatchedRow, "note.tags", 2)

	supportID := seedEvidenceRecord(t, harness, incidentID, adminUserID, "Support packet capture")
	partyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-task-requester-party",
		"party.display_name": "Incident Commander",
		"party.party_kind":   "person",
	})
	partyRow := partyData["row"].(map[string]any)
	partyID := workbookscenariotest.MustUUID(t, partyRow["record_id"].(string))

	confidenceSupportRef := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.decisions.v1", uuid.Nil, map[string]any{
		"client_txn_id":          "txn-workbook-decision-confidence-rejected",
		"decision.summary":       "Invalid support confidence",
		"decision.decision_type": "evidence",
		"decision.rationale":     "Client confidence is not authoritative.",
		"decision.support_refs": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 100},
			},
		},
	})
	httptestx.RequireErrorEnvelope(t, confidenceSupportRef, http.StatusBadRequest, "invalid_mutation_payload")
	confidenceAffectedRef := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.decisions.v1", uuid.Nil, map[string]any{
		"client_txn_id":          "txn-workbook-decision-affected-confidence-rejected",
		"decision.summary":       "Invalid affected confidence",
		"decision.decision_type": "evidence",
		"decision.rationale":     "Client confidence is not authoritative.",
		"decision.affected_record_ids": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 100},
			},
		},
	})
	httptestx.RequireErrorEnvelope(t, confidenceAffectedRef, http.StatusBadRequest, "invalid_mutation_payload")

	beforeTaskFailure := countIncidentRecords(t, harness, incidentID)
	invalidTask := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.task_requests.v1", uuid.Nil, map[string]any{
		"client_txn_id":  "txn-workbook-task-invalid",
		"task.title":     "Blocked task without reason",
		"task.task_kind": "collection",
		"task.status":    "blocked",
	})
	httptestx.RequireErrorEnvelope(t, invalidTask, http.StatusConflict, "illegal_transition")
	if got := countIncidentRecords(t, harness, incidentID); got != beforeTaskFailure {
		t.Fatalf("rejected task create wrote records: got %d want %d", got, beforeTaskFailure)
	}

	decisionData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.decisions.v1", map[string]any{
		"client_txn_id":                "txn-workbook-decision-create",
		"decision.summary":             "Contain endpoint",
		"decision.decision_type":       "containment",
		"decision.rationale":           "Containment is required to preserve evidence.",
		"decision.support_refs":        collectionActions(addRecordRef(supportID)),
		"decision.affected_record_ids": collectionActions(addRecordRef(supportID), addRecordRef(supportID)),
	})
	decisionRow := decisionData["row"].(map[string]any)
	decisionID := workbookscenariotest.MustUUID(t, decisionRow["record_id"].(string))
	requireCellValue(t, decisionRow, "decision.status", "proposed")
	requireCellValue(t, decisionRow, "decision.owner_user_id", adminUserID.String())
	requireCollectionItemCount(t, decisionRow, "decision.support_refs", 1)
	requireCollectionItemCount(t, decisionRow, "decision.affected_record_ids", 1)
	requireCellValue(t, decisionRow, "decision.affected_record_count", float64(1))

	supersededDecision := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.decisions.v1", uuid.Nil, map[string]any{
		"client_txn_id":          "txn-workbook-decision-superseded",
		"decision.summary":       "Invalid superseded create",
		"decision.decision_type": "scope",
		"decision.rationale":     "Superseded cannot be directly created.",
		"decision.status":        "superseded",
	})
	httptestx.RequireErrorEnvelope(t, supersededDecision, http.StatusConflict, "illegal_transition")

	confidenceTaskRef := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.task_requests.v1", uuid.Nil, map[string]any{
		"client_txn_id":  "txn-workbook-task-confidence-rejected",
		"task.title":     "Invalid linked confidence",
		"task.task_kind": "collection",
		"task.linked_record_ids": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 100},
			},
		},
	})
	httptestx.RequireErrorEnvelope(t, confidenceTaskRef, http.StatusBadRequest, "invalid_mutation_payload")

	taskData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.task_requests.v1", map[string]any{
		"client_txn_id":             "txn-workbook-task-create",
		"task.title":                "Collect endpoint logs",
		"task.task_kind":            "collection",
		"task.requester_party_id":   partyID.String(),
		"task.decision_record_id":   decisionID.String(),
		"task.linked_record_ids":    collectionActions(addRecordRef(supportID), addRecordRef(supportID)),
		"task.requester_party_text": "Incident Commander",
	})
	taskRow := taskData["row"].(map[string]any)
	taskID := workbookscenariotest.MustUUID(t, taskRow["record_id"].(string))
	requireCellValue(t, taskRow, "task.status", "open")
	requireCellValue(t, taskRow, "task.owner_user_id", adminUserID.String())
	requireCellValue(t, taskRow, "task.priority", "normal")
	requireCellValue(t, taskRow, "task.decision_record_id", decisionID.String())
	requireCellValue(t, taskRow, "task.linked_record_count", float64(1))
	requireCollectionItemCount(t, taskRow, "task.linked_record_ids", 1)

	taskConfidencePatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", taskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-task-confidence-patch-rejected",
		"changes": []map[string]any{
			{"field_key": "task.linked_record_ids", "action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_record_ref", "linked_record_id": noteID.String(), "confidence": 100},
				},
			}},
		},
	})
	httptestx.RequireErrorEnvelope(t, taskConfidencePatch, http.StatusBadRequest, "invalid_mutation_payload")
	if got := countActiveRecordLinks(t, harness, taskID, "task.linked_record_ids"); got != 1 {
		t.Fatalf("task confidence patch changed active links: got %d want 1", got)
	}

	decisionSupportConfidencePatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", decisionID, map[string]any{
		"view_schema_id":   "cartulary.view.decisions.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-decision-support-confidence-patch-rejected",
		"changes": []map[string]any{
			{"field_key": "decision.support_refs", "action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_record_ref", "linked_record_id": noteID.String(), "confidence": 100},
				},
			}},
		},
	})
	httptestx.RequireErrorEnvelope(t, decisionSupportConfidencePatch, http.StatusBadRequest, "invalid_mutation_payload")
	if got := countActiveRecordLinks(t, harness, decisionID, "decision.support_refs"); got != 1 {
		t.Fatalf("decision support confidence patch changed active links: got %d want 1", got)
	}

	decisionAffectedConfidencePatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", decisionID, map[string]any{
		"view_schema_id":   "cartulary.view.decisions.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-decision-affected-confidence-patch-rejected",
		"changes": []map[string]any{
			{"field_key": "decision.affected_record_ids", "action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_record_ref", "linked_record_id": noteID.String(), "confidence": 100},
				},
			}},
		},
	})
	httptestx.RequireErrorEnvelope(t, decisionAffectedConfidencePatch, http.StatusBadRequest, "invalid_mutation_payload")
	if got := countActiveRecordLinks(t, harness, decisionID, "decision.affected_record_ids"); got != 1 {
		t.Fatalf("decision affected confidence patch changed active links: got %d want 1", got)
	}

	taskDone := requireWorkbookPatch(t, harness, adminLogin, taskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-task-done",
		"changes": []map[string]any{
			{"field_key": "task.status", "value": "done"},
		},
	})
	taskDoneRow := taskDone["row"].(map[string]any)
	requireCellValue(t, taskDoneRow, "task.status", "done")
	requireNonEmptyCellValue(t, taskDoneRow, "task.completed_at")

	taskDoneToCanceled := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", taskID, map[string]any{
		"view_schema_id":   "cartulary.view.task_requests.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-workbook-task-done-canceled",
		"changes": []map[string]any{
			{"field_key": "task.status", "value": "canceled"},
		},
	})
	httptestx.RequireErrorEnvelope(t, taskDoneToCanceled, http.StatusConflict, "illegal_transition")

	decisionApproved := requireWorkbookPatch(t, harness, adminLogin, decisionID, map[string]any{
		"view_schema_id":   "cartulary.view.decisions.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-decision-approved",
		"changes": []map[string]any{
			{"field_key": "decision.status", "value": "approved"},
		},
	})
	decisionApprovedRow := decisionApproved["row"].(map[string]any)
	requireCellValue(t, decisionApprovedRow, "decision.status", "approved")

	decisionApprovedToRejected := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", decisionID, map[string]any{
		"view_schema_id":   "cartulary.view.decisions.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-workbook-decision-approved-rejected",
		"changes": []map[string]any{
			{"field_key": "decision.status", "value": "rejected"},
		},
	})
	httptestx.RequireErrorEnvelope(t, decisionApprovedToRejected, http.StatusConflict, "illegal_transition")
}

func requireWorkbookCreate(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, viewSchemaID, uuid.Nil, body)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func requireWorkbookPatch(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, recordID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, body)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func doWorkbookJSON(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, method string, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	url := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String()
	if method == http.MethodPost {
		url = harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + viewSchemaID + "/rows"
	}
	return workbookscenariotest.DoJSON(
		t,
		method,
		url,
		body,
		workbookscenariotest.WithCookies(login.SessionCookie, login.CSRFCookie),
		workbookscenariotest.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func requireCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	got := cells[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("unexpected %s value: got %#v want %#v", fieldKey, got, want)
	}
}

func requireNonEmptyCellValue(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	got := cells[fieldKey].(map[string]any)["value"]
	if got == nil || got == "" {
		t.Fatalf("expected non-empty %s value, got %#v", fieldKey, got)
	}
}

func requireCollectionItemCount(t testing.TB, row map[string]any, fieldKey string, want int) []any {
	t.Helper()
	cells := row["cells"].(map[string]any)
	value := cells[fieldKey].(map[string]any)["value"].(map[string]any)
	if value["kind"] != "collection_value_v1" {
		t.Fatalf("expected %s to be collection_value_v1, got %#v", fieldKey, value)
	}
	items := value["items"].([]any)
	if len(items) != want {
		t.Fatalf("unexpected %s item count: got %d want %d items=%#v", fieldKey, len(items), want, items)
	}
	return items
}

type workbookConflictSideEffects struct {
	ChangeSets        int
	MutationRows      int
	RecordRevisions   int
	RouteIdempotency  int
	ActiveRecordLinks int
	ActiveRiskRefs    int
}

func requireCollectionSameFieldConflict(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, viewSchemaID string, fieldKey string, baseVersion int64, serverAction map[string]any, clientAction map[string]any, actorID uuid.UUID, expectedItemKind string, txnPrefix string) int64 {
	t.Helper()
	serverData := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": baseVersion,
		"client_txn_id":    "txn-workbook-conflict-server-" + txnPrefix,
		"changes": []map[string]any{
			{"field_key": fieldKey, "action_payload": serverAction},
		},
	})
	serverRow := serverData["row"].(map[string]any)
	serverVersion := baseVersion + 1
	requireCollectionValueHasItemKind(t, cellMapValue(t, serverRow, fieldKey), expectedItemKind)

	before := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	stale := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": baseVersion,
		"client_txn_id":    "txn-workbook-conflict-client-" + txnPrefix,
		"changes": []map[string]any{
			{"field_key": fieldKey, "action_payload": clientAction},
		},
	})
	body := httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if _, ok := conflict["current_field_value"]; ok {
		t.Fatalf("same-field conflict preserved legacy current_field_value alias: %#v", conflict)
	}
	if _, ok := conflict["conflict_resolution"]; ok {
		t.Fatalf("same-field conflict preserved legacy conflict_resolution alias: %#v", conflict)
	}
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != fieldKey ||
		conflict["conflict_resolution_class"] != "collection_review" ||
		conflict["server_updated_by"] != actorID.String() {
		t.Fatalf("unexpected conflict envelope identity fields for %s: %#v", fieldKey, conflict)
	}
	if got := int64(conflict["base_row_version"].(float64)); got != baseVersion {
		t.Fatalf("unexpected base_row_version for %s: got %d want %d conflict=%#v", fieldKey, got, baseVersion, conflict)
	}
	if got := int64(conflict["current_row_version"].(float64)); got != serverVersion {
		t.Fatalf("unexpected current_row_version for %s: got %d want %d conflict=%#v", fieldKey, got, serverVersion, conflict)
	}
	if _, err := time.Parse(time.RFC3339Nano, conflict["server_updated_at"].(string)); err != nil {
		t.Fatalf("server_updated_at for %s was not RFC3339Nano: %v conflict=%#v", fieldKey, err, conflict)
	}
	for _, key := range []string{"client_value", "server_value", "base_value"} {
		value := conflict[key].(map[string]any)
		if value["kind"] != "collection_value_v1" {
			t.Fatalf("expected %s.%s to be collection_value_v1, got %#v", fieldKey, key, value)
		}
		requireCollectionValueHasItemKind(t, value, expectedItemKind)
	}
	after := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if before != after {
		t.Fatalf("same-field conflict for %s wrote durable side effects: before=%+v after=%+v", fieldKey, before, after)
	}
	return serverVersion
}

func cellMapValue(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()
	cells := row["cells"].(map[string]any)
	return cells[fieldKey].(map[string]any)["value"].(map[string]any)
}

func requireCollectionValueHasItemKind(t testing.TB, value map[string]any, itemKind string) {
	t.Helper()
	items := value["items"].([]any)
	for _, item := range items {
		if item.(map[string]any)["item_kind"] == itemKind {
			return
		}
	}
	t.Fatalf("expected collection to contain item_kind %s, got %#v", itemKind, value)
}

func requireCollectionValueHasRecordRef(t testing.TB, value map[string]any, recordID uuid.UUID) {
	t.Helper()
	items := value["items"].([]any)
	for _, item := range items {
		itemMap := item.(map[string]any)
		if itemMap["item_kind"] == "record_ref" && itemMap["linked_record_id"] == recordID.String() {
			return
		}
	}
	t.Fatalf("expected collection to contain record_ref %s, got %#v", recordID, value)
}

func snapshotWorkbookConflictSideEffects(t testing.TB, harness *workbookscenariotest.ServerHarness, incidentID uuid.UUID, recordID uuid.UUID) workbookConflictSideEffects {
	t.Helper()
	var snapshot workbookConflictSideEffects
	row := harness.DB.QueryRowContext(context.Background(), `
SELECT
    (SELECT count(*) FROM change_sets WHERE incident_id = $1),
    (SELECT count(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1),
    (SELECT count(*) FROM record_revisions WHERE record_id = $2),
    (SELECT count(*) FROM route_idempotency WHERE scope_key = $2::text),
    (SELECT count(*) FROM record_links WHERE incident_id = $1 AND src_record_id = $2 AND deleted_at IS NULL),
    (SELECT count(*) FROM handoff_risk_refs WHERE incident_id = $1 AND handoff_record_id = $2 AND deleted_at IS NULL)
`, incidentID, recordID)
	if err := row.Scan(&snapshot.ChangeSets, &snapshot.MutationRows, &snapshot.RecordRevisions, &snapshot.RouteIdempotency, &snapshot.ActiveRecordLinks, &snapshot.ActiveRiskRefs); err != nil {
		t.Fatalf("snapshot workbook conflict side effects: %v", err)
	}
	return snapshot
}

func countIncidentRecords(t testing.TB, harness *workbookscenariotest.ServerHarness, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM records WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count incident records: %v", err)
	}
	return count
}

func countActiveRecordLinks(t testing.TB, harness *workbookscenariotest.ServerHarness, sourceID uuid.UUID, fieldKey string) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM record_links
 WHERE src_record_id = $1
   AND field_key = $2
   AND deleted_at IS NULL
`, sourceID, fieldKey).Scan(&count); err != nil {
		t.Fatalf("count active record links: %v", err)
	}
	return count
}

func countViewRows(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, incidentID uuid.UUID, viewSchemaID string) int {
	t.Helper()
	resp := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		workbookscenariotest.WithCookies(login.SessionCookie),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	return len(body["data"].(map[string]any)["rows"].([]any))
}

func viewSchemaListContains(items []any, viewSchemaID string) bool {
	for _, item := range items {
		if item.(map[string]any)["view_schema_id"] == viewSchemaID {
			return true
		}
	}
	return false
}

func seedDecisionRecord(t testing.TB, harness *workbookscenariotest.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, summary string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	seedRecordEnvelope(t, harness, incidentID, actorID, recordID, "decision")
	execSeed(t, harness, `
INSERT INTO decisions (record_id, incident_id, summary, status, decision_type, decided_at)
VALUES ($1, $2, $3, 'approved', 'containment', '2026-04-24T12:00:00Z')
`, recordID, incidentID, summary)
	seedDecisionProjection(t, harness, recordID)
	return recordID
}

func seedTaskRecord(t testing.TB, harness *workbookscenariotest.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, title string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	seedRecordEnvelope(t, harness, incidentID, actorID, recordID, "task_request")
	execSeed(t, harness, `
INSERT INTO task_requests (record_id, incident_id, title, status, priority, updated_at)
VALUES ($1, $2, $3, 'open', 'high', '2026-04-24T11:00:00Z')
`, recordID, incidentID, title)
	seedTaskRequestProjection(t, harness, recordID)
	return recordID
}

func seedEvidenceRecord(t testing.TB, harness *workbookscenariotest.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, title string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	seedRecordEnvelope(t, harness, incidentID, actorID, recordID, "evidence")
	execSeed(t, harness, `
INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, requested_at)
VALUES ($1, $2, $3, 'received', 'complete', '2026-04-24T10:00:00Z')
`, recordID, incidentID, title)
	return recordID
}

func collectionActions(actions ...map[string]any) map[string]any {
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func addRecordRef(recordID uuid.UUID) map[string]any {
	return map[string]any{"op": "add_record_ref", "linked_record_id": recordID.String()}
}

func addPartyRef(partyID uuid.UUID) map[string]any {
	return map[string]any{"op": "add_party_ref", "party_id": partyID.String()}
}

func addRiskRef(text string) map[string]any {
	return map[string]any{"op": "add_risk_ref", "risk_ref_text": text}
}

func addToken(text string) map[string]any {
	return map[string]any{"op": "add_tag", "tag_name": text}
}
