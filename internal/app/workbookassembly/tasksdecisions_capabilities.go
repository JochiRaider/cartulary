package workbookassembly

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type taskDecisionIdempotency struct {
	store *authn.Store
}

func (a taskDecisionIdempotency) Get(
	ctx context.Context,
	key tasksdecisions.IdempotencyKey,
) (tasksdecisions.IdempotencyRecord, error) {
	record, err := a.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return tasksdecisions.IdempotencyRecord{}, tasksdecisions.ErrIdempotencyNotFound
	}
	if err != nil {
		return tasksdecisions.IdempotencyRecord{}, err
	}
	return tasksdecisions.IdempotencyRecord{
		RequestHash:  record.RequestHash,
		ResponseJSON: record.ResponseJSON,
	}, nil
}

func (taskDecisionIdempotency) PutTx(
	ctx context.Context,
	tx pgx.Tx,
	key tasksdecisions.IdempotencyKey,
	requestHash []byte,
	outcome tasksdecisions.IdempotencyOutcome,
	payload any,
) error {
	status := http.StatusOK
	if outcome == tasksdecisions.IdempotencyOutcomeCreated {
		status = http.StatusCreated
	}
	err := authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, status, payload)
	return tasksdecisions.ClassifyIdempotencyWriteError(err)
}

type taskDecisionRevisions struct {
	appender *revisions.Appender
	history  historyquery.Reader
}

func (a taskDecisionRevisions) AppendChangeSetTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendChangeSetParams,
) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, params)
}

func (a taskDecisionRevisions) AppendMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendMutationParams,
) error {
	return a.appender.AppendMutationTx(ctx, tx, params)
}

func (a taskDecisionRevisions) AppendRecordRevisionTx(
	ctx context.Context,
	tx pgx.Tx,
	params revisions.AppendRecordRevisionParams,
) error {
	return a.appender.AppendRecordRevisionTx(ctx, tx, params)
}

func (a taskDecisionRevisions) LoadRevisionWindowTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	baseVersion int64,
	currentVersion int64,
) ([]historyquery.RevisionWindowRow, error) {
	return a.history.LoadRevisionWindowTx(ctx, tx, recordID, baseVersion, currentVersion)
}

func newTaskDecisionMutationCapabilities(
	pool postgres.DB,
	appender *revisions.Appender,
) tasksdecisions.MutationCapabilities {
	projectionContribution := tasksdecisions.NewProjectionContribution()
	authStore := authn.NewStore(pool)
	return tasksdecisions.MutationCapabilities{
		IncidentState:    incidents.NewAccess(pool),
		MemberReferences: tasksdecisions.NewMemberReferenceCapability(),
		Idempotency:      taskDecisionIdempotency{store: authStore},
		RecordEnvelopes:  records.NewStore(),
		Links:            links.NewStore(),
		Projections: projections.NewTaskDecisionRows(
			pool,
			projectionContribution.Source(),
			projectionContribution.QuerySurfaces()...,
		),
		Revisions:           taskDecisionRevisions{appender: appender, history: historyquery.NewReader()},
		ConflictIdempotency: authStore,
	}
}
