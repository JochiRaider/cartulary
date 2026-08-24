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

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *Application) AppendIndicatorLifecycleInterval(ctx context.Context, actorUserID uuid.UUID, params IndicatorLifecycleAppendParams) (IndicatorLifecycleMutationResult, error) {
	if actorUserID == uuid.Nil {
		return IndicatorLifecycleMutationResult{}, ErrInvalidCreateRequest
	}
	requestHash := lifecycleAppendRequestHash(params)
	if err := validateChildMutationIdentity(params.ClientTxnID, params.RequestID, params.BaseRowVersion); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := normalizeLifecycleAppendParams(&params); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	replayKey := s.childReplayKey(lifecycleAppendRouteKey, actorUserID, params.IndicatorRecordID, params.ClientTxnID)
	if replay, found, err := loadLifecycleReplay(ctx, s.idempotency, replayKey, requestHash); err != nil || found {
		return replay, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, fmt.Errorf("begin Indicator lifecycle transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.incidentState.RequireOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	lockIDs := sortedRecordIDs(append([]uuid.UUID{params.IndicatorRecordID}, params.SupportRefs...)...)
	envelopes, err := s.lockAffectedRecordsTx(ctx, tx, lockIDs)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := validateAddressedIndicatorEnvelope(envelopes, params.IncidentID, params.IndicatorRecordID); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := validateLifecycleSupportEnvelopes(envelopes, params.IncidentID, params.SupportRefs); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	indicatorEnvelope := envelopes[params.IndicatorRecordID]
	if indicatorEnvelope.RowVersion != params.BaseRowVersion {
		return IndicatorLifecycleMutationResult{}, ErrRowVersionConflict
	}
	beforeRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, params.IndicatorRecordID)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	beforeSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisions, []uuid.UUID{params.IndicatorRecordID})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	createdAt := s.now().UTC().Truncate(time.Microsecond)
	record, err := insertIndicatorLifecycleIntervalTx(ctx, tx, actorUserID, params, createdAt)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	changeSetID, err := s.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: params.IncidentID, ActorUserID: actorUserID, Source: lifecycleAppendSource,
		ClientTxnID: &params.ClientTxnID, RequestID: &params.RequestID, CreatedAt: createdAt,
	})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := s.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "indicator_state_interval",
		TargetID: record.IntervalID.String(), OperationKind: "create",
		AfterVersionID: stringPointer(fmt.Sprintf("indicator_state_interval:%s:%d", record.IntervalID, record.RowVersion)),
		AfterValue:     buildIndicatorLifecycleValue(record),
	}); err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	versions, err := s.advanceAffectedRecordsTx(ctx, tx, actorUserID, createdAt, []uuid.UUID{params.IndicatorRecordID})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	afterRow, err := s.refreshAndLoadProjectionRowTx(ctx, tx, params.IndicatorRecordID)
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	afterSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisions, []uuid.UUID{params.IndicatorRecordID})
	if err != nil {
		return IndicatorLifecycleMutationResult{}, err
	}
	if err := appendAffectedRecordRevisionsTx(ctx, tx, s.revisions, changeSetID, versions, beforeSnapshots, afterSnapshots,
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
	if err := s.idempotency.InsertRouteIdempotencyPayload(ctx, tx, replayKey, requestHash, http.StatusCreated, payload); err != nil {
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
	if !vocabulary.IsLifecycleState(params.LifecycleState) {
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
	params.SupportRefs = append([]uuid.UUID(nil), params.SupportRefs...)
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

func (s *Application) ListIndicatorLifecycleIntervals(ctx context.Context, indicatorID uuid.UUID, afterValidFrom *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorLifecycleIntervalRecord, error) {
	return listIndicatorLifecycleIntervals(ctx, s.pool, indicatorID, afterValidFrom, afterID, limit)
}
