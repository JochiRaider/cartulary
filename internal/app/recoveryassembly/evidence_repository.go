package recoveryassembly

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type recoveryEvidenceRepository struct {
	db      postgres.DB
	loadKey func() (recovery.RecoveryEncryptionKey, error)
}

type recoveryJournalAdmissionPayload struct {
	SchemaID           string                `json:"schema_id"`
	RecordKind         string                `json:"record_kind"`
	OperationID        uuid.UUID             `json:"operation_id"`
	Operation          application.Operation `json:"operation"`
	AttemptID          *string               `json:"attempt_id"`
	StartedAt          time.Time             `json:"started_at"`
	BackupSetID        *uuid.UUID            `json:"backup_set_id"`
	ConsistencyPointAt *time.Time            `json:"consistency_point_at"`
	ArtifactKinds      []string              `json:"artifact_kinds"`
}

type recoveryJournalCompletionPayload struct {
	SchemaID                  string                                         `json:"schema_id"`
	RecordKind                string                                         `json:"record_kind"`
	OperationID               uuid.UUID                                      `json:"operation_id"`
	Operation                 application.Operation                          `json:"operation"`
	AttemptID                 *string                                        `json:"attempt_id"`
	StartedAt                 time.Time                                      `json:"started_at"`
	CompletedAt               time.Time                                      `json:"completed_at"`
	Result                    application.ResultStatus                       `json:"result"`
	BackupSetID               *uuid.UUID                                     `json:"backup_set_id"`
	ConsistencyPointAt        *time.Time                                     `json:"consistency_point_at"`
	ArtifactCounts            []application.ArtifactCount                    `json:"artifact_counts"`
	ErrorCode                 *string                                        `json:"error_code"`
	ErrorReason               *string                                        `json:"error_reason"`
	GraphProjectionCompletion *application.GraphProjectionCompletionEvidence `json:"graph_projection_completion"`
}

type recoveryAuditSummary struct {
	SchemaID           string                      `json:"schema_id"`
	OperationID        uuid.UUID                   `json:"operation_id"`
	Operation          application.Operation       `json:"operation"`
	AttemptID          *string                     `json:"attempt_id"`
	Result             application.ResultStatus    `json:"result"`
	StartedAt          time.Time                   `json:"started_at"`
	CompletedAt        time.Time                   `json:"completed_at"`
	BackupSetID        *uuid.UUID                  `json:"backup_set_id"`
	ConsistencyPointAt *time.Time                  `json:"consistency_point_at"`
	ArtifactCounts     []application.ArtifactCount `json:"artifact_counts"`
	ErrorCode          *string                     `json:"error_code"`
	ErrorReason        *string                     `json:"error_reason"`
}

func NewRecoveryEvidenceRepository(
	db postgres.DB,
	loadKey func() (recovery.RecoveryEncryptionKey, error),
) (application.RecoveryEvidenceRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("recovery evidence repository requires writable database")
	}
	if loadKey == nil {
		return nil, fmt.Errorf("recovery evidence repository requires recovery key loader")
	}
	return &recoveryEvidenceRepository{db: db, loadKey: loadKey}, nil
}

func (repository *recoveryEvidenceRepository) AppendAdmission(ctx context.Context, record application.RecoveryAdmissionRecord) error {
	record, err := application.NormalizeAdmissionRecord(record)
	if err != nil {
		return err
	}
	payload := recoveryJournalAdmissionPayload{
		SchemaID:           application.RecoveryJournalPayloadSchemaID,
		RecordKind:         "admission",
		OperationID:        record.OperationID,
		Operation:          record.Operation,
		AttemptID:          record.AttemptID,
		StartedAt:          record.StartedAt,
		BackupSetID:        record.BackupSetID,
		ConsistencyPointAt: record.ConsistencyPointAt,
		ArtifactKinds:      record.ArtifactKinds,
	}
	envelope, err := repository.encryptPayload(record.OperationID, record.Operation, "admission", payload)
	if err != nil {
		return err
	}
	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin recovery admission evidence transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := appendRecoveryJournalEnvelope(ctx, tx, recoveryJournalEnvelopeRecord{
		OperationID: record.OperationID,
		Operation:   record.Operation,
		Result:      "started",
		BackupSetID: record.BackupSetID,
		Envelope:    envelope,
		RecordedAt:  record.StartedAt,
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit recovery admission evidence transaction: %w", err)
	}
	return nil
}

func (repository *recoveryEvidenceRepository) AppendCompletion(ctx context.Context, record application.RecoveryCompletionRecord) error {
	record, err := application.NormalizeCompletionRecord(record)
	if err != nil {
		return err
	}
	payload := recoveryJournalCompletionPayload{
		SchemaID:                  application.RecoveryJournalPayloadSchemaID,
		RecordKind:                "completion",
		OperationID:               record.OperationID,
		Operation:                 record.Operation,
		AttemptID:                 record.AttemptID,
		StartedAt:                 record.StartedAt,
		CompletedAt:               record.CompletedAt,
		Result:                    record.Result,
		BackupSetID:               record.BackupSetID,
		ConsistencyPointAt:        record.ConsistencyPointAt,
		ArtifactCounts:            record.ArtifactCounts,
		ErrorCode:                 record.ErrorCode,
		ErrorReason:               record.ErrorReason,
		GraphProjectionCompletion: record.GraphProjectionCompletion,
	}
	envelope, err := repository.encryptPayload(record.OperationID, record.Operation, "completion", payload)
	if err != nil {
		return err
	}
	summary := recoveryAuditSummary{
		SchemaID:           application.RecoveryAuditSummarySchemaID,
		OperationID:        record.OperationID,
		Operation:          record.Operation,
		AttemptID:          record.AttemptID,
		Result:             record.Result,
		StartedAt:          record.StartedAt,
		CompletedAt:        record.CompletedAt,
		BackupSetID:        record.BackupSetID,
		ConsistencyPointAt: record.ConsistencyPointAt,
		ArtifactCounts:     record.ArtifactCounts,
		ErrorCode:          record.ErrorCode,
		ErrorReason:        record.ErrorReason,
	}

	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin recovery terminal evidence transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := appendRecoveryJournalEnvelope(ctx, tx, recoveryJournalEnvelopeRecord{
		OperationID: record.OperationID,
		Operation:   record.Operation,
		Result:      string(record.Result),
		BackupSetID: record.BackupSetID,
		ErrorCode:   record.ErrorCode,
		ErrorReason: record.ErrorReason,
		Envelope:    envelope,
		RecordedAt:  record.CompletedAt,
	}); err != nil {
		return err
	}
	if err := appendRecoveryAuditSummary(ctx, tx, record, summary); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit recovery terminal evidence transaction: %w", err)
	}
	return nil
}

func (repository *recoveryEvidenceRepository) FindSuccessfulCompletion(
	ctx context.Context,
	operationID uuid.UUID,
	operation application.Operation,
	attemptID *string,
	backupSetID uuid.UUID,
) (*application.RecoveryCompletionRecord, error) {
	if repository == nil || repository.db == nil || operationID == uuid.Nil || backupSetID == uuid.Nil {
		return nil, fmt.Errorf("successful Recovery completion lookup requires exact identities")
	}
	rows, err := repository.db.Query(ctx, `
SELECT envelope_schema_id, encryption_mode, key_fingerprint_sha256,
       payload_sha256, nonce, ciphertext
  FROM operator_recovery_journal
 WHERE operation_id = $1
   AND operation = $2
   AND backup_set_id = $3
   AND result = 'succeeded'
 ORDER BY created_at DESC, operator_recovery_journal_id ASC
`, operationID, operation, backupSetID)
	if err != nil {
		return nil, fmt.Errorf("query successful Recovery completion: %w", err)
	}
	defer rows.Close()
	key, err := repository.loadKey()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var envelope recovery.OperatorRecoveryJournalEnvelope
		if err := rows.Scan(
			&envelope.SchemaID,
			&envelope.EncryptionMode,
			&envelope.KeyFingerprintSHA256,
			&envelope.PayloadSHA256,
			&envelope.Nonce,
			&envelope.Ciphertext,
		); err != nil {
			return nil, fmt.Errorf("scan successful Recovery completion: %w", err)
		}
		body, err := recovery.DecryptOperatorRecoveryJournalPayload(
			key,
			recoveryEvidenceAAD(operationID, operation, "completion"),
			envelope,
		)
		if err != nil {
			return nil, err
		}
		var payload recoveryJournalCompletionPayload
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&payload); err != nil {
			return nil, fmt.Errorf("decode successful Recovery completion: %w", err)
		}
		if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("successful Recovery completion has trailing content")
		}
		if payload.SchemaID != application.RecoveryJournalPayloadSchemaID || payload.RecordKind != "completion" ||
			payload.OperationID != operationID || payload.Operation != operation || payload.Result != application.ResultSucceeded ||
			payload.BackupSetID == nil || *payload.BackupSetID != backupSetID || !sameOptionalString(payload.AttemptID, attemptID) ||
			payload.GraphProjectionCompletion == nil {
			continue
		}
		record, err := application.NormalizeCompletionRecord(application.RecoveryCompletionRecord{
			OperationID: payload.OperationID, Operation: payload.Operation, AttemptID: payload.AttemptID,
			StartedAt: payload.StartedAt, CompletedAt: payload.CompletedAt, Result: payload.Result,
			BackupSetID: payload.BackupSetID, ConsistencyPointAt: payload.ConsistencyPointAt,
			ArtifactCounts: payload.ArtifactCounts, ErrorCode: payload.ErrorCode, ErrorReason: payload.ErrorReason,
			GraphProjectionCompletion: payload.GraphProjectionCompletion,
		})
		if err != nil {
			return nil, err
		}
		return &record, nil
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate successful Recovery completions: %w", err)
	}
	return nil, nil
}

func sameOptionalString(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func (repository *recoveryEvidenceRepository) encryptPayload(
	operationID uuid.UUID,
	operation application.Operation,
	recordKind string,
	payload any,
) (recovery.OperatorRecoveryJournalEnvelope, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return recovery.OperatorRecoveryJournalEnvelope{}, fmt.Errorf("encode operator recovery journal payload: %w", err)
	}
	key, err := repository.loadKey()
	if err != nil {
		return recovery.OperatorRecoveryJournalEnvelope{}, err
	}
	envelope, err := recovery.EncryptOperatorRecoveryJournalPayload(
		key,
		recoveryEvidenceAAD(operationID, operation, recordKind),
		body,
	)
	if err != nil {
		return recovery.OperatorRecoveryJournalEnvelope{}, err
	}
	return envelope, nil
}

type recoveryJournalEnvelopeRecord struct {
	OperationID uuid.UUID
	Operation   application.Operation
	Result      string
	BackupSetID *uuid.UUID
	ErrorCode   *string
	ErrorReason *string
	Envelope    recovery.OperatorRecoveryJournalEnvelope
	RecordedAt  time.Time
}

func appendRecoveryJournalEnvelope(ctx context.Context, tx pgx.Tx, record recoveryJournalEnvelopeRecord) error {
	if _, err := tx.Exec(ctx, `
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
`, record.OperationID, record.Operation, record.Result, record.BackupSetID, record.ErrorCode, record.ErrorReason, record.Envelope.SchemaID, record.Envelope.EncryptionMode, record.Envelope.KeyFingerprintSHA256, record.Envelope.PayloadSHA256, record.Envelope.Nonce, record.Envelope.Ciphertext, record.RecordedAt); err != nil {
		return fmt.Errorf("append operator recovery journal: %w", err)
	}
	return nil
}

func appendRecoveryAuditSummary(
	ctx context.Context,
	tx pgx.Tx,
	record application.RecoveryCompletionRecord,
	summary recoveryAuditSummary,
) error {
	reasonCode := record.ErrorReason
	operationID := record.OperationID.String()
	raw := administrativeaudit.RawEvent{
		EventSource: "operator.recovery." + string(record.Operation),
		EventKind:   "operator_recovery_" + string(record.Result),
		ReasonCode:  reasonCode,
		RequestID:   &operationID,
		After:       summary,
		OccurredAt:  record.CompletedAt,
	}
	actionCode, targetKind, targetID, changes, projected := recoveryAuditProjection(record)
	if !projected {
		if _, err := administrativeaudit.AppendRawTx(ctx, tx, raw); err != nil {
			return fmt.Errorf("append operator recovery raw audit summary: %w", err)
		}
		return nil
	}
	if _, err := administrativeaudit.AppendTx(ctx, tx, raw, administrativeaudit.Event{
		ScopeKind:  administrativeaudit.ScopeDeployment,
		OccurredAt: record.CompletedAt,
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

func recoveryAuditProjection(record application.RecoveryCompletionRecord) (string, string, string, []administrativeaudit.Change, bool) {
	changes := []administrativeaudit.Change{
		administrativeaudit.Visible("operation", nil, record.Operation),
		administrativeaudit.Visible("result", nil, record.Result),
	}
	if record.ErrorCode != nil {
		changes = append(changes, administrativeaudit.Visible("error_code", nil, *record.ErrorCode))
	}
	if record.ErrorReason != nil {
		changes = append(changes, administrativeaudit.Visible("reason_code", nil, *record.ErrorReason))
	}
	switch record.Operation {
	case application.OperationBackupCreate:
		if record.Result != application.ResultSucceeded || record.BackupSetID == nil {
			return "", "", "", nil, false
		}
		targetID := record.BackupSetID.String()
		changes = append(changes, administrativeaudit.Visible("backup_set_id", nil, targetID))
		return administrativeaudit.ActionBackupCreated, administrativeaudit.TargetBackupSet, targetID, changes, true
	case application.OperationRestoreLatest:
		actionCode := administrativeaudit.ActionRestoreCompleted
		if record.Result == application.ResultStatus("failed") {
			actionCode = administrativeaudit.ActionRestoreFailed
		}
		return actionCode, administrativeaudit.TargetRestoreOperation, record.OperationID.String(), changes, true
	case application.OperationRestoreVerifyLatest, application.OperationRestoreVerifyDue:
		return administrativeaudit.ActionRestoreVerificationCompleted, administrativeaudit.TargetRestoreOperation, record.OperationID.String(), changes, true
	default:
		return "", "", "", nil, false
	}
}

func recoveryEvidenceAAD(operationID uuid.UUID, operation application.Operation, recordKind string) string {
	return application.RecoveryJournalPayloadSchemaID + "\n" + operationID.String() + "\n" + string(operation) + "\n" + recordKind
}
