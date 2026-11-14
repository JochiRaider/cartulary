package tasksdecisions_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestIncidentBundleTasksDecisionsExactSourceState(t *testing.T) {
	t.Run("envelope capability", requirePortableEnvelopeCapability)
	t.Run("collection endpoint capability", requirePortableCollectionEnvelopes)
	ctx := t.Context()
	harness := newTasksDecisionsPortabilityHarness(t, "exact-state")
	now := taskDecisionPortableTime()
	taskID, decisionID, extraTask, extraDecision, party := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, item := range []struct {
		id   uuid.UUID
		kind string
	}{
		{taskID, "task_request"}, {decisionID, "decision"}, {extraTask, "task_request"}, {extraDecision, "decision"}, {party, "party"},
	} {
		seedTaskDecisionEnvelope(t, harness, item.id, item.kind, now)
	}
	if _, err := harness.db.Exec(ctx, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Retained requester','organization')`, party, harness.incidentID); err != nil {
		t.Fatal(err)
	}
	other := authstoretest.SeedActor(t, harness.db, authstoretest.ActorAttributes{Email: "td-state-" + uuid.NewString() + "@example.test", DisplayName: "Other owner"})
	foreign := appsupport.CreateIncidentInStore(t, harness.db, harness.actor, "txn-td-state-foreign", "TD-STATE-FOREIGN", "Foreign source state")
	foreignTask, foreignDecision := uuid.New(), uuid.New()
	seedEnvelopeForIncident(t, harness, foreign.ID, foreignTask, "task_request", now)
	seedEnvelopeForIncident(t, harness, foreign.ID, foreignDecision, "decision", now)
	portableActor := uuid.New()
	input := taskDecisionImportContext(harness, "exact-state")
	var err error
	input.Actors, err = sourceport.NewActorCatalog([]sourceport.ActorDescriptor{
		{SourceActorID: portableActor.String(), DisplayName: "Portable owner"},
		{SourceActorID: harness.actor.ID.String(), DisplayName: "Importer"},
	})
	if err != nil {
		t.Fatal(err)
	}
	task := validPortableTask(harness, taskID, now)
	task["owner_user_id"], task["status"], task["completed_at"] = portableActor.String(), "done", now.Add(time.Second).Format(time.RFC3339Nano)
	task["updated_at"] = now.Add(2 * time.Second).Format(time.RFC3339Nano)
	task["workstream"], task["due_at"] = "Triage", now.Add(time.Hour).Format(time.RFC3339Nano)
	task["requester_party_text"], task["requester_party_id"] = "Responder", party.String()
	task["external_ticket_ref"], task["closure_summary"], task["decision_record_id"] = "CASE-42", "Completed action", decisionID.String()
	decision := validPortableDecision(harness, decisionID, now)
	decision["owner_user_id"] = portableActor.String()
	decision["decided_at"], decision["updated_at"] = now.Add(3*time.Second).Format(time.RFC3339Nano), now.Add(4*time.Second).Format(time.RFC3339Nano)
	port := newTaskDecisionIncidentBundleSourcePort(t)
	apply := func(t testing.TB, tx pgx.Tx, tasks, decisions []map[string]any) sourceport.Prepared {
		t.Helper()
		prepared, err := port.PrepareImport(ctx, taskDecisionBundle(tasks, decisions), input)
		if err != nil {
			t.Fatal(err)
		}
		if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		return prepared
	}
	begin := func(t testing.TB) pgx.Tx {
		t.Helper()
		input.Attributions = &taskDecisionAttributions{local: harness.actor.ID}
		tx, err := harness.db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx
	}
	validate := func(t testing.TB, tx pgx.Tx, prepared sourceport.Prepared) {
		t.Helper()
		if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
	}
	for _, nullOwner := range []bool{false, true} {
		t.Run(fmt.Sprintf("exact populated values null owner %t", nullOwner), func(t *testing.T) {
			tx := begin(t)
			row := maps.Clone(task)
			if nullOwner {
				row["owner_user_id"] = nil
			}
			prepared := apply(t, tx, []map[string]any{row}, []map[string]any{decision})
			validate(t, tx, prepared)
			var taskOwner *uuid.UUID
			var decisionOwner uuid.UUID
			if err := tx.QueryRow(ctx, `SELECT owner_user_id FROM task_requests WHERE record_id=$1`, taskID).Scan(&taskOwner); err != nil {
				t.Fatal(err)
			}
			if err := tx.QueryRow(ctx, `SELECT owner_user_id FROM decisions WHERE record_id=$1`, decisionID).Scan(&decisionOwner); err != nil {
				t.Fatal(err)
			}
			if decisionOwner != harness.actor.ID || (nullOwner && taskOwner != nil) || (!nullOwner && (taskOwner == nil || *taskOwner != harness.actor.ID)) {
				t.Fatalf("incorrect actor substitution: task=%v decision=%s", taskOwner, decisionOwner)
			}
			for table, source := range map[string]map[string]any{"task_requests": row, "decisions": decision} {
				expected := maps.Clone(source)
				if expected["owner_user_id"] != nil {
					expected["owner_user_id"] = harness.actor.ID.String()
				}
				for _, key := range []string{"created_at", "updated_at", "completed_at", "due_at", "decided_at"} {
					if value, ok := expected[key].(string); ok {
						expected[key] = strings.TrimSuffix(value, "Z") + "+00:00"
					}
				}
				var payload []byte
				if err := tx.QueryRow(ctx, "SELECT to_jsonb(s) FROM "+table+" s WHERE record_id=$1", source["record_id"]).Scan(&payload); err != nil {
					t.Fatal(err)
				}
				var actual map[string]any
				if err := json.Unmarshal(payload, &actual); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatalf("independent %s values differ: got %#v want %#v", table, actual, expected)
				}
			}
		})
	}
	for _, specimen := range []struct {
		table                string
		id, extra, foreignID uuid.UUID
		changes              []struct {
			field string
			value any
		}
	}{
		{"task_requests", taskID, extraTask, foreignTask, []struct {
			field string
			value any
		}{
			{"record_id", extraTask}, {"incident_id", foreign.ID}, {"title", "Different title"}, {"status", "canceled"},
			{"owner_user_id", other.ID}, {"priority", "urgent"}, {"task_kind", "containment"}, {"workstream", nil},
			{"due_at", nil}, {"requester_party_text", nil}, {"requester_party_id", nil}, {"blocked_reason", "New block"},
			{"completed_at", nil}, {"external_ticket_ref", nil}, {"closure_summary", nil}, {"decision_record_id", nil},
			{"created_at", now.Add(-time.Second)}, {"updated_at", now.Add(time.Second)},
			{"title", nil}, {"priority", nil}, {"task_kind", nil}, {"owner_user_id", nil},
		}},
		{"decisions", decisionID, extraDecision, foreignDecision, []struct {
			field string
			value any
		}{
			{"record_id", extraDecision}, {"incident_id", foreign.ID}, {"summary", "Different summary"}, {"status", "proposed"},
			{"owner_user_id", other.ID}, {"decision_type", "containment"}, {"decided_at", now.Add(time.Second)},
			{"rationale", "Different rationale"}, {"created_at", now.Add(-time.Second)}, {"updated_at", now.Add(time.Second)},
			{"summary", nil}, {"owner_user_id", nil}, {"decision_type", nil}, {"decided_at", nil}, {"rationale", nil},
		}},
	} {
		for index, change := range specimen.changes {
			t.Run(fmt.Sprintf("%s drift %s %d", specimen.table, change.field, index), func(t *testing.T) {
				tx := begin(t)
				prepared := apply(t, tx, []map[string]any{task}, []map[string]any{decision})
				validate(t, tx, prepared)
				// Table/column names come only from this closed test matrix. Moving
				// incident also moves the record ID to satisfy the composite FK.
				sql := "UPDATE " + specimen.table + " SET " + change.field + "=$2 WHERE record_id=$1"
				args := []any{specimen.id, change.value}
				if change.field == "incident_id" {
					sql = "UPDATE " + specimen.table + " SET incident_id=$2,record_id=$3 WHERE record_id=$1"
					args = append(args, specimen.foreignID)
				}
				if _, err := tx.Exec(ctx, sql, args...); err != nil {
					t.Fatal(err)
				}
				requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.source_state_exact")
			})
		}
		for _, variant := range []string{"missing", "extra", "empty remains empty", "empty rejects extra"} {
			t.Run(specimen.table+" "+variant, func(t *testing.T) {
				tx := begin(t)
				var tasks, decisions []map[string]any
				if variant == "missing" || variant == "extra" {
					tasks, decisions = []map[string]any{task}, []map[string]any{decision}
				}
				prepared := apply(t, tx, tasks, decisions)
				validate(t, tx, prepared)
				if variant == "empty remains empty" {
					return
				}
				if variant == "missing" {
					if _, err := tx.Exec(ctx, "DELETE FROM "+specimen.table+" WHERE record_id=$1", specimen.id); err != nil {
						t.Fatal(err)
					}
				} else if specimen.table == "task_requests" {
					apply(t, tx, []map[string]any{validPortableTask(harness, specimen.extra, now)}, nil)
				} else {
					apply(t, tx, nil, []map[string]any{validPortableDecision(harness, specimen.extra, now)})
				}
				requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.source_state_exact")
			})
		}
	}
	// Wrong source state must not hide the higher-priority envelope mismatch.
	t.Run("envelope binding precedes source drift", func(t *testing.T) {
		tx := begin(t)
		prepared := apply(t, tx, []map[string]any{task}, []map[string]any{decision})
		if _, err := tx.Exec(ctx, `UPDATE records SET record_type='decision' WHERE record_id=$1`, taskID); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET title='Changed' WHERE record_id=$1`, taskID); err != nil {
			t.Fatal(err)
		}
		requireTasksDecisionsPortabilityFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "tasks_decisions.envelope_type_scope")
	})
}
