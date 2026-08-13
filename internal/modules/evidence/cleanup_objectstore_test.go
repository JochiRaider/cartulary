package evidence

import (
	"context"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestEvidenceCleanupObjectDeleterUsesTypedPurposeAndPreservesNotFound(t *testing.T) {
	typed := &cleanupTypedStoreFixture{}
	deleter, err := newCleanupObjectDeleter(typed)
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

func TestEvidenceCleanupObjectDeleterRequiresTypedStore(t *testing.T) {
	field, ok := reflect.TypeOf(OwnerRuntimeDependencies{}).FieldByName("ObjectStore")
	if !ok {
		t.Fatal("Evidence owner runtime is missing its object-store dependency")
	}
	typedStore := reflect.TypeOf((*objectstore.TypedStore)(nil)).Elem()
	if field.Type != typedStore {
		t.Fatalf("Evidence owner object-store dependency = %v, want %v", field.Type, typedStore)
	}
}

type cleanupTypedStoreFixture struct {
	objectstore.TypedStore
	request objectstore.DeleteObjectRequest
	err     error
}

func (store *cleanupTypedStoreFixture) Delete(_ context.Context, request objectstore.DeleteObjectRequest) error {
	store.request = request
	return store.err
}
