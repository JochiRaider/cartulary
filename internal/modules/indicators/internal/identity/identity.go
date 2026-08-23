package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/netip"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

var hashPattern = regexp.MustCompile(`^[A-Fa-f0-9]{64}$`)

type Input struct {
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
}

type Canonical struct {
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DedupeKey       string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
}

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid indicator identity: %s %s", e.Field, e.ReasonCode)
}

func Canonicalize(input Input) (Canonical, error) {
	for _, required := range []struct {
		field string
		value string
	}{
		{field: "indicator_type", value: input.IndicatorType},
		{field: "value_kind", value: input.ValueKind},
		{field: "display_value", value: input.DisplayValue},
	} {
		if strings.TrimSpace(required.value) == "" {
			return Canonical{}, invalid(required.field, "missing_required_field")
		}
	}

	indicatorType, err := NormalizeIndicatorType(input.IndicatorType)
	if err != nil {
		return Canonical{}, invalid("indicator_type", "invalid_value")
	}
	valueKind, err := NormalizeValueKind(input.ValueKind)
	if err != nil {
		return Canonical{}, invalid("value_kind", "invalid_value")
	}
	if IsIPType(indicatorType) && valueKind != "atomic" {
		return Canonical{}, invalid("value_kind", "invalid_value")
	}
	displayValue, normalizedValue, err := NormalizeValue(indicatorType, input.DisplayValue, input.NormalizedValue)
	if err != nil {
		field := "display_value"
		if strings.Contains(err.Error(), "normalized_value") {
			field = "normalized_value"
		}
		return Canonical{}, invalid(field, "invalid_value")
	}
	hashAlgorithm, hashValue, err := normalizeHashPair(input.HashAlgorithm, input.HashValue)
	if err != nil {
		field := "hash_algorithm"
		if strings.Contains(err.Error(), "hash_value") {
			field = "hash_value"
		}
		return Canonical{}, invalid(field, "invalid_value")
	}
	if IsIPType(indicatorType) && (hashAlgorithm != nil || hashValue != nil) {
		return Canonical{}, invalid("hash_algorithm", "invalid_value")
	}
	canonical := Canonical{
		IndicatorType:   indicatorType,
		ValueKind:       valueKind,
		DisplayValue:    displayValue,
		NormalizedValue: cloneString(normalizedValue),
		DefangedValue:   cloneString(input.DefangedValue),
		HashAlgorithm:   cloneString(hashAlgorithm),
		HashValue:       cloneString(hashValue),
		STIXPattern:     cloneString(input.STIXPattern),
	}
	canonical.DedupeKey = DedupeKey(canonical)
	return canonical, nil
}

func NormalizeIndicatorType(raw string) (string, error) {
	if normalized, ok := vocabulary.CanonicalIndicatorType(raw); ok {
		return normalized, nil
	}
	return "", fmt.Errorf("unsupported indicator_type")
}

func NormalizeValueKind(raw string) (string, error) {
	if normalized, ok := vocabulary.CanonicalValueKind(raw); ok {
		return normalized, nil
	}
	return "", fmt.Errorf("unsupported value_kind")
}

func NormalizeValue(indicatorType string, rawDisplay string, rawNormalized *string) (string, *string, error) {
	displayValue := strings.TrimSpace(rawDisplay)
	switch indicatorType {
	case "ipv4_addr", "ipv6_addr":
		value, err := canonicalizeIPValue(indicatorType, displayValue)
		if err != nil {
			return "", nil, fmt.Errorf("invalid %s display_value", indicatorType)
		}
		if rawNormalized != nil && strings.TrimSpace(*rawNormalized) != "" {
			normalizedValue, err := canonicalizeIPValue(indicatorType, *rawNormalized)
			if err != nil || normalizedValue != value {
				return "", nil, fmt.Errorf("invalid normalized_value")
			}
		}
		return value, stringPointer(value), nil
	case "domain_name":
		value := strings.ToLower(strings.ReplaceAll(displayValue, "[.]", "."))
		return value, stringPointer(value), nil
	case "url":
		value := canonicalizeURL(displayValue)
		return value, stringPointer(value), nil
	case "email_addr":
		value := strings.ToLower(displayValue)
		return value, stringPointer(value), nil
	case "sha256":
		value := strings.ToLower(displayValue)
		if !hashPattern.MatchString(value) {
			return "", nil, fmt.Errorf("invalid sha256 display_value")
		}
		return value, stringPointer(value), nil
	default:
		if rawNormalized == nil || strings.TrimSpace(*rawNormalized) == "" {
			return displayValue, stringPointer(displayValue), nil
		}
		normalized, ok := fieldnorm.NormalizeLine(*rawNormalized)
		if !ok {
			return "", nil, fmt.Errorf("invalid normalized_value")
		}
		return displayValue, &normalized, nil
	}
}

func NormalizeObservationCandidate(parsedType *string, normalizedCandidate *string, observedText string) (*string, *string, error) {
	if parsedType != nil && strings.TrimSpace(*parsedType) != "" {
		indicatorType, err := NormalizeIndicatorType(*parsedType)
		if err != nil {
			return nil, nil, err
		}
		normalizedText := observedText
		if normalizedCandidate != nil && strings.TrimSpace(*normalizedCandidate) != "" {
			normalizedText = strings.TrimSpace(*normalizedCandidate)
		}
		switch indicatorType {
		case "ipv4_addr", "ipv6_addr":
			value, err := canonicalizeIPValue(indicatorType, normalizedText)
			if err != nil {
				return nil, nil, err
			}
			normalizedText = value
		case "domain_name":
			normalizedText = strings.ToLower(strings.ReplaceAll(normalizedText, "[.]", "."))
		case "url":
			normalizedText = canonicalizeURL(normalizedText)
		case "sha256":
			normalizedText = strings.ToLower(normalizedText)
		}
		return stringPointer(indicatorType), stringPointer(normalizedText), nil
	}

	guessType, candidate := guessObservationCandidate(observedText)
	if guessType == "" || candidate == "" {
		return nil, nil, nil
	}
	return stringPointer(guessType), stringPointer(candidate), nil
}

func IsIPType(indicatorType string) bool {
	return indicatorType == "ipv4_addr" || indicatorType == "ipv6_addr"
}

func DedupeKey(input Canonical) string {
	parts := []string{
		input.IndicatorType,
		input.ValueKind,
		input.DisplayValue,
		derefString(input.NormalizedValue),
		derefString(input.HashAlgorithm),
		derefString(input.HashValue),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))
	return hex.EncodeToString(sum[:])
}

func invalid(field string, reasonCode string) *ValidationError {
	return &ValidationError{Field: field, ReasonCode: reasonCode}
}

func normalizeHashPair(rawAlgorithm *string, rawValue *string) (*string, *string, error) {
	switch {
	case rawAlgorithm == nil && rawValue == nil:
		return nil, nil, nil
	case rawAlgorithm == nil:
		return nil, nil, fmt.Errorf("missing hash_algorithm for hash_value")
	case rawValue == nil:
		return nil, nil, fmt.Errorf("missing hash_value for hash_algorithm")
	}
	algorithm := strings.ToLower(strings.TrimSpace(*rawAlgorithm))
	value := strings.ToLower(strings.TrimSpace(*rawValue))
	if algorithm == "" || value == "" {
		return nil, nil, fmt.Errorf("empty hash pair")
	}
	if !isHexString(value) {
		return nil, nil, fmt.Errorf("invalid hash_value")
	}
	return &algorithm, &value, nil
}

func canonicalizeIPValue(indicatorType string, raw string) (string, error) {
	switch indicatorType {
	case "ipv4_addr":
		return canonicalizeIPv4(raw)
	case "ipv6_addr":
		return canonicalizeIPv6(raw)
	default:
		return "", fmt.Errorf("unsupported ip indicator_type")
	}
}

func canonicalizeIPv4(raw string) (string, error) {
	candidate := strings.ReplaceAll(strings.TrimSpace(raw), "[.]", ".")
	parts := strings.Split(candidate, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid ipv4 literal")
	}
	var octets [4]byte
	for index, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return "", fmt.Errorf("invalid ipv4 literal")
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return "", fmt.Errorf("invalid ipv4 literal")
			}
		}
		value, err := strconv.Atoi(part)
		if err != nil || value > 255 {
			return "", fmt.Errorf("invalid ipv4 literal")
		}
		octets[index] = byte(value)
	}
	return netip.AddrFrom4(octets).String(), nil
}

func canonicalizeIPv6(raw string) (string, error) {
	candidate := strings.TrimSpace(raw)
	if strings.Contains(candidate, "%") || strings.Contains(candidate, ".") {
		return "", fmt.Errorf("invalid ipv6 literal")
	}
	addr, err := netip.ParseAddr(candidate)
	if err != nil || !addr.Is6() || addr.Is4In6() {
		return "", fmt.Errorf("invalid ipv6 literal")
	}
	return addr.String(), nil
}

func canonicalizeURL(raw string) string {
	candidate := strings.TrimSpace(raw)
	candidate = strings.ReplaceAll(candidate, "hxxp://", "http://")
	candidate = strings.ReplaceAll(candidate, "hxxps://", "https://")
	candidate = strings.ReplaceAll(candidate, "[.]", ".")
	parsed, err := neturl.Parse(candidate)
	if err != nil {
		return strings.ToLower(candidate)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	return parsed.String()
}

func guessObservationCandidate(observedText string) (string, string) {
	defanged := strings.ReplaceAll(observedText, "[.]", ".")
	if value, err := canonicalizeIPValue("ipv4_addr", defanged); err == nil {
		return "ipv4_addr", value
	}
	if value, err := canonicalizeIPValue("ipv6_addr", defanged); err == nil {
		return "ipv6_addr", value
	}
	if strings.HasPrefix(strings.ToLower(defanged), "http://") || strings.HasPrefix(strings.ToLower(defanged), "https://") || strings.HasPrefix(strings.ToLower(defanged), "hxxp://") || strings.HasPrefix(strings.ToLower(defanged), "hxxps://") {
		return "url", canonicalizeURL(defanged)
	}
	if hashPattern.MatchString(defanged) {
		return "sha256", strings.ToLower(defanged)
	}
	if strings.Contains(defanged, ".") && !strings.Contains(defanged, " ") {
		return "domain_name", strings.ToLower(defanged)
	}
	return "", ""
}

func isHexString(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func stringPointer(value string) *string {
	return &value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
