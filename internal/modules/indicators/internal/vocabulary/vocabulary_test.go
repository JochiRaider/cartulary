package vocabulary_test

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
)

func TestClosedVocabularyMembershipAndCanonicalization(t *testing.T) {
	t.Parallel()
	families := []struct {
		name     string
		values   []string
		want     []string
		contains func(string) bool
	}{
		{name: "indicator types", values: vocabulary.IndicatorTypes(), want: []string{"ipv4_addr", "ipv6_addr", "domain_name", "url", "sha256", "email_addr", "registry_key", "process_name", "text"}, contains: vocabulary.IsIndicatorType},
		{name: "value kinds", values: vocabulary.ValueKinds(), want: []string{"atomic", "pattern", "reference"}, contains: vocabulary.IsValueKind},
		{name: "observation statuses", values: vocabulary.ObservationStatuses(), want: []string{"unresolved", "resolved", "dismissed"}, contains: vocabulary.IsObservationStatus},
		{name: "lifecycle states", values: vocabulary.LifecycleStates(), want: []string{"active", "benign", "false_positive", "retired"}, contains: vocabulary.IsLifecycleState},
	}
	for _, family := range families {
		family := family
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			if !reflect.DeepEqual(family.values, family.want) {
				t.Fatalf("registry = %#v, want %#v", family.values, family.want)
			}
			for _, value := range family.want {
				if !family.contains(value) {
					t.Fatalf("exact member %q was rejected", value)
				}
			}
			for _, value := range []string{"", "unknown", "alias", " active", "active ", "ACTIVE", "False_Positive"} {
				if family.contains(value) {
					t.Fatalf("non-member %q was accepted", value)
				}
			}
			family.values[0] = "mutated"
			if family.contains("mutated") || !family.contains(family.want[0]) {
				t.Fatal("caller mutation changed registry membership")
			}
		})
	}

	for input, want := range map[string]string{
		"domain_name":   "domain_name",
		" DOMAIN_NAME ": "domain_name",
		"IPv4_ADDR":     "ipv4_addr",
	} {
		if got, ok := vocabulary.CanonicalIndicatorType(input); !ok || got != want {
			t.Fatalf("canonical Indicator type %q = %q, %t, want %q", input, got, ok, want)
		}
	}
	for input, want := range map[string]string{
		"atomic":      "atomic",
		" REFERENCE ": "reference",
		"Pattern":     "pattern",
	} {
		if got, ok := vocabulary.CanonicalValueKind(input); !ok || got != want {
			t.Fatalf("canonical value kind %q = %q, %t, want %q", input, got, ok, want)
		}
	}
	for _, input := range []string{"domain", "ipv4", "literal", "scalar", "unknown"} {
		if _, ok := vocabulary.CanonicalIndicatorType(input); ok {
			t.Fatalf("Indicator type alias %q was canonicalized", input)
		}
		if _, ok := vocabulary.CanonicalValueKind(input); ok {
			t.Fatalf("value-kind alias %q was canonicalized", input)
		}
	}
}
