package links_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

func testCollectionTargetLocking(t *testing.T, db dbport.DB, incident, actor uuid.UUID) {
	for _, mode := range []string{"record", "party"} {
		t.Run("collection projection writers/"+mode, func(t *testing.T) {
			testLinkProjectionWriterOverlap(t, db, incident, actor, mode)
		})
		for _, scenario := range []string{"hold through completion", "wait for deletion", "missing", "wrong type", "foreign", "deleted"} {
			t.Run("collection target lock/"+mode+"/"+scenario, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
				defer cancel()
				source, target := uuid.New(), uuid.New()
				entitytest.SeedHostRecord(t, db, incident, actor, source, "Collection source", source.String(), "", "")
				if scenario != "missing" {
					kind := "party"
					if scenario == "wrong type" {
						kind = "host"
					}
					envelopetest.SeedRecordEnvelope(t, db, incident, actor, target, kind)
					if kind == "party" {
						if _, err := db.Exec(ctx, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Collection target','organization')`, target, incident); err != nil {
							t.Fatal(err)
						}
					}
				}
				if scenario == "deleted" {
					if _, err := db.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, target, actor); err != nil {
						t.Fatal(err)
					}
				}
				commandIncident := incident
				if scenario == "foreign" {
					commandIncident = uuid.New()
				}
				const field = "host.subject_party_ids"
				now := time.Date(2026, 11, 14, 1, 0, 0, 0, time.UTC)
				apply := func(tx pgx.Tx) (links.CollectionMutationResult, error) {
					if mode == "record" {
						return links.NewStore().ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{IncidentID: commandIncident, SourceRecordID: source, ActorUserID: actor, FieldKey: field, LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddRecordIDs: []uuid.UUID{target}, Now: now})
					}
					return links.NewStore().ApplyPartyRefCollectionWithMutationValuesTx(ctx, tx, links.PartyRefCollectionCommand{IncidentID: commandIncident, SourceRecordID: source, ActorUserID: actor, FieldKey: field, LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddPartyIDs: []uuid.UUID{target}, Now: now})
				}
				snapshot := func(tx pgx.Tx) string {
					t.Helper()
					var state string
					if err := tx.QueryRow(ctx, `SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY record_link_id),'[]'::jsonb)::text FROM record_links l WHERE incident_id=$1`, incident).Scan(&state); err != nil {
						t.Fatal(err)
					}
					return state
				}
				writer, err := db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer writer.Rollback(context.Background())
				before := snapshot(writer)
				var result links.CollectionMutationResult
				if scenario == "wait for deletion" {
					holder, beginErr := db.BeginTx(ctx, pgx.TxOptions{})
					if beginErr != nil {
						t.Fatal(beginErr)
					}
					defer holder.Rollback(context.Background())
					var holderPID int32
					if err := holder.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&holderPID); err != nil {
						t.Fatal(err)
					}
					if _, err := holder.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, target, actor); err != nil {
						t.Fatal(err)
					}
					type outcome struct {
						result links.CollectionMutationResult
						err    error
					}
					done := make(chan outcome, 1)
					go func() { r, e := apply(writer); done <- outcome{r, e} }()
					consumed := false
					defer func() {
						cancel()
						if !consumed {
							select {
							case <-done:
							case <-time.After(5 * time.Second):
								t.Error("collection worker did not drain")
							}
						}
					}()
					for {
						var waiting bool
						if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, holderPID).Scan(&waiting); err != nil {
							t.Fatal(err)
						}
						if waiting {
							break
						}
						select {
						case r := <-done:
							consumed = true
							t.Fatalf("collection did not wait for target deletion: %v", r.err)
						case <-ctx.Done():
							t.Fatal(ctx.Err())
						case <-time.After(10 * time.Millisecond):
						}
					}
					if err := holder.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					select {
					case r := <-done:
						consumed = true
						result, err = r.result, r.err
					case <-ctx.Done():
						t.Fatal(ctx.Err())
					}
				} else {
					result, err = apply(writer)
				}
				if scenario != "hold through completion" {
					var invalid *links.CollectionValidationError
					if !errors.As(err, &invalid) || invalid.Field != field || invalid.ReasonCode != "invalid_value" || len(result.Mutations()) != 0 {
						t.Fatalf("collection rejection=%#v error=%v", result, err)
					}
					if snapshot(writer) != before {
						t.Fatal("rejected collection changed retained Links")
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				mutations := result.Mutations()
				if len(mutations) != 1 || mutations[0].OperationKind() != "create" || mutations[0].TargetKind() != "record_link" || mutations[0].BeforeValue() != nil {
					t.Fatalf("created relation mutation=%#v", mutations)
				}
				var exact bool
				if err := writer.QueryRow(ctx, `SELECT src_record_id=$2 AND dst_record_id=$3 AND incident_id=$4 AND field_key=$5 AND link_type='references_record' AND owner_user_id=$6 AND created_by_user_id=$6 AND decided_at=$7 AND created_at=$7 AND deleted_at IS NULL AND deleted_by_user_id IS NULL FROM record_links WHERE record_link_id=$1`, mutations[0].TargetID(), source, target, incident, field, actor, now).Scan(&exact); err != nil || !exact {
					t.Fatalf("exact collection relation=%v error=%v", exact, err)
				}
				contender, err := db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer contender.Rollback(context.Background())
				_, err = contender.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, target)
				var postgresError *pgconn.PgError
				if !errors.As(err, &postgresError) || postgresError.Code != "55P03" {
					t.Fatalf("collection released target before caller completion: %v", err)
				}
				if err := contender.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
				if err := writer.Commit(ctx); err != nil {
					t.Fatal(err)
				}
				released, err := db.BeginTx(ctx, pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer released.Rollback(context.Background())
				if _, err := released.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, target); err != nil {
					t.Fatalf("caller commit did not release target lock: %v", err)
				}
			})
		}
	}
}
