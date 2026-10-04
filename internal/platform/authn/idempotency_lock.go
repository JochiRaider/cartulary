package authn

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LockRouteIdempotencyTx serializes admission of one route identity, including
// when no receipt exists yet. Call it before any owner mutation guards. The
// transaction releases the lock; semantic replay still compares the full key
// and request hash, so a hash collision can only cause additional waiting.
func LockRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key RouteIdempotencyKey) error {
	if tx == nil || key.ActorUserID == uuid.Nil || key.RouteKey == "" || key.ScopeKey == "" || key.ClientTxnID == "" {
		return errors.New("auth: incomplete route idempotency identity")
	}
	encoded, err := json.Marshal([]string{"cartulary.auth.route-idempotency.v1", key.RouteKey, key.ActorUserID.String(), key.ScopeKey, key.ClientTxnID})
	if err != nil {
		return err
	}
	digest := sha256.Sum256(encoded)
	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(binary.BigEndian.Uint64(digest[:8])&0x7fffffffffffffff))
	return err
}
