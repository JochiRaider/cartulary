package runtime_test

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentadmission "github.com/JochiRaider/cartulary/internal/modules/assessments/admission"
	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	projectiontestsupport "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

func TestAssessmentFixtureMutationSQLTransaction(t *testing.T) {
	ctx := context.Background()
	harness := pgtest.Start(t)
	db := harness.OpenIsolatedDatabaseT(t, "assessment-projection-sql-fixture", postgres.PurposeRecovery)
	actorID := uuid.MustParse(authflowtest.SeedLocalUser(
		t,
		db,
		"assessment-projection-sql@example.test",
		"Assessment Projection SQL",
		"AssessmentProjectionSQL1!",
		false,
	))
	incidentID := uuid.MustParse(incidentstoretest.SeedIncidentMembershipSQL(
		t,
		db,
		actorID.String(),
		"assessment_projection_sql",
	))
	recordID := uuid.New()
	assessmenttest.SeedAssessment(
		t,
		db,
		incidentID,
		actorID,
		recordID,
		uuid.New(),
		"host",
		"suspected",
	)

	var rowCount int
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM assessment_grid_projection WHERE record_id = $1
`, recordID).Scan(&rowCount); err != nil {
		t.Fatalf("count upserted assessment projection fixture: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("upserted assessment projection fixture rows = %d, want 1", rowCount)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin assessment projection fixture delete: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := projectiontestsupport.ApplyAssessmentFixtureMutationSQLTx(
		ctx,
		tx,
		assessmentprojection.ProjectionMutation{
			Kind:     assessmentprojection.ProjectionMutationDelete,
			RecordID: recordID,
		},
	); err != nil {
		t.Fatalf("delete assessment projection fixture: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit assessment projection fixture delete: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM assessment_grid_projection WHERE record_id = $1
`, recordID).Scan(&rowCount); err != nil {
		t.Fatalf("count deleted assessment projection fixture: %v", err)
	}
	if rowCount != 0 {
		t.Fatalf("deleted assessment projection fixture rows = %d, want 0", rowCount)
	}
}

func TestProjectionStoreQueryRowsAndLoadRowTxParity(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "projection-query-load-row-parity")
	projectionCatalog := projectiontestsupport.MustBuild(t, harness.DB)
	workbookStore := appsupport.NewWorkbookCatalog(
		harness.DB,
		conflicttest.NewCodec("workbook"),
	)
	assessmentOwner := appsupport.NewAssessmentOwner(harness.DB)
	partyOwner := appsupport.NewPartyOwner(harness.DB, conflicttest.NewCodec("workbook"))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "projection-parity@example.test", "Projection Parity", "ProjectionParity1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-projection-parity-incident", "IR-PROJECTION-PARITY", "Projection parity")

	hostID := uuid.New()
	supportID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Projection parity host", "projection-parity-host", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	confidenceScore := 75
	assessmentInput := assessments.CreateInput{
		ClientTxnID:     "txn-projection-parity-assessment",
		SubjectRef:      hostID,
		SubjectType:     "host",
		AssessmentState: "confirmed",
		ConfidenceScore: &confidenceScore,
		Rationale:       "Projection parity assessment.",
		SupportRefs:     []uuid.UUID{supportID},
	}
	assessment, err := assessmentOwner.Create(ctx, assessments.CreateCommand{
		ActorUserID: actor.ID,
		IncidentID:  incident.ID,
		Input:       assessmentInput,
		Idempotency: assessments.CreateIdempotencyKey{
			RouteKey:    "assessments.rows.create",
			ActorUserID: actor.ID,
			ScopeKey:    incident.ID.String() + ":" + assessments.AssessmentsViewSchemaID,
			ClientTxnID: assessmentInput.ClientTxnID,
			RequestHash: assessmentadmission.CreateRequestHash(assessmentInput),
		},
		RequestID: "req-projection-parity-assessment",
		Now:       time.Date(2026, 7, 1, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create assessment parity row: %v", err)
	}

	importedAssessment := createImportedAssessmentProjectionRow(
		t,
		ctx,
		harness.DB,
		projectionCatalog.Assessments().Rows,
		actor,
		incident.ID,
		hostID,
	)

	evidenceTitle := "Projection parity evidence"
	evidenceRequest := evidence.CreateRequest{
		ViewSchemaID: evidence.ViewSchemaID, ClientTxnID: "txn-projection-parity-evidence",
		Values: map[string]evidence.FieldValue{"evidence.title": {Text: &evidenceTitle}},
	}
	evidenceResult, err := appsupport.NewEvidenceMutationOwner(harness.DB, conflicttest.NewCodec("workbook")).Create(
		ctx,
		evidence.CreateCommand{
			Actor: actor, IncidentID: incident.ID, Request: evidenceRequest,
			RequestHash: evidence.CreateRequestHash(evidenceRequest), RequestID: "req-txn-projection-parity-evidence",
			RouteKey: "workbook.rows.create", Now: time.Date(2026, 7, 1, 13, 1, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("create Evidence projection parity row: %v", err)
	}
	note := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, artifacts.NotesViewSchemaID, "txn-projection-parity-note", map[string]workbookroutetest.ValueChange{
		"note.title": textProjectionValue("Projection parity note"),
		"note.body":  textProjectionValue("Projection parity note body."),
	}, map[string]workbookroutetest.CollectionActionPayload{
		"note.tags": {Actions: []workbookroutetest.CollectionAction{{Op: "add_tag", RawText: "projection-parity", NormalizedText: "projection-parity"}}},
	}, time.Date(2026, 7, 1, 13, 2, 0, 0, time.UTC))
	partyAdmission, apiErr := parties.AdmitCreateJSON(strings.NewReader(fmt.Sprintf(
		`{"client_txn_id":"txn-projection-parity-party","party.display_name":%q,"party.party_kind":"person"}`,
		"Projection Party",
	)))
	if apiErr != nil {
		t.Fatalf("admit Party projection parity row: %#v", apiErr)
	}
	party, err := partyOwner.Create(ctx, parties.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incident.ID, Admission: partyAdmission, RequestID: "req-txn-projection-parity-party",
		RouteKey: "workbook.rows.create", Now: time.Date(2026, 7, 1, 13, 3, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("create Party projection parity row: %v", err)
	}
	decision := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, tasksdecisions.DecisionsViewSchemaID, "txn-projection-parity-decision", map[string]workbookroutetest.ValueChange{
		"decision.summary":       textProjectionValue("Projection parity decision"),
		"decision.decision_type": textProjectionValue("containment"),
		"decision.rationale":     textProjectionValue("Projection parity decision rationale."),
	}, map[string]workbookroutetest.CollectionActionPayload{
		"decision.support_refs": {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidenceResult.RecordID}}},
	}, time.Date(2026, 7, 1, 13, 4, 0, 0, time.UTC))

	dueAt := time.Date(2026, 7, 3, 9, 0, 0, 0, time.UTC)
	task := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, tasksdecisions.TaskRequestsViewSchemaID, "txn-projection-parity-task", map[string]workbookroutetest.ValueChange{
		"task.title":              textProjectionValue("Projection parity task"),
		"task.task_kind":          textProjectionValue("collection"),
		"task.due_at":             {Kind: "timestamp", Timestamp: &dueAt},
		"task.decision_record_id": {Kind: "uuid", UUID: &decision.RecordID},
	}, map[string]workbookroutetest.CollectionActionPayload{
		"task.linked_record_ids": {Actions: []workbookroutetest.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidenceResult.RecordID}}},
	}, time.Date(2026, 7, 1, 13, 5, 0, 0, time.UTC))

	targets := []struct {
		viewSchemaID string
		recordID     uuid.UUID
	}{
		{viewSchemaID: assessments.AssessmentsViewSchemaID, recordID: assessment.RecordID},
		{viewSchemaID: artifacts.NotesViewSchemaID, recordID: note.RecordID},
		{viewSchemaID: evidence.ViewSchemaID, recordID: evidenceResult.RecordID},
		{viewSchemaID: parties.ViewSchemaID, recordID: party.RecordID},
		{viewSchemaID: tasksdecisions.DecisionsViewSchemaID, recordID: decision.RecordID},
		{viewSchemaID: tasksdecisions.TaskRequestsViewSchemaID, recordID: task.RecordID},
	}
	for _, target := range targets {
		t.Run(target.viewSchemaID, func(t *testing.T) {
			queryRows, err := projectionCatalog.RestoreProbeQuery().QueryRows(ctx, incident.ID, target.viewSchemaID, defaultProjectionQuery(t, target.viewSchemaID))
			if err != nil {
				t.Fatalf("query projection rows: %v", err)
			}
			queried := requireProjectionRow(t, queryRows, target.recordID)
			loaded := loadProjectionRowTx(t, ctx, harness.DB, projectionCatalog.RevisionLiveRecords(), target.viewSchemaID, target.recordID)
			if !reflect.DeepEqual(queried, loaded) {
				t.Fatalf("QueryRows and LoadRowTx diverged for %s\nquery: %s\nload:  %s", target.viewSchemaID, prettyRow(queried), prettyRow(loaded))
			}
		})
	}

	createdAssessmentRow := assessment.CanonicalRow
	queriedAssessments, err := projectionCatalog.RestoreProbeQuery().QueryRows(
		ctx,
		incident.ID,
		assessments.AssessmentsViewSchemaID,
		defaultProjectionQuery(t, assessments.AssessmentsViewSchemaID),
	)
	if err != nil {
		t.Fatalf("query created assessment row: %v", err)
	}
	if queried := requireProjectionRow(t, queriedAssessments, assessment.RecordID); !reflect.DeepEqual(
		createdAssessmentRow,
		queried,
	) {
		t.Fatalf(
			"assessment create and query rows diverged\ncreate: %s\nquery:  %s",
			prettyRow(createdAssessmentRow),
			prettyRow(queried),
		)
	}
	if queried := requireProjectionRow(t, queriedAssessments, importedAssessment.RecordID); !reflect.DeepEqual(
		importedAssessment.RowRefresh,
		queried,
	) {
		t.Fatalf(
			"assessment import and query rows diverged\nimport: %s\nquery:  %s",
			prettyRow(importedAssessment.RowRefresh),
			prettyRow(queried),
		)
	}

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin assessment projection rebuild parity: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(
		ctx,
		`DELETE FROM assessment_grid_projection WHERE record_id = $1`,
		assessment.RecordID,
	); err != nil {
		t.Fatalf("clear assessment projection before rebuild: %v", err)
	}
	if err := projectionCatalog.RevisionRebuilder().RebuildIncidentTx(
		ctx,
		tx,
		incident.ID,
	); err != nil {
		t.Fatalf("rebuild assessment projection: %v", err)
	}
	rebuilt, err := projectionCatalog.RevisionLiveRecords().LoadRowTx(
		ctx,
		tx,
		assessments.AssessmentsViewSchemaID,
		assessment.RecordID,
	)
	if err != nil {
		t.Fatalf("load rebuilt assessment projection: %v", err)
	}
	if !reflect.DeepEqual(createdAssessmentRow, rebuilt) {
		t.Fatalf(
			"assessment create and rebuilt rows diverged\ncreate:  %s\nrebuild: %s",
			prettyRow(createdAssessmentRow),
			prettyRow(rebuilt),
		)
	}
}

func createImportedAssessmentProjectionRow(
	t testing.TB,
	ctx context.Context,
	db postgres.DB,
	rows assessmentprojection.Rows,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	subjectID uuid.UUID,
) ownerfacade.ImportOwnerCreateResponse {
	t.Helper()
	recordID := uuid.New()
	assessmenttest.SeedAssessment(
		t,
		db,
		incidentID,
		actor.ID,
		recordID,
		subjectID,
		"host",
		"suspected",
	)

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin imported assessment projection refresh: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := rows.RefreshAssessmentTx(ctx, tx, recordID); err != nil {
		t.Fatalf("refresh imported assessment projection: %v", err)
	}
	row, err := rows.LoadAssessmentTx(ctx, tx, recordID)
	if err != nil {
		t.Fatalf("load imported assessment projection: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit imported assessment projection refresh: %v", err)
	}
	return ownerfacade.ImportOwnerCreateResponse{
		RecordID:        recordID,
		RowVersion:      1,
		CreatedOrReused: "created",
		OwnerResultCode: "created",
		RowRefresh:      row,
	}
}

func mustCreateWorkbookProjectionRow(
	t testing.TB,
	store *workbook.WorkbookContributionCatalog,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	viewSchemaID string,
	clientTxnID string,
	values map[string]workbookroutetest.ValueChange,
	collections map[string]workbookroutetest.CollectionActionPayload,
	now time.Time,
) workbook.MutationResult {
	t.Helper()

	result, err := workbookroutetest.CreateWorkbookRow(store, context.Background(), actor, incidentID, workbookroutetest.CreateRequest{
		ViewSchemaID: viewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
		Collections:  collections,
	}, []byte(clientTxnID), "req-"+clientTxnID, now)
	if err != nil {
		t.Fatalf("create %s row: %v", viewSchemaID, err)
	}
	return result
}

func textProjectionValue(value string) workbookroutetest.ValueChange {
	return workbookroutetest.ValueChange{Kind: "text", Text: &value}
}

func defaultProjectionQuery(t testing.TB, viewSchemaID string) viewschema.QueryMeta {
	t.Helper()

	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok {
		t.Fatalf("missing view schema %s", viewSchemaID)
	}
	return schema.DefaultQueryMeta()
}

func requireProjectionRow(t testing.TB, rows []map[string]any, recordID uuid.UUID) map[string]any {
	t.Helper()

	for _, row := range rows {
		if row["record_id"] == recordID.String() {
			return row
		}
	}
	t.Fatalf("missing projection row %s in %#v", recordID, rows)
	return nil
}

func loadProjectionRowTx(t testing.TB, ctx context.Context, db postgres.DB, store revisions.LiveRecordReader, viewSchemaID string, recordID uuid.UUID) map[string]any {
	t.Helper()

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin projection load tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	row, err := store.LoadRowTx(ctx, tx, viewSchemaID, recordID)
	if err != nil {
		t.Fatalf("load projection row tx: %v", err)
	}
	return row
}

func prettyRow(row map[string]any) string {
	data, err := json.MarshalIndent(row, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
