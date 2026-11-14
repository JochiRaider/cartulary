package tasksdecisions_test

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type taskDecisionAttributions struct {
	local               uuid.UUID
	rows                []incidentportability.ImportedAttribution
	recordErr, visitErr error
}

func (a *taskDecisionAttributions) RecordImportedAttribution(table, row, column, actor string) error {
	if a.recordErr != nil {
		return a.recordErr
	}
	a.rows = append(a.rows, incidentportability.ImportedAttribution{SourceTable: table, SourceRowID: row, SourceColumn: column, SourceActorID: actor, LocalUserID: a.local})
	return nil
}

func (a *taskDecisionAttributions) VisitImportedAttributions(table, row string, visit func(incidentportability.ImportedAttribution) error) error {
	if a.visitErr != nil {
		return a.visitErr
	}
	for _, value := range a.rows {
		if value.SourceTable == table && (row == "" || value.SourceRowID == row) {
			if err := visit(value); err != nil {
				return err
			}
		}
	}
	return nil
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
	want := []incidentportability.ImportedAttribution{
		{SourceTable: "task_requests", SourceRowID: taskID.String(), SourceColumn: "owner_user_id", SourceActorID: sourceActor.String(), LocalUserID: harness.actor.ID},
		{SourceTable: "decisions", SourceRowID: decisionID.String(), SourceColumn: "owner_user_id", SourceActorID: sourceActor.String(), LocalUserID: harness.actor.ID},
	}
	for _, variant := range []string{"both owners", "null task owner", "empty", "null owner extra", "empty task extra", "empty decision extra"} {
		t.Run(variant, func(t *testing.T) {
			taskRow := maps.Clone(task)
			tasks, decisions := []map[string]any{taskRow}, []map[string]any{decision}
			expected := want
			if variant == "null task owner" || variant == "null owner extra" {
				taskRow["status"], taskRow["owner_user_id"] = "canceled", nil
				expected = want[1:]
			}
			if variant == "empty" || variant == "empty task extra" || variant == "empty decision extra" {
				tasks, decisions, expected = nil, nil, nil
			}
			tx, prepared, input, attrs := prepare(t, tasks, decisions)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(attrs.rows, expected) {
				t.Fatalf("sidecars=%#v want=%#v", attrs.rows, expected)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			switch variant {
			case "null owner extra", "empty task extra":
				attrs.rows = append(attrs.rows, want[0])
			case "empty decision extra":
				attrs.rows = append(attrs.rows, want[1])
			default:
				return
			}
			requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.attribution_exact")
		})
	}
	for _, tableIndex := range []int{0, 1} {
		for _, variant := range []string{"missing", "duplicate", "table", "row", "column", "source actor", "local actor", "extra row"} {
			t.Run(want[tableIndex].SourceTable+" "+variant, func(t *testing.T) {
				tx, prepared, input, attrs := prepare(t, []map[string]any{task}, []map[string]any{decision})
				if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
					t.Fatal(err)
				}
				if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
					t.Fatal(err)
				}
				switch variant {
				case "missing":
					attrs.rows = append(attrs.rows[:tableIndex], attrs.rows[tableIndex+1:]...)
				case "duplicate":
					attrs.rows = append(attrs.rows, attrs.rows[tableIndex])
				case "table":
					attrs.rows[tableIndex].SourceTable = "records"
				case "row":
					attrs.rows[tableIndex].SourceRowID = uuid.NewString()
				case "column":
					attrs.rows[tableIndex].SourceColumn = "created_by_user_id"
				case "source actor":
					attrs.rows[tableIndex].SourceActorID = uuid.NewString()
				case "local actor":
					attrs.rows[tableIndex].LocalUserID = uuid.New()
				case "extra row":
					extra := attrs.rows[tableIndex]
					extra.SourceRowID = uuid.NewString()
					attrs.rows = append(attrs.rows, extra)
				}
				requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.attribution_exact")
			})
		}
	}
	t.Run("source drift precedes attribution drift", func(t *testing.T) {
		tx, prepared, input, attrs := prepare(t, []map[string]any{task}, []map[string]any{decision})
		if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET title='Changed' WHERE record_id=$1`, taskID); err != nil {
			t.Fatal(err)
		}
		attrs.rows = nil
		requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.source_state_exact")
	})
	for _, variant := range []string{"nil before apply", "nil empty", "nil before validate", "nil empty before validate", "record task failure", "record decision failure", "visit failure"} {
		t.Run(variant, func(t *testing.T) {
			tasks, decisions := []map[string]any{task}, []map[string]any{decision}
			if variant == "nil empty" || variant == "nil empty before validate" {
				tasks, decisions = nil, nil
			}
			if variant == "record decision failure" {
				tasks = nil
			}
			tx, prepared, input, attrs := prepare(t, tasks, decisions)
			failure := errors.New("attribution capability failure")
			if variant == "nil before apply" || variant == "nil empty" {
				input.Attributions = nil
			}
			if variant == "record task failure" || variant == "record decision failure" {
				attrs.recordErr = failure
			}
			err := port.ApplyImportTx(ctx, tx, prepared, input)
			if variant == "nil before apply" || variant == "nil empty" {
				requireTasksDecisionsPortabilityFailure(t, err, "tasks_decisions.attribution_exact")
			} else if attrs.recordErr != nil {
				if !errors.Is(err, failure) {
					t.Fatalf("recorder failure changed: %v", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if variant == "nil before validate" || variant == "nil empty before validate" {
					input.Attributions = nil
					requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.attribution_exact")
				} else {
					attrs.visitErr = failure
					if err := port.ValidateImportTx(ctx, tx, prepared, input); !errors.Is(err, failure) {
						t.Fatalf("visitor failure changed: %v", err)
					}
				}
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
