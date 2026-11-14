package revisions_test

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"math"
	"testing"
	"time"
)

type historyOrderQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func insertHistoryOrder(ctx context.Context, q historyOrderQuery, h revisionAllocatorHarness, explicit any) (int64, error) {
	var order int64
	err := q.QueryRow(ctx, `INSERT INTO change_sets(incident_id,actor_user_id,source,history_order) VALUES($1,$2,'order.test',$3) RETURNING history_order`, h.incidentID, h.actor.ID, explicit).Scan(&order)
	return order, err
}
func exerciseHistoryOrderAllocator(t *testing.T) {
	t.Run("import maximum, lower imported order, and rollback", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "history-order")
		for _, pair := range []struct {
			input any
			want  int64
		}{{nil, 1}, {int64(100), 100}, {int64(20), 20}, {nil, 101}} {
			if got, err := insertHistoryOrder(t.Context(), h.db, h, pair.input); err != nil || got != pair.want {
				t.Fatalf("order=%d want=%d error=%v", got, pair.want, err)
			}
		}
		tx, err := h.db.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(context.Background())
		if got, err := insertHistoryOrder(t.Context(), tx, h, nil); err != nil || got != 102 {
			t.Fatalf("uncommitted=%d %v", got, err)
		}
		if err := tx.Rollback(t.Context()); err != nil {
			t.Fatal(err)
		}
		if got, err := insertHistoryOrder(t.Context(), h.db, h, nil); err != nil || got != 102 {
			t.Fatalf("after rollback=%d %v", got, err)
		}
		if _, err := insertHistoryOrder(t.Context(), h.db, h, int64(100)); err == nil {
			t.Fatal("duplicate committed order admitted")
		}
		if got, err := insertHistoryOrder(t.Context(), h.db, h, nil); err != nil || got != 103 {
			t.Fatalf("duplicate changed allocation=%d %v", got, err)
		}
	})
	t.Run("exhaustion is atomic", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "history-order-exhausted")
		if _, err := insertHistoryOrder(t.Context(), h.db, h, int64(math.MaxInt64)); err != nil {
			t.Fatal(err)
		}
		_, err := insertHistoryOrder(t.Context(), h.db, h, nil)
		var pgerr *pgconn.PgError
		if !errors.As(err, &pgerr) || pgerr.Code != "22003" {
			t.Fatalf("exhaustion=%v", err)
		}
		var count, high int64
		if err := h.db.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM change_sets WHERE incident_id=$1),last_order FROM revision_history_orders WHERE incident_id=$1`, h.incidentID).Scan(&count, &high); err != nil || count != 1 || high != math.MaxInt64 {
			t.Fatalf("exhaustion retained count=%d high=%d error=%v", count, high, err)
		}
	})
	t.Run("assignment waits for prior transaction", func(t *testing.T) {
		h := newRevisionAllocatorHarness(t, "history-order-concurrent")
		tx, err := h.db.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(context.Background())
		if got, err := insertHistoryOrder(t.Context(), tx, h, int64(50)); err != nil || got != 50 {
			t.Fatalf("first=%d %v", got, err)
		}
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		writer, err := h.db.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer writer.Release()
		type outcome struct {
			value int64
			err   error
		}
		result := make(chan outcome, 1)
		finished := make(chan struct{})
		go func() {
			defer close(finished)
			got, err := insertHistoryOrder(ctx, writer, h, nil)
			result <- outcome{got, err}
		}()
		defer func() { cancel(); <-finished }()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			var blocked bool
			if err := h.db.QueryRow(ctx, `SELECT $1::integer=ANY(pg_blocking_pids($2::integer))`, int32(tx.Conn().PgConn().PID()), int32(writer.Conn().PgConn().PID())).Scan(&blocked); err != nil {
				t.Fatal(err)
			}
			if blocked {
				break
			}
			select {
			case got := <-result:
				t.Fatalf("assignment escaped commit serialization: %v", got)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-ticker.C:
			}
		}
		if err := tx.Commit(t.Context()); err != nil {
			t.Fatal(err)
		}
		select {
		case got := <-result:
			if got.err != nil || got.value != 51 {
				t.Fatalf("successor=%v", got)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	})
}
