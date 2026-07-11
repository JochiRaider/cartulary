package contractassert

import (
	"reflect"
	"slices"
	"testing"
)

type AuthorizationOutcome struct {
	Status int
	Code   string
}

type ReplayCounts struct {
	ChangeSets   int
	MutationRows int
	Revisions    int
}

type ReplayExpectation struct {
	FirstStatus     int
	ReplayStatus    int
	DivergentStatus int
	DivergentCode   string
	StableBefore    ReplayCounts
	StableAfter     ReplayCounts
}

func RequireAuthorizationReDerived(t testing.TB, before AuthorizationOutcome, after AuthorizationOutcome) {
	t.Helper()
	if before.Status == after.Status && before.Code == after.Code {
		t.Fatalf("expected authorization outcome to change after re-derivation: before=%+v after=%+v", before, after)
	}
}

func RequireReplayScaffold(t testing.TB, got ReplayExpectation) {
	t.Helper()
	if got.FirstStatus == 0 || got.ReplayStatus == 0 || got.DivergentStatus == 0 {
		t.Fatalf("expected replay scaffold statuses to be populated, got %+v", got)
	}
	if got.DivergentCode == "" {
		t.Fatal("expected divergent replay code")
	}
	if got.StableBefore != got.StableAfter {
		t.Fatalf("expected replay counts to remain stable, before=%+v after=%+v", got.StableBefore, got.StableAfter)
	}
}

func RequireDivergentReplayRejected(t testing.TB, status int, code string, wantCode string) {
	t.Helper()
	if status == 0 {
		t.Fatal("expected divergent replay status")
	}
	if code != wantCode {
		t.Fatalf("unexpected divergent replay code: got %q want %q", code, wantCode)
	}
}

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
