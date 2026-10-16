package authn_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func assertRepeatableRouteAdmissionLifecycle(t *testing.T, pool *pgxpool.Pool, actorA, actorB uuid.UUID) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	key := authn.RouteIdempotencyKey{RouteKey: "test.repeatable", ActorUserID: actorA, ScopeKey: "scope", ClientTxnID: "txn"}
	holder, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = holder.Rollback(context.Background()) }()
	if err := authn.LockRouteIdempotencyTx(ctx, holder, key); err != nil {
		t.Fatal(err)
	}
	waitCtx, stopWaiting := context.WithCancel(ctx)
	defer stopWaiting()
	type result struct {
		tx      pgx.Tx
		release func()
		err     error
	}
	ready := make(chan result, 1)
	go func() {
		tx, release, err := authn.BeginRepeatableRouteIdempotencyTx(waitCtx, pool, key)
		ready <- result{tx, release, err}
	}()
	for {
		var waiting bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, int32(holder.Conn().PgConn().PID())).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case early := <-ready:
			if early.release != nil {
				early.release()
			}
			t.Fatal("same-key admission escaped held transaction lock", early.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	stopWaiting()
	select {
	case canceled := <-ready:
		if canceled.release != nil {
			canceled.release()
		}
		if canceled.tx != nil || !errors.Is(canceled.err, context.Canceled) {
			t.Fatal("waiting admission did not cancel cleanly", canceled.err)
		}
	case <-ctx.Done():
		t.Fatal("waiting admission did not finish", ctx.Err())
	}
	if err := holder.Rollback(ctx); err != nil {
		t.Fatal(err)
	}

	// An abandoned transaction must roll back even if the caller is canceled.
	abandonCtx, abandon := context.WithCancel(ctx)
	tx, release, err := authn.BeginRepeatableRouteIdempotencyTx(abandonCtx, pool, key)
	if err != nil {
		t.Fatal(err)
	}
	var isolation string
	if err := tx.QueryRow(ctx, `SHOW transaction_isolation`).Scan(&isolation); err != nil || isolation != "repeatable read" {
		release()
		t.Fatal("lost consistent view", isolation, err)
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, []byte("request"), 200, map[string]any{"abandoned": true}); err != nil {
		release()
		t.Fatal(err)
	}
	abandon()
	release()
	if _, err := authn.NewStore(pool).GetRouteIdempotency(ctx, key); !errors.Is(err, authn.ErrNotFound) {
		t.Fatal("canceled unfinished admission committed a receipt", err)
	}

	tx, release, err = authn.BeginRepeatableRouteIdempotencyTx(ctx, pool, key)
	if err != nil {
		t.Fatal("admission lock leaked after cancellation", err)
	}
	defer release()
	pid := int32(tx.Conn().PgConn().PID())
	other := key
	other.ActorUserID = actorB
	independent, releaseOther, err := authn.BeginRepeatableRouteIdempotencyTx(ctx, pool, other)
	if err != nil {
		t.Fatal("another actor was blocked by this receipt scope", err)
	}
	if err := independent.Commit(ctx); err != nil {
		releaseOther()
		t.Fatal(err)
	}
	releaseOther()
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var locks int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM pg_locks WHERE pid=$1 AND locktype='advisory'`, pid).Scan(&locks); err != nil || locks != 0 {
		t.Fatal("committed admission retained a session lock", locks, err)
	}
}
