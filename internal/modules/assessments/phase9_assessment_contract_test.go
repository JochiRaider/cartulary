package assessments_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
)

func TestPhase9_AssessmentsAppendOnlyStatesAndBands_U_9_06(t *testing.T) {
	ctx := context.Background()
	harness := phase4storetest.StartStore(t, "phase9-assessments-u-9-06")
	assessmentStore := assessments.NewStore(harness.DB)
	workbookStore := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "phase9-u906@example.test", "Phase 9 U906", "Phase9U906Pass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-06-incident", "IR-PHASE9-U-9-06", "Phase 9 U-9-06 assessments")

	hostID := uuid.New()
	identityID := uuid.New()
	phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Phase 9 assessment host", "phase9-assessment-host", "", "")
	phase4storetest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, identityID, "Phase 9 assessment identity", "phase9@example.test", "phase9@example.test", "phase9")

	created := map[string]uuid.UUID{}
	for index, tc := range []struct {
		key         string
		subjectRef  uuid.UUID
		subjectType string
		state       string
		score       *int
		assessedAt  *time.Time
		wantBand    string
	}{
		{key: "unknown", subjectRef: hostID, subjectType: "host", state: "unknown", score: nil, assessedAt: nil, wantBand: "unset"},
		{key: "suspected", subjectRef: hostID, subjectType: "host", state: "suspected", score: intPtr(25), assessedAt: timePtr(time.Date(2026, 4, 24, 11, 0, 0, 0, time.UTC)), wantBand: "low"},
		{key: "confirmed", subjectRef: hostID, subjectType: "host", state: "confirmed", score: intPtr(55), assessedAt: timePtr(time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)), wantBand: "medium"},
		{key: "disproven", subjectRef: identityID, subjectType: "identity", state: "disproven", score: intPtr(69), assessedAt: timePtr(time.Date(2026, 4, 24, 13, 0, 0, 0, time.UTC)), wantBand: "medium"},
		{key: "cleared", subjectRef: hostID, subjectType: "host", state: "cleared", score: intPtr(85), assessedAt: timePtr(time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)), wantBand: "high"},
	} {
		request := assessments.CreateRequest{
			ClientTxnID:     "txn-phase9-u-9-06-" + tc.key,
			SubjectRef:      &tc.subjectRef,
			SubjectType:     tc.subjectType,
			AssessmentState: tc.state,
			ConfidenceScore: tc.score,
			Rationale:       "Phase 9 " + tc.state + " assessment rationale.",
			AssessedAt:      tc.assessedAt,
		}
		result, err := assessmentStore.CreateAssessmentRow(
			ctx,
			actor,
			incident.ID,
			request,
			assessments.CreateRequestHash(request),
			"req-phase9-u-9-06-"+tc.key,
			time.Date(2026, 5, 17, 17, index, 0, 0, time.UTC),
		)
		if err != nil {
			t.Fatalf("create %s assessment: %v", tc.key, err)
		}
		created[tc.key] = result.RecordID
		row := result.Payload["row"].(map[string]any)
		cells := row["cells"].(map[string]any)
		requireCellValue(t, cells, "assessment.subject_ref", tc.subjectRef.String())
		requireCellValue(t, cells, "assessment.subject_type", tc.subjectType)
		requireCellValue(t, cells, "assessment.assessment_state", tc.state)
		requireCellValue(t, cells, "assessment.confidence_band", tc.wantBand)
		requireCellValue(t, cells, "assessment.assessor", actor.ID.String())
		if tc.score == nil {
			requireCellValue(t, cells, "assessment.confidence_score", nil)
		}
		if got := cells["assessment.assessed_at"].(map[string]any)["value"]; got == "" {
			t.Fatalf("expected assessed_at to default to a persisted timestamp")
		}
	}

	if got := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID); got != 5 {
		t.Fatalf("expected five append-only assessment rows, got %d", got)
	}
	if got := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1 AND subject_record_id = $2`, incident.ID, hostID); got != 4 {
		t.Fatalf("expected host compromise judgments to be assessment history rows, got %d", got)
	}

	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.assessment_state", "cleared"), []uuid.UUID{created["cleared"]})
	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.assessment_state", "disproven"), []uuid.UUID{created["disproven"]})
	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.confidence_band", "unset"), []uuid.UUID{created["unknown"]})
	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.confidence_band", "low"), []uuid.UUID{created["suspected"]})
	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.confidence_band", "medium"), []uuid.UUID{created["disproven"], created["confirmed"]})
	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.confidence_band", "high"), []uuid.UUID{created["cleared"]})

	query := assessmentQueryMeta(t)
	query.Sort = []viewschema.SortEntry{
		{FieldKey: "assessment.confidence_band", Direction: "asc"},
		{FieldKey: "assessment.assessed_at", Direction: "desc"},
		{FieldKey: "record_id", Direction: "asc"},
	}
	rows, err := workbookStore.QueryRows(ctx, incident.ID, workbook.AssessmentsViewSchemaID, query)
	if err != nil {
		t.Fatalf("query assessment rows sorted by confidence band: %v", err)
	}
	requireRecordIDOrder(t, rows, []uuid.UUID{
		created["unknown"],
		created["suspected"],
		created["disproven"],
		created["confirmed"],
		created["cleared"],
	})

	for _, state := range []string{"contained", "isolated", "disabled", "reset", "monitored"} {
		before := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID)
		request := validCreateRequest(hostID, "host", state)
		request.ClientTxnID = "txn-phase9-u-9-06-operational-" + state
		if _, err := assessmentStore.CreateAssessmentRow(ctx, actor, incident.ID, request, assessments.CreateRequestHash(request), "req-operational-"+state, time.Now().UTC()); err == nil {
			t.Fatalf("expected operational state %q to fail closed", state)
		}
		if after := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID); after != before {
			t.Fatalf("operational state %q left partial assessment rows: before=%d after=%d", state, before, after)
		}
	}

	for _, tc := range []struct {
		name    string
		request assessments.CreateRequest
	}{
		{name: "missing subject", request: assessments.CreateRequest{
			ClientTxnID:     "txn-phase9-u-9-06-missing-subject",
			SubjectType:     "host",
			AssessmentState: "confirmed",
			Rationale:       "Subject was not supplied.",
		}},
		{name: "support refs do not satisfy minimum semantic set", request: assessments.CreateRequest{
			ClientTxnID: "txn-phase9-u-9-06-support-only",
			SupportRefs: []uuid.UUID{
				hostID,
			},
		}},
		{name: "empty rationale", request: assessments.CreateRequest{
			ClientTxnID:     "txn-phase9-u-9-06-empty-rationale",
			SubjectRef:      &hostID,
			SubjectType:     "host",
			AssessmentState: "confirmed",
		}},
	} {
		before := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID)
		if _, err := assessmentStore.CreateAssessmentRow(ctx, actor, incident.ID, tc.request, assessments.CreateRequestHash(tc.request), "req-"+tc.name, time.Now().UTC()); err == nil {
			t.Fatalf("expected %s to fail closed", tc.name)
		}
		if after := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID); after != before {
			t.Fatalf("%s left partial assessment rows: before=%d after=%d", tc.name, before, after)
		}
	}

	expectDecodeCreateRejected(t, map[string]any{
		"client_txn_id":               "txn-phase9-u-9-06-null-assessed-at",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "Explicit null timestamp.",
		"assessment.assessed_at":      nil,
	})
	expectDecodeCreateRejected(t, map[string]any{
		"client_txn_id":               "txn-phase9-u-9-06-no-zone-assessed-at",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "Timezone-less timestamp.",
		"assessment.assessed_at":      "2026-04-24T12:00:00",
	})
	expectDecodeCreateRejected(t, map[string]any{
		"client_txn_id":               "txn-phase9-u-9-06-padded-assessed-at",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "Whitespace padded timestamp.",
		"assessment.assessed_at":      " 2026-04-24T12:00:00Z",
	})

	patchPayload := map[string]any{
		"view_schema_id":   assessments.AssessmentsViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-u-9-06-semantic-patch",
		"changes": []map[string]any{
			{"field_key": "assessment.assessment_state", "value": "cleared"},
		},
	}
	data, err := json.Marshal(patchPayload)
	if err != nil {
		t.Fatalf("marshal patch payload: %v", err)
	}
	if _, apiErr := workbook.DecodePatchRequest(strings.NewReader(string(data))); apiErr == nil {
		t.Fatalf("expected in-place assessment semantic patch to be rejected")
	}
	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.assessment_state", "confirmed"), []uuid.UUID{created["confirmed"]})
}

func TestPhase9_U_9_12_RelationshipConfidenceRejectedAndManualLinksRemainNull(t *testing.T) {
	ctx := context.Background()
	harness := phase4storetest.StartStore(t, "phase9-assessments-u-9-12")
	assessmentStore := assessments.NewStore(harness.DB)
	workbookStore := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "phase9-u912@example.test", "Phase 9 U912", "Phase9U912Pass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-12-incident", "IR-PHASE9-U-9-12", "Phase 9 U-9-12 assessment links")
	hostID := uuid.New()
	supportID := uuid.New()
	phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Phase 9 assessment support host", "phase9-assessment-support", "", "")
	phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	request := validCreateRequest(hostID, "host", "confirmed")
	request.ClientTxnID = "txn-phase9-u-9-12-valid"
	request.SupportRefs = []uuid.UUID{supportID}
	result, err := assessmentStore.CreateAssessmentRow(ctx, actor, incident.ID, request, assessments.CreateRequestHash(request), "req-phase9-u-9-12-valid", time.Date(2026, 5, 17, 18, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create assessment with support ref: %v", err)
	}
	link := phase4storetest.LookupActiveLink(t, harness.DB, incident.ID, result.RecordID, supportID, "supported_by")
	if link.Provenance != "manual" || link.Confidence != nil {
		t.Fatalf("manual assessment support link must preserve provenance=manual confidence=NULL, got %#v", link)
	}

	before := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID)
	expectDecodeCreateRejected(t, map[string]any{
		"client_txn_id":               "txn-phase9-u-9-12-client-confidence",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "Client confidence must be rejected.",
		"assessment.support_refs": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
			},
		},
	})
	if after := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID); after != before {
		t.Fatalf("client-supplied support confidence left partial assessment rows: before=%d after=%d", before, after)
	}
	if got := queryCount(t, harness, `SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND src_record_id = $2 AND link_type = 'supported_by' AND confidence IS NOT NULL`, incident.ID, result.RecordID); got != 0 {
		t.Fatalf("expected no manual assessment support links with client confidence, got %d", got)
	}

	expectWorkbookDecodeCreateRejected(t, workbook.TaskRequestsViewSchemaID, map[string]any{
		"client_txn_id":  "txn-phase9-u-9-12-task-confidence-create",
		"task.title":     "Task confidence must be rejected",
		"task.task_kind": "collection",
		"task.linked_record_ids": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
			},
		},
	})
	expectWorkbookDecodePatchRejected(t, map[string]any{
		"view_schema_id":   workbook.TaskRequestsViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-u-9-12-task-confidence-patch",
		"changes": []map[string]any{
			{
				"field_key": "task.linked_record_ids",
				"action_payload": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
					},
				},
			},
		},
	})
	expectWorkbookDecodeCreateRejected(t, workbook.DecisionsViewSchemaID, map[string]any{
		"client_txn_id":          "txn-phase9-u-9-12-decision-support-confidence-create",
		"decision.summary":       "Decision confidence must be rejected",
		"decision.decision_type": "containment",
		"decision.rationale":     "Client confidence is not authoritative.",
		"decision.support_refs": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
			},
		},
	})
	expectWorkbookDecodePatchRejected(t, map[string]any{
		"view_schema_id":   workbook.DecisionsViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-u-9-12-decision-support-confidence-patch",
		"changes": []map[string]any{
			{
				"field_key": "decision.support_refs",
				"action_payload": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
					},
				},
			},
		},
	})
	expectWorkbookDecodeCreateRejected(t, workbook.DecisionsViewSchemaID, map[string]any{
		"client_txn_id":          "txn-phase9-u-9-12-decision-affected-confidence-create",
		"decision.summary":       "Decision affected confidence must be rejected",
		"decision.decision_type": "containment",
		"decision.rationale":     "Client confidence is not authoritative.",
		"decision.affected_record_ids": map[string]any{
			"kind": "collection_actions_v1",
			"actions": []map[string]any{
				{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
			},
		},
	})
	expectWorkbookDecodePatchRejected(t, map[string]any{
		"view_schema_id":   workbook.DecisionsViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-u-9-12-decision-affected-confidence-patch",
		"changes": []map[string]any{
			{
				"field_key": "decision.affected_record_ids",
				"action_payload": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
					},
				},
			},
		},
	})

	taskResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-12-task-null-confidence",
		Values: map[string]workbook.ValueChange{
			"task.title":     {Kind: "text", Text: textPtr("Task confidence remains null")},
			"task.task_kind": {Kind: "text", Text: textPtr("collection")},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"task.linked_record_ids": {
				Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}},
			},
		},
	}, []byte("txn-phase9-u-9-12-task-null-confidence"), "req-phase9-u-9-12-task-null-confidence", time.Date(2026, 5, 17, 18, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create task with manual link: %v", err)
	}
	decisionResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-12-decision-null-confidence",
		Values: map[string]workbook.ValueChange{
			"decision.summary":       {Kind: "text", Text: textPtr("Decision confidence remains null")},
			"decision.decision_type": {Kind: "text", Text: textPtr("containment")},
			"decision.rationale":     {Kind: "text", Text: textPtr("Manual relationship confidence remains null.")},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"decision.support_refs": {
				Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}},
			},
			"decision.affected_record_ids": {
				Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}},
			},
		},
	}, []byte("txn-phase9-u-9-12-decision-null-confidence"), "req-phase9-u-9-12-decision-null-confidence", time.Date(2026, 5, 17, 18, 45, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create decision with manual links: %v", err)
	}
	requireManualLinkConfidenceNull(t, harness, incident.ID, taskResult.RecordID, supportID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, decisionResult.RecordID, supportID, "supported_by")
	requireManualLinkConfidenceNull(t, harness, incident.ID, decisionResult.RecordID, supportID, "references_record")
}

func validCreateRequest(subjectRef uuid.UUID, subjectType string, state string) assessments.CreateRequest {
	return assessments.CreateRequest{
		ClientTxnID:     "txn-phase9-assessment-valid",
		SubjectRef:      &subjectRef,
		SubjectType:     subjectType,
		AssessmentState: state,
		Rationale:       "Valid rationale.",
	}
}

func filterEq(fieldKey string, value any) viewschema.Filter {
	return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": value}}
}

func requireQueriedRecordIDs(t testing.TB, store *workbook.Store, incidentID uuid.UUID, filter viewschema.Filter, want []uuid.UUID) {
	t.Helper()
	query := assessmentQueryMeta(t)
	query.Filters = []viewschema.Filter{filter}
	rows, err := store.QueryRows(context.Background(), incidentID, workbook.AssessmentsViewSchemaID, query)
	if err != nil {
		t.Fatalf("query assessment rows for %#v: %v", filter, err)
	}
	requireRecordIDOrder(t, rows, want)
}

func requireRecordIDOrder(t testing.TB, rows []map[string]any, want []uuid.UUID) {
	t.Helper()
	got := make([]string, 0, len(rows))
	for _, row := range rows {
		got = append(got, row["record_id"].(string))
	}
	if len(got) != len(want) {
		t.Fatalf("unexpected queried rows: got %#v want %#v", got, uuidStrings(want))
	}
	for idx := range want {
		if got[idx] != want[idx].String() {
			t.Fatalf("unexpected queried rows: got %#v want %#v", got, uuidStrings(want))
		}
	}
}

func assessmentQueryMeta(t testing.TB) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(workbook.AssessmentsViewSchemaID)
	if !ok {
		t.Fatalf("missing assessments view schema")
	}
	return schema.DefaultQueryMeta()
}

func queryCount(t testing.TB, harness *phase4storetest.StoreHarness, query string, args ...any) int {
	t.Helper()
	var got int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return got
}

func expectDecodeCreateRejected(t testing.TB, body map[string]any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal create body: %v", err)
	}
	if _, apiErr := assessments.DecodeCreateRequest(strings.NewReader(string(data))); apiErr == nil {
		t.Fatalf("expected create body to be rejected: %#v", body)
	}
}

func expectWorkbookDecodeCreateRejected(t testing.TB, viewSchemaID string, body map[string]any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal workbook create body: %v", err)
	}
	if _, apiErr := workbook.DecodeCreateRequest(viewSchemaID, strings.NewReader(string(data))); apiErr == nil {
		t.Fatalf("expected workbook create body to be rejected: %#v", body)
	} else if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
		t.Fatalf("unexpected workbook create error: %#v", apiErr)
	}
}

func expectWorkbookDecodePatchRejected(t testing.TB, body map[string]any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal workbook patch body: %v", err)
	}
	if _, apiErr := workbook.DecodePatchRequest(strings.NewReader(string(data))); apiErr == nil {
		t.Fatalf("expected workbook patch body to be rejected: %#v", body)
	} else if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" {
		t.Fatalf("unexpected workbook patch error: %#v", apiErr)
	}
}

func requireManualLinkConfidenceNull(t testing.TB, harness *phase4storetest.StoreHarness, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) {
	t.Helper()
	link := phase4storetest.LookupActiveLink(t, harness.DB, incidentID, sourceID, targetID, linkType)
	if link.Provenance != "manual" || link.Confidence != nil {
		t.Fatalf("manual %s link must preserve provenance=manual confidence=NULL, got %#v", linkType, link)
	}
}

func uuidStrings(values []uuid.UUID) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value.String())
	}
	return result
}

func textPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
