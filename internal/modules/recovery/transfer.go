package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/google/uuid"
	"io"
	"sort"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const TransferManifestSchemaID = "cartulary.recovery_transfer_manifest.v1"
const TransferManifestName = "transfer-manifest.sealed"
const TransferManifestMaximumBytes = 16 << 20
const TransferMaximumArtifacts = 4096

var ErrInvalidTransfer = errors.New("recovery: invalid transfer bundle")
var ErrTransferReleaseMismatch = errors.New("recovery: transfer requires matching release")

// TransferSource is an external read capability. Authentication and capture must
// never require staging plaintext or writing inside the supplied transfer media.
type TransferSource interface {
	ReadArtifact(context.Context, string, int64) ([]byte, error)
	OpenStoredArtifact(context.Context, string) (io.ReadCloser, int64, error)
	ListArtifacts(context.Context) ([]string, error)
}

// TransferDirectory is a private, confined publication capability. Until Publish
// succeeds Close discards its unpublished files. It never replaces a destination.
type TransferDirectory interface {
	StoredStreamingBackupStorage
	ListArtifacts(context.Context) ([]string, error)
	Publish(context.Context) error
	Close() error
}

type TransferArtifact struct {
	Reference string `json:"reference"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type TransferManifest struct {
	SchemaID                string             `json:"schema_id"`
	ApplicationCryptoFormat string             `json:"application_crypto_format"`
	ReleaseSHA256           string             `json:"release_sha256"`
	Generation              TransferGeneration `json:"generation"`
	Selection               TransferSelection  `json:"selection"`
	Artifacts               []TransferArtifact `json:"artifacts"`
	ExportedAt              time.Time          `json:"exported_at"`
}

// ExportTransfer copies exactly the authenticated catalog closure, retaining the
// current encrypted artifact bytes. The caller holds source retention exclusion.
func (catalog *BackupCatalog) ExportTransfer(ctx context.Context, backup BackupSet, release string, now time.Time, destination TransferDirectory) (TransferManifest, error) {
	if catalog == nil {
		return TransferManifest{}, ErrInvalidTransfer
	}
	encrypted, ok := catalog.storage.(encryptedBackupStorage)
	if !ok || !validSHA256Hex(release) || destination == nil {
		return TransferManifest{}, ErrInvalidTransfer
	}
	if err := catalog.VerifyBackupSetDurability(ctx, backup); err != nil {
		return TransferManifest{}, err
	}
	selection, err := readVNextBackupSetGeneration(ctx, catalog.storage, catalog.stateCatalog, backup)
	if err != nil {
		return TransferManifest{}, err
	}
	streaming, err := RequireStreamingBackupStorage(catalog.storage)
	if err != nil {
		return TransferManifest{}, err
	}
	integrity, err := VNextProofFromMetadata(ctx, streaming, backup.IntegrityManifestKey, vNextJSONContentType, backup.IntegrityManifestSizeBytes, backup.IntegrityManifestSHA256)
	if err != nil {
		return TransferManifest{}, err
	}
	expected, err := transferClosure(ctx, selection, integrity)
	if err != nil {
		return TransferManifest{}, err
	}

	refs := make([]string, 0, len(expected))
	for ref := range expected {
		refs = append(refs, ref)
	}
	sort.Strings(refs)
	manifest := TransferManifest{SchemaID: TransferManifestSchemaID, ApplicationCryptoFormat: ApplicationCryptoFormatID, ReleaseSHA256: release, Generation: transferGeneration(selection.generation.identity()), Selection: transferSelection(backup), Artifacts: make([]TransferArtifact, 0, len(refs)), ExportedAt: now.UTC()}
	for _, ref := range refs {
		if err := ctx.Err(); err != nil {
			return TransferManifest{}, err
		}
		reader, size, err := encrypted.backend.OpenStoredArtifact(ctx, ref)
		if err != nil {
			return TransferManifest{}, err
		}
		proof, copyErr := destination.WriteStoredArtifact(ctx, ref, "application/octet-stream", func(w io.Writer) error {
			return copyTransferArtifact(ctx, w, reader, size, expected[ref])
		})
		closeErr := reader.Close()
		if err := errors.Join(copyErr, closeErr); err != nil {
			return TransferManifest{}, err
		}
		if proof.SizeBytes != size || proof.SHA256 != expected[ref] {
			return TransferManifest{}, ErrInvalidTransfer
		}
		manifest.Artifacts = append(manifest.Artifacts, TransferArtifact{Reference: ref, SizeBytes: size, SHA256: expected[ref]})
	}
	body, err := canonicaljson.Marshal(manifest)
	if err != nil || len(body) > TransferManifestMaximumBytes {
		return TransferManifest{}, ErrInvalidTransfer
	}
	sealed, err := encrypted.key.key.Seal(TransferManifestSchemaID, []string{ApplicationCryptoFormatID}, body, nil)
	if err != nil {
		return TransferManifest{}, err
	}
	if _, err := destination.WriteArtifact(ctx, TransferManifestName, sealed, "application/octet-stream"); err != nil {
		return TransferManifest{}, err
	}
	if _, err := catalog.ValidateTransfer(ctx, destination, encrypted.key, release); err != nil {
		return TransferManifest{}, err
	}
	return manifest, nil
}

// ValidateTransfer authenticates metadata, exact membership and every encrypted
// byte, then asks the existing catalog to authenticate and validate the closure.
// No source database or object service is used by this path.
func (catalog *BackupCatalog) ValidateTransfer(ctx context.Context, storage TransferDirectory, key RecoveryEncryptionKey, release string) (TransferManifest, error) {
	manifest, err := authenticateTransferMembers(ctx, storage, key, release)
	if err != nil {
		return TransferManifest{}, err
	}
	encrypted, err := NewEncryptedBackupStorage(storage, key)
	if err != nil {
		return TransferManifest{}, err
	}
	validator := NewBackupCatalog(nil, encrypted, catalog.extensionBackups, catalog.stateCatalog)
	if err := validator.VerifyBackupSetDurability(ctx, manifest.Selection.BackupSet()); err != nil {
		return TransferManifest{}, err
	}
	selected, err := readVNextBackupSetGeneration(ctx, encrypted, catalog.stateCatalog, manifest.Selection.BackupSet())
	if err != nil {
		return TransferManifest{}, err
	}
	if transferGeneration(selected.generation.identity()) != manifest.Generation {
		return TransferManifest{}, ErrInvalidTransfer
	}
	streaming, err := RequireStreamingBackupStorage(encrypted)
	if err != nil {
		return TransferManifest{}, err
	}
	integrity, err := VNextProofFromMetadata(ctx, streaming, manifest.Selection.Integrity.Reference, vNextJSONContentType, manifest.Selection.Integrity.SizeBytes, manifest.Selection.Integrity.SHA256)
	if err != nil {
		return TransferManifest{}, ErrInvalidTransfer
	}
	closure, err := transferClosure(ctx, selected, integrity)
	if err != nil {
		return TransferManifest{}, err
	}
	if len(closure) != len(manifest.Artifacts) {
		return TransferManifest{}, ErrInvalidTransfer
	}
	for _, artifact := range manifest.Artifacts {
		if closure[artifact.Reference] != artifact.SHA256 {
			return TransferManifest{}, ErrInvalidTransfer
		}
	}

	return manifest, nil
}

func copyTransferArtifact(ctx context.Context, destination io.Writer, source io.Reader, size int64, digest string) error {
	// The current chunked envelope bound is shared with encrypted artifact reads.
	maximum, err := maximumBackupArtifactEnvelopeV3Bytes(int64(BackupArtifactChunkPlaintextBytes) * int64(BackupArtifactMaximumChunks))
	if err != nil || size <= 0 || uint64(size) > maximum {
		return ErrInvalidTransfer
	}
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(destination, hash), &transferContextReader{ctx: ctx, reader: io.LimitReader(source, size+1)})
	if err != nil {
		return err
	}
	if written != size || hex.EncodeToString(hash.Sum(nil)) != digest {
		return ErrInvalidTransfer
	}
	return nil
}

type transferContextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *transferContextReader) Read(b []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(b)
}

type TransferGeneration struct {
	RecoveryStateCatalogSHA256       string `json:"recovery_state_catalog_sha256"`
	CodecRegistrySHA256              string `json:"codec_registry_sha256"`
	GraphSourceRegistrySHA256        string `json:"graph_source_registry_sha256"`
	GraphImplementationBindingSHA256 string `json:"graph_implementation_binding_sha256"`
	GraphAlgorithmID                 string `json:"graph_algorithm_id"`
}

func transferGeneration(g RecoveryGenerationIdentity) TransferGeneration {
	return TransferGeneration(g)
}

type TransferSelection struct {
	BackupSetID        uuid.UUID        `json:"backup_set_id"`
	ConsistencyPointAt time.Time        `json:"consistency_point_at"`
	CreatedAt          time.Time        `json:"created_at"`
	RetainedUntil      time.Time        `json:"retained_until"`
	Postgres           TransferArtifact `json:"postgres"`
	Objects            TransferArtifact `json:"objects"`
	Integrity          TransferArtifact `json:"integrity"`
}

func transferSelection(b BackupSet) TransferSelection {
	return TransferSelection{
		BackupSetID: b.BackupSetID, ConsistencyPointAt: b.ConsistencyPointAt, CreatedAt: b.CreatedAt, RetainedUntil: b.RetainedUntil,
		Postgres: TransferArtifact{b.PostgresArtifactKey, b.PostgresArtifactSizeBytes, b.PostgresArtifactSHA256}, Objects: TransferArtifact{b.ObjectStoreArtifactKey, b.ObjectStoreArtifactSizeBytes, b.ObjectStoreArtifactSHA256}, Integrity: TransferArtifact{b.IntegrityManifestKey, b.IntegrityManifestSizeBytes, b.IntegrityManifestSHA256},
	}
}
func (s TransferSelection) BackupSet() BackupSet {
	return BackupSet{
		BackupSetID: s.BackupSetID, ConsistencyPointAt: s.ConsistencyPointAt, CreatedAt: s.CreatedAt, RetainedUntil: s.RetainedUntil,
		PostgresArtifactKey: s.Postgres.Reference, PostgresArtifactSizeBytes: s.Postgres.SizeBytes, PostgresArtifactSHA256: s.Postgres.SHA256,
		ObjectStoreArtifactKey: s.Objects.Reference, ObjectStoreArtifactSizeBytes: s.Objects.SizeBytes, ObjectStoreArtifactSHA256: s.Objects.SHA256,
		IntegrityManifestKey: s.Integrity.Reference, IntegrityManifestSizeBytes: s.Integrity.SizeBytes, IntegrityManifestSHA256: s.Integrity.SHA256,
		VerificationState: VerificationUnverified,
	}
}

// authenticateTransferMembers bounds and authenticates the external directory
// using encrypted bytes only. Full semantic validation uses private captured storage.
func authenticateTransferMembers(ctx context.Context, storage TransferSource, key RecoveryEncryptionKey, release string) (TransferManifest, error) {
	sealed, err := storage.ReadArtifact(ctx, TransferManifestName, TransferManifestMaximumBytes+65536)
	if err != nil {
		return TransferManifest{}, err
	}
	body, err := key.key.Open(TransferManifestSchemaID, []string{ApplicationCryptoFormatID}, sealed, nil)
	if err != nil || len(body) > TransferManifestMaximumBytes {
		return TransferManifest{}, ErrInvalidTransfer
	}
	var manifest TransferManifest
	if err := strictDecodeJSON(body, &manifest); err != nil {
		return TransferManifest{}, ErrInvalidTransfer
	}
	canonical, err := canonicaljson.Marshal(manifest)
	if err != nil || !bytes.Equal(body, canonical) || manifest.SchemaID != TransferManifestSchemaID || manifest.ApplicationCryptoFormat != ApplicationCryptoFormatID || len(manifest.Artifacts) == 0 || len(manifest.Artifacts) > TransferMaximumArtifacts || manifest.ExportedAt.IsZero() || manifest.Selection.ConsistencyPointAt.After(manifest.ExportedAt) {
		return TransferManifest{}, ErrInvalidTransfer
	}
	if manifest.Selection.BackupSetID == uuid.Nil || manifest.Selection.CreatedAt.IsZero() || manifest.Selection.CreatedAt.After(manifest.ExportedAt) || !manifest.Selection.RetainedUntil.After(manifest.Selection.ConsistencyPointAt) {
		return TransferManifest{}, ErrInvalidTransfer
	}
	if !validSHA256Hex(release) || manifest.ReleaseSHA256 != release {
		return TransferManifest{}, ErrTransferReleaseMismatch
	}
	names, err := storage.ListArtifacts(ctx)
	if err != nil {
		return TransferManifest{}, err
	}
	expected := map[string]bool{TransferManifestName: true}
	previous := ""
	for _, artifact := range manifest.Artifacts {
		if _, err := validateBackupLogicalRef(artifact.Reference); err != nil || artifact.Reference <= previous || expected[artifact.Reference] || !validSHA256Hex(artifact.SHA256) {
			return TransferManifest{}, ErrInvalidTransfer
		}
		previous = artifact.Reference
		expected[artifact.Reference] = true
		reader, size, err := storage.OpenStoredArtifact(ctx, artifact.Reference)
		if err != nil {
			return TransferManifest{}, err
		}
		validationErr := error(nil)
		if size != artifact.SizeBytes {
			validationErr = ErrInvalidTransfer
		} else {
			validationErr = copyTransferArtifact(ctx, io.Discard, reader, size, artifact.SHA256)
		}
		if err := errors.Join(validationErr, reader.Close()); err != nil {
			return TransferManifest{}, err
		}
	}
	if len(names) != len(expected) {
		return TransferManifest{}, ErrInvalidTransfer
	}
	for _, name := range names {
		if !expected[name] {
			return TransferManifest{}, ErrInvalidTransfer
		}
	}
	return manifest, nil
}

// CaptureTransfer validates the external input, copies its exact encrypted bytes
// into private storage, and validates the captured closure again before restore.
func (catalog *BackupCatalog) CaptureTransfer(ctx context.Context, input TransferSource, captured TransferDirectory, key RecoveryEncryptionKey, release string) (TransferManifest, string, error) {
	manifest, err := authenticateTransferMembers(ctx, input, key, release)
	if err != nil {
		return TransferManifest{}, "", err
	}
	sealed, err := input.ReadArtifact(ctx, TransferManifestName, TransferManifestMaximumBytes+65536)
	if err != nil {
		return TransferManifest{}, "", err
	}
	for _, artifact := range manifest.Artifacts {
		reader, size, err := input.OpenStoredArtifact(ctx, artifact.Reference)
		if err != nil {
			return TransferManifest{}, "", err
		}
		_, copyErr := captured.WriteStoredArtifact(ctx, artifact.Reference, "application/octet-stream", func(w io.Writer) error {
			if size != artifact.SizeBytes {
				return ErrInvalidTransfer
			}
			return copyTransferArtifact(ctx, w, reader, size, artifact.SHA256)
		})
		if err := errors.Join(copyErr, reader.Close()); err != nil {
			return TransferManifest{}, "", err
		}
	}
	if _, err := captured.WriteArtifact(ctx, TransferManifestName, sealed, "application/octet-stream"); err != nil {
		return TransferManifest{}, "", err
	}
	verified, err := catalog.ValidateTransfer(ctx, captured, key, release)
	if err != nil {
		return TransferManifest{}, "", err
	}
	first, _ := canonicaljson.Marshal(manifest)
	second, _ := canonicaljson.Marshal(verified)
	if !bytes.Equal(first, second) {
		return TransferManifest{}, "", ErrInvalidTransfer
	}
	digest := sha256.Sum256(sealed)
	return verified, hex.EncodeToString(digest[:]), nil
}

func transferClosure(ctx context.Context, selection vNextBackupGenerationSelection, integrity BackupArtifactStreamProof) (map[string]string, error) {
	expected := map[string]string{integrity.EnvelopeRef: integrity.EnvelopeSHA256}
	add := func(proof BackupArtifactStreamProof) error {
		if _, duplicate := expected[proof.EnvelopeRef]; duplicate {
			return ErrInvalidTransfer
		}
		expected[proof.EnvelopeRef] = proof.EnvelopeSHA256
		if len(expected) > TransferMaximumArtifacts {
			return ErrInvalidTransfer
		}
		return nil
	}
	for _, proof := range selection.proofs {
		if err := add(streamProof(proof)); err != nil {
			return nil, err
		}
	}
	proof, ok := selection.proofs[selection.integrity.ObjectStoreManifestRef]
	if !ok {
		return nil, ErrInvalidTransfer
	}
	_, objects, err := selection.restore.readObjectClosure(ctx, selection.integrity, selection.generation, proof)
	if err != nil {
		return nil, err
	}
	for _, proof := range objects {
		if err := add(proof); err != nil {
			return nil, err
		}
	}
	return expected, nil
}
