package revisions_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func exerciseTaskInverseAssociation(t *testing.T) {
	for _, scenario := range []string{"matching", "mismatch", "clear", "hidden endpoint", "wrong type", "multiple", "no bindings", "whole set"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			h := appsupport.StartStore(t, "task-inverse-association")
			actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "task-association@example.test", DisplayName: "Task association"})
			incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "task-association-incident", "IR-TASK-ASSOCIATION", "Task association")
			owner := appsupport.NewTaskDecisionOwner(h.DB, conflicttest.NewCodec("task-association"))
			now := time.Now().UTC().Truncate(time.Microsecond)
			text := func(value string) tasksdecisions.FieldValue { return tasksdecisions.FieldValue{Text: &value} }
			create := func(view, txn string, values map[string]tasksdecisions.FieldValue) tasksdecisions.MutationResult {
				t.Helper()
				request := tasksdecisions.CreateRequest{ViewSchemaID: view, ClientTxnID: txn, Values: values}
				result, err := owner.Create(ctx, tasksdecisions.CreateCommand{ActorUserID: actor.ID, IncidentID: incident.ID, Request: request, RequestHash: tasksdecisions.CreateRequestHash(request), RequestID: txn, RouteKey: "workbook.rows.create", Now: now})
				if err != nil {
					t.Fatal(err)
				}
				return result
			}
			decisions := []uuid.UUID{}
			for _, name := range []string{"Decision A", "Decision B"} {
				result := create(tasksdecisions.DecisionsViewSchemaID, name, map[string]tasksdecisions.FieldValue{"decision.summary": text(name), "decision.decision_type": text("containment"), "decision.rationale": text("Rationale")})
				decisions = append(decisions, result.RecordID)
			}
			values := map[string]tasksdecisions.FieldValue{"task.title": text("Original title"), "task.task_kind": text("collection")}
			if scenario != "clear" {
				values["task.decision_record_id"] = tasksdecisions.FieldValue{UUID: &decisions[0]}
			}
			task := create(tasksdecisions.TaskRequestsViewSchemaID, "task-create", values)
			change := tasksdecisions.PatchChange{FieldKey: "task.title", Value: new(tasksdecisions.FieldValue)}
			*change.Value = text("Changed title")
			if scenario == "mismatch" || scenario == "clear" || scenario == "hidden endpoint" || scenario == "whole set" {
				change = tasksdecisions.PatchChange{FieldKey: "task.decision_record_id", Value: &tasksdecisions.FieldValue{UUID: &decisions[1]}}
			}
			patch := tasksdecisions.PatchRequest{ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID, BaseRowVersion: 1, ClientTxnID: "task-patch", Changes: []tasksdecisions.PatchChange{change}}
			patched, err := owner.Patch(ctx, tasksdecisions.PatchCommand{ActorUserID: actor.ID, RecordID: task.RecordID, Request: patch, RequestHash: tasksdecisions.PatchRequestHash(patch), RequestID: "task-patch", RouteKey: "workbook.records.patch", ConflictRouteKey: "workbook.records.conflicts.resolve", Now: now.Add(time.Minute)})
			if err != nil {
				t.Fatal(err)
			}
			exec := func(query string, args ...any) {
				t.Helper()
				if _, err := h.DB.Exec(ctx, query, args...); err != nil {
					t.Fatal(err)
				}
			}
			switch scenario {
			case "hidden endpoint":
				exec(`UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, decisions[1], now, actor.ID)
			case "wrong type":
				exec(`UPDATE record_links SET link_type='supported_by' WHERE src_record_id=$1 AND field_key='task.decision_record_id' AND deleted_at IS NULL`, task.RecordID)
			case "multiple":
				exec(`INSERT INTO record_links(incident_id,src_record_id,dst_record_id,link_type,field_key,provenance,owner_user_id,created_by_user_id,decided_at,created_at) VALUES($1,$2,$3,'references_record','task.decision_record_id','manual',$4,$4,$5,$5)`, incident.ID, task.RecordID, decisions[1], actor.ID, now)
			case "no bindings":
				exec(`UPDATE record_links SET deleted_at=$2,deleted_by_user_id=$3 WHERE src_record_id=$1 AND field_key='task.decision_record_id' AND deleted_at IS NULL`, task.RecordID, now, actor.ID)
			}
			projections, err := projectionassembly.Build(h.DB)
			if err != nil {
				t.Fatal(err)
			}
			composition := revisionsupport.MustComposition(t)
			service, err := composition.Runtime.NewCommandService(h.DB, h.DB, rollbackNoImportedAttribution{}, projections.RevisionRows(), projections.RevisionLiveRecords(), func() time.Time { return now.Add(2 * time.Minute) }, composition.RecordChanges)
			if err != nil {
				t.Fatal(err)
			}
			state := func() map[string]string {
				t.Helper()
				result := map[string]string{}
				for _, table := range []string{"records", "task_requests", "decisions", "record_links", "task_request_grid_projection", "decision_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
					var value string
					if err := h.DB.QueryRow(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
						t.Fatal(err)
					}
					result[table] = value
				}
				return result
			}
			before, publication := state(), collaborationsupport.SnapshotPublicationState(t, h.DB, incident.ID)
			target := revisions.RollbackTarget{Kind: "row_restore", RestoreToRevisionNo: 1}
			if scenario == "whole set" {
				target = revisions.RollbackTarget{Kind: "change_set", ChangeSetID: patched.ChangeSetID.String()}
			}
			_, err = service.RollbackRecord(ctx, revisions.RollbackCommand{Actor: revisions.NewActorID(actor.ID), RecordID: task.RecordID, RequestHash: []byte("association-rollback"), Request: revisions.RollbackRequest{BaseRowVersion: 2, ClientTxnID: "association-rollback", Target: target}})
			valid := scenario == "matching" || scenario == "no bindings" || scenario == "whole set"
			if !valid {
				var precondition *revisions.RollbackPreconditionError
				if !errors.As(err, &precondition) || precondition.ReasonCode != "target_not_reversible" {
					t.Errorf("association inverse error=%v; want target_not_reversible", err)
				}
				if !reflect.DeepEqual(before, state()) || !reflect.DeepEqual(publication, collaborationsupport.SnapshotPublicationState(t, h.DB, incident.ID)) {
					t.Error("rejected association inverse changed retained state/publication")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var exact bool
			if err := h.DB.QueryRow(ctx, `SELECT t.decision_record_id=$2 AND p.decision_record_id=$2 AND t.title='Original title' AND p.title='Original title' AND r.row_version=3 AND p.row_version=3 FROM task_requests t JOIN records r USING(record_id) JOIN task_request_grid_projection p USING(record_id) WHERE t.record_id=$1`, task.RecordID, decisions[0]).Scan(&exact); err != nil || !exact {
				t.Fatalf("restored source/projection exact=%v error=%v", exact, err)
			}
			if scenario != "whole set" {
				if before["record_links"] != state()["record_links"] {
					t.Fatal("row restore mutated retained Links")
				}
			} else {
				var count int
				if err := h.DB.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND field_key='task.decision_record_id' AND deleted_at IS NULL AND link_type='references_record' AND dst_record_id=$2`, task.RecordID, decisions[0]).Scan(&count); err != nil || count != 1 {
					t.Fatalf("whole-set companion association count=%d error=%v", count, err)
				}
			}
		})
	}
}
