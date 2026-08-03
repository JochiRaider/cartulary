package indicators

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type indicatorLifecycleService struct {
	owner *Store
}

func (s *Store) AppendIndicatorLifecycleInterval(ctx context.Context, actor authn.UserRecord, params IndicatorLifecycleAppendParams) (IndicatorLifecycleIntervalRecord, uuid.UUID, error) {
	return s.lifecycleService.appendInterval(ctx, actor, params)
}

func (service indicatorLifecycleService) appendInterval(ctx context.Context, actor authn.UserRecord, params IndicatorLifecycleAppendParams) (IndicatorLifecycleIntervalRecord, uuid.UUID, error) {
	s := service.owner
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, fmt.Errorf("begin indicator lifecycle transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := s.sources.validateIncidentTx(ctx, tx, params.IncidentID, params.IndicatorRecordID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	record, err := s.lifecycles.insertTx(ctx, tx, actor.ID, params, createdAt)
	if err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}

	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: params.IncidentID, ActorUserID: actor.ID, Source: lifecycleAppendSource,
		ClientTxnID: params.ClientTxnID, RequestID: params.RequestID, CreatedAt: createdAt,
	})
	if err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "indicator_state_interval",
		TargetID: record.IntervalID.String(), OperationKind: "create",
		AfterVersionID: stringPointer(fmt.Sprintf("indicator_state_interval:%s:%d", record.IntervalID, record.RowVersion)),
		AfterValue:     buildIndicatorLifecycleValue(record),
	}); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if err := s.refreshProjectionRowTx(ctx, tx, params.IndicatorRecordID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return IndicatorLifecycleIntervalRecord{}, uuid.UUID{}, fmt.Errorf("commit indicator lifecycle transaction: %w", err)
	}
	return record, changeSetID, nil
}
