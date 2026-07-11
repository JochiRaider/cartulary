package workbook_test

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/phase4test"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

const phase6NotesViewSchemaID = "cartulary.view.notes.v1"

func TestPhase6_GridWriteConcurrencyRoute_U_6_01(t *testing.T) {
	harness, login, _, incidentID := phase6ConflictFixture(t, "phase6-u-6-01-grid-write-concurrency", "IR-PHASE6-U-6-01")
	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-01-create", "Base title", "Base body")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))

	routeLeft := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-01-route-left-create", "Route left", "Left base body")
	routeLeftID := phase4test.MustUUID(t, routeLeft["record_id"].(string))
	routeRight := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-01-route-right-create", "Route right", "Right base body")
	routeRightID := phase4test.MustUUID(t, routeRight["record_id"].(string))
	routeBoundPatchBody := map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-route-bound-body",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "Path-selected body"}},
	}
	routeLeftPatch := requireWorkbookPatch(t, harness, login, routeLeftID, routeBoundPatchBody)
	requireCellValue(t, routeLeftPatch["row"].(map[string]any), "note.body", "Path-selected body")
	routeRightAfterLeft := phase6RequireQueriedRow(t, harness, login, incidentID, phase6NotesViewSchemaID, routeRightID)
	requireCellValue(t, routeRightAfterLeft, "note.body", "Right base body")
	routeRightPatch := requireWorkbookPatch(t, harness, login, routeRightID, routeBoundPatchBody)
	requireCellValue(t, routeRightPatch["row"].(map[string]any), "note.body", "Path-selected body")
	routeLeftAfterRight := phase6RequireQueriedRow(t, harness, login, incidentID, phase6NotesViewSchemaID, routeLeftID)
	requireCellValue(t, routeLeftAfterRight, "note.body", "Path-selected body")

	missingBase := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id": phase6NotesViewSchemaID,
		"client_txn_id":  "txn-phase6-u-6-01-missing-base",
		"changes":        []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	httptestx.RequireErrorEnvelope(t, missingBase, http.StatusBadRequest, "invalid_mutation_payload")
	fullRowPatch := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"record_id":        recordID.String(),
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-record-id-in-body",
		"note.body":        "Client body",
	})
	httptestx.RequireErrorEnvelope(t, fullRowPatch, http.StatusBadRequest, "invalid_mutation_payload")

	titlePatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-server-title",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})
	requireCellValue(t, titlePatch["row"].(map[string]any), "note.title", "Server title")

	bodyPatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-client-body",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	bodyRow := bodyPatch["row"].(map[string]any)
	requireCellValue(t, bodyRow, "note.title", "Server title")
	requireCellValue(t, bodyRow, "note.body", "Client body")
	if got := int64(bodyRow["row_version"].(float64)); got != 3 {
		t.Fatalf("different-field stale edit row_version = %d want 3", got)
	}
	phase6RequireMutationChangedFields(t, harness, bodyPatch["change_set_id"].(string), []string{"note.body"})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	sameField := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-01-client-title",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Client title"}},
	})
	body := httptestx.RequireErrorEnvelope(t, sameField, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["field_key"] != "note.title" || conflict["client_value"] != "Client title" || conflict["server_value"] != "Server title" {
		t.Fatalf("unexpected same-field conflict payload: %#v", conflict)
	}
	afterConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("same-field stale edit wrote durable side effects: before=%+v after=%+v", beforeConflict, afterConflict)
	}
}

func TestPhase6_SameFieldConflictHTTP_U_6_02(t *testing.T) {
	harness, login, actorID, incidentID := phase6ConflictFixture(t, "phase6-u-6-02-same-field-http", "IR-PHASE6-U-6-02")
	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-02-create", "Base title", "Base body")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-02-server",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})

	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-02-client",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Client title"}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	errorObject := body["error"].(map[string]any)
	if errorObject["status"] != float64(http.StatusConflict) {
		t.Fatalf("same-field conflict error.status = %#v want 409", errorObject["status"])
	}
	conflict := errorObject["conflict"].(map[string]any)
	phase6RequireNoLegacyConflictAliases(t, conflict)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "note.title" ||
		conflict["conflict_resolution_class"] != "text_compare_merge" ||
		conflict["base_value"] != "Base title" ||
		conflict["server_value"] != "Server title" ||
		conflict["client_value"] != "Client title" ||
		conflict["server_updated_by"] != actorID.String() {
		t.Fatalf("unexpected same-field conflict identity/value fields: %#v", conflict)
	}
	phase6RequireConflictVersion(t, conflict, "base_row_version", 1)
	phase6RequireConflictVersion(t, conflict, "current_row_version", 2)
	if _, err := time.Parse(time.RFC3339Nano, conflict["server_updated_at"].(string)); err != nil {
		t.Fatalf("server_updated_at was not RFC3339Nano: %v conflict=%#v", err, conflict)
	}
}

func TestPhase6_TextCompareMergeDurability_U_6_03(t *testing.T) {
	harness, login, _, incidentID := phase6ConflictFixture(t, "phase6-u-6-03-text-merge-durability", "IR-PHASE6-U-6-03")
	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-03-clean-create", "Merge note", "one\ntwo\nthree")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-clean-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nTWO\nthree"}},
	})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-clean-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\ntwo\nTHREE"}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["suggested_merged_value"] != "one\nTWO\nTHREE" {
		t.Fatalf("expected clean text suggestion without accepting patch, got %#v", conflict)
	}
	afterConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if beforeConflict != afterConflict {
		t.Fatalf("clean text conflict wrote durable side effects: before=%+v after=%+v", beforeConflict, afterConflict)
	}
	current := phase6RequireQueriedRow(t, harness, login, incidentID, phase6NotesViewSchemaID, recordID)
	requireCellValue(t, current, "note.body", "one\nTWO\nthree")
	if got := int64(current["row_version"].(float64)); got != 2 {
		t.Fatalf("rejected clean text conflict changed row_version = %d want 2", got)
	}

	resolved := phase6ResolveConflict(t, harness, login, recordID, conflict["conflict_token"].(string), map[string]any{
		"conflict_token":  conflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-phase6-u-6-03-clean-resolve",
		"resolved_value":  "one\nTWO\nTHREE",
	})
	requireCellValue(t, resolved["row"].(map[string]any), "note.body", "one\nTWO\nTHREE")
	afterResolve := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if afterResolve.ChangeSets != beforeConflict.ChangeSets+1 || afterResolve.RecordRevisions != beforeConflict.RecordRevisions+1 {
		t.Fatalf("explicit text resolution should be the next durable revision: before=%+v after=%+v", beforeConflict, afterResolve)
	}

	overlap := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-03-overlap-create", "Overlap note", "one\ntwo")
	overlapID := phase4test.MustUUID(t, overlap["record_id"].(string))
	requireWorkbookPatch(t, harness, login, overlapID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-overlap-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nserver"}},
	})
	overlapResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", overlapID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-overlap-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nclient"}},
	})
	overlapBody := httptestx.RequireErrorEnvelope(t, overlapResp, http.StatusConflict, "same_field_conflict")
	overlapConflict := overlapBody["error"].(map[string]any)["conflict"].(map[string]any)
	if _, ok := overlapConflict["suggested_merged_value"]; ok {
		t.Fatalf("overlapping text conflict must omit suggested_merged_value: %#v", overlapConflict)
	}

	renderedLookingBase := "# Title\r\n<div data-x=\"base\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-1\">HOST</entity-chip>"
	renderedLookingServer := "# Title\n<div data-x=\"server\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-1\">HOST</entity-chip>"
	renderedLookingClient := "# Title\n<div data-x=\"base\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-2\">HOST</entity-chip>"
	renderedLookingSuggestion := "# Title\n<div data-x=\"server\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-2\">HOST</entity-chip>"
	rendered := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-03-rendered-create", "Rendered-looking note", renderedLookingBase)
	renderedID := phase4test.MustUUID(t, rendered["record_id"].(string))
	requireWorkbookPatch(t, harness, login, renderedID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-rendered-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": renderedLookingServer}},
	})
	beforeRenderedConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, renderedID)
	renderedResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", renderedID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-03-rendered-client",
		"changes":          []map[string]any{{"field_key": "note.body", "value": renderedLookingClient}},
	})
	renderedBody := httptestx.RequireErrorEnvelope(t, renderedResp, http.StatusConflict, "same_field_conflict")
	renderedConflict := renderedBody["error"].(map[string]any)["conflict"].(map[string]any)
	if renderedConflict["base_value"] != "# Title\n<div data-x=\"base\">safe</div>\n&amp; entity stays text\n<entity-chip record_id=\"host-1\">HOST</entity-chip>" ||
		renderedConflict["server_value"] != renderedLookingServer ||
		renderedConflict["client_value"] != renderedLookingClient ||
		renderedConflict["suggested_merged_value"] != renderedLookingSuggestion {
		t.Fatalf("rendered-looking text must remain raw plain text in conflict payload: %#v", renderedConflict)
	}
	afterRenderedConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, renderedID)
	if beforeRenderedConflict != afterRenderedConflict {
		t.Fatalf("rendered-looking clean text conflict wrote durable side effects: before=%+v after=%+v", beforeRenderedConflict, afterRenderedConflict)
	}
	renderedCurrent := phase6RequireQueriedRow(t, harness, login, incidentID, phase6NotesViewSchemaID, renderedID)
	requireCellValue(t, renderedCurrent, "note.body", renderedLookingServer)
}

func TestPhase6_CollectionReviewRouteResolve_U_6_04(t *testing.T) {
	harness, login, actorID, incidentID := phase6ConflictFixture(t, "phase6-u-6-04-collection-review-route", "IR-PHASE6-U-6-04")
	phase6RequireCollectionReviewCaseInventory(t, phase6CollectionReviewCases())

	partyData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase6-u-6-04-party-create",
		"party.display_name": "Incident Commander",
		"party.party_kind":   "person",
	})
	partyID := phase4test.MustUUID(t, partyData["row"].(map[string]any)["record_id"].(string))
	secondPartyData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase6-u-6-04-party-second-create",
		"party.display_name": "Security Lead",
		"party.party_kind":   "person",
	})
	secondPartyID := phase4test.MustUUID(t, secondPartyData["row"].(map[string]any)["record_id"].(string))
	thirdPartyData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-phase6-u-6-04-party-third-create",
		"party.display_name": "Legal Observer",
		"party.party_kind":   "person",
	})
	thirdPartyID := phase4test.MustUUID(t, thirdPartyData["row"].(map[string]any)["record_id"].(string))

	decisionID := seedDecisionRecord(t, harness, incidentID, actorID, "Approve containment")
	secondDecisionID := seedDecisionRecord(t, harness, incidentID, actorID, "Approve credential rotation")
	thirdDecisionID := seedDecisionRecord(t, harness, incidentID, actorID, "Approve network block")
	taskID := seedTaskRecord(t, harness, incidentID, actorID, "Collect endpoint logs")
	secondTaskID := seedTaskRecord(t, harness, incidentID, actorID, "Collect VPN logs")
	thirdTaskID := seedTaskRecord(t, harness, incidentID, actorID, "Collect firewall logs")
	evidenceID := seedEvidenceRecord(t, harness, incidentID, actorID, "Packet capture")
	secondEvidenceID := seedEvidenceRecord(t, harness, incidentID, actorID, "VPN log bundle")
	thirdEvidenceID := seedEvidenceRecord(t, harness, incidentID, actorID, "Firewall log bundle")

	noteData := requireWorkbookCreate(t, harness, login, incidentID, phase6NotesViewSchemaID, map[string]any{
		"client_txn_id": "txn-phase6-u-6-04-create",
		"note.title":    "Collection note",
		"note.tags":     collectionActions(addToken("base-tag")),
	})
	note := noteData["row"].(map[string]any)
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireCollectionValueHasItemKind(t, cellMapValue(t, note, "note.tags"), "tag")
	queried := phase6RequireQueriedRow(t, harness, login, incidentID, phase6NotesViewSchemaID, recordID)
	requireCollectionValueHasItemKind(t, cellMapValue(t, queried, "note.tags"), "tag")

	commData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.comm_log.v1", map[string]any{
		"client_txn_id":               "txn-phase6-u-6-04-comm-create",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "leadership",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Initial coordination update",
		"comm_log.decision_ids":       collectionActions(addRecordRef(decisionID)),
		"comm_log.action_task_ids":    collectionActions(addRecordRef(taskID)),
		"comm_log.audience_party_ids": collectionActions(addPartyRef(partyID)),
		"comm_log.attendee_party_ids": collectionActions(addPartyRef(partyID)),
	})
	commID := phase4test.MustUUID(t, commData["row"].(map[string]any)["record_id"].(string))
	handoffData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.handoff.v1", map[string]any{
		"client_txn_id":                  "txn-phase6-u-6-04-handoff-create",
		"handoff.incoming_owner_user_id": actorID.String(),
		"handoff.current_state_summary":  "Night shift owns containment",
		"handoff.open_task_ids":          collectionActions(addRecordRef(taskID)),
		"handoff.open_decision_ids":      collectionActions(addRecordRef(decisionID)),
		"handoff.open_risk_refs":         collectionActions(addRiskRef("Base risk")),
	})
	handoffID := phase4test.MustUUID(t, handoffData["row"].(map[string]any)["record_id"].(string))
	statusData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.status_review.v1", map[string]any{
		"client_txn_id":                       "txn-phase6-u-6-04-status-create",
		"status_review.current_state_summary": "Containment is stable",
		"status_review.blocked_task_ids":      collectionActions(addRecordRef(taskID)),
		"status_review.pending_evidence_ids":  collectionActions(addRecordRef(evidenceID)),
		"status_review.open_decision_ids":     collectionActions(addRecordRef(decisionID)),
	})
	statusID := phase4test.MustUUID(t, statusData["row"].(map[string]any)["record_id"].(string))
	lessonData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.lesson.v1", map[string]any{
		"client_txn_id":             "txn-phase6-u-6-04-lesson-create",
		"lesson.summary":            "Preserve VPN logs earlier",
		"lesson.follow_up_task_ids": collectionActions(addRecordRef(taskID)),
		"lesson.evidence_refs":      collectionActions(addRecordRef(evidenceID)),
	})
	lessonID := phase4test.MustUUID(t, lessonData["row"].(map[string]any)["record_id"].(string))
	decisionData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.decisions.v1", map[string]any{
		"client_txn_id":                "txn-phase6-u-6-04-decision-create",
		"decision.summary":             "Contain endpoint",
		"decision.decision_type":       "containment",
		"decision.rationale":           "Containment is required.",
		"decision.support_refs":        collectionActions(addRecordRef(evidenceID)),
		"decision.affected_record_ids": collectionActions(addRecordRef(evidenceID)),
	})
	decisionRowID := phase4test.MustUUID(t, decisionData["row"].(map[string]any)["record_id"].(string))
	taskData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.task_requests.v1", map[string]any{
		"client_txn_id":          "txn-phase6-u-6-04-task-create",
		"task.title":             "Collect endpoint logs",
		"task.task_kind":         "collection",
		"task.linked_record_ids": collectionActions(addRecordRef(evidenceID)),
	})
	taskRowID := phase4test.MustUUID(t, taskData["row"].(map[string]any)["record_id"].(string))

	versions := map[uuid.UUID]int64{
		recordID:      1,
		commID:        1,
		handoffID:     1,
		statusID:      1,
		lessonID:      1,
		decisionRowID: 1,
		taskRowID:     1,
	}
	recordIDs := map[string]uuid.UUID{
		"note.tags":                          recordID,
		"comm_log.decision_ids":              commID,
		"comm_log.action_task_ids":           commID,
		"comm_log.audience_party_ids":        commID,
		"comm_log.attendee_party_ids":        commID,
		"handoff.open_task_ids":              handoffID,
		"handoff.open_decision_ids":          handoffID,
		"handoff.open_risk_refs":             handoffID,
		"status_review.blocked_task_ids":     statusID,
		"status_review.pending_evidence_ids": statusID,
		"status_review.open_decision_ids":    statusID,
		"lesson.follow_up_task_ids":          lessonID,
		"lesson.evidence_refs":               lessonID,
		"decision.support_refs":              decisionRowID,
		"decision.affected_record_ids":       decisionRowID,
		"task.linked_record_ids":             taskRowID,
	}
	cases := phase6CollectionReviewCases()
	phase6BindCollectionActions(cases, phase6CollectionActionIDs{
		SecondDecisionID: secondDecisionID,
		ThirdDecisionID:  thirdDecisionID,
		SecondTaskID:     secondTaskID,
		ThirdTaskID:      thirdTaskID,
		SecondEvidenceID: secondEvidenceID,
		ThirdEvidenceID:  thirdEvidenceID,
		SecondPartyID:    secondPartyID,
		ThirdPartyID:     thirdPartyID,
	})
	keys := make([]string, 0, len(cases))
	for fieldKey := range cases {
		keys = append(keys, fieldKey)
	}
	sort.Strings(keys)
	for _, fieldKey := range keys {
		tc := cases[fieldKey]
		rowID, ok := recordIDs[fieldKey]
		if !ok {
			t.Fatalf("missing collection-review record fixture for %s", fieldKey)
		}
		versions[rowID] = phase6ExerciseCollectionReviewField(t, harness, login, incidentID, actorID, rowID, versions[rowID], tc)
	}

	for _, invalid := range []struct {
		name    string
		payload any
	}{
		{name: "raw-array", payload: []any{}},
		{name: "raw-string", payload: "client-tag"},
		{name: "replace-all", payload: map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "replace_all", "raw_text": "client-tag"}}}},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			invalidResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
				"view_schema_id":   phase6NotesViewSchemaID,
				"base_row_version": 3,
				"client_txn_id":    "txn-phase6-u-6-04-invalid-" + invalid.name,
				"changes": []map[string]any{{
					"field_key":      "note.tags",
					"action_payload": invalid.payload,
				}},
			})
			httptestx.RequireErrorEnvelope(t, invalidResp, http.StatusBadRequest, "invalid_mutation_payload")
		})
	}
}

func TestPhase6_ConflictResolveDurability_U_6_06(t *testing.T) {
	harness, login, actorID, incidentID := phase6ConflictFixture(t, "phase6-u-6-06-conflict-resolve-durability", "IR-PHASE6-U-6-06")

	keepRecordID, keepConflict := phase6CreateNoteTitleConflict(t, harness, login, incidentID, "keep", "Keep saved", "Keep local")
	beforeKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID)
	keepBody := map[string]any{
		"conflict_token":  keepConflict["conflict_token"].(string),
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-phase6-u-6-06-keep-resolve",
	}
	keepData := phase6ResolveConflict(t, harness, login, keepRecordID, keepConflict["conflict_token"].(string), keepBody)
	phase6RequireNoChangeSetID(t, keepData)
	requireCellValue(t, keepData["row"].(map[string]any), "note.title", "Keep saved")
	afterKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID)
	phase6RequireSideEffectDelta(t, "keep_saved", beforeKeep, afterKeep, workbookConflictSideEffects{
		RouteIdempotency: 1,
	})
	phase6ResolveConflict(t, harness, login, keepRecordID, keepConflict["conflict_token"].(string), keepBody)
	phase6RequireSideEffectDelta(t, "keep_saved replay", beforeKeep, snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID), workbookConflictSideEffects{
		RouteIdempotency: 1,
	})

	useRecordID, useConflict := phase6CreateNoteTitleConflict(t, harness, login, incidentID, "use", "Use saved", "Use local")
	beforeUse := snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID)
	useBody := map[string]any{
		"conflict_token":  useConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-phase6-u-6-06-use-resolve",
		"resolved_value":  "Use local",
	}
	useData := phase6ResolveConflict(t, harness, login, useRecordID, useConflict["conflict_token"].(string), useBody)
	requireCellValue(t, useData["row"].(map[string]any), "note.title", "Use local")
	phase4test.RequireChangeSetAttribution(t, harness.DB, useData["change_set_id"].(string), actorID.String(), "workbook.records.conflicts.resolve", "txn-phase6-u-6-06-use-resolve")
	afterUse := snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID)
	phase6RequireSideEffectDelta(t, "use_unsaved", beforeUse, afterUse, workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
	phase6ResolveConflict(t, harness, login, useRecordID, useConflict["conflict_token"].(string), useBody)
	phase6RequireSideEffectDelta(t, "use_unsaved replay", beforeUse, snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID), workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})

	mergedRecordID, mergedConflict := phase6CreateNoteTitleConflict(t, harness, login, incidentID, "merged", "Merged saved", "Merged local")
	beforeMerged := snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID)
	mergedBody := map[string]any{
		"conflict_token":  mergedConflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-phase6-u-6-06-merged-resolve",
		"resolved_value":  "Merged final",
	}
	mergedData := phase6ResolveConflict(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), mergedBody)
	requireCellValue(t, mergedData["row"].(map[string]any), "note.title", "Merged final")
	phase4test.RequireChangeSetAttribution(t, harness.DB, mergedData["change_set_id"].(string), actorID.String(), "workbook.records.conflicts.resolve", "txn-phase6-u-6-06-merged-resolve")
	afterMerged := snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID)
	phase6RequireSideEffectDelta(t, "merged_value", beforeMerged, afterMerged, workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
	phase6ResolveConflict(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), mergedBody)
	phase6RequireSideEffectDelta(t, "merged_value replay", beforeMerged, snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID), workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
}

type phase6CollectionActionIDs struct {
	SecondDecisionID uuid.UUID
	ThirdDecisionID  uuid.UUID
	SecondTaskID     uuid.UUID
	ThirdTaskID      uuid.UUID
	SecondEvidenceID uuid.UUID
	ThirdEvidenceID  uuid.UUID
	SecondPartyID    uuid.UUID
	ThirdPartyID     uuid.UUID
}

type phase6CollectionReviewCase struct {
	viewSchemaID     string
	fieldKey         string
	expectedItemKind string
	serverPayload    map[string]any
	clientPayload    map[string]any
	resolvePayload   map[string]any
}

func phase6CollectionReviewCases() map[string]phase6CollectionReviewCase {
	return map[string]phase6CollectionReviewCase{
		"note.tags":                          {viewSchemaID: phase6NotesViewSchemaID, fieldKey: "note.tags", expectedItemKind: "tag"},
		"comm_log.decision_ids":              {viewSchemaID: "cartulary.view.comm_log.v1", fieldKey: "comm_log.decision_ids", expectedItemKind: "record_ref"},
		"comm_log.action_task_ids":           {viewSchemaID: "cartulary.view.comm_log.v1", fieldKey: "comm_log.action_task_ids", expectedItemKind: "record_ref"},
		"comm_log.audience_party_ids":        {viewSchemaID: "cartulary.view.comm_log.v1", fieldKey: "comm_log.audience_party_ids", expectedItemKind: "party_ref"},
		"comm_log.attendee_party_ids":        {viewSchemaID: "cartulary.view.comm_log.v1", fieldKey: "comm_log.attendee_party_ids", expectedItemKind: "party_ref"},
		"decision.support_refs":              {viewSchemaID: "cartulary.view.decisions.v1", fieldKey: "decision.support_refs", expectedItemKind: "record_ref"},
		"decision.affected_record_ids":       {viewSchemaID: "cartulary.view.decisions.v1", fieldKey: "decision.affected_record_ids", expectedItemKind: "record_ref"},
		"handoff.open_task_ids":              {viewSchemaID: "cartulary.view.handoff.v1", fieldKey: "handoff.open_task_ids", expectedItemKind: "record_ref"},
		"handoff.open_decision_ids":          {viewSchemaID: "cartulary.view.handoff.v1", fieldKey: "handoff.open_decision_ids", expectedItemKind: "record_ref"},
		"handoff.open_risk_refs":             {viewSchemaID: "cartulary.view.handoff.v1", fieldKey: "handoff.open_risk_refs", expectedItemKind: "risk_ref"},
		"lesson.follow_up_task_ids":          {viewSchemaID: "cartulary.view.lesson.v1", fieldKey: "lesson.follow_up_task_ids", expectedItemKind: "record_ref"},
		"lesson.evidence_refs":               {viewSchemaID: "cartulary.view.lesson.v1", fieldKey: "lesson.evidence_refs", expectedItemKind: "record_ref"},
		"status_review.blocked_task_ids":     {viewSchemaID: "cartulary.view.status_review.v1", fieldKey: "status_review.blocked_task_ids", expectedItemKind: "record_ref"},
		"status_review.pending_evidence_ids": {viewSchemaID: "cartulary.view.status_review.v1", fieldKey: "status_review.pending_evidence_ids", expectedItemKind: "record_ref"},
		"status_review.open_decision_ids":    {viewSchemaID: "cartulary.view.status_review.v1", fieldKey: "status_review.open_decision_ids", expectedItemKind: "record_ref"},
		"task.linked_record_ids":             {viewSchemaID: "cartulary.view.task_requests.v1", fieldKey: "task.linked_record_ids", expectedItemKind: "record_ref"},
	}
}

func phase6BindCollectionActions(cases map[string]phase6CollectionReviewCase, ids phase6CollectionActionIDs) {
	bind := func(fieldKey string, server map[string]any, client map[string]any) {
		tc := cases[fieldKey]
		tc.serverPayload = collectionActions(server)
		tc.clientPayload = collectionActions(client)
		tc.resolvePayload = collectionActions(client)
		cases[fieldKey] = tc
	}
	bind("note.tags", addToken("server-tag"), addToken("client-tag"))
	bind("comm_log.decision_ids", addRecordRef(ids.SecondDecisionID), addRecordRef(ids.ThirdDecisionID))
	bind("comm_log.action_task_ids", addRecordRef(ids.SecondTaskID), addRecordRef(ids.ThirdTaskID))
	bind("comm_log.audience_party_ids", addPartyRef(ids.SecondPartyID), addPartyRef(ids.ThirdPartyID))
	bind("comm_log.attendee_party_ids", addPartyRef(ids.SecondPartyID), addPartyRef(ids.ThirdPartyID))
	bind("decision.support_refs", addRecordRef(ids.SecondEvidenceID), addRecordRef(ids.ThirdEvidenceID))
	bind("decision.affected_record_ids", addRecordRef(ids.SecondEvidenceID), addRecordRef(ids.ThirdEvidenceID))
	bind("handoff.open_task_ids", addRecordRef(ids.SecondTaskID), addRecordRef(ids.ThirdTaskID))
	bind("handoff.open_decision_ids", addRecordRef(ids.SecondDecisionID), addRecordRef(ids.ThirdDecisionID))
	bind("handoff.open_risk_refs", addRiskRef("Server risk"), addRiskRef("Client risk"))
	bind("lesson.follow_up_task_ids", addRecordRef(ids.SecondTaskID), addRecordRef(ids.ThirdTaskID))
	bind("lesson.evidence_refs", addRecordRef(ids.SecondEvidenceID), addRecordRef(ids.ThirdEvidenceID))
	bind("status_review.blocked_task_ids", addRecordRef(ids.SecondTaskID), addRecordRef(ids.ThirdTaskID))
	bind("status_review.pending_evidence_ids", addRecordRef(ids.SecondEvidenceID), addRecordRef(ids.ThirdEvidenceID))
	bind("status_review.open_decision_ids", addRecordRef(ids.SecondDecisionID), addRecordRef(ids.ThirdDecisionID))
	bind("task.linked_record_ids", addRecordRef(ids.SecondEvidenceID), addRecordRef(ids.ThirdEvidenceID))
}

func phase6RequireCollectionReviewCaseInventory(t testing.TB, cases map[string]phase6CollectionReviewCase) {
	t.Helper()
	got := make([]string, 0, len(cases))
	for fieldKey := range cases {
		got = append(got, fieldKey)
	}
	sort.Strings(got)
	want := phase6RouteSupportedCollectionReviewFields()
	if !phase6StringSlicesEqual(got, want) {
		t.Fatalf("collection_review case inventory drift: cases=%#v registry=%#v", got, want)
	}
}

func phase6RouteSupportedCollectionReviewFields() []string {
	routeSupportedSchemas := map[string]struct{}{
		"cartulary.view.comm_log.v1":      {},
		"cartulary.view.decisions.v1":     {},
		"cartulary.view.handoff.v1":       {},
		"cartulary.view.lesson.v1":        {},
		phase6NotesViewSchemaID:           {},
		"cartulary.view.status_review.v1": {},
		"cartulary.view.task_requests.v1": {},
	}
	fields := make([]string, 0)
	for _, resource := range viewschema.ListPublicResources() {
		if _, ok := routeSupportedSchemas[resource.ViewSchemaID]; !ok {
			continue
		}
		for _, field := range resource.Fields {
			if field.ConflictResolutionClass != nil && *field.ConflictResolutionClass == "collection_review" {
				fields = append(fields, field.FieldKey)
			}
		}
	}
	sort.Strings(fields)
	return fields
}

func phase6ExerciseCollectionReviewField(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, baseVersion int64, tc phase6CollectionReviewCase) int64 {
	t.Helper()
	serverData := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   tc.viewSchemaID,
		"base_row_version": baseVersion,
		"client_txn_id":    "txn-phase6-u-6-04-" + phase6TxnFieldSuffix(tc.fieldKey) + "-server",
		"changes": []map[string]any{{
			"field_key":      tc.fieldKey,
			"action_payload": tc.serverPayload,
		}},
	})
	serverRow := serverData["row"].(map[string]any)
	serverVersion := int64(serverRow["row_version"].(float64))
	requireCollectionValueHasItemKind(t, cellMapValue(t, serverRow, tc.fieldKey), tc.expectedItemKind)

	before := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   tc.viewSchemaID,
		"base_row_version": baseVersion,
		"client_txn_id":    "txn-phase6-u-6-04-" + phase6TxnFieldSuffix(tc.fieldKey) + "-client",
		"changes": []map[string]any{{
			"field_key":      tc.fieldKey,
			"action_payload": tc.clientPayload,
		}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	phase6RequireNoLegacyConflictAliases(t, conflict)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != tc.fieldKey ||
		conflict["conflict_resolution_class"] != "collection_review" ||
		conflict["server_updated_by"] != actorID.String() {
		t.Fatalf("unexpected collection conflict identity fields for %s: %#v", tc.fieldKey, conflict)
	}
	phase6RequireConflictVersion(t, conflict, "base_row_version", baseVersion)
	phase6RequireConflictVersion(t, conflict, "current_row_version", serverVersion)
	for _, key := range []string{"base_value", "server_value", "client_value"} {
		value := conflict[key].(map[string]any)
		if value["kind"] != "collection_value_v1" {
			t.Fatalf("%s.%s must use collection_value_v1, got %#v", tc.fieldKey, key, value)
		}
		requireCollectionValueHasItemKind(t, value, tc.expectedItemKind)
	}
	after := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if before != after {
		t.Fatalf("same-field collection conflict for %s wrote durable side effects: before=%+v after=%+v", tc.fieldKey, before, after)
	}

	resolved := phase6ResolveConflict(t, harness, login, recordID, conflict["conflict_token"].(string), map[string]any{
		"conflict_token":  conflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-phase6-u-6-04-" + phase6TxnFieldSuffix(tc.fieldKey) + "-resolve",
		"resolved_value":  tc.resolvePayload,
	})
	resolvedRow := resolved["row"].(map[string]any)
	requireCollectionValueHasItemKind(t, cellMapValue(t, resolvedRow, tc.fieldKey), tc.expectedItemKind)
	return int64(resolvedRow["row_version"].(float64))
}

func phase6TxnFieldSuffix(fieldKey string) string {
	replacer := strings.NewReplacer(".", "-", "_", "-")
	return replacer.Replace(fieldKey)
}

func phase6ConflictFixture(t testing.TB, name string, incidentKey string) (*phase4test.ServerHarness, phase4test.LoginResult, uuid.UUID, uuid.UUID) {
	t.Helper()
	harness := phase4test.StartServer(t, name)
	login, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-" + incidentKey,
		"incident_key":  incidentKey,
		"title":         name,
	})
	return harness, login, actorID, phase4test.MustUUID(t, incident["incident_id"].(string))
}

func phase6CreateNote(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, clientTxnID string, title string, body string) map[string]any {
	t.Helper()
	data := requireWorkbookCreate(t, harness, login, incidentID, phase6NotesViewSchemaID, map[string]any{
		"client_txn_id": clientTxnID,
		"note.title":    title,
		"note.body":     body,
	})
	return data["row"].(map[string]any)
}

func phase6CreateNoteTitleConflict(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, suffix string, savedValue string, localValue string) (uuid.UUID, map[string]any) {
	t.Helper()
	note := phase6CreateNote(t, harness, login, incidentID, "txn-phase6-u-6-06-"+suffix+"-create", suffix+" base", suffix+" body")
	recordID := phase4test.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-06-" + suffix + "-server",
		"changes":          []map[string]any{{"field_key": "note.title", "value": savedValue}},
	})
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   phase6NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase6-u-6-06-" + suffix + "-client",
		"changes":          []map[string]any{{"field_key": "note.title", "value": localValue}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "note.title" ||
		conflict["server_value"] != savedValue ||
		conflict["client_value"] != localValue {
		t.Fatalf("unexpected U-6-06 conflict payload: %#v", conflict)
	}
	return recordID, conflict
}

func phase6RequireNoChangeSetID(t testing.TB, data map[string]any) {
	t.Helper()
	if _, ok := data["change_set_id"]; ok {
		t.Fatalf("response must not carry change_set_id: %#v", data)
	}
}

func phase6RequireSideEffectDelta(t testing.TB, label string, before workbookConflictSideEffects, after workbookConflictSideEffects, want workbookConflictSideEffects) {
	t.Helper()
	got := workbookConflictSideEffects{
		ChangeSets:        after.ChangeSets - before.ChangeSets,
		MutationRows:      after.MutationRows - before.MutationRows,
		RecordRevisions:   after.RecordRevisions - before.RecordRevisions,
		RouteIdempotency:  after.RouteIdempotency - before.RouteIdempotency,
		ActiveRecordLinks: after.ActiveRecordLinks - before.ActiveRecordLinks,
		ActiveRiskRefs:    after.ActiveRiskRefs - before.ActiveRiskRefs,
	}
	if got != want {
		t.Fatalf("%s side-effect delta = %+v want %+v (before=%+v after=%+v)", label, got, want, before, after)
	}
}

func phase6RequireMutationChangedFields(t testing.TB, harness *phase4test.ServerHarness, changeSetID string, want []string) {
	t.Helper()
	var beforeJSON, afterJSON []byte
	row := harness.DB.QueryRowContext(context.Background(), `
SELECT before_value, after_value
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND sequence_no = 1
`, changeSetID)
	if err := row.Scan(&beforeJSON, &afterJSON); err != nil {
		t.Fatalf("load change-set mutation %s: %v", changeSetID, err)
	}
	beforeRow := phase6DecodeRowJSON(t, beforeJSON)
	afterRow := phase6DecodeRowJSON(t, afterJSON)
	got := phase6ChangedCellKeys(beforeRow, afterRow)
	if !phase6StringSlicesEqual(got, want) {
		t.Fatalf("change-set mutation changed fields = %#v want %#v", got, want)
	}
}

func phase6DecodeRowJSON(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	var row map[string]any
	if err := json.Unmarshal(payload, &row); err != nil {
		t.Fatalf("decode row json: %v payload=%s", err, payload)
	}
	return row
}

func phase6ChangedCellKeys(beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	keys := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		if phase6ServerManagedRevisionCell(fieldKey) {
			continue
		}
		beforeCell, ok := beforeCells[fieldKey]
		if !ok || !phase6JSONEqual(beforeCell, afterCell) {
			keys = append(keys, fieldKey)
		}
	}
	sort.Strings(keys)
	return keys
}

func phase6ServerManagedRevisionCell(fieldKey string) bool {
	switch fieldKey {
	case "note.updated_at":
		return true
	default:
		return false
	}
}

func phase6JSONEqual(left any, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}

func phase6StringSlicesEqual(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func phase6RequireNoLegacyConflictAliases(t testing.TB, conflict map[string]any) {
	t.Helper()
	for _, key := range []string{"current_field_value", "conflict_resolution"} {
		if _, ok := conflict[key]; ok {
			t.Fatalf("same-field conflict preserved legacy alias %q: %#v", key, conflict)
		}
	}
}

func phase6RequireConflictVersion(t testing.TB, conflict map[string]any, key string, want int64) {
	t.Helper()
	got, ok := conflict[key].(float64)
	if !ok || int64(got) != want {
		t.Fatalf("unexpected %s: got %#v want %d conflict=%#v", key, conflict[key], want, conflict)
	}
}

func phase6RequireQueriedRow(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		phase4test.WithCookies(login.SessionCookie),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	rows := body["data"].(map[string]any)["rows"].([]any)
	for _, row := range rows {
		rowMap := row.(map[string]any)
		if rowMap["record_id"] == recordID.String() {
			return rowMap
		}
	}
	t.Fatalf("query did not return record %s in %s rows=%#v", recordID, viewSchemaID, rows)
	return nil
}
