package conflicts

import (
	"crypto/sha256"
	"testing"
	"time"
)

func TestConflictTokenV4RejectsUnadmittedKeys(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	key := sha256.Sum256([]byte("revisions-conflict-token-entropy-failure"))
	ring := &ConflictTokenKeyRing{
		activeKeyID: "active",
		keys: map[string]conflictTokenKeyMaterial{
			"active": {key: key[:], state: conflictTokenKeyStateActive},
		},
	}
	codec, err := NewConflictTokenCodec(
		ring,
		WithClock(func() time.Time { return now }),
	)
	if err != nil {
		t.Fatalf("construct codec: %v", err)
	}
	claims := ConflictTokenClaims{
		RouteKey:                "workbook.records.conflicts.resolve",
		RecordID:                "00000000-0000-4000-8000-000000000001",
		ViewSchemaID:            "cartulary.view.notes.v1",
		FieldKey:                "note.title",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             RequestHashTokenValue([]byte("request")),
	}
	token, err := codec.Issue(claims)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	for _, size := range []int{0, 16, 24, 31, 33, 64} {
		ring.keys["active"] = conflictTokenKeyMaterial{key: make([]byte, size), state: conflictTokenKeyStateActive}
		if _, err := NewConflictTokenCodec(ring); err == nil {
			t.Fatalf("admitted %d-byte key", size)
		}
	}
	other := sha256.Sum256([]byte("other key"))
	ring.keys["active"] = conflictTokenKeyMaterial{key: other[:], state: conflictTokenKeyStateActive}
	wrong, err := NewConflictTokenCodec(ring, WithClock(func() time.Time { return now }))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := wrong.Parse(token); ok {
		t.Fatal("wrong key accepted")
	}
}
