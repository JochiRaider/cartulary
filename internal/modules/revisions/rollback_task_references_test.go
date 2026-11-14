package revisions_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func exerciseTaskInverseReferences(t *testing.T) {
	for _, referenceType := range []string{"decision", "party"} {
		t.Run(referenceType, func(t *testing.T) { exerciseTaskInverseReference(t, referenceType) })
	}
}

func exerciseTaskInverseReference(t *testing.T, referenceType string) {
	for _, kind := range []string{"row_restore", "change_set"} {
		for _, targetState := range []string{"active", "deleted", "missing source", "foreign", "wrong type", "null"} {
			t.Run(kind+"/"+targetState, func(t *testing.T) {
				ctx := context.Background()
				h := appsupport.StartStore(t, "task-inverse-reference")
				actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "task-inverse@example.test", DisplayName: "Task inverse"})
				incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "task-inverse-incident", "IR-TASK-INVERSE", "Task inverse")
				targetIncident := incident.ID
				if targetState == "foreign" {
					targetIncident = appsupport.CreateIncidentInStore(t, h.DB, actor, "task-inverse-foreign", "IR-TASK-FOREIGN", "Foreign reference").ID
				}
				id, reference := uuid.New(), uuid.New()
				var retainedReference any = reference
				if targetState == "null" {
					retainedReference = nil
				}
				composition := revisionsupport.MustComposition(t)
				appender := composition.Runtime.Appender()
				projections, err := projectionassembly.Build(h.DB)
				if err != nil {
					t.Fatal(err)
				}
				now := time.Now().UTC().Truncate(time.Microsecond)
				tx, err := h.DB.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = tx.Rollback(ctx) }()
				seed := func(query string, args ...any) {
					t.Helper()
					if _, err := tx.Exec(ctx, query, args...); err != nil {
						t.Fatal(err)
					}
				}
				targetType := referenceType
				if targetState == "wrong type" {
					targetType = "host"
				}
				seed(`INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id,created_at,updated_at)
VALUES($1,$2,$3,$4,$4,$5,$5)`, reference, targetIncident, targetType, actor.ID, now)
				if targetState != "missing source" && targetState != "wrong type" {
					if referenceType == "decision" {
						seed(`INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale,created_at,updated_at)
VALUES($1,$2,'Decision','approved',$3,'containment',$4,'Rationale',$4,$4)`, reference, targetIncident, actor.ID, now)
					} else {
						seed(`INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Requester','organization')`, reference, targetIncident)
					}
				}
				seed(`INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id,created_at,updated_at)
VALUES($1,$2,'task_request',$3,$3,$4,$4)`, id, incident.ID, actor.ID, now)
				var partyValue, decisionValue any
				if referenceType == "party" {
					partyValue = retainedReference
				} else {
					decisionValue = retainedReference
				}
				seed(`INSERT INTO task_requests(record_id,incident_id,title,status,owner_user_id,priority,task_kind,requester_party_id,decision_record_id,requester_party_text,created_at,updated_at)
VALUES($1,$2,'Restored title','open',$3,'normal','collection',$4,$5,'Original requester wording',$6,$6)`, id, incident.ID, actor.ID, partyValue, decisionValue, now)
				before, err := appender.CaptureRecordSnapshotTx(ctx, tx, id)
				if err != nil {
					t.Fatal(err)
				}
				created, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: "workbook.records.create", CreatedAt: now})
				if err != nil {
					t.Fatal(err)
				}
				if err := appender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{ChangeSetID: created, RecordID: id, RowVersion: 1, AfterSnapshot: &before}); err != nil {
					t.Fatal(err)
				}
				selected, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: "workbook.records.patch", CreatedAt: now})
				if err != nil {
					t.Fatal(err)
				}
				seed(`UPDATE records SET row_version=2 WHERE record_id=$1`, id)
				seed(`UPDATE task_requests SET title='Current title',decision_record_id=NULL,requester_party_id=NULL WHERE record_id=$1`, id)
				after, err := appender.CaptureRecordSnapshotTx(ctx, tx, id)
				if err != nil {
					t.Fatal(err)
				}
				if err := appender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{ChangeSetID: selected, SequenceNo: 1, TargetKind: "record", RecordID: id, OperationKind: "field_update", BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
					t.Fatal(err)
				}
				if err := appender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{ChangeSetID: selected, RecordID: id, RowVersion: 2, BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
					t.Fatal(err)
				}
				if targetState == "deleted" || targetState == "null" {
					seed(`UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, reference, now, actor.ID)
				}
				if err := projections.RevisionRows().RefreshRowTx(ctx, tx, "cartulary.view.task_requests.v1", id); err != nil {
					t.Fatal(err)
				}
				if err := tx.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				service, err := composition.Runtime.NewCommandService(h.DB, h.DB, rollbackNoImportedAttribution{}, projections.RevisionRows(), projections.RevisionLiveRecords(), func() time.Time { return now.Add(time.Minute) }, composition.RecordChanges)
				if err != nil {
					t.Fatal(err)
				}
				state := func() map[string]string {
					result := map[string]string{}
					for _, table := range []string{"records", "task_requests", "decisions", "parties", "party_active_key_claims", "record_links", "task_request_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
						var value string
						if err := h.DB.QueryRow(ctx, "SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
							t.Fatal(err)
						}
						result[table] = value
					}
					return result
				}
				unchanged, publication := state(), collaborationsupport.SnapshotPublicationState(t, h.DB, incident.ID)
				target := revisions.RollbackTarget{Kind: kind, ChangeSetID: selected.String()}
				if kind == "row_restore" {
					target = revisions.RollbackTarget{Kind: kind, RestoreToRevisionNo: 1}
				}
				_, err = service.RollbackRecord(ctx, revisions.RollbackCommand{Actor: revisions.NewActorID(actor.ID), RecordID: id, RequestHash: []byte("task-inverse-hash"), Request: revisions.RollbackRequest{BaseRowVersion: 2, ClientTxnID: "task-inverse-rollback", Target: target}})
				if targetState != "active" && targetState != "null" {
					var precondition *revisions.RollbackPreconditionError
					if !errors.As(err, &precondition) || precondition.ReasonCode != "target_not_reversible" {
						t.Errorf("invalid reference inverse error=%v, want target_not_reversible", err)
					}
					if !reflect.DeepEqual(unchanged, state()) || !reflect.DeepEqual(publication, collaborationsupport.SnapshotPublicationState(t, h.DB, incident.ID)) {
						t.Error("rejected inverse changed source/history/projection/receipt/allocator/publication")
					}
					return
				}
				if err != nil {
					t.Fatalf("valid reference inverse: %v", err)
				}
				var exact bool
				if err := h.DB.QueryRow(ctx, `SELECT t.title='Restored title' AND p.title='Restored title' AND r.row_version=3 AND p.row_version=3
AND t.requester_party_id IS NOT DISTINCT FROM $2::uuid AND p.requester_party_id IS NOT DISTINCT FROM $2::uuid
AND t.decision_record_id IS NOT DISTINCT FROM $3::uuid AND p.decision_record_id IS NOT DISTINCT FROM $3::uuid
AND t.requester_party_text='Original requester wording' AND p.requester_party_text='Original requester wording'
AND (SELECT count(*) FROM record_revisions WHERE record_id=$1)=3
FROM task_requests t JOIN records r USING(record_id) JOIN task_request_grid_projection p USING(record_id) WHERE t.record_id=$1`, id, partyValue, decisionValue).Scan(&exact); err != nil || !exact {
					t.Fatalf("restored Task source/projection/history exact=%v error=%v", exact, err)
				}
			})
		}
	}
}
