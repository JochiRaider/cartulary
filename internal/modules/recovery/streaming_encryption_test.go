package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

const streamingRecoveryMasterKey = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
const wrongStreamingRecoveryMasterKey = "YWJjZGVmMDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODk="

type streamingEnvelopeFixture struct {
	SchemaID                string                          `json:"schema_id"`
	ApplicationCryptoFormat string                          `json:"application_crypto_format"`
	LogicalRef              string                          `json:"logical_ref"`
	ContentType             string                          `json:"content_type"`
	KDF                     string                          `json:"kdf"`
	Cipher                  string                          `json:"cipher"`
	SaltBase64              string                          `json:"salt_base64"`
	ChunkPlaintextSize      int                             `json:"chunk_plaintext_bytes"`
	Chunks                  []streamingEnvelopeChunkFixture `json:"chunks"`
}

type streamingEnvelopeChunkFixture struct {
	Index            uint64 `json:"index"`
	PlaintextLength  int    `json:"plaintext_length"`
	Final            bool   `json:"final"`
	CiphertextBase64 string `json:"ciphertext_base64"`
}

type generatedPatternReader struct {
	remaining  int64
	offset     int64
	maxRequest int
	hasher     hash.Hash
}

func (reader *generatedPatternReader) Read(body []byte) (int, error) {
	if len(body) > reader.maxRequest {
		reader.maxRequest = len(body)
	}
	if reader.remaining == 0 {
		return 0, io.EOF
	}
	count := len(body)
	if int64(count) > reader.remaining {
		count = int(reader.remaining)
	}
	for index := 0; index < count; index++ {
		body[index] = byte((reader.offset + int64(index)) % 251)
	}
	_, _ = reader.hasher.Write(body[:count])
	reader.offset += int64(count)
	reader.remaining -= int64(count)
	return count, nil
}

type observedStreamWriter struct {
	count    int64
	maxWrite int
	hasher   hash.Hash
}

func (writer *observedStreamWriter) Write(body []byte) (int, error) {
	if len(body) > writer.maxWrite {
		writer.maxWrite = len(body)
	}
	writer.count += int64(len(body))
	_, _ = writer.hasher.Write(body)
	return len(body), nil
}

func TestStreamingBackupArtifactEnvelopeV2FailsClosedAndBoundsMemory_Unit(t *testing.T) {
	ctx := context.Background()
	t.Run("small and empty artifacts use the same current stream", func(t *testing.T) {
		root, storage, _ := writeStreamingArtifact(t, strings.NewReader("seed"), "seed")
		for index, body := range [][]byte{nil, []byte("small artifact")} {
			key := "small-" + strconv.Itoa(index)
			proof, err := storage.WriteArtifact(ctx, key, body, "application/octet-stream")
			if err != nil {
				t.Fatal(err)
			}
			opened, err := recovery.VerifyArtifactProof(ctx, storage, proof)
			if err != nil || !bytes.Equal(opened, body) {
				t.Fatalf("small read: %q %v", opened, err)
			}
			raw := readStreamingEnvelopeBody(t, root, proof.Key)
			streamProof := recovery.BackupArtifactStreamProof{LogicalRef: proof.Key, EnvelopeRef: proof.Key, ContentType: proof.ContentType, PlaintextBytes: proof.SizeBytes, PlaintextSHA256: proof.SHA256, EnvelopeBytes: int64(len(raw)), EnvelopeSHA256: sha256HexForStreamingTest(raw)}
			var output bytes.Buffer
			if err := storage.ReadArtifactStream(ctx, streamProof, &output); err != nil || !bytes.Equal(output.Bytes(), body) {
				t.Fatalf("small stream read: %v", err)
			}
			if _, err := storage.WriteArtifact(ctx, key, body, "application/octet-stream"); err == nil {
				t.Fatal("immutable artifact was overwritten")
			}
			if _, err := storage.WriteArtifactStream(ctx, recovery.BackupArtifactStreamWriteRequest{LogicalRef: key, EnvelopeRef: key + "-copy", ContentType: proof.ContentType, Plaintext: bytes.NewReader(body)}); err != nil {
				t.Fatal(err)
			}
			first := readStreamingEnvelopeFixture(t, root, key)
			second := readStreamingEnvelopeFixture(t, root, key+"-copy")
			if first.SaltBase64 == second.SaltBase64 {
				t.Fatal("new artifact reused its salt")
			}
		}
	})
	t.Run("interrupted writes never publish an envelope", func(t *testing.T) {
		root, storage, _ := writeStreamingArtifact(t, strings.NewReader("seed"), "seed")
		interrupted := errors.New("source interrupted")
		source := io.MultiReader(bytes.NewReader(make([]byte, recovery.BackupArtifactChunkPlaintextBytes+1)), streamReaderFunc(func([]byte) (int, error) { return 0, interrupted }))
		proof, err := storage.WriteArtifactStream(ctx, recovery.BackupArtifactStreamWriteRequest{LogicalRef: "interrupted", EnvelopeRef: "interrupted.json", ContentType: "application/octet-stream", Plaintext: source})
		if !errors.Is(err, interrupted) || proof.EnvelopeRef != "" {
			t.Fatalf("interrupted publication: %#v %v", proof, err)
		}
		if _, err := os.Stat(filepath.Join(root, "interrupted.json")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("interrupted envelope remains: %v", err)
		}
		cancelled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := storage.WriteArtifactStream(cancelled, recovery.BackupArtifactStreamWriteRequest{LogicalRef: "cancelled", EnvelopeRef: "cancelled.json", ContentType: "application/octet-stream", Plaintext: strings.NewReader("secret")}); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(root, "cancelled.json")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("cancelled envelope remains: %v", err)
		}
	})
	t.Run("authenticated output never reopens mutable source", func(t *testing.T) {
		plaintext := bytes.Repeat([]byte("immutable-content"), 4096)
		rootPath, _, proof := writeStreamingArtifact(t, bytes.NewReader(plaintext), "mutable")
		raw, err := recoveryassembly.NewFilesystemStorage(rootPath)
		if err != nil {
			t.Fatal(err)
		}
		backend := &mutatingStreamBackend{FilesystemStorage: raw, mutate: func() {
			if err := os.WriteFile(filepath.Join(rootPath, filepath.FromSlash(proof.EnvelopeRef)), []byte("replaced after authentication"), 0o600); err != nil {
				t.Fatal(err)
			}
		}}
		key, err := recovery.ParseRecoveryEncryptionKey(streamingRecoveryMasterKey)
		if err != nil {
			t.Fatal(err)
		}
		encrypted, err := recovery.NewEncryptedBackupStorage(backend, key)
		if err != nil {
			t.Fatal(err)
		}
		defer recovery.CloseBackupStorage(encrypted)
		stream, err := recovery.RequireStreamingBackupStorage(encrypted)
		if err != nil {
			t.Fatal(err)
		}
		var output bytes.Buffer
		if err := stream.ReadArtifactStream(ctx, proof, &output); err != nil {
			t.Fatal(err)
		}
		if backend.opens != 1 || !bytes.Equal(output.Bytes(), plaintext) {
			t.Fatal("output did not use exactly the authenticated input")
		}
	})

	t.Run("multi-chunk and zero-byte artifacts round trip with exact envelope facts", func(t *testing.T) {
		multiChunk := bytes.Repeat([]byte("cartulary-stream"), recovery.BackupArtifactChunkPlaintextBytes/8)
		multiChunk = append(multiChunk, bytes.Repeat([]byte{0x5a}, 73)...)
		rootPath, storage, proof := writeStreamingArtifact(t, bytes.NewReader(multiChunk), "multi")
		var restored bytes.Buffer
		if err := storage.ReadArtifactStream(ctx, proof, &restored); err != nil {
			t.Fatalf("read multi-chunk artifact: %v", err)
		}
		if !bytes.Equal(restored.Bytes(), multiChunk) {
			t.Fatal("multi-chunk plaintext changed")
		}
		envelope := readStreamingEnvelopeFixture(t, rootPath, proof.EnvelopeRef)
		if envelope.SchemaID != recovery.BackupArtifactEnvelopeV3SchemaID ||
			envelope.KDF != recovery.BackupArtifactEnvelopeV3KDF ||
			envelope.Cipher != recovery.BackupArtifactEnvelopeV3Cipher ||
			envelope.ChunkPlaintextSize != recovery.BackupArtifactChunkPlaintextBytes {
			t.Fatalf("streaming envelope algorithms = %#v", envelope)
		}
		salt, err := base64.StdEncoding.Strict().DecodeString(envelope.SaltBase64)
		if err != nil || len(salt) != 32 {
			t.Fatalf("streaming envelope salt = %d bytes, %v", len(salt), err)
		}
		if len(envelope.Chunks) < 2 {
			t.Fatalf("multi-chunk envelope has %d chunk", len(envelope.Chunks))
		}
		for index, chunk := range envelope.Chunks {
			if chunk.Index != uint64(index) || chunk.Final != (index == len(envelope.Chunks)-1) {
				t.Fatalf("chunk %d sequence = %#v", index, chunk)
			}
			if !chunk.Final && chunk.PlaintextLength != recovery.BackupArtifactChunkPlaintextBytes {
				t.Fatalf("non-final chunk %d length = %d", index, chunk.PlaintextLength)
			}
		}
		raw, err := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(proof.EnvelopeRef)))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(raw, []byte("cartulary-stream")) {
			t.Fatal("streaming envelope disclosed plaintext")
		}

		zeroRoot, zeroStorage, zeroProof := writeStreamingArtifact(t, strings.NewReader(""), "zero")
		if zeroProof.PlaintextBytes != 0 ||
			zeroProof.PlaintextSHA256 != sha256HexForStreamingTest(nil) {
			t.Fatalf("zero-byte proof = %#v", zeroProof)
		}
		zeroDestination := &observedStreamWriter{hasher: sha256.New()}
		if err := zeroStorage.ReadArtifactStream(ctx, zeroProof, zeroDestination); err != nil {
			t.Fatalf("read zero-byte artifact: %v", err)
		}
		if zeroDestination.count != 0 {
			t.Fatalf("zero-byte artifact wrote %d bytes", zeroDestination.count)
		}
		zeroEnvelope := readStreamingEnvelopeFixture(t, zeroRoot, zeroProof.EnvelopeRef)
		if len(zeroEnvelope.Chunks) != 1 ||
			zeroEnvelope.Chunks[0].Index != 0 ||
			zeroEnvelope.Chunks[0].PlaintextLength != 0 ||
			!zeroEnvelope.Chunks[0].Final {
			t.Fatalf("zero-byte chunks = %#v", zeroEnvelope.Chunks)
		}
		zeroCiphertext, err := base64.StdEncoding.Strict().DecodeString(
			zeroEnvelope.Chunks[0].CiphertextBase64,
		)
		if err != nil || len(zeroCiphertext) != 28 {
			t.Fatalf("zero-byte authenticated ciphertext = %d bytes, %v", len(zeroCiphertext), err)
		}
	})

	t.Run("wrong keys and every structural mutation fail before destination use", func(t *testing.T) {
		plaintext := bytes.Repeat(
			[]byte{0x42},
			recovery.BackupArtifactChunkPlaintextBytes+37,
		)
		mutations := []struct {
			name                     string
			invalidateEnvelopeDigest bool
			mutate                   func([]byte, *streamingEnvelopeFixture, *recovery.BackupArtifactStreamProof) []byte
		}{
			{name: "wrong application format", mutate: func(_ []byte, e *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
				e.ApplicationCryptoFormat = "cartulary.application_crypto_format.v0"
				return marshalStreamingEnvelopeFixture(t, e)
			}},
			{name: "old envelope", mutate: func(_ []byte, e *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
				e.SchemaID = "cartulary.backup_artifact_envelope.v2"
				return marshalStreamingEnvelopeFixture(t, e)
			}},
			{name: "substituted artifact salt", mutate: func(_ []byte, e *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
				e.SaltBase64 = base64.StdEncoding.EncodeToString(make([]byte, 32))
				return marshalStreamingEnvelopeFixture(t, e)
			}},
			{name: "changed chunk length", mutate: func(_ []byte, e *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
				e.Chunks[0].PlaintextLength--
				return marshalStreamingEnvelopeFixture(t, e)
			}},
			{name: "oversized header", mutate: func(_ []byte, e *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
				e.SchemaID = strings.Repeat("a", 32768)
				return marshalStreamingEnvelopeFixture(t, e)
			}},
			{
				name: "ciphertext corruption",
				mutate: func(_ []byte, envelope *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					ciphertext, err := base64.StdEncoding.Strict().DecodeString(envelope.Chunks[0].CiphertextBase64)
					if err != nil {
						t.Fatal(err)
					}
					ciphertext[0] ^= 0x80
					envelope.Chunks[0].CiphertextBase64 = base64.StdEncoding.EncodeToString(ciphertext)
					return marshalStreamingEnvelopeFixture(t, envelope)
				},
			},
			{
				name: "chunk reordering",
				mutate: func(_ []byte, envelope *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					envelope.Chunks[0], envelope.Chunks[1] = envelope.Chunks[1], envelope.Chunks[0]
					return marshalStreamingEnvelopeFixture(t, envelope)
				},
			},
			{
				name: "duplicate index",
				mutate: func(_ []byte, envelope *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					envelope.Chunks[1].Index = envelope.Chunks[0].Index
					return marshalStreamingEnvelopeFixture(t, envelope)
				},
			},
			{
				name: "duplicate chunk",
				mutate: func(_ []byte, envelope *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					envelope.Chunks = append(envelope.Chunks, envelope.Chunks[0])
					return marshalStreamingEnvelopeFixture(t, envelope)
				},
			},
			{
				name: "missing final flag",
				mutate: func(_ []byte, envelope *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					envelope.Chunks[len(envelope.Chunks)-1].Final = false
					return marshalStreamingEnvelopeFixture(t, envelope)
				},
			},
			{
				name: "truncation",
				mutate: func(body []byte, _ *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					return body[:len(body)-1]
				},
			},
			{
				name: "trailing data",
				mutate: func(body []byte, _ *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					return append(body, 'x')
				},
			},
			{
				name: "AAD logical reference",
				mutate: func(_ []byte, envelope *streamingEnvelopeFixture, proof *recovery.BackupArtifactStreamProof) []byte {
					envelope.LogicalRef += "-changed"
					proof.LogicalRef = envelope.LogicalRef
					return marshalStreamingEnvelopeFixture(t, envelope)
				},
			},
			{
				name: "AAD content type",
				mutate: func(_ []byte, envelope *streamingEnvelopeFixture, proof *recovery.BackupArtifactStreamProof) []byte {
					envelope.ContentType = "application/x-cartulary-changed"
					proof.ContentType = envelope.ContentType
					return marshalStreamingEnvelopeFixture(t, envelope)
				},
			},
			{
				name: "duplicate outer member",
				mutate: func(body []byte, _ *streamingEnvelopeFixture, _ *recovery.BackupArtifactStreamProof) []byte {
					return bytes.Replace(
						body,
						[]byte(`"logical_ref":`),
						[]byte(`"schema_id":"cartulary.backup_artifact_envelope.v2","logical_ref":`),
						1,
					)
				},
			},
			{
				name: "plaintext digest mismatch",
				mutate: func(body []byte, _ *streamingEnvelopeFixture, proof *recovery.BackupArtifactStreamProof) []byte {
					proof.PlaintextSHA256 = strings.Repeat("0", 64)
					return body
				},
			},
			{
				name:                     "envelope digest mismatch",
				invalidateEnvelopeDigest: true,
				mutate: func(body []byte, _ *streamingEnvelopeFixture, proof *recovery.BackupArtifactStreamProof) []byte {
					return body
				},
			},
		}
		for _, test := range mutations {
			t.Run(test.name, func(t *testing.T) {
				rootPath, storage, proof := writeStreamingArtifact(t, bytes.NewReader(plaintext), test.name)
				body := readStreamingEnvelopeBody(t, rootPath, proof.EnvelopeRef)
				var envelope streamingEnvelopeFixture
				if err := json.Unmarshal(body, &envelope); err != nil {
					t.Fatal(err)
				}
				mutated := test.mutate(body, &envelope, &proof)
				rewriteStreamingEnvelope(t, rootPath, &proof, mutated)
				if test.invalidateEnvelopeDigest {
					proof.EnvelopeSHA256 = strings.Repeat("0", 64)
				}
				destination := &observedStreamWriter{hasher: sha256.New()}
				err := storage.ReadArtifactStream(ctx, proof, destination)
				if !errors.Is(err, recovery.ErrInvalidBackupArtifact) {
					t.Fatalf("mutated envelope error = %v; want ErrInvalidBackupArtifact", err)
				}
				if destination.count != 0 {
					t.Fatalf("mutated envelope released %d plaintext bytes", destination.count)
				}
			})
		}

		rootPath, _, proof := writeStreamingArtifact(t, bytes.NewReader(plaintext), "wrong-key")
		raw, err := recoveryassembly.NewFilesystemStorage(rootPath)
		if err != nil {
			t.Fatal(err)
		}
		wrongKey, err := recovery.ParseRecoveryEncryptionKey(wrongStreamingRecoveryMasterKey)
		if err != nil {
			t.Fatal(err)
		}
		wrongEncrypted, err := recovery.NewEncryptedBackupStorage(raw, wrongKey)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = recovery.CloseBackupStorage(wrongEncrypted) })
		wrongStorage, err := recovery.RequireStreamingBackupStorage(wrongEncrypted)
		if err != nil {
			t.Fatal(err)
		}
		destination := &observedStreamWriter{hasher: sha256.New()}
		err = wrongStorage.ReadArtifactStream(ctx, proof, destination)
		if !errors.Is(err, recovery.ErrInvalidBackupArtifact) {
			t.Fatalf("wrong-key envelope error = %v; want ErrInvalidBackupArtifact", err)
		}
		if destination.count != 0 {
			t.Fatalf("wrong-key envelope released %d plaintext bytes", destination.count)
		}
	})

	t.Run("large artifact uses fixed-size reads and writes", func(t *testing.T) {
		const largeBytes = int64(64*1024*1024 + 19)
		source := &generatedPatternReader{
			remaining: largeBytes,
			hasher:    sha256.New(),
		}
		_, storage, proof := writeStreamingArtifact(t, source, "large")
		if proof.PlaintextBytes != largeBytes ||
			proof.PlaintextSHA256 != hex.EncodeToString(source.hasher.Sum(nil)) {
			t.Fatalf("large streaming proof = %#v", proof)
		}
		if source.maxRequest > recovery.BackupArtifactChunkPlaintextBytes {
			t.Fatalf("large source read request = %d", source.maxRequest)
		}
		destination := &observedStreamWriter{hasher: sha256.New()}
		if err := storage.ReadArtifactStream(ctx, proof, destination); err != nil {
			t.Fatalf("read large streaming artifact: %v", err)
		}
		if destination.count != largeBytes ||
			hex.EncodeToString(destination.hasher.Sum(nil)) != proof.PlaintextSHA256 {
			t.Fatalf("large destination bytes=%d digest=%s", destination.count, hex.EncodeToString(destination.hasher.Sum(nil)))
		}
		if destination.maxWrite > recovery.BackupArtifactChunkPlaintextBytes {
			t.Fatalf("large destination write = %d", destination.maxWrite)
		}
	})
}

type mutatingStreamBackend struct {
	*recoveryassembly.FilesystemStorage
	opens  int
	mutate func()
}

type streamReaderFunc func([]byte) (int, error)

func (read streamReaderFunc) Read(body []byte) (int, error) { return read(body) }

func (backend *mutatingStreamBackend) OpenStoredArtifact(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	reader, size, err := backend.FilesystemStorage.OpenStoredArtifact(ctx, key)
	if err != nil {
		return nil, 0, err
	}
	backend.opens++
	return &mutatingStreamReader{ReadCloser: reader, mutate: backend.mutate}, size, nil
}

type mutatingStreamReader struct {
	io.ReadCloser
	mutate func()
}

func (reader *mutatingStreamReader) Close() error {
	err := reader.ReadCloser.Close()
	reader.mutate()
	return err
}

func writeStreamingArtifact(
	t testing.TB,
	plaintext io.Reader,
	label string,
) (string, recovery.StreamingBackupStorage, recovery.BackupArtifactStreamProof) {
	t.Helper()
	rootPath := filepath.Join(t.TempDir(), "backups")
	raw, err := recoveryassembly.NewFilesystemStorage(rootPath)
	if err != nil {
		t.Fatalf("create raw streaming storage: %v", err)
	}
	key, err := recovery.ParseRecoveryEncryptionKey(streamingRecoveryMasterKey)
	if err != nil {
		t.Fatalf("parse streaming key: %v", err)
	}
	encrypted, err := recovery.NewEncryptedBackupStorage(raw, key)
	if err != nil {
		t.Fatalf("create encrypted streaming storage: %v", err)
	}
	t.Cleanup(func() { _ = recovery.CloseBackupStorage(encrypted) })
	storage, err := recovery.RequireStreamingBackupStorage(encrypted)
	if err != nil {
		t.Fatalf("require streaming storage: %v", err)
	}
	canonicalLabel := strings.NewReplacer(" ", "-", "_", "-").Replace(label)
	proof, err := storage.WriteArtifactStream(context.Background(), recovery.BackupArtifactStreamWriteRequest{
		LogicalRef:  "backup/00000000-0000-0000-0000-000000000207/objects/" + canonicalLabel,
		EnvelopeRef: "backup/00000000-0000-0000-0000-000000000207/envelopes/" + canonicalLabel,
		ContentType: "application/octet-stream",
		Plaintext:   plaintext,
	})
	if err != nil {
		t.Fatalf("write streaming artifact: %v", err)
	}
	return rootPath, storage, proof
}

func readStreamingEnvelopeFixture(
	t testing.TB,
	rootPath string,
	envelopeRef string,
) streamingEnvelopeFixture {
	t.Helper()
	body := readStreamingEnvelopeBody(t, rootPath, envelopeRef)
	var envelope streamingEnvelopeFixture
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode streaming envelope fixture: %v", err)
	}
	return envelope
}

func readStreamingEnvelopeBody(t testing.TB, rootPath string, envelopeRef string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(rootPath, filepath.FromSlash(envelopeRef)))
	if err != nil {
		t.Fatalf("read streaming envelope: %v", err)
	}
	return body
}

func rewriteStreamingEnvelope(
	t testing.TB,
	rootPath string,
	proof *recovery.BackupArtifactStreamProof,
	body []byte,
) {
	t.Helper()
	if err := os.WriteFile(
		filepath.Join(rootPath, filepath.FromSlash(proof.EnvelopeRef)),
		body,
		0o600,
	); err != nil {
		t.Fatalf("rewrite streaming envelope: %v", err)
	}
	proof.EnvelopeBytes = int64(len(body))
	proof.EnvelopeSHA256 = sha256HexForStreamingTest(body)
}

func marshalStreamingEnvelopeFixture(t testing.TB, envelope *streamingEnvelopeFixture) []byte {
	t.Helper()
	body, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal streaming envelope fixture: %v", err)
	}
	return body
}

func sha256HexForStreamingTest(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
