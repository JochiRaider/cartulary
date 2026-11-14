package linkfacts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type FieldReferenceFact struct {
	DestinationRecordID uuid.UUID
	LinkType            string
}

// FieldReferenceReader supplies non-deleted owner-field bindings independently
// of endpoint visibility. Errors return no partial facts; empty success is non-nil.
type FieldReferenceReader interface {
	LoadFieldReferencesTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, string) ([]FieldReferenceFact, error)
}
