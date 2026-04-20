package phase0test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func WriteUnreadableRegularFile(t testing.TB, dir string, name string, content []byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write unreadable regular file fixture: %v", err)
	}
	if err := os.Chmod(path, 0o200); err != nil {
		t.Fatalf("chmod unreadable regular file fixture: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(path, 0o600)
	})

	if _, err := os.ReadFile(path); err == nil {
		t.Skip("current environment can still read write-only files; unreadable regular-file assertions are unsupported here")
	} else if !errors.Is(err, fs.ErrPermission) && !os.IsPermission(err) {
		t.Fatalf("expected permission error for unreadable regular file, got %v", err)
	}

	return path
}

func RequireBootstrapUserLocalAuthOnly(t testing.TB, dsn string, userID string, wantEmail string) {
	t.Helper()

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		t.Fatalf("parse bootstrap user id: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("open authn store pool: %v", err)
	}
	defer pool.Close()

	user, err := authn.NewStore(pool).GetUserByID(context.Background(), parsedUserID)
	if err != nil {
		t.Fatalf("load bootstrap user via authn store: %v", err)
	}

	resource := auth.BuildSafeUserResource(user)
	authBindings, ok := resource["auth_bindings"].([]map[string]any)
	if !ok {
		t.Fatalf("expected []map[string]any auth_bindings, got %#v", resource["auth_bindings"])
	}
	if len(authBindings) != 1 {
		t.Fatalf("expected exactly one local auth binding summary, got %#v", authBindings)
	}

	binding := authBindings[0]
	if got := binding["provider_type"]; got != "local" {
		t.Fatalf("unexpected auth binding provider_type: got %v want %q", got, "local")
	}
	if got := binding["provider_key"]; got != "local" {
		t.Fatalf("unexpected auth binding provider_key: got %v want %q", got, "local")
	}
	if got := binding["username"]; got != wantEmail {
		t.Fatalf("unexpected auth binding username: got %v want %q", got, wantEmail)
	}
}
