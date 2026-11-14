package harnessredact

import (
	"encoding/json"
	"regexp"
	"strings"
)

var sensitiveCLIFlagPattern = regexp.MustCompile(`(?i)^--(?:password|passwd|pwd|secret|token|jwt|api[_-]?key|access[_-]?key|secret[_-]?key|private[_-]?key|client[_-]?secret|dsn)$`)
var structuredSecretKeyTokens = map[string]struct{}{
	"PASSWORD":                     {},
	"PASS":                         {},
	"PWD":                          {},
	"TOKEN":                        {},
	"JWT":                          {},
	"BEARER":                       {},
	"API_KEY":                      {},
	"ACCESS_KEY":                   {},
	"SECRET_KEY":                   {},
	"AUTHORIZATION":                {},
	"COOKIE":                       {},
	"SET_COOKIE":                   {},
	"X_CARTULARY_TEST_ROUTE_TOKEN": {},
}

type compiledRule struct {
	pattern     *regexp.Regexp
	replacement string
}

type compiledManifest struct {
	replacement string
	valueRules  []compiledRule
}

func String(value string) string {
	if value == "" {
		return ""
	}
	for _, rule := range rules.valueRules {
		value = rule.pattern.ReplaceAllString(value, rule.replacement)
	}
	return value
}

func StructuredString(value string) string {
	var decoded any
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return String(value)
	}
	payload, err := json.Marshal(Value(decoded, ""))
	if err != nil {
		return String(value)
	}
	return string(payload) + "\n"
}

func Value(value any, key string) any {
	if key != "" && isStructuredSecretKey(key) {
		return rules.replacement
	}
	switch typed := value.(type) {
	case string:
		return String(typed)
	case []string:
		items := make([]string, len(typed))
		redactNext := false
		for index, item := range typed {
			if redactNext {
				items[index] = rules.replacement
				redactNext = false
				continue
			}
			items[index] = String(item)
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

func canonicalStructuredKey(value string) string {
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range strings.ToUpper(value) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			builder.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(builder.String(), "_")
}

func isStructuredSecretKey(key string) bool {
	canonical := canonicalStructuredKey(key)
	if canonical == "" {
		return false
	}
	if _, ok := structuredSecretKeyTokens[canonical]; ok {
		return true
	}
	for token := range structuredSecretKeyTokens {
		if strings.HasPrefix(canonical, token+"_") || strings.HasSuffix(canonical, "_"+token) {
			return true
		}
	}
	return false
}
