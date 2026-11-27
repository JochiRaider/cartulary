package harnessruntime

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
)

// runRootFamilies are the run-level entries any graph run may retain at its
// root's top level (TH-HARNESS-REQ-874).
var runRootFamilies = map[string]bool{
	"run-manifest.json": true, "plan.json": true, "unit-events.ndjson": true,
	"run-summary.json": true, "retained-secret-scan.json": true, "cleanup-results.json": true,
	"error-code-coverage.json": true, "contract-evidence": true, "performance-fixtures": true,
	"_shared": true, "diagnostics": true, "rows": true, "target-summaries": true,
	"unit-artifacts": true, "unit-logs": true, "unit-results": true,
}

// checkRunRootLayout rejects a top-level entry that no run-level family,
// projected target, unit report target or declared unit evidence output
// names. A unit keeps its own test evidence and fixture copies under
// unit-artifacts/<unit>/, so the top level stays a closed set.
func checkRunRootLayout(runRoot string, plan *Plan) error {
	allowed := maps.Clone(runRootFamilies)
	for name := range plan.facts.Projections {
		allowed[name] = true
	}
	for _, u := range plan.facts.Graph.Units {
		if target := u.Command.Environment["CARTULARY_TEST_TARGET"]; target != "" {
			allowed[target] = true
		}
		for _, output := range u.Evidence {
			allowed[strings.SplitN(output, "/", 2)[0]] = true
		}
	}
	entries, err := os.ReadDir(runRoot)
	if err != nil {
		return errors.Join(failed("artifact", "artifact_error"), err)
	}
	undeclared := []string{}
	for _, entry := range entries {
		if !allowed[entry.Name()] {
			undeclared = append(undeclared, entry.Name())
		}
	}
	if len(undeclared) == 0 {
		return nil
	}
	slices.Sort(undeclared)
	shown := undeclared[:min(len(undeclared), 8)]
	return errors.Join(failed("artifact", "artifact_error"),
		fmt.Errorf("run root has %d undeclared top-level entries (TH-HARNESS-REQ-874): %s", len(undeclared), strings.Join(shown, ", ")))
}
