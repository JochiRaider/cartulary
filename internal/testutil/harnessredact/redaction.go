package harnessredact

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const fallbackReplacement = "[REDACTED]"

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

type manifest struct {
	Replacement          string        `json:"replacement"`
	SensitiveKeyPatterns []string      `json:"sensitive_key_patterns"`
	ValuePatterns        []patternRule `json:"value_patterns"`
}

type patternRule struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
}

type compiledRule struct {
	pattern     *regexp.Regexp
	replacement string
}

type compiledManifest struct {
	replacement string
	keyPatterns []*regexp.Regexp
	valueRules  []compiledRule
}

var (
	rulesOnce sync.Once
	rules     compiledManifest
)

func String(value string) string {
	if value == "" {
		return ""
	}
	for _, rule := range compiledRules().valueRules {
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
	compiled := compiledRules()
	if key != "" && isStructuredSecretKey(key) {
		return compiled.replacement
	}
	switch typed := value.(type) {
	case string:
		return String(typed)
	case []string:
		items := make([]string, len(typed))
		redactNext := false
		for index, item := range typed {
			if redactNext {
				items[index] = compiled.replacement
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
				items[index] = compiled.replacement
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

func compiledRules() compiledManifest {
	rulesOnce.Do(func() {
		rules = loadCompiledRules()
	})
	return rules
}

func loadCompiledRules() compiledManifest {
	raw, err := os.ReadFile(manifestPath())
	if err != nil {
		return fallbackRules()
	}
	var parsed manifest
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fallbackRules()
	}
	replacement := parsed.Replacement
	if replacement == "" {
		replacement = fallbackReplacement
	}
	compiled := compiledManifest{replacement: replacement}
	for _, pattern := range parsed.SensitiveKeyPatterns {
		if pattern == "" {
			continue
		}
		regex, err := regexp.Compile("(?i)" + pattern)
		if err != nil {
			continue
		}
		compiled.keyPatterns = append(compiled.keyPatterns, regex)
	}
	for _, rule := range parsed.ValuePatterns {
		if rule.Pattern == "" {
			continue
		}
		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue
		}
		ruleReplacement := rule.Replacement
		if ruleReplacement == "" {
			ruleReplacement = replacement
		}
		compiled.valueRules = append(compiled.valueRules, compiledRule{
			pattern:     pattern,
			replacement: ruleReplacement,
		})
	}
	if len(compiled.valueRules) == 0 {
		return fallbackRules()
	}
	return compiled
}

func fallbackRules() compiledManifest {
	return compiledManifest{
		replacement: fallbackReplacement,
		keyPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^(?:authorization|cookie|set-cookie|x-cartulary-test-route-token)$`),
			regexp.MustCompile(`(?i)^(?:password|passwd|pwd|secret|token|jwt|bearer|dsn)$`),
			regexp.MustCompile(`(?i)(?:^|[_-])(?:password|passwd|pwd|secret|token|jwt|dsn)$`),
			regexp.MustCompile(`(?i)(?:^|[_-])api[_-]?key$`),
			regexp.MustCompile(`(?i)(?:^|[_-])access[_-]?key$`),
			regexp.MustCompile(`(?i)(?:^|[_-])secret[_-]?key$`),
			regexp.MustCompile(`(?i)(?:^|[_-])private[_-]?key$`),
			regexp.MustCompile(`(?i)(?:^|[_-])client[_-]?secret$`),
		},
		valueRules: []compiledRule{
			{pattern: regexp.MustCompile(`(?i)\bBearer\s+[-._~+/A-Za-z0-9]+=*`), replacement: fallbackReplacement},
			{pattern: regexp.MustCompile(`(?i)\b(Authorization\s*:\s*)Bearer\s+[-._~+/A-Za-z0-9]+=*`), replacement: `$1` + fallbackReplacement},
			{pattern: regexp.MustCompile(`(?i)\b((?:Cookie|Set-Cookie|X-Cartulary-Test-Route-Token)\s*:\s*)([^\r\n]+)`), replacement: `$1` + fallbackReplacement},
			{pattern: regexp.MustCompile(`(?i)postgres(?:ql)?://([^:\s/@]+):([^@\s]+)@`), replacement: `postgres://$1:[REDACTED]@`},
			{pattern: regexp.MustCompile(`(?i)\b(password|passwd|secret|token|api[_-]?key|access[_-]?key|secret[_-]?key|client[_-]?secret)\s*=\s*([^\s&]+)`), replacement: `$1=[REDACTED]`},
			{pattern: regexp.MustCompile(`(?i)\b((?:minio|s3|aws)?[_-]?(?:access[_-]?key|secret[_-]?key|access[_-]?key[_-]?id|secret[_-]?access[_-]?key)\s*=\s*)([^\s&]+)`), replacement: `$1[REDACTED]`},
			{pattern: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`), replacement: fallbackReplacement},
		},
	}
}

func manifestPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("tools", "harness_redaction_manifest.json")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "tools", "harness_redaction_manifest.json")
}
