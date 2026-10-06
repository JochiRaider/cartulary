package recoveryassembly

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	pathpkg "path"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

var ErrUnsupportedBackupBinding = errors.New("recovery assembly: unsupported backup storage binding")

type FilesystemStorage struct {
	root *rootedfs.Root
}

type storedArtifactWriter struct {
	writer io.Writer
	size   int64
}

func (writer *storedArtifactWriter) Write(body []byte) (int, error) {
	written, err := writer.writer.Write(body)
	writer.size += int64(written)
	return written, err
}

func NewFilesystemStorage(rootPath string) (*FilesystemStorage, error) {
	if strings.TrimSpace(rootPath) == "" {
		return nil, fmt.Errorf("%w: backup storage root path is required", ErrUnsupportedBackupBinding)
	}
	root, err := rootedfs.OpenOrCreate(rootPath)
	if err != nil {
		return nil, fmt.Errorf("open backup storage capability: %w", err)
	}
	return &FilesystemStorage{root: root}, nil
}

func NewBackupStorage(bindingKind string, rootPath string, env map[string]string) (recovery.BackupStorage, error) {
	switch bindingKind {
	case "filesystem_root":
		raw, err := NewFilesystemStorage(rootPath)
		if err != nil {
			return nil, err
		}
		encrypted, err := recovery.NewEncryptedBackupStorageFromEnv(raw, env)
		if err != nil {
			_ = raw.Close()
			return nil, err
		}
		return encrypted, nil
	case "managed_service":
		return nil, fmt.Errorf("%w: managed-service backup storage is not implemented", ErrUnsupportedBackupBinding)
	default:
		return nil, fmt.Errorf("%w: binding kind must be configured", ErrUnsupportedBackupBinding)
	}
}

func (storage *FilesystemStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (recovery.BackupArtifactProof, error) {
	if storage == nil || storage.root == nil {
		return recovery.BackupArtifactProof{}, fmt.Errorf("write backup artifact: storage is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return recovery.BackupArtifactProof{}, err
	}
	reference, err := rootedfs.ParseReference(key)
	if err != nil {
		return recovery.BackupArtifactProof{}, fmt.Errorf("write backup artifact: invalid logical reference: %w", err)
	}
	if len(body) == 0 {
		return recovery.BackupArtifactProof{}, fmt.Errorf("%w: artifact body is empty", recovery.ErrInvalidBackupArtifact)
	}
	if err := makeParent(storage.root, reference); err != nil {
		return recovery.BackupArtifactProof{}, err
	}
	if err := storage.root.CreateExclusive(ctx, reference, func(writer io.Writer) error {
		_, writeErr := writer.Write(body)
		return writeErr
	}); err != nil {
		return recovery.BackupArtifactProof{}, fmt.Errorf("write backup artifact: %w", err)
	}
	written, metadata, err := storage.root.ReadRegular(reference, int64(len(body)))
	if err != nil {
		_ = storage.root.RemoveRegular(reference)
		return recovery.BackupArtifactProof{}, fmt.Errorf("verify backup artifact: %w", err)
	}
	if len(written) != len(body) || metadata.Size != int64(len(body)) {
		_ = storage.root.RemoveRegular(reference)
		return recovery.BackupArtifactProof{}, fmt.Errorf("%w: artifact size mismatch", recovery.ErrInvalidBackupArtifact)
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	digest := sha256.Sum256(body)
	return recovery.BackupArtifactProof{
		Key:         reference.String(),
		SHA256:      hex.EncodeToString(digest[:]),
		SizeBytes:   metadata.Size,
		ContentType: contentType,
	}, nil
}

func (storage *FilesystemStorage) ReadArtifact(ctx context.Context, key string, maxBytes int64) ([]byte, error) {
	if storage == nil || storage.root == nil {
		return nil, fmt.Errorf("read backup artifact: storage is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	reference, err := rootedfs.ParseReference(key)
	if err != nil {
		return nil, fmt.Errorf("read backup artifact: invalid logical reference: %w", err)
	}
	body, _, err := storage.root.ReadRegular(reference, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("read backup artifact: %w", err)
	}
	return body, nil
}

func (storage *FilesystemStorage) WriteStoredArtifact(
	ctx context.Context,
	key string,
	contentType string,
	write func(io.Writer) error,
) (recovery.BackupArtifactProof, error) {
	if storage == nil || storage.root == nil {
		return recovery.BackupArtifactProof{}, fmt.Errorf("write stored backup artifact: storage is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return recovery.BackupArtifactProof{}, err
	}
	if write == nil {
		return recovery.BackupArtifactProof{}, fmt.Errorf(
			"%w: stored artifact writer is required",
			recovery.ErrInvalidBackupArtifact,
		)
	}
	reference, err := rootedfs.ParseReference(key)
	if err != nil {
		return recovery.BackupArtifactProof{}, fmt.Errorf(
			"write stored backup artifact: invalid logical reference: %w",
			err,
		)
	}
	if err := makeParent(storage.root, reference); err != nil {
		return recovery.BackupArtifactProof{}, err
	}
	hasher := sha256.New()
	var size int64
	if err := storage.root.CreateExclusive(ctx, reference, func(destination io.Writer) error {
		counted := &storedArtifactWriter{writer: io.MultiWriter(destination, hasher)}
		if err := write(counted); err != nil {
			return err
		}
		if counted.size <= 0 {
			return fmt.Errorf("%w: stored artifact is empty", recovery.ErrInvalidBackupArtifact)
		}
		size = counted.size
		return nil
	}); err != nil {
		return recovery.BackupArtifactProof{}, fmt.Errorf("write stored backup artifact: %w", err)
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}
	return recovery.BackupArtifactProof{
		Key:         reference.String(),
		SHA256:      hex.EncodeToString(hasher.Sum(nil)),
		SizeBytes:   size,
		ContentType: contentType,
	}, nil
}

func (storage *FilesystemStorage) OpenStoredArtifact(
	ctx context.Context,
	key string,
) (io.ReadCloser, int64, error) {
	if storage == nil || storage.root == nil {
		return nil, 0, fmt.Errorf("open stored backup artifact: storage is unavailable")
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	reference, err := rootedfs.ParseReference(key)
	if err != nil {
		return nil, 0, fmt.Errorf("open stored backup artifact: invalid logical reference: %w", err)
	}
	reader, metadata, err := storage.root.OpenRegular(reference)
	if err != nil {
		return nil, 0, fmt.Errorf("open stored backup artifact: %w", err)
	}
	if metadata.Size <= 0 {
		_ = reader.Close()
		return nil, 0, fmt.Errorf("%w: stored artifact is empty", recovery.ErrInvalidBackupArtifact)
	}
	return reader, metadata.Size, nil
}

func (storage *FilesystemStorage) StageArtifact(ctx context.Context, maxBytes int64, write func(io.Writer) error) (io.ReadCloser, error) {
	if storage == nil || storage.root == nil {
		return nil, fmt.Errorf("stage backup artifact: storage is unavailable")
	}
	return storage.root.Stage(ctx, maxBytes, write)
}

func (storage *FilesystemStorage) ReadTargetMarker(maxMarkerBytes int64, maxGenerationBytes int64) ([]byte, []byte, error) {
	if storage == nil || storage.root == nil {
		return nil, nil, fmt.Errorf("read recovery marker: storage is unavailable")
	}

	markerBody, _, markerErr := storage.root.ReadRegular(rootedfs.MustParseReference("restore-target-marker.json"), maxMarkerBytes)
	generationBody, _, generationErr := storage.root.ReadRegular(rootedfs.MustParseReference("restore-target-generation"), maxGenerationBytes)
	if markerErr != nil && !errors.Is(markerErr, os.ErrNotExist) {
		return markerBody, generationBody, markerErr
	}
	if generationErr != nil && !errors.Is(generationErr, os.ErrNotExist) {
		return markerBody, generationBody, generationErr
	}
	return markerBody, generationBody, errors.Join(markerErr, generationErr)

}

func (storage *FilesystemStorage) Close() error {
	if storage == nil || storage.root == nil {
		return nil
	}
	err := storage.root.Close()
	storage.root = nil
	return err
}

func makeParent(root *rootedfs.Root, reference rootedfs.Reference) error {
	parent := pathpkg.Dir(reference.String())
	if parent == "." {
		return nil
	}
	parentReference, err := rootedfs.ParseReference(parent)
	if err != nil {
		return fmt.Errorf("resolve backup artifact parent: %w", err)
	}
	if err := root.MakePrivateDir(parentReference); err != nil {
		return fmt.Errorf("create backup artifact parent: %w", err)
	}
	return nil
}

// WriteTargetMarker commits the generation first. An interrupted first issuance
// retains that identity; Recovery must inspect pristine state before finishing it.
func (storage *FilesystemStorage) WriteTargetMarker(ctx context.Context, previousMarker, previousGeneration, marker, generation []byte) error {
	currentMarker, currentGeneration, err := storage.ReadTargetMarker(65536, 64)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if !bytes.Equal(currentMarker, previousMarker) || !bytes.Equal(currentGeneration, previousGeneration) {
		return errors.New("target proof changed during preparation")
	}
	if len(previousGeneration) == 0 {
		if err := storage.root.CreateExclusive(ctx, rootedfs.MustParseReference("restore-target-generation"), func(w io.Writer) error { _, err := w.Write(generation); return err }); err != nil {
			return err
		}
	} else if strings.TrimSpace(string(previousGeneration)) != strings.TrimSpace(string(generation)) {
		return errors.New("target generation cannot change")
	}
	write := func(w io.Writer) error { _, err := w.Write(marker); return err }
	ref := rootedfs.MustParseReference("restore-target-marker.json")
	if len(previousMarker) == 0 {
		return storage.root.CreateExclusive(ctx, ref, write)
	}
	return storage.root.AtomicReplace(ctx, ref, write)
}
