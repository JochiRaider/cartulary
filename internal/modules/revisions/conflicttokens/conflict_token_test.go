package conflicttokens_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestConflictTokenCodecRoundTripsAndRejectsTampering(t *testing.T) {
	codec := conflicttest.NewCodec("revisions-owner")
	recordID := uuid.New()
	claims := conflicttokens.ConflictTokenClaims{
		RouteKey:                "workbook.records.conflicts.resolve",
		RecordID:                recordID.String(),
		ViewSchemaID:            "cartulary.view.notes.v1",
		FieldKey:                "note.title",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       2,
		RequestHash:             conflicttokens.RequestHashTokenValue([]byte("request")),
	}

	token := codec.Issue(claims)
	parsed, ok := codec.Parse(token)
	if !ok {
		t.Fatal("expected issued conflict token to parse")
	}
	if parsed.Version != conflicttokens.ConflictTokenVersion ||
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

func TestConflictTokenCodecRejectsInvalidClaims(t *testing.T) {
	codec := conflicttest.NewCodec("revisions-owner-invalid")
	valid := conflicttokens.ConflictTokenClaims{
		RouteKey:                "timeline.records.conflicts.resolve",
		RecordID:                uuid.NewString(),
		ViewSchemaID:            "cartulary.view.timeline.v2",
		FieldKey:                "timeline.analyst",
		ConflictResolutionClass: "text_compare_merge",
		BaseRowVersion:          1,
		CurrentRowVersion:       1,
		RequestHash:             conflicttokens.RequestHashTokenValue([]byte("request")),
	}

	cases := map[string]func(conflicttokens.ConflictTokenClaims) conflicttokens.ConflictTokenClaims{
		"nil_record": func(claims conflicttokens.ConflictTokenClaims) conflicttokens.ConflictTokenClaims {
			claims.RecordID = uuid.Nil.String()
			return claims
		},
		"missing_route": func(claims conflicttokens.ConflictTokenClaims) conflicttokens.ConflictTokenClaims {
			claims.RouteKey = ""
			return claims
		},
		"missing_field": func(claims conflicttokens.ConflictTokenClaims) conflicttokens.ConflictTokenClaims {
			claims.FieldKey = ""
			return claims
		},
		"current_before_base": func(claims conflicttokens.ConflictTokenClaims) conflicttokens.ConflictTokenClaims {
			claims.CurrentRowVersion = claims.BaseRowVersion - 1
			return claims
		},
		"missing_request_hash": func(claims conflicttokens.ConflictTokenClaims) conflicttokens.ConflictTokenClaims {
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
