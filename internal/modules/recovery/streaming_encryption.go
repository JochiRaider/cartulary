package recovery

import (
	"bufio"
	"context"
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

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
)

const (
	BackupArtifactEnvelopeV3SchemaID = "cartulary.backup_artifact_envelope.v3"
	BackupArtifactEnvelopeV3KDF      = "hkdf_sha256_v2"
	BackupArtifactEnvelopeV3Cipher   = "aes_256_gcm_random_nonce_chunked_v2"

	BackupArtifactChunkPlaintextBytes = 4 * 1024 * 1024
	backupArtifactEnvelopeV3SaltBytes = cryptography.SaltBytes
	backupArtifactEnvelopeV3Overhead  = 12 + 16
	BackupArtifactMaximumChunks       = cryptography.MaxArtifactSeals
)

var ErrStreamingBackupStorage = errors.New("recovery: streaming backup storage required")

// StoredStreamingBackupStorage is the persistence-side streaming capability.
// Its write callback is published atomically or not at all. Opened input may be
// mutable: authentication must complete into private staging before exposure.
type StoredStreamingBackupStorage interface {
	BackupStorage
	WriteStoredArtifact(
		ctx context.Context,
		key string,
		contentType string,
		write func(io.Writer) error,
	) (BackupArtifactProof, error)
	OpenStoredArtifact(ctx context.Context, key string) (io.ReadCloser, int64, error)
	// StageArtifact bounds private, unpublished output and returns its immutable
	// read capability only after the callback succeeds. Close discards the stage.
	StageArtifact(ctx context.Context, maxBytes int64, write func(io.Writer) error) (io.ReadCloser, error)
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

// StreamingBackupStorage and BackupStorage share the sole current v3 codec. ReadArtifactStream performs a
// complete authentication preflight before writing to destination. The
// destination remains staging material until the method succeeds.
type StreamingBackupStorage interface {
	BackupStorage
	WriteArtifactStream(context.Context, BackupArtifactStreamWriteRequest) (BackupArtifactStreamProof, error)
	ReadArtifactStream(context.Context, BackupArtifactStreamProof, io.Writer) error
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

func (storage encryptedBackupStorage) WriteArtifactStream(
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
	artifactKey, err := storage.key.key.NewArtifactSealer(BackupArtifactEnvelopeV3SchemaID, ApplicationCryptoFormatID, logicalRef, contentType)
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
			plaintextBytes, plaintextSHA256, writeErr = writeBackupArtifactEnvelopeV3(
				ctx,
				writer,
				request.Plaintext,
				logicalRef,
				contentType,
				artifactKey,
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

func (storage encryptedBackupStorage) ReadArtifactStream(
	ctx context.Context,
	proof BackupArtifactStreamProof,
	destination io.Writer,
) (resultErr error) {
	if destination == nil {
		return fmt.Errorf("%w: streaming artifact destination is required", ErrInvalidBackupArtifact)
	}
	normalized, err := validateBackupArtifactStreamProof(proof)
	if err != nil {
		return err
	}
	stage, err := storage.backend.StageArtifact(ctx, max(1, normalized.PlaintextBytes), func(writer io.Writer) error {
		return storage.processBackupArtifactEnvelopeV3(ctx, normalized, writer)
	})
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, stage.Close()) }()
	_, err = io.CopyBuffer(contextStreamWriter{ctx: ctx, writer: destination}, stage, make([]byte, 64*1024))
	return err
}

type contextStreamWriter struct {
	ctx    context.Context
	writer io.Writer
}

func (writer contextStreamWriter) Write(body []byte) (int, error) {
	if err := writer.ctx.Err(); err != nil {
		return 0, err
	}
	return writer.writer.Write(body)
}

func (storage encryptedBackupStorage) ResolveObjectProof(
	ctx context.Context,
	object VNextObjectManifestEntry,
) (result BackupArtifactStreamProof, resultErr error) {
	envelopeRef := object.ArtifactRef + ".envelope.json"
	reader, storedBytes, err := storage.backend.OpenStoredArtifact(ctx, envelopeRef)
	if err != nil {
		return BackupArtifactStreamProof{}, fmt.Errorf("resolve object envelope proof: %w", err)
	}
	defer func() {
		if closeErr := reader.Close(); resultErr == nil && closeErr != nil {
			resultErr = fmt.Errorf("close object envelope proof: %w", closeErr)
		}
	}()
	hasher := sha256.New()
	copied, err := io.Copy(hasher, reader)
	if err != nil {
		return BackupArtifactStreamProof{}, fmt.Errorf("digest object envelope proof: %w", err)
	}
	if copied != storedBytes || storedBytes <= 0 {
		return BackupArtifactStreamProof{}, fmt.Errorf(
			"%w: object envelope size changed while resolving proof",
			ErrInvalidBackupArtifact,
		)
	}
	return BackupArtifactStreamProof{
		LogicalRef:      object.ArtifactRef,
		ContentType:     object.ContentType,
		PlaintextBytes:  object.PlaintextBytes,
		PlaintextSHA256: object.PlaintextSHA256,
		EnvelopeRef:     envelopeRef,
		EnvelopeSHA256:  hex.EncodeToString(hasher.Sum(nil)),
		EnvelopeBytes:   storedBytes,
	}, nil
}

func (storage encryptedBackupStorage) processBackupArtifactEnvelopeV3(
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
	maxEnvelopeBytes, err := maximumBackupArtifactEnvelopeV3Bytes(proof.PlaintextBytes)
	if err != nil {
		return err
	}
	if uint64(storedBytes) > maxEnvelopeBytes {
		return fmt.Errorf("%w: streaming envelope exceeds its plaintext-derived bound", ErrInvalidBackupArtifact)
	}

	hasher := sha256.New()
	counted := &countedReader{reader: reader}
	envelopeReader := io.TeeReader(counted, hasher)
	plaintext, err := decryptBackupArtifactEnvelopeV3(ctx, envelopeReader, storage.key, proof.LogicalRef, proof.ContentType, proof.PlaintextBytes, destination)
	if err != nil {
		return err
	}
	if plaintext.SizeBytes != proof.PlaintextBytes || plaintext.SHA256 != proof.PlaintextSHA256 {
		return fmt.Errorf("%w: streaming plaintext proof mismatch", ErrInvalidBackupArtifact)
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
	if _, err := expectedBackupArtifactEnvelopeV3Chunks(proof.PlaintextBytes); err != nil {
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

func writeBackupArtifactEnvelopeV3(
	ctx context.Context,
	writer io.Writer,
	plaintext io.Reader,
	logicalRef string,
	contentType string,
	artifactKey cryptography.ArtifactSealer,
) (int64, string, error) {
	if err := writeBackupArtifactEnvelopeV3Header(writer, logicalRef, contentType, artifactKey.Salt()); err != nil {
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
		if index >= BackupArtifactMaximumChunks {
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
		aad := backupArtifactEnvelopeV3AAD(logicalRef, contentType, uint32(index), count, final)
		ciphertext, err := artifactKey.Seal(plaintextChunk, aad)
		if err != nil {
			return 0, "", err
		}
		if err := writeBackupArtifactEnvelopeV3Chunk(
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

func writeBackupArtifactEnvelopeV3Header(
	writer io.Writer,
	logicalRef string,
	contentType string,
	salt [cryptography.SaltBytes]byte,
) error {
	logicalRefJSON, _ := json.Marshal(logicalRef)
	contentTypeJSON, _ := json.Marshal(contentType)
	header := `{"schema_id":"` + BackupArtifactEnvelopeV3SchemaID +
		`","application_crypto_format":"` + ApplicationCryptoFormatID + `","logical_ref":` + string(logicalRefJSON) +
		`,"content_type":` + string(contentTypeJSON) +
		`,"kdf":"` + BackupArtifactEnvelopeV3KDF +
		`","cipher":"` + BackupArtifactEnvelopeV3Cipher +
		`","salt_base64":"` + base64.StdEncoding.EncodeToString(salt[:]) +
		`","chunk_plaintext_bytes":` + strconv.Itoa(BackupArtifactChunkPlaintextBytes) +
		`,"chunks":[`
	if _, err := io.WriteString(writer, header); err != nil {
		return fmt.Errorf("write streaming backup artifact envelope header: %w", err)
	}
	return nil
}

func writeBackupArtifactEnvelopeV3Chunk(
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

func decryptBackupArtifactEnvelopeV3(ctx context.Context, reader io.Reader, masterKey RecoveryEncryptionKey, logicalRef, contentType string, maxBytes int64, destination io.Writer) (BackupArtifactProof, error) {
	fail := func() (BackupArtifactProof, error) { return BackupArtifactProof{}, ErrInvalidBackupArtifact }
	if maxBytes < 0 {
		return fail()
	}
	decoder := newEnvelopeDecoder(reader)
	if err := expectJSONDelimiter(decoder, '{', "envelope object"); err != nil {
		return BackupArtifactProof{}, err
	}
	var schemaID, formatID, encodedRef, encodedType, kdfID, cipherID, saltBase64 string
	for _, field := range []struct {
		name  string
		value *string
	}{
		{"schema_id", &schemaID}, {"application_crypto_format", &formatID}, {"logical_ref", &encodedRef}, {"content_type", &encodedType}, {"kdf", &kdfID}, {"cipher", &cipherID}, {"salt_base64", &saltBase64},
	} {
		if err := decodeExactJSONField(decoder, field.name, field.value); err != nil {
			return BackupArtifactProof{}, err
		}
	}
	if schemaID != BackupArtifactEnvelopeV3SchemaID || formatID != ApplicationCryptoFormatID || encodedRef != logicalRef || (contentType != "" && encodedType != contentType) || kdfID != BackupArtifactEnvelopeV3KDF || cipherID != BackupArtifactEnvelopeV3Cipher {
		return fail()
	}
	if _, err := validateBackupContentType(encodedType); err != nil {
		return fail()
	}
	salt, err := decodeExactBase64(saltBase64, cryptography.SaltBytes, "salt")
	if err != nil {
		return fail()
	}
	var saltValue [cryptography.SaltBytes]byte
	copy(saltValue[:], salt)
	opener, err := masterKey.key.ArtifactOpener(BackupArtifactEnvelopeV3SchemaID, []string{ApplicationCryptoFormatID, logicalRef, encodedType}, saltValue)
	if err != nil {
		return fail()
	}
	var chunkSize json.Number
	if err := decodeExactJSONField(decoder, "chunk_plaintext_bytes", &chunkSize); err != nil {
		return BackupArtifactProof{}, err
	}
	if chunkSize.String() != strconv.Itoa(BackupArtifactChunkPlaintextBytes) {
		return fail()
	}
	if err := expectJSONFieldName(decoder, "chunks"); err != nil {
		return BackupArtifactProof{}, err
	}
	if err := expectJSONDelimiter(decoder, '[', "chunks array"); err != nil {
		return BackupArtifactProof{}, err
	}
	hasher := sha256.New()
	var size int64
	for index := uint64(0); ; index++ {
		if err := ctx.Err(); err != nil {
			return BackupArtifactProof{}, err
		}
		if index >= BackupArtifactMaximumChunks || !decoder.More() {
			return fail()
		}
		chunkIndex, length, final, encoded, err := decodeBackupArtifactEnvelopeV3Chunk(decoder)
		if err != nil {
			return BackupArtifactProof{}, err
		}
		if chunkIndex != index || (!final && length != BackupArtifactChunkPlaintextBytes) || (length == 0 && index != 0) || int64(length) > maxBytes-size {
			return fail()
		}
		ciphertext, err := decodeExactBase64(encoded, length+backupArtifactEnvelopeV3Overhead, "ciphertext")
		if err != nil {
			return fail()
		}
		plaintext, err := opener.Open(ciphertext, backupArtifactEnvelopeV3AAD(logicalRef, encodedType, uint32(index), length, final))
		if err != nil || len(plaintext) != length {
			return fail()
		}
		_, _ = hasher.Write(plaintext)
		size += int64(length)
		if length != 0 {
			n, err := destination.Write(plaintext)
			if err != nil {
				return BackupArtifactProof{}, err
			}
			if n != length {
				return BackupArtifactProof{}, io.ErrShortWrite
			}
		}
		if final {
			break
		}
	}
	if decoder.More() {
		return fail()
	}
	if err := expectJSONDelimiter(decoder, ']', "chunks array"); err != nil {
		return BackupArtifactProof{}, err
	}
	if decoder.More() {
		return fail()
	}
	if err := expectJSONDelimiter(decoder, '}', "envelope object"); err != nil {
		return BackupArtifactProof{}, err
	}
	if err := rejectBackupArtifactEnvelopeV3TrailingData(decoder.Decoder, reader); err != nil {
		return BackupArtifactProof{}, err
	}
	return BackupArtifactProof{Key: logicalRef, ContentType: encodedType, SizeBytes: size, SHA256: hex.EncodeToString(hasher.Sum(nil))}, nil
}

// Reset the input budget for each JSON member, accounting for decoder read-ahead.
// This keeps a hostile string or whitespace run from allocating an artifact-sized
// buffer. Ordinary fields get 2 KiB; only ciphertext gets the chunk-sized bound.
type envelopeDecoder struct {
	*json.Decoder
	limited *io.LimitedReader
	counted *countedReader
}

func newEnvelopeDecoder(reader io.Reader) *envelopeDecoder {
	counted := &countedReader{reader: reader}
	limited := &io.LimitedReader{R: counted, N: 2048}
	decoder := json.NewDecoder(limited)
	decoder.UseNumber()
	return &envelopeDecoder{Decoder: decoder, limited: limited, counted: counted}
}
func (decoder *envelopeDecoder) budget(n int64) {
	decoder.limited.N = max(0, n-(decoder.counted.count-decoder.InputOffset()))
}
func (decoder *envelopeDecoder) Token() (json.Token, error) {
	decoder.budget(2048)
	return decoder.Decoder.Token()
}
func (decoder *envelopeDecoder) More() bool { decoder.budget(2048); return decoder.Decoder.More() }
func (decoder *envelopeDecoder) field(name string, value any) error {
	bound := int64(2048)
	if name == "ciphertext_base64" {
		bound = int64(base64.StdEncoding.EncodedLen(BackupArtifactChunkPlaintextBytes+backupArtifactEnvelopeV3Overhead) + 32)
	}
	decoder.budget(bound)
	return decoder.Decoder.Decode(value)
}

func decodeBackupArtifactEnvelopeV3Chunk(
	decoder *envelopeDecoder,
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

func decodeExactJSONField(decoder *envelopeDecoder, name string, destination any) error {
	if err := expectJSONFieldName(decoder, name); err != nil {
		return err
	}
	if err := decoder.field(name, destination); err != nil {
		return fmt.Errorf("%w: decode streaming envelope member %s", ErrInvalidBackupArtifact, name)
	}
	return nil
}

func expectJSONFieldName(decoder *envelopeDecoder, expected string) error {
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

func expectJSONDelimiter(decoder *envelopeDecoder, expected json.Delim, label string) error {
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

func rejectBackupArtifactEnvelopeV3TrailingData(decoder *json.Decoder, reader io.Reader) error {
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
	if err != nil || len(decoded) != expectedBytes || base64.StdEncoding.EncodeToString(decoded) != encoded {
		return nil, fmt.Errorf("%w: streaming envelope %s encoding is invalid", ErrInvalidBackupArtifact, label)
	}
	return decoded, nil
}

func isASCIIAlphaNumeric(value byte) bool {
	return (value >= 'a' && value <= 'z') ||
		(value >= 'A' && value <= 'Z') ||
		(value >= '0' && value <= '9')
}

func backupArtifactEnvelopeV3AAD(logicalRef, contentType string, index uint32, plaintextLength int, final bool) []byte {
	var result []byte
	for _, value := range []string{ApplicationCryptoFormatID, BackupArtifactEnvelopeV3SchemaID, logicalRef, contentType} {
		result = binary.BigEndian.AppendUint32(result, uint32(len(value)))
		result = append(result, value...)
	}
	result = binary.BigEndian.AppendUint32(result, index)
	result = binary.BigEndian.AppendUint32(result, uint32(plaintextLength))
	if final {
		return append(result, 1)
	}
	return append(result, 0)
}

func expectedBackupArtifactEnvelopeV3Chunks(plaintextBytes int64) (uint64, error) {
	if plaintextBytes < 0 {
		return 0, fmt.Errorf("%w: streaming artifact size is invalid", ErrInvalidBackupArtifact)
	}
	if plaintextBytes == 0 {
		return 1, nil
	}
	chunks := (uint64(plaintextBytes) + BackupArtifactChunkPlaintextBytes - 1) /
		BackupArtifactChunkPlaintextBytes
	if chunks > BackupArtifactMaximumChunks {
		return 0, fmt.Errorf("%w: streaming artifact has too many chunks", ErrInvalidBackupArtifact)
	}
	return chunks, nil
}

func expectedBackupArtifactEnvelopeV3ChunkLength(
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

func maximumBackupArtifactEnvelopeV3Bytes(plaintextBytes int64) (uint64, error) {
	chunks, err := expectedBackupArtifactEnvelopeV3Chunks(plaintextBytes)
	if err != nil {
		return 0, err
	}
	fullChunkEncodedBytes := uint64(base64.StdEncoding.EncodedLen(
		BackupArtifactChunkPlaintextBytes + backupArtifactEnvelopeV3Overhead,
	))
	fullChunks := chunks - 1
	if fullChunks > math.MaxUint64/fullChunkEncodedBytes {
		return 0, fmt.Errorf("%w: streaming envelope bound overflows", ErrInvalidBackupArtifact)
	}
	encodedCiphertextBytes := fullChunks * fullChunkEncodedBytes
	lastPlaintextBytes := expectedBackupArtifactEnvelopeV3ChunkLength(
		plaintextBytes,
		chunks,
		chunks-1,
	)
	lastEncodedBytes := uint64(base64.StdEncoding.EncodedLen(
		lastPlaintextBytes + backupArtifactEnvelopeV3Overhead,
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
