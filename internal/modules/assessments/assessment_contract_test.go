package assessments_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentadmission "github.com/JochiRaider/cartulary/internal/modules/assessments/admission"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

func TestAssessmentsAppendOnlyStatesAndBands_Unit(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-assessments-u-9-06")
	workbookStore := appsupport.NewWorkbookCatalog(harness.DB, conflicttest.NewCodec("workbook"))
	assessmentOwner := appsupport.NewAssessmentOwner(harness.DB)
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "workbook_interaction-u906@example.test", DisplayName: "Workbook inspector U906"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-06-incident", "IR-WORKBOOK-INTERACTION-assessment-storage", "Workbook inspector assessment-storage assessments")

	hostID := uuid.New()
	identityID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Workbook inspector assessment host", "workbook_interaction-assessment-host", "", "")
	entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, identityID, "Workbook inspector assessment identity", "workbook_interaction@example.test", "workbook_interaction@example.test", "workbook_interaction")

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
		request := assessmentCreateRequest{
			ClientTxnID:     "txn-workbook_interaction-u-9-06-" + tc.key,
			SubjectRef:      &tc.subjectRef,
			SubjectType:     tc.subjectType,
			AssessmentState: tc.state,
			ConfidenceScore: tc.score,
			Rationale:       "Workbook inspector " + tc.state + " assessment rationale.",
			AssessedAt:      tc.assessedAt,
		}
		result, err := createAssessment(
			ctx,
			assessmentOwner,
			actor,
			incident.ID,
			request,
			"req-workbook_interaction-u-9-06-"+tc.key,
			time.Date(2026, 5, 17, 17, index, 0, 0, time.UTC),
		)
		if err != nil {
			t.Fatalf("create %s assessment: %v", tc.key, err)
		}
		created[tc.key] = result.RecordID
		row := result.CanonicalRow
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
	queryProvider1, ok := workbookStore.QueryFor(assessments.AssessmentsViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", assessments.AssessmentsViewSchemaID)
	}
	rowsPage, err := queryProvider1.QueryRowsPage(ctx, workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: assessments.AssessmentsViewSchemaID, Query: query, Window: querypage.Window{Limit: 100}})
	if rowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	rows := rowsPage.Rows
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
		if _, err := createAssessment(ctx, assessmentOwner, actor, incident.ID, request, "req-operational-"+state, time.Now().UTC()); err == nil {
			t.Fatalf("expected operational state %q to fail closed", state)
		}
		if after := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID); after != before {
			t.Fatalf("operational state %q left partial assessment rows: before=%d after=%d", state, before, after)
		}
	}

	for _, tc := range []struct {
		name    string
		request assessmentCreateRequest
	}{
		{name: "missing subject", request: assessmentCreateRequest{
			ClientTxnID:     "txn-workbook_interaction-u-9-06-missing-subject",
			SubjectType:     "host",
			AssessmentState: "confirmed",
			Rationale:       "Subject was not supplied.",
		}},
		{name: "support refs do not satisfy minimum semantic set", request: assessmentCreateRequest{
			ClientTxnID: "txn-workbook_interaction-u-9-06-support-only",
			SupportRefs: []uuid.UUID{
				hostID,
			},
		}},
		{name: "empty rationale", request: assessmentCreateRequest{
			ClientTxnID:     "txn-workbook_interaction-u-9-06-empty-rationale",
			SubjectRef:      &hostID,
			SubjectType:     "host",
			AssessmentState: "confirmed",
		}},
	} {
		before := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID)
		if _, err := createAssessment(ctx, assessmentOwner, actor, incident.ID, tc.request, "req-"+tc.name, time.Now().UTC()); err == nil {
			t.Fatalf("expected %s to fail closed", tc.name)
		}
		if after := queryCount(t, harness, `SELECT COUNT(*) FROM assessments WHERE incident_id = $1`, incident.ID); after != before {
			t.Fatalf("%s left partial assessment rows: before=%d after=%d", tc.name, before, after)
		}
	}

	expectAssessmentCreateRejected(t, assessmentOwner, actor, incident.ID, map[string]any{
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

	assessmentSchema, ok := viewschema.Lookup(assessments.AssessmentsViewSchemaID)
	if !ok {
		t.Fatal("assessment view schema is missing")
	}
	for fieldKey, field := range assessmentSchema.Fields() {
		if field.Writable {
			t.Fatalf("append-only assessment field %s is unexpectedly patch-writable", fieldKey)
		}
	}
	requireQueriedRecordIDs(t, workbookStore, incident.ID, filterEq("assessment.assessment_state", "confirmed"), []uuid.UUID{created["confirmed"]})

	for index, boundary := range []struct {
		score int
		band  string
	}{
		{score: 0, band: "low"},
		{score: 39, band: "low"},
		{score: 40, band: "medium"},
		{score: 69, band: "medium"},
		{score: 70, band: "high"},
		{score: 100, band: "high"},
	} {
		request := validCreateRequest(hostID, "host", "confirmed")
		request.ClientTxnID = fmt.Sprintf("txn-workbook_interaction-u-9-06-boundary-%d", boundary.score)
		request.ConfidenceScore = &boundary.score
		result, err := createAssessment(
			ctx,
			assessmentOwner,
			actor,
			incident.ID,
			request,
			fmt.Sprintf("req-workbook_interaction-u-9-06-boundary-%d", boundary.score),
			time.Date(2026, 5, 17, 18, index, 0, 0, time.UTC),
		)
		if err != nil {
			t.Fatalf("create confidence boundary %d: %v", boundary.score, err)
		}
		cells := result.CanonicalRow["cells"].(map[string]any)
		requireNumericCellValue(t, cells, "assessment.confidence_score", boundary.score)
		requireCellValue(t, cells, "assessment.confidence_band", boundary.band)
	}
}

func TestRelationshipConfidenceRejectedAndManualLinksRemainNull_Unit(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-assessments-u-9-12")
	workbookStore := appsupport.NewWorkbookCatalog(harness.DB, conflicttest.NewCodec("workbook"))
	assessmentOwner := appsupport.NewAssessmentOwner(harness.DB)
	partyOwner := appsupport.NewPartyOwner(harness.DB, conflicttest.NewCodec("workbook"))
	evidenceOwner := appsupport.NewEvidenceMutationOwner(harness.DB, conflicttest.NewCodec("workbook"))
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "workbook_interaction-u912@example.test", DisplayName: "Workbook inspector U912"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-12-incident", "IR-WORKBOOK-INTERACTION-assessment-storage", "Workbook inspector assessment-storage assessment links")
	hostID := uuid.New()
	supportID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Workbook inspector assessment support host", "workbook_interaction-assessment-support", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	request := validCreateRequest(hostID, "host", "confirmed")
	request.ClientTxnID = "txn-workbook_interaction-u-9-12-valid"
	request.SupportRefs = []uuid.UUID{supportID}
	result, err := createAssessment(
		ctx,
		assessmentOwner,
		actor,
		incident.ID,
		request,
		"req-workbook_interaction-u-9-12-valid",
		time.Date(2026, 5, 17, 18, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("create assessment with support ref: %v", err)
	}
	link := linktest.LookupActiveLink(t, harness.DB, incident.ID, result.RecordID, supportID, "supported_by")
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

	expectOwnerDecodeCreateRejected(t, tasksdecisions.TaskRequestsViewSchemaID, map[string]any{
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
	expectOwnerDecodePatchRejected(t, map[string]any{
		"view_schema_id":   tasksdecisions.TaskRequestsViewSchemaID,
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
	expectOwnerDecodeCreateRejected(t, tasksdecisions.DecisionsViewSchemaID, map[string]any{
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
	expectOwnerDecodePatchRejected(t, map[string]any{
		"view_schema_id":   tasksdecisions.DecisionsViewSchemaID,
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
	expectOwnerDecodeCreateRejected(t, tasksdecisions.DecisionsViewSchemaID, map[string]any{
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
	expectOwnerDecodePatchRejected(t, map[string]any{
		"view_schema_id":   tasksdecisions.DecisionsViewSchemaID,
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
			viewSchemaID: artifacts.CommLogViewSchemaID,
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
			viewSchemaID: artifacts.CommLogViewSchemaID,
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
			viewSchemaID: artifacts.CommLogViewSchemaID,
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
			viewSchemaID: artifacts.CommLogViewSchemaID,
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
			viewSchemaID: artifacts.HandoffViewSchemaID,
			fieldKey:     "handoff.open_task_ids",
			createBody: map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "handoff-decision",
			viewSchemaID: artifacts.HandoffViewSchemaID,
			fieldKey:     "handoff.open_decision_ids",
			createBody: map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "handoff-risk",
			viewSchemaID: artifacts.HandoffViewSchemaID,
			fieldKey:     "handoff.open_risk_refs",
			createBody: map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_risk_ref", "risk_ref_text": "Non-authoritative risk confidence", "confidence": 80},
		},
		{
			name:         "status-blocked-task",
			viewSchemaID: artifacts.StatusReviewViewSchemaID,
			fieldKey:     "status_review.blocked_task_ids",
			createBody: map[string]any{
				"status_review.current_state_summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "status-pending-evidence",
			viewSchemaID: artifacts.StatusReviewViewSchemaID,
			fieldKey:     "status_review.pending_evidence_ids",
			createBody: map[string]any{
				"status_review.current_state_summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "status-open-decision",
			viewSchemaID: artifacts.StatusReviewViewSchemaID,
			fieldKey:     "status_review.open_decision_ids",
			createBody: map[string]any{
				"status_review.current_state_summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "lesson-follow-up-task",
			viewSchemaID: artifacts.LessonViewSchemaID,
			fieldKey:     "lesson.follow_up_task_ids",
			createBody: map[string]any{
				"lesson.summary": "Coordination confidence must be rejected.",
			},
			action: map[string]any{"op": "add_record_ref", "linked_record_id": supportID.String(), "confidence": 80},
		},
		{
			name:         "lesson-evidence",
			viewSchemaID: artifacts.LessonViewSchemaID,
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
		expectOwnerDecodeCreateRejected(t, tc.viewSchemaID, createBody)
		expectOwnerDecodePatchRejected(t, map[string]any{
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

	createOutcome1, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-task-null-confidence", Now: time.Date(2026, 5, 17, 18, 30, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-task-null-confidence",
		"task.title":             "Task confidence remains null",
		"task.task_kind":         "collection",
		"task.linked_record_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": supportID}}}}))
	taskResult := workbookroutetest.RequireMutationResult(t, createOutcome1, err)
	createOutcome2, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: tasksdecisions.DecisionsViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-decision-null-confidence", Now: time.Date(2026, 5, 17, 18, 45, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-decision-null-confidence",
		"decision.summary":             "Decision confidence remains null",
		"decision.decision_type":       "containment",
		"decision.rationale":           "Manual relationship confidence remains null.",
		"decision.support_refs":        map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": supportID}}},
		"decision.affected_record_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": supportID}}}}))
	decisionResult := workbookroutetest.RequireMutationResult(t, createOutcome2, err)
	requireManualLinkConfidenceNull(t, harness, incident.ID, taskResult.RecordID, supportID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, decisionResult.RecordID, supportID, "supported_by")
	requireManualLinkConfidenceNull(t, harness, incident.ID, decisionResult.RecordID, supportID, "references_record")

	partyAdmission, apiErr := parties.AdmitCreateJSON(strings.NewReader(
		`{"client_txn_id":"txn-workbook_interaction-u-9-12-coord-party","party.display_name":"Coordination party confidence remains null","party.party_kind":"team"}`,
	))
	if apiErr != nil {
		t.Fatalf("admit coordination party target: %#v", apiErr)
	}
	coordParty, err := partyOwner.Create(ctx, parties.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incident.ID, Admission: partyAdmission,
		RequestID: "req-workbook_interaction-u-9-12-coord-party",
		Now:       time.Date(2026, 5, 17, 18, 50, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create coordination party target: %v", err)
	}
	createOutcome3, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-coord-task", Now: time.Date(2026, 5, 17, 18, 51, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-coord-task",
		"task.title":     "Coordination task confidence remains null",
		"task.task_kind": "follow_up"}))
	coordTask := workbookroutetest.RequireMutationResult(t, createOutcome3, err)
	createOutcome4, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: tasksdecisions.DecisionsViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-coord-decision", Now: time.Date(2026, 5, 17, 18, 52, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-coord-decision",
		"decision.summary":       "Coordination decision confidence remains null",
		"decision.decision_type": "containment",
		"decision.rationale":     "Coordination manual relationship confidence remains null."}))
	coordDecision := workbookroutetest.RequireMutationResult(t, createOutcome4, err)
	coordEvidenceRequest, admissionFailure := evidence.AdmitCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-workbook_interaction-u-9-12-coord-evidence",
		"evidence.title":"Coordination evidence confidence remains null"
	}`))
	if admissionFailure != nil {
		t.Fatalf("admit coordination evidence target: %#v", admissionFailure)
	}
	coordEvidence, err := evidenceOwner.Create(ctx, evidence.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incident.ID, Admission: coordEvidenceRequest,
		RequestID: "req-workbook_interaction-u-9-12-coord-evidence",
		Now:       time.Date(2026, 5, 17, 18, 53, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create coordination evidence target: %v", err)
	}

	createOutcome5, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: artifacts.CommLogViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-comm-null-confidence", Now: time.Date(2026, 5, 17, 18, 54, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-comm-null-confidence",
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "Coordination party",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Manual confidence remains null.",
		"comm_log.decision_ids":       map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordDecision.RecordID}}},
		"comm_log.action_task_ids":    map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordTask.RecordID}}},
		"comm_log.audience_party_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_party_ref", "party_id": coordParty.RecordID}}}}))
	commResult := workbookroutetest.RequireMutationResult(t, createOutcome5, err)
	requireManualLinkConfidenceNull(t, harness, incident.ID, commResult.RecordID, coordDecision.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, commResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, commResult.RecordID, coordParty.RecordID, "references_record")

	createOutcome6, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: artifacts.HandoffViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-handoff-null-confidence", Now: time.Date(2026, 5, 17, 18, 55, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-handoff-null-confidence",
		"handoff.incoming_owner_user_id": actor.ID,
		"handoff.current_state_summary":  "Manual confidence remains null.",
		"handoff.open_task_ids":          map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordTask.RecordID}}},
		"handoff.open_decision_ids":      map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordDecision.RecordID}}},
		"handoff.open_risk_refs":         map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_risk_ref", "risk_ref_text": "Manual risk refs have no confidence"}}}}))
	handoffResult := workbookroutetest.RequireMutationResult(t, createOutcome6, err)
	requireManualLinkConfidenceNull(t, harness, incident.ID, handoffResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, handoffResult.RecordID, coordDecision.RecordID, "references_record")
	if got := queryCount(t, harness, `SELECT COUNT(*) FROM record_links WHERE incident_id = $1 AND src_record_id = $2 AND field_key = 'handoff.open_risk_refs'`, incident.ID, handoffResult.RecordID); got != 0 {
		t.Fatalf("risk refs must not persist as record links, got %d", got)
	}
	if got := queryCount(t, harness, `SELECT COUNT(*) FROM handoff_risk_refs WHERE incident_id = $1 AND handoff_record_id = $2 AND deleted_at IS NULL`, incident.ID, handoffResult.RecordID); got != 1 {
		t.Fatalf("expected one handoff risk child row, got %d", got)
	}

	createOutcome7, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: artifacts.StatusReviewViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-status-null-confidence", Now: time.Date(2026, 5, 17, 18, 56, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-status-null-confidence",
		"status_review.current_state_summary": "Manual confidence remains null.",
		"status_review.blocked_task_ids":      map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordTask.RecordID}}},
		"status_review.pending_evidence_ids":  map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordEvidence.RecordID}}},
		"status_review.open_decision_ids":     map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordDecision.RecordID}}}}))
	statusResult := workbookroutetest.RequireMutationResult(t, createOutcome7, err)
	requireManualLinkConfidenceNull(t, harness, incident.ID, statusResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, statusResult.RecordID, coordEvidence.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, statusResult.RecordID, coordDecision.RecordID, "references_record")

	createOutcome8, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: artifacts.LessonViewSchemaID, RequestID: "req-workbook_interaction-u-9-12-lesson-null-confidence", Now: time.Date(2026, 5, 17, 18, 57, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-u-9-12-lesson-null-confidence",
		"lesson.summary":            "Manual confidence remains null.",
		"lesson.follow_up_task_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordTask.RecordID}}},
		"lesson.evidence_refs":      map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": coordEvidence.RecordID}}}}))
	lessonResult := workbookroutetest.RequireMutationResult(t, createOutcome8, err)
	requireManualLinkConfidenceNull(t, harness, incident.ID, lessonResult.RecordID, coordTask.RecordID, "references_record")
	requireManualLinkConfidenceNull(t, harness, incident.ID, lessonResult.RecordID, coordEvidence.RecordID, "references_record")
}

type assessmentCreateRequest struct {
	ClientTxnID     string
	SubjectRef      *uuid.UUID
	SubjectType     string
	AssessmentState string
	ConfidenceScore *int
	Rationale       string
	Assessor        *uuid.UUID
	AssessedAt      *time.Time
	SupportRefs     []uuid.UUID
}

func validCreateRequest(
	subjectRef uuid.UUID,
	subjectType string,
	state string,
) assessmentCreateRequest {
	return assessmentCreateRequest{
		ClientTxnID:     "txn-workbook_interaction-assessment-valid",
		SubjectRef:      &subjectRef,
		SubjectType:     subjectType,
		AssessmentState: state,
		Rationale:       "Valid rationale.",
	}
}

func createAssessment(
	ctx context.Context,
	owner *assessments.Facade,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request assessmentCreateRequest,
	requestID string,
	now time.Time,
) (assessments.CreateResult, error) {
	input := assessments.CreateInput{
		ClientTxnID:     request.ClientTxnID,
		SubjectType:     request.SubjectType,
		AssessmentState: request.AssessmentState,
		ConfidenceScore: request.ConfidenceScore,
		Rationale:       request.Rationale,
		Assessor:        request.Assessor,
		AssessedAt:      request.AssessedAt,
		SupportRefs:     append([]uuid.UUID(nil), request.SupportRefs...),
	}
	if request.SubjectRef != nil {
		input.SubjectRef = *request.SubjectRef
	}
	return owner.Create(ctx, assessments.CreateCommand{
		ActorUserID: actor.ID,
		IncidentID:  incidentID,
		Input:       input,
		RequestID:   requestID,
		Now:         now,
	})
}

func filterEq(fieldKey string, value any) viewschema.Filter {
	return viewschema.Filter{FieldKey: fieldKey, Op: "eq", Arg: map[string]any{"value": value}}
}

func requireNumericCellValue(t testing.TB, cells map[string]any, fieldKey string, want int) {
	t.Helper()
	value := cells[fieldKey].(map[string]any)["value"]
	var got int
	switch typed := value.(type) {
	case int:
		got = typed
	case int32:
		got = int(typed)
	case int64:
		got = int(typed)
	case float64:
		got = int(typed)
	default:
		t.Fatalf("unexpected %s numeric type: %T", fieldKey, value)
	}
	if got != want {
		t.Fatalf("unexpected %s value: got %d want %d", fieldKey, got, want)
	}
}

func requireQueriedRecordIDs(t testing.TB, store *workbook.WorkbookContributionCatalog, incidentID uuid.UUID, filter viewschema.Filter, want []uuid.UUID) {
	t.Helper()
	query := assessmentQueryMeta(t)
	query.Filters = []viewschema.Filter{filter}
	queryProvider2, ok := store.QueryFor(assessments.AssessmentsViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", assessments.AssessmentsViewSchemaID)
	}
	rowsPage, err := queryProvider2.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incidentID, ViewSchemaID: assessments.AssessmentsViewSchemaID, Query: query, Window: querypage.Window{Limit: 100}})
	if rowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	rows := rowsPage.Rows
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
	schema, ok := viewschema.Lookup(assessments.AssessmentsViewSchemaID)
	if !ok {
		t.Fatalf("missing assessments view schema")
	}
	return schema.DefaultQueryMeta()
}

func queryCount(t testing.TB, harness *appsupport.StoreHarness, query string, args ...any) int {
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
	if _, apiErr := assessmentadmission.DecodeCreateRequest(strings.NewReader(string(data))); apiErr == nil {
		t.Fatalf("expected create body to be rejected: %#v", body)
	}
}

func expectAssessmentCreateRejected(
	t testing.TB,
	owner *assessments.Facade,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	body map[string]any,
) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal assessment create body: %v", err)
	}
	input, apiErr := assessmentadmission.DecodeCreateRequest(strings.NewReader(string(data)))
	if apiErr != nil {
		return
	}
	if _, err := owner.Create(context.Background(), assessments.CreateCommand{
		ActorUserID: actor.ID,
		IncidentID:  incidentID,
		Input:       input,
		RequestID:   "req-" + input.ClientTxnID,
		Now:         time.Now().UTC(),
	}); err == nil {
		t.Fatalf("expected assessment create body to be rejected: %#v", body)
	}
}

func expectOwnerDecodeCreateRejected(t testing.TB, viewSchemaID string, body map[string]any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal workbook create body: %v", err)
	}
	if viewSchemaID == tasksdecisions.TaskRequestsViewSchemaID || viewSchemaID == tasksdecisions.DecisionsViewSchemaID {
		_, admissionFailure := tasksdecisions.AdmitCreateJSON(viewSchemaID, strings.NewReader(string(data)))
		if admissionFailure == nil {
			t.Fatalf("expected Tasks/Decisions mutation body to be rejected: %#v", body)
		}
		return
	}
	if _, admissionErr := artifacts.AdmitCreate(viewSchemaID, strings.NewReader(string(data))); admissionErr == nil {
		t.Fatalf("expected owner mutation body to be rejected: %#v", body)
	}
}

func expectOwnerDecodePatchRejected(t testing.TB, body map[string]any) {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal workbook patch body: %v", err)
	}
	viewSchemaID, _ := body["view_schema_id"].(string)
	if viewSchemaID == tasksdecisions.TaskRequestsViewSchemaID || viewSchemaID == tasksdecisions.DecisionsViewSchemaID {
		_, admissionFailure := tasksdecisions.AdmitPatchJSON(strings.NewReader(string(data)))
		if admissionFailure == nil {
			t.Fatalf("expected Tasks/Decisions mutation body to be rejected: %#v", body)
		}
		return
	}
	if _, admissionErr := artifacts.AdmitPatch(strings.NewReader(string(data))); admissionErr == nil {
		t.Fatalf("expected owner mutation body to be rejected: %#v", body)
	}
}

func requireManualLinkConfidenceNull(t testing.TB, harness *appsupport.StoreHarness, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) {
	t.Helper()
	link := linktest.LookupActiveLink(t, harness.DB, incidentID, sourceID, targetID, linkType)
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

func intPtr(value int) *int {
	return &value
}

func timePtr(value time.Time) *time.Time {
	return &value
}
