package harnessredact

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

// The marker, key tokens, CLI flag names, value rules and the retained-secret
// scan's vocabulary are generated from tools/harness_redaction_manifest.json
// into policy_generated.go (TH-HARNESS-REQ-617). This file holds only the
// mechanics that interpret them.

type compiledRule struct {
	pattern     *regexp.Regexp
	replacement string
}

type compiledManifest struct {
	replacement string
	valueRules  []compiledRule
}

// scanTerm is a retained-scan JSON-member term; a non-zero excluded letter
// stops the term from matching when the next character folds to it.
type scanTerm struct {
	text     string
	excluded rune
}

// RetainedScanContext is one syntactic context of the retained-secret scan: a
// prefix, the bytes its value excludes, and the value's terminator.
type RetainedScanContext struct {
	Prefix, Excluded, Terminator string
}

func String(value string) string {
	if value == "" {
		return ""
	}
	for _, rule := range rules.valueRules {
		value = rule.pattern.ReplaceAllString(value, rule.replacement)
	}
	return redactJSONMembers(value)
}

func StructuredString(value string) string {
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return String(value)
	}
	payload, err := JSON(Value(decoded, ""))
	if err != nil {
		return String(value)
	}
	return string(payload)
}

// JSON encodes redacted values as one line without HTML escaping, so Marker
// stays literal and the retained-secret scan can recognize it
// (TH-HARNESS-REQ-617). encoding/json would otherwise write < and >.
func JSON(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Value(value any, key string) any {
	if key != "" && isStructuredSecretKey(key) {
		return rules.replacement
	}
	switch typed := value.(type) {
	case string:
		return structuredString(typed, key)
	case []string:
		items := make([]string, len(typed))
		redactNext := false
		for index, item := range typed {
			if redactNext {
				items[index] = rules.replacement
				redactNext = false
				continue
			}
			items[index] = structuredString(item, "")
			if sensitiveCLIFlagPattern.MatchString(item) {
				redactNext = true
			}
		}
		return items
	case []any:
		items := make([]any, len(typed))
		redactNext := false
		for index, item := range typed {
			if redactNext {
				items[index] = rules.replacement
				redactNext = false
				continue
			}
			items[index] = Value(item, "")
			if itemString, ok := item.(string); ok && sensitiveCLIFlagPattern.MatchString(itemString) {
				redactNext = true
			}
		}
		return items
	case map[string]any:
		items := make(map[string]any, len(typed))
		for entryKey, entryValue := range typed {
			items[entryKey] = Value(entryValue, entryKey)
		}
		return items
	default:
		return value
	}
}

// structuredString redacts a string value. A non-empty string under a key
// that the retained-secret scan's JSON-member rule matches is replaced whole
// (TH-HARNESS-REQ-604), as is the value of a whole-string assignment to a
// secret-named key. Containers and other scalars never reach this closure.
func structuredString(value, key string) string {
	if value == "" {
		return ""
	}
	if key != "" && scanKeyMatches(key) {
		return rules.replacement
	}
	if match := wholeStringAssignmentPattern.FindStringSubmatch(value); match != nil && isStructuredSecretKey(match[1]) {
		return match[1] + "=" + rules.replacement
	}
	return String(value)
}

// canonicalStructuredKey implements TH-HARNESS-REQ-604: a lowercase letter or
// digit followed by an uppercase letter gains a separator, ASCII letters are
// uppercased, and every other run, including non-ASCII bytes, becomes one
// underscore without leading or trailing separators.
func canonicalStructuredKey(value string) string {
	var builder strings.Builder
	separator := false
	var previous byte
	for index := 0; index < len(value); index++ {
		character := value[index]
		lowerOrDigit := (previous >= 'a' && previous <= 'z') || (previous >= '0' && previous <= '9')
		previous = character
		switch {
		case character >= 'a' && character <= 'z':
			character -= 'a' - 'A'
		case character >= 'A' && character <= 'Z':
			separator = separator || lowerOrDigit
		case character >= '0' && character <= '9':
		default:
			separator = true
			continue
		}
		if separator && builder.Len() > 0 {
			builder.WriteByte('_')
		}
		separator = false
		builder.WriteByte(character)
	}
	return builder.String()
}

func isStructuredSecretKey(key string) bool {
	canonical := canonicalStructuredKey(key)
	if canonical == "" {
		return false
	}
	for _, token := range structuredSecretKeyTokens {
		if canonical == token || strings.HasPrefix(canonical, token+"_") || strings.HasSuffix(canonical, "_"+token) {
			return true
		}
	}
	return false
}

// foldScanRune applies simple case folding as the scan's case-insensitive
// matching does to its ASCII terms: ASCII letters, the long s and the Kelvin
// sign.
func foldScanRune(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return r + ('a' - 'A')
	case r == '\u017f':
		return 's'
	case r == '\u212a':
		return 'k'
	}
	return r
}

// scanKeyMatches reports whether a whole key contains a retained-scan
// JSON-member term. Membership is decided in code, in linear time
// (TH-HARNESS-REQ-604, TH-HARNESS-REQ-617).
func scanKeyMatches(key string) bool {
	folded := make([]rune, 0, len(key))
	for _, r := range key {
		folded = append(folded, foldScanRune(r))
	}
	for _, term := range scanMemberTerms {
	search:
		for index := 0; index+len(term.text) <= len(folded); index++ {
			for offset := 0; offset < len(term.text); offset++ {
				if folded[index+offset] != rune(term.text[offset]) {
					continue search
				}
			}
			end := index + len(term.text)
			if term.excluded == 0 || end == len(folded) || folded[end] != term.excluded {
				return true
			}
		}
	}
	return false
}

func isASCIISpace(character byte) bool {
	return character == '\t' || character == '\n' || character == '\f' || character == '\r' || character == ' '
}

// redactJSONMembers is the scan closure for JSON members embedded in raw text
// (TH-HARNESS-REQ-604): a quoted key, ASCII whitespace, a colon, ASCII
// whitespace and a non-empty quoted value has its value replaced when the key
// contains a scan term. Every quote is a candidate key start, as in the scan,
// and each candidate inspects only its own key, separator and value, so the
// pass is linear.
func redactJSONMembers(text string) string {
	quotes := []int{}
	for index := 0; index < len(text); index++ {
		if text[index] == '"' {
			quotes = append(quotes, index)
		}
	}
	var builder strings.Builder
	previous := 0
	for quote := 0; quote+3 < len(quotes); quote++ {
		cursor := quotes[quote+1] + 1
		for cursor < len(text) && isASCIISpace(text[cursor]) {
			cursor++
		}
		if cursor >= len(text) || text[cursor] != ':' {
			continue
		}
		cursor++
		for cursor < len(text) && isASCIISpace(text[cursor]) {
			cursor++
		}
		if cursor != quotes[quote+2] || quotes[quote+3] == cursor+1 || !scanKeyMatches(text[quotes[quote]+1:quotes[quote+1]]) {
			continue
		}
		builder.WriteString(text[previous : cursor+1])
		builder.WriteString(rules.replacement)
		previous = quotes[quote+3]
	}
	if previous == 0 {
		return text
	}
	builder.WriteString(text[previous:])
	return builder.String()
}
