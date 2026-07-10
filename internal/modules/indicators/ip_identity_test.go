package indicators

import (
	"errors"
	"testing"
)

func TestIPLiteralIndicatorNormalization(t *testing.T) {
	tests := []struct {
		name          string
		indicatorType string
		display       string
		normalized    string
		want          string
	}{
		{
			name:          "ipv4 canonical",
			indicatorType: "ipv4_addr",
			display:       "203[.]0[.]113[.]7",
			want:          "203.0.113.7",
		},
		{
			name:          "ipv6 rfc5952 lowercase compression",
			indicatorType: "ipv6_addr",
			display:       "2001:0DB8:0000:0000:0000:0000:0000:0001",
			want:          "2001:db8::1",
		},
		{
			name:          "ipv6 normalized input must match canonical display",
			indicatorType: "ipv6_addr",
			display:       "2001:db8::1",
			normalized:    "2001:0db8:0:0:0:0:0:1",
			want:          "2001:db8::1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var normalized *string
			if test.normalized != "" {
				normalized = &test.normalized
			}
			display, normalizedValue, err := normalizeIndicatorValue(test.indicatorType, test.display, normalized)
			if err != nil {
				t.Fatalf("normalize indicator value: %v", err)
			}
			if display != test.want {
				t.Fatalf("display mismatch: got %q want %q", display, test.want)
			}
			if normalizedValue == nil || *normalizedValue != test.want {
				t.Fatalf("normalized mismatch: got %#v want %q", normalizedValue, test.want)
			}
		})
	}
}

func TestIPLiteralIndicatorNormalizationRejectsAmbiguousForms(t *testing.T) {
	tests := []struct {
		name          string
		indicatorType string
		display       string
		normalized    string
	}{
		{name: "ipv4 leading zero", indicatorType: "ipv4_addr", display: "203.0.113.007"},
		{name: "ipv4 family mismatch", indicatorType: "ipv4_addr", display: "2001:db8::1"},
		{name: "ipv6 family mismatch", indicatorType: "ipv6_addr", display: "203.0.113.7"},
		{name: "ipv6 zone id", indicatorType: "ipv6_addr", display: "fe80::1%eth0"},
		{name: "ipv6 mapped ipv4", indicatorType: "ipv6_addr", display: "::ffff:192.0.2.1"},
		{name: "ipv6 dotted quad suffix", indicatorType: "ipv6_addr", display: "64:ff9b::192.0.2.1"},
		{name: "normalized mismatch", indicatorType: "ipv6_addr", display: "2001:db8::1", normalized: "2001:db8::2"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var normalized *string
			if test.normalized != "" {
				normalized = &test.normalized
			}
			if _, _, err := normalizeIndicatorValue(test.indicatorType, test.display, normalized); err == nil {
				t.Fatalf("expected %s %q to be rejected", test.indicatorType, test.display)
			}
		})
	}
}

func TestIPLiteralIndicatorCreateValidation(t *testing.T) {
	tests := []struct {
		name      string
		values    map[string]string
		wantField string
	}{
		{
			name: "ip value kind must be atomic",
			values: map[string]string{
				"indicator.indicator_type": "ipv6_addr",
				"indicator.value_kind":     "pattern",
				"indicator.display_value":  "2001:db8::1",
			},
			wantField: "indicator.value_kind",
		},
		{
			name: "ip indicator rejects hash pair",
			values: map[string]string{
				"indicator.indicator_type": "ipv4_addr",
				"indicator.value_kind":     "atomic",
				"indicator.display_value":  "203.0.113.7",
				"indicator.hash_algorithm": "sha256",
				"indicator.hash_value":     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			wantField: "indicator.hash_algorithm",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := indicatorInputFromCreateRequest(CreateRequest{Values: test.values})
			if err == nil {
				t.Fatalf("expected create validation failure")
			}
			var validation *IndicatorCreateValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("expected IndicatorCreateValidationError, got %T", err)
			}
			if validation.Field != test.wantField || validation.ReasonCode != "invalid_value" {
				t.Fatalf("validation mismatch: got %s/%s want %s/invalid_value", validation.Field, validation.ReasonCode, test.wantField)
			}
		})
	}
}

func TestIPv6IndicatorDedupeUsesCanonicalValue(t *testing.T) {
	display, normalizedValue, err := normalizeIndicatorValue("ipv6_addr", "2001:0db8:0:0:0:0:0:1", nil)
	if err != nil {
		t.Fatalf("normalize expanded ipv6: %v", err)
	}
	canonical := indicatorUpsertInput{
		IndicatorType:   "ipv6_addr",
		ValueKind:       "atomic",
		DisplayValue:    display,
		NormalizedValue: normalizedValue,
	}
	shortDisplay, shortNormalizedValue, err := normalizeIndicatorValue("ipv6_addr", "2001:db8::1", nil)
	if err != nil {
		t.Fatalf("normalize compressed ipv6: %v", err)
	}
	compressed := indicatorUpsertInput{
		IndicatorType:   "ipv6_addr",
		ValueKind:       "atomic",
		DisplayValue:    shortDisplay,
		NormalizedValue: shortNormalizedValue,
	}
	if buildIndicatorDedupeKey(canonical) != buildIndicatorDedupeKey(compressed) {
		t.Fatalf("canonical-equivalent IPv6 values produced different dedupe keys")
	}
}
