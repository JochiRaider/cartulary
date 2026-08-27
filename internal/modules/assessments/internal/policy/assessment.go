package policy

const MaxInitialSupportReferences = 64

func ValidSubjectType(value string) bool {
	return value == "host" || value == "identity"
}

func ValidState(value string) bool {
	switch value {
	case "unknown", "suspected", "confirmed", "disproven", "cleared":
		return true
	default:
		return false
	}
}

func ConfidenceBand(score *int) (string, bool) {
	switch {
	case score == nil:
		return "unset", true
	case *score >= 0 && *score <= 39:
		return "low", true
	case *score >= 40 && *score <= 69:
		return "medium", true
	case *score >= 70 && *score <= 100:
		return "high", true
	default:
		return "", false
	}
}
