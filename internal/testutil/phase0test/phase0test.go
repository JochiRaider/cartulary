package phase0test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func WriteBootstrapManifest(t testing.TB, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "bootstrap-admin.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write bootstrap manifest: %v", err)
	}
	return path
}

func WriteMalformedBootstrapManifest(t testing.TB) string {
	t.Helper()
	return WriteBootstrapManifest(t, `{"bootstrap_schema_id":`)
}

func WriteWrongSchemaBootstrapManifest(t testing.TB) string {
	t.Helper()
	return WriteBootstrapManifest(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v2","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!"}`)
}

func WriteExplicitFalseMFABootstrapManifest(t testing.TB) string {
	t.Helper()
	return WriteBootstrapManifest(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","mfa_required":false}`)
}

func WriteUnknownMemberBootstrapManifest(t testing.TB) string {
	t.Helper()
	return WriteBootstrapManifest(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","unexpected":"surprise"}`)
}

func WriteForbiddenIncidentMembershipBootstrapManifest(t testing.TB) string {
	t.Helper()
	return WriteBootstrapManifest(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","incident_memberships":[{"incident_id":"11111111-1111-1111-1111-111111111111","role":"admin"}]}`)
}

func WriteForbiddenProviderBootstrapManifest(t testing.TB) string {
	t.Helper()
	return WriteBootstrapManifest(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","provider_subject":"oidc-subject"}`)
}

func WriteForbiddenDeploymentAdminBootstrapManifest(t testing.TB) string {
	t.Helper()
	return WriteBootstrapManifest(t, `{"bootstrap_schema_id":"cartulary.bootstrap_admin.v1","bootstrap_artifact_id":"11111111-1111-1111-1111-111111111111","email":"bootstrap-admin@example.test","display_name":"Bootstrap Admin","initial_password":"BootstrapPass1!","is_deployment_admin":true}`)
}

func WriteNonRegularBootstrapManifestPath(t testing.TB) string {
	t.Helper()
	return t.TempDir()
}

func WriteUnreadableBootstrapManifest(t testing.TB) string {
	t.Helper()

	path := WriteBootstrapManifest(t, string(fixtures.MustRead("bootstrap-admin", "canonical.json")))
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("chmod unreadable bootstrap manifest: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(path, 0o600)
	})

	return path
}

func CanonicalBootstrapManifestPath() string {
	return fixtures.Path("bootstrap-admin", "canonical.json")
}

func SeedBootstrapEmailConflict(t testing.TB, db *sql.DB) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `INSERT INTO users (email, display_name, password_hash) VALUES ($1, $2, $3)`, "bootstrap-admin@example.test", "Existing User", "existing-hash"); err != nil {
		t.Fatalf("seed conflicting user: %v", err)
	}
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
