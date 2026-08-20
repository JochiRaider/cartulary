package mutationpolicy

import (
	"slices"
	"strings"
	"testing"
)

func TestTimelineMutationPolicyIsClosedAndImmutable(t *testing.T) {
	wantFields := []string{
		"timeline.activity_local_text",
		"timeline.activity_synopsis_text",
		"timeline.activity_utc_text",
		"timeline.analyst_text",
		"timeline.data_source_text",
		"timeline.date_entered_text",
		"timeline.device_object_text",
		"timeline.ip_address_text",
		"timeline.mitre_stage_text",
		"timeline.raw_activity_text",
	}
	gotFields := DirectWritableFieldKeys()
	if !slices.Equal(gotFields, wantFields) {
		t.Fatalf("direct writable fields = %#v, want %#v", gotFields, wantFields)
	}
	gotFields[0] = "timeline.mutated"
	if !slices.Equal(DirectWritableFieldKeys(), wantFields) {
		t.Fatal("caller mutation changed the Timeline policy field set")
	}
	for _, fieldKey := range wantFields {
		if !IsDirectWritableField(fieldKey) {
			t.Fatalf("expected direct writable field %q", fieldKey)
		}
	}
	for _, fieldKey := range []string{"", "timeline.tags", "timeline.recorded_at", "timeline.unknown"} {
		if IsDirectWritableField(fieldKey) {
			t.Fatalf("unexpected direct writable field %q", fieldKey)
		}
	}
	if MaxPatchChanges != 32 || MaxCollectionActions != 64 || MaxVisibleTextRunes != 32_768 {
		t.Fatalf("unexpected mutation limits: changes=%d actions=%d runes=%d", MaxPatchChanges, MaxCollectionActions, MaxVisibleTextRunes)
	}
}

func TestTimelineVisibleTextPolicyBoundaries(t *testing.T) {
	for _, runeCount := range []int{MaxVisibleTextRunes - 1, MaxVisibleTextRunes} {
		if !IsValidVisibleText(strings.Repeat("界", runeCount)) {
			t.Fatalf("expected %d-rune value to be admitted", runeCount)
		}
	}
	if IsValidVisibleText(strings.Repeat("界", MaxVisibleTextRunes+1)) {
		t.Fatal("oversized visible text unexpectedly admitted")
	}
	for _, value := range []string{"", "plain", "tab\there", "line\nfeed", "carriage\rreturn"} {
		if !IsValidVisibleText(value) {
			t.Fatalf("allowed visible text rejected: %q", value)
		}
	}
	for _, value := range []string{"bad\x00value", "bad\x01value", "bad\x7fvalue", "bad\u0085value", "bad\u009fvalue"} {
		if IsValidVisibleText(value) {
			t.Fatalf("forbidden control admitted: %q", value)
		}
	}
}
