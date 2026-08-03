package identity

import (
	"errors"
	"strings"
	"testing"
)

func TestCanonicalizeRegistryAndDedupe(t *testing.T) {
	t.Parallel()
	types := []struct {
		name        string
		display     string
		wantDisplay string
	}{
		{name: "ipv4_addr", display: "203[.]0[.]113[.]7", wantDisplay: "203.0.113.7"},
		{name: "ipv6_addr", display: "2001:0db8:0:0:0:0:0:1", wantDisplay: "2001:db8::1"},
		{name: "domain_name", display: "VPN[.]EXAMPLE.TEST", wantDisplay: "vpn.example.test"},
		{name: "url", display: "hxxps://EXAMPLE[.]TEST/A", wantDisplay: "https://example.test/A"},
		{name: "sha256", display: strings.Repeat("A", 64), wantDisplay: strings.Repeat("a", 64)},
		{name: "email_addr", display: "User@Example.TEST", wantDisplay: "user@example.test"},
		{name: "registry_key", display: `HKLM\Software\Example`, wantDisplay: `HKLM\Software\Example`},
		{name: "process_name", display: "PowerShell.EXE", wantDisplay: "PowerShell.EXE"},
		{name: "text", display: "  suspicious payload  ", wantDisplay: "suspicious payload"},
	}
	for _, indicatorType := range types {
		for _, valueKind := range []string{"atomic", "pattern", "reference"} {
			indicatorType, valueKind := indicatorType, valueKind
			t.Run(indicatorType.name+"/"+valueKind, func(t *testing.T) {
				t.Parallel()
				input := Input{
					IndicatorType: indicatorType.name,
					ValueKind:     valueKind,
					DisplayValue:  indicatorType.display,
					DefangedValue: stringPointer("presentation-only"),
					STIXPattern:   stringPointer("[presentation:only = true]"),
				}
				if IsIPType(indicatorType.name) && valueKind != "atomic" {
					assertValidationField(t, input, "value_kind")
					return
				}
				if indicatorType.name == "sha256" {
					input.HashAlgorithm = stringPointer("SHA256")
					input.HashValue = stringPointer(strings.Repeat("B", 64))
				}
				canonical, err := Canonicalize(input)
				if err != nil {
					t.Fatalf("canonicalize identity: %v", err)
				}
				if canonical.IndicatorType != indicatorType.name || canonical.ValueKind != valueKind {
					t.Fatalf("registry values = %s/%s", canonical.IndicatorType, canonical.ValueKind)
				}
				if canonical.DisplayValue != indicatorType.wantDisplay || canonical.NormalizedValue == nil || *canonical.NormalizedValue != indicatorType.wantDisplay {
					t.Fatalf("canonical value = %q/%v, want %q", canonical.DisplayValue, canonical.NormalizedValue, indicatorType.wantDisplay)
				}
				if len(canonical.DedupeKey) != 64 {
					t.Fatalf("dedupe key = %q", canonical.DedupeKey)
				}

				input.DefangedValue = nil
				input.STIXPattern = nil
				identityOnly, err := Canonicalize(input)
				if err != nil {
					t.Fatalf("canonicalize without presentation fields: %v", err)
				}
				if identityOnly.DedupeKey != canonical.DedupeKey {
					t.Fatalf("presentation fields changed dedupe key: %q != %q", canonical.DedupeKey, identityOnly.DedupeKey)
				}
			})
		}
	}
}

func TestCanonicalizeRejectsAliasesAndMalformedRepresentations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input Input
		field string
	}{
		{name: "indicator type alias", input: Input{IndicatorType: "domain", ValueKind: "atomic", DisplayValue: "example.test"}, field: "indicator_type"},
		{name: "value kind alias", input: Input{IndicatorType: "domain_name", ValueKind: "literal", DisplayValue: "example.test"}, field: "value_kind"},
		{name: "hash algorithm without value", input: Input{IndicatorType: "sha256", ValueKind: "atomic", DisplayValue: strings.Repeat("a", 64), HashAlgorithm: stringPointer("sha256")}, field: "hash_value"},
		{name: "hash value without algorithm", input: Input{IndicatorType: "sha256", ValueKind: "atomic", DisplayValue: strings.Repeat("a", 64), HashValue: stringPointer(strings.Repeat("b", 64))}, field: "hash_value"},
		{name: "non hex hash", input: Input{IndicatorType: "sha256", ValueKind: "atomic", DisplayValue: strings.Repeat("a", 64), HashAlgorithm: stringPointer("sha256"), HashValue: stringPointer("not-hex")}, field: "hash_value"},
		{name: "ip hash pair", input: Input{IndicatorType: "ipv4_addr", ValueKind: "atomic", DisplayValue: "203.0.113.7", HashAlgorithm: stringPointer("sha256"), HashValue: stringPointer(strings.Repeat("a", 64))}, field: "hash_algorithm"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assertValidationField(t, test.input, test.field)
		})
	}
}

func TestIPCanonicalizationAndDedupe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		indicatorType string
		display       string
		normalized    *string
		want          string
	}{
		{name: "ipv4 canonical", indicatorType: "ipv4_addr", display: "203[.]0[.]113[.]7", want: "203.0.113.7"},
		{name: "ipv6 canonical", indicatorType: "ipv6_addr", display: "2001:0DB8:0000:0000:0000:0000:0000:0001", want: "2001:db8::1"},
		{name: "ipv6 normalized input", indicatorType: "ipv6_addr", display: "2001:db8::1", normalized: stringPointer("2001:0db8:0:0:0:0:0:1"), want: "2001:db8::1"},
	}
	var expandedKey string
	for _, test := range tests {
		canonical, err := Canonicalize(Input{IndicatorType: test.indicatorType, ValueKind: "atomic", DisplayValue: test.display, NormalizedValue: test.normalized})
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if canonical.DisplayValue != test.want || canonical.NormalizedValue == nil || *canonical.NormalizedValue != test.want {
			t.Fatalf("%s = %q/%v, want %q", test.name, canonical.DisplayValue, canonical.NormalizedValue, test.want)
		}
		if test.indicatorType == "ipv6_addr" {
			if expandedKey == "" {
				expandedKey = canonical.DedupeKey
			} else if canonical.DedupeKey != expandedKey {
				t.Fatalf("canonical-equivalent IPv6 values produced different dedupe keys")
			}
		}
	}

	for _, input := range []Input{
		{IndicatorType: "ipv4_addr", ValueKind: "atomic", DisplayValue: "203.0.113.007"},
		{IndicatorType: "ipv4_addr", ValueKind: "atomic", DisplayValue: "2001:db8::1"},
		{IndicatorType: "ipv6_addr", ValueKind: "atomic", DisplayValue: "203.0.113.7"},
		{IndicatorType: "ipv6_addr", ValueKind: "atomic", DisplayValue: "fe80::1%eth0"},
		{IndicatorType: "ipv6_addr", ValueKind: "atomic", DisplayValue: "::ffff:192.0.2.1"},
		{IndicatorType: "ipv6_addr", ValueKind: "atomic", DisplayValue: "64:ff9b::192.0.2.1"},
		{IndicatorType: "ipv6_addr", ValueKind: "atomic", DisplayValue: "2001:db8::1", NormalizedValue: stringPointer("2001:db8::2")},
	} {
		assertValidationField(t, input, "display_value", "normalized_value")
	}
}

func assertValidationField(t testing.TB, input Input, wantFields ...string) {
	t.Helper()
	_, err := Canonicalize(input)
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("validation = %#v, want ValidationError", err)
	}
	for _, want := range wantFields {
		if validation.Field == want && validation.ReasonCode == "invalid_value" {
			return
		}
	}
	t.Fatalf("validation = %s/%s, want one of %v/invalid_value", validation.Field, validation.ReasonCode, wantFields)
}
