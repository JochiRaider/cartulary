package recoveryassembly

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	pathpkg "path"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

var ErrUnsupportedBackupBinding = errors.New("recovery assembly: unsupported backup storage binding")

type FilesystemStorage struct {
	root *rootedfs.Root
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

func (storage *FilesystemStorage) ReadMarker(maxBytes int64) ([]byte, error) {
	if storage == nil || storage.root == nil {
		return nil, fmt.Errorf("read recovery marker: storage is unavailable")
	}
	body, _, err := storage.root.ReadRegular(rootedfs.MustParseReference("restore-verification-target.json"), maxBytes)
	if err != nil {
		return nil, fmt.Errorf("read recovery marker: %w", err)
	}
	return body, nil
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
