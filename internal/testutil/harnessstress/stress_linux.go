//go:build linux

// Package harnessstress runs exact test rows repeatedly beside a declared
// synthetic host load (TH-HARNESS-REQ-871). It finds load-dependent defects
// without an aggregate run: each repetition is an ordinary owner slice with
// its own run root, so its failures carry the usual evidence and the
// failure-scoped host pressure of TH-HARNESS-REQ-870.
package harnessstress

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/runcontext"
)

const (
	summarySchema = "cartulary.harness_stress_summary.v1"
	maxRepeat     = 50
	// Repetitions run the owner slice command.
	sliceTarget = "test-slice"
	// sliceStressTarget is this command's target directory in its run root.
	sliceStressTarget = "test-stress"
)

var runIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,95}$`)

// repetition is one slice run's outcome.
type repetition struct {
	Index          int      `json:"index"`
	RunID          string   `json:"run_id"`
	Status         string   `json:"status"`
	ExitCode       int      `json:"exit_code"`
	WallDurationMS int64    `json:"wall_duration_ms"`
	FailedRows     []string `json:"failed_rows"`
}

// sliceRunner runs one owner slice invocation; production uses the native
// runtime in process.
type sliceRunner func(ctx context.Context, environment []string, freshRows []string) (status string, code int)

// Main runs the stress command. Inputs arrive as the Make wrapper's
// environment, exactly as for an owner slice.
func Main(ctx context.Context, root, binary string, environment []string, stdout, stderr io.Writer) int {
	run := func(ctx context.Context, child []string, fresh []string) (string, int) {
		report, code, err := harnessruntime.RunWithOptions(ctx, root, binary, []string{"--selection", "command", "--target", sliceTarget}, child, stderr,
			harnessruntime.RunOptions{FreshRows: fresh})
		status := "pass"
		if report != nil {
			if value, ok := report.Summary["status"].(string); ok {
				status = value
			}
		}
		if (code != 0 || err != nil) && status == "pass" {
			status = "fail"
		}
		if code == 0 && err != nil {
			code = 1
		}
		return status, code
	}
	return stress(ctx, root, binary, environment, stdout, stderr, run)
}

func stress(ctx context.Context, root, binary string, environment []string, stdout, stderr io.Writer, run sliceRunner) int {
	env := map[string]string{}
	for _, entry := range environment {
		if key, value, ok := strings.Cut(entry, "="); ok {
			env[key] = value
		}
	}
	owner, rows := strings.TrimSpace(env["OWNER"]), strings.TrimSpace(env["ROWS"])
	repeat, err := strconv.Atoi(defaulted(env["REPEAT"], "10"))
	profile, known := profiles[defaulted(env["LOAD"], "mixed")]
	base := env["CARTULARY_TEST_RUN_ID"]
	if base == "" {
		base = fmt.Sprintf("%s-p%d", time.Now().UTC().Format("20060102T150405000Z"), os.Getpid())
	}
	if owner == "" || rows == "" || err != nil || repeat < 1 || repeat > maxRepeat || !known || !runIDPattern.MatchString(base) {
		fmt.Fprintln(stderr, "failure_class=config failure_reason=usage_error")
		return 2
	}
	results := defaulted(env["CARTULARY_TEST_RESULTS_DIR"], ".cartulary/test-results")
	if !filepath.IsAbs(results) {
		results = filepath.Join(root, results)
	}
	// The command's own run root belongs to its step wrapper; the summary
	// lives in the command's target directory, and each repetition is a
	// sibling run root that must not exist yet.
	stressRoot := filepath.Join(results, base, sliceStressTarget)
	for index := 1; index <= repeat; index++ {
		if _, err := os.Lstat(filepath.Join(results, fmt.Sprintf("%s-r%02d", base, index))); !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(stderr, "failure_class=config failure_reason=usage_error")
			return 2
		}
	}
	if _, err := os.Lstat(filepath.Join(stressRoot, "stress-summary.json")); !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(stderr, "failure_class=config failure_reason=usage_error")
		return 2
	}
	for _, directory := range []string{filepath.Join(results, base), stressRoot} {
		handle, err := harnesslocal.PrivateDirectory(directory)
		if err != nil {
			fmt.Fprintln(stderr, "failure_class=artifact failure_reason=artifact_error")
			return 11
		}
		handle.Close()
	}
	stopLoad, err := startLoad(binary, profile, filepath.Join(stressRoot, "load"))
	if err != nil {
		fmt.Fprintln(stderr, "failure_class=harness failure_reason=fixture_error")
		return 11
	}
	defer stopLoad()

	summary := map[string]any{"schema_id": summarySchema, "run_id": base, "owner_id": owner, "row_ids": strings.Split(rows, ","),
		"load": profile.name, "load_profile": profile.document(), "repeat": repeat}
	outcomes := []repetition{}
	exit := 0
	for index := 1; index <= repeat && ctx.Err() == nil; index++ {
		runID := fmt.Sprintf("%s-r%02d", base, index)
		child := sliceEnvironment(env, owner, rows, runID)
		started := time.Now()
		status, code := run(ctx, child, strings.Split(rows, ","))
		if !slices.Contains([]string{"pass", "fail", "cancelled"}, status) || code != 0 && status == "pass" {
			status = "fail"
		}
		outcome := repetition{Index: index, RunID: runID, Status: status, ExitCode: code, WallDurationMS: time.Since(started).Milliseconds(), FailedRows: []string{}}
		if code != 0 {
			outcome.FailedRows = failedRows(filepath.Join(results, runID))
			if exit == 0 {
				exit = code
			}
		}
		outcomes = append(outcomes, outcome)
		fmt.Fprintf(stdout, "[STRESS] repetition=%d/%d run_id=%s status=%s duration_ms=%d failed_rows=%d\n", index, repeat, runID, outcome.Status, outcome.WallDurationMS, len(outcome.FailedRows))
	}
	if ctx.Err() != nil && exit == 0 {
		exit = 130
	}
	passed := 0
	for _, outcome := range outcomes {
		if outcome.Status == "pass" {
			passed++
		}
	}
	summary["repetitions"], summary["passed"], summary["failed"] = outcomes, passed, len(outcomes)-passed
	if err := writeSummary(filepath.Join(stressRoot, "stress-summary.json"), summary); err != nil {
		fmt.Fprintln(stderr, "failure_class=artifact failure_reason=artifact_error")
		return max(exit, 11)
	}
	fmt.Fprintf(stdout, "[STRESS-SUMMARY] run_id=%s load=%s repetitions=%d passed=%d failed=%d summary=%s\n", base, profile.name, len(outcomes), passed, len(outcomes)-passed, filepath.Join(stressRoot, "stress-summary.json"))
	return exit
}

func defaulted(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

// sliceEnvironment is an owner slice invocation's environment: the stress
// command's own environment with the slice's selection and a fresh run ID.
// The caller's cache mode applies to preparation work; the selected rows
// always execute (RunOptions.FreshRows).
func sliceEnvironment(env map[string]string, owner, rows, runID string) []string {
	child := maps.Clone(env)
	for _, key := range []string{"REPEAT", "LOAD"} {
		delete(child, key)
	}
	child["OWNER"], child["ROWS"] = owner, rows
	runcontext.Context{RunID: runID, Target: sliceTarget}.Apply(child)
	sources := []string{"OWNER=cli", "ROWS=cli"}
	for _, input := range []string{"CARTULARY_HARNESS_CACHE_MODE", "HARNESS_DIAGNOSTICS"} {
		if child[input] != "" {
			sources = append(sources, input+"=cli")
		}
	}
	maps.Copy(child, runcontext.InputSources(sources))
	return runcontext.Entries(child)
}

// failedRows lists the rows of a repetition that did not pass, from its
// published row results.
func failedRows(runRoot string) []string {
	entries, err := os.ReadDir(filepath.Join(runRoot, "rows"))
	if err != nil {
		return []string{}
	}
	failed := []string{}
	for _, entry := range entries {
		data, err := harnesslocal.ReadFile(filepath.Join(runRoot, "rows", entry.Name()), 1<<20, false)
		if err != nil {
			continue
		}
		var row struct {
			RowID         string `json:"row_id"`
			TerminalState string `json:"terminal_state"`
		}
		if json.Unmarshal(data, &row) == nil && row.TerminalState != "passed" {
			failed = append(failed, row.RowID)
		}
	}
	slices.Sort(failed)
	return failed
}

func writeSummary(path string, summary map[string]any) error {
	data, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	value, err := harnessschema.Decode(data)
	if err == nil {
		err = harnessschema.Validate(summarySchema, value)
	}
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFile(path, append(data, '\n'))
}

// startLoad starts the load generator in its own process group. The child
// dies with this process, and stop kills its group and removes its files.
func startLoad(binary string, profile loadProfile, directory string) (func(), error) {
	if profile.name == "none" {
		return func() {}, nil
	}
	handle, err := harnesslocal.PrivateDirectory(directory)
	if err != nil {
		return nil, err
	}
	handle.Close()
	command := exec.Command(binary, "stress-load", "--profile", profile.name, "--directory", directory)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	if err := command.Start(); err != nil {
		return nil, err
	}
	return func() {
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		_ = command.Wait()
		_ = os.RemoveAll(directory)
	}, nil
}
