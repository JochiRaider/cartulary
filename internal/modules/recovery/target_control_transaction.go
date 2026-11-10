package recovery

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

type TargetControlTransaction interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// TargetControlCapability binds durable control publication to the original
// exclusive serving session. Payload transactions remain separate.
type TargetControlCapability interface {
	Context() context.Context
	AssertHeld() error
	TargetDatabase() dbport.DB
	WithControlTransaction(context.Context, func(context.Context, TargetControlTransaction) error) error
}
