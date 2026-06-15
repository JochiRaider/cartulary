package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *Store) ClipboardPaste(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request ClipboardPasteRequest, requestHash []byte, requestID string, now time.Time) (ClipboardPasteResult, error) {
	scopeKey := incidentID.String() + ":" + TimelineViewSchemaID
	routeKey := request.routeKey()
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return ClipboardPasteResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return ClipboardPasteResult{}, fmt.Errorf("decode replayed timeline clipboard paste payload: %w", err)
		}
		return ClipboardPasteResult{
			Payload:     payload,
			StatusCode:  http.StatusOK,
			Replayed:    true,
			IncidentID:  incidentID,
			ClientTxnID: request.ClientTxnID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return ClipboardPasteResult{}, fmt.Errorf("query timeline clipboard paste idempotency: %w", err)
	}

	plan, err := BuildClipboardPastePlan(request)
	if err != nil {
		return ClipboardPasteResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ClipboardPasteResult{}, fmt.Errorf("begin timeline clipboard paste transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	applied := make([]clipboardAppliedRow, 0, len(plan.Rows))
	conflicts := make([]map[string]any, 0)
	for index, rowPlan := range plan.Rows {
		target := request.Targets[index]
		switch target.Kind {
		case "create":
			row, err := s.applyClipboardPasteCreateTx(ctx, tx, actor, incidentID, rowPlan, request.sourceKind(), now.UTC())
			if err != nil {
				return ClipboardPasteResult{}, err
			}
			applied = append(applied, row)
		case "record":
			row, rowConflicts, err := s.applyClipboardPastePatchTx(ctx, tx, actor, incidentID, target.RecordID, target.BaseRowVersion, requestHash, rowPlan, request.sourceKind(), now.UTC())
			if err != nil {
				return ClipboardPasteResult{}, err
			}
			conflicts = append(conflicts, rowConflicts...)
			if row.RecordID != uuid.Nil {
				applied = append(applied, row)
			}
		default:
			return ClipboardPasteResult{}, fmt.Errorf("unsupported paste target kind: %s", target.Kind)
		}
	}
	if len(applied) == 0 {
		payload := map[string]any{
			"view_schema_id": TimelineViewSchemaID,
			"conflicts":      conflicts,
			"rows":           []any{},
		}
		if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
			if authn.IsUniqueViolation(err) {
				return ClipboardPasteResult{}, authn.ErrClientTxnConflict
			}
			return ClipboardPasteResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return ClipboardPasteResult{}, fmt.Errorf("commit timeline clipboard paste conflicts-only transaction: %w", err)
		}
		return ClipboardPasteResult{
			Payload:     payload,
			StatusCode:  http.StatusOK,
			IncidentID:  incidentID,
			ClientTxnID: request.ClientTxnID,
		}, nil
	}

	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return ClipboardPasteResult{}, err
	}

	sequenceNo := 1
	resultRows := make([]ClipboardPasteRowResult, 0, len(applied))
	payloadRows := make([]map[string]any, 0, len(applied))
	for _, row := range applied {
		beforeVersion := ""
		afterVersion := versionID(row.After.RecordID, row.After.RowVersion)
		params := revisions.MutationParams{
			ChangeSetID:    changeSetID,
			SequenceNo:     sequenceNo,
			TargetKind:     "timeline_record",
			TargetID:       row.After.RecordID.String(),
			OperationKind:  row.Operation,
			AfterVersionID: &afterVersion,
			AfterValue:     row.AfterRow,
		}
		if row.Before != nil {
			beforeVersion = versionID(row.Before.RecordID, row.Before.RowVersion)
			params.BeforeVersionID = &beforeVersion
			params.BeforeValue = row.BeforeRow
		}
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, params); err != nil {
			return ClipboardPasteResult{}, err
		}
		sequenceNo++
		if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, sequenceNo, row.AttachedEvidenceMutations); err != nil {
			return ClipboardPasteResult{}, err
		}
		sequenceNo += len(row.AttachedEvidenceMutations)
		if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, sequenceNo, row.TagMutations); err != nil {
			return ClipboardPasteResult{}, err
		}
		sequenceNo += len(row.TagMutations)
		revision := revisions.RecordRevisionParams{
			ChangeSetID: changeSetID,
			RecordID:    row.After.RecordID,
			RowVersion:  row.After.RowVersion,
			AfterValue:  row.AfterRow,
		}
		if row.Before != nil {
			revision.BeforeValue = row.BeforeRow
		}
		if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revision); err != nil {
			return ClipboardPasteResult{}, err
		}
		if err := s.projectionStore.UpsertTimelineRowTx(ctx, tx, projectionInput(row.After)); err != nil {
			return ClipboardPasteResult{}, err
		}
		result := ClipboardPasteRowResult{
			RecordID:         row.After.RecordID,
			RowVersion:       row.After.RowVersion,
			ChangedFieldKeys: row.ChangedFieldKeys,
			Row:              row.AfterRow,
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
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return ClipboardPasteResult{}, authn.ErrClientTxnConflict
		}
		return ClipboardPasteResult{}, err
	}
	for _, row := range applied {
		if err := s.beforeCommit(routeKey, row.After.RecordID); err != nil {
			return ClipboardPasteResult{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return ClipboardPasteResult{}, fmt.Errorf("commit timeline clipboard paste transaction: %w", err)
	}
	return ClipboardPasteResult{
		Payload:     payload,
		StatusCode:  http.StatusOK,
		IncidentID:  incidentID,
		ChangeSetID: changeSetID,
		ClientTxnID: request.ClientTxnID,
		Rows:        resultRows,
	}, nil
}

type clipboardAppliedRow struct {
	Operation                 string
	Before                    *projectedRecord
	After                     projectedRecord
	BeforeRow                 map[string]any
	AfterRow                  map[string]any
	ChangedFieldKeys          []string
	TagMutations              []recordTagMutation
	AttachedEvidenceMutations []attachedEvidenceMutation
	RecordID                  uuid.UUID
}

func (s *Store) applyClipboardPasteCreateTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, rowPlan ClipboardPasteRowPlan, originKind string, now time.Time) (clipboardAppliedRow, error) {
	recordID := uuid.New()
	current := sourceRecord{
		RecordID:        recordID,
		IncidentID:      incidentID,
		RawCapture:      rawCaptureWithImportColumns(nil, rowPlan.Unknown),
		CaptureState:    InitialCaptureState(),
		RowVersion:      1,
		RecordedAt:      now.UTC(),
		EditedAt:        now.UTC(),
		CreatedByUserID: actor.ID,
		UpdatedByUserID: actor.ID,
	}
	for _, cell := range rowPlan.Cells {
		applyPatchChangeToSource(&current, cell.Change)
	}
	rawCaptureJSON, err := json.Marshal(current.RawCapture)
	if err != nil {
		return clipboardAppliedRow{}, fmt.Errorf("encode raw capture: %w", err)
	}
	if _, err := tx.Exec(ctx, `
WITH inserted_record AS (
    INSERT INTO records (
        record_id,
        incident_id,
        record_type,
        created_by_user_id,
        created_at,
        updated_by_user_id,
        updated_at,
        row_version
    )
    VALUES ($1, $2, 'timeline_event', $3, $4, $3, $4, 1)
)
INSERT INTO timeline_events (
    record_id,
    incident_id,
    occurred_at,
    summary,
    details,
    source_text,
    raw_capture,
    capture_state,
    row_version,
    recorded_at,
    edited_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $5, $6, $7, $8, $9, 'rough', 1, $4, $4, $3, $3)
`, current.RecordID, incidentID, actor.ID, now.UTC(), current.OccurredAt, current.Summary, current.Details, current.SourceText, rawCaptureJSON); err != nil {
		return clipboardAppliedRow{}, fmt.Errorf("insert clipboard paste timeline row: %w", err)
	}
	mentionProjectionRefresh, err := applyPasteMentionActionsTx(ctx, tx, actor, current.IncidentID, current.RecordID, rowPlan.Cells, originKind, now.UTC())
	if err != nil {
		return clipboardAppliedRow{}, err
	}
	if err := s.rebuildMentionEntityProjectionsTx(ctx, tx, current.IncidentID, mentionProjectionRefresh); err != nil {
		return clipboardAppliedRow{}, err
	}
	tagMutations, err := applyPasteTagActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, rowPlan.Cells, now.UTC())
	if err != nil {
		return clipboardAppliedRow{}, err
	}
	attachedEvidenceMutations, err := applyPasteAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, rowPlan.Cells, now.UTC())
	if err != nil {
		return clipboardAppliedRow{}, err
	}
	projected := projectRecord(current, nil)
	if err := hydrateProjectedCollections(ctx, tx, &projected); err != nil {
		return clipboardAppliedRow{}, err
	}
	afterRow := BuildRow(projected)
	return clipboardAppliedRow{
		Operation:                 "create",
		After:                     projected,
		AfterRow:                  afterRow,
		ChangedFieldKeys:          ComputeChangedFieldKeys(nil, projected),
		TagMutations:              tagMutations,
		AttachedEvidenceMutations: attachedEvidenceMutations,
		RecordID:                  projected.RecordID,
	}, nil
}

func ensureClipboardPasteRecordIncident(current sourceRecord, incidentID uuid.UUID) error {
	if current.IncidentID != incidentID {
		return ErrRecordNotFound
	}
	return nil
}

func (s *Store) applyClipboardPastePatchTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, baseRowVersion int64, requestHash []byte, rowPlan ClipboardPasteRowPlan, originKind string, now time.Time) (clipboardAppliedRow, []map[string]any, error) {
	current, err := loadSourceRecordForIncidentTx(ctx, tx, incidentID, recordID)
	if err != nil {
		return clipboardAppliedRow{}, nil, err
	}
	if err := ensureClipboardPasteRecordIncident(current, incidentID); err != nil {
		return clipboardAppliedRow{}, nil, err
	}
	if current.RowVersion < baseRowVersion {
		return clipboardAppliedRow{}, nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	acceptedCells := append([]clipboardPasteCell{}, rowPlan.Cells...)
	conflicts := make([]map[string]any, 0)
	if current.RowVersion > baseRowVersion && len(rowPlan.Cells) > 0 {
		window, err := loadPatchConflictWindowTx(ctx, tx, recordID, baseRowVersion, current.RowVersion)
		if err != nil {
			return clipboardAppliedRow{}, nil, err
		}
		currentProjected := projectRecord(current, nil)
		if err := hydrateProjectedCollections(ctx, tx, &currentProjected); err != nil {
			return clipboardAppliedRow{}, nil, err
		}
		kept := acceptedCells[:0]
		for _, cell := range acceptedCells {
			if changed, ok := window.ChangedFields[cell.FieldKey]; ok {
				conflict, err := buildSameFieldConflict(recordID, currentProjected, baseRowVersion, requestHash, window, cell.Change, changed)
				if err != nil {
					return clipboardAppliedRow{}, nil, err
				}
				conflicts = append(conflicts, conflict.Conflict)
				continue
			}
			kept = append(kept, cell)
		}
		acceptedCells = kept
	}
	if current.CaptureState == "superseded" {
		return clipboardAppliedRow{}, nil, newIllegalTransitionError("superseded_terminal", current.CaptureState, captureStateEnriched)
	}
	next := current
	for _, cell := range acceptedCells {
		applyPatchChangeToSource(&next, cell.Change)
	}
	next.RawCapture = rawCaptureWithImportColumns(current.RawCapture, rowPlan.Unknown)
	beforeProjected := projectRecord(current, nil)
	if err := hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return clipboardAppliedRow{}, nil, err
	}
	mentionChanged := pasteCellsIncludeField(acceptedCells, "timeline.host_refs") || pasteCellsIncludeField(acceptedCells, "timeline.identity_refs")
	tagChanged := pasteCellsIncludeField(acceptedCells, "timeline.tags")
	evidenceChanged := pasteCellsIncludeField(acceptedCells, "timeline.attached_evidence_ids")
	materialChanged := hasMaterialChange(current, next)
	if mentionChanged {
		mentionProjectionRefresh, err := applyPasteMentionActionsTx(ctx, tx, actor, current.IncidentID, recordID, acceptedCells, originKind, now.UTC())
		if err != nil {
			return clipboardAppliedRow{}, nil, err
		}
		if err := s.rebuildMentionEntityProjectionsTx(ctx, tx, current.IncidentID, mentionProjectionRefresh); err != nil {
			return clipboardAppliedRow{}, nil, err
		}
	}
	var tagMutations []recordTagMutation
	if tagChanged {
		tagMutations, err = applyPasteTagActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, acceptedCells, now.UTC())
		if err != nil {
			return clipboardAppliedRow{}, nil, err
		}
		tagChanged = len(tagMutations) > 0
	}
	var attachedEvidenceMutations []attachedEvidenceMutation
	if evidenceChanged {
		attachedEvidenceMutations, err = applyPasteAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, acceptedCells, now.UTC())
		if err != nil {
			return clipboardAppliedRow{}, nil, err
		}
		evidenceChanged = len(attachedEvidenceMutations) > 0
	}
	if !materialChanged && !mentionChanged && !tagChanged && !evidenceChanged {
		return clipboardAppliedRow{}, conflicts, nil
	}
	nextState, err := CaptureStateAfterMaterialPatch(current.CaptureState)
	if err != nil {
		return clipboardAppliedRow{}, nil, err
	}
	next.CaptureState = nextState
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return clipboardAppliedRow{}, nil, err
	}
	next.EditedAt = now.UTC()
	next.UpdatedByUserID = actor.ID
	if current.CaptureState == captureStateReviewed {
		next.ReviewedAt = nil
		next.ReviewedByUserID = nil
	}
	rawCaptureJSON, err := json.Marshal(next.RawCapture)
	if err != nil {
		return clipboardAppliedRow{}, nil, fmt.Errorf("encode raw capture: %w", err)
	}
	if err := tx.QueryRow(ctx, `
UPDATE timeline_events
   SET occurred_at = $2,
       summary = $3,
       details = $4,
       source_text = $5,
       raw_capture = $6,
       capture_state = $7,
       row_version = $8,
       edited_at = $9,
       updated_by_user_id = $10,
       reviewed_at = $11,
       reviewed_by_user_id = $12
 WHERE record_id = $1
   AND incident_id = $13
RETURNING recorded_at
`, recordID, next.OccurredAt, next.Summary, next.Details, next.SourceText, rawCaptureJSON, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID, incidentID).Scan(&next.RecordedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return clipboardAppliedRow{}, nil, ErrRecordNotFound
		}
		return clipboardAppliedRow{}, nil, fmt.Errorf("update timeline clipboard paste record: %w", err)
	}
	afterProjected := projectRecord(next, nil)
	if err := hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return clipboardAppliedRow{}, nil, err
	}
	beforeRow := BuildRow(beforeProjected)
	afterRow := BuildRow(afterProjected)
	return clipboardAppliedRow{
		Operation:                 "patch",
		Before:                    &beforeProjected,
		After:                     afterProjected,
		BeforeRow:                 beforeRow,
		AfterRow:                  afterRow,
		ChangedFieldKeys:          ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		TagMutations:              tagMutations,
		AttachedEvidenceMutations: attachedEvidenceMutations,
		RecordID:                  afterProjected.RecordID,
	}, conflicts, nil
}

func applyPatchChangeToSource(record *sourceRecord, change PatchChange) {
	switch change.FieldKey {
	case "timeline.occurred_at":
		record.OccurredAt = change.OccurredAt
	case "timeline.summary":
		record.Summary = change.TextValue
	case "timeline.details":
		record.Details = change.TextValue
	case "timeline.source_text":
		record.SourceText = change.TextValue
	}
}

func pasteCellsIncludeField(cells []clipboardPasteCell, fieldKey string) bool {
	return slices.ContainsFunc(cells, func(cell clipboardPasteCell) bool {
		return cell.FieldKey == fieldKey && cell.Change.ActionPayload != nil && len(cell.Change.ActionPayload.Actions) > 0
	})
}

func applyPasteMentionActionsTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, cells []clipboardPasteCell, originKind string, now time.Time) (mentionProjectionRefresh, error) {
	var refresh mentionProjectionRefresh
	for _, cell := range cells {
		if cell.Change.ActionPayload == nil || (cell.FieldKey != "timeline.host_refs" && cell.FieldKey != "timeline.identity_refs") {
			continue
		}
		entityType := "host"
		if cell.FieldKey == "timeline.identity_refs" {
			entityType = "identity"
		}
		linked, err := insertMentionActionsTx(ctx, tx, actor.ID, incidentID, recordID, cell.FieldKey, entityType, cell.Change.ActionPayload, mentionInsertOptions{
			allowInteractiveAutoResolution: true,
			originKind:                     originKind,
		}, now)
		if err != nil {
			return mentionProjectionRefresh{}, err
		}
		if linked {
			refresh.include(cell.FieldKey)
		}
	}
	return refresh, nil
}

func applyPasteTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, cells []clipboardPasteCell, now time.Time) ([]recordTagMutation, error) {
	mutations := make([]recordTagMutation, 0)
	for _, cell := range cells {
		if cell.FieldKey != "timeline.tags" || cell.Change.ActionPayload == nil {
			continue
		}
		applied, err := insertTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, cell.Change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func applyPasteAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, cells []clipboardPasteCell, now time.Time) ([]attachedEvidenceMutation, error) {
	mutations := make([]attachedEvidenceMutation, 0)
	for _, cell := range cells {
		if cell.FieldKey != "timeline.attached_evidence_ids" || cell.Change.ActionPayload == nil {
			continue
		}
		applied, err := applyAttachedEvidenceActionsTx(ctx, tx, actorUserID, incidentID, recordID, cell.Change.ActionPayload, now)
		if err != nil {
			return nil, err
		}
		mutations = append(mutations, applied...)
	}
	return mutations, nil
}

func rawCaptureWithImportColumns(existing map[string]any, additions []ClipboardRawImportColumn) map[string]any {
	next := map[string]any{}
	for key, value := range existing {
		next[key] = value
	}
	if len(additions) == 0 {
		return next
	}
	columns := make([]any, 0)
	if existingColumns, ok := next["import_columns"].([]any); ok {
		columns = append(columns, existingColumns...)
	}
	for _, addition := range additions {
		columns = append(columns, map[string]any{
			"source_kind":           addition.SourceKind,
			"paste_client_txn_id":   addition.PasteClientTxnID,
			"source_row_ordinal":    addition.SourceRowOrdinal,
			"source_column_ordinal": addition.SourceColumnOrdinal,
			"source_header_text":    addition.SourceHeaderText,
			"raw_value":             addition.RawValue,
		})
	}
	next["import_columns"] = columns
	return next
}
