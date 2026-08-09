package artifacts

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type artifactIncidentAdmissionPort interface {
	EnsureOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type artifactRecordEnvelopePort interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (records.Envelope, error)
}

type artifactRevisionPort interface {
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, revisions.AppendMutationParams) error
	AppendRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
}

type artifactRevisionHistoryPort interface {
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicts.RevisionWindowRow, error)
}

type artifactSourceMutationPort interface {
	InsertRowTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, CreateParams, time.Time) error
	ApplyDirectChangeTx(context.Context, pgx.Tx, uuid.UUID, string, FieldValue, time.Time) (bool, error)
	ApplyHandoffRiskRefPayloadTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, RiskRefActionPayload, time.Time) (bool, error)
	NormalizeFindingLifecycleTx(context.Context, pgx.Tx, uuid.UUID, time.Time) (bool, error)
	TouchRowTx(context.Context, pgx.Tx, uuid.UUID, time.Time) error
}

type artifactIdempotencyPort interface {
	Get(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error)
	PutTx(context.Context, pgx.Tx, authn.RouteIdempotencyKey, []byte, int, any) error
}

type artifactIdempotencyAdapter struct {
	store *authn.Store
}

func (a artifactIdempotencyAdapter) Get(
	ctx context.Context,
	key authn.RouteIdempotencyKey,
) (authn.RouteIdempotencyRecord, error) {
	return a.store.GetRouteIdempotency(ctx, key)
}

func (artifactIdempotencyAdapter) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key authn.RouteIdempotencyKey,
	requestHash []byte,
	statusCode int,
	payload any,
) error {
	return authn.InsertRouteIdempotencyPayload(
		ctx,
		tx,
		key,
		nil,
		requestHash,
		statusCode,
		payload,
	)
}
