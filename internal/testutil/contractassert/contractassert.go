package contractassert

import (
	"reflect"
	"slices"
	"testing"
)

func RequireClosedVocabularyRejected(t testing.TB, code string, details map[string]any, wantField string, wantReasonCode string) {
	t.Helper()
	if code == "" {
		t.Fatal("expected closed-vocabulary rejection code")
	}
	if code != "invalid_auth_request" && code != "invalid_mutation_payload" && code != "invalid_view_query" {
		t.Fatalf("unexpected closed-vocabulary rejection code: %q", code)
	}
	if details == nil {
		t.Fatal("expected closed-vocabulary rejection details")
	}
	if wantField != "" && details["field"] != wantField {
		t.Fatalf("unexpected closed-vocabulary field: got %v want %q", details["field"], wantField)
	}
	if wantReasonCode != "" && details["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected closed-vocabulary reason_code: got %v want %q", details["reason_code"], wantReasonCode)
	}
}

func RequireWritableStringNormalization(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected normalized string: got %q want %q", got, want)
	}
}

func RequireFieldKeyConformance(t testing.TB, fieldKeys []string, allowed []string) {
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

func RequireProjectionDeterminism(t testing.TB, first any, second any) {
	t.Helper()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("projection rebuild was not deterministic:\nfirst: %#v\nsecond: %#v", first, second)
	}
}

func RequireDefaultQueryMeta(t testing.TB, body map[string]any, viewSchemaID string) {
	t.Helper()
	_ = viewSchemaID

	metaValue, ok := body["meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected success envelope meta object, got %T", body["meta"])
	}
	pagingValue, ok := metaValue["paging"].(map[string]any)
	if !ok {
		t.Fatalf("expected meta.paging object, got %T", metaValue["paging"])
	}
	if pagingValue["limit"] != float64(100) {
		t.Fatalf("expected default meta.paging.limit=100, got %#v", pagingValue)
	}
	if _, ok := pagingValue["has_more"].(bool); !ok {
		t.Fatalf("expected meta.paging.has_more boolean, got %#v", pagingValue)
	}
	if nextCursor := pagingValue["next_cursor"]; nextCursor != nil {
		if _, ok := nextCursor.(string); !ok {
			t.Fatalf("expected meta.paging.next_cursor string or null, got %#v", pagingValue)
		}
	}
}
