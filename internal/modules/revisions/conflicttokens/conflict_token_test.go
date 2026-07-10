package conflicttokens

import (
	"testing"

	"github.com/google/uuid"
)

func TestSupportPhase6_ConflictTokenCodecRoundTripsAndRejectsTampering(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("revisions-owner")
	recordID := uuid.New()
	claims := ConflictTokenClaims{
		RouteKey:                "workbook.records.conflicts.resolve",
		RecordID:                recordID.String(),
		ViewSchemaID:            "cartulary.view.notes.v1",
		FieldKey:                "note.title",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             RequestHashTokenValue([]byte("request")),
	}

	token := codec.Issue(claims)
	parsed, ok := codec.Parse(token)
	if !ok {
		t.Fatal("expected issued conflict token to parse")
	}
	if parsed.Version != ConflictTokenVersion ||
		parsed.RouteKey != claims.RouteKey ||
		parsed.RecordID != claims.RecordID ||
		parsed.ViewSchemaID != claims.ViewSchemaID ||
		parsed.FieldKey != claims.FieldKey ||
		parsed.ConflictResolutionClass != claims.ConflictResolutionClass ||
		parsed.BaseRowVersion != claims.BaseRowVersion ||
		parsed.CurrentRowVersion != claims.CurrentRowVersion ||
		parsed.RequestHash != claims.RequestHash ||
		parsed.Signature == "" {
		t.Fatalf("unexpected parsed conflict token claims: %#v", parsed)
	}
	if _, ok := codec.Parse(token + "x"); ok {
		t.Fatal("tampered conflict token parsed")
	}
	if _, ok := codec.Parse(token[:len(token)-1]); ok {
		t.Fatal("truncated conflict token parsed")
	}
}

func TestSupportPhase6_ConflictTokenCodecRejectsInvalidClaims(t *testing.T) {
	codec := NewConflictTokenCodecForTesting("revisions-owner-invalid")
	valid := ConflictTokenClaims{
		RouteKey:                "timeline.records.conflicts.resolve",
		RecordID:                uuid.NewString(),
		ViewSchemaID:            "cartulary.view.timeline.v2",
		FieldKey:                "timeline.analyst",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       1,
		RequestHash:             RequestHashTokenValue([]byte("request")),
	}

	cases := map[string]func(ConflictTokenClaims) ConflictTokenClaims{
		"nil_record": func(claims ConflictTokenClaims) ConflictTokenClaims {
			claims.RecordID = uuid.Nil.String()
			return claims
		},
		"missing_route": func(claims ConflictTokenClaims) ConflictTokenClaims {
			claims.RouteKey = ""
			return claims
		},
		"missing_field": func(claims ConflictTokenClaims) ConflictTokenClaims {
			claims.FieldKey = ""
			return claims
		},
		"current_before_base": func(claims ConflictTokenClaims) ConflictTokenClaims {
			claims.CurrentRowVersion = claims.BaseRowVersion - 1
			return claims
		},
		"missing_request_hash": func(claims ConflictTokenClaims) ConflictTokenClaims {
			claims.RequestHash = ""
			return claims
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, ok := codec.Parse(codec.Issue(mutate(valid))); ok {
				t.Fatalf("invalid %s conflict token parsed", name)
			}
		})
	}
}
