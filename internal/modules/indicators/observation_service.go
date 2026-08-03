package indicators

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type indicatorObservationService struct {
	owner *Store
}

func (s *Store) CreateIndicatorObservation(ctx context.Context, actor authn.UserRecord, params IndicatorObservationCreateParams) (IndicatorObservationRecord, uuid.UUID, error) {
	return s.observationService.createManualObservation(ctx, actor, params)
}

func (service indicatorObservationService) createManualObservation(ctx context.Context, actor authn.UserRecord, params IndicatorObservationCreateParams) (IndicatorObservationRecord, uuid.UUID, error) {
	s := service.owner
	params.originKind = indicatororigin.ManualEntry
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("begin indicator observation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	record, err := s.observations.insertTx(ctx, tx, actor.ID, params, createdAt)
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}

	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: params.IncidentID, ActorUserID: actor.ID, Source: observationCreateSource,
		ClientTxnID: params.ClientTxnID, RequestID: params.RequestID, CreatedAt: createdAt,
	})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "indicator_observation",
		TargetID: record.ObservationID.String(), OperationKind: "create",
		AfterVersionID: stringPointer(fmt.Sprintf("indicator_observation:%s:%d", record.ObservationID, record.RowVersion)),
		AfterValue:     buildIndicatorObservationValue(record),
	}); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if record.ResolvedIndicatorRecordID != nil {
		if err := s.refreshProjectionRowTx(ctx, tx, *record.ResolvedIndicatorRecordID); err != nil {
			return IndicatorObservationRecord{}, uuid.UUID{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("commit indicator observation transaction: %w", err)
	}
	return record, changeSetID, nil
}

func (s *Store) ResolveIndicatorObservation(ctx context.Context, actor authn.UserRecord, params IndicatorObservationResolveParams) (IndicatorObservationRecord, uuid.UUID, error) {
	return s.observationService.resolveObservation(ctx, actor, params)
}

func (service indicatorObservationService) resolveObservation(ctx context.Context, actor authn.UserRecord, params IndicatorObservationResolveParams) (IndicatorObservationRecord, uuid.UUID, error) {
	s := service.owner
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("begin indicator observation resolve transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	current, err := s.observations.loadTx(ctx, tx, params.ObservationID, true)
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, current.IncidentID); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.sources.validateIncidentTx(ctx, tx, current.IncidentID, params.ResolvedIndicatorRecordID); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}

	resolvedAt := time.Now().UTC().Truncate(time.Microsecond)
	next := current
	next.ResolutionStatus = "resolved"
	next.ResolvedIndicatorRecordID = &params.ResolvedIndicatorRecordID
	next.ResolvedByUserID = &actor.ID
	next.ResolvedAt = &resolvedAt
	method := observationResolveSource
	next.ResolutionMethod = &method
	next.RowVersion = current.RowVersion + 1
	if err := s.observations.resolveTx(ctx, tx, next); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}

	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: current.IncidentID, ActorUserID: actor.ID, Source: method,
		ClientTxnID: params.ClientTxnID, RequestID: params.RequestID, CreatedAt: resolvedAt,
	})
	if err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "indicator_observation",
		TargetID: next.ObservationID.String(), OperationKind: "resolve",
		BeforeVersionID: stringPointer(fmt.Sprintf("indicator_observation:%s:%d", current.ObservationID, current.RowVersion)),
		AfterVersionID:  stringPointer(fmt.Sprintf("indicator_observation:%s:%d", next.ObservationID, next.RowVersion)),
		BeforeValue:     buildIndicatorObservationValue(current), AfterValue: buildIndicatorObservationValue(next),
	}); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, err
	}
	projectionIDs := []uuid.UUID{params.ResolvedIndicatorRecordID}
	if current.ResolvedIndicatorRecordID != nil && *current.ResolvedIndicatorRecordID != params.ResolvedIndicatorRecordID {
		projectionIDs = append(projectionIDs, *current.ResolvedIndicatorRecordID)
	}
	for _, recordID := range projectionIDs {
		if err := s.refreshProjectionRowTx(ctx, tx, recordID); err != nil {
			return IndicatorObservationRecord{}, uuid.UUID{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return IndicatorObservationRecord{}, uuid.UUID{}, fmt.Errorf("commit indicator observation resolve transaction: %w", err)
	}
	return next, changeSetID, nil
}
