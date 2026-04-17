package httptestx

import (
	"reflect"
	"slices"
	"testing"
	"time"
)

type AuthorizationOutcome struct {
	Status int
	Code   string
}

type MutationAttribution struct {
	ActorUserID string
	Source      string
	ClientTxnID string
	RequestID   string
	CreatedAt   time.Time
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

func RequireMutationAttribution(t testing.TB, got MutationAttribution, wantActorUserID string, wantSource string, wantClientTxnID string) {
	t.Helper()
	if got.ActorUserID == "" || got.Source == "" || got.ClientTxnID == "" || got.RequestID == "" || got.CreatedAt.IsZero() {
		t.Fatalf("expected non-empty mutation attribution, got %+v", got)
	}
	if wantActorUserID != "" && got.ActorUserID != wantActorUserID {
		t.Fatalf("unexpected actor_user_id: got %q want %q", got.ActorUserID, wantActorUserID)
	}
	if wantSource != "" && got.Source != wantSource {
		t.Fatalf("unexpected mutation source: got %q want %q", got.Source, wantSource)
	}
	if wantClientTxnID != "" && got.ClientTxnID != wantClientTxnID {
		t.Fatalf("unexpected client_txn_id: got %q want %q", got.ClientTxnID, wantClientTxnID)
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

func RequireClosedVocabularyRejected(t testing.TB, code string, details map[string]any, wantField string, wantReasonCode string) {
	t.Helper()
	if code == "" {
		t.Fatal("expected closed-vocabulary rejection code")
	}
	if code != "invalid_mutation_payload" && code != "invalid_view_query" {
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
