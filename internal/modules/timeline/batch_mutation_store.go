package timeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/versionid"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const (
	clipboardPasteRouteKey = "timeline.clipboard_paste"
	bulkMutationRouteKey   = "workbook.bulk_mutations"
)

type ownerBatchApplyV1 struct {
	ClientTxnID string
	Operation   string
	Targets     []OwnerBatchTargetV1
	Rows        []ownerBatchRowPlanV1
	RequestHash []byte
	RequestID   string
	Now         time.Time
}

func (s *store) applyOwnerBatchV1(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request ownerBatchApplyV1) (BatchMutationResult, error) {
	if err := validateOwnerBatchShape(request); err != nil {
		return BatchMutationResult{}, err
	}
	scopeKey := incidentID.String() + ":" + TimelineViewSchemaID
	routeKey, originKind, err := ownerBatchOperationMetadata(request.Operation)
	if err != nil {
		return BatchMutationResult{}, err
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, request.RequestHash) {
			return BatchMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return BatchMutationResult{}, fmt.Errorf("decode replayed timeline batch mutation payload: %w", err)
		}
		return BatchMutationResult{
			Payload:     payload,
			StatusCode:  http.StatusOK,
			Replayed:    true,
			IncidentID:  incidentID,
			ClientTxnID: request.ClientTxnID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return BatchMutationResult{}, fmt.Errorf("query timeline batch mutation idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return BatchMutationResult{}, fmt.Errorf("begin timeline batch mutation transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, incidentID); err != nil {
		return BatchMutationResult{}, err
	}
	if err := s.validateOwnerBatchTargetsTx(ctx, tx, incidentID, request.Operation, request.Targets); err != nil {
		return BatchMutationResult{}, err
	}

	applied := make([]batchAppliedRow, 0, len(request.Rows))
	conflicts := make([]map[string]any, 0)
	for index, rowPlan := range request.Rows {
		target := request.Targets[index]
		switch target.Kind {
		case "create":
			row, err := s.applyOwnerBatchCreateTx(ctx, tx, actor, incidentID, rowPlan, originKind, request.Now.UTC())
			if err != nil {
				return BatchMutationResult{}, err
			}
			applied = append(applied, row)
		case "record":
			row, rowConflicts, err := s.applyOwnerBatchPatchTx(ctx, tx, actor, incidentID, target.RecordID, target.BaseRowVersion, request.RequestHash, rowPlan, originKind, request.Now.UTC())
			if err != nil {
				return BatchMutationResult{}, err
			}
			conflicts = append(conflicts, rowConflicts...)
			if row.RecordID != uuid.Nil {
				applied = append(applied, row)
			}
		default:
			return BatchMutationResult{}, fmt.Errorf("unsupported batch target kind: %s", target.Kind)
		}
	}
	if len(applied) == 0 {
		payload := map[string]any{
			"view_schema_id": TimelineViewSchemaID,
			"conflicts":      conflicts,
			"rows":           []any{},
		}
		if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, request.RequestHash, http.StatusOK, payload); err != nil {
			if authn.IsUniqueViolation(err) {
				return BatchMutationResult{}, authn.ErrClientTxnConflict
			}
			return BatchMutationResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return BatchMutationResult{}, fmt.Errorf("commit timeline batch mutation conflicts-only transaction: %w", err)
		}
		return BatchMutationResult{
			Payload:     payload,
			StatusCode:  http.StatusOK,
			IncidentID:  incidentID,
			ClientTxnID: request.ClientTxnID,
		}, nil
	}

	changeSetID, err := s.revisionsStore.AppendChangeSetTx(ctx, tx, ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &request.RequestID,
		CreatedAt:   request.Now.UTC(),
	})
	if err != nil {
		return BatchMutationResult{}, err
	}

	sequenceNo := 1
	resultRows := make([]BatchMutationRowResult, 0, len(applied))
	payloadRows := make([]map[string]any, 0, len(applied))
	for _, row := range applied {
		beforeVersion := ""
		afterVersion := versionid.Format(row.After.RecordID, row.After.RowVersion)
		params := revisions.AppendRecordMutationParams{
			ChangeSetID:    changeSetID,
			SequenceNo:     sequenceNo,
			TargetKind:     "timeline_record",
			RecordID:       row.After.RecordID,
			OperationKind:  row.Operation,
			AfterVersionID: &afterVersion,
			AfterSnapshot:  &row.AfterSnapshot,
		}
		if row.Before != nil {
			beforeVersion = versionid.Format(row.Before.RecordID, row.Before.RowVersion)
			params.BeforeVersionID = &beforeVersion
			params.BeforeSnapshot = row.BeforeSnapshot
		}
		if err := s.revisionsStore.AppendRecordMutationTx(ctx, tx, params); err != nil {
			return BatchMutationResult{}, err
		}
		sequenceNo++
		if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, sequenceNo, row.LinkMutations); err != nil {
			return BatchMutationResult{}, err
		}
		sequenceNo += len(row.LinkMutations)
		if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, sequenceNo, row.TagMutations); err != nil {
			return BatchMutationResult{}, err
		}
		sequenceNo += len(row.TagMutations)
		revision := revisions.LiveRevisionInput{
			ChangeSetID:   changeSetID,
			RecordID:      row.After.RecordID,
			RowVersion:    row.After.RowVersion,
			AfterSnapshot: &row.AfterSnapshot,
			ConflictFacts: timelineRevisionFacts(row.BeforeRow, row.AfterRow, row.ChangedFieldKeys),
		}
		if row.Before != nil {
			revision.BeforeSnapshot = row.BeforeSnapshot
		}
		if err := s.revisionsStore.AppendLiveRevisionTx(ctx, tx, revision); err != nil {
			return BatchMutationResult{}, err
		}
		if err := s.upsertProjectionTx(ctx, tx, row.After); err != nil {
			return BatchMutationResult{}, err
		}
		result := BatchMutationRowResult{
			RecordID:         row.After.RecordID,
			RowVersion:       row.After.RowVersion,
			ChangedFieldKeys: row.ChangedFieldKeys,
			Row:              row.AfterRow,
		}
		if err := s.appendRecordChangeIntentTx(
			ctx,
			tx,
			incidentID,
			row.After.RecordID,
			row.After.RowVersion,
			changeSetID,
			request.ClientTxnID,
			actor.ID,
			row.ChangedFieldKeys,
			row.AfterRow,
			len(resultRows),
			request.Now,
		); err != nil {
			return BatchMutationResult{}, err
		}
		resultRows = append(resultRows, result)
		payloadRows = append(payloadRows, row.AfterRow)
	}

	payload := map[string]any{
		"view_schema_id": TimelineViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"rows":           payloadRows,
		"conflicts":      conflicts,
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, request.RequestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return BatchMutationResult{}, authn.ErrClientTxnConflict
		}
		return BatchMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return BatchMutationResult{}, fmt.Errorf("commit timeline batch mutation transaction: %w", err)
	}
	return BatchMutationResult{
		Payload:     payload,
		StatusCode:  http.StatusOK,
		IncidentID:  incidentID,
		ChangeSetID: changeSetID,
		ClientTxnID: request.ClientTxnID,
		Rows:        resultRows,
	}, nil
}

func validateOwnerBatchShape(request ownerBatchApplyV1) error {
	if len(request.Targets) > mutationpolicy.MaxOwnerBatchTargets {
		return fmt.Errorf("owner_batch_apply_v1 target count exceeds %d", mutationpolicy.MaxOwnerBatchTargets)
	}
	if strings.TrimSpace(request.ClientTxnID) == "" {
		return fmt.Errorf("owner_batch_apply_v1 client transaction ID is required")
	}
	if len(request.Targets) == 0 || len(request.Targets) != len(request.Rows) {
		return fmt.Errorf("owner_batch_apply_v1 targets and rows must be nonempty and aligned")
	}
	if request.RequestHash == nil {
		return fmt.Errorf("owner_batch_apply_v1 request hash is required")
	}
	_, _, err := ownerBatchOperationMetadata(request.Operation)
	return err
}

func ownerBatchOperationMetadata(operation string) (routeKey string, originKind string, err error) {
	switch operation {
	case OwnerBatchOperationClipboardPasteV1:
		return clipboardPasteRouteKey, "clipboard_paste", nil
	case OwnerBatchOperationFillDownV1, OwnerBatchOperationMultiRowTagAssignmentV1:
		return bulkMutationRouteKey, "bulk_edit", nil
	default:
		return "", "", fmt.Errorf("unsupported owner_batch_apply_v1 operation %q", operation)
	}
}

func (s *store) validateOwnerBatchTargetsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, operation string, targets []OwnerBatchTargetV1) error {
	recordIDs := make([]uuid.UUID, 0, len(targets))
	for _, target := range targets {
		switch target.Kind {
		case "create":
			if operation != OwnerBatchOperationClipboardPasteV1 || target.RecordID != uuid.Nil || target.BaseRowVersion != 0 {
				return ErrRecordNotFound
			}
		case "record":
			if target.RecordID == uuid.Nil || target.BaseRowVersion < 1 {
				return ErrRecordNotFound
			}
			recordIDs = append(recordIDs, target.RecordID)
		default:
			return ErrRecordNotFound
		}
	}
	envelopes, err := s.recordStore.LoadEnvelopesTx(ctx, tx, recordIDs, true)
	if err != nil {
		return err
	}
	for _, recordID := range recordIDs {
		envelope, ok := envelopes[recordID]
		if !ok || envelope.IncidentID != incidentID || envelope.RecordType != "timeline_event" || envelope.DeletedAt != nil {
			return ErrRecordNotFound
		}
		if _, err := s.loadSourceRecordForIncidentTx(ctx, tx, incidentID, recordID); err != nil {
			return err
		}
	}
	return nil
}

type batchAppliedRow struct {
	Operation        string
	Before           *workbookprojection.DerivedRecord
	After            workbookprojection.DerivedRecord
	BeforeRow        map[string]any
	AfterRow         map[string]any
	BeforeSnapshot   *revisions.RecordSnapshot
	AfterSnapshot    revisions.RecordSnapshot
	ChangedFieldKeys []string
	TagMutations     []recordTagMutation
	LinkMutations    []attachedEvidenceMutation
	RecordID         uuid.UUID
}

func (s *store) applyOwnerBatchCreateTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, rowPlan ownerBatchRowPlanV1, originKind string, now time.Time) (batchAppliedRow, error) {
	recordID := uuid.New()
	current := sourcerepository.Snapshot{
		RecordID:              recordID,
		IncidentID:            incidentID,
		ActivityTimePairState: "disabled",
		CaptureState:          initialCaptureState(),
		RowVersion:            1,
		RecordedAt:            now.UTC(),
		EditedAt:              now.UTC(),
		CreatedByUserID:       actor.ID,
		UpdatedByUserID:       actor.ID,
	}
	for _, cell := range rowPlan.Cells {
		applyPatchChangeToSource(&current, cell.Change)
	}
	profile, err := getTimeConversionProfileTx(ctx, tx, incidentID, now.UTC())
	if err != nil {
		return batchAppliedRow{}, err
	}
	applyTimelineTimeConversion(&current, profile)
	if _, err := s.recordStore.InsertTx(ctx, tx, RecordCreateParams{
		RecordID:        &current.RecordID,
		IncidentID:      incidentID,
		RecordType:      "timeline_event",
		CreatedByUserID: actor.ID,
		CreatedAt:       now.UTC(),
		UpdatedByUserID: actor.ID,
		UpdatedAt:       now.UTC(),
		RowVersion:      1,
	}); err != nil {
		return batchAppliedRow{}, fmt.Errorf("insert timeline batch record envelope: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO timeline_events (
    record_id,
    incident_id,
    date_entered_text,
    analyst_text,
    mitre_stage_text,
    device_object_text,
    ip_address_text,
    activity_utc_text,
    activity_local_text,
    raw_activity_text,
    activity_synopsis_text,
    data_source_text,
    activity_utc_generated,
    activity_local_generated,
    activity_time_pair_state,
    capture_state,
    row_version,
    recorded_at,
    edited_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 'rough', 1, $4, $4, $3, $3)
`, current.RecordID, incidentID, actor.ID, now.UTC(), current.DateEnteredText, current.AnalystText, current.MitreStageText, current.DeviceObjectText, current.IPAddressText, current.ActivityUTCText, current.ActivityLocalText, current.RawActivityText, current.ActivitySynopsisText, current.DataSourceText, current.ActivityUTCGenerated, current.ActivityLocalGenerated, current.ActivityTimePairState); err != nil {
		return batchAppliedRow{}, fmt.Errorf("insert timeline batch row: %w", err)
	}
	if err := insertSourceProvenanceTx(ctx, tx, current.RecordID, rowPlan.Unmapped); err != nil {
		return batchAppliedRow{}, err
	}
	mentionResult, err := s.applyBatchMentionActionsTx(ctx, tx, actor, current.IncidentID, current.RecordID, rowPlan.Cells, originKind, now.UTC())
	if err != nil {
		return batchAppliedRow{}, err
	}
	if err := s.refreshMentionEntityProjectionsTx(ctx, tx, mentionResult.Projection); err != nil {
		return batchAppliedRow{}, err
	}
	tagMutations, err := s.applyBatchTagActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, rowPlan.Cells, now.UTC())
	if err != nil {
		return batchAppliedRow{}, err
	}
	attachedEvidenceMutations, err := s.applyBatchAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, rowPlan.Cells, now.UTC())
	if err != nil {
		return batchAppliedRow{}, err
	}
	projected := projectRecord(current, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &projected); err != nil {
		return batchAppliedRow{}, err
	}
	afterRow := buildRow(projected)
	afterSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return batchAppliedRow{}, err
	}
	linkMutations := append(mentionResult.LinkMutations, attachedEvidenceMutations...)
	return batchAppliedRow{
		Operation:        "create",
		After:            projected,
		AfterRow:         afterRow,
		AfterSnapshot:    afterSnapshot,
		ChangedFieldKeys: computeChangedFieldKeys(nil, projected),
		TagMutations:     tagMutations,
		LinkMutations:    linkMutations,
		RecordID:         projected.RecordID,
	}, nil
}

func ensureOwnerBatchRecordIncident(current sourcerepository.Snapshot, incidentID uuid.UUID) error {
	if current.IncidentID != incidentID {
		return ErrRecordNotFound
	}
	return nil
}

func (s *store) applyOwnerBatchPatchTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, baseRowVersion int64, requestHash []byte, rowPlan ownerBatchRowPlanV1, originKind string, now time.Time) (batchAppliedRow, []map[string]any, error) {
	current, err := s.loadSourceRecordForIncidentTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return batchAppliedRow{}, nil, err
	}
	if err := ensureOwnerBatchRecordIncident(current, incidentID); err != nil {
		return batchAppliedRow{}, nil, err
	}
	if current.RowVersion < baseRowVersion {
		return batchAppliedRow{}, nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	acceptedCells := append([]ownerBatchCellV1{}, rowPlan.Cells...)
	conflicts := make([]map[string]any, 0)
	if current.RowVersion > baseRowVersion && len(rowPlan.Cells) > 0 {
		window, err := s.loadPatchConflictWindowTx(ctx, tx, recordID, baseRowVersion, current.RowVersion)
		if err != nil {
			return batchAppliedRow{}, nil, err
		}
		currentProjected := projectRecord(current, nil)
		if err := s.hydrateProjectedCollections(ctx, tx, &currentProjected); err != nil {
			return batchAppliedRow{}, nil, err
		}
		kept := acceptedCells[:0]
		for _, cell := range acceptedCells {
			if changed, ok := window.ChangedFields[cell.FieldKey]; ok {
				conflict, err := s.buildSameFieldConflict(recordID, currentProjected, baseRowVersion, requestHash, window, cell.Change, changed)
				if err != nil {
					return batchAppliedRow{}, nil, err
				}
				conflicts = append(conflicts, conflict.Conflict.PublicValue())
				continue
			}
			kept = append(kept, cell)
		}
		acceptedCells = kept
	}
	if current.CaptureState == "superseded" {
		return batchAppliedRow{}, nil, newIllegalTransitionError("superseded_terminal", current.CaptureState, captureStateEnriched)
	}
	next := current
	for _, cell := range acceptedCells {
		applyPatchChangeToSource(&next, cell.Change)
	}
	profile, err := getTimeConversionProfileTx(ctx, tx, current.IncidentID, now.UTC())
	if err != nil {
		return batchAppliedRow{}, nil, err
	}
	applyTimelineTimeConversion(&next, profile)
	beforeProjected := projectRecord(current, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return batchAppliedRow{}, nil, err
	}
	mentionChanged := batchCellsIncludeField(acceptedCells, "timeline.host_refs") || batchCellsIncludeField(acceptedCells, "timeline.identity_refs")
	tagChanged := batchCellsIncludeField(acceptedCells, "timeline.tags")
	evidenceChanged := batchCellsIncludeField(acceptedCells, "timeline.attached_evidence_ids")
	provenanceChanged := len(rowPlan.Unmapped) > 0
	materialChanged := hasMaterialChange(current, next) || provenanceChanged
	var mentionLinkMutations []attachedEvidenceMutation
	if mentionChanged {
		mentionResult, err := s.applyBatchMentionActionsTx(ctx, tx, actor, current.IncidentID, recordID, acceptedCells, originKind, now.UTC())
		if err != nil {
			return batchAppliedRow{}, nil, err
		}
		if err := s.refreshMentionEntityProjectionsTx(ctx, tx, mentionResult.Projection); err != nil {
			return batchAppliedRow{}, nil, err
		}
		mentionLinkMutations = mentionResult.LinkMutations
	}
	var tagMutations []recordTagMutation
	if tagChanged {
		tagMutations, err = s.applyBatchTagActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, acceptedCells, now.UTC())
		if err != nil {
			return batchAppliedRow{}, nil, err
		}
		tagChanged = len(tagMutations) > 0
	}
	var attachedEvidenceMutations []attachedEvidenceMutation
	if evidenceChanged {
		attachedEvidenceMutations, err = s.applyBatchAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, acceptedCells, now.UTC())
		if err != nil {
			return batchAppliedRow{}, nil, err
		}
		evidenceChanged = len(attachedEvidenceMutations) > 0
	}
	linkMutations := append(mentionLinkMutations, attachedEvidenceMutations...)
	if !materialChanged && !mentionChanged && !tagChanged && !evidenceChanged {
		return batchAppliedRow{}, conflicts, nil
	}
	stateMaterialChanged := materialChanged || mentionChanged || evidenceChanged
	if stateMaterialChanged {
		nextState, err := captureStateAfterMaterialPatch(current.CaptureState)
		if err != nil {
			return batchAppliedRow{}, nil, err
		}
		next.CaptureState = nextState
	} else {
		next.CaptureState = current.CaptureState
	}
	beforeSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return batchAppliedRow{}, nil, err
	}
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return batchAppliedRow{}, nil, err
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
   AND incident_id = $21
RETURNING recorded_at
`, recordID, next.DateEnteredText, next.AnalystText, next.MitreStageText, next.DeviceObjectText, next.IPAddressText, next.ActivityUTCText, next.ActivityLocalText, next.RawActivityText, next.ActivitySynopsisText, next.DataSourceText, next.ActivityUTCGenerated, next.ActivityLocalGenerated, next.ActivityTimePairState, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID, incidentID).Scan(&next.RecordedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return batchAppliedRow{}, nil, ErrRecordNotFound
		}
		return batchAppliedRow{}, nil, fmt.Errorf("update timeline batch record: %w", err)
	}
	if err := insertSourceProvenanceTx(ctx, tx, recordID, rowPlan.Unmapped); err != nil {
		return batchAppliedRow{}, nil, err
	}
	afterProjected := projectRecord(next, nil)
	if err := s.hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return batchAppliedRow{}, nil, err
	}
	beforeRow := buildRow(beforeProjected)
	afterRow := buildRow(afterProjected)
	afterSnapshot, err := s.revisionsStore.CaptureRecordSnapshotTx(ctx, tx, current.RecordID)
	if err != nil {
		return batchAppliedRow{}, nil, err
	}
	return batchAppliedRow{
		Operation:        "patch",
		Before:           &beforeProjected,
		After:            afterProjected,
		BeforeRow:        beforeRow,
		AfterRow:         afterRow,
		BeforeSnapshot:   &beforeSnapshot,
		AfterSnapshot:    afterSnapshot,
		ChangedFieldKeys: computeChangedFieldKeys(&beforeProjected, afterProjected),
		TagMutations:     tagMutations,
		LinkMutations:    linkMutations,
		RecordID:         afterProjected.RecordID,
	}, conflicts, nil
}

func applyPatchChangeToSource(record *sourcerepository.Snapshot, change PatchChange) {
	switch change.FieldKey {
	case "timeline.date_entered_text":
		record.DateEnteredText = change.TextValue
	case "timeline.analyst_text":
		record.AnalystText = change.TextValue
	case "timeline.mitre_stage_text":
		record.MitreStageText = change.TextValue
	case "timeline.device_object_text":
		record.DeviceObjectText = change.TextValue
	case "timeline.ip_address_text":
		record.IPAddressText = change.TextValue
	case "timeline.activity_utc_text":
		record.ActivityUTCText = change.TextValue
		record.ActivityUTCGenerated = false
	case "timeline.activity_local_text":
		record.ActivityLocalText = change.TextValue
		record.ActivityLocalGenerated = false
	case "timeline.raw_activity_text":
		record.RawActivityText = change.TextValue
	case "timeline.activity_synopsis_text":
		record.ActivitySynopsisText = change.TextValue
	case "timeline.data_source_text":
		record.DataSourceText = change.TextValue
	}
}

func batchCellsIncludeField(cells []ownerBatchCellV1, fieldKey string) bool {
	return slices.ContainsFunc(cells, func(cell ownerBatchCellV1) bool {
		return cell.FieldKey == fieldKey && cell.Change.ActionPayload != nil && len(cell.Change.ActionPayload.Actions) > 0
	})
}

func (s *store) applyBatchMentionActionsTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, cells []ownerBatchCellV1, originKind string, now time.Time) (mentionApplicationResult, error) {
	var result mentionApplicationResult
	for _, cell := range cells {
		if cell.Change.ActionPayload == nil || (cell.FieldKey != "timeline.host_refs" && cell.FieldKey != "timeline.identity_refs") {
			continue
		}
		entityType := "host"
		if cell.FieldKey == "timeline.identity_refs" {
			entityType = "identity"
		}
		applied, err := s.insertMentionActionsTx(ctx, tx, s.linkStore, actor.ID, incidentID, recordID, cell.FieldKey, entityType, cell.Change.ActionPayload, mentionInsertOptions{
			allowInteractiveAutoResolution: true,
			originKind:                     originKind,
		}, now)
		if err != nil {
			return mentionApplicationResult{}, err
		}
		result.merge(applied)
	}
	return result, nil
}

func (s *store) applyBatchTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, cells []ownerBatchCellV1, now time.Time) ([]recordTagMutation, error) {
	mutations := make([]recordTagMutation, 0)
	for _, cell := range cells {
		if cell.FieldKey != "timeline.tags" || cell.Change.ActionPayload == nil {
			continue
		}
		applied, err := s.applyTimelineTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, cell.Change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func (s *store) applyBatchAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, cells []ownerBatchCellV1, now time.Time) ([]attachedEvidenceMutation, error) {
	mutations := make([]attachedEvidenceMutation, 0)
	for _, cell := range cells {
		if cell.FieldKey != "timeline.attached_evidence_ids" || cell.Change.ActionPayload == nil {
			continue
		}
		applied, err := s.applyAttachedEvidenceActionsTx(ctx, tx, actorUserID, incidentID, recordID, cell.Change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}
