package incidentportabilityassembly

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

// RecoveryStorage preserves published bundle references in the configured
// export root. It never interprets an object-store key as a filesystem path.
type RecoveryStorage struct{ root *rootedfs.Root }

var _ recovery.RootObjectStorage = (*RecoveryStorage)(nil)

func NewRecoveryStorage(exportRoot string) (*RecoveryStorage, error) {
	root, err := rootedfs.OpenOrCreate(exportRoot)
	if err != nil {
		return nil, err
	}
	return &RecoveryStorage{root: root}, nil
}
func (s *RecoveryStorage) Close() { _ = s.root.Close() }

func recoveryBundleReference(key string) (rootedfs.Reference, error) {
	token, ok := strings.CutPrefix(key, "incident-bundles/")
	token, suffix := strings.CutSuffix(token, ".zip")
	id, err := uuid.Parse(token)
	if !ok || !suffix || err != nil || id.String() != token {
		return rootedfs.Reference{}, errors.New("invalid retained bundle reference")
	}
	return rootedfs.ParseReference(key)
}
func (s *RecoveryStorage) OpenRecoveryObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ref, err := recoveryBundleReference(key)
	if err != nil {
		return nil, err
	}
	file, _, err := s.root.OpenRegular(ref)
	return file, err
}
func (s *RecoveryStorage) StatRecoveryObject(ctx context.Context, key string) (recovery.VNextObjectSourceInfo, error) {
	if err := ctx.Err(); err != nil {
		return recovery.VNextObjectSourceInfo{}, err
	}
	ref, err := recoveryBundleReference(key)
	if err != nil {
		return recovery.VNextObjectSourceInfo{}, err
	}
	file, metadata, err := s.root.OpenRegular(ref)
	if err != nil {
		return recovery.VNextObjectSourceInfo{}, err
	}
	defer file.Close()
	return recovery.VNextObjectSourceInfo{PlaintextBytes: metadata.Size, ContentType: "application/zip"}, nil
}
func (s *RecoveryStorage) RestoreMember(ctx context.Context, key, digest string, size int64, source io.Reader) error {
	if size < 0 || size == math.MaxInt64 || !incidentBundleSHA256Pattern.MatchString(digest) || source == nil {
		return errors.New("invalid retained bundle integrity")
	}
	ref, err := recoveryBundleReference(key)
	if err != nil {
		return err
	}
	if err := s.root.MakePrivateDir(rootedfs.MustParseReference("incident-bundles")); err != nil {
		return err
	}
	return s.root.CreateExclusive(ctx, ref, func(w io.Writer) error {
		hash := sha256.New()
		n, err := io.Copy(io.MultiWriter(w, hash), io.LimitReader(source, size+1))
		if err != nil {
			return err
		}
		if n != size || hex.EncodeToString(hash.Sum(nil)) != digest {
			return errors.New("retained bundle integrity mismatch")
		}
		return ctx.Err()
	})
}
func (s *RecoveryStorage) RequireEmpty(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	files, err := s.root.ListRegular()
	if err != nil {
		return err
	}
	if len(files) != 0 {
		return errors.New("export restore root is not empty")
	}
	return nil
}
func (s *RecoveryStorage) ResetVerificationTarget(ctx context.Context) error {
	files, err := s.root.ListRegular()
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.root.RemoveRegular(file.Reference); err != nil {
			return err
		}
	}
	return nil
}
