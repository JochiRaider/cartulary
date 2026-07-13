package networkflow

import (
	"testing"

	norm "github.com/JochiRaider/cartulary/internal/gen/networkflowunicode"
)

func TestUnicode17TextCanonicalization(t *testing.T) {
	t.Parallel()
	assertUnicode17TextCanonicalization(t)
}

func assertUnicode17TextCanonicalization(t *testing.T) {
	t.Helper()
	if norm.Version != "17.0.0" {
		t.Fatalf("Network Flow NFC tables = %q, want Unicode 17.0.0", norm.Version)
	}
	included := []rune{
		0x0009, 0x000a, 0x000b, 0x000c, 0x000d, 0x0020, 0x0085, 0x00a0,
		0x1680, 0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006,
		0x2007, 0x2008, 0x2009, 0x200a, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000,
	}
	for _, scalar := range included {
		if got := trimUnicodeWhitespace(string(scalar) + "e\u0301" + string(scalar)); got != "e\u0301" {
			t.Fatalf("U+%04X was not trimmed: %q", scalar, got)
		}
	}
	for _, scalar := range []rune{0x180e, 0x200b, 0xfeff} {
		value := string(scalar) + "x" + string(scalar)
		if got := trimUnicodeWhitespace(value); got != value {
			t.Fatalf("excluded U+%04X was trimmed: %q", scalar, got)
		}
	}
	if got, err := NormalizeTableDisplayNameInput(" \u0065\u0301 "); err != nil || got != "\u00e9" {
		t.Fatalf("NFC normalization = %q, %v", got, err)
	}
}
