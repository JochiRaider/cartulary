package links_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func testFreshLinkEndpointLocks(t *testing.T, db dbport.DB, incident, actor uuid.UUID) {
	t.Run("overlapping projection writers", func(t *testing.T) {
		testLinkProjectionWriterOverlap(t, db, incident, actor, "upsert")
	})
	for _, mode := range []string{"upsert", "field", "supersedes"} {
		for _, endpoint := range []string{"source", "target"} {
			for _, scenario := range []string{"hold through completion", "wait for deletion"} {
				t.Run("fresh Link lock/"+mode+"/"+endpoint+"/"+scenario, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
					defer cancel()
					source, target := uuid.New(), uuid.New()
					source[0], target[0] = 0xf0, 0x10 // Caller order is deliberately opposite canonical lock order.
					now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
					for _, id := range []uuid.UUID{source, target} {
						kind := "decision"
						if mode == "field" && id == source {
							kind = "task_request"
						}
						envelopetest.SeedRecordEnvelope(t, db, incident, actor, id, kind)
						if kind == "task_request" {
							if _, err := db.Exec(ctx, `INSERT INTO task_requests(record_id,incident_id,title,status,owner_user_id,task_kind) VALUES($1,$2,'Source','open',$3,'collection')`, id, incident, actor); err != nil {
								t.Fatal(err)
							}
						} else {
							status := "approved"
							if mode == "supersedes" && id == target {
								status = "superseded"
							}
							if _, err := db.Exec(ctx, `INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale) VALUES($1,$2,'Endpoint',$3,$4,'containment',$5,'Retained rationale')`, id, incident, status, actor, now); err != nil {
								t.Fatal(err)
							}
						}
					}
					selected := source
					if endpoint == "target" {
						selected = target
					}
					begin := func() pgx.Tx {
						t.Helper()
						tx, err := db.BeginTx(ctx, pgx.TxOptions{})
						if err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
						return tx
					}
					snapshot := func(tx pgx.Tx) string {
						t.Helper()
						var state string
						if err := tx.QueryRow(ctx, `SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY record_link_id),'[]'::jsonb)::text FROM record_links l WHERE incident_id=$1`, incident).Scan(&state); err != nil {
							t.Fatal(err)
						}
						return state
					}
					writer := begin()
					before := snapshot(writer)
					store := links.NewStore()
					apply := func() ([]links.Mutation, error) {
						if mode == "field" {
							r, err := store.SyncFieldReferenceWithMutationValuesTx(ctx, writer, links.SyncFieldReferenceCommand{IncidentID: incident, SrcRecordID: source, FieldKey: "task.decision_record_id", LinkType: links.LinkTypeReferencesRecord, TargetID: &target, ActorUserID: actor, Now: now})
							return r.Mutations(), err
						}
						var result links.RecordLinkCommandResult
						var err error
						if mode == "supersedes" {
							result, err = store.InsertSupersedesCommandTx(ctx, writer, links.InsertSupersedesCommand{IncidentID: incident, ReplacementRecordID: source, SupersededRecordID: target, OwnerUserID: actor, Now: now})
						} else {
							result, err = store.UpsertLinkCommandTx(ctx, writer, links.UpsertLinkCommand{IncidentID: incident, SrcRecordID: source, DstRecordID: target, LinkType: links.LinkTypeReferencesRecord, Provenance: links.LinkProvenanceManual, OwnerUserID: actor, Now: now})
						}
						if mutation, ok := result.Mutation(); ok {
							return []links.Mutation{mutation}, err
						}
						return nil, err
					}
					requireLocked := func(id uuid.UUID) {
						t.Helper()
						probe := begin()
						_, err := probe.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, id)
						var pgErr *pgconn.PgError
						if !errors.As(err, &pgErr) || pgErr.Code != "55P03" {
							t.Fatalf("endpoint %s lacks writer protection: %v", id, err)
						}
						if err := probe.Rollback(ctx); err != nil {
							t.Fatal(err)
						}
					}
					var mutations []links.Mutation
					var err error
					if scenario == "wait for deletion" {
						holder := begin()
						var pid int32
						if err := holder.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
							t.Fatal(err)
						}
						if _, err := holder.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, selected, now, actor); err != nil {
							t.Fatal(err)
						}
						type outcome struct {
							mutations []links.Mutation
							err       error
						}
						done := make(chan outcome, 1)
						go func() { m, e := apply(); done <- outcome{m, e} }()
						consumed := false
						defer func() {
							cancel()
							if !consumed {
								select {
								case <-done:
								case <-time.After(5 * time.Second):
									t.Error("Link worker did not drain")
								}
							}
						}()
						for {
							var waiting bool
							if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, pid).Scan(&waiting); err != nil {
								t.Fatal(err)
							}
							if waiting {
								break
							}
							select {
							case r := <-done:
								consumed = true
								t.Fatalf("Link did not wait for deletion: mutations=%d error=%v", len(r.mutations), r.err)
							case <-ctx.Done():
								t.Fatal(ctx.Err())
							case <-time.After(10 * time.Millisecond):
							}
						}
						if selected == source {
							requireLocked(target)
						} // Lower UUID must already be protected while waiting for the higher one.
						if err := holder.Commit(ctx); err != nil {
							t.Fatal(err)
						}
						select {
						case r := <-done:
							consumed = true
							mutations, err = r.mutations, r.err
						case <-ctx.Done():
							t.Fatal(ctx.Err())
						}
						if err == nil || !strings.Contains(err.Error(), "invalid record link") || len(mutations) != 0 {
							t.Fatalf("deleted endpoint accepted: mutations=%#v error=%v", mutations, err)
						}
						if snapshot(writer) != before {
							t.Fatal("rejected endpoint changed retained Links")
						}
						return
					}
					mutations, err = apply()
					if err != nil || len(mutations) != 1 || mutations[0].OperationKind() != "create" || mutations[0].TargetKind() != "record_link" || mutations[0].BeforeValue() != nil {
						t.Fatalf("fresh relation=%#v error=%v", mutations, err)
					}
					var field *string
					if mode == "field" {
						value := "task.decision_record_id"
						field = &value
					}
					kind := "references_record"
					if mode == "supersedes" {
						kind = "supersedes"
					}
					var exact bool
					if err := writer.QueryRow(ctx, `SELECT incident_id=$2 AND src_record_id=$3 AND dst_record_id=$4 AND field_key IS NOT DISTINCT FROM $5::text AND link_type=$6 AND provenance='manual' AND confidence IS NULL AND owner_user_id=$7 AND created_by_user_id=$7 AND decided_at=$8 AND created_at=$8 AND deleted_at IS NULL AND deleted_by_user_id IS NULL FROM record_links WHERE record_link_id=$1`, mutations[0].TargetID(), incident, source, target, field, kind, actor, now).Scan(&exact); err != nil || !exact {
						t.Fatalf("relation fields=%t error=%v", exact, err)
					}
					if mode != "supersedes" {
						stable := snapshot(writer)
						again, err := apply()
						if err != nil || len(again) != 0 || snapshot(writer) != stable {
							t.Fatalf("exact no-op changed relation: mutations=%#v error=%v", again, err)
						}
					}
					requireLocked(selected)
					if err := writer.Commit(ctx); err != nil {
						t.Fatal(err)
					}
					released := begin()
					if _, err := released.Exec(ctx, `SELECT record_id FROM records WHERE record_id=$1 FOR NO KEY UPDATE NOWAIT`, selected); err != nil {
						t.Fatalf("endpoint lock survived caller completion: %v", err)
					}
				})
			}
		}
	}
}

type endpointLockProjection struct {
	arrived chan pgx.Tx
	release <-chan struct{}
}

func (p endpointLockProjection) RefreshLinkedRecordsTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID, ids []uuid.UUID) error {
	if p.release != nil {
		p.arrived <- tx
		select {
		case <-p.release:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	_, err := records.LockIncidentEnvelopeIdentitiesTx(ctx, tx, incident, ids)
	return err
}

func testLinkProjectionWriterOverlap(t *testing.T, db dbport.DB, incident, actor uuid.UUID, mode string) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	target := uuid.New()
	target[0] = 0x10
	sources := []uuid.UUID{uuid.New(), uuid.New()}
	sources[0][0], sources[1][0] = 0xe0, 0xf0
	now := time.Date(2026, 11, 14, 8, 0, 0, 0, time.UTC)
	var field *string
	if mode != "upsert" {
		value := "host.subject_party_ids"
		field = &value
	}
	for _, id := range []uuid.UUID{target, sources[0], sources[1]} {
		if mode != "upsert" {
			if id != target {
				entitytest.SeedHostRecord(t, db, incident, actor, id, "Collection source", id.String(), "", "")
			} else {
				envelopetest.SeedRecordEnvelope(t, db, incident, actor, id, "party")
				if _, err := db.Exec(ctx, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Collection target','organization')`, id, incident); err != nil {
					t.Fatal(err)
				}
			}
			continue
		}
		envelopetest.SeedRecordEnvelope(t, db, incident, actor, id, "decision")
		if _, err := db.Exec(ctx, `INSERT INTO decisions(record_id,incident_id,summary,status,owner_user_id,decision_type,decided_at,rationale) VALUES($1,$2,'Endpoint','approved',$3,'containment',$4,'Retained rationale')`, id, incident, actor, now); err != nil {
			t.Fatal(err)
		}
	}
	release := make(chan struct{})
	arrived := make(chan pgx.Tx, 2)
	projection := endpointLockProjection{arrived: arrived}
	if mode == "upsert" {
		projection.release = release
	}
	store, err := links.NewMutationStore(projection)
	if err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		index     int
		mutations []links.Mutation
		err       error
	}
	done := make(chan outcome, 2)
	started, consumed := 0, 0
	defer func() {
		cancel()
		for consumed < started {
			select {
			case <-done:
				consumed++
			case <-time.After(5 * time.Second):
				t.Error("projection writers did not drain")
				return
			}
		}
	}()
	txs := make([]pgx.Tx, 2)
	pids := make([]int32, 2)
	for i := range txs {
		txs[i], err = db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		tx := txs[i]
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		if err := tx.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pids[i]); err != nil {
			t.Fatal(err)
		}
	}
	launch := func(i int) {
		started++
		go func() {
			var mutations []links.Mutation
			var err error
			if mode != "upsert" {
				if mode == "record" {
					err = store.ValidateRecordRefCollectionTx(ctx, txs[i], links.RecordRefCollectionValidation{IncidentID: incident, FieldKey: *field, LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddRecordIDs: []uuid.UUID{target, target}})
				} else {
					err = store.ValidatePartyRefCollectionTx(ctx, txs[i], links.PartyRefCollectionValidation{IncidentID: incident, FieldKey: *field, LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddPartyIDs: []uuid.UUID{target, target}})
				}
				if err == nil {
					arrived <- txs[i]
					select {
					case <-release:
					case <-ctx.Done():
						err = ctx.Err()
					}
				}
				if err != nil {
					_ = txs[i].Rollback(context.Background())
					done <- outcome{index: i, err: err}
					return
				}
			}
			switch mode {
			case "record":
				r, e := store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, txs[i], links.RecordRefCollectionCommand{IncidentID: incident, SourceRecordID: sources[i], ActorUserID: actor, FieldKey: *field, LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddRecordIDs: []uuid.UUID{target, target}, Now: now})
				mutations, err = r.Mutations(), e
			case "party":
				r, e := store.ApplyPartyRefCollectionWithMutationValuesTx(ctx, txs[i], links.PartyRefCollectionCommand{IncidentID: incident, SourceRecordID: sources[i], ActorUserID: actor, FieldKey: *field, LinkType: links.LinkTypeReferencesRecord, ExpectedTargetType: "party", AddPartyIDs: []uuid.UUID{target, target}, Now: now})
				mutations, err = r.Mutations(), e
			default:
				r, e := store.UpsertLinkCommandTx(ctx, txs[i], links.UpsertLinkCommand{IncidentID: incident, SrcRecordID: sources[i], DstRecordID: target, LinkType: links.LinkTypeReferencesRecord, Provenance: links.LinkProvenanceManual, OwnerUserID: actor, Now: now})
				err = e
				if m, ok := r.Mutation(); ok {
					mutations = append(mutations, m)
				}
			}
			if err == nil {
				err = txs[i].Commit(ctx)
			} else {
				_ = txs[i].Rollback(context.Background())
			}
			done <- outcome{i, mutations, err}
		}()
	}
	launch(0)
	select {
	case tx := <-arrived:
		if tx != txs[0] {
			t.Fatal("admission received another transaction")
		}
	case r := <-done:
		consumed++
		t.Fatalf("first writer bypassed admission: %v", r.err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	launch(1)
	for {
		var waiting bool
		if err := db.QueryRow(ctx, `SELECT $1::integer=ANY(pg_blocking_pids($2::integer))`, pids[0], pids[1]).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case tx := <-arrived:
			if tx != txs[1] {
				t.Fatal("second admission received another transaction")
			}
			waiting = true
		case r := <-done:
			consumed++
			t.Fatalf("second writer bypassed admission: %v", r.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
		if waiting {
			break
		}
	}
	close(release)
	for consumed < started {
		select {
		case r := <-done:
			consumed++
			if r.err != nil || len(r.mutations) != 1 {
				t.Errorf("writer %d: mutations=%#v error=%v", r.index, r.mutations, r.err)
				continue
			}
			m := r.mutations[0]
			if m.OperationKind() != "create" || m.TargetKind() != "record_link" || m.BeforeValue() != nil {
				t.Errorf("writer %d mutation=%#v", r.index, m)
			}
			var exact bool
			if err := db.QueryRow(ctx, `SELECT incident_id=$2 AND src_record_id=$3 AND dst_record_id=$4 AND field_key IS NOT DISTINCT FROM $7::text AND link_type='references_record' AND provenance='manual' AND confidence IS NULL AND owner_user_id=$5 AND created_by_user_id=$5 AND decided_at=$6 AND created_at=$6 AND deleted_at IS NULL AND deleted_by_user_id IS NULL FROM record_links WHERE record_link_id=$1`, m.TargetID(), incident, sources[r.index], target, actor, now, field).Scan(&exact); err != nil || !exact {
				t.Errorf("writer %d relation fields=%t error=%v", r.index, exact, err)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM record_links WHERE src_record_id=ANY($1::uuid[]) AND dst_record_id=$2`, sources, target).Scan(&count); err != nil || count != 2 {
		t.Fatalf("committed relations=%d error=%v", count, err)
	}
}
