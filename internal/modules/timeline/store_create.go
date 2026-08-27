package timeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type createRowOptions struct {
	allowInteractiveAutoResolution bool
}

func (s *store) CreateRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.createRow(ctx, actor, incidentID, request, requestHash, requestID, now, createRowOptions{
		allowInteractiveAutoResolution: true,
	})
}

func (s *store) createRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time, options createRowOptions) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + TimelineViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    createRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			IncidentID: incidentID,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline create idempotency: %w", err)
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return MutationResult{}, err
	}
	changeSetID := uuid.New()
	if _, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, ChangeSetParams{
		ChangeSetID: &changeSetID,
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      createRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	}); err != nil {
		return MutationResult{}, err
	}
	result, _, err := s.createRowTx(
		ctx,
		tx,
		actor.ID,
		incidentID,
		request,
		changeSetID,
		func(int) (int, error) { return 1, nil },
		now.UTC(),
		options,
	)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, result.Payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline create transaction: %w", err)
	}
	return result, nil
}

func (s *store) createRowTx(
	ctx context.Context,
	tx pgx.Tx,
	actorUserID uuid.UUID,
	incidentID uuid.UUID,
	request CreateRequest,
	changeSetID uuid.UUID,
	allocateMutationSequence func(int) (int, error),
	now time.Time,
	options createRowOptions,
) (MutationResult, int, error) {
	recordID := uuid.New()
	current := sourcerepository.Snapshot{
		RecordID:              recordID,
		IncidentID:            incidentID,
		DateEnteredText:       request.DateEnteredText,
		AnalystText:           request.AnalystText,
		MitreStageText:        request.MitreStageText,
		DeviceObjectText:      request.DeviceObjectText,
		IPAddressText:         request.IPAddressText,
		ActivityUTCText:       request.ActivityUTCText,
		ActivityLocalText:     request.ActivityLocalText,
		RawActivityText:       request.RawActivityText,
		ActivitySynopsisText:  request.ActivitySynopsisText,
		DataSourceText:        request.DataSourceText,
		ActivityTimePairState: "disabled",
		CaptureState:          initialCaptureState(),
		RowVersion:            1,
		RecordedAt:            now.UTC(),
		EditedAt:              now.UTC(),
		CreatedByUserID:       actorUserID,
		UpdatedByUserID:       actorUserID,
	}
	profile, err := getTimeConversionProfileTx(ctx, tx, incidentID, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}
	applyTimelineTimeConversion(&current, profile)
	if _, err := s.recordStore.InsertTx(ctx, tx, RecordCreateParams{
		RecordID:        &current.RecordID,
		IncidentID:      incidentID,
		RecordType:      "timeline_event",
		CreatedByUserID: actorUserID,
		CreatedAt:       now.UTC(),
		UpdatedByUserID: actorUserID,
		UpdatedAt:       now.UTC(),
		RowVersion:      1,
	}); err != nil {
		return MutationResult{}, 0, fmt.Errorf("insert timeline record envelope: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO timeline_events (
    record_id, incident_id, date_entered_text, analyst_text, mitre_stage_text,
    device_object_text, ip_address_text, activity_utc_text, activity_local_text,
    raw_activity_text, activity_synopsis_text, data_source_text,
    activity_utc_generated, activity_local_generated, activity_time_pair_state,
    capture_state, row_version, recorded_at, edited_at,
    created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 'rough', 1, $4, $4, $3, $3)
`, current.RecordID, incidentID, actorUserID, now.UTC(), current.DateEnteredText, current.AnalystText, current.MitreStageText, current.DeviceObjectText, current.IPAddressText, current.ActivityUTCText, current.ActivityLocalText, current.RawActivityText, current.ActivitySynopsisText, current.DataSourceText, current.ActivityUTCGenerated, current.ActivityLocalGenerated, current.ActivityTimePairState); err != nil {
		return MutationResult{}, 0, fmt.Errorf("insert timeline source row: %w", err)
	}
	if err := insertSourceProvenanceTx(ctx, tx, current.RecordID, request.RawCaptureColumns); err != nil {
		return MutationResult{}, 0, err
	}

	mentionResult, err := s.applyCreateMentionActionsTx(ctx, tx, actorUserID, current.IncidentID, current.RecordID, request.HostRefs, request.IdentityRefs, options, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.refreshMentionEntityProjectionsTx(ctx, tx, mentionResult.Projection); err != nil {
		return MutationResult{}, 0, err
	}
	tagMutations, err := s.applyCreateTagActionsTx(ctx, tx, actorUserID, current.IncidentID, current.RecordID, request.Tags, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}
	attachedEvidenceMutations, err := s.applyAttachedEvidenceActionsTx(ctx, tx, actorUserID, current.IncidentID, current.RecordID, request.AttachedEvidence, now.UTC())
	if err != nil {
		return MutationResult{}, 0, err
	}

	projected := projectRecord(current, nil)
	if createRequestHasCollectionActions(request) {
		if err := s.hydrateProjectedCollections(ctx, tx, &projected); err != nil {
			return MutationResult{}, 0, err
		}
	}
	linkMutations := append(mentionResult.LinkMutations, attachedEvidenceMutations...)
	mutationSequence, err := allocateMutationSequence(1 + len(linkMutations) + len(tagMutations))
	if err != nil {
		return MutationResult{}, 0, err
	}

	afterRow := buildRow(projected)
	afterSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return MutationResult{}, 0, err
	}
	afterVersion := versionid.Format(current.RecordID, projected.RowVersion)
	if err := s.revisionsStore.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     mutationSequence,
		TargetKind:     "timeline_record",
		RecordID:       current.RecordID,
		OperationKind:  "create",
		AfterVersionID: &afterVersion,
		AfterSnapshot:  &afterSnapshot,
	}); err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, mutationSequence+1, linkMutations); err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, mutationSequence+1+len(linkMutations), tagMutations); err != nil {
		return MutationResult{}, 0, err
	}
	changedFieldKeys := computeChangedFieldKeys(nil, projected)
	if err := s.revisionsStore.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:   changeSetID,
		RecordID:      current.RecordID,
		RowVersion:    projected.RowVersion,
		AfterSnapshot: &afterSnapshot,
		ConflictFacts: timelineRevisionFacts(nil, afterRow, changedFieldKeys),
	}); err != nil {
		return MutationResult{}, 0, err
	}
	if err := s.upsertProjectionTx(ctx, tx, projected); err != nil {
		return MutationResult{}, 0, err
	}

	payload := buildMutationPayload(projected, changeSetID)
	if err := s.appendRecordChangeIntentTx(
		ctx,
		tx,
		incidentID,
		current.RecordID,
		projected.RowVersion,
		changeSetID,
		request.ClientTxnID,
		actorUserID,
		changedFieldKeys,
		afterRow,
		0,
		now,
	); err != nil {
		return MutationResult{}, 0, err
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         current.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       projected.RowVersion,
		ChangedFieldKeys: computeChangedFieldKeys(nil, projected),
		Row:              projected,
	}, mutationSequence, nil
}

func createRequestHasCollectionActions(request CreateRequest) bool {
	return (request.HostRefs != nil && len(request.HostRefs.Actions) > 0) ||
		(request.IdentityRefs != nil && len(request.IdentityRefs.Actions) > 0) ||
		(request.Tags != nil && len(request.Tags.Actions) > 0) ||
		(request.AttachedEvidence != nil && len(request.AttachedEvidence.Actions) > 0)
}
