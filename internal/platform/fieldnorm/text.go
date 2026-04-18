package fieldnorm

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func NormalizeLine(raw string) (string, bool) {
	normalized := norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
	if normalized == "" {
		return "", false
	}
	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r) {
			return "", false
		}
	}
	return normalized, true
}

func NormalizeNote(raw string) (string, bool) {
	normalized := norm.NFC.String(raw)
	normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	if normalized == "" {
		return "", false
	}
	for _, r := range normalized {
		switch {
		case r == '\n' || r == '\t':
		case unicode.Is(unicode.Cc, r) || unicode.Is(unicode.Cf, r):
			return "", false
		}
	}
	return normalized, true
}

func NormalizeIdentifier(identifierClass string, raw string) (string, bool) {
	normalized, ok := NormalizeLine(raw)
	if !ok {
		return "", false
	}
	switch identifierClass {
	case "aad_device_id", "fqdn", "hostname", "aad_object_id", "upn", "email", "sam_account_name":
		return strings.ToLower(normalized), true
	case "sid":
		return strings.ToUpper(normalized), true
	default:
		return normalized, true
	}
}
