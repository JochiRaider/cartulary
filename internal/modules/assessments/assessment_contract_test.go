package assessments_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	recordstoretest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestAssessmentsAppendOnlyStatesAndBands_Unit(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "workbook_interaction-assessments-u-9-06")
	assessmentStore := assessments.NewStore(harness.DB)
	projectionQuery := timelineassembly.NewBundle(harness.DB, conflicttest.NewCodec("timeline")).ProjectionCatalog.Query
	workbookStore := workbook.NewStore(harness.DB, conflicttest.NewCodec("workbook"), projectionQuery)
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "workbook_interaction-u906@example.test", "Workbook inspector U906", "WorkbookInteractionU906Pass1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-06-incident", "IR-WORKBOOK-INTERACTION-assessment-storage", "Workbook inspector assessment-storage assessments")

	hostID := uuid.New()
	identityID := uuid.New()
	recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Workbook inspector assessment host", "workbook_interaction-assessment-host", "", "")
	recordstoretest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, identityID, "Workbook inspector assessment identity", "workbook_interaction@example.test", "workbook_interaction@example.test", "workbook_interaction")

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
			ClientTxnID:     "txn-workbook_interaction-u-9-06-" + tc.key,
			SubjectRef:      &tc.subjectRef,
			SubjectType:     tc.subjectType,
			AssessmentState: tc.state,
			ConfidenceScore: tc.score,
			Rationale:       "Workbook inspector " + tc.state + " assessment rationale.",
			AssessedAt:      tc.assessedAt,
		}
		result, err := assessmentStore.CreateAssessmentRow(
			ctx,
			actor,
			incident.ID,
			request,
			assessments.CreateRequestHash(request),
			"req-workbook_interaction-u-9-06-"+tc.key,
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
		request.ClientTxnID = "txn-workbook_interaction-u-9-06-operational-" + state
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
			ClientTxnID:     "txn-workbook_interaction-u-9-06-missing-subject",
			SubjectType:     "host",
			AssessmentState: "confirmed",
			Rationale:       "Subject was not supplied.",
		}},
		{name: "support refs do not satisfy minimum semantic set", request: assessments.CreateRequest{
			ClientTxnID: "txn-workbook_interaction-u-9-06-support-only",
			SupportRefs: []uuid.UUID{
				hostID,
			},
		}},
		{name: "empty rationale", request: assessments.CreateRequest{
			ClientTxnID:     "txn-workbook_interaction-u-9-06-empty-rationale",
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
		"client_txn_id":               "txn-workbook_interaction-u-9-06-null-assessed-at",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "Explicit null timestamp.",
		"assessment.assessed_at":      nil,
	})
	expectDecodeCreateRejected(t, map[string]any{
		"client_txn_id":               "txn-workbook_interaction-u-9-06-no-zone-assessed-at",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "Timezone-less timestamp.",
		"assessment.assessed_at":      "2026-04-24T12:00:00",
	})
	expectDecodeCreateRejected(t, map[string]any{
		"client_txn_id":               "txn-workbook_interaction-u-9-06-padded-assessed-at",
		"assessment.subject_ref":      hostID.String(),
		"assessment.subject_type":     "host",
		"assessment.assessment_state": "confirmed",
		"assessment.rationale":        "Whitespace padded timestamp.",
		"assessment.assessed_at":      " 2026-04-24T12:00:00Z",
	})

	patchPayload := map[string]any{
		"view_schema_id":   assessments.AssessmentsViewSchemaID,
		"base_row_version": 1,
		"client_txn_id":    "txn-workbook_interaction-u-9-06-semantic-patch",
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

func TestRelationshipConfidenceRejectedAndManualLinksRemainNull_Unit(t *testing.T) {
	ctx := context.Background()
	harness := recordstoretest.StartStore(t, "workbook_interaction-assessments-u-9-12")
	assessmentStore := assessments.NewStore(harness.DB)
	workbookStore := appsupport.NewWorkbookStore(harness.DB, conflicttest.NewCodec("workbook"))
	actor := recordstoretest.SeedLocalUserFlags(t, harness.DB, "workbook_interaction-u912@example.test", "Workbook inspector U912", "WorkbookInteractionU912Pass1!", false, false, true)
	incident := recordstoretest.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-12-incident", "IR-WORKBOOK-INTERACTION-assessment-storage", "Workbook inspector assessment-storage assessment links")
	hostID := uuid.New()
	supportID := uuid.New()
	recordstoretest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Workbook inspector assessment support host", "workbook_interaction-assessment-support", "", "")
	recordstoretest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	request := validCreateRequest(hostID, "host", "confirmed")
	request.ClientTxnID = "txn-workbook_interaction-u-9-12-valid"
	request.SupportRefs = []uuid.UUID{supportID}
	result, err := assessmentStore.CreateAssessmentRow(ctx, actor, incident.ID, request, assessments.CreateRequestHash(request), "req-workbook_interaction-u-9-12-valid", time.Date(2026, 5, 17, 18, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create assessment with support ref: %v", err)
	}
	link := recordstoretest.LookupActiveLink(t, harness.DB, incident.ID, result.RecordID, supportID, "supported_by")
	if link.Provenance != "manual" || link.Confidence != nil {
		t.Fatalf("manual assessment support link must preserve provenance=manual confidence=NULL, got %#v", link)
	}

	before := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID)
	expectDecodeCreateRejected(t, map[string]any{
		"client_txn_id":               "txn-workbook_interaction-u-9-12-client-confidence",
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
		"client_txn_id":  "txn-workbook_interaction-u-9-12-task-confidence-create",
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
		"client_txn_id":    "txn-workbook_interaction-u-9-12-task-confidence-patch",
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
		"client_txn_id":          "txn-workbook_interaction-u-9-12-decision-support-confidence-create",
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
		"client_txn_id":    "txn-workbook_interaction-u-9-12-decision-support-confidence-patch",
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
		"client_txn_id":          "txn-workbook_interaction-u-9-12-decision-affected-confidence-create",
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
		"client_txn_id":    "txn-workbook_interaction-u-9-12-decision-affected-confidence-patch",
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

	for _, tc := range []struct {
		name         string
		viewSchemaID string
		fieldKey     string
		createBody   map[string]any
		action       map[string]any
	}{
		{
			name:         "comm-decision",
			viewSchemaID: workbook.CommLogViewSchemaID,
			fieldKey:     "comm_log.decision_ids",
			createBody: map[string]any{
				"comm_log.comm_type":          "briefing",
				"comm_log.audience":           "leadership",
				"comm_log.channel_or_meeting": "Bridge",
				"comm_log.summary":            "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "comm-action-task",
			viewSchemaID: workbook.CommLogViewSchemaID,
			fieldKey:     "comm_log.action_task_ids",
			createBody: map[string]any{
				"comm_log.comm_type":          "briefing",
				"comm_log.audience":           "leadership",
				"comm_log.channel_or_meeting": "Bridge",
				"comm_log.summary":            "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "comm-audience-party",
			viewSchemaID: workbook.CommLogViewSchemaID,
			fieldKey:     "comm_log.audience_party_ids",
			createBody: map[string]any{
				"comm_log.comm_type":          "briefing",
				"comm_log.audience":           "leadership",
				"comm_log.channel_or_meeting": "Bridge",
				"comm_log.summary":            "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_party_ref", "party_id": hostID.String(), "confidence": 80},
		},
		{
			name:         "comm-attendee-party",
			viewSchemaID: workbook.CommLogViewSchemaID,
			fieldKey:     "comm_log.attendee_party_ids",
			createBody: map[string]any{
				"comm_log.comm_type":          "briefing",
				"comm_log.audience":           "leadership",
				"comm_log.channel_or_meeting": "Bridge",
				"comm_log.summary":            "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_party_ref", "party_id": hostID.String(), "confidence": 80},
		},
		{
			name:         "handoff-task",
			viewSchemaID: workbook.HandoffViewSchemaID,
			fieldKey:     "handoff.open_task_ids",
			createBody: map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "handoff-decision",
			viewSchemaID: workbook.HandoffViewSchemaID,
			fieldKey:     "handoff.open_decision_ids",
			createBody: map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "handoff-risk",
			viewSchemaID: workbook.HandoffViewSchemaID,
			fieldKey:     "handoff.open_risk_refs",
			createBody: map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_risk_ref", "risk_ref_text": "Non-authoritative risk confidence", "confidence": 80},
		},
		{
			name:         "status-blocked-task",
			viewSchemaID: workbook.StatusReviewViewSchemaID,
			fieldKey:     "status_review.blocked_task_ids",
			createBody: map[string]any{
				"status_review.current_state_summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "status-pending-evidence",
			viewSchemaID: workbook.StatusReviewViewSchemaID,
			fieldKey:     "status_review.pending_evidence_ids",
			createBody: map[string]any{
				"status_review.current_state_summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "status-open-decision",
			viewSchemaID: workbook.StatusReviewViewSchemaID,
			fieldKey:     "status_review.open_decision_ids",
			createBody: map[string]any{
				"status_review.current_state_summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "lesson-follow-up-task",
			viewSchemaID: workbook.LessonViewSchemaID,
			fieldKey:     "lesson.follow_up_task_ids",
			createBody: map[string]any{
				"lesson.summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "lesson-evidence",
			viewSchemaID: workbook.LessonViewSchemaID,
			fieldKey:     "lesson.evidence_refs",
			createBody: map[string]any{
				"lesson.summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
	} {
		createBody := cloneBody(tc.createBody)
		createBody["client_txn_id"] = "txn-workbook_interaction-u-9-12-" + tc.name + "-confidence-create"
		createBody[tc.fieldKey] = map[string]any{
			"kind":    "collection_actions_v1",
			"actions": []map[string]any{tc.action},
		}
		expectWorkbookDecodeCreateRejected(t, tc.viewSchemaID, createBody)
		expectWorkbookDecodePatchRejected(t, map[string]any{
			"view_schema_id":   tc.viewSchemaID,
			"base_row_version": 1,
			"client_txn_id":    "txn-workbook_interaction-u-9-12-" + tc.name + "-confidence-patch",
			"changes": []map[string]any{
				{
					"field_key": tc.fieldKey,
					"action_payload": map[string]any{
						"kind":    "collection_actions_v1",
						"actions": []map[string]any{tc.action},
					},
				},
			},
		})
	}

	taskResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-task-null-confidence",
		Values: map[string]workbook.ValueChange{
			"task.title":     {Kind: "text", Text: textPtr("Task confidence remains null")},
			"task.task_kind": {Kind: "text", Text: textPtr("collection")},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"task.linked_record_ids": {
				Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}},
			},
		},
	}, []byte("txn-workbook_interaction-u-9-12-task-null-confidence"), "req-workbook_interaction-u-9-12-task-null-confidence", time.Date(2026, 5, 17, 18, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create task with manual link: %v", err)
	}
	decisionResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-decision-null-confidence",
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
	}, []byte("txn-workbook_interaction-u-9-12-decision-null-confidence"), "req-workbook_interaction-u-9-12-decision-null-confidence", time.Date(2026, 5, 17, 18, 45, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create decision with manual links: %v", err)
	}
	requireManualLinkConfidenceNull(t, harness, incident.ID, taskResult.RecordID, supportID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, decisionResult.RecordID, supportID, "supported_by")
	requireManualLinkConfidenceNull(t, harness, incident.ID, decisionResult.RecordID, supportID, "references_record")

	coordParty, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-coord-party",
		Values: map[string]workbook.ValueChange{
			"party.display_name": {Kind: "text", Text: textPtr("Coordination party confidence remains null")},
			"party.party_kind":   {Kind: "text", Text: textPtr("team")},
		},
	}, []byte("txn-workbook_interaction-u-9-12-coord-party"), "req-workbook_interaction-u-9-12-coord-party", time.Date(2026, 5, 17, 18, 50, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create coordination party target: %v", err)
	}
	coordTask, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-coord-task",
		Values: map[string]workbook.ValueChange{
			"task.title":     {Kind: "text", Text: textPtr("Coordination task confidence remains null")},
			"task.task_kind": {Kind: "text", Text: textPtr("follow_up")},
		},
	}, []byte("txn-workbook_interaction-u-9-12-coord-task"), "req-workbook_interaction-u-9-12-coord-task", time.Date(2026, 5, 17, 18, 51, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create coordination task target: %v", err)
	}
	coordDecision, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-coord-decision",
		Values: map[string]workbook.ValueChange{
			"decision.summary":       {Kind: "text", Text: textPtr("Coordination decision confidence remains null")},
			"decision.decision_type": {Kind: "text", Text: textPtr("containment")},
			"decision.rationale":     {Kind: "text", Text: textPtr("Coordination manual relationship confidence remains null.")},
		},
	}, []byte("txn-workbook_interaction-u-9-12-coord-decision"), "req-workbook_interaction-u-9-12-coord-decision", time.Date(2026, 5, 17, 18, 52, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create coordination decision target: %v", err)
	}
	coordEvidence, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.EvidenceViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-coord-evidence",
		Values: map[string]workbook.ValueChange{
			"evidence.title": {Kind: "text", Text: textPtr("Coordination evidence confidence remains null")},
		},
	}, []byte("txn-workbook_interaction-u-9-12-coord-evidence"), "req-workbook_interaction-u-9-12-coord-evidence", time.Date(2026, 5, 17, 18, 53, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create coordination evidence target: %v", err)
	}

	commResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.CommLogViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-comm-null-confidence",
		Values: map[string]workbook.ValueChange{
			"comm_log.comm_type":          {Kind: "text", Text: textPtr("briefing")},
			"comm_log.audience":           {Kind: "text", Text: textPtr("Coordination party")},
			"comm_log.channel_or_meeting": {Kind: "text", Text: textPtr("Bridge")},
			"comm_log.summary":            {Kind: "text", Text: textPtr("Manual confidence remains null.")},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"comm_log.decision_ids":       {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordDecision.RecordID}}},
			"comm_log.action_task_ids":    {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordTask.RecordID}}},
			"comm_log.audience_party_ids": {Actions: []workbook.CollectionAction{{Op: "add_party_ref", PartyID: &coordParty.RecordID}}},
		},
	}, []byte("txn-workbook_interaction-u-9-12-comm-null-confidence"), "req-workbook_interaction-u-9-12-comm-null-confidence", time.Date(2026, 5, 17, 18, 54, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create comm log manual links: %v", err)
	}
	requireManualLinkConfidenceNull(t, harness, incident.ID, commResult.RecordID, coordDecision.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, commResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, commResult.RecordID, coordParty.RecordID, "references_record")

	handoffResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.HandoffViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-handoff-null-confidence",
		Values: map[string]workbook.ValueChange{
			"handoff.incoming_owner_user_id": {Kind: "uuid", UUID: &actor.ID},
			"handoff.current_state_summary":  {Kind: "text", Text: textPtr("Manual confidence remains null.")},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"handoff.open_task_ids":     {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordTask.RecordID}}},
			"handoff.open_decision_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordDecision.RecordID}}},
			"handoff.open_risk_refs":    {Actions: []workbook.CollectionAction{{Op: "add_risk_ref", RiskRefText: "Manual risk refs have no confidence", NormalizedText: "manual risk refs have no confidence"}}},
		},
	}, []byte("txn-workbook_interaction-u-9-12-handoff-null-confidence"), "req-workbook_interaction-u-9-12-handoff-null-confidence", time.Date(2026, 5, 17, 18, 55, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create handoff manual links: %v", err)
	}
	requireManualLinkConfidenceNull(t, harness, incident.ID, handoffResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, handoffResult.RecordID, coordDecision.RecordID, "references_record")
	if got := queryCount(t, harness, `SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND src_record_id = $2 AND field_key = 'handoff.open_risk_refs'`, incident.ID, handoffResult.RecordID); got != 0 {
		t.Fatalf("risk refs must not persist as record links, got %d", got)
	}
	if got := queryCount(t, harness, `SELECT COUNT(*) FROM handoff_risk_refs WHERE incident_id = $1 AND handoff_record_id = $2 AND deleted_at IS NULL`, incident.ID, handoffResult.RecordID); got != 1 {
		t.Fatalf("expected one handoff risk child row, got %d", got)
	}

	statusResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.StatusReviewViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-status-null-confidence",
		Values: map[string]workbook.ValueChange{
			"status_review.current_state_summary": {Kind: "text", Text: textPtr("Manual confidence remains null.")},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"status_review.blocked_task_ids":     {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordTask.RecordID}}},
			"status_review.pending_evidence_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordEvidence.RecordID}}},
			"status_review.open_decision_ids":    {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordDecision.RecordID}}},
		},
	}, []byte("txn-workbook_interaction-u-9-12-status-null-confidence"), "req-workbook_interaction-u-9-12-status-null-confidence", time.Date(2026, 5, 17, 18, 56, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create status review manual links: %v", err)
	}
	requireManualLinkConfidenceNull(t, harness, incident.ID, statusResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, statusResult.RecordID, coordEvidence.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, statusResult.RecordID, coordDecision.RecordID, "references_record")

	lessonResult, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.LessonViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-u-9-12-lesson-null-confidence",
		Values: map[string]workbook.ValueChange{
			"lesson.summary": {Kind: "text", Text: textPtr("Manual confidence remains null.")},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"lesson.follow_up_task_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordTask.RecordID}}},
			"lesson.evidence_refs":      {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &coordEvidence.RecordID}}},
		},
	}, []byte("txn-workbook_interaction-u-9-12-lesson-null-confidence"), "req-workbook_interaction-u-9-12-lesson-null-confidence", time.Date(2026, 5, 17, 18, 57, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create lesson manual links: %v", err)
	}
	requireManualLinkConfidenceNull(t, harness, incident.ID, lessonResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, lessonResult.RecordID, coordEvidence.RecordID, "references_record")
}

func validCreateRequest(subjectRef uuid.UUID, subjectType string, state string) assessments.CreateRequest {
	return assessments.CreateRequest{
		ClientTxnID:     "txn-workbook_interaction-assessment-valid",
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

func queryCount(t testing.TB, harness *recordstoretest.StoreHarness, query string, args ...any) int {
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

func requireManualLinkConfidenceNull(t testing.TB, harness *recordstoretest.StoreHarness, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) {
	t.Helper()
	link := recordstoretest.LookupActiveLink(t, harness.DB, incidentID, sourceID, targetID, linkType)
	if link.Provenance != "manual" || link.Confidence != nil {
		t.Fatalf("manual %s link must preserve provenance=manual confidence=NULL, got %#v", linkType, link)
	}
}

func cloneBody(body map[string]any) map[string]any {
	cloned := make(map[string]any, len(body)+2)
	for key, value := range body {
		cloned[key] = value
	}
	return cloned
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
