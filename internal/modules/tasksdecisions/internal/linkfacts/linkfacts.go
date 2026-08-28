package linkfacts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Fact is the consumer-owned scalar projection of one active Links row.
// Optional field keys use an explicit presence bit so no pointer ownership
// crosses the capability boundary.
type Fact struct {
	SourceRecordID      uuid.UUID
	DestinationRecordID uuid.UUID
	LinkType            string
	FieldKey            string
	HasFieldKey         bool
}

// Capability supplies active Links-owned facts in the caller's transaction.
// Implementations return a non-nil empty slice and no partial facts on error.
type Capability interface {
	LoadRecordLinkFactsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]Fact, error)
}
