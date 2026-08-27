package tasksdecisions_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

func TestDecisionLifecycleSupersessionAndConsistency_Unit(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-decisions")
	codec := conflicttest.NewCodec("workbook")
	store := appsupport.NewWorkbookCatalog(harness.DB, codec)
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-decision@example.test", "TaskDecision Decision", "TaskDecisionDecision1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-decision-incident", "IR-TASK-DECISION-DECISION", "Workbook inspector task and decision workflow decisions")

	beforeRecords := countRecords(t, harness.DB, incident.ID)
	minimumRequest := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-task-decision-decision-minimum-fail",
		Values: map[string]tasksdecisions.FieldValue{
			"decision.decision_type": {Text: stringPtr("scope")},
		},
	}
	_, err := createTaskDecision(owner, actor, incident.ID, minimumRequest, "req-workbook_interaction-task-decision-decision-minimum-fail", testTime(0))
	requireMutationValidation(t, err, "decision.summary", "missing_required_field")
	if got := countRecords(t, harness.DB, incident.ID); got != beforeRecords {
		t.Fatalf("rejected minimum decision create wrote records: got %d want %d", got, beforeRecords)
	}
	supersededRequest := tasksdecisions.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  "txn-workbook_interaction-task-decision-decision-create-superseded",
		Values: map[string]tasksdecisions.FieldValue{
			"decision.summary":       {Text: stringPtr("Bad superseded create")},
			"decision.decision_type": {Text: stringPtr("scope")},
			"decision.rationale":     {Text: stringPtr("Superseded must be explicit.")},
			"decision.status":        {Text: stringPtr("superseded")},
		},
	}
	_, err = createTaskDecision(owner, actor, incident.ID, supersededRequest, "req-workbook_interaction-task-decision-decision-create-superseded", testTime(0))
	requireLifecycle(t, err)

	support := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-support", "Decision support record")
	affectedOne := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-affected-one", "Decision affected record one")
	affectedTwo := mustCreateEvidence(t, harness.DB, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-affected-two", "Decision affected record two")
	relationshipDecision := mustCreateDecisionWithCollections(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-relationships", "proposed", "Relationship decision", map[string]tasksdecisions.CollectionActionPayload{
		"decision.support_refs":        collectionActions(addOptionalSurfaceRecordRef(support)),
		"decision.affected_record_ids": collectionActions(addOptionalSurfaceRecordRef(affectedOne), addOptionalSurfaceRecordRef(affectedOne)),
	})
	relationshipRow := relationshipDecision.Row
	requireCollectionItemCount(t, relationshipRow, "decision.support_refs", 1)
	requireCollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 1)
	requireCellNumericValue(t, relationshipRow, "decision.affected_record_count", 1)
	requireManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, support, "decision.support_refs", "supported_by")
	requireManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, affectedOne, "decision.affected_record_ids", "references_record")

	relationshipDecision = mustPatch(t, owner, actor, relationshipDecision.RecordID, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-affected-add",
		collectionChange("decision.affected_record_ids", collectionActions(addOptionalSurfaceRecordRef(affectedTwo))))
	relationshipRow = relationshipDecision.Row
	requireCollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 2)
	requireCellNumericValue(t, relationshipRow, "decision.affected_record_count", 2)
	requireManualReferenceLink(t, harness.DB, relationshipDecision.RecordID, affectedTwo, "decision.affected_record_ids", "references_record")

	relationshipDecision = mustPatch(t, owner, actor, relationshipDecision.RecordID, tasksdecisions.DecisionsViewSchemaID, 2, "txn-workbook_interaction-task-decision-decision-affected-remove",
		collectionChange("decision.affected_record_ids", collectionActions(removeRecordRef(affectedOne))))
	relationshipRow = relationshipDecision.Row
	requireCollectionItemCount(t, relationshipRow, "decision.affected_record_ids", 1)
	requireCellNumericValue(t, relationshipRow, "decision.affected_record_count", 1)
	if got := countReferenceLinks(t, harness.DB, relationshipDecision.RecordID, "decision.affected_record_ids"); got != 1 {
		t.Fatalf("decision affected record links after remove: got %d want 1", got)
	}

	target := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-target", "proposed", "Target decision")
	source := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-source", "approved", "Superseding decision")
	executed := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-executed", "approved", "Executed decision")
	executedRow := mustPatch(t, owner, actor, executed, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-approved-executed",
		valueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("executed")}))
	requireCellValue(t, executedRow.Row, "decision.status", "executed")

	_, err = patchRecord(owner, actor, target, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-direct-superseded",
		valueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("superseded")}))
	requireLifecycle(t, err)
	rejected := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-rejected", "rejected", "Rejected decision")
	_, err = patchRecord(owner, actor, rejected, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-rejected-proposed",
		valueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("proposed")}))
	requireLifecycle(t, err)
	_, err = patchRecord(owner, actor, executed, tasksdecisions.DecisionsViewSchemaID, 2, "txn-workbook_interaction-task-decision-decision-executed-approved",
		valueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("approved")}))
	requireLifecycle(t, err)

	request := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-supersede",
		Reason:              "Supersede with better containment rationale.",
		ReplacementRecordID: &source,
	}
	result, err := supersedeDecision(ctx, owner, actor, target, request, "req-workbook_interaction-task-decision-decision-supersede", testTime(time.Hour))
	if err != nil {
		t.Fatalf("supersede proposed target: %v", err)
	}
	if result.ViewSchemaID != tasksdecisions.DecisionsViewSchemaID || result.Facts.TargetStatus != "superseded" {
		t.Fatalf("unexpected decision supersede result: %#v", result)
	}
	if len(result.AdditionalRecordChanges) != 2 {
		t.Fatalf("expected two changed decision rows, got %d", len(result.AdditionalRecordChanges))
	}
	requireDecisionSupersessionChangeSetEffects(t, harness.DB, result.ChangeSetID, source, target)
	if got := countSupersedesLinks(t, harness.DB, source, target); got != 1 {
		t.Fatalf("decision supersedes link count: got %d want 1", got)
	}
	decisionRows, err := workbookroutetest.QueryRows(store, ctx, incident.ID, tasksdecisions.DecisionsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "decision.is_superseded", Op: "eq", Arg: map[string]any{"value": true}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query decision superseded projection: %v", err)
	}
	if !rowsContain(decisionRows, target) {
		t.Fatalf("superseded target missing from projection rows: %#v", decisionRows)
	}
	sourceRows, err := workbookroutetest.QueryRows(store, ctx, incident.ID, tasksdecisions.DecisionsViewSchemaID, viewschema.QueryMeta{
		Filters: []viewschema.Filter{{FieldKey: "decision.supersedes_record_id", Op: "eq", Arg: map[string]any{"value": target.String()}}},
		Sort:    []viewschema.SortEntry{{FieldKey: "decision.updated_at", Direction: "desc"}},
	})
	if err != nil {
		t.Fatalf("query decision supersedes projection: %v", err)
	}
	if len(sourceRows) != 1 || sourceRows[0]["record_id"] != source.String() {
		t.Fatalf("superseding decision missing from projection rows: %#v", sourceRows)
	}
	afterSupersessionEffects := decisionSupersessionIncidentEffects(t, harness.DB, incident.ID)
	replay, err := supersedeDecision(ctx, owner, actor, target, request, "req-workbook_interaction-task-decision-decision-supersede-replay", testTime(2*time.Hour))
	if err != nil {
		t.Fatalf("replay decision supersede: %v", err)
	}
	if !replay.Replayed {
		t.Fatalf("expected idempotent replay")
	}
	if got := countSupersedesLinks(t, harness.DB, source, target); got != 1 {
		t.Fatalf("decision supersedes link count after replay: got %d want 1", got)
	}
	if got := decisionSupersessionIncidentEffects(t, harness.DB, incident.ID); got != afterSupersessionEffects {
		t.Fatalf("decision supersede replay changed durable effects: before=%+v after=%+v", afterSupersessionEffects, got)
	}
	rejectedRequest := request
	rejectedRequest.BaseRowVersion = result.RowVersion
	rejectedRequest.ClientTxnID = "txn-workbook_interaction-task-decision-decision-supersede-rejected"
	_, err = supersedeDecision(ctx, owner, actor, target, rejectedRequest, "req-workbook_interaction-task-decision-decision-supersede-rejected", testTime(2*time.Hour))
	requireLifecycle(t, err)
	if got := decisionSupersessionIncidentEffects(t, harness.DB, incident.ID); got != afterSupersessionEffects {
		t.Fatalf("rejected decision supersede changed durable effects: before=%+v after=%+v", afterSupersessionEffects, got)
	}
	requireSupersessionPublicationRollback(t, harness.DB, owner, codec, actor, incident.ID)

	executedTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-executed-target", "executed", "Executed target")
	executedSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-executed-source", "approved", "Executed target replacement")
	executedRequest := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-supersede-executed",
		Reason:              "Supersede executed decision.",
		ReplacementRecordID: &executedSource,
	}
	if _, err := supersedeDecision(ctx, owner, actor, executedTarget, executedRequest, "req-workbook_interaction-task-decision-decision-supersede-executed", testTime(3*time.Hour)); err != nil {
		t.Fatalf("supersede executed target: %v", err)
	}
	row := queryOne(t, store, incident.ID, tasksdecisions.DecisionsViewSchemaID, "decision.is_superseded", true, executedTarget)
	requireCellValue(t, row, "decision.status", "executed")
	requireCellValue(t, row, "decision.is_superseded", true)

	badSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-bad-source", "proposed", "Inconsistent source")
	badTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-bad-target", "proposed", "Inconsistent target")
	if _, err := harness.DB.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'supersedes', NULL, 'manual', NULL, $4, $4, $5, $5)
`, incident.ID, badSource, badTarget, actor.ID, testTime(4*time.Hour)); err != nil {
		t.Fatalf("seed inconsistent supersedes link: %v", err)
	}
	_, err = patchRecord(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-fail",
		valueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr("approved")}))
	requireLifecycle(t, err)
	beforeInconsistentVersion := recordVersion(t, harness.DB, badSource)
	_, err = patchRecord(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-rationale-fail",
		valueChange("decision.rationale", tasksdecisions.FieldValue{Text: stringPtr("Ordinary scalar edits must fail closed.")}))
	requireLifecycle(t, err)
	if got := recordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision scalar patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
	beforeSupportLinks := countReferenceLinks(t, harness.DB, badSource, "decision.support_refs")
	_, err = patchRecord(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-support-fail",
		collectionChange("decision.support_refs", collectionActions(addOptionalSurfaceRecordRef(support))))
	requireLifecycle(t, err)
	if got := countReferenceLinks(t, harness.DB, badSource, "decision.support_refs"); got != beforeSupportLinks {
		t.Fatalf("inconsistent decision support patch changed links: got %d want %d", got, beforeSupportLinks)
	}
	if got := recordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision support patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
	beforeAffectedLinks := countReferenceLinks(t, harness.DB, badSource, "decision.affected_record_ids")
	_, err = patchRecord(owner, actor, badSource, tasksdecisions.DecisionsViewSchemaID, 1, "txn-workbook_interaction-task-decision-decision-inconsistent-affected-fail",
		collectionChange("decision.affected_record_ids", collectionActions(addOptionalSurfaceRecordRef(affectedTwo))))
	requireLifecycle(t, err)
	if got := countReferenceLinks(t, harness.DB, badSource, "decision.affected_record_ids"); got != beforeAffectedLinks {
		t.Fatalf("inconsistent decision affected patch changed links: got %d want %d", got, beforeAffectedLinks)
	}
	if got := recordVersion(t, harness.DB, badSource); got != beforeInconsistentVersion {
		t.Fatalf("inconsistent decision affected patch changed row version: got %d want %d", got, beforeInconsistentVersion)
	}
}

var errInjectedSupersessionPublication = errors.New("injected supersession publication failure")

type failNthRecordChangedAppender struct {
	delegate collaboration.RecordChangedAppender
	failOn   int
	calls    int
}

func (a *failNthRecordChangedAppender) AppendRecordChangedTx(
	ctx context.Context,
	tx pgx.Tx,
	input collaboration.RecordChangeIntentInput,
) error {
	a.calls++
	if a.calls == a.failOn {
		return errInjectedSupersessionPublication
	}
	return a.delegate.AppendRecordChangedTx(ctx, tx, input)
}

func requireSupersessionPublicationRollback(
	t *testing.T,
	pool postgres.DB,
	successOwner *tasksdecisions.MutationFacade,
	codec conflicttokens.ConflictTokenCodec,
	actor authn.UserRecord,
	incidentID uuid.UUID,
) {
	t.Helper()
	for _, failOn := range []int{1, 2} {
		t.Run(fmt.Sprintf("publication_%d_failure_rolls_back_every_prior_effect", failOn), func(t *testing.T) {
			clientTxnID := fmt.Sprintf("txn-workbook_interaction-task-decision-supersede-publication-failure-%d", failOn)
			target := mustCreateDecision(t, successOwner, actor, incidentID, clientTxnID+"-target", "proposed", "Publication rollback target")
			source := mustCreateDecision(t, successOwner, actor, incidentID, clientTxnID+"-source", "approved", "Publication rollback source")
			targetBefore := decisionSnapshot(t, pool, target)
			sourceBefore := decisionSnapshot(t, pool, source)
			effectsBefore := decisionSupersessionIncidentEffects(t, pool, incidentID)
			request := tasksdecisions.SupersedeRequest{
				BaseRowVersion:      targetBefore.RowVersion,
				ClientTxnID:         clientTxnID,
				Reason:              "Characterize atomic rollback across the publication boundary.",
				ReplacementRecordID: &source,
			}
			failingPublications := &failNthRecordChangedAppender{
				delegate: collaborationsupport.NewPublicationAppender(),
				failOn:   failOn,
			}
			failingOwner := appsupport.NewTaskDecisionOwnerWithPublications(pool, codec, failingPublications)
			_, err := supersedeDecision(context.Background(), failingOwner, actor, target, request, "req-"+clientTxnID, testTime(5*time.Hour))
			if !errors.Is(err, errInjectedSupersessionPublication) {
				t.Fatalf("supersession publication failure %d = %v; want injected failure", failOn, err)
			}
			requireDecisionSnapshot(t, decisionSnapshot(t, pool, target), targetBefore, "target after injected publication failure")
			requireDecisionSnapshot(t, decisionSnapshot(t, pool, source), sourceBefore, "source after injected publication failure")
			if got := countSupersedesLinks(t, pool, source, target); got != 0 {
				t.Fatalf("publication failure %d retained supersedes link: got %d want 0", failOn, got)
			}
			if got := decisionSupersessionIncidentEffects(t, pool, incidentID); got != effectsBefore {
				t.Fatalf("publication failure %d retained durable effects: before=%+v after=%+v", failOn, effectsBefore, got)
			}

			retry, err := supersedeDecision(context.Background(), successOwner, actor, target, request, "req-"+clientTxnID+"-retry", testTime(6*time.Hour))
			if err != nil {
				t.Fatalf("retry after publication failure %d: %v", failOn, err)
			}
			if retry.Replayed {
				t.Fatalf("retry after rolled-back publication failure %d was incorrectly replayed", failOn)
			}
			requireDecisionSupersessionChangeSetEffects(t, pool, retry.ChangeSetID, source, target)
		})
	}
}

type decisionSupersessionEffects struct {
	changeSets int
	revisions  int
	intents    int
}

func requireDecisionSupersessionChangeSetEffects(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, changeSetID uuid.UUID, sourceRecordID uuid.UUID, targetRecordID uuid.UUID) {
	t.Helper()
	var effects decisionSupersessionEffects
	if err := db.QueryRow(context.Background(), `
SELECT
    (SELECT count(*) FROM change_sets WHERE change_set_id = $1),
    (SELECT count(*) FROM record_revisions WHERE change_set_id = $1)
`, changeSetID).Scan(&effects.changeSets, &effects.revisions); err != nil {
		t.Fatalf("query decision supersession change-set effects: %v", err)
	}
	effects.intents = collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{SourceChangeSetID: changeSetID.String()})
	if effects != (decisionSupersessionEffects{changeSets: 1, revisions: 2, intents: 2}) {
		t.Fatalf("decision supersession change-set effects = %+v; want one change set, two revisions, and two intents", effects)
	}
	orderedRecordIDs := collaborationsupport.OrderedRecordChangeIntentRecordIDs(t, db, changeSetID.String())
	wantRecordIDs := []string{sourceRecordID.String(), targetRecordID.String()}
	if !slices.Equal(orderedRecordIDs, wantRecordIDs) {
		t.Fatalf("decision supersession intent order = %v; want source then target %v", orderedRecordIDs, wantRecordIDs)
	}
}

func decisionSupersessionIncidentEffects(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, incidentID uuid.UUID) decisionSupersessionEffects {
	t.Helper()
	var effects decisionSupersessionEffects
	if err := db.QueryRow(context.Background(), `
SELECT
    (SELECT count(*) FROM change_sets WHERE incident_id = $1),
    (SELECT count(*) FROM record_revisions revision
       JOIN records record USING (record_id)
      WHERE record.incident_id = $1)
`, incidentID).Scan(&effects.changeSets, &effects.revisions); err != nil {
		t.Fatalf("query decision supersession incident effects: %v", err)
	}
	effects.intents = collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{IncidentID: incidentID.String()})
	return effects
}

func TestSupersedeDecisionRejectsInconsistentSourceOrTarget_Unit(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-decision-supersede-inconsistent")
	codec := conflicttest.NewCodec("workbook")
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-decision-supersede-inconsistent@example.test", "TaskDecision Decision Supersede Inconsistent", "TaskDecisionDecisionSupersede1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-decision-supersede-inconsistent-incident", "IR-TASK-DECISION-DECISION-SUPERSEDE-INCONSISTENT", "Workbook inspector task and decision workflow decision supersede inconsistent")

	inconsistentSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-source", "proposed", "Inconsistent source")
	sourceExistingTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-source-existing-target", "proposed", "Existing target")
	validTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-source-fail-target", "proposed", "Source fail target")
	insertSupersedesLink(t, harness.DB, incident.ID, inconsistentSource, sourceExistingTarget, actor.ID, testTime(time.Hour))

	sourceBefore := decisionSnapshot(t, harness.DB, inconsistentSource)
	sourceExistingTargetBefore := decisionSnapshot(t, harness.DB, sourceExistingTarget)
	validTargetBefore := decisionSnapshot(t, harness.DB, validTarget)
	sourceRequest := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      validTargetBefore.RowVersion,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-inconsistent-source-route",
		Reason:              "Attempt explicit supersession with inconsistent source.",
		ReplacementRecordID: &inconsistentSource,
	}
	_, err := supersedeDecision(ctx, owner, actor, validTarget, sourceRequest, "req-workbook_interaction-task-decision-decision-inconsistent-source-route", testTime(2*time.Hour))
	requireLifecycle(t, err)
	requireDecisionSnapshot(t, decisionSnapshot(t, harness.DB, inconsistentSource), sourceBefore, "inconsistent source supersede")
	requireDecisionSnapshot(t, decisionSnapshot(t, harness.DB, sourceExistingTarget), sourceExistingTargetBefore, "inconsistent source existing target")
	requireDecisionSnapshot(t, decisionSnapshot(t, harness.DB, validTarget), validTargetBefore, "valid target rejected by inconsistent source")
	if got := countSupersedesLinks(t, harness.DB, inconsistentSource, validTarget); got != 0 {
		t.Fatalf("inconsistent source supersede wrote link to attempted target: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, inconsistentSource, "decision.support_refs"); got != 0 {
		t.Fatalf("inconsistent source supersede wrote support links: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, validTarget, "decision.affected_record_ids"); got != 0 {
		t.Fatalf("inconsistent source supersede wrote affected links: got %d want 0", got)
	}

	validSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-target-fail-source", "approved", "Valid source")
	inconsistentTarget := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-target", "proposed", "Inconsistent target")
	targetExistingSource := mustCreateDecision(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-inconsistent-target-existing-source", "approved", "Existing superseding source")
	insertSupersedesLink(t, harness.DB, incident.ID, targetExistingSource, inconsistentTarget, actor.ID, testTime(3*time.Hour))

	validSourceBefore := decisionSnapshot(t, harness.DB, validSource)
	inconsistentTargetBefore := decisionSnapshot(t, harness.DB, inconsistentTarget)
	targetExistingSourceBefore := decisionSnapshot(t, harness.DB, targetExistingSource)
	targetRequest := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      inconsistentTargetBefore.RowVersion,
		ClientTxnID:         "txn-workbook_interaction-task-decision-decision-inconsistent-target-route",
		Reason:              "Attempt explicit supersession against inconsistent target.",
		ReplacementRecordID: &validSource,
	}
	_, err = supersedeDecision(ctx, owner, actor, inconsistentTarget, targetRequest, "req-workbook_interaction-task-decision-decision-inconsistent-target-route", testTime(4*time.Hour))
	requireLifecycle(t, err)
	requireDecisionSnapshot(t, decisionSnapshot(t, harness.DB, validSource), validSourceBefore, "valid source rejected by inconsistent target")
	requireDecisionSnapshot(t, decisionSnapshot(t, harness.DB, inconsistentTarget), inconsistentTargetBefore, "inconsistent target supersede")
	requireDecisionSnapshot(t, decisionSnapshot(t, harness.DB, targetExistingSource), targetExistingSourceBefore, "inconsistent target existing source")
	if got := countSupersedesLinks(t, harness.DB, validSource, inconsistentTarget); got != 0 {
		t.Fatalf("inconsistent target supersede wrote attempted link: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, validSource, "decision.support_refs"); got != 0 {
		t.Fatalf("inconsistent target supersede wrote support links: got %d want 0", got)
	}
	if got := countReferenceLinks(t, harness.DB, inconsistentTarget, "decision.affected_record_ids"); got != 0 {
		t.Fatalf("inconsistent target supersede wrote affected links: got %d want 0", got)
	}
}

func supersedeDecision(
	ctx context.Context,
	owner *tasksdecisions.MutationFacade,
	actor authn.UserRecord,
	targetRecordID uuid.UUID,
	request tasksdecisions.SupersedeRequest,
	requestID string,
	now time.Time,
) (tasksdecisions.SupersedeMutationResult, error) {
	return owner.SupersedeDecision(ctx, tasksdecisions.SupersedeCommand{
		ActorUserID: actor.ID, TargetRecordID: targetRecordID, Request: request,
		RequestHash: tasksdecisions.SupersedeRequestHash(request), RequestID: requestID,
		RouteKey: "workbook.records.supersede", Now: now,
	})
}
