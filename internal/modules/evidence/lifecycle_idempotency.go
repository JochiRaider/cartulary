package evidence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// LifecycleOperationID is the closed set of Evidence lifecycle operations
// whose durable replay records are owned by the application idempotency store.
type LifecycleOperationID string

const (
	LifecycleOperationBlobCreate LifecycleOperationID = "object_blobs.create"
	LifecycleOperationBlobAttach LifecycleOperationID = "evidence.attach_blob"
)

type LifecycleIdempotencyKey struct {
	OperationID LifecycleOperationID
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

// LifecycleIdempotencyCapability hides the shared authentication store and its
// transport-shaped record from Evidence while retaining the established
// durable replay payload.
type LifecycleIdempotencyCapability interface {
	Get(context.Context, LifecycleIdempotencyKey, []byte) (map[string]any, bool, error)
	GetTx(context.Context, pgx.Tx, LifecycleIdempotencyKey, []byte) (map[string]any, bool, error)
	PutTx(context.Context, pgx.Tx, LifecycleIdempotencyKey, []byte, map[string]any) error
}
