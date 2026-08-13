package evidence

import (
	"context"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestEvidenceCleanupObjectDeleterUsesTypedPurposeAndPreservesNotFound(t *testing.T) {
	typed := &cleanupTypedStoreFixture{}
	deleter, err := NewCleanupObjectDeleter(typed)
	if err != nil {
		t.Fatalf("compose typed cleanup deleter: %v", err)
	}
	if err := deleter.DeleteObject(context.Background(), "incidents/test/object-blobs/test"); err != nil {
		t.Fatalf("typed cleanup delete: %v", err)
	}
	if typed.request.Key != "incidents/test/object-blobs/test" || typed.request.Purpose != objectstore.PurposeEvidenceCleanup {
		t.Fatalf("typed cleanup request = %#v", typed.request)
	}

	typed.err = &objectstore.AdapterError{
		Code:      objectstore.ErrorCodeObjectNotFound,
		Reason:    objectstore.ReasonObjectMissing,
		Operation: objectstore.OperationDeleteObject,
	}
	err = deleter.DeleteObject(context.Background(), "incidents/test/object-blobs/missing")
	if !objectstore.IsObjectNotFound(err) {
		t.Fatalf("typed not-found classification was not preserved: %v", err)
	}
}

func TestEvidenceCleanupObjectDeleterRejectsLegacyOnlyStore(t *testing.T) {
	if _, err := NewCleanupObjectDeleter(cleanupLegacyStoreFixture{}); err == nil {
		t.Fatal("legacy-only object store was accepted for Evidence cleanup")
	}
}

type cleanupTypedStoreFixture struct {
	objectstore.Store
	request objectstore.DeleteObjectRequest
	err     error
}

func (store *cleanupTypedStoreFixture) Delete(_ context.Context, request objectstore.DeleteObjectRequest) error {
	store.request = request
	return store.err
}

type cleanupLegacyStoreFixture struct {
	objectstore.Store
}
