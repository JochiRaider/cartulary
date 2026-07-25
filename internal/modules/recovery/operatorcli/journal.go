package operatorcli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const journalPayloadSchemaID = "cartulary.operator_recovery_journal_payload.v1"

type JournalStore struct {
	DB      postgres.DB
	LoadKey func() (recovery.RecoveryEncryptionKey, error)
	Now     func() time.Time
}

type JournalRecord struct {
	OperationID string
	Operation   string
	Result      string
	BackupSetID *string
	ErrorCode   string
	ReasonCode  string
	Summary     map[string]any
	CompletedAt time.Time
}

type journalPayload struct {
	SchemaID    string         `json:"schema_id"`
	OperationID string         `json:"operation_id"`
	Operation   string         `json:"operation"`
	Result      string         `json:"result"`
	BackupSetID *string        `json:"backup_set_id"`
	ErrorCode   string         `json:"error_code,omitempty"`
	ReasonCode  string         `json:"reason_code,omitempty"`
	Summary     map[string]any `json:"summary,omitempty"`
	RecordedAt  time.Time      `json:"recorded_at"`
}

func (store JournalStore) Append(ctx context.Context, record JournalRecord) error {
	if store.DB == nil {
		return fmt.Errorf("operator recovery journal requires writable database")
	}
	if store.LoadKey == nil {
		return fmt.Errorf("operator recovery journal requires recovery key loader")
	}
	operationID, err := uuid.Parse(record.OperationID)
	if err != nil {
		return fmt.Errorf("operator recovery journal operation_id must be UUID: %w", err)
	}
	var backupSetID any
	if record.BackupSetID != nil {
		parsed, err := uuid.Parse(*record.BackupSetID)
		if err != nil {
			return fmt.Errorf("operator recovery journal backup_set_id must be UUID: %w", err)
		}
		backupSetID = parsed
	}
	recordedAt := record.CompletedAt.UTC()
	if recordedAt.IsZero() {
		recordedAt = store.now()().UTC()
	}
	body, err := json.Marshal(journalPayload{
		SchemaID:    journalPayloadSchemaID,
		OperationID: record.OperationID,
		Operation:   record.Operation,
		Result:      record.Result,
		BackupSetID: record.BackupSetID,
		ErrorCode:   record.ErrorCode,
		ReasonCode:  record.ReasonCode,
		Summary:     safeJournalSummary(record.Summary),
		RecordedAt:  recordedAt,
	})
	if err != nil {
		return fmt.Errorf("encode operator recovery journal payload: %w", err)
	}
	key, err := store.LoadKey()
	if err != nil {
		return err
	}
	envelope, err := recovery.EncryptOperatorRecoveryJournalPayload(key, recovery.OperatorRecoveryJournalSchemaID+"\n"+record.OperationID+"\n"+record.Operation, body)
	if err != nil {
		return err
	}
	if _, err := store.DB.Exec(ctx, `
INSERT INTO operator_recovery_journal (
    operation_id,
    operation,
    result,
    backup_set_id,
    error_code,
    reason_code,
    envelope_schema_id,
    encryption_mode,
    key_fingerprint_sha256,
    payload_sha256,
    nonce,
    ciphertext,
    created_at
)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9, $10, $11, $12, $13)
`, operationID, record.Operation, record.Result, backupSetID, record.ErrorCode, record.ReasonCode, envelope.SchemaID, envelope.EncryptionMode, envelope.KeyFingerprintSHA256, envelope.PayloadSHA256, envelope.Nonce, envelope.Ciphertext, recordedAt); err != nil {
		return fmt.Errorf("append operator recovery journal: %w", err)
	}
	return nil
}

func (store JournalStore) AppendAuditSummary(ctx context.Context, record JournalRecord) error {
	if store.DB == nil {
		return fmt.Errorf("operator recovery audit requires writable database")
	}
	recordedAt := record.CompletedAt.UTC()
	if recordedAt.IsZero() {
		recordedAt = store.now()().UTC()
	}
	after := map[string]any{
		"schema_id":     "cartulary.operator_recovery_audit_summary.v1",
		"operation_id":  record.OperationID,
		"operation":     record.Operation,
		"result":        record.Result,
		"backup_set_id": record.BackupSetID,
		"error_code":    record.ErrorCode,
		"reason_code":   record.ReasonCode,
		"summary":       safeJournalSummary(record.Summary),
		"recorded_at":   recordedAt,
	}
	var reasonCode *string
	if record.ReasonCode != "" {
		reasonCode = &record.ReasonCode
	}
	raw := administrativeaudit.RawEvent{
		EventSource: "operator.recovery." + record.Operation,
		EventKind:   "operator_recovery_" + record.Result,
		ReasonCode:  reasonCode,
		RequestID:   &record.OperationID,
		After:       after,
		OccurredAt:  recordedAt,
	}
	actionCode, targetKind, targetID, changes, projected := recoveryAuditProjection(record)
	if !projected {
		if _, err := administrativeaudit.AppendRaw(ctx, store.DB, raw); err != nil {
			return fmt.Errorf("append operator recovery raw audit summary: %w", err)
		}
		return nil
	}
	if _, err := administrativeaudit.Append(ctx, store.DB, raw, administrativeaudit.Event{
		ScopeKind:  administrativeaudit.ScopeDeployment,
		OccurredAt: recordedAt,
		ActorKind:  administrativeaudit.ActorOperator,
		Source:     administrativeaudit.SourceOperator,
		ActionCode: actionCode,
		TargetKind: targetKind,
		TargetID:   &targetID,
		Changes:    changes,
		ReasonCode: reasonCode,
	}); err != nil {
		return fmt.Errorf("append operator recovery audit summary: %w", err)
	}
	return nil
}

func recoveryAuditProjection(record JournalRecord) (string, string, string, []administrativeaudit.Change, bool) {
	changes := []administrativeaudit.Change{
		administrativeaudit.Visible("operation", nil, record.Operation),
		administrativeaudit.Visible("result", nil, record.Result),
	}
	if record.ErrorCode != "" {
		changes = append(changes, administrativeaudit.Visible("error_code", nil, record.ErrorCode))
	}
	if record.ReasonCode != "" {
		changes = append(changes, administrativeaudit.Visible("reason_code", nil, record.ReasonCode))
	}
	switch record.Operation {
	case "backup_create":
		if record.Result != "succeeded" || record.BackupSetID == nil || *record.BackupSetID == "" {
			return "", "", "", nil, false
		}
		changes = append(changes, administrativeaudit.Visible("backup_set_id", nil, *record.BackupSetID))
		return administrativeaudit.ActionBackupCreated, administrativeaudit.TargetBackupSet, *record.BackupSetID, changes, true
	case "restore_latest":
		actionCode := administrativeaudit.ActionRestoreCompleted
		switch record.Result {
		case "started":
			actionCode = administrativeaudit.ActionRestoreStarted
		case "failed":
			actionCode = administrativeaudit.ActionRestoreFailed
		}
		return actionCode, administrativeaudit.TargetRestoreOperation, record.OperationID, changes, true
	case "restore_verify_latest", "restore_verify_due":
		return administrativeaudit.ActionRestoreVerificationCompleted, administrativeaudit.TargetRestoreOperation, record.OperationID, changes, true
	default:
		return "", "", "", nil, false
	}
}

func (store JournalStore) now() func() time.Time {
	if store.Now != nil {
		return store.Now
	}
	return func() time.Time { return time.Now().UTC() }
}

func safeJournalSummary(summary map[string]any) map[string]any {
	if len(summary) == 0 {
		return nil
	}
	out := make(map[string]any, len(summary))
	for key, value := range summary {
		switch key {
		case "dsn", "raw_dsn", "endpoint", "host", "bucket", "object_key", "path", "recovery_key", "secret", "incident_content":
			continue
		default:
			out[key] = value
		}
	}
	return out
}
