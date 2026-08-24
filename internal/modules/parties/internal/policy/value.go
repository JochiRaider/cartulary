package policy

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/gen/partypolicy"
	"github.com/JochiRaider/cartulary/internal/gen/stringcontracts"
)

type Error struct {
	ReasonCode string
}

func (e *Error) Error() string { return "parties: field admission failed" }

// Value is the immutable result of applying one Party registry entry. All
// representations are calculated together so consumers cannot normalize the
// same source value differently.
type Value struct {
	fieldKey             string
	storedValue          *string
	equalityValue        *string
	exactMatchClaimValue *string
	canonicalHashValue   any
}

func (v Value) FieldKey() string { return v.fieldKey }

func (v Value) StoredValue() (string, bool) {
	return optionalString(v.storedValue)
}

func (v Value) EqualityValue() (string, bool) {
	return optionalString(v.equalityValue)
}

func (v Value) ExactMatchClaimValue() (string, bool) {
	return optionalString(v.exactMatchClaimValue)
}

func (v Value) CanonicalHashValue() any { return v.canonicalHashValue }

func FieldKeys() []string { return partypolicy.FieldKeys() }

func LookupField(fieldKey string) (partypolicy.Field, bool) {
	return partypolicy.LookupField(fieldKey)
}

// Admit applies the sole Party field registry. A nil raw value represents JSON
// null. Optional normalized-empty strings produce the same authoritative-null
// result; required fields reject both forms.
func Admit(fieldKey string, raw *string) (Value, *Error) {
	field, ok := partypolicy.LookupField(fieldKey)
	if !ok {
		return Value{}, &Error{ReasonCode: "unsupported_field_key"}
	}
	if raw == nil {
		if field.Clearable {
			return nullValue(fieldKey), nil
		}
		return Value{}, &Error{ReasonCode: "field_not_nullable"}
	}

	stored, reason := normalize(field, *raw)
	if reason != "" {
		if reason == "normalized_empty" && field.Clearable {
			return nullValue(fieldKey), nil
		}
		if reason == "normalized_empty" && field.Required {
			reason = "field_required"
		}
		return Value{}, &Error{ReasonCode: reason}
	}

	equality := stored
	if field.EqualityPosture == "locale_independent_case_insensitive" {
		equality = strings.ToLower(stored)
	}
	value := Value{
		fieldKey:           fieldKey,
		storedValue:        stringPointer(stored),
		equalityValue:      stringPointer(equality),
		canonicalHashValue: stored,
	}
	if field.ClaimRole != "" {
		claim := equality
		value.exactMatchClaimValue = stringPointer(claim)
	}
	if field.CanonicalHashPosture == "equality_value_or_null" {
		value.canonicalHashValue = equality
	}
	return value, nil
}

// AdmitStored verifies a stored or portable Party representation and returns
// the same immutable value used for request admission. Callers that require
// canonical storage can compare the returned StoredValue with their input.
func AdmitStored(fieldKey string, stored *string) (Value, *Error) {
	return Admit(fieldKey, stored)
}

func normalize(field partypolicy.Field, raw string) (string, string) {
	if field.FieldKey == "party.party_kind" {
		if slices.Contains(field.EnumValues, raw) {
			return raw, ""
		}
		return "", "invalid_enum_value"
	}

	var normalized string
	if field.StringContractID == "multiline_body_v1" {
		normalized = norm.NFC.String(raw)
		normalized = strings.ReplaceAll(normalized, "\r\n", "\n")
		normalized = strings.ReplaceAll(normalized, "\r", "\n")
		normalized = strings.TrimFunc(normalized, unicode.IsSpace)
	} else {
		normalized = norm.NFC.String(strings.TrimFunc(raw, unicode.IsSpace))
		if field.StringContractID == "display_name_line_v1" {
			normalized = collapseWhitespace(normalized)
		}
	}
	if normalized == "" {
		return "", "normalized_empty"
	}
	if !utf8.ValidString(normalized) {
		return "", "invalid_utf8"
	}
	for _, character := range normalized {
		if field.StringContractID == "multiline_body_v1" && (character == '\n' || character == '\t') {
			continue
		}
		if unicode.Is(unicode.Cc, character) || unicode.Is(unicode.Cf, character) {
			return "", "control_character_not_allowed"
		}
	}
	if utf8.RuneCountInString(normalized) > field.MaxUnicodeScalars {
		return "", "max_unicode_scalars_exceeded"
	}

	switch field.StringContractID {
	case "email_address_v1":
		for _, character := range normalized {
			if unicode.IsSpace(character) {
				return "", "invalid_email_address"
			}
		}
		parts := strings.Split(normalized, "@")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", "invalid_email_address"
		}
	case "timezone_name_v1":
		if _, ok := stringcontracts.LookupTimezoneName(normalized); !ok {
			return "", "invalid_timezone_name"
		}
	}
	return normalized, ""
}

func collapseWhitespace(value string) string {
	var output strings.Builder
	output.Grow(len(value))
	pendingSpace := false
	for _, character := range value {
		if unicode.IsSpace(character) {
			pendingSpace = output.Len() > 0
			continue
		}
		if pendingSpace {
			output.WriteByte(' ')
			pendingSpace = false
		}
		output.WriteRune(character)
	}
	return output.String()
}

func nullValue(fieldKey string) Value {
	return Value{fieldKey: fieldKey, canonicalHashValue: nil}
}

func optionalString(value *string) (string, bool) {
	if value == nil {
		return "", false
	}
	return *value, true
}

func stringPointer(value string) *string { return &value }
