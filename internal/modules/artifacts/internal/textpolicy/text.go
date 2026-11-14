package textpolicy

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Normalize admits the string contracts used by Artifact scalar fields. Empty
// values are returned to the caller, which owns the field's nullability and
// create-only minimum signal. Unknown contract identities fail closed.
func Normalize(raw, contractID string) (string, bool) {
	limit, multiline := 0, false
	switch contractID {
	case "single_line_title_v1":
		limit = 512
	case "party_text_v1":
		limit = 256
	case "multiline_body_v1":
		limit, multiline = 16384, true
	case "reason_note_v1":
		limit, multiline = 4096, true
	case "tag_label_v1":
		limit = 64
	default:
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
	count := 0
	for range normalized {
		count++
		if count > limit {
			return "", false
		}
	}
	return normalized, true
}
