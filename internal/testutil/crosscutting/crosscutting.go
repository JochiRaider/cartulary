package crosscutting

import (
	"fmt"
	"slices"
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

func RequireSecretSafePayload(t testing.TB, payload map[string]any, forbiddenKeys []string) {
	t.Helper()
	requireSecretSafeValue(t, payload, forbiddenKeys, "")
}

func requireSecretSafeValue(t testing.TB, value any, forbiddenKeys []string, path string) {
	t.Helper()

	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if slices.Contains(forbiddenKeys, key) {
				t.Fatalf("payload exposed forbidden key %q at %s", key, joinPath(path, key))
			}
			requireSecretSafeValue(t, item, forbiddenKeys, joinPath(path, key))
		}
	case []any:
		for index, item := range typed {
			requireSecretSafeValue(t, item, forbiddenKeys, joinPath(path, index))
		}
	}
}

func joinPath(path string, part any) string {
	if path == "" {
		return toPathPart(part)
	}
	return path + "." + toPathPart(part)
}

func toPathPart(part any) string {
	switch typed := part.(type) {
	case string:
		return typed
	default:
		return "[" + fmt.Sprint(part) + "]"
	}
}
