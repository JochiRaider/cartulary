package blobref_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
)

func TestObjectBlobStorageKeyContract(t *testing.T) {
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000210001")
	objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000210002")
	key, err := blobref.ObjectBlobStorageKey(incidentID, objectBlobID)
	if err != nil {
		t.Fatalf("build key: %v", err)
	}
	want := "incidents/00000000-0000-0000-0000-000000210001/object-blobs/00000000-0000-0000-0000-000000210002"
	if key != want {
		t.Fatalf("key got %q want %q", key, want)
	}
	parts, err := blobref.ParseObjectBlobStorageKey(key)
	if err != nil {
		t.Fatalf("parse key: %v", err)
	}
	if parts.IncidentID != incidentID || parts.ObjectBlobID != objectBlobID {
		t.Fatalf("parsed parts got %#v", parts)
	}

	for _, invalid := range []string{
		"",
		"/incidents/00000000-0000-0000-0000-000000210001/object-blobs/00000000-0000-0000-0000-000000210002",
		"../incidents/00000000-0000-0000-0000-000000210001/object-blobs/00000000-0000-0000-0000-000000210002",
		"incidents//object-blobs/00000000-0000-0000-0000-000000210002",
		"incidents/00000000-0000-0000-0000-000000210001/object-blobs/00000000-0000-0000-0000-000000210002\n",
		"incidents/00000000-0000-0000-0000-000000210001/object-blobs/" + strings.Repeat("a", 1025),
	} {
		if blobref.ValidObjectBlobStorageKey(invalid) {
			t.Fatalf("invalid key accepted: %q", invalid)
		}
	}
}

func TestObjectBlobStorageRefContract(t *testing.T) {
	objectBlobID := uuid.MustParse("00000000-0000-0000-0000-000000210003")
	ref, err := blobref.ObjectBlobStorageRef(objectBlobID)
	if err != nil {
		t.Fatalf("build ref: %v", err)
	}
	if ref != "object://00000000-0000-0000-0000-000000210003" {
		t.Fatalf("ref got %q", ref)
	}
	parsed, err := blobref.ParseObjectBlobStorageRef(ref)
	if err != nil {
		t.Fatalf("parse ref: %v", err)
	}
	if parsed != objectBlobID {
		t.Fatalf("parsed ref got %s want %s", parsed, objectBlobID)
	}
	if !blobref.IsServerManagedStorageRef(" " + ref + " ") {
		t.Fatalf("trimmed server-managed ref was not detected")
	}
	for _, value := range []string{"ticket://collect", "object://NOT-A-UUID", "object://00000000-0000-0000-0000-000000000000"} {
		if blobref.IsServerManagedStorageRef(value) {
			t.Fatalf("non-server ref detected as reserved: %q", value)
		}
	}
}
