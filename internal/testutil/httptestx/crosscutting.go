package httptestx

import (
	"reflect"
	"slices"
	"testing"
)

type AuthorizationOutcome struct {
	Allowed bool
	Code    string
}

type MutationAttribution struct {
	Actor     string
	Source    string
	Timestamp string
}

type ReplayExpectation struct {
	FirstStatus     int
	ReplayStatus    int
	DivergentStatus int
	DivergentCode   string
}

func RequireAuthorizationReDerived(t testing.TB, before AuthorizationOutcome, after AuthorizationOutcome) {
	t.Helper()
	if before == after {
		t.Fatalf("expected authorization outcome to change after re-derivation: %+v", before)
	}
}

func RequireMutationAttribution(t testing.TB, got MutationAttribution) {
	t.Helper()
	if got.Actor == "" || got.Source == "" || got.Timestamp == "" {
		t.Fatalf("expected non-empty mutation attribution, got %+v", got)
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
}

func RequireClosedVocabularyRejected(t testing.TB, code string, details map[string]any) {
	t.Helper()
	if code == "" {
		t.Fatal("expected closed-vocabulary rejection code")
	}
	if details == nil {
		t.Fatal("expected closed-vocabulary rejection details")
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
