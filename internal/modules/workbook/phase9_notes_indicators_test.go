package workbook_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	phase4storetest "github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
)

func TestPhase9_NotesAreArtifactBackedRows_U_9_03(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-u-9-03-notes")
	store := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u903@example.test", "U903 Notes", "U903NotesPass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-03-incident", "IR-U903", "Phase 9 U-9-03")
	sourceRecordID := uuid.New()
	phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceRecordID)

	created, err := store.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.NotesViewSchemaID,
		ClientTxnID:  "txn-phase9-u-9-03-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Phase 9 artifact note"),
			"note.body":  textChange("Artifact-backed note body"),
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"note.tags": {
				Actions: []workbook.CollectionAction{{Op: "add_tag", RawText: "phase9-sprint3", NormalizedText: "phase9-sprint3"}},
			},
		},
	}, []byte("txn-phase9-u-9-03-note"), "req-phase9-u-9-03-note", time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create note through workbook store: %v", err)
	}

	requireScalarCount(t, harness, `
SELECT count(*)
  FROM records r
  JOIN artifacts a ON a.incident_id = r.incident_id AND a.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type = 'note'
`, created.RecordID, 1)
	requireScalarCount(t, harness, `SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'notes'`, 0)
	requireScalarCount(t, harness, `SELECT count(*) FROM record_tags WHERE incident_id = $1 AND record_id = $2 AND normalized_tag_name = 'phase9-sprint3' AND deleted_at IS NULL`, incident.ID, created.RecordID, 1)
	requireScalarCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1 AND row_version = 1`, created.RecordID, 1)

	linked, err := store.CreateLinkedNote(context.Background(), actor, sourceRecordID, workbook.LinkedNoteCreateRequest{
		ClientTxnID: "txn-phase9-u-9-03-linked-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Phase 9 linked note"),
			"note.body":  textChange("Linked through references_artifact"),
		},
	}, workbook.LinkedNoteCreateRequestHash(sourceRecordID, workbook.LinkedNoteCreateRequest{
		ClientTxnID: "txn-phase9-u-9-03-linked-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Phase 9 linked note"),
			"note.body":  textChange("Linked through references_artifact"),
		},
	}), "req-phase9-u-9-03-linked-note", time.Date(2026, 5, 17, 15, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create linked note: %v", err)
	}
	requireScalarCount(t, harness, `
SELECT count(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'references_artifact'
   AND field_key IS NULL
   AND deleted_at IS NULL
`, incident.ID, sourceRecordID, linked.RecordID, 1)

	rows, err := store.QueryRows(context.Background(), incident.ID, workbook.NotesViewSchemaID, mustQueryMeta(t, workbook.NotesViewSchemaID))
	if err != nil {
		t.Fatalf("query notes rows: %v", err)
	}
	linkedRow := requireQueriedRow(t, rows, linked.RecordID)
	if got := linkedRow["cells"].(map[string]any)["note.linked_record_count"].(map[string]any)["value"]; got != float64(1) && got != int32(1) && got != int64(1) && got != 1 {
		t.Fatalf("expected linked note count to include incoming contextual link, got %#v", got)
	}
}

func TestPhase9_NotesAndIndicatorsQueryThroughWorkbookProjections_I_9_02(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-i-9-02-notes-indicators")
	workbookStore := workbook.NewStore(harness.DB)
	entityStore := entities.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "i902@example.test", "I902 Projection", "I902ProjectionPass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-i-9-02-incident", "IR-I902", "Phase 9 I-9-02")

	note, err := workbookStore.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.NotesViewSchemaID,
		ClientTxnID:  "txn-phase9-i-9-02-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Projection-backed note"),
			"note.body":  textChange("query token phase9-note-projection"),
		},
	}, []byte("txn-phase9-i-9-02-note"), "req-phase9-i-9-02-note", time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection note: %v", err)
	}
	noteRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.NotesViewSchemaID, mustQueryMeta(t, workbook.NotesViewSchemaID))
	if err != nil {
		t.Fatalf("query notes projection: %v", err)
	}
	requireQueriedRow(t, noteRows, note.RecordID)

	indicator, err := entityStore.CreateIndicatorRow(context.Background(), actor, incident.ID, entities.CreateRequest{
		ClientTxnID: "txn-phase9-i-9-02-indicator",
		Values: map[string]string{
			"indicator.indicator_type": "ipv4_addr",
			"indicator.value_kind":     "atomic",
			"indicator.display_value":  "203.0.113.45",
		},
	}, []byte("txn-phase9-i-9-02-indicator"), "req-phase9-i-9-02-indicator", time.Date(2026, 5, 17, 16, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection indicator: %v", err)
	}
	indicatorRows, err := workbookStore.QueryRows(context.Background(), incident.ID, entities.IndicatorsViewSchemaID, mustQueryMeta(t, entities.IndicatorsViewSchemaID))
	if err != nil {
		t.Fatalf("query indicators through workbook store: %v", err)
	}
	requireQueriedRow(t, indicatorRows, indicator.RecordID)
}

func TestPhase9_AssessmentsQueryThroughWorkbookProjections_I_9_02(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-i-9-02-assessments")
	workbookStore := workbook.NewStore(harness.DB)
	assessmentStore := assessments.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "i902-assessments@example.test", "I902 Assessments", "I902AssessmentsPass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-i-9-02-assessment-incident", "IR-I902-ASSESS", "Phase 9 I-9-02 assessments")

	hostID := uuid.New()
	supportID := uuid.New()
	phase4storetest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Phase 9 projection assessment host", "phase9-projection-assessment", "", "")
	phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	confidenceScore := 85
	request := assessments.CreateRequest{
		ClientTxnID:     "txn-phase9-i-9-02-assessment",
		SubjectRef:      &hostID,
		SubjectType:     "host",
		AssessmentState: "confirmed",
		ConfidenceScore: &confidenceScore,
		Rationale:       "Projection-backed assessment row.",
		SupportRefs:     []uuid.UUID{supportID},
	}
	created, err := assessmentStore.CreateAssessmentRow(
		context.Background(),
		actor,
		incident.ID,
		request,
		assessments.CreateRequestHash(request),
		"req-phase9-i-9-02-assessment",
		time.Date(2026, 5, 17, 16, 10, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("create projection assessment: %v", err)
	}

	query := mustQueryMeta(t, workbook.AssessmentsViewSchemaID)
	query.Filters = []viewschema.Filter{
		{FieldKey: "assessment.confidence_band", Op: "eq", Arg: map[string]any{"value": "high"}},
	}
	rows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.AssessmentsViewSchemaID, query)
	if err != nil {
		t.Fatalf("query assessments through workbook projection: %v", err)
	}
	row := requireQueriedRow(t, rows, created.RecordID)
	cells := row["cells"].(map[string]any)
	if got := cells["assessment.confidence_band"].(map[string]any)["value"]; got != "high" {
		t.Fatalf("expected projected confidence band high, got %#v", got)
	}
	if got := cells["assessment.supporting_link_count"].(map[string]any)["value"]; got != int64(1) && got != int32(1) && got != 1 {
		t.Fatalf("expected projected supporting link count, got %#v", got)
	}
}

func TestPhase9_TaskRequestsAndDecisionsQueryThroughWorkbookProjections_I_9_02(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-i-9-02-tasks-decisions")
	workbookStore := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "i902-tasks-decisions@example.test", "I902 Tasks Decisions", "I902TasksDecisions1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-i-9-02-task-decision-incident", "IR-I902-TD", "Phase 9 I-9-02 tasks decisions")

	supportID := mustCreateEvidenceForU911(t, workbookStore, actor, incident.ID, "txn-phase9-i-9-02-decision-support", "I-9-02 decision support")
	affectedID := mustCreateEvidenceForU911(t, workbookStore, actor, incident.ID, "txn-phase9-i-9-02-decision-affected", "I-9-02 affected record")
	decision, err := workbookStore.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  "txn-phase9-i-9-02-decision",
		Values: map[string]workbook.ValueChange{
			"decision.summary":       textChange("Projection-backed decision"),
			"decision.decision_type": textChange("containment"),
			"decision.rationale":     textChange("Decision projection includes support and affected links."),
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"decision.support_refs":        {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}}},
			"decision.affected_record_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &affectedID}}},
		},
	}, []byte("txn-phase9-i-9-02-decision"), "req-phase9-i-9-02-decision", time.Date(2026, 5, 17, 16, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection decision: %v", err)
	}

	decisionQuery := mustQueryMeta(t, workbook.DecisionsViewSchemaID)
	decisionQuery.Filters = []viewschema.Filter{{FieldKey: "decision.status", Op: "eq", Arg: map[string]any{"value": "proposed"}}}
	decisionRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.DecisionsViewSchemaID, decisionQuery)
	if err != nil {
		t.Fatalf("query decisions through workbook projection: %v", err)
	}
	decisionRow := requireQueriedRow(t, decisionRows, decision.RecordID)
	requireProjectedCollectionCount(t, decisionRow, "decision.support_refs", 1)
	requireProjectedCollectionCount(t, decisionRow, "decision.affected_record_ids", 1)
	requireProjectedNumericCell(t, decisionRow, "decision.affected_record_count", 1)

	dueAt := time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC)
	decisionID := decision.RecordID
	task, err := workbookStore.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  "txn-phase9-i-9-02-task",
		Values: map[string]workbook.ValueChange{
			"task.title":              textChange("Projection-backed task"),
			"task.task_kind":          textChange("collection"),
			"task.due_at":             {Kind: "timestamp", Timestamp: &dueAt},
			"task.decision_record_id": {Kind: "uuid", UUID: &decisionID},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"task.linked_record_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}}},
		},
	}, []byte("txn-phase9-i-9-02-task"), "req-phase9-i-9-02-task", time.Date(2026, 5, 17, 16, 20, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection task: %v", err)
	}

	taskQuery := mustQueryMeta(t, workbook.TaskRequestsViewSchemaID)
	taskQuery.Filters = []viewschema.Filter{
		{FieldKey: "task.status", Op: "eq", Arg: map[string]any{"value": "open"}},
		{FieldKey: "task.owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
		{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": dueAt.Format(time.RFC3339Nano)}},
	}
	taskRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.TaskRequestsViewSchemaID, taskQuery)
	if err != nil {
		t.Fatalf("query task requests through workbook projection: %v", err)
	}
	taskRow := requireQueriedRow(t, taskRows, task.RecordID)
	cells := taskRow["cells"].(map[string]any)
	if got := cells["task.decision_record_id"].(map[string]any)["value"]; got != decision.RecordID.String() {
		t.Fatalf("expected projected task decision ref, got %#v", got)
	}
	requireProjectedCollectionCount(t, taskRow, "task.linked_record_ids", 1)
	requireProjectedNumericCell(t, taskRow, "task.linked_record_count", 1)
}

func TestPhase9_CoordinationSurfacesQueryThroughWorkbookProjections_I_9_02(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-i-9-02-coordination")
	workbookStore := workbook.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "i902-coordination@example.test", "I902 Coordination", "I902Coordination1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-i-9-02-coordination-incident", "IR-I902-COORD", "Phase 9 I-9-02 coordination")

	partyID := mustCreatePartyForU911(t, workbookStore, actor, incident.ID, "txn-phase9-i-9-02-coordination-party", "Projection coordination party")
	taskID := mustCreateTaskForU911(t, workbookStore, actor, incident.ID, "txn-phase9-i-9-02-coordination-task", "Projection coordination task")
	evidenceID := mustCreateEvidenceForU911(t, workbookStore, actor, incident.ID, "txn-phase9-i-9-02-coordination-evidence", "Projection coordination evidence")
	decisionID := mustCreateSprint6Decision(t, workbookStore, actor, incident.ID, "txn-phase9-i-9-02-coordination-decision", "approved", "Projection coordination decision")

	nextReport := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	comm := mustCreateSprint7Row(t, workbookStore, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-phase9-i-9-02-comm", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      sprint7Timestamp(time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          textChange("briefing"),
		"comm_log.audience":           textChange("Projection leadership"),
		"comm_log.channel_or_meeting": textChange("Bridge"),
		"comm_log.summary":            textChange("Projection-backed comm log"),
		"comm_log.next_report_at":     sprint7Timestamp(nextReport),
	}, map[string]workbook.CollectionActionPayload{
		"comm_log.decision_ids":       {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &decisionID}}},
		"comm_log.action_task_ids":    {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"comm_log.audience_party_ids": {Actions: []workbook.CollectionAction{{Op: "add_party_ref", PartyID: &partyID}}},
	}, time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC))
	commQuery := mustQueryMeta(t, workbook.CommLogViewSchemaID)
	commQuery.Filters = []viewschema.Filter{
		{FieldKey: "comm_log.comm_type", Op: "eq", Arg: map[string]any{"value": "briefing"}},
		{FieldKey: "comm_log.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-24"}},
	}
	commRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.CommLogViewSchemaID, commQuery)
	if err != nil {
		t.Fatalf("query comm log projection: %v", err)
	}
	commRow := requireQueriedRow(t, commRows, comm.RecordID)
	requireProjectedCollectionCount(t, commRow, "comm_log.decision_ids", 1)
	requireProjectedCollectionCount(t, commRow, "comm_log.action_task_ids", 1)
	requireProjectedCollectionCount(t, commRow, "comm_log.audience_party_ids", 1)

	acknowledgedAt := time.Date(2026, 5, 22, 14, 0, 0, 0, time.UTC)
	handoff := mustCreateSprint7Row(t, workbookStore, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-phase9-i-9-02-handoff", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          sprint7Timestamp(time.Date(2026, 5, 22, 13, 0, 0, 0, time.UTC)),
		"handoff.incoming_owner_user_id": sprint7UUID(actor.ID),
		"handoff.current_state_summary":  textChange("Projection-backed handoff"),
		"handoff.acknowledged_at":        sprint7Timestamp(acknowledgedAt),
	}, map[string]workbook.CollectionActionPayload{
		"handoff.open_task_ids":     {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"handoff.open_decision_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &decisionID}}},
		"handoff.open_risk_refs":    {Actions: []workbook.CollectionAction{{Op: "add_risk_ref", RiskRefText: "Projection-backed handoff risk", NormalizedText: "projection-backed handoff risk"}}},
	}, time.Date(2026, 5, 22, 13, 0, 0, 0, time.UTC))
	handoffQuery := mustQueryMeta(t, workbook.HandoffViewSchemaID)
	handoffQuery.Filters = []viewschema.Filter{{FieldKey: "handoff.ack_state", Op: "eq", Arg: map[string]any{"value": "acknowledged"}}}
	handoffRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.HandoffViewSchemaID, handoffQuery)
	if err != nil {
		t.Fatalf("query handoff projection: %v", err)
	}
	handoffRow := requireQueriedRow(t, handoffRows, handoff.RecordID)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_task_ids", 1)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_decision_ids", 1)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_risk_refs", 1)

	statusNextReport := time.Date(2026, 5, 25, 15, 0, 0, 0, time.UTC)
	status := mustCreateSprint7Row(t, workbookStore, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-phase9-i-9-02-status", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         sprint7Timestamp(time.Date(2026, 5, 23, 15, 0, 0, 0, time.UTC)),
		"status_review.current_state_summary": textChange("Projection-backed status review"),
		"status_review.next_report_at":        sprint7Timestamp(statusNextReport),
	}, map[string]workbook.CollectionActionPayload{
		"status_review.blocked_task_ids":     {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"status_review.pending_evidence_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidenceID}}},
		"status_review.open_decision_ids":    {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &decisionID}}},
	}, time.Date(2026, 5, 23, 15, 0, 0, 0, time.UTC))
	statusQuery := mustQueryMeta(t, workbook.StatusReviewViewSchemaID)
	statusQuery.Filters = []viewschema.Filter{{FieldKey: "status_review.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-25"}}}
	statusRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.StatusReviewViewSchemaID, statusQuery)
	if err != nil {
		t.Fatalf("query status review projection: %v", err)
	}
	statusRow := requireQueriedRow(t, statusRows, status.RecordID)
	requireProjectedCollectionCount(t, statusRow, "status_review.blocked_task_ids", 1)
	requireProjectedCollectionCount(t, statusRow, "status_review.pending_evidence_ids", 1)
	requireProjectedCollectionCount(t, statusRow, "status_review.open_decision_ids", 1)

	lesson := mustCreateSprint7Row(t, workbookStore, actor, incident.ID, workbook.LessonViewSchemaID, "txn-phase9-i-9-02-lesson", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": sprint7Timestamp(time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC)),
		"lesson.summary":       textChange("Projection-backed lesson"),
		"lesson.closure_state": textChange("closed"),
	}, map[string]workbook.CollectionActionPayload{
		"lesson.follow_up_task_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"lesson.evidence_refs":      {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidenceID}}},
	}, time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC))
	lessonQuery := mustQueryMeta(t, workbook.LessonViewSchemaID)
	lessonQuery.Filters = []viewschema.Filter{{FieldKey: "lesson.closure_state", Op: "eq", Arg: map[string]any{"value": "closed"}}}
	lessonRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.LessonViewSchemaID, lessonQuery)
	if err != nil {
		t.Fatalf("query lesson projection: %v", err)
	}
	lessonRow := requireQueriedRow(t, lessonRows, lesson.RecordID)
	requireProjectedCollectionCount(t, lessonRow, "lesson.follow_up_task_ids", 1)
	requireProjectedCollectionCount(t, lessonRow, "lesson.evidence_refs", 1)
}

func textChange(value string) workbook.ValueChange {
	return workbook.ValueChange{Kind: "text", Text: &value}
}

func mustQueryMeta(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("missing view schema %s", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

func requireQueriedRow(t testing.TB, rows []map[string]any, recordID uuid.UUID) map[string]any {
	t.Helper()
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("missing queried row %s in %#v", recordID, rows)
	return nil
}

func requireScalarCount(t testing.TB, harness *phase4storetest.StoreHarness, query string, args ...any) {
	t.Helper()
	want := args[len(args)-1].(int)
	args = args[:len(args)-1]
	var got int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query scalar count: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}

func requireProjectedCollectionCount(t testing.TB, row map[string]any, fieldKey string, want int) {
	t.Helper()
	value := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"].(map[string]any)
	if value["kind"] != "collection_value_v1" {
		t.Fatalf("expected %s collection value, got %#v", fieldKey, value)
	}
	var got int
	switch items := value["items"].(type) {
	case []any:
		got = len(items)
	case []map[string]any:
		got = len(items)
	default:
		t.Fatalf("unexpected %s items shape: %#v", fieldKey, value["items"])
	}
	if got != want {
		t.Fatalf("unexpected %s item count: got %d want %d items=%#v", fieldKey, got, want, value["items"])
	}
}

func requireProjectedNumericCell(t testing.TB, row map[string]any, fieldKey string, want int64) {
	t.Helper()
	got := row["cells"].(map[string]any)[fieldKey].(map[string]any)["value"]
	switch value := got.(type) {
	case int:
		if int64(value) == want {
			return
		}
	case int32:
		if int64(value) == want {
			return
		}
	case int64:
		if value == want {
			return
		}
	case float64:
		if int64(value) == want {
			return
		}
	}
	t.Fatalf("unexpected %s value: got %#v want %d", fieldKey, got, want)
}
