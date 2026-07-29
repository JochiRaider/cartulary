package recovery

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

const (
	BackupArtifactEnvelopeV2SchemaID = "cartulary.backup_artifact_envelope.v2"
	BackupArtifactEnvelopeV2KDF      = "hkdf_sha256_v1"
	BackupArtifactEnvelopeV2Cipher   = "aes_256_gcm_chunked_v1"

	BackupArtifactChunkPlaintextBytes   = 4 * 1024 * 1024
	backupArtifactEnvelopeV2SaltBytes   = 32
	backupArtifactEnvelopeV2PrefixBytes = 8
	backupArtifactEnvelopeV2TagBytes    = 16
	backupArtifactEnvelopeV2InfoPrefix  = "CARTULARY-BACKUP-ARTIFACT-ENVELOPE-V2\n"
)

var ErrStreamingBackupStorage = errors.New("recovery: streaming backup storage required")

// StoredStreamingBackupStorage is the persistence-side streaming capability.
// Its callback is published atomically or not at all, and opened artifacts are
// immutable for the lifetime of the returned reader.
type StoredStreamingBackupStorage interface {
	BackupStorage
	WriteStoredArtifact(
		ctx context.Context,
		key string,
		contentType string,
		write func(io.Writer) error,
	) (BackupArtifactProof, error)
	OpenStoredArtifact(ctx context.Context, key string) (io.ReadCloser, int64, error)
}

type BackupArtifactStreamWriteRequest struct {
	LogicalRef  string
	EnvelopeRef string
	ContentType string
	Plaintext   io.Reader
}

type BackupArtifactStreamProof struct {
	LogicalRef      string `json:"logical_ref"`
	ContentType     string `json:"content_type"`
	PlaintextBytes  int64  `json:"plaintext_bytes"`
	PlaintextSHA256 string `json:"plaintext_sha256"`
	EnvelopeRef     string `json:"envelope_ref"`
	EnvelopeSHA256  string `json:"envelope_sha256"`
	EnvelopeBytes   int64  `json:"-"`
}

// StreamingBackupStorage is the authenticated v2 capability beside the
// historical byte-slice v1 BackupStorage API. ReadArtifactStream performs a
// complete authentication preflight before writing to destination. The
// destination remains staging material until the method succeeds.
type StreamingBackupStorage interface {
	BackupStorage
	WriteArtifactStream(context.Context, BackupArtifactStreamWriteRequest) (BackupArtifactStreamProof, error)
	ReadArtifactStream(context.Context, BackupArtifactStreamProof, io.Writer) error
}

type streamingEncryptedBackupStorage struct {
	encryptedBackupStorage
	backend StoredStreamingBackupStorage
}

type countedReader struct {
	reader io.Reader
	count  int64
}

func (reader *countedReader) Read(body []byte) (int, error) {
	count, err := reader.reader.Read(body)
	reader.count += int64(count)
	return count, err
}

func RequireStreamingBackupStorage(storage BackupStorage) (StreamingBackupStorage, error) {
	streaming, ok := storage.(StreamingBackupStorage)
	if !ok {
		return nil, ErrStreamingBackupStorage
	}
	return streaming, nil
}

func (storage streamingEncryptedBackupStorage) WriteArtifactStream(
	ctx context.Context,
	request BackupArtifactStreamWriteRequest,
) (BackupArtifactStreamProof, error) {
	if err := ctx.Err(); err != nil {
		return BackupArtifactStreamProof{}, err
	}
	logicalRef, envelopeRef, contentType, err := validateBackupArtifactStreamWriteRequest(request)
	if err != nil {
		return BackupArtifactStreamProof{}, err
	}
	salt := make([]byte, backupArtifactEnvelopeV2SaltBytes)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return BackupArtifactStreamProof{}, fmt.Errorf("generate backup artifact envelope salt: %w", err)
	}
	noncePrefix := make([]byte, backupArtifactEnvelopeV2PrefixBytes)
	if _, err := io.ReadFull(rand.Reader, noncePrefix); err != nil {
		return BackupArtifactStreamProof{}, fmt.Errorf("generate backup artifact envelope nonce prefix: %w", err)
	}
	artifactKey, err := deriveBackupArtifactEnvelopeV2Key(storage.key, salt, logicalRef)
	if err != nil {
		return BackupArtifactStreamProof{}, err
	}

	var plaintextBytes int64
	var plaintextSHA256 string
	storedProof, err := storage.backend.WriteStoredArtifact(
		ctx,
		envelopeRef,
		"application/json",
		func(writer io.Writer) error {
			var writeErr error
			plaintextBytes, plaintextSHA256, writeErr = writeBackupArtifactEnvelopeV2(
				ctx,
				writer,
				request.Plaintext,
				logicalRef,
				contentType,
				artifactKey,
				salt,
				noncePrefix,
			)
			return writeErr
		},
	)
	if err != nil {
		return BackupArtifactStreamProof{}, fmt.Errorf("write streaming backup artifact envelope: %w", err)
	}
	if storedProof.Key != envelopeRef ||
		storedProof.SizeBytes <= 0 ||
		!validSHA256Hex(storedProof.SHA256) ||
		storedProof.ContentType != "application/json" {
		return BackupArtifactStreamProof{}, fmt.Errorf(
			"%w: stored backup artifact envelope proof is invalid",
			ErrInvalidBackupArtifact,
		)
	}
	return BackupArtifactStreamProof{
		LogicalRef:      logicalRef,
		ContentType:     contentType,
		PlaintextBytes:  plaintextBytes,
		PlaintextSHA256: plaintextSHA256,
		EnvelopeRef:     envelopeRef,
		EnvelopeSHA256:  storedProof.SHA256,
		EnvelopeBytes:   storedProof.SizeBytes,
	}, nil
}

func (storage streamingEncryptedBackupStorage) ReadArtifactStream(
	ctx context.Context,
	proof BackupArtifactStreamProof,
	destination io.Writer,
) error {
	if destination == nil {
		return fmt.Errorf("%w: streaming artifact destination is required", ErrInvalidBackupArtifact)
	}
	normalized, err := validateBackupArtifactStreamProof(proof)
	if err != nil {
		return err
	}
	if err := storage.processBackupArtifactEnvelopeV2(ctx, normalized, io.Discard); err != nil {
		return err
	}
	if err := storage.processBackupArtifactEnvelopeV2(ctx, normalized, destination); err != nil {
		return err
	}
	return nil
}

func (storage streamingEncryptedBackupStorage) processBackupArtifactEnvelopeV2(
	ctx context.Context,
	proof BackupArtifactStreamProof,
	destination io.Writer,
) (resultErr error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	reader, storedBytes, err := storage.backend.OpenStoredArtifact(ctx, proof.EnvelopeRef)
	if err != nil {
		return fmt.Errorf("open streaming backup artifact envelope: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); resultErr == nil && closeErr != nil {
			resultErr = fmt.Errorf("close streaming backup artifact envelope: %w", closeErr)
		}
	}()
	if storedBytes <= 0 {
		return fmt.Errorf("%w: streaming envelope is empty", ErrInvalidBackupArtifact)
	}
	if proof.EnvelopeBytes > 0 && storedBytes != proof.EnvelopeBytes {
		return fmt.Errorf("%w: streaming envelope size mismatch", ErrInvalidBackupArtifact)
	}
	maxEnvelopeBytes, err := maximumBackupArtifactEnvelopeV2Bytes(proof.PlaintextBytes)
	if err != nil {
		return err
	}
	if uint64(storedBytes) > maxEnvelopeBytes {
		return fmt.Errorf("%w: streaming envelope exceeds its plaintext-derived bound", ErrInvalidBackupArtifact)
	}

	hasher := sha256.New()
	counted := &countedReader{reader: reader}
	envelopeReader := io.TeeReader(counted, hasher)
	if err := decryptBackupArtifactEnvelopeV2(
		ctx,
		envelopeReader,
		storage.key,
		proof,
		destination,
	); err != nil {
		return err
	}
	if counted.count != storedBytes {
		return fmt.Errorf("%w: streaming envelope size changed during read", ErrInvalidBackupArtifact)
	}
	if hex.EncodeToString(hasher.Sum(nil)) != proof.EnvelopeSHA256 {
		return fmt.Errorf("%w: streaming envelope digest mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func validateBackupArtifactStreamWriteRequest(
	request BackupArtifactStreamWriteRequest,
) (string, string, string, error) {
	if request.Plaintext == nil {
		return "", "", "", fmt.Errorf("%w: streaming artifact plaintext is required", ErrInvalidBackupArtifact)
	}
	logicalRef, err := validateBackupLogicalRef(request.LogicalRef)
	if err != nil {
		return "", "", "", err
	}
	envelopeRef, err := validateBackupLogicalRef(request.EnvelopeRef)
	if err != nil {
		return "", "", "", err
	}
	contentType, err := validateBackupContentType(request.ContentType)
	if err != nil {
		return "", "", "", err
	}
	return logicalRef, envelopeRef, contentType, nil
}

func validateBackupArtifactStreamProof(
	proof BackupArtifactStreamProof,
) (BackupArtifactStreamProof, error) {
	logicalRef, err := validateBackupLogicalRef(proof.LogicalRef)
	if err != nil {
		return BackupArtifactStreamProof{}, err
	}
	envelopeRef, err := validateBackupLogicalRef(proof.EnvelopeRef)
	if err != nil {
		return BackupArtifactStreamProof{}, err
	}
	contentType, err := validateBackupContentType(proof.ContentType)
	if err != nil {
		return BackupArtifactStreamProof{}, err
	}
	if proof.PlaintextBytes < 0 ||
		!validSHA256Hex(proof.PlaintextSHA256) ||
		!validSHA256Hex(proof.EnvelopeSHA256) ||
		proof.EnvelopeBytes < 0 {
		return BackupArtifactStreamProof{}, fmt.Errorf("%w: streaming artifact proof is invalid", ErrInvalidBackupArtifact)
	}
	if _, err := expectedBackupArtifactEnvelopeV2Chunks(proof.PlaintextBytes); err != nil {
		return BackupArtifactStreamProof{}, err
	}
	proof.LogicalRef = logicalRef
	proof.EnvelopeRef = envelopeRef
	proof.ContentType = contentType
	return proof, nil
}

func validateBackupLogicalRef(value string) (string, error) {
	if value == "" || len(value) > 512 {
		return "", fmt.Errorf("%w: logical reference is invalid", ErrInvalidBackupArtifact)
	}
	if !isASCIIAlphaNumeric(value[0]) {
		return "", fmt.Errorf("%w: logical reference is invalid", ErrInvalidBackupArtifact)
	}
	for index := range len(value) {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') &&
			character != '_' &&
			character != '.' &&
			character != ':' &&
			character != '/' &&
			character != '-' {
			return "", fmt.Errorf("%w: logical reference is invalid", ErrInvalidBackupArtifact)
		}
	}
	if _, err := normalizeArtifactKey(value); err != nil {
		return "", err
	}
	return value, nil
}

func validateBackupContentType(value string) (string, error) {
	if value == "" || len(value) > 128 {
		return "", fmt.Errorf("%w: content type is invalid", ErrInvalidBackupArtifact)
	}
	if !isASCIIAlphaNumeric(value[0]) {
		return "", fmt.Errorf("%w: content type is invalid", ErrInvalidBackupArtifact)
	}
	for index := range len(value) {
		character := value[index]
		if (character < 'a' || character > 'z') &&
			(character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') &&
			!strings.ContainsRune("!#$&^_.+/-", rune(character)) {
			return "", fmt.Errorf("%w: content type is invalid", ErrInvalidBackupArtifact)
		}
	}
	return value, nil
}

func deriveBackupArtifactEnvelopeV2Key(
	masterKey RecoveryEncryptionKey,
	salt []byte,
	logicalRef string,
) ([32]byte, error) {
	if masterKey.fingerprint == "" {
		return [32]byte{}, ErrRecoveryMasterKeyRequired
	}
	if len(salt) != backupArtifactEnvelopeV2SaltBytes {
		return [32]byte{}, fmt.Errorf("%w: streaming envelope salt size is invalid", ErrInvalidBackupArtifact)
	}
	derived, err := hkdf.Key(
		sha256.New,
		masterKey.key[:],
		salt,
		backupArtifactEnvelopeV2InfoPrefix+logicalRef,
		32,
	)
	if err != nil {
		return [32]byte{}, fmt.Errorf("derive backup artifact envelope key: %w", err)
	}
	var key [32]byte
	copy(key[:], derived)
	return key, nil
}

func writeBackupArtifactEnvelopeV2(
	ctx context.Context,
	writer io.Writer,
	plaintext io.Reader,
	logicalRef string,
	contentType string,
	artifactKey [32]byte,
	salt []byte,
	noncePrefix []byte,
) (int64, string, error) {
	block, err := aes.NewCipher(artifactKey[:])
	if err != nil {
		return 0, "", fmt.Errorf("create streaming backup artifact cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, "", fmt.Errorf("create streaming backup artifact gcm: %w", err)
	}
	if gcm.NonceSize() != backupArtifactEnvelopeV2PrefixBytes+4 ||
		gcm.Overhead() != backupArtifactEnvelopeV2TagBytes {
		return 0, "", fmt.Errorf("%w: unsupported streaming cipher parameters", ErrInvalidBackupArtifact)
	}

	if err := writeBackupArtifactEnvelopeV2Header(writer, logicalRef, contentType, salt, noncePrefix); err != nil {
		return 0, "", err
	}
	buffered := bufio.NewReaderSize(plaintext, 64*1024)
	chunk := make([]byte, BackupArtifactChunkPlaintextBytes)
	plaintextHasher := sha256.New()
	var plaintextBytes int64
	var index uint64
	firstChunk := true
	for {
		if err := ctx.Err(); err != nil {
			return 0, "", err
		}
		count, readErr := io.ReadFull(buffered, chunk)
		final := false
		switch {
		case readErr == nil:
			_, peekErr := buffered.Peek(1)
			switch {
			case peekErr == nil:
			case errors.Is(peekErr, io.EOF):
				final = true
			default:
				return 0, "", fmt.Errorf("look ahead streaming backup artifact plaintext: %w", peekErr)
			}
		case errors.Is(readErr, io.ErrUnexpectedEOF):
			final = true
		case errors.Is(readErr, io.EOF) && count == 0:
			if !firstChunk {
				return 0, "", fmt.Errorf("%w: streaming plaintext ended after a non-final chunk", ErrInvalidBackupArtifact)
			}
			final = true
		default:
			return 0, "", fmt.Errorf("read streaming backup artifact plaintext: %w", readErr)
		}
		if index > math.MaxUint32 {
			return 0, "", fmt.Errorf("%w: streaming artifact has too many chunks", ErrInvalidBackupArtifact)
		}
		if !firstChunk {
			if _, err := io.WriteString(writer, ","); err != nil {
				return 0, "", fmt.Errorf("write streaming backup artifact chunk separator: %w", err)
			}
		}
		plaintextChunk := chunk[:count]
		if _, err := plaintextHasher.Write(plaintextChunk); err != nil {
			return 0, "", fmt.Errorf("hash streaming backup artifact plaintext: %w", err)
		}
		if int64(count) > math.MaxInt64-plaintextBytes {
			return 0, "", fmt.Errorf("%w: streaming artifact size overflows", ErrInvalidBackupArtifact)
		}
		plaintextBytes += int64(count)
		nonce := backupArtifactEnvelopeV2Nonce(noncePrefix, uint32(index))
		aad := backupArtifactEnvelopeV2AAD(logicalRef, contentType, uint32(index), count, final)
		ciphertext := gcm.Seal(nil, nonce, plaintextChunk, aad)
		if err := writeBackupArtifactEnvelopeV2Chunk(
			writer,
			uint32(index),
			count,
			final,
			ciphertext,
		); err != nil {
			return 0, "", err
		}
		firstChunk = false
		if final {
			break
		}
		index++
	}
	if _, err := io.WriteString(writer, "]}"); err != nil {
		return 0, "", fmt.Errorf("finish streaming backup artifact envelope: %w", err)
	}
	return plaintextBytes, hex.EncodeToString(plaintextHasher.Sum(nil)), nil
}

func writeBackupArtifactEnvelopeV2Header(
	writer io.Writer,
	logicalRef string,
	contentType string,
	salt []byte,
	noncePrefix []byte,
) error {
	logicalRefJSON, _ := json.Marshal(logicalRef)
	contentTypeJSON, _ := json.Marshal(contentType)
	header := `{"schema_id":"` + BackupArtifactEnvelopeV2SchemaID +
		`","logical_ref":` + string(logicalRefJSON) +
		`,"content_type":` + string(contentTypeJSON) +
		`,"kdf":"` + BackupArtifactEnvelopeV2KDF +
		`","cipher":"` + BackupArtifactEnvelopeV2Cipher +
		`","salt_base64":"` + base64.StdEncoding.EncodeToString(salt) +
		`","nonce_prefix_base64":"` + base64.StdEncoding.EncodeToString(noncePrefix) +
		`","chunk_plaintext_bytes":` + strconv.Itoa(BackupArtifactChunkPlaintextBytes) +
		`,"chunks":[`
	if _, err := io.WriteString(writer, header); err != nil {
		return fmt.Errorf("write streaming backup artifact envelope header: %w", err)
	}
	return nil
}

func writeBackupArtifactEnvelopeV2Chunk(
	writer io.Writer,
	index uint32,
	plaintextLength int,
	final bool,
	ciphertext []byte,
) error {
	chunkHeader := `{"index":` + strconv.FormatUint(uint64(index), 10) +
		`,"plaintext_length":` + strconv.Itoa(plaintextLength) +
		`,"final":` + strconv.FormatBool(final) +
		`,"ciphertext_base64":"`
	if _, err := io.WriteString(writer, chunkHeader); err != nil {
		return fmt.Errorf("write streaming backup artifact chunk header: %w", err)
	}
	encoder := base64.NewEncoder(base64.StdEncoding, writer)
	if _, err := encoder.Write(ciphertext); err != nil {
		_ = encoder.Close()
		return fmt.Errorf("write streaming backup artifact chunk ciphertext: %w", err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("finish streaming backup artifact chunk ciphertext: %w", err)
	}
	if _, err := io.WriteString(writer, `"}`); err != nil {
		return fmt.Errorf("finish streaming backup artifact chunk: %w", err)
	}
	return nil
}

func decryptBackupArtifactEnvelopeV2(
	ctx context.Context,
	reader io.Reader,
	masterKey RecoveryEncryptionKey,
	proof BackupArtifactStreamProof,
	destination io.Writer,
) error {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := expectJSONDelimiter(decoder, '{', "envelope object"); err != nil {
		return err
	}
	var schemaID string
	if err := decodeExactJSONField(decoder, "schema_id", &schemaID); err != nil {
		return err
	}
	if schemaID != BackupArtifactEnvelopeV2SchemaID {
		return fmt.Errorf("%w: streaming envelope schema is invalid", ErrInvalidBackupArtifact)
	}
	var logicalRef string
	if err := decodeExactJSONField(decoder, "logical_ref", &logicalRef); err != nil {
		return err
	}
	if logicalRef != proof.LogicalRef {
		return fmt.Errorf("%w: streaming envelope logical reference mismatch", ErrInvalidBackupArtifact)
	}
	var contentType string
	if err := decodeExactJSONField(decoder, "content_type", &contentType); err != nil {
		return err
	}
	if contentType != proof.ContentType {
		return fmt.Errorf("%w: streaming envelope content type mismatch", ErrInvalidBackupArtifact)
	}
	var kdfID string
	if err := decodeExactJSONField(decoder, "kdf", &kdfID); err != nil {
		return err
	}
	if kdfID != BackupArtifactEnvelopeV2KDF {
		return fmt.Errorf("%w: streaming envelope KDF is invalid", ErrInvalidBackupArtifact)
	}
	var cipherID string
	if err := decodeExactJSONField(decoder, "cipher", &cipherID); err != nil {
		return err
	}
	if cipherID != BackupArtifactEnvelopeV2Cipher {
		return fmt.Errorf("%w: streaming envelope cipher is invalid", ErrInvalidBackupArtifact)
	}
	var saltBase64 string
	if err := decodeExactJSONField(decoder, "salt_base64", &saltBase64); err != nil {
		return err
	}
	salt, err := decodeExactBase64(saltBase64, backupArtifactEnvelopeV2SaltBytes, "salt")
	if err != nil {
		return err
	}
	var noncePrefixBase64 string
	if err := decodeExactJSONField(decoder, "nonce_prefix_base64", &noncePrefixBase64); err != nil {
		return err
	}
	noncePrefix, err := decodeExactBase64(
		noncePrefixBase64,
		backupArtifactEnvelopeV2PrefixBytes,
		"nonce prefix",
	)
	if err != nil {
		return err
	}
	var chunkPlaintextBytes json.Number
	if err := decodeExactJSONField(decoder, "chunk_plaintext_bytes", &chunkPlaintextBytes); err != nil {
		return err
	}
	if chunkPlaintextBytes.String() != strconv.Itoa(BackupArtifactChunkPlaintextBytes) {
		return fmt.Errorf("%w: streaming envelope chunk size is invalid", ErrInvalidBackupArtifact)
	}
	if err := expectJSONFieldName(decoder, "chunks"); err != nil {
		return err
	}
	if err := expectJSONDelimiter(decoder, '[', "chunks array"); err != nil {
		return err
	}

	artifactKey, err := deriveBackupArtifactEnvelopeV2Key(masterKey, salt, logicalRef)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(artifactKey[:])
	if err != nil {
		return fmt.Errorf("create streaming backup artifact cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("create streaming backup artifact gcm: %w", err)
	}
	expectedChunks, err := expectedBackupArtifactEnvelopeV2Chunks(proof.PlaintextBytes)
	if err != nil {
		return err
	}
	plaintextHasher := sha256.New()
	var plaintextBytes int64
	for index := uint64(0); index < expectedChunks; index++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !decoder.More() {
			return fmt.Errorf("%w: streaming envelope is truncated", ErrInvalidBackupArtifact)
		}
		chunkIndex, plaintextLength, final, ciphertextBase64, err := decodeBackupArtifactEnvelopeV2Chunk(decoder)
		if err != nil {
			return err
		}
		expectedLength := expectedBackupArtifactEnvelopeV2ChunkLength(
			proof.PlaintextBytes,
			expectedChunks,
			index,
		)
		expectedFinal := index == expectedChunks-1
		if chunkIndex != index ||
			plaintextLength != expectedLength ||
			final != expectedFinal {
			return fmt.Errorf("%w: streaming envelope chunk sequence is invalid", ErrInvalidBackupArtifact)
		}
		expectedCiphertextBytes := plaintextLength + gcm.Overhead()
		if len(ciphertextBase64) != base64.StdEncoding.EncodedLen(expectedCiphertextBytes) {
			return fmt.Errorf("%w: streaming envelope ciphertext length is invalid", ErrInvalidBackupArtifact)
		}
		ciphertext, err := base64.StdEncoding.Strict().DecodeString(ciphertextBase64)
		if err != nil || len(ciphertext) != expectedCiphertextBytes {
			return fmt.Errorf("%w: streaming envelope ciphertext encoding is invalid", ErrInvalidBackupArtifact)
		}
		nonce := backupArtifactEnvelopeV2Nonce(noncePrefix, uint32(index))
		aad := backupArtifactEnvelopeV2AAD(
			logicalRef,
			contentType,
			uint32(index),
			plaintextLength,
			final,
		)
		plaintext, err := gcm.Open(nil, nonce, ciphertext, aad)
		if err != nil {
			return fmt.Errorf("%w: authenticate streaming envelope chunk", ErrInvalidBackupArtifact)
		}
		if _, err := plaintextHasher.Write(plaintext); err != nil {
			return fmt.Errorf("hash streaming backup artifact plaintext: %w", err)
		}
		if len(plaintext) > 0 {
			written, err := destination.Write(plaintext)
			if err != nil {
				return fmt.Errorf("write authenticated streaming backup artifact: %w", err)
			}
			if written != len(plaintext) {
				return fmt.Errorf("write authenticated streaming backup artifact: %w", io.ErrShortWrite)
			}
		}
		plaintextBytes += int64(len(plaintext))
	}
	if decoder.More() {
		return fmt.Errorf("%w: streaming envelope has duplicate or trailing chunks", ErrInvalidBackupArtifact)
	}
	if err := expectJSONDelimiter(decoder, ']', "chunks array"); err != nil {
		return err
	}
	if decoder.More() {
		return fmt.Errorf("%w: streaming envelope has unknown or duplicate members", ErrInvalidBackupArtifact)
	}
	if err := expectJSONDelimiter(decoder, '}', "envelope object"); err != nil {
		return err
	}
	if err := rejectBackupArtifactEnvelopeV2TrailingData(decoder, reader); err != nil {
		return err
	}
	if plaintextBytes != proof.PlaintextBytes {
		return fmt.Errorf("%w: streaming artifact plaintext size mismatch", ErrInvalidBackupArtifact)
	}
	if hex.EncodeToString(plaintextHasher.Sum(nil)) != proof.PlaintextSHA256 {
		return fmt.Errorf("%w: streaming artifact plaintext digest mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func decodeBackupArtifactEnvelopeV2Chunk(
	decoder *json.Decoder,
) (uint64, int, bool, string, error) {
	if err := expectJSONDelimiter(decoder, '{', "chunk object"); err != nil {
		return 0, 0, false, "", err
	}
	var index json.Number
	if err := decodeExactJSONField(decoder, "index", &index); err != nil {
		return 0, 0, false, "", err
	}
	indexValue, err := strconv.ParseUint(index.String(), 10, 32)
	if err != nil {
		return 0, 0, false, "", fmt.Errorf("%w: streaming envelope chunk index is invalid", ErrInvalidBackupArtifact)
	}
	var plaintextLength json.Number
	if err := decodeExactJSONField(decoder, "plaintext_length", &plaintextLength); err != nil {
		return 0, 0, false, "", err
	}
	lengthValue, err := strconv.ParseUint(plaintextLength.String(), 10, 32)
	if err != nil || lengthValue > BackupArtifactChunkPlaintextBytes {
		return 0, 0, false, "", fmt.Errorf("%w: streaming envelope plaintext length is invalid", ErrInvalidBackupArtifact)
	}
	var final bool
	if err := decodeExactJSONField(decoder, "final", &final); err != nil {
		return 0, 0, false, "", err
	}
	var ciphertextBase64 string
	if err := decodeExactJSONField(decoder, "ciphertext_base64", &ciphertextBase64); err != nil {
		return 0, 0, false, "", err
	}
	if decoder.More() {
		return 0, 0, false, "", fmt.Errorf(
			"%w: streaming envelope chunk has unknown or duplicate members",
			ErrInvalidBackupArtifact,
		)
	}
	if err := expectJSONDelimiter(decoder, '}', "chunk object"); err != nil {
		return 0, 0, false, "", err
	}
	return indexValue, int(lengthValue), final, ciphertextBase64, nil
}

func decodeExactJSONField(decoder *json.Decoder, name string, destination any) error {
	if err := expectJSONFieldName(decoder, name); err != nil {
		return err
	}
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("%w: decode streaming envelope member %s", ErrInvalidBackupArtifact, name)
	}
	return nil
}

func expectJSONFieldName(decoder *json.Decoder, expected string) error {
	if !decoder.More() {
		return fmt.Errorf("%w: streaming envelope member %s is missing", ErrInvalidBackupArtifact, expected)
	}
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: decode streaming envelope member name", ErrInvalidBackupArtifact)
	}
	name, ok := token.(string)
	if !ok || name != expected {
		return fmt.Errorf(
			"%w: streaming envelope member order or identity is invalid; expected %s",
			ErrInvalidBackupArtifact,
			expected,
		)
	}
	return nil
}

func expectJSONDelimiter(decoder *json.Decoder, expected json.Delim, label string) error {
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("%w: decode streaming envelope %s", ErrInvalidBackupArtifact, label)
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != expected {
		return fmt.Errorf("%w: streaming envelope %s is invalid", ErrInvalidBackupArtifact, label)
	}
	return nil
}

func rejectBackupArtifactEnvelopeV2TrailingData(decoder *json.Decoder, reader io.Reader) error {
	var trailing [1]byte
	if count, _ := decoder.Buffered().Read(trailing[:]); count != 0 {
		return fmt.Errorf("%w: streaming envelope has trailing data", ErrInvalidBackupArtifact)
	}
	if count, err := reader.Read(trailing[:]); count != 0 || (err != nil && !errors.Is(err, io.EOF)) {
		return fmt.Errorf("%w: streaming envelope has trailing data", ErrInvalidBackupArtifact)
	}
	return nil
}

func decodeExactBase64(encoded string, expectedBytes int, label string) ([]byte, error) {
	if len(encoded) != base64.StdEncoding.EncodedLen(expectedBytes) {
		return nil, fmt.Errorf("%w: streaming envelope %s encoding is invalid", ErrInvalidBackupArtifact, label)
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(encoded)
	if err != nil || len(decoded) != expectedBytes {
		return nil, fmt.Errorf("%w: streaming envelope %s encoding is invalid", ErrInvalidBackupArtifact, label)
	}
	return decoded, nil
}

func isASCIIAlphaNumeric(value byte) bool {
	return (value >= 'a' && value <= 'z') ||
		(value >= 'A' && value <= 'Z') ||
		(value >= '0' && value <= '9')
}

func backupArtifactEnvelopeV2Nonce(prefix []byte, index uint32) []byte {
	nonce := make([]byte, backupArtifactEnvelopeV2PrefixBytes+4)
	copy(nonce, prefix)
	binary.BigEndian.PutUint32(nonce[backupArtifactEnvelopeV2PrefixBytes:], index)
	return nonce
}

func backupArtifactEnvelopeV2AAD(
	logicalRef string,
	contentType string,
	index uint32,
	plaintextLength int,
	final bool,
) []byte {
	result := make([]byte, 0, 256+len(logicalRef)+len(contentType))
	result = append(result, BackupArtifactEnvelopeV2SchemaID...)
	result = append(result, '\n')
	result = append(result, logicalRef...)
	result = append(result, '\n')
	result = append(result, contentType...)
	result = append(result, '\n')
	result = strconv.AppendUint(result, uint64(index), 10)
	result = append(result, '\n')
	result = strconv.AppendInt(result, int64(plaintextLength), 10)
	result = append(result, '\n')
	result = strconv.AppendBool(result, final)
	return result
}

func expectedBackupArtifactEnvelopeV2Chunks(plaintextBytes int64) (uint64, error) {
	if plaintextBytes < 0 {
		return 0, fmt.Errorf("%w: streaming artifact size is invalid", ErrInvalidBackupArtifact)
	}
	if plaintextBytes == 0 {
		return 1, nil
	}
	chunks := (uint64(plaintextBytes) + BackupArtifactChunkPlaintextBytes - 1) /
		BackupArtifactChunkPlaintextBytes
	if chunks > uint64(math.MaxUint32)+1 {
		return 0, fmt.Errorf("%w: streaming artifact has too many chunks", ErrInvalidBackupArtifact)
	}
	return chunks, nil
}

func expectedBackupArtifactEnvelopeV2ChunkLength(
	plaintextBytes int64,
	chunks uint64,
	index uint64,
) int {
	if plaintextBytes == 0 {
		return 0
	}
	if index < chunks-1 {
		return BackupArtifactChunkPlaintextBytes
	}
	remainder := int(uint64(plaintextBytes) % BackupArtifactChunkPlaintextBytes)
	if remainder == 0 {
		return BackupArtifactChunkPlaintextBytes
	}
	return remainder
}

func maximumBackupArtifactEnvelopeV2Bytes(plaintextBytes int64) (uint64, error) {
	chunks, err := expectedBackupArtifactEnvelopeV2Chunks(plaintextBytes)
	if err != nil {
		return 0, err
	}
	fullChunkEncodedBytes := uint64(base64.StdEncoding.EncodedLen(
		BackupArtifactChunkPlaintextBytes + backupArtifactEnvelopeV2TagBytes,
	))
	fullChunks := chunks - 1
	if fullChunks > math.MaxUint64/fullChunkEncodedBytes {
		return 0, fmt.Errorf("%w: streaming envelope bound overflows", ErrInvalidBackupArtifact)
	}
	encodedCiphertextBytes := fullChunks * fullChunkEncodedBytes
	lastPlaintextBytes := expectedBackupArtifactEnvelopeV2ChunkLength(
		plaintextBytes,
		chunks,
		chunks-1,
	)
	lastEncodedBytes := uint64(base64.StdEncoding.EncodedLen(
		lastPlaintextBytes + backupArtifactEnvelopeV2TagBytes,
	))
	if encodedCiphertextBytes > math.MaxUint64-lastEncodedBytes {
		return 0, fmt.Errorf("%w: streaming envelope bound overflows", ErrInvalidBackupArtifact)
	}
	encodedCiphertextBytes += lastEncodedBytes
	const envelopeFixedBound = uint64(4096)
	const chunkMetadataBound = uint64(192)
	if chunks > (math.MaxUint64-envelopeFixedBound-encodedCiphertextBytes)/chunkMetadataBound {
		return 0, fmt.Errorf("%w: streaming envelope bound overflows", ErrInvalidBackupArtifact)
	}
	return envelopeFixedBound + encodedCiphertextBytes + chunks*chunkMetadataBound, nil
}
