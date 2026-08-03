package workbook_test

import (
	"context"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestLinkedNotesCreateContextualArtifactLinks_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-u-9-03-notes")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle := timelineassembly.NewBundle(harness.DB, workbookTestConflictTokens(), appender, revisionComposition.Intents, evidence.NewTimelineAttachmentContribution(harness.DB))
	store := newCatalogBackedWorkbookStore(t, harness.DB, timelineBundle, appender, revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u903@example.test", "U903 Notes", "U903NotesPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-03-incident", "IR-U903", "Workbook inspector workbook-storage")
	sourceRecordID := uuid.New()
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceRecordID)

	linked, err := store.CreateLinkedNote(context.Background(), actor, sourceRecordID, workbook.LinkedNoteCreateRequest{
		ClientTxnID: "txn-workbook_interaction-u-9-03-linked-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Workbook inspector linked note"),
			"note.body":  textChange("Linked through references_artifact"),
		},
	}, workbook.LinkedNoteCreateRequestHash(sourceRecordID, workbook.LinkedNoteCreateRequest{
		ClientTxnID: "txn-workbook_interaction-u-9-03-linked-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Workbook inspector linked note"),
			"note.body":  textChange("Linked through references_artifact"),
		},
	}), "req-workbook_interaction-u-9-03-linked-note", time.Date(2026, 5, 17, 15, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create linked note: %v", err)
	}
	requireScalarCount(t, harness, `
SELECT count(*)
  FROM records r
  JOIN artifacts a ON a.incident_id = r.incident_id AND a.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type = 'note'
`, linked.RecordID, 1)
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
	var linkedRecordLinkID string
	if err := harness.DB.QueryRow(context.Background(), `
SELECT record_link_id::text
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'references_artifact'
   AND field_key IS NULL
   AND deleted_at IS NULL
`, incident.ID, sourceRecordID, linked.RecordID).Scan(&linkedRecordLinkID); err != nil {
		t.Fatalf("query linked note record_link_id: %v", err)
	}
	requireScalarCount(t, harness, `
SELECT count(*)
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND sequence_no = 2
   AND target_kind = 'record_link'
   AND target_id = $2
   AND operation_kind = 'create'
   AND after_value ->> 'record_link_id' = $2
   AND after_value ->> 'src_record_id' = $3
   AND after_value ->> 'dst_record_id' = $4
   AND after_value ->> 'link_type' = 'references_artifact'
`, linked.ChangeSetID, linkedRecordLinkID, sourceRecordID.String(), linked.RecordID.String(), 1)
	requireScalarCount(t, harness, `
SELECT count(*)
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND target_kind = 'record_link'
   AND target_id = $2
`, linked.ChangeSetID, sourceRecordID.String()+":references_artifact:"+linked.RecordID.String(), 0)

	rows, err := store.QueryRows(context.Background(), incident.ID, workbook.NotesViewSchemaID, mustQueryMeta(t, workbook.NotesViewSchemaID))
	if err != nil {
		t.Fatalf("query notes rows: %v", err)
	}
	linkedRow := requireQueriedRow(t, rows, linked.RecordID)
	if got := linkedRow["cells"].(map[string]any)["note.linked_record_count"].(map[string]any)["value"]; got != float64(1) && got != int32(1) && got != int64(1) && got != 1 {
		t.Fatalf("expected linked note count to include incoming contextual link, got %#v", got)
	}
}

func TestNotesAndIndicatorsQueryThroughWorkbookProjections_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-i-9-02-notes-indicators")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle := timelineassembly.NewBundle(harness.DB, workbookTestConflictTokens(), appender, revisionComposition.Intents, evidence.NewTimelineAttachmentContribution(harness.DB))
	workbookStore := newCatalogBackedWorkbookStore(t, harness.DB, timelineBundle, appender, revisionComposition.Intents)
	indicatorStore, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    harness.DB,
		Revisions:   appender,
		Projections: timelineBundle.ProjectionCoordinator,
	})
	if err != nil {
		t.Fatalf("compose Indicator test owner: %v", err)
	}
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902@example.test", "I902 Projection", "I902ProjectionPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-incident", "IR-I902", "Workbook inspector workbook-interaction")

	note, err := workbookStore.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.NotesViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Projection-backed note"),
			"note.body":  textChange("query token workbook_interaction-note-projection"),
		},
	}, []byte("txn-workbook_interaction-i-9-02-note"), "req-workbook_interaction-i-9-02-note", time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection note: %v", err)
	}
	noteRows, err := workbookStore.QueryRows(context.Background(), incident.ID, workbook.NotesViewSchemaID, mustQueryMeta(t, workbook.NotesViewSchemaID))
	if err != nil {
		t.Fatalf("query notes projection: %v", err)
	}
	requireQueriedRow(t, noteRows, note.RecordID)

	indicator, err := indicatorStore.CreateIndicatorRow(context.Background(), actor, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-workbook_interaction-i-9-02-indicator",
		IndicatorType: "ipv4_addr",
		ValueKind:     "atomic",
		DisplayValue:  "203.0.113.45",
	}, []byte("txn-workbook_interaction-i-9-02-indicator"), "req-workbook_interaction-i-9-02-indicator", time.Date(2026, 5, 17, 16, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection indicator: %v", err)
	}
	indicatorRows, err := workbookStore.QueryRows(context.Background(), incident.ID, indicators.ViewSchemaID, mustQueryMeta(t, indicators.ViewSchemaID))
	if err != nil {
		t.Fatalf("query indicators through workbook store: %v", err)
	}
	requireQueriedRow(t, indicatorRows, indicator.RecordID)
}

func TestAssessmentsQueryThroughWorkbookProjections_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-i-9-02-assessments")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle := timelineassembly.NewBundle(harness.DB, workbookTestConflictTokens(), appender, revisionComposition.Intents, evidence.NewTimelineAttachmentContribution(harness.DB))
	workbookStore := newCatalogBackedWorkbookStore(t, harness.DB, timelineBundle, appender, revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902-assessments@example.test", "I902 Assessments", "I902AssessmentsPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-assessment-incident", "IR-I902-ASSESS", "Workbook inspector workbook-interaction assessments")

	hostID := uuid.New()
	supportID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Workbook inspector projection assessment host", "workbook_interaction-projection-assessment", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	confidenceScore := int64(85)
	subjectType := "host"
	assessmentState := "confirmed"
	rationale := "Projection-backed assessment row."
	request := workbook.CreateRequest{
		ViewSchemaID: assessments.AssessmentsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-assessment",
		Values: map[string]workbook.ValueChange{
			"assessment.subject_ref": {
				Kind: "uuid",
				UUID: &hostID,
			},
			"assessment.subject_type": {
				Kind: "text",
				Text: &subjectType,
			},
			"assessment.assessment_state": {
				Kind: "text",
				Text: &assessmentState,
			},
			"assessment.confidence_score": {
				Kind:   "number",
				Number: &confidenceScore,
			},
			"assessment.rationale": {
				Kind: "text",
				Text: &rationale,
			},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"assessment.support_refs": {
				Actions: []workbook.CollectionAction{{
					Op:             "add_record_ref",
					LinkedRecordID: &supportID,
				}},
			},
		},
	}
	created, err := workbookStore.CreateWorkbookRow(
		context.Background(),
		actor,
		incident.ID,
		request,
		workbook.CreateRequestHash(request),
		"req-workbook_interaction-i-9-02-assessment",
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

func TestTaskRequestsAndDecisionsQueryThroughWorkbookProjections_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-i-9-02-tasks-decisions")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle := timelineassembly.NewBundle(harness.DB, workbookTestConflictTokens(), appender, revisionComposition.Intents, evidence.NewTimelineAttachmentContribution(harness.DB))
	workbookStore := newCatalogBackedWorkbookStore(t, harness.DB, timelineBundle, appender, revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902-tasks-decisions@example.test", "I902 Tasks Decisions", "I902TasksDecisions1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-task-decision-incident", "IR-I902-TD", "Workbook inspector workbook-interaction tasks decisions")

	supportID := mustCreateEvidenceFor(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-decision-support", "workbook-interaction decision support")
	affectedID := mustCreateEvidenceFor(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-decision-affected", "workbook-interaction affected record")
	decision, err := workbookStore.CreateWorkbookRow(context.Background(), actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-decision",
		Values: map[string]workbook.ValueChange{
			"decision.summary":       textChange("Projection-backed decision"),
			"decision.decision_type": textChange("containment"),
			"decision.rationale":     textChange("Decision projection includes support and affected links."),
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"decision.support_refs":        {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}}},
			"decision.affected_record_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &affectedID}}},
		},
	}, []byte("txn-workbook_interaction-i-9-02-decision"), "req-workbook_interaction-i-9-02-decision", time.Date(2026, 5, 17, 16, 15, 0, 0, time.UTC))
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
		ClientTxnID:  "txn-workbook_interaction-i-9-02-task",
		Values: map[string]workbook.ValueChange{
			"task.title":              textChange("Projection-backed task"),
			"task.task_kind":          textChange("collection"),
			"task.due_at":             {Kind: "timestamp", Timestamp: &dueAt},
			"task.decision_record_id": {Kind: "uuid", UUID: &decisionID},
		},
		Collections: map[string]workbook.CollectionActionPayload{
			"task.linked_record_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}}},
		},
	}, []byte("txn-workbook_interaction-i-9-02-task"), "req-workbook_interaction-i-9-02-task", time.Date(2026, 5, 17, 16, 20, 0, 0, time.UTC))
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

func TestWorkbookHotProjectionTablesRebuild_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-i-9-02-hot-projections")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle := timelineassembly.NewBundle(harness.DB, workbookTestConflictTokens(), appender, revisionComposition.Intents, evidence.NewTimelineAttachmentContribution(harness.DB))
	workbookStore := newCatalogBackedWorkbookStore(t, harness.DB, timelineBundle, appender, revisionComposition.Intents)
	projectionStore := projections.NewStore(harness.DB, timelineBundle.ProjectionCatalog.Catalog)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902-hot-projections@example.test", "I902 Hot Projections", "I902HotProjection1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-hot-incident", "IR-I902-HOT", "Workbook inspector workbook-interaction hot projections")

	requireScalarCount(t, harness, `
SELECT count(*)
  FROM information_schema.tables
 WHERE table_schema = 'public'
   AND table_name = 'artifact_grid_projection'
   AND table_type = 'BASE TABLE'
`, 1)
	requireScalarCount(t, harness, `
SELECT count(*)
  FROM information_schema.views
 WHERE table_schema = 'public'
   AND table_name = 'artifact_grid_projection'
`, 0)
	requireScalarCount(t, harness, `SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'evidence_grid_projection'`, 1)
	requireScalarCount(t, harness, `SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'party_grid_projection'`, 1)

	party, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-hot-party",
		Values: map[string]workbook.ValueChange{
			"party.display_name": textChange("Projection table party"),
			"party.party_kind":   textChange("person"),
		},
	}, []byte("txn-workbook_interaction-i-9-02-hot-party"), "req-workbook_interaction-i-9-02-hot-party", time.Date(2026, 6, 30, 13, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party projection row: %v", err)
	}
	evidence, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.EvidenceViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-hot-evidence",
		Values: map[string]workbook.ValueChange{
			"evidence.title": textChange("Projection table evidence"),
		},
	}, []byte("txn-workbook_interaction-i-9-02-hot-evidence"), "req-workbook_interaction-i-9-02-hot-evidence", time.Date(2026, 6, 30, 13, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create evidence projection row: %v", err)
	}
	note, err := workbookStore.CreateWorkbookRow(ctx, actor, incident.ID, workbook.CreateRequest{
		ViewSchemaID: workbook.NotesViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-hot-note",
		Values: map[string]workbook.ValueChange{
			"note.title": textChange("Projection table note"),
		},
	}, []byte("txn-workbook_interaction-i-9-02-hot-note"), "req-workbook_interaction-i-9-02-hot-note", time.Date(2026, 6, 30, 13, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create artifact projection row: %v", err)
	}

	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, workbook.PartiesViewSchemaID), party.RecordID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, workbook.EvidenceViewSchemaID), evidence.RecordID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, workbook.NotesViewSchemaID), note.RecordID)
	requireQueriedRow(t, mustProjectionStoreRows(t, projectionStore, incident.ID, workbook.PartiesViewSchemaID), party.RecordID)
	requireQueriedRow(t, mustProjectionStoreRows(t, projectionStore, incident.ID, workbook.EvidenceViewSchemaID), evidence.RecordID)
	requireQueriedRow(t, mustProjectionStoreRows(t, projectionStore, incident.ID, workbook.NotesViewSchemaID), note.RecordID)

	sourceBefore := snapshotHotProjectionSourceState(t, harness, incident.ID)
	execProjectionSQL(t, harness, `DELETE FROM party_grid_projection WHERE record_id = $1`, party.RecordID)
	execProjectionSQL(t, harness, `DELETE FROM evidence_grid_projection WHERE record_id = $1`, evidence.RecordID)
	execProjectionSQL(t, harness, `DELETE FROM artifact_grid_projection WHERE record_id = $1`, note.RecordID)
	if hasQueriedRow(mustProjectionRows(t, workbookStore, incident.ID, workbook.PartiesViewSchemaID), party.RecordID) {
		t.Fatalf("party query ignored projection table deletion")
	}
	if hasQueriedRow(mustProjectionRows(t, workbookStore, incident.ID, workbook.EvidenceViewSchemaID), evidence.RecordID) {
		t.Fatalf("evidence query ignored projection table deletion")
	}
	if hasQueriedRow(mustProjectionRows(t, workbookStore, incident.ID, workbook.NotesViewSchemaID), note.RecordID) {
		t.Fatalf("artifact query ignored projection table deletion")
	}

	if err := projectionStore.RebuildIncidentParties(ctx, incident.ID); err != nil {
		t.Fatalf("rebuild party projections: %v", err)
	}
	if err := projectionStore.RebuildIncidentEvidence(ctx, incident.ID); err != nil {
		t.Fatalf("rebuild evidence projections: %v", err)
	}
	if err := projectionStore.RebuildIncidentArtifacts(ctx, incident.ID); err != nil {
		t.Fatalf("rebuild artifact projections: %v", err)
	}
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, workbook.PartiesViewSchemaID), party.RecordID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, workbook.EvidenceViewSchemaID), evidence.RecordID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, workbook.NotesViewSchemaID), note.RecordID)
	requireQueriedRow(t, mustProjectionStoreRows(t, projectionStore, incident.ID, workbook.PartiesViewSchemaID), party.RecordID)
	requireQueriedRow(t, mustProjectionStoreRows(t, projectionStore, incident.ID, workbook.EvidenceViewSchemaID), evidence.RecordID)
	requireQueriedRow(t, mustProjectionStoreRows(t, projectionStore, incident.ID, workbook.NotesViewSchemaID), note.RecordID)
	if sourceAfter := snapshotHotProjectionSourceState(t, harness, incident.ID); sourceAfter != sourceBefore {
		t.Fatalf("projection rebuild mutated source/history state: before=%+v after=%+v", sourceBefore, sourceAfter)
	}
}

func TestCoordinationSurfacesQueryThroughWorkbookProjections_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-i-9-02-coordination")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle := timelineassembly.NewBundle(harness.DB, workbookTestConflictTokens(), appender, revisionComposition.Intents, evidence.NewTimelineAttachmentContribution(harness.DB))
	workbookStore := newCatalogBackedWorkbookStore(t, harness.DB, timelineBundle, appender, revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902-coordination@example.test", "I902 Coordination", "I902Coordination1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-coordination-incident", "IR-I902-COORD", "Workbook inspector workbook-interaction coordination")

	partyID := mustCreatePartyFor(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-party", "Projection coordination party")
	taskID := mustCreateTaskFor(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-task", "Projection coordination task")
	evidenceID := mustCreateEvidenceFor(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-evidence", "Projection coordination evidence")
	decisionID := mustCreateDecision(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-decision", "approved", "Projection coordination decision")

	nextReport := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	comm := mustCreateRow(t, workbookStore, actor, incident.ID, workbook.CommLogViewSchemaID, "txn-workbook_interaction-i-9-02-comm", map[string]workbook.ValueChange{
		"comm_log.timestamp_utc":      Timestamp(time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          textChange("briefing"),
		"comm_log.audience":           textChange("Projection leadership"),
		"comm_log.channel_or_meeting": textChange("Bridge"),
		"comm_log.summary":            textChange("Projection-backed comm log"),
		"comm_log.next_report_at":     Timestamp(nextReport),
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
	handoff := mustCreateRow(t, workbookStore, actor, incident.ID, workbook.HandoffViewSchemaID, "txn-workbook_interaction-i-9-02-handoff", map[string]workbook.ValueChange{
		"handoff.timestamp_utc":          Timestamp(time.Date(2026, 5, 22, 13, 0, 0, 0, time.UTC)),
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  textChange("Projection-backed handoff"),
		"handoff.acknowledged_at":        Timestamp(acknowledgedAt),
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
	status := mustCreateRow(t, workbookStore, actor, incident.ID, workbook.StatusReviewViewSchemaID, "txn-workbook_interaction-i-9-02-status", map[string]workbook.ValueChange{
		"status_review.timestamp_utc":         Timestamp(time.Date(2026, 5, 23, 15, 0, 0, 0, time.UTC)),
		"status_review.current_state_summary": textChange("Projection-backed status review"),
		"status_review.next_report_at":        Timestamp(statusNextReport),
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

	lesson := mustCreateRow(t, workbookStore, actor, incident.ID, workbook.LessonViewSchemaID, "txn-workbook_interaction-i-9-02-lesson", map[string]workbook.ValueChange{
		"lesson.timestamp_utc": Timestamp(time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC)),
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

func newCatalogBackedWorkbookStore(
	t testing.TB,
	pool postgres.DB,
	timelineBundle *timelineassembly.Bundle,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) *workbook.Store {
	t.Helper()
	conflictTokens := workbookTestConflictTokens()
	conflictFields, err := revisionassembly.CurrentConflictFieldResolver()
	if err != nil {
		t.Fatalf("compose conflict field resolver: %v", err)
	}
	evidenceContribution := evidence.NewWorkbookContribution(pool, conflictTokens, appender, intents, conflictFields, workbookassembly.NewConflictIdempotencyPort(pool))
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(pool, conflictTokens, appender, conflictFields)
	if err != nil {
		t.Fatalf("compose Tasks/Decisions mutation contribution: %v", err)
	}
	indicatorOwner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    pool,
		Revisions:   appender,
		Projections: timelineBundle.ProjectionCoordinator,
	})
	if err != nil {
		t.Fatalf("compose Indicators owner: %v", err)
	}
	catalog, err := workbookassembly.NewContributionCatalog(
		pool,
		timelineBundle.ProjectionCatalog.Catalog,
		timelineBundle.ProjectionCatalog.Query,
		indicatorOwner,
		timelineBundle.Facade,
		evidenceContribution,
		taskDecisionMutation,
		conflictTokens,
		conflictFields,
		appender,
		intents,
	)
	if err != nil {
		t.Fatalf("compose workbook contribution catalog: %v", err)
	}
	store, err := workbookassembly.NewMutationStore(
		pool,
		catalog,
		appender,
		taskDecisionMutation,
	)
	if err != nil {
		t.Fatalf("compose Workbook mutation store: %v", err)
	}
	return store
}

func hasQueriedRow(rows []map[string]any, recordID uuid.UUID) bool {
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return true
		}
	}
	return false
}

func mustProjectionRows(t testing.TB, store *workbook.Store, incidentID uuid.UUID, viewSchemaID string) []map[string]any {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, viewSchemaID, mustQueryMeta(t, viewSchemaID))
	if err != nil {
		t.Fatalf("query %s projection rows: %v", viewSchemaID, err)
	}
	return rows
}

func mustProjectionStoreRows(t testing.TB, store *projections.Store, incidentID uuid.UUID, viewSchemaID string) []map[string]any {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, viewSchemaID, mustQueryMeta(t, viewSchemaID))
	if err != nil {
		t.Fatalf("query %s through projection store: %v", viewSchemaID, err)
	}
	return rows
}

type HotProjectionSourceState struct {
	Records         int
	Artifacts       int
	Evidence        int
	Parties         int
	ChangeSets      int
	RecordRevisions int
}

func snapshotHotProjectionSourceState(t testing.TB, harness *appsupport.StoreHarness, incidentID uuid.UUID) HotProjectionSourceState {
	t.Helper()
	return HotProjectionSourceState{
		Records:         appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM records WHERE incident_id = $1`, incidentID),
		Artifacts:       appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM artifacts WHERE incident_id = $1`, incidentID),
		Evidence:        appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM evidence WHERE incident_id = $1`, incidentID),
		Parties:         appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM parties WHERE incident_id = $1`, incidentID),
		ChangeSets:      appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incidentID),
		RecordRevisions: appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM record_revisions rr JOIN records r ON r.record_id = rr.record_id WHERE r.incident_id = $1`, incidentID),
	}
}

func execProjectionSQL(t testing.TB, harness *appsupport.StoreHarness, query string, args ...any) {
	t.Helper()
	if _, err := harness.DB.Exec(context.Background(), query, args...); err != nil {
		t.Fatalf("exec projection sql: %v", err)
	}
}

func requireScalarCount(t testing.TB, harness *appsupport.StoreHarness, query string, args ...any) {
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
