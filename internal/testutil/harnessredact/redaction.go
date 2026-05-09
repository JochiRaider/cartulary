package harnessredact

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
)

const fallbackReplacement = "[REDACTED]"

type manifest struct {
	Replacement   string        `json:"replacement"`
	ValuePatterns []patternRule `json:"value_patterns"`
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

var (
	rulesOnce sync.Once
	rules     []compiledRule
)

func String(value string) string {
	if value == "" {
		return ""
	}
	for _, rule := range compiledRules() {
		value = rule.pattern.ReplaceAllString(value, rule.replacement)
	}
	return value
}

func compiledRules() []compiledRule {
	rulesOnce.Do(func() {
		rules = loadCompiledRules()
	})
	return rules
}

func loadCompiledRules() []compiledRule {
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
	compiled := make([]compiledRule, 0, len(parsed.ValuePatterns))
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
		compiled = append(compiled, compiledRule{
			pattern:     pattern,
			replacement: ruleReplacement,
		})
	}
	if len(compiled) == 0 {
		return fallbackRules()
	}
	return compiled
}

func fallbackRules() []compiledRule {
	return []compiledRule{
		{pattern: regexp.MustCompile(`(?i)\bBearer\s+[-._~+/A-Za-z0-9]+=*`), replacement: fallbackReplacement},
		{pattern: regexp.MustCompile(`(?i)\bBasic\s+[-._~+/A-Za-z0-9]+=*`), replacement: fallbackReplacement},
		{pattern: regexp.MustCompile(`(?i)postgres(?:ql)?://([^:\s/@]+):([^@\s]+)@`), replacement: `postgres://$1:[REDACTED]@`},
		{pattern: regexp.MustCompile(`(?i)\b(password|passwd|secret|token|api[_-]?key|access[_-]?key|client[_-]?secret)=([^\s&]+)`), replacement: `$1=[REDACTED]`},
		{pattern: regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`), replacement: fallbackReplacement},
	}
}

func manifestPath() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join("tools", "harness_redaction_manifest.json")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "tools", "harness_redaction_manifest.json")
}
