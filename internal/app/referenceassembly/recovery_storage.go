package referenceassembly

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

type RecoveryStorage struct {
	storage *RootStorage
	limits  reference_data.Limits
}

var _ recovery.ReferencePackStorage = (*RecoveryStorage)(nil)

func NewRecoveryStorage(temporaryRoot, publishedRoot string, limits reference_data.Limits) (*RecoveryStorage, error) {
	storage, err := NewRootStorage(temporaryRoot, publishedRoot)
	if err != nil {
		return nil, err
	}
	return &RecoveryStorage{storage: storage, limits: limits}, nil
}

func (s *RecoveryStorage) Close() { s.storage.Close() }

type recoveryMemberReader struct {
	io.Reader
	io.Closer
}

func (s *RecoveryStorage) OpenRecoveryObject(ctx context.Context, key string) (io.ReadCloser, error) {
	ref, err := reference_data.ParseStorageRef(key)
	if err != nil {
		return nil, err
	}
	r, size, err := s.storage.OpenPublished(ctx, ref)
	if err != nil {
		return nil, err
	}
	return &recoveryMemberReader{io.NewSectionReader(r, 0, size), r}, nil
}

func (s *RecoveryStorage) StatRecoveryObject(ctx context.Context, key string) (recovery.VNextObjectSourceInfo, error) {
	ref, err := reference_data.ParseStorageRef(key)
	if err != nil {
		return recovery.VNextObjectSourceInfo{}, err
	}
	r, size, err := s.storage.OpenPublished(ctx, ref)
	if err != nil {
		return recovery.VNextObjectSourceInfo{}, err
	}
	if err := r.Close(); err != nil {
		return recovery.VNextObjectSourceInfo{}, err
	}
	return recovery.VNextObjectSourceInfo{PlaintextBytes: size, ContentType: "application/octet-stream"}, nil
}

func (s *RecoveryStorage) RestoreMember(ctx context.Context, key, digest string, size int64, source io.Reader) error {
	if size < 0 || size > max(s.limits.ReferencePacks.MaxExtractedBytes, s.limits.ReferencePacks.MaxContainerBytes, s.limits.Archives.DefaultMaxExtractedBytes) || !referencePackSHA256Pattern.MatchString(digest) || source == nil {
		return errArtifactIntegrity
	}
	// The retained opaque identity is restored exactly. It has the same closed
	// namespace and digest/UUID shape as normal publication, never an archive path.
	token, ok := strings.CutPrefix(key, "reference-packs/objects/"+digest+"-")
	if !ok {
		return errArtifactIntegrity
	}
	id, err := uuid.Parse(token)
	if err != nil || id.String() != token {
		return errArtifactIntegrity
	}
	ref, err := rootedfs.ParseReference(key)
	if err != nil {
		return err
	}
	if err := s.storage.published.MakePrivateDir(rootedfs.MustParseReference("reference-packs/objects")); err != nil {
		return err
	}
	return s.storage.published.CreateExclusive(ctx, ref, func(w io.Writer) error {
		h := sha256.New()
		n, err := copyArtifact(ctx, io.MultiWriter(w, h), source, size)
		if err != nil {
			return err
		}
		if n != size || hex.EncodeToString(h.Sum(nil)) != digest {
			return errArtifactIntegrity
		}
		return nil
	})
}

func (s *RecoveryStorage) RequireEmpty(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	files, err := s.storage.published.ListRegular()
	if err != nil {
		return err
	}
	if len(files) != 0 {
		return errors.New("reference pack restore target is not empty")
	}
	return nil
}

// The caller owns Recovery's exclusive disposable-target admission. No path
// recursion or deletion outside the admitted root is available here.
func (s *RecoveryStorage) ResetVerificationTarget(ctx context.Context) error {
	files, err := s.storage.published.ListRegular()
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.storage.published.RemoveRegular(file.Reference); err != nil {
			return err
		}
	}
	return nil
}

func (s *RecoveryStorage) ValidateHistoricalState(ctx context.Context, db postgres.DB) error {
	return reference_data.RestoreHistoricalState(ctx, db, s.storage, s.limits)
}
