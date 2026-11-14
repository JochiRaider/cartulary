// Package textpolicy owns normalization for Task and Decision source strings.
package textpolicy

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Normalize returns normalized empty text to the binding owner, which decides
// requiredness. It does not infer lifecycle rules or normalize closed tokens.
func Normalize(raw, contractID string) (string, bool) {
	limit, multiline := 0, false
	switch contractID {
	case "single_line_title_v1":
		limit = 512
	case "party_text_v1":
		limit = 256
	case "locator_text_v1":
		limit = 1024
	case "reason_note_v1":
		limit, multiline = 4096, true
	case "multiline_body_v1":
		limit, multiline = 16384, true
	default:
		return "", false
	}
	if !utf8.ValidString(raw) {
		return "", false
	}
	if multiline {
		raw = strings.ReplaceAll(strings.ReplaceAll(raw, "\r\n", "\n"), "\r", "\n")
	}
	for _, r := range raw {
		if (r <= 31 && !(multiline && (r == '\n' || r == '\t'))) || (r >= 127 && r <= 159) {
			return "", false
		}
	}
	normalized := strings.TrimFunc(norm.NFC.String(raw), unicode.IsSpace)
	return normalized, utf8.RuneCountInString(normalized) <= limit
}
