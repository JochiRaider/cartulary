package revisions_test

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

type historyOrderQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

var historyOrderClock = time.Date(2026, time.November, 18, 9, 0, 0, 0, time.UTC)

// allocateHistoryOrderAtStatementEnd serves transaction fixtures, which run
// every logical commit inside one transaction that never commits. Firing the
// commit-time allocator at each statement end assigns the orders those
// sequential logical commits would receive. Outside a transaction block the
// statement only warns, so committed fixtures keep real commit allocation.
func allocateHistoryOrderAtStatementEnd(t testing.TB, db interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}) {
	t.Helper()
	if _, err := db.Exec(context.Background(), `SET CONSTRAINTS change_sets_allocate_history_order IMMEDIATE`); err != nil {
		t.Fatalf("fire commit-time history order allocation: %v", err)
	}
}

// appendNativeChangeSet uses the production native writer, which never
// supplies an order. The clock runs backwards so timestamps cannot explain
// any order the test observes.
func appendNativeChangeSet(ctx context.Context, tx pgx.Tx, incidentID, actorID uuid.UUID, step int) (uuid.UUID, error) {
	return (&revisions.Appender{}).AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: incidentID, ActorUserID: actorID, Source: "order.test",
		CreatedAt: historyOrderClock.Add(-time.Duration(step) * time.Minute),
	})
}

func commitNativeChangeSet(t *testing.T, h revisionAllocatorHarness, step int) uuid.UUID {
	t.Helper()
	tx, err := h.db.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	id, err := appendNativeChangeSet(t.Context(), tx, h.incidentID, h.actor.ID, step)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertExplicitHistoryOrder(ctx context.Context, q historyOrderQuery, h revisionAllocatorHarness, order int64) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.QueryRow(ctx, `INSERT INTO change_sets(incident_id,actor_user_id,source,history_order) VALUES($1,$2,'order.import',$3) RETURNING change_set_id`, h.incidentID, h.actor.ID, order).Scan(&id)
	return id, err
}

func storedHistoryOrder(t *testing.T, q historyOrderQuery, changeSetID uuid.UUID) pgtype.Int8 {
	t.Helper()
	var order pgtype.Int8
	if err := q.QueryRow(t.Context(), `SELECT history_order FROM change_sets WHERE change_set_id=$1`, changeSetID).Scan(&order); err != nil {
		t.Fatal(err)
	}
	return order
}

func requireHistoryOrders(t *testing.T, q historyOrderQuery, want map[uuid.UUID]int64) {
	t.Helper()
	for id, order := range want {
		if got := storedHistoryOrder(t, q, id); !got.Valid || got.Int64 != order {
			t.Fatalf("change set %s history_order=%v want %d", id, got, order)
		}
	}
}

func requireAllocatorHighWater(t *testing.T, q historyOrderQuery, incidentID uuid.UUID, want int64) {
	t.Helper()
	var got int64
	if err := q.QueryRow(t.Context(), `SELECT last_order FROM revision_history_orders WHERE incident_id=$1`, incidentID).Scan(&got); err != nil || got != want {
		t.Fatalf("allocator last_order=%d want %d: %v", got, want, err)
	}
}

// Every committed change set must carry an order (REQ-02-213), whatever
// writer, isolation or failure produced the surrounding history.
func requireEveryCommittedChangeSetOrdered(t *testing.T, q historyOrderQuery) {
	t.Helper()
	var unordered int
	if err := q.QueryRow(t.Context(), `SELECT count(*) FROM change_sets WHERE history_order IS NULL`).Scan(&unordered); err != nil || unordered != 0 {
		t.Fatalf("committed change sets without history_order=%d: %v", unordered, err)
	}
}

// requireUnblocked runs work on its own connection and fails as soon as
// PostgreSQL reports that connection waiting for another backend.
func requireUnblocked(t *testing.T, observer *pgxpool.Pool, pid uint32, work func(context.Context) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- work(ctx) }()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
			return
		case <-ticker.C:
		}
		var blockers []int32
		if err := observer.QueryRow(ctx, `SELECT pg_blocking_pids($1::integer)`, int32(pid)).Scan(&blockers); err != nil {
			t.Fatal(err)
		}
		if len(blockers) > 0 {
			cancel()
			<-done
			t.Fatalf("change-set writer waited on in-progress backends %v", blockers)
		}
	}
}

// awaitBlockedBy returns once PostgreSQL reports waiter blocked by blocker.
func awaitBlockedBy(t *testing.T, observer *pgxpool.Pool, waiter, blocker uint32, done <-chan error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		if err := observer.QueryRow(ctx, `SELECT $1::integer = ANY(pg_blocking_pids($2::integer))`, int32(blocker), int32(waiter)).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked {
			return
		}
		select {
		case err := <-done:
			t.Fatalf("commit finished without waiting for the allocator lock: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
	}
}

func beginOnConnection(t *testing.T, db *pgxpool.Pool) (*pgxpool.Conn, pgx.Tx) {
	t.Helper()
	conn, err := db.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(conn.Release)
	tx, err := conn.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return conn, tx
}

func TestRevisionsHistoryOrderCommitAllocation_Integration(t *testing.T) {
	t.Run("native orders follow commit order", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "commit-order")
		first := commitNativeChangeSet(t, h, 0)
		_, early := beginOnConnection(t, h.db)
		earlyID, err := appendNativeChangeSet(t.Context(), early, h.incidentID, h.actor.ID, 1)
		if err != nil {
			t.Fatal(err)
		}
		if got := storedHistoryOrder(t, early, earlyID); got.Valid {
			t.Fatalf("uncommitted native change set already has history_order=%d", got.Int64)
		}
		lateConn, late := beginOnConnection(t, h.db)
		var lateID uuid.UUID
		requireUnblocked(t, h.db, lateConn.Conn().PgConn().PID(), func(ctx context.Context) error {
			var err error
			lateID, err = appendNativeChangeSet(ctx, late, h.incidentID, h.actor.ID, 2)
			if err != nil {
				return err
			}
			return late.Commit(ctx)
		})
		if err := early.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		// The later insert committed first and therefore precedes the earlier one.
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{first: 1, lateID: 2, earlyID: 3})
		requireAllocatorHighWater(t, h.db, h.incidentID, 3)
		requireEveryCommittedChangeSetOrdered(t, h.db)
	})

	t.Run("an in-progress writer does not block another commit", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "in-progress-writer")
		first := commitNativeChangeSet(t, h, 0)
		// The held transaction stands in for an import apply unit that has
		// appended its change set and is still writing owner rows.
		_, held := beginOnConnection(t, h.db)
		heldID, err := appendNativeChangeSet(t.Context(), held, h.incidentID, h.actor.ID, 1)
		if err != nil {
			t.Fatal(err)
		}
		for sequence := 1; sequence <= 3; sequence++ {
			if _, err := held.Exec(t.Context(), `INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,history_record_ids,history_entry_record_ids) VALUES($1,$2,'host',$3,'create',$4,$4)`, heldID, sequence, h.recordID.String(), []uuid.UUID{h.recordID}); err != nil {
				t.Fatal(err)
			}
		}
		editorConn, editor := beginOnConnection(t, h.db)
		var editID uuid.UUID
		requireUnblocked(t, h.db, editorConn.Conn().PgConn().PID(), func(ctx context.Context) error {
			var err error
			editID, err = appendNativeChangeSet(ctx, editor, h.incidentID, h.actor.ID, 2)
			if err != nil {
				return err
			}
			return editor.Commit(ctx)
		})
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{first: 1, editID: 2})
		if err := held.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{heldID: 3})
		requireEveryCommittedChangeSetOrdered(t, h.db)
	})

	t.Run("commit takes the allocator lock", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "commit-lock")
		first := commitNativeChangeSet(t, h, 0)
		holderConn, holder := beginOnConnection(t, h.db)
		if _, err := holder.Exec(t.Context(), `SELECT last_order FROM revision_history_orders WHERE incident_id=$1 FOR UPDATE`, h.incidentID); err != nil {
			t.Fatal(err)
		}
		writerConn, writer := beginOnConnection(t, h.db)
		var writerID uuid.UUID
		// Appending takes no allocator lock; only the commit serializes.
		requireUnblocked(t, h.db, writerConn.Conn().PgConn().PID(), func(ctx context.Context) error {
			var err error
			writerID, err = appendNativeChangeSet(ctx, writer, h.incidentID, h.actor.ID, 1)
			return err
		})
		done := make(chan error, 1)
		go func() { done <- writer.Commit(t.Context()) }()
		awaitBlockedBy(t, h.db, writerConn.Conn().PgConn().PID(), holderConn.Conn().PgConn().PID(), done)
		if err := holder.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		if err := <-done; err != nil {
			t.Fatal(err)
		}
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{first: 1, writerID: 2})
		requireEveryCommittedChangeSetOrdered(t, h.db)
	})

	t.Run("rolled back allocation leaves no gap", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "rollback-gap")
		first := commitNativeChangeSet(t, h, 0)
		for _, allocateEarly := range []bool{false, true} {
			tx, err := h.db.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			id, err := appendNativeChangeSet(t.Context(), tx, h.incidentID, h.actor.ID, 1)
			if err != nil {
				t.Fatal(err)
			}
			if allocateEarly {
				// Firing the deferred allocator before rollback proves that an
				// allocation already made is undone with its transaction.
				if _, err := tx.Exec(t.Context(), `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
					t.Fatal(err)
				}
				requireHistoryOrders(t, tx, map[uuid.UUID]int64{id: 2})
				requireAllocatorHighWater(t, tx, h.incidentID, 2)
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
		}
		next := commitNativeChangeSet(t, h, 2)
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{first: 1, next: 2})
		requireAllocatorHighWater(t, h.db, h.incidentID, 2)
		requireEveryCommittedChangeSetOrdered(t, h.db)
	})

	t.Run("explicit orders are preserved and followed", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "explicit-order")
		first := commitNativeChangeSet(t, h, 0)
		tx, err := h.db.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		// Portable orders may arrive out of order; the high water never falls.
		high, err := insertExplicitHistoryOrder(t.Context(), tx, h, 100)
		if err != nil {
			t.Fatal(err)
		}
		low, err := insertExplicitHistoryOrder(t.Context(), tx, h, 20)
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{first: 1, high: 100, low: 20})
		requireAllocatorHighWater(t, h.db, h.incidentID, 100)
		next := commitNativeChangeSet(t, h, 1)
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{next: 101})
		_, err = insertExplicitHistoryOrder(t.Context(), h.db, h, 100)
		var duplicate *pgconn.PgError
		if !errors.As(err, &duplicate) || duplicate.Code != "23505" || duplicate.ConstraintName != "change_sets_incident_history_order_key" {
			t.Fatalf("duplicate committed order=%v", err)
		}
		after := commitNativeChangeSet(t, h, 2)
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{after: 102})
		requireEveryCommittedChangeSetOrdered(t, h.db)
	})

	t.Run("allocation failure rolls back the whole transaction", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "exhausted")
		maximum, err := insertExplicitHistoryOrder(t.Context(), h.db, h, math.MaxInt64)
		if err != nil {
			t.Fatal(err)
		}
		tx, err := h.db.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		id, err := appendNativeChangeSet(t.Context(), tx, h.incidentID, h.actor.ID, 1)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(t.Context(), `INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,history_record_ids,history_entry_record_ids) VALUES($1,1,'host',$2,'create',$3,$3)`, id, h.recordID.String(), []uuid.UUID{h.recordID}); err != nil {
			t.Fatal(err)
		}
		err = tx.Commit(t.Context())
		var exhausted *pgconn.PgError
		if !errors.As(err, &exhausted) || exhausted.Code != "22003" || exhausted.Message != "revision history order is exhausted" {
			t.Fatalf("exhausted commit=%v", err)
		}
		var changeSets, mutations int
		if err := h.db.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM change_sets WHERE incident_id=$1),(SELECT count(*) FROM change_set_mutations WHERE change_set_id=$2)`, h.incidentID, id).Scan(&changeSets, &mutations); err != nil || changeSets != 1 || mutations != 0 {
			t.Fatalf("failed commit retained change_sets=%d mutations=%d: %v", changeSets, mutations, err)
		}
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{maximum: math.MaxInt64})
		requireAllocatorHighWater(t, h.db, h.incidentID, math.MaxInt64)
		requireEveryCommittedChangeSetOrdered(t, h.db)
	})

	t.Run("runtime role allocates at commit", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "runtime-role")
		tx, err := h.db.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		// The allocator needs only the runtime role's table privileges, never
		// an EXECUTE grant on its private routine.
		if _, err := tx.Exec(t.Context(), `SET LOCAL ROLE cartulary_runtime`); err != nil {
			t.Fatal(err)
		}
		id, err := appendNativeChangeSet(t.Context(), tx, h.incidentID, h.actor.ID, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		requireHistoryOrders(t, h.db, map[uuid.UUID]int64{id: 1})
		requireEveryCommittedChangeSetOrdered(t, h.db)
	})

	t.Run("bundle import orders are preserved and followed", func(t *testing.T) {
		h := newRevisionsPortabilityHarness(t, "import-order")
		bundle := h.validBundle(t)
		rows := decodePortableRows(t, bundle[revisionsTestChangeSetsPath])
		rows[0]["history_order"] = json.Number("100")
		bundle[revisionsTestChangeSetsPath] = encodePortableRows(t, rows)
		input := h.freshImportContext()
		prepared, err := h.port.PrepareImport(t.Context(), bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		tx, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(context.Background()) }()
		if err := h.port.ApplyImportTx(t.Context(), tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		if err := h.port.ValidateImportTx(t.Context(), tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		native, err := h.db.BeginTx(t.Context(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = native.Rollback(context.Background()) }()
		nativeID, err := appendNativeChangeSet(t.Context(), native, h.incidentID, h.actor.ID, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := native.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		exportOrders := func() (map[string]any, error) {
			read, err := h.db.BeginTx(t.Context(), pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = read.Rollback(context.Background()) }()
			files, err := portabilitysupport.Export(h.port, t.Context(), sourceport.ExportContext{Query: read, IncidentID: h.incidentID})
			if err != nil {
				return nil, err
			}
			orders := map[string]any{}
			for _, file := range files {
				if file.Path == revisionsTestChangeSetsPath {
					for _, row := range decodePortableRows(t, file.Payload) {
						orders[row["change_set_id"].(string)] = row["history_order"]
					}
				}
			}
			return orders, nil
		}
		orders, err := exportOrders()
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]any{h.changeSet.String(): json.Number("100"), nativeID.String(): json.Number("101")}
		if len(orders) != len(want) || orders[h.changeSet.String()] != want[h.changeSet.String()] || orders[nativeID.String()] != want[nativeID.String()] {
			t.Fatalf("exported orders=%v want %v", orders, want)
		}
		// A committed change set without an order is an invariant break that
		// export reports, never a zero or inferred order.
		if _, err := h.db.Exec(t.Context(), `UPDATE change_sets SET history_order=NULL WHERE change_set_id=$1`, nativeID); err != nil {
			t.Fatal(err)
		}
		if _, err := exportOrders(); !errors.Is(err, revisions.ErrHistoryOrderMissing) {
			t.Fatalf("export of unordered committed history=%v", err)
		}
	})
}
