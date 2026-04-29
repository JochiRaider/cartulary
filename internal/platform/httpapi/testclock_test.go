package httpapi

import (
	"testing"
	"time"
)

func TestTestClockFixedTimeAndAdvance(t *testing.T) {
	clock := NewTestClock()
	fixed := time.Date(2026, time.April, 29, 13, 35, 4, 261991000, time.FixedZone("EDT", -4*60*60))
	wantFixed := fixed.UTC()

	if got := clock.SetFixed(fixed); !got.Equal(wantFixed) {
		t.Fatalf("SetFixed returned %s, want %s", got, wantFixed)
	}
	if got := clock.Now(); !got.Equal(wantFixed) {
		t.Fatalf("Now returned %s, want fixed %s", got, wantFixed)
	}

	wantAdvanced := wantFixed.Add(1500 * time.Millisecond)
	if got := clock.Advance(1500 * time.Millisecond); !got.Equal(wantAdvanced) {
		t.Fatalf("Advance returned %s, want %s", got, wantAdvanced)
	}
	if got := clock.Now(); !got.Equal(wantAdvanced) {
		t.Fatalf("Now returned %s, want advanced fixed %s", got, wantAdvanced)
	}
}

func TestTestClockSetOffsetClearsFixedTime(t *testing.T) {
	clock := NewTestClock()
	clock.SetFixed(time.Date(2035, time.January, 1, 0, 0, 0, 0, time.UTC))

	before := time.Now().UTC().Add(-100 * time.Millisecond)
	if got := clock.SetOffset(0); got.Before(before) {
		t.Fatalf("SetOffset returned stale time %s before %s", got, before)
	}
	after := time.Now().UTC().Add(100 * time.Millisecond)
	if got := clock.Now(); got.Before(before) || got.After(after) {
		t.Fatalf("SetOffset should clear fixed time; Now returned %s outside [%s, %s]", got, before, after)
	}
}
