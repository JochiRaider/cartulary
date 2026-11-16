package fieldnorm

import (
	"strings"
	"testing"
)

func TestNormalizeAliasTextV1(t *testing.T) {
	normalized, ok := NormalizeAliasText("\u2003Cafe\u0301  Gateway\u2003")
	if !ok || normalized != "Café  Gateway" {
		t.Fatalf("unexpected normalized alias: %q ok=%v", normalized, ok)
	}
	if _, ok := NormalizeAliasText("gateway\u0085alias"); ok {
		t.Fatal("C1 control must be rejected")
	}
	if _, ok := NormalizeAliasText(strings.Repeat("x", 257)); ok {
		t.Fatal("aliases longer than 256 Unicode scalars must be rejected")
	}
	// The bound counts scalars, not bytes: 256 two-byte scalars are accepted.
	if normalized, ok := NormalizeAliasText(strings.Repeat("é", 256)); !ok || normalized != strings.Repeat("é", 256) {
		t.Fatalf("256-scalar alias must be accepted: %q ok=%v", normalized, ok)
	}
	if _, ok := NormalizeAliasText(strings.Repeat("é", 257)); ok {
		t.Fatal("257-scalar alias must be rejected")
	}
	if normalized, ok := NormalizeAliasText("gateway\u200Balias"); !ok || normalized != "gateway\u200Balias" {
		t.Fatalf("non-C0/C1 format character must not be broadened into rejection: %q ok=%v", normalized, ok)
	}
}
