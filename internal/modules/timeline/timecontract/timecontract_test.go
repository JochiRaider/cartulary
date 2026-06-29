package timecontract

import (
	"testing"
	"time"
)

func TestTimelineTimeContractParsingAndFormatting(t *testing.T) {
	for _, input := range []string{
		"2026-06-28T17:34Z",
		"2026-06-28T17:34:00Z",
		"2026-06-28 17:34Z",
		"2026-06-28 17:34:00Z",
	} {
		parsed, ok := ParseUTC(&input)
		if !ok {
			t.Fatalf("expected UTC input %q to parse", input)
		}
		if got := FormatUTC(parsed); got != "2026-06-28T17:34:00Z" {
			t.Fatalf("unexpected UTC normalization for %q: %s", input, got)
		}
	}

	for _, input := range []string{
		"2026-06-28T12:34",
		"2026-06-28T12:34:00",
		"2026-06-28 12:34",
		"2026-06-28 12:34:00",
	} {
		parsed, ok := ParseLocalWithFixedOffset(&input, -300)
		if !ok {
			t.Fatalf("expected local input %q to parse", input)
		}
		if got := FormatUTC(parsed.UTC); got != "2026-06-28T17:34:00Z" {
			t.Fatalf("unexpected local-to-UTC normalization for %q: %s", input, got)
		}
		if parsed.OffsetSeconds != -18000 || parsed.HasWireOffset {
			t.Fatalf("unexpected local parse metadata for %q: %#v", input, parsed)
		}
	}

	offsetInput := "2026-06-28T12:34:56-05:00"
	parsed, ok := ParseLocalWithFixedOffset(&offsetInput, -300)
	if !ok {
		t.Fatalf("expected offset local input to parse")
	}
	if got := FormatUTC(parsed.UTC); got != "2026-06-28T17:34:56Z" {
		t.Fatalf("unexpected offset local UTC normalization: %s", got)
	}
	if parsed.OffsetSeconds != -18000 || !parsed.HasWireOffset {
		t.Fatalf("unexpected offset parse metadata: %#v", parsed)
	}
	if got := FormatLocal(parsed.UTC, -300); got != "2026-06-28T12:34:56-05:00" {
		t.Fatalf("unexpected local offset formatting: %s", got)
	}

	for _, input := range []string{
		"2026-06-28T17:34:00+00:00",
		"2026-06-28",
		" 2026-06-28T17:34:00Z",
		"2026-06-28T17:34:00Z ",
	} {
		if _, ok := ParseUTC(&input); ok {
			t.Fatalf("expected UTC parser to reject %q", input)
		}
	}
	if _, ok := ParseLocalWithFixedOffset(nil, -300); ok {
		t.Fatalf("nil local input must not parse")
	}
	if got := FormatUTC(time.Date(2026, 6, 28, 17, 34, 0, 0, time.UTC)); got != "2026-06-28T17:34:00Z" {
		t.Fatalf("unexpected UTC format: %s", got)
	}
}
