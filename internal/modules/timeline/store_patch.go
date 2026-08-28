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
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *store) PatchRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyPatch(ctx, actor, recordID, request, requestHash, requestID, now, patchRouteKey)
}

func (s *store) ResolveConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims conflicttokens.ConflictTokenClaims, request ConflictResolveRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if request.ResolutionKind == "keep_saved" {
		return s.clearConflict(ctx, actor, recordID, claims, request, requestHash)
	}
	if request.ResolvedChange == nil {
		return MutationResult{}, ErrNoEffectiveChange
	}
	patch := PatchRequest{
		ViewSchemaID:    TimelineViewSchemaID,
		BaseRowVersion:  claims.CurrentRowVersion,
		ClientTxnID:     request.ClientTxnID,
		CanonicalChange: []PatchChange{*request.ResolvedChange},
	}
	return s.applyPatch(ctx, actor, recordID, patch, requestHash, requestID, now, conflictResolveRouteKey)
}

func (s *store) clearConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims conflicttokens.ConflictTokenClaims, request ConflictResolveRequest, requestHash []byte) (MutationResult, error) {
	if claims.ViewSchemaID != TimelineViewSchemaID {
		return MutationResult{}, ErrRecordNotFound
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    conflictResolveRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline conflict clear payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline conflict clear idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline conflict clear transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := s.loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	projected := projectRecord(current, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &projected); err != nil {
		return MutationResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": TimelineViewSchemaID,
		"row":            buildRow(projected),
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline conflict clear transaction: %w", err)
	}
	return MutationResult{
		Payload:     payload,
		StatusCode:  http.StatusOK,
		IncidentID:  current.IncidentID,
		RecordID:    recordID,
		ClientTxnID: request.ClientTxnID,
		RowVersion:  current.RowVersion,
		Row:         projected,
	}, nil
}

func (s *store) applyPatch(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time, routeKey string) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline patch payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline patch idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := s.loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, current.IncidentID); err != nil {
		return MutationResult{}, err
	}
	if current.RowVersion < request.BaseRowVersion {
		return MutationResult{}, &RowVersionConflictError{
			RecordID:          recordID,
			BaseRowVersion:    request.BaseRowVersion,
			CurrentRowVersion: current.RowVersion,
		}
	}
	if current.RowVersion > request.BaseRowVersion {
		window, err := s.loadPatchConflictWindowTx(ctx, tx, recordID, request.BaseRowVersion, current.RowVersion)
		if err != nil {
			return MutationResult{}, err
		}
		if change, changed, ok := overlappingPatchChange(request.CanonicalChange, window.ChangedFields); ok {
			currentProjected := projectRecord(current, nil)
			if err := s.hydrateProjectedCollections(ctx, tx, &currentProjected); err != nil {
				return MutationResult{}, err
			}
			conflict, err := s.buildSameFieldConflict(recordID, currentProjected, request.BaseRowVersion, requestHash, window, change, changed)
			if err != nil {
				return MutationResult{}, err
			}
			return MutationResult{}, conflict
		}
	}
	if current.CaptureState == "superseded" {
		return MutationResult{}, newIllegalTransitionError("superseded_terminal", current.CaptureState, captureStateEnriched)
	}

	next := current
	mentionChanged := false
	tagChanged := false
	evidenceChanged := false
	for _, change := range request.CanonicalChange {
		switch {
		case change.FieldKey == "timeline.date_entered_text":
			next.DateEnteredText = change.TextValue
		case change.FieldKey == "timeline.analyst_text":
			next.AnalystText = change.TextValue
		case change.FieldKey == "timeline.mitre_stage_text":
			next.MitreStageText = change.TextValue
		case change.FieldKey == "timeline.device_object_text":
			next.DeviceObjectText = change.TextValue
		case change.FieldKey == "timeline.ip_address_text":
			next.IPAddressText = change.TextValue
		case change.FieldKey == "timeline.activity_utc_text":
			next.ActivityUTCText = change.TextValue
			next.ActivityUTCGenerated = false
		case change.FieldKey == "timeline.activity_local_text":
			next.ActivityLocalText = change.TextValue
			next.ActivityLocalGenerated = false
		case change.FieldKey == "timeline.raw_activity_text":
			next.RawActivityText = change.TextValue
		case change.FieldKey == "timeline.activity_synopsis_text":
			next.ActivitySynopsisText = change.TextValue
		case change.FieldKey == "timeline.data_source_text":
			next.DataSourceText = change.TextValue
		case isTimelineMentionCollection(change.FieldKey):
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				mentionChanged = true
			}
		case isTimelineTagCollection(change.FieldKey):
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				tagChanged = true
			}
		case isTimelineAttachedEvidenceCollection(change.FieldKey):
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				evidenceChanged = true
			}
		}
	}
	profile, err := getTimeConversionProfileTx(ctx, tx, current.IncidentID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	applyTimelineTimeConversion(&next, profile)

	beforeProjected := projectRecord(current, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return MutationResult{}, err
	}
	materialChanged := hasMaterialChange(current, next)
	var mentionLinkMutations []attachedEvidenceMutation
	if mentionChanged {
		mentionResult, err := s.applyPatchMentionActionsTx(ctx, tx, actor, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
		if err := s.refreshMentionEntityProjectionsTx(ctx, tx, mentionResult.Projection); err != nil {
			return MutationResult{}, err
		}
		mentionLinkMutations = mentionResult.LinkMutations
	}
	var tagMutations []recordTagMutation
	if tagChanged {
		var err error
		tagMutations, err = s.applyPatchTagActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
		tagChanged = len(tagMutations) > 0
	}
	var attachedEvidenceMutations []attachedEvidenceMutation
	if evidenceChanged {
		var err error
		attachedEvidenceMutations, err = s.applyPatchAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
	}
	linkMutations := append(mentionLinkMutations, attachedEvidenceMutations...)
	if !materialChanged && !mentionChanged && !tagChanged && !evidenceChanged {
		return MutationResult{}, ErrNoEffectiveChange
	}
	stateMaterialChanged := materialChanged || mentionChanged || evidenceChanged
	if stateMaterialChanged {
		nextState, err := captureStateAfterMaterialPatch(current.CaptureState)
		if err != nil {
			return MutationResult{}, err
		}
		next.CaptureState = nextState
	} else {
		next.CaptureState = current.CaptureState
	}
	beforeSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return MutationResult{}, err
	}
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	next.EditedAt = now.UTC()
	next.UpdatedByUserID = actor.ID
	if stateMaterialChanged && current.CaptureState == captureStateReviewed {
		next.ReviewedAt = nil
		next.ReviewedByUserID = nil
	}

	if err := tx.QueryRow(ctx, `
UPDATE timeline_events
   SET date_entered_text = $2,
       analyst_text = $3,
       mitre_stage_text = $4,
       device_object_text = $5,
       ip_address_text = $6,
       activity_utc_text = $7,
       activity_local_text = $8,
       raw_activity_text = $9,
       activity_synopsis_text = $10,
       data_source_text = $11,
       activity_utc_generated = $12,
       activity_local_generated = $13,
       activity_time_pair_state = $14,
       capture_state = $15,
       row_version = $16,
       edited_at = $17,
       updated_by_user_id = $18,
       reviewed_at = $19,
       reviewed_by_user_id = $20
 WHERE record_id = $1
RETURNING recorded_at
`, recordID, next.DateEnteredText, next.AnalystText, next.MitreStageText, next.DeviceObjectText, next.IPAddressText, next.ActivityUTCText, next.ActivityLocalText, next.RawActivityText, next.ActivitySynopsisText, next.DataSourceText, next.ActivityUTCGenerated, next.ActivityLocalGenerated, next.ActivityTimePairState, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID).Scan(&next.RecordedAt); err != nil {
		return MutationResult{}, fmt.Errorf("update timeline record: %w", err)
	}

	afterProjected := projectRecord(next, nil)
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
		ClientTxnID: &request.ClientTxnID,
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
	if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, 2, linkMutations); err != nil {
		return MutationResult{}, err
	}
	if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, 2+len(linkMutations), tagMutations); err != nil {
		return MutationResult{}, err
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

	payload := buildMutationPayload(afterProjected, changeSetID)
	if err := s.appendRecordChangeIntentTx(
		ctx,
		tx,
		current.IncidentID,
		current.RecordID,
		afterProjected.RowVersion,
		changeSetID,
		request.ClientTxnID,
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
		return MutationResult{}, fmt.Errorf("commit timeline patch transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       current.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       afterProjected.RowVersion,
		ChangedFieldKeys: changedFieldKeys,
		Row:              afterProjected,
	}, nil
}

func (s *store) loadPatchConflictWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) (patchConflictWindow, error) {
	rows, err := s.revisionsStore.ListRecordRevisionWindowTx(ctx, tx, recordID, baseRowVersion, currentRowVersion)
	if err != nil {
		return patchConflictWindow{}, fmt.Errorf("query timeline patch conflict window: %w", err)
	}
	generic, err := conflicttokens.BuildCanonicalPatchConflictWindow(
		recordID,
		baseRowVersion,
		currentRowVersion,
		rows,
		s.conflictFields,
		s.conflictSnapshots,
	)
	if err != nil {
		return patchConflictWindow{}, newRowVersionConflict(recordID, baseRowVersion, currentRowVersion)
	}
	ensureEmptyTimelineCollectionCells(generic.BaseRow)
	window := patchConflictWindow{
		BaseRow:       generic.BaseRow,
		ChangedFields: make(map[string]patchChangedField, len(generic.ChangedFields)),
	}
	for fieldKey, changed := range generic.ChangedFields {
		window.ChangedFields[fieldKey] = patchChangedField{
			FieldKey:        fieldKey,
			ServerUpdatedBy: changed.ServerUpdatedBy,
			ServerUpdatedAt: changed.ServerUpdatedAt.UTC(),
		}
	}
	return window, nil
}

func ensureEmptyTimelineCollectionCells(row map[string]any) {
	cells, _ := row["cells"].(map[string]any)
	for _, fieldKey := range []string{"timeline.host_refs", "timeline.identity_refs", "timeline.tags", "timeline.attached_evidence_ids"} {
		if _, present := cells[fieldKey]; !present {
			policy, _ := LookupCollectionPolicy(fieldKey)
			cells[fieldKey] = map[string]any{"value": valuecodec.Collection(policy.Ordered, nil)}
		}
	}
}
