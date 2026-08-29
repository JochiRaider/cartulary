package rollback

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestSourceForRollbackValuePreservesPresenceAndAssociations(t *testing.T) {
	t.Parallel()
	value := map[string]any{"source": map[string]any{
		"title":           "Memory image",
		"source_party_id": nil,
	}}
	got, ok := sourceForRollbackValue(value)
	if !ok {
		t.Fatal("sourceForRollbackValue returned ok=false")
	}
	want := map[string]any{"title": "Memory image", "source_party_id": nil}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("evidence source = %#v, want %#v", got, want)
	}
	if _, present := got["object_blob_id"]; present {
		t.Fatal("absent blob identity unexpectedly became present")
	}
	if _, ok := sourceForRollbackValue(map[string]any{"cells": map[string]any{"evidence.title": map[string]any{"value": "unsupported"}}}); ok {
		t.Fatal("schema-less projection row was accepted")
	}
}

func TestProviderRejectsInvalidLifecycleAndReferences(t *testing.T) {
	t.Parallel()
	provider := NewProvider()
	objectBlobID := uuid.New()
	otherObjectBlobID := uuid.New()
	storageRef, err := blobref.ObjectBlobStorageRef(objectBlobID)
	if err != nil {
		t.Fatal(err)
	}
	mismatchedStorageRef, err := blobref.ObjectBlobStorageRef(otherObjectBlobID)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []map[string]any{
		{"source": map[string]any{"lifecycle_state": "invalid"}},
		{"source": map[string]any{"upload_state": "invalid"}},
		{"source": map[string]any{"collector_party_id": "not-a-uuid"}},
		{"source": map[string]any{"lifecycle_state": "available", "upload_state": "available", "object_blob_id": nil}},
		{"source": map[string]any{"lifecycle_state": "released", "upload_state": "quarantined", "object_blob_id": objectBlobID.String()}},
		{"source": map[string]any{"lifecycle_state": "quarantined", "upload_state": "available", "object_blob_id": objectBlobID.String()}},
		{"source": map[string]any{"lifecycle_state": "received", "upload_state": "available", "object_blob_id": objectBlobID.String(), "storage_ref": mismatchedStorageRef}},
		{"source": map[string]any{"lifecycle_state": "received", "upload_state": "available", "object_blob_id": nil, "storage_ref": storageRef}},
		{"source": map[string]any{"lifecycle_state": "received", "upload_state": "available", "object_blob_id": nil, "storage_ref": "object://not-a-uuid"}},
	} {
		if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
			t.Fatalf("ValidateRollbackValue error = %v, want target-not-reversible", err)
		}
	}
	for _, source := range []map[string]any{
		{"lifecycle_state": "requested", "upload_state": "pending", "object_blob_id": nil, "storage_ref": "ticket://external"},
		{"lifecycle_state": "received", "upload_state": "available", "object_blob_id": objectBlobID.String(), "storage_ref": storageRef},
		{"lifecycle_state": "available", "upload_state": "available", "object_blob_id": objectBlobID.String(), "storage_ref": storageRef},
		{"lifecycle_state": "quarantined", "upload_state": "quarantined", "object_blob_id": objectBlobID.String(), "storage_ref": storageRef},
	} {
		if err := provider.ValidateRollbackValue(map[string]any{"source": source}); err != nil {
			t.Errorf("ValidateRollbackValue(%#v) = %v", source, err)
		}
	}
}
