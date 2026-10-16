//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// RecoverRun admits one exact preserved runtime. It delegates every physical
// cleanup and dependency decision to the same owner used during normal closure.
// Original execution evidence is never rewritten into a successful result.
func RecoverRun(ctx context.Context, repository string, environment map[string]string, out io.Writer) int {
	for _, key := range []string{"RESULTS_DIR", "RUNTIME_ROOT", "RUNTIME_LEASE_ID"} {
		if environment[key] == "" {
			fmt.Fprintln(out, "failure_class=config failure_reason=configuration_error recovery rejected: exact run and runtime identity required")
			return 2
		}
	}
	run, err := filepath.Abs(environment["RESULTS_DIR"])
	if err != nil {
		return 2
	}
	actor, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return 2
	}
	runtime, err := harnesssession.OwnRuntime(repository, run, environment["RUNTIME_ROOT"], filepath.Base(run), environment["RUNTIME_LEASE_ID"], actor, true)
	if err != nil {
		fmt.Fprintln(out, "failure_class=config failure_reason=configuration_error recovery rejected: runtime ownership or run binding is not valid")
		return 2
	}
	defer runtime.ReleaseHandles()
	outcomes := map[string]map[string]int{}
	defer func() { _ = json.NewEncoder(out).Encode(outcomes) }()
	p := ProductionOptions{Root: repository, RunRoot: run, Target: "test-run-recover", Runtime: runtime, Environment: environment,
		ReportCleanup: func(result CleanupOutcome) {
			if outcomes[result.Kind] == nil {
				outcomes[result.Kind] = map[string]int{}
			}
			outcomes[result.Kind][result.Outcome]++
		},
	}
	if err := p.Recover(ctx); err != nil {
		fmt.Fprintln(out, "failure_class=harness failure_reason=cleanup_error recovery incomplete: exact ownership proof retained")
		return 12
	}
	if err := runtime.Close(); err != nil {
		fmt.Fprintln(out, "failure_class=harness failure_reason=cleanup_error recovery incomplete: runtime settlement not established")
		return 12
	}
	fmt.Fprintln(out, "recovery complete; original execution result unchanged")
	return 0
}
