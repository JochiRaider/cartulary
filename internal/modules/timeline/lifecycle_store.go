package timeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *store) MarkReviewed(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request ActionRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyAction(ctx, actor, reviewRouteKey, recordID, request.BaseRowVersion, request.ClientTxnID, requestHash, requestID, now, request.Reason, nil, func(current sourceRecord) (sourceRecord, *string, error) {
		if !CaptureStateAllowsMarkReviewed(current.CaptureState) {
			return sourceRecord{}, nil, newIllegalTransitionError("mark_reviewed_not_allowed", current.CaptureState, captureStateReviewed)
		}
		next := current
		next.CaptureState = captureStateReviewed
		next.RowVersion = current.RowVersion + 1
		next.EditedAt = now.UTC()
		next.UpdatedByUserID = actor.ID
		next.ReviewedAt = &next.EditedAt
		reviewerID := actor.ID
		next.ReviewedByUserID = &reviewerID
		return next, request.Reason, nil
	})
}

func (s *store) Supersede(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request SupersedeRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyAction(ctx, actor, supersedeRouteKey, recordID, request.BaseRowVersion, request.ClientTxnID, requestHash, requestID, now, &request.Reason, request.ReplacementRecordID, func(current sourceRecord) (sourceRecord, *string, error) {
		if !CaptureStateAllowsSupersede(current.CaptureState) {
			return sourceRecord{}, nil, newIllegalTransitionError("supersede_not_allowed", current.CaptureState, captureStateSuperseded)
		}

		next := current
		next.CaptureState = captureStateSuperseded
		next.RowVersion = current.RowVersion + 1
		next.EditedAt = now.UTC()
		next.UpdatedByUserID = actor.ID
		next.SupersededAt = &next.EditedAt
		supersededBy := actor.ID
		next.SupersededByUserID = &supersededBy
		return next, &request.Reason, nil
	})
}

func (s *store) applyAction(
	ctx context.Context,
	actor authn.UserRecord,
	routeKey string,
	recordID uuid.UUID,
	baseRowVersion int64,
	clientTxnID string,
	requestHash []byte,
	requestID string,
	now time.Time,
	reason *string,
	replacementRecordID *uuid.UUID,
	prepare func(sourceRecord) (sourceRecord, *string, error),
) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline action payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline action idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline action transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, current.IncidentID); err != nil {
		return MutationResult{}, err
	}
	if current.RowVersion != baseRowVersion {
		return MutationResult{}, ErrRowVersionConflict
	}

	next, effectiveReason, err := prepare(current)
	if err != nil {
		return MutationResult{}, err
	}

	var validatedReplacementID *uuid.UUID
	if routeKey == supersedeRouteKey {
		if err := s.validateSupersedeReplacementTx(ctx, tx, current, replacementRecordID); err != nil {
			return MutationResult{}, err
		}
		if replacementRecordID != nil {
			replacementID := *replacementRecordID
			validatedReplacementID = &replacementID
		}
	} else if replacementRecordID != nil {
		replacementID := *replacementRecordID
		validatedReplacementID = &replacementID
	}
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}

	beforeProjected := projectRecord(current, nil)
	if err := hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return MutationResult{}, err
	}
	if err := tx.QueryRow(ctx, `
UPDATE timeline_events
   SET capture_state = $2,
       row_version = $3,
       edited_at = $4,
       updated_by_user_id = $5,
       reviewed_at = $6,
       reviewed_by_user_id = $7,
       superseded_at = $8,
       superseded_by_user_id = $9
 WHERE record_id = $1
RETURNING recorded_at
`, recordID, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID, next.SupersededAt, next.SupersededByUserID).Scan(&next.RecordedAt); err != nil {
		return MutationResult{}, fmt.Errorf("update timeline action state: %w", err)
	}

	var insertedLink *supersedesLink
	if routeKey == supersedeRouteKey && validatedReplacementID != nil {
		link, err := s.linkStore.InsertSupersedesTx(ctx, tx, current.IncidentID, *validatedReplacementID, current.RecordID, actor.ID, now.UTC())
		if err != nil {
			if isRecordLinkConflict(err) {
				return MutationResult{}, newIllegalTransitionError("supersede_not_allowed", current.CaptureState, captureStateSuperseded, supersedeGuardTargetMustNotHaveActiveReplacement)
			}
			return MutationResult{}, err
		}
		insertedLink = &link
	}

	afterProjected := projectRecord(next, validatedReplacementID)
	if err := hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, timelineChangeSetParams{
		IncidentID:  current.IncidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		Reason:      effectiveReason,
		ClientTxnID: &clientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	beforeRow := buildRow(beforeProjected)
	afterRow := buildRow(afterProjected)
	beforeVersion := versionID(current.RecordID, current.RowVersion)
	afterVersion := versionID(next.RecordID, next.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, timelineMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "timeline_record",
		TargetID:        current.RecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersion,
		AfterVersionID:  &afterVersion,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if insertedLink != nil {
		linkAfter, err := s.linkStore.LoadRecordLinkValueTx(ctx, tx, insertedLink.RecordLinkID)
		if err != nil {
			return MutationResult{}, err
		}
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, timelineMutationParams{
			ChangeSetID:    changeSetID,
			SequenceNo:     2,
			TargetKind:     "record_link",
			TargetID:       insertedLink.RecordLinkID.String(),
			OperationKind:  "create",
			AfterVersionID: nil,
			AfterValue:     linkAfter,
		}); err != nil {
			return MutationResult{}, err
		}
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, timelineRecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    current.RecordID,
		RowVersion:  next.RowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.projectionStore.UpsertTimelineRowTx(ctx, tx, projectionInput(afterProjected)); err != nil {
		return MutationResult{}, err
	}

	payload := BuildActionPayload(afterProjected, changeSetID, effectiveReason)
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := s.beforeCommit(routeKey, recordID); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline action transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       current.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      clientTxnID,
		RowVersion:       afterProjected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		Row:              afterProjected,
	}, nil
}

func (s *store) validateSupersedeReplacementTx(ctx context.Context, tx pgx.Tx, current sourceRecord, replacementRecordID *uuid.UUID) error {
	if replacementRecordID == nil {
		return nil
	}

	guards := make([]string, 0, 3)
	if *replacementRecordID == current.RecordID {
		guards = append(guards, supersedeGuardReplacementDifferent)
	}

	targetHasActiveReplacement, err := s.linkStore.HasActiveIncomingSupersedesLinkForUpdateTx(ctx, tx, current.IncidentID, current.RecordID)
	if err != nil {
		return err
	}
	if targetHasActiveReplacement {
		guards = append(guards, supersedeGuardTargetMustNotHaveActiveReplacement)
	}

	if *replacementRecordID != current.RecordID {
		replacement, err := loadSourceRecordTx(ctx, tx, *replacementRecordID)
		if errors.Is(err, ErrRecordNotFound) {
			guards = append(guards, supersedeGuardReplacementVisibleActiveSameIncident)
		} else if err != nil {
			return err
		} else {
			if replacement.IncidentID != current.IncidentID {
				guards = append(guards, supersedeGuardReplacementVisibleActiveSameIncident)
			}
			if replacement.CaptureState == captureStateSuperseded {
				guards = append(guards, supersedeGuardReplacementNotSuperseded)
			}
		}
	}

	if len(guards) > 0 {
		return newIllegalTransitionError("supersede_not_allowed", current.CaptureState, captureStateSuperseded, guards...)
	}
	return nil
}
