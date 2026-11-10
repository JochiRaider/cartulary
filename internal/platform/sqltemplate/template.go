// Package sqltemplate owns lexical admission for closed SQL construction.
// It does not classify relations, authorize operations, or parse SQL grammar.
package sqltemplate

import (
	"fmt"
	"strings"
)

// Parse returns fixed chunks around whole-token formatting holes. Literal mode
// has no holes. Every chunk ends in ordinary SQL code, so composition cannot
// change a quoted token's meaning. Parentheses may span independent fragments.
func Parse(source string, format bool) ([]string, error) {
	if strings.ContainsRune(source, 0) {
		return nil, fmt.Errorf("SQL template: NUL is forbidden")
	}
	chunks := []string{}
	start := 0
	fail := func(at int, reason string) ([]string, error) {
		return nil, fmt.Errorf("SQL template byte %d: %s", at+1, reason)
	}
	for i := 0; i < len(source); {
		if source[i] == 0 {
			return fail(i, "NUL is forbidden")
		}
		if strings.HasPrefix(source[i:], "--") || strings.HasPrefix(source[i:], "/*") || strings.HasPrefix(source[i:], "*/") {
			return fail(i, "comments are forbidden in closed templates")
		}
		if format && strings.HasPrefix(source[i:], "%s") {
			if i > 0 && word(source[i-1]) || i+2 < len(source) && word(source[i+2]) {
				return fail(i, "placeholder must occur between complete tokens")
			}
			chunks = append(chunks, source[start:i])
			i += 2
			start = i
			continue
		}
		// Dollar strings must close using the identical tag. Positional parameters
		// and dollar characters in identifiers are not dollar strings.
		if source[i] == '$' && (i == 0 || !word(source[i-1])) {
			j := i + 1
			if j < len(source) && (letter(source[j]) || source[j] == '_') {
				for j < len(source) && (letter(source[j]) || source[j] == '_' || digit(source[j])) {
					j++
				}
			}
			if j < len(source) && source[j] == '$' {
				tag := source[i : j+1]
				end := strings.Index(source[j+1:], tag)
				if end < 0 {
					return fail(i, "unterminated dollar string")
				}
				end += j + 1 + len(tag)
				if format && strings.Contains(source[i:end], "%s") {
					return fail(i, "placeholder inside quoted token")
				}
				i = end
				continue
			}
		}
		if source[i] == '\'' || source[i] == '"' {
			quote := source[i]
			begin := i
			escape := false
			if i > 0 && (source[i-1] == 'e' || source[i-1] == 'E') && (i == 1 || !word(source[i-2])) && quote == '\'' {
				escape = true
			}
			if i >= 2 && source[i-1] == '&' && (source[i-2] == 'U' || source[i-2] == 'u') {
				return fail(i, "Unicode escape tokens are not admitted")
			}
			i++
			closed := false
			for i < len(source) {
				if source[i] == 0 {
					return fail(i, "NUL is forbidden")
				}
				if source[i] == quote {
					if i+1 < len(source) && source[i+1] == quote {
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				if escape && source[i] == '\\' {
					i++
					if i == len(source) {
						break
					}
				}
				i++
			}
			if !closed {
				return fail(begin, "unterminated quoted token")
			}
			if format && strings.Contains(source[begin:i], "%s") {
				return fail(begin, "placeholder inside quoted token")
			}
			continue
		}
		i++
	}
	chunks = append(chunks, source[start:])
	return chunks, nil
}

func letter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 128 }
func digit(c byte) bool  { return c >= '0' && c <= '9' }
func word(c byte) bool   { return letter(c) || digit(c) || c == '_' || c == '$' }
