package conflicts

import "testing"

func TestConflictTokenCodecRoundTripsV2Claims(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("unit")
	claims := ConflictTokenClaims{
		RouteKey:                "workbook.conflicts.resolve",
		RecordID:                "11111111-1111-4111-8111-111111111111",
		ViewSchemaID:            "cartulary.view.timeline.v2",
		FieldKey:                "timeline.raw_activity_text",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          3,
		CurrentRowVersion:       5,
		RequestHash:             "abc",
	}
	token := codec.Issue(claims)
	parsed, ok := codec.Parse(token)
	if !ok {
		t.Fatalf("parse returned false for issued token %q", token)
	}
	if parsed.Version != ConflictTokenVersion {
		t.Fatalf("version = %d, want %d", parsed.Version, ConflictTokenVersion)
	}
	if parsed.RouteKey != claims.RouteKey || parsed.RecordID != claims.RecordID || parsed.FieldKey != claims.FieldKey || parsed.RequestHash != claims.RequestHash {
		t.Fatalf("parsed claims mismatch: %#v", parsed)
	}
}

func TestConflictTokenCodecRejectsTampering(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("unit")
	token := codec.Issue(ConflictTokenClaims{
		RouteKey:                "workbook.conflicts.resolve",
		RecordID:                "11111111-1111-4111-8111-111111111111",
		ViewSchemaID:            "cartulary.view.timeline.v2",
		FieldKey:                "timeline.raw_activity_text",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             "abc",
	})
	tampered := token[:len(token)-1] + "A"
	if _, ok := codec.Parse(tampered); ok {
		t.Fatalf("tampered token parsed successfully")
	}
}

func TestConflictTokenCodecRejectsInvalidRecordID(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("unit")
	token := codec.Issue(ConflictTokenClaims{
		RouteKey:                "workbook.conflicts.resolve",
		RecordID:                "not-a-uuid",
		ViewSchemaID:            "cartulary.view.timeline.v2",
		FieldKey:                "timeline.raw_activity_text",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             "abc",
	})
	if _, ok := codec.Parse(token); ok {
		t.Fatalf("invalid record id token parsed successfully")
	}
}
