package tasksdecisions_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/tasksdecisionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

const (
	tasksBundlePath     = "data/task_requests.ndjson"
	decisionsBundlePath = "data/decisions.ndjson"
)

type tasksDecisionsPortabilityHarness struct {
	db         dbport.DB
	actor      authn.UserRecord
	incidentID uuid.UUID
}

func TestIncidentBundleTasksDecisionsRoundTrip(t *testing.T) {
	t.Run("microsecond source clocks", requireTaskDecisionMicrosecondRoundTrip)
	t.Run("deleted_sources", func(t *testing.T) {
		ctx := context.Background()
		harness := newTasksDecisionsPortabilityHarness(t, "deleted-sources")
		now := taskDecisionPortableTime()
		taskID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
		decisionID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
		for _, fixture := range []struct {
			id   uuid.UUID
			kind string
		}{{taskID, "task_request"}, {decisionID, "decision"}} {
			seedTaskDecisionEnvelope(t, harness, fixture.id, fixture.kind, now)
			if _, err := harness.db.Exec(ctx, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, fixture.id, now, harness.actor.ID); err != nil {
				t.Fatalf("mark source envelope deleted: %v", err)
			}
		}
		task := validPortableTask(harness, taskID, now)
		decision := validPortableDecision(harness, decisionID, now)
		port := newTaskDecisionIncidentBundleSourcePort(t)
		importContext := taskDecisionImportContext(harness, "deleted-sources")
		prepared, err := port.PrepareImport(ctx, taskDecisionBundle([]map[string]any{task}, []map[string]any{decision}), importContext)
		if err != nil {
			t.Fatalf("prepare deleted sources: %v", err)
		}
		tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin deleted-source transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
			t.Fatalf("apply deleted sources: %v", err)
		}
		if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
			t.Fatalf("validate legal deleted sources: %v", err)
		}
		files, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{
			Query: tx, IncidentID: harness.incidentID, PortableAttributions: incidentbundles.PortableSourceAttributions(),
		})
		if err != nil {
			t.Fatalf("export deleted sources: %v", err)
		}
		bundle := filesToTaskDecisionBundle(t, files)
		expected := map[string]map[string]any{tasksBundlePath: task, decisionsBundlePath: decision}
		for path, row := range expected {
			row["created_at"] = "2026-07-31T13:00:00.123+00:00"
			row["updated_at"] = "2026-07-31T13:00:00.123+00:00"
			if path == decisionsBundlePath {
				row["decided_at"] = "2026-07-31T13:00:00.123+00:00"
			}
			rows, err := incidentportability.DecodeNDJSON(bundle[path])
			if err != nil || !reflect.DeepEqual(rows, []map[string]any{row}) {
				t.Fatalf("deleted source %s = %#v, err=%v; want %#v", path, rows, err, row)
			}
		}
		for _, id := range []uuid.UUID{taskID, decisionID} {
			var deletedAt time.Time
			var deletedBy uuid.UUID
			if err := tx.QueryRow(ctx, `SELECT deleted_at, deleted_by_user_id FROM records WHERE record_id = $1`, id).Scan(&deletedAt, &deletedBy); err != nil || !deletedAt.Equal(now) || deletedBy != harness.actor.ID {
				t.Fatalf("deleted envelope %s = (%v, %s), err=%v", id, deletedAt, deletedBy, err)
			}
		}
	})
	ctx := context.Background()
	harness := newTasksDecisionsPortabilityHarness(t, "round-trip")
	now := time.Date(2026, time.July, 31, 12, 0, 0, 123000000, time.UTC)
	decisionID := uuid.MustParse("00000000-0000-4000-8000-000000000003")
	taskID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	canceledTaskID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
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
	exported, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{
		Query: harness.db, IncidentID: harness.incidentID, PortableAttributions: incidentbundles.PortableSourceAttributions(),
	})
	if err != nil {
		t.Fatalf("export tasks/decisions: %v", err)
	}
	bundle := filesToTaskDecisionBundle(t, exported)
	expected := map[string][]map[string]any{
		tasksBundlePath: {
			{
				"record_id": taskID.String(), "incident_id": harness.incidentID.String(),
				"title": "Portable task", "status": "open", "owner_user_id": harness.actor.ID.String(),
				"priority": "normal", "task_kind": "follow_up", "workstream": nil,
				"due_at": nil, "requester_party_text": nil, "requester_party_id": nil,
				"blocked_reason": nil, "completed_at": nil, "external_ticket_ref": nil,
				"closure_summary": nil, "decision_record_id": decisionID.String(),
				"created_at": "2026-07-31T12:00:00.123+00:00", "updated_at": "2026-07-31T12:00:00.123+00:00",
			},
			{
				"record_id": canceledTaskID.String(), "incident_id": harness.incidentID.String(),
				"title": "Canceled unowned task", "status": "canceled", "owner_user_id": nil,
				"priority": "normal", "task_kind": "request", "workstream": nil,
				"due_at": nil, "requester_party_text": nil, "requester_party_id": nil,
				"blocked_reason": nil, "completed_at": nil, "external_ticket_ref": nil,
				"closure_summary": nil, "decision_record_id": nil,
				"created_at": "2026-07-31T12:00:00.123+00:00", "updated_at": "2026-07-31T12:00:00.123+00:00",
			},
		},
		decisionsBundlePath: {{
			"record_id": decisionID.String(), "incident_id": harness.incidentID.String(),
			"summary": "Portable decision", "status": "approved", "owner_user_id": harness.actor.ID.String(),
			"decision_type": "scope", "decided_at": "2026-07-31T12:00:00.123+00:00",
			"rationale": "Portable rationale", "created_at": "2026-07-31T12:00:00.123+00:00",
			"updated_at": "2026-07-31T12:00:00.123+00:00",
		}},
	}
	requireSourceRows := func(bundle portabilitysupport.MapBundle) {
		t.Helper()
		for _, path := range []string{tasksBundlePath, decisionsBundlePath} {
			rows, err := incidentportability.DecodeNDJSON(bundle[path])
			if err != nil || !reflect.DeepEqual(rows, expected[path]) {
				t.Fatalf("exported %s = %#v, err=%v; want %#v", path, rows, err, expected[path])
			}
		}
	}
	requireSourceRows(bundle)
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
	reexported, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{
		Query: tx, IncidentID: harness.incidentID, PortableAttributions: incidentbundles.PortableSourceAttributions(),
	})
	if err != nil {
		t.Fatalf("re-export tasks/decisions: %v", err)
	}
	reexportedBundle := filesToTaskDecisionBundle(t, reexported)
	requireSourceRows(reexportedBundle)
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
	t.Run("retained Decision endpoint lifecycle", requirePortableDecisionEndpointLifecycle)
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
	t.Run("retained Decision associations", requirePortableTaskDecisionBindings)
	requireTaskDecisionPortableStoredText(t)
	requireTaskDecisionPortableTimestampAdmission(t)
	for _, matches := range []bool{false, true} {
		t.Run(fmt.Sprintf("decision_field_relation_matches_%t", matches), func(t *testing.T) {
			harness := newTasksDecisionsPortabilityHarness(t, "decision-field-relation")
			now := taskDecisionPortableTime()
			taskID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
			decisionID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
			otherID := uuid.MustParse("00000000-0000-4000-8000-000000000003")
			seedTaskDecisionEnvelope(t, harness, taskID, "task_request", now)
			seedTaskDecisionEnvelope(t, harness, decisionID, "decision", now)
			seedTaskDecisionEnvelope(t, harness, otherID, "decision", now)
			linkedID := otherID
			if matches {
				linkedID = decisionID
			}
			if _, err := harness.db.Exec(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'references_record', 'task.decision_record_id', 'manual', NULL, $4, $4, $5, $5)
`, harness.incidentID, taskID, linkedID, harness.actor.ID, now); err != nil {
				t.Fatalf("seed authoritative Decision field link: %v", err)
			}
			row := validPortableTask(harness, taskID, now)
			row["decision_record_id"] = decisionID.String()
			err := applyAndValidateTaskDecisionBundle(t, harness, taskDecisionBundle(
				[]map[string]any{row},
				[]map[string]any{validPortableDecision(harness, decisionID, now), validPortableDecision(harness, otherID, now)},
			), "decision-field-relation")
			if matches {
				if err != nil {
					t.Fatalf("matching Decision scalar and relation rejected: %v", err)
				}
			} else {
				requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
			}
		})
	}
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
	t.Run("retained direct-reference lifecycle", requirePortableTaskReferenceLifecycle)
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
	for _, envelopeFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("canonical_envelope_first_%t", envelopeFirst), func(t *testing.T) {
			ctx := context.Background()
			harness := newTasksDecisionsPortabilityHarness(t, fmt.Sprintf("canonical-order-%t", envelopeFirst))
			now := taskDecisionPortableTime()
			ids := []uuid.UUID{
				uuid.MustParse("00000000-0000-4000-8000-000000000001"),
				uuid.MustParse("00000000-0000-4000-8000-000000000002"),
			}
			var rows, decisions []map[string]any
			for i, id := range ids {
				row := validPortableTask(harness, id, now)
				kind := "task_request"
				if (i == 0) == envelopeFirst {
					kind = "decision"
				}
				seedTaskDecisionEnvelope(t, harness, id, kind, now)
				if kind == "task_request" {
					decisionID, decisionRows := seedTaskDecisionFieldMismatch(t, harness, id, now)
					row["decision_record_id"] = decisionID.String()
					decisions = decisionRows
				}
				rows = append(rows, row)
			}
			port := newTaskDecisionIncidentBundleSourcePort(t)
			importContext := taskDecisionImportContext(harness, "canonical-order")
			prepared, err := port.PrepareImport(ctx, taskDecisionBundle(rows, decisions), importContext)
			if err != nil {
				t.Fatalf("prepare canonically ordered competing defects: %v", err)
			}
			tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin competing-defects transaction: %v", err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
				t.Fatalf("apply competing defects before validation: %v", err)
			}
			err = port.ValidateImportTx(ctx, tx, prepared, importContext)
			requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.envelope_type_scope")
		})
	}

}

func TestIncidentBundleTasksDecisionsSourceTransactionRollback(t *testing.T) {
	ctx := context.Background()
	harness := newTasksDecisionsPortabilityHarness(t, "atomic")
	now := taskDecisionPortableTime()
	recordID := uuid.New()
	seedTaskDecisionEnvelope(t, harness, recordID, "task_request", now)
	row := validPortableTask(harness, recordID, now)
	decisionID, decisions := seedTaskDecisionFieldMismatch(t, harness, recordID, now)
	row["decision_record_id"] = decisionID.String()
	port := newTaskDecisionIncidentBundleSourcePort(t)
	importContext := taskDecisionImportContext(harness, "atomic")
	prepared, err := port.PrepareImport(ctx, taskDecisionBundle([]map[string]any{row}, decisions), importContext)
	if err != nil {
		t.Fatalf("prepare source rollback fixture: %v", err)
	}
	tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin source rollback transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatalf("apply source rollback fixture: %v", err)
	}
	var count int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM task_requests WHERE record_id = $1`, recordID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("inserted task source rows = %d, err=%v; want one", count, err)
	}
	err = port.ValidateImportTx(ctx, tx, prepared, importContext)
	requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("roll back rejected task source: %v", err)
	}
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
	actor := authstoretest.SeedActor(t, storeHarness.DB, authstoretest.ActorAttributes{Email: "tasks-decisions-portability-" + uuid.NewString() + "@example.test", DisplayName: "Tasks Decisions Portability"})
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
		BundleVersion: 8,
		OperationID:   "tasks-decisions-" + operationID,
		Actors:        actors,
		Attributions:  &taskDecisionAttributions{local: harness.actor.ID},
	}
}

func seedTaskDecisionFieldMismatch(t testing.TB, harness tasksDecisionsPortabilityHarness, taskID uuid.UUID, now time.Time) (uuid.UUID, []map[string]any) {
	t.Helper()
	decisionID := uuid.MustParse("00000000-0000-4000-8000-000000000003")
	otherID := uuid.MustParse("00000000-0000-4000-8000-000000000004")
	seedTaskDecisionEnvelope(t, harness, decisionID, "decision", now)
	seedTaskDecisionEnvelope(t, harness, otherID, "decision", now)
	if _, err := harness.db.Exec(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'references_record', 'task.decision_record_id', 'manual', NULL, $4, $4, $5, $5)
`, harness.incidentID, taskID, otherID, harness.actor.ID, now); err != nil {
		t.Fatalf("seed mismatched authoritative Decision field link: %v", err)
	}
	return decisionID, []map[string]any{validPortableDecision(harness, decisionID, now), validPortableDecision(harness, otherID, now)}
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

func taskDecisionBundle(tasks []map[string]any, decisions []map[string]any) portabilitysupport.MapBundle {
	return portabilitysupport.MapBundle{
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

func filesToTaskDecisionBundle(t testing.TB, files []portabilitysupport.File) portabilitysupport.MapBundle {
	t.Helper()
	if len(files) != 2 {
		t.Fatalf("Tasks/Decisions export returned %d files, want exactly two", len(files))
	}
	bundle := portabilitysupport.MapBundle{}
	for _, file := range files {
		if file.Path != tasksBundlePath && file.Path != decisionsBundlePath {
			t.Fatalf("unexpected Tasks/Decisions export path %q", file.Path)
		}
		if _, duplicate := bundle[file.Path]; duplicate {
			t.Fatalf("duplicate Tasks/Decisions export path %q", file.Path)
		}
		bundle[file.Path] = append([]byte(nil), file.Payload...)
	}
	for _, path := range []string{tasksBundlePath, decisionsBundlePath} {
		if _, ok := bundle[path]; !ok {
			t.Fatalf("export omitted %s", path)
		}
	}
	return bundle
}

func applyAndValidateTaskDecisionBundle(t testing.TB, harness tasksDecisionsPortabilityHarness, bundle portabilitysupport.MapBundle, operationID string) error {
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
	port, err := tasksdecisions.NewIncidentBundleSourcePort(tasksdecisionassembly.NewLinkFactsCapability(), tasksdecisionassembly.NewFieldReferenceReader(), tasksdecisionassembly.NewEnvelopeFactReader())
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

func requireTaskDecisionPortableStoredText(t *testing.T) {
	t.Helper()
	harness := tasksDecisionsPortabilityHarness{incidentID: uuid.New(), actor: authn.UserRecord{ID: uuid.New()}}
	port := newTaskDecisionIncidentBundleSourcePort(t)
	importContext := taskDecisionImportContext(harness, "stored-text")
	fields := []struct {
		kind, field         string
		limit               int
		optional, multiline bool
	}{
		{"task_request", "title", 512, false, false},
		{"task_request", "workstream", 512, true, false},
		{"task_request", "requester_party_text", 256, true, false},
		{"task_request", "blocked_reason", 4096, true, true},
		{"task_request", "external_ticket_ref", 1024, true, false},
		{"task_request", "closure_summary", 16384, true, true},
		{"decision", "summary", 512, false, false},
		{"decision", "rationale", 16384, false, true},
	}
	for _, field := range fields {
		t.Run("portable stored text "+field.kind+" "+field.field, func(t *testing.T) {
			check := func(value any, valid bool) {
				t.Helper()
				row := validPortableTask(harness, uuid.MustParse("00000000-0000-4000-8000-000000000001"), taskDecisionPortableTime())
				if field.kind == "decision" {
					row = validPortableDecision(harness, uuid.MustParse("00000000-0000-4000-8000-000000000002"), taskDecisionPortableTime())
				}
				row[field.field] = value
				if field.field == "blocked_reason" && value != nil {
					row["status"] = "blocked"
				}
				bundle := taskDecisionBundle([]map[string]any{row}, nil)
				if field.kind == "decision" {
					bundle = taskDecisionBundle(nil, []map[string]any{row})
				}
				before := maps.Clone(bundle)
				for path, payload := range before {
					before[path] = bytes.Clone(payload)
				}
				_, err := port.PrepareImport(context.Background(), bundle, importContext)
				if valid {
					if err != nil {
						t.Fatalf("canonical %s rejected: %v", field.field, err)
					}
				} else {
					requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
				}
				if !reflect.DeepEqual(bundle, before) {
					t.Fatal("portable source was rewritten")
				}
			}
			check(nil, field.optional)
			for _, valid := range []string{"Café  MiXeD\u200dText", strings.Repeat("😀", field.limit), strings.Repeat("é", field.limit)} {
				check(valid, true)
			}
			for _, invalid := range []string{"", " ", " leading", "trailing\u00a0", "Cafe\u0301", strings.Repeat("😀", field.limit+1), "A\r\nB", "A\rB"} {
				check(invalid, false)
			}
			check("A\n\t B", field.multiline)
			for control := rune(0); control <= 159; control++ {
				if control > 31 && control < 127 {
					continue
				}
				if field.multiline && (control == '\n' || control == '\t') {
					continue
				}
				for _, value := range []string{string(control) + "Text", "Te" + string(control) + "xt", "Text" + string(control)} {
					check(value, false)
				}
			}
		})
	}
}

func requireTaskDecisionPortableTimestampAdmission(t *testing.T) {
	t.Helper()
	harness := tasksDecisionsPortabilityHarness{incidentID: uuid.New(), actor: authn.UserRecord{ID: uuid.New()}}
	port := newTaskDecisionIncidentBundleSourcePort(t)
	importContext := taskDecisionImportContext(harness, "timestamp-admission")
	for _, test := range []struct {
		kind, field string
		optional    bool
	}{
		{"task_request", "due_at", true}, {"task_request", "completed_at", true},
		{"task_request", "created_at", false}, {"task_request", "updated_at", false},
		{"decision", "decided_at", false}, {"decision", "created_at", false}, {"decision", "updated_at", false},
	} {
		t.Run(test.kind+" timestamp "+test.field, func(t *testing.T) {
			check := func(value any, valid bool) {
				t.Helper()
				row := validPortableTask(harness, uuid.MustParse("00000000-0000-4000-8000-000000000001"), taskDecisionPortableTime())
				if test.kind == "decision" {
					row = validPortableDecision(harness, uuid.MustParse("00000000-0000-4000-8000-000000000002"), taskDecisionPortableTime())
				}
				// Isolate each lexical check from unrelated source-clock ordering.
				row["created_at"], row["updated_at"] = "0001-01-01T00:00:00Z", "9999-12-31T23:59:59.999999Z"
				row[test.field] = value
				if test.field == "completed_at" && value != nil {
					row["status"] = "done"
				}
				bundle := taskDecisionBundle([]map[string]any{row}, nil)
				if test.kind == "decision" {
					bundle = taskDecisionBundle(nil, []map[string]any{row})
				}
				before := maps.Clone(bundle)
				for path, payload := range before {
					before[path] = bytes.Clone(payload)
				}
				_, err := port.PrepareImport(context.Background(), bundle, importContext)
				if valid {
					if err != nil {
						t.Fatalf("timestamp %s rejected: %v", test.field, err)
					}
				} else {
					requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.dependent_fields_legal")
				}
				if !reflect.DeepEqual(bundle, before) {
					t.Fatal("portable timestamps changed")
				}
			}
			check(nil, test.optional)
			for _, value := range []string{"0001-01-01T00:00:00Z", "9999-12-31T23:59:59.999999Z", "2026-01-01T00:00:00.000001+00:00", "2026-01-01T00:00:00.123456Z", "2026-01-01T00:00:00.1Z"} {
				check(value, true)
			}
			for _, value := range []any{"0000-01-01T00:00:00Z", "10000-01-01T00:00:00Z", "2026-01-01T00:00:00.1234567Z", "2026-01-01T00:00:00.123456789Z", "2026-01-01T00:00:00.1234560Z", "2026-01-01T00:00:00.100000Z", "2026-01-01T00:00:00+01:00", "2026-01-01T00:00:00", "2026-01-01", "", 1, true} {
				check(value, false)
			}
		})
	}
}

func requireTaskDecisionMicrosecondRoundTrip(t *testing.T) {
	ctx := context.Background()
	harness := newTasksDecisionsPortabilityHarness(t, "microsecond-source")
	taskID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	decisionID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	now := taskDecisionPortableTime()
	seedTaskDecisionEnvelope(t, harness, taskID, "task_request", now)
	seedTaskDecisionEnvelope(t, harness, decisionID, "decision", now)
	task := validPortableTask(harness, taskID, now)
	task["status"] = "done"
	task["created_at"], task["updated_at"] = "2026-07-31T13:00:00.123451Z", "2026-07-31T13:00:00.123456Z"
	task["due_at"], task["completed_at"] = "2020-01-01T00:00:00.123452Z", "2026-07-31T13:00:00.123453Z"
	decision := validPortableDecision(harness, decisionID, now)
	decision["decided_at"] = "1999-01-01T00:00:00.123454Z"
	decision["created_at"], decision["updated_at"] = "2026-07-31T13:00:00.123455Z", "2026-07-31T13:00:00.123456Z"
	port := newTaskDecisionIncidentBundleSourcePort(t)
	importContext := taskDecisionImportContext(harness, "microseconds")
	prepared, err := port.PrepareImport(ctx, taskDecisionBundle([]map[string]any{task}, []map[string]any{decision}), importContext)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := port.ApplyImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatal(err)
	}
	if err := port.ValidateImportTx(ctx, tx, prepared, importContext); err != nil {
		t.Fatal(err)
	}
	files, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: harness.incidentID, PortableAttributions: incidentbundles.PortableSourceAttributions()})
	if err != nil {
		t.Fatal(err)
	}
	bundle := filesToTaskDecisionBundle(t, files)
	expectedTask, expectedDecision := maps.Clone(task), maps.Clone(decision)
	expectedTask["created_at"], expectedTask["updated_at"] = "2026-07-31T13:00:00.123451+00:00", "2026-07-31T13:00:00.123456+00:00"
	expectedTask["due_at"], expectedTask["completed_at"] = "2020-01-01T00:00:00.123452+00:00", "2026-07-31T13:00:00.123453+00:00"
	expectedDecision["decided_at"] = "1999-01-01T00:00:00.123454+00:00"
	expectedDecision["created_at"], expectedDecision["updated_at"] = "2026-07-31T13:00:00.123455+00:00", "2026-07-31T13:00:00.123456+00:00"
	for path, expected := range map[string]map[string]any{tasksBundlePath: expectedTask, decisionsBundlePath: expectedDecision} {
		rows, err := incidentportability.DecodeNDJSON(bundle[path])
		if err != nil || !reflect.DeepEqual(rows, []map[string]any{expected}) {
			t.Fatalf("microsecond export %s=%#v err=%v want=%#v", path, rows, err, expected)
		}
	}
}
