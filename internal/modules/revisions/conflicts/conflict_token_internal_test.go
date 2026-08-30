package conflicts

import (
	"crypto/sha256"
	"errors"
	"io"
	"testing"
	"time"
)

func TestConflictTokenV3PropagatesEntropyFailure(t *testing.T) {
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
		withEntropySource(failingEntropyReader{}),
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
	if _, err := codec.Issue(claims); !errors.Is(err, errConflictTokenUnavailable) {
		t.Fatalf("entropy failure = %v, want closed unavailable error", err)
	}
}

type failingEntropyReader struct{}

func (failingEntropyReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
