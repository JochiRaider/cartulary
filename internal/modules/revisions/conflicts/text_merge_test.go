package conflicts

import "testing"

func TestSuggestedTextMergeValueSuggestsCleanNonOverlappingLineEdits(t *testing.T) {
	suggested, ok := SuggestedTextMergeValue("one\r\ntwo\nthree", "one\r\nTWO\nthree", "one\ntwo\nTHREE")
	if !ok {
		t.Fatal("expected clean non-overlapping edits to produce a suggestion")
	}
	if suggested != "one\nTWO\nTHREE" {
		t.Fatalf("suggested merge = %q, want %q", suggested, "one\nTWO\nTHREE")
	}
}

func TestSuggestedTextMergeValueRejectsOverlappingOrNonTextEdits(t *testing.T) {
	if suggested, ok := SuggestedTextMergeValue("one\ntwo", "one\nserver", "one\nclient"); ok {
		t.Fatalf("overlapping line merge must not suggest a value, got %q", suggested)
	}
	if suggested, ok := SuggestedTextMergeValue("one\ntwo", 12, "one\nclient"); ok {
		t.Fatalf("non-text merge must not suggest a value, got %q", suggested)
	}
}
