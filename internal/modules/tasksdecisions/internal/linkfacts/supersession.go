package linkfacts

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// SupersessionFact is a copied retained relation, independent of endpoint visibility.
type SupersessionFact struct {
	SourceRecordID      uuid.UUID
	DestinationRecordID uuid.UUID
}

type SupersessionReader interface {
	LoadRetainedSupersessionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]SupersessionFact, error)
}
