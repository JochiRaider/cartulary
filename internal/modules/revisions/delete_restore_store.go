package revisions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var (
	ErrRowVersionConflict      = errors.New("revisions: row version conflict")
	ErrRecordDeletedUseRestore = errors.New("revisions: record deleted use restore")
	ErrRecordAlreadyDeleted    = errors.New("revisions: record already deleted")
	ErrRecordDeleteBlocked     = errors.New("revisions: record delete blocked")
	ErrRecordNotDeleted        = errors.New("revisions: record not deleted")
	ErrRecordLocked            = errors.New("revisions: record locked")
	ErrUnsupportedRecordType   = errors.New("revisions: unsupported record type")
)

const activeIncomingPartyReferenceReason = "active_incoming_party_reference"

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return ErrRowVersionConflict.Error()
}

func (e *RowVersionConflictError) Unwrap() error {
	return ErrRowVersionConflict
}

func (e *RowVersionConflictError) Details() map[string]any {
	return map[string]any{
		"record_id":           e.RecordID.String(),
		"base_row_version":    e.BaseRowVersion,
		"current_row_version": e.CurrentRowVersion,
	}
}

type RecordLockedError struct {
	RecordID uuid.UUID
}

func (e *RecordLockedError) Error() string {
	return ErrRecordLocked.Error()
}

func (e *RecordLockedError) Unwrap() error {
	return ErrRecordLocked
}

type RecordDeleteBlockedError struct {
	RecordID   uuid.UUID
	ReasonCode string
}

func (e *RecordDeleteBlockedError) Error() string {
	return ErrRecordDeleteBlocked.Error()
}

func (e *RecordDeleteBlockedError) Unwrap() error {
	return ErrRecordDeleteBlocked
}

func (e *RecordDeleteBlockedError) Details() map[string]any {
	return map[string]any{
		"record_id":   e.RecordID.String(),
		"reason_code": e.ReasonCode,
	}
}

type DeleteRestoreResult struct {
	Payload      map[string]any
	StatusCode   int
	IncidentID   uuid.UUID
	RecordID     uuid.UUID
	RowVersion   int64
	ChangeSetID  uuid.UUID
	ClientTxnID  string
	ViewSchemaID string
	ChangeKind   string
	Replayed     bool
}

type deleteRestoreRecord struct {
	IncidentID      uuid.UUID
	RecordID        uuid.UUID
	RecordType      string
	RowVersion      int64
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

type recordDeleteRestoreAdapter struct {
	RecordType      string
	SourceTable     string
	SourceRecordCol string
	ViewSchemaID    string
	SourceTombstone bool
}

var deleteRestoreAdapters = map[string]recordDeleteRestoreAdapter{
	"timeline_event": {RecordType: "timeline_event", SourceTable: "timeline_events", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.timeline.v2"},
	"host":           {RecordType: "host", SourceTable: "hosts", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.hosts.v1"},
	"identity":       {RecordType: "identity", SourceTable: "identities", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.identities.v1"},
	"party":          {RecordType: "party", SourceTable: "parties", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.parties.v1"},
	"indicator":      {RecordType: "indicator", SourceTable: "indicators", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.indicators.v1", SourceTombstone: true},
	"artifact":       {RecordType: "artifact", SourceTable: "artifacts", SourceRecordCol: "record_id"},
	"task_request":   {RecordType: "task_request", SourceTable: "task_requests", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.task_requests.v1"},
	"decision":       {RecordType: "decision", SourceTable: "decisions", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.decisions.v1"},
	"evidence":       {RecordType: "evidence", SourceTable: "evidence", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.evidence.v1"},
	"assessment":     {RecordType: "assessment", SourceTable: "assessments", SourceRecordCol: "record_id", ViewSchemaID: "cartulary.view.assessments.v1", SourceTombstone: true},
}

func DeleteRestoreAdapterTypes() []string {
	types := make([]string, 0, len(deleteRestoreAdapters))
	for recordType := range deleteRestoreAdapters {
		types = append(types, recordType)
	}
	return types
}

func (s *Store) SoftDeleteRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.applyDeleteRestore(ctx, actor, recordID, request, requestHash, requestID, now.UTC(), deleteRouteKey, true)
}

func (s *Store) RestoreRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.applyDeleteRestore(ctx, actor, recordID, request, requestHash, requestID, now.UTC(), restoreRouteKey, false)
}

func (s *Store) applyDeleteRestore(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time, routeKey string, deleting bool) (DeleteRestoreResult, error) {
	authStore := authn.NewStore(s.db)
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return DeleteRestoreResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredDeleteRestorePayload(existing.ResponseJSON)
		if err != nil {
			return DeleteRestoreResult{}, err
		}
		result := deleteRestoreResultFromPayload(payload)
		result.StatusCode = http.StatusOK
		result.ClientTxnID = request.ClientTxnID
		result.Replayed = true
		return result, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return DeleteRestoreResult{}, fmt.Errorf("query delete restore idempotency: %w", err)
	}

	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DeleteRestoreResult{}, fmt.Errorf("begin delete restore transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if !deleting {
		if err := LockRecordEnvelopesNowaitTx(ctx, tx, []uuid.UUID{recordID}); err != nil {
			return DeleteRestoreResult{}, err
		}
	}

	record, err := loadDeleteRestoreRecordTx(ctx, tx, recordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	if err := s.incidentAccess.EnsureOpenTx(ctx, tx, record.IncidentID); err != nil {
		return DeleteRestoreResult{}, err
	}
	adapter, ok := deleteRestoreAdapters[record.RecordType]
	if !ok {
		return DeleteRestoreResult{}, ErrUnsupportedRecordType
	}
	viewSchemaID, err := adapter.viewSchemaID(ctx, tx, record.RecordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	if record.RowVersion != request.BaseRowVersion {
		return DeleteRestoreResult{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: record.RowVersion}
	}
	if deleting && record.DeletedAt != nil {
		return DeleteRestoreResult{}, ErrRecordAlreadyDeleted
	}
	if !deleting && record.DeletedAt == nil {
		return DeleteRestoreResult{}, ErrRecordNotDeleted
	}
	if deleting {
		if err := validateDeletePreconditionsTx(ctx, tx, record); err != nil {
			return DeleteRestoreResult{}, err
		}
	}

	beforeSnapshot, err := adapter.snapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	nextRowVersion, err := updateEnvelopeDeleteStateTx(ctx, tx, record.RecordID, actor.ID, now, deleting)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	if adapter.SourceTombstone {
		if err := adapter.updateSourceDeleteStateTx(ctx, tx, record.RecordID, actor.ID, now, deleting); err != nil {
			return DeleteRestoreResult{}, err
		}
	}
	if err := rebuildDeleteRestoreProjectionsTx(ctx, tx, record.IncidentID); err != nil {
		return DeleteRestoreResult{}, err
	}
	afterSnapshot, err := adapter.snapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}

	source := deleteRouteKey
	operation := "soft_delete"
	deleted := true
	changeKind := "remove"
	if !deleting {
		source = restoreRouteKey
		operation = "restore"
		deleted = false
		changeKind = "invalidate"
	}
	changeSetID, err := s.InsertChangeSetTx(ctx, tx, ChangeSetParams{
		IncidentID:  record.IncidentID,
		ActorUserID: actor.ID,
		Source:      source,
		Reason:      request.Reason,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now,
	})
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	beforeVersionID := fmt.Sprintf("record:%s:%d", record.RecordID, record.RowVersion)
	afterVersionID := fmt.Sprintf("record:%s:%d", record.RecordID, nextRowVersion)
	if err := s.InsertMutationTx(ctx, tx, MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        record.RecordID.String(),
		OperationKind:   operation,
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeSnapshot,
		AfterValue:      afterSnapshot,
	}); err != nil {
		return DeleteRestoreResult{}, err
	}
	if err := s.InsertRecordRevisionTx(ctx, tx, RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    record.RecordID,
		RowVersion:  nextRowVersion,
		BeforeValue: beforeSnapshot,
		AfterValue:  afterSnapshot,
	}); err != nil {
		return DeleteRestoreResult{}, err
	}
	current, err := loadDeleteRestoreRecordTx(ctx, tx, recordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	payload := buildDeleteRestorePayload(current, changeSetID, deleted)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return DeleteRestoreResult{}, authn.ErrClientTxnConflict
		}
		return DeleteRestoreResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DeleteRestoreResult{}, fmt.Errorf("commit delete restore transaction: %w", err)
	}
	return DeleteRestoreResult{
		Payload:      payload,
		StatusCode:   http.StatusOK,
		IncidentID:   record.IncidentID,
		RecordID:     record.RecordID,
		RowVersion:   nextRowVersion,
		ChangeSetID:  changeSetID,
		ClientTxnID:  request.ClientTxnID,
		ViewSchemaID: viewSchemaID,
		ChangeKind:   changeKind,
	}, nil
}

func loadDeleteRestoreRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (deleteRestoreRecord, error) {
	var (
		record       deleteRestoreRecord
		deletedAt    sql.NullTime
		deletedByRaw sql.NullString
	)
	if err := tx.QueryRow(ctx, `
SELECT incident_id, record_id, record_type, row_version, deleted_at, deleted_by_user_id::text
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&record.IncidentID, &record.RecordID, &record.RecordType, &record.RowVersion, &deletedAt, &deletedByRaw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return deleteRestoreRecord{}, ErrRecordNotFound
		}
		return deleteRestoreRecord{}, err
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	if deletedByRaw.Valid {
		parsed, err := uuid.Parse(deletedByRaw.String)
		if err != nil {
			return deleteRestoreRecord{}, err
		}
		record.DeletedByUserID = &parsed
	}
	return record, nil
}

func validateDeletePreconditionsTx(ctx context.Context, tx pgx.Tx, record deleteRestoreRecord) error {
	switch record.RecordType {
	case "party":
		hasIncoming, err := hasActiveIncomingPartyReferenceTx(ctx, tx, record.IncidentID, record.RecordID)
		if err != nil {
			return err
		}
		if hasIncoming {
			return &RecordDeleteBlockedError{
				RecordID:   record.RecordID,
				ReasonCode: activeIncomingPartyReferenceReason,
			}
		}
	}
	return nil
}

func hasActiveIncomingPartyReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, partyID uuid.UUID) (bool, error) {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM evidence e
      JOIN records r
        ON r.incident_id = e.incident_id
       AND r.record_id = e.record_id
       AND r.deleted_at IS NULL
     WHERE e.incident_id = $1
       AND (e.collector_party_id = $2 OR e.source_party_id = $2)
    UNION ALL
    SELECT 1
      FROM task_requests t
      JOIN records r
        ON r.incident_id = t.incident_id
       AND r.record_id = t.record_id
       AND r.deleted_at IS NULL
     WHERE t.incident_id = $1
       AND t.requester_party_id = $2
    UNION ALL
    SELECT 1
      FROM active_record_links_v1 rl
      JOIN records src
        ON src.incident_id = rl.incident_id
       AND src.record_id = rl.src_record_id
       AND src.deleted_at IS NULL
     WHERE rl.incident_id = $1
       AND rl.dst_record_id = $2
       AND rl.link_type = 'references_record'
       AND rl.field_key IN ('comm_log.audience_party_ids', 'comm_log.attendee_party_ids')
       AND rl.deleted_at IS NULL
)
`, incidentID, partyID).Scan(&exists); err != nil {
		return false, fmt.Errorf("validate party delete references: %w", err)
	}
	return exists, nil
}

func LockRecordEnvelopesNowaitTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	ordered := append([]uuid.UUID(nil), recordIDs...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].String() < ordered[j].String()
	})
	for i := 0; i < len(ordered); i++ {
		if i > 0 && ordered[i] == ordered[i-1] {
			continue
		}
		var locked uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE NOWAIT`, ordered[i]).Scan(&locked); err != nil {
			if isLockUnavailable(err) {
				return &RecordLockedError{RecordID: ordered[i]}
			}
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrRecordNotFound
			}
			return err
		}
	}
	return nil
}

func updateEnvelopeDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, deleting bool) (int64, error) {
	var rowVersion int64
	if deleting {
		if err := tx.QueryRow(ctx, `
UPDATE records
   SET row_version = row_version + 1,
       deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2,
       updated_by_user_id = $3
 WHERE record_id = $1
RETURNING row_version
`, recordID, now.UTC(), actorUserID).Scan(&rowVersion); err != nil {
			return 0, err
		}
		return rowVersion, nil
	}
	if err := tx.QueryRow(ctx, `
UPDATE records
   SET row_version = row_version + 1,
       deleted_at = NULL,
       deleted_by_user_id = NULL,
       updated_at = $2,
       updated_by_user_id = $3
 WHERE record_id = $1
RETURNING row_version
`, recordID, now.UTC(), actorUserID).Scan(&rowVersion); err != nil {
		return 0, err
	}
	return rowVersion, nil
}

func (a recordDeleteRestoreAdapter) snapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	query := fmt.Sprintf(`
SELECT jsonb_build_object('record', to_jsonb(r), 'source', to_jsonb(s))
  FROM records r
  JOIN %s s
    ON s.%s = r.record_id
 WHERE r.record_id = $1
`, a.SourceTable, a.SourceRecordCol)
	var raw []byte
	if err := tx.QueryRow(ctx, query, recordID).Scan(&raw); err != nil {
		return nil, err
	}
	var snapshot map[string]any
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (a recordDeleteRestoreAdapter) updateSourceDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, deleting bool) error {
	if deleting {
		_, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s
   SET deleted_at = $2,
       deleted_by_user_id = $3,
       updated_at = $2
 WHERE %s = $1
`, a.SourceTable, a.SourceRecordCol), recordID, now.UTC(), actorUserID)
		return err
	}
	_, err := tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s
   SET deleted_at = NULL,
       deleted_by_user_id = NULL,
       updated_at = $2
 WHERE %s = $1
`, a.SourceTable, a.SourceRecordCol), recordID, now.UTC())
	return err
}

func (a recordDeleteRestoreAdapter) viewSchemaID(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
	if a.RecordType != "artifact" {
		return a.ViewSchemaID, nil
	}
	var artifactType string
	if err := tx.QueryRow(ctx, `SELECT artifact_type FROM artifacts WHERE record_id = $1`, recordID).Scan(&artifactType); err != nil {
		return "", err
	}
	variant, ok := viewschema.LookupArtifactVariantByArtifactType(artifactType)
	if !ok {
		switch artifactType {
		case "investigative_query":
			return "cartulary.view.investigative_queries.v1", nil
		case "forensic_keyword":
			return "cartulary.view.forensic_keywords.v1", nil
		default:
			return "", fmt.Errorf("unsupported artifact type %q", artifactType)
		}
	}
	return variant.PublicSurfaceRef, nil
}

func rebuildDeleteRestoreProjectionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return projectionadapters.NewRowProjector(nil).RebuildIncidentTx(ctx, tx, incidentID)
}

func buildDeleteRestorePayload(record deleteRestoreRecord, changeSetID uuid.UUID, deleted bool) map[string]any {
	return map[string]any{
		"record_id":          record.RecordID.String(),
		"incident_id":        record.IncidentID.String(),
		"row_version":        record.RowVersion,
		"deleted":            deleted,
		"deleted_at":         formatTimePointer(record.DeletedAt),
		"deleted_by_user_id": formatUUIDPointer(record.DeletedByUserID),
		"change_set_id":      changeSetID.String(),
	}
}

func decodeStoredDeleteRestorePayload(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode delete restore idempotency payload: %w", err)
	}
	return payload, nil
}

func deleteRestoreResultFromPayload(payload map[string]any) DeleteRestoreResult {
	result := DeleteRestoreResult{Payload: payload}
	if raw, ok := payload["incident_id"].(string); ok {
		result.IncidentID, _ = uuid.Parse(raw)
	}
	if raw, ok := payload["record_id"].(string); ok {
		result.RecordID, _ = uuid.Parse(raw)
	}
	if raw, ok := payload["change_set_id"].(string); ok {
		result.ChangeSetID, _ = uuid.Parse(raw)
	}
	switch raw := payload["row_version"].(type) {
	case float64:
		result.RowVersion = int64(raw)
	case int64:
		result.RowVersion = raw
	}
	return result
}

func formatTimePointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func isLockUnavailable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "55P03"
}
