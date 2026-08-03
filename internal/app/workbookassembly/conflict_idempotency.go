package workbookassembly

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type conflictIdempotencyAdapter struct {
	store *authn.Store
}

func NewConflictIdempotencyPort(database postgres.DB) conflicts.IdempotencyPort {
	return conflictIdempotencyAdapter{store: authn.NewStore(database)}
}

func (adapter conflictIdempotencyAdapter) Get(
	ctx context.Context,
	key conflicts.IdempotencyKey,
	requestHash []byte,
) (conflicts.IdempotencyRecord, error) {
	record, err := adapter.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return conflicts.IdempotencyRecord{}, conflicts.ErrIdempotencyNotFound
	}
	if err != nil {
		return conflicts.IdempotencyRecord{}, err
	}
	if !bytes.Equal(record.RequestHash, requestHash) {
		return conflicts.IdempotencyRecord{RequestHash: record.RequestHash}, nil
	}
	target, err := conflicts.DecodeStoredTarget(record.ResponseJSON)
	if err != nil {
		return conflicts.IdempotencyRecord{}, err
	}
	return conflicts.IdempotencyRecord{RequestHash: record.RequestHash, Target: target}, nil
}

func (conflictIdempotencyAdapter) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key conflicts.IdempotencyKey,
	requestHash []byte,
	target conflicts.StoredTarget,
) error {
	err := authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, http.StatusOK, map[string]any{
		"view_schema_id": target.ViewSchemaID,
		"row":            target.Row,
	})
	if authn.IsUniqueViolation(err) {
		return conflicts.ErrClientTxnConflict
	}
	return err
}
