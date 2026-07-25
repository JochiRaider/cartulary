package listquery

import (
	"net/url"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

const (
	ErrorKindList       = "list"
	ErrorKindPagination = "pagination"

	ReasonDuplicateQueryMember   = "duplicate_query_member"
	ReasonInvalidFilterRange     = "invalid_filter_range"
	ReasonInvalidFilterValue     = "invalid_filter_value"
	ReasonInvalidSearch          = "invalid_search"
	ReasonSearchTokenCountExceed = "search_token_count_exceeded"
	ReasonUnknownQueryMember     = "unknown_query_member"
)

type ExactFilter struct {
	Allowed []string
}

type RangeFilter struct{}

type Config struct {
	ExactFilters map[string]ExactFilter
	RangeFilters map[string]RangeFilter
	Search       bool
}

type Error struct {
	Kind       string
	ReasonCode string
}

type Result struct {
	Scope  map[string]string
	Values url.Values
}

var searchCaseFolder = cases.Fold()

func Parse(rawQuery string, config Config) (Result, *Error) {
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return Result{}, listError(ReasonInvalidSearch)
	}
	scope := make(map[string]string)
	if config.Search {
		scope["search"] = ""
	}
	for key := range config.ExactFilters {
		scope[key] = ""
	}
	for key := range config.RangeFilters {
		scope[key] = ""
	}

	for _, rawValues := range values {
		if len(rawValues) > 1 {
			return Result{}, listError(ReasonDuplicateQueryMember)
		}
	}
	for key := range values {
		switch {
		case key == "limit" || key == "cursor_token":
			continue
		case isPaginationAlias(key):
			return Result{}, paginationError(pagination.ReasonInvalidLimit)
		case key == "search" && config.Search:
			continue
		case hasExactFilter(config, key):
			continue
		case hasRangeFilter(config, key):
			continue
		default:
			return Result{}, listError(ReasonUnknownQueryMember)
		}
	}

	if config.Search {
		raw, present := singleValue(values, "search")
		if present {
			tokens, apiErr := NormalizeSearch(raw)
			if apiErr != nil {
				return Result{}, apiErr
			}
			scope["search"] = strings.Join(tokens, " ")
		}
	}

	for key, filter := range config.ExactFilters {
		raw, present := singleValue(values, key)
		if !present {
			continue
		}
		value := strings.TrimSpace(raw)
		if !validFilterValue(value) || !allowedExactValue(value, filter.Allowed) {
			return Result{}, listError(ReasonInvalidFilterValue)
		}
		scope[key] = value
	}

	for key := range config.RangeFilters {
		raw, present := singleValue(values, key)
		if !present {
			continue
		}
		value := strings.TrimSpace(raw)
		if !validFilterValue(value) {
			return Result{}, listError(ReasonInvalidFilterRange)
		}
		scope[key] = value
	}

	return Result{Scope: scope, Values: values}, nil
}

func NormalizeSearch(raw string) ([]string, *Error) {
	if !utf8.ValidString(raw) {
		return nil, listError(ReasonInvalidSearch)
	}
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	if normalized == "" {
		return nil, nil
	}
	if countRunes(normalized) > 256 {
		return nil, listError(ReasonInvalidSearch)
	}
	for _, r := range normalized {
		if (r >= 0x00 && r <= 0x1F) || (r >= 0x7F && r <= 0x9F) {
			return nil, listError(ReasonInvalidSearch)
		}
	}

	tokenSet := uniqueTokenSet(tokenizeSearchString(normalized))

	if len(tokenSet) == 0 {
		return nil, listError(ReasonInvalidSearch)
	}
	if len(tokenSet) > 16 {
		return nil, listError(ReasonSearchTokenCountExceed)
	}
	tokens := make([]string, 0, len(tokenSet))
	for token := range tokenSet {
		tokens = append(tokens, token)
	}
	slices.Sort(tokens)
	return tokens, nil
}

func MatchSearchTokens(tokens []string, values ...string) bool {
	if len(tokens) == 0 {
		return true
	}
	sourceTokens := make([]string, 0, len(values))
	for _, value := range values {
		sourceTokens = append(sourceTokens, tokenizeSearchString(value)...)
	}
	for _, token := range tokens {
		matched := false
		for _, sourceToken := range sourceTokens {
			if strings.HasPrefix(sourceToken, token) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
}

func tokenizeSearchString(value string) []string {
	normalized := norm.NFC.String(value)
	tokens := []string{}
	var token strings.Builder
	flush := func() {
		if token.Len() == 0 {
			return
		}
		tokens = append(tokens, searchCaseFolder.String(token.String()))
		token.Reset()
	}
	for _, r := range normalized {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			token.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return tokens
}

func uniqueTokenSet(tokens []string) map[string]struct{} {
	tokenSet := map[string]struct{}{}
	for _, token := range tokens {
		tokenSet[token] = struct{}{}
	}
	return tokenSet
}

func listError(reasonCode string) *Error {
	return &Error{Kind: ErrorKindList, ReasonCode: reasonCode}
}

func paginationError(reasonCode string) *Error {
	return &Error{Kind: ErrorKindPagination, ReasonCode: reasonCode}
}

func singleValue(values url.Values, key string) (string, bool) {
	rawValues, ok := values[key]
	if !ok || len(rawValues) == 0 {
		return "", false
	}
	return rawValues[0], true
}

func hasExactFilter(config Config, key string) bool {
	_, ok := config.ExactFilters[key]
	return ok
}

func hasRangeFilter(config Config, key string) bool {
	_, ok := config.RangeFilters[key]
	return ok
}

func isPaginationAlias(key string) bool {
	switch key {
	case "page", "offset", "page_size", "block_size":
		return true
	default:
		return false
	}
}

func validFilterValue(value string) bool {
	if value == "" || value == "null" {
		return false
	}
	if strings.Contains(value, ",") {
		return false
	}
	if strings.HasPrefix(value, "[") || strings.HasSuffix(value, "]") {
		return false
	}
	return true
}

func allowedExactValue(value string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	return slices.Contains(allowed, value)
}

func countRunes(value string) int {
	count := 0
	for range value {
		count++
	}
	return count
}
