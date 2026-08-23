package vocabulary

import "strings"

var (
	indicatorTypes = [...]string{
		"ipv4_addr",
		"ipv6_addr",
		"domain_name",
		"url",
		"sha256",
		"email_addr",
		"registry_key",
		"process_name",
		"text",
	}
	valueKinds = [...]string{
		"atomic",
		"pattern",
		"reference",
	}
	observationStatuses = [...]string{
		"unresolved",
		"resolved",
		"dismissed",
	}
	lifecycleStates = [...]string{
		"active",
		"benign",
		"false_positive",
		"retired",
	}

	indicatorTypeSet     = membership(indicatorTypes[:])
	valueKindSet         = membership(valueKinds[:])
	observationStatusSet = membership(observationStatuses[:])
	lifecycleStateSet    = membership(lifecycleStates[:])
)

// IndicatorTypes returns a defensive copy of the closed Indicator type registry.
func IndicatorTypes() []string { return clone(indicatorTypes[:]) }

// ValueKinds returns a defensive copy of the closed Indicator value-kind registry.
func ValueKinds() []string { return clone(valueKinds[:]) }

// ObservationStatuses returns a defensive copy of the closed observation-status registry.
func ObservationStatuses() []string { return clone(observationStatuses[:]) }

// LifecycleStates returns a defensive copy of the closed lifecycle-state registry.
func LifecycleStates() []string { return clone(lifecycleStates[:]) }

// IsIndicatorType reports exact membership without canonicalizing the input.
func IsIndicatorType(value string) bool { return contains(indicatorTypeSet, value) }

// IsValueKind reports exact membership without canonicalizing the input.
func IsValueKind(value string) bool { return contains(valueKindSet, value) }

// IsObservationStatus reports exact membership without canonicalizing the input.
func IsObservationStatus(value string) bool { return contains(observationStatusSet, value) }

// IsLifecycleState reports exact membership without canonicalizing the input.
func IsLifecycleState(value string) bool { return contains(lifecycleStateSet, value) }

// CanonicalIndicatorType admits the deliberately tolerant identity-creation form.
func CanonicalIndicatorType(value string) (string, bool) {
	canonical := strings.ToLower(strings.TrimSpace(value))
	return canonical, IsIndicatorType(canonical)
}

// CanonicalValueKind admits the deliberately tolerant identity-creation form.
func CanonicalValueKind(value string) (string, bool) {
	canonical := strings.ToLower(strings.TrimSpace(value))
	return canonical, IsValueKind(canonical)
}

func membership(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func contains(values map[string]struct{}, value string) bool {
	_, present := values[value]
	return present
}

func clone(values []string) []string { return append([]string(nil), values...) }
