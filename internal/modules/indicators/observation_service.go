package indicators

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *Application) CreateIndicatorObservation(ctx context.Context, actorUserID uuid.UUID, params IndicatorObservationCreateParams) (IndicatorObservationMutationResult, error) {
	if actorUserID == uuid.Nil {
		return IndicatorObservationMutationResult{}, ErrInvalidCreateRequest
	}
	requestHash := observationCreateRequestHash(params)
	if err := validateChildMutationIdentity(params.ClientTxnID, params.RequestID, params.BaseRowVersion); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if params.IncidentID == uuid.Nil || params.SourceRecordID == uuid.Nil || params.SourceFieldKey == "" {
		return IndicatorObservationMutationResult{}, ErrInvalidCreateRequest
	}
	replayKey := s.childReplayKey(observationCreateRouteKey, actorUserID, params.SourceRecordID, params.ClientTxnID)
	if replay, found, err := loadObservationReplay(ctx, s.idempotency, replayKey, requestHash); err != nil || found {
		return replay, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorObservationMutationResult{}, fmt.Errorf("begin Indicator observation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := s.incidentState.RequireOpenTx(ctx, tx, params.IncidentID); err != nil {
		return IndicatorObservationMutationResult{}, err
	}

	affectedIDs := []uuid.UUID{params.SourceRecordID}
	if params.ResolvedIndicatorRecordID != nil {
		affectedIDs = append(affectedIDs, *params.ResolvedIndicatorRecordID)
	}
	affectedIDs = sortedRecordIDs(affectedIDs...)
	envelopes, err := s.lockAffectedRecordsTx(ctx, tx, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := validateObservationSourceEnvelope(envelopes, params.IncidentID, params.SourceRecordID, params.BaseRowVersion); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if params.ResolvedIndicatorRecordID != nil {
		if err := validateResolvedIndicatorEnvelope(envelopes, params.IncidentID, *params.ResolvedIndicatorRecordID); err != nil {
			return IndicatorObservationMutationResult{}, err
		}
	}
	source, err := s.sourceText.LoadTextTx(ctx, tx, params.SourceRecordID, envelopes[params.SourceRecordID].RecordType, params.SourceFieldKey)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	observedText, err := sourceSpan(source.Text, params.SpanStartByte, params.SpanEndByte)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	params.originKind = indicatororigin.ManualEntry
	params.originLocator = sourceOriginLocator(params.SourceRecordID, params.SourceFieldKey, params.SpanStartByte, params.SpanEndByte)
	params.observedText = observedText

	beforeRows, err := s.beforeObservationAffectedRowsTx(ctx, tx, params.SourceRecordID, source.Row, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	beforeSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisions, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	createdAt := s.now().UTC().Truncate(time.Microsecond)
	record, err := insertIndicatorObservationTx(ctx, tx, actorUserID, params, createdAt)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	changeSetID, err := s.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: params.IncidentID, ActorUserID: actorUserID, Source: observationCreateSource,
		ClientTxnID: &params.ClientTxnID, RequestID: &params.RequestID, CreatedAt: createdAt,
	})
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := s.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "indicator_observation",
		TargetID: record.ObservationID.String(), OperationKind: "create",
		AfterVersionID: stringPointer(fmt.Sprintf("indicator_observation:%s:%d", record.ObservationID, record.RowVersion)),
		AfterValue:     buildIndicatorObservationValue(record),
	}); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	versions, err := s.advanceAffectedRecordsTx(ctx, tx, actorUserID, createdAt, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	afterRows, err := s.afterObservationAffectedRowsTx(ctx, tx, params.SourceRecordID, params.SourceFieldKey, envelopes, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	afterSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisions, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := appendAffectedRecordRevisionsTx(ctx, tx, s.revisions, changeSetID, versions, beforeSnapshots, afterSnapshots, beforeRows, afterRows); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	result := IndicatorObservationMutationResult{Observation: record, ChangeSetID: changeSetID, AffectedRecords: versions}
	payload, err := mutationPayload(result)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := s.idempotency.InsertRouteIdempotencyPayload(ctx, tx, replayKey, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return IndicatorObservationMutationResult{}, authn.ErrClientTxnConflict
		}
		return IndicatorObservationMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return IndicatorObservationMutationResult{}, fmt.Errorf("commit Indicator observation transaction: %w", err)
	}
	return result, nil
}

func (s *Application) ResolveIndicatorObservation(ctx context.Context, actorUserID uuid.UUID, params IndicatorObservationResolveParams) (IndicatorObservationMutationResult, error) {
	return s.transitionIndicatorObservation(ctx, actorUserID, observationTransitionResolve, params.ObservationID, params.ResolvedIndicatorRecordID, params.BaseRowVersion, params.ClientTxnID, params.RequestID, observationResolveRequestHash(params))
}

func (s *Application) DismissIndicatorObservation(ctx context.Context, actorUserID uuid.UUID, params IndicatorObservationActionParams) (IndicatorObservationMutationResult, error) {
	return s.transitionIndicatorObservation(ctx, actorUserID, observationTransitionDismiss, params.ObservationID, uuid.Nil, params.BaseRowVersion, params.ClientTxnID, params.RequestID, observationActionRequestHash(params))
}

func (s *Application) RestoreIndicatorObservation(ctx context.Context, actorUserID uuid.UUID, params IndicatorObservationActionParams) (IndicatorObservationMutationResult, error) {
	return s.transitionIndicatorObservation(ctx, actorUserID, observationTransitionRestore, params.ObservationID, uuid.Nil, params.BaseRowVersion, params.ClientTxnID, params.RequestID, observationActionRequestHash(params))
}

type observationTransition string

const (
	observationTransitionResolve observationTransition = "resolve"
	observationTransitionDismiss observationTransition = "dismiss"
	observationTransitionRestore observationTransition = "restore"
)

func (s *Application) transitionIndicatorObservation(ctx context.Context, actorUserID uuid.UUID, transition observationTransition, observationID uuid.UUID, targetID uuid.UUID, baseRowVersion int64, clientTxnID string, requestID string, requestHash []byte) (IndicatorObservationMutationResult, error) {
	if actorUserID == uuid.Nil {
		return IndicatorObservationMutationResult{}, ErrInvalidCreateRequest
	}
	if observationID == uuid.Nil || (transition == observationTransitionResolve && targetID == uuid.Nil) {
		return IndicatorObservationMutationResult{}, ErrInvalidCreateRequest
	}
	if err := validateChildMutationIdentity(clientTxnID, requestID, baseRowVersion); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	routeKey := map[observationTransition]string{
		observationTransitionResolve: observationResolveRouteKey,
		observationTransitionDismiss: observationDismissRouteKey,
		observationTransitionRestore: observationRestoreRouteKey,
	}[transition]
	replayKey := s.childReplayKey(routeKey, actorUserID, observationID, clientTxnID)
	if replay, found, err := loadObservationReplay(ctx, s.idempotency, replayKey, requestHash); err != nil || found {
		return replay, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return IndicatorObservationMutationResult{}, fmt.Errorf("begin Indicator observation transition: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	current, err := loadIndicatorObservation(ctx, tx, observationID, true)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if current.RowVersion != baseRowVersion {
		return IndicatorObservationMutationResult{}, ErrRowVersionConflict
	}
	if err := s.incidentState.RequireOpenTx(ctx, tx, current.IncidentID); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	transitionAt := s.now().UTC().Truncate(time.Microsecond)
	next, err := nextObservationTransition(current, transition, targetID, actorUserID, transitionAt)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	affectedIDs := []uuid.UUID{current.SourceRecordID}
	if current.ResolvedIndicatorRecordID != nil {
		affectedIDs = append(affectedIDs, *current.ResolvedIndicatorRecordID)
	}
	if next.ResolvedIndicatorRecordID != nil {
		affectedIDs = append(affectedIDs, *next.ResolvedIndicatorRecordID)
	}
	affectedIDs = sortedRecordIDs(affectedIDs...)
	envelopes, err := s.lockAffectedRecordsTx(ctx, tx, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := validatePriorObservationDependencies(envelopes, current); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if transition == observationTransitionResolve {
		if err := validateResolvedIndicatorEnvelope(envelopes, current.IncidentID, targetID); err != nil {
			return IndicatorObservationMutationResult{}, err
		}
	}
	sourceRow, err := s.sourceText.LoadRowTx(ctx, tx, current.SourceRecordID, envelopes[current.SourceRecordID].RecordType, current.SourceFieldKey)
	if err != nil {
		if errors.Is(err, ErrSourceTextUnavailable) {
			return IndicatorObservationMutationResult{}, ErrIndicatorObservationNotFound
		}
		return IndicatorObservationMutationResult{}, err
	}
	beforeRows, err := s.beforeObservationAffectedRowsTx(ctx, tx, current.SourceRecordID, sourceRow, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	beforeSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisions, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := updateIndicatorObservationTransitionTx(ctx, tx, next, current.RowVersion); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	changeSetID, err := s.revisions.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: current.IncidentID, ActorUserID: actorUserID, Source: "indicators.observations." + string(transition),
		ClientTxnID: &clientTxnID, RequestID: &requestID, CreatedAt: transitionAt,
	})
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := s.revisions.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams{
		ChangeSetID: changeSetID, SequenceNo: 1, TargetKind: "indicator_observation",
		TargetID: next.ObservationID.String(), OperationKind: string(transition),
		BeforeVersionID: stringPointer(fmt.Sprintf("indicator_observation:%s:%d", current.ObservationID, current.RowVersion)),
		AfterVersionID:  stringPointer(fmt.Sprintf("indicator_observation:%s:%d", next.ObservationID, next.RowVersion)),
		BeforeValue:     buildIndicatorObservationValue(current), AfterValue: buildIndicatorObservationValue(next),
	}); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	versions, err := s.advanceAffectedRecordsTx(ctx, tx, actorUserID, transitionAt, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	afterRows, err := s.afterObservationAffectedRowsTx(ctx, tx, current.SourceRecordID, current.SourceFieldKey, envelopes, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	afterSnapshots, err := captureAffectedRecordSnapshotsTx(ctx, tx, s.revisions, affectedIDs)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := appendAffectedRecordRevisionsTx(ctx, tx, s.revisions, changeSetID, versions, beforeSnapshots, afterSnapshots, beforeRows, afterRows); err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	result := IndicatorObservationMutationResult{Observation: next, ChangeSetID: changeSetID, AffectedRecords: versions}
	payload, err := mutationPayload(result)
	if err != nil {
		return IndicatorObservationMutationResult{}, err
	}
	if err := s.idempotency.InsertRouteIdempotencyPayload(ctx, tx, replayKey, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return IndicatorObservationMutationResult{}, authn.ErrClientTxnConflict
		}
		return IndicatorObservationMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return IndicatorObservationMutationResult{}, fmt.Errorf("commit Indicator observation transition: %w", err)
	}
	return result, nil
}

func nextObservationTransition(current IndicatorObservationRecord, transition observationTransition, targetID uuid.UUID, actorID uuid.UUID, now time.Time) (IndicatorObservationRecord, error) {
	next := current
	next.RowVersion++
	method := "indicators.observations." + string(transition)
	switch transition {
	case observationTransitionResolve:
		if current.ResolutionStatus == "dismissed" || (current.ResolutionStatus == "resolved" && current.ResolvedIndicatorRecordID != nil && *current.ResolvedIndicatorRecordID == targetID) {
			return IndicatorObservationRecord{}, ErrIllegalTransition
		}
		next.ResolutionStatus = "resolved"
		next.ResolvedIndicatorRecordID = &targetID
		next.ResolvedByUserID = &actorID
		next.ResolvedAt = &now
		next.ResolutionMethod = &method
	case observationTransitionDismiss:
		if current.ResolutionStatus != "unresolved" && current.ResolutionStatus != "resolved" {
			return IndicatorObservationRecord{}, ErrIllegalTransition
		}
		next.ResolutionStatus = "dismissed"
		next.ResolvedIndicatorRecordID = nil
		next.ResolvedByUserID = &actorID
		next.ResolvedAt = &now
		next.ResolutionMethod = &method
	case observationTransitionRestore:
		if current.ResolutionStatus != "dismissed" {
			return IndicatorObservationRecord{}, ErrIllegalTransition
		}
		next.ResolutionStatus = "unresolved"
		next.ResolvedIndicatorRecordID = nil
		next.ResolvedByUserID = nil
		next.ResolvedAt = nil
		next.ResolutionMethod = nil
	default:
		return IndicatorObservationRecord{}, ErrIllegalTransition
	}
	return next, nil
}

func (s *Application) beforeObservationAffectedRowsTx(ctx context.Context, tx pgx.Tx, sourceRecordID uuid.UUID, sourceRow map[string]any, affectedIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	rows := map[uuid.UUID]map[string]any{sourceRecordID: sourceRow}
	for _, recordID := range affectedIDs {
		if recordID == sourceRecordID {
			continue
		}
		row, err := s.refreshAndLoadProjectionRowTx(ctx, tx, recordID)
		if err != nil {
			return nil, err
		}
		rows[recordID] = row
	}
	return rows, nil
}

func (s *Application) afterObservationAffectedRowsTx(ctx context.Context, tx pgx.Tx, sourceRecordID uuid.UUID, sourceFieldKey string, envelopes map[uuid.UUID]records.Envelope, affectedIDs []uuid.UUID) (map[uuid.UUID]map[string]any, error) {
	rows := make(map[uuid.UUID]map[string]any, len(affectedIDs))
	for _, recordID := range affectedIDs {
		var (
			row map[string]any
			err error
		)
		if recordID == sourceRecordID {
			row, err = s.sourceText.RefreshAndLoadRowTx(ctx, tx, recordID, envelopes[recordID].RecordType, sourceFieldKey)
		} else {
			row, err = s.refreshAndLoadProjectionRowTx(ctx, tx, recordID)
		}
		if err != nil {
			return nil, err
		}
		rows[recordID] = row
	}
	return rows, nil
}

func (s *Application) ListSourceRecordIndicatorObservations(ctx context.Context, sourceRecordID uuid.UUID, afterCreatedAt *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorObservationRecord, error) {
	return listSourceRecordIndicatorObservations(ctx, s.pool, sourceRecordID, afterCreatedAt, afterID, limit)
}

func (s *Application) ListIndicatorObservations(ctx context.Context, indicatorID uuid.UUID, afterCreatedAt *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorObservationRecord, error) {
	return listResolvedIndicatorObservations(ctx, s.pool, indicatorID, afterCreatedAt, afterID, limit)
}

func (s *Application) GetIndicatorObservation(ctx context.Context, observationID uuid.UUID) (IndicatorObservationRecord, error) {
	if observationID == uuid.Nil {
		return IndicatorObservationRecord{}, ErrIndicatorObservationNotFound
	}
	return loadIndicatorObservation(ctx, s.pool, observationID, false)
}
