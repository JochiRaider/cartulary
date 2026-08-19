package bootstrapport

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// InitialPreferenceInput contains the complete authored state for the two
// Workbook preference rows created with an incident. The caller owns the
// transaction and timestamp.
type InitialPreferenceInput struct {
	IncidentID      uuid.UUID
	UserID          uuid.UUID
	CommitTimestamp time.Time
}

// Writer inserts the initial Workbook preference rows in a caller-owned
// transaction. It does not own transaction, retry, authorization, or audit
// behavior.
type Writer interface {
	InsertInitialTx(context.Context, pgx.Tx, InitialPreferenceInput) error
}
