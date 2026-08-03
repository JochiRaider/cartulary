package indicators

import (
	"strings"
	"testing"
)

func TestIndicatorIdentityCharacterization(t *testing.T) {
	t.Parallel()
	types := []struct {
		indicatorType string
		display       string
		wantDisplay   string
	}{
		{indicatorType: "ipv4_addr", display: "203[.]0[.]113[.]7", wantDisplay: "203.0.113.7"},
		{indicatorType: "ipv6_addr", display: "2001:0db8:0:0:0:0:0:1", wantDisplay: "2001:db8::1"},
		{indicatorType: "domain_name", display: "VPN[.]EXAMPLE.TEST", wantDisplay: "vpn.example.test"},
		{indicatorType: "url", display: "hxxps://EXAMPLE[.]TEST/A", wantDisplay: "https://example.test/A"},
		{indicatorType: "sha256", display: strings.Repeat("A", 64), wantDisplay: strings.Repeat("a", 64)},
		{indicatorType: "email_addr", display: "User@Example.TEST", wantDisplay: "user@example.test"},
		{indicatorType: "registry_key", display: `HKLM\Software\Example`, wantDisplay: `HKLM\Software\Example`},
		{indicatorType: "process_name", display: "PowerShell.EXE", wantDisplay: "PowerShell.EXE"},
		{indicatorType: "text", display: "  suspicious payload  ", wantDisplay: "suspicious payload"},
	}
	valueKinds := []string{"atomic", "pattern", "reference"}
	for _, indicatorType := range types {
		for _, valueKind := range valueKinds {
			if (indicatorType.indicatorType == "ipv4_addr" || indicatorType.indicatorType == "ipv6_addr") && valueKind != "atomic" {
				continue
			}
			name := indicatorType.indicatorType + "/" + valueKind
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				values := map[string]string{
					"indicator.indicator_type": indicatorType.indicatorType,
					"indicator.value_kind":     valueKind,
					"indicator.display_value":  indicatorType.display,
					"indicator.defanged_value": "presentation-only",
					"indicator.stix_pattern":   "[presentation:only = true]",
				}
				if indicatorType.indicatorType == "sha256" {
					values["indicator.hash_algorithm"] = "SHA256"
					values["indicator.hash_value"] = strings.Repeat("B", 64)
				}
				got, err := indicatorInputFromCreateRequest(CreateRequest{Values: values})
				if err != nil {
					t.Fatalf("canonicalize identity: %v", err)
				}
				if got.IndicatorType != indicatorType.indicatorType || got.ValueKind != valueKind {
					t.Fatalf("registry values = %s/%s", got.IndicatorType, got.ValueKind)
				}
				if got.DisplayValue != indicatorType.wantDisplay || got.NormalizedValue == nil || *got.NormalizedValue != indicatorType.wantDisplay {
					t.Fatalf("canonical value = %q/%v, want %q", got.DisplayValue, got.NormalizedValue, indicatorType.wantDisplay)
				}
				if len(got.DedupeKey) != 64 {
					t.Fatalf("dedupe key = %q", got.DedupeKey)
				}

				withoutPresentation := cloneStringMap(values)
				delete(withoutPresentation, "indicator.defanged_value")
				delete(withoutPresentation, "indicator.stix_pattern")
				identityOnly, err := indicatorInputFromCreateRequest(CreateRequest{Values: withoutPresentation})
				if err != nil {
					t.Fatalf("canonicalize identity without presentation fields: %v", err)
				}
				if identityOnly.DedupeKey != got.DedupeKey {
					t.Fatalf("presentation fields changed dedupe key: %q != %q", got.DedupeKey, identityOnly.DedupeKey)
				}
			})
		}
	}
}

func TestIndicatorIdentityCharacterizationRejectsAliasesAndIncompleteHashes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		values map[string]string
		field  string
	}{
		{name: "indicator type alias", values: identityValues("domain", "atomic", "example.test"), field: "indicator.indicator_type"},
		{name: "value kind alias", values: identityValues("domain_name", "literal", "example.test"), field: "indicator.value_kind"},
		{name: "hash algorithm without value", values: withIdentityValue(identityValues("sha256", "atomic", strings.Repeat("a", 64)), "indicator.hash_algorithm", "sha256"), field: "indicator.hash_value"},
		{name: "hash value without algorithm", values: withIdentityValue(identityValues("sha256", "atomic", strings.Repeat("a", 64)), "indicator.hash_value", strings.Repeat("b", 64)), field: "indicator.hash_value"},
		{name: "non hex hash", values: withIdentityValues(identityValues("sha256", "atomic", strings.Repeat("a", 64)), map[string]string{"indicator.hash_algorithm": "sha256", "indicator.hash_value": "not-hex"}), field: "indicator.hash_value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := indicatorInputFromCreateRequest(CreateRequest{Values: test.values})
			validation, ok := err.(*IndicatorCreateValidationError)
			if !ok || validation.Field != test.field || validation.ReasonCode != "invalid_value" {
				t.Fatalf("validation = %#v, want %s/invalid_value", err, test.field)
			}
		})
	}
}

func cloneStringMap(source map[string]string) map[string]string {
	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func identityValues(indicatorType string, valueKind string, displayValue string) map[string]string {
	return map[string]string{
		"indicator.indicator_type": indicatorType,
		"indicator.value_kind":     valueKind,
		"indicator.display_value":  displayValue,
	}
}

func withIdentityValue(values map[string]string, key string, value string) map[string]string {
	values[key] = value
	return values
}

func withIdentityValues(values map[string]string, additions map[string]string) map[string]string {
	for key, value := range additions {
		values[key] = value
	}
	return values
}
