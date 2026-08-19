package importfinalizerport

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrInitialAdminUnavailable = errors.New("incident bundle initial admin unavailable")

type Params struct {
	IncidentID        uuid.UUID
	SubmittedByUserID uuid.UUID
	PublishedAt       time.Time
	RequestID         *string
	ClientTxnID       *string
}

type Finalizer interface {
	FinalizeIncidentBundleImportTx(context.Context, pgx.Tx, Params) error
}
