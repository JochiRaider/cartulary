package tasksdecisions_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/tasksdecisionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

const (
	tasksBundlePath     = "data/task_requests.ndjson"
	decisionsBundlePath = "data/decisions.ndjson"
)

type tasksDecisionsPortabilityHarness struct {
	db         postgres.DB
	actor      authn.UserRecord
	incidentID uuid.UUID
}

func TestIncidentBundleTasksDecisionsRoundTrip(t *testing.T) {
	ctx := context.Background()
	harness := newTasksDecisionsPortabilityHarness(t, "round-trip")
	now := time.Date(2026, time.July, 31, 12, 0, 0, 123000000, time.UTC)
	decisionID := uuid.New()
	taskID := uuid.New()
	canceledTaskID := uuid.New()
	seedTaskDecisionEnvelope(t, harness, decisionID, "decision", now)
	seedTaskDecisionEnvelope(t, harness, taskID, "task_request", now)
	seedTaskDecisionEnvelope(t, harness, canceledTaskID, "task_request", now)
	if _, err := harness.db.Exec(ctx, `
INSERT INTO decisions (
    record_id, incident_id, summary, status, owner_user_id, decision_type,
    decided_at, rationale, created_at, updated_at
) VALUES ($1, $2, 'Portable decision', 'approved', $3, 'scope', $4, 'Portable rationale', $4, $4)
`, decisionID, harness.incidentID, harness.actor.ID, now); err != nil {
		t.Fatalf("seed portable decision: %v", err)
	}
	if _, err := harness.db.Exec(ctx, `
INSERT INTO task_requests (
    record_id, incident_id, title, status, owner_user_id, priority, task_kind,
    decision_record_id, created_at, updated_at
) VALUES ($1, $2, 'Portable task', 'open', $3, 'normal', 'follow_up', $4, $5, $5)
`, taskID, harness.incidentID, harness.actor.ID, decisionID, now); err != nil {
		t.Fatalf("seed portable task: %v", err)
	}
	if _, err := harness.db.Exec(ctx, `
INSERT INTO task_requests (
    record_id, incident_id, title, status, owner_user_id, priority, task_kind,
    created_at, updated_at
) VALUES ($1, $2, 'Canceled unowned task', 'canceled', NULL, 'normal', 'request', $3, $3)
`, canceledTaskID, harness.incidentID, now); err != nil {
		t.Fatalf("seed portable canceled task: %v", err)
	}

	port := newTaskDecisionIncidentBundleSourcePort(t)
	exported, err := port.Export(ctx, sourceport.ExportContext{
		Query: harness.db, IncidentID: harness.incidentID,
	})
	if err != nil {
		t.Fatalf("export tasks/decisions: %v", err)
	}
	bundle := filesToTaskDecisionBundle(t, exported)
	if _, err := harness.db.Exec(ctx, `DELETE FROM task_requests WHERE incident_id = $1`, harness.incidentID); err != nil {
		t.Fatalf("clear exported tasks: %v", err)
	}
	if _, err := harness.db.Exec(ctx, `DELETE FROM decisions WHERE incident_id = $1`, harness.incidentID); err != nil {
		t.Fatalf("clear exported decisions: %v", err)
	}

	importContext := taskDecisionImportContext(harness, "round-trip")
	prepared, err := port.PrepareImport(ctx, bundle, importContext)
	if err != nil {
		t.Fatalf("prepare tasks/decisions round trip: %v", err)
	}
	tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin tasks/decisions round-trip transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("apply tasks/decisions round trip: %v", err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("validate tasks/decisions round trip: %v", err)
	}
	reexported, err := port.Export(ctx, sourceport.ExportContext{
		Query: tx, IncidentID: harness.incidentID,
	})
	if err != nil {
		t.Fatalf("re-export tasks/decisions: %v", err)
	}
	reexportedBundle := filesToTaskDecisionBundle(t, reexported)
	for _, path := range []string{tasksBundlePath, decisionsBundlePath} {
		if !bytes.Equal(bundle[path], reexportedBundle[path]) {
			t.Fatalf("round-trip payload changed for %s\noriginal=%s\nreexported=%s", path, bundle[path], reexportedBundle[path])
		}
	}
}

func TestIncidentBundleTasksDecisionsRejectsEnvelopeTypeScope(t *testing.T) {
	harness := newTasksDecisionsPortabilityHarness(t, "envelope")
	now := taskDecisionPortableTime()
	recordID := uuid.New()
	seedTaskDecisionEnvelope(t, harness, recordID, "decision", now)
	err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle(
		[]map[string]any{validPortableTask(harness, recordID, now)}, nil,
	), "envelope")
	requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.envelope_type_scope")
}

func TestIncidentBundleTasksDecisionsRejectsLifecycleIllegal(t *testing.T) {
	harness := newTasksDecisionsPortabilityHarness(t, "lifecycle")
	now := taskDecisionPortableTime()
	recordID := uuid.New()
	seedTaskDecisionEnvelope(t, harness, recordID, "decision", now)
	row := validPortableDecision(harness, recordID, now)
	row["status"] = "superseded"
	err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle(nil, []map[string]any{row}), "lifecycle")
	requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.lifecycle_legal")
}

func TestIncidentBundleTasksDecisionsRejectsDependentFieldsIllegal(t *testing.T) {
	harness := newTasksDecisionsPortabilityHarness(t, "dependent")
	now := taskDecisionPortableTime()
	recordID := uuid.New()
	seedTaskDecisionEnvelope(t, harness, recordID, "task_request", now)
	row := validPortableTask(harness, recordID, now)
	row["status"] = "blocked"
	row["blocked_reason"] = nil
	err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle([]map[string]any{row}, nil), "dependent")
	requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
}

func TestIncidentBundleTasksDecisionsRejectsReferencesOutsideIncident(t *testing.T) {
	harness := newTasksDecisionsPortabilityHarness(t, "references")
	now := taskDecisionPortableTime()
	taskID := uuid.New()
	seedTaskDecisionEnvelope(t, harness, taskID, "task_request", now)
	foreign := appsupport.CreateIncidentInStore(
		t,
		harness.db,
		harness.actor,
		"txn-tasks-decisions-portability-foreign-"+uuid.NewString(),
		"IR-TD-FOREIGN-"+strings.ToUpper(uuid.NewString()[:8]),
		"Tasks decisions portability foreign incident",
	)
	foreignDecisionID := uuid.New()
	seedEnvelopeForIncident(t, harness, foreign.ID, foreignDecisionID, "decision", now)
	row := validPortableTask(harness, taskID, now)
	row["decision_record_id"] = foreignDecisionID.String()
	err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle([]map[string]any{row}, nil), "references")
	requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.references_same_incident")
}

func TestIncidentBundleTasksDecisionsFailureSelectionIsOrderIndependent(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse_%t", reverse), func(t *testing.T) {
			harness := newTasksDecisionsPortabilityHarness(t, fmt.Sprintf("order-%t", reverse))
			now := taskDecisionPortableTime()
			envelopeID := uuid.New()
			dependentID := uuid.New()
			seedTaskDecisionEnvelope(t, harness, envelopeID, "decision", now)
			seedTaskDecisionEnvelope(t, harness, dependentID, "task_request", now)
			envelopeRow := validPortableTask(harness, envelopeID, now)
			dependentRow := validPortableTask(harness, dependentID, now)
			dependentRow["status"] = "blocked"
			dependentRow["blocked_reason"] = nil
			rows := []map[string]any{envelopeRow, dependentRow}
			if reverse {
				rows[0], rows[1] = rows[1], rows[0]
			}
			err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle(rows, nil), fmt.Sprintf("order-%t", reverse))
			requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.envelope_type_scope")
		})
	}
}

func TestIncidentBundleTasksDecisionsCoordinatorRejectsWithoutPartialPublication(t *testing.T) {
	harness := newTasksDecisionsPortabilityHarness(t, "atomic")
	now := taskDecisionPortableTime()
	recordID := uuid.New()
	seedTaskDecisionEnvelope(t, harness, recordID, "task_request", now)
	row := validPortableTask(harness, recordID, now)
	row["status"] = "done"
	row["completed_at"] = nil
	err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle([]map[string]any{row}, nil), "atomic")
	requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
	var count int
	if queryErr := harness.db.QueryRow(context.Background(), `SELECT count(*) FROM task_requests WHERE record_id = $1`, recordID).Scan(&count); queryErr != nil {
		t.Fatalf("query rolled-back task source: %v", queryErr)
	}
	if count != 0 {
		t.Fatalf("invalid import left %d task source rows", count)
	}
}

func TestIncidentBundleTasksDecisionsDiagnosticsAreSafe(t *testing.T) {
	t.Run("semantic_failure", func(t *testing.T) {
		harness := newTasksDecisionsPortabilityHarness(t, "diagnostics")
		now := taskDecisionPortableTime()
		recordID := uuid.New()
		seedTaskDecisionEnvelope(t, harness, recordID, "decision", now)
		row := validPortableTask(harness, recordID, now)
		row["title"] = "secret-value-should-never-escape"
		err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle([]map[string]any{row}, nil), "diagnostics")
		requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.envelope_type_scope")
		for _, forbidden := range []string{"secret-value-should-never-escape", "task_requests", "SELECT", tasksBundlePath} {
			if strings.Contains(err.Error(), forbidden) {
				t.Fatalf("safe portability failure exposed %q: %v", forbidden, err)
			}
		}
	})

	t.Run("strict_prepare", func(t *testing.T) {
		harness := newTasksDecisionsPortabilityHarness(t, "strict-prepare")
		now := taskDecisionPortableTime()
		recordID := uuid.New()
		row := validPortableTask(harness, recordID, now)
		row["hostile_unknown_member"] = "secret-value-should-never-escape"
		_, err := newTaskDecisionIncidentBundleSourcePort(t).PrepareImport(
			context.Background(),
			taskDecisionBundle([]map[string]any{row}, nil),
			taskDecisionImportContext(harness, "strict-prepare"),
		)
		if err == nil {
			t.Fatal("unknown task source member unexpectedly passed strict preparation")
		}
		if strings.Contains(err.Error(), "secret-value-should-never-escape") || strings.Contains(err.Error(), "hostile_unknown_member") {
			t.Fatalf("strict preparation exposed hostile input: %v", err)
		}
	})
}

func newTasksDecisionsPortabilityHarness(t testing.TB, suffix string) tasksDecisionsPortabilityHarness {
	t.Helper()
	storeHarness := appsupport.StartStore(t, "tasks-decisions-portability-"+suffix)
	actor := authstoretest.SeedLocalUserRecord(
		t,
		storeHarness.DB,
		"tasks-decisions-portability-"+uuid.NewString()+"@example.test",
		"Tasks Decisions Portability",
		"TasksDecisionsPortability1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		storeHarness.DB,
		actor,
		"txn-tasks-decisions-portability-"+uuid.NewString(),
		"IR-TD-"+strings.ToUpper(uuid.NewString()[:8]),
		"Tasks decisions portability "+suffix,
	)
	return tasksDecisionsPortabilityHarness{db: storeHarness.DB, actor: actor, incidentID: incident.ID}
}

func taskDecisionImportContext(harness tasksDecisionsPortabilityHarness, operationID string) sourceport.ImportContext {
	actors, _ := sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{
		SourceActorID: harness.actor.ID.String(),
		DisplayName:   "Portable actor",
	}})
	return sourceport.ImportContext{
		IncidentID:    harness.incidentID,
		ActorUserID:   harness.actor.ID,
		BundleVersion: 3,
		OperationID:   "tasks-decisions-" + operationID,
		Actors:        actors,
	}
}

func seedTaskDecisionEnvelope(t testing.TB, harness tasksDecisionsPortabilityHarness, recordID uuid.UUID, recordType string, now time.Time) {
	t.Helper()
	seedEnvelopeForIncident(t, harness, harness.incidentID, recordID, recordType, now)
}

func seedEnvelopeForIncident(t testing.TB, harness tasksDecisionsPortabilityHarness, incidentID uuid.UUID, recordID uuid.UUID, recordType string, now time.Time) {
	t.Helper()
	if _, err := harness.db.Exec(context.Background(), `
INSERT INTO records (
    record_id, incident_id, record_type,
    created_by_user_id, created_at, updated_by_user_id, updated_at, row_version
) VALUES ($1, $2, $3, $4, $5, $4, $5, 1)
`, recordID, incidentID, recordType, harness.actor.ID, now); err != nil {
		t.Fatalf("seed %s envelope: %v", recordType, err)
	}
}

func taskDecisionPortableTime() time.Time {
	return time.Date(2026, time.July, 31, 13, 0, 0, 123000000, time.UTC)
}

func validPortableTask(harness tasksDecisionsPortabilityHarness, recordID uuid.UUID, now time.Time) map[string]any {
	return map[string]any{
		"record_id": recordID.String(), "incident_id": harness.incidentID.String(),
		"title": "Portable task", "status": "open", "owner_user_id": harness.actor.ID.String(),
		"priority": "normal", "task_kind": "follow_up", "workstream": nil,
		"due_at": nil, "requester_party_text": nil, "requester_party_id": nil,
		"blocked_reason": nil, "completed_at": nil, "external_ticket_ref": nil,
		"closure_summary": nil, "decision_record_id": nil,
		"created_at": now.Format(time.RFC3339Nano), "updated_at": now.Format(time.RFC3339Nano),
	}
}

func validPortableDecision(harness tasksDecisionsPortabilityHarness, recordID uuid.UUID, now time.Time) map[string]any {
	return map[string]any{
		"record_id": recordID.String(), "incident_id": harness.incidentID.String(),
		"summary": "Portable decision", "status": "approved", "owner_user_id": harness.actor.ID.String(),
		"decision_type": "scope", "decided_at": now.Format(time.RFC3339Nano),
		"rationale": "Portable rationale", "created_at": now.Format(time.RFC3339Nano),
		"updated_at": now.Format(time.RFC3339Nano),
	}
}

func taskDecisionBundle(tasks []map[string]any, decisions []map[string]any) sourceport.MapBundle {
	return sourceport.MapBundle{
		tasksBundlePath:     encodeTaskDecisionRows(tasks),
		decisionsBundlePath: encodeTaskDecisionRows(decisions),
	}
}

func encodeTaskDecisionRows(rows []map[string]any) []byte {
	var payload []byte
	for _, row := range rows {
		encoded, err := incidentportability.CanonicalJSONString(row)
		if err != nil {
			panic(err)
		}
		payload = append(payload, encoded...)
	}
	return payload
}

func filesToTaskDecisionBundle(t testing.TB, files []incidentportability.File) sourceport.MapBundle {
	t.Helper()
	bundle := sourceport.MapBundle{}
	for _, file := range files {
		bundle[file.Path] = append([]byte(nil), file.Payload...)
	}
	for _, path := range []string{tasksBundlePath, decisionsBundlePath} {
		if _, ok := bundle[path]; !ok {
			t.Fatalf("export omitted %s", path)
		}
	}
	return bundle
}

func applyAndValidateTaskDecisionBundle(t testing.TB, harness tasksDecisionsPortabilityHarness, bundle sourceport.MapBundle, operationID string) error {
	t.Helper()
	ctx := context.Background()
	port := newTaskDecisionIncidentBundleSourcePort(t)
	importContext := taskDecisionImportContext(harness, operationID)
	prepared, err := port.PrepareImport(ctx, bundle, importContext)
	if err != nil {
		return err
	}
	tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin tasks/decisions portability transaction: %v", err)
	}
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	err = port.ValidateImportTx(ctx, tx, prepared, importContext)
	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
		t.Fatalf("roll back tasks/decisions portability transaction: %v", rollbackErr)
	}
	return err
}

func newTaskDecisionIncidentBundleSourcePort(t testing.TB) sourceport.Port {
	t.Helper()
	port, err := tasksdecisions.NewIncidentBundleSourcePort(tasksdecisionassembly.NewLinkFactsCapability())
	if err != nil {
		t.Fatalf("compose Tasks/Decisions source port: %v", err)
	}
	return port
}

func requireTasksDecisionsPortabilityFailure(t testing.TB, err error, invariantID string) {
	t.Helper()
	var failure *sourceport.Failure
	if !errors.As(err, &failure) || failure.FamilyID() != "tasks_decisions" || failure.InvariantID() != invariantID {
		t.Fatalf("tasks/decisions portability failure = %#v, %v; want %s", failure, err, invariantID)
	}
}
