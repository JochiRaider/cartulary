package timeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *store) MarkReviewed(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request ActionRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyAction(ctx, actor, reviewRouteKey, recordID, request.BaseRowVersion, request.ClientTxnID, requestHash, requestID, now, request.Reason, nil, func(current sourcerepository.Snapshot) (sourcerepository.Snapshot, *string, error) {
		if !captureStateAllowsMarkReviewed(current.CaptureState) {
			return sourcerepository.Snapshot{}, nil, newIllegalTransitionError(reasonTransitionNotAllowed, current.CaptureState, captureStateReviewed)
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
	return s.applyAction(ctx, actor, supersedeRouteKey, recordID, request.BaseRowVersion, request.ClientTxnID, requestHash, requestID, now, &request.Reason, request.ReplacementRecordID, func(current sourcerepository.Snapshot) (sourcerepository.Snapshot, *string, error) {
		if !captureStateAllowsSupersede(current.CaptureState) {
			return sourcerepository.Snapshot{}, nil, newIllegalTransitionError(reasonTransitionNotAllowed, current.CaptureState, captureStateSuperseded)
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
	prepare func(sourcerepository.Snapshot) (sourcerepository.Snapshot, *string, error),
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

	if routeKey == supersedeRouteKey && replacementRecordID != nil {
		if _, err := s.recordStore.LoadEnvelopesForNoKeyUpdateTx(ctx, tx, []uuid.UUID{recordID, *replacementRecordID}); err != nil {
			return MutationResult{}, fmt.Errorf("protect timeline supersession endpoints: %w", err)
		}
	}
	requestedState := captureStateReviewed
	if routeKey == supersedeRouteKey {
		requestedState = captureStateSuperseded
	}
	if err := s.resolveActionTargetTx(ctx, tx, recordID, requestedState); err != nil {
		return MutationResult{}, err
	}
	current, err := s.loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, current.IncidentID); err != nil {
		return MutationResult{}, err
	}
	if current.RowVersion != baseRowVersion {
		return MutationResult{}, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
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
	beforeSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}

	beforeProjected := projectRecord(current, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
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

	var insertedLink *RecordLinkMutation
	if routeKey == supersedeRouteKey && validatedReplacementID != nil {
		link, err := s.linkStore.InsertSupersedesCommandTx(ctx, tx, InsertSupersedesCommand{
			IncidentID:          current.IncidentID,
			ReplacementRecordID: *validatedReplacementID,
			SupersededRecordID:  current.RecordID,
			OwnerUserID:         actor.ID,
			Now:                 now.UTC(),
		})
		if err != nil {
			if isRecordLinkConflict(err) {
				return MutationResult{}, newIllegalTransitionError(reasonGuardsNotSatisfied, current.CaptureState, captureStateSuperseded, supersedeGuardTargetMustNotHaveActiveReplacement)
			}
			return MutationResult{}, err
		}
		if link.Mutation == nil {
			return MutationResult{}, errors.New("supersede timeline event: Links returned no create mutation")
		}
		insertedLink = link.Mutation
	}

	afterProjected := projectRecord(next, validatedReplacementID)
	if err := s.hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return MutationResult{}, err
	}
	afterSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, ChangeSetParams{
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
	beforeVersion := versionid.Format(current.RecordID, current.RowVersion)
	afterVersion := versionid.Format(next.RecordID, next.RowVersion)
	if err := s.revisionsStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "timeline_record",
		RecordID:        current.RecordID,
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersion,
		AfterVersionID:  &afterVersion,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return MutationResult{}, err
	}
	if insertedLink != nil {
		if err := s.revisionsStore.AppendMutationTx(ctx, tx, MutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    2,
			TargetKind:    "record_link",
			TargetID:      insertedLink.RecordLinkID.String(),
			OperationKind: insertedLink.Operation,
			BeforeValue:   insertedLink.BeforeValue,
			AfterValue:    insertedLink.AfterValue,
		}); err != nil {
			return MutationResult{}, err
		}
	}
	changedFieldKeys := computeChangedFieldKeys(&beforeProjected, afterProjected)
	if err := s.revisionsStore.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:    changeSetID,
		RecordID:       current.RecordID,
		RowVersion:     next.RowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		ConflictFacts:  timelineRevisionFacts(beforeRow, afterRow, changedFieldKeys),
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.upsertProjectionTx(ctx, tx, afterProjected); err != nil {
		return MutationResult{}, err
	}

	payload := buildActionPayload(afterProjected, changeSetID, effectiveReason)
	if err := s.appendRecordChangeIntentTx(
		ctx,
		tx,
		current.IncidentID,
		current.RecordID,
		afterProjected.RowVersion,
		changeSetID,
		clientTxnID,
		actor.ID,
		changedFieldKeys,
		afterRow,
		0,
		now,
	); err != nil {
		return MutationResult{}, err
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
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
		ChangedFieldKeys: changedFieldKeys,
		Row:              afterProjected,
	}, nil
}

func (s *store) validateSupersedeReplacementTx(ctx context.Context, tx pgx.Tx, current sourcerepository.Snapshot, replacementRecordID *uuid.UUID) error {
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
		replacement, err := s.loadSourceRecordTx(ctx, tx, *replacementRecordID)
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
		return newIllegalTransitionError(reasonGuardsNotSatisfied, current.CaptureState, captureStateSuperseded, guards...)
	}
	return nil
}

// resolveActionTargetTx classifies a lifecycle action target from its record
// envelope before the Timeline source row is read, in REQ-01-085 and
// REQ-01-705 order: an absent record is not found, a soft-deleted record must
// be restored first, and any other record type defines no Timeline
// capture-state machine.
func (s *store) resolveActionTargetTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, requestedState string) error {
	envelope, err := s.classifyAddressedRecordTx(ctx, tx, recordID)
	if err != nil {
		return err
	}
	if envelope.RecordType != "timeline_event" {
		return newIllegalTransitionError(reasonTransitionNotAllowed, "", requestedState, "record_type_has_machine")
	}
	return nil
}

// classifyAddressedRecordTx applies the first two REQ-01-705 steps to the
// record a single-record route addresses, through the Records classifier: a
// missing record is not found and a soft-deleted one must be restored. The
// route applies its own type, incident-lifecycle and version steps next.
func (s *store) classifyAddressedRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (sourcerepository.Envelope, error) {
	envelope, err := s.recordStore.ClassifyTargetTx(ctx, tx, recordID, records.TargetLockNone)
	switch {
	case errors.Is(err, records.ErrTargetNotFound):
		return sourcerepository.Envelope{}, ErrRecordNotFound
	case errors.Is(err, records.ErrTargetDeleted):
		return sourcerepository.Envelope{}, ErrRecordDeleted
	case err != nil:
		return sourcerepository.Envelope{}, fmt.Errorf("classify timeline record target: %w", err)
	}
	return envelope, nil
}
