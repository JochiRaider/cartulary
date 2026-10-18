//go:build linux

package harnessruntime

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// RunResult keeps the published summary separate from resolved presentation inputs.
// Presentation never changes the graph's execution identity or evidence schema.
type RunResult struct {
	Summary           map[string]any
	JSONOutput        bool
	CleanupDiagnostic *harnesssession.CleanupDiagnostic
}

// Run accepts only ordinary graph selectors. Prepare owns their interpretation;
// no external executable plan, shell fragment or runtime-selection flag exists.
func Run(ctx context.Context, root, binary string, args, environment []string) (report *RunResult, code int, resultErr error) {
	env := map[string]string{}
	for _, entry := range environment {
		if k, v, ok := strings.Cut(entry, "="); ok {
			env[k] = v
		}
	}
	mode := env["HARNESS_DIAGNOSTICS"]
	if mode == "" {
		mode = "off"
	}
	if mode != "off" && mode != "basic" {
		return nil, 2, failed("config", "configuration_error")
	}
	diagnostics := newGraphObservation(mode == "basic")
	if env["CARTULARY_TEST_RUN_ID"] == "" {
		env["CARTULARY_TEST_RUN_ID"] = fmt.Sprintf("%s-p%d", time.Now().UTC().Format("20060102T150405000Z"), os.Getpid())
	}
	input := []string{}
	for k, v := range env {
		input = append(input, k+"="+v)
	}
	plan, err := Prepare(ctx, root, args, input)
	if err != nil {
		var lifecycle *failure
		if errors.As(err, &lifecycle) && lifecycle.Reason == "cleanup_error" {
			return nil, 12, err
		}
		if errors.Is(err, context.Canceled) {
			return nil, 130, err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, 13, err
		}
		return nil, 2, errors.Join(failed("config", "configuration_error"), err)
	}
	// This is an externally measured preparation interval in the graph clock.
	// Compiler subphases have a different clock and are never fabricated as
	// exact graph-clock positions or added to canonical execution duration.
	diagnostics.mark("runtime_prepare")
	policy, err := readInstrumentationPolicy(root)
	if err != nil {
		return nil, 11, err
	}
	digest, err := semanticDigest(policy)
	if err != nil || digest != plan.facts.Manifest["instrumentation"].(map[string]any)["policy_digest"] {
		return nil, 11, errPlan
	}
	results := env["CARTULARY_TEST_RESULTS_DIR"]
	if results == "" {
		results = ".cartulary/test-results"
	}
	if !filepath.IsAbs(results) {
		results = filepath.Join(root, results)
	}
	runID := plan.facts.Manifest["run_id"].(string)
	runRoot := filepath.Join(results, runID)
	publication, err := newGraphPublication(plan, runRoot)
	if err != nil {
		return nil, 11, err
	}
	defer publication.Abort()
	diagnostics.mark("manifest_publication")
	runtime, err := harnesssession.NewRuntime(root, runRoot, runID, env["CARTULARY_HARNESS_SCRATCH_ROOT"])
	if err != nil {
		return nil, 11, err
	}
	defer runtime.ReleaseHandles()
	admission := filepath.Join(os.TempDir(), fmt.Sprintf("cartulary-host-admission-%d", os.Getuid()))
	// Admission's own transaction creates and validates this root. Collection
	// may start before the first transaction and waits for that existing owner.
	diagnostics.start(binary, runRoot, admission, runtime, plan, policy, env)
	publication.Forbidden = runtime.ForbiddenValues
	child := plan.unitEnvironment(env)
	for _, k := range []string{"CARTULARY_TEST_SERVICES_MODE", "CARTULARY_TEST_SERVICES_PERSISTENT_BORROWER", "CARTULARY_TEST_SERVICES_SESSION_FILE", "HARNESS_DIAGNOSTICS", "MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS"} {
		delete(child, k)
	}
	target := plan.facts.Manifest["target"].(string)
	maps.Copy(child, plan.facts.Bindings)
	delete(child, "CARTULARY_TEST_SERVICES_MODE")
	delete(child, "CARTULARY_TEST_SERVICES_SESSION_FILE")
	maps.Copy(child, runtime.Environment())
	// The graph starts with no fixture consumers. A fixture lease supplies its
	// exact retained resource references when it extends this environment.
	child["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"] = "[]"
	child["CARTULARY_TEST_RESULTS_DIR"] = results
	child["CARTULARY_TEST_RUN_ID"] = runID
	for _, k := range []string{"CARTULARY_HARNESS_GRAPH_CHILD", "CARTULARY_HARNESS_IDENTITY_PREPARED", "CARTULARY_HARNESS_SKIP_PREREQUISITES", "CARTULARY_SUPPRESS_CHILD_SUCCESS"} {
		child[k] = "1"
	}
	// A nested graph cannot inherit its parent's event authority. Each launch
	// receives this publisher's complete-record prefix below.
	delete(child, "CARTULARY_HARNESS_LIVE_UNIT_EVENTS_FILE")
	delete(child, "CARTULARY_HARNESS_UNIT_EVENT_PREFIX")
	if mode == "basic" {
		child["CARTULARY_HARNESS_EXECUTION_ENABLED"] = "1"
	} else {
		delete(child, "CARTULARY_HARNESS_EXECUTION_ENABLED")
	}
	fixtureEnv := maps.Clone(child)
	services := map[string]bool{}
	for _, unit := range plan.facts.Graph.Units {
		for _, service := range unit.Services {
			services[service] = true
		}
	}
	// The shared physical owner acquires the union, never a fixed service pair.
	fixtureEnv["CARTULARY_HARNESS_SERVICE_DEPENDENCIES"] = strings.Join(slices.Sorted(maps.Keys(services)), ",")
	maps.Copy(fixtureEnv, plan.facts.Bindings)
	declared := plan.facts.Manifest["declared_inputs"].(map[string]any)
	sources := strings.Fields(env["CARTULARY_MAKE_INPUT_SOURCES"])
	for _, k := range []string{"OWNER", "ROWS", "SCOPE", "SUITE"} {
		if v, ok := declared[k].(string); ok {
			fixtureEnv[k] = v
			if k != "SCOPE" {
				next := []string{}
				for _, s := range sources {
					if !strings.HasPrefix(s, k+"=") {
						next = append(next, s)
					}
				}
				sources = append(next, k+"=cli")
			}
		}
	}
	fixtureEnv["CARTULARY_MAKE_INPUT_SOURCES"] = strings.Join(sources, " ")
	var cleanupMu sync.Mutex
	var fixtureCleanupDiagnostic *harnesssession.CleanupDiagnostic
	cleanup := []cleanupRecord{}
	recordCleanup := func(r cleanupRecord) { cleanupMu.Lock(); defer cleanupMu.Unlock(); cleanup = append(cleanup, r) }
	production := harnessfixtures.NewProduction(harnessfixtures.ProductionOptions{Root: root, RunRoot: runRoot, Target: target, Runtime: runtime, Environment: fixtureEnv, ContainerProof: diagnostics.collector.containers}, harnessfixtures.Options{Observe: diagnostics.execution.fixture, Cleanup: func(r harnessfixtures.CleanupRecord) {
		v := cleanupRecord{Operation: r.Operation, UnitID: r.UnitID, LeaseID: r.LeaseID, Outcome: "completed"}
		if r.Err != nil {
			v.Outcome = "failed"
			v.Failure = cleanupFailureOf(r.Err)
			cleanupMu.Lock()
			if fixtureCleanupDiagnostic == nil {
				fixtureCleanupDiagnostic = harnesssession.DescribeCleanupFailure(r.Err)
			}
			cleanupMu.Unlock()
		}
		recordCleanup(v)
	}})
	fixtures := &nativeFixtures{production: production}
	finalized, scanned := false, false
	var primary *failure
	secrets := []string{}
	var runtimeCloseErr error
	closeRuntime := func(unresolved bool) (closeErr error) {
		defer func() { runtimeCloseErr = closeErr }()
		diagnostics.execution.capture()
		var secretErr error
		secrets, secretErr = runtime.ForbiddenValues()
		resources, err := runtime.Resources(4096)
		unresolved = unresolved || err != nil || len(resources) > 0
		if unresolved {
			return errors.Join(err, secretErr, runtime.PreserveRecovery())
		}
		return errors.Join(secretErr, runtime.Close())
	}
	defer func() {
		secondary := []error{}
		if !finalized {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			e := production.Close(cleanupCtx)
			cancel()
			secondary = append(secondary, e, closeRuntime(fixtures.Unresolved()))
			finalized = true
		}
		if !scanned {
			secondary = append(secondary, diagnostics.scan(secrets))
		}
		diagnostics.wrapper.Complete()
		if e := errors.Join(secondary...); e != nil {
			resultErr = errors.Join(resultErr, e)
			if primary != nil {
				resultErr = errors.Join(primary, resultErr)
			}
			if code == 0 {
				code = failureExitCode(failureOf(e), "")
			}
		}
		if report != nil {
			cleanupMu.Lock()
			report.CleanupDiagnostic = fixtureCleanupDiagnostic
			cleanupMu.Unlock()
			if report.CleanupDiagnostic == nil {
				report.CleanupDiagnostic = harnesssession.DescribeCleanupFailure(runtimeCloseErr)
			}
		}
	}()
	cache := newNativeCache(plan, root, runRoot)
	defer cache.Close()
	commands := map[string]string{}
	for _, unit := range plan.facts.Graph.Units {
		commands[unit.Command.Environment["CARTULARY_TEST_TARGET"]] = plan.facts.Manifest["command_id"].(string)
	}
	for name, projection := range plan.facts.Projections {
		commands[name] = projection.CommandID
	}
	processes := processRunner{Root: root, Environment: child, Commands: commands, FailureScope: func(identity map[string]any) (failureScope, error) {
		return createFailureScope(runtime.Directory, identity)
	}, Observe: diagnostics.execution.launch}
	var parent *string
	if v, ok := env["CARTULARY_HOST_ADMISSION_LEASE"]; ok {
		if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(v) {
			return nil, 4, failed("infra", "resource_conflict")
		}
		parent = &v
	}
	diagnostics.canonicalEpoch = monotonicNS()
	diagnostics.mark("runtime_prepare")
	requirements := []harnessfixtures.WorkRequirement{}
	for _, unit := range plan.facts.Graph.Units {
		requirements = append(requirements, workRequirement(unit))
	}
	production.ConfigureAdmission(admission, parent, plan.facts.Capacities, requirements)

	ports := enginePorts{Now: func() time.Duration { return time.Duration(monotonicNS() - diagnostics.canonicalEpoch) }, Execute: func(ctx context.Context, r executionRequest) unitResult {
		artifact := filepath.Join(runRoot, "unit-artifacts", safeUnitName.ReplaceAllString(r.Unit.ID, "-"))
		d, err := harnesslocal.PrivateDirectory(artifact)
		if err != nil {
			return unitResult{Status: "failed", ExitCode: 11, Failure: failed("artifact", "artifact_error")}
		}
		d.Close()
		r.Environment = maps.Clone(r.Environment)
		if r.Environment == nil {
			r.Environment = map[string]string{}
		}
		r.Environment["CARTULARY_WORK_UNIT_ID"] = r.Unit.ID
		r.Environment["CARTULARY_STEP_ARTIFACT_DIR"] = artifact
		prefix, err := publication.Prefix()
		if err != nil {
			return unitResult{Status: "failed", ExitCode: 11, Failure: failed("artifact", "artifact_error")}
		}
		r.Environment["CARTULARY_HARNESS_UNIT_EVENT_PREFIX"] = prefix
		return processes.Execute(ctx, r)
	}, Fixtures: fixtures, Cache: cache, Admit: nativeAdmission(plan.facts.Capacities, parent, admission, production), Terminal: publication.Terminal,
		Cleanup: recordCleanup, CleanupServices: production.Close, Finalize: func(_ context.Context, unresolved bool, _ []cleanupRecord) error {
			err := closeRuntime(unresolved)
			finalized = true
			return err
		}, Pause: diagnostics.pause, Resume: diagnostics.resume, Observe: diagnostics.live.update,
		Event: func(event map[string]any) error {
			if seq, ok := event["seq"].(int64); ok {
				diagnostics.live.transition(seq)
			}
			if event["event"] == "run_completed" {
				diagnostics.mark("canonical_execution_reference")
			}
			return publication.Event(event)
		}}
	result, err := executeGraph(ctx, plan, ports)
	primary = result.Failure
	if err != nil {
		if primary != nil {
			return nil, failureExitCode(primary, ""), errors.Join(primary, err)
		}
		return nil, failureExitCode(failureOf(err), ""), err
	}
	cleanupMu.Lock()
	result.Cleanup = append([]cleanupRecord{}, cleanup...)
	cleanupMu.Unlock()
	summary, err := publication.Finish(result)
	if err != nil {
		if primary != nil {
			return nil, failureExitCode(primary, ""), errors.Join(primary, err)
		}
		return nil, 11, err
	}
	report = &RunResult{Summary: summary, JSONOutput: declared["JSON"] == "1"}
	diagnostics.mark("evidence_publication")
	scanned = true
	err = diagnostics.scan(secrets)
	if result.Status != "passed" {
		code = failureExitCode(result.Failure, "")
		if summary["status"] == "cancelled" {
			code = 130
		}
	}
	if err != nil {
		if code == 0 {
			code = failureExitCode(failureOf(err), "")
		}
		if result.Failure != nil {
			err = errors.Join(result.Failure, err)
		}
		return report, code, err
	}
	return report, code, nil
}
