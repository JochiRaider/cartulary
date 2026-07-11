package auditassert

import (
	"testing"
	"time"
)

type SystemMutationAttribution struct {
	ActorUserID string
	Source      string
	EventKind   string
	RequestID   string
	CreatedAt   time.Time
}

type MutationAttribution struct {
	ActorUserID string
	Source      string
	ClientTxnID string
	RequestID   string
	CreatedAt   time.Time
}

func RequireSystemMutationAttribution(t testing.TB, got SystemMutationAttribution, wantSource string, wantEventKind string) {
	t.Helper()
	if got.Source == "" || got.CreatedAt.IsZero() {
		t.Fatalf("expected non-empty system mutation attribution, got %+v", got)
	}
	if got.ActorUserID != "" {
		t.Fatalf("startup-owned mutation must not record a user actor, got %+v", got)
	}
	if wantSource != "" && got.Source != wantSource {
		t.Fatalf("unexpected system mutation source: got %q want %q", got.Source, wantSource)
	}
	if wantEventKind != "" && got.EventKind != wantEventKind {
		t.Fatalf("unexpected system mutation event kind: got %q want %q", got.EventKind, wantEventKind)
	}
}

func RequireMutationAttribution(t testing.TB, got MutationAttribution, wantActorUserID string, wantSource string, wantClientTxnID string) {
	t.Helper()
	if got.ActorUserID == "" || got.Source == "" || got.RequestID == "" || got.CreatedAt.IsZero() {
		t.Fatalf("expected non-empty mutation attribution, got %+v", got)
	}
	if wantActorUserID != "" && got.ActorUserID != wantActorUserID {
		t.Fatalf("unexpected actor_user_id: got %q want %q", got.ActorUserID, wantActorUserID)
	}
	if wantSource != "" && got.Source != wantSource {
		t.Fatalf("unexpected mutation source: got %q want %q", got.Source, wantSource)
	}
	if wantClientTxnID != "" {
		if got.ClientTxnID == "" {
			t.Fatalf("expected non-empty client_txn_id, got %+v", got)
		}
		if got.ClientTxnID != wantClientTxnID {
			t.Fatalf("unexpected client_txn_id: got %q want %q", got.ClientTxnID, wantClientTxnID)
		}
	}
}
