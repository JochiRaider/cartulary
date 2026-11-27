package harnessruntime

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRootLayout(t *testing.T) {
	writer := engineUnit("writer", 1)
	writer.Command.Environment = map[string]string{"CARTULARY_TEST_TARGET": "lint-biome"}
	writer.Evidence = []string{"rows/writer.json", "visual-registry/registry.json"}
	p := enginePlan(writer)
	p.facts.Projections = map[string]projection{"check": {}}
	root := t.TempDir()
	for _, name := range []string{"run-manifest.json", "unit-events.ndjson", "_shared", "unit-artifacts", "check", "lint-biome", "visual-registry", "rows"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := checkRunRootLayout(root, p); err != nil {
		t.Fatal("declared layout rejected", err)
	}
	// Test evidence and fixture copies at the top level are undeclared; a unit
	// keeps them under unit-artifacts/<unit>/.
	for _, name := range []string{"NF-FIX-001-cisco-sna-minimal-1", "adhoc", "job-attempts-1.json"} {
		if err := os.WriteFile(filepath.Join(root, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	err := checkRunRootLayout(root, p)
	var f *failure
	if !errors.As(err, &f) || f.Class != "artifact" || !strings.Contains(err.Error(), "3 undeclared top-level entries") || !strings.Contains(err.Error(), "NF-FIX-001-cisco-sna-minimal-1, adhoc, job-attempts-1.json") {
		t.Fatal("undeclared entries accepted", err)
	}
}
