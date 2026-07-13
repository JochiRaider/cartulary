package networkflow

import (
	"strconv"
	"strings"
	"unicode/utf8"

	norm "github.com/JochiRaider/cartulary/internal/gen/networkflowunicode"
)

func SanitizeSourceFilenameDisplay(filenameHint string) string {
	value := norm.NFC.String(filenameHint)
	value = strings.ReplaceAll(value, "\\", "/")
	segments := strings.Split(value, "/")
	candidate := "uploaded.csv"
	for i := len(segments) - 1; i >= 0; i-- {
		if utf8.RuneCountInString(segments[i]) > 0 {
			candidate = segments[i]
			break
		}
	}
	candidate = removeC0C1Controls(candidate)
	candidate = trimUnicodeWhitespace(candidate)
	if candidate == "" {
		candidate = "uploaded.csv"
	}
	return firstRunes(candidate, 256)
}

func NormalizeTableDisplayNameInput(value string) (string, error) {
	normalized := norm.NFC.String(value)
	trimmed := trimUnicodeWhitespace(normalized)
	if trimmed == "" {
		return "", nil
	}
	if containsC0C1Control(normalized) {
		return "", &InvalidDisplayNameError{ReasonCode: "forbidden_control"}
	}
	return trimmed, nil
}

func DeriveTableDisplayName(originalFilename string, existingActiveDisplayNames map[string]struct{}) (string, error) {
	sourceDisplay := SanitizeSourceFilenameDisplay(originalFilename)
	stem := filenameStemAfterPathStripping(sourceDisplay)
	candidate, err := NormalizeTableDisplayNameInput(stem)
	if err != nil {
		return "", err
	}
	if candidate == "" {
		candidate = "Imported NetFlow"
	}
	candidate = firstRunes(candidate, 64)
	if _, exists := existingActiveDisplayNames[candidate]; !exists {
		return candidate, nil
	}
	for n := 2; n <= 9999; n++ {
		suffix := " (" + strconv.Itoa(n) + ")"
		base := firstRunes(candidate, 64-utf8.RuneCountInString(suffix))
		suffixed := base + suffix
		if _, exists := existingActiveDisplayNames[suffixed]; !exists {
			return suffixed, nil
		}
	}
	return "", ErrTableNameExhausted
}

func normalizeExplicitDisplayName(value string) (string, error) {
	normalized, err := NormalizeTableDisplayNameInput(value)
	if err != nil {
		return "", err
	}
	switch {
	case normalized == "":
		return "", &InvalidDisplayNameError{ReasonCode: "empty_display_name"}
	case utf8.RuneCountInString(normalized) > 64:
		return "", &InvalidDisplayNameError{ReasonCode: "display_name_too_long"}
	default:
		return normalized, nil
	}
}

func filenameStemAfterPathStripping(sourceFilenameDisplay string) string {
	runes := []rune(sourceFilenameDisplay)
	if len(runes) > 1 && runes[len(runes)-1] == '.' {
		return string(runes[:len(runes)-1])
	}
	lastDot := -1
	for i, r := range runes {
		if r == '.' {
			lastDot = i
		}
	}
	if lastDot > 0 && lastDot < len(runes)-1 {
		return string(runes[:lastDot])
	}
	return sourceFilenameDisplay
}

func removeC0C1Controls(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if isC0C1Control(r) {
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func containsC0C1Control(value string) bool {
	for _, r := range value {
		if isC0C1Control(r) {
			return true
		}
	}
	return false
}

func isC0C1Control(r rune) bool {
	return (r >= 0x00 && r <= 0x1f) || (r >= 0x7f && r <= 0x9f)
}

func trimUnicodeWhitespace(value string) string {
	return strings.TrimFunc(value, isUnicodeWhitespaceV1)
}

// isUnicodeWhitespaceV1 is the closed Unicode White_Space scalar set owned by
// trim_unicode_whitespace_v1. Keeping the set explicit prevents a Go runtime
// or Unicode-table upgrade from changing Network Flow names and fingerprints.
func isUnicodeWhitespaceV1(r rune) bool {
	switch r {
	case 0x0009, 0x000a, 0x000b, 0x000c, 0x000d,
		0x0020, 0x0085, 0x00a0, 0x1680,
		0x2028, 0x2029, 0x202f, 0x205f, 0x3000:
		return true
	default:
		return r >= 0x2000 && r <= 0x200a
	}
}

func firstRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit])
}
