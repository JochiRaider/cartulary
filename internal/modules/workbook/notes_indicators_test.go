package workbook_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Runtime.ConflictFieldResolver(), revisionComposition.RecordChanges)
	store, artifactOwner := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.RecordChanges)
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "u903@example.test", DisplayName: "U903 Notes"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-03-incident", "IR-U903", "Workbook inspector workbook-storage")
	sourceRecordID := uuid.New()
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceRecordID)

	title := "Workbook inspector linked note"
	body := "Linked through references_artifact"
	linkedPayload, err := json.Marshal(map[string]any{
		"client_txn_id": "txn-workbook_interaction-u-9-03-linked-note",
		"note.title":    title,
		"note.body":     body,
	})
	if err != nil {
		t.Fatalf("encode linked note admission: %v", err)
	}
	linkedAdmission, admissionErr := artifacts.AdmitContextualNote(bytes.NewReader(linkedPayload))
	if admissionErr != nil {
		t.Fatalf("admit linked note: %v", admissionErr)
	}
	linked, err := artifactOwner.CreateContextualNote(context.Background(), artifacts.ContextualNoteCreateCommand{
		ActorUserID: actor.ID, SourceRecordID: sourceRecordID, Admission: linkedAdmission,
		RequestID: "req-workbook_interaction-u-9-03-linked-note",
		Now:       time.Date(2026, 5, 17, 15, 5, 0, 0, time.UTC),
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

	queryProvider1, ok := store.QueryFor(artifacts.NotesViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", artifacts.NotesViewSchemaID)
	}
	rowsPage, err := queryProvider1.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: artifacts.NotesViewSchemaID, Query: mustQueryMeta(t, artifacts.NotesViewSchemaID), Window: querypage.Window{Limit: 100}})
	if rowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	rows := rowsPage.Rows
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Runtime.ConflictFieldResolver(), revisionComposition.RecordChanges)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.RecordChanges)
	indicatorApplication, err := indicators.NewApplication(indicators.ApplicationDependencies{
		ReferencePacks:  referencefixture.IndicatorRegistry{},
		Postgres:        harness.DB,
		Idempotency:     indicatorassembly.NewIdempotencyPort(authn.NewStore(harness.DB)),
		IncidentState:   admission.NewChecker(harness.DB),
		Revisions:       appender,
		Collaboration:   revisionComposition.RecordChanges,
		RecordEnvelopes: records.NewStore(harness.DB),
		Projections:     projections.IndicatorPorts().Rows,
		SourceText:      indicatorassembly.NewSourceTextPort(projections.SourceTextRows()),
		Clock:           func() time.Time { return time.Date(2026, 5, 17, 16, 5, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("compose Indicator test owner: %v", err)
	}
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "i902@example.test", DisplayName: "I902 Projection"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-incident", "IR-I902", "Workbook inspector workbook-interaction")

	createOutcome1, err := workbookroutetest.CreateWorkbookRow(workbookStore, context.Background(), workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: artifacts.NotesViewSchemaID, RequestID: "req-workbook_interaction-i-9-02-note", Now: time.Date(2026, 5, 17, 16, 0, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-i-9-02-note",
		"note.title": "Projection-backed note",
		"note.body":  "query token workbook_interaction-note-projection"}))
	note := workbookroutetest.RequireMutationResult(t, createOutcome1, err)
	queryProvider2, ok := workbookStore.QueryFor(artifacts.NotesViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", artifacts.NotesViewSchemaID)
	}
	noteRowsPage, err := queryProvider2.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: artifacts.NotesViewSchemaID, Query: mustQueryMeta(t, artifacts.NotesViewSchemaID), Window: querypage.Window{Limit: 100}})
	if noteRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	noteRows := noteRowsPage.Rows
	if err != nil {
		t.Fatalf("query notes projection: %v", err)
	}
	requireQueriedRow(t, noteRows, note.RecordID)

	indicator, err := indicatorApplication.CreateIndicatorRow(context.Background(), actor.ID, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-workbook_interaction-i-9-02-indicator",
		IndicatorType: "ipv4_addr",
		ValueKind:     "atomic",
		DisplayValue:  "203.0.113.45",
	}, "req-workbook_interaction-i-9-02-indicator")
	if err != nil {
		t.Fatalf("create projection indicator: %v", err)
	}
	queryProvider3, ok := workbookStore.QueryFor(indicators.ViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", indicators.ViewSchemaID)
	}
	indicatorRowsPage, err := queryProvider3.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: indicators.ViewSchemaID, Query: mustQueryMeta(t, indicators.ViewSchemaID), Window: querypage.Window{Limit: 100}})
	if indicatorRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	indicatorRows := indicatorRowsPage.Rows
	if err != nil {
		t.Fatalf("query indicators through workbook store: %v", err)
	}
	requireQueriedRow(t, indicatorRows, indicator.RecordID)
}

func TestAssessmentsQueryThroughWorkbookProjections_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-i-9-02-assessments")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Runtime.ConflictFieldResolver(), revisionComposition.RecordChanges)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.RecordChanges)
	assessmentOwner := appsupport.NewAssessmentOwner(harness.DB)
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "i902-assessments@example.test", DisplayName: "I902 Assessments"})
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
		RequestID:   "req-workbook_interaction-i-9-02-assessment",
		Now:         time.Date(2026, 5, 17, 16, 10, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create projection assessment: %v", err)
	}

	query := mustQueryMeta(t, assessments.AssessmentsViewSchemaID)
	query.Filters = []viewschema.Filter{
		{FieldKey: "assessment.confidence_band", Op: "eq", Arg: map[string]any{"value": "high"}},
	}
	queryProvider4, ok := workbookStore.QueryFor(assessments.AssessmentsViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", assessments.AssessmentsViewSchemaID)
	}
	rowsPage, err := queryProvider4.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: assessments.AssessmentsViewSchemaID, Query: query, Window: querypage.Window{Limit: 100}})
	if rowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	rows := rowsPage.Rows
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Runtime.ConflictFieldResolver(), revisionComposition.RecordChanges)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.RecordChanges)
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "i902-tasks-decisions@example.test", DisplayName: "I902 Tasks Decisions"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-task-decision-incident", "IR-I902-TD", "Workbook inspector workbook-interaction tasks decisions")

	supportID := mustCreateEvidenceFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-decision-support", "workbook-interaction decision support")
	affectedID := mustCreateEvidenceFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-decision-affected", "workbook-interaction affected record")
	createOutcome2, err := workbookroutetest.CreateWorkbookRow(workbookStore, context.Background(), workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: tasksdecisions.DecisionsViewSchemaID, RequestID: "req-workbook_interaction-i-9-02-decision", Now: time.Date(2026, 5, 17, 16, 15, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-i-9-02-decision",
		"decision.summary":             "Projection-backed decision",
		"decision.decision_type":       "containment",
		"decision.rationale":           "Decision projection includes support and affected links.",
		"decision.support_refs":        map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": supportID}}},
		"decision.affected_record_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": affectedID}}}}))
	decision := workbookroutetest.RequireMutationResult(t, createOutcome2, err)

	decisionQuery := mustQueryMeta(t, tasksdecisions.DecisionsViewSchemaID)
	decisionQuery.Filters = []viewschema.Filter{{FieldKey: "decision.status", Op: "eq", Arg: map[string]any{"value": "proposed"}}}
	queryProvider5, ok := workbookStore.QueryFor(tasksdecisions.DecisionsViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", tasksdecisions.DecisionsViewSchemaID)
	}
	decisionRowsPage, err := queryProvider5.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: tasksdecisions.DecisionsViewSchemaID, Query: decisionQuery, Window: querypage.Window{Limit: 100}})
	if decisionRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	decisionRows := decisionRowsPage.Rows
	if err != nil {
		t.Fatalf("query decisions through workbook projection: %v", err)
	}
	decisionRow := requireQueriedRow(t, decisionRows, decision.RecordID)
	requireProjectedCollectionCount(t, decisionRow, "decision.support_refs", 1)
	requireProjectedCollectionCount(t, decisionRow, "decision.affected_record_ids", 1)
	requireProjectedNumericCell(t, decisionRow, "decision.affected_record_count", 1)

	dueAt := time.Date(2026, 5, 20, 10, 30, 0, 0, time.UTC)
	decisionID := decision.RecordID
	createOutcome3, err := workbookroutetest.CreateWorkbookRow(workbookStore, context.Background(), workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID, RequestID: "req-workbook_interaction-i-9-02-task", Now: time.Date(2026, 5, 17, 16, 20, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-i-9-02-task",
		"task.title":              "Projection-backed task",
		"task.task_kind":          "collection",
		"task.due_at":             dueAt,
		"task.decision_record_id": decisionID,
		"task.linked_record_ids":  map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": supportID}}}}))
	task := workbookroutetest.RequireMutationResult(t, createOutcome3, err)

	taskQuery := mustQueryMeta(t, tasksdecisions.TaskRequestsViewSchemaID)
	taskQuery.Filters = []viewschema.Filter{
		{FieldKey: "task.status", Op: "eq", Arg: map[string]any{"value": "open"}},
		{FieldKey: "task.owner_user_id", Op: "eq", Arg: map[string]any{"value": actor.ID.String()}},
		{FieldKey: "task.due_at", Op: "range", Arg: map[string]any{"lte": dueAt.Format(time.RFC3339Nano)}},
	}
	queryProvider6, ok := workbookStore.QueryFor(tasksdecisions.TaskRequestsViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", tasksdecisions.TaskRequestsViewSchemaID)
	}
	taskRowsPage, err := queryProvider6.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID, Query: taskQuery, Window: querypage.Window{Limit: 100}})
	if taskRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	taskRows := taskRowsPage.Rows
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Runtime.ConflictFieldResolver(), revisionComposition.RecordChanges)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.RecordChanges)
	projectionRebuilder := projections.ImportRebuilder()
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "i902-hot-projections@example.test", DisplayName: "I902 Hot Projections"})
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
	createOutcome4, err := workbookroutetest.CreateWorkbookRow(workbookStore, ctx, workbook.CreateCommand{Actor: actor, IncidentID: incident.ID, ViewSchemaID: artifacts.NotesViewSchemaID, RequestID: "req-workbook_interaction-i-9-02-hot-note", Now: time.Date(2026, 6, 30, 13, 2, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": "txn-workbook_interaction-i-9-02-hot-note",
		"note.title": "Projection table note"}))
	note := workbookroutetest.RequireMutationResult(t, createOutcome4, err)

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
	if err := projectionRebuilder.RebuildImportedIncidentTx(ctx, tx, incident.ID); err != nil {
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
	timelineBundle, projections := newWorkbookTimelineComposition(t, harness.DB, appender, revisionComposition.Runtime.ConflictFieldResolver(), revisionComposition.RecordChanges)
	workbookStore, _ := newCatalogBackedWorkbookCatalog(t, harness.DB, timelineBundle, projections, appender, revisionComposition.RecordChanges)
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "i902-coordination@example.test", DisplayName: "I902 Coordination"})
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-i-9-02-coordination-incident", "IR-I902-COORD", "Workbook inspector workbook-interaction coordination")

	partyID := mustCreatePartyFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-party", "Projection coordination party")
	taskID := mustCreateTaskFor(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-task", "Projection coordination task")
	evidenceID := mustCreateEvidenceFor(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-evidence", "Projection coordination evidence")
	decisionID := mustCreateDecision(t, workbookStore, actor, incident.ID, "txn-workbook_interaction-i-9-02-coordination-decision", "approved", "Projection coordination decision")

	nextReport := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	comm := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.CommLogViewSchemaID, "txn-workbook_interaction-i-9-02-comm", map[string]any{"comm_log.timestamp_utc": time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC),
		"comm_log.comm_type":          "briefing",
		"comm_log.audience":           "Projection leadership",
		"comm_log.channel_or_meeting": "Bridge",
		"comm_log.summary":            "Projection-backed comm log",
		"comm_log.next_report_at":     nextReport}, map[string]any{"comm_log.decision_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": decisionID}}},
		"comm_log.action_task_ids":    map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": taskID}}},
		"comm_log.audience_party_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_party_ref", "party_id": partyID}}}}, time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC))
	commQuery := mustQueryMeta(t, artifacts.CommLogViewSchemaID)
	commQuery.Filters = []viewschema.Filter{
		{FieldKey: "comm_log.comm_type", Op: "eq", Arg: map[string]any{"value": "briefing"}},
		{FieldKey: "comm_log.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-24"}},
	}
	queryProvider7, ok := workbookStore.QueryFor(artifacts.CommLogViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", artifacts.CommLogViewSchemaID)
	}
	commRowsPage, err := queryProvider7.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: artifacts.CommLogViewSchemaID, Query: commQuery, Window: querypage.Window{Limit: 100}})
	if commRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	commRows := commRowsPage.Rows
	if err != nil {
		t.Fatalf("query comm log projection: %v", err)
	}
	commRow := requireQueriedRow(t, commRows, comm.RecordID)
	requireProjectedCollectionCount(t, commRow, "comm_log.decision_ids", 1)
	requireProjectedCollectionCount(t, commRow, "comm_log.action_task_ids", 1)
	requireProjectedCollectionCount(t, commRow, "comm_log.audience_party_ids", 1)

	acknowledgedAt := time.Date(2026, 5, 22, 14, 0, 0, 0, time.UTC)
	handoff := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.HandoffViewSchemaID, "txn-workbook_interaction-i-9-02-handoff", map[string]any{"handoff.timestamp_utc": time.Date(2026, 5, 22, 13, 0, 0, 0, time.UTC),
		"handoff.incoming_owner_user_id": actor.ID,
		"handoff.current_state_summary":  "Projection-backed handoff",
		"handoff.acknowledged_at":        acknowledgedAt}, map[string]any{"handoff.open_task_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": taskID}}},
		"handoff.open_decision_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": decisionID}}},
		"handoff.open_risk_refs":    map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_risk_ref", "risk_ref_text": "Projection-backed handoff risk"}}}}, time.Date(2026, 5, 22, 13, 0, 0, 0, time.UTC))
	handoffQuery := mustQueryMeta(t, artifacts.HandoffViewSchemaID)
	handoffQuery.Filters = []viewschema.Filter{{FieldKey: "handoff.ack_state", Op: "eq", Arg: map[string]any{"value": "acknowledged"}}}
	queryProvider8, ok := workbookStore.QueryFor(artifacts.HandoffViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", artifacts.HandoffViewSchemaID)
	}
	handoffRowsPage, err := queryProvider8.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: artifacts.HandoffViewSchemaID, Query: handoffQuery, Window: querypage.Window{Limit: 100}})
	if handoffRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	handoffRows := handoffRowsPage.Rows
	if err != nil {
		t.Fatalf("query handoff projection: %v", err)
	}
	handoffRow := requireQueriedRow(t, handoffRows, handoff.RecordID)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_task_ids", 1)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_decision_ids", 1)
	requireProjectedCollectionCount(t, handoffRow, "handoff.open_risk_refs", 1)

	statusNextReport := time.Date(2026, 5, 25, 15, 0, 0, 0, time.UTC)
	status := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.StatusReviewViewSchemaID, "txn-workbook_interaction-i-9-02-status", map[string]any{"status_review.timestamp_utc": time.Date(2026, 5, 23, 15, 0, 0, 0, time.UTC),
		"status_review.current_state_summary": "Projection-backed status review",
		"status_review.next_report_at":        statusNextReport}, map[string]any{"status_review.blocked_task_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": taskID}}},
		"status_review.pending_evidence_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": evidenceID}}},
		"status_review.open_decision_ids":    map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": decisionID}}}}, time.Date(2026, 5, 23, 15, 0, 0, 0, time.UTC))
	statusQuery := mustQueryMeta(t, artifacts.StatusReviewViewSchemaID)
	statusQuery.Filters = []viewschema.Filter{{FieldKey: "status_review.next_report_day", Op: "eq", Arg: map[string]any{"value": "2026-05-25"}}}
	queryProvider9, ok := workbookStore.QueryFor(artifacts.StatusReviewViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", artifacts.StatusReviewViewSchemaID)
	}
	statusRowsPage, err := queryProvider9.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: artifacts.StatusReviewViewSchemaID, Query: statusQuery, Window: querypage.Window{Limit: 100}})
	if statusRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	statusRows := statusRowsPage.Rows
	if err != nil {
		t.Fatalf("query status review projection: %v", err)
	}
	statusRow := requireQueriedRow(t, statusRows, status.RecordID)
	requireProjectedCollectionCount(t, statusRow, "status_review.blocked_task_ids", 1)
	requireProjectedCollectionCount(t, statusRow, "status_review.pending_evidence_ids", 1)
	requireProjectedCollectionCount(t, statusRow, "status_review.open_decision_ids", 1)

	lesson := mustCreateRow(t, workbookStore, actor, incident.ID, artifacts.LessonViewSchemaID, "txn-workbook_interaction-i-9-02-lesson", map[string]any{"lesson.timestamp_utc": time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC),
		"lesson.summary":       "Projection-backed lesson",
		"lesson.closure_state": "closed"}, map[string]any{"lesson.follow_up_task_ids": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": taskID}}},
		"lesson.evidence_refs": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{map[string]any{"op": "add_record_ref", "linked_record_id": evidenceID}}}}, time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC))
	lessonQuery := mustQueryMeta(t, artifacts.LessonViewSchemaID)
	lessonQuery.Filters = []viewschema.Filter{{FieldKey: "lesson.closure_state", Op: "eq", Arg: map[string]any{"value": "closed"}}}
	queryProvider10, ok := workbookStore.QueryFor(artifacts.LessonViewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", artifacts.LessonViewSchemaID)
	}
	lessonRowsPage, err := queryProvider10.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incident.ID, ViewSchemaID: artifacts.LessonViewSchemaID, Query: lessonQuery, Window: querypage.Window{Limit: 100}})
	if lessonRowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	lessonRows := lessonRowsPage.Rows
	if err != nil {
		t.Fatalf("query lesson projection: %v", err)
	}
	lessonRow := requireQueriedRow(t, lessonRows, lesson.RecordID)
	requireProjectedCollectionCount(t, lessonRow, "lesson.follow_up_task_ids", 1)
	requireProjectedCollectionCount(t, lessonRow, "lesson.evidence_refs", 1)
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
	intents collaboration.RecordChangedAppender,
) (*workbook.WorkbookContributionCatalog, *artifacts.MutationFacade) {
	t.Helper()
	conflictTokens := workbookTestConflictTokens()
	conflictFields := revisionsupport.MustConflictFieldResolver(t)
	evidenceOwner := appsupport.NewEvidenceSourceRuntime(
		pool,
		conflictTokens,
		appender,
		intents,
		conflictFields,
		workbookassembly.NewConflictIdempotencyPort(pool),
		projections,
	)
	taskDecisionMutation, err := workbookassembly.NewTaskDecisionMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		projections.TaskDecisionMutationRows(),
		intents,
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
		intents,
	)
	if err != nil {
		t.Fatalf("compose Artifacts mutation contribution: %v", err)
	}
	indicatorOwner, err := indicators.NewApplication(indicators.ApplicationDependencies{
		ReferencePacks:  referencefixture.IndicatorRegistry{},
		Postgres:        pool,
		Idempotency:     indicatorassembly.NewIdempotencyPort(authn.NewStore(pool)),
		IncidentState:   admission.NewChecker(pool),
		Revisions:       appender,
		RecordEnvelopes: records.NewStore(pool),
		Projections:     projections.IndicatorPorts().Rows,
		SourceText:      indicatorassembly.NewSourceTextPort(projections.SourceTextRows()),
		Collaboration:   intents,
		Clock:           func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("compose Indicators owner: %v", err)
	}
	dependencies := workbookassembly.ContributionDependencies{
		Postgres:              pool,
		ProjectionDescriptors: projections.DescriptorSet(),
		ProjectionQueries:     projections,
		EntityMutationRows:    projections.EntityMutationRows(),
		EntityQueryReader:     projections.EntityQueryReader(),
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
		{name: "Entity mutation rows", edit: func(input *workbookassembly.ContributionDependencies) { input.EntityMutationRows = nil }},
		{name: "Entity query reader", edit: func(input *workbookassembly.ContributionDependencies) { input.EntityQueryReader = nil }},
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
	conflictFields conflicttokens.FieldResolver,
	intents collaboration.RecordChangedAppender,
) (*timelineassembly.Bundle, *projectionassembly.Runtime) {
	t.Helper()
	projections, err := projectionassembly.Build(pool)
	if err != nil {
		t.Fatalf("compose projection runtime: %v", err)
	}
	conflictTokens := workbookTestConflictTokens()
	evidenceOwner := appsupport.NewEvidenceSourceRuntimeForTimeline(
		pool,
		conflictTokens,
		appender,
		intents,
		projections,
	)
	bundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            pool,
		ConflictTokens:      conflictTokens,
		ConflictFields:      conflictFields,
		Revisions:           appender,
		Collaboration:       intents,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
		TimelineProjection:  projections.TimelinePorts().Writer,
		EntityProjection:    projections.EntityMutationRows(),
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
	queryProvider11, ok := store.QueryFor(viewSchemaID)
	if !ok {
		t.Fatalf("query provider missing for %s", viewSchemaID)
	}
	rowsPage, err := queryProvider11.QueryRowsPage(context.Background(), workbook.QueryCommand{IncidentID: incidentID, ViewSchemaID: viewSchemaID, Query: mustQueryMeta(t, viewSchemaID), Window: querypage.Window{Limit: 100}})
	if rowsPage.HasMore {
		t.Fatal("fixture exceeds explicit 100-row query window")
	}
	rows := rowsPage.Rows
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
