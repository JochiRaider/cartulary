package evidence

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type cleanupObjectStoreDeleter struct {
	store objectstore.TypedStore
}

func newCleanupObjectDeleter(store objectstore.TypedStore) (cleanupObjectDeleter, error) {
	if store == nil {
		return nil, errors.New("compose Evidence cleanup object deleter: object store is required")
	}
	return cleanupObjectStoreDeleter{store: store}, nil
}

func (deleter cleanupObjectStoreDeleter) DeleteObject(ctx context.Context, key string) error {
	return deleter.store.Delete(ctx, objectstore.DeleteObjectRequest{
		Key:     key,
		Purpose: objectstore.PurposeEvidenceCleanup,
	})
}
