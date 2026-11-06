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
	actor, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return 2
	}
	runtime, err := harnesssession.OwnRuntime(repository, run, runtimeRoot, filepath.Base(run), leaseID, actor, true)
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
		fmt.Fprintln(out, harnesssession.DescribeCleanupFailure(err))
		fmt.Fprintln(out, "failure_class=harness failure_reason=cleanup_error recovery incomplete: exact ownership proof retained")
		return 12
	}
	if err := runtime.Close(); err != nil {
		fmt.Fprintln(out, harnesssession.DescribeCleanupFailure(err))
		fmt.Fprintln(out, "failure_class=harness failure_reason=cleanup_error recovery incomplete: runtime settlement not established")
		return 12
	}
	fmt.Fprintln(out, "recovery complete; original execution result unchanged")
	return 0
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
