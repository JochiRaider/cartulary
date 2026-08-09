package evidence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
)

type evidenceIncidentAdmissionPort interface {
	EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type evidenceRecordEnvelopePort interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
}

type evidenceSourceMutationPort interface {
	InsertWorkbookRowTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, WorkbookCreateParams, time.Time) error
	ValidateWorkbookLifecyclePatchTx(context.Context, pgx.Tx, uuid.UUID, []WorkbookLifecyclePatchChange) error
	ApplyWorkbookDirectChangeTx(context.Context, pgx.Tx, uuid.UUID, string, WorkbookFieldValue, time.Time) (bool, error)
	TouchWorkbookRowTx(context.Context, pgx.Tx, uuid.UUID, time.Time) error
}
