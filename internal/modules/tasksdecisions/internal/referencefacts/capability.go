package referencefacts

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/envelopefacts"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Capability supplies only peer-owned facts and locks required by Task references.
type Capability interface {
	LockEnvelopeFactTx(context.Context, pgx.Tx, uuid.UUID, string, bool) (envelopefacts.Fact, bool, error)
	LockPartySourceTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, bool) (bool, error)
}
