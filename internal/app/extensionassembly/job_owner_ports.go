package extensionassembly

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// JobOwnerTransactionAdapters is the application composition boundary for
// atomic operations spanning Jobs and its Auth/Extensions consumers.
type JobOwnerTransactionAdapters struct{}

func (JobOwnerTransactionAdapters) LookupRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey) (jobs.RouteIdempotencyRecord, bool, error) {
	record, err := authn.GetRouteIdempotencyTx(ctx, tx, authRouteIdempotencyKey(key))
	if errors.Is(err, authn.ErrNotFound) {
		return jobs.RouteIdempotencyRecord{}, false, nil
	}
	if err != nil {
		return jobs.RouteIdempotencyRecord{}, false, err
	}
	return jobs.RouteIdempotencyRecord{
		RequestHash:  append([]byte(nil), record.RequestHash...),
		ResponseJSON: append([]byte(nil), record.ResponseJSON...),
	}, true, nil
}

func (JobOwnerTransactionAdapters) CommitRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, requestHash []byte, statusCode int, payload any) error {
	return authn.InsertRouteIdempotencyPayload(ctx, tx, authRouteIdempotencyKey(key), nil, requestHash, statusCode, payload)
}

func (JobOwnerTransactionAdapters) ReconcileFinalIdempotencyOutcomeTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, requestHash []byte, resource jobs.Resource) (bool, error) {
	return authn.UpdateRouteIdempotencyPayload(ctx, tx, authRouteIdempotencyKey(key), requestHash, resource)
}

func (JobOwnerTransactionAdapters) AppendExtensionCancellationObservationTx(ctx context.Context, tx pgx.Tx, observation jobs.ExtensionCancellationObservation) error {
	key := observation.Key
	return extensionstore.AppendRouteJobCancellationObservationTx(ctx, tx, extensionstore.RouteCancellationIdentity{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, observation.JobID, observation.ObservedAt)
}

func authRouteIdempotencyKey(key jobs.RouteIdempotencyKey) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}
}
