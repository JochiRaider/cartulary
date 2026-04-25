package workbook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (s *Store) CreateWorkbookRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + request.ViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: incidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook create idempotency: %w", err)
	}
	if err := validateCreateRequest(request); err != nil {
		return MutationResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := validateCreateReferencesTx(ctx, tx, incidentID, request); err != nil {
		return MutationResult{}, err
	}

	recordType := "artifact"
	if request.ViewSchemaID == PartiesViewSchemaID {
		recordType = "party"
	}
	recordID, err := s.recordStore.InsertTx(ctx, tx, recordsInsertParams(incidentID, recordType, actor.ID, now.UTC()))
	if err != nil {
		return MutationResult{}, err
	}
	if request.ViewSchemaID == PartiesViewSchemaID {
		if err := insertPartyTx(ctx, tx, recordID, incidentID, request, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	} else if err := insertArtifactTx(ctx, tx, recordID, incidentID, actor.ID, request, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := applyCollectionPayloadsTx(ctx, tx, incidentID, recordID, actor.ID, request.Collections, now.UTC()); err != nil {
		return MutationResult{}, err
	}

	row, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      workbookCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(request.ViewSchemaID, changeSetID, row)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook create transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (s *Store) PatchWorkbookRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookPatchRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook patch payload: %w", err)
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook patch idempotency: %w", err)
	}
	if len(request.Changes) == 0 {
		return MutationResult{}, mutationValidationError("changes", "empty_changes")
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadRecordMetaForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if !recordTypeMatchesView(meta.RecordType, request.ViewSchemaID) {
		return MutationResult{}, pgx.ErrNoRows
	}
	if meta.RowVersion != request.BaseRowVersion {
		if collectionChange := firstCollectionChange(request.Changes); collectionChange != nil {
			current, loadErr := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
			if loadErr == nil {
				return MutationResult{}, buildCollectionSameFieldConflict(recordID, request.BaseRowVersion, meta.RowVersion, *collectionChange, current)
			}
		}
		return MutationResult{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
	}
	beforeRow, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := validatePatchReferencesTx(ctx, tx, meta.IncidentID, request); err != nil {
		return MutationResult{}, err
	}
	changed, err := applyPatchTx(ctx, tx, meta.IncidentID, recordID, actor.ID, request, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if !changed {
		return MutationResult{}, mutationValidationError("changes", "no_effective_change")
	}
	rowVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := touchSourceRowTx(ctx, tx, request.ViewSchemaID, recordID, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	afterRow, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: actor.ID,
		Source:      workbookPatchRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	beforeVersionID := workbookVersionID(recordID, request.BaseRowVersion)
	afterVersionID := workbookVersionID(recordID, rowVersion)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        recordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook patch transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       meta.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func (s *Store) RecordIncident(ctx context.Context, recordID uuid.UUID, viewSchemaID string) (uuid.UUID, error) {
	var incidentID uuid.UUID
	switch {
	case viewSchemaID == PartiesViewSchemaID:
		err := s.pool.QueryRow(ctx, `
SELECT r.incident_id
  FROM records r
  JOIN parties p
    ON p.incident_id = r.incident_id
   AND p.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'party'
   AND r.deleted_at IS NULL
`, recordID).Scan(&incidentID)
		return incidentID, err
	case isWorkbookMutationSurface(viewSchemaID):
		err := s.pool.QueryRow(ctx, `
SELECT r.incident_id
  FROM records r
  JOIN artifacts a
    ON a.incident_id = r.incident_id
   AND a.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type = $2
   AND r.deleted_at IS NULL
`, recordID, artifactTypeForView(viewSchemaID)).Scan(&incidentID)
		return incidentID, err
	case viewSchemaID == "cartulary.view.timeline.v1":
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type = 'timeline_event'
   AND deleted_at IS NULL
`, recordID).Scan(&incidentID)
		return incidentID, err
	default:
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND deleted_at IS NULL
`, recordID).Scan(&incidentID)
		return incidentID, err
	}
}

type recordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func recordsInsertParams(incidentID uuid.UUID, recordType string, actorID uuid.UUID, now time.Time) records.InsertParams {
	return records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      recordType,
		CreatedByUserID: actorID,
		CreatedAt:       now,
		UpdatedByUserID: actorID,
		UpdatedAt:       now,
		RowVersion:      1,
	}
}

func loadRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (recordMeta, error) {
	var meta recordMeta
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version
  FROM records
 WHERE record_id = $1
   AND deleted_at IS NULL
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion)
	if err != nil {
		return recordMeta{}, err
	}
	return meta, nil
}

func (s *Store) loadGenericRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	definition, ok := genericSurfaces[viewSchemaID]
	if !ok {
		return nil, fmt.Errorf("workbook mutation surface %q not mapped", viewSchemaID)
	}
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(", r.row_version")
	for _, field := range definition.fields {
		builder.WriteString(", ")
		builder.WriteString(field.expr)
	}
	builder.WriteString(" ")
	builder.WriteString(definition.fromSQL)
	builder.WriteString(" WHERE ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(" = $1 AND r.deleted_at IS NULL")
	if definition.whereSQL != "" {
		builder.WriteString(" AND ")
		builder.WriteString(definition.whereSQL)
	}
	row := tx.QueryRow(ctx, builder.String(), recordID)
	values := make([]any, len(definition.fields)+2)
	scanTargets := make([]any, len(values))
	for index := range values {
		scanTargets[index] = &values[index]
	}
	if err := row.Scan(scanTargets...); err != nil {
		return nil, err
	}
	return buildGenericRow(definition, nil, values)
}

func validateCreateRequest(request CreateRequest) error {
	switch request.ViewSchemaID {
	case PartiesViewSchemaID:
		if !hasTextValue(request.Values, "party.display_name") {
			return mutationValidationError("party.display_name", "missing_required_field")
		}
		if !validValueText(request.Values, "party.party_kind", validPartyKind) {
			return mutationValidationError("party.party_kind", "missing_required_field")
		}
	case CommLogViewSchemaID:
		for _, field := range []string{"comm_log.comm_type", "comm_log.audience", "comm_log.channel_or_meeting", "comm_log.summary"} {
			if !hasTextValue(request.Values, field) {
				return mutationValidationError(field, "missing_required_field")
			}
		}
		if !validValueText(request.Values, "comm_log.comm_type", validCommType) {
			return mutationValidationError("comm_log.comm_type", "invalid_value")
		}
	case HandoffViewSchemaID:
		if !hasUUIDValue(request.Values, "handoff.incoming_owner_user_id") {
			return mutationValidationError("handoff.incoming_owner_user_id", "missing_required_field")
		}
		if !hasTextValue(request.Values, "handoff.current_state_summary") {
			return mutationValidationError("handoff.current_state_summary", "missing_required_field")
		}
	case StatusReviewViewSchemaID:
		if !hasTextValue(request.Values, "status_review.current_state_summary") {
			return mutationValidationError("status_review.current_state_summary", "missing_required_field")
		}
	case LessonViewSchemaID:
		if !hasTextValue(request.Values, "lesson.summary") {
			return mutationValidationError("lesson.summary", "missing_required_field")
		}
		if value, ok := request.Values["lesson.closure_state"]; ok && !validClosureState(derefText(value.Text)) {
			return mutationValidationError("lesson.closure_state", "invalid_value")
		}
	}
	return nil
}

func insertPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, request CreateRequest, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, organization_name, role_title,
    primary_email, timezone_name, external_ref, notes, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
`, recordID, incidentID,
		textValue(request.Values, "party.display_name"),
		textValue(request.Values, "party.party_kind"),
		nullableTextValue(request.Values, "party.organization_name"),
		nullableTextValue(request.Values, "party.role_title"),
		nullableTextValue(request.Values, "party.primary_email"),
		nullableTextValue(request.Values, "party.timezone_name"),
		nullableTextValue(request.Values, "party.external_ref"),
		nullableTextValue(request.Values, "party.notes"),
		now)
	if err != nil {
		return fmt.Errorf("insert party: %w", err)
	}
	return nil
}

func insertArtifactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, request CreateRequest, now time.Time) error {
	artifactType := artifactTypeForView(request.ViewSchemaID)
	timestamp := now
	if value, ok := request.Values[artifactType+".timestamp_utc"]; ok && value.Timestamp != nil {
		timestamp = value.Timestamp.UTC()
	}
	commID, handoffID, statusReviewID, lessonID := any(nil), any(nil), any(nil), any(nil)
	switch request.ViewSchemaID {
	case CommLogViewSchemaID:
		commID = uuid.NewString()
	case HandoffViewSchemaID:
		handoffID = uuid.NewString()
	case StatusReviewViewSchemaID:
		statusReviewID = uuid.NewString()
	case LessonViewSchemaID:
		lessonID = uuid.NewString()
	}
	outgoingOwner := nullableUUIDValue(request.Values, "handoff.outgoing_owner_user_id")
	if request.ViewSchemaID == HandoffViewSchemaID && outgoingOwner == nil {
		outgoingOwner = actorID
	}
	reviewOwner := nullableUUIDValue(request.Values, "status_review.review_owner_user_id")
	if request.ViewSchemaID == StatusReviewViewSchemaID && reviewOwner == nil {
		reviewOwner = actorID
	}
	lessonOwner := nullableUUIDValue(request.Values, "lesson.owner_user_id")
	if request.ViewSchemaID == LessonViewSchemaID && lessonOwner == nil {
		lessonOwner = actorID
	}
	closureState := nullableTextValue(request.Values, "lesson.closure_state")
	if request.ViewSchemaID == LessonViewSchemaID && closureState == nil {
		closureState = "open"
	}
	_, err := tx.Exec(ctx, `
INSERT INTO artifacts (
    record_id, incident_id, artifact_type, timestamp_utc, updated_at, created_at,
    comm_id, comm_type, audience, channel_or_meeting, summary, next_report_at, privilege_tag,
    handoff_id, outgoing_owner_user_id, incoming_owner_user_id, current_state_summary, next_checks, acknowledged_at,
    status_review_id, review_owner_user_id, active_risks_summary,
    lesson_id, owner_user_id, closure_state, created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $5,
    $6, $7, $8, $9, $10, $11, $12,
    $13, $14, $15, $16, $17, $18,
    $19, $20, $21,
    $22, $23, $24, $25
)
`, recordID, incidentID, artifactType, timestamp, now,
		commID, nullableTextValue(request.Values, "comm_log.comm_type"), nullableTextValue(request.Values, "comm_log.audience"), nullableTextValue(request.Values, "comm_log.channel_or_meeting"), nullableTextValue(request.Values, "comm_log.summary"), nullableTimestampValue(request.Values, "comm_log.next_report_at"), nullableTextValue(request.Values, "comm_log.privilege_tag"),
		handoffID, outgoingOwner, nullableUUIDValue(request.Values, "handoff.incoming_owner_user_id"), nullableTextValue(request.Values, "handoff.current_state_summary"), nullableTextValue(request.Values, "handoff.next_checks"), nullableTimestampValue(request.Values, "handoff.acknowledged_at"),
		statusReviewID, reviewOwner, nullableTextValue(request.Values, "status_review.active_risks_summary"),
		lessonID, lessonOwner, closureState, actorID)
	if err != nil {
		return fmt.Errorf("insert artifact: %w", err)
	}
	return nil
}

func validateCreateReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request CreateRequest) error {
	for fieldKey, value := range request.Values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateActiveUserTx(ctx, tx, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range request.Collections {
		if err := validateCollectionPayloadTx(ctx, tx, incidentID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}

func validatePatchReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request PatchRequest) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && strings.HasSuffix(change.FieldKey, "_user_id") {
			if err := validateActiveUserTx(ctx, tx, *change.Value.UUID, change.FieldKey); err != nil {
				return err
			}
		}
		if change.Collection != nil {
			if err := validateCollectionPayloadTx(ctx, tx, incidentID, change.FieldKey, *change.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, payload CollectionActionPayload) error {
	for _, action := range payload.Actions {
		switch {
		case action.LinkedRecordID != nil:
			if err := validateTargetRecordTx(ctx, tx, incidentID, *action.LinkedRecordID, expectedTargetType(fieldKey), fieldKey); err != nil {
				return err
			}
		case action.PartyID != nil:
			if err := validateTargetRecordTx(ctx, tx, incidentID, *action.PartyID, "party", fieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("validate user: %w", err)
	}
	if !exists {
		return mutationValidationError(field, "invalid_value")
	}
	return nil
}

func validateTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate collection target: %w", err)
	}
	if !exists {
		return mutationValidationError(field, "invalid_value")
	}
	return nil
}

func applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request PatchRequest, now time.Time) (bool, error) {
	changed := false
	for _, change := range request.Changes {
		if change.Value != nil {
			applied, err := applyDirectChangeTx(ctx, tx, recordID, request.ViewSchemaID, change, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
			continue
		}
		if change.Collection != nil {
			applied, err := applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		}
	}
	return changed, nil
}

func applyDirectChangeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, viewSchemaID string, change PatchChange, now time.Time) (bool, error) {
	table, column := tableColumnForField(change.FieldKey)
	if table == "" || column == "" {
		return false, mutationValidationError(change.FieldKey, "unsupported_field_key")
	}
	value := directDBValue(*change.Value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, table, column, column), recordID, value, now)
	if err != nil {
		return false, fmt.Errorf("apply direct change: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func applyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]CollectionActionPayload, now time.Time) error {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if _, err := applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, key, collections[key], now); err != nil {
			return err
		}
	}
	return nil
}

func applyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload CollectionActionPayload, now time.Time) (bool, error) {
	changed := false
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			applied, err := upsertReferenceLinkTx(ctx, tx, incidentID, recordID, *action.LinkedRecordID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_record_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "record_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := tombstoneReferenceLinkTx(ctx, tx, incidentID, recordID, dst, fieldKey, expectedTargetType(fieldKey), actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "add_party_ref":
			applied, err := upsertReferenceLinkTx(ctx, tx, incidentID, recordID, *action.PartyID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_party_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "party_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := tombstoneReferenceLinkTx(ctx, tx, incidentID, recordID, dst, fieldKey, "party", actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "add_risk_ref":
			applied, err := upsertRiskRefTx(ctx, tx, incidentID, recordID, action.RiskRefText, action.NormalizedText, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_risk_ref":
			riskRefID, err := uuidFromItemRef(action.ItemRef, "risk_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := tombstoneRiskRefTx(ctx, tx, incidentID, recordID, riskRefID, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		default:
			return false, mutationValidationError(fieldKey, "invalid_value")
		}
	}
	return changed, nil
}

func upsertReferenceLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'references_record', $4, 'manual', NULL, $5, $5, $6, $6)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type, field_key)
WHERE deleted_at IS NULL AND field_key IS NOT NULL
DO NOTHING
`, incidentID, src, dst, fieldKey, actorID, now)
	if err != nil {
		return false, fmt.Errorf("upsert reference link: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func tombstoneReferenceLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, expectedTargetType string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $6,
       deleted_by_user_id = $5
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'references_record'
   AND field_key = $4
   AND deleted_at IS NULL
   AND EXISTS (
       SELECT 1
         FROM records dst
        WHERE dst.incident_id = record_links.incident_id
          AND dst.record_id = record_links.dst_record_id
          AND dst.record_type = $7
          AND dst.deleted_at IS NULL
   )
`, incidentID, src, dst, fieldKey, actorID, now, expectedTargetType)
	if err != nil {
		return false, fmt.Errorf("remove reference link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, mutationValidationError(fieldKey, "invalid_value")
	}
	return true, nil
}

func upsertRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, text string, normalized string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO handoff_risk_refs (
    incident_id, handoff_record_id, risk_ref_text, normalized_risk_ref_text,
    created_by_user_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (handoff_record_id, normalized_risk_ref_text)
WHERE deleted_at IS NULL
DO NOTHING
`, incidentID, recordID, text, normalized, actorID, now)
	if err != nil {
		return false, fmt.Errorf("upsert risk ref: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func tombstoneRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, riskRefID uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE handoff_risk_refs
   SET deleted_at = $5,
       deleted_by_user_id = $4
 WHERE incident_id = $1
   AND handoff_record_id = $2
   AND risk_ref_id = $3
   AND deleted_at IS NULL
`, incidentID, recordID, riskRefID, actorID, now)
	if err != nil {
		return false, fmt.Errorf("remove risk ref: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, mutationValidationError("handoff.open_risk_refs", "invalid_value")
	}
	return true, nil
}

func touchSourceRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	table := "artifacts"
	if viewSchemaID == PartiesViewSchemaID {
		table = "parties"
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET updated_at = $2 WHERE record_id = $1`, table), recordID, now); err != nil {
		return fmt.Errorf("touch source row: %w", err)
	}
	return nil
}

func buildCollectionSameFieldConflict(recordID uuid.UUID, base int64, current int64, change PatchChange, currentRow map[string]any) *SameFieldConflictError {
	return &SameFieldConflictError{Conflict: map[string]any{
		"record_id":           recordID.String(),
		"field_key":           change.FieldKey,
		"base_row_version":    base,
		"current_row_version": current,
		"current_field_value": cellValue(currentRow, change.FieldKey),
		"conflict_resolution": "collection_review",
	}}
}

func firstCollectionChange(changes []PatchChange) *PatchChange {
	for index := range changes {
		if changes[index].Collection != nil {
			return &changes[index]
		}
	}
	return nil
}

func changedFieldKeys(before map[string]any, after map[string]any) []string {
	afterCells, _ := after["cells"].(map[string]any)
	beforeCells := map[string]any{}
	if before != nil {
		beforeCells, _ = before["cells"].(map[string]any)
	}
	keys := make([]string, 0)
	for fieldKey, afterValue := range afterCells {
		if beforeValue, ok := beforeCells[fieldKey]; !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			keys = append(keys, fieldKey)
		}
	}
	slices.Sort(keys)
	return keys
}

func cellValue(row map[string]any, fieldKey string) any {
	cells, _ := row["cells"].(map[string]any)
	cell, _ := cells[fieldKey].(map[string]any)
	return cell["value"]
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractPayloadUUID(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	return uuid.Parse(text)
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func recordTypeMatchesView(recordType string, viewSchemaID string) bool {
	if viewSchemaID == PartiesViewSchemaID {
		return recordType == "party"
	}
	return recordType == "artifact"
}

func artifactTypeForView(viewSchemaID string) string {
	switch viewSchemaID {
	case CommLogViewSchemaID:
		return "comm_log"
	case HandoffViewSchemaID:
		return "handoff"
	case StatusReviewViewSchemaID:
		return "status_review"
	case LessonViewSchemaID:
		return "lesson"
	default:
		return ""
	}
}

func tableColumnForField(fieldKey string) (string, string) {
	switch {
	case strings.HasPrefix(fieldKey, "party."):
		return "parties", strings.TrimPrefix(fieldKey, "party.")
	case strings.HasPrefix(fieldKey, "comm_log."):
		return "artifacts", strings.TrimPrefix(fieldKey, "comm_log.")
	case strings.HasPrefix(fieldKey, "handoff."):
		return "artifacts", strings.TrimPrefix(fieldKey, "handoff.")
	case strings.HasPrefix(fieldKey, "status_review."):
		return "artifacts", strings.TrimPrefix(fieldKey, "status_review.")
	case strings.HasPrefix(fieldKey, "lesson."):
		return "artifacts", strings.TrimPrefix(fieldKey, "lesson.")
	default:
		return "", ""
	}
}

func directDBValue(value ValueChange) any {
	switch value.Kind {
	case "timestamp":
		if value.Timestamp == nil {
			return nil
		}
		return value.Timestamp.UTC()
	case "uuid":
		if value.UUID == nil {
			return nil
		}
		return *value.UUID
	case "text":
		if value.Text == nil {
			return nil
		}
		return *value.Text
	default:
		return nil
	}
}

func uuidFromItemRef(itemRef string, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	return uuid.Parse(strings.TrimPrefix(itemRef, prefix))
}

func hasTextValue(values map[string]ValueChange, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && *value.Text != ""
}

func hasUUIDValue(values map[string]ValueChange, field string) bool {
	value, ok := values[field]
	return ok && value.UUID != nil
}

func validValueText(values map[string]ValueChange, field string, valid func(string) bool) bool {
	value, ok := values[field]
	return ok && value.Text != nil && valid(*value.Text)
}

func textValue(values map[string]ValueChange, field string) string {
	value := values[field]
	return derefText(value.Text)
}

func nullableTextValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.Text == nil {
		return nil
	}
	return *value.Text
}

func nullableUUIDValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.UUID == nil {
		return nil
	}
	return *value.UUID
}

func nullableTimestampValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.Timestamp == nil {
		return nil
	}
	return value.Timestamp.UTC()
}

func derefText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func expectedTargetType(fieldKey string) string {
	switch fieldKey {
	case "comm_log.decision_ids", "handoff.open_decision_ids", "status_review.open_decision_ids":
		return "decision"
	case "comm_log.action_task_ids", "handoff.open_task_ids", "status_review.blocked_task_ids", "lesson.follow_up_task_ids":
		return "task_request"
	case "status_review.pending_evidence_ids", "lesson.evidence_refs":
		return "evidence"
	default:
		return ""
	}
}

func validPartyKind(value string) bool {
	switch value {
	case "person", "team", "organization", "distribution_list", "other":
		return true
	default:
		return false
	}
}

func validCommType(value string) bool {
	switch value {
	case "meeting", "notification", "approval", "briefing", "handoff":
		return true
	default:
		return false
	}
}

func validClosureState(value string) bool {
	return value == "open" || value == "closed"
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
