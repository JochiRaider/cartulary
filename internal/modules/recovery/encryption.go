package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"os"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/google/uuid"
)

const (
	RecoveryMasterKeyEnv              = "CARTULARY_RECOVERY_MASTER_KEY"
	ApplicationCryptoFormatID         = "cartulary.application_crypto_format.v1"
	BackupArtifactEnvelopeSchemaID    = "cartulary.backup_artifact_envelope.v3"
	OperatorRecoveryJournalSchemaID   = "cartulary.operator_recovery_journal_envelope.v2"
	BackupStorageEncryptionModeAESGCM = "aes-256-gcm-envelope"
)

var (
	ErrRecoveryMasterKeyRequired = errors.New("recovery: recovery master key required")
	ErrRecoveryMasterKeyInvalid  = errors.New("recovery: recovery master key invalid")
	ErrEncryptedBackupStorage    = errors.New("recovery: encrypted backup storage required")
)

type BackupStorageEncryptionProof struct {
	Mode                 string `json:"mode"`
	EnvelopeSchemaID     string `json:"envelope_schema_id"`
	KeyFingerprintSHA256 string `json:"key_fingerprint_sha256"`
}

type OperatorRecoveryJournalEnvelope struct {
	RecordID             string
	SchemaID             string
	EncryptionMode       string
	KeyFingerprintSHA256 string
	PayloadSHA256        string
	SealedPayload        []byte
}

type RecoveryEncryptionKey struct {
	key         cryptography.MasterKey
	fingerprint string
}

type encryptedBackupStorage struct {
	backend StoredStreamingBackupStorage
	key     RecoveryEncryptionKey
}

type backupStorageEncryptionReporter interface {
	BackupStorageEncryptionProof() BackupStorageEncryptionProof
}

func LoadRecoveryEncryptionKey(env map[string]string) (RecoveryEncryptionKey, error) {
	raw, ok := lookupRecoveryEnv(env, RecoveryMasterKeyEnv)
	if !ok || raw == "" {
		return RecoveryEncryptionKey{}, ErrRecoveryMasterKeyRequired
	}
	return ParseRecoveryEncryptionKey(raw)
}

func ParseRecoveryEncryptionKey(raw string) (RecoveryEncryptionKey, error) {
	if raw == "" {
		return RecoveryEncryptionKey{}, ErrRecoveryMasterKeyRequired
	}
	if len(raw) != 43 && len(raw) != 44 {
		return RecoveryEncryptionKey{}, ErrRecoveryMasterKeyInvalid
	}
	encoding := base64.RawStdEncoding
	if len(raw) == 44 {
		encoding = base64.StdEncoding
	}
	decoded, err := encoding.Strict().DecodeString(raw)
	if err != nil || encoding.EncodeToString(decoded) != raw {
		return RecoveryEncryptionKey{}, ErrRecoveryMasterKeyInvalid
	}
	defer clear(decoded)
	key, err := cryptography.AdmitKey(decoded)
	if err != nil {
		return RecoveryEncryptionKey{}, ErrRecoveryMasterKeyInvalid
	}
	sum := sha256.Sum256(decoded)
	return RecoveryEncryptionKey{key: key, fingerprint: hex.EncodeToString(sum[:])}, nil
}

func NewEncryptedBackupStorage(inner BackupStorage, key RecoveryEncryptionKey) (BackupStorage, error) {
	if inner == nil {
		return nil, ErrEncryptedBackupStorage
	}
	if key.fingerprint == "" {
		return nil, ErrRecoveryMasterKeyRequired
	}
	backend, ok := inner.(StoredStreamingBackupStorage)
	if !ok {
		return nil, ErrStreamingBackupStorage
	}
	return encryptedBackupStorage{backend: backend, key: key}, nil
}

func NewEncryptedBackupStorageFromEnv(inner BackupStorage, env map[string]string) (BackupStorage, error) {
	key, err := LoadRecoveryEncryptionKey(env)
	if err != nil {
		return nil, err
	}
	return NewEncryptedBackupStorage(inner, key)
}

func EncryptOperatorRecoveryJournalPayload(key RecoveryEncryptionKey, recordID string, aad string, body []byte) (OperatorRecoveryJournalEnvelope, error) {
	if key.fingerprint == "" {
		return OperatorRecoveryJournalEnvelope{}, ErrRecoveryMasterKeyRequired
	}
	id, err := uuid.Parse(recordID)
	if err != nil || id == uuid.Nil || id.String() != recordID || strings.TrimSpace(aad) == "" || len(body) == 0 {
		return OperatorRecoveryJournalEnvelope{}, ErrInvalidBackupArtifact
	}
	sealed, err := key.key.Seal(OperatorRecoveryJournalSchemaID, []string{ApplicationCryptoFormatID, recordID, aad}, body, nil)
	if err != nil {
		return OperatorRecoveryJournalEnvelope{}, ErrInvalidBackupArtifact
	}
	return OperatorRecoveryJournalEnvelope{RecordID: recordID, SchemaID: OperatorRecoveryJournalSchemaID, EncryptionMode: BackupStorageEncryptionModeAESGCM, KeyFingerprintSHA256: key.fingerprint, PayloadSHA256: sha256Hex(body), SealedPayload: sealed}, nil
}

func DecryptOperatorRecoveryJournalPayload(key RecoveryEncryptionKey, aad string, envelope OperatorRecoveryJournalEnvelope) ([]byte, error) {
	if key.fingerprint == "" {
		return nil, ErrRecoveryMasterKeyRequired
	}
	id, err := uuid.Parse(envelope.RecordID)
	if err != nil || id == uuid.Nil || id.String() != envelope.RecordID || strings.TrimSpace(aad) == "" ||
		envelope.SchemaID != OperatorRecoveryJournalSchemaID || envelope.EncryptionMode != BackupStorageEncryptionModeAESGCM || envelope.KeyFingerprintSHA256 != key.fingerprint || !validSHA256Hex(envelope.PayloadSHA256) {
		return nil, ErrInvalidBackupArtifact
	}
	body, err := key.key.Open(OperatorRecoveryJournalSchemaID, []string{ApplicationCryptoFormatID, envelope.RecordID, aad}, envelope.SealedPayload, nil)
	if err != nil || len(body) == 0 || sha256Hex(body) != envelope.PayloadSHA256 {
		return nil, ErrInvalidBackupArtifact
	}
	return body, nil
}

func (storage encryptedBackupStorage) BackupStorageEncryptionProof() BackupStorageEncryptionProof {
	return BackupStorageEncryptionProof{Mode: BackupStorageEncryptionModeAESGCM, EnvelopeSchemaID: BackupArtifactEnvelopeSchemaID, KeyFingerprintSHA256: storage.key.fingerprint}
}

func (storage encryptedBackupStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (BackupArtifactProof, error) {
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	proof, err := storage.WriteArtifactStream(ctx, BackupArtifactStreamWriteRequest{LogicalRef: key, EnvelopeRef: key, ContentType: contentType, Plaintext: bytes.NewReader(body)})
	if err != nil {
		return BackupArtifactProof{}, err
	}
	return BackupArtifactProof{Key: proof.LogicalRef, SHA256: proof.PlaintextSHA256, SizeBytes: proof.PlaintextBytes, ContentType: proof.ContentType}, nil
}

func (storage encryptedBackupStorage) ReadArtifact(ctx context.Context, key string, maxBytes int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	key, err := validateBackupLogicalRef(key)
	if err != nil {
		return nil, err
	}
	if maxBytes < 0 {
		return nil, ErrInvalidBackupArtifact
	}
	bound, err := maximumBackupArtifactEnvelopeV3Bytes(maxBytes)
	if err != nil || bound > math.MaxInt64 {
		return nil, ErrInvalidBackupArtifact
	}
	envelope, err := storage.backend.ReadArtifact(ctx, key, int64(bound))
	if err != nil {
		return nil, err
	}
	var body bytes.Buffer
	if _, err := decryptBackupArtifactEnvelopeV3(ctx, bytes.NewReader(envelope), storage.key, key, "", maxBytes, &body); err != nil {
		return nil, err
	}
	return body.Bytes(), nil
}

func (storage encryptedBackupStorage) Close() error { return CloseBackupStorage(storage.backend) }

func backupStorageEncryptionProof(storage BackupStorage) (BackupStorageEncryptionProof, error) {
	reporter, ok := storage.(backupStorageEncryptionReporter)
	if !ok {
		return BackupStorageEncryptionProof{}, ErrEncryptedBackupStorage
	}
	proof := reporter.BackupStorageEncryptionProof()
	if proof.Mode != BackupStorageEncryptionModeAESGCM || proof.EnvelopeSchemaID != BackupArtifactEnvelopeSchemaID || !validSHA256Hex(proof.KeyFingerprintSHA256) {
		return BackupStorageEncryptionProof{}, ErrEncryptedBackupStorage
	}
	return proof, nil
}

func lookupRecoveryEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}
	return os.LookupEnv(key)
}

var _ StreamingBackupStorage = encryptedBackupStorage{}
var _ io.Closer = encryptedBackupStorage{}
