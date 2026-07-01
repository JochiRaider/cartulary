package entities

import "strings"

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string {
	return "entities: invalid mutation request"
}

func ValidatePartyCreateParams(params PartyCreateParams) error {
	if !hasPartyText(params.Values, "party.display_name") {
		return &ValidationError{Field: "party.display_name", ReasonCode: "missing_required_field"}
	}
	if !validPartyText(params.Values, "party.party_kind", ValidPartyKind) {
		return &ValidationError{Field: "party.party_kind", ReasonCode: "missing_required_field"}
	}
	return nil
}

func ValidPartyKind(value string) bool {
	switch value {
	case "person", "team", "organization", "distribution_list", "other":
		return true
	default:
		return false
	}
}

func hasPartyText(values map[string]PartyFieldValue, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && strings.TrimSpace(*value.Text) != ""
}

func validPartyText(values map[string]PartyFieldValue, field string, predicate func(string) bool) bool {
	value, ok := values[field]
	if !ok || value.Text == nil {
		return false
	}
	return predicate(*value.Text)
}
