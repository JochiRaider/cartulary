package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RouteIdempotencyKey is the Jobs-owned identity required to replay a route
// mutation. It deliberately carries no Auth persistence types.
type RouteIdempotencyKey struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

func NewRouteIdempotencyKey(routeKey string, actorUserID uuid.UUID, scopeKey string, clientTxnID string) RouteIdempotencyKey {
	return RouteIdempotencyKey{
		RouteKey: routeKey, ActorUserID: actorUserID,
		ScopeKey: scopeKey, ClientTxnID: clientTxnID,
	}
}

// RouteIdempotencyRecord is the minimum Auth-owned result Jobs needs to
// classify a replay without receiving Auth storage records.
type RouteIdempotencyRecord struct {
	RequestHash  []byte
	ResponseJSON json.RawMessage
}

// RouteIdempotencyPort preserves route replay and publication in the caller's
// transaction while leaving Auth responsible for its schema.
type RouteIdempotencyPort interface {
	LookupRouteIdempotencyTx(context.Context, pgx.Tx, RouteIdempotencyKey) (RouteIdempotencyRecord, bool, error)
	CommitRouteIdempotencyTx(context.Context, pgx.Tx, RouteIdempotencyKey, []byte, int, any) error
}

// ExtensionCancellationObservation is the minimal event passed to the
// Extensions owner when cancellation reaches an extension-owned job.
type ExtensionCancellationObservation struct {
	Key        RouteIdempotencyKey
	JobID      uuid.UUID
	ObservedAt time.Time
}

// ExtensionCancellationPort lets Extensions append its ownership evidence in
// the same transaction without exposing its tables to Jobs.
type ExtensionCancellationPort interface {
	AppendExtensionCancellationObservationTx(context.Context, pgx.Tx, ExtensionCancellationObservation) error
}

// OwnerTransactionPorts contains the consumer-owned persistence participants
// required by Jobs' cross-owner atomic operations.
type OwnerTransactionPorts struct {
	RouteIdempotency      RouteIdempotencyPort
	ExtensionCancellation ExtensionCancellationPort
}
