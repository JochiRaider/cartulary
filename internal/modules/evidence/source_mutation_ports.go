package evidence

// Source mutation ports keep transaction coordination owner-private.

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type evidenceIncidentAdmissionPort interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type evidenceRecordEnvelopePort interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
}

type evidenceSourceMutationPort interface {
	insertRowTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, createParams, time.Time) error
	validateLifecyclePatchTx(context.Context, pgx.Tx, uuid.UUID, []lifecyclePatchChange) error
	applyDirectChangeTx(context.Context, pgx.Tx, uuid.UUID, string, FieldValue, time.Time) (bool, error)
	touchRowTx(context.Context, pgx.Tx, uuid.UUID, time.Time) error
}
