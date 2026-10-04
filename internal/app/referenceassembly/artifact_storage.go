package referenceassembly

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

var _ reference_data.ArtifactStorage = (*RootStorage)(nil)

var errArtifactBound = errors.New("reference pack: artifact byte bound exceeded")
var errArtifactIntegrity = errors.New("reference pack: artifact integrity mismatch")

func (s *RootStorage) StageStream(ctx context.Context, source io.Reader, maximum int64) (reference_data.StagingRef, string, int64, error) {
	if s == nil || s.temporary == nil || source == nil || maximum < 1 {
		return reference_data.StagingRef{}, "", 0, errors.New("reference pack: invalid staging request")
	}
	dir := rootedfs.MustParseReference("reference-packs/imports")
	if err := s.temporary.MakePrivateDir(dir); err != nil {
		return reference_data.StagingRef{}, "", 0, err
	}
	ref := rootedfs.MustParseReference(dir.String() + "/" + uuid.NewString() + ".bundle")
	var size int64
	hash := sha256.New()
	err := s.temporary.CreateExclusive(ctx, ref, func(destination io.Writer) error {
		var err error
		size, err = copyArtifact(ctx, io.MultiWriter(destination, hash), source, maximum)
		return err
	})
	if err != nil {
		return reference_data.StagingRef{}, "", 0, err
	}
	staged, err := reference_data.ParseStagingRef(ref.String())
	return staged, hex.EncodeToString(hash.Sum(nil)), size, err
}

func (s *RootStorage) PublishStream(ctx context.Context, digest string, size int64, source io.Reader) (reference_data.StorageRef, io.Closer, error) {
	if s == nil || s.published == nil || source == nil || size < 0 || !referencePackSHA256Pattern.MatchString(digest) {
		return reference_data.StorageRef{}, nil, errors.New("reference pack: invalid publication request")
	}
	dir := rootedfs.MustParseReference("reference-packs/objects")
	if err := s.published.MakePrivateDir(dir); err != nil {
		return reference_data.StorageRef{}, nil, err
	}
	lease, err := s.published.LockSharedDirectory(ctx, dir)
	if err != nil {
		return reference_data.StorageRef{}, nil, err
	}
	// A unique physical object makes concurrent rollback cleanup incapable of
	// deleting another operation's successful publication of the same digest.
	ref := rootedfs.MustParseReference(dir.String() + "/" + digest + "-" + uuid.NewString())
	err = s.published.CreateExclusive(ctx, ref, func(destination io.Writer) error {
		hash := sha256.New()
		n, err := copyArtifact(ctx, io.MultiWriter(destination, hash), source, size)
		if err != nil {
			return err
		}
		if n != size || hex.EncodeToString(hash.Sum(nil)) != digest {
			return errArtifactIntegrity
		}
		return nil
	})
	if err != nil {
		return reference_data.StorageRef{}, nil, errors.Join(err, lease.Close())
	}
	retained, err := reference_data.ParseStorageRef(ref.String())
	return retained, lease, err
}

// copyArtifact never computes maximum+1, which could overflow an admitted
// int64 byte limit. It probes exactly one additional byte without publishing it.
func copyArtifact(ctx context.Context, destination io.Writer, source io.Reader, maximum int64) (int64, error) {
	buffer := make([]byte, 32768)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		remaining := maximum - written
		if remaining == 0 {
			var extra [1]byte
			n, err := source.Read(extra[:])
			if n != 0 {
				return written, errArtifactBound
			}
			if errors.Is(err, io.EOF) {
				return written, nil
			}
			if err != nil {
				return written, err
			}
			continue
		}
		next := buffer
		if remaining < int64(len(next)) {
			next = next[:remaining]
		}
		n, err := source.Read(next)
		if n > 0 {
			count, writeErr := destination.Write(next[:n])
			written += int64(count)
			if writeErr != nil {
				return written, writeErr
			}
			if count != n {
				return written, io.ErrShortWrite
			}
		}
		if errors.Is(err, io.EOF) {
			return written, nil
		}
		if err != nil {
			return written, err
		}
	}
}
