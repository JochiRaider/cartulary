package securityassert

import (
	"fmt"
	"slices"
	"testing"
)

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
