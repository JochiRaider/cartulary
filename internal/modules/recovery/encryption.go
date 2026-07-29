package recovery

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

const (
	RecoveryMasterKeyEnv              = "CARTULARY_RECOVERY_MASTER_KEY"
	BackupArtifactEnvelopeSchemaID    = "cartulary.backup_artifact_envelope.v1"
	OperatorRecoveryJournalSchemaID   = "cartulary.operator_recovery_journal_envelope.v1"
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
	SchemaID             string
	EncryptionMode       string
	KeyFingerprintSHA256 string
	PayloadSHA256        string
	Nonce                []byte
	Ciphertext           []byte
}

type RecoveryEncryptionKey struct {
	key         [32]byte
	fingerprint string
}

type encryptedBackupStorage struct {
	inner BackupStorage
	key   RecoveryEncryptionKey
}

type backupArtifactEnvelope struct {
	SchemaID             string `json:"schema_id"`
	EncryptionMode       string `json:"encryption_mode"`
	KeyFingerprintSHA256 string `json:"key_fingerprint_sha256"`
	ArtifactKey          string `json:"artifact_key"`
	PlaintextContentType string `json:"plaintext_content_type"`
	NonceBase64          string `json:"nonce_base64"`
	CiphertextBase64     string `json:"ciphertext_base64"`
}

type backupStorageEncryptionReporter interface {
	BackupStorageEncryptionProof() BackupStorageEncryptionProof
}

func LoadRecoveryEncryptionKey(env map[string]string) (RecoveryEncryptionKey, error) {
	raw, ok := lookupRecoveryEnv(env, RecoveryMasterKeyEnv)
	if !ok || strings.TrimSpace(raw) == "" {
		return RecoveryEncryptionKey{}, ErrRecoveryMasterKeyRequired
	}
	return ParseRecoveryEncryptionKey(raw)
}

func ParseRecoveryEncryptionKey(raw string) (RecoveryEncryptionKey, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return RecoveryEncryptionKey{}, ErrRecoveryMasterKeyRequired
	}
	decoded, err := base64.RawStdEncoding.DecodeString(trimmed)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(trimmed)
		if err != nil {
			return RecoveryEncryptionKey{}, fmt.Errorf("%w: decode %s: %v", ErrRecoveryMasterKeyInvalid, RecoveryMasterKeyEnv, err)
		}
	}
	if len(decoded) < 32 {
		return RecoveryEncryptionKey{}, fmt.Errorf("%w: %s must decode to at least 32 bytes", ErrRecoveryMasterKeyInvalid, RecoveryMasterKeyEnv)
	}
	var key [32]byte
	copy(key[:], decoded[:32])
	sum := sha256.Sum256(decoded[:32])
	return RecoveryEncryptionKey{
		key:         key,
		fingerprint: hex.EncodeToString(sum[:]),
	}, nil
}

func NewEncryptedBackupStorage(inner BackupStorage, key RecoveryEncryptionKey) (BackupStorage, error) {
	if inner == nil {
		return nil, fmt.Errorf("%w: inner backup storage is required", ErrEncryptedBackupStorage)
	}
	if key.fingerprint == "" {
		return nil, ErrRecoveryMasterKeyRequired
	}
	legacy := encryptedBackupStorage{inner: inner, key: key}
	if backend, ok := inner.(StoredStreamingBackupStorage); ok {
		return streamingEncryptedBackupStorage{
			encryptedBackupStorage: legacy,
			backend:                backend,
		}, nil
	}
	return legacy, nil
}

func NewEncryptedBackupStorageFromEnv(inner BackupStorage, env map[string]string) (BackupStorage, error) {
	key, err := LoadRecoveryEncryptionKey(env)
	if err != nil {
		return nil, err
	}
	return NewEncryptedBackupStorage(inner, key)
}

func EncryptOperatorRecoveryJournalPayload(key RecoveryEncryptionKey, aad string, body []byte) (OperatorRecoveryJournalEnvelope, error) {
	if key.fingerprint == "" {
		return OperatorRecoveryJournalEnvelope{}, ErrRecoveryMasterKeyRequired
	}
	if len(body) == 0 {
		return OperatorRecoveryJournalEnvelope{}, errors.New("recovery: operator recovery journal payload is empty")
	}
	block, err := aes.NewCipher(key.key[:])
	if err != nil {
		return OperatorRecoveryJournalEnvelope{}, fmt.Errorf("create operator recovery journal cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return OperatorRecoveryJournalEnvelope{}, fmt.Errorf("create operator recovery journal gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return OperatorRecoveryJournalEnvelope{}, fmt.Errorf("generate operator recovery journal nonce: %w", err)
	}
	if strings.TrimSpace(aad) == "" {
		aad = OperatorRecoveryJournalSchemaID
	}
	return OperatorRecoveryJournalEnvelope{
		SchemaID:             OperatorRecoveryJournalSchemaID,
		EncryptionMode:       BackupStorageEncryptionModeAESGCM,
		KeyFingerprintSHA256: key.fingerprint,
		PayloadSHA256:        sha256Hex(body),
		Nonce:                nonce,
		Ciphertext:           gcm.Seal(nil, nonce, body, []byte(aad)),
	}, nil
}

func DecryptOperatorRecoveryJournalPayload(
	key RecoveryEncryptionKey,
	aad string,
	envelope OperatorRecoveryJournalEnvelope,
) ([]byte, error) {
	if key.fingerprint == "" {
		return nil, ErrRecoveryMasterKeyRequired
	}
	if envelope.SchemaID != OperatorRecoveryJournalSchemaID ||
		envelope.EncryptionMode != BackupStorageEncryptionModeAESGCM ||
		envelope.KeyFingerprintSHA256 != key.fingerprint ||
		!validSHA256Hex(envelope.PayloadSHA256) {
		return nil, fmt.Errorf("%w: operator recovery journal envelope metadata mismatch", ErrInvalidBackupArtifact)
	}
	block, err := aes.NewCipher(key.key[:])
	if err != nil {
		return nil, fmt.Errorf("create operator recovery journal cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create operator recovery journal gcm: %w", err)
	}
	if len(envelope.Nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("%w: operator recovery journal nonce size is invalid", ErrInvalidBackupArtifact)
	}
	if strings.TrimSpace(aad) == "" {
		aad = OperatorRecoveryJournalSchemaID
	}
	body, err := gcm.Open(nil, envelope.Nonce, envelope.Ciphertext, []byte(aad))
	if err != nil {
		return nil, fmt.Errorf("%w: decrypt operator recovery journal payload: %v", ErrInvalidBackupArtifact, err)
	}
	if sha256Hex(body) != envelope.PayloadSHA256 {
		return nil, fmt.Errorf("%w: operator recovery journal payload digest mismatch", ErrInvalidBackupArtifact)
	}
	return body, nil
}

func (storage encryptedBackupStorage) BackupStorageEncryptionProof() BackupStorageEncryptionProof {
	return BackupStorageEncryptionProof{
		Mode:                 BackupStorageEncryptionModeAESGCM,
		EnvelopeSchemaID:     BackupArtifactEnvelopeSchemaID,
		KeyFingerprintSHA256: storage.key.fingerprint,
	}
}

func (storage encryptedBackupStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (BackupArtifactProof, error) {
	if err := ctx.Err(); err != nil {
		return BackupArtifactProof{}, err
	}
	normalizedKey, err := normalizeArtifactKey(key)
	if err != nil {
		return BackupArtifactProof{}, err
	}
	if len(body) == 0 {
		return BackupArtifactProof{}, fmt.Errorf("%w: artifact body is empty", ErrInvalidBackupArtifact)
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	block, err := aes.NewCipher(storage.key.key[:])
	if err != nil {
		return BackupArtifactProof{}, fmt.Errorf("create backup artifact cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return BackupArtifactProof{}, fmt.Errorf("create backup artifact gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return BackupArtifactProof{}, fmt.Errorf("generate backup artifact nonce: %w", err)
	}
	envelope := backupArtifactEnvelope{
		SchemaID:             BackupArtifactEnvelopeSchemaID,
		EncryptionMode:       BackupStorageEncryptionModeAESGCM,
		KeyFingerprintSHA256: storage.key.fingerprint,
		ArtifactKey:          normalizedKey,
		PlaintextContentType: contentType,
		NonceBase64:          base64.StdEncoding.EncodeToString(nonce),
		CiphertextBase64:     base64.StdEncoding.EncodeToString(gcm.Seal(nil, nonce, body, backupArtifactAAD(normalizedKey, contentType))),
	}
	envelopeBody, err := json.Marshal(envelope)
	if err != nil {
		return BackupArtifactProof{}, fmt.Errorf("encode backup artifact envelope: %w", err)
	}
	if _, err := storage.inner.WriteArtifact(ctx, normalizedKey, envelopeBody, "application/json"); err != nil {
		return BackupArtifactProof{}, err
	}
	return BackupArtifactProof{
		Key:         normalizedKey,
		SHA256:      sha256Hex(body),
		SizeBytes:   int64(len(body)),
		ContentType: contentType,
	}, nil
}

func (storage encryptedBackupStorage) ReadArtifact(ctx context.Context, key string, maxBytes int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if maxBytes <= 0 || maxBytes > (math.MaxInt64-65536)/2 {
		return nil, fmt.Errorf("%w: invalid artifact read bound", ErrInvalidBackupArtifact)
	}
	normalizedKey, err := normalizeArtifactKey(key)
	if err != nil {
		return nil, err
	}
	envelopeBody, err := storage.inner.ReadArtifact(ctx, normalizedKey, maxBytes*2+65536)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(envelopeBody))
	decoder.DisallowUnknownFields()
	var envelope backupArtifactEnvelope
	if err := decoder.Decode(&envelope); err != nil {
		return nil, fmt.Errorf("%w: decode backup artifact envelope: %v", ErrInvalidBackupArtifact, err)
	}
	if envelope.SchemaID != BackupArtifactEnvelopeSchemaID ||
		envelope.EncryptionMode != BackupStorageEncryptionModeAESGCM ||
		envelope.KeyFingerprintSHA256 != storage.key.fingerprint ||
		envelope.ArtifactKey != normalizedKey {
		return nil, fmt.Errorf("%w: backup artifact envelope metadata mismatch for %s", ErrInvalidBackupArtifact, normalizedKey)
	}
	nonce, err := base64.StdEncoding.DecodeString(envelope.NonceBase64)
	if err != nil {
		return nil, fmt.Errorf("%w: decode backup artifact envelope nonce for %s: %v", ErrInvalidBackupArtifact, normalizedKey, err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(envelope.CiphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("%w: decode backup artifact envelope ciphertext for %s: %v", ErrInvalidBackupArtifact, normalizedKey, err)
	}
	block, err := aes.NewCipher(storage.key.key[:])
	if err != nil {
		return nil, fmt.Errorf("create backup artifact cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create backup artifact gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, backupArtifactAAD(normalizedKey, envelope.PlaintextContentType))
	if err != nil {
		return nil, fmt.Errorf("%w: decrypt backup artifact envelope for %s: %v", ErrInvalidBackupArtifact, normalizedKey, err)
	}
	if int64(len(plaintext)) > maxBytes {
		return nil, fmt.Errorf("%w: backup artifact exceeds admitted size for %s", ErrInvalidBackupArtifact, normalizedKey)
	}
	return plaintext, nil
}

func (storage encryptedBackupStorage) Close() error {
	return CloseBackupStorage(storage.inner)
}

func backupStorageEncryptionProof(storage BackupStorage) (BackupStorageEncryptionProof, error) {
	reporter, ok := storage.(backupStorageEncryptionReporter)
	if !ok {
		return BackupStorageEncryptionProof{}, ErrEncryptedBackupStorage
	}
	proof := reporter.BackupStorageEncryptionProof()
	if proof.Mode != BackupStorageEncryptionModeAESGCM ||
		proof.EnvelopeSchemaID != BackupArtifactEnvelopeSchemaID ||
		!validSHA256Hex(proof.KeyFingerprintSHA256) {
		return BackupStorageEncryptionProof{}, ErrEncryptedBackupStorage
	}
	return proof, nil
}

func backupArtifactAAD(key string, contentType string) []byte {
	return []byte(BackupArtifactEnvelopeSchemaID + "\n" + key + "\n" + contentType)
}

func lookupRecoveryEnv(env map[string]string, key string) (string, bool) {
	if env != nil {
		value, ok := env[key]
		return value, ok
	}
	return os.LookupEnv(key)
}
