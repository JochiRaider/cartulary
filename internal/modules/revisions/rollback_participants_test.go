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
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

type rollbackTransactionWitness struct {
	postgres.DB
	tx        pgx.Tx
	committed int
}
type rollbackParticipantTx struct {
	pgx.Tx
	owner *rollbackTransactionWitness
}

func (w *rollbackTransactionWitness) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := w.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	w.tx = &rollbackParticipantTx{Tx: tx, owner: w}
	return w.tx, nil
}
func (tx *rollbackParticipantTx) Commit(ctx context.Context) error {
	tx.owner.committed++
	return tx.Tx.Commit(ctx)
}

type rollbackProjectionWitness struct {
	revisions.ProjectionRefresher
	t        *testing.T
	runner   *rollbackTransactionWitness
	recordID uuid.UUID
	calls    int
	fail     error
}

func (w *rollbackProjectionWitness) RefreshRowTx(ctx context.Context, tx pgx.Tx, view string, id uuid.UUID) error {
	w.calls++
	if tx != w.runner.tx || id != w.recordID || view != "cartulary.view.hosts.v1" {
		w.t.Fatalf("projection participant boundary: %v %s %s", tx == w.runner.tx, view, id)
	}
	var name string
	if err := tx.QueryRow(ctx, "SELECT display_name FROM hosts WHERE record_id=$1", id).Scan(&name); err != nil {
		return err
	}
	if name != "before" {
		w.t.Fatalf("projection cannot see restored source: %s", name)
	}
	if err := w.ProjectionRefresher.RefreshRowTx(ctx, tx, view, id); err != nil {
		return err
	}
	return w.fail
}

type rollbackPublicationWitness struct {
	collaboration.RecordChangedAppender
	t                         *testing.T
	runner                    *rollbackTransactionWitness
	recordID, incident, actor uuid.UUID
	calls                     int
	fail                      error
}

func (w *rollbackPublicationWitness) AppendRecordChangedTx(ctx context.Context, tx pgx.Tx, input collaboration.RecordChangeIntentInput) error {
	w.calls++
	if tx != w.runner.tx || input.RecordID != w.recordID || input.IncidentID != w.incident || input.ActorUserID != w.actor || input.RowVersion != 3 {
		w.t.Fatalf("publication participant boundary: tx=%v input=%#v", tx == w.runner.tx, input)
	}
	if err := w.RecordChangedAppender.AppendRecordChangedTx(ctx, tx, input); err != nil {
		return err
	}
	return w.fail
}

type rollbackNoImportedAttribution struct{}

func (rollbackNoImportedAttribution) ResolveImportedSourceActorsTx(context.Context, pgx.Tx, uuid.UUID, string, string, []string) (map[string]string, error) {
	return map[string]string{}, nil
}

func TestRollbackParticipantsAndHistoricalAppend_Integration(t *testing.T) {
	for _, stage := range []string{"success", "projection failure", "publication failure"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			harness := appsupport.StartStore(t, "rollback-participants")
			actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "rollback-participants@example.test", DisplayName: "Rollback participants"})
			incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "rollback-participants-incident", "IR-PARTICIPANTS", "Rollback participants")
			id := uuid.New()
			entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, id, "before", "PARTICIPANT", "", "")
			composition := revisionsupport.MustComposition(t)
			appender := composition.Runtime.Appender()
			projections, err := projectionassembly.Build(harness.DB)
			if err != nil {
				t.Fatal(err)
			}
			now := time.Now().UTC().Truncate(time.Microsecond)
			tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback(ctx) }()
			before, err := appender.CaptureRecordSnapshotTx(ctx, tx, id)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: "workbook.records.patch", CreatedAt: now})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, "UPDATE records SET row_version=2,updated_at=$2 WHERE record_id=$1", id, now); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, "UPDATE hosts SET row_version=2,display_name='after',updated_at=$2 WHERE record_id=$1", id, now); err != nil {
				t.Fatal(err)
			}
			after, err := appender.CaptureRecordSnapshotTx(ctx, tx, id)
			if err != nil {
				t.Fatal(err)
			}
			if err := appender.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{ChangeSetID: changed, SequenceNo: 1, TargetKind: "host", RecordID: id, OperationKind: "field_update", BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
				t.Fatal(err)
			}
			if err := appender.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{ChangeSetID: changed, RecordID: id, RowVersion: 2, BeforeSnapshot: &before, AfterSnapshot: &after}); err != nil {
				t.Fatal(err)
			}
			if err := projections.RevisionRows().RefreshRowTx(ctx, tx, "cartulary.view.hosts.v1", id); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			runner := &rollbackTransactionWitness{DB: harness.DB}
			projection := &rollbackProjectionWitness{ProjectionRefresher: projections.RevisionRows(), t: t, runner: runner, recordID: id}
			publication := &rollbackPublicationWitness{RecordChangedAppender: composition.RecordChanges, t: t, runner: runner, recordID: id, incident: incident.ID, actor: actor.ID}
			sentinel := errors.New("injected participant failure")
			if stage == "projection failure" {
				projection.fail = sentinel
			}
			if stage == "publication failure" {
				publication.fail = sentinel
			}
			service, err := composition.Runtime.NewCommandService(harness.DB, runner, rollbackNoImportedAttribution{}, projection, projections.RevisionLiveRecords(), func() time.Time { return now.Add(time.Minute) }, publication)
			if err != nil {
				t.Fatal(err)
			}
			state := func() map[string]string {
				result := map[string]string{}
				for _, table := range []string{"records", "hosts", "host_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "route_idempotency"} {
					var value string
					if err := harness.DB.QueryRow(ctx, "SELECT COALESCE(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
						t.Fatal(err)
					}
					result[table] = value
				}
				return result
			}
			initial, initialPublication := state(), collaborationsupport.SnapshotPublicationState(t, harness.DB, incident.ID)
			result, err := service.RollbackRecord(ctx, revisions.RollbackCommand{Actor: revisions.NewActorID(actor.ID), RecordID: id, RequestHash: []byte("rollback-participants-hash"), Request: revisions.RollbackRequest{BaseRowVersion: 2, ClientTxnID: "rollback-participants-command", Target: revisions.RollbackTarget{Kind: "change_set", ChangeSetID: changed.String()}}})
			if stage != "success" {
				if !errors.Is(err, sentinel) {
					t.Fatalf("participant failure=%v", err)
				}
				if runner.committed != 0 || !reflect.DeepEqual(initial, state()) || !reflect.DeepEqual(initialPublication, collaborationsupport.SnapshotPublicationState(t, harness.DB, incident.ID)) {
					t.Fatal("participant failure committed source/history/projection/publication/receipt effects")
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if runner.committed != 1 {
					t.Fatalf("commits=%d", runner.committed)
				}
				var name string
				var version int64
				if err := harness.DB.QueryRow(ctx, "SELECT display_name,row_version FROM hosts WHERE record_id=$1", id).Scan(&name, &version); err != nil {
					t.Fatal(err)
				}
				if name != "before" || version != 3 {
					t.Fatalf("restored source=%s v%d result=%#v", name, version, result)
				}
				for _, table := range []string{"records", "host_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "route_idempotency"} {
					if state()[table] == initial[table] {
						t.Fatalf("missing committed effect in %s", table)
					}
				}
				if reflect.DeepEqual(initialPublication, collaborationsupport.SnapshotPublicationState(t, harness.DB, incident.ID)) {
					t.Fatal("no committed publication")
				}
				// Historical append retains history without touching live state or publishing.
				live := state()
				published := collaborationsupport.SnapshotPublicationState(t, harness.DB, incident.ID)
				historical, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = historical.Rollback(ctx) }()
				change, err := appender.AppendChangeSetTx(ctx, historical, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: "history.import", CreatedAt: now.Add(-time.Hour)})
				if err != nil {
					t.Fatal(err)
				}
				if err := appender.AppendHistoricalRevisionTx(ctx, historical, revisions.HistoricalRevisionInput{ChangeSetID: change, RecordID: id, RowVersion: 1, AfterSnapshot: &before}); err != nil {
					t.Fatal(err)
				}
				if err := historical.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				current := state()
				for _, table := range []string{"records", "hosts", "host_grid_projection", "record_revision_conflict_facts", "route_idempotency"} {
					if live[table] != current[table] {
						t.Fatalf("historical append changed %s", table)
					}
				}
				if live["record_revisions"] == current["record_revisions"] || !reflect.DeepEqual(published, collaborationsupport.SnapshotPublicationState(t, harness.DB, incident.ID)) {
					t.Fatal("historical append history/publication isolation failed")
				}
			}
			expectedPublications := 1
			if stage == "projection failure" {
				expectedPublications = 0
			}
			if projection.calls != 1 || publication.calls != expectedPublications {
				t.Fatalf("participant calls: projection=%d publication=%d", projection.calls, publication.calls)
			}
		})
	}
}
