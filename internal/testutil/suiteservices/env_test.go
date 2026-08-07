package suiteservices

import (
	"path/filepath"
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
