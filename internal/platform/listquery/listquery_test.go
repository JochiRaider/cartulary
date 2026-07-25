package listquery

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func TestParseNormalizesSearchAndFilters(t *testing.T) {
	result, queryErr := Parse("search=Beta+alpha+alpha&status=active&limit=1", Config{
		Search: true,
		ExactFilters: map[string]ExactFilter{
			"status": {Allowed: []string{"active", "closed"}},
		},
	})
	if queryErr != nil {
		t.Fatalf("parse valid list query: %#v", queryErr)
	}
	if result.Scope["search"] != "alpha beta" {
		t.Fatalf("unexpected canonical search: %#v", result.Scope)
	}
	if result.Scope["status"] != "active" {
		t.Fatalf("unexpected status scope: %#v", result.Scope)
	}
	if result.Values.Get("limit") != "1" {
		t.Fatalf("pagination members must remain available: %#v", result.Values)
	}
}

func TestParseRejectsListQueryFailures(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"duplicate", "search=a&search=b", ReasonDuplicateQueryMember},
		{"duplicate precedes unknown", "unknown=value&status=active&status=closed", ReasonDuplicateQueryMember},
		{"unknown", "sort=title", ReasonUnknownQueryMember},
		{"invalid search", "search=%00", ReasonInvalidSearch},
		{"zero tokens", "search=---", ReasonInvalidSearch},
		{"token count", "search=" + strings.Join([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q"}, "+"), ReasonSearchTokenCountExceed},
		{"invalid exact filter", "status=open", ReasonInvalidFilterValue},
		{"comma filter", "status=active,closed", ReasonInvalidFilterValue},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, queryErr := Parse(tc.raw, Config{
				Search: true,
				ExactFilters: map[string]ExactFilter{
					"status": {Allowed: []string{"active", "closed"}},
				},
			})
			if queryErr == nil || queryErr.Kind != ErrorKindList || queryErr.ReasonCode != tc.want {
				t.Fatalf("unexpected error: %#v", queryErr)
			}
		})
	}
}

func TestParseRejectsPaginationAliasesAsPaginationErrors(t *testing.T) {
	_, queryErr := Parse("page=2", Config{Search: true})
	if queryErr == nil || queryErr.Kind != ErrorKindPagination || queryErr.ReasonCode != pagination.ReasonInvalidLimit {
		t.Fatalf("unexpected alias error: %#v", queryErr)
	}
}

func TestMatchSearchTokensUsesSourceFieldTokens(t *testing.T) {
	uuidTokens, queryErr := NormalizeSearch("d56d8685-f36e-448c-8f44-bd2978aa26d8")
	if queryErr != nil {
		t.Fatalf("normalize uuid search: %#v", queryErr)
	}
	if !MatchSearchTokens(uuidTokens, "d56d8685-f36e-448c-8f44-bd2978aa26d8") {
		t.Fatal("full hyphenated UUID search must match UUID source-field tokens")
	}
	if MatchSearchTokens(uuidTokens, "d56d8685-f36e-448c-8f44-000000000000") {
		t.Fatal("every query token must match at least one source token")
	}

	emailTokens, queryErr := NormalizeSearch("ADMIN.User Example")
	if queryErr != nil {
		t.Fatalf("normalize email search: %#v", queryErr)
	}
	if !MatchSearchTokens(emailTokens, "admin.user@example.test") {
		t.Fatal("dotted email source fields must match their source tokens")
	}

	keyTokens, queryErr := NormalizeSearch("IR-SEARCH-001")
	if queryErr != nil {
		t.Fatalf("normalize incident key search: %#v", queryErr)
	}
	if !MatchSearchTokens(keyTokens, "IR-SEARCH-001") {
		t.Fatal("hyphenated keys must match their source tokens")
	}

	accentTokens, queryErr := NormalizeSearch("café")
	if queryErr != nil {
		t.Fatalf("normalize accented search: %#v", queryErr)
	}
	if !MatchSearchTokens(accentTokens, "CAFÉ") {
		t.Fatal("case folding must remain locale independent for source tokens")
	}
	if MatchSearchTokens(accentTokens, "CAFE") {
		t.Fatal("diacritics must remain significant")
	}
}
