// Package textpolicy owns Evidence's bound human-authored source strings.
package textpolicy

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Normalize leaves normalized-empty handling to the field binding owner.
func Normalize(raw, contractID string) (string, bool) {
	limit := 0
	switch contractID {
	case "single_line_title_v1":
		limit = 512
	case "locator_text_v1":
		limit = 1024
	case "party_text_v1":
		limit = 256
	default:
		return "", false
	}
	if !utf8.ValidString(raw) {
		return "", false
	}
	for _, r := range raw {
		if r <= 31 || (r >= 127 && r <= 159) {
			return "", false
		}
	}
	value := strings.TrimFunc(norm.NFC.String(raw), unicode.IsSpace)
	return value, utf8.RuneCountInString(value) <= limit
}
