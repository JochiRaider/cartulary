package referenceassembly

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

var referencePackSHA256Pattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// RootStorage realizes the Reference Data storage port with admitted,
// descriptor-anchored filesystem roots.
type RootStorage struct {
	temporary *rootedfs.Root
	published *rootedfs.Root
}

// NewRootStorage opens the temporary and published-pack root capabilities.
// The caller owns the returned storage and must close it.
func NewRootStorage(temporaryRoot string, publishedRoot string) (*RootStorage, error) {
	temporary, err := rootedfs.OpenOrCreate(temporaryRoot)
	if err != nil {
		return nil, fmt.Errorf("open reference pack temporary capability: %w", err)
	}
	published, err := rootedfs.OpenOrCreate(publishedRoot)
	if err != nil {
		_ = temporary.Close()
		return nil, fmt.Errorf("open reference pack storage capability: %w", err)
	}
	return &RootStorage{temporary: temporary, published: published}, nil
}

func (storage *RootStorage) Close() {
	if storage == nil {
		return
	}
	_ = storage.published.Close()
	_ = storage.temporary.Close()
}

func (storage *RootStorage) RemoveStaged(reference reference_data.StagingRef) error {
	if storage == nil || storage.temporary == nil {
		return errors.New("reference pack temporary storage is unavailable")
	}
	return removeReferencePackRegular(storage.temporary, reference.String())
}

func (storage *RootStorage) RemovePublished(reference reference_data.StorageRef) error {
	if storage == nil || storage.published == nil {
		return errors.New("reference pack published storage is unavailable")
	}
	return removeReferencePackRegular(storage.published, reference.String())
}

func removeReferencePackRegular(root *rootedfs.Root, rawReference string) error {
	reference, err := rootedfs.ParseReference(rawReference)
	if err != nil {
		return err
	}
	err = root.RemoveRegular(reference)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// OpenIncoming admits exactly one operator filename beneath the already
// confined storage root. RootedFS rejects links and non-regular objects before
// opening bytes; incoming names never become published object references.
func (storage *RootStorage) OpenIncoming(ctx context.Context, name string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if storage == nil || storage.published == nil || !reference_data.ValidOperatorBundleName(name) {
		return nil, reference_data.ErrInvalidReferencePackStorageReference
	}
	ref, err := rootedfs.ParseReference("incoming/" + name)
	if err != nil {
		return nil, err
	}
	file, _, err := storage.published.OpenRegular(ref)
	return file, err
}
