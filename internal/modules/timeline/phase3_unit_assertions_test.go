package timeline

import (
	"slices"
	"testing"
)

func requireClosedVocabularyRejected(t testing.TB, code string, details map[string]any, wantField string, wantReasonCode string) {
	t.Helper()
	if code != "invalid_mutation_payload" && code != "invalid_view_query" {
		t.Fatalf("unexpected rejection code: %q", code)
	}
	if details["field"] != wantField {
		t.Fatalf("unexpected rejection field: got %v want %q", details["field"], wantField)
	}
	if details["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected rejection reason_code: got %v want %q", details["reason_code"], wantReasonCode)
	}
}

func requireWritableStringNormalization(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected normalized string: got %q want %q", got, want)
	}
}

func requireFieldKeyConformance(t testing.TB, fieldKeys []string, allowed []string) {
	t.Helper()
	if !slices.IsSorted(fieldKeys) {
		t.Fatalf("expected sorted field keys, got %v", fieldKeys)
	}
	for _, fieldKey := range fieldKeys {
		if !slices.Contains(allowed, fieldKey) {
			t.Fatalf("unexpected field key %q not in allowed set %v", fieldKey, allowed)
		}
	}
}
