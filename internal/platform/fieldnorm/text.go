package fieldnorm

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var autoResolutionCaseFolder = cases.Fold()

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

func NormalizeMentionToken(raw string) (string, bool) {
	normalized := norm.NFC.String(raw)
	if normalized == "" {
		return "", false
	}

	var builder strings.Builder
	builder.Grow(len(normalized))

	runeCount := 0
	wroteToken := false
	pendingSpace := false

	for _, r := range normalized {
		if unicode.Is(unicode.Cc, r) {
			return "", false
		}
		if unicode.IsSpace(r) {
			if wroteToken {
				pendingSpace = true
			}
			continue
		}
		if pendingSpace {
			builder.WriteByte(' ')
			runeCount++
			pendingSpace = false
			if runeCount > 256 {
				return "", false
			}
		}
		builder.WriteRune(r)
		runeCount++
		if runeCount > 256 {
			return "", false
		}
		wroteToken = true
	}

	mentionToken := builder.String()
	if mentionToken == "" {
		return "", false
	}
	return mentionToken, true
}

func AutoResolutionCandidateText(raw string) (string, bool) {
	normalized, ok := NormalizeMentionToken(raw)
	if !ok {
		return "", false
	}
	return autoResolutionCaseFolder.String(normalized), true
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
