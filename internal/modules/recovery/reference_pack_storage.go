package recovery

import (
	"context"
	"io"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// RootObjectStorage confines recovery to an explicitly admitted filesystem root.
// Dispatch is selected by the admitted owner family, never by an object key.
type RootObjectStorage interface {
	VNextObjectSource
	RestoreMember(context.Context, string, string, int64, io.Reader) error
	RequireEmpty(context.Context) error
	ResetVerificationTarget(context.Context) error
	Close()
}

// ReferencePackStorage also validates the source owner's retained semantic state.
type ReferencePackStorage interface {
	RootObjectStorage
	ValidateHistoricalState(context.Context, postgres.DB) error
}
