package configtest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/configassembly"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/diagnosticstest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type TempRoots struct {
	Base  string
	Paths map[string]string
}

const revisionsConflictTokenFixtureSecret = "oVmbXT5kH1Q59Lur9tmdNgYUW3L41EGpcjT73_5CgSQ"

func EnsureRevisionsConflictTokenTestEnvironment(env map[string]string) {
	if env == nil {
		return
	}
	if _, exists := env["CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH"]; !exists {
		env["CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH"] = fixtures.Path("revisions", "conflict-token-key-ring.json")
	}
	if _, exists := env["CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_FIXTURE_V1"]; !exists {
		env["CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_FIXTURE_V1"] = revisionsConflictTokenFixtureSecret
	}
	env[conflicts.ConflictTokenFixtureRuntimeEnvName] = conflicts.ConflictTokenFixtureRuntimeMarker
}

func EffectiveConfigEnv(fixtureParts []string, overlays map[string]string) map[string]string {
	env := map[string]string{
		config.ConfigFileEnv: fixtures.Path(fixtureParts...),
	}
	for key, value := range overlays {
		env[key] = value
	}
	EnsureRevisionsConflictTokenTestEnvironment(env)
	return env
}

func LoadFixture(t testing.TB, fixtureParts []string, overlays map[string]string) configassembly.Loaded {
	t.Helper()

	loaded, err := configassembly.Load(configassembly.LoadOptions{
		Env: EffectiveConfigEnv(fixtureParts, overlays),
	})
	if err != nil {
		t.Fatalf("load config fixture: %v", err)
	}

	return loaded
}

func LoadPath(t testing.TB, path string, overlays map[string]string) configassembly.Loaded {
	t.Helper()
	env := make(map[string]string, len(overlays)+2)
	for key, value := range overlays {
		env[key] = value
	}
	EnsureRevisionsConflictTokenTestEnvironment(env)
	loaded, err := configassembly.Load(configassembly.LoadOptions{Path: path, Env: env})
	if err != nil {
		t.Fatalf("load config path %s: %v", path, err)
	}
	return loaded
}

func LoadInvalidFixture(t testing.TB, fixtureParts []string, overlays map[string]string) error {
	t.Helper()

	_, err := configassembly.Load(configassembly.LoadOptions{
		Env: EffectiveConfigEnv(fixtureParts, overlays),
	})
	if err == nil {
		t.Fatalf("expected invalid config fixture %v to fail", fixtureParts)
	}

	return err
}

func Overlay(pairs ...string) map[string]string {
	overlays := make(map[string]string)
	for i := 0; i+1 < len(pairs); i += 2 {
		overlays[pairs[i]] = pairs[i+1]
	}
	return overlays
}

func RequireDiagnosticsMatchGolden(t testing.TB, err error, goldenParts []string) {
	t.Helper()

	diagnosticstest.RequireJSONMatchesGolden(t, DiagnosticsJSON(t, err), goldenParts)
}

func DiagnosticsJSON(t testing.TB, err error) string {
	t.Helper()

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
	if ok {
		return diagnosticsErr.JSON()
	}

	t.Fatalf("expected diagnostics error, got %T", err)
	return ""
}

func RequireDiagnostic(t testing.TB, err error, wantPath string, wantReasonCode string) {
	t.Helper()

	diagnosticsErr, ok := err.(*config.DiagnosticsError)
	if !ok {
		t.Fatalf("expected diagnostics error, got %T", err)
	}

	for _, diagnostic := range diagnosticsErr.Diagnostics {
		if diagnostic.Path == wantPath && diagnostic.ReasonCode == wantReasonCode {
			return
		}
	}

	t.Fatalf("missing diagnostic path=%q reason_code=%q in %#v", wantPath, wantReasonCode, diagnosticsErr.Diagnostics)
}

func SetupTempRoots(t testing.TB) TempRoots {
	t.Helper()

	base := filepath.Join(t.TempDir(), "cartulary-runtime-roots")
	paths := map[string]string{
		"CARTULARY__ROOTS__DATABASE_STORAGE__PATH":       filepath.Join(base, "database-storage"),
		"CARTULARY__ROOTS__OBJECT_STORAGE__PATH":         filepath.Join(base, "object-storage"),
		"CARTULARY__ROOTS__BACKUP_STORAGE__PATH":         filepath.Join(base, "backup-storage"),
		"CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH": filepath.Join(base, "reference-pack-storage"),
		"CARTULARY__ROOTS__TEMPORARY_WORK__PATH":         filepath.Join(base, "temporary-work"),
		"CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH":         filepath.Join(base, "export-outputs"),
	}

	for _, path := range paths {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatalf("create temp root %s: %v", path, err)
		}
	}

	return TempRoots{
		Base:  base,
		Paths: paths,
	}
}

func BindPostgresDSNToDatabaseRoot(t testing.TB, rootPath string, dsn string) {
	t.Helper()

	if rootPath == "" || dsn == "" || !filepath.IsAbs(rootPath) {
		return
	}
	if err := os.MkdirAll(rootPath, 0o700); err != nil {
		t.Fatalf("create postgres root %s: %v", rootPath, err)
	}
	dsnPath := filepath.Join(rootPath, postgres.FilesystemRootDSNFile)
	if err := os.WriteFile(dsnPath, []byte(dsn+"\n"), 0o600); err != nil {
		t.Fatalf("write root-bound postgres dsn %s: %v", dsnPath, err)
	}
}

func BindPostgresEnvToDatabaseRoot(t testing.TB, rootPath string, env map[string]string) {
	t.Helper()

	if dsn, ok := env[suiteservices.PostgresDSNEnv]; ok {
		BindPostgresDSNToDatabaseRoot(t, rootPath, dsn)
	}
}
