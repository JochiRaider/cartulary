package tasksdecisions_test

import (
	"context"
	"maps"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

type taskDecisionAttributions struct {
	portabilitysupport.AttributionCapability
	local uuid.UUID
	rows  []incidentportability.ImportedAttribution
}

func TestIncidentBundleTasksDecisionsExactAttribution(t *testing.T) {
	ctx := t.Context()
	harness := newTasksDecisionsPortabilityHarness(t, "exact-attribution")
	now := taskDecisionPortableTime()
	taskID, decisionID, sourceActor := uuid.New(), uuid.New(), uuid.New()
	seedTaskDecisionEnvelope(t, harness, taskID, "task_request", now)
	seedTaskDecisionEnvelope(t, harness, decisionID, "decision", now)
	task := validPortableTask(harness, taskID, now)
	task["owner_user_id"] = sourceActor.String()
	decision := validPortableDecision(harness, decisionID, now)
	decision["owner_user_id"] = sourceActor.String()
	port := newTaskDecisionIncidentBundleSourcePort(t)
	prepare := func(t testing.TB, tasks, decisions []map[string]any) (pgx.Tx, sourceport.Prepared, sourceport.ImportContext, *taskDecisionAttributions) {
		t.Helper()
		input := taskDecisionImportContext(harness, "exact-attribution")
		var err error
		input.Actors, err = sourceport.NewActorCatalog([]sourceport.ActorDescriptor{{SourceActorID: sourceActor.String(), DisplayName: "Historical owner"}})
		if err != nil {
			t.Fatal(err)
		}
		prepared, err := port.PrepareImport(ctx, taskDecisionBundle(tasks, decisions), input)
		if err != nil {
			t.Fatal(err)
		}
		tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx, prepared, input, input.Attributions.(*taskDecisionAttributions)
	}
	// Core 01 REQ-01-641: attribution is staged while preparing and flushed by
	// the core in step 9; the port records none row by row.
	want := []incidentportability.AttributionRow{
		{SourceTable: "task_requests", SourceRowID: taskID.String(), SourceColumn: "owner_user_id", SourceActorID: sourceActor.String()},
		{SourceTable: "decisions", SourceRowID: decisionID.String(), SourceColumn: "owner_user_id", SourceActorID: sourceActor.String()},
	}
	for _, variant := range []string{"both owners", "null task owner", "empty"} {
		t.Run(variant, func(t *testing.T) {
			taskRow := maps.Clone(task)
			tasks, decisions := []map[string]any{taskRow}, []map[string]any{decision}
			expected := want
			if variant == "null task owner" {
				taskRow["status"], taskRow["owner_user_id"] = "canceled", nil
				expected = want[1:]
			}
			if variant == "empty" {
				tasks, decisions, expected = nil, nil, nil
			}
			tx, prepared, input, attrs := prepare(t, tasks, decisions)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			staged, err := portabilitysupport.ReadStagedAttributions(ctx, port, prepared, input)
			if err != nil || !reflect.DeepEqual(staged, expected) || len(attrs.rows) != 0 {
				t.Fatalf("staged=%#v want=%#v recorded=%#v err=%v", staged, expected, attrs.rows, err)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("source drift precedes attribution drift", func(t *testing.T) {
		tx, prepared, input, _ := prepare(t, []map[string]any{task}, []map[string]any{decision})
		if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET title='Changed' WHERE record_id=$1`, taskID); err != nil {
			t.Fatal(err)
		}
		input.Attributions = nil
		requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.source_state_exact")
	})
	for _, variant := range []string{"nil before apply", "nil empty", "nil before validate", "nil empty before validate"} {
		t.Run(variant, func(t *testing.T) {
			tasks, decisions := []map[string]any{task}, []map[string]any{decision}
			if variant == "nil empty" || variant == "nil empty before validate" {
				tasks, decisions = nil, nil
			}
			tx, prepared, input, _ := prepare(t, tasks, decisions)
			if variant == "nil before apply" || variant == "nil empty" {
				input.Attributions = nil
			}
			err := port.ApplyImportTx(ctx, tx, prepared, input)
			if variant == "nil before apply" || variant == "nil empty" {
				requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.attribution_exact")
			} else {
				if err != nil {
					t.Fatal(err)
				}
				input.Attributions = nil
				requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.attribution_exact")
				return
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM task_requests WHERE incident_id=$1)+(SELECT count(*) FROM decisions WHERE incident_id=$1)`, harness.incidentID).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if count != 0 {
				t.Fatalf("rejected attribution admission wrote %d source rows", count)
			}
		})
	}
}
