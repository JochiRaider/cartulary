// Package artifactport defines the Incident Bundles transport boundary for
// historical owner artifacts. Implementations validate and publish their own
// state; transport never inspects the owner's tables or interprets its bytes.
package artifactport

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type WriteFile func(context.Context, string, int64, io.Reader) error

type Bundle interface {
	File(string) ([]byte, bool)
	Paths() []string
}

type ImportRequest struct {
	IncidentID  uuid.UUID
	OperationID uuid.UUID
	References  []byte
	Bundle      Bundle
}

// Prepared is opaque owner state. It publishes only within the parent's
// transaction after incident and Reference Pack retention publication.
type Prepared interface {
	ApplyTx(context.Context, pgx.Tx) error
}

// ErrInvalid is deliberately value-free. No hostile member name, payload value,
// storage reference or source identifier enters the public diagnostic.
var ErrInvalid = errors.New("incident bundle historical artifacts are invalid")
