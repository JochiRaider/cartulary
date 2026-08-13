package evidence

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

type cleanupTypedObjectStore interface {
	Delete(context.Context, objectstore.DeleteObjectRequest) error
}

type cleanupObjectStoreDeleter struct {
	store cleanupTypedObjectStore
}

func NewCleanupObjectDeleter(store objectstore.Store) (CleanupObjectDeleter, error) {
	if store == nil {
		return nil, errors.New("compose Evidence cleanup object deleter: object store is required")
	}
	typed, ok := store.(cleanupTypedObjectStore)
	if !ok {
		return nil, errors.New("compose Evidence cleanup object deleter: typed delete capability is required")
	}
	return cleanupObjectStoreDeleter{store: typed}, nil
}

func (deleter cleanupObjectStoreDeleter) DeleteObject(ctx context.Context, key string) error {
	return deleter.store.Delete(ctx, objectstore.DeleteObjectRequest{
		Key:     key,
		Purpose: objectstore.PurposeEvidenceCleanup,
	})
}
