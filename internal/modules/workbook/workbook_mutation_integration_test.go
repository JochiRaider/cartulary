package workbook_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase4_PartiesSurface_I_4_PARTIES_01(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-parties-surface")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-parties-surface-incident",
		"incident_key":  "IR-PHASE4-PARTIES",
		"title":         "Phase 4 parties surface",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	listResp := phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/view-schemas", nil, phase4test.WithCookies(adminLogin.SessionCookie))
	listData := httptestx.RequireSuccessEnvelope(t, listResp, http.StatusOK)["data"].(map[string]any)
	if !viewSchemaListContains(listData["view_schemas"].([]any), "cartulary.view.parties.v1") {
		t.Fatalf("Parties schema missing from base-profile discovery: %#v", listData["view_schemas"])
	}
	singleResp := phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/view-schemas/cartulary.view.parties.v1", nil, phase4test.WithCookies(adminLogin.SessionCookie))
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

	partyData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":       "txn-phase4-party-create",
		"party.display_name":  "  Acme Legal  ",
		"party.party_kind":    "organization",
		"party.primary_email": " legal@example.test ",
	})
	partyRow := partyData["row"].(map[string]any)
	partyID := phase4test.MustUUID(t, partyRow["record_id"].(string))
	requireCellValue(t, partyRow, "party.display_name", "Acme Legal")
	requireCellValue(t, partyRow, "party.party_kind", "organization")

	queryResp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.parties.v1/query",
		map[string]any{"filters": []map[string]any{prefixFilter("party.display_name", "acme")}},
		phase4test.WithCookies(adminLogin.SessionCookie),
	)
	queryData := httptestx.RequireSuccessEnvelope(t, queryResp, http.StatusOK)["data"].(map[string]any)
	rows := queryData["rows"].([]any)
	if len(rows) != 1 || rows[0].(map[string]any)["record_id"] != partyID.String() {
		t.Fatalf("expected incident-scoped Parties query to return created row, got %#v", rows)
	}
}

func TestPhase4_CoordinationDefaults_I_4_COORD_01(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-coordination-defaults")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-coordination-defaults-incident",
		"incident_key":  "IR-PHASE4-COORD-DEFAULTS",
		"title":         "Phase 4 coordination defaults",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

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
	commID := phase4test.MustUUID(t, commRow["record_id"].(string))
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
	requireCellValue(t, lessonRow, "lesson.owner_user_id", adminUserID.String())
	requireCellValue(t, lessonRow, "lesson.closure_state", "open")
	requireCollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 0)
	requireCollectionItemCount(t, lessonRow, "lesson.evidence_refs", 0)
}

func TestPhase4_CoordinationCollections_I_4_COORD_02(t *testing.T) {
	harness := phase4test.StartServer(t, "workbook-coordination-mutations")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-coordination-incident",
		"incident_key":  "IR-WORKBOOK-MUTATE",
		"title":         "Workbook coordination mutations",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

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
	partyID := phase4test.MustUUID(t, partyRow["record_id"].(string))
	requireCellValue(t, partyRow, "party.display_name", "Acme Legal")
	requireCellValue(t, partyRow, "party.party_kind", "organization")

	decisionID := seedDecisionRecord(t, harness, incidentID, adminUserID, "Approve containment")
	taskID := seedTaskRecord(t, harness, incidentID, adminUserID, "Collect endpoint logs")
	evidenceID := seedEvidenceRecord(t, harness, incidentID, adminUserID, "Packet capture")

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
	commID := phase4test.MustUUID(t, commRow["record_id"].(string))
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

	staleCollection := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-comm-stale-collection",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "action_payload": collectionActions(addRecordRef(decisionID))},
		},
	})
	staleBody := httptestx.RequireErrorEnvelope(t, staleCollection, http.StatusConflict, "same_field_conflict")
	conflict := staleBody["error"].(map[string]any)["conflict"].(map[string]any)
	currentValue := conflict["current_field_value"].(map[string]any)
	if currentValue["kind"] != "collection_value_v1" {
		t.Fatalf("expected typed collection conflict value, got %#v", currentValue)
	}
	rawArrayPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-workbook-comm-raw-array",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "value": []any{}},
		},
	})
	httptestx.RequireErrorEnvelope(t, rawArrayPatch, http.StatusBadRequest, "invalid_mutation_payload")
	rawNullPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-workbook-comm-raw-null",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "value": nil},
		},
	})
	httptestx.RequireErrorEnvelope(t, rawNullPatch, http.StatusBadRequest, "invalid_mutation_payload")
	wrongTarget := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 2,
		"client_txn_id":    "txn-workbook-comm-wrong-target",
		"changes": []map[string]any{
			{"field_key": "comm_log.decision_ids", "action_payload": collectionActions(addRecordRef(partyID))},
		},
	})
	httptestx.RequireErrorEnvelope(t, wrongTarget, http.StatusBadRequest, "invalid_mutation_payload")

	otherIncident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-coordination-other-incident",
		"incident_key":  "IR-WORKBOOK-MUTATE-OTHER",
		"title":         "Workbook coordination other incident",
	})
	otherIncidentID := phase4test.MustUUID(t, otherIncident["incident_id"].(string))
	foreignPartyData := requireWorkbookCreate(t, harness, adminLogin, otherIncidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-foreign-party-create",
		"party.display_name": "Foreign Party",
		"party.party_kind":   "person",
	})
	foreignPartyID := phase4test.MustUUID(t, foreignPartyData["row"].(map[string]any)["record_id"].(string))
	foreignPartyRef := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", commID, map[string]any{
		"view_schema_id":   "cartulary.view.comm_log.v1",
		"base_row_version": 2,
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
	requireCollectionItemCount(t, statusRow, "status_review.pending_evidence_ids", 1)

	lessonData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.lesson.v1", map[string]any{
		"client_txn_id":             "txn-workbook-lesson-create",
		"lesson.summary":            "Preserve VPN logs earlier",
		"lesson.follow_up_task_ids": collectionActions(addRecordRef(taskID)),
		"lesson.evidence_refs":      collectionActions(addRecordRef(evidenceID)),
	})
	lessonRow := lessonData["row"].(map[string]any)
	lessonID := phase4test.MustUUID(t, lessonRow["record_id"].(string))
	requireCellValue(t, lessonRow, "lesson.owner_user_id", adminUserID.String())
	requireCellValue(t, lessonRow, "lesson.closure_state", "open")
	requireCollectionItemCount(t, lessonRow, "lesson.follow_up_task_ids", 1)
	requireCollectionItemCount(t, lessonRow, "lesson.evidence_refs", 1)

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
	harness := phase4test.StartServer(t, "workbook-notes-tasks-decisions-mutations")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-workbook-required-surfaces-incident",
		"incident_key":  "IR-WORKBOOK-REQUIRED-MUTATE",
		"title":         "Workbook required surface mutations",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

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
	noteID := phase4test.MustUUID(t, noteRow["record_id"].(string))
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
	partyID := phase4test.MustUUID(t, partyRow["record_id"].(string))

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
		"client_txn_id":          "txn-workbook-decision-create",
		"decision.summary":       "Contain endpoint",
		"decision.decision_type": "containment",
		"decision.rationale":     "Containment is required to preserve evidence.",
		"decision.support_refs":  collectionActions(addRecordRef(supportID)),
	})
	decisionRow := decisionData["row"].(map[string]any)
	decisionID := phase4test.MustUUID(t, decisionRow["record_id"].(string))
	requireCellValue(t, decisionRow, "decision.status", "proposed")
	requireCellValue(t, decisionRow, "decision.owner_user_id", adminUserID.String())
	requireCollectionItemCount(t, decisionRow, "decision.support_refs", 1)

	supersededDecision := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.decisions.v1", uuid.Nil, map[string]any{
		"client_txn_id":          "txn-workbook-decision-superseded",
		"decision.summary":       "Invalid superseded create",
		"decision.decision_type": "scope",
		"decision.rationale":     "Superseded cannot be directly created.",
		"decision.status":        "superseded",
	})
	httptestx.RequireErrorEnvelope(t, supersededDecision, http.StatusConflict, "illegal_transition")

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
	taskID := phase4test.MustUUID(t, taskRow["record_id"].(string))
	requireCellValue(t, taskRow, "task.status", "open")
	requireCellValue(t, taskRow, "task.owner_user_id", adminUserID.String())
	requireCellValue(t, taskRow, "task.priority", "normal")
	requireCellValue(t, taskRow, "task.decision_record_id", decisionID.String())
	requireCellValue(t, taskRow, "task.linked_record_count", float64(1))
	requireCollectionItemCount(t, taskRow, "task.linked_record_ids", 1)

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

func requireWorkbookCreate(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := doWorkbookJSON(t, harness, login, http.MethodPost, incidentID, viewSchemaID, uuid.Nil, body)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func requireWorkbookPatch(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, recordID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, body)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func doWorkbookJSON(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, method string, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	url := harness.Server.HTTP.URL + "/api/v1/records/" + recordID.String()
	if method == http.MethodPost {
		url = harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + viewSchemaID + "/rows"
	}
	return phase4test.DoJSON(
		t,
		method,
		url,
		body,
		phase4test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
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

func countIncidentRecords(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM records WHERE incident_id = $1`, incidentID).Scan(&count); err != nil {
		t.Fatalf("count incident records: %v", err)
	}
	return count
}

func countViewRows(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string) int {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		phase4test.WithCookies(login.SessionCookie),
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

func seedDecisionRecord(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, summary string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	seedRecordEnvelope(t, harness, incidentID, actorID, recordID, "decision")
	execSeed(t, harness, `
INSERT INTO decisions (record_id, incident_id, summary, status, decision_type, decided_at)
VALUES ($1, $2, $3, 'approved', 'containment', '2026-04-24T12:00:00Z')
`, recordID, incidentID, summary)
	return recordID
}

func seedTaskRecord(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, title string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	seedRecordEnvelope(t, harness, incidentID, actorID, recordID, "task_request")
	execSeed(t, harness, `
INSERT INTO task_requests (record_id, incident_id, title, status, priority, updated_at)
VALUES ($1, $2, $3, 'open', 'high', '2026-04-24T11:00:00Z')
`, recordID, incidentID, title)
	return recordID
}

func seedEvidenceRecord(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, title string) uuid.UUID {
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
	return map[string]any{"op": "add_token", "raw_text": text}
}
