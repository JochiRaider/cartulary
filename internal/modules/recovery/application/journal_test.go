package application_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

const operatorRecoveryJournalTestKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="

func TestOperatorRecoveryJournalEncryptsPayloadAndAuditIsSafe(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "operator-recovery-journal")
	ctx := context.Background()
	key, err := recovery.ParseRecoveryEncryptionKey(operatorRecoveryJournalTestKey)
	if err != nil {
		t.Fatalf("parse recovery key: %v", err)
	}
	store := application.JournalStore{
		DB: db,
		LoadKey: func() (recovery.RecoveryEncryptionKey, error) {
			return key, nil
		},
		Now: func() time.Time {
			return time.Date(2026, 7, 6, 2, 55, 0, 0, time.UTC)
		},
	}
	backupSetID := "00000000-0000-0000-0000-000000000428"
	record := application.JournalRecord{
		OperationID: "00000000-0000-0000-0000-000000004284",
		Operation:   "backup_create",
		Result:      "failed",
		BackupSetID: &backupSetID,
		ErrorCode:   "backup_create_failed",
		ReasonCode:  "postgres_artifact_failed",
		Summary: map[string]any{
			"artifact_ref_count": 1,
			"dsn":                "postgres://user:pass@db.internal/cartulary",
			"bucket":             "operator-secret-bucket",
			"object_key":         "backup_sets/raw-object-key",
			"recovery_key":       "cleartext-recovery-key",
		},
	}
	if err := store.Append(ctx, record); err != nil {
		t.Fatalf("append journal: %v", err)
	}
	if err := store.AppendAuditSummary(ctx, record); err != nil {
		t.Fatalf("append audit summary: %v", err)
	}

	var envelopeSchemaID, encryptionMode, keyFingerprint, payloadSHA256 string
	var nonce, ciphertext []byte
	if err := db.QueryRow(ctx, `
SELECT envelope_schema_id, encryption_mode, key_fingerprint_sha256, payload_sha256, nonce, ciphertext
FROM operator_recovery_journal
WHERE operation_id = $1
`, record.OperationID).Scan(&envelopeSchemaID, &encryptionMode, &keyFingerprint, &payloadSHA256, &nonce, &ciphertext); err != nil {
		t.Fatalf("query journal row: %v", err)
	}
	if envelopeSchemaID != recovery.OperatorRecoveryJournalSchemaID || encryptionMode != recovery.BackupStorageEncryptionModeAESGCM {
		t.Fatalf("unexpected encryption envelope: schema=%q mode=%q", envelopeSchemaID, encryptionMode)
	}
	if len(keyFingerprint) != 64 || len(payloadSHA256) != 64 || len(nonce) == 0 || len(ciphertext) == 0 {
		t.Fatalf("journal row missing encryption metadata: fingerprint=%q payload=%q nonce=%d ciphertext=%d", keyFingerprint, payloadSHA256, len(nonce), len(ciphertext))
	}
	forbiddenValues := [][]byte{
		[]byte("postgres://user:pass"),
		[]byte("operator-secret-bucket"),
		[]byte("backup_sets/raw-object-key"),
		[]byte("cleartext-recovery-key"),
	}
	for _, forbidden := range forbiddenValues {
		if bytes.Contains(ciphertext, forbidden) {
			t.Fatalf("journal ciphertext exposed forbidden value %q", forbidden)
		}
	}

	var auditAfter string
	if err := db.QueryRow(ctx, `
SELECT after_json::text
FROM deployment_admin_audit_events
WHERE request_id = $1
`, record.OperationID).Scan(&auditAfter); err != nil {
		t.Fatalf("query audit summary: %v", err)
	}
	for _, forbidden := range []string{"postgres://user:pass", "operator-secret-bucket", "backup_sets/raw-object-key", "cleartext-recovery-key"} {
		if strings.Contains(auditAfter, forbidden) {
			t.Fatalf("audit summary exposed forbidden value %q in %s", forbidden, auditAfter)
		}
	}
	if !strings.Contains(auditAfter, `"artifact_ref_count": 1`) {
		t.Fatalf("audit summary lost safe operational summary: %s", auditAfter)
	}
}
