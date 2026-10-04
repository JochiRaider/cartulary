package referenceassembly

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/google/uuid"
)

var publishedParent = rootedfs.MustParseReference("reference-packs/objects")
var publishedObjectName = regexp.MustCompile(`^[a-f0-9]{64}-[a-f0-9-]{36}$`)
var publicationTemporaryName = regexp.MustCompile(`^\.cartulary-tmp-[a-f0-9]{32}$`)

// Filesystem ownership precedes database retention guards throughout the
// publication/collection path. The namespace directory is never replaced.
func (s *RootStorage) WithPublishedCollection(ctx context.Context, run func(reference_data.PublishedObjects) error) (resultErr error) {
	if s == nil || s.published == nil || run == nil {
		return errors.New("reference pack: incomplete collection dependencies")
	}
	if err := s.published.MakePrivateDir(publishedParent); err != nil {
		return err
	}
	lease, err := s.published.LockDirectory(ctx, publishedParent)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, lease.Close()) }()
	return run(publishedObjects{s})
}

type publishedObjects struct{ storage *RootStorage }

func (p publishedObjects) Visit(ctx context.Context, visit func(reference_data.StorageRef) error) error {
	return p.storage.published.VisitDirectory(ctx, publishedParent, func(entry rootedfs.DirectoryEntry) error {
		name := strings.TrimPrefix(entry.Reference.String(), publishedParent.String()+"/")
		if entry.Directory {
			return errors.New("reference pack: unexpected published object")
		}
		if !publicationTemporaryName.MatchString(name) {
			if !publishedObjectName.MatchString(name) {
				return errors.New("reference pack: unexpected published object")
			}
			id, err := uuid.Parse(name[65:])
			if err != nil || id == uuid.Nil || id.String() != name[65:] {
				return errors.New("reference pack: unexpected published object")
			}
		}
		ref, err := reference_data.ParseStorageRef(entry.Reference.String())
		if err != nil {
			return err
		}
		return visit(ref)
	})
}
func (p publishedObjects) Remove(ref reference_data.StorageRef) error {
	return p.storage.RemovePublished(ref)
}
