package revisions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestConflictTokenCodecRoundTripsV2Claims(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("unit")
	claims := ConflictTokenClaims{
		RouteKey:                "route-a",
		RecordID:                uuid.NewString(),
		ViewSchemaID:            "cartulary.view.example.v1",
		FieldKey:                "example.title",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          2,
		CurrentRowVersion:       3,
		RequestHash:             RequestHashTokenValue([]byte("request")),
	}

	parsed, ok := codec.Parse(codec.Issue(claims))
	if !ok {
		t.Fatal("expected issued token to parse")
	}
	if parsed.Version != ConflictTokenVersion {
		t.Fatalf("version = %d, want %d", parsed.Version, ConflictTokenVersion)
	}
	if parsed.RecordID != claims.RecordID || parsed.RouteKey != claims.RouteKey || parsed.FieldKey != claims.FieldKey {
		t.Fatalf("parsed claims mismatch: %#v", parsed)
	}
}

func TestConflictTokenCodecRejectsTampering(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("unit")
	token := codec.Issue(ConflictTokenClaims{
		RouteKey:                "route-a",
		RecordID:                uuid.NewString(),
		ViewSchemaID:            "cartulary.view.example.v1",
		FieldKey:                "example.title",
		ConflictResolutionClass: "atomic_replace",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             RequestHashTokenValue([]byte("request")),
	})
	tampered := strings.TrimRight(token, "A") + "A"
	if tampered == token {
		tampered = token[:len(token)-1] + "B"
	}
	if _, ok := codec.Parse(tampered); ok {
		t.Fatal("expected tampered token to be rejected")
	}
}

func TestConflictTokenCodecRejectsInvalidRecordID(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("unit")
	token := codec.Issue(ConflictTokenClaims{
		RouteKey:                "route-a",
		RecordID:                "not-a-uuid",
		ViewSchemaID:            "cartulary.view.example.v1",
		FieldKey:                "example.title",
		ConflictResolutionClass: "atomic_replace",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             RequestHashTokenValue([]byte("request")),
	})
	if _, ok := codec.Parse(token); ok {
		t.Fatal("expected invalid record id to be rejected")
	}
}
