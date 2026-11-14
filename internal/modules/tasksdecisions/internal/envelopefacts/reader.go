package envelopefacts

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Fact is the current envelope identity needed for portable source admission.
type Fact struct {
	IncidentID uuid.UUID
	RecordType string
	Deleted    bool
}

type Reader interface {
	LoadEnvelopeFactTx(context.Context, pgx.Tx, uuid.UUID) (Fact, bool, error)
}
