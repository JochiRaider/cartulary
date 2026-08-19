package indicators

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type indicatorLifecycleService struct {
	owner *Store
}

func (s *Store) AppendIndicatorLifecycleInterval(ctx context.Context, actor authn.UserRecord, params IndicatorLifecycleAppendParams) (IndicatorLifecycleMutationResult, error) {
	return s.lifecycleService.appendInterval(ctx, actor, params)
}

func (service indicatorLifecycleService) appendInterval(ctx context.Context, actor authn.UserRecord, params IndicatorLifecycleAppendParams) (IndicatorLifecycleMutationResult, error) {
	s := service.owner
	if err := validateChildMutationIdentity(params.ClientTxnID, params.RequestID, params.RequestHash, params.BaseRowVersion); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := normalizeLifecycleAppendParams(&params); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	replayKey := s.childReplayKey(lifecycleAppendRouteKey, actor.ID, params.IndicatorRecordID, params.ClientTxnID)
	if replay, found, err := loadLifecycleReplay(ctx, s.authStore, replayKey, params.RequestHash); err != nil || found {
		return replay, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, fmt.Errorf("begin Indicator lifecycle transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	lockIDs := sortedRecordIDs(append([]uuid.UUID{params.IndicatorRecordID}, params.SupportRefs...)...)
	envelopes, err := s.lockAffectedRecordsTx(ctx, tx, params.IncidentID, lockIDs)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, ErrInvalidCreateRequest
	}
	indicatorEnvelope := envelopes[params.IndicatorRecordID]
	if err := validateIndicatorEnvelope(indicatorEnvelope); err != nil {
		return IndicatorLifecycleMutationResult{}, ErrIndicatorNotFound
	}
	if indicatorEnvelope.RowVersion != params.BaseRowVersion {
		return IndicatorLifecycleMutationResult{}, ErrRowVersionConflict
	}
	beforeRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, params.IndicatorRecordID)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	beforeSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisionsStore, []uuid.UUID{params.IndicatorRecordID})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	createdAt := s.now().UTC().Truncate(time.Microsecond)
	record, err := s.lifecycles.insertTx(ctx, tx, actor.ID, params, createdAt)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: params.IncidentID, ActorUserID: actor.ID, Source: lifecycleAppendSource,
		ClientTxnID: &params.ClientTxnID, RequestID: &params.RequestID, CreatedAt: createdAt,
	})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := s.revisionsStore.AppendMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "indicator_state_interval",
		TargetID: record.IntervalID.String(), OperationKind: "create",
		AfterVersionID: stringPointer(fmt.Sprintf("indicator_state_interval:%s:%d", record.IntervalID, record.RowVersion)),
		AfterValue:     buildIndicatorLifecycleValue(record),
	}); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	versions, err := s.advanceAffectedRecordsTx(ctx, tx, actor.ID, createdAt, []uuid.UUID{params.IndicatorRecordID})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	afterRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, params.IndicatorRecordID)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	afterSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisionsStore, []uuid.UUID{params.IndicatorRecordID})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := appendAffectedRecordRevisionsTx(ctx, tx, s.revisionsStore, changeSetID, versions, beforeSnapshots, afterSnapshots,
		map[uuid.UUID]map[string]any{params.IndicatorRecordID: beforeRow},
		map[uuid.UUID]map[string]any{params.IndicatorRecordID: afterRow},
	); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	result := IndicatorLifecycleMutationResult{Interval: record, ChangeSetID: changeSetID, AffectedRecords: versions}
	payload, err := mutationPayload(result)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, replayKey, nil, params.RequestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return IndicatorLifecycleMutationResult{}, authn.ErrClientTxnConflict
		}
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return IndicatorLifecycleMutationResult{}, fmt.Errorf("commit Indicator lifecycle transaction: %w", err)
	}
	return result, nil
}

func normalizeLifecycleAppendParams(params *IndicatorLifecycleAppendParams) error {
	if params == nil || params.IncidentID == uuid.Nil || params.IndicatorRecordID == uuid.Nil || params.ValidFrom.IsZero() {
		return ErrInvalidCreateRequest
	}
	switch params.LifecycleState {
	case "active", "benign", "false_positive", "retired":
	default:
		return ErrInvalidCreateRequest
	}
	params.ValidFrom = params.ValidFrom.UTC().Truncate(time.Microsecond)
	if params.ValidTo != nil {
		validTo := params.ValidTo.UTC().Truncate(time.Microsecond)
		if validTo.Before(params.ValidFrom) {
			return ErrInvalidCreateRequest
		}
		params.ValidTo = &validTo
	}
	if params.Confidence != nil && (*params.Confidence < 0 || *params.Confidence > 100) {
		return ErrInvalidCreateRequest
	}
	if !validOptionalLifecycleText(params.Rationale) || !validOptionalLifecycleText(params.Assessor) || len(params.SupportRefs) > 64 {
		return ErrInvalidCreateRequest
	}
	seen := make(map[uuid.UUID]struct{}, len(params.SupportRefs))
	for _, recordID := range params.SupportRefs {
		if recordID == uuid.Nil {
			return ErrInvalidCreateRequest
		}
		if _, duplicate := seen[recordID]; duplicate {
			return ErrInvalidCreateRequest
		}
		seen[recordID] = struct{}{}
	}
	sort.Slice(params.SupportRefs, func(left int, right int) bool {
		return params.SupportRefs[left].String() < params.SupportRefs[right].String()
	})
	return nil
}

func validOptionalLifecycleText(value *string) bool {
	return value == nil || (*value != "" && !strings.ContainsRune(*value, 0))
}

func (s *Store) ListIndicatorLifecycleIntervals(ctx context.Context, indicatorID uuid.UUID, afterValidFrom *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorLifecycleIntervalRecord, error) {
	return s.lifecycles.list(ctx, s.pool, indicatorID, afterValidFrom, afterID, limit)
}
