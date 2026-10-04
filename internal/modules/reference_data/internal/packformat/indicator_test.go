package packformat

import (
	"strings"
	"testing"
)

// Expected canonical values are authored independently of the normalizers.
func TestIndicatorAlgorithmsCanonicalVectors_Unit(t *testing.T) {
	vectors := []struct{ typ, raw, canonical, defanged string }{
		{"ipv4_addr", " 192.0.2.7\u00a0", "192.0.2.7", "192[.]0[.]2[.]7"},
		{"ipv6_addr", "2001:0DB8:0:0:1:0:0:1", "2001:db8::1:0:0:1", "2001[:]db8[:][:]1[:]0[:]0[:]1"},
		{"domain_name", "EXAMPLE.COM.", "example.com", "example[.]com"},
		{"url", "HTTPS://Example.COM:0443/a/%2e%2E/b/%7e?q=%2f&x=%7e#", "https://example.com/b/~?q=%2F&x=~#", "hxxps://example[.]com/b/~?q=%2F&x=~#"},
		{"sha256", strings.Repeat("AB", 32), strings.Repeat("ab", 32), strings.Repeat("ab", 32)},
		{"email_addr", "User+Case@EXAMPLE.COM.", "User+Case@example.com", "User+Case[@]example[.]com"},
		{"registry_key", `hklm//Software\Vendor\`, `HKEY_LOCAL_MACHINE\SOFTWARE\VENDOR`, `HKEY_LOCAL_MACHINE\SOFTWARE\VENDOR`},
		{"process_name", " Cafe\u0301.exe ", "Café.exe", "Café.exe"},
		{"text", " A\r\nB\tC\rD ", "A\nB\tC\nD", "A\nB\tC\nD"},
	}
	for _, v := range vectors {
		t.Run(v.typ, func(t *testing.T) {
			got, err := Evaluate(indicatorPolicies[v.typ], "atomic", v.raw)
			if err != nil || !got.Valid || got.Normalized == nil || *got.Normalized != v.canonical || *got.Display != v.canonical || *got.Defanged != v.defanged || got.Code != nil {
				t.Fatalf("evaluation: %#v, %v", got, err)
			}
			again, err := Evaluate(indicatorPolicies[v.typ], "atomic", v.canonical)
			if err != nil || !again.Valid || *again.Normalized != v.canonical || *again.DedupeKey != *got.DedupeKey {
				t.Fatalf("non-idempotent evaluation: %#v, %v", again, err)
			}
		})
	}
	invalid := map[string][]string{
		"ipv4_addr":    {"01.2.3.4", "256.2.3.4", "1.2.3.4/32"},
		"ipv6_addr":    {"::ffff:c000:201", "::c000:201", "2001:db8::1%eth0", "2001:db8::192.0.2.1"},
		"domain_name":  {"example..", "é.example", "-example.com", "example[.]com"},
		"email_addr":   {"a..b@example.com", "\"a\"@example.com", "a@"},
		"process_name": {"a/b.exe", "a\\b.exe", "x\x7f"},
		"registry_key": {"OTHER\\Path", "HKLM\\x\x00"},
		"text":         {"\n\t", "a\u0085b"},
	}
	for typ, values := range invalid {
		for _, raw := range values {
			got, err := Evaluate(indicatorPolicies[typ], "atomic", raw)
			if err != nil || got.Valid || got.Code == nil || got.Display != nil || got.Normalized != nil || got.Defanged != nil || got.DedupeKey != nil || got.NormalizationAlgorithm == "" {
				t.Fatalf("invalid %s result: %#v, %v", typ, got, err)
			}
		}
	}
	for _, typ := range []string{"sha256", "email_addr"} {
		if _, err := Evaluate(indicatorPolicies[typ], "reference", "private value"); err == nil {
			t.Fatal("unsupported kind admitted")
		}
	}
	got, err := Evaluate(indicatorPolicies["text"], "atomic", strings.Repeat("é", 8193))
	if err == nil || got.Raw != "" {
		t.Fatal("oversize input was echoed")
	}
}

func TestURLNormalizationPreservesDelimitersAndIsIdempotent_Unit(t *testing.T) {
	vectors := map[string]string{
		"http://EXAMPLE.com":                          "http://example.com/",
		"http://example.com?#":                        "http://example.com/?#",
		"http://example.com/a/%2e/b/%2E%2e/c":         "http://example.com/a/c",
		"http://example.com/a%2fb/%3f?q=%23%3f%26%3d": "http://example.com/a%2Fb/%3F?q=%23%3F%26%3D",
		"http://example.com/a//b/../":                 "http://example.com/a//",
		"https://[2001:DB8::1]:0443?q=a&q=b":          "https://[2001:db8::1]/?q=a&q=b",
		"http://192.0.2.1.":                           "http://192.0.2.1/",
		"http://example.com/a/../../b":                "http://example.com/b",
	}
	for raw, want := range vectors {
		got, ok := NormalizeURL(raw)
		if !ok || got != want {
			t.Fatalf("%q = %q, %v; want %q", raw, got, ok, want)
		}
		again, ok := NormalizeURL(got)
		if !ok || again != got {
			t.Fatalf("not idempotent: %q", got)
		}
	}
	for _, raw := range []string{"http://01.2.3.4.", "http://example.com..", "http://example.com/%", "http://example.com/a b", "http://user@example.com", "http://example.com:0", "http://[::ffff:c000:201]", "http://example.com/#a#b", "http://example.com/\\x"} {
		if got, ok := NormalizeURL(raw); ok {
			t.Fatalf("invalid URI admitted: %q", got)
		}
	}
}
