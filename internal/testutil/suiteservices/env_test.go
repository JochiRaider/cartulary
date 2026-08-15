package suiteservices

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveResultsRootCanonicalPaths(t *testing.T) {
	repoRoot, err := FindRepoRoot()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("absolute", func(t *testing.T) {
		absolute := filepath.Join(t.TempDir(), "results")
		got, err := ResolveResultsRoot(map[string]string{testResultsDirEnv: absolute})
		if err != nil || got != absolute {
			t.Fatalf("ResolveResultsRoot absolute got %q err=%v want %q", got, err, absolute)
		}
	})

	t.Run("relative", func(t *testing.T) {
		relative := filepath.Join(".cartulary", "test-results", "suiteservices-relative")
		got, err := ResolveResultsRoot(map[string]string{testResultsDirEnv: relative})
		want := filepath.Join(repoRoot, relative)
		if err != nil || got != want {
			t.Fatalf("ResolveResultsRoot relative got %q err=%v want %q", got, err, want)
		}
	})

	t.Run("missing configuration defaults under repository", func(t *testing.T) {
		got, err := ResolveResultsRoot(map[string]string{})
		want := filepath.Join(repoRoot, ".cartulary", "test-results")
		if err != nil || got != want {
			t.Fatalf("ResolveResultsRoot default got %q err=%v want %q", got, err, want)
		}
	})
}

func TestResolveSuiteRuntimeDirRejectsUnsafeRoots(t *testing.T) {
	resultsRoot := t.TempDir()
	baseEnv := map[string]string{
		SuiteIDEnv:        "suite-runtime-contract",
		testResultsDirEnv: resultsRoot,
		testRunIDEnv:      "run-runtime-contract",
	}

	t.Run("external private root", func(t *testing.T) {
		runtimeRoot := t.TempDir()
		if err := os.Chmod(runtimeRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		owner := fmt.Sprintf("{\"schema_id\":\"cartulary.harness_suite_runtime_owner.v1\",\"lease_id\":\"00000000-0000-4000-8000-000000000001\",\"run_id\":\"run-runtime-contract\",\"owner_uid\":%d,\"created_at\":\"2026-08-14T00:00:00Z\"}\n", os.Getuid())
		if err := os.WriteFile(filepath.Join(runtimeRoot, "runtime-owner.json"), []byte(owner), 0o600); err != nil {
			t.Fatal(err)
		}
		env := map[string]string{}
		for key, value := range baseEnv {
			env[key] = value
		}
		env[SuiteRuntimeRootEnv] = runtimeRoot
		env[SuiteRuntimeLeaseIDEnv] = "00000000-0000-4000-8000-000000000001"
		env[SuiteRuntimeRunIDEnv] = "run-runtime-contract"
		got, ok, err := ResolveSuiteRuntimeDir(env)
		if err != nil || !ok || got != filepath.Join(runtimeRoot, "test-services") {
			t.Fatalf("ResolveSuiteRuntimeDir got %q ok=%v err=%v", got, ok, err)
		}
		info, err := os.Lstat(got)
		if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
			t.Fatalf("private service root mode is invalid: info=%v err=%v", info, err)
		}
	})

	t.Run("missing", func(t *testing.T) {
		_, _, err := ResolveSuiteRuntimeDir(baseEnv)
		if err == nil || !strings.Contains(err.Error(), SuiteRuntimeRootEnv) {
			t.Fatalf("missing runtime root error = %v", err)
		}
	})

	t.Run("retained containment", func(t *testing.T) {
		runtimeRoot := filepath.Join(resultsRoot, "private")
		if err := os.Mkdir(runtimeRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		env := map[string]string{}
		for key, value := range baseEnv {
			env[key] = value
		}
		env[SuiteRuntimeRootEnv] = runtimeRoot
		if _, _, err := ResolveSuiteRuntimeDir(env); err == nil || !strings.Contains(err.Error(), "outside") {
			t.Fatalf("retained-contained runtime root error = %v", err)
		}
	})

	t.Run("symlink", func(t *testing.T) {
		target := t.TempDir()
		link := filepath.Join(t.TempDir(), "runtime-link")
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}
		env := map[string]string{}
		for key, value := range baseEnv {
			env[key] = value
		}
		env[SuiteRuntimeRootEnv] = link
		if _, _, err := ResolveSuiteRuntimeDir(env); err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("symlink runtime root error = %v", err)
		}
	})

	t.Run("permissions", func(t *testing.T) {
		runtimeRoot := t.TempDir()
		if err := os.Chmod(runtimeRoot, 0o755); err != nil {
			t.Fatal(err)
		}
		env := map[string]string{}
		for key, value := range baseEnv {
			env[key] = value
		}
		env[SuiteRuntimeRootEnv] = runtimeRoot
		if _, _, err := ResolveSuiteRuntimeDir(env); err == nil || !strings.Contains(err.Error(), "0700") {
			t.Fatalf("permissive runtime root error = %v", err)
		}
	})
}

func TestResolveRunIDAndTargetEnvironment(t *testing.T) {
	env := map[string]string{
		testRunIDEnv: " explicit-run ",
		TargetEnv:    "backend-process",
	}
	if got := ResolveRunID(env); got != "explicit-run" {
		t.Fatalf("ResolveRunID explicit got %q", got)
	}
	if got := LookupEnvValue(env, TargetEnv); got != "backend-process" {
		t.Fatalf("LookupEnvValue target got %q", got)
	}
	if got := ResolveRunID(map[string]string{}); got != "adhoc" {
		t.Fatalf("ResolveRunID default got %q want adhoc", got)
	}
}
