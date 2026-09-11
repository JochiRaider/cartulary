package fieldnorm

import (
	"encoding/json"
	"slices"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	contractentities "github.com/JochiRaider/cartulary/internal/gen/contractentities"
	"golang.org/x/text/unicode/norm"
)

// Entity comparison uses the database's Unicode 16 normalization identity.
// Keep this separate from request decoding, which owns retained request hashes.
type entityUnicodeContract struct {
	UnicodeVersion   string            `json:"unicode_version"`
	AssignedRanges   [][2]rune         `json:"assigned_ranges"`
	CombiningClasses map[rune]uint8    `json:"combining_classes"`
	Lowercase        entityCaseMapping `json:"lowercase"`
	SIDUppercase     entityCaseMapping `json:"sid_uppercase"`
	RejectedRanges   [][2]rune         `json:"identifier_rejected_ranges"`
	WhitespaceRanges [][2]rune         `json:"whitespace_ranges"`
}

type entityCaseMapping struct {
	Source string `json:"source"`
	Target string `json:"target"`
	values map[rune]rune
}

var entityUnicode = sync.OnceValue(loadEntityUnicode)

func loadEntityUnicode() entityUnicodeContract {
	var contract entityUnicodeContract
	artifact := contractentities.Index["contracts/entities/identifier-unicode.v1.json"]
	if err := json.Unmarshal([]byte(artifact.JSON), &contract); err != nil || contract.UnicodeVersion != "16.0.0" {
		panic("invalid Entities Unicode contract")
	}
	for _, mapping := range []*entityCaseMapping{&contract.Lowercase, &contract.SIDUppercase} {
		sources, targets := []rune(mapping.Source), []rune(mapping.Target)
		if len(sources) != len(targets) || len(sources) == 0 {
			panic("invalid Entities casing projection")
		}
		mapping.values = make(map[rune]rune, len(sources))
		for index, source := range sources {
			mapping.values[source] = targets[index]
		}
	}
	return contract
}

func (contract entityUnicodeContract) assigned(value rune) bool {
	return entityRangeContains(contract.AssignedRanges, value)
}

func entityRangeContains(ranges [][2]rune, value rune) bool {
	index := sort.Search(len(ranges), func(index int) bool {
		return ranges[index][1] >= value
	})
	return index < len(ranges) && ranges[index][0] <= value
}

func entityWhitespace(value rune) bool {
	return entityRangeContains(entityUnicode().WhitespaceRanges, value)
}

// NormalizeEntityAliasText retains alias_text_v1 case, C0/C1 policy and length,
// with the same NFC identity as the Entities source constraint.
func NormalizeEntityAliasText(raw string) (string, bool) {
	if !utf8.ValidString(raw) {
		return "", false
	}
	value := entityNFC(strings.TrimFunc(raw, entityWhitespace))
	if value == "" || utf8.RuneCountInString(value) > 256 {
		return "", false
	}
	for _, scalar := range value {
		if scalar <= 31 || scalar >= 127 && scalar <= 159 {
			return "", false
		}
	}
	return value, true
}

func normalizeEntityIdentifier(raw string) (string, bool) {
	if !utf8.ValidString(raw) {
		return "", false
	}
	value := entityNFC(strings.TrimFunc(raw, entityWhitespace))
	if value == "" {
		return "", false
	}
	for _, scalar := range value {
		if entityRangeContains(entityUnicode().RejectedRanges, scalar) {
			return "", false
		}
	}
	return value, true
}

func (mapping entityCaseMapping) apply(value string) string {
	return strings.Map(func(scalar rune) rune {
		if mapped, found := mapping.values[scalar]; found {
			return mapped
		}
		return scalar
	}, value)
}

// Normalize assigned scalars, order combining sequences, compose unblocked pairs.
// Pair normalization cannot insert x/text's stream-safe CGJ into long sequences.
// Unassigned Unicode 16 scalars stay inert starters in newer runtimes. Unicode
// normalization stability preserves decomposition/composition of the assigned
// repertoire; the contract owns assignment and canonical combining classes.
func entityNFC(raw string) string {
	if strings.IndexFunc(raw, func(value rune) bool { return value >= utf8.RuneSelf }) == -1 {
		return raw
	}
	decomposed := make([]rune, 0, utf8.RuneCountInString(raw))
	for _, scalar := range raw {
		if entityUnicode().assigned(scalar) {
			decomposed = append(decomposed, []rune(norm.NFD.String(string(scalar)))...)
		} else {
			decomposed = append(decomposed, scalar)
		}
	}
	for start := 0; start < len(decomposed); {
		if entityUnicode().CombiningClasses[decomposed[start]] == 0 {
			start++
			continue
		}
		end := start + 1
		for end < len(decomposed) && entityUnicode().CombiningClasses[decomposed[end]] != 0 {
			end++
		}
		slices.SortStableFunc(decomposed[start:end], func(left, right rune) int {
			return int(entityUnicode().CombiningClasses[left]) - int(entityUnicode().CombiningClasses[right])
		})
		start = end
	}
	composed := make([]rune, 0, len(decomposed))
	starter, previousClass := -1, uint8(0)
	for _, scalar := range decomposed {
		class := entityUnicode().CombiningClasses[scalar]
		if starter >= 0 && (previousClass == 0 || previousClass < class) &&
			entityUnicode().assigned(scalar) && entityUnicode().assigned(composed[starter]) {
			pair := []rune(norm.NFC.String(string([]rune{composed[starter], scalar})))
			if len(pair) == 1 {
				composed[starter] = pair[0]
				continue
			}
		}
		composed = append(composed, scalar)
		if class == 0 {
			starter = len(composed) - 1
		}
		previousClass = class
	}
	return string(composed)
}
