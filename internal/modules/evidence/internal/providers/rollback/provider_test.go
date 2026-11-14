package rollback

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestSourceForRollbackValuePreservesPresenceAndAssociations(t *testing.T) {
	t.Parallel()
	value := evidenceRollbackFixture(map[string]any{"title": "Memory image"})
	before := evidenceRollbackFixture(map[string]any{"title": "Memory image"})
	got, ok := sourceForRollbackValue(value)
	if !ok || len(got) != 17 || got["title"] != "Memory image" || got["source_party_id"] != nil || got["object_blob_id"] != nil || got["owner_user_id"] != nil {
		t.Fatalf("complete Evidence source lost retained facts: %#v", got)
	}
	if !reflect.DeepEqual(value, before) {
		t.Fatal("retained source was rewritten")
	}
	partial := map[string]any{"source": map[string]any{"title": "Memory image"}}
	retired := evidenceRollbackFixture(nil)
	retired["snapshot_schema_id"] = "cartulary.revisions.snapshot.evidence.v1"
	missingOwner := evidenceRollbackFixture(nil)
	delete(missingOwner["source"].(map[string]any), "owner_user_id")
	for _, invalid := range []map[string]any{partial, retired, missingOwner, {"cells": map[string]any{"evidence.title": map[string]any{"value": "unsupported"}}}} {
		if _, ok := sourceForRollbackValue(invalid); ok {
			t.Fatal("incomplete or retired Evidence source accepted")
		}
	}
}

func TestProviderRejectsInvalidLifecycleAndReferences(t *testing.T) {
	t.Parallel()
	provider := NewProvider()
	for _, key := range []string{"collector_party_id", "source_party_id", "object_blob_id", "owner_user_id"} {
		t.Run(key, func(t *testing.T) {
			for _, raw := range []any{"", " \t", "not-a-uuid", uuid.Nil.String(), 42, true, []any{}, map[string]any{}} {
				value := evidenceRollbackFixture(map[string]any{key: raw})
				if err := provider.ValidateRollbackValue(value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
					t.Fatalf("ValidateRollbackValue(%#v) = %v, want target-not-reversible", raw, err)
				}
				if err := provider.RestoreTx(context.Background(), nil, rollbackcontract.RestoreRequest{RetainedValue: value}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
					t.Fatalf("RestoreTx(%#v) = %v, want rejection before database access", raw, err)
				}
			}
			id := uuid.MustParse("542e8768-d0c8-4186-9794-387959fc705b")
			for _, tc := range []struct {
				source  map[string]any
				present bool
				want    any
			}{
				{source: map[string]any{"title": "Retained evidence"}},
				{source: map[string]any{key: nil}, present: true},
				{source: map[string]any{key: id.String()}, present: true, want: id},
			} {
				if err := provider.ValidateRollbackValue(evidenceRollbackFixture(tc.source)); err != nil {
					t.Fatalf("valid association %#v rejected: %v", tc.source, err)
				}
				got, present, err := nullableUUID(tc.source, key)
				if err != nil || present != tc.present || got != tc.want {
					t.Fatalf("association %#v decoded as (%#v, %v, %v), want (%#v, %v, nil)", tc.source, got, present, err, tc.want, tc.present)
				}
			}
		})
	}
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
		value = evidenceRollbackFixture(value["source"].(map[string]any))
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
		if err := provider.ValidateRollbackValue(evidenceRollbackFixture(source)); err != nil {
			t.Errorf("ValidateRollbackValue(%#v) = %v", source, err)
		}
	}
}

func evidenceRollbackFixture(overrides map[string]any) map[string]any {
	const id = "20000000-0000-4000-8000-000000000001"
	const incident = "30000000-0000-4000-8000-000000000001"
	source := map[string]any{
		"record_id": id, "incident_id": incident, "title": nil, "lifecycle_state": "requested",
		"requested_at": nil, "received_at": nil, "storage_ref": nil, "blob_hash": nil,
		"collector_party_text": nil, "collector_party_id": nil, "source_party_text": nil, "source_party_id": nil,
		"upload_state": "pending", "object_blob_id": nil, "owner_user_id": nil,
		"created_at": "2026-11-13T00:00:00Z", "updated_at": "2026-11-13T00:00:00Z",
	}
	for key, value := range overrides {
		source[key] = value
	}
	return map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.evidence.v2", "record": map[string]any{"record_id": id, "incident_id": incident}, "source": source}
}
