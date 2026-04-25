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

func TestWorkbook_PartiesAndCoordinationMutations(t *testing.T) {
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
