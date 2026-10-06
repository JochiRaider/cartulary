package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"
)

const backupStorageAnchorScheme = "backup-storage://"

var ErrInvalidBackupArtifact = errors.New("recovery: invalid backup artifact")

type BackupStorage interface {
	WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (BackupArtifactProof, error)
	ReadArtifact(ctx context.Context, key string, maxBytes int64) ([]byte, error)
}

type BackupArtifactProof struct {
	Key         string `json:"key"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
}

func CloseBackupStorage(storage BackupStorage) error {
	if closer, ok := storage.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

func VerifyArtifactProof(ctx context.Context, storage BackupStorage, proof BackupArtifactProof) ([]byte, error) {
	if storage == nil {
		return nil, fmt.Errorf("%w: backup storage is required", ErrInvalidBackupArtifact)
	}
	body, err := storage.ReadArtifact(ctx, proof.Key, proof.SizeBytes)
	if err != nil {
		return nil, err
	}
	if int64(len(body)) != proof.SizeBytes {
		return nil, fmt.Errorf("%w: artifact size mismatch for %s", ErrInvalidBackupArtifact, proof.Key)
	}
	if sha256Hex(body) != proof.SHA256 {
		return nil, fmt.Errorf("%w: artifact sha256 mismatch for %s", ErrInvalidBackupArtifact, proof.Key)
	}
	return body, nil
}

func backupTimestamp(value time.Time) time.Time {
	return value.UTC().Truncate(time.Microsecond)
}

func normalizeArtifactKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", fmt.Errorf("%w: artifact key is required", ErrInvalidBackupArtifact)
	}
	if path.IsAbs(trimmed) {
		return "", fmt.Errorf("%w: artifact key must be relative", ErrInvalidBackupArtifact)
	}
	normalized := path.Clean(trimmed)
	if normalized == "." || normalized == ".." || strings.HasPrefix(normalized, "../") || strings.Contains(normalized, "/../") {
		return "", fmt.Errorf("%w: artifact key escapes backup storage root", ErrInvalidBackupArtifact)
	}
	return normalized, nil
}

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
