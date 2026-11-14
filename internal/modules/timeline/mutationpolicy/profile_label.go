package mutationpolicy

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// NormalizeProfileLabel applies the profile's single_line_title_v1 binding.
// Null handling belongs to the route or retained-source shape admission.
func NormalizeProfileLabel(raw string) (string, bool) {
	if !utf8.ValidString(raw) {
		return "", false
	}
	for _, r := range raw {
		if r <= 31 || (r >= 127 && r <= 159) {
			return "", false
		}
	}
	label := strings.TrimFunc(norm.NFC.String(raw), unicode.IsSpace)
	return label, utf8.RuneCountInString(label) <= 512
}
