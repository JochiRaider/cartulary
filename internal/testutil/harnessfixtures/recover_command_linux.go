//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
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
	if environment["RESULTS_DIR"] == "" || (environment["RUNTIME_ROOT"] == "") != (environment["RUNTIME_LEASE_ID"] == "") {
		fmt.Fprintln(out, "failure_class=config failure_reason=configuration_error recovery rejected: exact run and complete optional runtime identity required")
		return 2
	}
	run, err := filepath.Abs(environment["RESULTS_DIR"])
	if err != nil {
		return 2
	}
	runtimeRoot, leaseID := environment["RUNTIME_ROOT"], environment["RUNTIME_LEASE_ID"]
	if runtimeRoot == "" {
		locator, err := harnesssession.ResolveRecovery(repository, run, filepath.Base(run), environment["CARTULARY_HARNESS_SCRATCH_ROOT"])
		if err != nil {
			fmt.Fprintln(out, "failure_class=config failure_reason=configuration_error recovery unresolved: exact owner-issued locator unavailable")
			return 2
		}
		if locator.State == "settled" {
			fmt.Fprintln(out, "recovery already settled; original execution result unchanged")
			return 0
		}
		runtimeRoot, leaseID = locator.Root, locator.LeaseID
	}
	outcomes, code, messages := recoverRuntime(ctx, repository, run, runtimeRoot, leaseID, environment)
	for _, message := range messages {
		fmt.Fprintln(out, message)
	}
	// Outcome counts exist only once exact ownership was admitted.
	if code != 2 {
		_ = json.NewEncoder(out).Encode(outcomes)
	}
	return code
}

// RecoverDeadRuntimes settles, before a top-level graph admits work, every
// retained runtime of the repository whose original owner is proven dead. Each
// root is admitted and cleaned exactly as RecoverRun would; a stopped or killed
// run therefore never leaves its services or admission capacity charged for
// later runs. A failure retains that root's exact recovery proof and is
// reported with its explicit recovery command; it does not stop the caller,
// whose admission still fails rather than waiting on capacity no live owner
// can release.
func RecoverDeadRuntimes(ctx context.Context, repository string, environment map[string]string, out io.Writer) {
	runtimes, err := harnesssession.DeadRuntimes(repository, environment["CARTULARY_HARNESS_SCRATCH_ROOT"])
	if err != nil {
		fmt.Fprintln(out, "[RECOVERY-FAIL] failure_class=harness failure_reason=cleanup_error dead runtime discovery failed")
		return
	}
	for _, dead := range runtimes {
		outcomes, code, messages := recoverRuntime(ctx, repository, dead.RunRoot, dead.Root, dead.LeaseID, environment)
		if code == 0 {
			resources := 0
			for _, kinds := range outcomes {
				for _, count := range kinds {
					resources += count
				}
			}
			fmt.Fprintf(out, "[RECOVERED] run_id=%s resources=%d\n", filepath.Base(dead.RunRoot), resources)
			continue
		}
		fmt.Fprintf(out, "[RECOVERY-FAIL] run_id=%s failure_class=harness failure_reason=cleanup_error recover with: make test-run-recover RESULTS_DIR=%s\n", filepath.Base(dead.RunRoot), dead.RunRoot)
		for _, message := range messages {
			fmt.Fprintln(out, message)
		}
	}
}

// recoverRuntime admits one exact retained runtime for recovery and settles it
// through the shared cleanup evaluator. Exit codes follow RecoverRun: 2 when
// ownership or run binding is invalid, 12 when cleanup or settlement is
// incomplete. Original execution evidence is never rewritten.
func recoverRuntime(ctx context.Context, repository, run, runtimeRoot, leaseID string, environment map[string]string) (map[string]map[string]int, int, []string) {
	outcomes := map[string]map[string]int{}
	actor, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return outcomes, 2, nil
	}
	runtime, err := harnesssession.OwnRuntime(repository, run, runtimeRoot, filepath.Base(run), leaseID, actor, true)
	if err != nil {
		return outcomes, 2, []string{"failure_class=config failure_reason=configuration_error recovery rejected: runtime ownership or run binding is not valid"}
	}
	defer runtime.ReleaseHandles()
	p := ProductionOptions{Root: repository, RunRoot: run, Target: "test-run-recover", Runtime: runtime, Environment: environment,
		ReportCleanup: func(result CleanupOutcome) {
			if outcomes[result.Kind] == nil {
				outcomes[result.Kind] = map[string]int{}
			}
			outcomes[result.Kind][result.Outcome]++
		},
	}
	if err := p.Recover(ctx); err != nil {
		return outcomes, 12, []string{fmt.Sprint(harnesssession.DescribeCleanupFailure(err)), "failure_class=harness failure_reason=cleanup_error recovery incomplete: exact ownership proof retained"}
	}
	if err := runtime.Close(); err != nil {
		return outcomes, 12, []string{fmt.Sprint(harnesssession.DescribeCleanupFailure(err)), "failure_class=harness failure_reason=cleanup_error recovery incomplete: runtime settlement not established"}
	}
	return outcomes, 0, []string{"recovery complete; original execution result unchanged"}
}

// RecoveryStatus reports only opaque binding state. Private coordinates remain
// inside the native owner, and an active receipt never grants cleanup authority.
func RecoveryStatus(repository string, environment map[string]string, out io.Writer) int {
	if environment["RESULTS_DIR"] == "" {
		return 2
	}
	run, err := filepath.Abs(environment["RESULTS_DIR"])
	if err != nil {
		return 2
	}
	locator, err := harnesssession.ResolveRecovery(repository, run, filepath.Base(run), environment["CARTULARY_HARNESS_SCRATCH_ROOT"])
	status := "runtime_bound"
	var lease any
	if errors.Is(err, os.ErrNotExist) {
		status = "unresolved"
	} else if err != nil {
		fmt.Fprintln(out, "failure_class=config failure_reason=configuration_error recovery locator rejected")
		return 2
	} else {
		lease = locator.LeaseID
		if locator.State == "settled" {
			status = "settled"
		}
	}
	if err := json.NewEncoder(out).Encode(map[string]any{"schema_id": "cartulary.harness_recovery_status.v1", "run_id": filepath.Base(run), "lease_id": lease, "status": status}); err != nil {
		return 2
	}
	return 0
}
