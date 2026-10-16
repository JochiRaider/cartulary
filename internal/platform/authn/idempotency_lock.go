package authn

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LockRouteIdempotencyTx serializes admission of one route identity, including
// when no receipt exists yet. Call it before any owner mutation guards. The
// transaction releases the lock; semantic replay still compares the full key
// and request hash, so a hash collision can only cause additional waiting.
func LockRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key RouteIdempotencyKey) error {
	if tx == nil {
		return errors.New("auth: route idempotency transaction is nil")
	}
	identity, err := routeIdempotencyLockIdentity(key)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, identity)
	return err
}

func routeIdempotencyLockIdentity(key RouteIdempotencyKey) (int64, error) {
	if key.ActorUserID == uuid.Nil || key.RouteKey == "" || key.ScopeKey == "" || key.ClientTxnID == "" {
		return 0, errors.New("auth: incomplete route idempotency identity")
	}
	encoded, err := json.Marshal([]string{"cartulary.auth.route-idempotency.v1", key.RouteKey, key.ActorUserID.String(), key.ScopeKey, key.ClientTxnID})
	if err != nil {
		return 0, err
	}
	digest := sha256.Sum256(encoded)
	return int64(binary.BigEndian.Uint64(digest[:8]) & 0x7fffffffffffffff), nil
}

// BeginRepeatableRouteIdempotencyTx acquires admission before the consistent
// view is established. Waiting on a transaction-level SELECT inside repeatable
// read would freeze an earlier view and miss the competing committed receipt.
// A session lock bridges BEGIN to the same transaction lock without a gap. It
// is released before returning; uncertain acquisition/handoff destroys the
// connection so no session lock can escape into the pool. Call release even
// after Commit; it also rolls back unfinished work independently of cancellation.
func BeginRepeatableRouteIdempotencyTx(ctx context.Context, pool *pgxpool.Pool, key RouteIdempotencyKey) (pgx.Tx, func(), error) {
	identity, err := routeIdempotencyLockIdentity(key)
	if err != nil {
		return nil, nil, err
	}
	if pool == nil {
		return nil, nil, errors.New("auth: route idempotency pool is nil")
	}
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}
	discard := func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = conn.Conn().Close(cleanup)
		conn.Release()
	}
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, identity); err != nil {
		discard()
		return nil, nil, err
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		discard()
		return nil, nil, err
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, identity); err != nil {
		discard()
		return nil, nil, err
	}
	var unlocked bool
	if err := tx.QueryRow(ctx, `SELECT pg_advisory_unlock($1)`, identity).Scan(&unlocked); err != nil || !unlocked {
		discard()
		return nil, nil, errors.Join(err, errors.New("auth: route admission lock handoff failed"))
	}
	release := func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tx.Rollback(cleanup); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			_ = conn.Conn().Close(cleanup)
		}
		conn.Release()
	}
	return tx, release, nil
}
