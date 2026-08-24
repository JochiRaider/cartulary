package indicatorassembly

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type idempotencyPort struct {
	store *authn.Store
}

func NewIdempotencyPort(store *authn.Store) indicators.IdempotencyPort {
	if store == nil {
		return nil
	}
	return idempotencyPort{store: store}
}

func (port idempotencyPort) GetRouteIdempotency(ctx context.Context, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	return port.store.GetRouteIdempotency(ctx, key)
}

func (port idempotencyPort) InsertRouteIdempotencyPayload(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, requestHash []byte, statusCode int, payload any) error {
	return authn.InsertRouteIdempotencyPayload(ctx, tx, key, nil, requestHash, statusCode, payload)
}
