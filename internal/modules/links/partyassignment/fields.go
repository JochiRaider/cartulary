package partyassignment

import "strings"

func Fields() []string {
	return []string{
		"host.subject_party_ids", "host.responsible_party_ids", "host.observer_party_ids",
		"identity.subject_party_ids", "identity.responsible_party_ids", "identity.observer_party_ids",
	}
}

// IsField identifies only the explicitly owned collections.
// Assignment membership is not an evidence-support relationship.
func IsField(field string) bool {
	_, _, ok := Field(field)
	return ok
}

func Field(field string) (string, string, bool) {
	family, member, ok := strings.Cut(field, ".")
	if !ok || (family != "host" && family != "identity") {
		return "", "", false
	}
	switch member {
	case "subject_party_ids":
		return family, "subject_party", true
	case "responsible_party_ids":
		return family, "responsible_party", true
	case "observer_party_ids":
		return family, "observer_party", true
	default:
		return "", "", false
	}
}
