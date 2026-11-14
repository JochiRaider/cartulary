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
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func exerciseRowTargetCurrentness(t *testing.T) {
	for _, kinds := range [][2]string{{"host", "record"}, {"record", "host"}} {
		for _, later := range []string{"ordinary", "rollback", "other record"} {
			t.Run(kinds[0]+" then "+kinds[1]+"/"+later, func(t *testing.T) {
				ctx := context.Background()
				h := appsupport.StartStore(t, "row-target-currentness")
				actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "row-currentness@example.test", DisplayName: "Row currentness"})
				incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "row-currentness-incident", "IR-ROW-CURRENTNESS", "Row currentness")
				id := uuid.New()
				entitytest.SeedHostRecord(t, h.DB, incident.ID, actor.ID, id, "before", "ROW-CURRENTNESS", "", "")
				composition := revisionsupport.MustComposition(t)
				appender := composition.Runtime.Appender()
				projections, err := projectionassembly.Build(h.DB)
				if err != nil {
					t.Fatal(err)
				}
				now := time.Now().UTC().Truncate(time.Microsecond)
				mutate := func(recordID uuid.UUID, kind, name, source string, version int64) uuid.UUID {
					t.Helper()
					tx, err := h.DB.BeginTx(ctx, pgx.TxOptions{})
					if err != nil {
						t.Fatal(err)
					}
					defer func() { _ = tx.Rollback(ctx) }()
					before, err := appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
					if err != nil {
						t.Fatal(err)
					}
					change, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: source, CreatedAt: now})
					if err != nil {
						t.Fatal(err)
					}
					if _, err := tx.Exec(ctx, `UPDATE records SET row_version=$2,updated_at=$3 WHERE record_id=$1`, recordID, version, now); err != nil {
						t.Fatal(err)
					}
					if _, err := tx.Exec(ctx, `UPDATE hosts SET row_version=$2,updated_at=$3,display_name=$4 WHERE record_id=$1`, recordID, version, now, name); err != nil {
						t.Fatal(err)
					}
					after, err := appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
					if err != nil {
						t.Fatal(err)
					}
					operation := "field_update"
					if source == "rollback" {
						operation = "rollback"
					}
					if err := appender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{ChangeSetID: change, SequenceNo: 1, TargetKind: kind, RecordID: recordID, OperationKind: operation, BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
						t.Fatal(err)
					}
					if err := appender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{ChangeSetID: change, RecordID: recordID, RowVersion: version, BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
						t.Fatal(err)
					}
					if err := projections.RevisionRows().RefreshRowTx(ctx, tx, "cartulary.view.hosts.v1", recordID); err != nil {
						t.Fatal(err)
					}
					if err := tx.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					return change
				}
				selected := mutate(id, kinds[0], "selected after", "workbook.records.patch", 2)
				laterID, laterVersion, baseVersion := id, int64(3), int64(3)
				if later == "other record" {
					laterID, laterVersion, baseVersion = uuid.New(), 2, 2
					entitytest.SeedHostRecord(t, h.DB, incident.ID, actor.ID, laterID, "other before", "OTHER-CURRENTNESS", "", "")
				}
				source, laterName := "workbook.records.patch", "later after"
				if later == "rollback" {
					source, laterName = "rollback", "before"
				}
				mutate(laterID, kinds[1], laterName, source, laterVersion)
				service, err := composition.Runtime.NewCommandService(h.DB, h.DB, rollbackNoImportedAttribution{}, projections.RevisionRows(), projections.RevisionLiveRecords(), func() time.Time { return now.Add(time.Minute) }, composition.RecordChanges)
				if err != nil {
					t.Fatal(err)
				}
				state := func() map[string]string {
					result := map[string]string{}
					for _, table := range []string{"records", "hosts", "host_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
						var value string
						if err := h.DB.QueryRow(ctx, "SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
							t.Fatal(err)
						}
						result[table] = value
					}
					return result
				}
				before, publication := state(), collaborationsupport.SnapshotPublicationState(t, h.DB, incident.ID)
				result, err := service.RollbackRecord(ctx, revisions.RollbackCommand{Actor: revisions.NewActorID(actor.ID), RecordID: id, RequestHash: []byte("row-currentness-hash"), Request: revisions.RollbackRequest{BaseRowVersion: baseVersion, ClientTxnID: "row-currentness-rollback", Target: revisions.RollbackTarget{Kind: "change_set", ChangeSetID: selected.String()}}})
				if later == "other record" {
					if err != nil {
						t.Fatalf("unrelated row blocked reversal: %v", err)
					}
					var exact bool
					if err := h.DB.QueryRow(ctx, `SELECT h.display_name='before' AND h.row_version=3 AND r.row_version=3 AND p.display_name='before' AND p.row_version=3 FROM hosts h JOIN records r USING(record_id) JOIN host_grid_projection p USING(record_id) WHERE h.record_id=$1`, id).Scan(&exact); err != nil || !exact {
						t.Fatalf("unrelated-row reversal source/projection=%v error=%v result=%#v", exact, err, result)
					}
					return
				}
				want := "dependent_later_changes"
				if later == "rollback" {
					want = "stale_target"
				}
				var precondition *revisions.RollbackPreconditionError
				if !errors.As(err, &precondition) || precondition.ReasonCode != want {
					t.Errorf("old row reversal error=%v, want %s", err, want)
				}
				if !reflect.DeepEqual(before, state()) || !reflect.DeepEqual(publication, collaborationsupport.SnapshotPublicationState(t, h.DB, incident.ID)) {
					t.Error("old row reversal changed source/history/projection/allocator/receipt/publication")
				}
			})
		}
	}
}
