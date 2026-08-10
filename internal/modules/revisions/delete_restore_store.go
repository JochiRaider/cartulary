package revisions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
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
	IncidentID   uuid.UUID
	RecordID     uuid.UUID
	RowVersion   int64
	ChangeSetID  uuid.UUID
	ClientTxnID  string
	ViewSchemaID string
	ChangeKind   string
	Replayed     bool
}

type deleteRestoreRecord = RecordEnvelope

func (s *commandStore) SoftDeleteRecord(ctx context.Context, command DeleteRestoreCommand) (DeleteRestoreResult, error) {
	return s.applyDeleteRestore(ctx, command, deleteRouteKey, true)
}

func (s *commandStore) RestoreRecord(ctx context.Context, command DeleteRestoreCommand) (DeleteRestoreResult, error) {
	return s.applyDeleteRestore(ctx, command, restoreRouteKey, false)
}

func (s *commandStore) applyDeleteRestore(ctx context.Context, command DeleteRestoreCommand, routeKey string, deleting bool) (DeleteRestoreResult, error) {
	actorID := command.Actor.UUID()
	recordID := command.RecordID
	request := command.Request
	requestHash := command.RequestHash
	requestID := command.RequestID
	now := command.effectiveAt
	idempotencyKey := IdempotencyKey{
		RouteKey:    routeKey,
		ActorID:     command.Actor,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotency.Get(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return DeleteRestoreResult{}, ErrClientTxnConflict
		}
		payload, err := decodeStoredDeleteRestorePayload(existing.ResponseJSON)
		if err != nil {
			return DeleteRestoreResult{}, err
		}
		result := deleteRestoreResultFromPayload(payload)
		result.ClientTxnID = request.ClientTxnID
		result.Replayed = true
		return result, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return DeleteRestoreResult{}, fmt.Errorf("query delete restore idempotency: %w", err)
	}

	tx, err := s.transactions.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return DeleteRestoreResult{}, fmt.Errorf("begin delete restore transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if !deleting {
		if err := s.lockDestructiveOperationRecordsNowaitTx(ctx, tx, []uuid.UUID{recordID}); err != nil {
			return DeleteRestoreResult{}, err
		}
	}

	record, err := s.loadDeleteRestoreRecordTx(ctx, tx, recordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	commandKind := CommandSoftDelete
	if !deleting {
		commandKind = CommandRestore
	}
	if err := s.authorization.AuthorizeCommandTx(ctx, tx, record.IncidentID, command.Actor, commandKind); err != nil {
		return DeleteRestoreResult{}, err
	}
	sourceAdapter, ok := s.deleteRestoreSources.Source(record.RecordType)
	if !ok {
		return DeleteRestoreResult{}, ErrUnsupportedRecordType
	}
	viewSchemaID, err := sourceAdapter.ViewSchemaID(ctx, tx, record.RecordID)
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
		if err := validateDeletePreconditionsTx(ctx, tx, sourceAdapter, record); err != nil {
			return DeleteRestoreResult{}, err
		}
	}

	beforeSnapshot, err := s.appender.CaptureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	beforeLiveRecord, err := s.loadLiveRecordTx(ctx, tx, viewSchemaID, record.RecordID, sourceAdapter)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	nextRowVersion, err := s.envelopes.SetDeleteStateTx(ctx, tx, record.RecordID, actorID, now, deleting)
	if err != nil {
		return DeleteRestoreResult{}, adaptEnvelopeError(err)
	}
	if err := sourceAdapter.UpdateSourceDeleteStateTx(ctx, tx, record.RecordID, actorID, now, deleting); err != nil {
		return DeleteRestoreResult{}, err
	}
	if err := s.rebuildProjectionsTx(ctx, tx, record.IncidentID); err != nil {
		return DeleteRestoreResult{}, err
	}
	afterSnapshot, err := s.appender.CaptureRecordSnapshotTx(ctx, tx, record.RecordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	afterLiveRecord, err := s.loadLiveRecordTx(ctx, tx, viewSchemaID, record.RecordID, sourceAdapter)
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
		ActorUserID: actorID,
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
	if err := s.appender.AppendCapturedRecordMutationTx(ctx, tx, AppendCapturedRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		RecordID:        record.RecordID,
		OperationKind:   operation,
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  &beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return DeleteRestoreResult{}, err
	}
	if err := s.appender.AppendCapturedRecordRevisionTx(ctx, tx, AppendCapturedRecordRevisionParams{
		ChangeSetID:    changeSetID,
		RecordID:       record.RecordID,
		RowVersion:     nextRowVersion,
		BeforeSnapshot: &beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		LiveChange: LiveRecordChange{
			BeforeValue: beforeLiveRecord,
			AfterValue:  afterLiveRecord,
		},
	}); err != nil {
		return DeleteRestoreResult{}, err
	}
	current, err := s.loadDeleteRestoreRecordTx(ctx, tx, recordID)
	if err != nil {
		return DeleteRestoreResult{}, err
	}
	payload := buildDeleteRestorePayload(current, changeSetID, deleted)
	if err := s.idempotency.PutSuccessTx(ctx, tx, idempotencyKey, requestHash, payload); err != nil {
		return DeleteRestoreResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return DeleteRestoreResult{}, fmt.Errorf("commit delete restore transaction: %w", err)
	}
	return DeleteRestoreResult{
		Payload:      payload,
		IncidentID:   record.IncidentID,
		RecordID:     record.RecordID,
		RowVersion:   nextRowVersion,
		ChangeSetID:  changeSetID,
		ClientTxnID:  request.ClientTxnID,
		ViewSchemaID: viewSchemaID,
		ChangeKind:   changeKind,
	}, nil
}

func (s *commandStore) loadDeleteRestoreRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (deleteRestoreRecord, error) {
	record, err := s.envelopes.LoadEnvelopeTx(ctx, tx, recordID, true)
	if err != nil {
		return deleteRestoreRecord{}, adaptEnvelopeError(err)
	}
	return record, nil
}

func validateDeletePreconditionsTx(ctx context.Context, tx pgx.Tx, source deleterestorecontract.DeleteRestoreSource, record deleteRestoreRecord) error {
	reasonCode, blocked, err := source.ValidateDeletePreconditionsTx(ctx, tx, record.IncidentID, record.RecordID)
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

func (s *commandStore) lockDestructiveOperationRecordsNowaitTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if err := s.envelopes.LockDestructiveRecordsNowaitTx(ctx, tx, recordIDs); err != nil {
		var locked *EnvelopeLockError
		if errors.As(err, &locked) {
			return &RecordLockedError{RecordID: locked.RecordID}
		}
		if errors.Is(err, ErrEnvelopeNotFound) {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func adaptEnvelopeError(err error) error {
	if errors.Is(err, ErrEnvelopeNotFound) {
		return ErrRecordNotFound
	}
	return err
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
