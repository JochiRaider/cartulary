package recovery

import (
	"context"
	"io"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// ReferencePackStorage is a separately confined deployment root. Recovery
// dispatches only this owner's object family through it; generic object-store
// keys must never select or escape into this storage capability.
type ReferencePackStorage interface {
	VNextObjectSource
	RestoreMember(context.Context, string, string, int64, io.Reader) error
	RequireEmpty(context.Context) error
	ResetVerificationTarget(context.Context) error
	ValidateHistoricalState(context.Context, postgres.DB) error
	Close()
}
