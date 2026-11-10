package providercontract

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// RecordEndpoint contains only the identity needed to refresh linked providers.
type RecordEndpoint struct {
	RecordID   uuid.UUID
	RecordType string
}

// LockRecordEndpoints borrows the source transaction and returns incident-local
// envelopes in deterministic identity order, locked for dependent refresh.
type LockRecordEndpoints func(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) ([]RecordEndpoint, error)
