package workbook_test

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

const NotesViewSchemaID = "cartulary.view.notes.v1"

func TestGridWriteConcurrencyRoute_Unit(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "collaboration-u-6-01-grid-write-concurrency", "IR-COLLABORATION-workbook-storage")
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-01-create", "Base title", "Base body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))

	routeLeft := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-01-route-left-create", "Route left", "Left base body")
	routeLeftID := appsupport.MustUUID(t, routeLeft["record_id"].(string))
	routeRight := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-01-route-right-create", "Route right", "Right base body")
	routeRightID := appsupport.MustUUID(t, routeRight["record_id"].(string))
	routeBoundPatchBody := map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-route-bound-body",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "Path-selected body"}},
	}
	routeLeftPatch := requireWorkbookPatch(t, harness, login, routeLeftID, routeBoundPatchBody)
	requireCellValue(t, routeLeftPatch["row"].(map[string]any), "note.body", "Path-selected body")
	routeRightAfterLeft := RequireQueriedRow(t, harness, login, incidentID, NotesViewSchemaID, routeRightID)
	requireCellValue(t, routeRightAfterLeft, "note.body", "Right base body")
	routeRightPatch := requireWorkbookPatch(t, harness, login, routeRightID, routeBoundPatchBody)
	requireCellValue(t, routeRightPatch["row"].(map[string]any), "note.body", "Path-selected body")
	routeLeftAfterRight := RequireQueriedRow(t, harness, login, incidentID, NotesViewSchemaID, routeLeftID)
	requireCellValue(t, routeLeftAfterRight, "note.body", "Path-selected body")

	missingBase := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id": NotesViewSchemaID,
		"client_txn_id":  "txn-collaboration-u-6-01-missing-base",
		"changes":        []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	httptestx.RequireErrorEnvelope(t, missingBase, http.StatusBadRequest, "invalid_mutation_payload")
	fullRowPatch := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"record_id":        recordID.String(),
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-record-id-in-body",
		"note.body":        "Client body",
	})
	httptestx.RequireErrorEnvelope(t, fullRowPatch, http.StatusBadRequest, "invalid_mutation_payload")

	titlePatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-server-title",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})
	requireCellValue(t, titlePatch["row"].(map[string]any), "note.title", "Server title")

	bodyPatch := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-client-body",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "Client body"}},
	})
	bodyRow := bodyPatch["row"].(map[string]any)
	requireCellValue(t, bodyRow, "note.title", "Server title")
	requireCellValue(t, bodyRow, "note.body", "Client body")
	if got := int64(bodyRow["row_version"].(float64)); got != 3 {
		t.Fatalf("different-field stale edit row_version = %d want 3", got)
	}
	RequireMutationChangedFields(t, harness, bodyPatch["change_set_id"].(string), []string{"note.body"})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	sameField := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-01-client-title",
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

func TestSameFieldConflictHTTP_Unit(t *testing.T) {
	harness, login, actorID, incidentID := ConflictFixture(t, "collaboration-u-6-02-same-field-http", "IR-COLLABORATION-workbook-storage")
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-02-create", "Base title", "Base body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-02-server",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Server title"}},
	})

	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-02-client",
		"changes":          []map[string]any{{"field_key": "note.title", "value": "Client title"}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	errorObject := body["error"].(map[string]any)
	if errorObject["status"] != float64(http.StatusConflict) {
		t.Fatalf("same-field conflict error.status = %#v want 409", errorObject["status"])
	}
	conflict := errorObject["conflict"].(map[string]any)
	RequireNoLegacyConflictAliases(t, conflict)
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
	RequireConflictVersion(t, conflict, "base_row_version", 1)
	RequireConflictVersion(t, conflict, "current_row_version", 2)
	if _, err := time.Parse(time.RFC3339Nano, conflict["server_updated_at"].(string)); err != nil {
		t.Fatalf("server_updated_at was not RFC3339Nano: %v conflict=%#v", err, conflict)
	}
}

func TestTextCompareMergeDurability_Unit(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "collaboration-u-6-03-text-merge-durability", "IR-COLLABORATION-workbook-storage")
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-03-clean-create", "Merge note", "one\ntwo\nthree")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-clean-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nTWO\nthree"}},
	})

	beforeConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-clean-client",
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
	current := RequireQueriedRow(t, harness, login, incidentID, NotesViewSchemaID, recordID)
	requireCellValue(t, current, "note.body", "one\nTWO\nthree")
	if got := int64(current["row_version"].(float64)); got != 2 {
		t.Fatalf("rejected clean text conflict changed row_version = %d want 2", got)
	}

	resolved := ResolveConflict(t, harness, login, recordID, conflict["conflict_token"].(string), map[string]any{
		"conflict_token":  conflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-collaboration-u-6-03-clean-resolve",
		"resolved_value":  "one\nTWO\nTHREE",
	})
	requireCellValue(t, resolved["row"].(map[string]any), "note.body", "one\nTWO\nTHREE")
	afterResolve := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if afterResolve.ChangeSets != beforeConflict.ChangeSets+1 || afterResolve.RecordRevisions != beforeConflict.RecordRevisions+1 {
		t.Fatalf("explicit text resolution should be the next durable revision: before=%+v after=%+v", beforeConflict, afterResolve)
	}

	overlap := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-03-overlap-create", "Overlap note", "one\ntwo")
	overlapID := appsupport.MustUUID(t, overlap["record_id"].(string))
	requireWorkbookPatch(t, harness, login, overlapID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-overlap-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": "one\nserver"}},
	})
	overlapResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", overlapID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-overlap-client",
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
	rendered := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-03-rendered-create", "Rendered-looking note", renderedLookingBase)
	renderedID := appsupport.MustUUID(t, rendered["record_id"].(string))
	requireWorkbookPatch(t, harness, login, renderedID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-rendered-server",
		"changes":          []map[string]any{{"field_key": "note.body", "value": renderedLookingServer}},
	})
	beforeRenderedConflict := snapshotWorkbookConflictSideEffects(t, harness, incidentID, renderedID)
	renderedResp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", renderedID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-03-rendered-client",
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
	renderedCurrent := RequireQueriedRow(t, harness, login, incidentID, NotesViewSchemaID, renderedID)
	requireCellValue(t, renderedCurrent, "note.body", renderedLookingServer)
}

func TestCollectionReviewRouteResolve_Unit(t *testing.T) {
	harness, login, actorID, incidentID := ConflictFixture(t, "collaboration-u-6-04-collection-review-route", "IR-COLLABORATION-workbook-storage")
	RequireCollectionReviewCaseInventory(t, CollectionReviewCases())

	partyData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-collaboration-u-6-04-party-create",
		"party.display_name": "Incident Commander",
		"party.party_kind":   "person",
	})
	partyID := appsupport.MustUUID(t, partyData["row"].(map[string]any)["record_id"].(string))
	secondPartyData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-collaboration-u-6-04-party-second-create",
		"party.display_name": "Security Lead",
		"party.party_kind":   "person",
	})
	secondPartyID := appsupport.MustUUID(t, secondPartyData["row"].(map[string]any)["record_id"].(string))
	thirdPartyData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-collaboration-u-6-04-party-third-create",
		"party.display_name": "Legal Observer",
		"party.party_kind":   "person",
	})
	thirdPartyID := appsupport.MustUUID(t, thirdPartyData["row"].(map[string]any)["record_id"].(string))

	decisionID := seedDecisionRecord(t, harness, incidentID, actorID, "Approve containment")
	secondDecisionID := seedDecisionRecord(t, harness, incidentID, actorID, "Approve credential rotation")
	thirdDecisionID := seedDecisionRecord(t, harness, incidentID, actorID, "Approve network block")
	taskID := seedTaskRecord(t, harness, incidentID, actorID, "Collect endpoint logs")
	secondTaskID := seedTaskRecord(t, harness, incidentID, actorID, "Collect VPN logs")
	thirdTaskID := seedTaskRecord(t, harness, incidentID, actorID, "Collect firewall logs")
	evidenceID := seedEvidenceRecord(t, harness, incidentID, actorID, "Packet capture")
	secondEvidenceID := seedEvidenceRecord(t, harness, incidentID, actorID, "VPN log bundle")
	thirdEvidenceID := seedEvidenceRecord(t, harness, incidentID, actorID, "Firewall log bundle")

	noteData := requireWorkbookCreate(t, harness, login, incidentID, NotesViewSchemaID, map[string]any{
		"client_txn_id": "txn-collaboration-u-6-04-create",
		"note.title":    "Collection note",
		"note.tags":     collectionActions(addToken("base-tag")),
	})
	note := noteData["row"].(map[string]any)
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireCollectionValueHasItemKind(t, cellMapValue(t, note, "note.tags"), "tag")
	queried := RequireQueriedRow(t, harness, login, incidentID, NotesViewSchemaID, recordID)
	requireCollectionValueHasItemKind(t, cellMapValue(t, queried, "note.tags"), "tag")

	commData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.comm_log.v1", map[string]any{
		"client_txn_id":               "txn-collaboration-u-6-04-comm-create",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "leadership",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Initial coordination update",
		"comm_log.decision_ids":       collectionActions(addRecordRef(decisionID)),
		"comm_log.action_task_ids":    collectionActions(addRecordRef(taskID)),
		"comm_log.audience_party_ids": collectionActions(addPartyRef(partyID)),
		"comm_log.attendee_party_ids": collectionActions(addPartyRef(partyID)),
	})
	commID := appsupport.MustUUID(t, commData["row"].(map[string]any)["record_id"].(string))
	handoffData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.handoff.v1", map[string]any{
		"client_txn_id":                  "txn-collaboration-u-6-04-handoff-create",
		"handoff.incoming_owner_user_id": actorID.String(),
		"handoff.current_state_summary":  "Night shift owns containment",
		"handoff.open_task_ids":          collectionActions(addRecordRef(taskID)),
		"handoff.open_decision_ids":      collectionActions(addRecordRef(decisionID)),
		"handoff.open_risk_refs":         collectionActions(addRiskRef("Base risk")),
	})
	handoffID := appsupport.MustUUID(t, handoffData["row"].(map[string]any)["record_id"].(string))
	statusData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.status_review.v1", map[string]any{
		"client_txn_id":                       "txn-collaboration-u-6-04-status-create",
		"status_review.current_state_summary": "Containment is stable",
		"status_review.blocked_task_ids":      collectionActions(addRecordRef(taskID)),
		"status_review.pending_evidence_ids":  collectionActions(addRecordRef(evidenceID)),
		"status_review.open_decision_ids":     collectionActions(addRecordRef(decisionID)),
	})
	statusID := appsupport.MustUUID(t, statusData["row"].(map[string]any)["record_id"].(string))
	lessonData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.lesson.v1", map[string]any{
		"client_txn_id":             "txn-collaboration-u-6-04-lesson-create",
		"lesson.summary":            "Preserve VPN logs earlier",
		"lesson.follow_up_task_ids": collectionActions(addRecordRef(taskID)),
		"lesson.evidence_refs":      collectionActions(addRecordRef(evidenceID)),
	})
	lessonID := appsupport.MustUUID(t, lessonData["row"].(map[string]any)["record_id"].(string))
	decisionData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.decisions.v1", map[string]any{
		"client_txn_id":                "txn-collaboration-u-6-04-decision-create",
		"decision.summary":             "Contain endpoint",
		"decision.decision_type":       "containment",
		"decision.rationale":           "Containment is required.",
		"decision.support_refs":        collectionActions(addRecordRef(evidenceID)),
		"decision.affected_record_ids": collectionActions(addRecordRef(evidenceID)),
	})
	decisionRowID := appsupport.MustUUID(t, decisionData["row"].(map[string]any)["record_id"].(string))
	taskData := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.task_requests.v1", map[string]any{
		"client_txn_id":          "txn-collaboration-u-6-04-task-create",
		"task.title":             "Collect endpoint logs",
		"task.task_kind":         "collection",
		"task.linked_record_ids": collectionActions(addRecordRef(evidenceID)),
	})
	taskRowID := appsupport.MustUUID(t, taskData["row"].(map[string]any)["record_id"].(string))

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
	cases := CollectionReviewCases()
	BindCollectionActions(cases, CollectionActionIDs{
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
		versions[rowID] = ExerciseCollectionReviewField(t, harness, login, incidentID, actorID, rowID, versions[rowID], tc)
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
				"view_schema_id":   NotesViewSchemaID,
				"base_row_version": 3,
				"client_txn_id":    "txn-collaboration-u-6-04-invalid-" + invalid.name,
				"changes": []map[string]any{{
					"field_key":      "note.tags",
					"action_payload": invalid.payload,
				}},
			})
			httptestx.RequireErrorEnvelope(t, invalidResp, http.StatusBadRequest, "invalid_mutation_payload")
		})
	}
}

func TestConflictResolveDurability_Unit(t *testing.T) {
	harness, login, actorID, incidentID := ConflictFixture(t, "collaboration-u-6-06-conflict-resolve-durability", "IR-COLLABORATION-workbook-storage")

	keepRecordID, keepConflict := CreateNoteTitleConflict(t, harness, login, incidentID, "keep", "Keep saved", "Keep local")
	beforeKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID)
	keepBody := map[string]any{
		"conflict_token":  keepConflict["conflict_token"].(string),
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-collaboration-u-6-06-keep-resolve",
	}
	keepData := ResolveConflict(t, harness, login, keepRecordID, keepConflict["conflict_token"].(string), keepBody)
	RequireNoChangeSetID(t, keepData)
	requireCellValue(t, keepData["row"].(map[string]any), "note.title", "Keep saved")
	afterKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID)
	RequireSideEffectDelta(t, "keep_saved", beforeKeep, afterKeep, workbookConflictSideEffects{
		RouteIdempotency: 1,
	})
	keepReplayData := ResolveConflict(t, harness, login, keepRecordID, keepConflict["conflict_token"].(string), keepBody)
	if !reflect.DeepEqual(keepReplayData, keepData) {
		t.Fatalf("keep_saved exact replay changed the original success:\nfirst  %#v\nreplay %#v", keepData, keepReplayData)
	}
	RequireSideEffectDelta(t, "keep_saved replay", beforeKeep, snapshotWorkbookConflictSideEffects(t, harness, incidentID, keepRecordID), workbookConflictSideEffects{
		RouteIdempotency: 1,
	})

	useRecordID, useConflict := CreateNoteTitleConflict(t, harness, login, incidentID, "use", "Use saved", "Use local")
	beforeUse := snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID)
	useBody := map[string]any{
		"conflict_token":  useConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-collaboration-u-6-06-use-resolve",
		"resolved_value":  "Use local",
	}
	useData := ResolveConflict(t, harness, login, useRecordID, useConflict["conflict_token"].(string), useBody)
	requireCellValue(t, useData["row"].(map[string]any), "note.title", "Use local")
	workbookscenariotest.RequireChangeSetAttribution(t, harness.DB, useData["change_set_id"].(string), actorID.String(), "workbook.records.conflicts.resolve", "txn-collaboration-u-6-06-use-resolve")
	afterUse := snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID)
	RequireSideEffectDelta(t, "use_unsaved", beforeUse, afterUse, workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
	useReplayData := ResolveConflict(t, harness, login, useRecordID, useConflict["conflict_token"].(string), useBody)
	if !reflect.DeepEqual(useReplayData, useData) {
		t.Fatalf("use_unsaved exact replay changed the original success:\nfirst  %#v\nreplay %#v", useData, useReplayData)
	}
	RequireSideEffectDelta(t, "use_unsaved replay", beforeUse, snapshotWorkbookConflictSideEffects(t, harness, incidentID, useRecordID), workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})

	mergedRecordID, mergedConflict := CreateNoteTitleConflict(t, harness, login, incidentID, "merged", "Merged saved", "Merged local")
	beforeMerged := snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID)
	mergedBody := map[string]any{
		"conflict_token":  mergedConflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-collaboration-u-6-06-merged-resolve",
		"resolved_value":  "Merged final",
	}
	mergedData := ResolveConflict(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), mergedBody)
	requireCellValue(t, mergedData["row"].(map[string]any), "note.title", "Merged final")
	workbookscenariotest.RequireChangeSetAttribution(t, harness.DB, mergedData["change_set_id"].(string), actorID.String(), "workbook.records.conflicts.resolve", "txn-collaboration-u-6-06-merged-resolve")
	afterMerged := snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID)
	RequireSideEffectDelta(t, "merged_value", beforeMerged, afterMerged, workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
	mergedReplayData := ResolveConflict(t, harness, login, mergedRecordID, mergedConflict["conflict_token"].(string), mergedBody)
	if !reflect.DeepEqual(mergedReplayData, mergedData) {
		t.Fatalf("merged_value exact replay changed the original success:\nfirst  %#v\nreplay %#v", mergedData, mergedReplayData)
	}
	RequireSideEffectDelta(t, "merged_value replay", beforeMerged, snapshotWorkbookConflictSideEffects(t, harness, incidentID, mergedRecordID), workbookConflictSideEffects{
		ChangeSets:       1,
		MutationRows:     1,
		RecordRevisions:  1,
		RouteIdempotency: 1,
	})
}

func TestSourceOwnerConflictContributions_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "workbook-source-owner-conflict-contributions", "IR-WORKBOOK-SOURCE-CONFLICTS")

	party := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.parties.v1", map[string]any{
		"client_txn_id":      "txn-workbook-source-conflict-party-create",
		"party.display_name": "Party base",
		"party.party_kind":   "organization",
	})["row"].(map[string]any)
	partyID := appsupport.MustUUID(t, party["record_id"].(string))
	partyConflict := createAtomicConflict(
		t,
		harness,
		login,
		partyID,
		"cartulary.view.parties.v1",
		"party.display_name",
		"Party saved",
		"Party local",
		"party",
	)
	partyResolved := ResolveConflict(t, harness, login, partyID, partyConflict["conflict_token"].(string), map[string]any{
		"conflict_token":  partyConflict["conflict_token"].(string),
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-workbook-source-conflict-party-resolve",
		"resolved_value":  "Party local",
	})
	requireCellValue(t, partyResolved["row"].(map[string]any), "party.display_name", "Party local")

	evidence := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-workbook-source-conflict-evidence-create",
		"evidence.title": "Evidence base",
	})["row"].(map[string]any)
	evidenceID := appsupport.MustUUID(t, evidence["record_id"].(string))
	evidenceConflict := createAtomicConflict(
		t,
		harness,
		login,
		evidenceID,
		"cartulary.view.evidence.v1",
		"evidence.title",
		"Evidence saved",
		"Evidence local",
		"evidence",
	)
	beforeKeep := snapshotWorkbookConflictSideEffects(t, harness, incidentID, evidenceID)
	keepBody := map[string]any{
		"conflict_token":  evidenceConflict["conflict_token"].(string),
		"resolution_kind": "keep_saved",
		"client_txn_id":   "txn-workbook-source-conflict-evidence-keep",
	}
	evidenceKept := ResolveConflict(t, harness, login, evidenceID, evidenceConflict["conflict_token"].(string), keepBody)
	RequireNoChangeSetID(t, evidenceKept)
	requireCellValue(t, evidenceKept["row"].(map[string]any), "evidence.title", "Evidence saved")
	RequireSideEffectDelta(
		t,
		"source-owned evidence keep_saved",
		beforeKeep,
		snapshotWorkbookConflictSideEffects(t, harness, incidentID, evidenceID),
		workbookConflictSideEffects{RouteIdempotency: 1},
	)
	evidenceReplay := ResolveConflict(t, harness, login, evidenceID, evidenceConflict["conflict_token"].(string), keepBody)
	if !reflect.DeepEqual(evidenceReplay, evidenceKept) {
		t.Fatalf("source-owned evidence keep_saved replay changed the original success:\nfirst  %#v\nreplay %#v", evidenceKept, evidenceReplay)
	}
	RequireSideEffectDelta(
		t,
		"source-owned evidence keep_saved replay",
		beforeKeep,
		snapshotWorkbookConflictSideEffects(t, harness, incidentID, evidenceID),
		workbookConflictSideEffects{RouteIdempotency: 1},
	)
}

func TestConflictResolutionRevalidatesStaleToken_Unit(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "collaboration-u-6-07-stale-conflict-resolution", "IR-COLLABORATION-STALE-CONFLICT")
	recordID, conflict := CreateNoteTitleConflict(t, harness, login, incidentID, "stale", "Saved at token issue", "Unsaved at token issue")
	oldToken := conflict["conflict_token"].(string)

	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 2,
		"client_txn_id":    "txn-collaboration-u-6-07-server-advanced",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "Saved after token issue",
		}},
	})
	beforeRejectedResolution := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)

	response := ResolveConflictRaw(t, harness, login, recordID, oldToken, map[string]any{
		"conflict_token":  oldToken,
		"resolution_kind": "use_unsaved",
		"client_txn_id":   "txn-collaboration-u-6-07-stale-resolution",
		"resolved_value":  "Unsaved at token issue",
	})
	body := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
	fresh := body["error"].(map[string]any)["conflict"].(map[string]any)
	if fresh["conflict_token"] == oldToken {
		t.Fatalf("stale resolution returned the consumed token instead of a fresh conflict: %#v", fresh)
	}
	RequireConflictVersion(t, fresh, "base_row_version", 2)
	RequireConflictVersion(t, fresh, "current_row_version", 3)
	if got := fresh["server_value"]; got != "Saved after token issue" {
		t.Fatalf("fresh conflict did not revalidate current source value: %#v", fresh)
	}
	afterRejectedResolution := snapshotWorkbookConflictSideEffects(t, harness, incidentID, recordID)
	if afterRejectedResolution != beforeRejectedResolution {
		t.Fatalf("stale conflict resolution wrote durable effects: before=%+v after=%+v", beforeRejectedResolution, afterRejectedResolution)
	}
}

func createAtomicConflict(
	t testing.TB,
	harness *appsupport.ServerHarness,
	login appsupport.LoginResult,
	recordID uuid.UUID,
	viewSchemaID string,
	fieldKey string,
	savedValue any,
	localValue any,
	txnSuffix string,
) map[string]any {
	t.Helper()
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-source-conflict-" + txnSuffix + "-server",
		"changes":          []map[string]any{{"field_key": fieldKey, "value": savedValue}},
	})
	response := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook-source-conflict-" + txnSuffix + "-client",
		"changes":          []map[string]any{{"field_key": fieldKey, "value": localValue}},
	})
	body := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != fieldKey ||
		conflict["server_value"] != savedValue ||
		conflict["client_value"] != localValue ||
		conflict["conflict_token"] == "" {
		t.Fatalf("unexpected source-owner conflict payload for %s: %#v", fieldKey, conflict)
	}
	return conflict
}

type CollectionActionIDs struct {
	SecondDecisionID uuid.UUID
	ThirdDecisionID  uuid.UUID
	SecondTaskID     uuid.UUID
	ThirdTaskID      uuid.UUID
	SecondEvidenceID uuid.UUID
	ThirdEvidenceID  uuid.UUID
	SecondPartyID    uuid.UUID
	ThirdPartyID     uuid.UUID
}

type CollectionReviewCase struct {
	viewSchemaID     string
	fieldKey         string
	expectedItemKind string
	serverPayload    map[string]any
	clientPayload    map[string]any
	resolvePayload   map[string]any
}

func CollectionReviewCases() map[string]CollectionReviewCase {
	return map[string]CollectionReviewCase{
		"note.tags":                          {viewSchemaID: NotesViewSchemaID, fieldKey: "note.tags", expectedItemKind: "tag"},
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

func BindCollectionActions(cases map[string]CollectionReviewCase, ids CollectionActionIDs) {
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

func RequireCollectionReviewCaseInventory(t testing.TB, cases map[string]CollectionReviewCase) {
	t.Helper()
	got := make([]string, 0, len(cases))
	for fieldKey := range cases {
		got = append(got, fieldKey)
	}
	sort.Strings(got)
	want := RouteSupportedCollectionReviewFields()
	if !StringSlicesEqual(got, want) {
		t.Fatalf("collection_review case inventory drift: cases=%#v registry=%#v", got, want)
	}
}

func RouteSupportedCollectionReviewFields() []string {
	routeSupportedSchemas := map[string]struct{}{
		"cartulary.view.comm_log.v1":      {},
		"cartulary.view.decisions.v1":     {},
		"cartulary.view.handoff.v1":       {},
		"cartulary.view.lesson.v1":        {},
		NotesViewSchemaID:                 {},
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

func ExerciseCollectionReviewField(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, baseVersion int64, tc CollectionReviewCase) int64 {
	t.Helper()
	serverData := requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   tc.viewSchemaID,
		"base_row_version": baseVersion,
		"client_txn_id":    "txn-collaboration-u-6-04-" + TxnFieldSuffix(tc.fieldKey) + "-server",
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
		"client_txn_id":    "txn-collaboration-u-6-04-" + TxnFieldSuffix(tc.fieldKey) + "-client",
		"changes": []map[string]any{{
			"field_key":      tc.fieldKey,
			"action_payload": tc.clientPayload,
		}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	RequireNoLegacyConflictAliases(t, conflict)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != tc.fieldKey ||
		conflict["conflict_resolution_class"] != "collection_review" ||
		conflict["server_updated_by"] != actorID.String() {
		t.Fatalf("unexpected collection conflict identity fields for %s: %#v", tc.fieldKey, conflict)
	}
	RequireConflictVersion(t, conflict, "base_row_version", baseVersion)
	RequireConflictVersion(t, conflict, "current_row_version", serverVersion)
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

	resolved := ResolveConflict(t, harness, login, recordID, conflict["conflict_token"].(string), map[string]any{
		"conflict_token":  conflict["conflict_token"].(string),
		"resolution_kind": "merged_value",
		"client_txn_id":   "txn-collaboration-u-6-04-" + TxnFieldSuffix(tc.fieldKey) + "-resolve",
		"resolved_value":  tc.resolvePayload,
	})
	resolvedRow := resolved["row"].(map[string]any)
	requireCollectionValueHasItemKind(t, cellMapValue(t, resolvedRow, tc.fieldKey), tc.expectedItemKind)
	return int64(resolvedRow["row_version"].(float64))
}

func TxnFieldSuffix(fieldKey string) string {
	replacer := strings.NewReplacer(".", "-", "_", "-")
	return replacer.Replace(fieldKey)
}

func ConflictFixture(t testing.TB, name string, incidentKey string) (*appsupport.ServerHarness, appsupport.LoginResult, uuid.UUID, uuid.UUID) {
	t.Helper()
	harness := appsupport.StartServer(t, name)
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-" + incidentKey,
		"incident_key":  incidentKey,
		"title":         name,
	})
	return harness, login, actorID, appsupport.MustUUID(t, incident["incident_id"].(string))
}

func CreateNote(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, clientTxnID string, title string, body string) map[string]any {
	t.Helper()
	data := requireWorkbookCreate(t, harness, login, incidentID, NotesViewSchemaID, map[string]any{
		"client_txn_id": clientTxnID,
		"note.title":    title,
		"note.body":     body,
	})
	return data["row"].(map[string]any)
}

func CreateNoteTitleConflict(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, suffix string, savedValue string, localValue string) (uuid.UUID, map[string]any) {
	t.Helper()
	note := CreateNote(t, harness, login, incidentID, "txn-collaboration-u-6-06-"+suffix+"-create", suffix+" base", suffix+" body")
	recordID := appsupport.MustUUID(t, note["record_id"].(string))
	requireWorkbookPatch(t, harness, login, recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-06-" + suffix + "-server",
		"changes":          []map[string]any{{"field_key": "note.title", "value": savedValue}},
	})
	resp := doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   NotesViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-collaboration-u-6-06-" + suffix + "-client",
		"changes":          []map[string]any{{"field_key": "note.title", "value": localValue}},
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "same_field_conflict")
	conflict := body["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["conflict_token"] == "" ||
		conflict["record_id"] != recordID.String() ||
		conflict["field_key"] != "note.title" ||
		conflict["server_value"] != savedValue ||
		conflict["client_value"] != localValue {
		t.Fatalf("unexpected workbook-storage conflict payload: %#v", conflict)
	}
	return recordID, conflict
}

func RequireNoChangeSetID(t testing.TB, data map[string]any) {
	t.Helper()
	if _, ok := data["change_set_id"]; ok {
		t.Fatalf("response must not carry change_set_id: %#v", data)
	}
}

func RequireSideEffectDelta(t testing.TB, label string, before workbookConflictSideEffects, after workbookConflictSideEffects, want workbookConflictSideEffects) {
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

func RequireMutationChangedFields(t testing.TB, harness *appsupport.ServerHarness, changeSetID string, want []string) {
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
	beforeRow := DecodeRowJSON(t, beforeJSON)
	afterRow := DecodeRowJSON(t, afterJSON)
	got := ChangedCellKeys(beforeRow, afterRow)
	if !StringSlicesEqual(got, want) {
		t.Fatalf("change-set mutation changed fields = %#v want %#v", got, want)
	}
}

func DecodeRowJSON(t testing.TB, payload []byte) map[string]any {
	t.Helper()
	var row map[string]any
	if err := json.Unmarshal(payload, &row); err != nil {
		t.Fatalf("decode row json: %v payload=%s", err, payload)
	}
	return row
}

func ChangedCellKeys(beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	keys := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		if ServerManagedRevisionCell(fieldKey) {
			continue
		}
		beforeCell, ok := beforeCells[fieldKey]
		if !ok || !JSONEqual(beforeCell, afterCell) {
			keys = append(keys, fieldKey)
		}
	}
	sort.Strings(keys)
	return keys
}

func ServerManagedRevisionCell(fieldKey string) bool {
	switch fieldKey {
	case "note.updated_at":
		return true
	default:
		return false
	}
}

func JSONEqual(left any, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}

func StringSlicesEqual(left []string, right []string) bool {
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

func RequireNoLegacyConflictAliases(t testing.TB, conflict map[string]any) {
	t.Helper()
	for _, key := range []string{"current_field_value", "conflict_resolution"} {
		if _, ok := conflict[key]; ok {
			t.Fatalf("same-field conflict preserved legacy alias %q: %#v", key, conflict)
		}
	}
}

func RequireConflictVersion(t testing.TB, conflict map[string]any, key string, want int64) {
	t.Helper()
	got, ok := conflict[key].(float64)
	if !ok || int64(got) != want {
		t.Fatalf("unexpected %s: got %#v want %d conflict=%#v", key, conflict[key], want, conflict)
	}
}

func RequireQueriedRow(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string, recordID uuid.UUID) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		appsupport.WithCookies(login.SessionCookie),
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
