package workbook_test

import (
	"context"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentadmission "github.com/JochiRaider/cartulary/internal/modules/assessments/admission"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

type typedNilProjectionQueryCatalog struct{}

func (*typedNilProjectionQueryCatalog) WorkbookQueryProvider(string) (workbook.QueryProvider, bool) {
	return nil, false
}

func TestLinkedNotesCreateContextualArtifactLinks_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-u-9-03-notes")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Intents)
	store, artifactOwner := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u903@example.test", "U903 Notes", "U903NotesPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-03-incident", "IR-U903", "Workbook inspector workbook-storage")
	sourceRecordID := uuid.New()
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceRecordID)

	title := "Workbook inspector linked note"
	body := "Linked through references_artifact"
	linkedRequest := artifacts.ContextualNoteCreateRequest{
		ClientTxnID: "txn-workbook_interaction-u-9-03-linked-note",
		Values: map[string]artifacts.FieldValue{
			"note.title": {Text: &title},
			"note.body":  {Text: &body},
		},
	}
	linked, err := artifactOwner.CreateContextualNote(context.Background(), artifacts.ContextualNoteCreateCommand{
		ActorUserID: actor.ID, SourceRecordID: sourceRecordID, Request: linkedRequest,
		RequestHash: artifacts.ContextualNoteCreateRequestHash(sourceRecordID, linkedRequest),
		RequestID:   "req-workbook_interaction-u-9-03-linked-note",
		OperationID: artifacts.OperationLinkedNoteCreate,
		Now:         time.Date(2026, 5, 17, 15, 5, 0, 0, time.UTC),
	})
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

	rows, err := store.QueryRows(context.Background(), incident.ID, artifacts.NotesViewSchemaID, mustQueryMeta(t, artifacts.NotesViewSchemaID))
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Intents)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.Intents)
	indicatorStore, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:        harness.DB,
		Revisions:       appender,
		RecordEnvelopes: records.NewStore(harness.DB),
		Projections:     projections.IndicatorPorts().Rows,
		SourceText:      indicatorassembly.NewSourceTextPort(projections.SourceTextRows()),
		Clock:           func() time.Time { return time.Date(2026, 5, 17, 16, 5, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("compose Indicator test owner: %v", err)
	}
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902@example.test", "I902 Projection", "I902ProjectionPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-incident", "IR-I902", "Workbook inspector workbook-interaction")

	note, err := workbookroutetest.CreateWorkbookRow(workbookStore, context.Background(), actor, incident.ID, workbookroutetest.CreateRequest{
		ViewSchemaID: artifacts.NotesViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-note",
		Values: map[string]workbookroutetest.ValueChange{
			"note.title": textChange("Projection-backed note"),
			"note.body":  textChange("query token workbook_interaction-note-projection"),
		},
	}, []byte("txn-workbook_interaction-i-9-02-note"), "req-workbook_interaction-i-9-02-note", time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection note: %v", err)
	}
	noteRows, err := workbookStore.QueryRows(context.Background(), incident.ID, artifacts.NotesViewSchemaID, mustQueryMeta(t, artifacts.NotesViewSchemaID))
	if err != nil {
		t.Fatalf("query notes projection: %v", err)
	}
	requireQueriedRow(t, noteRows, note.RecordID)

	indicator, err := indicatorStore.CreateIndicatorRow(context.Background(), actor, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-workbook_interaction-i-9-02-indicator",
		IndicatorType: "ipv4_addr",
		ValueKind:     "atomic",
		DisplayValue:  "203.0.113.45",
	}, "req-workbook_interaction-i-9-02-indicator")
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Intents)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.Intents)
	assessmentOwner := appsupport.NewAssessmentOwner(harness.DB)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902-assessments@example.test", "I902 Assessments", "I902AssessmentsPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-assessment-incident", "IR-I902-ASSESS", "Workbook inspector workbook-interaction assessments")

	hostID := uuid.New()
	supportID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Workbook inspector projection assessment host", "workbook_interaction-projection-assessment", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	confidenceScore := 85
	input := assessments.CreateInput{
		ClientTxnID:     "txn-workbook_interaction-i-9-02-assessment",
		SubjectRef:      hostID,
		SubjectType:     "host",
		AssessmentState: "confirmed",
		ConfidenceScore: &confidenceScore,
		Rationale:       "Projection-backed assessment row.",
		SupportRefs:     []uuid.UUID{supportID},
	}
	created, err := assessmentOwner.Create(context.Background(), assessments.CreateCommand{
		ActorUserID: actor.ID,
		IncidentID:  incident.ID,
		Input:       input,
		Idempotency: assessments.CreateIdempotencyKey{
			RouteKey:    "assessments.rows.create",
			ActorUserID: actor.ID,
			ScopeKey:    incident.ID.String() + ":" + assessments.AssessmentsViewSchemaID,
			ClientTxnID: input.ClientTxnID,
			RequestHash: assessmentadmission.CreateRequestHash(input),
		},
		RequestID: "req-workbook_interaction-i-9-02-assessment",
		Now:       time.Date(2026, 5, 17, 16, 10, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create projection assessment: %v", err)
	}

	query := mustQueryMeta(t, assessments.AssessmentsViewSchemaID)
	query.Filters = []viewschema.Filter{
		{FieldKey: "assessment.confidence_band", Op: "eq", Arg: map[string]any{"value": "high"}},
	}
	rows, err := workbookStore.QueryRows(context.Background(), incident.ID, assessments.AssessmentsViewSchemaID, query)
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Intents)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902-tasks-decisions@example.test", "I902 Tasks Decisions", "I902TasksDecisions1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-task-decision-incident", "IR-I902-TD", "Workbook inspector workbook-interaction tasks decisions")

	supportID := mustCreateEvidenceFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-decision-support", "workbook-interaction decision support")
	affectedID := mustCreateEvidenceFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-decision-affected", "workbook-interaction affected record")
	decision, err := workbookroutetest.CreateWorkbookRow(workbookStore, context.Background(), actor, incident.ID, workbookroutetest.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-decision",
		Values: map[string]workbookroutetest.ValueChange{
			"decision.summary":       textChange("Projection-backed decision"),
			"decision.decision_type": textChange("containment"),
			"decision.rationale":     textChange("Decision projection includes support and affected links."),
		},
		Collections: map[string]workbookroutetest.CollectionActionPayload{
			"decision.support_refs":        {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}}},
			"decision.affected_record_ids": {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &affectedID}}},
		},
	}, []byte("txn-workbook_interaction-i-9-02-decision"), "req-workbook_interaction-i-9-02-decision", time.Date(2026, 5, 17, 16, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection decision: %v", err)
	}

	decisionQuery := mustQueryMeta(t, tasksdecisions.DecisionsViewSchemaID)
	decisionQuery.Filters = []viewschema.Filter{{FieldKey: "decision.status", Op: "eq", Arg: map[string]any{"value": "proposed"}}}
	decisionRows, err := workbookStore.QueryRows(context.Background(), incident.ID, tasksdecisions.DecisionsViewSchemaID, decisionQuery)
	if err != nil {
		t.Fatalf("query decisions through workbook projection: %v", err)
	}
	decisionRow := requireQueriedRow(t, decisionRows, decision.RecordID)
	requireProjectedCollectionCount(t, decisionRow, "decision.support_refs", 1)
	requireProjectedCollectionCount(t, decisionRow, "decision.affected_record_ids", 1)
	requireProjectedNumericCell(t, decisionRow, "decision.affected_record_count", 1)

	dueAt := time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC)
	decisionID := decision.RecordID
	task, err := workbookroutetest.CreateWorkbookRow(workbookStore, context.Background(), actor, incident.ID, workbookroutetest.CreateRequest{
		ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-task",
		Values: map[string]workbookroutetest.ValueChange{
			"task.title":              textChange("Projection-backed task"),
			"task.task_kind":          textChange("collection"),
			"task.due_at":             {Kind: "timestamp", Timestamp: &dueAt},
			"task.decision_record_id": {Kind: "uuid", UUID: &decisionID},
		},
		Collections: map[string]workbookroutetest.CollectionActionPayload{
			"task.linked_record_ids": {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &supportID}}},
		},
	}, []byte("txn-workbook_interaction-i-9-02-task"), "req-workbook_interaction-i-9-02-task", time.Date(2026, 5, 17, 16, 20, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create projection task: %v", err)
	}

	taskQuery := mustQueryMeta(t, tasksdecisions.TaskRequestsViewSchemaID)
	taskQuery.Filters = []viewschema.Filter{
		{FieldKey: "task.status", Op: "eq", Arg: map[string]any{"value": "open"}},
		{FieldKey: "task.owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
		{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": dueAt.Format(time.RFC3339Nano)}},
	}
	taskRows, err := workbookStore.QueryRows(context.Background(), incident.ID, tasksdecisions.TaskRequestsViewSchemaID, taskQuery)
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Intents)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.Intents)
	projectionRebuilder := projections.RevisionRebuilder()
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

	partyID := mustCreatePartyFor(
		t, harness.DB, actor, incident.ID,
		"txn-workbook_interaction-i-9-02-hot-party", "Projection table party",
	)
	evidenceID := mustCreateEvidenceFor(
		t, harness.DB, actor, incident.ID,
		"txn-workbook_interaction-i-9-02-hot-evidence", "Projection table evidence",
	)
	note, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, actor, incident.ID, workbookroutetest.CreateRequest{
		ViewSchemaID: artifacts.NotesViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-i-9-02-hot-note",
		Values: map[string]workbookroutetest.ValueChange{
			"note.title": textChange("Projection table note"),
		},
	}, []byte("txn-workbook_interaction-i-9-02-hot-note"), "req-workbook_interaction-i-9-02-hot-note", time.Date(2026, 6, 30, 13, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create artifact projection row: %v", err)
	}

	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, parties.ViewSchemaID), partyID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, evidence.ViewSchemaID), evidenceID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, artifacts.NotesViewSchemaID), note.RecordID)

	sourceBefore := snapshotHotProjectionSourceState(t, harness, incident.ID)
	execProjectionSQL(t, harness, `DELETE FROM party_grid_projection WHERE record_id = $1`, partyID)
	execProjectionSQL(t, harness, `DELETE FROM evidence_grid_projection WHERE record_id = $1`, evidenceID)
	execProjectionSQL(t, harness, `DELETE FROM artifact_grid_projection WHERE record_id = $1`, note.RecordID)
	if hasQueriedRow(mustProjectionRows(t, workbookStore, incident.ID, parties.ViewSchemaID), partyID) {
		t.Fatalf("party query ignored projection table deletion")
	}
	if hasQueriedRow(mustProjectionRows(t, workbookStore, incident.ID, evidence.ViewSchemaID), evidenceID) {
		t.Fatalf("evidence query ignored projection table deletion")
	}
	if hasQueriedRow(mustProjectionRows(t, workbookStore, incident.ID, artifacts.NotesViewSchemaID), note.RecordID) {
		t.Fatalf("artifact query ignored projection table deletion")
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin projection rebuild transaction: %v", err)
	}
	if err := projectionRebuilder.RebuildIncidentTx(ctx, tx, incident.ID); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("rebuild incident projections: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit projection rebuild transaction: %v", err)
	}
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, parties.ViewSchemaID), partyID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, evidence.ViewSchemaID), evidenceID)
	requireQueriedRow(t, mustProjectionRows(t, workbookStore, incident.ID, artifacts.NotesViewSchemaID), note.RecordID)
	if sourceAfter := snapshotHotProjectionSourceState(t, harness, incident.ID); sourceAfter != sourceBefore {
		t.Fatalf("projection rebuild mutated source/history state: before=%+v after=%+v", sourceBefore, sourceAfter)
	}
}

func TestCoordinationSurfacesQueryThroughWorkbookProjections_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-i-9-02-coordination")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Intents)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "i902-coordination@example.test", "I902 Coordination", "I902Coordination1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-coordination-incident", "IR-I902-COORD", "Workbook inspector workbook-interaction coordination")

	partyID := mustCreatePartyFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-party", "Projection coordination party")
	taskID := mustCreateTaskFor(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-task", "Projection coordination task")
	evidenceID := mustCreateEvidenceFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-evidence", "Projection coordination evidence")
	decisionID := mustCreateDecision(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-decision", "approved", "Projection coordination decision")

	nextReport := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	comm := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.CommLogViewSchemaID, "txn-workbook_interaction-i-9-02-comm", map[string]workbookroutetest.ValueChange{
		"comm_log.timestamp_utc":      Timestamp(time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)),
		"comm_log.comm_type":          textChange("briefing"),
		"comm_log.audience":           textChange("Projection leadership"),
		"comm_log.channel_or_meeting": textChange("Bridge"),
		"comm_log.summary":            textChange("Projection-backed comm log"),
		"comm_log.next_report_at":     Timestamp(nextReport),
	}, map[string]workbookroutetest.CollectionActionPayload{
		"comm_log.decision_ids":       {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &decisionID}}},
		"comm_log.action_task_ids":    {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"comm_log.audience_party_ids": {Actions: []workbookroutetest.CollectionAction{{Op: "add_party_ref", PartyID: &partyID}}},
	}, time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC))
	commQuery := mustQueryMeta(t, artifacts.CommLogViewSchemaID)
	commQuery.Filters = []viewschema.Filter{
		{FieldKey: "comm_log.comm_type", Op: "eq", Arg: map[string]any{"value": "briefing"}},
		{FieldKey: "comm_log.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-24"}},
	}
	commRows, err := workbookStore.QueryRows(context.Background(), incident.ID, artifacts.CommLogViewSchemaID, commQuery)
	if err != nil {
		t.Fatalf("query comm log projection: %v", err)
	}
	commRow := requireQueriedRow(t, commRows, comm.RecordID)
	requireProjectedCollectionCount(t, commRow, "comm_log.decision_ids", 1)
	requireProjectedCollectionCount(t, commRow, "comm_log.action_task_ids", 1)
	requireProjectedCollectionCount(t, commRow, "comm_log.audience_party_ids", 1)

	acknowledgedAt := time.Date(2026, 5, 22, 14, 0, 0, 0, time.UTC)
	handoff := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.HandoffViewSchemaID, "txn-workbook_interaction-i-9-02-handoff", map[string]workbookroutetest.ValueChange{
		"handoff.timestamp_utc":          Timestamp(time.Date(2026, 5, 22, 13, 0, 0, 0, time.UTC)),
		"handoff.incoming_owner_user_id": UUID(actor.ID),
		"handoff.current_state_summary":  textChange("Projection-backed handoff"),
		"handoff.acknowledged_at":        Timestamp(acknowledgedAt),
	}, map[string]workbookroutetest.CollectionActionPayload{
		"handoff.open_task_ids":     {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"handoff.open_decision_ids": {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &decisionID}}},
		"handoff.open_risk_refs":    {Actions: []workbookroutetest.CollectionAction{{Op: "add_risk_ref", RiskRefText: "Projection-backed handoff risk", NormalizedText: "projection-backed handoff risk"}}},
	}, time.Date(2026, 5, 22, 13, 0, 0, 0, time.UTC))
	handoffQuery := mustQueryMeta(t, artifacts.HandoffViewSchemaID)
	handoffQuery.Filters = []viewschema.Filter{{FieldKey: "handoff.ack_state", Op: "eq", Arg: map[string]any{"value": "acknowledged"}}}
	handoffRows, err := workbookStore.QueryRows(context.Background(), incident.ID, artifacts.HandoffViewSchemaID, handoffQuery)
	if err != nil {
		t.Fatalf("query handoff projection: %v", err)
	}
	handoffRow := requireQueriedRow(t, handoffRows, handoff.RecordID)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_task_ids", 1)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_decision_ids", 1)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_risk_refs", 1)

	statusNextReport := time.Date(2026, 5, 25, 15, 0, 0, 0, time.UTC)
	status := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.StatusReviewViewSchemaID, "txn-workbook_interaction-i-9-02-status", map[string]workbookroutetest.ValueChange{
		"status_review.timestamp_utc":         Timestamp(time.Date(2026, 5, 23, 15, 0, 0, 0, time.UTC)),
		"status_review.current_state_summary": textChange("Projection-backed status review"),
		"status_review.next_report_at":        Timestamp(statusNextReport),
	}, map[string]workbookroutetest.CollectionActionPayload{
		"status_review.blocked_task_ids":     {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"status_review.pending_evidence_ids": {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidenceID}}},
		"status_review.open_decision_ids":    {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &decisionID}}},
	}, time.Date(2026, 5, 23, 15, 0, 0, 0, time.UTC))
	statusQuery := mustQueryMeta(t, artifacts.StatusReviewViewSchemaID)
	statusQuery.Filters = []viewschema.Filter{{FieldKey: "status_review.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-25"}}}
	statusRows, err := workbookStore.QueryRows(context.Background(), incident.ID, artifacts.StatusReviewViewSchemaID, statusQuery)
	if err != nil {
		t.Fatalf("query status review projection: %v", err)
	}
	statusRow := requireQueriedRow(t, statusRows, status.RecordID)
	requireProjectedCollectionCount(t, statusRow, "status_review.blocked_task_ids", 1)
	requireProjectedCollectionCount(t, statusRow, "status_review.pending_evidence_ids", 1)
	requireProjectedCollectionCount(t, statusRow, "status_review.open_decision_ids", 1)

	lesson := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.LessonViewSchemaID, "txn-workbook_interaction-i-9-02-lesson", map[string]workbookroutetest.ValueChange{
		"lesson.timestamp_utc": Timestamp(time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC)),
		"lesson.summary":       textChange("Projection-backed lesson"),
		"lesson.closure_state": textChange("closed"),
	}, map[string]workbookroutetest.CollectionActionPayload{
		"lesson.follow_up_task_ids": {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &taskID}}},
		"lesson.evidence_refs":      {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidenceID}}},
	}, time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC))
	lessonQuery := mustQueryMeta(t, artifacts.LessonViewSchemaID)
	lessonQuery.Filters = []viewschema.Filter{{FieldKey: "lesson.closure_state", Op: "eq", Arg: map[string]any{"value": "closed"}}}
	lessonRows, err := workbookStore.QueryRows(context.Background(), incident.ID, artifacts.LessonViewSchemaID, lessonQuery)
	if err != nil {
		t.Fatalf("query lesson projection: %v", err)
	}
	lessonRow := requireQueriedRow(t, lessonRows, lesson.RecordID)
	requireProjectedCollectionCount(t, lessonRow, "lesson.follow_up_task_ids", 1)
	requireProjectedCollectionCount(t, lessonRow, "lesson.evidence_refs", 1)
}

func textChange(value string) workbookroutetest.ValueChange {
	return workbookroutetest.ValueChange{Kind: "text", Text: &value}
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

func newCatalogBackedWorkbookCatalog(
	t testing.TB,
	pool postgres.DB,
	timelineBundle *timelineassembly.Bundle,
	projections *projectionassembly.Runtime,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) (*workbook.WorkbookContributionCatalog, *artifacts.MutationFacade) {
	t.Helper()
	conflictTokens := workbookTestConflictTokens()
	conflictFields, err := revisionassembly.CurrentConflictFieldResolver()
	if err != nil {
		t.Fatalf("compose conflict field resolver: %v", err)
	}
	evidenceOwner := appsupport.NewEvidenceOwnerRuntime(
		pool,
		conflictTokens,
		appender,
		intents,
		appsupport.UnavailableEvidenceObjectStore(),
		conflictFields,
		workbookassembly.NewConflictIdempotencyPort(pool),
		projections,
	)
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		projections.TaskDecisionPorts().Rows,
	)
	if err != nil {
		t.Fatalf("compose Tasks/Decisions mutation contribution: %v", err)
	}
	artifactMutation, err := workbookassembly.NewArtifactMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		projections.ArtifactPorts().Rows,
	)
	if err != nil {
		t.Fatalf("compose Artifacts mutation contribution: %v", err)
	}
	indicatorOwner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:        pool,
		Revisions:       appender,
		RecordEnvelopes: records.NewStore(pool),
		Projections:     projections.IndicatorPorts().Rows,
		SourceText:      indicatorassembly.NewSourceTextPort(projections.SourceTextRows()),
		Clock:           func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("compose Indicators owner: %v", err)
	}
	dependencies := workbookassembly.ContributionDependencies{
		Postgres:              pool,
		ProjectionDescriptors: projections.DescriptorSet(),
		ProjectionQueries:     projections,
		EntityProjections:     projections.EntityPorts(),
		AssessmentProjections: projections.AssessmentPorts().Rows,
		PartyProjections:      projections.PartyPorts().Rows,
		IndicatorOwner:        indicatorOwner,
		TimelineOwner:         timelineBundle.Facade,
		EvidenceOwner:         evidenceOwner.MutationContribution(),
		ArtifactOwner:         artifactMutation,
		TaskDecisionOwner:     taskDecisionMutation,
		ConflictTokens:        conflictTokens,
		ConflictFields:        conflictFields,
		Revisions:             appender,
		CollaborationIntents:  intents,
	}
	requireContributionDependenciesFailClosed(t, dependencies)
	catalog, err := workbookassembly.NewContributionCatalog(dependencies)
	if err != nil {
		t.Fatalf("compose workbook contribution catalog: %v", err)
	}
	return catalog, artifactMutation
}

func requireContributionDependenciesFailClosed(
	t testing.TB,
	valid workbookassembly.ContributionDependencies,
) {
	t.Helper()
	tests := []struct {
		name string
		edit func(*workbookassembly.ContributionDependencies)
	}{
		{name: "Postgres", edit: func(input *workbookassembly.ContributionDependencies) { input.Postgres = nil }},
		{name: "projection descriptors", edit: func(input *workbookassembly.ContributionDependencies) {
			input.ProjectionDescriptors = providercontract.DescriptorSet{}
		}},
		{name: "projection queries", edit: func(input *workbookassembly.ContributionDependencies) { input.ProjectionQueries = nil }},
		{name: "typed-nil projection queries", edit: func(input *workbookassembly.ContributionDependencies) {
			var typedNil *typedNilProjectionQueryCatalog
			input.ProjectionQueries = typedNil
		}},
		{name: "Entity writer", edit: func(input *workbookassembly.ContributionDependencies) { input.EntityProjections.Writer = nil }},
		{name: "Entity reader", edit: func(input *workbookassembly.ContributionDependencies) { input.EntityProjections.Reader = nil }},
		{name: "Assessment rows", edit: func(input *workbookassembly.ContributionDependencies) { input.AssessmentProjections = nil }},
		{name: "Party rows", edit: func(input *workbookassembly.ContributionDependencies) { input.PartyProjections = nil }},
		{name: "Indicator owner", edit: func(input *workbookassembly.ContributionDependencies) { input.IndicatorOwner = nil }},
		{name: "Timeline owner", edit: func(input *workbookassembly.ContributionDependencies) { input.TimelineOwner = nil }},
		{name: "Evidence owner", edit: func(input *workbookassembly.ContributionDependencies) { input.EvidenceOwner = nil }},
		{name: "Artifact owner", edit: func(input *workbookassembly.ContributionDependencies) { input.ArtifactOwner = nil }},
		{name: "Task Decision owner", edit: func(input *workbookassembly.ContributionDependencies) { input.TaskDecisionOwner = nil }},
		{name: "conflict tokens", edit: func(input *workbookassembly.ContributionDependencies) {
			input.ConflictTokens = conflicttokens.ConflictTokenCodec{}
		}},
		{name: "conflict fields", edit: func(input *workbookassembly.ContributionDependencies) { input.ConflictFields = nil }},
		{name: "Revisions", edit: func(input *workbookassembly.ContributionDependencies) { input.Revisions = nil }},
		{name: "Collaboration intents", edit: func(input *workbookassembly.ContributionDependencies) { input.CollaborationIntents = nil }},
	}
	for _, test := range tests {
		input := valid
		test.edit(&input)
		if _, err := workbookassembly.NewContributionCatalog(input); err == nil {
			t.Fatalf("contribution dependencies accepted missing %s", test.name)
		}
	}
}

func newWorkbookTimelineComposition(
	t testing.TB,
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) (*timelineassembly.Bundle, *projectionassembly.Runtime) {
	t.Helper()
	projections, err := projectionassembly.Build(pool)
	if err != nil {
		t.Fatalf("compose projection runtime: %v", err)
	}
	conflictTokens := workbookTestConflictTokens()
	evidenceOwner := appsupport.NewEvidenceOwnerRuntimeForTimeline(
		pool,
		conflictTokens,
		appender,
		intents,
		projections,
	)
	bundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            pool,
		ConflictTokens:      conflictTokens,
		Revisions:           appender,
		Collaboration:       intents,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projections.TimelinePorts().Writer,
		EntityProjection:    projections.EntityPorts().Writer,
		AssessmentRows:      projections.AssessmentPorts().Rows,
	})
	if err != nil {
		t.Fatalf("compose Timeline bundle: %v", err)
	}
	return bundle, projections
}

func hasQueriedRow(rows []map[string]any, recordID uuid.UUID) bool {
	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return true
		}
	}
	return false
}

func mustProjectionRows(t testing.TB, store *workbook.WorkbookContributionCatalog, incidentID uuid.UUID, viewSchemaID string) []map[string]any {
	t.Helper()
	rows, err := store.QueryRows(context.Background(), incidentID, viewSchemaID, mustQueryMeta(t, viewSchemaID))
	if err != nil {
		t.Fatalf("query %s projection rows: %v", viewSchemaID, err)
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
