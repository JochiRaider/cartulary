package runtime_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/importassembly"
	"github.com/JochiRaider/cartulary/internal/app/indicatorassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestProjectionStoreQueryRowsAndLoadRowTxParity(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "projection-query-load-row-parity")
	revisionComposition := revisionsupport.MustComposition(t)
	appender := revisionComposition.Runtime.Appender()
	timelineBundle := timelineassembly.NewBundle(
		harness.DB,
		conflicttest.NewCodec("timeline"),
		appender,
		revisionComposition.Intents,
		evidence.NewTimelineAttachmentContribution(harness.DB),
	)
	projectionCatalog := timelineBundle.Projections
	workbookStore := appsupport.NewWorkbookStore(
		harness.DB,
		conflicttest.NewCodec("workbook"),
	)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "projection-parity@example.test", "Projection Parity", "ProjectionParity1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-projection-parity-incident", "IR-PROJECTION-PARITY", "Projection parity")

	hostID := uuid.New()
	supportID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, hostID, "Projection parity host", "projection-parity-host", "", "")
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, supportID)

	confidenceScore := 75
	confidenceNumber := int64(confidenceScore)
	subjectType := "host"
	assessmentState := "confirmed"
	rationale := "Projection parity assessment."
	assessmentRequest := workbook.CreateRequest{
		ViewSchemaID: assessments.AssessmentsViewSchemaID,
		ClientTxnID:  "txn-projection-parity-assessment",
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
				Number: &confidenceNumber,
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
	assessment, err := workbookStore.CreateWorkbookRow(
		ctx,
		actor,
		incident.ID,
		assessmentRequest,
		workbook.CreateRequestHash(assessmentRequest),
		"req-projection-parity-assessment",
		time.Date(2026, 7, 1, 13, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("create assessment parity row: %v", err)
	}

	indicatorOwner, err := indicators.NewStore(indicators.StoreDependencies{
		Postgres:    harness.DB,
		Revisions:   appender,
		Projections: timelineBundle.IndicatorProjections.Rows,
		SourceText:  indicatorassembly.NewSourceTextPort(projectionCatalog.SourceTextRows()),
	})
	if err != nil {
		t.Fatalf("compose Indicators owner: %v", err)
	}
	importRegistry, err := importassembly.NewOwnerCreateRegistry(
		importassembly.OwnerRegistryDependencies{
			Postgres:                harness.DB,
			RevisionAppender:        appender,
			Intents:                 revisionComposition.Intents,
			Timeline:                timelineBundle.Facade,
			EntityProjections:       timelineBundle.EntityProjections.Writer,
			AssessmentProjections:   timelineBundle.AssessmentProjections.Rows,
			ArtifactProjections:     timelineBundle.ArtifactProjections.Rows,
			EvidenceProjections:     timelineBundle.EvidenceProjections.Rows,
			PartyProjections:        timelineBundle.PartyProjections.Rows,
			TaskDecisionProjections: timelineBundle.TaskDecisionProjections.Rows,
			Indicators:              indicatorOwner,
		},
	)
	if err != nil {
		t.Fatalf("compose import owner registry: %v", err)
	}
	assessmentImportFacade, ok := importRegistry.Resolve(
		assessments.AssessmentsViewSchemaID,
		"assessments.import_create",
	)
	if !ok {
		t.Fatal("assessment import owner facade is not registered")
	}
	importedAssessment := createImportedAssessmentProjectionRow(
		t,
		ctx,
		harness.DB,
		appender,
		assessmentImportFacade,
		actor,
		incident.ID,
		hostID,
	)

	evidence := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, workbook.EvidenceViewSchemaID, "txn-projection-parity-evidence", map[string]workbook.ValueChange{
		"evidence.title": textProjectionValue("Projection parity evidence"),
	}, nil, time.Date(2026, 7, 1, 13, 1, 0, 0, time.UTC))
	note := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, workbook.NotesViewSchemaID, "txn-projection-parity-note", map[string]workbook.ValueChange{
		"note.title": textProjectionValue("Projection parity note"),
		"note.body":  textProjectionValue("Projection parity note body."),
	}, map[string]workbook.CollectionActionPayload{
		"note.tags": {Actions: []workbook.CollectionAction{{Op: "add_tag", RawText: "projection-parity", NormalizedText: "projection-parity"}}},
	}, time.Date(2026, 7, 1, 13, 2, 0, 0, time.UTC))
	party := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, workbook.PartiesViewSchemaID, "txn-projection-parity-party", map[string]workbook.ValueChange{
		"party.display_name": textProjectionValue("Projection Party"),
		"party.party_kind":   textProjectionValue("person"),
	}, nil, time.Date(2026, 7, 1, 13, 3, 0, 0, time.UTC))
	decision := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, workbook.DecisionsViewSchemaID, "txn-projection-parity-decision", map[string]workbook.ValueChange{
		"decision.summary":       textProjectionValue("Projection parity decision"),
		"decision.decision_type": textProjectionValue("containment"),
		"decision.rationale":     textProjectionValue("Projection parity decision rationale."),
	}, map[string]workbook.CollectionActionPayload{
		"decision.support_refs": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidence.RecordID}}},
	}, time.Date(2026, 7, 1, 13, 4, 0, 0, time.UTC))

	dueAt := time.Date(2026, 7, 3, 9, 0, 0, 0, time.UTC)
	task := mustCreateWorkbookProjectionRow(t, workbookStore, actor, incident.ID, workbook.TaskRequestsViewSchemaID, "txn-projection-parity-task", map[string]workbook.ValueChange{
		"task.title":              textProjectionValue("Projection parity task"),
		"task.task_kind":          textProjectionValue("collection"),
		"task.due_at":             {Kind: "timestamp", Timestamp: &dueAt},
		"task.decision_record_id": {Kind: "uuid", UUID: &decision.RecordID},
	}, map[string]workbook.CollectionActionPayload{
		"task.linked_record_ids": {Actions: []workbook.CollectionAction{{Op: "add_record_ref", LinkedRecordID: &evidence.RecordID}}},
	}, time.Date(2026, 7, 1, 13, 5, 0, 0, time.UTC))

	targets := []struct {
		viewSchemaID string
		recordID     uuid.UUID
	}{
		{viewSchemaID: workbook.AssessmentsViewSchemaID, recordID: assessment.RecordID},
		{viewSchemaID: workbook.NotesViewSchemaID, recordID: note.RecordID},
		{viewSchemaID: workbook.EvidenceViewSchemaID, recordID: evidence.RecordID},
		{viewSchemaID: workbook.PartiesViewSchemaID, recordID: party.RecordID},
		{viewSchemaID: workbook.DecisionsViewSchemaID, recordID: decision.RecordID},
		{viewSchemaID: workbook.TaskRequestsViewSchemaID, recordID: task.RecordID},
	}
	for _, target := range targets {
		t.Run(target.viewSchemaID, func(t *testing.T) {
			queryRows, err := projectionCatalog.RestoreProbeQuery().QueryRows(ctx, incident.ID, target.viewSchemaID, defaultProjectionQuery(t, target.viewSchemaID))
			if err != nil {
				t.Fatalf("query projection rows: %v", err)
			}
			queried := requireProjectionRow(t, queryRows, target.recordID)
			loaded := loadProjectionRowTx(t, ctx, harness.DB, projectionCatalog.RevisionServices(), target.viewSchemaID, target.recordID)
			if !reflect.DeepEqual(queried, loaded) {
				t.Fatalf("QueryRows and LoadRowTx diverged for %s\nquery: %s\nload:  %s", target.viewSchemaID, prettyRow(queried), prettyRow(loaded))
			}
		})
	}

	createdAssessmentRow := assessment.Payload["row"].(map[string]any)
	queriedAssessments, err := projectionCatalog.RestoreProbeQuery().QueryRows(
		ctx,
		incident.ID,
		workbook.AssessmentsViewSchemaID,
		defaultProjectionQuery(t, workbook.AssessmentsViewSchemaID),
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
	if err := projectionCatalog.RevisionServices().RebuildIncidentTx(
		ctx,
		tx,
		incident.ID,
	); err != nil {
		t.Fatalf("rebuild assessment projection: %v", err)
	}
	rebuilt, err := projectionCatalog.RevisionServices().LoadRowTx(
		ctx,
		tx,
		workbook.AssessmentsViewSchemaID,
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
	appender *revisions.Appender,
	facade ownerfacade.ImportOwnerCreateFacade,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	subjectID uuid.UUID,
) ownerfacade.ImportOwnerCreateResponse {
	t.Helper()

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin assessment import transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Date(2026, 7, 1, 13, 30, 0, 0, time.UTC)
	changeSetID, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      "imports.unit.apply",
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("append assessment import change set: %v", err)
	}
	subjectType := "host"
	state := "suspected"
	rationale := "Imported projection parity assessment."
	score := int64(55)
	response, err := facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
		Request: ownerfacade.ImportOwnerCreateRequest{
			IncidentID:         incidentID,
			ActorUserID:        actor.ID,
			TargetViewSchemaID: assessments.AssessmentsViewSchemaID,
			ImportSessionID:    uuid.New(),
			ImportUnitID:       uuid.New(),
			ClientTxnID:        "txn-projection-parity-assessment-import",
			FieldValues: []ownerfacade.ImportFieldValue{
				{
					FieldKey: "assessment.subject_ref",
					NormalizedValue: ownerfacade.ImportScalarValue{
						Kind: "uuid",
						UUID: &subjectID,
					},
				},
				{
					FieldKey: "assessment.subject_type",
					NormalizedValue: ownerfacade.ImportScalarValue{
						Kind: "text",
						Text: &subjectType,
					},
				},
				{
					FieldKey: "assessment.assessment_state",
					NormalizedValue: ownerfacade.ImportScalarValue{
						Kind: "text",
						Text: &state,
					},
				},
				{
					FieldKey: "assessment.confidence_score",
					NormalizedValue: ownerfacade.ImportScalarValue{
						Kind:   "number",
						Number: &score,
					},
				},
				{
					FieldKey: "assessment.rationale",
					NormalizedValue: ownerfacade.ImportScalarValue{
						Kind: "text",
						Text: &rationale,
					},
				},
			},
		},
		ChangeSetID: changeSetID,
		SequenceNo:  1,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("create imported assessment: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit assessment import transaction: %v", err)
	}
	return response
}

func mustCreateWorkbookProjectionRow(
	t testing.TB,
	store *workbook.Store,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	viewSchemaID string,
	clientTxnID string,
	values map[string]workbook.ValueChange,
	collections map[string]workbook.CollectionActionPayload,
	now time.Time,
) workbook.MutationResult {
	t.Helper()

	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
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

func textProjectionValue(value string) workbook.ValueChange {
	return workbook.ValueChange{Kind: "text", Text: &value}
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

func loadProjectionRowTx(t testing.TB, ctx context.Context, db postgres.DB, store revisions.ProjectionServices, viewSchemaID string, recordID uuid.UUID) map[string]any {
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
