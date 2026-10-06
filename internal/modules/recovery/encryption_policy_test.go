package recovery

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	"github.com/google/uuid"
)

func TestRecoveryCurrentEncryptionPolicy_Unit(t *testing.T) {
	const encoded = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
	key, err := ParseRecoveryEncryptionKey(encoded)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"", encoded + "\n", " " + encoded, base64.StdEncoding.EncodeToString(make([]byte, 31)), base64.StdEncoding.EncodeToString(make([]byte, 33)), strings.Repeat("x", 1<<20)} {
		if _, err := ParseRecoveryEncryptionKey(bad); err == nil {
			t.Fatal("invalid key admitted")
		}
	}
	if _, err := ParseRecoveryEncryptionKey(strings.TrimSuffix(encoded, "=")); err != nil {
		t.Fatal("canonical unpadded key rejected")
	}
	const aad = "operation/backup-create/completion"
	id := uuid.NewString()
	envelope, err := EncryptOperatorRecoveryJournalPayload(key, id, aad, []byte("journal"))
	if err != nil {
		t.Fatal(err)
	}
	body, err := DecryptOperatorRecoveryJournalPayload(key, aad, envelope)
	if err != nil || string(body) != "journal" {
		t.Fatalf("journal: %q %v", body, err)
	}
	for _, mutate := range []func(*OperatorRecoveryJournalEnvelope){
		func(e *OperatorRecoveryJournalEnvelope) { e.RecordID = uuid.NewString() },
		func(e *OperatorRecoveryJournalEnvelope) {
			e.SchemaID = "cartulary.operator_recovery_journal_envelope.v1"
		},
		func(e *OperatorRecoveryJournalEnvelope) { e.PayloadSHA256 = strings.Repeat("0", 64) },
		func(e *OperatorRecoveryJournalEnvelope) {
			e.SealedPayload = append([]byte(nil), e.SealedPayload...)
			e.SealedPayload[len(e.SealedPayload)-1] ^= 1
		},
	} {
		copy := envelope
		mutate(&copy)
		if _, err := DecryptOperatorRecoveryJournalPayload(key, aad, copy); err == nil {
			t.Fatal("substituted journal admitted")
		}
	}
	if _, err := DecryptOperatorRecoveryJournalPayload(key, aad+"other", envelope); err == nil {
		t.Fatal("cross-purpose journal admitted")
	}
	if _, err := EncryptOperatorRecoveryJournalPayload(key, "", aad, body); err == nil {
		t.Fatal("missing record identity admitted")
	}
	if _, err := EncryptOperatorRecoveryJournalPayload(key, id, "", body); err == nil {
		t.Fatal("missing owner context admitted")
	}
	if _, err := expectedBackupArtifactEnvelopeV3Chunks(int64(BackupArtifactMaximumChunks) * BackupArtifactChunkPlaintextBytes); err != nil {
		t.Fatal(err)
	}
	if _, err := expectedBackupArtifactEnvelopeV3Chunks(int64(BackupArtifactMaximumChunks)*BackupArtifactChunkPlaintextBytes + 1); err == nil {
		t.Fatal("excess key usage admitted")
	}

	// Independently produced with Python HKDF/HMAC and AESGCM, without a
	// production nonce/entropy hook. Generated JSON is reordered into wire order.
	for _, artifact := range contractrecovery.Artifacts {
		if artifact.Path != "contracts/recovery/fixtures/backup-artifact-envelope.v3.json" {
			continue
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(artifact.JSON), &fields); err != nil {
			t.Fatal(err)
		}
		var wire bytes.Buffer
		wire.WriteByte('{')
		for i, field := range []string{"schema_id", "application_crypto_format", "logical_ref", "content_type", "kdf", "cipher", "salt_base64", "chunk_plaintext_bytes", "chunks"} {
			if i != 0 {
				wire.WriteByte(',')
			}
			name, _ := json.Marshal(field)
			wire.Write(name)
			wire.WriteByte(':')
			if field == "chunks" {
				var chunks []struct {
					Index      uint32 `json:"index"`
					Length     int    `json:"plaintext_length"`
					Final      bool   `json:"final"`
					Ciphertext string `json:"ciphertext_base64"`
				}
				if err := json.Unmarshal(fields[field], &chunks); err != nil {
					t.Fatal(err)
				}
				raw, _ := json.Marshal(chunks)
				wire.Write(raw)
			} else {
				wire.Write(fields[field])
			}
		}
		wire.WriteByte('}')
		var ref string
		_ = json.Unmarshal(fields["logical_ref"], &ref)
		proof, err := decryptBackupArtifactEnvelopeV3(context.Background(), bytes.NewReader(wire.Bytes()), key, ref, "application/octet-stream", 0, io.Discard)
		if err != nil || proof.SizeBytes != 0 || proof.SHA256 != sha256Hex(nil) {
			t.Fatalf("independent vector: %#v %v", proof, err)
		}
		for _, attack := range []struct {
			field       string
			size, limit int
		}{
			{"schema_id", 32768, 4096},
			{"ciphertext_base64", 8 * 1024 * 1024, 6 * 1024 * 1024},
		} {
			prefix := `"` + attack.field + `":"`
			malformed := bytes.Replace(wire.Bytes(), []byte(prefix), []byte(prefix+strings.Repeat("A", attack.size)), 1)
			input := &countedReader{reader: bytes.NewReader(malformed)}
			if _, err := decryptBackupArtifactEnvelopeV3(context.Background(), input, key, ref, "", 0, io.Discard); err == nil || input.count > int64(attack.limit) {
				t.Fatalf("unbounded %s parsing: bytes=%d err=%v", attack.field, input.count, err)
			}
		}
		for _, old := range []string{"cartulary.backup_artifact_envelope.v1", "cartulary.backup_artifact_envelope.v2"} {
			bad := bytes.ReplaceAll(wire.Bytes(), []byte(BackupArtifactEnvelopeV3SchemaID), []byte(old))
			if _, err := decryptBackupArtifactEnvelopeV3(context.Background(), bytes.NewReader(bad), key, ref, "", 0, io.Discard); err == nil {
				t.Fatal("old envelope admitted")
			}
		}
		return
	}
	t.Fatal("independent fixture missing")
}
