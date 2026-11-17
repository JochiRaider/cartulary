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

// REQ-01-488 / REQ-01-067: the label sort key is NFC then locale-independent
// case folding, and for tags it equals the stored normalized_tag_name.
func TestTextSortKeyFoldsNFCAndMatchesTagNormalization(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{"Alpha", "alpha"},
		{"ALPHA", "alpha"},
		{"Straße", "strasse"},
		{"Café", "café"},
		{"Café", "café"},
		{"Ω", "ω"},
		{"", ""},
	} {
		if got := TextSortKey(tc.raw); got != tc.want {
			t.Fatalf("TextSortKey(%q) = %q, want %q", tc.raw, got, tc.want)
		}
	}
	// Diacritics stay significant under the key.
	if TextSortKey("cafe") == TextSortKey("café") {
		t.Fatal("diacritics must remain significant")
	}
	for _, raw := range []string{"  Evidence Follow-Up  ", "Hidden tag Ω 東京", "Straße", "Café"} {
		label, normalized, ok := NormalizeTagLabel(raw)
		if !ok {
			t.Fatalf("tag %q rejected", raw)
		}
		if key := TextSortKey(label); key != normalized {
			t.Fatalf("tag %q: TextSortKey(label)=%q, normalized=%q", raw, key, normalized)
		}
	}
}
