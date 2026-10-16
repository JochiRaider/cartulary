package reporting

import (
	"strings"
	"unicode/utf8"
)

// Identifiers preserve exact Unicode code points. Never trim, case-fold or
// encode a rejected identity into an apparently acceptable substitute.
func validReportingIdentifier(value string) bool {
	if !utf8.ValidString(value) || value == "" || utf8.RuneCountInString(value) > 128 || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r >= 0x7f && r <= 0x9f || r == '/' || r == '\\' || r == '#' {
			return false
		}
	}
	return true
}
