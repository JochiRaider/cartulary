package revisions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	recorddeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
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

func (s *commandStore) SoftDeleteRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.applyDeleteRestore(ctx, actor, recordID, request, requestHash, requestID, now.UTC(), deleteRouteKey, true)
}

func (s *commandStore) RestoreRecord(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time) (DeleteRestoreResult, error) {
	return s.applyDeleteRestore(ctx, actor, recordID, request, requestHash, requestID, now.UTC(), restoreRouteKey, false)
}

func (s *commandStore) applyDeleteRestore(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request DeleteRestoreRequest, requestHash []byte, requestID string, now time.Time, routeKey string, deleting bool) (DeleteRestoreResult, error) {
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
		if err := lockDestructiveOperationRecordsNowaitTx(ctx, tx, []uuid.UUID{recordID}); err != nil {
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
	provider, ok := s.deleteRestoreProviders.Provider(record.RecordType)
	if !ok {
		return DeleteRestoreResult{}, ErrUnsupportedRecordType
	}
	viewSchemaID, err := provider.ViewSchemaID(ctx, tx, record.RecordID)
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
		if err := validateDeletePreconditionsTx(ctx, tx, provider, record); err != nil {
			return DeleteRestoreResult{}, err
		}
	}

	beforeSnapshot, err := provider.SnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	nextRowVersion, err := updateEnvelopeDeleteStateTx(ctx, tx, record.RecordID, actor.ID, now, deleting)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	if err := provider.UpdateSourceDeleteStateTx(ctx, tx, record.RecordID, actor.ID, now, deleting); err != nil {
		return DeleteRestoreResult{}, err
	}
	if err := s.rebuildProjectionsTx(ctx, tx, record.IncidentID); err != nil {
		return DeleteRestoreResult{}, err
	}
	afterSnapshot, err := provider.SnapshotTx(ctx, tx, record.RecordID)
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
	changeSetID, err := s.appender.AppendChangeSetTx(ctx, tx, AppendChangeSetParams{
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
	if err := s.appender.AppendMutationTx(ctx, tx, AppendMutationParams{
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
	if err := s.appender.AppendRecordRevisionTx(ctx, tx, AppendRecordRevisionParams{
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
	if err := s.appendDeleteRestoreRecordChangeIntentTx(
		ctx,
		tx,
		record.IncidentID,
		actor.ID,
		request.ClientTxnID,
		changeSetID,
		record.RecordID,
		nextRowVersion,
		viewSchemaID,
		changeKind,
		now,
	); err != nil {
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

func validateDeletePreconditionsTx(ctx context.Context, tx pgx.Tx, provider recorddeleterestore.SourceProvider, record deleteRestoreRecord) error {
	reasonCode, blocked, err := provider.ValidateDeletePreconditionsTx(ctx, tx, record.IncidentID, record.RecordID)
	if err != nil {
		return err
	}
	if blocked {
		return &RecordDeleteBlockedError{
			RecordID:   record.RecordID,
			ReasonCode: reasonCode,
		}
	}
	return nil
}

func lockDestructiveOperationRecordsNowaitTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if err := records.LockDestructiveOperationRecordsNowaitTx(ctx, tx, recordIDs); err != nil {
		var locked *records.DestructiveOperationRecordLockedError
		if errors.As(err, &locked) {
			return &RecordLockedError{RecordID: locked.RecordID}
		}
		if errors.Is(err, records.ErrRecordEnvelopeNotFound) {
			return ErrRecordNotFound
		}
		return err
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
