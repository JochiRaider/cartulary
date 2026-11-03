package recoveryassembly

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const recoveryEvidenceTestKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func TestRecoveryEvidenceCompletionIsTypedEncryptedAndAtomic_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "recovery-evidence-typed")
	ctx := context.Background()
	key, err := recovery.ParseRecoveryEncryptionKey(recoveryEvidenceTestKey)
	if err != nil {
		t.Fatalf("parse recovery key: %v", err)
	}
	repository, err := NewRecoveryEvidenceRepository(db, func() (recovery.RecoveryEncryptionKey, error) {
		return key, nil
	})
	if err != nil {
		t.Fatalf("construct recovery evidence repository: %v", err)
	}

	operationID := uuid.MustParse("00000000-0000-0000-0000-000000004284")
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000000428")
	startedAt := time.Date(2026, 7, 29, 2, 55, 0, 0, time.UTC)
	completedAt := startedAt.Add(5 * time.Minute)
	if err := repository.AppendAdmission(ctx, application.RecoveryAdmissionRecord{
		OperationID:   operationID,
		Operation:     application.OperationBackupCreate,
		StartedAt:     startedAt,
		ArtifactKinds: []string{"postgres_snapshot", "object_snapshot", "postgres_snapshot"},
	}); err != nil {
		t.Fatalf("append recovery admission: %v", err)
	}
	if err := repository.AppendCompletion(ctx, application.RecoveryCompletionRecord{
		OperationID:        operationID,
		Operation:          application.OperationBackupCreate,
		StartedAt:          startedAt,
		CompletedAt:        completedAt,
		Result:             application.ResultSucceeded,
		BackupSetID:        &backupSetID,
		ConsistencyPointAt: &startedAt,
		ArtifactCounts: []application.ArtifactCount{
			{Kind: "postgres_snapshot", Count: 1},
			{Kind: "backup_attestation", Count: 1},
		},
	}); err != nil {
		t.Fatalf("append recovery completion: %v", err)
	}

	var admissionEnvelope recovery.OperatorRecoveryJournalEnvelope
	if err := db.QueryRow(ctx, `
SELECT envelope_schema_id, encryption_mode, key_fingerprint_sha256,
       payload_sha256, operator_recovery_journal_id::text, sealed_payload
FROM operator_recovery_journal
WHERE operation_id = $1 AND result = 'started'
`, operationID).Scan(
		&admissionEnvelope.SchemaID,
		&admissionEnvelope.EncryptionMode,
		&admissionEnvelope.KeyFingerprintSHA256,
		&admissionEnvelope.PayloadSHA256,
		&admissionEnvelope.RecordID,
		&admissionEnvelope.SealedPayload,
	); err != nil {
		t.Fatalf("query admission recovery journal envelope: %v", err)
	}
	admissionBody, err := recovery.DecryptOperatorRecoveryJournalPayload(
		key,
		recoveryEvidenceAAD(operationID, application.OperationBackupCreate, "admission"),
		admissionEnvelope,
	)
	if err != nil {
		t.Fatalf("decrypt admission recovery journal payload: %v", err)
	}
	var admission map[string]any
	if err := json.Unmarshal(admissionBody, &admission); err != nil {
		t.Fatalf("decode admission recovery journal payload: %v", err)
	}
	wantAdmissionKeys := []string{
		"artifact_kinds", "attempt_id", "backup_set_id", "consistency_point_at",
		"operation", "operation_id", "record_kind", "schema_id", "started_at",
	}
	if got := sortedJSONKeys(admission); strings.Join(got, ",") != strings.Join(wantAdmissionKeys, ",") {
		t.Fatalf("admission journal keys got %v want %v", got, wantAdmissionKeys)
	}
	admittedKinds := admission["artifact_kinds"].([]any)
	if len(admittedKinds) != 2 || admittedKinds[0] != "object_snapshot" || admittedKinds[1] != "postgres_snapshot" {
		t.Fatalf("admission artifact kinds were not sorted and unique: %#v", admittedKinds)
	}

	var envelope recovery.OperatorRecoveryJournalEnvelope
	if err := db.QueryRow(ctx, `
SELECT envelope_schema_id, encryption_mode, key_fingerprint_sha256,
       payload_sha256, operator_recovery_journal_id::text, sealed_payload
FROM operator_recovery_journal
WHERE operation_id = $1 AND result = 'succeeded'
`, operationID).Scan(
		&envelope.SchemaID,
		&envelope.EncryptionMode,
		&envelope.KeyFingerprintSHA256,
		&envelope.PayloadSHA256,
		&envelope.RecordID,
		&envelope.SealedPayload,
	); err != nil {
		t.Fatalf("query terminal recovery journal envelope: %v", err)
	}
	body, err := recovery.DecryptOperatorRecoveryJournalPayload(
		key,
		recoveryEvidenceAAD(operationID, application.OperationBackupCreate, "completion"),
		envelope,
	)
	if err != nil {
		t.Fatalf("decrypt terminal recovery journal payload: %v", err)
	}
	var payload map[string]any
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("decode terminal recovery journal payload: %v", err)
	}
	wantKeys := []string{
		"artifact_counts", "attempt_id", "backup_set_id", "completed_at",
		"consistency_point_at", "error_code", "error_reason", "operation",
		"operation_id", "record_kind", "restore_generation", "result", "schema_id", "started_at", "target_binding_digests",
	}
	if got := sortedJSONKeys(payload); strings.Join(got, ",") != strings.Join(wantKeys, ",") {
		t.Fatalf("terminal journal keys got %v want %v", got, wantKeys)
	}
	if payload["restore_generation"] != nil || payload["target_binding_digests"] != nil {
		t.Fatalf("non-restore completion unexpectedly carries restore evidence: %#v", payload)
	}
	counts := payload["artifact_counts"].([]any)
	if first := counts[0].(map[string]any)["kind"]; first != "backup_attestation" {
		t.Fatalf("artifact counts were not sorted: %#v", counts)
	}

	var afterJSON string
	if err := db.QueryRow(ctx, `
SELECT after_json::text
FROM deployment_admin_audit_events
WHERE request_id = $1
`, operationID.String()).Scan(&afterJSON); err != nil {
		t.Fatalf("query recovery audit summary: %v", err)
	}
	var summary map[string]any
	if err := json.Unmarshal([]byte(afterJSON), &summary); err != nil {
		t.Fatalf("decode recovery audit summary: %v", err)
	}
	wantAuditKeys := []string{
		"artifact_counts", "attempt_id", "backup_set_id", "completed_at",
		"consistency_point_at", "error_code", "error_reason", "operation",
		"operation_id", "result", "schema_id", "started_at",
	}
	if got := sortedJSONKeys(summary); strings.Join(got, ",") != strings.Join(wantAuditKeys, ",") {
		t.Fatalf("audit summary keys got %v want %v", got, wantAuditKeys)
	}
}

func TestRecoveryTerminalEvidenceRollsBackJournalWhenAuditFails_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "recovery-evidence-atomic")
	ctx := context.Background()
	key, err := recovery.ParseRecoveryEncryptionKey(recoveryEvidenceTestKey)
	if err != nil {
		t.Fatalf("parse recovery key: %v", err)
	}
	repository, err := NewRecoveryEvidenceRepository(db, func() (recovery.RecoveryEncryptionKey, error) {
		return key, nil
	})
	if err != nil {
		t.Fatalf("construct recovery evidence repository: %v", err)
	}
	if _, err := db.Exec(ctx, `
CREATE OR REPLACE FUNCTION fail_recovery_audit_insert()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'injected recovery audit failure';
END;
$$;
CREATE TRIGGER fail_recovery_audit_insert
BEFORE INSERT ON deployment_admin_audit_events
FOR EACH ROW EXECUTE FUNCTION fail_recovery_audit_insert();
`); err != nil {
		t.Fatalf("install recovery audit failure injection: %v", err)
	}

	operationID := uuid.MustParse("00000000-0000-0000-0000-000000004285")
	startedAt := time.Date(2026, 7, 29, 3, 0, 0, 0, time.UTC)
	errorCode := "restore_failed"
	errorReason := "invariant_check_failed"
	err = repository.AppendCompletion(ctx, application.RecoveryCompletionRecord{
		OperationID: operationID,
		Operation:   application.OperationRestoreLatest,
		StartedAt:   startedAt,
		CompletedAt: startedAt.Add(time.Minute),
		Result:      application.ResultStatus("failed"),
		ErrorCode:   &errorCode,
		ErrorReason: &errorReason,
	})
	if err == nil || !strings.Contains(err.Error(), "injected recovery audit failure") {
		t.Fatalf("terminal evidence failure got %v", err)
	}

	var journalCount int
	if err := db.QueryRow(ctx, `
SELECT count(*) FROM operator_recovery_journal WHERE operation_id = $1
`, operationID).Scan(&journalCount); err != nil {
		t.Fatalf("count rolled-back journal rows: %v", err)
	}
	if journalCount != 0 {
		t.Fatalf("terminal journal committed without audit: count=%d", journalCount)
	}
	var auditCount int
	if err := db.QueryRow(ctx, `
SELECT count(*) FROM deployment_admin_audit_events WHERE request_id = $1
`, operationID.String()).Scan(&auditCount); err != nil {
		t.Fatalf("count rolled-back audit rows: %v", err)
	}
	if auditCount != 0 {
		t.Fatalf("terminal audit committed despite failure: count=%d", auditCount)
	}
}

func TestRecoveryCompletionRecordsCloseEveryOperationAndFailureTiming_Unit(t *testing.T) {
	startedAt := time.Date(2026, 7, 29, 4, 0, 0, 0, time.UTC)
	completedAt := startedAt.Add(time.Minute)
	backupSetID := uuid.MustParse("00000000-0000-0000-0000-000000000429")
	consistencyPointAt := startedAt.Add(-time.Hour)
	generation := recovery.RestoreGenerationEvidence{
		TargetGenerationID:         uuid.MustParse("00000000-0000-0000-0000-000000005001"),
		RecoveryStateCatalogSHA256: strings.Repeat("5", 64),
		CodecRegistrySHA256:        strings.Repeat("6", 64),
	}
	bindings := recovery.TargetBindingDigests{
		DatabaseSHA256:             strings.Repeat("1", 64),
		ObjectStoreSHA256:          strings.Repeat("2", 64),
		ReferencePackStorageSHA256: strings.Repeat("3", 64),
		ExportOutputsSHA256:        strings.Repeat("4", 64),
	}
	errorCode := "verification_failed"
	errorReason := "invariant_check_failed"
	tests := []struct {
		name      string
		operation application.Operation
		result    application.ResultStatus
		late      bool
		attempt   bool
		admitted  bool
	}{
		{"backup success", application.OperationBackupCreate, application.ResultSucceeded, true, false, false},
		{"backup early failure", application.OperationBackupCreate, application.ResultFailed, false, false, false},
		{"backup late failure", application.OperationBackupCreate, application.ResultFailed, true, false, false},
		{"export success", application.OperationBackupExportLatest, application.ResultSucceeded, true, false, false},
		{"export early failure", application.OperationBackupExportLatest, application.ResultFailed, false, false, false},
		{"export late failure", application.OperationBackupExportLatest, application.ResultFailed, true, false, false},
		{"bundle restore success", application.OperationRestoreBundle, application.ResultSucceeded, true, false, true},
		{"bundle restore early failure", application.OperationRestoreBundle, application.ResultFailed, false, false, false},
		{"bundle restore late failure", application.OperationRestoreBundle, application.ResultFailed, true, false, false},
		{"restore success", application.OperationRestoreLatest, application.ResultSucceeded, true, false, true},
		{"restore early failure", application.OperationRestoreLatest, application.ResultFailed, false, false, false},
		{"restore late failure", application.OperationRestoreLatest, application.ResultFailed, true, false, false},
		{"verify latest success", application.OperationRestoreVerifyLatest, application.ResultSucceeded, true, false, true},
		{"verify latest early failure", application.OperationRestoreVerifyLatest, application.ResultFailed, false, false, false},
		{"verify latest late failure", application.OperationRestoreVerifyLatest, application.ResultFailed, true, false, false},
		{"verify due success", application.OperationRestoreVerifyDue, application.ResultSucceeded, true, false, false},
		{"verify due no op", application.OperationRestoreVerifyDue, application.ResultNoOp, false, false, false},
		{"verify due early failure", application.OperationRestoreVerifyDue, application.ResultFailed, false, false, false},
		{"verify due late failure", application.OperationRestoreVerifyDue, application.ResultFailed, true, false, false},
		{"verify due attempt success", application.OperationRestoreVerifyDue, application.ResultSucceeded, true, true, true},
		{"verify due attempt early failure", application.OperationRestoreVerifyDue, application.ResultFailed, false, true, false},
		{"verify due admitted attempt failure", application.OperationRestoreVerifyDue, application.ResultFailed, true, true, true},
	}

	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := application.RecoveryCompletionRecord{
				OperationID: uuid.MustParse("00000000-0000-0000-0000-" + leftPadDecimal(index+5000, 12)),
				Operation:   test.operation,
				StartedAt:   startedAt,
				CompletedAt: completedAt,
				Result:      test.result,
			}
			if test.result == application.ResultFailed {
				record.ErrorCode = &errorCode
				record.ErrorReason = &errorReason
			}
			if test.late {
				record.BackupSetID = &backupSetID
				record.ConsistencyPointAt = &consistencyPointAt
				record.ArtifactCounts = []application.ArtifactCount{{Kind: "restore_verification", Count: 1}}
			}
			if test.attempt {
				attemptID := "00000000-0000-0000-0000-000000005002"
				record.AttemptID = &attemptID
			}
			if test.admitted {
				record.RestoreGeneration = &generation
				record.TargetBindings = &bindings
			}
			if test.admitted && test.result == application.ResultSucceeded {
				for _, omitted := range []string{"generation", "bindings", "both"} {
					incomplete := record
					if omitted != "bindings" {
						incomplete.RestoreGeneration = nil
					}
					if omitted != "generation" {
						incomplete.TargetBindings = nil
					}
					if _, err := application.NormalizeCompletionRecord(incomplete); err == nil {
						t.Fatalf("successful target completion accepted missing %s", omitted)
					}
				}
			}
			normalized, err := application.NormalizeCompletionRecord(record)
			if err != nil {
				t.Fatalf("normalize completion record: %v", err)
			}
			body, err := application.EncodeRecoveryCompletion(normalized)
			if err != nil {
				t.Fatalf("encode completion payload: %v", err)
			}
			for _, forbidden := range []string{"dsn", "endpoint", "bucket", "object_key", "recovery_key", "secret", "incident_content"} {
				if bytes.Contains(body, []byte(forbidden)) {
					t.Fatalf("typed completion payload exposed forbidden field %q: %s", forbidden, body)
				}
			}
			var decoded map[string]any
			if err := json.Unmarshal(body, &decoded); err != nil {
				t.Fatalf("decode completion payload: %v", err)
			}
			if _, exists := decoded["graph_projection_completion"]; exists {
				t.Fatal("retired Graph completion field entered the current journal")
			}
			current, err := application.DecodeRecoveryJournalPayload(body)
			if err != nil {
				t.Fatalf("current typed journal rejected completion: %v", err)
			}
			if current.SchemaID != application.RecoveryJournalPayloadSchemaID || current.RecordKind != "completion" ||
				(current.RestoreGeneration != nil) != test.admitted || (current.TargetBindings != nil) != test.admitted {
				t.Fatalf("completion lost current identity or proof presence: %#v", current)
			}
			if test.admitted && (*current.RestoreGeneration != generation || *current.TargetBindings != bindings) {
				t.Fatal("completion changed the admitted generation or target bindings")
			}
		})
	}

	if _, err := application.NormalizeCompletionRecord(application.RecoveryCompletionRecord{
		OperationID: uuid.MustParse("00000000-0000-0000-0000-000000009998"),
		Operation:   application.OperationBackupInspectLatest,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Result:      application.ResultSucceeded,
	}); err == nil {
		t.Fatal("read-only inspection accepted a mutating completion record")
	}
	_, err := application.NormalizeCompletionRecord(application.RecoveryCompletionRecord{
		OperationID: uuid.MustParse("00000000-0000-0000-0000-000000009999"),
		Operation:   application.OperationBackupCreate,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Result:      application.ResultNoOp,
	})
	if err == nil {
		t.Fatal("backup no-op completion must be rejected")
	}
}

func sortedJSONKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func leftPadDecimal(value int, width int) string {
	rendered := fmt.Sprintf("%d", value)
	return strings.Repeat("0", width-len(rendered)) + rendered
}
